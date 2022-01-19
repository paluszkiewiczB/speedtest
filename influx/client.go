package influx

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"speedtestMonitoring/core"
	"time"
)

func NewClient(cfg Cfg) (core.Storage, error) {
	c := influxdb2.NewClient(cfg.Url, cfg.Token)
	err := testConnection(c)
	if err != nil {
		return nil, err
	}
	w := c.WriteAPI(cfg.Organization, cfg.Bucket)
	return &influx{writer: w, client: c, organization: cfg.Organization, bucket: cfg.Bucket}, nil
}

type Cfg struct {
	Url, Token, Organization, Bucket string
}

type influx struct {
	client               influxdb2.Client
	writer               api.WriteAPI
	organization, bucket string
}

func testConnection(c influxdb2.Client) error {
	timeout, cancelTimeout := context.WithTimeout(context.Background(), time.Second*15)
	defer cancelTimeout()
	ticker := time.NewTicker(time.Second * 1)
	eC := make(chan error, 1)
	go func() {
		ping := func() {
			fmt.Printf("Pinging...\n")
			ok, err := c.Ping(timeout)
			if ok && err == nil {
				println("Connection obtained")
				eC <- nil
			} else {
				fmt.Printf("error pinging: %v\n", err)
			}
		}

		ping()
		for {
			select {
			case <-ticker.C:
				ping()
			case <-timeout.Done():
				eC <- timeoutErr("pinging influxdb when to test connection")
			}
		}

	}()
	return <-eC
}

// Push TODO configurable measurements and tags?
func (db *influx) Push(ctx context.Context, speed core.Speed) error {
	eC := make(chan error, 1)
	go func() {
		measurement := "speedtest"
		tags := map[string]string{
			"connection": "wifi",
			"client":     "raspberry-pi-zero-w",
		}
		fields := map[string]interface{}{
			"download": speed.Download,
			"upload":   speed.Upload,
			"ping":     speed.Ping,
		}
		p := influxdb2.NewPoint(measurement, tags, fields, speed.Timestamp)
		fmt.Printf("Writing point at: %v", speed.Timestamp)
		db.writer.WritePoint(p)
		eC <- nil
	}()

	select {
	case <-ctx.Done():
		return timeoutErr("writing point to influxdb")
	case err := <-eC:
		return err
	}

}

func (db *influx) Close() error {
	db.writer.Flush()
	db.client.Close()
	return nil
}

func timeoutErr(task string) error {
	return fmt.Errorf("timed out executing task: %s", task)
}
