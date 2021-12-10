package forms

import (
	"github.com/kiprotect/go-helpers/forms"
)

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
