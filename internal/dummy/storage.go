package dummy

import (
	"context"
	"github.com/paluszkiewiczB/speedtest/internal/core"
)

func NewStorage() *Storage {
	s := make([]core.Speed, 0)
	return &Storage{s}
}

type Storage struct {
	s []core.Speed
}

func (s *Storage) Push(_ context.Context, speed core.Speed) error {
	s.s = append(s.s, speed)
	return nil
}

func (s *Storage) Close() error {
	return nil
}

func (s *Storage) GetAll() []core.Speed {
	return s.s
}
