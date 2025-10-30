package cmd

import (
	"strings"
	"testing"

	"github.com/google/go-github/v37/github"
	"github.com/dexterity-inc/envi/internal/encryption"
)

func TestCreateSharingReadmeContentMultipleScenarios(t *testing.T) {
	login := "testuser"
	user := &github.User{
		Login: &login,
	}

	tests := []struct {
		name                string
		recipientUsername   string
		keyFilePath         string
		useEncryption       bool
		useMaskedEncryption bool
		minLength           int
	}{
		{
			name:                "basic unencrypted",
			recipientUsername:   "recipient",
			keyFilePath:         "",
			useEncryption:       false,
			useMaskedEncryption: false,
			minLength:           200,
		},
		{
			name:                "with key file",
			recipientUsername:   "recipient",
			keyFilePath:         "/path/to/key.file",
			useEncryption:       true,
			useMaskedEncryption: false,
			minLength:           500,
		},
		{
			name:                "with password",
			recipientUsername:   "recipient",
			keyFilePath:         "",
			useEncryption:       true,
			useMaskedEncryption: false,
			minLength:           500,
		},
		{
			name:                "masked encryption",
			recipientUsername:   "recipient",
			keyFilePath:         "",
			useEncryption:       false,
			useMaskedEncryption: true,
			minLength:           500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set encryption state
			originalUseEncryption := encryption.UseEncryption
			originalUseMaskedEncryption := encryption.UseMaskedEncryption
			defer func() {
				encryption.UseEncryption = originalUseEncryption
				encryption.UseMaskedEncryption = originalUseMaskedEncryption
			}()

			encryption.UseEncryption = tt.useEncryption
			encryption.UseMaskedEncryption = tt.useMaskedEncryption

			result := createSharingReadmeContent(user, tt.recipientUsername, tt.keyFilePath)

			// Check length
			if len(result) < tt.minLength {
				t.Errorf("Content too short: got %d, expected at least %d", len(result), tt.minLength)
			}

			// Verify markdown structure
			if !strings.HasPrefix(result, "#") {
				t.Error("Should start with markdown header")
			}

			// Verify usernames are included
			if !strings.Contains(result, "testuser") {
				t.Error("Should contain sender username")
			}
			if !strings.Contains(result, tt.recipientUsername) {
				t.Error("Should contain recipient username")
			}

			// Verify attribution
			if !strings.Contains(result, "github.com/dexterity-inc/envi") {
				t.Error("Should contain attribution")
			}

			// Verify installation instructions
			if !strings.Contains(result, "brew install envi") {
				t.Error("Should contain brew installation")
			}
			if !strings.Contains(result, "scoop install envi") {
				t.Error("Should contain scoop installation")
			}
		})
	}
}

func TestCreateSharingReadmeContentInstructionsForKeyFile(t *testing.T) {
	login := "sender"
	user := &github.User{Login: &login}

	// Set encryption with key file
	originalUseEncryption := encryption.UseEncryption
	defer func() {
		encryption.UseEncryption = originalUseEncryption
	}()
	encryption.UseEncryption = true

	result := createSharingReadmeContent(user, "recipient", "/path/to/key.file")

	// Should mention key file
	if !strings.Contains(result, "key file") {
		t.Error("Should mention key file when key file path is provided")
	}

	// Should show correct command with key file
	if !strings.Contains(result, "--use-key-file") {
		t.Error("Should show --use-key-file flag")
	}
	if !strings.Contains(result, "--key-file") {
		t.Error("Should show --key-file flag")
	}
}

func TestCreateSharingReadmeContentInstructionsForPassword(t *testing.T) {
	login := "sender"
	user := &github.User{Login: &login}

	// Set encryption without key file
	originalUseEncryption := encryption.UseEncryption
	defer func() {
		encryption.UseEncryption = originalUseEncryption
	}()
	encryption.UseEncryption = true

	result := createSharingReadmeContent(user, "recipient", "")

	// Should mention password
	if !strings.Contains(result, "password") {
		t.Error("Should mention password when no key file is provided")
	}

	// Should not show key file flags
	if strings.Contains(result, "--use-key-file") {
		t.Error("Should not show --use-key-file flag without key file")
	}

	// Should mention password prompt
	if !strings.Contains(result, "prompted") {
		t.Error("Should mention password prompt")
	}
}

