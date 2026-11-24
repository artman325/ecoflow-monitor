package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/artman325/ecoflow-monitor/internal/config"
	"github.com/artman325/ecoflow-monitor/internal/ecoflow"
)

func main() {

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

	devices, err := c.DeviceList()
	handleError(err, cfg)

	device, err := ecoflow.SelectDevice(devices, cfg.SN)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if errors.Is(err, ecoflow.ErrMultipleDevices) {
			fmt.Fprintln(os.Stderr, "hint: set EF_SN in .env to one of the SNs above")
		}
		os.Exit(1)
	}

	deviceStatus, err := c.DeviceQuota(device.SN)
	handleError(err, cfg)

	fmt.Println(devices)
	fmt.Println(*deviceStatus)

}

func handleError(err error, cfg *config.Config) {
	if err == nil {
		return
	}
	if cfg.SN == "" {

	}
	var apiErr *ecoflow.APIError
	if errors.As(err, &apiErr) {
		if apiErr.IsCredentialsError() {
			fmt.Fprintln(os.Stderr, "credentials:", err)
		} else {
			fmt.Fprintln(os.Stderr, "api:", err)
		}

	} else {
		fmt.Fprintln(os.Stderr, "connection:", err)

	}
	os.Exit(1)

}
