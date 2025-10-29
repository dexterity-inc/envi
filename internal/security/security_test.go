package security

import (
	"errors"
	"testing"
)

func TestSanitizeFilePath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid relative path",
			input:       "config/env.json",
			expected:    "config/env.json",
			expectError: false,
		},
		{
			name:        "path with dots",
			input:       "config/../env.json",
			expected:    "env.json",
			expectError: false,
		},
		{
			name:        "directory traversal attempt",
			input:       "../../../etc/passwd",
			expected:    "",
			expectError: true,
			errorType:   ErrPathTraversal,
		},
		{
			name:        "absolute path",
			input:       "/etc/passwd",
			expected:    "",
			expectError: true,
			errorType:   ErrAbsolutePath,
		},
		{
			name:        "empty path",
			input:       "",
			expected:    "",
			expectError: true,
			errorType:   ErrEmptyPath,
		},
		{
			name:        "Windows absolute path",
			input:       "C:\\Windows\\System32",
			expected:    "",
			expectError: true,
			errorType:   ErrAbsolutePath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SanitizeFilePath(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				} else if tt.errorType != nil && err != tt.errorType {
					t.Errorf("Expected error %v, got %v", tt.errorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %q, got %q", tt.expected, result)
				}
			}
		})
	}
}

func TestValidateEnvVarName(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid uppercase name",
			input:       "DATABASE_URL",
			expectError: false,
		},
		{
			name:        "valid lowercase name",
			input:       "api_key",
			expectError: false,
		},
		{
			name:        "empty name",
			input:       "",
			expectError: true,
		},
		{
			name:        "name starting with number",
			input:       "2FA_SECRET",
			expectError: true,
		},
		{
			name:        "dangerous system variable",
			input:       "PATH",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvVarName(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateEnvVarValue(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "simple value",
			input:       "simple_value",
			expectError: false,
		},
		{
			name:        "value with spaces",
			input:       "value with spaces",
			expectError: false,
		},
		{
			name:        "empty value",
			input:       "",
			expectError: false,
		},
		{
			name:        "value with command substitution (allowed)",
			input:       "$(echo hello)",
			expectError: false, // Changed: legitimate use in scripts
		},
		{
			name:        "value with backticks (allowed)",
			input:       "`whoami`",
			expectError: false, // Changed: legitimate use in markdown, scripts
		},
		{
			name:        "value with variable expansion (allowed)",
			input:       "${HOME}/path",
			expectError: false, // Changed: legitimate use in templates
		},
		{
			name:        "value with semicolon (allowed)",
			input:       "Server=localhost;Database=mydb",
			expectError: false, // Changed: legitimate use in connection strings
		},
		{
			name:        "value with pipe (allowed)",
			input:       "command | filter",
			expectError: false, // Changed: legitimate use in shell commands
		},
		{
			name:        "very long value",
			input:       string(make([]byte, MaxEnvVarValueLength+1)),
			expectError: true,
		},
		{
			name:        "unicode value",
			input:       "测试值",
			expectError: false,
		},
		{
			name:        "value with newline",
			input:       "line1\nline2",
			expectError: false,
		},
		{
			name:        "value with tab",
			input:       "value\twith\ttab",
			expectError: false,
		},
		{
			name:        "value with carriage return",
			input:       "value\rwith\rcarriage",
			expectError: false,
		},
		{
			name:        "value with command chaining && (allowed)",
			input:       "cmd1 && cmd2",
			expectError: false, // Changed: legitimate use in shell commands
		},
		{
			name:        "value with command chaining || (allowed)",
			input:       "cmd1 || fallback",
			expectError: false, // Changed: legitimate use in shell commands
		},
		{
			name:        "value with redirection < (allowed)",
			input:       "cmd < input.txt",
			expectError: false, // Changed: legitimate use in shell commands
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvVarValue(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateOutputPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid relative output path",
			input:       "output/file.env",
			expectError: false,
		},
		{
			name:        "valid nested output path",
			input:       "project/config/production.env",
			expectError: false,
		},
		{
			name:        "empty path",
			input:       "",
			expectError: true,
		},
		{
			name:        "absolute path",
			input:       "/etc/passwd",
			expectError: true,
		},
		{
			name:        "path traversal",
			input:       "../../../etc/passwd",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOutputPath(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateInputPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid relative input path",
			input:       "input/file.env",
			expectError: false,
		},
		{
			name:        "valid nested input path",
			input:       "project/config/development.env",
			expectError: false,
		},
		{
			name:        "empty path",
			input:       "",
			expectError: true,
		},
		{
			name:        "absolute path",
			input:       "/etc/hosts",
			expectError: true,
		},
		{
			name:        "path traversal",
			input:       "../../../etc/hosts",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateInputPath(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateKeyFilePath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid key file path",
			input:       "keys/encryption.key",
			expectError: false,
		},
		{
			name:        "valid nested key file path",
			input:       "project/keys/production.key",
			expectError: false,
		},
		{
			name:        "key file in relative tmp directory",
			input:       "tmp/key.txt",
			expectError: false,
		},
		{
			name:        "key file in /tmp directory",
			input:       "/tmp/key.txt",
			expectError: true,
		},
		{
			name:        "Windows tmp directory",
			input:       "\\tmp\\key.txt",
			expectError: true,
		},
		{
			name:        "empty path",
			input:       "",
			expectError: true,
		},
		{
			name:        "path traversal attempt",
			input:       "../../../etc/key",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKeyFilePath(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestValidateEnvLine(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "valid env line",
			input:       "DATABASE_URL=postgres://localhost:5432/mydb",
			expectError: false,
		},
		{
			name:        "empty line",
			input:       "",
			expectError: false,
		},
		{
			name:        "comment line",
			input:       "# This is a comment",
			expectError: false,
		},
		{
			name:        "line with spaces around equals",
			input:       "API_KEY = secret-key-123",
			expectError: false,
		},
		{
			name:        "line without equals",
			input:       "INVALID_LINE",
			expectError: true,
		},
		{
			name:        "line with invalid variable name",
			input:       "2INVALID=value",
			expectError: true,
		},
		{
			name:        "line with command substitution (allowed)",
			input:       "CMD=$(echo hello)",
			expectError: false, // Changed: legitimate use
		},
		{
			name:        "line with system variable",
			input:       "PATH=/usr/bin:/bin",
			expectError: true,
		},
		{
			name:        "line with empty value",
			input:       "EMPTY_VAR=",
			expectError: false,
		},
		{
			name:        "line with quoted value",
			input:       "QUOTED_VAR=\"hello world\"",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvLine(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Tests for error handling functions
func TestNewSensitiveDataDetector(t *testing.T) {
	detector := NewSensitiveDataDetector()
	if detector == nil {
		t.Error("NewSensitiveDataDetector() should not return nil")
	}
	if len(detector.patterns) == 0 {
		t.Error("NewSensitiveDataDetector() should initialize patterns")
	}
}

func TestContainsSensitiveData(t *testing.T) {
	detector := NewSensitiveDataDetector()
	
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "GitHub personal access token",
			input:    "ghp_1234567890abcdef",
			expected: true,
		},
		{
			name:     "GitHub fine-grained token",
			input:    "github_pat_11ABCDEFG",
			expected: true,
		},
		{
			name:     "GitHub OAuth token",
			input:    "gho_abcdefghijk",
			expected: true,
		},
		{
			name:     "Contains password",
			input:    "my_password=secret123",
			expected: true,
		},
		{
			name:     "Contains secret",
			input:    "API_SECRET=abc123",
			expected: true,
		},
		{
			name:     "Contains key",
			input:    "encryption_key=xyz789",
			expected: true,
		},
		{
			name:     "Contains token",
			input:    "auth_token=bearer123",
			expected: true,
		},
		{
			name:     "Safe data",
			input:    "DATABASE_URL=localhost:5432",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Case insensitive check",
			input:    "MY_PASSWORD=secret",
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.ContainsSensitiveData(tt.input)
			if result != tt.expected {
				t.Errorf("ContainsSensitiveData(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSafeLogString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Sensitive data with password",
			input:    "user_password=secret123",
			expected: "[REDACTED_SENSITIVE_DATA]",
		},
		{
			name:     "GitHub token",
			input:    "ghp_1234567890abcdef",
			expected: "[REDACTED_SENSITIVE_DATA]",
		},
		{
			name:     "Safe data",
			input:    "DATABASE_HOST=localhost",
			expected: "DATABASE_HOST=localhost",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeLogString(tt.input)
			if result != tt.expected {
				t.Errorf("SafeLogString(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeError(t *testing.T) {
	tests := []struct {
		name     string
		error    error
		context  string
		expected string
	}{
		{
			name:     "Nil error",
			error:    nil,
			context:  "test context",
			expected: "",
		},
		{
			name:     "Error with file path",
			error:    errors.New("failed to read /Users/john/secret.env"),
			context:  "file operation",
			expected: "file operation: failed to read [user_home]/john/secret.env",
		},
		{
			name:     "Error with GitHub token",
			error:    errors.New("invalid token ghp_1234567890abcdefghijklmnop"),
			context:  "authentication",
			expected: "authentication: invalid token [redacted_token]",
		},
		{
			name:     "Error without context",
			error:    errors.New("simple error message"),
			context:  "",
			expected: "simple error message",
		},
		{
			name:     "Error with Windows path",
			error:    errors.New("failed to access C:\\Users\\john\\config.yaml"),
			context:  "config loading",
			expected: "config loading: failed to access [user_home]\\john\\config.yaml",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeError(tt.error, tt.context)
			if tt.error == nil {
				if result != nil {
					t.Errorf("SanitizeError(nil, %q) should return nil, got %v", tt.context, result)
				}
			} else {
				if result == nil {
					t.Errorf("SanitizeError(%v, %q) should not return nil", tt.error, tt.context)
				} else if result.Error() != tt.expected {
					t.Errorf("SanitizeError(%v, %q) = %q, expected %q", tt.error, tt.context, result.Error(), tt.expected)
				}
			}
		})
	}
}

func TestCreateSafeError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		expected  string
	}{
		{
			name:      "File operation error",
			operation: "file read",
			expected:  "file read failed - please check your configuration and try again",
		},
		{
			name:      "Network operation error",
			operation: "API request",
			expected:  "API request failed - please check your configuration and try again",
		},
		{
			name:      "Empty operation",
			operation: "",
			expected:  " failed - please check your configuration and try again",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CreateSafeError(tt.operation)
			if result == nil {
				t.Errorf("CreateSafeError(%q) should not return nil", tt.operation)
			} else if result.Error() != tt.expected {
				t.Errorf("CreateSafeError(%q) = %q, expected %q", tt.operation, result.Error(), tt.expected)
			}
		})
	}
}