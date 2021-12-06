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

package services

import (
	"time"
)

type DatabaseDefinition struct {
	Name              string            `json:"name"`
	Description       string            `json:"description"`
	Maker             DatabaseMaker     `json:"-"`
	SettingsValidator SettingsValidator `json:"-"`
}

type SettingsValidator func(settings map[string]interface{}) (interface{}, error)
type DatabaseDefinitions map[string]DatabaseDefinition
type DatabaseMaker func(settings interface{}) (Database, error)

type DatabaseOps interface {
	Expire(table string, key []byte, ttl time.Duration) error
	Set(table string, key []byte) Set
	SortedSet(table string, key []byte) SortedSet
	List(table string, key []byte) List
	Map(table string, key []byte) Map
	Value(table string, key []byte) Value
}

type Lock interface {
	Release() error
}

// A database can deliver and accept message
type Database interface {
	Close() error
	Open() error
	Reset() error
	Lock(lockKey string) (Lock, error)

	DatabaseOps
}

type Object interface {
}

type Set interface {
	Add([]byte) error
	Has([]byte) (bool, error)
	Del(key []byte) error
	Members() ([]*SetEntry, error)
	Object
}

type SetEntry struct {
	Data []byte
}

type SortedSetEntry struct {
	Score int64
	Data  []byte
}

type SortedSet interface {
	Object
	Del([]byte) (bool, error)
	Add([]byte, int64) error
	Range(int64, int64) ([]*SortedSetEntry, error)
	RangeByScore(int64, int64) ([]*SortedSetEntry, error)
	At(int64) (*SortedSetEntry, error)
	Score([]byte) (int64, error)
	PopMin(int64) ([]*SortedSetEntry, error)
	RemoveRangeByScore(int64, int64) error
}

type List interface {
	Object
}

type Map interface {
	GetAll() (map[string][]byte, error)
	Get(key []byte) ([]byte, error)
	Del(key []byte) error
	Set(key []byte, value []byte) error
	Remove() error
	Object
}

type Value interface {
	Object
	Set(value []byte, ttl time.Duration) error
	Get() ([]byte, error)
	Del() error
}

type BaseDatabase struct {
}
