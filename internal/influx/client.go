package influx

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"log"
)

type Client interface {
	core.Storage
	Ping(ctx context.Context) error
}

func NewClient(cfg Cfg) (*AsyncClient, error) {
	c := influxdb2.NewClient(cfg.Url, cfg.Token)
	w := c.WriteAPI(cfg.Organization, cfg.Bucket)
	return &AsyncClient{writer: w, client: c, points: cfg.Points}, nil
}

type Cfg struct {
	Url, Token, Organization, Bucket string
	Points                           PointsCfg
}

type PointsCfg struct {
	Measurement string
	Tags        map[string]string
}

type AsyncClient struct {
	client influxdb2.Client
	writer api.WriteAPI
	points PointsCfg
}

func (c *AsyncClient) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx)
	return err
}

func (c *AsyncClient) Push(ctx context.Context, speed core.Speed) error {
	eC := make(chan error, 1)
	go func() {
		fields := map[string]interface{}{
			"download": speed.Download,
			"upload":   speed.Upload,
			"ping":     speed.Ping.Milliseconds(),
		}
		p := influxdb2.NewPoint(c.points.Measurement, c.points.Tags, fields, speed.Timestamp)
		log.Printf("Writing point at: %v", speed.Timestamp)
		c.writer.WritePoint(p)
		eC <- nil
	}()

	select {
	case <-ctx.Done():
		return context.Canceled
	case err := <-eC:
		return err
	}

}

func (c *AsyncClient) Close() error {
	c.writer.Flush()
	c.client.Close()
	return nil
}
