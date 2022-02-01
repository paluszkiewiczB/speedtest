package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gurkankaymak/hocon"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/errHandlers"
	"github.com/paluszkiewiczB/speedtest/internal/influx"
	"github.com/paluszkiewiczB/speedtest/internal/inmemory"
	"github.com/paluszkiewiczB/speedtest/internal/ookla"
	"github.com/paluszkiewiczB/speedtest/internal/schedule"
	"log"
	"os"
	"os/signal"
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

	tester := ookla.Logging(ookla.NewSpeedTester())
	log.Printf("speed tester created, creating storage")
	storage, err := createStorage(cfg.GetConfig("storage"))
	if err != nil {
		log.Fatalf("Could not create storage: %v\n", err)
	}

	scheduler := schedule.NewScheduler()

	ctx, cancelFunc := context.WithCancel(context.Background())
	shutdownC := make(chan os.Signal, 1)
	signal.Notify(shutdownC, os.Interrupt)
	go func() {
		<-shutdownC
		cancelFunc()
	}()

	bootCfg := core.Config{SpeedTestInterval: stc.schedulerCfg.duration}

	handler := errHandlers.NewPrintln()
	err = core.Boot(ctx, bootCfg, scheduler, tester, storage, handler)
	if err != nil {
		log.Fatal(err)
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
	sCfg, err := parseSchedulerCfg(config.GetConfig("scheduler"))
	if err != nil {
		return nil, err
	}

	return &speedTestCfg{
		schedulerCfg: sCfg,
	}, nil
}

type schedulerCfg struct {
	duration time.Duration
}

func parseSchedulerCfg(cfg *hocon.Config) (*schedulerCfg, error) {
	duration, err := parseDuration(cfg)
	if err != nil {
		return nil, err
	}
	return &schedulerCfg{duration: duration}, nil
}

func parseDuration(cfg *hocon.Config) (time.Duration, error) {
	duration := cfg.Get("duration")
	switch d := duration.(type) {
	case hocon.String:
		cheat, err := hocon.ParseString(fmt.Sprintf("duration: %s", d))
		if err != nil {
			return -1, err
		}
		return cheat.GetDuration("duration"), nil
	case hocon.Duration:
		return cfg.GetDuration("duration"), nil
	}

	return -1, errors.New("unsupported value type of duration")
}
