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
