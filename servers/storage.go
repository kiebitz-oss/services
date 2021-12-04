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
	"time"
)

type Storage struct {
	*Server
	settings *services.StorageSettings
	db       services.Database
	test     bool
}

func MakeStorage(settings *services.Settings) (*Storage, error) {

	storage := &Storage{
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
				Handler: storage.storeSettings,
			},
			{
				Name:    "getSettings",
				Form:    &forms.GetSettingsForm,
				Handler: storage.getSettings,
			},
			{
				Name:    "deleteSettings",
				Form:    &forms.DeleteSettingsForm,
				Handler: storage.deleteSettings,
			},
			{
				Name:    "resetDB",
				Form:    &forms.ResetDBForm,
				Handler: storage.resetDB,
			},
		},
	}

	var err error

	if storage.Server, err = MakeServer("storage", settings.Storage.HTTP, settings.Storage.JSONRPC, settings.Storage.REST, api); err != nil {
		return nil, err
	}

	return storage, nil

}

func (c *Storage) isRoot(context services.Context, data, signature []byte, timestamp *time.Time) services.Response {
	return isRoot(context, data, signature, timestamp, c.settings.Keys)
}
