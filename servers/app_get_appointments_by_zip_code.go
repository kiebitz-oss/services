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
	"github.com/kiebitz-oss/services/databases"
	"github.com/kiebitz-oss/services/forms"
	"github.com/kiebitz-oss/services/jsonrpc"
)

func (c *Appointments) getAppointmentsByZipCode(context *jsonrpc.Context, params *services.GetAppointmentsByZipCodeParams) *jsonrpc.Response {

	keys, err := c.getActorKeys()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	neighbors := c.db.SortedSet("distances::neighbors::zipCode", []byte(params.ZipCode))
	publicProviderData := c.db.Map("providerData", []byte("public"))

	allNeighbors, err := neighbors.Range(0, -1)
	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	distances := map[string]int64{}

	for _, neighbor := range allNeighbors {
		distances[string(neighbor.Data)] = neighbor.Score
	}

	providerAppointmentsList := []*services.ProviderAppointments{}

	for _, providerKey := range keys.Providers {

		pkd, err := providerKey.ProviderKeyData()
		if err != nil {
			services.Log.Error(err)
			continue
		}

		if pkd.QueueData.ZipCode != params.ZipCode {
			if distance, ok := distances[pkd.QueueData.ZipCode]; !ok {
				continue
			} else if distance > params.Radius {
				continue
			}
		}

		// the provider "ID" is the hash of the signing key
		hash := crypto.Hash(pkd.Signing)

		pd, err := publicProviderData.Get(hash)

		if err != nil {
			if err != databases.NotFound {
				services.Log.Error(err)
			}
			services.Log.Info("provider data not found")
			continue
		}

		providerData := &services.SignedProviderData{}
		var providerDataMap map[string]interface{}

		if err := json.Unmarshal(pd, &providerDataMap); err != nil {
			services.Log.Error(err)
			continue
		}

		if params, err := forms.SignedProviderDataForm.Validate(providerDataMap); err != nil {
			services.Log.Error(err)
			continue
		} else if err := forms.SignedProviderDataForm.Coerce(providerData, params); err != nil {
			services.Log.Error(err)
			continue
		}

		// appointments are stored in a provider-specific key
		appointmentsByID := c.db.Map("appointmentsByID", hash)
		// complexity: O(n) where n is the number of appointments of the provider
		allDates, err := appointmentsByID.GetAll()

		if err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		signedAppointments := make([]*services.SignedAppointment, 0)

		visitedDates := make(map[string]bool)

		for _, date := range allDates {

			if _, ok := visitedDates[string(date)]; ok {
				continue
			} else {
				visitedDates[string(date)] = true
			}

			dateKey := append(hash, date...)
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

				slots := make([]*services.Slot, len(signedAppointment.Bookings))

				for i, booking := range signedAppointment.Bookings {
					slots[i] = &services.Slot{ID: booking.ID}
				}

				// we remove the bookings as the user is not allowed to see them
				signedAppointment.Bookings = nil
				signedAppointment.BookedSlots = slots

				signedAppointments = append(signedAppointments, signedAppointment)
			}
		}

		if len(signedAppointments) == 0 {
			continue
		}

		mediatorKey, err := findActorKey(keys.Mediators, providerKey.PublicKey)

		if err != nil {
			services.Log.Error(err)
			continue
		}

		keyChain := &services.KeyChain{
			Provider: providerKey,
			Mediator: mediatorKey,
		}

		providerAppointments := &services.ProviderAppointments{
			Provider: providerData,
			Offers:   signedAppointments,
			KeyChain: keyChain,
		}

		providerAppointmentsList = append(providerAppointmentsList, providerAppointments)

	}

	return context.Result(providerAppointmentsList)
}
