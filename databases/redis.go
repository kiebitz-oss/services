// Kiebitz - Privacy-Friendly Appointment Scheduling
// Copyright (C) 2021-2021 The Kiebitz Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package databases

import (
	"context"
	"errors"
	"fmt"
	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
	"github.com/kiebitz-oss/services"
	"github.com/kiprotect/go-helpers/forms"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"sync"
	"time"
)

type Redis struct {
	client   redis.UniversalClient
	options  redis.UniversalOptions
	pipeline redis.Pipeliner
	mutex    sync.Mutex
	channel  chan bool
	ctx      context.Context
}

type RedisLock struct {
	lockKey     string
	client      redis.UniversalClient
	ctx         context.Context
	dLockClient *redislock.Client
	dLock       *redislock.Lock
}

func MakeRedisLock(ctx context.Context, lockKey string, client redis.UniversalClient) (RedisLock, error) {

	rL := RedisLock{
		client:      client,
		lockKey:     lockKey,
		ctx:         ctx,
		dLockClient: redislock.New(client),
	}

	return rL, nil
}

// Makes sure, that RedisLock implements services.Lock
var _ services.Lock = RedisLock{}

func (r RedisLock) Lock() error {

	lock, err := r.dLockClient.Obtain(r.ctx, r.lockKey, 100*time.Millisecond, nil)

	if err != nil {
		return err
	}

	r.dLock = lock
	return nil
}

func (r RedisLock) Release() error {
	if r.dLock == nil {
		services.Log.Error("Unable to release lock, which was never locked.")
		return errors.New("no lock")
	}

	return r.dLock.Release(r.ctx)
}

type RedisSettings struct {
	MasterName string   `json:"master_name"`
	Addresses  []string `json:"addresses`
	Database   int64    `json:"database"`
	Password   string   `json:"password"`
}

type MetricHook struct {
	redisDurations *prometheus.HistogramVec
}

func (m MetricHook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, "time", time.Now()), nil

}

func (m MetricHook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	if startTime, ok := ctx.Value("time").(time.Time); ok {
		elapsedTime := time.Since(startTime)
		m.redisDurations.WithLabelValues(cmd.Name()).Observe(elapsedTime.Seconds())
	} else {
		services.Log.Warning("Context without time value found, something is broken within the metric instrumentation")
	}

	return nil
}

func (m MetricHook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	return ctx, nil
}

func (m MetricHook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	return nil
}

var RedisForm = forms.Form{
	ErrorMsg: "invalid data encountered in the Redis config form",
	Fields: []forms.Field{
		{
			Name: "addresses",
			Validators: []forms.Validator{
				forms.IsRequired{},
				forms.IsStringList{},
			},
		},
		{
			Name: "master_name",
			Validators: []forms.Validator{
				forms.IsOptional{Default: ""},
				forms.IsString{},
			},
		},
		{
			Name: "database",
			Validators: []forms.Validator{
				forms.IsOptional{Default: 0},
				forms.IsInteger{Min: 0, Max: 100},
			},
		},
		{
			Name: "password",
			Validators: []forms.Validator{
				forms.IsRequired{},
				forms.IsString{},
			},
		},
	},
}

func ValidateRedisSettings(settings map[string]interface{}) (interface{}, error) {
	if params, err := RedisForm.Validate(settings); err != nil {
		return nil, err
	} else {
		redisSettings := &RedisSettings{}
		if err := RedisForm.Coerce(redisSettings, params); err != nil {
			return nil, err
		}
		return redisSettings, nil
	}
}

func MakeRedis(settings interface{}) (services.Database, error) {

	redisSettings := settings.(RedisSettings)
	ctx := context.TODO()

	options := redis.UniversalOptions{
		MasterName:   redisSettings.MasterName,
		Password:     redisSettings.Password,
		ReadTimeout:  time.Second * 1.0,
		WriteTimeout: time.Second * 1.0,
		Addrs:        redisSettings.Addresses,
		DB:           int(redisSettings.Database),
	}

	redisDurations := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_durations_seconds",
			Help:    "Redis command durations",
			Buckets: []float64{0, 0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0},
		},
		[]string{"command"},
	)

	prometheus.MustRegister(redisDurations)

	client := redis.NewUniversalClient(&options)
	metricHook := MetricHook{redisDurations: redisDurations}
	client.AddHook(metricHook)

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	database := &Redis{
		options: options,
		client:  client,
		channel: make(chan bool),
		ctx:     ctx,
	}

	return database, nil

}

