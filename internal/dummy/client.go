package dummy

import (
	"context"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"time"
)

type SpeedTester struct {
}

func (s *SpeedTester) Test(_ context.Context) (core.Speed, error) {
	return core.Speed{
		Download:  10,
		Upload:    10,
		Ping:      10,
		Timestamp: time.Now(),
	}, nil
}
