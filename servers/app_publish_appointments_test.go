package servers_test

import (
	"github.com/kiebitz-oss/services"
	at "github.com/kiebitz-oss/services/testing"
	af "github.com/kiebitz-oss/services/testing/fixtures"
	"testing"
)

func TestPublishAppointments(t *testing.T) {

	var fixturesConfig = []at.FC{

		// we create the settings
		at.FC{af.Settings{}, "settings"},

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

		at.FC{af.Appointments{
			Appointments: []*services.Appointment{
				//				{
				//					Duration: 30,
				//				},
			},
		}, "appointments"},
	}

	fixtures, err := at.SetupFixtures(fixturesConfig)
	defer at.TeardownFixtures(fixturesConfig, fixtures)

	if err != nil {
		t.Fatal(err)
	}

}
