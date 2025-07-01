package encryption

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dexterity-inc/envi/internal/utils"
)

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    bool
	}{
		{
			name:    "encrypted content",
			content: []byte(utils.EncryptionPrefix + "encrypted_data"),
			want:    true,
		},
		{
			name:    "unencrypted content",
			content: []byte("KEY=value"),
			want:    false,
		},
		{
			name:    "empty content",
			content: []byte(""),
			want:    false,
		},
		{
			name:    "partial prefix",
			content: []byte("ENVI_ENC"),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEncrypted(tt.content)
			if got != tt.want {
				t.Errorf("IsEncrypted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsMasked(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    bool
	}{
		{
			name:    "masked content",
			content: []byte("KEY=" + utils.MaskedPrefix + "masked_value"),
			want:    true,
		},
		{
			name:    "unmasked content",
			content: []byte("KEY=value"),
			want:    false,
		},
		{
			name:    "empty content",
			content: []byte(""),
			want:    false,
		},
		{
			name:    "partial prefix",
			content: []byte("ENVI_MAS"),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsMasked(tt.content)
			if got != tt.want {
				t.Errorf("IsMasked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSelfContainedShare(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    bool
	}{
		{
			name:    "self-contained content",
			content: []byte(utils.SelfContainedPrefix + "self_contained_data"),
			want:    true,
		},
		{
			name:    "regular content",
			content: []byte("KEY=value"),
			want:    false,
		},
		{
			name:    "empty content",
			content: []byte(""),
			want:    false,
		},
		{
			name:    "partial prefix",
			content: []byte("ENVI_SELF"),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSelfContainedShare(tt.content)
			if got != tt.want {
				t.Errorf("IsSelfContainedShare() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test encryption and decryption with password
func TestEncryptDecryptContent_WithPassword(t *testing.T) {
	// Set up test environment
	originalPassword := EncryptionPassword
	originalUseKeyFile := UseKeyFile
	defer func() {
		EncryptionPassword = originalPassword
		UseKeyFile = originalUseKeyFile
	}()

	// Use password for encryption
	UseKeyFile = false
	EncryptionPassword = "testpassword123"

	testData := []byte("TEST_VAR=secret_value\nANOTHER_VAR=another_secret")

	// Test encryption
	encrypted, err := EncryptContent(testData)
	if err != nil {
		t.Fatalf("EncryptContent() error = %v", err)
	}

	// Verify encrypted data has correct prefix
	if !IsEncrypted(encrypted) {
		t.Error("Encrypted content should have encryption prefix")
	}

	// Test decryption
	decrypted, err := DecryptContent(encrypted)
	if err != nil {
		t.Fatalf("DecryptContent() error = %v", err)
	}

	// Verify decrypted data matches original
	if !bytes.Equal(decrypted, testData) {
		t.Errorf("Decrypted content = %s, want %s", string(decrypted), string(testData))
	}
}

// Test encryption and decryption with key file
func TestEncryptDecryptContent_WithKeyFile(t *testing.T) {
	// Create temporary directory for key file
	tempDir, err := os.MkdirTemp("", "encryption_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	keyFile := filepath.Join(tempDir, "test.key")
	testKey := []byte("12345678901234567890123456789012") // 32 bytes for AES-256

	// Create key file
	if err := os.WriteFile(keyFile, testKey, 0600); err != nil {
		t.Fatalf("Failed to create key file: %v", err)
	}

	// Set up test environment
	originalUseKeyFile := UseKeyFile
	originalKeyFile := EncryptionKeyFile
	defer func() {
		UseKeyFile = originalUseKeyFile
		EncryptionKeyFile = originalKeyFile
	}()

	UseKeyFile = true
	EncryptionKeyFile = keyFile

	testData := []byte("TEST_VAR=secret_value\nANOTHER_VAR=another_secret")

	// Test encryption
	encrypted, err := EncryptContent(testData)
	if err != nil {
		t.Fatalf("EncryptContent() error = %v", err)
	}

	// Verify encrypted data has correct prefix
	if !IsEncrypted(encrypted) {
		t.Error("Encrypted content should have encryption prefix")
	}

	// Test decryption
	decrypted, err := DecryptContent(encrypted)
	if err != nil {
		t.Fatalf("DecryptContent() error = %v", err)
	}

	// Verify decrypted data matches original
	if !bytes.Equal(decrypted, testData) {
		t.Errorf("Decrypted content = %s, want %s", string(decrypted), string(testData))
	}
}

// Test masking and unmasking
func TestMaskUnmaskEnvContent(t *testing.T) {
	// Set up test environment
	originalPassword := EncryptionPassword
	originalUseKeyFile := UseKeyFile
	defer func() {
		EncryptionPassword = originalPassword
		UseKeyFile = originalUseKeyFile
	}()

	UseKeyFile = false
	EncryptionPassword = "testpassword123"

	testContent := `# This is a comment
TEST_VAR=secret_value
ANOTHER_VAR=another_secret
EMPTY_VAR=

# Another comment
FINAL_VAR=final_value`

	// Test masking
	masked, err := MaskEnvContent([]byte(testContent))
	if err != nil {
		t.Fatalf("MaskEnvContent() error = %v", err)
	}

	// Verify masked content contains comments but has masked values
	maskedStr := string(masked)
	if !strings.Contains(maskedStr, "# This is a comment") {
		t.Error("Masked content should preserve comments")
	}

	if !strings.Contains(maskedStr, utils.MaskedPrefix) {
		t.Error("Masked content should contain masked values")
	}

	if strings.Contains(maskedStr, "secret_value") {
		t.Error("Masked content should not contain original secret values")
	}

	// Test unmasking
	unmasked, err := UnmaskEnvContent(masked)
	if err != nil {
		t.Fatalf("UnmaskEnvContent() error = %v", err)
	}

	// Verify unmasked content matches original
	if !bytes.Equal(unmasked, []byte(testContent)) {
		t.Errorf("Unmasked content doesn't match original.\nGot:\n%s\nWant:\n%s",
			string(unmasked), testContent)
	}
}

// Test self-contained encryption and decryption
func TestSelfContainedEncryption(t *testing.T) {
	testData := []byte("TEST_VAR=secret_value\nANOTHER_VAR=another_secret")

	// Test that we can identify self-contained content
	selfContainedContent := []byte(utils.SelfContainedPrefix + "test_data")
	if !IsSelfContainedShare(selfContainedContent) {
		t.Error("Should identify self-contained content correctly")
	}

	// Test that regular content is not identified as self-contained
	if IsSelfContainedShare(testData) {
		t.Error("Regular content should not be identified as self-contained")
	}
}

// Test error cases
func TestDecryptContent_InvalidData(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		wantErr bool
	}{
		{
			name:    "unencrypted content",
			content: []byte("KEY=value"),
			wantErr: true,
		},
		{
			name:    "invalid base64",
			content: []byte(utils.EncryptionPrefix + "invalid_base64!@#"),
			wantErr: true,
		},
		{
			name:    "empty encrypted content",
			content: []byte(utils.EncryptionPrefix),
			wantErr: true,
		},
	}

	// Set up valid password for testing
	originalPassword := EncryptionPassword
	originalUseKeyFile := UseKeyFile
	defer func() {
		EncryptionPassword = originalPassword
		UseKeyFile = originalUseKeyFile
	}()

	UseKeyFile = false
	EncryptionPassword = "testpassword123"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSelfContainedShare_ValidationOnly(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    bool
	}{
		{
			name:    "valid self-contained content",
			content: []byte(utils.SelfContainedPrefix + "some_data"),
			want:    true,
		},
		{
			name:    "non-self-contained content",
			content: []byte("KEY=value"),
			want:    false,
		},
		{
			name:    "empty content",
			content: []byte(""),
			want:    false,
		},
		{
			name:    "partial prefix",
			content: []byte("ENVI_SELF"),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSelfContainedShare(tt.content)
			if got != tt.want {
				t.Errorf("IsSelfContainedShare() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test prefix detection functionality
func TestPrefixDetection(t *testing.T) {
	tests := []struct {
		name        string
		content     []byte
		isEncrypted bool
		isMasked    bool
		isSelfCont  bool
	}{
		{
			name:        "encrypted content",
			content:     []byte(utils.EncryptionPrefix + "data"),
			isEncrypted: true,
			isMasked:    false,
			isSelfCont:  false,
		},
		{
			name:        "masked content",
			content:     []byte("KEY=" + utils.MaskedPrefix + "data"),
			isEncrypted: false,
			isMasked:    true,
			isSelfCont:  false,
		},
		{
			name:        "self-contained content",
			content:     []byte(utils.SelfContainedPrefix + "data"),
			isEncrypted: false,
			isMasked:    false,
			isSelfCont:  true,
		},
		{
			name:        "plain content",
			content:     []byte("KEY=value"),
			isEncrypted: false,
			isMasked:    false,
			isSelfCont:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsEncrypted(tt.content) != tt.isEncrypted {
				t.Errorf("IsEncrypted() = %v, want %v", IsEncrypted(tt.content), tt.isEncrypted)
			}
			if IsMasked(tt.content) != tt.isMasked {
				t.Errorf("IsMasked() = %v, want %v", IsMasked(tt.content), tt.isMasked)
			}
			if IsSelfContainedShare(tt.content) != tt.isSelfCont {
				t.Errorf("IsSelfContainedShare() = %v, want %v", IsSelfContainedShare(tt.content), tt.isSelfCont)
			}
		})
	}
}

// Test masking empty values
func TestMaskEnvContent_EmptyValues(t *testing.T) {
	// Set up test environment
	originalPassword := EncryptionPassword
	originalUseKeyFile := UseKeyFile
	defer func() {
		EncryptionPassword = originalPassword
		UseKeyFile = originalUseKeyFile
	}()

	UseKeyFile = false
	EncryptionPassword = "testpassword123"

	testContent := `VAR_WITH_VALUE=some_value
VAR_EMPTY=
VAR_SPACES=   
VAR_TABS=		`

	masked, err := MaskEnvContent([]byte(testContent))
	if err != nil {
		t.Fatalf("MaskEnvContent() error = %v", err)
	}

	maskedStr := string(masked)

	// Empty values should not be masked
	if !strings.Contains(maskedStr, "VAR_EMPTY=") {
		t.Error("Empty variables should be preserved")
	}

	// Non-empty values should be masked
	if !strings.Contains(maskedStr, utils.MaskedPrefix) {
		t.Error("Non-empty values should be masked")
	}

	// Verify we can unmask successfully
	unmasked, err := UnmaskEnvContent(masked)
	if err != nil {
		t.Fatalf("UnmaskEnvContent() error = %v", err)
	}

	if !bytes.Equal(unmasked, []byte(testContent)) {
		t.Errorf("Unmasked content doesn't match original")
	}
}

// Benchmark tests
func BenchmarkEncryptContent(b *testing.B) {
	// Set up test environment
	UseKeyFile = false
	EncryptionPassword = "testpassword123"

	testData := []byte("TEST_VAR=secret_value\nANOTHER_VAR=another_secret")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := EncryptContent(testData)
		if err != nil {
			b.Fatalf("EncryptContent() error = %v", err)
		}
	}
}

func BenchmarkDecryptContent(b *testing.B) {
	// Set up test environment
	UseKeyFile = false
	EncryptionPassword = "testpassword123"

	testData := []byte("TEST_VAR=secret_value\nANOTHER_VAR=another_secret")
	encrypted, err := EncryptContent(testData)
	if err != nil {
		b.Fatalf("EncryptContent() setup error = %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := DecryptContent(encrypted)
		if err != nil {
			b.Fatalf("DecryptContent() error = %v", err)
		}
	}
}

func BenchmarkMaskEnvContent(b *testing.B) {
	// Set up test environment
	UseKeyFile = false
	EncryptionPassword = "testpassword123"

	testContent := []byte(`TEST_VAR=secret_value
ANOTHER_VAR=another_secret
THIRD_VAR=third_value`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := MaskEnvContent(testContent)
		if err != nil {
			b.Fatalf("MaskEnvContent() error = %v", err)
		}
	}
}

func BenchmarkPrefixDetection(b *testing.B) {
	testData := []byte(utils.EncryptionPrefix + "encrypted_data_here")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsEncrypted(testData)
	}
}
