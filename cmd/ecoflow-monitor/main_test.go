package main

import (
	"testing"
	"time"

	"github.com/artman325/ecoflow-monitor/internal/ecoflow"
)

func TestCalculatePause(t *testing.T) {

	type inputStruct struct {
		interval time.Duration
		failures int
	}

	tests := []struct {
		name       string
		input      inputStruct
		wantOutput time.Duration
	}{
		{
			name:       "Interval:60s, failures:0",
			input:      inputStruct{60 * time.Second, 0},
			wantOutput: 60 * time.Second,
		},
		{
			name:       "Interval:60s, failures:1",
			input:      inputStruct{60 * time.Second, 1},
			wantOutput: 1 * time.Minute,
		},
		{
			name:       "Interval:60s, failures:2",
			input:      inputStruct{60 * time.Second, 2},
			wantOutput: 2 * time.Minute,
		},
		{
			name:       "Interval:60s, failures:3",
			input:      inputStruct{60 * time.Second, 3},
			wantOutput: 3 * time.Minute,
		},
		{
			name:       "Interval:60s, failures:4",
			input:      inputStruct{60 * time.Second, 4},
			wantOutput: 5 * time.Minute,
		},
		{
			name:       "Interval:60s, failures:10",
			input:      inputStruct{60 * time.Second, 10},
			wantOutput: 5 * time.Minute,
		},
		{
			name:       "Interval:600s, failures:0",
			input:      inputStruct{600 * time.Second, 0},
			wantOutput: 10 * time.Minute,
		},
		{
			name:       "Interval:600s, failures:1",
			input:      inputStruct{600 * time.Second, 1},
			wantOutput: 10 * time.Minute,
		},
		{
			name:       "Interval:600s, failures:2",
			input:      inputStruct{600 * time.Second, 2},
			wantOutput: 10 * time.Minute,
		},
		{
			name:       "Interval:600s, failures:3",
			input:      inputStruct{600 * time.Second, 3},
			wantOutput: 10 * time.Minute,
		},
		{
			name:       "Interval:600s, failures:4",
			input:      inputStruct{600 * time.Second, 4},
			wantOutput: 10 * time.Minute,
		},
		{
			name:       "Interval:600s, failures:10",
			input:      inputStruct{600 * time.Second, 10},
			wantOutput: 10 * time.Minute,
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {

			d := calculatePause(tc.input.interval, tc.input.failures)

			if d != tc.wantOutput {
				t.Fatalf("got = %v, expected %v", d, tc.wantOutput)
			}

		})
	}

}

func TestNeedTelemetry(t *testing.T) {

	tests := []struct {
		name       string
		base       refreshState
		input      bool
		wantOutput bool
	}{
		{
			name:       "program just started. no telemetry yet. station offline",
			base:       refreshState{},
			input:      false,
			wantOutput: true,
		},
		{
			name:       "program just started. no telemetry yet. station online",
			base:       refreshState{},
			input:      true,
			wantOutput: true,
		},
		{
			name:       "telemetry present. station online",
			base:       refreshState{lastTelemetry: &ecoflow.Telemetry{}},
			input:      true,
			wantOutput: true,
		},
		{
			name:       "telemetry present. station offline",
			base:       refreshState{lastTelemetry: &ecoflow.Telemetry{}},
			input:      false,
			wantOutput: false,
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {

			got := tc.base.needTelemetry(tc.input)

			if got != tc.wantOutput {
				t.Fatalf("got = %v, expected %v", got, tc.wantOutput)
			}

		})
	}
}

func TestStateUpdateFailures(t *testing.T) {
	lastUpdate := time.Now().Add(time.Duration(-10) * time.Minute)
	now := time.Now()
	st := refreshState{
		lastTelemetry: &ecoflow.Telemetry{WattsInSum: 111},
		lastUpdate:    lastUpdate,
		failures:      0,
	}
	res := refreshResult{
		device:    nil,
		telemetry: &ecoflow.Telemetry{WattsInSum: 222},
		err:       ecoflow.ErrDeviceNotFound,
	}

	st.update(&res, now)

	expectedFailures := 1
	gotFailures := st.failures

	if gotFailures != expectedFailures {
		t.Fatalf("got = %v, expected %v", gotFailures, expectedFailures)
	}

	expectedWattsInSum := 111
	gotWattsInSum := st.lastTelemetry.WattsInSum
	if gotWattsInSum != expectedWattsInSum {
		t.Fatalf("got = %v, expected %v", gotWattsInSum, expectedWattsInSum)
	}

	expectedLastUpdate := lastUpdate
	gotLastUpdate := st.lastUpdate
	if !gotLastUpdate.Equal(expectedLastUpdate) {
		t.Fatalf("got = %v, expected %v", gotLastUpdate, expectedLastUpdate)
	}

}

