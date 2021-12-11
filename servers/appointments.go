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
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/api"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/forms"
)

// time windows for statistics generation
var tws = []services.TimeWindowFunc{
	services.Minute,
	services.QuarterHour,
	services.Hour,
	services.Day,
	services.Week,
	services.Month,
}

type Appointments struct {
	*Server
	db       services.Database
	backend  *AppointmentsBackend
	meter    services.Meter
	settings *services.AppointmentsSettings
	test     bool
}

func MakeAppointments(settings *services.Settings) (*Appointments, error) {

	appointments := &Appointments{
		db:       settings.DatabaseObj,
		backend:  &AppointmentsBackend{db: settings.DatabaseObj},
		meter:    settings.MeterObj,
		settings: settings.Appointments,
		test:     settings.Test,
	}

	api := &api.API{
		Version: 1,
		Name:    "appointments",
		Endpoints: []*api.Endpoint{
			{
				Name:        "getStats", // unauthenticated
				Description: "Returns various public statistics related to the system.",
				Form:        &forms.GetStatsForm,
				Handler:     appointments.getStats,
				ReturnType: &api.ReturnType{
					Validators: forms.GetStatsRVV,
				},
				REST: &api.REST{
					Path:   "stats",
					Method: api.GET,
				},
			},
			{
				Name:        "getKeys", // unauthenticated
				Description: "Returns various required public keys. Please note that you should have an independent verification mechanism for these keys and not blindly trust the ones provided by this API.",
				Form:        &forms.GetKeysForm,
				Handler:     appointments.getKeys,
				ReturnType: &api.ReturnType{
					Validators: forms.GetKeysRVV,
				},
				REST: &api.REST{
					Path:   "keys",
					Method: api.GET,
				},
			},
			{
				Name:        "getAppointmentsByZipCode", // unauthenticated
				Description: "Returns available appointments for a given zip code area.",
				Form:        &forms.GetAppointmentsByZipCodeForm,
				Handler:     appointments.getAppointmentsByZipCode,
				ReturnType: &api.ReturnType{
					Validators: forms.GetAppointmentsByZipCodeRVV,
				},
				REST: &api.REST{
					Path:   "appointments/zipCode/<zipCode>/<radius>",
					Method: api.GET,
				},
			},
			{
				Name:        "getAppointment", // unauthenticated
				Description: "Returns details about a specific appointment.",
				Form:        &forms.GetAppointmentForm,
				Handler:     appointments.getAppointment,
				ReturnType: &api.ReturnType{
					Validators: forms.GetAppointmentRVV,
				},
				REST: &api.REST{
					Path:   "provider/<providerID>/appointments/<id>",
					Method: api.GET,
				},
			},
			{
				Name:        "getToken", // unauthenticated
				Description: "Returns a signed token that allows users to book appointments.",
				Form:        &forms.GetTokenForm,
				Handler:     appointments.getToken,
				ReturnType: &api.ReturnType{
					Validators: forms.GetTokenRVV,
				},
				REST: &api.REST{
					Path:   "token",
					Method: api.POST,
				},
			},
			{
				Name:        "addMediatorPublicKeys", // authenticted (root)
				Description: "Adds the public key data and associated information of a mediator to the system.",
				Form:        &forms.AddMediatorPublicKeysForm,
				Handler:     appointments.addMediatorPublicKeys,
				ReturnType: &api.ReturnType{
					Validators: forms.IsAcknowledgeRVV,
				},
				REST: &api.REST{
					Path:   "mediators",
					Method: api.POST,
				},
			},
			{
				Name:        "addCodes", // authenticated (root)
				Description: "Adds signup codes to the system.",
				Form:        &forms.AddCodesForm,
				Handler:     appointments.addCodes,
				ReturnType: &api.ReturnType{
					Validators: forms.IsAcknowledgeRVV,
				},
				REST: &api.REST{
					Path:   "codes",
					Method: api.POST,
				},
			},
			{
				Name:        "uploadDistances", // authenticated (root)
				Description: "Uploads distance information to the system.",
				Form:        &forms.UploadDistancesForm,
				Handler:     appointments.uploadDistances,
				ReturnType: &api.ReturnType{
					Validators: forms.IsAcknowledgeRVV,
				},
				REST: &api.REST{
					Path:   "distances",
					Method: api.POST,
				},
			},
			{
				Name:        "resetDB", // authenticated (root)
				Description: "Resets the database. This endpoint is only active for test deployments.",
				Form:        &forms.ResetDBForm,
				Handler:     appointments.resetDB,
				ReturnType: &api.ReturnType{
					Validators: forms.IsAcknowledgeRVV,
				},
				REST: &api.REST{
					Path:   "db/reset",
					Method: api.DELETE,
				},
			},
			{
				Name:        "confirmProvider", // authenticated (mediator)
				Description: "Confirms a provider by adding its public key data and associated information to the system.",
				Form:        &forms.ConfirmProviderForm,
				Handler:     appointments.confirmProvider,
				ReturnType: &api.ReturnType{
					Validators: forms.IsAcknowledgeRVV,
				},
				REST: &api.REST{
					Path:   "providers",
					Method: api.POST,
				},
			},
			{
				Name:        "getPendingProviderData", // authenticated (mediator)
				Description: "Returns a list of provider data waiting for confirmation.",
				Form:        &forms.GetPendingProviderDataForm,
				Handler:     appointments.getPendingProviderData,
				ReturnType: &api.ReturnType{
					Validators: forms.GetProviderDataRVV,
				},
				REST: &api.REST{
					Path:   "providers/pending",
					Method: api.POST,
				},
			},
			{
				Name:        "getVerifiedProviderData", // authenticated (mediator)
				Description: "Returns a list of confirmed provider data.",
				Form:        &forms.GetVerifiedProviderDataForm,
				Handler:     appointments.getVerifiedProviderData,
				ReturnType: &api.ReturnType{
					Validators: forms.GetProviderDataRVV,
				},
				REST: &api.REST{
					Path:   "providers/verified",
					Method: api.POST,
				},
			},
			{
				Name:        "getProviderAppointments", // authenticated (provider)
				Description: "Returns a list of appointments for the given provider.",
				Form:        &forms.GetProviderAppointmentsForm,
				Handler:     appointments.getProviderAppointments,
				ReturnType: &api.ReturnType{
					Validators: forms.GetProviderAppointmentsRVV,
				},
				REST: &api.REST{
					Path:   "appointments",
					Method: api.POST,
				},
			},
			{
				Name:        "publishAppointments", // authenticated (provider)
				Description: "Publishes new or modified appointments to the system.",
				Form:        &forms.PublishAppointmentsForm,
				Handler:     appointments.publishAppointments,
				ReturnType: &api.ReturnType{
					Validators: forms.IsAcknowledgeRVV,
				},
				REST: &api.REST{
					Path:   "appointments/publish",
					Method: api.POST,
				},
			},
			{
				Name:        "storeProviderData", // authenticated (provider)
				Description: "Stores provider data for verification.",
				Form:        &forms.StoreProviderDataForm,
				Handler:     appointments.storeProviderData,
				ReturnType: &api.ReturnType{
					Validators: forms.IsAcknowledgeRVV,
				},
				REST: &api.REST{
					Path:   "providers/data",
					Method: api.POST,
				},
			},
			{
				Name:        "checkProviderData", // authenticated (provider)
				Description: "Checks the verification status of provider data.",
				Form:        &forms.CheckProviderDataForm,
				Handler:     appointments.checkProviderData,
				ReturnType: &api.ReturnType{
					Validators: forms.CheckProviderDataRVV,
				},
				REST: &api.REST{
					Path:   "providers/data",
					Method: api.POST,
				},
			},
			{
				Name:        "bookAppointment", // authenticated (user)
				Description: "Books an appointment.",
				Form:        &forms.BookAppointmentForm,
				Handler:     appointments.bookAppointment,
				ReturnType: &api.ReturnType{
					Validators: forms.BookAppointmentRVV,
				},
				REST: &api.REST{
					Path:   "appointments/book",
					Method: api.POST,
				},
			},
			{
				Name:        "cancelAppointment", // authenticated (user)
				Description: "Cancels a booking.",
				Form:        &forms.CancelAppointmentForm,
				Handler:     appointments.cancelAppointment,
				ReturnType: &api.ReturnType{
					Validators: forms.IsAcknowledgeRVV,
				},
				REST: &api.REST{
					Path:   "appointments/cancel",
					Method: api.DELETE,
				},
			},
		},
	}

	var err error

	if appointments.Server, err = MakeServer("appointments", settings.Appointments.HTTP, settings.Appointments.JSONRPC, settings.Appointments.REST, api); err != nil {
		return nil, err
	}

	return appointments, nil
}

