// Kiebitz - Privacy-Friendly Appointment Scheduling
// Copyright (C) 2021-2021 The Kiebitz Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version. Additional terms
// as defined in section 7 of the license (e.g. regarding attribution)
// are specified at https://kiebitz.eu/en/docs/open-source/additional-terms.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package servers

import (
	"encoding/json"
	"fmt"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/forms"
	"time"
)

// The appointments backend acts as an interface between the API and the
// database. It is mostly concerned with ensuring data is propery serialized
// and deserialized when stored or fetched from the database.
type AppointmentsBackend struct {
	db services.Database
}

func (a *AppointmentsBackend) Neighbors(neighborType, zipCode string) *Neighbors {
	return &Neighbors{
		neighbors: a.db.SortedSet(fmt.Sprintf("distances::neighbors::%s", neighborType), []byte(zipCode)),
	}
}

func (a *AppointmentsBackend) PriorityToken(name string) *PriorityToken {
	return &PriorityToken{
		token: a.db.Integer("priorityToken", []byte(name)),
	}
}

func (a *AppointmentsBackend) Keys(actor string) *Keys {
	return &Keys{
		keys: a.db.Map("keys", []byte(actor)),
	}
}

func (a *AppointmentsBackend) Codes(actor string) *Codes {
	return &Codes{
		codes:  a.db.Set("codes", []byte(actor)),
		scores: a.db.SortedSet("codeScores", []byte(actor)),
	}
}

func (a *AppointmentsBackend) PublicProviderData() *PublicProviderData {
	return &PublicProviderData{
		dbs: a.db.Map("providerData", []byte("public")),
	}
}

func (a *AppointmentsBackend) ConfirmedProviderData() *ConfirmedProviderData {
	return &ConfirmedProviderData{
		dbs: a.db.Map("providerData", []byte("confirmed")),
	}
}

func (a *AppointmentsBackend) UnverifiedProviderData() *RawProviderData {
	return &RawProviderData{
		dbs: a.db.Map("providerData", []byte("unverified")),
	}
}

func (a *AppointmentsBackend) VerifiedProviderData() *RawProviderData {
	return &RawProviderData{
		dbs: a.db.Map("providerData", []byte("verified")),
	}
}

func (a *AppointmentsBackend) AppointmentsByDate(providerID []byte, date string) *AppointmentsByDate {
	dateKey := append(providerID, []byte(date)...)
	return &AppointmentsByDate{
		dbs: a.db.Map("appointmentsByDate", dateKey),
	}
}

func (a *AppointmentsBackend) AppointmentDatesByID(providerID []byte) *AppointmentDatesByID {
	return &AppointmentDatesByID{
		providerID: providerID,
		db:         a.db,
		dbs:        a.db.Map("appointmentDatesByID", providerID),
	}
}

func (a *AppointmentsBackend) UsedTokens() *UsedTokens {
	return &UsedTokens{
		dbs: a.db.Set("bookings", []byte("tokens")),
	}
}

type PriorityToken struct {
	token services.Integer
}

func (p *PriorityToken) IncrBy(value int64) (int64, error) {
	return p.token.IncrBy(value)
}

type Neighbors struct {
	neighbors services.SortedSet
}

func (n *Neighbors) Add(to string, distance int64) error {
	return n.neighbors.Add([]byte(to), distance)
}

func (n *Neighbors) Range(from, to int64) ([]*services.SortedSetEntry, error) {
	return n.neighbors.Range(from, to)
}

type Keys struct {
	keys services.Map
}

func (k *Keys) Set(id []byte, key *services.ActorKey) error {
	if data, err := json.Marshal(key); err != nil {
		return err
	} else {
		return k.keys.Set(id, data)
	}
}

func (k *Keys) Get(id []byte) (*services.ActorKey, error) {
	if mk, err := k.keys.Get(id); err != nil {
		return nil, err
	} else {
		var key *services.ActorKey
		if err := json.Unmarshal(mk, &key); err != nil {
			return nil, err
		} else {
			// we set the ID to
			key.ID = id
			return key, nil
		}
	}
}

func (k *Keys) GetAll() ([]*services.ActorKey, error) {

	mk, err := k.keys.GetAll()

	if err != nil {
		return nil, err
	}

	actorKeys := []*services.ActorKey{}

	for id, v := range mk {
		var m *services.ActorKey
		if err := json.Unmarshal(v, &m); err != nil {
			return nil, err
		} else {
			m.ID = []byte(id)
			actorKeys = append(actorKeys, m)
		}
	}

	return actorKeys, nil

}

type Codes struct {
	codes  services.Set
	scores services.SortedSet
}

func (c *Codes) Has(code []byte) (bool, error) {
	return c.codes.Has(code)
}

func (c *Codes) Add(code []byte) error {
	return c.codes.Add(code)
}

func (c *Codes) Del(code []byte) error {
	return c.codes.Del(code)
}

func (c *Codes) Score(code []byte) (int64, error) {
	return c.scores.Score(code)
}

func (c *Codes) AddToScore(code []byte, score int64) error {
	return c.scores.Add(code, score)
}

type ConfirmedProviderData struct {
	dbs services.Map
}

func (c *ConfirmedProviderData) Set(providerID []byte, encryptedData *services.ConfirmedProviderData) error {
	if data, err := json.Marshal(encryptedData); err != nil {
		return err
	} else {
		return c.dbs.Set(providerID, data)
	}
}

