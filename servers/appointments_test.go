package servers_test

import (
	"github.com/kiebitz-oss/services"
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
	}

	fixtures, err := at.SetupFixtures(fixturesConfig)
	defer at.TeardownFixtures(fixturesConfig, fixtures)

	client := fixtures["client"].(*at.Client)

	resp, err := client.Appointments("getKeys", map[string]interface{}{})

	if err != nil {
		t.Fatal(err)
	}

	services.Log.Info(resp.JSON())

	if resp.StatusCode != 200 {
		t.Fatalf("expected a 200 status code, got %d instead", resp.StatusCode)
	}

}
