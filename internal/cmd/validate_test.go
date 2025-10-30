package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseEnvFile(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedVars    map[string]string
		expectedComments []string
		expectError     bool
	}{
		{
			name: "basic env file",
			content: `# This is a comment
DB_HOST=localhost
DB_PORT=5432
API_KEY=secret123`,
			expectedVars: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
				"API_KEY": "secret123",
			},
			expectedComments: []string{"# This is a comment"},
			expectError:      false,
		},
		{
			name: "env file with empty lines",
			content: `
DB_HOST=localhost

DB_PORT=5432

`,
			expectedVars: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
			},
			expectedComments: []string{},
			expectError:      false,
		},
		{
			name: "env file with quoted values",
			content: `API_KEY="my secret key"
DB_NAME='production_db'
PATH="/usr/local/bin"`,
			expectedVars: map[string]string{
				"API_KEY": "my secret key",
				"DB_NAME": "production_db",
				"PATH":    "/usr/local/bin",
			},
			expectedComments: []string{},
			expectError:      false,
		},
		{
			name: "env file with empty values",
			content: `EMPTY_VAR=
ANOTHER_VAR=value`,
			expectedVars: map[string]string{
				"EMPTY_VAR":   "",
				"ANOTHER_VAR": "value",
			},
			expectedComments: []string{},
			expectError:      false,
		},
		{
			name: "env file with multiple comments",
			content: `# Database configuration
# DO NOT COMMIT THIS FILE
DB_HOST=localhost
# API keys
API_KEY=secret`,
			expectedVars: map[string]string{
				"DB_HOST": "localhost",
				"API_KEY": "secret",
			},
			expectedComments: []string{
				"# Database configuration",
				"# DO NOT COMMIT THIS FILE",
				"# API keys",
			},
			expectError: false,
		},
		{
			name: "env file with complex values",
			content: `URL=https://example.com:8080/api/v1
CONNECTION_STRING=postgresql://user:pass@localhost:5432/db?sslmode=disable
JSON_CONFIG={"key":"value","nested":{"data":123}}`,
			expectedVars: map[string]string{
				"URL":               "https://example.com:8080/api/v1",
				"CONNECTION_STRING": "postgresql://user:pass@localhost:5432/db?sslmode=disable",
				"JSON_CONFIG":       `{"key":"value","nested":{"data":123}}`,
			},
			expectedComments: []string{},
			expectError:      false,
		},
		{
			name:             "empty file",
			content:          "",
			expectedVars:     map[string]string{},
			expectedComments: []string{},
			expectError:      false,
		},
		{
			name: "only comments",
			content: `# Comment 1
# Comment 2
# Comment 3`,
			expectedVars:     map[string]string{},
			expectedComments: []string{"# Comment 1", "# Comment 2", "# Comment 3"},
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile := filepath.Join(t.TempDir(), ".env.test")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0600)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			// Parse the file
			vars, comments, err := parseEnvFile(tmpFile)

			// Check error
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Check variables
			if len(vars) != len(tt.expectedVars) {
				t.Errorf("Expected %d variables, got %d", len(tt.expectedVars), len(vars))
			}
			for key, expectedValue := range tt.expectedVars {
				if actualValue, exists := vars[key]; !exists {
					t.Errorf("Expected variable %s not found", key)
				} else if actualValue != expectedValue {
					t.Errorf("Variable %s: expected %q, got %q", key, expectedValue, actualValue)
				}
			}

			// Check comments
			if len(comments) != len(tt.expectedComments) {
				t.Errorf("Expected %d comments, got %d", len(tt.expectedComments), len(comments))
			}
		})
	}
}

func TestParseEnvFileNonExistent(t *testing.T) {
	_, _, err := parseEnvFile("/nonexistent/file/path/.env")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestCheckStrictAndRequired(t *testing.T) {
	tests := []struct {
		name            string
		vars            map[string]string
		strict          bool
		required        []string
		expectStrictErr bool
		expectReqErr    bool
	}{
		{
			name: "all valid with strict",
			vars: map[string]string{
				"DB_HOST": "localhost",
				"API_KEY": "secret",
			},
			strict:          true,
			required:        []string{"DB_HOST", "API_KEY"},
			expectStrictErr: false,
			expectReqErr:    false,
		},
		{
			name: "empty values with strict",
			vars: map[string]string{
				"DB_HOST": "",
				"API_KEY": "secret",
			},
			strict:          true,
			required:        []string{},
			expectStrictErr: true,
			expectReqErr:    false,
		},
		{
			name: "missing required variables",
			vars: map[string]string{
				"DB_HOST": "localhost",
			},
			strict:          false,
			required:        []string{"DB_HOST", "API_KEY", "SECRET"},
			expectStrictErr: false,
			expectReqErr:    true,
		},
		{
			name: "no validation",
			vars: map[string]string{
				"DB_HOST": "localhost",
			},
			strict:          false,
			required:        []string{},
			expectStrictErr: false,
			expectReqErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global flags
			originalStrict := validateStrict
			originalRequired := validateRequired
			defer func() {
				validateStrict = originalStrict
				validateRequired = originalRequired
			}()

			validateStrict = tt.strict
			validateRequired = tt.required

			// This function logs errors but doesn't return them
			// We're just testing it doesn't panic
			checkStrictAndRequired(tt.vars)
		})
	}
}

