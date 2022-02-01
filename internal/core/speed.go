package core

import (
	"context"
	"time"
)

var InvalidSpeed = Speed{-1, -1, -1, time.Unix(0, 0)}

type Speed struct {
	Download  float64
	Upload    float64
	Ping      time.Duration
	Timestamp time.Time
}

type Storage interface {
	Push(ctx context.Context, speed Speed) error
	Close() error
}

type SpeedTester interface {
	Test(context.Context) (Speed, error)
}

type Scheduler interface {
	Schedule(ctx context.Context, key string, d time.Duration, task func()) error
	Cancel(key string) error
	Close() error
}

type ErrorHandler interface {
	Handle(err error)
}

type Config struct {
	SpeedTestInterval time.Duration
}
