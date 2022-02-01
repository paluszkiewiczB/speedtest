package core_test

import (
	"context"
	"errors"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/schedule"
	"testing"
	"time"
)

func Test_Boot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	storage := newInMemoryStorage()

	handler := &countingHandler{}
	err := core.Boot(ctx, core.Config{SpeedTestInterval: 1 * time.Hour}, schedule.NewScheduler(), &dummyTester{}, storage, handler)
	if err != nil {
		t.Fatal(err)
	}
	if handler.i != 0 {
		t.Fatalf("expected 0 errors count, actual: %d", handler.i)
	}

	if len(storage.s) != 1 {
		t.Fatalf("expected to have 1 stored speedtest result")
	}
}

func Test_BootFailed(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	storage := newInMemoryStorage()

	handler := &countingHandler{}
	err := core.Boot(ctx, core.Config{SpeedTestInterval: 1 * time.Hour}, &failingScheduler{}, &dummyTester{}, storage, handler)
	if err == nil {
		t.Fatal("expected err, but it's nil")
	}
	if handler.i != 0 {
		t.Fatalf("expected 0 errors count, actual: %d", handler.i)
	}

	if len(storage.s) != 0 {
		t.Fatalf("expected to have 0 stored speedtest result")
	}
}

func Test_BootWithTestErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	storage := newInMemoryStorage()

	handler := &countingHandler{}
	err := core.Boot(ctx, core.Config{SpeedTestInterval: 1 * time.Hour}, schedule.NewScheduler(), &failingTester{}, storage, handler)
	if err != nil {
		t.Fatal(err)
	}

	if handler.i == 0 {
		t.Fatal("expected errors count but 0 handled")
	}

	if len(storage.s) != 0 {
		t.Fatalf("expected to have 0 stored speedtest result")
	}
}

func Test_BootWithSaveErrors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	storage := &failingStorage{}

	handler := &countingHandler{}
	err := core.Boot(ctx, core.Config{SpeedTestInterval: 1 * time.Hour}, schedule.NewScheduler(), &dummyTester{}, storage, handler)
	if err != nil {
		t.Fatal(err)
	}

	if handler.i == 0 {
		t.Fatal("expected errors count but 0 handled")
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

type countingHandler struct {
	i int
}

func (h *countingHandler) Handle(error) {
	h.i++
}

type failingScheduler struct{}

func (f failingScheduler) Schedule(_ context.Context, _ string, _ time.Duration, _ func()) error {
	return errors.New("scheduler error")
}

func (f *failingScheduler) Cancel(_ string) error {
	return errors.New("scheduler error")
}

func (f *failingScheduler) Close() error {
	return errors.New("scheduler error")
}

type failingTester struct{}

func (f *failingTester) Test(_ context.Context) (core.Speed, error) {
	return core.InvalidSpeed, errors.New("test failed")
}

type failingStorage struct{}

func (f *failingStorage) Push(_ context.Context, _ core.Speed) error {
	return errors.New("storage failed")
}

func (f *failingStorage) Close() error {
	return errors.New("storage failed")
}
