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
			Name: "name",
			Validators: []forms.Validator{
				forms.IsString{},
			},
		},
		{
			Name: "from",
			Validators: []forms.Validator{
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
		{
			Name: "to",
			Validators: []forms.Validator{
				forms.IsTime{
					Format: "rfc3339",
				},
			},
		},
		{
			Name: "data",
			Validators: []forms.Validator{
				forms.IsOptional{},
				forms.IsStringMap{},
			},
		},
		{
			Name: "value",
			Validators: []forms.Validator{
				forms.IsInteger{},
			},
		},
	},
}

var IsAcknowledgeRVV = []forms.Validator{
	forms.IsString{},
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
			Name:       "providerData",
			Validators: PublicKeyValidators,
		},
		{
			Name:       "rootKey",
			Validators: PublicKeyValidators,
		},
		{
			Name:       "tokenKey",
			Validators: PublicKeyValidators,
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
			Name: "provider",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &ActorKeyForm,
				},
			},
		},
		{
			Name: "mediator",
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
			Name: "provider",
			Validators: []forms.Validator{
				forms.IsStringMap{
					Form: &SignedProviderDataForm,
				},
			},
		},
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
