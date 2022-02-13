package influx_test

import (
	"context"
	"errors"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/influx"
	"testing"
	"time"
)

var (
	clientErr = errors.New("sleeping error")
)

func TestTimeOutingClient_Close(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}

	t.Run("should time out and return error", func(t *testing.T) {
		tests := map[string]struct {
			client    influx.Client
			timeoutMs int
		}{
			"when exec time is 10 ms and timeout is 1ms": {
				client:    &sleepingClient{ms: 10},
				timeoutMs: 1,
			},
			"when exec time is 10ms and timeout is 0ms": {
				client:    &sleepingClient{ms: 10},
				timeoutMs: 0,
			},
		}

		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				timeout := time.Duration(int64(tt.timeoutMs) * time.Millisecond.Nanoseconds())
				client := influx.TimeOutingClient{Delegate: tt.client, Max: timeout}
				err := client.Close()
				if context.DeadlineExceeded != err {
					t.Fatalf("expected deadline exceeded, actual: %s", err)
				}
			})
		}
	})

	t.Run("should not time out and return nil", func(t *testing.T) {
		tests := map[string]struct {
			client    influx.Client
			timeoutMs int
		}{
			"when exec time is 0 ms and timeout is 10ms": {
				client:    &sleepingClient{ms: 0},
				timeoutMs: 10,
			},
			"when exec time is 1ms and timeout is 10ms": {
				client:    &sleepingClient{ms: 1},
				timeoutMs: 10,
			},
		}

		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				timeout := time.Duration(int64(tt.timeoutMs) * time.Millisecond.Nanoseconds())
				client := influx.TimeOutingClient{Delegate: tt.client, Max: timeout}
				err := client.Close()
				if err != nil {
					t.Fatalf("unexpected err: %s", err)
				}
			})
		}
	})

	t.Run("should not time out and return error", func(t *testing.T) {
		tests := map[string]struct {
			client    influx.Client
			timeoutMs int
		}{
			"when exec time is 0 ms and timeout is 10ms": {
				client:    &sleepingClient{ms: 0, failure: true},
				timeoutMs: 10,
			},
			"when exec time is 1ms and timeout is 10ms": {
				client:    &sleepingClient{ms: 1, failure: true},
				timeoutMs: 10,
			},
		}

		for name, tt := range tests {
			t.Run(name, func(t *testing.T) {
				timeout := time.Duration(int64(tt.timeoutMs) * time.Millisecond.Nanoseconds())
				client := influx.TimeOutingClient{Delegate: tt.client, Max: timeout}
				err := client.Close()
				if err != clientErr {
					t.Fatalf("expected client error, actual: %s", err)
				}
			})
		}
	})
}

func TestTimeOutingClient_Ping(t *testing.T) {
}

func TestTimeOutingClient_Push(t *testing.T) {

}

type sleepingClient struct {
	ms      int
	failure bool
}

func (c *sleepingClient) Push(_ context.Context, _ core.Speed) error {
	return c.sleep()
}

func (c *sleepingClient) Close() error {
	return c.sleep()
}

func (c *sleepingClient) Ping(_ context.Context) error {
	return c.sleep()
}

func (c *sleepingClient) sleep() error {
	time.Sleep(time.Duration(int64(c.ms) * time.Millisecond.Nanoseconds()))
	if c.failure {
		return clientErr
	}
	return nil
}
