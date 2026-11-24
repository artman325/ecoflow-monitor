package ecoflow

import (
	"errors"
	"fmt"
	"strings"
)

var ErrDeviceWithSNNotFound = errors.New("device: SN not found")
var ErrDeviceNotFound = errors.New("device: not found")
var ErrMultipleDevices = errors.New("device: multiple devices found")

func SelectDevice(devices []DeviceInfo, sn string) (*DeviceInfo, error) {

	if len(devices) == 0 {
		return nil, fmt.Errorf("%w: available device list is empty", ErrDeviceNotFound)
	}

	if sn == "" && len(devices) == 1 {
		return &devices[0], nil
	}

	var foundSNList = []string{}

	for _, d := range devices {
		foundSNList = append(foundSNList, d.SN)
		if sn != "" && d.SN == sn {
			return &d, nil
		}
	}

	// not found

	if sn == "" {
		return nil, fmt.Errorf("%w: specify SN from [%s]", ErrMultipleDevices, strings.Join(foundSNList, ", "))
	}
	return nil, fmt.Errorf("%w: %q, available: [%s]", ErrDeviceWithSNNotFound, sn, strings.Join(foundSNList, ", "))

}