// Method Handlers

func (c *Appointments) Key(key string) *crypto.Key {
	return c.settings.Key(key)
}

func (c *Appointments) getKeysData() (*services.Keys, error) {

	providerDataKey := c.settings.Key("provider")

	return &services.Keys{
		ProviderData: providerDataKey.PublicKey,
		RootKey:      c.settings.Key("root").PublicKey,
		TokenKey:     c.settings.Key("token").PublicKey,
	}, nil

}

func (c *Appointments) getActorKeys() (*services.KeyLists, error) {

	mediatorKeys, err := c.backend.Keys("mediators").GetAll()

	if err != nil {
		return nil, err
	}

	providerKeys, err := c.backend.Keys("providers").GetAll()

	if err != nil {
		return nil, err
	}

	return &services.KeyLists{
		Providers: providerKeys,
		Mediators: mediatorKeys,
	}, nil
}

// authentication helpers

func (c *Appointments) isUser(context services.Context, params *services.SignedParams) services.Response {

	signedData := &crypto.SignedStringData{
		Data:      params.JSON,
		Signature: params.Signature,
	}

	tokenKey := c.settings.Key("token")

	if tokenKey == nil {
		services.Log.Error("token key missing")
		return context.InternalError()
	}

	// first we verify the signed token against the token key
	if ok, err := tokenKey.VerifyString(signedData); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if !ok {
		return context.Error(400, "invalid token", nil)
	}

	signedTokenData := params.ExtraData.(*services.SignedTokenData)

	// then we ensure the public key matches the key from the signed token data
	if !bytes.Equal(signedTokenData.Data.PublicKey, params.PublicKey) {
		return context.Error(400, "invalid key", nil)
	}

	// then we verify the data was signed with the same key
	if ok, err := crypto.VerifyWithBytes([]byte(params.JSON), params.Signature, params.PublicKey); err != nil {
		services.Log.Error(err)
		return context.InternalError()
	} else if !ok {
		return context.Error(400, "invalid signature", nil)
	}

	if expired(params.Timestamp) {
		return context.Error(410, "signature expired", nil)
	}

	return nil

}

