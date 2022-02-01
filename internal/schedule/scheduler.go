package schedule

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

func NewScheduler() *Scheduler {
	c := make(map[string]*scheduledTask)
	mu := &sync.Mutex{}
	return &Scheduler{cancels: c, mu: mu}
}

type Scheduler struct {
	cancels map[string]*scheduledTask
	mu      *sync.Mutex
}

func (s *Scheduler) Schedule(ctx context.Context, key string, d time.Duration, task func()) error {
	if s.cancels == nil {
		return errors.New("scheduler was not properly initialized or is closed")
	}

	taskCtx, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(d)
	scheduled := &scheduledTask{cancel: cancel, ticker: ticker}
	err := s.putCancel(key, scheduled)
	if err != nil {
		ticker.Stop()
		return err
	}
	go func() {
		log.Printf("starting task: %s", key)
		task()

		for {
			select {
			case <-taskCtx.Done():
				log.Printf("context for task: %s was cancelled", key)
				err := s.Cancel(key)
				if err != nil {
					log.Printf("error when removing task")
				}
				return
			case <-ticker.C:
				log.Printf("starting task: %s", key)
				task()
			}
		}
	}()
	log.Printf("scheduled task %s", key)
	return nil
}

// Cancel cancels scheduled task
// Calling Cancel second time is no-op
func (s *Scheduler) Cancel(key string) error {
	s.mu.Lock()
	c, ok := s.cancels[key]
	delete(s.cancels, key)
	s.mu.Unlock()
	if ok {
		c.cancel()
		c.ticker.Stop()
	}
	log.Printf("task cancelled: %s", key)
	return nil
}

// Close closes Scheduler - scheduling new task after Close was called causes panic
// Calling Close second time is no-op
func (s *Scheduler) Close() error {
	s.mu.Lock()
	for _, c := range s.cancels {
		c.ticker.Stop()
		c.cancel()
	}
	s.cancels = nil
	s.mu.Unlock()
	return nil
}

func (s *Scheduler) putCancel(key string, task *scheduledTask) error {
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
	cancel func()
	ticker *time.Ticker
}
