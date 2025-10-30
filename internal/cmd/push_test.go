package cmd

import (
	"strings"
	"testing"
)

func TestGenerateDescription(t *testing.T) {
	tests := []struct {
		name     string
		envFile  string
		contains []string
	}{
		{
			name:     "default env file",
			envFile:  ".env",
			contains: []string{"environment variables"},
		},
		{
			name:     "production env file",
			envFile:  ".env.production",
			contains: []string{"production", "environment variables"},
		},
		{
			name:     "development env file",
			envFile:  ".env.development",
			contains: []string{"development", "environment variables"},
		},
		{
			name:     "staging env file",
			envFile:  ".env.staging",
			contains: []string{"staging", "environment variables"},
		},
		{
			name:     "test env file",
			envFile:  ".env.test",
			contains: []string{"test", "environment variables"},
		},
		{
			name:     "custom env file",
			envFile:  ".env.custom",
			contains: []string{"custom", "environment variables"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateDescription(tt.envFile)
			
			if result == "" {
				t.Error("Description should not be empty")
			}

			// Check for expected content
			for _, expected := range tt.contains {
				if !strings.Contains(strings.ToLower(result), strings.ToLower(expected)) {
					t.Errorf("Expected description to contain %q, got: %s", expected, result)
				}
			}

			// Verify it contains "envi" attribution
			if !strings.Contains(strings.ToLower(result), "envi") {
				t.Error("Description should contain 'envi' attribution")
			}
		})
	}
}

func TestGetProjectName(t *testing.T) {
	// This function tries to get project name from git or directory
	result := getProjectName()
	
	// We can't test exact value since it depends on environment,
	// but we can verify it returns a string
	if result == "" {
		t.Log("Project name is empty (expected if not in a git repo)")
	} else {
		// Verify it's a reasonable project name
		if len(result) > 200 {
			t.Error("Project name seems unreasonably long")
		}
		
		// Project name should not contain path separators
		if strings.Contains(result, "/") || strings.Contains(result, "\\") {
			t.Error("Project name should not contain path separators")
		}
	}
}

func TestCreateReadmeContent(t *testing.T) {
	tests := []struct {
		name              string
		fullEncryption    bool
		maskedEncryption  bool
		expectedContains  []string
		unexpectedContains []string
	}{
		{
			name:             "unencrypted",
			fullEncryption:   false,
			maskedEncryption: false,
			expectedContains: []string{
				"Environment Variables",
				"envi",
				"Install envi",
			},
			unexpectedContains: []string{
				"encrypted",
				"password",
				"decrypt",
			},
		},
		{
			name:             "full encryption",
			fullEncryption:   true,
			maskedEncryption: false,
			expectedContains: []string{
				"Environment Variables",
				"encrypted",
				"envi pull",
				"decrypt",
				"password",
			},
			unexpectedContains: []string{},
		},
		{
			name:             "masked encryption",
			fullEncryption:   false,
			maskedEncryption: true,
			expectedContains: []string{
				"Environment Variables",
				"masked",
				"encrypted",
				"envi pull",
				"--unmask",
			},
			unexpectedContains: []string{},
		},
		{
			name:             "both flags true - full encryption takes precedence",
			fullEncryption:   true,
			maskedEncryption: true,
			expectedContains: []string{
				"Environment Variables",
				"encrypted",
				"envi pull",
			},
			unexpectedContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createReadmeContent(tt.fullEncryption, tt.maskedEncryption)
			
			if result == "" {
				t.Error("README content should not be empty")
			}

			// Check for expected content
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected README to contain %q, but it didn't.\nGot: %s", expected, result)
				}
			}

			// Check for unexpected content
			for _, unexpected := range tt.unexpectedContains {
				if strings.Contains(result, unexpected) {
					t.Errorf("Expected README to NOT contain %q, but it did.\nGot: %s", unexpected, result)
				}
			}

			// Verify markdown header
			if !strings.HasPrefix(result, "#") {
				t.Error("README should start with a markdown header")
			}

			// Verify attribution
			if !strings.Contains(result, "github.com/dexterity-inc/envi") {
				t.Error("README should contain attribution link")
			}
		})
	}
}

func TestCreateReadmeContentStructure(t *testing.T) {
	readme := createReadmeContent(false, false)
	
	// Test that README has reasonable structure
	lines := strings.Split(readme, "\n")
	if len(lines) < 5 {
		t.Error("README should have multiple lines of content")
	}

	// Test for markdown formatting
	headerCount := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			headerCount++
		}
	}

	if headerCount == 0 {
		t.Error("README should contain at least one markdown header")
	}
}

func TestGenerateDescriptionWithProjectName(t *testing.T) {
	// Test that description contains project name if available
	description := generateDescription(".env")
	
	// Should contain meaningful content
	if len(description) < 10 {
		t.Error("Description seems too short to be useful")
	}

	// Should not contain template placeholders
	if strings.Contains(description, "{{") || strings.Contains(description, "}}") {
		t.Error("Description should not contain template placeholders")
	}
}

func TestCreateReadmeContentMarkdown(t *testing.T) {
	tests := []struct {
		name             string
		fullEncryption   bool
		maskedEncryption bool
	}{
		{"unencrypted", false, false},
		{"full encryption", true, false},
		{"masked encryption", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readme := createReadmeContent(tt.fullEncryption, tt.maskedEncryption)
			
			// Verify markdown structure
			if !strings.Contains(readme, "##") {
				t.Error("README should contain subheadings (##)")
			}

			// Verify code blocks if instructions are present
			if strings.Contains(readme, "envi pull") {
				if !strings.Contains(readme, "```") {
					t.Error("README with commands should contain code blocks")
				}
			}
		})
	}
}

func TestGenerateDescriptionConsistency(t *testing.T) {
	// Test that generateDescription is consistent
	desc1 := generateDescription(".env.production")
	desc2 := generateDescription(".env.production")
	
	// Should generate similar descriptions (though project name might differ in edge cases)
	if !strings.Contains(desc1, "production") || !strings.Contains(desc2, "production") {
		t.Error("Description should be consistent for same input")
	}
}

func TestCreateReadmeContentLength(t *testing.T) {
	tests := []struct {
		name             string
		fullEncryption   bool
		maskedEncryption bool
		minLength        int
	}{
		{"unencrypted", false, false, 100},
		{"full encryption", true, false, 200},
		{"masked encryption", false, true, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			readme := createReadmeContent(tt.fullEncryption, tt.maskedEncryption)
			
			if len(readme) < tt.minLength {
				t.Errorf("README too short: got %d chars, expected at least %d", len(readme), tt.minLength)
			}
		})
	}
}
