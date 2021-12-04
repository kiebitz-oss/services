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
	"github.com/kiebitz-oss/services/http"
	"github.com/kiebitz-oss/services/jsonrpc"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"time"
)

type Handler func(*Context) *Response

type RESTServer struct {
	metricsPrefix string
	httpDurations *prometheus.HistogramVec
	settings      *services.RESTServerSettings
	server        *http.HTTPServer
	ownServer     bool
	handler       Handler
}

func (s *RESTServer) REST(handler Handler) http.Handler {

	return func(c *http.Context) {

		startTime := time.Now()

		context := &Context{
			HTTP: c,
		}

		response := handler(context)

		if response == nil {
			response = context.Nil().(*Response)
		}

		c.JSON(response.StatusCode, response.Data)

		elapsedTime := time.Since(startTime)
		codeString := strconv.Itoa(response.StatusCode)
		s.httpDurations.WithLabelValues(c.Request.URL.Path, codeString).Observe(elapsedTime.Seconds())
	}
}

func MakeRESTServer(settings *services.RESTServerSettings, handler Handler, metricsPrefix string, httpServer *http.HTTPServer) (*RESTServer, error) {

	server := &RESTServer{
		settings:      settings,
		metricsPrefix: metricsPrefix,
	}

	routeGroups := []*http.RouteGroup{
		{
			// these handlers will be executed for all routes in the group
			Handlers: []http.Handler{
				jsonrpc.Cors(settings.Cors, false),
			},
			Routes: []*http.Route{
				{
					Pattern: "^.*$",
					Handlers: []http.Handler{
						server.REST(handler),
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

func (s *RESTServer) Start() error {

	s.httpDurations = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    fmt.Sprintf("%s_%s", s.metricsPrefix, "rest_durations_seconds"),
			Help:    "REST latency distributions",
			Buckets: []float64{0, 0.1, 0.2, 0.5, 1, 2, 5, 10},
		},
		[]string{"method", "code"},
	)

	if err := prometheus.Register(s.httpDurations); err != nil {
		return fmt.Errorf("error registering collector for RESET server (%s): %v", s.metricsPrefix, err)
	}

	if !s.ownServer {
		return nil
	}

	return s.server.Start()
}

func (s *RESTServer) Stop() error {
	prometheus.Unregister(s.httpDurations)

	if !s.ownServer {
		return nil
	}

	return s.server.Stop()
}
