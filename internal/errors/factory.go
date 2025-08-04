package errors

import "time"

// NewAuthError creates an authentication error
func NewAuthError(message, resolution string) *ToolError {
	return &ToolError{
		Category:   CategoryAuth,
		Code:       "AUTH_FAILED",
		Message:    message,
		Resolution: resolution,
		Timestamp:  time.Now(),
		Retryable:  false,
	}
}

// NewValidationError creates a validation error
func NewValidationError(code, message, resolution string) *ToolError {
	return &ToolError{
		Category:   CategoryValidation,
		Code:       code,
		Message:    message,
		Resolution: resolution,
		Timestamp:  time.Now(),
		Retryable:  false,
	}
}

// NewAPIError creates an API error with retry information
func NewAPIError(statusCode int, code, message string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
		Retryable:  isRetryableStatusCode(statusCode),
	}
}

// NewVPNError creates a VPN-related error
func NewVPNError(code, message, resolution string) *ToolError {
	return &ToolError{
		Category:   CategoryVPN,
		Code:       code,
		Message:    message,
		Resolution: resolution,
		Timestamp:  time.Now(),
		Retryable:  true,
	}
}

// NewDeploymentError creates a deployment error
func NewDeploymentError(code, message, resolution string) *ToolError {
	return &ToolError{
		Category:   CategoryDeployment,
		Code:       code,
		Message:    message,
		Resolution: resolution,
		Timestamp:  time.Now(),
		Retryable:  false,
	}
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string, duration time.Duration) *ToolError {
	return &ToolError{
		Category:   CategoryTimeout,
		Code:       "OPERATION_TIMEOUT",
		Message:    "Operation timed out",
		Resolution: "Try again or increase the timeout duration",
		Timestamp:  time.Now(),
		Retryable:  true,
		Metadata: map[string]interface{}{
			"operation": operation,
			"duration":  duration.String(),
		},
	}
}

// NewInternalError creates an internal error
func NewInternalError(err error) *ToolError {
	return &ToolError{
		Category:   CategoryInternal,
		Code:       "INTERNAL_ERROR",
		Message:    "An internal error occurred",
		Resolution: "This is likely a bug. Please report it.",
		Timestamp:  time.Now(),
		Retryable:  false,
		Metadata: map[string]interface{}{
			"error": err.Error(),
		},
	}
}

// Helper function to determine if HTTP status code is retryable
func isRetryableStatusCode(code int) bool {
	switch code {
	case 429, 502, 503, 504:
		return true
	default:
		return code >= 500
	}
}