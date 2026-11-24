package ecoflow

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

type apiResponse struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type DeviceInfo struct {
	SN          string `json:"sn"`
	DeviceName  string `json:"deviceName"`
	Online      int    `json:"online"`
	ProductName string `json:"productName"`
}

type DeviceStatus struct {
	LcdShowSoc  int `json:"ems.lcdShowSoc"` // battery level shown on the station's screen, %
	WattsInSum  int `json:"pd.wattsInSum"`  // total input power, W
	WattsOutSum int `json:"pd.wattsOutSum"` // total output power (all ports), W
	RemainTime  int `json:"pd.remainTime"`  // minutes until full (charging) or until empty (discharging), always positive; 5999 = neither

	InputWatts     int `json:"inv.inputWatts"`    // AC input (grid) power, W; 0 while plugged in means grid power is out
	OutputWatts    int `json:"inv.outputWatts"`   // AC outlets output power, W
	Temp           int `json:"bmsMaster.temp"`    // main battery temperature, °C
	SysChgDsgState int `json:"pd.sysChgDsgState"` // observed: 1 = discharging, 2 = charging; passthrough TBD
}

type APIError struct {
	Code    string
	Message string
}

const (
	codeInvalidAccessKey   = "8513"
	codeInvalidSignature   = "8521"
	codeDeviceNoPermission = "8512" // observed for an SN that is not bound to the account
)

func (e *APIError) Error() string {
	return fmt.Sprintf("api error %s: %s", e.Code, e.Message)
}
func (e *APIError) IsCredentialsError() bool {
	return e.Code == codeInvalidAccessKey || e.Code == codeInvalidSignature
}

func (c *Client) DeviceList() ([]DeviceInfo, error) {
	path := "/iot-open/sign/device/list"

	var params map[string]string
	res, err := c.get(path, params)
	if err != nil {
		return nil, fmt.Errorf("device list: %w", err)
	}

	var devices []DeviceInfo
	if err := json.Unmarshal(res, &devices); err != nil {
		return nil, fmt.Errorf("device list: decode data: %w", err)

	}

	return devices, nil
}

func (c *Client) DeviceQuota(sn string) (*DeviceStatus, error) {
	path := "/iot-open/sign/device/quota/all"

	params := make(map[string]string)
	params["sn"] = sn

	res, err := c.get(path, params)
	if err != nil {
		return nil, fmt.Errorf("device quota: %w", err)
	}

	var deviceStatus DeviceStatus
	if err := json.Unmarshal(res, &deviceStatus); err != nil {
		return nil, fmt.Errorf("device quota: decode data: %w", err)
	}

	return &deviceStatus, nil

}

func (c *Client) get(path string, params map[string]string) (json.RawMessage, error) {
	endpoint := c.host + path
	nonce := generateNonce()
	timestamp := currentTimestamp()

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	var v = make(url.Values)
	for key, val := range params {
		v.Set(key, val)
	}
	req.URL.RawQuery = v.Encode()

	req.Header.Set("accessKey", c.accessKey)
	req.Header.Set("nonce", nonce)
	req.Header.Set("timestamp", timestamp)
	req.Header.Set("sign", sign(params, c.accessKey, c.secretKey, nonce, timestamp))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %s: %s", resp.Status, res[:min(len(res), 200)])
	}

	var rawResponse apiResponse
	if err := json.Unmarshal(res, &rawResponse); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if rawResponse.Code != "0" {
		return nil, &APIError{Code: rawResponse.Code, Message: rawResponse.Message}
	}

	return rawResponse.Data, nil
}
