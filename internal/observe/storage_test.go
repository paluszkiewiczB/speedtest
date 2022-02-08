package observe_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/observe"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	success = core.Speed{Download: 1.0, Upload: 1.0, Ping: 1, Timestamp: time.Unix(1, 1)}
	failure = core.Speed{Timestamp: time.Unix(0, 0)}
)

func TestMetricsStorage_Push(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	successCount, failureCount := 7, 12
	speeds := prepareSpeeds(successCount, failureCount)

	s := failingStorage{}
	storage := observe.Storage(s)
	ctx := context.Background()
	port, err := getFreePort()
	if err != nil {
		panic(err)
	}
	cfg := observe.PrometheusConfig{
		Endpoint: "/testMetrics",
		Port:     port,
	}
	withPrometheusEndpoint(ctx, cfg, func() {
		for _, s := range speeds {
			_ = storage.Push(ctx, s)
		}

		sCount, fCount := checkMetrics(cfg)

		if successCount != sCount {
			t.Errorf("Expected: %d successes, got: %d", successCount, sCount)
		}

		if failureCount != fCount {
			t.Errorf("Expected: %d failures, got: %d", failureCount, fCount)
		}
	})
}

func TestMetricsStorage_Close(t *testing.T) {
	t.Run("should close delegate storage", func(t *testing.T) {
		delegate := &closeCountingStorage{}
		storage := observe.Storage(delegate)
		err := storage.Close()
		if err != nil {
			t.Fatal(err)
		}
		if 1 != delegate.counter {
			t.Errorf("expected delegate to be closed once, actual count: %d", delegate.counter)
		}
	})

	t.Run("should return error if delegate returns error on Close", func(t *testing.T) {
		storage := observe.Storage(failingStorage{})
		err := storage.Close()
		if err == nil {
			t.Fatalf("expected error when closing storage, but it is nil")
		}
	})
}

func prepareSpeeds(successCount int, failureCount int) []core.Speed {
	speeds := make([]core.Speed, 0)
	for i := 0; i < successCount; i++ {
		speeds = append(speeds, success)
	}
	for i := 0; i < failureCount; i++ {
		speeds = append(speeds, failure)
	}
	return speeds
}

func withPrometheusEndpoint(c context.Context, cfg observe.PrometheusConfig, f func()) {
	ctx, cancel := context.WithCancel(c)
	defer cancel()
	observe.ExposePrometheus(ctx, cfg)
	f()
}

func checkMetrics(config observe.PrometheusConfig) (int, int) {
	response, err := http.Get(fmt.Sprintf("http://localhost:%d%s", config.Port, config.Endpoint))
	if err != nil {
		panic(err)
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	bodyS := string(body)
	s, f := -1, -1
	lines := strings.Split(bodyS, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, observe.SuccessfulPushesCounterName) {
			s = parseInt(line[len(observe.SuccessfulPushesCounterName)+1:])
		} else if strings.HasPrefix(line, observe.FailedPushesCounterName) {
			f = parseInt(line[len(observe.FailedPushesCounterName)+1:])
		}
	}

	return s, f
}

func parseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return i
}

type failingStorage struct {
}

func (f failingStorage) Push(_ context.Context, speed core.Speed) error {
	if speed == success {
		return nil
	}
	return errors.New("error")
}

func (f failingStorage) Close() error {
	return errors.New("error")
}

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

type closeCountingStorage struct {
	counter int
}

func (c *closeCountingStorage) Push(ctx context.Context, speed core.Speed) error {
	panic("implement me")
}

func (c *closeCountingStorage) Close() error {
	c.counter++
	return nil
}
