package servers_test

import (
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/helpers"
	at "github.com/kiebitz-oss/services/testing"
	af "github.com/kiebitz-oss/services/testing/fixtures"
	"testing"
)

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
		at.FC{af.Provider{}, "provider"},
	}

	fixtures, err := at.SetupFixtures(fixturesConfig)

	if err != nil {
		t.Fatal(err)
	}

	defer at.TeardownFixtures(fixturesConfig, fixtures)

	client := fixtures["client"].(*helpers.Client)

	resp, err := client.Appointments.GetKeys()

	if err != nil {
		t.Fatal(err)
	}

	services.Log.Info(resp.JSON())

	if resp.StatusCode != 200 {
		t.Fatalf("expected a 200 status code, got %d instead", resp.StatusCode)
	}

}
