package security_test

import (
	"testing"

	"github.com/dexterity-inc/envi/internal/security"
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
			input:       "config/app.yaml",
			expected:    "config/app.yaml",
			expectError: false,
		},
		{
			name:        "path with dots",
			input:       "./config/app.yaml",
			expected:    "config/app.yaml",
			expectError: false,
		},
		{
			name:        "directory traversal attempt",
			input:       "../../../etc/passwd",
			expected:    "",
			expectError: true,
			errorType:   security.ErrPathTraversal,
		},
		{
			name:        "absolute path",
			input:       "/etc/passwd",
			expected:    "",
			expectError: true,
			errorType:   security.ErrAbsolutePath,
		},
		{
			name:        "empty path",
			input:       "",
			expected:    "",
			expectError: true,
			errorType:   security.ErrEmptyPath,
		},
		{
			name:        "path starting with dots",
			input:       "../config",
			expected:    "",
			expectError: true,
			errorType:   security.ErrPathTraversal,
		},
		{
			name:        "Windows absolute path",
			input:       "C:\\Windows\\System32",
			expected:    "",
			expectError: true,
			errorType:   security.ErrAbsolutePath,
		},
		{
			name:        "complex valid path",
			input:       "app/config/environments/production.yaml",
			expected:    "app/config/environments/production.yaml",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := security.SanitizeFilePath(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
					return
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("Expected error type %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidateEnvVarName(t *testing.T) {
	tests := []struct {
		name        string
		varName     string
		expectError bool
		errorType   error
	}{
		{
			name:        "valid uppercase name",
			varName:     "DATABASE_URL",
			expectError: false,
		},
		{
			name:        "valid lowercase name",
			varName:     "api_key",
			expectError: false,
		},
		{
			name:        "valid name with numbers",
			varName:     "PORT_8080",
			expectError: false,
		},
		{
			name:        "valid name starting with underscore",
			varName:     "_PRIVATE_KEY",
			expectError: false,
		},
		{
			name:        "empty name",
			varName:     "",
			expectError: true,
			errorType:   security.ErrEnvVarEmpty,
		},
		{
			name:        "name starting with number",
			varName:     "1_INVALID",
			expectError: true,
			errorType:   security.ErrInvalidEnvVarName,
		},
		{
			name:        "name with special characters",
			varName:     "INVALID-NAME",
			expectError: true,
			errorType:   security.ErrInvalidEnvVarName,
		},
		{
			name:        "name with spaces",
			varName:     "INVALID NAME",
			expectError: true,
			errorType:   security.ErrInvalidEnvVarName,
		},
		{
			name:        "dangerous system variable",
			varName:     "PATH",
			expectError: true,
		},
		{
			name:        "dangerous system variable HOME",
			varName:     "HOME",
			expectError: true,
		},
		{
			name:        "too long name",
			varName:     "THIS_IS_A_VERY_LONG_ENVIRONMENT_VARIABLE_NAME_THAT_EXCEEDS_THE_MAXIMUM_ALLOWED_LENGTH_LIMIT_FOR_VARIABLE_NAMES_IN_THIS_SYSTEM_123456789",
			expectError: true,
			errorType:   security.ErrEnvVarTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := security.ValidateEnvVarName(tt.varName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for name '%s', but got none", tt.varName)
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("Expected error type %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for name '%s': %v", tt.varName, err)
			}
		})
	}
}

func TestValidateEnvVarValue(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
		errorType   error
	}{
		{
			name:        "simple value",
			value:       "simple_value",
			expectError: false,
		},
		{
			name:        "value with spaces",
			value:       "value with spaces",
			expectError: false,
		},
		{
			name:        "empty value",
			value:       "",
			expectError: false,
		},
		{
			name:        "value with newlines",
			value:       "line1\nline2",
			expectError: false,
		},
		{
			name:        "value with tabs",
			value:       "value\twith\ttabs",
			expectError: false,
		},
		{
			name:        "value with command substitution",
			value:       "$(dangerous command)",
			expectError: true,
		},
		{
			name:        "value with backticks",
			value:       "`dangerous command`",
			expectError: true,
		},
		{
			name:        "value with variable expansion",
			value:       "${OTHER_VAR}",
			expectError: true,
		},
		{
			name:        "value with semicolon",
			value:       "value; rm -rf /",
			expectError: true,
		},
		{
			name:        "value with pipe",
			value:       "value | dangerous_command",
			expectError: true,
		},
		{
			name:        "value with redirection",
			value:       "value > /etc/passwd",
			expectError: true,
		},
		{
			name:        "very long value",
			value:       generateLongString(5000),
			expectError: true,
			errorType:   security.ErrEnvVarTooLong,
		},
		{
			name:        "unicode value",
			value:       "Hello 世界 🌍",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := security.ValidateEnvVarValue(tt.value)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for value '%s', but got none", tt.value)
				}
				if tt.errorType != nil && err != tt.errorType {
					t.Errorf("Expected error type %v, got %v", tt.errorType, err)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for value '%s': %v", tt.value, err)
			}
		})
	}
}