func (c *Appointments) isRoot(context services.Context, params *services.SignedParams) services.Response {
	return isRoot(context, []byte(params.JSON), params.Signature, params.Timestamp, c.settings.Keys)
}

func (c *Appointments) isMediator(context services.Context, params *services.SignedParams) (services.Response, *services.ActorKey) {

	keys, err := c.getActorKeys()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError(), nil
	}

	if resp, key := c.isValidActorSignature(context, []byte(params.JSON), params.Signature, params.PublicKey, keys.Mediators); resp != nil {
		return resp, nil
	} else if expired(params.Timestamp) {
		return context.Error(410, "signature expired", nil), nil
	} else {
		return nil, key
	}
}

func (c *Appointments) isProvider(context services.Context, params *services.SignedParams) (services.Response, *services.ActorKey) {

	keys, err := c.getActorKeys()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError(), nil
	}

	if resp, key := c.isValidActorSignature(context, []byte(params.JSON), params.Signature, params.PublicKey, keys.Providers); resp != nil {
		return resp, nil
	} else if expired(params.Timestamp) {
		return context.Error(410, "signature expired", nil), nil
	} else {
		return nil, key
	}
}

func (c *Appointments) isValidActorSignature(context services.Context, data, signature, publicKey []byte, keyList []*services.ActorKey) (services.Response, *services.ActorKey) {

	actorKey, err := findActorKey(keyList, publicKey)

	if err != nil {
		services.Log.Error(err)
		return context.InternalError(), nil
	}

	if actorKey == nil {
		return context.Error(403, "not authorized", nil), nil
	}

	if ok, err := crypto.VerifyWithBytes(data, signature, publicKey); err != nil {
		services.Log.Error(err)
		return context.InternalError(), nil
	} else if !ok {
		return context.Error(401, "invalid signature", nil), nil
	}

	return nil, actorKey

}
