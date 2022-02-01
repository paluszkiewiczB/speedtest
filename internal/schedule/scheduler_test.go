package schedule_test

import (
	"context"
	"github.com/paluszkiewiczB/speedtest/internal/schedule"
	"testing"
	"time"
)

func Test_CancellingContextShouldStopTask(t *testing.T) {
	scheduler := schedule.NewScheduler()
	timeout, cancel := context.WithCancel(context.Background())
	cancelContext := time.NewTimer(1001 * time.Microsecond)
	stopCounting := time.NewTimer(3 * time.Millisecond)
	go func() {
		<-cancelContext.C
		cancel()
	}()

	task := &testTask{}
	err := scheduler.Schedule(timeout, "test", 1*time.Millisecond, task.inc)
	if err != nil {
		t.Fatal(err)
	}
	<-stopCounting.C
	if 2 != task.counter {
		t.Fatalf("expected counter to be 2, actual: %d", task.counter)
	}
}

func TestScheduler_Cancel(t *testing.T) {
	scheduler := schedule.NewScheduler()
	cancelContext := time.NewTimer(1001 * time.Microsecond)
	stopCounting := time.NewTimer(3 * time.Millisecond)
	go func() {
		<-cancelContext.C
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

type testTask struct {
	counter int
}

func (t *testTask) inc() {
	t.counter++
}
