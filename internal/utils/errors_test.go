package utils

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *AppError
		wantMsg string
	}{
		{
			name: "validation error",
			err: &AppError{
				Type:    ErrorTypeValidation,
				Message: "validation failed",
				Context: "username field",
			},
			wantMsg: "[username field] validation failed",
		},
		{
			name: "config error",
			err: &AppError{
				Type:    ErrorTypeConfig,
				Message: "config file not found",
			},
			wantMsg: "config file not found",
		},
		{
			name: "file error with original",
			err: &AppError{
				Type:        ErrorTypeFile,
				Message:     "file read failed",
				OriginalErr: errors.New("permission denied"),
			},
			wantMsg: "file read failed (permission denied)",
		},
		{
			name: "network error with context",
			err: &AppError{
				Type:        ErrorTypeNetwork,
				Message:     "connection timeout",
				Context:     "GitHub API",
				OriginalErr: errors.New("dial tcp: timeout"),
			},
			wantMsg: "[GitHub API] connection timeout (dial tcp: timeout)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantMsg {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.wantMsg)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := AppError{
		Type:        ErrorTypeValidation,
		Message:     "validation failed",
		OriginalErr: originalErr,
	}

	unwrapped := appErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// Test with no original error
	appErrNoOriginal := AppError{
		Type:    ErrorTypeConfig,
		Message: "config error",
	}

	unwrappedNil := appErrNoOriginal.Unwrap()
	if unwrappedNil != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrappedNil)
	}
}

func TestAppError_WithContext(t *testing.T) {
	appErr := AppError{
		Type:    ErrorTypeValidation,
		Message: "validation failed",
	}

	contextErr := appErr.WithContext("user input")

	if contextErr.Context != "user input" {
		t.Errorf("WithContext() context = %v, want %v", contextErr.Context, "user input")
	}

	// Ensure original error is unchanged
	if appErr.Context != "" {
		t.Errorf("Original error context changed: %v", appErr.Context)
	}
}

func TestAppError_WithSuggestions(t *testing.T) {
	appErr := AppError{
		Type:    ErrorTypeConfig,
		Message: "config file not found",
	}

	suggestions := []string{"Create config file", "Check file path"}
	suggestionErr := appErr.WithSuggestions(suggestions...)

	if len(suggestionErr.Suggestions) != len(suggestions) {
		t.Errorf("WithSuggestions() suggestions length = %v, want %v", len(suggestionErr.Suggestions), len(suggestions))
	}

	for i, suggestion := range suggestions {
		if suggestionErr.Suggestions[i] != suggestion {
			t.Errorf("WithSuggestions() suggestion[%d] = %v, want %v", i, suggestionErr.Suggestions[i], suggestion)
		}
	}
}

func TestAppError_WithExitCode(t *testing.T) {
	appErr := AppError{
		Type:     ErrorTypeSystem,
		Message:  "system error",
		ExitCode: 1,
	}

	exitCodeErr := appErr.WithExitCode(42)

	if exitCodeErr.ExitCode != 42 {
		t.Errorf("WithExitCode() exit code = %v, want %v", exitCodeErr.ExitCode, 42)
	}

	// Ensure original error is unchanged
	if appErr.ExitCode != 1 {
		t.Errorf("Original error exit code changed: %v", appErr.ExitCode)
	}
}

