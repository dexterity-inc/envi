package utils

import (
	"fmt"
	"log"
	"os"
)

// Time format constants
const (
	TimeFormatShort = "2006-01-02 15:04:05"
)

// File permission constants
const (
	// EnvFilePerms is the recommended secure file permissions for .env files (rw-------)
	EnvFilePerms = 0600
)

// Logger provides logging functionality for the envi CLI
type Logger struct {
	useTUI bool
}

var defaultLogger *Logger

// InitLogger initializes the logger with TUI setting
func InitLogger(useTUI bool) {
	defaultLogger = &Logger{
		useTUI: useTUI,
	}
}

// Info logs an info message
func Info(format string, args ...interface{}) {
	if defaultLogger == nil {
		InitLogger(true)
	}
	message := fmt.Sprintf(format, args...)
	if defaultLogger.useTUI {
		fmt.Println("ℹ️ " + message)
	} else {
		log.Printf("INFO: %s", message)
	}
}

// Success logs a success message
func Success(format string, args ...interface{}) {
	if defaultLogger == nil {
		InitLogger(true)
	}
	message := fmt.Sprintf(format, args...)
	if defaultLogger.useTUI {
		fmt.Println("✅ " + message)
	} else {
		log.Printf("SUCCESS: %s", message)
	}
}

// Error logs an error message
func Error(format string, args ...interface{}) {
	if defaultLogger == nil {
		InitLogger(true)
	}
	message := fmt.Sprintf(format, args...)
	if defaultLogger.useTUI {
		fmt.Fprintf(os.Stderr, "❌ %s\n", message)
	} else {
		log.Printf("ERROR: %s", message)
	}
}

// Warn logs a warning message
func Warn(format string, args ...interface{}) {
	if defaultLogger == nil {
		InitLogger(true)
	}
	message := fmt.Sprintf(format, args...)
	if defaultLogger.useTUI {
		fmt.Println("⚠️ " + message)
	} else {
		log.Printf("WARN: %s", message)
	}
}

// Fatal logs a fatal error and exits
func Fatal(format string, args ...interface{}) {
	if defaultLogger == nil {
		InitLogger(true)
	}
	message := fmt.Sprintf(format, args...)
	if defaultLogger.useTUI {
		fmt.Fprintf(os.Stderr, "💥 FATAL: %s\n", message)
	} else {
		log.Fatalf("FATAL: %s", message)
	}
	os.Exit(1)
}

// FatalError logs a fatal error with error details and exits
func FatalError(err error, context string) {
	if err != nil {
		Fatal("%s: %v", context, err)
	}
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		InitLogger(true)
	}
	return defaultLogger
}

// Instance methods for Logger to keep backwards compatibility where
// callers expect a logger instance with Info/Warn/Error/Success methods.
func (l *Logger) Info(format string, args ...interface{}) {
	Info(format, args...)
}

func (l *Logger) Success(format string, args ...interface{}) {
	Success(format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	Error(format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	Warn(format, args...)
}

// WrapFileError wraps a file-related error with context (nil-safe)
func WrapFileError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// WrapEncryptionError wraps encryption related errors with context (nil-safe)
func WrapEncryptionError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// NewInputError creates a typed input error
func NewInputError(msg string) error {
	return fmt.Errorf("input error: %s", msg)
}

// NewValidationError creates a typed validation error
func NewValidationError(msg string) error {
	return fmt.Errorf("validation error: %s", msg)
}

// FatalMessage logs a fatal message and exits
func FatalMessage(format string, args ...interface{}) {
	Fatal(format, args...)
}

// HandleError handles an error by logging it
func HandleError(err error, message string) {
	if err != nil {
		Error("%s: %v", message, err)
	}
}

// WrapInputError wraps an error with additional context
func WrapInputError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// WrapGitHubError wraps GitHub API related errors with context (nil-safe)
func WrapGitHubError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// DefaultKeyFile returns the default key file path
func DefaultKeyFile() string {
	return "envi.key"
}

// Confirm prompts for user confirmation
func Confirm(title, message string) (bool, error) {
	fmt.Printf("%s\n%s (y/N): ", title, message)
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false, err
	}
	return response == "y" || response == "Y" || response == "yes" || response == "Yes", nil
}