func (c *ConfirmedProviderData) Get(providerID []byte) (*services.ConfirmedProviderData, error) {
	if data, err := c.dbs.Get(providerID); err != nil {
		return nil, err
	} else {
		var mapData map[string]interface{}
		encryptedData := &services.ConfirmedProviderData{}
		if err := json.Unmarshal(data, &mapData); err != nil {
			return nil, err
		} else if params, err := forms.ConfirmedProviderDataForm.Validate(mapData); err != nil {
			return nil, err
		} else if err := forms.ConfirmedProviderDataForm.Coerce(encryptedData, params); err != nil {
			return nil, err
		} else {
			return encryptedData, nil
		}
	}
}

type RawProviderData struct {
	dbs services.Map
}

func (c *RawProviderData) Set(providerID []byte, rawData *services.RawProviderData) error {
	if data, err := json.Marshal(rawData); err != nil {
		return err
	} else {
		return c.dbs.Set(providerID, data)
	}
}

func (c *RawProviderData) Del(providerID []byte) error {
	return c.dbs.Del(providerID)
}

func (c *RawProviderData) Get(providerID []byte) (*services.RawProviderData, error) {
	if data, err := c.dbs.Get(providerID); err != nil {
		return nil, err
	} else {
		var mapData map[string]interface{}
		rawData := &services.RawProviderData{}
		if err := json.Unmarshal(data, &mapData); err != nil {
			return nil, err
		} else if params, err := forms.RawProviderDataForm.Validate(mapData); err != nil {
			return nil, err
		} else if err := forms.RawProviderDataForm.Coerce(rawData, params); err != nil {
			return nil, err
		} else {
			return rawData, nil
		}
	}
}

func (c *RawProviderData) GetAll() (map[string]*services.RawProviderData, error) {
	if dataMap, err := c.dbs.GetAll(); err != nil {
		return nil, err
	} else {
		providerDataMap := map[string]*services.RawProviderData{}
		for id, data := range dataMap {
			var mapData map[string]interface{}
			rawData := &services.RawProviderData{}
			if err := json.Unmarshal(data, &mapData); err != nil {
				return nil, err
			} else if params, err := forms.RawProviderDataForm.Validate(mapData); err != nil {
				return nil, err
			} else if err := forms.RawProviderDataForm.Coerce(rawData, params); err != nil {
				return nil, err
			} else {
				providerDataMap[id] = rawData
			}
		}
		return providerDataMap, nil
	}
}

type UsedTokens struct {
	dbs services.Set
}

func (t *UsedTokens) Del(token []byte) error {
	return t.dbs.Del(token)
}

func (t *UsedTokens) Has(token []byte) (bool, error) {
	return t.dbs.Has(token)
}

func (t *UsedTokens) Add(token []byte) error {
	return t.dbs.Add(token)
}

type AppointmentDatesByID struct {
	providerID []byte
	dbs        services.Map
	db         services.Database
}

func (a *AppointmentDatesByID) GetAll() (map[string][]byte, error) {
	return a.dbs.GetAll()
}

func (a *AppointmentDatesByID) Get(id []byte) (string, error) {
	if data, err := a.dbs.Get(id); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func (a *AppointmentDatesByID) Set(id []byte, date string) error {
	// ID map will auto-delete after one year (purely for storage reasons, it does not contain sensitive data)
	if err := a.db.Expire("appointmentDatesByID", a.providerID, time.Hour*24*365); err != nil {
		return err
	}
	return a.dbs.Set(id, []byte(date))
}

func (a *AppointmentDatesByID) Del(id []byte) error {
	return a.dbs.Del(id)
}

type PublicProviderData struct {
	dbs services.Map
}

func (p *PublicProviderData) Get(id []byte) (*services.SignedProviderData, error) {
	if data, err := p.dbs.Get(id); err != nil {
		return nil, err
	} else if signedProviderData, err := SignedProviderData(data); err != nil {
		return nil, err
	} else {
		return signedProviderData, nil
	}
}

func (p *PublicProviderData) Set(id []byte, signedProviderData *services.SignedProviderData) error {
	if data, err := json.Marshal(signedProviderData); err != nil {
		return err
	} else {
		return p.dbs.Set(id, data)
	}
}

type AppointmentsByDate struct {
	dbs services.Map
}

func (a *AppointmentsByDate) Del(id []byte) error {
	return a.dbs.Del(id)
}

func (a *AppointmentsByDate) Set(appointment *services.SignedAppointment) error {
	if data, err := json.Marshal(appointment); err != nil {
		return err
	} else {
		return a.dbs.Set(appointment.Data.ID, data)
	}
}

func (a *AppointmentsByDate) Get(id []byte) (*services.SignedAppointment, error) {
	if appointmentData, err := a.dbs.Get(id); err != nil {
		return nil, err
	} else {
		if signedAppointment, err := SignedAppointment(appointmentData); err != nil {
			return nil, err
		} else {
			return signedAppointment, nil
		}
	}
}

func (a *AppointmentsByDate) GetAll() (map[string]*services.SignedAppointment, error) {

	signedAppointments := make(map[string]*services.SignedAppointment)

	if allAppointments, err := a.dbs.GetAll(); err != nil {
		return nil, err
	} else {
		for id, appointmentData := range allAppointments {
			if signedAppointment, err := SignedAppointment(appointmentData); err != nil {
				return nil, err
			} else {
				signedAppointments[id] = signedAppointment
			}
		}

		return signedAppointments, nil
	}
}
