package core

import (
	"context"
	"errors"
	"log"
)

func Boot(ctx context.Context, cfg Config, scheduler Scheduler, tester SpeedTester, storage Storage, errH ErrorHandler) error {
	defer func() {
		err := scheduler.Close()
		if err != nil {
			log.Printf("error when closing scheduler: %v", err)
		}
	}()

	speedC := make(chan Speed)
	testErrC := make(chan error)
	defer close(speedC)
	handleErrors(testErrC, errH)

	err := scheduler.Schedule(ctx, "SpeedTest", cfg.SpeedTestInterval, func() {
		s, err := tester.Test(ctx)
		if err != nil {
			testErrC <- err
			return
		}
		speedC <- s
	})
	if err != nil {
		return errors.New("could not schedule task for speedtest")
	}
	for {
		select {
		case <-ctx.Done():
			log.Println("context cancelled, exiting")
			return nil
		case s := <-speedC:
			log.Printf("speedtest result: %v", s)
			err := storage.Push(ctx, s)
			if err != nil {
				testErrC <- err
			}
		}
	}
}

func handleErrors(c chan error, h ErrorHandler) {
	go func() {
		for err := range c {
			h.Handle(err)
		}
	}()
}
