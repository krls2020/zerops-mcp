package api

import (
	"context"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"
)

// RetryConfig defines retry behavior
type RetryConfig struct {
	MaxRetries   int
	BaseDelay    time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	JitterFactor float64
}

// DefaultRetryConfig returns sensible retry defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		BaseDelay:    100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.1,
	}
}

// doRequestWithRetry performs a request with exponential backoff retry
func (c *Client) doRequestWithRetry(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	config := c.retryConfig
	if config.MaxRetries == 0 {
		config = DefaultRetryConfig()
	}
	
	var lastErr error
	
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Make the request
		resp, err := c.doRequest(ctx, method, path, body)
		
		// Success
		if err == nil {
			return resp, nil
		}
		
		// Check if error is retryable
		if !isRetryableError(err) || attempt == config.MaxRetries {
			return nil, err
		}
		
		lastErr = err
		
		// Calculate delay with exponential backoff and jitter
		delay := calculateBackoff(attempt, config)
		
		// Wait with context awareness
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
		
		// Log retry attempt if debug is enabled
		if c.debug {
			log.Printf("[DEBUG] Retrying request %s %s (attempt %d/%d) after %v", 
				method, path, attempt+1, config.MaxRetries, delay)
		}
	}
	
	return nil, lastErr
}

// calculateBackoff calculates the backoff duration for a retry attempt
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	// Exponential backoff
	backoff := float64(config.BaseDelay) * math.Pow(config.Multiplier, float64(attempt))
	
	// Cap at max delay
	if backoff > float64(config.MaxDelay) {
		backoff = float64(config.MaxDelay)
	}
	
	// Add jitter
	jitter := backoff * config.JitterFactor * (rand.Float64()*2 - 1)
	backoff += jitter
	
	return time.Duration(backoff)
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	
	errStr := err.Error()
	
	// Network errors
	networkErrors := []string{
		"connection refused",
		"connection reset",
		"i/o timeout",
		"temporary failure",
		"no such host",
		"network is unreachable",
	}
	
	for _, pattern := range networkErrors {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	
	// HTTP status codes that are retryable
	retryableStatuses := []string{
		"429", // Too Many Requests
		"502", // Bad Gateway
		"503", // Service Unavailable
		"504", // Gateway Timeout
	}
	
	for _, status := range retryableStatuses {
		if strings.Contains(errStr, status) {
			return true
		}
	}
	
	return false
}

// WithRetry creates a new client with custom retry configuration
func (c *Client) WithRetry(config RetryConfig) *Client {
	c.retryConfig = config
	return c
}