package observe

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func ExposePrometheus(ctx context.Context, cfg PrometheusConfig) {
	http.Handle(cfg.Endpoint, promhttp.Handler())
	server := http.Server{Addr: fmt.Sprintf(":%d", cfg.Port), Handler: nil}
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Printf("could not start prometheus server: %s\n", err)
		}
	}()

	go func() {
		<-ctx.Done()
		err := server.Close()
		if err != nil {
			log.Printf("could not stop prometheus http server")
		}
	}()
}

type PrometheusConfig struct {
	Endpoint string
	Port     int
}
