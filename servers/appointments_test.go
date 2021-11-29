package servers_test

import (
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/helpers"
	at "github.com/kiebitz-oss/services/testing"
	af "github.com/kiebitz-oss/services/testing/fixtures"
	"testing"
)

func TestAppointmentsApi(t *testing.T) {

	var fixturesConfig = []at.FC{

		// we create the settings
		at.FC{af.Settings{}, "settings"},

		// we create the appointments API
		at.FC{af.Appointments{}, "appointments"},

		// we create a client (without a key)
		at.FC{af.Client{}, "client"},

		// we create a mediator
		at.FC{af.Mediator{}, "mediator"},
	}

	fixtures, err := at.SetupFixtures(fixturesConfig)
	defer at.TeardownFixtures(fixturesConfig, fixtures)

	if err != nil {
		t.Fatal(err)
	}

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
