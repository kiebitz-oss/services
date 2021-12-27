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

package metrics

import (
	"context"
	"github.com/kiebitz-oss/services"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sync"
	"time"
)

type PrometheusMetricsServer struct {
	server  *http.Server
	mutex   sync.Mutex
	err     error
	running bool
}

type PrometheusMetricsServerSettings struct {
	BindAddress string
}

func MakePrometheusMetricsServer(bindAddress string) (*PrometheusMetricsServer, error) {

	p := &PrometheusMetricsServer{
		server: &http.Server{Addr: bindAddress, Handler: promhttp.Handler()},
	}

	return p, nil
}

func (p *PrometheusMetricsServer) Start() error {

	go func() {

		if err := p.server.ListenAndServe(); err != http.ErrServerClosed {

			// something went wrong, we log and store the error...

			services.Log.Error(err)

			p.mutex.Lock()
			p.err = err
			p.running = false
			p.mutex.Unlock()
		} else {
			// the server shut down gracefully...
			p.mutex.Lock()
			p.running = false
			p.err = nil
			p.mutex.Unlock()
		}
	}()

	time.Sleep(time.Millisecond * 100)

	p.mutex.Lock()
	running := p.running
	p.mutex.Unlock()

	if !running {
		return p.err
	}

	return nil
}

func (p *PrometheusMetricsServer) Stop() error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.server.Shutdown(ctx)
}
