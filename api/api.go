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
