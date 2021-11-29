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

package fixtures

import (
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/definitions"
	"github.com/kiebitz-oss/services/helpers"
)

type Settings struct {
}

func (c Settings) Setup(fixtures map[string]interface{}) (interface{}, error) {
	// we set the loglevel to 'debug' so we can see which settings files are being loaded
	services.Log.SetLevel(services.DebugLogLevel)

	defs, ok := fixtures["definitions"].(services.Definitions)

	if !ok {
		defs = definitions.Default
	}

	paths := helpers.SettingsPaths()

	if settings, err := helpers.Settings(paths, &defs); err != nil {
		return nil, err
	} else if db, err := helpers.InitializeDatabase(settings); err != nil {
		return nil, err
	} else if meter, err := helpers.InitializeMeter(settings); err != nil {
		return nil, err
	} else if err := db.Reset(); err != nil {
		return nil, err
	} else {
		settings.DatabaseObj = db
		settings.MeterObj = meter
		return settings, nil
	}
}

func (c Settings) Teardown(fixture interface{}) error {
	if fixture == nil {
		return nil
	}
	settings := fixture.(*services.Settings)
	// we close the database
	if err := settings.DatabaseObj.Close(); err != nil {
		return err
	}
	return nil
}
