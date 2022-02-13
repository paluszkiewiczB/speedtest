package main

import (
	"context"
	"github.com/gurkankaymak/hocon"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/errHandlers"
	"github.com/paluszkiewiczB/speedtest/internal/influx"
	"github.com/paluszkiewiczB/speedtest/internal/observe"
	"github.com/paluszkiewiczB/speedtest/internal/schedule"
	"log"
	"os"
	"os/signal"
)

func main() {
	cfg, err := hocon.ParseResource("reference.conf")
	if err != nil {
		log.Fatalf("could not parse config: %v\n", err)
		return
	}
	stc, err := parseSpeedTestCfg(cfg.GetConfig("speedtest"))
	if err != nil {
		log.Fatalf("could not parse speed test cfg: %v\n", err)
	}

	tester, err := createSpeedTester(stc.clientCfg)
	if err != nil {
		log.Fatalf("could not create speed tester: %v", err)
	}

	storage, err := createStorage(cfg.GetConfig("storage"))
	if err != nil {
		log.Fatalf("could not create storage: %v\n", err)
	}

	scheduler := schedule.NewScheduler()

	ctx, cancelFunc := context.WithCancel(context.Background())
	shutdownC := make(chan os.Signal, 1)
	signal.Notify(shutdownC, os.Interrupt)
	go func() {
		<-shutdownC
		cancelFunc()
	}()

	promCfg := parsePrometheusCfg(cfg)
	if promCfg.enabled {
		observe.ExposePrometheus(ctx, promCfg.cfg)
		if promCfg.storageEnabled {
			storage = observe.Storage(storage)
		}
	}

	if i, ok := storage.(influx.Client); ok {
		err = i.Ping(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}

	handler := errHandlers.NewPrintln()
	bootCfg := core.Config{SpeedTestInterval: stc.schedulerCfg.duration}
	err = core.Boot(ctx, bootCfg, scheduler, tester, storage, handler)
	if err != nil {
		log.Fatal(err)
	}
}
