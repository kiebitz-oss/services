package servers

import (
	"github.com/kiebitz-oss/services"
)

// The appointments backend acts as an interface between the API and the
// database. It is mostly concerned with ensuring data is propery serialized
// and deserialized when stored or fetched from the database.
type AppointmentsBackend struct {
	db services.Database
}

func (a *AppointmentsBackend) PublicProviderData() *PublicProviderData {
	return &PublicProviderData{
		dbs: a.db.Map("providerData", []byte("public")),
	}
}

func (a *AppointmentsBackend) AppointmentsByDate(providerID []byte, date string) *AppointmentsByDate {
	dateKey := append(providerID, []byte(date)...)
	return &AppointmentsByDate{
		dbs: a.db.Map("appointmentsByDate", dateKey),
	}
}

func (a *AppointmentsBackend) AppointmentDatesByID(providerID []byte) *AppointmentDatesByID {
	return &AppointmentDatesByID{
		dbs: a.db.Map("appointmentDatesByID", providerID),
	}
}

type AppointmentDatesByID struct {
	dbs services.Map
}

func (a *AppointmentDatesByID) GetAll() (map[string][]byte, error) {
	return a.dbs.GetAll()
}

type PublicProviderData struct {
	dbs services.Map
}

func (p *PublicProviderData) Get(hash []byte) (*services.SignedProviderData, error) {
	if data, err := p.dbs.Get(hash); err != nil {
		return nil, err
	} else if signedProviderData, err := SignedProviderData(data); err != nil {
		return nil, err
	} else {
		return signedProviderData, nil
	}
}

type AppointmentsByDate struct {
	dbs services.Map
}

func (a *AppointmentsByDate) Get(id []byte) (*services.SignedAppointment, error) {
	if appointmentData, err := a.dbs.Get(id); err != nil {
		return nil, err
	} else {
		if signedAppointment, err := SignedAppointment(appointmentData); err != nil {
			return nil, err
		} else {
			return signedAppointment, nil
		}
	}
}

func (a *AppointmentsByDate) GetAll() (map[string]*services.SignedAppointment, error) {

	signedAppointments := make(map[string]*services.SignedAppointment)

	if allAppointments, err := a.dbs.GetAll(); err != nil {
		return nil, err
	} else {
		for id, appointmentData := range allAppointments {
			if signedAppointment, err := SignedAppointment(appointmentData); err != nil {
				return nil, err
			} else {
				signedAppointments[id] = signedAppointment
			}
		}

		return signedAppointments, nil
	}
}
