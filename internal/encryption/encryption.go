package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/dexterity-inc/envi/internal/tui"
	"github.com/dexterity-inc/envi/internal/utils"
	"golang.org/x/crypto/pbkdf2"
)

// Encryption command flags
var (
	UseEncryption       bool
	UseMaskedEncryption bool
	UseKeyFile          bool
	EncryptionKeyFile   string
	EncryptionPassword  string
	UseTUI              bool = true
)

// Encryption constants
const (
	EncryptionKeyLength = 32 // 256-bit key
)

// InitEncryptionFlags initializes encryption-related flags for commands
func InitEncryptionFlags(cmd *cobra.Command) {
	// These flags are added to the root command for all subcommands
	cmd.PersistentFlags().BoolVar(&UseEncryption, "encrypt", false, "Encrypt data using AES-256")
	cmd.PersistentFlags().BoolVarP(&UseMaskedEncryption, "mask", "m", false, "Mask values (keep keys visible)")
	cmd.PersistentFlags().BoolVar(&UseKeyFile, "use-key-file", false, "Use key file instead of password")
	cmd.PersistentFlags().StringVarP(&EncryptionKeyFile, "key-file", "k", ".envi.key", "Path to encryption key file")
}

// IsEncrypted checks if content is encrypted with full encryption
func IsEncrypted(content []byte) bool {
	return bytes.HasPrefix(content, []byte(utils.EncryptionPrefix))
}

// IsMasked checks if content is encrypted with masked encryption
func IsMasked(content []byte) bool {
	return bytes.Contains(content, []byte(utils.MaskedPrefix))
}

// EncryptContent encrypts the given content using AES-256-GCM
func EncryptContent(content []byte) ([]byte, error) {
	// Get the encryption key
	key, err := getEncryptionKey()
	if err != nil {
		return nil, errors.New("failed to retrieve encryption key")
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.New("failed to create AES cipher block")
	}

	// Create a new GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.New("failed to create GCM")
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.New("failed to generate nonce")
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, content, nil)

	// Encode as base64 with prefix
	result := []byte(utils.EncryptionPrefix + base64.StdEncoding.EncodeToString(ciphertext))

	return result, nil
}

// DecryptContent decrypts the given content using AES-256-GCM
func DecryptContent(content []byte) ([]byte, error) {
	// Remove the prefix
	if !IsEncrypted(content) {
		return nil, errors.New("content is not encrypted or has invalid format")
	}
	cipherTextB64 := string(content)[len(utils.EncryptionPrefix):]

	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		return nil, errors.New("invalid encrypted data format")
	}

	// Get the encryption key
	key, err := getEncryptionKey()
	if err != nil {
		return nil, errors.New("failed to retrieve encryption key")
	}

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Verify ciphertext length
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("invalid encrypted data: ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decryption failed: invalid password or corrupted data")
	}

	return plaintext, nil
}

// MaskEnvContent masks the values in a .env file while keeping the keys visible
func MaskEnvContent(content []byte) ([]byte, error) {
	// Get the encryption key
	key, err := getEncryptionKey()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	// Pre-allocate with exact capacity to avoid resizing
	maskedLines := make([]string, 0, len(lines))

	// Create AES cipher block once for reuse
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM once for reuse
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		// Skip comments and empty lines
		if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
			maskedLines = append(maskedLines, line)
			continue
		}

		// Find the first equals sign (key=value)
		eqIdx := strings.Index(line, "=")
		if eqIdx == -1 {
			// Not a key=value line, keep as is
			maskedLines = append(maskedLines, line)
			continue
		}

		// Split into key and value
		k, v := line[:eqIdx+1], line[eqIdx+1:]

		// Encrypt the value
		if v == "" {
			// Empty value, no need to encrypt
			maskedLines = append(maskedLines, line)
			continue
		}

		// Create a nonce
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return nil, err
		}

		// Encrypt the value
		ciphertext := gcm.Seal(nonce, nonce, []byte(v), nil)

		// Encode as base64
		maskedValue := utils.MaskedPrefix + base64.StdEncoding.EncodeToString(ciphertext)

		// Add to masked lines
		maskedLines = append(maskedLines, k+maskedValue)
	}

	return []byte(strings.Join(maskedLines, "\n")), nil
}

