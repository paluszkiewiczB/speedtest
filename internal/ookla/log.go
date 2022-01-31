package ookla

import (
	"github.com/google/uuid"
	"log"
	"time"
)

func Logging(tester *SpeedTester) *SpeedTester {
	logActions(tester.actions)
	return tester
}

func logActions(actions chan action) {
	go func() {
		for a := range actions {
			go func(a action) {
				log.Printf("[%s] speed test started action: %s at: %v\n", a.id, a.name, a.start)
				dots(a.id, a.name, a.finished)
			}(a)
		}
	}()
}

func dots(id uuid.UUID, name string, stop <-chan struct{}) {
	timer := time.NewTicker(5 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-timer.C:
			log.Printf("[%s] %s...\n", id, name)
		case <-stop:
			return
		}
	}
}
