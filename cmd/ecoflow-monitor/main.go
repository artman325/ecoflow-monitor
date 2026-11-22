package main

import (
	"encoding/json"
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
	resp, err := c.DeviceList()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var deviceInfo []ecoflow.DeviceInfo
	if err := json.Unmarshal(resp, &deviceInfo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	resp, err = c.DeviceQuota(deviceInfo[0].SN)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var deviceStatus ecoflow.DeviceStatus
	if err := json.Unmarshal(resp, &deviceStatus); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(deviceInfo)
	fmt.Println(deviceStatus)

}
