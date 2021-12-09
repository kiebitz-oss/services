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

var PublicKeyValidators = []forms.Validator{
	forms.IsBytes{
		Encoding:  "base64",
		MaxLength: 128,
		MinLength: 64,
	},
}

var PublicKeyField = forms.Field{
	Name:       "publicKey",
	Validators: PublicKeyValidators,
}

var SignatureField = forms.Field{
	Name:       "signature",
	Validators: PublicKeyValidators,
}

var OptionalIDField = forms.Field{
	Name: "id",
	Validators: []forms.Validator{
		forms.IsOptional{},
		ID,
	},
}

var IDField = forms.Field{
	Name: "id",
	Validators: []forms.Validator{
		ID,
	},
}

var ProviderIDField = forms.Field{
	Name: "providerID",
	Validators: []forms.Validator{
		ID,
	},
}

var TimestampField = forms.Field{
	Name: "timestamp",
	Validators: []forms.Validator{
		forms.IsTime{
			Format: "rfc3339",
		},
	},
}

var SignedDataFields = func(form *forms.Form) []forms.Field {
	return []forms.Field{
		forms.Field{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsString{},
				JSON{
					Key: "json",
				},
				forms.IsStringMap{
					Form: form,
				},
			},
		},
		SignatureField,
		PublicKeyField,
	}
}

var ConfirmProviderForm = forms.Form{
	Fields: SignedDataFields(&ConfirmProviderDataForm),
}

var RawProviderDataForm = forms.Form{
	Fields: []forms.Field{
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

var ConfirmProviderDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
		{
			Name: "encryptedProviderData",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &EncryptedProviderDataForm,
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
					Form: SignedKeyDataForm(&ProviderKeyDataForm),
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
	Fields: append(SignedDataFields(&ProviderDataForm), OptionalIDField),
}

var SignedKeyDataForm = func(form *forms.Form) *forms.Form {
	return &forms.Form{
		Fields: SignedDataFields(form),
	}
}

var ProviderKeyDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name:       "signing",
			Validators: PublicKeyValidators,
		},
		{
			Name:       "encryption",
			Validators: PublicKeyValidators,
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
	Fields: SignedDataFields(&ResetDBDataForm),
}

var ResetDBDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
	},
}

var AddMediatorPublicKeysForm = forms.Form{
	Fields: SignedDataFields(&AddMediatorPublicKeysDataForm),
}

var AddMediatorPublicKeysDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "signedKeyData",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: SignedKeyDataForm(&MediatorKeyDataForm),
				},
			},
		},
		TimestampField,
	},
}

var MediatorKeyDataForm = forms.Form{
	Fields: []forms.Field{
		{
			Name:       "signing",
			Validators: PublicKeyValidators,
		},
		{
			Name:       "encryption",
			Validators: PublicKeyValidators,
		},
	},
}

// admin endpoints

var AddCodesForm = forms.Form{
	Fields: SignedDataFields(&CodesDataForm),
}

var CodesDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
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
	Fields: SignedDataFields(&DistancesDataForm),
}

var DistancesDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
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
		PublicKeyField,
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
	Fields: SignedDataFields(&TokenDataForm),
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
					HasMin:  true,
					HasMax:  true,
					Min:     5,
					Max:     80,
					Convert: true,
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
		{
			Name: "from",
			Validators: []forms.Validator{
				forms.IsTime{Format: "rfc3339-date"},
			},
		},
		{
			Name: "to",
			Validators: []forms.Validator{
				forms.IsTime{Format: "rfc3339-date"},
			},
		},
		{
			Name: "aggregate",
			Validators: []forms.Validator{
				forms.IsOptional{Default: false},
				forms.IsBoolean{},
			},
		},
	},
	Validator: func(values map[string]interface{}, errorAdder forms.ErrorAdder) error {
		from := values["from"].(time.Time)
		to := values["to"].(time.Time)
		if from.After(to) {
			return fmt.Errorf("'from' value is after 'to' value")
		}
		if to.Sub(from) > time.Hour*24 {
			return fmt.Errorf("date span exceeds 1 day")
		}
		return nil
	},
}

