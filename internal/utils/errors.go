package utils

import (
	"fmt"
	"os"
	"strings"
)

// ErrorType represents the type of error
type ErrorType int

const (
	ErrorTypeValidation ErrorType = iota
	ErrorTypeConfig
	ErrorTypeFile
	ErrorTypeNetwork
	ErrorTypeEncryption
	ErrorTypeGitHub
	ErrorTypePermission
	ErrorTypeInput
	ErrorTypeSystem
)

// AppError represents an application error with context
type AppError struct {
	Type        ErrorType
	Message     string
	Context     string
	OriginalErr error
	Suggestions []string
	ExitCode    int
}

func (e AppError) Error() string {
	var parts []string

	if e.Context != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Context))
	}

	if e.Message != "" {
		parts = append(parts, e.Message)
	}

	if e.OriginalErr != nil {
		parts = append(parts, fmt.Sprintf("(%v)", e.OriginalErr))
	}

	return strings.Join(parts, " ")
}

// Unwrap returns the original error
func (e AppError) Unwrap() error {
	return e.OriginalErr
}

// WithContext adds context to an error
func (e AppError) WithContext(context string) AppError {
	e.Context = context
	return e
}

// WithSuggestions adds suggestions to an error
func (e AppError) WithSuggestions(suggestions ...string) AppError {
	e.Suggestions = suggestions
	return e
}

// WithExitCode sets the exit code for the error
func (e AppError) WithExitCode(exitCode int) AppError {
	e.ExitCode = exitCode
	return e
}

// NewError creates a new application error
func NewError(errorType ErrorType, message string) AppError {
	return AppError{
		Type:     errorType,
		Message:  message,
		ExitCode: 1,
	}
}

// WrapError wraps an existing error with additional context
func WrapError(err error, errorType ErrorType, message string) AppError {
	return AppError{
		Type:        errorType,
		Message:     message,
		OriginalErr: err,
		ExitCode:    1,
	}
}

// Error constructors for common error types

// NewValidationError creates a validation error
func NewValidationError(message string) AppError {
	return NewError(ErrorTypeValidation, message)
}

// NewConfigError creates a configuration error
func NewConfigError(message string) AppError {
	return NewError(ErrorTypeConfig, message)
}

// NewFileError creates a file operation error
func NewFileError(message string) AppError {
	return NewError(ErrorTypeFile, message)
}

// NewNetworkError creates a network error
func NewNetworkError(message string) AppError {
	return NewError(ErrorTypeNetwork, message)
}

// NewEncryptionError creates an encryption error
func NewEncryptionError(message string) AppError {
	return NewError(ErrorTypeEncryption, message)
}

// NewGitHubError creates a GitHub API error
func NewGitHubError(message string) AppError {
	return NewError(ErrorTypeGitHub, message)
}

// NewPermissionError creates a permission error
func NewPermissionError(message string) AppError {
	return NewError(ErrorTypePermission, message)
}

// NewInputError creates an input error
func NewInputError(message string) AppError {
	return NewError(ErrorTypeInput, message)
}

// NewSystemError creates a system error
func NewSystemError(message string) AppError {
	return NewError(ErrorTypeSystem, message)
}

// WrapValidationError wraps an error as a validation error
func WrapValidationError(err error, message string) AppError {
	return WrapError(err, ErrorTypeValidation, message)
}

// WrapConfigError wraps an error as a configuration error
func WrapConfigError(err error, message string) AppError {
	return WrapError(err, ErrorTypeConfig, message)
}

// WrapFileError wraps an error as a file error
func WrapFileError(err error, message string) AppError {
	return WrapError(err, ErrorTypeFile, message)
}

// WrapNetworkError wraps an error as a network error
func WrapNetworkError(err error, message string) AppError {
	return WrapError(err, ErrorTypeNetwork, message)
}

// WrapEncryptionError wraps an error as an encryption error
func WrapEncryptionError(err error, message string) AppError {
	return WrapError(err, ErrorTypeEncryption, message)
}

// WrapGitHubError wraps an error as a GitHub error
func WrapGitHubError(err error, message string) AppError {
	return WrapError(err, ErrorTypeGitHub, message)
}

// WrapPermissionError wraps an error as a permission error
func WrapPermissionError(err error, message string) AppError {
	return WrapError(err, ErrorTypePermission, message)
}

// WrapInputError wraps an error as an input error
func WrapInputError(err error, message string) AppError {
	return WrapError(err, ErrorTypeInput, message)
}

// WrapSystemError wraps an error as a system error
func WrapSystemError(err error, message string) AppError {
	return WrapError(err, ErrorTypeSystem, message)
}

// Error handling functions

