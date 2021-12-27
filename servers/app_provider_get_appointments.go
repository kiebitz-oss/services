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
	appointmentDatesByID := c.backend.AppointmentDatesByID(hash)
	allDates, err := appointmentDatesByID.GetAll()
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

		if date.Before(params.Data.From) || date.After(params.Data.To) {
			continue
		}

		appointmentsByDate := c.backend.AppointmentsByDate(hash, string(dateStr))

		allAppointments, err := appointmentsByDate.GetAll()

		if err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		for _, appointment := range allAppointments {
			// if the updatedSince parameter is given we only return appointments that have
			// been updated since the given time
			if params.Data.UpdatedSince != nil && (params.Data.UpdatedSince.After(appointment.UpdatedAt) || params.Data.UpdatedSince.Equal(appointment.UpdatedAt)) {
				continue
			}
			signedAppointments = append(signedAppointments, appointment)
		}
	}

	return context.Result(signedAppointments)
}
