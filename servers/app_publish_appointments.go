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
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/forms"
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

	lock, err := c.db.Lock("bookAppointment_" + string(hash[:]))
	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	defer lock.Release()

	// appointments are stored in a provider-specific key
	appointmentDatesByID := c.db.Map("appointmentDatesByID", hash)

	// appointments expire automatically after 120 days
	if err := c.db.Expire("appointments", hash, time.Hour*24*120); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	usedTokens := c.db.Set("bookings", []byte("tokens"))

	var bookedSlots, openSlots int64

	for _, appointment := range params.Data.Offers {

		// check if there's an existing appointment
		if date, err := appointmentDatesByID.Get(appointment.Data.ID); err == nil {

			if err := appointmentDatesByID.Del(appointment.Data.ID); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

			appointmentsByDate := c.db.Map("appointmentsByDate", append(hash, date...))

			existingAppointment := &services.SignedAppointment{}
			var mapData map[string]interface{}

			if data, err := appointmentsByDate.Get(appointment.Data.ID); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := appointmentsByDate.Del(appointment.Data.ID); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := json.Unmarshal(data, &mapData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if params, err := forms.SignedAppointmentForm.Validate(mapData); err != nil {
				services.Log.Error(err)
			} else if err := forms.SignedAppointmentForm.Coerce(existingAppointment, params); err != nil {
				services.Log.Error(err)
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

		date := []byte(appointment.Data.Timestamp.Format("2006-01-02"))
		// the hash is under our control so it's safe to concatenate it directly with the date
		dateKey := append(hash, date...)

		appointmentsByDate := c.db.Map("appointmentsByDate", dateKey)

		// appointments will auto-delete one day after their timestamp
		if err := c.db.Expire("appointmentsByDate", dateKey, appointment.Data.Timestamp.Sub(time.Now())+time.Hour*24); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		// ID map will auto-delete after one year (purely for storage reasons, it does not contain sensitive data)
		if err := c.db.Expire("appointmentDatesByID", hash, time.Hour*24*365); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		if err := appointmentDatesByID.Set(appointment.Data.ID, date); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		}

		appointment.UpdatedAt = time.Now()

		if jsonData, err := json.Marshal(appointment); err != nil {
			services.Log.Error(err)
			return context.InternalError()
		} else if err := appointmentsByDate.Set(appointment.Data.ID, jsonData); err != nil {
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
