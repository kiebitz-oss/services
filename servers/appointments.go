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
	"github.com/kiebitz-oss/services/api"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/forms"
	"time"
)

type Appointments struct {
	*Server
	db       services.Database
	meter    services.Meter
	settings *services.AppointmentsSettings
	test     bool
}

func MakeAppointments(settings *services.Settings) (*Appointments, error) {

	appointments := &Appointments{
		db:       settings.DatabaseObj,
		meter:    settings.MeterObj,
		settings: settings.Appointments,
		test:     settings.Test,
	}

	api := &api.API{
		Version: 1,
		Endpoints: []*api.Endpoint{
			{
				Name:    "getStats", // unauthenticated
				Form:    &forms.GetStatsForm,
				Handler: appointments.getStats,
				REST: &api.REST{
					Path:   "stats",
					Method: api.GET,
				},
			},
			{
				Name:    "getKeys", // unauthenticated
				Form:    &forms.GetKeysForm,
				Handler: appointments.getKeys,
				REST: &api.REST{
					Path:   "keys",
					Method: api.GET,
				},
			},
			{
				Name:    "getAppointmentsByZipCode", // unauthenticated
				Form:    &forms.GetAppointmentsByZipCodeForm,
				Handler: appointments.getAppointmentsByZipCode,
				REST: &api.REST{
					Path:   "appointments/zipCode/<zipCode>/<radius>",
					Method: api.GET,
				},
			},
			{
				Name:    "getAppointment", // unauthenticated
				Form:    &forms.GetAppointmentForm,
				Handler: appointments.getAppointment,
				REST: &api.REST{
					Path:   "provider/<providerID>/appointments/<id>",
					Method: api.GET,
				},
			},
			{
				Name:    "getToken", // unauthenticated
				Form:    &forms.GetTokenForm,
				Handler: appointments.getToken,
				REST: &api.REST{
					Path:   "token",
					Method: api.POST,
				},
			},
			{
				Name:    "addMediatorPublicKeys", // authenticted (root)
				Form:    &forms.AddMediatorPublicKeysForm,
				Handler: appointments.addMediatorPublicKeys,
				REST: &api.REST{
					Path:   "mediators",
					Method: api.POST,
				},
			},
			{
				Name:    "revokeMediator", // authenticated (mediator)
				Form:    &forms.RevokeMediatorForm,
				Handler: appointments.revokeMediator,
				REST: &api.REST{
					Path:   "mediators/revoke",
					Method: api.POST,
				},
			},
			{
				Name:    "addCodes", // authenticated (root)
				Form:    &forms.AddCodesForm,
				Handler: appointments.addCodes,
				REST: &api.REST{
					Path:   "codes",
					Method: api.POST,
				},
			},
			{
				Name:    "uploadDistances", // authenticated (root)
				Form:    &forms.UploadDistancesForm,
				Handler: appointments.uploadDistances,
				REST: &api.REST{
					Path:   "distances",
					Method: api.POST,
				},
			},
			{
				Name:    "resetDB", // authenticated (root)
				Form:    &forms.ResetDBForm,
				Handler: appointments.resetDB,
				REST: &api.REST{
					Path:   "db/reset",
					Method: api.DELETE,
				},
			},
			{
				Name:    "confirmProvider", // authenticated (mediator)
				Form:    &forms.ConfirmProviderForm,
				Handler: appointments.confirmProvider,
				REST: &api.REST{
					Path:   "providers",
					Method: api.POST,
				},
			},
			{
				Name:    "revokeProvider", // authenticated (mediator)
				Form:    &forms.RevokeProviderForm,
				Handler: appointments.revokeProvider,
				REST: &api.REST{
					Path:   "providers/revoke",
					Method: api.POST,
				},
			},
			{
				Name:    "getPendingProviderData", // authenticated (mediator)
				Form:    &forms.GetPendingProviderDataForm,
				Handler: appointments.getPendingProviderData,
				REST: &api.REST{
					Path:   "providers/pending",
					Method: api.POST,
				},
			},
			{
				Name:    "getVerifiedProviderData", // authenticated (mediator)
				Form:    &forms.GetVerifiedProviderDataForm,
				Handler: appointments.getVerifiedProviderData,
				REST: &api.REST{
					Path:   "providers/verified",
					Method: api.POST,
				},
			},
			{
				Name:    "getProviderAppointments", // authenticated (provider)
				Form:    &forms.GetProviderAppointmentsForm,
				Handler: appointments.getProviderAppointments,
				REST: &api.REST{
					Path:   "appointments",
					Method: api.POST,
				},
			},
			{
				Name:    "publishAppointments", // authenticated (provider)
				Form:    &forms.PublishAppointmentsForm,
				Handler: appointments.publishAppointments,
				REST: &api.REST{
					Path:   "appointments/publish",
					Method: api.POST,
				},
			},
			{
				Name:    "storeProviderData", // authenticated (provider)
				Form:    &forms.StoreProviderDataForm,
				Handler: appointments.storeProviderData,
				REST: &api.REST{
					Path:   "providers/data",
					Method: api.POST,
				},
			},
			{
				Name:    "checkProviderData", // authenticated (provider)
				Form:    &forms.CheckProviderDataForm,
				Handler: appointments.checkProviderData,
				REST: &api.REST{
					Path:   "providers/data",
					Method: api.POST,
				},
			},
			{
				Name:    "bookAppointment", // authenticated (user)
				Form:    &forms.BookAppointmentForm,
				Handler: appointments.bookAppointment,
				REST: &api.REST{
					Path:   "appointments/book",
					Method: api.POST,
				},
			},
			{
				Name:    "cancelAppointment", // authenticated (user)
				Form:    &forms.CancelAppointmentForm,
				Handler: appointments.cancelAppointment,
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

// signed requests are valid only 1 minute
func expired(timestamp *time.Time) bool {
	return time.Now().Add(-time.Minute).After(*timestamp)
}

// public endpoints

func findActorKey(keys []*services.ActorKey, publicKey []byte) (*services.ActorKey, error) {
	for _, key := range keys {
		if akd, err := key.KeyData(); err != nil {
			services.Log.Error(err)
			continue
		} else if bytes.Equal(akd.Signing, publicKey) {
			return key, nil
		}
	}
	return nil, nil
}

func (c *Appointments) getListKeys(key string) ([]*services.ActorKey, error) {
	mk, err := c.db.Map("keys", []byte(key)).GetAll()

	if err != nil {
		services.Log.Error(err)
		return nil, err
	}

	actorKeys := []*services.ActorKey{}

	for _, v := range mk {
		var m *services.ActorKey
		if err := json.Unmarshal(v, &m); err != nil {
			services.Log.Error(err)
			continue
		} else {
			actorKeys = append(actorKeys, m)
		}
	}

	return actorKeys, nil

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
	mediatorKeys, err := c.getListKeys("mediators")

	if err != nil {
		return nil, err
	}

	providerKeys, err := c.getListKeys("providers")

	if err != nil {
		return nil, err
	}

	return &services.KeyLists{
		Providers: providerKeys,
		Mediators: mediatorKeys,
	}, nil
}

// provider-only endpoints

var tws = []services.TimeWindowFunc{
	services.Minute,
	services.QuarterHour,
	services.Hour,
	services.Day,
	services.Week,
	services.Month,
}

func (c *Appointments) isRoot(context services.Context, data, signature []byte, timestamp *time.Time) services.Response {
	return isRoot(context, data, signature, timestamp, c.settings.Keys)
}

func isRoot(context services.Context, data, signature []byte, timestamp *time.Time, keys []*crypto.Key) services.Response {
	rootKey := services.Key(keys, "root")
	if rootKey == nil {
		services.Log.Error("root key missing")
		return context.InternalError()
	}
	if ok, err := rootKey.Verify(&crypto.SignedData{
		Data:      data,
		Signature: signature,
	}); !ok {
		return context.Error(403, "invalid signature", nil)
	} else if err != nil {
		services.Log.Error(err)
		return context.InternalError()
	}
	if expired(timestamp) {
		return context.Error(410, "signature expired", nil)
	}
	return nil
}

func (c *Appointments) isMediator(context services.Context, data, signature, publicKey []byte) (services.Response, *services.ActorKey) {

	keys, err := c.getActorKeys()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError(), nil
	}

	return c.isOnKeyList(context, data, signature, publicKey, keys.Mediators)
}

func (c *Appointments) isProvider(context services.Context, data, signature, publicKey []byte) (services.Response, *services.ActorKey) {

	keys, err := c.getActorKeys()

	if err != nil {
		services.Log.Error(err)
		return context.InternalError(), nil
	}

	return c.isOnKeyList(context, data, signature, publicKey, keys.Providers)
}

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

func (c *Appointments) isOnKeyList(context services.Context, data, signature, publicKey []byte, keyList []*services.ActorKey) (services.Response, *services.ActorKey) {

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
