package influx

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"log"
	"time"
)

func NewClient(cfg Cfg) (*Client, error) {
	c := influxdb2.NewClient(cfg.Url, cfg.Token)
	err := testConnection(c)
	if err != nil {
		return nil, err
	}
	w := c.WriteAPI(cfg.Organization, cfg.Bucket)
	return &Client{writer: w, client: c, organization: cfg.Organization, bucket: cfg.Bucket}, nil
}

type Cfg struct {
	Url, Token, Organization, Bucket string
}

type Client struct {
	client               influxdb2.Client
	writer               api.WriteAPI
	organization, bucket string
}

func testConnection(c influxdb2.Client) error {
	log.Println("testing connection to influxdb")
	timeout, cancelTimeout := context.WithTimeout(context.Background(), time.Second*15)
	defer cancelTimeout()
	ticker := time.NewTicker(time.Second * 1)
	eC := make(chan error, 1)
	go func() {
		ping := func() {
			log.Println("pinging influxdb...")
			ok, err := c.Ping(timeout)
			if ok && err == nil {
				log.Println("influxdb connection obtained")
				eC <- nil
			} else {
				log.Printf("error pinging: %v\n", err)
			}
		}

		ping()
		for {
			select {
			case <-ticker.C:
				ping()
			case <-timeout.Done():
				eC <- timeoutErr("pinging influxdb to test connection")
			}
		}

	}()
	return <-eC
}

// Push TODO configurable measurements and tags?
func (c *Client) Push(ctx context.Context, speed core.Speed) error {
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
		log.Printf("Writing point at: %v", speed.Timestamp)
		c.writer.WritePoint(p)
		eC <- nil
	}()

	select {
	case <-ctx.Done():
		return timeoutErr("writing point to influxdb")
	case err := <-eC:
		return err
	}

}

func (c *Client) Close() error {
	c.writer.Flush()
	c.client.Close()
	return nil
}

func timeoutErr(task string) error {
	return fmt.Errorf("timed out executing task: %s", task)
}
