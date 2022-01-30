package inmemory

import (
	"context"
	"github.com/paluszkiewiczB/speedtest/internal/core"
)

func NewStorage() core.Storage {
	s := make([]core.Speed, 0)
	return &storage{s}
}

type storage struct {
	s []core.Speed
}

func (s *storage) Push(_ context.Context, speed core.Speed) error {
	s.s = append(s.s, speed)
	return nil
}

func (s *storage) Close() error {
	return nil
}
