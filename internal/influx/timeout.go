package influx

import (
	"context"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/resilience"
	"time"
)

type TimeOutingClient struct {
	Delegate Client
	Max      time.Duration
}

func (c *TimeOutingClient) Push(ctx context.Context, speed core.Speed) error {
	return c.TimeOut(ctx, func(ctx context.Context) error {
		return c.Delegate.Push(ctx, speed)
	})
}

func (c *TimeOutingClient) Close() error {
	return c.TimeOut(context.Background(), func(ctx context.Context) error {
		return c.Delegate.Close()
	})
}

func (c *TimeOutingClient) Ping(ctx context.Context) error {
	return c.TimeOut(ctx, func(ctx context.Context) error {
		return c.Delegate.Ping(ctx)
	})
}

func (c *TimeOutingClient) TimeOut(ctx context.Context, f func(context.Context) error) error {
	return resilience.Timeout(ctx, c.Max, f)
}
