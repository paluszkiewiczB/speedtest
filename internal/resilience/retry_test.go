package resilience_test

import (
	"context"
	"errors"
	"github.com/paluszkiewiczB/speedtest/internal/resilience"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	t.Run("should stop retrying after first success or when max attempts reached", func(t *testing.T) {
		tests := map[string]struct {
			cfg     cfg
			fails   int
			tries   int
			wantErr bool
		}{
			"should return nil when task does not fail": {
				cfg:     cfg{times: 10},
				fails:   0,
				tries:   1,
				wantErr: false,
			},
			"should return nil when task fails once and 3 tries are allowed": {
				cfg:     cfg{times: 3},
				fails:   1,
				tries:   2,
				wantErr: false,
			},
			"should return error when task twice once and 2 tries are allowed": {
				cfg:     cfg{times: 2},
				fails:   2,
				tries:   2,
				wantErr: true,
			},
		}

		ctx := context.Background()
		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				task := failingTask{timeToFail: tt.fails}
				if err := resilience.Retry(ctx, tt.cfg, task.tryToFail); (err != nil) != tt.wantErr {
					t.Fatalf("Retry() error = %v, wantErr %v", err, tt.wantErr)
				}
				if task.tries != tt.tries {
					t.Fatalf("expected: %d tries, actual: %d", tt.tries, task.tries)
				}
				if task.fails != tt.fails {
					t.Fatalf("expected: %d fails, actual: %d", tt.fails, task.fails)
				}
			})
		}
	})

	t.Run("should fail after first try when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := resilience.Retry(ctx, cfg{times: 10}, (&failingTask{timeToFail: 9999}).tryToFail)
		if err != context.Canceled {
			t.Fatalf("expected context.Canceled, actual error is: %s", err)
		}
	})

	t.Run("first try should not be be stopped by cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := resilience.Retry(ctx, cfg{times: 1}, (&failingTask{timeToFail: 9999}).tryToFail)
		if err == nil {
			t.Fatal("expected error, but it's nil")
		}
	})
}

type cfg struct {
	times int
	wait  time.Duration
}

func (c cfg) MaxAttempts() int { return c.times }

func (c cfg) Interval() time.Duration { return c.wait }

type failingTask struct {
	timeToFail, fails, tries int
}

func (f *failingTask) tryToFail(_ context.Context) error {
	if f.timeToFail > f.fails {
		return f.fail()
	}

	f.tries++
	return nil
}

func (f *failingTask) fail() error {
	f.fails++
	f.tries++
	return errors.New("error")
}
