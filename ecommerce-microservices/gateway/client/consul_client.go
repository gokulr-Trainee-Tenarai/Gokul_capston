package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ConsulServiceEntry struct {
	ServiceID   string `json:"ServiceID"`
	ServiceName string `json:"ServiceName"`
	ServiceAddr string `json:"ServiceAddress"`
	ServicePort int    `json:"ServicePort"`
}

type ConsulClient struct {
	consulURL  string
	httpClient *http.Client
	mu         sync.Mutex
	indices    map[string]int
}

func NewConsulClient(consulURL string) *ConsulClient {
	return &ConsulClient{
		consulURL:  consulURL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
		indices:    make(map[string]int),
	}
}

func (c *ConsulClient) GetServiceURL(serviceName string) (string, error) {
	url := fmt.Sprintf("%s/v1/catalog/service/%s", c.consulURL, serviceName)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		if serviceName == "product-service" {
			return "http://localhost:8081", nil
		}
		if serviceName == "order-service" {
			return "http://localhost:8082", nil
		}
		return "", fmt.Errorf("failed to contact Consul: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("consul returned status %d", resp.StatusCode)
	}

	var entries []ConsulServiceEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return "", fmt.Errorf("failed to decode consul response: %w", err)
	}

	if len(entries) == 0 {
		if serviceName == "product-service" {
			return "http://localhost:8081", nil
		}
		if serviceName == "order-service" {
			return "http://localhost:8082", nil
		}
		return "", fmt.Errorf("service '%s' not registered in Consul", serviceName)
	}

	c.mu.Lock()
	idx := c.indices[serviceName]
	selected := entries[idx%len(entries)]
	c.indices[serviceName] = idx + 1
	c.mu.Unlock()

	addr := selected.ServiceAddr
	if addr == "" {
		addr = "localhost"
	}
	return fmt.Sprintf("http://%s:%d", addr, selected.ServicePort), nil
}
