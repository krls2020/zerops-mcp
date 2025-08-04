package config

import (
	"fmt"
	"os"
	"time"
)

// Config holds the application configuration
type Config struct {
	ZeropsAPIKey  string
	ZeropsAPIURL  string
	APITimeout    time.Duration
	VPNWaitTime   time.Duration
	Debug         bool
}

// Load loads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		ZeropsAPIKey: os.Getenv("ZEROPS_API_KEY"),
		ZeropsAPIURL: os.Getenv("ZEROPS_API_URL"),
		Debug:        os.Getenv("ZEROPS_DEBUG") == "true" || os.Getenv("DEBUG") == "true",
	}

	// Set defaults
	if cfg.ZeropsAPIURL == "" {
		cfg.ZeropsAPIURL = "https://api.app-prg1.zerops.io"
	}

	// Parse timeout from environment or use default
	timeoutStr := os.Getenv("ZEROPS_API_TIMEOUT")
	if timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			cfg.APITimeout = timeout
		}
	}
	if cfg.APITimeout == 0 {
		cfg.APITimeout = 30 * time.Second
	}

	// Parse VPN wait time from environment or use default
	vpnWaitStr := os.Getenv("ZEROPS_VPN_WAIT_TIME")
	if vpnWaitStr != "" {
		if wait, err := time.ParseDuration(vpnWaitStr); err == nil {
			cfg.VPNWaitTime = wait
		}
	}
	if cfg.VPNWaitTime == 0 {
		cfg.VPNWaitTime = 2 * time.Second
	}

	return cfg
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.ZeropsAPIKey == "" {
		return fmt.Errorf("ZEROPS_API_KEY environment variable not set")
	}
	return nil
}