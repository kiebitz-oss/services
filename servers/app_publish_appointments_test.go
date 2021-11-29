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

package servers_test

import (
	"github.com/kiebitz-oss/services"
	at "github.com/kiebitz-oss/services/testing"
	af "github.com/kiebitz-oss/services/testing/fixtures"
	"testing"
)

func TestPublishAppointments(t *testing.T) {

	var fixturesConfig = []at.FC{

		// we create the settings
		at.FC{af.Settings{}, "settings"},

		// we create the appointments API
		at.FC{af.AppointmentsServer{}, "appointmentsServer"},

		// we create a client (without a key)
		at.FC{af.Client{}, "client"},

		// we create a mediator
		at.FC{af.Mediator{}, "mediator"},

		// we create a mediator
		at.FC{af.Provider{
			ZipCode:   "10707",
			StoreData: true,
			Confirm:   true,
		}, "provider"},

		at.FC{af.Appointments{
			Appointments: []*services.Appointment{
				//				{
				//					Duration: 30,
				//				},
			},
		}, "appointments"},
	}

	fixtures, err := at.SetupFixtures(fixturesConfig)
	defer at.TeardownFixtures(fixturesConfig, fixtures)

	if err != nil {
		t.Fatal(err)
	}

}
