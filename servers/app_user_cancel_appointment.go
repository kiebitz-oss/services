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
	"bytes"
	"github.com/kiebitz-oss/services"
	"time"
)

func (c *Appointments) cancelAppointment(context services.Context, params *services.CancelAppointmentSignedParams) services.Response {

	if resp := c.isUser(context, &services.SignedParams{
		JSON:      params.JSON,
		Signature: params.Signature,
		PublicKey: params.PublicKey,
		ExtraData: params.Data.SignedTokenData,
		Timestamp: params.Data.Timestamp,
	}); resp != nil {
		return resp
	}

	appointmentDatesByID := c.backend.AppointmentDatesByID(params.Data.ProviderID)

	if date, err := appointmentDatesByID.Get(params.Data.ID); err != nil {
		services.Log.Errorf("Cannot get appointment by ID: %v", err)
		return context.InternalError()
	} else {

		appointmentsByDate := c.backend.AppointmentsByDate(params.Data.ProviderID, date)

		if signedAppointment, err := appointmentsByDate.Get(params.Data.ID); err != nil {
			services.Log.Errorf("Cannot get appointment by date: %v", err)
			return context.InternalError()
		} else {
			newBookings := make([]*services.Booking, 0)

			token := params.Data.SignedTokenData.Data.Token

			found := false
			for _, booking := range signedAppointment.Bookings {
				if bytes.Equal(booking.Token, token) {
					found = true
					continue
				}
				newBookings = append(newBookings, booking)
			}

			if !found {
				return context.NotFound()
			}

			signedAppointment.Bookings = newBookings

			usedTokens := c.backend.UsedTokens()

			// we mark the token as unused
			if err := usedTokens.Del(token); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

			signedAppointment.UpdatedAt = time.Now()

			// we update the appointment
			if err := appointmentsByDate.Set(signedAppointment); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

		}

	}

	return context.Acknowledge()

}