var GetProviderAppointmentsForm = forms.Form{
	Fields: SignedDataFields(&GetProviderAppointmentsDataForm),
}

var GetProviderAppointmentsDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
		{
			Name: "from",
			Validators: []forms.Validator{
				forms.IsTime{Format: "rfc3339-date"},
			},
		},
		{
			Name: "to",
			Validators: []forms.Validator{
				forms.IsTime{Format: "rfc3339-date"},
			},
		},
		{
			Name: "updatedSince",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsTime{Format: "rfc3339"},
			},
		},
	},
	Validator: func(values map[string]interface{}, errorAdder forms.ErrorAdder) error {
		// form validator only gets called if values are valid, so we can
		// perform a type assertion without check here
		from := values["from"].(time.Time)
		to := values["to"].(time.Time)
		if from.After(to) {
			return fmt.Errorf("'from' value is after 'to' value")
		}
		if to.Sub(from) > time.Hour*24*14 {
			return fmt.Errorf("date span exceeds 14 days")
		}
		return nil
	},
}

var PublishAppointmentsForm = forms.Form{
	Fields: SignedDataFields(&PublishAppointmentsDataForm),
}

var PublishAppointmentsDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
		{
			Name: "offers",
			Validators: []forms.Validator{
				forms.IsList{
					Validators: []forms.Validator{
						forms.IsStringMap{
							Form: &SignedAppointmentForm,
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
		IDField,
		PublicKeyField,
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

var SignedAppointmentForm = forms.Form{
	Fields: append(SignedDataFields(&AppointmentDataForm), []forms.Field{
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
	}...),
}

var AppointmentDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
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
		PublicKeyField,
		IDField,
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
		IDField,
	},
}

var GetBookedAppointmentsDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
	},
}
var GetBookedAppointmentsForm = forms.Form{
	Fields: SignedDataFields(&GetBookedAppointmentsDataForm),
}

var CancelBookingDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
		IDField,
	},
}
var CancelBookingForm = forms.Form{
	Fields: SignedDataFields(&CancelBookingDataForm),
}

var BookAppointmentForm = forms.Form{
	Fields: SignedDataFields(&BookAppointmentDataForm),
}

var BookAppointmentDataForm = forms.Form{
	Fields: []forms.Field{
		ProviderIDField,
		IDField,
		TimestampField,
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
	Fields: SignedDataFields(&GetAppointmentDataForm),
}

var GetAppointmentDataForm = forms.Form{
	Fields: []forms.Field{
		IDField,
		ProviderIDField,
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
	Fields: SignedDataFields(&CancelAppointmentDataForm),
}

var CancelAppointmentDataForm = forms.Form{
	Fields: []forms.Field{
		IDField,
		ProviderIDField,
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
	Fields: SignedDataFields(&CheckProviderDataDataForm),
}

var CheckProviderDataDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
	},
}

var StoreProviderDataForm = forms.Form{
	Fields: SignedDataFields(&StoreProviderDataDataForm),
}

var EncryptedProviderDataForm = forms.Form{
	Fields: SignedDataFields(&ECDHEncryptedDataForm),
}

var StoreProviderDataDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
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
	Fields: SignedDataFields(&GetPendingProviderDataDataForm),
}

var GetPendingProviderDataDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
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
	Fields: SignedDataFields(&GetVerifiedProviderDataDataForm),
}

var GetVerifiedProviderDataDataForm = forms.Form{
	Fields: []forms.Field{
		TimestampField,
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
				forms.MatchesRegex{Regexp: regexp.MustCompile(`^[\w\d\-]{0,50}$`)},
			},
		},
		{
			Name: "metric",
			Validators: []forms.Validator{
				forms.IsOptional{Default: ""},
				forms.MatchesRegex{Regexp: regexp.MustCompile(`^[\w\d\-]{0,50}$`)},
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
