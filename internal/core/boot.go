package core

import (
	"context"
	"log"
)

func Boot(ctx context.Context, cfg Config, scheduler Scheduler, tester SpeedTester, storage Storage) {
	defer func() {
		err := scheduler.Close()
		if err != nil {
			log.Printf("error when closing scheduler: %v", err)
		}
	}()

	speedC := make(chan Speed)
	testErrC := make(chan error)
	defer close(speedC)
	defer close(testErrC)
	logErrors(testErrC)

	err := scheduler.Schedule(ctx, "SpeedTest", cfg.SpeedTestInterval, func() {
		s, err := tester.Test(ctx)
		if err != nil {
			testErrC <- err
			return
		}
		speedC <- s
	})
	if err != nil {
		log.Fatalf("could not schedule task for speedtest")
	}
	for {
		select {
		case <-ctx.Done():
			log.Println("context cancelled, exiting")
			return
		case s := <-speedC:
			log.Printf("speedtest result: %v", s)
			err := storage.Push(ctx, s)
			if err != nil {
				log.Printf("error when storing speedtest result: %v", err)
			}
		}
	}
}

func logErrors(c chan error) {
	go func() {
		for err := range c {
			log.Printf("speedtest error: %v\n", err)
		}
	}()
}
