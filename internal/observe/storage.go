package observe

import (
	"context"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	SuccessfulPushesCounterName = "speedtest_successful_storage_pushes"
	FailedPushesCounterName     = "speedtest_failed_storage_pushes"
)

var (
	successfulPushes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "speedtest_successful_storage_pushes",
		Help: "Number of successful pushes of speed measurements to storage",
	})
	failedPushes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "speedtest_failed_storage_pushes",
		Help: "Number of failed pushes of speed measurements to storage",
	})
)

func Storage(delegate core.Storage) *MetricsStorage {
	return &MetricsStorage{
		delegate: delegate,
	}
}

type MetricsStorage struct {
	delegate core.Storage
}

func (s *MetricsStorage) Push(ctx context.Context, speed core.Speed) error {
	err := s.delegate.Push(ctx, speed)
	if err == nil {
		successfulPushes.Inc()
	} else {
		failedPushes.Inc()
	}
	return err
}

func (s *MetricsStorage) Close() error {
	return s.delegate.Close()
}
