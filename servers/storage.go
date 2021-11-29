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
	"github.com/kiebitz-oss/services/forms"
	"github.com/kiebitz-oss/services/jsonrpc"
	"github.com/kiebitz-oss/services/metrics"
)

type Storage struct {
	settings      *services.StorageSettings
	server        *jsonrpc.JSONRPCServer
	metricsServer *metrics.PrometheusMetricsServer
	db            services.Database
}

func MakeStorage(settings *services.Settings) (*Storage, error) {

	Storage := &Storage{
		db:       settings.DatabaseObj,
		settings: settings.Storage,
	}

	methods := map[string]*jsonrpc.Method{
		"storeSettings": {
			Form:    &forms.StoreSettingsForm,
			Handler: Storage.storeSettings,
		},
		"getSettings": {
			Form:    &forms.GetSettingsForm,
			Handler: Storage.getSettings,
		},
		"deleteSettings": {
			Form:    &forms.DeleteSettingsForm,
			Handler: Storage.deleteSettings,
		},
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

func (c *Storage) Start() error {
	return c.server.Start()
}

func (c *Storage) Stop() error {
	return c.server.Stop()
}
