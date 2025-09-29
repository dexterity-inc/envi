package encryption

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"github.com/spf13/cobra"
)

func TestEncryptionConstants(t *testing.T) {
	// Test that encryption constants are properly defined
	if EncryptionPrefix == "" {
		t.Error("EncryptionPrefix should not be empty")
	}
	if MaskedPrefix == "" {
		t.Error("MaskedPrefix should not be empty")
	}
	if EncryptionKeyLength != 32 {
		t.Errorf("Expected EncryptionKeyLength to be 32, got %d", EncryptionKeyLength)
	}
	
	// Test PBKDF2 parameters meet OWASP recommendations
	if PBKDF2Iterations < 100000 {
		t.Errorf("PBKDF2Iterations should be at least 100000 (OWASP recommended), got %d", PBKDF2Iterations)
	}
	if PBKDF2SaltLength < 16 {
		t.Errorf("PBKDF2SaltLength should be at least 16 bytes, got %d", PBKDF2SaltLength)
	}
}

func TestInitEncryptionFlags(t *testing.T) {
	cmd := &cobra.Command{
		Use: "test",
	}
	
	// Test flag initialization
	InitEncryptionFlags(cmd)
	
	// Check that required flags are added
	encryptFlag := cmd.PersistentFlags().Lookup("encrypt")
	if encryptFlag == nil {
		t.Error("encrypt flag should be initialized")
	}
	
	maskFlag := cmd.PersistentFlags().Lookup("mask")
	if maskFlag == nil {
		t.Error("mask flag should be initialized")
	}
	
	useKeyFileFlag := cmd.PersistentFlags().Lookup("use-key-file")
	if useKeyFileFlag == nil {
		t.Error("use-key-file flag should be initialized")
	}
	
	keyFileFlag := cmd.PersistentFlags().Lookup("key-file")
	if keyFileFlag == nil {
		t.Error("key-file flag should be initialized")
	}
	
	// Check default values
	if keyFileFlag.DefValue != ".envi.key" {
		t.Errorf("Expected default key-file to be '.envi.key', got %s", keyFileFlag.DefValue)
	}
}

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "encrypted content",
			content:  []byte(EncryptionPrefix + "base64data"),
			expected: true,
		},
		{
			name:     "non-encrypted content",
			content:  []byte("DATABASE_URL=postgres://localhost"),
			expected: false,
		},
		{
			name:     "empty content",
			content:  []byte(""),
			expected: false,
		},
		{
			name:     "partial prefix",
			content:  []byte("ENVI_ENC"),
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEncrypted(tt.content)
			if result != tt.expected {
				t.Errorf("IsEncrypted() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsMasked(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "masked content",
			content:  []byte("DATABASE_URL=" + MaskedPrefix + "base64data"),
			expected: true,
		},
		{
			name:     "non-masked content",
			content:  []byte("DATABASE_URL=postgres://localhost"),
			expected: false,
		},
		{
			name:     "empty content",
			content:  []byte(""),
			expected: false,
		},
		{
			name:     "multiple masked values",
			content:  []byte("API_KEY=" + MaskedPrefix + "data1\nSECRET=" + MaskedPrefix + "data2"),
			expected: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsMasked(tt.content)
			if result != tt.expected {
				t.Errorf("IsMasked() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsSelfContainedShare(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected bool
	}{
		{
			name:     "encrypted content is self-contained",
			content:  []byte(EncryptionPrefix + "base64data"),
			expected: true,
		},
		{
			name:     "non-encrypted content is not self-contained",
			content:  []byte("DATABASE_URL=postgres://localhost"),
			expected: false,
		},
		{
			name:     "empty content is not self-contained",
			content:  []byte(""),
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSelfContainedShare(tt.content)
			if result != tt.expected {
				t.Errorf("IsSelfContainedShare() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "simple password",
			password: "testpassword",
		},
		{
			name:     "complex password",
			password: "P@ssw0rd!Complex#123",
		},
		{
			name:     "unicode password", 
			password: "测试密码",
		},
		{
			name:     "empty password",
			password: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := hashPassword(tt.password)
			hash2 := hashPassword(tt.password)
			
			// Check that hash length is correct
			if len(hash1) != EncryptionKeyLength {
				t.Errorf("Expected hash length %d, got %d", EncryptionKeyLength, len(hash1))
			}
			
			// Check that hash is deterministic
			if !bytes.Equal(hash1, hash2) {
				t.Error("hashPassword should produce deterministic results")
			}
			
			// Check that different passwords produce different hashes
			if tt.password != "" {
				differentHash := hashPassword(tt.password + "different")
				if bytes.Equal(hash1, differentHash) {
					t.Error("Different passwords should produce different hashes")
				}
			}
		})
	}
}

func TestGenerateKeyFile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	keyFilePath := filepath.Join(tempDir, "test.key")
	
	// Test key file generation
	err := GenerateKeyFile(keyFilePath)
	if err != nil {
		t.Fatalf("GenerateKeyFile() returned error: %v", err)
	}
	
	// Check that file was created
	if _, err := os.Stat(keyFilePath); os.IsNotExist(err) {
		t.Error("Key file was not created")
	}
	
	// Check file permissions
	info, err := os.Stat(keyFilePath)
	if err != nil {
		t.Fatalf("Error stating key file: %v", err)
	}
	
	expectedPerm := os.FileMode(0600)
	if info.Mode().Perm() != expectedPerm {
		t.Errorf("Expected key file permissions %v, got %v", expectedPerm, info.Mode().Perm())
	}
	
	// Check file content
	content, err := os.ReadFile(keyFilePath)
	if err != nil {
		t.Fatalf("Error reading key file: %v", err)
	}
	
	if len(content) == 0 {
		t.Error("Key file should not be empty")
	}
	
	// Check that it's valid base64
	decoded, err := base64DecodeString(strings.TrimSpace(string(content)))
	if err != nil {
		t.Errorf("Key file content should be valid base64: %v", err)
	}
	
	if len(decoded) != EncryptionKeyLength {
		t.Errorf("Decoded key should be %d bytes, got %d", EncryptionKeyLength, len(decoded))
	}
}

func TestMaskEnvContent(t *testing.T) {
	// Set up test environment with a password
	originalPassword := EncryptionPassword
	originalUseKeyFile := UseKeyFile
	defer func() {
		EncryptionPassword = originalPassword
		UseKeyFile = originalUseKeyFile
	}()
	
	EncryptionPassword = "testpassword"
	UseKeyFile = false
	
	tests := []struct {
		name     string
		content  []byte
		wantErr  bool
	}{
		{
			name: "simple env file",
			content: []byte(`DATABASE_URL=postgres://localhost:5432/mydb
API_KEY=secret123
# This is a comment
EMPTY_VAR=
PORT=3000`),
			wantErr: false,
		},
		{
			name: "env file with comments only",
			content: []byte(`# Comment 1
# Comment 2

# Another comment`),
			wantErr: false,
		},
		{
			name:    "empty content",
			content: []byte(""),
			wantErr: false,
		},
		{
			name: "malformed lines",
			content: []byte(`VALID_VAR=value
INVALID_LINE_NO_EQUALS
ANOTHER_VALID=value2`),
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MaskEnvContent(tt.content)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("MaskEnvContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// Check that result is not empty for non-empty input
				if len(tt.content) > 0 && len(result) == 0 {
					t.Error("MaskEnvContent() should not return empty result for non-empty input")
				}
				
				// Check that masked content contains the masked prefix for non-comment lines
				resultStr := string(result)
				if strings.Contains(string(tt.content), "=") && !strings.Contains(resultStr, MaskedPrefix) {
					// Only expect masked prefix if there were actual key=value pairs
					lines := strings.Split(string(tt.content), "\n")
					hasValues := false
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line != "" && !strings.HasPrefix(line, "#") && strings.Contains(line, "=") {
							parts := strings.SplitN(line, "=", 2)
							if len(parts) == 2 && parts[1] != "" {
								hasValues = true
								break
							}
						}
					}
					if hasValues {
						t.Error("MaskEnvContent() should contain masked prefix for files with values")
					}
				}
				
				// Check that comments are preserved
				originalLines := strings.Split(string(tt.content), "\n")
				resultLines := strings.Split(resultStr, "\n")
				
				for i, line := range originalLines {
					if strings.HasPrefix(strings.TrimSpace(line), "#") {
						if i < len(resultLines) && resultLines[i] != line {
							t.Errorf("Comment line should be preserved: expected %q, got %q", line, resultLines[i])
						}
					}
				}
			}
		})
	}
}

func TestDecryptSelfContainedShare(t *testing.T) {
	// Test that the function exists and handles basic cases
	tests := []struct {
		name     string
		content  []byte
		password string
		wantErr  bool
	}{
		{
			name:     "non-encrypted content",
			content:  []byte("DATABASE_URL=value"),
			password: "password",
			wantErr:  true,
		},
		{
			name:     "empty content",
			content:  []byte(""),
			password: "password",
			wantErr:  true,
		},
		{
			name:     "invalid encrypted content",
			content:  []byte(EncryptionPrefix + "invalid_base64!"),
			password: "password",
			wantErr:  true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptSelfContainedShare(tt.content, tt.password)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptSelfContainedShare() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to decode base64
func base64DecodeString(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

// Test EncryptContent function
func TestEncryptContent(t *testing.T) {
	// Save original encryption settings
	originalPassword := EncryptionPassword
	originalUseKeyFile := UseKeyFile
	defer func() {
		EncryptionPassword = originalPassword
		UseKeyFile = originalUseKeyFile
	}()
	
	// Set up test password
	EncryptionPassword = "testpassword123"
	UseKeyFile = false
	
	tests := []struct {
		name    string
		content []byte
		wantErr bool
	}{
		{
			name:    "simple text",
			content: []byte("Hello, World!"),
			wantErr: false,
		},
		{
			name:    "empty content",
			content: []byte(""),
			wantErr: false,
		},
		{
			name:    "multiline content",
			content: []byte("line1\nline2\nline3"),
			wantErr: false,
		},
		{
			name:    "binary content",
			content: []byte{0, 1, 2, 3, 255, 254, 253},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := EncryptContent(tt.content)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// Verify the result has the encryption prefix
				if !IsEncrypted(encrypted) {
					t.Error("Encrypted content should have encryption prefix")
				}
				
				// Verify we can decrypt it back
				decrypted, err := DecryptContent(encrypted)
				if err != nil {
					t.Errorf("Failed to decrypt encrypted content: %v", err)
				}
				
				if !bytes.Equal(decrypted, tt.content) {
					t.Errorf("Decrypted content doesn't match original. Got %v, want %v", decrypted, tt.content)
				}
			}
		})
	}
}

// Test UnmaskEnvContent function
func TestUnmaskEnvContent(t *testing.T) {
	// Save original encryption settings
	originalPassword := EncryptionPassword
	originalUseKeyFile := UseKeyFile
	defer func() {
		EncryptionPassword = originalPassword
		UseKeyFile = originalUseKeyFile
	}()
	
	// Set up test password
	EncryptionPassword = "testpassword123"
	UseKeyFile = false
	
	tests := []struct {
		name    string
		content []byte
		wantErr bool
	}{
		{
			name:    "empty content",
			content: []byte(""),
			wantErr: false,
		},
		{
			name:    "comments only",
			content: []byte("# Comment 1\n# Comment 2"),
			wantErr: false,
		},
		{
			name:    "unmasked content",
			content: []byte("KEY1=value1\nKEY2=value2"),
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First mask the content if it has values
			masked, err := MaskEnvContent(tt.content)
			if err != nil {
				t.Errorf("Failed to mask content: %v", err)
				return
			}
			
			// Then unmask it
			unmasked, err := UnmaskEnvContent(masked)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmaskEnvContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				// The unmasked content should match the original
				if !bytes.Equal(unmasked, tt.content) {
					t.Errorf("Unmasked content doesn't match original.\nGot: %s\nWant: %s", unmasked, tt.content)
				}
			}
		})
	}
}

// Test getKeyFromFile function
func TestGetKeyFromFile(t *testing.T) {
	// Save original settings
	originalKeyFile := EncryptionKeyFile
	defer func() {
		EncryptionKeyFile = originalKeyFile
	}()
	
	tests := []struct {
		name       string
		keyContent string
		wantErr    bool
		setupFile  bool
	}{
		{
			name:       "valid base64 key",
			keyContent: base64.StdEncoding.EncodeToString(make([]byte, EncryptionKeyLength)),
			wantErr:    false,
			setupFile:  true,
		},
		{
			name:       "raw key of correct length",
			keyContent: string(make([]byte, EncryptionKeyLength)),
			wantErr:    false,
			setupFile:  true,
		},
		{
			name:       "password to be hashed",
			keyContent: "mypassword123",
			wantErr:    false,
			setupFile:  true,
		},
		{
			name:      "nonexistent file",
			wantErr:   true,
			setupFile: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFile {
				// Create a temporary key file with relative path
				tempDir := t.TempDir()
				originalDir, _ := os.Getwd()
				defer os.Chdir(originalDir)
				
				os.Chdir(tempDir)
				keyFile := "test.key" // Use relative path to avoid security validation
				
				err := os.WriteFile(keyFile, []byte(tt.keyContent), 0600)
				if err != nil {
					t.Fatalf("Failed to create test key file: %v", err)
				}
				
				EncryptionKeyFile = keyFile
			} else {
				EncryptionKeyFile = "nonexistent.key" // Use relative path
			}
			
			key, err := getKeyFromFile()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("getKeyFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				if len(key) != EncryptionKeyLength {
					t.Errorf("Expected key length %d, got %d", EncryptionKeyLength, len(key))
				}
			}
		})
	}
}

