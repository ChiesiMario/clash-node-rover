package main

import (
	"bytes"
	"context"
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
	encodedName := url.PathEscape(name)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/proxies/"+encodedName, nil)
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

	var group ProxyGroup
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, err
	}

	return &group, nil
}

type DelayResponse struct {
	Delay int    `json:"delay"`
	Error string `json:"error"`
}

func (c *APIClient) TestProxyDelay(proxyName, testUrl string, timeout time.Duration) (int, error) {
	encodedName := url.PathEscape(proxyName)
	u := fmt.Sprintf("%s/proxies/%s/delay?timeout=%d&url=%s", c.BaseURL, encodedName, timeout.Milliseconds(), url.QueryEscape(testUrl))

	ctx, cancel := context.WithTimeout(context.Background(), timeout+2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
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

	payloadBytes, err := json.Marshal(map[string]string{"name": proxyName})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "PUT", u, bytes.NewReader(payloadBytes))
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

func (api *APIClient) TestBandwidth(testURL string, proxyURL string, timeout time.Duration) (float64, int64, error) {
	parsedProxy, err := url.Parse(proxyURL)
	if err != nil {
		return 0, 0, err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(parsedProxy),
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", testURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	startTime := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, 0, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var totalBytes int64
	buf := make([]byte, 32*1024)

	for {
		if time.Since(startTime) >= timeout {
			break
		}
		n, err := resp.Body.Read(buf)
		if n > 0 {
			totalBytes += int64(n)
		}
		if err != nil {
			break
		}
	}

	elapsed := time.Since(startTime)
	if elapsed.Seconds() <= 0 {
		return 0, totalBytes, nil
	}

	return (float64(totalBytes) / 1024.0) / elapsed.Seconds(), totalBytes, nil
}

type ProxyProvider struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	VehicleType string `json:"vehicleType"`
	Proxies     []struct {
		Name string `json:"name"`
	} `json:"proxies"`
}

type ProvidersResponse struct {
	Providers map[string]ProxyProvider `json:"providers"`
}

func (c *APIClient) GetProxyProviders() (map[string]ProxyProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/providers/proxies", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}

	var providerResp ProvidersResponse
	if err := json.NewDecoder(resp.Body).Decode(&providerResp); err != nil {
		return nil, err
	}

	return providerResp.Providers, nil
}