// Makes sure, that Redis implements Database
var _ services.Database = &Redis{}

func (d *Redis) Reset() error {
	return d.client.FlushDB(d.ctx).Err()
}

func (d *Redis) Lock(lockKey string) (services.Lock, error) {

	redisLock, err := MakeRedisLock(d.ctx, lockKey, d.client)

	if err != nil {
		return nil, err
	}

	err = redisLock.Lock()

	if err != nil {
		return nil, err
	}

	var sLock services.Lock
	sLock = redisLock

	return sLock, nil

}

func (d *Redis) Client() redis.Cmdable {
	return d.client
}

func (d *Redis) Open() error {
	return nil
}

func (d *Redis) Close() error {
	return d.client.Close()
}

func (d *Redis) Expire(table string, key []byte, ttl time.Duration) error {
	return d.Client().Expire(d.ctx, string(d.fullKey(table, key)), ttl).Err()
}

func (d *Redis) Set(table string, key []byte) services.Set {
	return &RedisSet{
		db:      d,
		fullKey: d.fullKey(table, key),
	}
}
func (d *Redis) SortedSet(table string, key []byte) services.SortedSet {
	return &RedisSortedSet{
		db:      d,
		fullKey: d.fullKey(table, key),
	}
}

func (d *Redis) List(table string, key []byte) services.List {
	return nil
}

func (d *Redis) Map(table string, key []byte) services.Map {
	return &RedisMap{
		db:      d,
		fullKey: d.fullKey(table, key),
	}
}

func (d *Redis) Value(table string, key []byte) services.Value {
	return &RedisValue{
		db:      d,
		fullKey: d.fullKey(table, key),
	}
}

func (d *Redis) fullKey(table string, key []byte) []byte {
	return []byte(fmt.Sprintf("%s::%s", table, string(key)))
}

type RedisMap struct {
	db      *Redis
	fullKey []byte
}

func (r *RedisMap) Del(key []byte) error {
	return r.db.Client().HDel(r.db.ctx, string(r.fullKey), string(key)).Err()
}

func (r *RedisMap) GetAll() (map[string][]byte, error) {
	result, err := r.db.Client().HGetAll(r.db.ctx, string(r.fullKey)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, NotFound
		}
		return nil, err
	}
	byteMap := map[string][]byte{}
	for k, v := range result {
		byteMap[k] = []byte(v)
	}
	return byteMap, nil
}

func (r *RedisMap) Get(key []byte) ([]byte, error) {
	result, err := r.db.Client().HGet(r.db.ctx, string(r.fullKey), string(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, NotFound
		}
		return nil, err
	}
	return []byte(result), nil
}

func (r *RedisMap) Set(key []byte, value []byte) error {
	return r.db.Client().HSet(r.db.ctx, string(r.fullKey), string(key), string(value)).Err()
}

type RedisSet struct {
	db      *Redis
	fullKey []byte
}

func (r *RedisSet) Add(data []byte) error {
	return r.db.Client().SAdd(r.db.ctx, string(r.fullKey), string(data)).Err()
}

func (r *RedisSet) Has(data []byte) (bool, error) {
	return r.db.Client().SIsMember(r.db.ctx, string(r.fullKey), string(data)).Result()
}

func (r *RedisSet) Del(data []byte) error {
	return r.db.Client().SRem(r.db.ctx, string(r.fullKey), string(data)).Err()
}

func (r *RedisSet) Members() ([]*services.SetEntry, error) {
	result, err := r.db.Client().SMembers(r.db.ctx, string(r.fullKey)).Result()
	if err != nil {
		return nil, err
	}

	var entries []*services.SetEntry

	for _, entry := range result {
		entries = append(entries, &services.SetEntry{
			Data: []byte(entry),
		})
	}
	return entries, nil
}