func TestAddMissingVars(t *testing.T) {
	tests := []struct {
		name        string
		initial     string
		missing     map[string]string
		current     map[string]string
		comments    []string
		expectError bool
	}{
		{
			name: "add to existing file",
			initial: `DB_HOST=localhost
DB_PORT=5432`,
			missing: map[string]string{
				"API_KEY": "secret123",
				"SECRET":  "value",
			},
			current: map[string]string{
				"DB_HOST": "localhost",
				"DB_PORT": "5432",
			},
			comments:    []string{},
			expectError: false,
		},
		{
			name:    "add to empty file",
			initial: "",
			missing: map[string]string{
				"NEW_VAR": "new_value",
			},
			current:     map[string]string{},
			comments:    []string{},
			expectError: false,
		},
		{
			name: "add multiple variables",
			initial: `# Existing config
VAR1=value1`,
			missing: map[string]string{
				"VAR2": "value2",
				"VAR3": "value3",
				"VAR4": "value4",
			},
			current: map[string]string{
				"VAR1": "value1",
			},
			comments:    []string{"# Existing config"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile := filepath.Join(t.TempDir(), ".env.test")
			err := os.WriteFile(tmpFile, []byte(tt.initial), 0600)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			// Add missing variables
			err = addMissingVars(tmpFile, tt.missing, tt.current, tt.comments)

			// Check error
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// Read the file and verify missing vars were added
				content, err := os.ReadFile(tmpFile)
				if err != nil {
					t.Fatalf("Failed to read file: %v", err)
				}

				contentStr := string(content)

				// Verify original content is preserved
				if tt.initial != "" && !strings.Contains(contentStr, strings.TrimSpace(tt.initial)) {
					t.Error("Original content not preserved")
				}

				// Verify missing vars were added
				for key, value := range tt.missing {
					expected := key + "=" + value
					if !strings.Contains(contentStr, expected) {
						t.Errorf("Missing variable %s=%s not found in file", key, value)
					}
				}

				// Verify comment was added
				if len(tt.missing) > 0 && !strings.Contains(contentStr, "# Added by envi validate --fix") {
					t.Error("Expected comment not found in file")
				}
			}
		})
	}
}

func TestAddMissingVarsNonExistent(t *testing.T) {
	err := addMissingVars("/nonexistent/path/.env", map[string]string{"VAR": "value"}, map[string]string{}, []string{})
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestParseEnvFileWithVariousFormats(t *testing.T) {
	content := `# Comment at top
SIMPLE_VAR=value
QUOTED_DOUBLE="double quoted value"
QUOTED_SINGLE='single quoted value'
EMPTY_VAR=
VAR_WITH_EQUALS=value=with=equals
VAR_WITH_SPACES=  spaces around  
# Comment in middle
LAST_VAR=last`

	tmpFile := filepath.Join(t.TempDir(), ".env.test")
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	vars, comments, err := parseEnvFile(tmpFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test specific cases
	if vars["SIMPLE_VAR"] != "value" {
		t.Errorf("SIMPLE_VAR: expected 'value', got %q", vars["SIMPLE_VAR"])
	}

	if vars["QUOTED_DOUBLE"] != "double quoted value" {
		t.Errorf("QUOTED_DOUBLE: expected 'double quoted value', got %q", vars["QUOTED_DOUBLE"])
	}

	if vars["QUOTED_SINGLE"] != "single quoted value" {
		t.Errorf("QUOTED_SINGLE: expected 'single quoted value', got %q", vars["QUOTED_SINGLE"])
	}

	if vars["EMPTY_VAR"] != "" {
		t.Errorf("EMPTY_VAR: expected empty string, got %q", vars["EMPTY_VAR"])
	}

	if vars["VAR_WITH_EQUALS"] != "value=with=equals" {
		t.Errorf("VAR_WITH_EQUALS: expected 'value=with=equals', got %q", vars["VAR_WITH_EQUALS"])
	}

	if len(comments) != 2 {
		t.Errorf("Expected 2 comments, got %d", len(comments))
	}
}
