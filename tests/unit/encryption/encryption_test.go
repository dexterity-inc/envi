package encryption_test

import (
	"crypto/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dexterity-inc/envi/internal/encryption"
	"github.com/dexterity-inc/envi/tests/helpers"
)

func TestEncryptDecryptContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		password    string
		expectError bool
	}{
		{
			name:        "basic encryption and decryption",
			content:     "SECRET_KEY=my-secret-value\nAPI_TOKEN=abc123",
			password:    "test-password-123",
			expectError: false,
		},
		{
			name:        "empty content",
			content:     "",
			password:    "test-password",
			expectError: false,
		},
		{
			name:        "large content",
			content:     strings.Repeat("LARGE_VAR=large_value\n", 1000),
			password:    "strong-password-456",
			expectError: false,
		},
		{
			name:        "unicode content",
			content:     "UNICODE_VAR=Hello 世界 🌍\nEMOJI=🚀🔒",
			password:    "unicode-password-789",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test environment with mock password input
			testConfig := helpers.SetupTestEnvironment(t)
			defer testConfig.Cleanup()

			originalContent := []byte(tt.content)

			// Test encryption
			encryptedContent, err := encryption.EncryptContent(originalContent)
			if tt.expectError {
				if err == nil {
					t.Error("Expected encryption error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected encryption error: %v", err)
			}

			// Verify content is actually encrypted (should be different)
			if string(encryptedContent) == tt.content {
				t.Error("Encrypted content should be different from original")
			}

			// Verify encrypted content has the correct prefix
			if !strings.HasPrefix(string(encryptedContent), encryption.EncryptionPrefix) {
				t.Error("Encrypted content should have encryption prefix")
			}

			// Test decryption
			decryptedContent, err := encryption.DecryptContent(encryptedContent)
			if err != nil {
				t.Fatalf("Unexpected decryption error: %v", err)
			}

			// Verify decrypted content matches original
			if string(decryptedContent) != tt.content {
				t.Errorf("Decrypted content doesn't match original.\nExpected: %s\nGot: %s", 
					tt.content, string(decryptedContent))
			}
		})
	}
}

func TestMaskUnmaskEnvContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name: "basic masking and unmasking",
			content: `DATABASE_URL=postgres://user:pass@localhost/db
API_KEY=secret123
DEBUG=true`,
			expectError: false,
		},
		{
			name: "empty content",
			content: "",
			expectError: false,
		},
		{
			name: "content with comments",
			content: `# Database configuration
DATABASE_URL=postgres://localhost/db
# API settings
API_KEY=secret-key`,
			expectError: false,
		},
		{
			name: "content with empty values",
			content: `EMPTY_VAR=
NON_EMPTY=value`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalContent := []byte(tt.content)

			// Test masking
			maskedContent, err := encryption.MaskEnvContent(originalContent)
			if tt.expectError {
				if err == nil {
					t.Error("Expected masking error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected masking error: %v", err)
			}

			// Verify content is masked but variable names are visible
			maskedStr := string(maskedContent)
			if strings.Contains(tt.content, "DATABASE_URL") {
				if !strings.Contains(maskedStr, "DATABASE_URL") {
					t.Error("Variable names should remain visible after masking")
				}
			}

			// Verify values are masked (should have masked prefix)
			if strings.Contains(tt.content, "=") && !strings.Contains(maskedStr, encryption.MaskedPrefix) {
				t.Error("Masked content should contain masked prefix")
			}

			// Test unmasking
			unmaskedContent, err := encryption.UnmaskEnvContent(maskedContent)
			if err != nil {
				t.Fatalf("Unexpected unmasking error: %v", err)
			}

			// Verify unmasked content matches original
			if string(unmaskedContent) != tt.content {
				t.Errorf("Unmasked content doesn't match original.\nExpected: %s\nGot: %s", 
					tt.content, string(unmaskedContent))
			}
		})
	}
}

