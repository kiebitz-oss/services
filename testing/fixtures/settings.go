// Kiebitz - Privacy-Friendly Appointment Scheduling
// Copyright (C) 2021-2021 The Kiebitz Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version. Additional terms
// as defined in section 7 of the license (e.g. regarding attribution)
// are specified at https://kiebitz.eu/en/docs/open-source/additional-terms.
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
	"github.com/kiebitz-oss/services/helpers"
)

type Settings struct {
	Definitions services.Definitions
	LogLevel    services.Level
}

func (c Settings) Setup(fixtures map[string]interface{}) (interface{}, error) {
	if c.LogLevel == services.PanicLogLevel {
		// panic log level is the default but we want debug as default
		c.LogLevel = services.DebugLogLevel
	}
	services.Log.SetLevel(c.LogLevel)

	paths, fs, err := helpers.SettingsPaths()

	services.Log.Info(paths)

	if err != nil {
		return nil, err
	}

	if settings, err := helpers.Settings(paths, fs, &c.Definitions); err != nil {
		return nil, err
	} else if db, err := helpers.InitializeDatabase(settings); err != nil {
		return nil, err
	} else if meter, err := helpers.InitializeMeter(settings); err != nil {
		return nil, err
	} else {
		if db != nil {
			if err := db.Reset(); err != nil {
				return nil, err
			}
		}
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
