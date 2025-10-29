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

	cleanPath := filepath.Clean(path)

	// Check for absolute paths (Unix-style)
	if filepath.IsAbs(cleanPath) {
		return "", ErrAbsolutePath
	}

	// Check for Windows-style absolute paths (even on non-Windows systems)
	if len(cleanPath) >= 3 && cleanPath[1] == ':' && (cleanPath[2] == '\\' || cleanPath[2] == '/') {
		return "", ErrAbsolutePath
	}

	// Check for directory traversal attempts
	if strings.HasPrefix(cleanPath, "..") || strings.Contains(cleanPath, string(filepath.Separator)+"..") {
		return "", ErrPathTraversal
	}

	// Check for paths that would escape the current directory
	if strings.HasPrefix(cleanPath, "/") || strings.HasPrefix(cleanPath, "\\") {
		return "", ErrPathTraversal
	}

	return cleanPath, nil
}

func checkDangerousPath(path string) error {
	dangerousPaths := []string{
		"/etc/", "/bin/", "/sbin/", "/usr/bin/", "/usr/sbin/",
		"/System/", "/Library/", "/var/", "/tmp/", "/boot/",
		"/proc/", "/sys/",
	}

	for _, dangerous := range dangerousPaths {
		if strings.HasPrefix(path, dangerous) {
			return fmt.Errorf("access to system directory %s is not allowed", dangerous)
		}
	}
	return nil
}

func ValidateOutputPath(path string) error {
	sanitized, err := SanitizeFilePath(path)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	return checkDangerousPath(sanitized)
}

// ValidateInputPath validates a user-provided input file path
func ValidateInputPath(path string) error {
	sanitized, err := SanitizeFilePath(path)
	if err != nil {
		return fmt.Errorf("invalid input path: %w", err)
	}

	return checkDangerousPath(sanitized)
}

// ValidateKeyFilePath validates a key file path
// Security note: Key files should not be in publicly accessible or temporary locations
func ValidateKeyFilePath(path string) error {
	sanitized, err := SanitizeFilePath(path)
	if err != nil {
		return fmt.Errorf("invalid key file path: %w", err)
	}

	if err := checkDangerousPath(sanitized); err != nil {
		return err
	}

	// Additional check: key files shouldn't be in temp directories
	lowerPath := strings.ToLower(sanitized)
	if strings.Contains(lowerPath, "/tmp/") || strings.Contains(lowerPath, "\\tmp\\") {
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

	envVarRegex := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)
	if !envVarRegex.MatchString(name) {
		return ErrInvalidEnvVarName
	}

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

	for _, r := range value {
		if unicode.IsControl(r) && r != '\t' && r != '\n' && r != '\r' {
			return fmt.Errorf("environment variable value contains invalid control character: %U", r)
		}
	}

	return nil
}

// ValidateEnvLine validates a complete environment variable line (KEY=VALUE format)
func ValidateEnvLine(line string) error {
	line = strings.TrimSpace(line)

	if line == "" || strings.HasPrefix(line, "#") {
		return nil
	}

	if !strings.Contains(line, "=") {
		return errors.New("invalid environment variable format: missing '=' separator")
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return errors.New("invalid environment variable format")
	}

	name := strings.TrimSpace(parts[0])
	value := parts[1]

	if err := ValidateEnvVarName(name); err != nil {
		return fmt.Errorf("invalid variable name '%s': %w", name, err)
	}

	if err := ValidateEnvVarValue(value); err != nil {
		return fmt.Errorf("invalid variable value for '%s': %w", name, err)
	}

	return nil
}