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
	"fmt"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/helpers"
)

type Appointments struct {
}

func (c Appointments) Setup(fixtures map[string]interface{}) (interface{}, error) {

	sett := fixtures["settings"]

	if sett == nil {
		return nil, fmt.Errorf("no settings found")
	}

	settingsObj, ok := sett.(*services.Settings)

	if !ok {
		return nil, fmt.Errorf("not a real settings object")
	}

	if appointments, err := helpers.InitializeAppointmentsServer(settingsObj); err != nil {
		return nil, err
	} else {
		return appointments, nil
	}

}

func (c Appointments) Teardown(fixture interface{}) error {
	return nil
}
