package servers

import (
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/jsonrpc"
)

func (c *Appointments) getProviderAppointments(context *jsonrpc.Context, params *services.GetProviderAppointmentsSignedParams) *jsonrpc.Response {

	// make sure this is a valid provider asking for tokens
	resp, providerKey := c.isProvider(context, []byte(params.JSON), params.Signature, params.PublicKey)

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

	// appointments are stored in a provider-specific key
	appointmentsByID := c.db.Map("appointmentsByID", hash)
	allDates, err := appointmentsByID.GetAll()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}

	signedAppointments := make([]*services.SignedAppointment, 0)

	for _, date := range allDates {
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
			signedAppointments = append(signedAppointments, signedAppointment)
		}
	}

	return context.Result(signedAppointments)
}
