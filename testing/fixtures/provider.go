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
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/helpers"
)

type Provider struct {
	Name    string
	Street  string
	City    string
	ZipCode string
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

	services.Log.Info(client)

	provider, err := crypto.MakeActor("provider")

	if err != nil {
		return nil, err
	}

	services.Log.Info(provider)

	// we add the provider public keys to the backend
	if _, err := client.Appointments.ConfirmProvider(nil, mediator); err != nil {
		return nil, err
	}

	return provider, nil

}

func (c Provider) Teardown(fixture interface{}) error {
	return nil
}
