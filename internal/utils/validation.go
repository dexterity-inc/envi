package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Pre-compiled regexes for better performance
var (
	uppercaseRegex = regexp.MustCompile(`[A-Z]`)
	lowercaseRegex = regexp.MustCompile(`[a-z]`)
	digitRegex     = regexp.MustCompile(`[0-9]`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s (value: %v)", e.Field, e.Message, e.Value)
}

// Validator provides validation functionality
type Validator struct {
	errors []ValidationError
}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{
		errors: make([]ValidationError, 0),
	}
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string, value interface{}) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	})
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// GetErrors returns all validation errors
func (v *Validator) GetErrors() []ValidationError {
	return v.errors
}

// Error returns a combined error message
func (v *Validator) Error() error {
	if !v.HasErrors() {
		return nil
	}

	var messages []string
	for _, err := range v.errors {
		messages = append(messages, err.Error())
	}
	return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
}

// Required validates that a field is not empty
func (v *Validator) Required(field string, value interface{}) bool {
	if value == nil {
		v.AddError(field, "field is required", value)
		return false
	}

	switch val := value.(type) {
	case string:
		if strings.TrimSpace(val) == "" {
			v.AddError(field, "field cannot be empty", value)
			return false
		}
	case []byte:
		if len(val) == 0 {
			v.AddError(field, "field cannot be empty", value)
			return false
		}
	case []string:
		if len(val) == 0 {
			v.AddError(field, "field cannot be empty", value)
			return false
		}
	}
	return true
}

// MinLength validates minimum length for strings
func (v *Validator) MinLength(field string, value string, min int) bool {
	if len(value) < min {
		v.AddError(field, fmt.Sprintf("minimum length is %d characters", min), value)
		return false
	}
	return true
}

// MaxLength validates maximum length for strings
func (v *Validator) MaxLength(field string, value string, max int) bool {
	if len(value) > max {
		v.AddError(field, fmt.Sprintf("maximum length is %d characters", max), value)
		return false
	}
	return true
}

// Pattern validates string against a regex pattern
func (v *Validator) Pattern(field, value, pattern string) bool {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil {
		v.AddError(field, "invalid pattern", pattern)
		return false
	}
	if !matched {
		v.AddError(field, "does not match required pattern", value)
		return false
	}
	return true
}

// FileExists validates that a file exists
func (v *Validator) FileExists(field, path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		v.AddError(field, "file does not exist", path)
		return false
	}
	return true
}

// FileReadable validates that a file is readable
func (v *Validator) FileReadable(field, path string) bool {
	if !v.FileExists(field, path) {
		return false
	}

	file, err := os.Open(path)
	if err != nil {
		v.AddError(field, "file is not readable", path)
		return false
	}
	defer file.Close()
	return true
}

// FileWritable validates that a file is writable
func (v *Validator) FileWritable(field, path string) bool {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Directory doesn't exist, check if we can create it
		if err := os.MkdirAll(dir, DirPerms); err != nil {
			v.AddError(field, "cannot create directory", dir)
			return false
		}
	}

	// Check if file exists and is writable
	if _, err := os.Stat(path); err == nil {
		file, err := os.OpenFile(path, os.O_WRONLY, 0)
		if err != nil {
			v.AddError(field, "file is not writable", path)
			return false
		}
		file.Close()
	}
	return true
}

// ValidGistID validates GitHub Gist ID format
func (v *Validator) ValidGistID(field, gistID string) bool {
	return v.Pattern(field, gistID, PatternGistID)
}

// ValidGitHubUser validates GitHub username format
func (v *Validator) ValidGitHubUser(field, username string) bool {
	return v.Pattern(field, username, PatternGitHubUser)
}

// ValidEnvVar validates environment variable name format
func (v *Validator) ValidEnvVar(field, envVar string) bool {
	return v.Pattern(field, envVar, PatternEnvVar)
}

// ValidFileName validates filename format
func (v *Validator) ValidFileName(field, filename string) bool {
	return v.Pattern(field, filename, PatternFileName)
}

// ValidURL validates URL format
func (v *Validator) ValidURL(field, url string) bool {
	// Basic URL validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		v.AddError(field, "must be a valid HTTP/HTTPS URL", url)
		return false
	}
	return true
}

// ValidEmail validates email format
func (v *Validator) ValidEmail(field, email string) bool {
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	return v.Pattern(field, email, emailPattern)
}

