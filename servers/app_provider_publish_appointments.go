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
	"bytes"
	"encoding/hex"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"time"
)

func (c *Appointments) publishAppointments(context services.Context, params *services.PublishAppointmentsSignedParams) services.Response {

	resp, providerKey := c.isProvider(context, &services.SignedParams{
		JSON:      params.JSON,
		Signature: params.Signature,
		PublicKey: params.PublicKey,
		Timestamp: params.Data.Timestamp,
	})

	if resp != nil {
		return resp
	}

	if expired(params.Data.Timestamp) {
		return context.Error(410, "signature expired", nil)
	}

	pkd, err := providerKey.ProviderKeyData()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	// the provider "ID" is the hash of the signing key
	hash := crypto.Hash(pkd.Signing)
	hexUID := hex.EncodeToString(hash)

	// appointments are stored in a provider-specific key
	appointmentDatesByID := c.backend.AppointmentDatesByID(hash)
	usedTokens := c.backend.UsedTokens()

	// to do: fix statistics generation
	var bookedSlots, openSlots int64

	for _, appointment := range params.Data.Offers {

		// check if there's an existing appointment
		if date, err := appointmentDatesByID.Get(appointment.Data.ID); err == nil {

			if err := appointmentDatesByID.Del(appointment.Data.ID); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

			appointmentsByDate := c.backend.AppointmentsByDate(hash, string(date))

			if existingAppointment, err := appointmentsByDate.Get(appointment.Data.ID); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := appointmentsByDate.Del(appointment.Data.ID); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else {
				bookings := make([]*services.Booking, 0)
				for _, existingSlotData := range existingAppointment.Data.SlotData {
					found := false
					for _, slotData := range appointment.Data.SlotData {
						if bytes.Equal(slotData.ID, existingSlotData.ID) {
							found = true
							break
						}
					}
					if found {
						// this slot has been preserved, if there's any booking for it we migrate it
						for _, booking := range existingAppointment.Bookings {
							if bytes.Equal(booking.ID, existingSlotData.ID) {
								bookings = append(bookings, booking)
								break
							}
						}
					} else {
						// this slot has been deleted, if there's any booking for it we delete it
						for _, booking := range existingAppointment.Bookings {
							if bytes.Equal(booking.ID, existingSlotData.ID) {
								// we re-enable the associated token
								if err := usedTokens.Del(booking.Token); err != nil {
									services.Log.Error(err)
									return context.InternalError()
								}
								break
							}
						}
					}
				}
				appointment.Bookings = bookings
			}
		}

		date := appointment.Data.Timestamp.Format("2006-01-02")

		appointmentsByDate := c.backend.AppointmentsByDate(hash, date)

		if err := appointmentDatesByID.Set(appointment.Data.ID, date); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		appointment.UpdatedAt = time.Now()

		if err := appointmentsByDate.Set(appointment); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}
	}

	if c.meter != nil {

		now := time.Now().UTC().UnixNano()

		addTokenStats := func(tw services.TimeWindow, data map[string]string) error {
			// we add the maximum of the open appointments
			if err := c.meter.AddMax("queues", "open", hexUID, data, tw, openSlots); err != nil {
				return err
			}
			// we add the maximum of the booked appointments
			if err := c.meter.AddMax("queues", "booked", hexUID, data, tw, bookedSlots); err != nil {
				return err
			}
			// we add the info that this provider is active
			if err := c.meter.AddOnce("queues", "active", hexUID, data, tw, 1); err != nil {
				return err
			}
			return nil
		}

		for _, twt := range tws {

			// generate the time window
			tw := twt(now)

			// global statistics
			if err := addTokenStats(tw, map[string]string{}); err != nil {
				services.Log.Error(err)
			}

			// statistics by zip code
			if err := addTokenStats(tw, map[string]string{
				"zipCode": pkd.QueueData.ZipCode,
			}); err != nil {
				services.Log.Error(err)
			}

		}

	}

	return context.Acknowledge()
}
