package schedule

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type Scheduler interface {
	Schedule(ctx context.Context, key string, d time.Duration, task func()) error
	Cancel(key string) error
	Close() error
}

func NewScheduler() Scheduler {
	m := make(map[string]*scheduledTask)
	mu := &sync.Mutex{}
	return &scheduler{m, mu}
}

type scheduler struct {
	cancels map[string]*scheduledTask
	mu      *sync.Mutex
}

func (s *scheduler) Schedule(ctx context.Context, key string, d time.Duration, task func()) error {
	cancel := make(chan struct{})
	ticker := time.NewTicker(d)
	scheduled := &scheduledTask{cancel, ticker}
	err := s.putCancel(key, scheduled)
	if err != nil {
		return err
	}
	go func() {
		log.Printf("Starting task: %s", key)
		task()

		for {
			select {
			case <-ctx.Done():
				return
			case <-cancel:
				return
			case <-ticker.C:
				log.Printf("Starting task: %s", key)
				task()
			}
		}
	}()
	log.Printf("Scheduled task %s\n", key)
	return nil
}

// Cancel cancels scheduled task
// Calling Cancel second time is no-op
func (s *scheduler) Cancel(key string) error {
	s.mu.Lock()
	c, ok := s.cancels[key]
	s.cancels[key] = nil
	s.mu.Unlock()
	if ok {
		c.cancel <- struct{}{}
		c.ticker.Stop()
	}
	return nil
}

// Close closes scheduler - scheduling new task after Close was called causes panic
// Calling Close second time is no-op
func (s *scheduler) Close() error {
	s.mu.Lock()
	for _, t := range s.cancels {
		t.cancel <- struct{}{}
		t.ticker.Stop()
	}
	s.cancels = nil
	s.mu.Unlock()
	return nil
}

func (s *scheduler) putCancel(key string, task *scheduledTask) error {
	s.mu.Lock()
	_, exists := s.cancels[key]
	if exists {
		return fmt.Errorf("task with key: %s is already scheduled", key)
	}
	s.cancels[key] = task
	s.mu.Unlock()
	return nil
}

type scheduledTask struct {
	cancel chan<- struct{}
	ticker *time.Ticker
}