// UnmaskEnvContent unmasks the values in a masked .env file
func UnmaskEnvContent(content []byte) ([]byte, error) {
	// Get the encryption key
	key, err := getEncryptionKey()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	// Pre-allocate with exact capacity to avoid resizing
	unmaskedLines := make([]string, 0, len(lines))

	// Create AES cipher block once for reuse
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Create GCM once for reuse
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()

	for _, line := range lines {
		// Skip comments and empty lines
		if strings.HasPrefix(strings.TrimSpace(line), "#") || strings.TrimSpace(line) == "" {
			unmaskedLines = append(unmaskedLines, line)
			continue
		}

		// Find the first equals sign (key=value)
		eqIdx := strings.Index(line, "=")
		if eqIdx == -1 {
			// Not a key=value line, keep as is
			unmaskedLines = append(unmaskedLines, line)
			continue
		}

		// Split into key and value
		k, v := line[:eqIdx+1], line[eqIdx+1:]

		// Check if value is masked
		if !strings.HasPrefix(v, utils.MaskedPrefix) {
			// Not masked, keep as is
			unmaskedLines = append(unmaskedLines, line)
			continue
		}

		// Remove prefix
		encryptedValue := v[len(utils.MaskedPrefix):]

		// Decode from base64
		ciphertext, err := base64.StdEncoding.DecodeString(encryptedValue)
		if err != nil {
			return nil, errors.New("invalid masked data format")
		}

		// Verify ciphertext length
		if len(ciphertext) < nonceSize {
			return nil, errors.New("invalid masked data: ciphertext too short")
		}

		// Extract nonce and ciphertext
		nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

		// Decrypt the value
		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return nil, errors.New("unmasking failed: invalid password or corrupted data")
		}

		// Add to unmasked lines
		unmaskedLines = append(unmaskedLines, k+string(plaintext))
	}

	return []byte(strings.Join(unmaskedLines, "\n")), nil
}

// getEncryptionKey gets the encryption key from password input or key file
func getEncryptionKey() ([]byte, error) {
	if UseKeyFile {
		// Use key file
		return getKeyFromFile()
	}

	// Use password
	if EncryptionPassword != "" {
		// Password provided in flag (not recommended)
		return hashPassword(EncryptionPassword), nil
	}

	// Get password from user
	var password string
	var err error

	if UseTUI {
		// Use TUI for password input
		password, err = tui.GetPassword("Enter encryption password", false)
		if err != nil {
			return nil, errors.New("failed to retrieve encryption password")
		}
	} else {
		// Use terminal input
		fmt.Print("Enter encryption password: ")
		passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return nil, err
		}
		fmt.Println()
		password = string(passwordBytes)
	}

	if password == "" {
		return nil, errors.New("password cannot be empty")
	}

	return hashPassword(password), nil
}

// getKeyFromFile reads the encryption key from a file
func getKeyFromFile() ([]byte, error) {
	keyData, err := os.ReadFile(EncryptionKeyFile)
	if err != nil {
		return nil, errors.New("failed to read encryption key file")
	}

	// Clean the key data
	key := bytes.TrimSpace(keyData)

	// Check if the key is base64 encoded
	decodedKey, err := base64.StdEncoding.DecodeString(string(key))
	if err == nil && len(decodedKey) == EncryptionKeyLength {
		// The key is valid base64 and has the correct length after decoding
		return decodedKey, nil
	}

	// Check if the key has the correct length directly
	if len(key) == EncryptionKeyLength {
		return key, nil
	}

	// Hash the key to get a fixed length key
	return hashPassword(string(key)), nil
}

