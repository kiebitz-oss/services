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

package fixtures

import (
	"fmt"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/helpers"
	"time"
)

func TS(dt string) time.Time {
	if t, err := time.Parse(time.RFC3339, dt); err != nil {
		panic(err)
	} else {
		return t
	}
}

type Appointments struct {
	Start      time.Time
	Duration   int64
	N          int64
	Slots      int64
	Properties map[string]interface{}
}

func (c Appointments) Setup(fixtures map[string]interface{}) (interface{}, error) {

	client, ok := fixtures["client"].(*helpers.Client)

	if !ok {
		return nil, fmt.Errorf("client missing")
	}

	provider, ok := fixtures["provider"].(*helpers.Provider)

	if !ok {
		return nil, fmt.Errorf("provider missing")
	}

	appointments := make([]*services.SignedAppointment, c.N)

	ct := c.Start
	for i := int64(0); i < c.N; i++ {
		if appointment, err := services.MakeAppointment(ct, c.Slots, c.Duration); err != nil {
			return nil, err
		} else {
			appointment.PublicKey = provider.Actor.EncryptionKey.PublicKey
			appointment.Properties = c.Properties
			if signedAppointment, err := appointment.Sign(provider.Actor.SigningKey); err != nil {
				return nil, err
			} else {
				appointments[i] = signedAppointment
			}
		}
		ct = ct.Add(time.Duration(c.Duration) * time.Minute)
	}

	params := &services.PublishAppointmentsParams{
		Timestamp:    time.Now(),
		Appointments: appointments,
	}

	// we confirm the provider data
	if resp, err := client.Appointments.PublishAppointments(params, provider); err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		json, _ := resp.JSON()
		return nil, fmt.Errorf("cannot publish appointments: %v", json)
	}

	return appointments, nil

}

func (c Appointments) Teardown(fixture interface{}) error {
	return nil
}