// Test getEncryptionKey function with different configurations
func TestGetEncryptionKey(t *testing.T) {
	// Save original settings
	originalPassword := EncryptionPassword
	originalUseKeyFile := UseKeyFile
	originalKeyFile := EncryptionKeyFile
	defer func() {
		EncryptionPassword = originalPassword
		UseKeyFile = originalUseKeyFile
		EncryptionKeyFile = originalKeyFile
	}()
	
	t.Run("use password flag", func(t *testing.T) {
		UseKeyFile = false
		EncryptionPassword = "testpassword"
		
		key, err := getEncryptionKey()
		if err != nil {
			t.Errorf("getEncryptionKey() with password failed: %v", err)
		}
		
		if len(key) != EncryptionKeyLength {
			t.Errorf("Expected key length %d, got %d", EncryptionKeyLength, len(key))
		}
	})
	
	t.Run("use key file", func(t *testing.T) {
		UseKeyFile = true
		EncryptionPassword = ""
		
		// Create a temporary key file with relative path
		tempDir := t.TempDir()
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		
		os.Chdir(tempDir)
		keyFile := "test.key" // Use relative path
		keyContent := base64.StdEncoding.EncodeToString(make([]byte, EncryptionKeyLength))
		
		err := os.WriteFile(keyFile, []byte(keyContent), 0600)
		if err != nil {
			t.Fatalf("Failed to create test key file: %v", err)
		}
		
		EncryptionKeyFile = keyFile
		
		key, err := getEncryptionKey()
		if err != nil {
			t.Errorf("getEncryptionKey() with key file failed: %v", err)
		}
		
		if len(key) != EncryptionKeyLength {
			t.Errorf("Expected key length %d, got %d", EncryptionKeyLength, len(key))
		}
	})
	
	t.Run("invalid key file path", func(t *testing.T) {
		UseKeyFile = true
		EncryptionPassword = ""
		EncryptionKeyFile = "../../../etc/passwd" // Invalid path
		
		_, err := getEncryptionKey()
		if err == nil {
			t.Error("getEncryptionKey() should fail with invalid key file path")
		}
	})
}
