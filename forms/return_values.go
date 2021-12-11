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
		Form: &EncryptedProviderDataForm,
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
			forms.IsStringMap{
				Form: &ProviderAppointmentsForm,
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
			Name:        "offers",
			Description: "Appointment offers for the provider.",
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
