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

package helpers

import (
	"fmt"
	"github.com/kiebitz-oss/services"
	at "github.com/kiebitz-oss/services/testing"
	af "github.com/kiebitz-oss/services/testing/fixtures"
	"github.com/urfave/cli"
)

type SettingsFixture struct {
	settings *services.Settings
}

func (s SettingsFixture) Setup(fixtures map[string]interface{}) (interface{}, error) {
	return s.settings, nil
}

func (s SettingsFixture) Teardown(fixture interface{}) error {
	return nil
}

func benchmark(settings *services.Settings) func(c *cli.Context) error {
	return func(c *cli.Context) error {

		safetyOff := c.Bool("safetyOff")
		appointments := c.Int("appointments")
		providers := c.Int("providers")
		slots := c.Int("slots")

		if !settings.Test && !safetyOff {
			return fmt.Errorf("Non-test system detected, aborting! Override this by setting --safetyOff.")
		}

		var fixturesConfig = []at.FC{

			// we create the settings
			at.FC{SettingsFixture{settings: settings}, "settings"},

			// we create a client (without a key)
			at.FC{af.Client{}, "client"},

			// we create a mediator
			at.FC{af.Mediator{}, "mediator"},

			at.FC{af.ProvidersAndAppointments{
				Providers: int64(providers),
				BaseProvider: af.Provider{
					ZipCode:   "10707",
					Name:      "Dr. Maier Müller",
					Street:    "Mühlenstr. 55",
					City:      "Berlin",
					StoreData: true,
					Confirm:   true,
				},
				BaseAppointments: af.Appointments{
					N:        int64(appointments),
					Start:    af.TS("2022-10-01T12:00:00Z"),
					Duration: 30,
					Slots:    int64(slots),
					Properties: map[string]interface{}{
						"vaccine": "moderna",
					},
				},
			}, "providersAndAppointments"},
		}

		fixtures, err := at.SetupFixtures(fixturesConfig)

		if err != nil {
			return err
		}

		if err := at.TeardownFixtures(fixturesConfig, fixtures); err != nil {
			return err
		}

		return nil
	}
}

func Testing(settings *services.Settings) ([]cli.Command, error) {

	return []cli.Command{
		{
			Name:    "testing",
			Aliases: []string{"a"},
			Flags:   []cli.Flag{},
			Usage:   "Test functionality.",
			Subcommands: []cli.Command{
				{
					Name: "benchmark",
					Flags: []cli.Flag{
						&cli.IntFlag{
							Name:  "providers",
							Value: 1000,
							Usage: "number of providers to create",
						},
						&cli.IntFlag{
							Name:  "appointments",
							Value: 1000,
							Usage: "number of appointments to create per provider",
						},
						&cli.IntFlag{
							Name:  "slots",
							Value: 20,
							Usage: "number of slots per appointments to create",
						},
						&cli.BoolFlag{
							Name:  "safetyOff",
							Usage: "set if you want to run this test on a non-test system",
						},
					},
					Usage:  "Run a comprehensive benchmark.",
					Action: benchmark(settings),
				},
			},
		},
	}, nil
}
