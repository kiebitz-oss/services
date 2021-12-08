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

package servers

import (
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"time"
)

func (c *Appointments) getProviderAppointments(context services.Context, params *services.GetProviderAppointmentsSignedParams) services.Response {

	resp, providerKey := c.isProvider(context, &services.SignedParams{
		JSON:      params.JSON,
		Signature: params.Signature,
		PublicKey: params.PublicKey,
		Timestamp: params.Data.Timestamp,
	})

	if resp != nil {
		return resp
	}

	pkd, err := providerKey.ProviderKeyData()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	// the provider "ID" is the hash of the signing key
	hash := crypto.Hash(pkd.Signing)

	// appointments are stored in a provider-specific key
	appointmentsByID := c.db.Map("appointmentsByID", hash)
	allDates, err := appointmentsByID.GetAll()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	signedAppointments := make([]*services.SignedAppointment, 0)

	for _, dateStr := range allDates {

		date, err := time.Parse("2006-01-02", string(dateStr))

		if err != nil {
			services.Log.Error(err)
			continue
		}

		if params.Data.From != nil && date.Before(*params.Data.From) {
			continue
		}

		if params.Data.To != nil && date.After(*params.Data.To) {
			continue
		}

		dateKey := append(hash, dateStr...)
		appointmentsByDate := c.db.Map("appointmentsByDate", dateKey)
		allAppointments, err := appointmentsByDate.GetAll()
		if err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		for _, appointment := range allAppointments {
			var signedAppointment *services.SignedAppointment
			if err := json.Unmarshal(appointment, &signedAppointment); err != nil {
				services.Log.Error(err)
				continue
			}
			signedAppointments = append(signedAppointments, signedAppointment)
		}
	}

	return context.Result(signedAppointments)
}