func TestCreateSharingReadmeContentMarkdownFormatting(t *testing.T) {
	login := "user"
	user := &github.User{Login: &login}

	result := createSharingReadmeContent(user, "recipient", "")

	// Check for proper markdown structure
	lines := strings.Split(result, "\n")
	
	hasMainHeader := false
	hasSubHeader := false
	hasCodeBlock := false

	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			hasMainHeader = true
		}
		if strings.HasPrefix(line, "## ") {
			hasSubHeader = true
		}
		if strings.Contains(line, "```") {
			hasCodeBlock = true
		}
	}

	if !hasMainHeader {
		t.Error("Should have main header (# )")
	}
	if !hasSubHeader {
		t.Error("Should have subheader (## )")
	}
	if !hasCodeBlock {
		t.Error("Should have code blocks (```)")
	}
}

func TestCreateSharingReadmeContentConsistency(t *testing.T) {
	login := "user"
	user := &github.User{Login: &login}

	// Generate twice with same parameters
	result1 := createSharingReadmeContent(user, "recipient", "")
	result2 := createSharingReadmeContent(user, "recipient", "")

	// Should be identical
	if result1 != result2 {
		t.Error("Should generate consistent output for same parameters")
	}
}

func TestCreateSharingReadmeContentDifferentRecipients(t *testing.T) {
	login := "sender"
	user := &github.User{Login: &login}

	result1 := createSharingReadmeContent(user, "recipient1", "")
	result2 := createSharingReadmeContent(user, "recipient2", "")

	// Should be different (contain different recipient names)
	if result1 == result2 {
		t.Error("Should generate different output for different recipients")
	}

	// But should both contain sender
	if !strings.Contains(result1, "sender") || !strings.Contains(result2, "sender") {
		t.Error("Both should contain sender username")
	}

	// And contain respective recipients
	if !strings.Contains(result1, "recipient1") {
		t.Error("Result1 should contain recipient1")
	}
	if !strings.Contains(result2, "recipient2") {
		t.Error("Result2 should contain recipient2")
	}
}

func TestCreateSharingReadmeContentNoEmptyLines(t *testing.T) {
	login := "user"
	user := &github.User{Login: &login}

	result := createSharingReadmeContent(user, "recipient", "")

	// Should not have excessive empty lines
	lines := strings.Split(result, "\n")
	consecutiveEmpty := 0
	maxConsecutiveEmpty := 0

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			consecutiveEmpty++
			if consecutiveEmpty > maxConsecutiveEmpty {
				maxConsecutiveEmpty = consecutiveEmpty
			}
		} else {
			consecutiveEmpty = 0
		}
	}

	if maxConsecutiveEmpty > 3 {
		t.Errorf("Too many consecutive empty lines: %d", maxConsecutiveEmpty)
	}
}

func TestCreateSharingReadmeContentSpecialCharacters(t *testing.T) {
	tests := []struct {
		name      string
		recipient string
	}{
		{"with hyphen", "user-name"},
		{"with underscore", "user_name"},
		{"with numbers", "user123"},
		{"mixed case", "UserName"},
	}

	login := "sender"
	user := &github.User{Login: &login}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := createSharingReadmeContent(user, tt.recipient, "")
			
			if !strings.Contains(result, tt.recipient) {
				t.Errorf("Should contain recipient name: %s", tt.recipient)
			}

			// Should still be valid markdown
			if !strings.HasPrefix(result, "#") {
				t.Error("Should still be valid markdown")
			}
		})
	}
}

func TestCreateSharingReadmeContentLongUsernames(t *testing.T) {
	longUsername := strings.Repeat("a", 100)
	login := longUsername
	user := &github.User{Login: &login}

	result := createSharingReadmeContent(user, "recipient", "")

	// Should handle long usernames without breaking
	if len(result) < 200 {
		t.Error("Should still generate reasonable content with long usernames")
	}

	// Should contain the long username
	if !strings.Contains(result, longUsername) {
		t.Error("Should contain the long username")
	}
}