func TestStateUpdateFailuresReset(t *testing.T) {

	st := refreshState{
		failures: 5,
	}
	res := refreshResult{
		device: &ecoflow.DeviceInfo{},
		err:    nil,
	}

	st.update(&res, time.Now())

	expectedFailures := 0
	gotFailures := st.failures

	if gotFailures != expectedFailures {
		t.Fatalf("got = %v, expected %v", gotFailures, expectedFailures)
	}

}

func TestStateUpdateTelemetryWhenOnline(t *testing.T) {
	lastUpdate := time.Now().Add(time.Duration(-10) * time.Minute)
	now := time.Now()

	st := refreshState{
		lastTelemetry: &ecoflow.Telemetry{WattsInSum: 111},
		lastUpdate:    lastUpdate,
		failures:      0,
	}
	res := refreshResult{
		device:    &ecoflow.DeviceInfo{Online: 1},
		telemetry: &ecoflow.Telemetry{WattsInSum: 222},
		err:       nil,
	}

	st.update(&res, now)

	expected := 222
	got := st.lastTelemetry.WattsInSum

	if got != expected {
		t.Fatalf("got = %v, expected %v", got, expected)
	}

	expectedLastUpdate := now
	gotLastUpdate := st.lastUpdate
	if !gotLastUpdate.Equal(expectedLastUpdate) {
		t.Fatalf("got = %v, expected %v", gotLastUpdate, expectedLastUpdate)
	}

}

func TestStateUpdateTelemetryWhenOffline(t *testing.T) {
	lastUpdate := time.Now().Add(time.Duration(-10) * time.Minute)
	now := time.Now()

	st := refreshState{
		lastTelemetry: &ecoflow.Telemetry{WattsInSum: 111},
		lastUpdate:    lastUpdate,
		failures:      0,
	}
	res := refreshResult{
		device:    &ecoflow.DeviceInfo{Online: 0},
		telemetry: &ecoflow.Telemetry{WattsInSum: 222},
		err:       nil,
	}

	st.update(&res, now)

	expectedFailures := 0
	gotFailures := st.failures

	if gotFailures != expectedFailures {
		t.Fatalf("got = %v, expected %v", gotFailures, expectedFailures)
	}

	expectedWattsInSum := 222
	gotWattsInSum := st.lastTelemetry.WattsInSum
	if gotWattsInSum != expectedWattsInSum {
		t.Fatalf("got = %v, expected %v", gotWattsInSum, expectedWattsInSum)
	}

	expectedLastUpdate := lastUpdate
	gotLastUpdate := st.lastUpdate
	if !gotLastUpdate.Equal(expectedLastUpdate) {
		t.Fatalf("got = %v, expected %v", gotLastUpdate, expectedLastUpdate)
	}

}

func TestStateUpdateKeepsTelemetryWhenOffline(t *testing.T) {
	lastUpdate := time.Now().Add(time.Duration(-10) * time.Minute)
	now := time.Now()

	st := refreshState{
		lastTelemetry: &ecoflow.Telemetry{WattsInSum: 111},
		lastUpdate:    lastUpdate,
		failures:      0,
	}
	res := refreshResult{
		device:    &ecoflow.DeviceInfo{Online: 0},
		telemetry: nil,
		err:       nil,
	}

	st.update(&res, now)

	expectedFailures := 0
	gotFailures := st.failures

	if gotFailures != expectedFailures {
		t.Fatalf("got = %v, expected %v", gotFailures, expectedFailures)
	}

	expectedWattsInSum := 111
	gotWattsInSum := st.lastTelemetry.WattsInSum
	if gotWattsInSum != expectedWattsInSum {
		t.Fatalf("got = %v, expected %v", gotWattsInSum, expectedWattsInSum)
	}

	expectedLastUpdate := lastUpdate
	gotLastUpdate := st.lastUpdate
	if !gotLastUpdate.Equal(expectedLastUpdate) {
		t.Fatalf("got = %v, expected %v", gotLastUpdate, expectedLastUpdate)
	}

}
