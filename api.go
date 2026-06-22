package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type APIClient struct {
	BaseURL    string
	Secret     string
	HTTPClient *http.Client
}

func NewAPIClient(baseURL, secret string) *APIClient {
	return &APIClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Secret:  secret,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) doRequest(req *http.Request) (*http.Response, error) {
	if c.Secret != "" {
		req.Header.Set("Authorization", "Bearer "+c.Secret)
	}
	return c.HTTPClient.Do(req)
}

type ProxyGroup struct {
	All  []string `json:"all"`
	Now  string   `json:"now"`
	Name string   `json:"name"`
	Type string   `json:"type"`
}

type ProxiesResponse struct {
	Proxies map[string]ProxyGroup `json:"proxies"`
}

func (c *APIClient) GetProxyGroup(name string) (*ProxyGroup, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/proxies", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data ProxiesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	if group, ok := data.Proxies[name]; ok {
		return &group, nil
	}

	return nil, fmt.Errorf("proxy group %s not found", name)
}

type DelayResponse struct {
	Delay int    `json:"delay"`
	Error string `json:"error"`
}

func (c *APIClient) TestProxyDelay(proxyName, testUrl string, timeout time.Duration) (int, error) {
	// API: /proxies/{name}/delay?timeout=5000&url=...
	encodedName := url.PathEscape(proxyName)
	u := fmt.Sprintf("%s/proxies/%s/delay?timeout=%d&url=%s", c.BaseURL, encodedName, timeout.Milliseconds(), url.QueryEscape(testUrl))

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data DelayResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	if data.Error != "" {
		return 0, fmt.Errorf("test error: %s", data.Error)
	}

	return data.Delay, nil
}

func (c *APIClient) SelectProxy(groupName, proxyName string) error {
	encodedGroup := url.PathEscape(groupName)
	u := fmt.Sprintf("%s/proxies/%s", c.BaseURL, encodedGroup)

	payload := fmt.Sprintf(`{"name": "%s"}`, proxyName)
	req, err := http.NewRequest("PUT", u, strings.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to select proxy, status: %d", resp.StatusCode)
	}

	return nil
}
