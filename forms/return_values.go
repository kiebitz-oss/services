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

package forms

import (
	"github.com/kiprotect/go-helpers/forms"
)

var GetTokenRVV = []forms.Validator{
	forms.IsStringMap{
		Form: &SignedTokenDataForm,
	},
}

var GetAppointmentRVV = []forms.Validator{
	forms.IsStringMap{
		Form: &SignedAppointmentForm,
	},
}

var BookAppointmentRVV = []forms.Validator{
	forms.IsStringMap{
		Form: &BookingForm,
	},
}

var GetProviderAppointmentsRVV = []forms.Validator{
	forms.IsList{
		Validators: []forms.Validator{
			forms.IsStringMap{
				Form: &SignedAppointmentForm,
			},
		},
	},
}

var CheckProviderDataRVV = []forms.Validator{
	forms.IsStringMap{
		Form: &ConfirmedProviderDataForm,
	},
}

var GetProviderDataRVV = []forms.Validator{
	forms.IsList{
		Validators: []forms.Validator{
			forms.IsStringMap{
				Form: &RawProviderDataForm,
			},
		},
	},
}

var StatsValueForm = forms.Form{
	Name: "statsValue",
	Fields: []forms.Field{
		{
			Name:        "name",
			Description: "Name of the statistics value.",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name:        "from",
			Description: "Beginning of the time window of the statistics value.",
			Validators: []forms.Validator{
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
		{
			Name:        "to",
			Description: "End of the time window of the statistics value.",
			Validators: []forms.Validator{
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
		{
			Name:        "data",
			Description: "Data associate with the statistics value.",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{},
			},
		},
		{
			Name:        "value",
			Description: "The statistics value itself.",
			Validators: []forms.Validator{
				forms.IsInteger{},
			},
		},
	},
}

var IsAcknowledgeRVV = []forms.Validator{
	forms.IsIn{Choices: []interface{}{"ok"}},
}

var GetStatsRVV = []forms.Validator{
	forms.IsList{
		Validators: []forms.Validator{
			forms.IsStringMap{
				Form: &StatsValueForm,
			},
		},
	},
}

var KeysForm = forms.Form{
	Name: "keys",
	Fields: []forms.Field{
		{
			Name:        "providerData",
			Description: "Public provider data key.",
			Validators:  PublicKeyValidators,
		},
		{
			Name:        "rootKey",
			Description: "Public root key.",
			Validators:  PublicKeyValidators,
		},
		{
			Name:        "tokenKey",
			Description: "Public token key.",
			Validators:  PublicKeyValidators,
		},
	},
}

var GetKeysRVV = []forms.Validator{
	forms.IsStringMap{
		Form: &KeysForm,
	},
}

var GetAppointmentsByZipCodeRVV = []forms.Validator{
	forms.IsList{
		Validators: []forms.Validator{
			forms.Or{
				Options: [][]forms.Validator{
					{
						forms.IsStringMap{
							Form: &ProviderAppointmentsForm,
						},
						forms.IsStringMap{
							Form: &AggregatedProviderAppointmentsForm,
						},
					},
				},
			},
		},
	},
}

var ActorKeyForm = forms.Form{
	Name:   "actorKey",
	Fields: SignedDataFields(nil),
}

var KeyChainForm = forms.Form{
	Name: "keyChain",
	Fields: []forms.Field{
		{
			Name:        "provider",
			Description: "Public provider key data.",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &ActorKeyForm,
				},
			},
		},
		{
			Name:        "mediator",
			Description: "Public mediator key data.",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &ActorKeyForm,
				},
			},
		},
	},
}

var ProviderAppointmentsForm = forms.Form{
	Name: "providerAppointments",
	Fields: []forms.Field{
		{
			Name:        "provider",
			Description: "Signed public provider data.",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &SignedProviderDataForm,
				},
			},
		},
		{
			Name:        "keyChain",
			Description: "The chain of keys that signed the provider key.",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &KeyChainForm,
				},
			},
		},
		{
			Name:        "appointments",
			Description: "Appointments offered by the provider.",
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

var AggregatedProviderAppointmentsForm = forms.Form{
	Name: "aggregatedProviderAppointments",
	Fields: []forms.Field{
		{
			Name:        "provider",
			Description: "Signed public provider data.",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &SignedProviderDataForm,
				},
			},
		},
		{
			Name:        "keyChain",
			Description: "The chain of keys that signed the provider key.",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &KeyChainForm,
				},
			},
		},
		{
			Name:        "openAppointments",
			Description: "Number of open appointments offered by the provider, grouped by day.",
			Validators: []forms.Validator{
				forms.IsStringMap{},
			},
		},
	},
}
