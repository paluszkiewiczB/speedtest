package resilience

import (
	"context"
	"time"
)

func Retry(ctx context.Context, cfg RetryCfg, f func(context.Context) error) error {
	err := f(ctx)
	if err == nil {
		return nil
	}

	if cfg.MaxAttempts() < 2 {
		return err
	}

	timer := time.NewTimer(cfg.Interval())
	for i := 1; i < cfg.MaxAttempts(); i++ {
		select {
		case <-timer.C:
			err = f(ctx)
			if err == nil {
				return nil
			}
			timer.Reset(cfg.Interval())
		case <-ctx.Done():
			return context.Canceled
		}
	}

	return err
}

type RetryCfg interface {
	MaxAttempts() int
	Interval() time.Duration
}
