package errors

import (
	"fmt"
	"time"
)

// Category defines the type of error
type Category string

const (
	CategoryAuth       Category = "AUTH"
	CategoryValidation Category = "VALIDATION"
	CategoryAPI        Category = "API"
	CategoryVPN        Category = "VPN"
	CategoryDeployment Category = "DEPLOYMENT"
	CategoryTimeout    Category = "TIMEOUT"
	CategoryInternal   Category = "INTERNAL"
)

// ToolError represents a structured error with user-friendly information
type ToolError struct {
	Category   Category
	Code       string
	Message    string
	Resolution string
	NextTool   string
	Metadata   map[string]interface{}
	Timestamp  time.Time
	Retryable  bool
}

// Error implements the error interface
func (e *ToolError) Error() string {
	return fmt.Sprintf("[%s:%s] %s", e.Category, e.Code, e.Message)
}

// IsRetryable returns whether the error is retryable
func (e *ToolError) IsRetryable() bool {
	return e.Retryable
}

// WithMetadata adds metadata to the error
func (e *ToolError) WithMetadata(key string, value interface{}) *ToolError {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

// APIError represents an error from the Zerops API
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
	Retryable  bool
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error %d: %s - %s (RequestID: %s)", e.StatusCode, e.Code, e.Message, e.RequestID)
}

// ZCLIError represents an error from zcli execution
type ZCLIError struct {
	Command    string
	Args       []string
	Err        error
	Stdout     string
	Stderr     string
	ExitCode   int
	Duration   time.Duration
}

func (e *ZCLIError) Error() string {
	return fmt.Sprintf("zcli %s failed (exit %d): %v", e.Command, e.ExitCode, e.Err)
}

// IsTimeout checks if the error is a timeout
func (e *ZCLIError) IsTimeout() bool {
	return e.Duration > 0 && e.Err != nil && e.Err.Error() == "context deadline exceeded"
}

// IsSudoRequired checks if the error requires sudo
func (e *ZCLIError) IsSudoRequired() bool {
	return contains(e.Stderr, "permission denied") || contains(e.Stderr, "requires sudo")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}