func TestNewError(t *testing.T) {
	errorType := ErrorTypeValidation
	message := "validation failed"

	appErr := NewError(errorType, message)

	if appErr.Type != errorType {
		t.Errorf("NewError() type = %v, want %v", appErr.Type, errorType)
	}

	if appErr.Message != message {
		t.Errorf("NewError() message = %v, want %v", appErr.Message, message)
	}

	if appErr.ExitCode != 1 {
		t.Errorf("NewError() exit code = %v, want %v", appErr.ExitCode, 1)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	errorType := ErrorTypeFile
	message := "file operation failed"

	appErr := WrapError(originalErr, errorType, message)

	if appErr.Type != errorType {
		t.Errorf("WrapError() type = %v, want %v", appErr.Type, errorType)
	}

	if appErr.Message != message {
		t.Errorf("WrapError() message = %v, want %v", appErr.Message, message)
	}

	if appErr.OriginalErr != originalErr {
		t.Errorf("WrapError() original error = %v, want %v", appErr.OriginalErr, originalErr)
	}

	if appErr.ExitCode != 1 {
		t.Errorf("WrapError() exit code = %v, want %v", appErr.ExitCode, 1)
	}
}

func TestErrorConstructors(t *testing.T) {
	tests := []struct {
		name        string
		constructor func(string) AppError
		errorType   ErrorType
		message     string
	}{
		{
			name:        "NewValidationError",
			constructor: NewValidationError,
			errorType:   ErrorTypeValidation,
			message:     "validation failed",
		},
		{
			name:        "NewConfigError",
			constructor: NewConfigError,
			errorType:   ErrorTypeConfig,
			message:     "config error",
		},
		{
			name:        "NewFileError",
			constructor: NewFileError,
			errorType:   ErrorTypeFile,
			message:     "file error",
		},
		{
			name:        "NewNetworkError",
			constructor: NewNetworkError,
			errorType:   ErrorTypeNetwork,
			message:     "network error",
		},
		{
			name:        "NewEncryptionError",
			constructor: NewEncryptionError,
			errorType:   ErrorTypeEncryption,
			message:     "encryption error",
		},
		{
			name:        "NewGitHubError",
			constructor: NewGitHubError,
			errorType:   ErrorTypeGitHub,
			message:     "github error",
		},
		{
			name:        "NewPermissionError",
			constructor: NewPermissionError,
			errorType:   ErrorTypePermission,
			message:     "permission error",
		},
		{
			name:        "NewInputError",
			constructor: NewInputError,
			errorType:   ErrorTypeInput,
			message:     "input error",
		},
		{
			name:        "NewSystemError",
			constructor: NewSystemError,
			errorType:   ErrorTypeSystem,
			message:     "system error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := tt.constructor(tt.message)

			if appErr.Type != tt.errorType {
				t.Errorf("%s() type = %v, want %v", tt.name, appErr.Type, tt.errorType)
			}

			if appErr.Message != tt.message {
				t.Errorf("%s() message = %v, want %v", tt.name, appErr.Message, tt.message)
			}
		})
	}
}

func TestWrapErrorConstructors(t *testing.T) {
	originalErr := errors.New("original error")
	message := "wrapped error"

	tests := []struct {
		name        string
		constructor func(error, string) AppError
		errorType   ErrorType
	}{
		{
			name:        "WrapValidationError",
			constructor: WrapValidationError,
			errorType:   ErrorTypeValidation,
		},
		{
			name:        "WrapConfigError",
			constructor: WrapConfigError,
			errorType:   ErrorTypeConfig,
		},
		{
			name:        "WrapFileError",
			constructor: WrapFileError,
			errorType:   ErrorTypeFile,
		},
		{
			name:        "WrapNetworkError",
			constructor: WrapNetworkError,
			errorType:   ErrorTypeNetwork,
		},
		{
			name:        "WrapEncryptionError",
			constructor: WrapEncryptionError,
			errorType:   ErrorTypeEncryption,
		},
		{
			name:        "WrapGitHubError",
			constructor: WrapGitHubError,
			errorType:   ErrorTypeGitHub,
		},
		{
			name:        "WrapPermissionError",
			constructor: WrapPermissionError,
			errorType:   ErrorTypePermission,
		},
		{
			name:        "WrapInputError",
			constructor: WrapInputError,
			errorType:   ErrorTypeInput,
		},
		{
			name:        "WrapSystemError",
			constructor: WrapSystemError,
			errorType:   ErrorTypeSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := tt.constructor(originalErr, message)

			if appErr.Type != tt.errorType {
				t.Errorf("%s() type = %v, want %v", tt.name, appErr.Type, tt.errorType)
			}

			if appErr.Message != message {
				t.Errorf("%s() message = %v, want %v", tt.name, appErr.Message, message)
			}

			if appErr.OriginalErr != originalErr {
				t.Errorf("%s() original error = %v, want %v", tt.name, appErr.OriginalErr, originalErr)
			}
		})
	}
}

func TestErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name      string
		error     error
		checker   func(error) bool
		errorType ErrorType
		want      bool
	}{
		{
			name:      "IsValidationError with validation error",
			error:     NewValidationError("validation failed"),
			checker:   IsValidationError,
			errorType: ErrorTypeValidation,
			want:      true,
		},
		{
			name:      "IsValidationError with config error",
			error:     NewConfigError("config failed"),
			checker:   IsValidationError,
			errorType: ErrorTypeConfig,
			want:      false,
		},
		{
			name:      "IsConfigError with config error",
			error:     NewConfigError("config failed"),
			checker:   IsConfigError,
			errorType: ErrorTypeConfig,
			want:      true,
		},
		{
			name:      "IsFileError with file error",
			error:     NewFileError("file failed"),
			checker:   IsFileError,
			errorType: ErrorTypeFile,
			want:      true,
		},
		{
			name:      "IsNetworkError with network error",
			error:     NewNetworkError("network failed"),
			checker:   IsNetworkError,
			errorType: ErrorTypeNetwork,
			want:      true,
		},
		{
			name:      "IsEncryptionError with encryption error",
			error:     NewEncryptionError("encryption failed"),
			checker:   IsEncryptionError,
			errorType: ErrorTypeEncryption,
			want:      true,
		},
		{
			name:      "IsGitHubError with github error",
			error:     NewGitHubError("github failed"),
			checker:   IsGitHubError,
			errorType: ErrorTypeGitHub,
			want:      true,
		},
		{
			name:      "IsPermissionError with permission error",
			error:     NewPermissionError("permission failed"),
			checker:   IsPermissionError,
			errorType: ErrorTypePermission,
			want:      true,
		},
		{
			name:      "IsInputError with input error",
			error:     NewInputError("input failed"),
			checker:   IsInputError,
			errorType: ErrorTypeInput,
			want:      true,
		},
		{
			name:      "IsSystemError with system error",
			error:     NewSystemError("system failed"),
			checker:   IsSystemError,
			errorType: ErrorTypeSystem,
			want:      true,
		},
		{
			name:    "error type checker with standard error",
			error:   errors.New("standard error"),
			checker: IsValidationError,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.checker(tt.error)
			if got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestErrorType_String(t *testing.T) {
	tests := []struct {
		name      string
		errorType ErrorType
		want      string
	}{
		{
			name:      "ErrorTypeValidation",
			errorType: ErrorTypeValidation,
			want:      "validation",
		},
		{
			name:      "ErrorTypeConfig",
			errorType: ErrorTypeConfig,
			want:      "configuration",
		},
		{
			name:      "ErrorTypeFile",
			errorType: ErrorTypeFile,
			want:      "file",
		},
		{
			name:      "ErrorTypeNetwork",
			errorType: ErrorTypeNetwork,
			want:      "network",
		},
		{
			name:      "ErrorTypeEncryption",
			errorType: ErrorTypeEncryption,
			want:      "encryption",
		},
		{
			name:      "ErrorTypeGitHub",
			errorType: ErrorTypeGitHub,
			want:      "github",
		},
		{
			name:      "ErrorTypePermission",
			errorType: ErrorTypePermission,
			want:      "permission",
		},
		{
			name:      "ErrorTypeInput",
			errorType: ErrorTypeInput,
			want:      "input",
		},
		{
			name:      "ErrorTypeSystem",
			errorType: ErrorTypeSystem,
			want:      "system",
		},
		{
			name:      "Unknown error type",
			errorType: ErrorType(999),
			want:      "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.errorType.String()
			if got != tt.want {
				t.Errorf("ErrorType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test error creation with method chaining
func TestErrorChaining(t *testing.T) {
	originalErr := errors.New("database connection failed")

	appErr := WrapError(originalErr, ErrorTypeNetwork, "failed to connect to database").
		WithContext("user authentication").
		WithSuggestions("Check network connection", "Verify database credentials").
		WithExitCode(5)

	if appErr.Type != ErrorTypeNetwork {
		t.Errorf("Error type = %v, want %v", appErr.Type, ErrorTypeNetwork)
	}

	if appErr.Context != "user authentication" {
		t.Errorf("Error context = %v, want %v", appErr.Context, "user authentication")
	}

	if len(appErr.Suggestions) != 2 {
		t.Errorf("Suggestions length = %v, want %v", len(appErr.Suggestions), 2)
	}

	if appErr.ExitCode != 5 {
		t.Errorf("Exit code = %v, want %v", appErr.ExitCode, 5)
	}

	if appErr.OriginalErr != originalErr {
		t.Errorf("Original error = %v, want %v", appErr.OriginalErr, originalErr)
	}
}

// Benchmark tests for error operations
func BenchmarkNewError(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewError(ErrorTypeValidation, "validation failed")
	}
}

func BenchmarkWrapError(b *testing.B) {
	originalErr := errors.New("original error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		WrapError(originalErr, ErrorTypeFile, "file operation failed")
	}
}

func BenchmarkAppError_Error(b *testing.B) {
	appErr := AppError{
		Type:        ErrorTypeValidation,
		Message:     "validation failed",
		Context:     "user input",
		OriginalErr: errors.New("original error"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = appErr.Error()
	}
}

func BenchmarkIsValidationError(b *testing.B) {
	err := NewValidationError("validation failed")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidationError(err)
	}
}
