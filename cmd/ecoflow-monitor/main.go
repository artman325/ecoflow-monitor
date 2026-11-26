package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/artman325/ecoflow-monitor/internal/config"
	"github.com/artman325/ecoflow-monitor/internal/ecoflow"
)

type refreshResult struct {
	device    *ecoflow.DeviceInfo
	telemetry *ecoflow.Telemetry
	err       error
}
type refreshState struct {
	lastTelemetry *ecoflow.Telemetry
	lastUpdate    time.Time
	failures      int
}

// minutes
var backoffSteps = []int{0, 1, 2, 3, 5}

func (st *refreshState) update(rr *refreshResult, now time.Time) {
	if rr.err != nil {
		st.failures++
	} else {
		st.failures = 0
	}
	if rr.err == nil && rr.telemetry != nil {
		st.lastTelemetry = rr.telemetry
	}
	if rr.err == nil && rr.telemetry != nil && rr.device.IsOnline() {
		st.lastUpdate = now
	}

}
func (st *refreshState) needTelemetry(online bool) bool {
	return st.lastTelemetry == nil || online
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	e, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	envPath := flag.String("env", filepath.Join(filepath.Dir(e), ".env"), "path to env file")
	flag.Parse()
	cfg, err := config.Load(*envPath)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	c := ecoflow.NewClient(cfg.Host, cfg.AccessKey, cfg.SecretKey)

	run(ctx, c, cfg)

	fmt.Println("Программа корректно завершена")
}

func run(ctx context.Context, c *ecoflow.Client, cfg *config.Config) {
	var st refreshState

	for {

		res := refresh(c, cfg.SN, &st)

		if res.err != nil {
			fmt.Fprintln(os.Stderr, res.err)
			fmt.Fprintln(os.Stderr, errorCategory(res.err))

			if errors.Is(res.err, ecoflow.ErrMultipleDevices) {
				fmt.Fprintln(os.Stderr, "hint: set EF_SN in .env to one of the SNs above")
			}
		}

		st.update(&res, time.Now())

		// send to registry. for now just print
		fmt.Println(res.device)
		fmt.Println(st.lastTelemetry)

		pause := calculatePause(cfg.Interval, st.failures)

		select {
		case <-ctx.Done():
			// Контекст отменен — выходим из функции и прерываем бесконечный цикл
			fmt.Println("Остановка цикла:", ctx.Err())
			return
		case <-time.After(pause):
			fmt.Println("Работаю...")
		}
	}
}

// calculatePause returns the delay before the next poll.
// After consecutive failures it backs off to 1, 2, 3 and then 5 minutes,
// but never polls more often than the configured interval:
// the backoff only exists to stop hammering the API, so if the interval
// is already longer than the current step, the interval is used as is.
func calculatePause(interval time.Duration, failures int) time.Duration {
	if failures < 0 {
		failures = 0
	}
	idx := failures
	if idx >= len(backoffSteps) {
		idx = len(backoffSteps) - 1
	}

	return max(interval, (time.Minute * time.Duration(backoffSteps[idx])))
}

func refresh(c *ecoflow.Client, sn string, st *refreshState) refreshResult {

	var res refreshResult
	var devices []ecoflow.DeviceInfo

	devices, res.err = c.DeviceList()
	if res.err != nil {
		return res
	}

	res.device, res.err = ecoflow.SelectDevice(devices, sn)

	if res.err != nil {
		return res
	}

	if st.needTelemetry(res.device.IsOnline()) {
		res.telemetry, res.err = c.DeviceTelemetry(res.device.SN)
	}

	return res
}

func errorCategory(err error) string {

	if err == nil {
		return ""
	}

	var apiErr *ecoflow.APIError
	if errors.As(err, &apiErr) {
		if apiErr.IsCredentialsError() {
			return "credentials"
		} else {
			return "api"
		}
	} else if ecoflow.IsDeviceError(err) {
		return "device"
	}
	return "connection"

}
