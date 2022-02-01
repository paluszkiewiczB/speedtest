package core_test

import (
	"context"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/schedule"
	"testing"
	"time"
)

func Test_Boot(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	storage := newInMemoryStorage()

	go func() {
		timer := time.NewTimer(1 * time.Millisecond)
		<-timer.C
		cancel()
	}()

	core.Boot(ctx, core.Config{SpeedTestInterval: 1 * time.Hour}, schedule.NewScheduler(), &dummyTester{}, storage)

	if len(storage.s) != 1 {
		t.Fatalf("expected to have 1 stored speedtest result")
	}
}

type dummyTester struct {
}

func (d *dummyTester) Test(_ context.Context) (core.Speed, error) {
	return core.Speed{Download: 1.0, Upload: 1.0, Ping: 1 * time.Millisecond, Timestamp: time.Now()}, nil
}

func newInMemoryStorage() *inMemoryStorage {
	s := make([]core.Speed, 0)
	return &inMemoryStorage{s: s}
}

type inMemoryStorage struct {
	s []core.Speed
}

func (i *inMemoryStorage) Push(_ context.Context, speed core.Speed) error {
	i.s = append(i.s, speed)
	return nil
}

func (i *inMemoryStorage) Close() error {
	return nil
}
