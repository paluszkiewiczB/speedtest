package influx

import (
	"context"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/resilience"
	"time"
)

func Retrying(client Client, cfg RetryCfg) *RetryingClient {
	return &RetryingClient{
		delegate: client,
		cfg:      cfg,
	}
}

type RetryingClient struct {
	delegate Client
	cfg      RetryCfg
}

type RetryCfg struct {
	Times int
	Wait  time.Duration
}

func (c *RetryCfg) MaxAttempts() int {
	return c.Times
}

func (c *RetryCfg) Interval() time.Duration {
	return c.Wait
}

func (c *RetryingClient) Push(ctx context.Context, speed core.Speed) error {
	return c.retry(ctx, func(ctx context.Context) error {
		return c.delegate.Push(ctx, speed)
	})
}

func (c *RetryingClient) Close() error {
	return c.retry(context.Background(), func(ctx context.Context) error {
		return c.delegate.Close()
	})
}

func (c *RetryingClient) Ping(ctx context.Context) error {
	return c.retry(ctx, func(ctx context.Context) error {
		return c.delegate.Ping(ctx)
	})
}

func (c *RetryingClient) retry(ctx context.Context, f func(context.Context) error) error {
	return resilience.Retry(ctx, &c.cfg, f)
}
