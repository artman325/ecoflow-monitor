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
	device *ecoflow.DeviceInfo
	status *ecoflow.Telemetry
	err    error
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

	pause := 5 * time.Second
	for {
		res := refresh(c, cfg)
		if res.err != nil {
			fmt.Fprintln(os.Stderr, res.err)
			fmt.Fprintln(os.Stderr, errorCategory(res.err))

			if errors.Is(res.err, ecoflow.ErrMultipleDevices) {
				fmt.Fprintln(os.Stderr, "hint: set EF_SN in .env to one of the SNs above")
			}
		} else {

			fmt.Println(res.device)
			fmt.Println(res.status)
		}

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

func refresh(c *ecoflow.Client, cfg *config.Config) refreshResult {

	var res refreshResult
	var devices []ecoflow.DeviceInfo

	devices, res.err = c.DeviceList()
	if res.err != nil {
		return res
	}

	res.device, res.err = ecoflow.SelectDevice(devices, cfg.SN)

	if res.err != nil {
		return res
	}

	res.status, res.err = c.DeviceTelemetry(res.device.SN)

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
