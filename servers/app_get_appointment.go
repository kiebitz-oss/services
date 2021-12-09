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
	"github.com/kiebitz-oss/services/forms"
)

func (c *Appointments) getAppointment(context services.Context, params *services.GetAppointmentSignedParams) services.Response {

	if resp := c.isUser(context, &services.SignedParams{
		JSON:      params.JSON,
		Signature: params.Signature,
		PublicKey: params.PublicKey,
		ExtraData: params.Data.SignedTokenData,
		Timestamp: params.Data.Timestamp,
	}); resp != nil {
		return resp
	}

	appointmentDatesByID := c.db.Map("appointmentDatesByID", params.Data.ProviderID)

	if date, err := appointmentDatesByID.Get(params.Data.ID); err != nil {
		services.Log.Errorf("Cannot get appointment by ID: %v", err)
		return context.InternalError()
	} else {

		dateKey := append(params.Data.ProviderID, date...)
		appointmentsByDate := c.db.Map("appointmentsByDate", dateKey)

		if appointment, err := appointmentsByDate.Get(params.Data.ID); err != nil {
			services.Log.Errorf("Cannot get appointment by date: %v", err)
			return context.InternalError()
		} else {
			signedAppointment := &services.SignedAppointment{}
			var mapData map[string]interface{}
			if err := json.Unmarshal(appointment, &mapData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if params, err := forms.SignedAppointmentForm.Validate(mapData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := forms.SignedAppointmentForm.Coerce(signedAppointment, params); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

			slots := make([]*services.Slot, len(signedAppointment.Bookings))

			for i, booking := range signedAppointment.Bookings {
				slots[i] = &services.Slot{ID: booking.ID}
			}

			// we remove the bookings as the user is not allowed to see them
			signedAppointment.Bookings = nil
			signedAppointment.BookedSlots = slots

			return context.Result(signedAppointment)

		}

	}

}
