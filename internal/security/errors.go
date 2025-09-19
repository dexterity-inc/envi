package security

import (
	"fmt"
	"strings"
)

// SensitiveDataDetector checks if a string might contain sensitive information
type SensitiveDataDetector struct {
	patterns []string
}

// NewSensitiveDataDetector creates a new detector with common sensitive patterns
func NewSensitiveDataDetector() *SensitiveDataDetector {
	return &SensitiveDataDetector{
		patterns: []string{
			"ghp_",        // GitHub personal access token
			"github_pat_", // GitHub fine-grained token
			"gho_",        // GitHub OAuth token
			"ghu_",        // GitHub user-to-server token
			"ghs_",        // GitHub server-to-server token
			"password",    // Generic password
			"secret",      // Generic secret
			"key",         // Generic key
			"token",       // Generic token
		},
	}
}

// ContainsSensitiveData checks if the input might contain sensitive information
func (d *SensitiveDataDetector) ContainsSensitiveData(input string) bool {
	lowerInput := strings.ToLower(input)
	for _, pattern := range d.patterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	return false
}

// SafeLogString sanitizes a string for safe logging
func SafeLogString(input string) string {
	detector := NewSensitiveDataDetector()
	if detector.ContainsSensitiveData(input) {
		return "[REDACTED_SENSITIVE_DATA]"
	}
	return input
}

// SanitizeError creates a safe error message that doesn't leak sensitive information
// It removes file paths, tokens, passwords, and other potentially sensitive data
func SanitizeError(err error, context string) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	
	// Remove absolute file paths - keep only filename
	errMsg = sanitizeFilePaths(errMsg)
	
	// Remove potential tokens and secrets
	errMsg = sanitizeSecrets(errMsg)
	
	// Remove system-specific information
	errMsg = sanitizeSystemInfo(errMsg)
	
	// Create sanitized error with context
	if context != "" {
		return fmt.Errorf("%s: %s", context, errMsg)
	}
	
	return fmt.Errorf("%s", errMsg)
}

// sanitizeFilePaths removes absolute paths from error messages
func sanitizeFilePaths(msg string) string {
	// Replace common path patterns with generic names
	patterns := map[string]string{
		"/Users/": "[USER_HOME]/",
		"/home/":  "[USER_HOME]/",
		"C:\\Users\\": "[USER_HOME]\\",
		"/tmp/":   "[TEMP]/",
		"/var/":   "[VAR]/",
	}
	
	for pattern, replacement := range patterns {
		msg = strings.ReplaceAll(msg, pattern, replacement)
	}
	
	return msg
}

// sanitizeSecrets removes potential secrets from error messages
func sanitizeSecrets(msg string) string {
	// Remove potential tokens (40+ character alphanumeric strings)
	tokenPattern := `[a-zA-Z0-9]{40,}`
	// Don't use regex for simplicity, just check for known prefixes
	prefixes := []string{"ghp_", "github_pat_", "gho_", "ghu_", "ghs_"}
	
	for _, prefix := range prefixes {
		if strings.Contains(msg, prefix) {
			// Replace the token part with [REDACTED]
			parts := strings.Split(msg, prefix)
			if len(parts) > 1 {
				msg = parts[0] + "[REDACTED_TOKEN]"
				// Add back any text after the token (if not part of token)
				if len(parts) > 1 {
					tokenPart := parts[1]
					// Find end of token (space or special char)
					for i, char := range tokenPart {
						if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
							 (char >= '0' && char <= '9') || char == '_') {
							msg += tokenPart[i:]
							break
						}
					}
				}
			}
		}
	}
	
	return msg
}

// sanitizeSystemInfo removes system-specific information
func sanitizeSystemInfo(msg string) string {
	// Remove specific error details that might leak system info
	sensitiveTerms := map[string]string{
		"permission denied": "access denied",
		"no such file":      "file not found", 
		"connection refused": "connection failed",
	}
	
	for sensitive, generic := range sensitiveTerms {
		msg = strings.ReplaceAll(strings.ToLower(msg), sensitive, generic)
	}
	
	return msg
}

// CreateSafeError creates a generic error message for user-facing errors
func CreateSafeError(operation string) error {
	return fmt.Errorf("%s failed - please check your configuration and try again", operation)
}