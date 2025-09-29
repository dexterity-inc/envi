package security

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// ErrPathTraversal is returned when a path contains directory traversal sequences
var ErrPathTraversal = errors.New("path contains directory traversal sequences")

// ErrAbsolutePath is returned when an absolute path is provided where relative is expected
var ErrAbsolutePath = errors.New("absolute paths are not allowed")

// ErrEmptyPath is returned when an empty path is provided
var ErrEmptyPath = errors.New("path cannot be empty")

// SanitizeFilePath validates and sanitizes a file path to prevent directory traversal attacks
// It ensures the path is relative, doesn't contain traversal sequences, and is within allowed bounds
func SanitizeFilePath(path string) (string, error) {
	if path == "" {
		return "", ErrEmptyPath
	}

	// Clean the path to resolve any . and .. components
	cleaned := filepath.Clean(path)

	// Check for absolute paths (Unix-style)
	if filepath.IsAbs(cleaned) {
		return "", ErrAbsolutePath
	}

	// Check for Windows-style absolute paths (even on non-Windows systems)
	if len(cleaned) >= 3 && cleaned[1] == ':' && (cleaned[2] == '\\' || cleaned[2] == '/') {
		return "", ErrAbsolutePath
	}

	// Check for directory traversal attempts
	if strings.Contains(cleaned, "..") {
		return "", ErrPathTraversal
	}

	// Check for paths that would escape the current directory
	if strings.HasPrefix(cleaned, "/") || strings.HasPrefix(cleaned, "\\") {
		return "", ErrPathTraversal
	}

	// Additional check: ensure the cleaned path doesn't start with ..
	if strings.HasPrefix(cleaned, "..") {
		return "", ErrPathTraversal
	}

	return cleaned, nil
}

// ValidateOutputPath validates a user-provided output file path
// It allows specific safe directories and prevents dangerous locations
func ValidateOutputPath(path string) error {
	sanitized, err := SanitizeFilePath(path)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Check for attempts to write to sensitive directories
	dangerousPaths := []string{
		"/etc/",
		"/bin/",
		"/sbin/",
		"/usr/bin/",
		"/usr/sbin/",
		"/System/",
		"/Library/",
		"/var/",
		"/tmp/",
		"/dev/",
		"/proc/",
		"/sys/",
	}

	// Convert to absolute path for checking
	absPath, err := filepath.Abs(sanitized)
	if err != nil {
		return fmt.Errorf("error resolving path: %w", err)
	}

	for _, dangerous := range dangerousPaths {
		if strings.HasPrefix(absPath, dangerous) {
			return fmt.Errorf("writing to system directory %s is not allowed", dangerous)
		}
	}

	return nil
}

// ValidateInputPath validates a user-provided input file path
// It ensures the path is safe to read from
func ValidateInputPath(path string) error {
	sanitized, err := SanitizeFilePath(path)
	if err != nil {
		return fmt.Errorf("invalid input path: %w", err)
	}

	// Additional validation can be added here if needed
	_ = sanitized

	return nil
}

// ValidateKeyFilePath validates a user-provided key file path
// It applies stricter validation for key files
func ValidateKeyFilePath(path string) error {
	sanitized, err := SanitizeFilePath(path)
	if err != nil {
		return fmt.Errorf("invalid key file path: %w", err)
	}

	// Key files should not be in certain directories for security
	if strings.Contains(sanitized, "/tmp/") || strings.Contains(sanitized, "\\tmp\\") {
		return errors.New("key files should not be stored in temporary directories")
	}

	return nil
}

// Environment variable validation errors
var (
	ErrInvalidEnvVarName  = errors.New("invalid environment variable name")
	ErrInvalidEnvVarValue = errors.New("invalid environment variable value")
	ErrEnvVarTooLong      = errors.New("environment variable too long")
	ErrEnvVarEmpty        = errors.New("environment variable cannot be empty")
)

// Environment variable constraints
const (
	MaxEnvVarNameLength  = 128  // Maximum length for variable names
	MaxEnvVarValueLength = 4096 // Maximum length for variable values
)

// ValidateEnvVarName validates an environment variable name
// Environment variable names should follow POSIX conventions:
// - Start with a letter or underscore
// - Contain only letters, digits, and underscores
// - Be in uppercase by convention (though not required)
func ValidateEnvVarName(name string) error {
	if name == "" {
		return ErrEnvVarEmpty
	}

	if len(name) > MaxEnvVarNameLength {
		return ErrEnvVarTooLong
	}

	// Check for valid POSIX environment variable name pattern
	envVarRegex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !envVarRegex.MatchString(name) {
		return ErrInvalidEnvVarName
	}

	// Check for potentially dangerous names (reserved or system variables)
	dangerousNames := []string{
		"PATH", "HOME", "USER", "SHELL", "PWD", "OLDPWD",
		"IFS", "PS1", "PS2", "PS3", "PS4",
		"TERM", "LANG", "LC_ALL", "TZ",
		"LD_PRELOAD", "LD_LIBRARY_PATH",
	}

	for _, dangerous := range dangerousNames {
		if name == dangerous {
			return fmt.Errorf("environment variable name '%s' conflicts with system variable", name)
		}
	}

	return nil
}

// ValidateEnvVarValue validates an environment variable value
// Checks for potentially dangerous content and length limits
func ValidateEnvVarValue(value string) error {
	if len(value) > MaxEnvVarValueLength {
		return ErrEnvVarTooLong
	}

	// Check for control characters (except tab, newline, carriage return)
	for _, r := range value {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return fmt.Errorf("environment variable value contains invalid control character: %U", r)
		}
	}

	// Check for potentially dangerous patterns
	dangerousPatterns := []string{
		"$(",    // Command substitution
		"`",     // Command substitution
		"${",    // Variable expansion
		";",     // Command separator
		"&&",    // Command chaining
		"||",    // Command chaining
		"|",     // Pipe
		">",     // Redirection
		"<",     // Redirection
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(value, pattern) {
			return fmt.Errorf("environment variable value contains potentially dangerous pattern: %s", pattern)
		}
	}

	return nil
}

// ValidateEnvLine validates a complete environment variable line (KEY=VALUE format)
func ValidateEnvLine(line string) error {
	line = strings.TrimSpace(line)
	
	// Skip empty lines and comments
	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	// Check for valid KEY=VALUE format
	if !strings.Contains(line, "=") {
		return errors.New("invalid environment variable format: missing '=' separator")
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return errors.New("invalid environment variable format")
	}

	name := strings.TrimSpace(parts[0])
	value := parts[1] // Don't trim value as it may have intentional whitespace

	// Validate name
	if err := ValidateEnvVarName(name); err != nil {
		return fmt.Errorf("invalid variable name '%s': %w", name, err)
	}

	// Validate value
	if err := ValidateEnvVarValue(value); err != nil {
		return fmt.Errorf("invalid variable value for '%s': %w", name, err)
	}

	return nil
}