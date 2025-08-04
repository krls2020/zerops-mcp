package utils

import (
	"regexp"
	"strings"
)

// Common patterns for sensitive data
var (
	// Matches typical API keys/tokens (long alphanumeric strings)
	apiKeyPattern = regexp.MustCompile(`[a-zA-Z0-9]{20,}`)
	
	// Matches JWT tokens
	jwtPattern = regexp.MustCompile(`eyJ[a-zA-Z0-9._-]+`)
	
	// Matches passwords in JSON or query parameters
	passwordPattern = regexp.MustCompile(`"password"\s*:\s*"[^"]+"|password=[^&\s]+`)
	
	// Matches Bearer tokens
	bearerPattern = regexp.MustCompile(`Bearer\s+[\w.-]+`)
	
	// Sensitive field names to mask in JSON
	sensitiveFields = []string{
		"password", "secret", "token", "apiKey", "api_key",
		"private_key", "privateKey", "access_token", "accessToken",
		"refresh_token", "refreshToken", "auth", "authorization",
	}
)

// MaskSensitive masks sensitive data in the given string
func MaskSensitive(input string) string {
	if input == "" {
		return input
	}
	
	// Make a copy to avoid modifying the original
	masked := input
	
	// Mask API keys
	masked = apiKeyPattern.ReplaceAllStringFunc(masked, func(match string) string {
		if len(match) > 8 {
			return match[:4] + strings.Repeat("*", len(match)-8) + match[len(match)-4:]
		}
		return strings.Repeat("*", len(match))
	})
	
	// Mask JWT tokens
	masked = jwtPattern.ReplaceAllString(masked, "eyJ****")
	
	// Mask passwords
	masked = passwordPattern.ReplaceAllStringFunc(masked, func(match string) string {
		if strings.Contains(match, ":") {
			// JSON format
			parts := strings.Split(match, ":")
			return parts[0] + ": \"****\""
		}
		// Query parameter format
		parts := strings.Split(match, "=")
		return parts[0] + "=****"
	})
	
	// Mask Bearer tokens
	masked = bearerPattern.ReplaceAllString(masked, "Bearer ****")
	
	// Mask sensitive fields in JSON-like structures
	for _, field := range sensitiveFields {
		// Match both quoted and unquoted field names
		pattern := regexp.MustCompile(`"` + field + `"\s*:\s*"[^"]+"|` + field + `\s*:\s*[^,}\s]+`)
		masked = pattern.ReplaceAllStringFunc(masked, func(match string) string {
			parts := strings.SplitN(match, ":", 2)
			if len(parts) == 2 {
				return parts[0] + ": \"****\""
			}
			return match
		})
	}
	
	return masked
}

// MaskAPIKey masks an API key keeping only first and last 4 characters
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return strings.Repeat("*", len(apiKey))
	}
	return apiKey[:4] + strings.Repeat("*", len(apiKey)-8) + apiKey[len(apiKey)-4:]
}

// IsSensitiveField checks if a field name is considered sensitive
func IsSensitiveField(fieldName string) bool {
	lowerField := strings.ToLower(fieldName)
	for _, sensitive := range sensitiveFields {
		if strings.Contains(lowerField, strings.ToLower(sensitive)) {
			return true
		}
	}
	return false
}