package errHandlers_test

import (
	"bytes"
	"errors"
	"github.com/paluszkiewiczB/speedtest/internal/errHandlers"
	"log"
	"os"
	"strings"
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

func Test_NewPrintln(t *testing.T) {
	errMsg := "test error message"
	output := captureOutput(func() {
		handler := errHandlers.NewPrintln()
		handler.Handle(errors.New(errMsg))
	})

	if !strings.Contains(output, "test error message") {
		t.Fatalf("expected log containing: '%s', got: '%s'", errMsg, output)
	}
}

type countingLogger struct {
	i int
}

func (c *countingLogger) Log(_ ...interface{}) {
	c.i++
}

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}
