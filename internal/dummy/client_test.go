package dummy_test

import (
	"context"
	"github.com/paluszkiewiczB/speedtest/internal/dummy"
	"testing"
)

func TestSpeedTester_Test(t *testing.T) {
	tester := dummy.SpeedTester{}
	_, err := tester.Test(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
