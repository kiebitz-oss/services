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

package servers_test

import (
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/definitions"
	"github.com/kiebitz-oss/services/helpers"
	at "github.com/kiebitz-oss/services/testing"
	af "github.com/kiebitz-oss/services/testing/fixtures"
	"testing"
)

func TestGetAppointmentsByZipCode(t *testing.T) {

	var fixturesConfig = []at.FC{

		// we create the settings
		at.FC{af.Settings{LogLevel: services.InfoLogLevel, Definitions: definitions.Default}, "settings"},

		// we create the appointments API
		at.FC{af.AppointmentsServer{}, "appointmentsServer"},

		// we create a client (without a key)
		at.FC{af.Client{}, "client"},

		// we create a mediator
		at.FC{af.Mediator{}, "mediator"},

		at.FC{af.ProvidersAndAppointments{
			Providers: 10,
			BaseProvider: af.Provider{
				ZipCode:   "10707",
				StoreData: true,
				Confirm:   true,
			},
			BaseAppointments: af.Appointments{
				N:        10,
				Start:    af.TS("2022-10-01T12:00:00Z"),
				Duration: 30,
				Slots:    20,
				Properties: map[string]interface{}{
					"vaccine": "moderna",
				},
			},
		}, "providersAndAppointments"},
	}

	fixtures, err := at.SetupFixtures(fixturesConfig)
	defer at.TeardownFixtures(fixturesConfig, fixtures)

	if err != nil {
		t.Fatal(err)
	}

	client := fixtures["client"].(*helpers.Client)

	if response, err := client.Appointments.GetAppointmentsByZipCode(&services.GetAppointmentsByZipCodeParams{
		ZipCode: "10707",
		Radius:  20,
	}); err != nil {
		t.Fatal(err)
	} else {
		json, err := response.JSON()
		if err != nil {
			t.Fatal(err)
		}
		services.Log.Info(json)
	}

}
