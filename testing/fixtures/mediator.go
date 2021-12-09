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
	"github.com/kiebitz-oss/services/crypto"
	"github.com/kiebitz-oss/services/helpers"
)

type Mediator struct {
}

// Creates a new mediator and
func (c Mediator) Setup(fixtures map[string]interface{}) (interface{}, error) {

	client, ok := fixtures["client"].(*helpers.Client)

	if !ok {
		return nil, fmt.Errorf("client missing")
	}

	mediator, err := crypto.MakeActor("mediator")

	if err != nil {
		return nil, err
	}

	// we add the mediator public keys to the backend
	if resp, err := client.Appointments.AddMediatorPublicKeys(mediator); err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("cannot add mediator keys")
	}

	return mediator, nil

}

func (c Mediator) Teardown(fixture interface{}) error {
	return nil
}