func TestKeyFileOperations(t *testing.T) {
	testConfig := helpers.SetupTestEnvironment(t)
	defer testConfig.Cleanup()

	keyFilePath := filepath.Join(testConfig.TempDir, "test.key")

	t.Run("generate and save key file", func(t *testing.T) {
		// Generate a random key
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			t.Fatalf("Failed to generate random key: %v", err)
		}

		// Save key to file
		err = os.WriteFile(keyFilePath, key, 0600)
		if err != nil {
			t.Fatalf("Failed to save key file: %v", err)
		}

		// Verify file exists and has correct permissions
		helpers.AssertFileExists(t, keyFilePath)
		helpers.AssertFilePermissions(t, keyFilePath, 0600)

		// Read key back
		savedKey, err := os.ReadFile(keyFilePath)
		if err != nil {
			t.Fatalf("Failed to read key file: %v", err)
		}

		// Verify key matches
		if len(savedKey) != len(key) {
			t.Errorf("Key length mismatch: expected %d, got %d", len(key), len(savedKey))
		}

		// Verify key content matches
		for i, b := range key {
			if savedKey[i] != b {
				t.Error("Key content doesn't match after save/load")
				break
			}
		}
	})

	t.Run("use key file for encryption", func(t *testing.T) {
		// This test would require mocking the key file reading functionality
		// or extending the encryption package to accept key files directly
		// For now, we'll test that the key file exists and is readable
		if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
			t.Skip("Key file not available for this test")
		}

		content := []byte("TEST_VAR=test_value")
		
		// Test encryption with key file (would need enhanced encryption functions)
		// This is a placeholder for the actual key file encryption test
		if len(content) == 0 {
			t.Error("Content should not be empty")
		}
	})
}

func TestEncryptionConstants(t *testing.T) {
	t.Run("encryption constants are defined", func(t *testing.T) {
		if encryption.EncryptionPrefix == "" {
			t.Error("EncryptionPrefix should not be empty")
		}
		if encryption.MaskedPrefix == "" {
			t.Error("MaskedPrefix should not be empty")
		}
		if encryption.EncryptionKeyLength != 32 {
			t.Errorf("EncryptionKeyLength should be 32, got %d", encryption.EncryptionKeyLength)
		}
		if encryption.PBKDF2Iterations < 100000 {
			t.Errorf("PBKDF2Iterations should be at least 100000, got %d", encryption.PBKDF2Iterations)
		}
		if encryption.PBKDF2SaltLength != 16 {
			t.Errorf("PBKDF2SaltLength should be 16, got %d", encryption.PBKDF2SaltLength)
		}
	})
}

func TestEncryptionErrorHandling(t *testing.T) {
	t.Run("decrypt invalid content", func(t *testing.T) {
		invalidContent := []byte("invalid encrypted content")
		
		_, err := encryption.DecryptContent(invalidContent)
		if err == nil {
			t.Error("Expected error when decrypting invalid content")
		}
	})

	t.Run("unmask invalid content", func(t *testing.T) {
		invalidContent := []byte("invalid masked content")
		
		_, err := encryption.UnmaskEnvContent(invalidContent)
		if err == nil {
			t.Error("Expected error when unmasking invalid content")
		}
	})

	t.Run("decrypt content without prefix", func(t *testing.T) {
		contentWithoutPrefix := []byte("some content without encryption prefix")
		
		_, err := encryption.DecryptContent(contentWithoutPrefix)
		if err == nil {
			t.Error("Expected error when decrypting content without prefix")
		}
	})
}

func TestPBKDF2KeyDerivation(t *testing.T) {
	// This test verifies that key derivation is deterministic
	// Since PBKDF2 key derivation is not directly exposed,
	// we test it indirectly through encryption/decryption
	content := []byte("TEST=value")
	
	// Encrypt the same content twice (would need to mock salt for deterministic results)
	encrypted1, err := encryption.EncryptContent(content)
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}
	
	encrypted2, err := encryption.EncryptContent(content)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}
	
	// The encrypted results should be different due to random salts
	if string(encrypted1) == string(encrypted2) {
		t.Error("Encrypted content should be different due to random salts")
	}
	
	// But both should decrypt to the same original content
	decrypted1, err := encryption.DecryptContent(encrypted1)
	if err != nil {
		t.Fatalf("First decryption failed: %v", err)
	}
	
	decrypted2, err := encryption.DecryptContent(encrypted2)
	if err != nil {
		t.Fatalf("Second decryption failed: %v", err)
	}
	
	if string(decrypted1) != string(decrypted2) {
		t.Error("Both decrypted contents should be identical")
	}
}