func TestValidateEnvLine(t *testing.T) {
	tests := []struct {
		name        string
		line        string
		expectError bool
	}{
		{
			name:        "valid env line",
			line:        "DATABASE_URL=postgres://localhost:5432/mydb",
			expectError: false,
		},
		{
			name:        "empty line",
			line:        "",
			expectError: false,
		},
		{
			name:        "comment line",
			line:        "# This is a comment",
			expectError: false,
		},
		{
			name:        "line with spaces around equals",
			line:        "API_KEY = secret123",
			expectError: false,
		},
		{
			name:        "line without equals",
			line:        "INVALID_LINE",
			expectError: true,
		},
		{
			name:        "line with invalid variable name",
			line:        "123INVALID=value",
			expectError: true,
		},
		{
			name:        "line with dangerous value",
			line:        "COMMAND=$(rm -rf /)",
			expectError: true,
		},
		{
			name:        "line with system variable",
			line:        "PATH=/usr/bin",
			expectError: true,
		},
		{
			name:        "line with empty value",
			line:        "EMPTY_VAR=",
			expectError: false,
		},
		{
			name:        "line with quoted value",
			line:        `QUOTED_VAR="quoted value"`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := security.ValidateEnvLine(tt.line)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for line '%s', but got none", tt.line)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for line '%s': %v", tt.line, err)
			}
		})
	}
}

func TestValidateOutputPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "valid output path",
			path:        "output.env",
			expectError: false,
		},
		{
			name:        "valid nested path",
			path:        "config/output.env",
			expectError: false,
		},
		{
			name:        "dangerous system path",
			path:        "/etc/passwd",
			expectError: true,
		},
		{
			name:        "path traversal attempt",
			path:        "../../../etc/passwd",
			expectError: true,
		},
		{
			name:        "tmp directory",
			path:        "/tmp/output.env",
			expectError: true,
		},
		{
			name:        "system bin directory",
			path:        "/usr/bin/malicious",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := security.ValidateOutputPath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for path '%s', but got none", tt.path)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for path '%s': %v", tt.path, err)
			}
		})
	}
}

func TestValidateKeyFilePath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "valid key file path",
			path:        ".envi.key",
			expectError: false,
		},
		{
			name:        "valid nested key file path",
			path:        "keys/encryption.key",
			expectError: false,
		},
		{
			name:        "key file in tmp directory",
			path:        "/tmp/key.key",
			expectError: true,
		},
		{
			name:        "Windows tmp directory",
			path:        "\\tmp\\key.key",
			expectError: true,
		},
		{
			name:        "path traversal attempt",
			path:        "../../../key.key",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := security.ValidateKeyFilePath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for path '%s', but got none", tt.path)
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error for path '%s': %v", tt.path, err)
			}
		})
	}
}

// Helper function to generate long strings for testing
func generateLongString(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = 'a'
	}
	return string(result)
}

