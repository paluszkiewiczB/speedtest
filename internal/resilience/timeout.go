package resilience

import (
	"context"
	"time"
)

func Timeout(ctx context.Context, max time.Duration, f func(context.Context) error) error {
	timeout, cancel := context.WithTimeout(ctx, max)
	defer cancel()

	errC := make(chan error)

	go func() {
		errC <- f(timeout)
	}()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-timeout.Done():
			return context.DeadlineExceeded
		case err := <-errC:
			return err
		}
	}
}
