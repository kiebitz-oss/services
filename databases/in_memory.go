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
	"crypto/sha256"
	"github.com/kiebitz-oss/services"
	"time"
)

type TTL struct {
	Key string
	TTL time.Time
}

type InMemory struct {
	Data map[string][][]byte
	TTLs []*TTL
}

// we hash all keys to SHA256 values to limit the storage size
func bhash(key []byte) []byte {
	v := sha256.Sum256(key)
	return v[:]
}

func hash(key []byte) string {
	return string(bhash(key))
}

func fk(table string, key []byte) []byte {
	return append([]byte(table), key...)
}

type InMemorySettings struct {
}

func ValidateInMemorySettings(settings map[string]interface{}) (interface{}, error) {
	return settings, nil
}

func MakeInMemory(settings interface{}) (services.Database, error) {
	return &InMemory{
		Data: make(map[string][][]byte),
		TTLs: make([]*TTL, 0, 100),
	}, nil
}

var _ services.Database = &InMemory{}

func (d *InMemory) Reset() error {
	return nil
}

func (d *InMemory) Close() error {
	return nil
}

func (d *InMemory) Open() error {
	return nil
}

func (d *InMemory) Lock(lockKey string) (services.Lock, error) {
	return nil, nil
}

func (d *InMemory) Expire(table string, key []byte, ttl time.Duration) error {
	return nil
}

func (d *InMemory) Set(table string, key []byte) services.Set {
	return nil
}
func (d *InMemory) SortedSet(table string, key []byte) services.SortedSet {
	return nil
}

func (d *InMemory) List(table string, key []byte) services.List {
	return nil
}

func (d *InMemory) Map(table string, key []byte) services.Map {
	return nil
}

func (d *InMemory) Value(table string, key []byte) services.Value {
	return nil
}

func (d *InMemory) Integer(table string, key []byte) services.Integer {
	return nil
}
