package observe_test

import (
	"context"
	"fmt"
	"github.com/paluszkiewiczB/speedtest/internal/observe"
	"log"
	"net/http"
	"testing"
)

func TestExposePrometheus(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("should return 200 when Prometheus endpoint is exposed", func(t *testing.T) {
		cfg, url := prepareCfg()
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(func() {
			cancel()
		})
		observe.ExposePrometheus(ctx, cfg)
		get, err := http.Get(url)
		if get.StatusCode != 200 {
			log.Fatalf("expected status 200, actual: %d", get.StatusCode)
		}
		if err != nil {
			log.Fatal(err)
		}
	})

	t.Run("should stop Prometheus endpoint when context gets cancelled", func(t *testing.T) {
		cfg, url := prepareCfg()
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(func() {
			cancel()
		})
		observe.ExposePrometheus(ctx, cfg)
		get, err := http.Get(url)
		if get.StatusCode != 200 {
			t.Fatalf("expected status code 200, actual: %d", get.StatusCode)
		}
		if err != nil {
			t.Fatal(err)
		}
		cancel()
		_, err = http.Get(url)
		if err == nil {
			t.Fatal("expected error, but it is nil")
		}
	})
}

// prepareCfg returns PrometheusConfig and url to endpoint
func prepareCfg() (observe.PrometheusConfig, string) {
	port, err := getFreePort()
	if err != nil {
		panic(err)
	}
	endpoint := fmt.Sprintf("/test%d", port)
	cfg := observe.PrometheusConfig{
		Endpoint: endpoint,
		Port:     port,
	}
	url := fmt.Sprintf("http://localhost:%d/%s", port, endpoint)
	return cfg, url
}
