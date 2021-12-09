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
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/forms"
	"time"
)

func (c *Appointments) isActiveProviderID(context services.Context, publicKey []byte) (services.Response, bool) {
	activeProvider, err := c.db.Map("keys", []byte("providers")).Get([]byte(publicKey))

	if len(activeProvider) == 0 {
		return context.Error(404, "provider not found", nil), false
	} else if err != nil {
		services.Log.Error(err)
		return context.InternalError(), false
	}

	return nil, true
}

func (c *Appointments) bookAppointment(context services.Context, params *services.BookAppointmentSignedParams) services.Response {

	// Not sure, if this lock makes any sense.
	lock, err := c.db.Lock("bookAppointment_" + string(params.Data.ID[:]))
	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	defer lock.Release()

	var result interface{}

	usedTokens := c.db.Set("bookings", []byte("tokens"))

	notAuthorized := context.Error(401, "not authorized", nil)

	signedData := &crypto.SignedStringData{
		Data:      params.Data.SignedTokenData.JSON,
		Signature: params.Data.SignedTokenData.Signature,
	}

	tokenKey := c.settings.Key("token")

	if ok, err := tokenKey.VerifyString(signedData); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if !ok {
		return context.Error(400, "invalid signature", nil)
	}

	token := params.Data.SignedTokenData.Data.Token

	if ok, err := usedTokens.Has(token); err != nil {
		services.Log.Error()
		return context.InternalError()
	} else if ok {
		return notAuthorized
	}

	// we verify the signature (without verifying e.g. the provenance of the key)
	if ok, err := crypto.VerifyWithBytes([]byte(params.JSON), params.Signature, params.PublicKey); err != nil {
		services.Log.Errorf("Cannot verify with bytes: %s", err)
		return context.InternalError()
	} else if !ok {
		return context.Error(400, "invalid signature", nil)
	}

	// test if provider of the appointment is still active
	if res, isActive := c.isActiveProviderID(context, params.Data.ProviderID); res != nil {
		return res
	} else if !isActive {
		return context.Error(404, "invalid provider id", nil)
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
			// we try to find an open slot

			foundSlot := false
			for _, slotData := range signedAppointment.Data.SlotData {

				found := false

				for _, booking := range signedAppointment.Bookings {
					if bytes.Equal(booking.ID, slotData.ID) {
						found = true
						break
					}
				}

				if found {
					continue
				}

				// this slot is open, we book it!

				booking := &services.Booking{
					PublicKey:     params.PublicKey,
					ID:            slotData.ID,
					Token:         token,
					EncryptedData: params.Data.EncryptedData,
				}

				signedAppointment.Bookings = append(signedAppointment.Bookings, booking)
				foundSlot = true

				result = booking

				break
			}

			if !foundSlot {
				return context.NotFound()
			}

			// we mark the token as used
			if err := usedTokens.Add(token); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

			signedAppointment.UpdatedAt = time.Now()

			// we update the appointment
			if jsonData, err := json.Marshal(signedAppointment); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := appointmentsByDate.Set(signedAppointment.Data.ID, jsonData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

		}

	}

	if c.meter != nil {

		now := time.Now().UTC().UnixNano()

		for _, twt := range tws {

			// generate the time window
			tw := twt(now)

			// we add the info that a booking was made
			if err := c.meter.Add("queues", "bookings", map[string]string{}, tw, 1); err != nil {
				services.Log.Error(err)
			}

		}

	}

	return context.Result(result)

}
