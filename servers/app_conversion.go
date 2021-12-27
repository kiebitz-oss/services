// Kiebitz - Privacy-Friendly Appointment Scheduling
// Copyright (C) 2021-2021 The Kiebitz Authors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version. Additional terms
// as defined in section 7 of the license (e.g. regarding attribution)
// are specified at https://kiebitz.eu/en/docs/open-source/additional-terms.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
