package ecoflow

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	accessKey  string
	secretKey  string
	host       string
}

func NewClient(host, accessKey, secretKey string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		accessKey: accessKey,
		secretKey: secretKey,
		host:      host,
	}
}

func (c *Client) DeviceList() ([]byte, error) {
	endpoint := c.host + "/iot-open/sign/device/list"
	nonce := generateNonce()
	timestamp := currentTimestamp()

	params := make(map[string]string)

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("device list: create request: %w", err)
	}

	req.Header.Set("accessKey", c.accessKey)
	req.Header.Set("nonce", nonce)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("sign", sign(params, c.accessKey, c.secretKey, nonce, timestamp))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device list: do request: %w", err)
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("device list: read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device list: unexpected status %s: %s", resp.Status, res[:min(len(res), 200)])
	}
	return res, nil
}

func (c *Client) DeviceQuota(sn string) ([]byte, error) {
	return nil, errors.New("device quota: not implemented")
}
