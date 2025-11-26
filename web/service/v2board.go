package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"x-ui/logger"
	"x-ui/web/entity"
)

type V2boardService struct{}

type ServerConfig struct {
	ServerPort         int                    `json:"server_port"`
	BaseConfig         BaseConfig             `json:"base_config"`
	Routes             []Route                `json:"routes"`
	Cipher             string                 `json:"cipher"`
	Obfs               string                 `json:"obfs"`
	ObfsSettings       map[string]interface{} `json:"obfs_settings"`
	ServerKey          string                 `json:"server_key"`
	Network            string                 `json:"network"`
	NetworkSettings    map[string]interface{} `json:"network_settings"`
	NetworkSettingsAlt map[string]interface{} `json:"networkSettings"`
	Flow               string                 `json:"flow"`
	TlsSettings        map[string]interface{} `json:"tls_settings"`
	Tls                int                    `json:"tls"`
	Host               string                 `json:"host"`
	ServerName         string                 `json:"server_name"`
}

type BaseConfig struct {
	PushInterval int `json:"push_interval"`
	PullInterval int `json:"pull_interval"`
}

type Route struct {
	Id          int      `json:"id"`
	Match       []string `json:"match"`
	Action      string   `json:"action"`
	ActionValue string   `json:"action_value"`
}

type UserList struct {
	Users []User `json:"users"`
}

type User struct {
	Id         int    `json:"id"`
	Uuid       string `json:"uuid"`
	SpeedLimit int    `json:"speed_limit"`
	Email      string `json:"email"`
}

type TrafficReport map[string][]int64

func NewV2boardService() *V2boardService {
	return &V2boardService{}
}

func (s *V2boardService) GetServerConfig(setting *entity.AllSetting, nodeId, nodeType string) (*ServerConfig, error) {
	if !setting.V2boardEnable || setting.V2boardUrl == "" || setting.V2boardToken == "" {
		return nil, fmt.Errorf("v2board integration not enabled or not configured")
	}

	url := fmt.Sprintf("%s/api/v1/server/UniProxy/config?node_id=%s&node_type=%s&token=%s",
		setting.V2boardUrl, nodeId, nodeType, setting.V2boardToken)

	resp, err := s.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 304 {
		return nil, fmt.Errorf("configuration unchanged")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var config ServerConfig
	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (s *V2boardService) GetUserList(setting *entity.AllSetting, nodeId, nodeType string) (*UserList, error) {
	if !setting.V2boardEnable || setting.V2boardUrl == "" || setting.V2boardToken == "" {
		return nil, fmt.Errorf("v2board integration not enabled or not configured")
	}

	url := fmt.Sprintf("%s/api/v1/server/UniProxy/user?node_id=%s&node_type=%s&token=%s",
		setting.V2boardUrl, nodeId, nodeType, setting.V2boardToken)

	resp, err := s.makeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 304 {
		return nil, fmt.Errorf("user list unchanged")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var userList UserList
	if err := json.NewDecoder(resp.Body).Decode(&userList); err != nil {
		return nil, err
	}

	return &userList, nil
}

func (s *V2boardService) ReportTraffic(setting *entity.AllSetting, nodeId, nodeType string, traffic TrafficReport) error {
	if !setting.V2boardEnable || setting.V2boardUrl == "" || setting.V2boardToken == "" {
		return fmt.Errorf("v2board integration not enabled or not configured")
	}

	url := fmt.Sprintf("%s/api/v1/server/UniProxy/push?node_id=%s&node_type=%s&token=%s",
		setting.V2boardUrl, nodeId, nodeType, setting.V2boardToken)

	data, err := json.Marshal(traffic)
	if err != nil {
		return err
	}

	resp, err := s.makeRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("traffic report failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (s *V2boardService) makeRequest(method, url string, body io.Reader) (*http.Response, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	logger.Info("Making v2board API request: ", method, " ", url)

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("v2board API request failed: ", err)
		return nil, err
	}

	return resp, nil
}
