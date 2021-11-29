package servers_test

import (
	"encoding/json"
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
		at.FC{af.Settings{}, "settings"},

		// we create the appointments API
		at.FC{af.Appointments{}, "appointments"},

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
