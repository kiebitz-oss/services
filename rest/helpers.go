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
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
)

var jsonContentTypeRegexp = regexp.MustCompile("(?i)^application/json(?:;.*)?$")

func getParams(url *url.URL) map[string]interface{} {
	params := map[string]interface{}{}
	for k, v := range url.Query() {
		if len(v) == 1 {
			params[k] = v[0]
		} else {
			params[k] = v
		}
	}
	return params
}

func ExtractRESTRequest(c *Context, methods map[string]*Method) (*Request, *Response) {

	var params map[string]interface{}

	if c.HTTP.Request.Method != "GET" {

		if !jsonContentTypeRegexp.MatchString(c.HTTP.Request.Header.Get("content-type")) {
			return nil, c.Error(400, "invalid JSON", nil).(*Response)
		}

		decoder := json.NewDecoder(c.HTTP.Request.Body)
		if err := decoder.Decode(&params); err != nil {
			return nil, c.Error(400, "invalid JSON", nil).(*Response)
		}
	} else {
		params = getParams(c.HTTP.Request.URL)
	}

	for _, method := range methods {
		if ok, pathParams := method.Matches(c.HTTP.Request.URL.Path, c.HTTP.Request.Method); ok {
			for k, v := range pathParams {
				if _, ok := params[k]; ok {
					return nil, c.Error(400, fmt.Sprintf("parameter %s supplied twice", k), nil).(*Response)
				}
				params[k] = v
			}

			return &Request{
				Method: method,
				Params: params,
			}, nil

		}
	}

	return nil, c.NotFound().(*Response)
}
