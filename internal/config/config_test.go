package config_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/artman325/ecoflow-monitor/internal/config"
)

const expectedDefaultInterval = 60 * time.Second
const expectedDefaultHost = "https://api-e.ecoflow.com"
const bomSymbol = "\uFEFF"
const nonBreakingSpace = "\u00A0"

func TestMissingEnvFile(t *testing.T) {
	fPath := filepath.Join(t.TempDir(), "nonexistent.env")
	_, err := config.Load(fPath)
	if !errors.Is(err, config.ErrEnvFileNotFound) {
		t.Fatalf("err = %v, want %v", err, config.ErrEnvFileNotFound)
	}
}

func TestUnreadableEnvFile(t *testing.T) {
	// simulate an unreadable .env: chmod 000 is unreliable (root, Windows),
	// so pass a directory — it exists but ReadFile fails on it
	fPath := t.TempDir()
	_, err := config.Load(fPath)
	if !errors.Is(err, config.ErrEnvFileUnreadable) {
		t.Fatalf("err = %v, want %v", err, config.ErrEnvFileUnreadable)
	}
}

func TestEnvVariables(t *testing.T) {

	tests := []struct {
		name        string
		input       []byte
		wantErr     error
		wantErrText string // "" если не нужно проверять текст
		extraCheck  func(t *testing.T, envData *config.Config)
	}{
		{
			name:    "Totally empty env",
			input:   []byte(""),
			wantErr: config.ErrInvalidAccessKey,
		},
		{
			name:    "Line missing '='",
			input:   []byte("EF_ACCESS_KEY...\nEF_SECRET_KEY"),
			wantErr: config.ErrInvalidLine,
		},
		//EF_ACCESS_KEY
		{
			name:        "EF_ACCESS_KEY is required",
			input:       []byte("AA=BB\nCCC=DDD"),
			wantErr:     config.ErrInvalidAccessKey,
			wantErrText: "key is missing",
		},
		{
			name:        "EF_ACCESS_KEY must not be empty",
			input:       []byte("EF_ACCESS_KEY=\nCCC=DDD"),
			wantErr:     config.ErrInvalidAccessKey,
			wantErrText: "value is empty",
		},
		//EF_SECRET_KEY
		{
			name:        "EF_SECRET_KEY is required",
			input:       []byte("EF_ACCESS_KEY=AAA\nCCC=DDD"),
			wantErr:     config.ErrInvalidSecretKey,
			wantErrText: "key is missing",
		},
		{
			name:        "EF_SECRET_KEY must not be empty",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY="),
			wantErr:     config.ErrInvalidSecretKey,
			wantErrText: "value is empty",
		},
		// minimal correct file
		{
			name:    "minimal .env without errors",
			input:   []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB"),
			wantErr: nil,
			extraCheck: func(t *testing.T, cf *config.Config) {
				t.Helper()
				if cf.AccessKey != "AAA" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "AccessKey", cf.AccessKey, "AAA")
				}
				if cf.SecretKey != "BBB" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "SecretKey", cf.SecretKey, "BBB")
				}
			},
		},
		// maximum correct file
		{
			name:    "maximum .env without errors",
			input:   []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_HOST=https://api-e.ecoflow.com\nEF_SN=LOREMIPSUM\nEF_INTERVAL_SECONDS=660\n"),
			wantErr: nil,
			extraCheck: func(t *testing.T, cf *config.Config) {
				t.Helper()
				if cf.AccessKey != "AAA" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "AccessKey", cf.AccessKey, "AAA")
				}
				if cf.SecretKey != "BBB" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "SecretKey", cf.SecretKey, "BBB")
				}
				if cf.Host != "https://api-e.ecoflow.com" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "Host", cf.Host, "https://api-e.ecoflow.com")
				}
				if cf.SN != "LOREMIPSUM" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "SN", cf.SN, "LOREMIPSUM")
				}
				var expectedInterval = time.Duration(660) * time.Second
				if cf.Interval != expectedInterval {
					t.Errorf("поле %s: получено %v, ожидалось %v", "Interval", cf.Interval, expectedInterval)
				}
			},
		},
		//EF_INTERVAL_SECONDS
		{
			name:    fmt.Sprintf("EF_INTERVAL_SECONDS should be %v by default", expectedDefaultInterval),
			input:   []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_INTERVAL_SECONDS="),
			wantErr: nil,
			extraCheck: func(t *testing.T, cf *config.Config) {
				t.Helper()
				if cf.Interval != expectedDefaultInterval {
					t.Errorf("поле %s: получено %v, ожидалось %v", "Interval", cf.Interval, expectedDefaultInterval)
				}
			},
		},
		{
			name:        "EF_INTERVAL_SECONDS cannot be negative",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_INTERVAL_SECONDS=-8"),
			wantErr:     config.ErrInvalidInterval,
			wantErrText: "value is invalid",
		},
		{
			name:        "EF_INTERVAL_SECONDS: quoted value is not supported",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_INTERVAL_SECONDS='123'"),
			wantErr:     config.ErrInvalidInterval,
			wantErrText: "value is invalid",
		},
		{
			name:        "EF_INTERVAL_SECONDS must be a number",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_INTERVAL_SECONDS=asdasd"),
			wantErr:     config.ErrInvalidInterval,
			wantErrText: "value is invalid",
		},
		{
			name:        "EF_INTERVAL_SECONDS cannot be zero",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_INTERVAL_SECONDS=0"),
			wantErr:     config.ErrInvalidInterval,
			wantErrText: "value is invalid",
		},
		{
			name:        "EF_INTERVAL_SECONDS cannot be less than 10 seconds",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_INTERVAL_SECONDS=9"),
			wantErr:     config.ErrInvalidInterval,
			wantErrText: "must be at least 10 seconds",
		},
		{
			name:    "EF_INTERVAL_SECONDS - 10sec is ok",
			input:   []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_INTERVAL_SECONDS=10"),
			wantErr: nil,
		},
		{
			name:    "EF_INTERVAL_SECONDS - 11sec is ok too",
			input:   []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_INTERVAL_SECONDS=11"),
			wantErr: nil,
		},
		{
			name:    "EF_INTERVAL_SECONDS - test correct value",
			input:   []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_INTERVAL_SECONDS=548"),
			wantErr: nil,
			extraCheck: func(t *testing.T, cf *config.Config) {
				t.Helper()
				var expectedInterval = time.Duration(548) * time.Second
				if cf.Interval != expectedInterval {
					t.Errorf("поле %s: получено %v, ожидалось %v", "Interval", cf.Interval, expectedInterval)
				}
			},
		},

		//EF_HOST
		{
			name:    fmt.Sprintf("EF_HOST should be %v by default", expectedDefaultHost),
			input:   []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_HOST="),
			wantErr: nil,
			extraCheck: func(t *testing.T, cf *config.Config) {
				t.Helper()
				if cf.Host != expectedDefaultHost {
					t.Errorf("поле %s: получено %v, ожидалось %v", "Host", cf.Host, expectedDefaultHost)
				}
			},
		},
		{
			name:        "EF_HOST: bare word",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_HOST=ASD"),
			wantErr:     config.ErrInvalidHost,
			wantErrText: "value is invalid",
		},
		{
			name:        "EF_HOST: no hostname or wrong scheme",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_HOST=/ftp"),
			wantErr:     config.ErrInvalidHost,
			wantErrText: "value is invalid",
		},
		{
			name:        "EF_HOST: http instead of https",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_HOST=http://example.com"),
			wantErr:     config.ErrInvalidHost,
			wantErrText: "value is invalid",
		},
		{
			name:        "EF_HOST: single slash after scheme",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_HOST=https:/example.com"),
			wantErr:     config.ErrInvalidHost,
			wantErrText: "value is invalid",
		},
		{
			name:        "EF_HOST: domain without scheme",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_HOST=example.com"),
			wantErr:     config.ErrInvalidHost,
			wantErrText: "value is invalid",
		},
		{
			name:        "EF_HOST: quotes in hostname",
			input:       []byte("EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_HOST=https://ex''ample.com"),
			wantErr:     config.ErrInvalidHost,
			wantErrText: "value is invalid",
		},
		//---Other----
		//bom smbol
		{
			name:  "correct example with BOM Prefix",
			input: []byte(bomSymbol + "EF_ACCESS_KEY=AAA\nEF_SECRET_KEY=BBB\nEF_HOST=https://example.com"),
			extraCheck: func(t *testing.T, cf *config.Config) {
				t.Helper()
				if cf.AccessKey != "AAA" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "AccessKey", cf.AccessKey, "AAA")
				}
			},
		},
		{
			name:    "spaces around key and value",
			input:   []byte("   EF_ACCESS_KEY   =   AAA    \nEF_SECRET_KEY=BBB\nEF_HOST=https://example.com"),
			wantErr: nil,
			extraCheck: func(t *testing.T, cf *config.Config) {
				t.Helper()
				if cf.AccessKey != "AAA" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "AccessKey", cf.AccessKey, "AAA")
				}
			},
		},
		{
			name:    "nonBreakingSpace around key and value",
			input:   []byte("" + nonBreakingSpace + "EF_ACCESS_KEY" + nonBreakingSpace + "=" + nonBreakingSpace + "AAA" + nonBreakingSpace + "\nEF_SECRET_KEY=BBB\nEF_HOST=https://example.com"),
			wantErr: nil,
			extraCheck: func(t *testing.T, cf *config.Config) {
				t.Helper()
				if cf.AccessKey != "AAA" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "AccessKey", cf.AccessKey, "AAA")
				}
			},
		},
		{
			name:    "'=' inside value is preserved",
			input:   []byte("EF_ACCESS_KEY=====AAA====\nEF_SECRET_KEY=BBB\nEF_HOST=https://example.com"),
			wantErr: nil,
			extraCheck: func(t *testing.T, cf *config.Config) {
				t.Helper()
				if cf.AccessKey != "====AAA====" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "AccessKey", cf.AccessKey, "====AAA====")
				}
			},
		},
		{
			name:    "empty lines are allowed",
			input:   []byte("\r\nEF_ACCESS_KEY=AAA\n\r\n\r\n\r\nEF_SECRET_KEY=BBB\n\r\n\r\n\r\n"),
			wantErr: nil,
			extraCheck: func(t *testing.T, cf *config.Config) {
				t.Helper()
				if cf.AccessKey != "AAA" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "AccessKey", cf.AccessKey, "AAA")
				}
				if cf.SecretKey != "BBB" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "SecretKey", cf.SecretKey, "BBB")
				}
			},
		},
		{
			name:    "comments lines are allowed",
			input:   []byte("EF_ACCESS_KEY=AAA\n#EF_ACCESS_KEY=CCC\n#EF_SECRET_KEY=DDD\nEF_SECRET_KEY=BBB\n"),
			wantErr: nil,
			extraCheck: func(t *testing.T, cf *config.Config) {
				t.Helper()
				if cf.AccessKey != "AAA" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "AccessKey", cf.AccessKey, "AAA")
				}
				if cf.SecretKey != "BBB" {
					t.Errorf("поле %s: получено %v, ожидалось %v", "SecretKey", cf.SecretKey, "BBB")
				}
			},
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {

			// Создадим файл налету
			fPath := filepath.Join(t.TempDir(), ".env")
			err := os.WriteFile(fPath, tc.input, 0644)
			if err != nil {
				t.Fatal(err)
			}

			envData, err := config.Load(fPath)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErrText != "" && !strings.Contains(err.Error(), tc.wantErrText) {
				t.Errorf("текст ошибки %q не содержит подстроку %q", err.Error(), tc.wantErrText)
			}

			if tc.wantErr == nil && tc.extraCheck != nil {
				tc.extraCheck(t, envData)
			}

		})
	}
}