// ValidDate validates date format
func (v *Validator) ValidDate(field, date, format string) bool {
	_, err := time.Parse(format, date)
	if err != nil {
		v.AddError(field, fmt.Sprintf("must be in format: %s", format), date)
		return false
	}
	return true
}

// ValidInteger validates integer format
func (v *Validator) ValidInteger(field, value string) bool {
	if value == "" {
		v.AddError(field, "must be a valid integer", value)
		return false
	}

	// Check if it's a valid integer
	for _, char := range value {
		if char < '0' || char > '9' {
			v.AddError(field, "must contain only digits", value)
			return false
		}
	}
	return true
}

// ValidFloat validates float format
func (v *Validator) ValidFloat(field, value string) bool {
	if value == "" {
		v.AddError(field, "must be a valid number", value)
		return false
	}

	// Basic float validation
	dotCount := 0
	for i, char := range value {
		if char == '.' {
			dotCount++
			if dotCount > 1 {
				v.AddError(field, "must contain only one decimal point", value)
				return false
			}
		} else if char < '0' || char > '9' {
			if i == 0 && char == '-' {
				// Allow negative sign at the beginning
				continue
			}
			v.AddError(field, "must be a valid number", value)
			return false
		}
	}
	return true
}

// InRange validates that a number is within a range
func (v *Validator) InRange(field string, value, min, max int) bool {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("must be between %d and %d", min, max), value)
		return false
	}
	return true
}

// OneOf validates that a value is one of the allowed values
func (v *Validator) OneOf(field string, value interface{}, allowed []interface{}) bool {
	for _, allowedVal := range allowed {
		if value == allowedVal {
			return true
		}
	}
	v.AddError(field, fmt.Sprintf("must be one of: %v", allowed), value)
	return false
}

// ValidPassword validates password strength
func (v *Validator) ValidPassword(field, password string) bool {
	if len(password) < MinPasswordLen {
		v.AddError(field, fmt.Sprintf("minimum length is %d characters", MinPasswordLen), password)
		return false
	}

	// Check for at least one uppercase letter using pre-compiled regex
	if !uppercaseRegex.MatchString(password) {
		v.AddError(field, "must contain at least one uppercase letter", password)
		return false
	}

	// Check for at least one lowercase letter using pre-compiled regex
	if !lowercaseRegex.MatchString(password) {
		v.AddError(field, "must contain at least one lowercase letter", password)
		return false
	}

	// Check for at least one digit using pre-compiled regex
	if !digitRegex.MatchString(password) {
		v.AddError(field, "must contain at least one digit", password)
		return false
	}

	return true
}

// ValidFileSize validates file size
func (v *Validator) ValidFileSize(field, path string, maxSize int64) bool {
	info, err := os.Stat(path)
	if err != nil {
		v.AddError(field, "cannot get file size", path)
		return false
	}

	if info.Size() > maxSize {
		v.AddError(field, fmt.Sprintf("file size exceeds maximum of %d bytes", maxSize), info.Size())
		return false
	}
	return true
}

// ValidFileExtension validates file extension
func (v *Validator) ValidFileExtension(field, path string, allowedExtensions []string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, allowed := range allowedExtensions {
		if strings.ToLower(allowed) == ext {
			return true
		}
	}
	v.AddError(field, fmt.Sprintf("file extension must be one of: %v", allowedExtensions), ext)
	return false
}

// Convenience functions for common validations

// ValidateGistID validates a Gist ID
func ValidateGistID(gistID string) error {
	v := NewValidator()
	if !v.Required("gist_id", gistID) || !v.ValidGistID("gist_id", gistID) {
		return v.Error()
	}
	return nil
}

// ValidateGitHubToken validates a GitHub token
func ValidateGitHubToken(token string) error {
	v := NewValidator()
	if !v.Required("token", token) || !v.MinLength("token", token, 40) {
		return v.Error()
	}
	return nil
}

// ValidateEnvFile validates an environment file
func ValidateEnvFile(path string) error {
	v := NewValidator()
	if !v.Required("env_file", path) || !v.FileExists("env_file", path) || !v.FileReadable("env_file", path) {
		return v.Error()
	}
	return nil
}

// ValidateConfig validates configuration
func ValidateConfig(config interface{}) error {
	// This would be implemented based on the specific config structure
	// For now, return nil as a placeholder
	return nil
}
