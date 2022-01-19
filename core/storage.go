package core

import "context"

type Storage interface {
	Push(ctx context.Context, speed Speed) error
	Close() error
}
