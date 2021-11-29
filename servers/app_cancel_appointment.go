package servers

import (
	"bytes"
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/forms"
	"github.com/kiebitz-oss/services/jsonrpc"
	"time"
)

func (c *Appointments) cancelAppointment(context *jsonrpc.Context, params *services.CancelAppointmentSignedParams) *jsonrpc.Response {
	// we verify the signature (without veryfing e.g. the provenance of the key)
	if ok, err := crypto.VerifyWithBytes([]byte(params.JSON), params.Signature, params.PublicKey); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if !ok {
		return context.Error(400, "invalid signature", nil)
	}

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

	lock, err := c.db.Lock("cancelAppointment_" + string(params.Data.ProviderID[:]))
	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	defer lock.Release()

	appointmentsByID := c.db.Map("appointmentsByID", params.Data.ProviderID)

	if date, err := appointmentsByID.Get(params.Data.ID); err != nil {
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
			} else if params, err := forms.AppointmentForm.Validate(mapData); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			} else if err := forms.AppointmentForm.Coerce(signedAppointment, params); err != nil {
				services.Log.Error(err)
				return context.InternalError()
			}

			// we try to find an open slot

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

			usedTokens := c.db.Set("bookings", []byte("tokens"))

			// we mark the token as unused
			if err := usedTokens.Del(token); err != nil {
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

	return context.Acknowledge()

}
