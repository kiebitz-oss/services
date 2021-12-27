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
	"github.com/kiebitz-oss/services/http"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type Handler func(*Context) *Response

type JSONRPCServer struct {
	metricsPrefix string
	httpDurations *prometheus.HistogramVec
	settings      *services.JSONRPCServerSettings
	server        *http.HTTPServer
	ownServer     bool
	handler       Handler
}

func (s *JSONRPCServer) JSONRPC(handler Handler) http.Handler {

	return func(c *http.Context) {

		startTime := time.Now()

		// the request data has been validated by the 'ExtractJSONRequest' handler
		request := c.Get("request").(*Request)

		context := &Context{
			Request: request,
		}

		response := handler(context)

		if response == nil {
			response = context.Nil().(*Response)
		}

		// people will forget this so we add it here in that case
		if response.JSONRPC == "" {
			response.JSONRPC = "2.0"
		}

		code := 200

		// if there was an error we return a 400 status instead of 200
		if response.Error != nil {
			code = 400
		}

		c.JSON(code, response)

		elapsedTime := time.Since(startTime)
		codeString := strconv.Itoa(code)

		s.httpDurations.WithLabelValues(request.Method, codeString).Observe(elapsedTime.Seconds())
	}
}

func MakeJSONRPCServer(settings *services.JSONRPCServerSettings, handler Handler, metricsPrefix string, httpServer *http.HTTPServer) (*JSONRPCServer, error) {

	server := &JSONRPCServer{
		settings:      settings,
		metricsPrefix: metricsPrefix,
	}

	routeGroups := []*http.RouteGroup{
		{
			// these handlers will be executed for all routes in the group
			Handlers: []http.Handler{
				Cors(settings.Cors, false),
			},
			Routes: []*http.Route{
				{
					Pattern: "^/jsonrpc$",
					Handlers: []http.Handler{
						ExtractJSONRequest,
						server.JSONRPC(handler),
					},
				},
			},
		},
	}

	if httpServer == nil {

		server.ownServer = true

		if settings.HTTP == nil {
			return nil, fmt.Errorf("HTTP settings missing")
		}

		var err error

		if httpServer, err = http.MakeHTTPServer(settings.HTTP, routeGroups, fmt.Sprintf("%s_http", metricsPrefix)); err != nil {
			return nil, err
		}
	} else if err := httpServer.AddRouteGroups(routeGroups); err != nil {
		return nil, err
	}

	server.server = httpServer

	return server, nil

}

func (s *JSONRPCServer) Start() error {

	s.httpDurations = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    fmt.Sprintf("%s_%s", s.metricsPrefix, "rpc_durations_seconds"),
			Help:    "RPC latency distributions",
			Buckets: []float64{0, 0.1, 0.2, 0.5, 1, 2, 5, 10},
		},
		[]string{"method", "code"},
	)

	if err := prometheus.Register(s.httpDurations); err != nil {
		return fmt.Errorf("error registering collector for jsonRPC server (%s): %v", s.metricsPrefix, err)
	}

	if !s.ownServer {
		return nil
	}

	return s.server.Start()
}

func (s *JSONRPCServer) Stop() error {
	prometheus.Unregister(s.httpDurations)

	if !s.ownServer {
		return nil
	}

	return s.server.Stop()
}
