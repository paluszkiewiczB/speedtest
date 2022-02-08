package schedule_test

import (
	"context"
	"github.com/paluszkiewiczB/speedtest/internal/schedule"
	"testing"
	"time"
)

// FIXME flaky tests :(

func Test_CancellingContextShouldStopTask(t *testing.T) {
	scheduler := schedule.NewScheduler()
	timeout, cancel := context.WithCancel(context.Background())
	stopCounting := time.NewTimer(3 * time.Millisecond)

	task := &limitTask{t: &testTask{}, limit: 2, cancel: cancel}
	err := scheduler.Schedule(timeout, "test", 1*time.Millisecond, task.inc)
	if err != nil {
		t.Fatal(err)
	}
	<-stopCounting.C
	if 2 != task.t.counter {
		t.Fatalf("expected counter to be 2, actual: %d", task.t.counter)
	}
}

func TestScheduler_Cancel(t *testing.T) {
	scheduler := schedule.NewScheduler()
	cancelContext := time.NewTimer(1100 * time.Microsecond)
	stopCounting := time.NewTimer(3 * time.Millisecond)
	go func() {
		<-cancelContext.C
		_ = scheduler.Cancel("test")
		_ = scheduler.Cancel("test")
		_ = scheduler.Cancel("test")
	}()
	task := &testTask{}
	err := scheduler.Schedule(context.Background(), "test", 1*time.Millisecond, task.inc)
	if err != nil {
		t.Fatal(err)
	}

	<-stopCounting.C
	if task.counter != 2 {
		t.Fatalf("expected counter to be 2, actual %d", task.counter)
	}
}

func TestScheduler_ScheduleAfterClose(t *testing.T) {
	scheduler := schedule.NewScheduler()
	err := scheduler.Close()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	err = scheduler.Schedule(ctx, "should fal", 1*time.Hour, func() {
		panic("should never happen")
	})

	if err == nil {
		t.Fatalf("expected error when scheduling task, but it is nil")
	}
}

func Test_SchedulerDuplicate(t *testing.T) {
	scheduler := schedule.NewScheduler()
	task := &testTask{}
	err := scheduler.Schedule(context.Background(), "test", 1*time.Millisecond, task.inc)
	if err != nil {
		t.Fatal(err)
	}
	err = scheduler.Schedule(context.Background(), "test", 1*time.Millisecond, task.inc)
	if err == nil {
		t.Fatal("expected error, but it's nil")
	}
}

type testTask struct {
	counter int
}

func (t *testTask) inc() {
	t.counter++
}

type limitTask struct {
	t      *testTask
	limit  int
	cancel func()
}

func (t *limitTask) inc() {
	t.t.inc()
	if t.t.counter == t.limit {
		t.cancel()
	}
}
