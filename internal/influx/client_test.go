package influx_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-cmp/cmp"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/paluszkiewiczB/speedtest/internal/core"
	"github.com/paluszkiewiczB/speedtest/internal/influx"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	username   = "username"
	password   = "influxdb123"
	org        = "testOrganization"
	bucket     = "testBucket"
	token      = "adminToken"
	mappedPort = "8086/tcp"
)

var influxEnvs = map[string]string{
	"DOCKER_INFLUXDB_INIT_USERNAME":    username,
	"DOCKER_INFLUXDB_INIT_PASSWORD":    password,
	"DOCKER_INFLUXDB_INIT_ORG":         org,
	"DOCKER_INFLUXDB_INIT_BUCKET":      bucket,
	"DOCKER_INFLUXDB_INIT_ADMIN_TOKEN": token,
	"DOCKER_INFLUXDB_INIT_MODE":        "setup",
}

func TestInflux_Push(t *testing.T) {
	ctx := context.Background()
	container, err := prepareContainer(ctx)
	defer terminate(container, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err != nil {
		t.Fatal(err)
	}
	ports, err := container.Ports(ctx)
	if err != nil {
		t.Fatal(err)
	}
	mP := ports[mappedPort][0]
	url := fmt.Sprintf("http://%s:%s", mP.HostIP, strings.Split(mP.HostPort, "/")[0])
	fmt.Printf("InfluxDB url: %s\n", url)
	client, err := influx.NewClient(influx.Cfg{
		Url:          url,
		Token:        token,
		Organization: org,
		Bucket:       bucket,
	})
	if err != nil {
		t.Fatal(err)
	}
	speed := core.Speed{
		Download:  10.1,
		Upload:    2.51,
		Ping:      14 * time.Second,
		Timestamp: time.Now(),
	}
	err = client.Push(ctx, speed)
	if err != nil {
		t.Fatal(err)
	}
	err = client.Close()
	if err != nil {
		t.Error(err)
	}

	read, err := readSpeed(ctx, url)
	if err != nil {
		t.Fatal(err)
	}

	if cmp.Diff(speed, read) != "" {
		t.Fatalf("read speed differs from written one. expected: %v, actual: %v", speed, read)
	}
}

func prepareContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{},
		Image:          "influxdb:2.1-alpine",
		ExposedPorts:   []string{mappedPort},
		Env:            influxEnvs,
		WaitingFor:     wait.ForLog("Starting log_id="),
	}
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	return c, nil
}

func terminate(c testcontainers.Container, ctx context.Context) {
	err := c.Terminate(ctx)
	if err != nil {
		fmt.Printf("error terminating influxdb container: %v\n", err)
	}
}

func readSpeed(ctx context.Context, url string) (core.Speed, error) {
	query := fmt.Sprintf(
		`	from(bucket:"%s")
					|> range(start: -1h)
					|> filter(fn: (r) =>r._measurement == "speedtest")
					|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")`, bucket)
	result, err := influxdb2.NewClient(url, token).QueryAPI(org).Query(ctx, query)
	if err != nil {
		return core.InvalidSpeed, err
	}

	if !result.Next() {
		return core.InvalidSpeed, errors.New("speed not found")
	}

	r := result.Record()
	return core.Speed{
		Download:  toFloat(r.ValueByKey("download")),
		Upload:    toFloat(r.ValueByKey("upload")),
		Ping:      toDuration(r.ValueByKey("ping")),
		Timestamp: r.Time(),
	}, nil
}

func toFloat(f interface{}) float64 {
	if r, ok := f.(float64); ok {
		return r
	}
	if s, ok := f.(string); ok {
		float, err := strconv.ParseFloat(s, 64)
		if err != nil {
			panic(err)
		}
		return float
	}
	panic("not supported type")
}

func toDuration(d interface{}) time.Duration {
	if s, ok := d.(string); ok {
		duration, err := time.ParseDuration(s)
		if err != nil {
			panic(err)
		}
		return duration
	}

	panic("not supported type")
}
