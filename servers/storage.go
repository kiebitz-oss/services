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
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/api"
	"github.com/kiebitz-oss/services/forms"
	"github.com/kiebitz-oss/services/jsonrpc"
	"github.com/kiebitz-oss/services/metrics"
	"time"
)

type Storage struct {
	settings      *services.StorageSettings
	server        *jsonrpc.JSONRPCServer
	metricsServer *metrics.PrometheusMetricsServer
	db            services.Database
	test          bool
}

func MakeStorage(settings *services.Settings) (*Storage, error) {

	Storage := &Storage{
		db:       settings.DatabaseObj,
		settings: settings.Storage,
		test:     settings.Test,
	}

	api := &api.API{
		Version: 1,
		Endpoints: []*api.Endpoint{
			{
				Name:    "storeSettings",
				Form:    &forms.StoreSettingsForm,
				Handler: Storage.storeSettings,
			},
			{
				Name:    "getSettings",
				Form:    &forms.GetSettingsForm,
				Handler: Storage.getSettings,
			},
			{
				Name:    "deleteSettings",
				Form:    &forms.DeleteSettingsForm,
				Handler: Storage.deleteSettings,
			},
			{
				Name:    "resetDB",
				Form:    &forms.ResetDBForm,
				Type:    api.Retrieve,
				Handler: Storage.resetDB,
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

	if jsonrpcServer, err := jsonrpc.MakeJSONRPCServer(settings.Storage.RPC, handler, "storage"); err != nil {
		return nil, err
	} else {
		Storage.server = jsonrpcServer
		return Storage, nil
	}
}

func (c *Storage) isRoot(context services.Context, data, signature []byte, timestamp *time.Time) services.Response {
	return isRoot(context, data, signature, timestamp, c.settings.Keys)
}

func (c *Storage) Start() error {
	return c.server.Start()
}

func (c *Storage) Stop() error {
	return c.server.Stop()
}
