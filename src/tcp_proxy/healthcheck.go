package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"
)

type HealthChecker struct {
	proxyAddress string
	client       *http.Client
	debug        bool
}

func NewHealthChecker(proxyAddress string, debug bool) *HealthChecker {
	proxyURL, err := url.Parse(fmt.Sprintf("socks5://%s", proxyAddress))
	if err != nil {
		log.Printf("Failed to parse proxy URL: %v", err)
		return nil
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	return &HealthChecker{
		proxyAddress: proxyAddress,
		client: &http.Client{
			Transport: transport,
			Timeout:   5 * time.Second,
		},
		debug: debug,
	}
}

func (h *HealthChecker) Check() bool {
	// Try to connect to test endpoint
	resp, err := h.client.Get("http://192.168.0.1")
	if err != nil {
		log.Printf("Health check failed: %v", err)
		return false
	}
	defer resp.Body.Close()

	if h.debug {
		// Debug log response headers
		log.Printf("[DEBUG] Health check response headers: %v", resp.Header)
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return true
	}
	log.Printf("Health check failed with status: %d", resp.StatusCode)
	return false
}
