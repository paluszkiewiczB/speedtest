package main

import (
	"context"
	"fmt"
	"github.com/gurkankaymak/hocon"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/influx"
	"github.com/paluszkiewiczB/speedtest/internal/inmemory"
	"github.com/paluszkiewiczB/speedtest/internal/schedule"
	"github.com/paluszkiewiczB/speedtest/internal/speedtest"
	"log"
	"time"
)

func main() {
	cfg, err := hocon.ParseResource("reference.conf")
	if err != nil {
		log.Fatalf("Could not parse config: %v\n", err)
		return
	}
	stc, err := parseSpeedTestCfg(cfg.GetConfig("speedtest"))
	if err != nil {
		log.Fatalf("Could not parse speed test cfg: %v\n", err)
	}

	ctx := context.TODO()
	tester := speedtest.NewOnlineSpeedTester()
	storage, err := createStorage(cfg.GetConfig("storage"))
	if err != nil {
		log.Fatalf("Could not create storage: %v\n", err)
	}

	speeds := make(chan core.Speed)
	speedErrors := make(chan error)
	defer close(speeds)
	defer close(speedErrors)
	scheduler := schedule.NewScheduler()
	defer func(scheduler schedule.Scheduler) {
		err := scheduler.Close()
		if err != nil {
			log.Printf("error when closing scheduler: %v", err)
		}
	}(scheduler)
	err = scheduler.Schedule(ctx, "speedtest", stc.schedulerCfg.duration, func() {
		s, err := tester.Test(ctx)
		if err != nil {
			speedErrors <- err
			return
		}
		speeds <- s
	})
	if err != nil {
		log.Fatalf("could not schedule task for speedtest")
	}

	go func() {
		for err := range speedErrors {
			log.Printf("error during speed test: %v", err)
		}
	}()

	for {
		select {
		case s := <-speeds:
			log.Printf("speedtest result: %v", s)
			err := storage.Push(ctx, s)
			if err != nil {
				log.Printf("error when storing speedtest result: %v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func createStorage(config *hocon.Config) (core.Storage, error) {
	storageType := config.GetString("type")
	switch storageType {
	case "influx":
		c, err := parseInfluxStorageCfg(config.GetConfig("influxdb"))
		if err != nil {
			return nil, err
		}
		return influx.NewClient(c)
	case "in-memory":
		return inmemory.NewStorage(), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

func parseInfluxStorageCfg(config *hocon.Config) (influx.Cfg, error) {
	url := fmt.Sprintf("%s:%d", config.GetString("host"), config.GetInt("port"))

	return influx.Cfg{Url: url, Token: config.GetString("token"), Organization: config.GetString("organization"), Bucket: config.GetString("bucket")}, nil
}

type speedTestCfg struct {
	schedulerCfg *schedulerCfg
}

func parseSpeedTestCfg(config *hocon.Config) (*speedTestCfg, error) {
	schedCfg, err := parseSchedulerCfg(config.GetConfig("scheduler"))
	if err != nil {
		return nil, err
	}

	return &speedTestCfg{
		schedulerCfg: schedCfg,
	}, nil
}

type schedulerCfg struct {
	duration time.Duration
}

func parseSchedulerCfg(cfg *hocon.Config) (*schedulerCfg, error) {
	return &schedulerCfg{duration: cfg.GetDuration("duration")}, nil
}
