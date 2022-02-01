package errHandlers_test

import (
	"errors"
	"github.com/paluszkiewiczB/speedtest/internal/errHandlers"
	"testing"
)

func Test_New(t *testing.T) {
	logger := &countingLogger{}
	handler := errHandlers.New(logger)
	handler.Handle(errors.New("test"))

	if logger.i != 1 {
		t.Fatalf("Expected 1 logged err, got: %d", logger.i)
	}
}

func Test_NewFuncDelegating(t *testing.T) {
	logger := &countingLogger{}
	handler := errHandlers.NewFuncDelegating(logger.Log)
	handler.Handle(errors.New("test"))

	if logger.i != 1 {
		t.Fatalf("Expected 1 logged err, got: %d", logger.i)
	}
}

type countingLogger struct {
	i int
}

func (c *countingLogger) Log(_ ...interface{}) {
	c.i++
}
