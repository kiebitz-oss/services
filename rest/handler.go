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

package rest

import (
	"fmt"
	"github.com/kiebitz-oss/services"
	"github.com/kiprotect/go-helpers/forms"
	"regexp"
)

type Method struct {
	Form       *forms.Form
	Handler    interface{}
	Path       string         `json:"path"`
	Method     string         `json:"method"`
	urlParams  []string       `json:"urlParams"`
	pathRegexp *regexp.Regexp `json:"-"`
}

var pathRegexp = regexp.MustCompile(`(?:<[a-zA-Z0-9_]+>)|(?:[^<]*)`)

func (m *Method) Matches(path, method string) (bool, map[string]interface{}) {
	if method != m.Method {
		return false, nil
	}
	if m.pathRegexp == nil {
		pathComponents := pathRegexp.FindAllStringSubmatch(m.Path, -1)
		pathParamsRegexp := "^/"
		urlParams := []string{}
		for _, component := range pathComponents {
			c := component[0]
			if c[0] == '<' {
				// this is a URL parameter
				name := c[1 : len(c)-1]
				urlParams = append(urlParams, name)
				pathParamsRegexp += fmt.Sprintf("(?P<%s>[^/]+)", name)
			} else {
				// this is a literal
				pathParamsRegexp += regexp.QuoteMeta(c)
			}
		}
		pathParamsRegexp += "$"

		// this should never fail but if it does we prefer to panic than log an error
		m.pathRegexp = regexp.MustCompile(pathParamsRegexp)
		m.urlParams = urlParams
	}

	if groups := m.pathRegexp.FindStringSubmatch(path); groups != nil {
		params := map[string]interface{}{}
		for i, group := range groups {
			if i == 0 {
				continue
			}
			params[m.urlParams[i-1]] = group
		}
		return true, params
	}

	return false, nil
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

		request, response := ExtractRESTRequest(context, methods)

		if response != nil {
			return response
		}

		context.Request = request

		return services.HandleAPICall(request.Method.Handler, request.Method.Form, context).(*Response)
	}, nil
}
