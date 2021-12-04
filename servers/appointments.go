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
	"github.com/kiebitz-oss/services/http"
	"github.com/kiebitz-oss/services/jsonrpc"
	"github.com/kiebitz-oss/services/metrics"
	"time"
)

type Appointments struct {
	server        *http.HTTPServer
	jsonRPCServer *jsonrpc.JSONRPCServer
	metricsServer *metrics.PrometheusMetricsServer
	db            services.Database
	meter         services.Meter
	settings      *services.AppointmentsSettings
	test          bool
}

func MakeAppointments(settings *services.Settings) (*Appointments, error) {

	Appointments := &Appointments{
		db:       settings.DatabaseObj,
		meter:    settings.MeterObj,
		settings: settings.Appointments,
		test:     settings.Test,
	}

	api := &api.API{
		Version: 1,
		Endpoints: []*api.Endpoint{
			{
				Name:    "resetDB",
				Form:    &forms.ResetDBForm,
				Type:    api.Retrieve,
				Handler: Appointments.resetDB,
			},
			{
				Name:    "confirmProvider",
				Form:    &forms.ConfirmProviderForm,
				Type:    api.Create,
				Handler: Appointments.confirmProvider,
			},
			{
				Name:    "addMediatorPublicKeys",
				Form:    &forms.AddMediatorPublicKeysForm,
				Type:    api.Create,
				Handler: Appointments.addMediatorPublicKeys,
			},
			{
				Name:    "addCodes",
				Form:    &forms.AddCodesForm,
				Handler: Appointments.addCodes,
			},
			{
				Name:    "uploadDistances",
				Form:    &forms.UploadDistancesForm,
				Handler: Appointments.uploadDistances,
			},
			{
				Name:    "getStats",
				Form:    &forms.GetStatsForm,
				Handler: Appointments.getStats,
			},
			{
				Name:    "getKeys",
				Form:    &forms.GetKeysForm,
				Handler: Appointments.getKeys,
			},
			{
				Name:    "getAppointmentsByZipCode",
				Form:    &forms.GetAppointmentsByZipCodeForm,
				Handler: Appointments.getAppointmentsByZipCode,
			},
			{
				Name:    "getAppointment",
				Form:    &forms.GetAppointmentForm,
				Handler: Appointments.getAppointment,
			},
			{
				Name:    "getProviderAppointments",
				Form:    &forms.GetProviderAppointmentsForm,
				Handler: Appointments.getProviderAppointments,
			},
			{
				Name:    "publishAppointments",
				Form:    &forms.PublishAppointmentsForm,
				Handler: Appointments.publishAppointments,
			},
			{
				Name:    "bookAppointment",
				Form:    &forms.BookAppointmentForm,
				Handler: Appointments.bookAppointment,
			},
			{
				Name:    "cancelAppointment",
				Form:    &forms.CancelAppointmentForm,
				Handler: Appointments.cancelAppointment,
			},
			{
				Name:    "getToken",
				Form:    &forms.GetTokenForm,
				Handler: Appointments.getToken,
			},
			{
				Name:    "storeProviderData",
				Form:    &forms.StoreProviderDataForm,
				Handler: Appointments.storeProviderData,
			},
			{
				Name:    "checkProviderData",
				Form:    &forms.CheckProviderDataForm,
				Handler: Appointments.checkProviderData,
			},
			{
				Name:    "getPendingProviderData",
				Form:    &forms.GetPendingProviderDataForm,
				Handler: Appointments.getPendingProviderData,
			},
			{
				Name:    "getVerifiedProviderData",
				Form:    &forms.GetVerifiedProviderDataForm,
				Handler: Appointments.getVerifiedProviderData,
			},
		},
	}

	methods, err := api.ToJSONRPC()

	if err != nil {
		return nil, err
	}

	handler, err := jsonrpc.MethodsHandler(methods)

	if err != nil {
		return nil, err
	}

	if server, err := http.MakeHTTPServer(settings.Appointments.HTTP, nil, "appointments"); err != nil {
		return nil, err
	} else if jsonrpcServer, err := jsonrpc.MakeJSONRPCServer(settings.Appointments.JSONRPC, handler, "appointments", server); err != nil {
		return nil, err
	} else {
		Appointments.jsonRPCServer = jsonrpcServer
		Appointments.server = server
		return Appointments, nil
	}
}

func (c *Appointments) Start() error {
	// we start the JSONRPC server first to avoid passing HTTP requests to it before it is initialized
	if err := c.jsonRPCServer.Start(); err != nil {
		return err
	}
	if err := c.server.Start(); err != nil {
		return err
	}
	return nil
}

func (c *Appointments) Stop() error {
	// we stop the HTTP server first to avoid the JSONRPC server receiving requests when it is already stopped
	if err := c.server.Stop(); err != nil {
		return err
	}
	if err := c.jsonRPCServer.Stop(); err != nil {
		return err
	}
	return nil
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
