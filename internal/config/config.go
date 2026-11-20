package config

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

const defaultInterval = 60 * time.Second
const defaultHost = "https://api-e.ecoflow.com"

var ErrInvalidAccessKey = errors.New("config: invalid AccessKey")
var ErrInvalidSecretKey = errors.New("config: invalid SecretKey")
var ErrInvalidHost = errors.New("config: invalid Host")
var ErrInvalidInterval = errors.New("config: invalid Interval")

var domainRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)

type Config struct {
	AccessKey string
	SecretKey string
	Host      string
	SN        string
	Interval  time.Duration
}

func Load(path string) (*Config, error) {
	conf := Config{}
	envData, err := parse(path)
	if err != nil {
		return nil, err
	}

	if _, present := envData["EF_ACCESS_KEY"]; !present {
		return nil, fmt.Errorf("%w: key is missing", ErrInvalidAccessKey)
	} else if envData["EF_ACCESS_KEY"] == "" {
		return nil, fmt.Errorf("%w: value is empty", ErrInvalidAccessKey)
	}
	conf.AccessKey = envData["EF_ACCESS_KEY"]

	if _, present := envData["EF_SECRET_KEY"]; !present {
		return nil, fmt.Errorf("%w: key is missing", ErrInvalidSecretKey)
	} else if envData["EF_SECRET_KEY"] == "" {
		return nil, fmt.Errorf("%w: value is empty", ErrInvalidSecretKey)
	}
	conf.SecretKey = envData["EF_SECRET_KEY"]

	//EF_INTERVAL_SECONDS
	if envData["EF_INTERVAL_SECONDS"] == "" {
		conf.Interval = defaultInterval
	} else {
		interval, err := strconv.ParseInt(envData["EF_INTERVAL_SECONDS"], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("%w: value is invalid", ErrInvalidInterval)
		}
		if interval <= 0 {
			return nil, fmt.Errorf("%w: value is invalid", ErrInvalidInterval)
		} else if interval < 10 {
			return nil, fmt.Errorf("%w: must be at least 10 seconds", ErrInvalidInterval)
		}
		conf.Interval = time.Duration(interval) * time.Second
	}

	//EF_HOST
	conf.Host = envData["EF_HOST"]
	if envData["EF_HOST"] == "" {
		conf.Host = defaultHost
	} else {
		u, err := url.ParseRequestURI(envData["EF_HOST"])
		if (err != nil) || (u.Scheme != "https") || (u.Host == "") {
			return nil, fmt.Errorf("%w: value is invalid", ErrInvalidHost)
		}
		// отсеем штуки типа "https://exmap'le.com" или "https://exmap      le.com"
		if !domainRegex.MatchString(u.Hostname()) {
			return nil, fmt.Errorf("%w: value is invalid", ErrInvalidHost)
		}
	}

	//EF_SN
	conf.SN = envData["EF_SN"]

	return &conf, nil

}
