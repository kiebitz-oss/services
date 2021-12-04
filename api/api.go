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

package api

import (
	"github.com/kiprotect/go-helpers/forms"
)

type Method string

const (
	POST   = Method("POST")
	GET    = Method("GET")
	PUT    = Method("PUT")
	DELETE = Method("DELETE")
	PATCH  = Method("PATCH")
)

type API struct {
	Version   int
	Endpoints []*Endpoint
}

type REST struct {
	Path   string `json:"path"`
	Method Method `json:"method"`
}

type Endpoint struct {
	Name    string      `json:"name"`
	Handler interface{} `json:"-"`
	REST    *REST       `json:"rest,omitempty"`
	Form    *forms.Form `json:"form"`
}