type RedisValue struct {
	db      *Redis
	fullKey []byte
}

func (r *RedisValue) Set(data []byte, ttl time.Duration) error {
	return r.db.Client().Set(r.db.ctx, string(r.fullKey), string(data), ttl).Err()
}

func (r *RedisValue) Get() ([]byte, error) {
	result, err := r.db.Client().Get(r.db.ctx, string(r.fullKey)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, NotFound
		}
		return nil, err
	}
	return []byte(result), nil
}

func (r *RedisValue) Del() error {
	return r.db.Client().Del(r.db.ctx, string(r.fullKey)).Err()
}

type RedisSortedSet struct {
	db      *Redis
	fullKey []byte
}

func (r *RedisSortedSet) Score(data []byte) (int64, error) {
	n, err := r.db.Client().ZScore(r.db.ctx, string(r.fullKey), string(data)).Result()
	if err == redis.Nil {
		return 0, NotFound
	} else if err != nil {
		return 0, err
	}
	return int64(n), nil
}

func (r *RedisSortedSet) Add(data []byte, score int64) error {
	return r.db.Client().ZAdd(r.db.ctx, string(r.fullKey), &redis.Z{Score: float64(score), Member: string(data)}).Err()
}

func (r *RedisSortedSet) Del(data []byte) (bool, error) {
	n, err := r.db.Client().ZRem(r.db.ctx, string(r.fullKey), string(data)).Result()
	return n > 0, err
}

func (r *RedisSortedSet) Range(from, to int64) ([]*services.SortedSetEntry, error) {
	result, err := r.db.Client().ZRangeWithScores(r.db.ctx, string(r.fullKey), from, to).Result()
	if err != nil {
		return nil, err
	}

	entries := []*services.SortedSetEntry{}

	for _, entry := range result {
		entries = append(entries, &services.SortedSetEntry{
			Score: int64(entry.Score),
			Data:  []byte(entry.Member.(string)),
		})
	}
	return entries, nil
}

func (r *RedisSortedSet) RangeByScore(from, to int64) ([]*services.SortedSetEntry, error) {
	result, err := r.db.Client().ZRangeByScoreWithScores(r.db.ctx, string(r.fullKey), &redis.ZRangeBy{
		Min: strconv.FormatInt(from, 10),
		Max: strconv.FormatInt(to, 10),
	}).Result()
	if err != nil {
		return nil, err
	}

	entries := []*services.SortedSetEntry{}

	for _, entry := range result {
		entries = append(entries, &services.SortedSetEntry{
			Score: int64(entry.Score),
			Data:  []byte(entry.Member.(string)),
		})
	}
	return entries, nil
}

func (r *RedisSortedSet) At(index int64) (*services.SortedSetEntry, error) {
	result, err := r.db.Client().ZRangeWithScores(r.db.ctx, string(r.fullKey), index, index).Result()
	if err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, NotFound
	}

	z := result[0]

	return &services.SortedSetEntry{
		Score: int64(z.Score),
		Data:  []byte(z.Member.(string)),
	}, nil
}

func (r *RedisSortedSet) PopMin(n int64) ([]*services.SortedSetEntry, error) {
	result, err := r.db.Client().ZPopMin(r.db.ctx, string(r.fullKey), n).Result()
	if err != nil {
		return nil, err
	}
	entries := []*services.SortedSetEntry{}
	for _, z := range result {
		entries = append(entries, &services.SortedSetEntry{
			Score: int64(z.Score),
			Data:  []byte(z.Member.(string)),
		})
	}
	return entries, nil
}

func (r *RedisSortedSet) RemoveRangeByScore(from, to int64) error {
	_, err := r.db.Client().ZRemRangeByScore(r.db.ctx, string(r.fullKey), strconv.FormatInt(from, 10), strconv.FormatInt(to, 10)).Result()
	if err != nil {
		return err
	}
	return nil
}