// HandleError handles an error with appropriate logging and exit
func HandleError(err error, context string) {
	if err == nil {
		return
	}

	// Check if it's an AppError
	if appErr, ok := err.(AppError); ok {
		handleAppError(appErr)
		return
	}

	// Convert to AppError
	appErr := WrapSystemError(err, "unexpected error")
	appErr.Context = context
	handleAppError(appErr)
}

// handleAppError handles an AppError specifically
func handleAppError(err AppError) {
	logger := GetLogger()

	// Log the error with appropriate level
	switch err.Type {
	case ErrorTypeValidation:
		logger.Error("Validation error: %s", err.Error())
	case ErrorTypeConfig:
		logger.Error("Configuration error: %s", err.Error())
	case ErrorTypeFile:
		logger.Error("File error: %s", err.Error())
	case ErrorTypeNetwork:
		logger.Error("Network error: %s", err.Error())
	case ErrorTypeEncryption:
		logger.Error("Encryption error: %s", err.Error())
	case ErrorTypeGitHub:
		logger.Error("GitHub API error: %s", err.Error())
	case ErrorTypePermission:
		logger.Error("Permission error: %s", err.Error())
	case ErrorTypeInput:
		logger.Error("Input error: %s", err.Error())
	case ErrorTypeSystem:
		logger.Error("System error: %s", err.Error())
	}

	// Show suggestions if available
	if len(err.Suggestions) > 0 {
		logger.Info("Suggestions:")
		for _, suggestion := range err.Suggestions {
			logger.Info("  • %s", suggestion)
		}
	}

	// Exit with appropriate code
	os.Exit(err.ExitCode)
}

// FatalError logs a fatal error and exits
func FatalError(err error, context string) {
	HandleError(err, context)
}

// FatalMessage logs a fatal message and exits
func FatalMessage(message string, context string) {
	err := NewSystemError(message).WithContext(context)
	HandleError(err, context)
}

// Error utilities

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	if appErr, ok := err.(AppError); ok {
		return appErr.Type == ErrorTypeValidation
	}
	return false
}

// IsConfigError checks if an error is a configuration error
func IsConfigError(err error) bool {
	if appErr, ok := err.(AppError); ok {
		return appErr.Type == ErrorTypeConfig
	}
	return false
}

// IsFileError checks if an error is a file error
func IsFileError(err error) bool {
	if appErr, ok := err.(AppError); ok {
		return appErr.Type == ErrorTypeFile
	}
	return false
}

// IsNetworkError checks if an error is a network error
func IsNetworkError(err error) bool {
	if appErr, ok := err.(AppError); ok {
		return appErr.Type == ErrorTypeNetwork
	}
	return false
}

// IsEncryptionError checks if an error is an encryption error
func IsEncryptionError(err error) bool {
	if appErr, ok := err.(AppError); ok {
		return appErr.Type == ErrorTypeEncryption
	}
	return false
}

// IsGitHubError checks if an error is a GitHub error
func IsGitHubError(err error) bool {
	if appErr, ok := err.(AppError); ok {
		return appErr.Type == ErrorTypeGitHub
	}
	return false
}

// IsPermissionError checks if an error is a permission error
func IsPermissionError(err error) bool {
	if appErr, ok := err.(AppError); ok {
		return appErr.Type == ErrorTypePermission
	}
	return false
}

// IsInputError checks if an error is an input error
func IsInputError(err error) bool {
	if appErr, ok := err.(AppError); ok {
		return appErr.Type == ErrorTypeInput
	}
	return false
}

// IsSystemError checks if an error is a system error
func IsSystemError(err error) bool {
	if appErr, ok := err.(AppError); ok {
		return appErr.Type == ErrorTypeSystem
	}
	return false
}

// Error type to string conversion
func (et ErrorType) String() string {
	switch et {
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeConfig:
		return "configuration"
	case ErrorTypeFile:
		return "file"
	case ErrorTypeNetwork:
		return "network"
	case ErrorTypeEncryption:
		return "encryption"
	case ErrorTypeGitHub:
		return "github"
	case ErrorTypePermission:
		return "permission"
	case ErrorTypeInput:
		return "input"
	case ErrorTypeSystem:
		return "system"
	default:
		return "unknown"
	}
}

// Common error messages
var (
	ErrMsgFileNotFound     = "file not found"
	ErrMsgPermissionDenied = "permission denied"
	ErrMsgInvalidInput     = "invalid input"
	ErrMsgNetworkTimeout   = "network timeout"
	ErrMsgConfigInvalid    = "invalid configuration"
	ErrMsgTokenInvalid     = "invalid token"
	ErrMsgEncryptionFailed = "encryption failed"
	ErrMsgDecryptionFailed = "decryption failed"
	ErrMsgGitHubAPIError   = "GitHub API error"
	ErrMsgValidationFailed = "validation failed"
)
