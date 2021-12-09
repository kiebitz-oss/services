package servers

import (
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"time"
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
		providerID: providerID,
		db:         a.db,
		dbs:        a.db.Map("appointmentDatesByID", providerID),
	}
}

func (a *AppointmentsBackend) UsedTokens() *UsedTokens {
	return &UsedTokens{
		dbs: a.db.Set("bookings", []byte("tokens")),
	}
}

type UsedTokens struct {
	dbs services.Set
}

func (t *UsedTokens) Del(token []byte) error {
	return t.dbs.Del(token)
}

func (t *UsedTokens) Has(token []byte) (bool, error) {
	return t.dbs.Has(token)
}

func (t *UsedTokens) Add(token []byte) error {
	return t.dbs.Add(token)
}

type AppointmentDatesByID struct {
	providerID []byte
	dbs        services.Map
	db         services.Database
}

func (a *AppointmentDatesByID) GetAll() (map[string][]byte, error) {
	return a.dbs.GetAll()
}

func (a *AppointmentDatesByID) Get(id []byte) (string, error) {
	if data, err := a.dbs.Get(id); err != nil {
		return "", err
	} else {
		return string(data), nil
	}
}

func (a *AppointmentDatesByID) Set(id []byte, date string) error {
	// ID map will auto-delete after one year (purely for storage reasons, it does not contain sensitive data)
	if err := a.db.Expire("appointmentDatesByID", a.providerID, time.Hour*24*365); err != nil {
		return err
	}
	return a.dbs.Set(id, []byte(date))
}

func (a *AppointmentDatesByID) Del(id []byte) error {
	return a.dbs.Del(id)
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

func (a *AppointmentsByDate) Del(id []byte) error {
	return a.dbs.Del(id)
}

func (a *AppointmentsByDate) Set(appointment *services.SignedAppointment) error {
	if data, err := json.Marshal(appointment); err != nil {
		return err
	} else {
		return a.dbs.Set(appointment.Data.ID, data)
	}
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
