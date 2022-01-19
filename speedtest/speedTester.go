package speedtest

import (
	"context"
	"fmt"
	"github.com/showwin/speedtest-go/speedtest"
	"log"
	"speedtestMonitoring/core"
	"time"
)

type SpeedTester interface {
	Test(context.Context) (core.Speed, error)
}

type DummySpeedTester struct{}

func (t DummySpeedTester) Test(ctx context.Context) (*core.Speed, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		done := make(chan *core.Speed, 1)

		go func() {
			time.Sleep(2 * time.Second)
			done <- &core.Speed{Download: 1.0, Upload: 2.0, Ping: time.Second}
		}()

		return <-done, nil
	}
}

func NewOnlineSpeedTester() SpeedTester {
	return &onlineSpeedTester{}
}

type onlineSpeedTester struct{}

// Test FIXME rewrite without pointless channels
func (t *onlineSpeedTester) Test(ctx context.Context) (core.Speed, error) {
	user, _ := speedtest.FetchUserInfo()

	serverList, _ := speedtest.FetchServerList(user)
	targets, _ := serverList.FindServer([]int{})
	server := targets[0]

	fmt.Printf("User:\n%v\n\nSelected server:\n%v\n\n", user, server)
	measurementTime := time.Now()
	errC := make(chan error)
	go func() {
		errC <- test(ctx, server, "ping", func(s *speedtest.Server) error {
			return s.PingTest()
		}, func(s *speedtest.Server) string {
			return s.Latency.String()
		})
	}()

	if err := <-errC; err != nil {
		fmt.Printf("Ping test failed with error: %v\n", err)
		return core.InvalidSpeed, err
	}

	go func() {
		errC <- test(ctx, server, "download", func(s *speedtest.Server) error {
			return s.DownloadTest(false)
		}, func(s *speedtest.Server) string {
			return fmt.Sprintf("%f Mb/s", s.DLSpeed)
		})

	}()
	if err := <-errC; err != nil {
		fmt.Printf("Download test failed with error: %v\n", err)
		return core.InvalidSpeed, err
	}

	go func() {
		errC <- test(ctx, server, "upload", func(s *speedtest.Server) error {
			return s.UploadTest(false)
		}, func(s *speedtest.Server) string {
			return fmt.Sprintf("%f Mb/s", s.ULSpeed)
		})
	}()

	if err := <-errC; err != nil {
		fmt.Printf("Upload test failed with error: %v\n", err)
		return core.InvalidSpeed, err
	}

	return core.Speed{
		Download:  server.DLSpeed,
		Upload:    server.ULSpeed,
		Ping:      server.Latency,
		Timestamp: measurementTime,
	}, nil
}

func test(ctx context.Context, server *speedtest.Server, name string, test func(*speedtest.Server) error, result func(*speedtest.Server) string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		log.Printf("Testing %s...\n", name)
		finished := make(chan bool, 1)
		errC := make(chan error, 1)
		dots(name, finished)
		go func() {
			err := test(server)
			if err != nil {
				errC <- err
			}
			finished <- true
			errC <- nil
		}()

		if err := <-errC; err != nil {
			log.Printf("Error occured while testing %s\n", name)
			return err
		}

		log.Printf("Test %s result: %s", name, result(server))
		return nil
	}
}

func dots(name string, stop chan bool) {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-timer.C:
		log.Printf("Testing %s...\n", name)
	case <-stop:
		return
	}
}
