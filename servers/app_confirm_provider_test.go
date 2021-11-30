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
	"encoding/json"
	"github.com/kiebitz-oss/services/definitions"
	"github.com/kiebitz-oss/services/helpers"
	at "github.com/kiebitz-oss/services/testing"
	af "github.com/kiebitz-oss/services/testing/fixtures"
	"testing"
)

type ConfirmProviderResult struct {
	EncryptedProviderData *EncryptedProviderData `json:"encryptedProviderData"`
	SignedKeyData         *SignedKeyData         `json:"signedKeyData"`
}

type SignedKeyData struct {
	Data      string `json:"data"`
	Signature string `json:"signature"`
	PublicKey string `json:"publicKey"`
}

func (s *SignedKeyData) KeyData() (*KeyData, error) {
	keyData := &KeyData{}
	if err := json.Unmarshal([]byte(s.Data), keyData); err != nil {
		return nil, err
	} else {
		return keyData, nil
	}
}

type KeyData struct {
	QueueData *QueueData `json:"queueData"`
}

type QueueData struct {
	ZipCode string `json:"zipCode"`
}

type EncryptedProviderData struct {
	Data      string `json:"data"`
	IV        string `json:"iv"`
	PublicKey string `json:"publicKey"`
}

func TestConfirmProvider(t *testing.T) {

	var fixturesConfig = []at.FC{

		// we create the settings
		at.FC{af.Settings{definitions.Default}, "settings"},

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
	}

	fixtures, err := at.SetupFixtures(fixturesConfig)
	defer at.TeardownFixtures(fixturesConfig, fixtures)

	if err != nil {
		t.Fatal(err)
	}

	client := fixtures["client"].(*helpers.Client)
	provider := fixtures["provider"].(*helpers.Provider)

	resp, err := client.Appointments.CheckProviderData(provider)

	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected a 200 status code, got %d instead", resp.StatusCode)
	}

	result := &ConfirmProviderResult{}

	if err := resp.CoerceResult(result); err != nil {
		t.Fatal(err)
	}

	if keyData, err := result.SignedKeyData.KeyData(); err != nil {
		t.Fatal(err)
	} else if keyData.QueueData.ZipCode != "10707" {
		t.Fatalf("zip code does not match")
	}

}
