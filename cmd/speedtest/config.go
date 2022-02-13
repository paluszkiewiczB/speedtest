package main

import (
	"errors"
	"fmt"
	"github.com/gurkankaymak/hocon"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/dummy"
	"github.com/paluszkiewiczB/speedtest/internal/influx"
	"github.com/paluszkiewiczB/speedtest/internal/observe"
	"github.com/paluszkiewiczB/speedtest/internal/ookla"
	"log"
	"strings"
	"time"
)

type prometheusCfg struct {
	enabled        bool
	storageEnabled bool
	cfg            observe.PrometheusConfig
}

func parsePrometheusCfg(config *hocon.Config) prometheusCfg {
	pCfg := config.GetConfig("prometheus")
	enabled := pCfg.GetBoolean("enabled")
	if !enabled {
		return prometheusCfg{enabled: false}
	}

	cfg := observe.PrometheusConfig{
		Endpoint: pCfg.GetString("endpoint"),
		Port:     pCfg.GetInt("port"),
	}

	return prometheusCfg{
		enabled:        true,
		cfg:            cfg,
		storageEnabled: pCfg.GetBoolean("storage"),
	}
}

func createStorage(cfg *hocon.Config) (core.Storage, error) {
	storageType := cfg.GetString("type")
	switch storageType {
	case "INFLUX":
		c, err := parseInfluxStorageCfg(cfg.GetConfig("influxdb"))
		if err != nil {
			return nil, err
		}
		client, err := influx.NewClient(c)
		return client, err
	case "IN-MEMORY":
		return dummy.NewStorage(), nil
	case "TIMEOUT":
		return createTimeoutStorage(cfg)
	case "RETRY":
		return createRetryStorage(cfg)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

func createTimeoutStorage(cfg *hocon.Config) (core.Storage, error) {
	return createInfluxDecorator(cfg, func(delegate influx.Client, cfg *hocon.Config) (influx.Client, error) {
		maxTime := cfg.GetConfig("timeout").GetDuration("time")
		return &influx.TimeOutingClient{Max: maxTime, Delegate: delegate}, nil
	})
}

func createRetryStorage(cfg *hocon.Config) (core.Storage, error) {
	return createInfluxDecorator(cfg, func(delegate influx.Client, cfg *hocon.Config) (influx.Client, error) {
		rCfg := cfg.GetConfig("retry")
		tries := rCfg.GetInt("tries")
		interval := rCfg.GetDuration("interval")
		return influx.Retrying(delegate, influx.RetryCfg{Times: tries, Wait: interval}), nil
	})
}

func createInfluxDecorator(cfg *hocon.Config, f func(influx.Client, *hocon.Config) (influx.Client, error)) (influx.Client, error) {
	delegate, err := createStorage(cfg.GetConfig("client"))
	if err != nil {
		return nil, err
	}
	if d, ok := delegate.(influx.Client); !ok {
		return nil, fmt.Errorf("could not create timeout client for non-influx delegate: %T", delegate)
	} else {
		return f(d, cfg)
	}
}

func parseInfluxStorageCfg(config *hocon.Config) (influx.Cfg, error) {
	url := fmt.Sprintf("%s:%d", config.GetString("host"), config.GetInt("port"))
	points := config.GetConfig("points")
	return influx.Cfg{
		Url:          url,
		Token:        config.GetString("token"),
		Organization: config.GetString("organization"),
		Bucket:       config.GetString("bucket"),
		Points: influx.PointsCfg{
			Measurement: points.GetString("measurement"),
			Tags:        parseTags(points.GetString("tags")),
		}}, nil
}

// parseTags does not support comma and colon inside the tag. Potential improvement - support escaping
func parseTags(tags string) map[string]string {
	out := make(map[string]string)
	for _, kv := range strings.Split(tags, ",") {
		split := strings.Split(kv, ":")
		if len(split) != 2 {
			panic(fmt.Sprintf("cannot parse tag: %v", kv))
		}
		key := split[0]
		value := split[1]

		if old, ok := out[key]; ok {
			log.Printf("overriding tag value for key: %s from %s to %s", key, old, value)
		}

		out[key] = value
	}
	return out
}

type speedTestCfg struct {
	schedulerCfg *schedulerCfg
	clientCfg    *clientCfg
}

func parseSpeedTestCfg(config *hocon.Config) (*speedTestCfg, error) {
	sCfg, err := parseSchedulerCfg(config.GetConfig("scheduler"))
	if err != nil {
		return nil, err
	}

	cCfg, err := parseClientCfg(config.GetConfig("client"))
	if err != nil {
		return nil, err
	}

	return &speedTestCfg{
		schedulerCfg: sCfg,
		clientCfg:    cCfg,
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

type clientCfg struct {
	clientType string
}

func parseClientCfg(cfg *hocon.Config) (*clientCfg, error) {
	return &clientCfg{clientType: cfg.GetString("type")}, nil
}

func createSpeedTester(cfg *clientCfg) (core.SpeedTester, error) {
	switch cfg.clientType {
	case "OOKLA":
		return ookla.NewSpeedTester(), nil
	case "OOKLA_LOGGING":
		return ookla.Logging(ookla.NewSpeedTester()), nil
	case "DUMMY":
		return &dummy.SpeedTester{}, nil
	}

	return nil, fmt.Errorf("unsupported speed tester type: %s", cfg.clientType)
}
