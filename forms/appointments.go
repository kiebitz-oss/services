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

package forms

import (
	"encoding/json"
	"fmt"
	"github.com/kiprotect/go-helpers/forms"
	"regexp"
	"time"
)

// An ID must be between 8 and 32 bytes long
var ID = forms.IsBytes{
	Encoding:  "base64",
	MinLength: 32,
	MaxLength: 32,
}

type JSON struct {
	Key string
}

func UsageValidator(values map[string]interface{}, addError forms.ErrorAdder) error {
	if values["from"] != nil && values["to"] == nil || values["to"] != nil && values["from"] == nil {
		return fmt.Errorf("both from and to must be specified")
	}
	if values["from"] != nil && values["n"] != nil {
		return fmt.Errorf("cannot specify both n and from/to")
	}
	if values["n"] == nil && values["from"] == nil {
		return fmt.Errorf("you need to specify either n or from/to")
	}
	if values["from"] != nil {
		fromT := values["from"].(time.Time)
		toT := values["to"].(time.Time)
		if fromT.UnixNano() > toT.UnixNano() {
			return fmt.Errorf("from date must be before to date")
		}
	}
	return nil
}

func (j JSON) Validate(value interface{}, values map[string]interface{}) (interface{}, error) {
	var jsonValue interface{}
	if err := json.Unmarshal([]byte(value.(string)), &jsonValue); err != nil {
		return nil, err
	}
	// we assign the original value to the given key
	if j.Key != "" {
		values[j.Key] = value
	}
	return jsonValue, nil
}

var ConfirmProviderForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &ConfirmProviderDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var ConfirmProviderDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "encryptedProviderData",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &ECDHEncryptedDataForm,
				},
			},
		},
		{
			Name: "publicProviderData",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{
					Form: &SignedProviderDataForm,
				},
			},
		},
		{
			Name: "signedKeyData",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &SignedDataForm,
				},
			},
		},
	},
}

var ProviderDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "street",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "city",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "zipCode",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
	},
}

var SignedProviderDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &ProviderDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 30,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 30,
				},
			},
		},
	},
}

var SignedDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &KeyDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 30,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 30,
				},
			},
		},
	},
}

var KeyDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "signing",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 30,
				},
			},
		},
		{
			Name: "encryption",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 30,
				},
			},
		},
		{
			Name: "queueData",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &ProviderQueueDataForm,
				},
			},
		},
	},
}

var ProviderQueueDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "zipCode",
			Validators: []forms.Validator{
				forms.IsString{
					MaxLength: 5,
					MinLength: 5,
				},
			},
		},
		{
			Name: "accessible",
			Validators: []forms.Validator{
				forms.IsOptional{Default: false},
				forms.IsBoolean{},
			},
		},
	},
}

var ResetDBForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &ResetDBDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var ResetDBDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "timestamp",
			Validators: []forms.Validator{
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
	},
}

var AddMediatorPublicKeysForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &AddMediatorPublicKeysDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var AddMediatorPublicKeysDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "timestamp",
			Validators: []forms.Validator{
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
		{
			Name: "encryption",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding: "base64",
				},
			},
		},
		{
			Name: "signing",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding: "base64",
				},
			},
		},
	},
}

// admin endpoints

var AddCodesForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &CodesDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var CodesDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "timestamp",
			Validators: []forms.Validator{
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
		{
			Name: "actor",
			Validators: []forms.Validator{
				forms.IsString{},
				forms.IsIn{Choices: []interface{}{"provider", "user"}},
			},
		},
		{
			Name: "codes",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsBytes{
							Encoding:  "hex",
							MaxLength: 32,
							MinLength: 16,
						},
					},
				},
			},
		},
	},
}

var UploadDistancesForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &DistancesDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var DistancesDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "timestamp",
			Validators: []forms.Validator{
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
		{
			Name: "type",
			Validators: []forms.Validator{
				forms.IsIn{Choices: []interface{}{"zipCode", "zipArea"}},
			},
		},
		{
			Name: "distances",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &DistanceForm,
						},
					},
				},
			},
		},
	},
}

var DistanceForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "from",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "to",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "distance",
			Validators: []forms.Validator{
				forms.IsFloat{
					HasMin: true,
					Min:    0.0,
					HasMax: true,
					Max:    200.0,
				},
			},
		},
	},
}

var GetKeysForm = forms.Form{
	Fields: []forms.Field{},
}

var GetTokenForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "hash",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "code",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "hex", // we encode this as hex since it gets passed in URLs
					MinLength: 16,
					MaxLength: 32,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var TokenQueueDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "zipCode",
			Validators: []forms.Validator{
				forms.IsString{
					MaxLength: 5,
					MinLength: 5,
				},
			},
		},
		{
			Name: "distance",
			Validators: []forms.Validator{
				forms.IsOptional{Default: 5},
				forms.IsInteger{
					HasMin: true,
					HasMax: true,
					Min:    5,
					Max:    50,
				},
			},
		},
		{
			Name: "accessible",
			Validators: []forms.Validator{
				forms.IsOptional{Default: false},
				forms.IsBoolean{},
			},
		},
		{
			Name: "offerReceived",
			Validators: []forms.Validator{
				forms.IsOptional{Default: false},
				forms.IsBoolean{},
			},
		},
		{
			Name: "offerAccepted",
			Validators: []forms.Validator{
				forms.IsOptional{Default: false},
				forms.IsBoolean{},
			},
		},
	},
}

var SignedTokenDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &TokenDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var TokenDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "hash",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "token",
			Validators: []forms.Validator{
				ID,
			},
		},
	},
}

var GetAppointmentsByZipCodeForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "radius",
			Validators: []forms.Validator{
				forms.IsOptional{Default: 50},
				forms.IsInteger{
					HasMin: true,
					HasMax: true,
					Min:    5,
					Max:    80,
				},
			},
		},
		{
			Name: "zipCode",
			Validators: []forms.Validator{
				forms.IsString{
					MaxLength: 5,
					MinLength: 5,
				},
			},
		},
	},
}

var GetProviderAppointmentsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &GetProviderAppointmentsDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var GetProviderAppointmentsDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "timestamp",
			Validators: []forms.Validator{
				forms.IsTime{Format: "rfc3339"},
			},
		},
	},
}

var PublishAppointmentsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &PublishAppointmentsDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var PublishAppointmentsDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "timestamp",
			Validators: []forms.Validator{
				forms.IsTime{Format: "rfc3339"},
			},
		},
		{
			Name: "offers",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &AppointmentForm,
						},
					},
				},
			},
		},
	},
}

var AppointmentPropertiesForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "vaccine",
			Validators: []forms.Validator{
				forms.IsIn{Choices: []interface{}{"biontech", "moderna", "astrazeneca", "johnson-johnson"}},
			},
		},
	},
}

var BookingForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "id",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "token",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "encryptedData",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &ECDHEncryptedDataForm,
				},
			},
		},
	},
}

var AppointmentForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "updatedAt",
			Validators: []forms.Validator{
				forms.IsOptional{}, // only for reading, not for submitting
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
		{
			Name: "bookings",
			Validators: []forms.Validator{
				forms.IsOptional{}, // only for reading, not for submitting
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &BookingForm,
						},
					},
				},
			},
		},
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &AppointmentDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var AppointmentDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "timestamp",
			Validators: []forms.Validator{
				forms.IsTime{Format: "rfc3339"},
			},
		},
		{
			Name: "duration",
			Validators: []forms.Validator{
				forms.IsInteger{
					HasMin: true,
					HasMax: true,
					Min:    5,
					Max:    300,
				},
			},
		},
		{
			Name: "properties",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &AppointmentPropertiesForm,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "id",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "slotData",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &SlotForm,
						},
					},
				},
			},
		},
	},
}

var SlotForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "id",
			Validators: []forms.Validator{
				ID,
			},
		},
	},
}

var GetBookedAppointmentsDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "timestamp",
			Validators: []forms.Validator{
				forms.IsTime{Format: "rfc3339"},
			},
		},
	},
}
var GetBookedAppointmentsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &GetBookedAppointmentsDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var CancelBookingDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "timestamp",
			Validators: []forms.Validator{
				forms.IsTime{Format: "rfc3339"},
			},
		},
		{
			Name: "id",
			Validators: []forms.Validator{
				ID,
			},
		},
	},
}
var CancelBookingForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &CancelBookingDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var BookAppointmentForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &BookAppointmentDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var BookAppointmentDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "providerID",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "id",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "timestamp",
			Validators: []forms.Validator{
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
		{
			Name: "signedTokenData",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &SignedTokenDataForm,
				},
			},
		},
		{
			Name: "encryptedData",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &ECDHEncryptedDataForm,
				},
			},
		},
	},
}

var GetAppointmentForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &GetAppointmentDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var GetAppointmentDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "id",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "providerID",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "signedTokenData",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &SignedTokenDataForm,
				},
			},
		},
	},
}

var CancelAppointmentForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &CancelAppointmentDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var CancelAppointmentDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "id",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "providerID",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "signedTokenData",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &SignedTokenDataForm,
				},
			},
		},
	},
}

var CheckProviderDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &CheckProviderDataDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var CheckProviderDataDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "timestamp",
			Validators: []forms.Validator{
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
	},
}

var StoreProviderDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &StoreProviderDataDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var StoreProviderDataDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "code",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "hex", // we encode this as hex since it gets passed in URLs
					MinLength: 16,
					MaxLength: 32,
				},
			},
		},
		{
			Name: "encryptedData",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &ECDHEncryptedDataForm,
				},
			},
		},
	},
}

var GetPendingProviderDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &GetPendingProviderDataDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var GetPendingProviderDataDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "limit",
			Validators: []forms.Validator{
				forms.IsOptional{Default: 1000},
				forms.IsInteger{
					HasMin: true,
					HasMax: true,
					Min:    1,
					Max:    10000,
				},
			},
		},
	},
}

var GetVerifiedProviderDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: &GetPendingProviderDataDataForm,
				},
			},
		},
		{
			Name: "signature",
			Validators: []forms.Validator{
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
		{
			Name: "publicKey",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsBytes{
					Encoding:  "base64",
					MaxLength: 1000,
					MinLength: 50,
				},
			},
		},
	},
}

var GetVerifiedProviderDataDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "limit",
			Validators: []forms.Validator{
				forms.IsOptional{Default: 1000},
				forms.IsInteger{
					HasMin: true,
					HasMax: true,
					Min:    1,
					Max:    10000,
				},
			},
		},
	},
}

var GetStatsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "id",
			Validators: []forms.Validator{
				forms.IsIn{Choices: []interface{}{"queues", "tokens"}},
			},
		},
		{
			Name: "type",
			Validators: []forms.Validator{
				forms.IsIn{Choices: []interface{}{"minute", "hour", "day", "quarterHour", "week", "month"}},
			},
		},
		{
			Name: "name",
			Validators: []forms.Validator{
				forms.IsOptional{Default: ""},
				forms.MatchesRegex{Regex: regexp.MustCompile(`^[\w\d\-]{0,50}$`)},
			},
		},
		{
			Name: "metric",
			Validators: []forms.Validator{
				forms.IsOptional{Default: ""},
				forms.MatchesRegex{Regex: regexp.MustCompile(`^[\w\d\-]{0,50}$`)},
			},
		},
		{
			Name: "filter",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{},
			},
		},
		{
			Name: "from",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsTime{Format: "rfc3339", ToUTC: true},
			},
		},
		{
			Name: "to",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsTime{Format: "rfc3339", ToUTC: true},
			},
		},
		{
			Name: "n",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsInteger{HasMin: true, Min: 1, HasMax: true, Max: 500, Convert: true},
			},
		},
	},
	Transforms: []forms.Transform{},
	Validator:  UsageValidator,
}
