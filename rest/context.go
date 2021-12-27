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

package rest

import (
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/http"
	"regexp"
)

var idRegexp = regexp.MustCompile(`^n:(-?\d{1,32})$`)
var idNRegexp = regexp.MustCompile(`^(n+):(-?\d{1,32})$`)

type Context struct {
	HTTP    *http.Context
	Request *Request
}

func (c *Context) Result(data interface{}) services.Response {

	return &Response{
		StatusCode: 200,
		Data:       data,
	}
}

func (c *Context) Error(code int, message string, data interface{}) services.Response {
	return &Response{
		StatusCode: code,
		Data: &Error{
			Message: message,
			Data:    data,
		},
	}
}

func (c *Context) Params() map[string]interface{} {
	return c.Request.Params
}

func (c *Context) NotFound() services.Response {
	return c.Error(404, "not found", nil)
}

func (c *Context) Acknowledge() services.Response {
	return c.Result("ok")
}

func (c *Context) Nil() services.Response {
	return c.Result(nil)
}

func (c *Context) MethodNotFound() services.Response {
	return c.Error(404, "method not found", nil)
}

func (c *Context) InvalidParams(err error) services.Response {
	return c.Error(400, "invalid params", err)
}

func (c *Context) InternalError() services.Response {
	return c.Error(500, "internal error", nil)
}
