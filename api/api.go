package api

import (
	"github.com/kiprotect/go-helpers/forms"
)

type EndpointType string

const (
	Retrieve = EndpointType("retrieve")
	Create   = EndpointType("create")
	Replace  = EndpointType("replace")
	Delete   = EndpointType("delete")
	Modify   = EndpointType("modify")
)

type API struct {
	Version   int
	Endpoints []*Endpoint
}

type Endpoint struct {
	Name    string       `json:"name"`
	Type    EndpointType `json:"type"`
	Handler interface{}  `json:"-"`
	Form    *forms.Form  `json:"form"`
}
