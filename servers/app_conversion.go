package servers

import (
	"encoding/json"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/forms"
)

func SignedAppointment(data []byte) (*services.SignedAppointment, error) {
	var mapData map[string]interface{}
	signedAppointment := &services.SignedAppointment{}
	if err := json.Unmarshal(data, &mapData); err != nil {
		return nil, err
	} else if params, err := forms.SignedAppointmentForm.Validate(mapData); err != nil {
		return nil, err
	} else if err := forms.SignedAppointmentForm.Coerce(signedAppointment, params); err != nil {
		return nil, err
	} else {
		return signedAppointment, nil
	}
}

func SignedProviderData(data []byte) (*services.SignedProviderData, error) {
	providerData := &services.SignedProviderData{}
	var providerDataMap map[string]interface{}
	if err := json.Unmarshal(data, &providerDataMap); err != nil {
		return nil, err
	}
	if params, err := forms.SignedProviderDataForm.Validate(providerDataMap); err != nil {
		return nil, err
	} else if err := forms.SignedProviderDataForm.Coerce(providerData, params); err != nil {
		return nil, err
	}
	return providerData, nil

}
