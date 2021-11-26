package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type PrometheusMetricsServer struct {
	server *http.Server
}

type PrometheusMetricsServerSettings struct {
	BindAddress string
}

func MakePrometheusMetricsServer(bindAddress string) (*PrometheusMetricsServer, error) {

	p := &PrometheusMetricsServer{
		server: &http.Server{Addr: bindAddress, Handler: promhttp.Handler()},
	}

	go func() {
		if err := p.server.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic("Could not start metrics server")
			}
		}
	}()

	return p, nil
}

func (p *PrometheusMetricsServer) Stop() error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.server.Shutdown(ctx)
}
