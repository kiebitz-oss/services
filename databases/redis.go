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
	"hash/fnv"
	"strconv"
	"sync"
	"time"
)

type Redis struct {
	metricsPrefix  string
	redisDurations *prometheus.HistogramVec
	clients        []redis.UniversalClient
	pipeline       redis.Pipeliner
	mutex          sync.Mutex
	channel        chan bool
	Ctx            context.Context
}

type RedisLock struct {
	lockKey     string
	client      redis.UniversalClient
	ctx         context.Context
	dLockClient *redislock.Client
	dLock       *redislock.Lock
}

func MakeRedisLock(ctx context.Context, lockKey string, client redis.UniversalClient) *RedisLock {
	return &RedisLock{
		client:      client,
		lockKey:     lockKey,
		ctx:         ctx,
		dLockClient: redislock.New(client),
	}
}

func (r *RedisLock) Lock() error {

	lock, err := r.dLockClient.Obtain(r.ctx, r.lockKey, 100*time.Millisecond, nil)

	if err != nil {
		return err
	}

	r.dLock = lock
	return nil
}

func (r *RedisLock) Release() error {
	if r.dLock == nil {
		services.Log.Errorf("Unable to release lock %s, which was never locked.", r.lockKey)
		return errors.New("no lock")
	}

	return r.dLock.Release(r.ctx)
}

type RedisShardSettings struct {
	Shards []RedisSettings `json:"shards"`
}

type RedisSettings struct {
	MasterName        string   `json:"master_name"`
	Addresses         []string `json:"addresses`
	SentinelAddresses []string `json:"sentinel_addresses`
	Database          int64    `json:"database"`
	Password          string   `json:"password"`
	SentinelUsername  string   `json:"sentinel_username"`
	SentinelPassword  string   `json:"sentinel_password"`
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
				forms.IsOptional{Default: []string{}},
				forms.IsStringList{},
			},
		},
		{
			Name: "sentinel_addresses",
			Validators: []forms.Validator{
				forms.IsOptional{Default: []string{}},
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
		{
			Name: "sentinel_username",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsString{},
			},
		},
		{
			Name: "sentinel_password",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsString{},
			},
		},
	},
}
var RedisShardForm = forms.Form{
	ErrorMsg: "invalid data encountered in the Redis config form",
	Fields: []forms.Field{
		{
			Name: "shards",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{Form: &RedisForm},
					},
				},
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

func ValidateRedisShardSettings(settings map[string]interface{}) (interface{}, error) {
	if params, err := RedisShardForm.Validate(settings); err != nil {
		return nil, err
	} else {
		redisSettings := &RedisShardSettings{}
		if err := RedisForm.Coerce(redisSettings, params); err != nil {
			return nil, err
		}
		return redisSettings, nil
	}
}

func MakeRedisClient(settings interface{}) (redis.UniversalClient, error) {
	redisSettings := settings.(RedisSettings)

	var client redis.UniversalClient

	if len(redisSettings.Addresses) > 0 {
		options := redis.UniversalOptions{
			MasterName:   redisSettings.MasterName,
			Password:     redisSettings.Password,
			ReadTimeout:  time.Second * 1.0,
			WriteTimeout: time.Second * 1.0,
			Addrs:        redisSettings.Addresses,
			DB:           int(redisSettings.Database),
		}

		client = redis.NewUniversalClient(&options)

		services.Log.Info("Creating redis connection")

	} else if len(redisSettings.SentinelAddresses) > 0 {
		options := redis.FailoverOptions{
			MasterName:       redisSettings.MasterName,
			Password:         redisSettings.Password,
			ReadTimeout:      time.Second * 1.0,
			WriteTimeout:     time.Second * 1.0,
			DB:               int(redisSettings.Database),
			SentinelAddrs:    redisSettings.SentinelAddresses,
			SentinelUsername: redisSettings.SentinelUsername,
			SentinelPassword: redisSettings.SentinelPassword,
		}

		client = redis.NewFailoverClient(&options)

		services.Log.Info("Creating sentinel-based redis connection")
	} else {
		return nil, errors.New("invalid database configuration, needed addresses or sentinelAddresses")
	}

	return client, nil
}

func MakeRedisShards(settings interface{}) (*Redis, error) {
	redisShardSettings := settings.(RedisShardSettings)
	ctx := context.TODO()

	clients := []redis.UniversalClient{}
	for _, redisSettings := range redisShardSettings.Shards {
		client, err := MakeRedisClient(redisSettings)

		if err != nil {
			return nil, err
		}

		if _, err := client.Ping(ctx).Result(); err != nil {
			return nil, err
		}

		clients = append(clients, client)
	}

	services.Log.Info("Creating redis-shard database")

	database := &Redis{
		clients: clients,
		channel: make(chan bool),
		Ctx:     ctx,
	}

	return database, nil

}

func MakeRedis(settings interface{}) (*Redis, error) {
	ctx := context.TODO()

	client, err := MakeRedisClient(settings)

	if err != nil {
		return nil, err
	}

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	services.Log.Info("Creating redis database")

	database := &Redis{
		clients: []redis.UniversalClient{client},
		channel: make(chan bool),
		Ctx:     ctx,
	}

	return database, nil
}

func MakeRedisAsDatabase(settings interface{}) (services.Database, error) {
	var db services.Database
	var err error

	db, err = MakeRedis(settings)

	return db, err
}

func MakeRedisShardAsDatabase(settings interface{}) (services.Database, error) {
	var db services.Database
	var err error

	db, err = MakeRedisShards(settings)

	return db, err
}

// Makes sure, that Redis implements Database
var _ services.Database = &Redis{}

func (d *Redis) Reset() error {
	for _, c := range d.clients {
		err := c.FlushDB(d.Ctx).Err()
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Redis) Lock(lockKey string) (services.Lock, error) {
	c := d.Client(lockKey)
	redisLock := MakeRedisLock(d.Ctx, lockKey, c)

	if err := redisLock.Lock(); err != nil {
		return nil, err
	}

	return redisLock, nil

}

func (d *Redis) getShardForKey(key string) uint32 {
	f := fnv.New32()
	f.Write([]byte(key))
	hashNum := f.Sum32()
	return hashNum % uint32(len(d.clients))
}

func (d *Redis) Client(key string) redis.UniversalClient {
	return d.clients[0]
	return d.clients[d.getShardForKey(key)]
}

func (d *Redis) Open() error {
	d.redisDurations = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_durations_seconds",
			Help:    "Redis command durations",
			Buckets: []float64{0, 0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1.0, 2.0},
		},
		[]string{"command"},
	)

	if err := prometheus.Register(d.redisDurations); err != nil {
		return err
	}

	metricHook := MetricHook{redisDurations: d.redisDurations}

	for _, client := range d.clients {
		client.AddHook(metricHook)
	}

	return nil
}

