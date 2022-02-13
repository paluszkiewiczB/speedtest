package dummy_test

import (
	"context"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/dummy"
	"testing"
)

func TestStorage(t *testing.T) {

	t.Run("should store a speed", func(t *testing.T) {
		storage := dummy.NewStorage()
		err := storage.Push(context.Background(), core.InvalidSpeed)
		if err != nil {
			t.Fatal(err)
		}
		all := storage.GetAll()
		if len(all) != 1 {
			t.Fatalf("expected one element, actual: %d", len(all))
		}
		if all[0] != core.InvalidSpeed {
			t.Fatalf("expected speed: %v, actual: %v", core.InvalidSpeed, all[0])
		}
	})
}
