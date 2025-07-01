package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestValidator_Required(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     interface{}
		wantValid bool
		wantError string
	}{
		{
			name:      "valid string",
			field:     "test_field",
			value:     "valid value",
			wantValid: true,
		},
		{
			name:      "empty string",
			field:     "test_field",
			value:     "",
			wantValid: false,
			wantError: "field cannot be empty",
		},
		{
			name:      "whitespace only string",
			field:     "test_field",
			value:     "   ",
			wantValid: false,
			wantError: "field cannot be empty",
		},
		{
			name:      "nil value",
			field:     "test_field",
			value:     nil,
			wantValid: false,
			wantError: "field is required",
		},
		{
			name:      "empty byte slice",
			field:     "test_field",
			value:     []byte{},
			wantValid: false,
			wantError: "field cannot be empty",
		},
		{
			name:      "valid byte slice",
			field:     "test_field",
			value:     []byte("test"),
			wantValid: true,
		},
		{
			name:      "empty string slice",
			field:     "test_field",
			value:     []string{},
			wantValid: false,
			wantError: "field cannot be empty",
		},
		{
			name:      "valid string slice",
			field:     "test_field",
			value:     []string{"test"},
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()

			got := v.Required(tt.field, tt.value)

			if got != tt.wantValid {
				t.Errorf("Required() = %v, want %v", got, tt.wantValid)
			}

			if !tt.wantValid {
				if !v.HasErrors() {
					t.Error("Expected validation errors but got none")
					return
				}

				errors := v.GetErrors()
				if len(errors) == 0 {
					t.Error("Expected validation errors but got empty slice")
					return
				}

				if errors[0].Field != tt.field {
					t.Errorf("Error field = %v, want %v", errors[0].Field, tt.field)
				}

				if tt.wantError != "" && !contains(errors[0].Message, tt.wantError) {
					t.Errorf("Error message = %v, want to contain %v", errors[0].Message, tt.wantError)
				}
			} else if v.HasErrors() {
				t.Errorf("Unexpected validation errors: %v", v.GetErrors())
			}
		})
	}
}

func TestValidator_MinLength(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		min       int
		wantValid bool
	}{
		{
			name:      "valid length",
			field:     "test_field",
			value:     "hello",
			min:       3,
			wantValid: true,
		},
		{
			name:      "exact minimum length",
			field:     "test_field",
			value:     "abc",
			min:       3,
			wantValid: true,
		},
		{
			name:      "too short",
			field:     "test_field",
			value:     "ab",
			min:       3,
			wantValid: false,
		},
		{
			name:      "empty string",
			field:     "test_field",
			value:     "",
			min:       1,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()

			got := v.MinLength(tt.field, tt.value, tt.min)

			if got != tt.wantValid {
				t.Errorf("MinLength() = %v, want %v", got, tt.wantValid)
			}

			if !tt.wantValid && !v.HasErrors() {
				t.Error("Expected validation errors but got none")
			}
		})
	}
}

func TestValidator_MaxLength(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		max       int
		wantValid bool
	}{
		{
			name:      "valid length",
			field:     "test_field",
			value:     "hello",
			max:       10,
			wantValid: true,
		},
		{
			name:      "exact maximum length",
			field:     "test_field",
			value:     "abc",
			max:       3,
			wantValid: true,
		},
		{
			name:      "too long",
			field:     "test_field",
			value:     "abcd",
			max:       3,
			wantValid: false,
		},
		{
			name:      "empty string",
			field:     "test_field",
			value:     "",
			max:       10,
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()

			got := v.MaxLength(tt.field, tt.value, tt.max)

			if got != tt.wantValid {
				t.Errorf("MaxLength() = %v, want %v", got, tt.wantValid)
			}

			if !tt.wantValid && !v.HasErrors() {
				t.Error("Expected validation errors but got none")
			}
		})
	}
}

func TestValidator_Pattern(t *testing.T) {
	tests := []struct {
		name      string
		field     string
		value     string
		pattern   string
		wantValid bool
	}{
		{
			name:      "valid email pattern",
			field:     "email",
			value:     "test@example.com",
			pattern:   `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			wantValid: true,
		},
		{
			name:      "invalid email pattern",
			field:     "email",
			value:     "invalid-email",
			pattern:   `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			wantValid: false,
		},
		{
			name:      "valid gist ID pattern",
			field:     "gist_id",
			value:     "a1b2c3d4e5f6789012345678901234567890",
			pattern:   PatternGistID,
			wantValid: false, // Too long for gist ID
		},
		{
			name:      "valid GitHub username",
			field:     "username",
			value:     "test-user123",
			pattern:   PatternGitHubUser,
			wantValid: true,
		},
		{
			name:      "invalid pattern syntax",
			field:     "test",
			value:     "test",
			pattern:   "[",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()

			got := v.Pattern(tt.field, tt.value, tt.pattern)

			if got != tt.wantValid {
				t.Errorf("Pattern() = %v, want %v", got, tt.wantValid)
			}

			if !tt.wantValid && !v.HasErrors() {
				t.Error("Expected validation errors but got none")
			}
		})
	}
}

