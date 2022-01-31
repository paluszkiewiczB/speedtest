package ookla

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/showwin/speedtest-go/speedtest"
	"log"
	"time"
)

func NewSpeedTester() *SpeedTester {
	actions := make(chan action)
	return &SpeedTester{actions: actions}
}

type SpeedTester struct {
	actions chan action
}

func (t *SpeedTester) Test(ctx context.Context) (core.Speed, error) {
	user, err := speedtest.FetchUserInfo()
	if err != nil {
		log.Println("fetching user info failed")
		return core.InvalidSpeed, err
	}
	serverList, err := speedtest.FetchServerList(user)
	if err != nil {
		log.Println("fetching server list failed")
		return core.InvalidSpeed, err
	}
	targets, err := serverList.FindServer([]int{})
	if err != nil {
		log.Println("finding server failed")
		return core.InvalidSpeed, err
	}
	server := targets[0]
	log.Printf("selected server:  %8.2fkm %s (%s)\n", server.Distance, server.Name, server.Country)

	measurementTime := time.Now()
	err = t.test(ctx, "ping test", func() error {
		return server.PingTest()
	})
	if err != nil {
		return core.InvalidSpeed, err
	}

	err = t.test(ctx, "download test", func() error {
		return server.DownloadTestContext(ctx, false)
	})
	if err != nil {
		log.Println("download test failed")
		return core.InvalidSpeed, err
	}

	err = t.test(ctx, "upload test", func() error {
		return server.UploadTest(false)
	})
	if err != nil {
		fmt.Println("upload test failed")
		return core.InvalidSpeed, err
	}

	return core.Speed{
		Download:  server.DLSpeed,
		Upload:    server.ULSpeed,
		Ping:      server.Latency,
		Timestamp: measurementTime,
	}, nil
}

func (t *SpeedTester) test(ctx context.Context, name string, test func() error) error {
	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
			actionFinished := make(chan struct{}, 1)
			t.actions <- newAction(name, actionFinished)
			err := test()
			if err != nil {
				log.Printf("%s failed\n", name)
				return err
			}
			actionFinished <- struct{}{}
			return err
		}
	}
}

func newAction(name string, finished <-chan struct{}) action {
	return action{
		id:       uuid.New(),
		name:     name,
		start:    time.Now(),
		finished: finished,
	}
}

type action struct {
	id       uuid.UUID
	name     string
	start    time.Time
	finished <-chan struct{}
}
