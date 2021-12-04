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

package servers

import (
	"fmt"
	"github.com/kiebitz-oss/services"
	"github.com/kiebitz-oss/services/api"
	"github.com/kiebitz-oss/services/http"
	"github.com/kiebitz-oss/services/jsonrpc"
	"github.com/kiebitz-oss/services/rest"
)

type Server struct {
	httpServer    *http.HTTPServer
	restServer    *rest.RESTServer
	jsonRPCServer *jsonrpc.JSONRPCServer
}

func MakeServer(name string, httpSettings *services.HTTPServerSettings, jsonRPCSettings *services.JSONRPCServerSettings, restSettings *services.RESTServerSettings, api *api.API) (*Server, error) {

	server := &Server{}

	httpServer, err := http.MakeHTTPServer(httpSettings, nil, name)

	if err != nil {
		return nil, err
	}

	server.httpServer = httpServer

	var serverDefined = false

	if jsonRPCSettings != nil {
		serverDefined = true
		if jsonRPCHandler, err := api.ToJSONRPC(); err != nil {
			return nil, err
		} else if jsonRPCServer, err := jsonrpc.MakeJSONRPCServer(jsonRPCSettings, jsonRPCHandler, name, httpServer); err != nil {
			return nil, err
		} else {
			server.jsonRPCServer = jsonRPCServer
		}
	}

	if restSettings != nil {
		serverDefined = true
		if restHandler, err := api.ToREST(); err != nil {
			return nil, err
		} else if restServer, err := rest.MakeRESTServer(restSettings, restHandler, name, httpServer); err != nil {
			return nil, err
		} else {
			server.restServer = restServer
		}
	}

	if !serverDefined {
		return nil, fmt.Errorf("you need to enable at least the JSONRPC or REST servers...")
	}

	return server, nil

}

func (c *Server) Start() error {
	// we start the JSONRPC server first to avoid passing HTTP requests to it before it is initialized
	if c.jsonRPCServer != nil {
		if err := c.jsonRPCServer.Start(); err != nil {
			return err
		}
	}
	if c.restServer != nil {
		if err := c.restServer.Start(); err != nil {
			return err
		}
	}
	if err := c.httpServer.Start(); err != nil {
		return err
	}
	return nil
}

func (c *Server) Stop() error {
	// we stop the HTTP server first to avoid the JSONRPC server receiving requests when it is already stopped
	if err := c.httpServer.Stop(); err != nil {
		return err
	}
	if c.restServer != nil {
		if err := c.restServer.Stop(); err != nil {
			return err
		}
	}
	if c.jsonRPCServer != nil {
		if err := c.jsonRPCServer.Stop(); err != nil {
			return err
		}
	}
	return nil
}
