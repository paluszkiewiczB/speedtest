package influx_test

import (
	"context"
	"errors"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/influx"
	"testing"
)

func TestRetryingClient_TestConnection(t *testing.T) {
	tests := map[string]struct {
		client       *failingClient
		cfg          influx.RetryCfg
		fails, tests int
		error        bool
	}{
		"should test connection 3 times and return error": {
			client: &failingClient{},
			cfg:    influx.RetryCfg{Times: 3, Wait: 0},
			fails:  3,
			tests:  3,
			error:  true,
		},
		"should test connection twice and return nil when delegate returns error only once": {
			client: &failingClient{timeToFail: 1},
			cfg:    influx.RetryCfg{Times: 5, Wait: 0},
			fails:  1,
			tests:  2,
			error:  false,
		},
		"should test connection once and return nil when client never fails": {
			client: &failingClient{timeToFail: -1},
			cfg:    influx.RetryCfg{Times: 10},
			fails:  0,
			tests:  1,
			error:  false,
		},
	}

	ctx := context.Background()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			client := tt.client
			retrying := influx.Retrying(client, tt.cfg)
			err := retrying.Ping(ctx)
			if tt.error && err == nil {
				t.Fatalf("expected error, but it is nil")
			}

			if !tt.error && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.fails != client.failedTimes {
				t.Fatalf("expected client to fail %d times, actual count: %d", tt.fails, client.failedTimes)
			}

			if tt.tests != client.test {
				t.Fatalf("expected client to test connection: %d times, actual count: %d", tt.tests, client.test)
			}

			if client.push != 0 || client.close != 0 {
				t.Fatalf("unexpected invocation of push: %d or close: %d", client.push, client.close)
			}

		})
	}
}

func TestRetryingClient_Push(t *testing.T) {
	tests := map[string]struct {
		client       *failingClient
		cfg          influx.RetryCfg
		fails, tests int
		error        bool
	}{
		"should push 3 times and return error": {
			client: &failingClient{},
			cfg:    influx.RetryCfg{Times: 3, Wait: 0},
			fails:  3,
			tests:  3,
			error:  true,
		},
		"should push twice and return nil when delegate returns error only once": {
			client: &failingClient{timeToFail: 1},
			cfg:    influx.RetryCfg{Times: 5, Wait: 0},
			fails:  1,
			tests:  2,
			error:  false,
		},
		"should push once and return nil when client never fails": {
			client: &failingClient{timeToFail: -1},
			cfg:    influx.RetryCfg{Times: 10},
			fails:  0,
			tests:  1,
			error:  false,
		},
	}

	ctx := context.Background()
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			client := tt.client
			retrying := influx.Retrying(client, tt.cfg)
			err := retrying.Push(ctx, core.InvalidSpeed)
			if tt.error && err == nil {
				t.Fatalf("expected error, but it is nil")
			}

			if !tt.error && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.fails != client.failedTimes {
				t.Fatalf("expected client to fail %d times, actual count: %d", tt.fails, client.failedTimes)
			}

			if tt.tests != client.push {
				t.Fatalf("expected client to push: %d times, actual count: %d", tt.tests, client.test)
			}

			if client.test != 0 || client.close != 0 {
				t.Fatalf("unexpected invocation of push: %d or close: %d", client.push, client.close)
			}

		})
	}
}

func TestRetryingClient_Close(t *testing.T) {
	tests := map[string]struct {
		client       *failingClient
		cfg          influx.RetryCfg
		fails, tests int
		error        bool
	}{
		"should close 3 times and return error": {
			client: &failingClient{},
			cfg:    influx.RetryCfg{Times: 3, Wait: 0},
			fails:  3,
			tests:  3,
			error:  true,
		},
		"should close twice and return nil when delegate returns error only once": {
			client: &failingClient{timeToFail: 1},
			cfg:    influx.RetryCfg{Times: 5, Wait: 0},
			fails:  1,
			tests:  2,
			error:  false,
		},
		"should close once and return nil when client never fails": {
			client: &failingClient{timeToFail: -1},
			cfg:    influx.RetryCfg{Times: 10},
			fails:  0,
			tests:  1,
			error:  false,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			client := tt.client
			retrying := influx.Retrying(client, tt.cfg)
			err := retrying.Close()
			if tt.error && err == nil {
				t.Fatalf("expected error, but it is nil")
			}

			if !tt.error && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.fails != client.failedTimes {
				t.Fatalf("expected client to fail %d times, actual count: %d", tt.fails, client.failedTimes)
			}

			if tt.tests != client.close {
				t.Fatalf("expected client to close: %d times, actual count: %d", tt.tests, client.test)
			}

			if client.push != 0 || client.test != 0 {
				t.Fatalf("unexpected invocation of push: %d or close: %d", client.push, client.close)
			}

		})
	}
}

// when timeToFail is 0, always fails; otherwise counts failures
type failingClient struct {
	timeToFail, failedTimes, push, close, test int
}

func (f *failingClient) Push(_ context.Context, _ core.Speed) error {
	f.push++
	return f.tryToFail()
}

func (f *failingClient) Close() error {
	f.close++
	return f.tryToFail()
}

func (f *failingClient) Ping(_ context.Context) error {
	f.test++
	return f.tryToFail()
}

func (f *failingClient) tryToFail() error {
	if f.timeToFail == 0 {
		return f.fail()
	}

	if f.timeToFail > f.failedTimes {
		return f.fail()
	}

	return nil
}

func (f *failingClient) fail() error {
	f.failedTimes++
	return errors.New("error")
}
