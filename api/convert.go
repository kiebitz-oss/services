package api

import (
	"github.com/kiebitz-oss/services/jsonrpc"
)

func (c *API) ToJSONRPC() (map[string]*jsonrpc.Method, error) {
	methods := map[string]*jsonrpc.Method{}
	for _, endpoint := range c.Endpoints {
		methods[endpoint.Name] = &jsonrpc.Method{
			Form:    endpoint.Form,
			Handler: endpoint.Handler,
		}
	}
	return methods, nil
}
