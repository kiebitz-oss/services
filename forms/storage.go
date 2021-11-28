package forms

import (
	"github.com/kiprotect/go-helpers/forms"
)

type IsAnything struct{}

func (a IsAnything) Validate(value interface{}, values map[string]interface{}) (interface{}, error) {
	return value, nil
}

var StoreSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "id",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name: "data",
			Validators: []forms.Validator{
				IsAnything{},
			},
		},
	},
}

var GetSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "id",
			Validators: []forms.Validator{
				ID,
			},
		},
	},
}

var DeleteSettingsForm = forms.Form{
	Fields: []forms.Field{
		{
			Name: "id",
			Validators: []forms.Validator{
				ID,
			},
		},
	},
}
