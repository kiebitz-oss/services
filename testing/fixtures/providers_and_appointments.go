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
	"github.com/kiebitz-oss/services/helpers"
)

type ProvidersAndAppointments struct {
	BaseProvider     Provider
	BaseAppointments Appointments
	Providers        int64
}

type ProviderAndAppointments struct {
	Provider     *helpers.Provider
	Appointments []*services.SignedAppointment
}

func (c ProvidersAndAppointments) Setup(fixtures map[string]interface{}) (interface{}, error) {

	providersAndAppointments := make([]*ProviderAndAppointments, c.Providers)

	for i := int64(0); i < c.Providers; i++ {
		if provider, err := c.BaseProvider.Setup(fixtures); err != nil {
			return nil, err
		} else {
			fixtures["provider"] = provider
			if appointments, err := c.BaseAppointments.Setup(fixtures); err != nil {
				return nil, err
			} else {
				providersAndAppointments = append(providersAndAppointments, &ProviderAndAppointments{
					Provider:     provider.(*helpers.Provider),
					Appointments: appointments.([]*services.SignedAppointment),
				})
			}
		}
	}

	return providersAndAppointments, nil

}

func (c ProvidersAndAppointments) Teardown(fixture interface{}) error {
	return nil
}
