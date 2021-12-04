package api

import (
	"github.com/kiebitz-oss/services/jsonrpc"
	"github.com/kiebitz-oss/services/rest"
)

func (c *API) ToJSONRPC() (jsonrpc.Handler, error) {
	methods := map[string]*jsonrpc.Method{}
	for _, endpoint := range c.Endpoints {
		methods[endpoint.Name] = &jsonrpc.Method{
			Form:    endpoint.Form,
			Handler: endpoint.Handler,
		}
	}
	return jsonrpc.MethodsHandler(methods)
}

func (c *API) ToREST() (rest.Handler, error) {
	methods := map[string]*rest.Method{}
	for _, endpoint := range c.Endpoints {
		if endpoint.REST == nil {
			continue
		}
		methods[endpoint.Name] = &rest.Method{
			Path:    endpoint.REST.Path,
			Method:  string(endpoint.REST.Method),
			Form:    endpoint.Form,
			Handler: endpoint.Handler,
		}
	}
	return rest.MethodsHandler(methods)
}