func (d *Redis) Close() error {
	prometheus.Unregister(d.redisDurations)
	for _, client := range d.clients {
		err := client.Close()

		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Redis) Expire(table string, key []byte, ttl time.Duration) error {
	stringKey := string(d.fullKey(table, key))
	return d.Client(stringKey).Expire(d.Ctx, stringKey, ttl).Err()
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

func (d *Redis) fullKey(table string, key []byte) string {
	return fmt.Sprintf("%s::%s", table, string(key))
}

type RedisMap struct {
	db      *Redis
	fullKey string
}

func (r *RedisMap) Del(key []byte) error {
	return r.db.Client(r.fullKey).HDel(r.db.Ctx, r.fullKey, string(key)).Err()
}

func (r *RedisMap) GetAll() (map[string][]byte, error) {
	result, err := r.db.Client(r.fullKey).HGetAll(r.db.Ctx, r.fullKey).Result()
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
	result, err := r.db.Client(r.fullKey).HGet(r.db.Ctx, r.fullKey, string(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, NotFound
		}
		return nil, err
	}
	return []byte(result), nil
}

func (r *RedisMap) Set(key []byte, value []byte) error {
	return r.db.Client(r.fullKey).HSet(r.db.Ctx, r.fullKey, string(key), string(value)).Err()
}

type RedisSet struct {
	db      *Redis
	fullKey string
}

func (r *RedisSet) Add(data []byte) error {
	return r.db.Client(r.fullKey).SAdd(r.db.Ctx, r.fullKey, string(data)).Err()
}

func (r *RedisSet) Has(data []byte) (bool, error) {
	return r.db.Client(r.fullKey).SIsMember(r.db.Ctx, r.fullKey, string(data)).Result()
}

func (r *RedisSet) Del(data []byte) error {
	return r.db.Client(r.fullKey).SRem(r.db.Ctx, r.fullKey, string(data)).Err()
}

func (r *RedisSet) Members() ([]*services.SetEntry, error) {
	result, err := r.db.Client(r.fullKey).SMembers(r.db.Ctx, r.fullKey).Result()
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
	fullKey string
}

func (r *RedisValue) Set(data []byte, ttl time.Duration) error {
	return r.db.Client(r.fullKey).Set(r.db.Ctx, string(r.fullKey), string(data), ttl).Err()
}

func (r *RedisValue) Get() ([]byte, error) {
	result, err := r.db.Client(r.fullKey).Get(r.db.Ctx, string(r.fullKey)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, NotFound
		}
		return nil, err
	}
	return []byte(result), nil
}

func (r *RedisValue) Del() error {
	return r.db.Client(r.fullKey).Del(r.db.Ctx, string(r.fullKey)).Err()
}

type RedisSortedSet struct {
	db      *Redis
	fullKey string
}

func (r *RedisSortedSet) Score(data []byte) (int64, error) {
	n, err := r.db.Client(r.fullKey).ZScore(r.db.Ctx, r.fullKey, string(data)).Result()
	if err == redis.Nil {
		return 0, NotFound
	} else if err != nil {
		return 0, err
	}
	return int64(n), nil
}

func (r *RedisSortedSet) Add(data []byte, score int64) error {
	return r.db.Client(r.fullKey).ZAdd(r.db.Ctx, r.fullKey, &redis.Z{Score: float64(score), Member: string(data)}).Err()
}

func (r *RedisSortedSet) Del(data []byte) (bool, error) {
	n, err := r.db.Client(r.fullKey).ZRem(r.db.Ctx, r.fullKey, string(data)).Result()
	return n > 0, err
}

func (r *RedisSortedSet) Range(from, to int64) ([]*services.SortedSetEntry, error) {
	result, err := r.db.Client(r.fullKey).ZRangeWithScores(r.db.Ctx, r.fullKey, from, to).Result()
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
	result, err := r.db.Client(r.fullKey).ZRangeByScoreWithScores(r.db.Ctx, r.fullKey, &redis.ZRangeBy{
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
	result, err := r.db.Client(r.fullKey).ZRangeWithScores(r.db.Ctx, r.fullKey, index, index).Result()
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
	result, err := r.db.Client(r.fullKey).ZPopMin(r.db.Ctx, r.fullKey, n).Result()
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
	_, err := r.db.Client(r.fullKey).ZRemRangeByScore(r.db.Ctx, r.fullKey, strconv.FormatInt(from, 10), strconv.FormatInt(to, 10)).Result()
	if err != nil {
		return err
	}
	return nil
}