// hashPassword creates a fixed-length encryption key from a password
func hashPassword(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

// GenerateKeyFile creates a new encryption key file with random data
func GenerateKeyFile(keyFilePath string) error {
	// Generate a random key
	key := make([]byte, EncryptionKeyLength)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return fmt.Errorf("failed to generate random key: %w", err)
	}

	// Encode as base64 for easier handling
	encodedKey := base64.StdEncoding.EncodeToString(key)

	// Create the key file with secure permissions
	if err := os.WriteFile(keyFilePath, []byte(encodedKey), 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}

// GenerateSelfContainedShare creates a self-contained encrypted share
// that includes the decryption key embedded in the content
func GenerateSelfContainedShare(content []byte, sharePassword string) ([]byte, error) {
	// Generate a random encryption key
	encryptionKey := make([]byte, EncryptionKeyLength)
	if _, err := io.ReadFull(rand.Reader, encryptionKey); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Encrypt the content with the random key
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, content, nil)

	// Derive a key from the share password to encrypt the encryption key
	derivedKey := deriveKeyFromPassword(sharePassword, utils.KeyDerivationSalt)

	// Encrypt the encryption key with the derived key
	keyBlock, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create key cipher: %w", err)
	}

	keyGCM, err := cipher.NewGCM(keyBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to create key GCM: %w", err)
	}

	keyNonce := make([]byte, keyGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, keyNonce); err != nil {
		return nil, fmt.Errorf("failed to generate key nonce: %w", err)
	}

	encryptedKey := keyGCM.Seal(keyNonce, keyNonce, encryptionKey, nil)

	// Combine encrypted key and encrypted content
	combined := append(encryptedKey, ciphertext...)

	// Encode as base64 with prefix
	result := []byte(utils.SelfContainedPrefix + base64.StdEncoding.EncodeToString(combined))

	return result, nil
}

// DecryptSelfContainedShare decrypts a self-contained encrypted share
func DecryptSelfContainedShare(encryptedContent []byte, sharePassword string) ([]byte, error) {
	// Remove the prefix
	if !bytes.HasPrefix(encryptedContent, []byte(utils.SelfContainedPrefix)) {
		return nil, errors.New("content is not a self-contained encrypted share")
	}

	// Decode from base64
	combined, err := base64.StdEncoding.DecodeString(string(encryptedContent)[len(utils.SelfContainedPrefix):])
	if err != nil {
		return nil, fmt.Errorf("invalid encrypted data format: %w", err)
	}

	// Derive key from password
	derivedKey := deriveKeyFromPassword(sharePassword, utils.KeyDerivationSalt)

	// Create cipher for key decryption
	keyBlock, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create key cipher: %w", err)
	}

	keyGCM, err := cipher.NewGCM(keyBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to create key GCM: %w", err)
	}

	// Extract encrypted key (first 48 bytes: 12 nonce + 32 key + 16 tag)
	keyNonceSize := keyGCM.NonceSize()
	keySize := EncryptionKeyLength
	keyTagSize := 16
	encryptedKeySize := keyNonceSize + keySize + keyTagSize

	if len(combined) < encryptedKeySize {
		return nil, errors.New("invalid encrypted data: too short")
	}

	encryptedKey := combined[:encryptedKeySize]
	ciphertext := combined[encryptedKeySize:]

	// Decrypt the encryption key
	keyNonce, encryptedKeyData := encryptedKey[:keyNonceSize], encryptedKey[keyNonceSize:]
	encryptionKey, err := keyGCM.Open(nil, keyNonce, encryptedKeyData, nil)
	if err != nil {
		return nil, errors.New("failed to decrypt encryption key: invalid password")
	}

	// Decrypt the content with the recovered encryption key
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create content cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create content GCM: %w", err)
	}

	// Extract nonce and ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("invalid encrypted data: ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the content
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("failed to decrypt content: corrupted data")
	}

	return plaintext, nil
}

// IsSelfContainedShare checks if content is a self-contained encrypted share
func IsSelfContainedShare(content []byte) bool {
	return bytes.HasPrefix(content, []byte(utils.SelfContainedPrefix))
}

// deriveKeyFromPassword derives a key from a password using PBKDF2
func deriveKeyFromPassword(password, salt string) []byte {
	// Use PBKDF2 to derive a key from the password
	key := pbkdf2.Key([]byte(password), []byte(salt), 10000, EncryptionKeyLength, sha256.New)
	return key
}
