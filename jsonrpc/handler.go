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

package jsonrpc

import (
	"fmt"
	"github.com/kiebitz-oss/services"
	"github.com/kiprotect/go-helpers/forms"
)

type Method struct {
	Form    *forms.Form
	Handler interface{}
}

func MethodsHandler(methods map[string]*Method) (Handler, error) {

	// we check that all provided methods have the correct type
	for key, method := range methods {
		if _, err := services.APIHandlerStruct(method.Handler); err != nil {
			return nil, err
		}
		if method.Form == nil {
			return nil, fmt.Errorf("form for method %s missing", key)
		}
	}

	return func(context *Context) *Response {
		if method, ok := methods[context.Request.Method]; !ok {
			return context.MethodNotFound().(*Response)
		} else {
			return services.HandleAPICall(method.Handler, method.Form, context).(*Response)
		}
	}, nil
}
