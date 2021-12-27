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

type IsAnything struct{}

func (a IsAnything) Validate(value interface{}, values map[string]interface{}) (interface{}, error) {
	return value, nil
}

var StoreSettingsForm = forms.Form{
	Name: "storeSettings",
	Fields: []forms.Field{
		{
			Name:        "id",
			Description: "ID under which to store settings.",
			Validators: []forms.Validator{
				ID,
			},
		},
		{
			Name:        "data",
			Description: "Settings to store under the given ID.",
			Validators: []forms.Validator{
				IsAnything{},
			},
		},
	},
}

var GetSettingsForm = forms.Form{
	Name: "getSettings",
	Fields: []forms.Field{
		{
			Name:        "id",
			Description: "ID for which to retrieve settings.",
			Validators: []forms.Validator{
				ID,
			},
		},
	},
}

var DeleteSettingsForm = forms.Form{
	Name: "deleteSettings",
	Fields: []forms.Field{
		{
			Name:        "id",
			Description: "ID for which to delete settings.",
			Validators: []forms.Validator{
				ID,
			},
		},
	},
}
