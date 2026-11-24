package ecoflow

import (
	"errors"
	"testing"
)

func TestSelectDevice(t *testing.T) {

	type inputStruct struct {
		devices []DeviceInfo
		sn      string
	}

	tests := []struct {
		name       string
		input      inputStruct
		wantErr    error
		wantOutput DeviceInfo
	}{
		{
			name: "The SN was provided and exists",
			input: inputStruct{
				devices: []DeviceInfo{
					{SN: "SN-001"},
					{SN: "SN-002"},
					{SN: "SN-003"},
				},
				sn: "SN-001",
			},
			wantOutput: DeviceInfo{SN: "SN-001"},
		},
		{
			name: "The SN was provided and does not exist",
			input: inputStruct{
				devices: []DeviceInfo{
					{SN: "SN-001"},
					{SN: "SN-002"},
					{SN: "SN-003"},
				},
				sn: "SN-004",
			},
			wantErr: ErrDeviceWithSNNotFound,
		},
		{
			name: "The SN wasn't provided and device list is empty",
			input: inputStruct{
				devices: []DeviceInfo{},
				sn:      "",
			},
			wantErr: ErrDeviceNotFound,
		},
		{
			name: "The SN wasn't provided and device list length is 1",
			input: inputStruct{
				devices: []DeviceInfo{
					{SN: "SN-001"},
				},
				sn: "",
			},
			wantErr:    nil,
			wantOutput: DeviceInfo{SN: "SN-001"},
		},

		{
			name: "The SN wasn't provided and device list length is greater than 1",
			input: inputStruct{
				devices: []DeviceInfo{
					{SN: "SN-001"},
					{SN: "SN-002"},
					{SN: "SN-003"},
				},
				sn: "",
			},
			wantErr: ErrMultipleDevices,
		},
		{
			name: "The SN was provided, device list length is 1, but it is not in the list",
			input: inputStruct{
				devices: []DeviceInfo{
					{SN: "SN-001"},
				},
				sn: "SN-004",
			},
			wantErr: ErrDeviceWithSNNotFound,
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {

			d, err := SelectDevice(tc.input.devices, tc.input.sn)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}

			if tc.wantErr == nil {
				if *d != tc.wantOutput {
					t.Fatalf("got = %v, expected %v", *d, tc.wantOutput)
				}

			}

		})
	}
}
