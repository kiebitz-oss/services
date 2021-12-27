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
	"fmt"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/helpers"
)

type Provider struct {
	Name        string
	Street      string
	City        string
	ZipCode     string
	Description string
	Accessible  bool
	Confirm     bool
	StoreData   bool
}

// Creates a new provider and
func (c Provider) Setup(fixtures map[string]interface{}) (interface{}, error) {

	client, ok := fixtures["client"].(*helpers.Client)

	if !ok {
		return nil, fmt.Errorf("client missing")
	}

	mediator, ok := fixtures["mediator"].(*crypto.Actor)

	if !ok {
		return nil, fmt.Errorf("mediator missing")
	}

	actor, err := crypto.MakeActor("provider")

	if err != nil {
		return nil, err
	}

	provider := &helpers.Provider{
		Actor: actor,
		PublicData: &services.ProviderData{
			Name:        c.Name,
			Street:      c.Street,
			City:        c.City,
			ZipCode:     c.ZipCode,
			Description: c.Description,
		},
		QueueData: &services.ProviderQueueData{
			ZipCode:    c.ZipCode,
			Accessible: c.Accessible,
		},
	}

	if c.StoreData {
		// we store the provider data in the backend
		if resp, err := client.Appointments.StoreProviderData(provider); err != nil {
			return nil, err
		} else if resp.StatusCode != 200 {
			return nil, fmt.Errorf("cannot store provider data")
		}
	}

	if c.Confirm {
		// we confirm the provider data
		if resp, err := client.Appointments.ConfirmProvider(provider, mediator); err != nil {
			return nil, err
		} else if resp.StatusCode != 200 {
			return nil, fmt.Errorf("cannot confirm provider data")
		}
	}

	return provider, nil

}

func (c Provider) Teardown(fixture interface{}) error {
	return nil
}
