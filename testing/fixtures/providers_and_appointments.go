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
	"sync"
	"time"
)

type ProvidersAndAppointments struct {
	BaseProvider     Provider
	BaseAppointments Appointments
	Providers        int64
	Concurrency      int64
}

type ProviderAndAppointments struct {
	Provider     *helpers.Provider
	Appointments []*services.SignedAppointment
}

func (c ProvidersAndAppointments) Setup(fixtures map[string]interface{}) (interface{}, error) {

	if c.Concurrency == 0 {
		c.Concurrency = 1
	}

	var mutex sync.Mutex
	var workerErr error
	workChannels := make(chan bool, c.Concurrency)

	providersAndAppointments := make([]*ProviderAndAppointments, c.Providers)

	start := time.Now()

	for i := int64(0); i < c.Providers; i++ {
		if i%10 == 0 && i > c.Concurrency*2 {
			elapsed := time.Now().Sub(start).Seconds()
			et := float64(i) / elapsed
			services.Log.Debugf("Created %d providers, %.2f providers / second (%.2f appointments / second)", i, et, et*float64(c.BaseAppointments.N))
		}

		mutex.Lock()
		if workerErr != nil {
			services.Log.Error(workerErr)
			return workerErr, nil
		}
		mutex.Unlock()

		workChannels <- true

		go func(x int64) {
			defer func() {
				<-workChannels
			}()
			if provider, err := c.BaseProvider.Setup(fixtures); err != nil {
				mutex.Lock()
				defer mutex.Unlock()
				workerErr = err
				return
			} else {
				newFixtures := map[string]interface{}{}
				for k, v := range fixtures {
					newFixtures[k] = v
				}
				newFixtures["provider"] = provider
				if appointments, err := c.BaseAppointments.Setup(newFixtures); err != nil {
					mutex.Lock()
					defer mutex.Unlock()
					workerErr = err
					return
				} else {
					providersAndAppointments = append(providersAndAppointments, &ProviderAndAppointments{
						Provider:     provider.(*helpers.Provider),
						Appointments: appointments.([]*services.SignedAppointment),
					})
				}
			}
		}(i)
	}

	// we wait for the remaining work to be done
	for i := int64(0); i < c.Concurrency; i++ {
		<-workChannels
	}

	return providersAndAppointments, nil

}

func (c ProvidersAndAppointments) Teardown(fixture interface{}) error {
	return nil
}
