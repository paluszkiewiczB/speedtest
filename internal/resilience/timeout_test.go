package resilience_test

import (
	"context"
	"errors"
	"github.com/paluszkiewiczB/speedtest/internal/resilience"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	t.Run("should finish before or max 1ms after timeout", func(t *testing.T) {
		if testing.Short() {
			t.SkipNow()
		}

		tests := map[string]struct {
			sleepMs, maxMs             int64
			wantTimeout, wantErr, fail bool
		}{
			"should timeout within 1ms": {
				sleepMs:     2,
				maxMs:       1,
				wantTimeout: true,
				wantErr:     true,
			},
			"should timeout within 5ms": {
				sleepMs:     6,
				maxMs:       5,
				wantTimeout: true,
				wantErr:     true,
			},
			"should not timeout and not fail": {
				sleepMs:     0,
				maxMs:       1,
				wantTimeout: false,
				wantErr:     false,
			},
			"should not timeout and fail": {
				sleepMs:     0,
				fail:        true,
				maxMs:       1,
				wantTimeout: false,
				wantErr:     true,
			},
		}

		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {

				maxTime := time.Duration(tt.maxMs * time.Millisecond.Nanoseconds())
				execT, err := timer(func(ctx context.Context) error {
					sleepT := time.Duration(tt.sleepMs * time.Millisecond.Nanoseconds())
					task := sleepMs(sleepT, tt.fail)
					return resilience.Timeout(context.Background(), maxTime, task)
				})

				if err == nil && tt.wantErr {
					t.Fatalf("expected error, but it is nil")
				}

				if err != nil && !tt.wantErr {
					t.Fatalf("unexpected error: %s", err)
				}

				if !tt.wantTimeout && execT > maxTime {
					t.Fatalf("exec time exceed max: %s, was: %s", maxTime, execT)
				}
			})
		}
	})

	t.Run("should stop immediately when context is cancelled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := timer(func(context.Context) error {
			return resilience.Timeout(ctx, time.Hour, func(ctx context.Context) error {
				sleepMs(10, false)
				return errors.New("should not wait that long")
			})
		})

		if err == nil {
			t.Fatal(err)
		}
	})
}

func sleepMs(ms time.Duration, fail bool) func(context.Context) error {
	return func(ctx context.Context) error {
		time.Sleep(ms * time.Millisecond)
		if fail {
			return errors.New("fail")
		}
		return nil
	}
}

func timer(f func(context.Context) error) (time.Duration, error) {
	start := time.Now()
	err := f(context.Background())
	took := time.Since(start)
	return took, err
}