func TestValidator_ValidPassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantValid bool
		wantError string
	}{
		{
			name:      "valid password",
			password:  "ValidPass123",
			wantValid: true,
		},
		{
			name:      "too short",
			password:  "Abc1",
			wantValid: false,
			wantError: "minimum length",
		},
		{
			name:      "no uppercase",
			password:  "validpass123",
			wantValid: false,
			wantError: "uppercase letter",
		},
		{
			name:      "no lowercase",
			password:  "VALIDPASS123",
			wantValid: false,
			wantError: "lowercase letter",
		},
		{
			name:      "no digit",
			password:  "ValidPassword",
			wantValid: false,
			wantError: "digit",
		},
		{
			name:      "minimum valid password",
			password:  "Valid123",
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()

			got := v.ValidPassword("password", tt.password)

			if got != tt.wantValid {
				t.Errorf("ValidPassword() = %v, want %v", got, tt.wantValid)
			}

			if !tt.wantValid {
				if !v.HasErrors() {
					t.Error("Expected validation errors but got none")
					return
				}

				errors := v.GetErrors()
				if len(errors) == 0 {
					t.Error("Expected validation errors but got empty slice")
					return
				}

				if tt.wantError != "" && !contains(errors[0].Message, tt.wantError) {
					t.Errorf("Error message = %v, want to contain %v", errors[0].Message, tt.wantError)
				}
			}
		})
	}
}

func TestValidator_ValidURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantValid bool
	}{
		{
			name:      "valid https URL",
			url:       "https://example.com",
			wantValid: true,
		},
		{
			name:      "valid http URL",
			url:       "http://example.com",
			wantValid: true,
		},
		{
			name:      "invalid URL without protocol",
			url:       "example.com",
			wantValid: false,
		},
		{
			name:      "invalid URL with wrong protocol",
			url:       "ftp://example.com",
			wantValid: false,
		},
		{
			name:      "empty URL",
			url:       "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()

			got := v.ValidURL("url", tt.url)

			if got != tt.wantValid {
				t.Errorf("ValidURL() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestValidator_ValidDate(t *testing.T) {
	tests := []struct {
		name      string
		date      string
		format    string
		wantValid bool
	}{
		{
			name:      "valid ISO date",
			date:      "2024-01-15T14:30:25Z",
			format:    time.RFC3339,
			wantValid: true,
		},
		{
			name:      "valid custom format",
			date:      "2024-01-15 14:30:25",
			format:    "2006-01-02 15:04:05",
			wantValid: true,
		},
		{
			name:      "invalid date format",
			date:      "2024/01/15",
			format:    "2006-01-02",
			wantValid: false,
		},
		{
			name:      "invalid date",
			date:      "invalid-date",
			format:    time.RFC3339,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()

			got := v.ValidDate("date", tt.date, tt.format)

			if got != tt.wantValid {
				t.Errorf("ValidDate() = %v, want %v", got, tt.wantValid)
			}
		})
	}
}

func TestValidator_FileOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "validator_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	readOnlyFile := filepath.Join(tempDir, "readonly.txt")
	if err := os.WriteFile(readOnlyFile, []byte("readonly"), 0444); err != nil {
		t.Fatalf("Failed to create readonly file: %v", err)
	}

	tests := []struct {
		name   string
		method string
		path   string
		want   bool
	}{
		{
			name:   "file exists - existing file",
			method: "FileExists",
			path:   testFile,
			want:   true,
		},
		{
			name:   "file exists - non-existing file",
			method: "FileExists",
			path:   filepath.Join(tempDir, "nonexistent.txt"),
			want:   false,
		},
		{
			name:   "file readable - readable file",
			method: "FileReadable",
			path:   testFile,
			want:   true,
		},
		{
			name:   "file readable - non-existing file",
			method: "FileReadable",
			path:   filepath.Join(tempDir, "nonexistent.txt"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			var got bool

			switch tt.method {
			case "FileExists":
				got = v.FileExists("test_field", tt.path)
			case "FileReadable":
				got = v.FileReadable("test_field", tt.path)
			}

			if got != tt.want {
				t.Errorf("%s() = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestValidateGistID(t *testing.T) {
	tests := []struct {
		name    string
		gistID  string
		wantErr bool
	}{
		{
			name:    "valid gist ID",
			gistID:  "a1b2c3d4e5f6789012345678901234ab",
			wantErr: false,
		},
		{
			name:    "empty gist ID",
			gistID:  "",
			wantErr: true,
		},
		{
			name:    "invalid format - too short",
			gistID:  "abc123",
			wantErr: true,
		},
		{
			name:    "invalid format - too long",
			gistID:  "a1b2c3d4e5f6789012345678901234abcd",
			wantErr: true,
		},
		{
			name:    "invalid format - invalid characters",
			gistID:  "g1h2i3j4k5l6789012345678901234ab",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGistID(tt.gistID)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGistID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateGitHubToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token - 40 chars",
			token:   "ghp_1234567890abcdef1234567890abcdef12345678",
			wantErr: false,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
		{
			name:    "too short token",
			token:   "short_token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitHubToken(tt.token)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGitHubToken() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEnvFile(t *testing.T) {
	// Create temporary file for testing
	tempDir, err := os.MkdirTemp("", "validate_env_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, ".env")
	if err := os.WriteFile(testFile, []byte("TEST_VAR=value"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid env file",
			path:    testFile,
			wantErr: false,
		},
		{
			name:    "non-existent file",
			path:    filepath.Join(tempDir, "nonexistent.env"),
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEnvFile(tt.path)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEnvFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(len(substr) == 0 ||
			(len(s) > 0 &&
				(s[:len(substr)] == substr ||
					(len(s) > len(substr) && contains(s[1:], substr)))))
}

// Benchmark tests for performance validation
func BenchmarkValidator_Required(b *testing.B) {
	v := NewValidator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.Required("test", "test value")
	}
}

func BenchmarkValidator_ValidPassword(b *testing.B) {
	v := NewValidator()
	password := "ValidPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.ValidPassword("password", password)
	}
}

func BenchmarkValidateGistID(b *testing.B) {
	gistID := "a1b2c3d4e5f6789012345678901234ab"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateGistID(gistID)
	}
}
