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
	"github.com/kiebitz-oss/services/databases"
)

func (c *Appointments) getAppointment(context services.Context, params *services.GetAppointmentParams) services.Response {

	// get all provider keys
	keys, err := c.getActorKeys()

	publicProviderData := c.backend.PublicProviderData()

	providerKey, err := findActorKey(keys.Providers, params.ProviderID)

	if err != nil {
		if err == databases.NotFound {
			return context.NotFound()
		}
		services.Log.Error(err)
		return context.InternalError()
	}

	// fetch the full public data of the provider
	providerData, err := publicProviderData.Get(params.ProviderID)

	if err != nil {
		if err == databases.NotFound {
			return context.NotFound()
		}
		services.Log.Error(err)
		return context.InternalError()
	}

	mediatorKey, err := findActorKey(keys.Mediators, providerKey.PublicKey)

	if err != nil {
		if err == databases.NotFound {
			return context.NotFound()
		}
		services.Log.Error(err)
		return context.InternalError()
	}

	keyChain := &services.KeyChain{
		Provider: providerKey,
		Mediator: mediatorKey,
	}

	providerData.ID = params.ProviderID

	appointmentDatesByID := c.backend.AppointmentDatesByID(params.ProviderID)

	if date, err := appointmentDatesByID.Get(params.ID); err != nil {
		services.Log.Errorf("Cannot get appointment by ID: %v", err)
		return context.InternalError()
	} else {

		appointmentsByDate := c.backend.AppointmentsByDate(params.ProviderID, date)

		if signedAppointment, err := appointmentsByDate.Get(params.ID); err != nil {
			if err == databases.NotFound {
				return context.NotFound()
			}
			services.Log.Errorf("Cannot get appointment by date: %v", err)
			return context.InternalError()
		} else {

			slots := make([]*services.Slot, len(signedAppointment.Bookings))

			for i, booking := range signedAppointment.Bookings {
				slots[i] = &services.Slot{ID: booking.ID}
			}

			// we remove the bookings as the user is not allowed to see them
			signedAppointment.Bookings = nil
			signedAppointment.BookedSlots = slots

			return context.Result(&services.ProviderAppointments{
				Provider:     providerData,
				Appointments: []*services.SignedAppointment{signedAppointment},
				KeyChain:     keyChain,
			})

		}

	}

}
