package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	
	"golang.org/x/term"
)

var (
	// Global variables for encryption (shared across commands)
	encryptionKeyFile string
	useKeyFile        bool
	autoGenKey        bool
	useEncryption     bool
	useMaskedEncryption bool
	encryptionPassword string
	// disableEncryption is declared in config_cmd.go
)

func init() {
	// Empty init function - flags will be registered in initEncryptionFlags()
}

// initEncryptionFlags registers all encryption-related flags for push and pull commands
func initEncryptionFlags() {
	// Register flags for the push command if it exists
	if pushCmd != nil {
		pushCmd.Flags().BoolVarP(&useEncryption, "encrypt", "e", false, "Encrypt the .env file before pushing")
		pushCmd.Flags().BoolVarP(&useMaskedEncryption, "mask", "m", false, "Mask the values in the .env file before pushing")
		pushCmd.Flags().BoolVar(&useKeyFile, "use-key-file", false, "Use a key file for encryption instead of password")
		pushCmd.Flags().StringVarP(&encryptionKeyFile, "key-file", "k", ".envi.key", "Path to encryption key file")
		pushCmd.Flags().StringVarP(&encryptionPassword, "password", "p", "", "Password for encryption (not recommended, use key file instead)")
		pushCmd.Flags().BoolVar(&autoGenKey, "generate-key", false, "Auto-generate a strong encryption key and save to key file")
	}

	// Register flags for the pull command if it exists
	if pullCmd != nil {
		pullCmd.Flags().BoolVarP(&useEncryption, "decrypt", "d", false, "Decrypt the .env file after pulling")
		pullCmd.Flags().BoolVar(&useMaskedEncryption, "unmask", false, "Unmask values in .env file (for masked encryption)")
		pullCmd.Flags().BoolVar(&useKeyFile, "use-key-file", false, "Use a key file for decryption instead of password")
		pullCmd.Flags().StringVarP(&encryptionKeyFile, "key-file", "k", ".envi.key", "Path to decryption key file")
		pullCmd.Flags().StringVarP(&encryptionPassword, "password", "p", "", "Password for decryption (not recommended, use key file instead)")
	}
}

// EncryptContent encrypts the given content with either a password or key file
func EncryptContent(content []byte) ([]byte, error) {
	var key []byte
	var err error
	
	if useKeyFile {
		key, err = getKeyFromFile()
		if err != nil {
			if autoGenKey {
				key, err = generateAndSaveKey()
				if err != nil {
					return nil, fmt.Errorf("failed to generate encryption key: %v", err)
				}
				fmt.Println("Generated new encryption key and saved to file")
			} else {
				return nil, fmt.Errorf("failed to get encryption key from file: %v", err)
			}
		}
	} else {
		key, err = getKeyFromPassword()
		if err != nil {
			return nil, fmt.Errorf("failed to get encryption key from password: %v", err)
		}
	}
	
	// Create cipher block using the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}
	
	// Create a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}
	
	// Encrypt and prepend nonce
	ciphertext := gcm.Seal(nonce, nonce, content, nil)
	
	// Base64 encode the result for safe storage
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	
	// Add a header to identify this as an encrypted file
	return []byte("ENVI_ENCRYPTED_V1\n" + encoded), nil
}

// DecryptContent decrypts the given content with either a password or key file
func DecryptContent(content []byte) ([]byte, error) {
	// Check if the content is encrypted
	contentStr := string(content)
	if !strings.HasPrefix(contentStr, "ENVI_ENCRYPTED_V1\n") {
		return nil, fmt.Errorf("content is not encrypted or uses an unsupported format")
	}
	
	// Extract the base64 encoded ciphertext
	encoded := strings.TrimPrefix(contentStr, "ENVI_ENCRYPTED_V1\n")
	
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted content: %v", err)
	}
	
	var key []byte
	
	if useKeyFile {
		key, err = getKeyFromFile()
		if err != nil {
			return nil, fmt.Errorf("failed to get decryption key from file: %v", err)
		}
	} else {
		key, err = getKeyFromPassword()
		if err != nil {
			return nil, fmt.Errorf("failed to get decryption key from password: %v", err)
		}
	}
	
	// Create cipher block using the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %v", err)
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}
	
	// Extract nonce from ciphertext
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	
	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %v", err)
	}
	
	return plaintext, nil
}

// IsEncrypted checks if the content is encrypted
func IsEncrypted(content []byte) bool {
	return strings.HasPrefix(string(content), "ENVI_ENCRYPTED_V1\n")
}

// getKeyFromPassword prompts for a password and derives a key
func getKeyFromPassword() ([]byte, error) {
	// Check if password was provided directly through flag (not recommended)
	if encryptionPassword != "" {
		fmt.Println("Warning: Using password provided through command line is insecure.")
		fmt.Println("Consider using interactive password input or key files instead.")
		
		if len(encryptionPassword) < 8 {
			return nil, fmt.Errorf("password must be at least 8 characters long")
		}
		
		// Use SHA-256 to create a fixed-size key from the password
		key := sha256.Sum256([]byte(encryptionPassword))
		return key[:], nil
	}

	// Interactive password input
	fmt.Print("Enter encryption password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // Add newline after password input
	
	if err != nil {
		return nil, fmt.Errorf("failed to read password: %v", err)
	}
	
	if len(password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters long")
	}
	
	// For new passwords (when encrypting), request confirmation
	if useEncryption || useMaskedEncryption {
		fmt.Print("Confirm encryption password: ")
		passwordConfirm, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println() // Add newline after password input
		
		if err != nil {
			return nil, fmt.Errorf("failed to read password confirmation: %v", err)
		}
		
		if string(password) != string(passwordConfirm) {
			return nil, fmt.Errorf("passwords do not match")
		}
	}
	
	// Use SHA-256 to create a fixed-size key from the password
	key := sha256.Sum256(password)
	return key[:], nil
}

// getKeyFromFile reads the key from a file
func getKeyFromFile() ([]byte, error) {
	// Expand ~ in file path if present
	if strings.HasPrefix(encryptionKeyFile, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %v", err)
		}
		encryptionKeyFile = filepath.Join(homeDir, encryptionKeyFile[2:])
	}
	
	// Check if file exists
	info, err := os.Stat(encryptionKeyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("key file not found: %s", encryptionKeyFile)
		}
		return nil, fmt.Errorf("failed to access key file: %v", err)
	}
	
	// Check file permissions
	perm := info.Mode().Perm()
	if perm&0077 != 0 {
		fmt.Printf("Warning: Key file %s has loose permissions: %o\n", encryptionKeyFile, perm)
		fmt.Println("Fixing permissions to 0600 (user read/write only)...")
		
		// Try to fix permissions
		if err := os.Chmod(encryptionKeyFile, 0600); err != nil {
			fmt.Printf("Error fixing permissions: %v\n", err)
			fmt.Println("Please fix manually for security: chmod 600", encryptionKeyFile)
		} else {
			fmt.Println("Key file permissions fixed.")
		}
	}
	
	// Read key from file
	key, err := os.ReadFile(encryptionKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %v", err)
	}
	
	// Verify key length
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid key file: expected 32 bytes, got %d", len(key))
	}
	
	return key, nil
}

// generateAndSaveKey generates a new random key and saves it to a file
func generateAndSaveKey() ([]byte, error) {
	// Expand ~ in file path if present
	if strings.HasPrefix(encryptionKeyFile, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %v", err)
		}
		encryptionKeyFile = filepath.Join(homeDir, encryptionKeyFile[2:])
	}
	
	// Generate random key (32 bytes for AES-256)
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %v", err)
	}
	
	// Check if file already exists
	if _, err := os.Stat(encryptionKeyFile); err == nil {
		return nil, fmt.Errorf("key file already exists: %s", encryptionKeyFile)
	}
	
	// Save key to file with restricted permissions
	if err := os.WriteFile(encryptionKeyFile, key, 0600); err != nil {
		return nil, fmt.Errorf("failed to save key to file: %v", err)
	}
	
	fmt.Printf("Encryption key saved to: %s\n", encryptionKeyFile)
	fmt.Println("IMPORTANT: Keep this key file safe and secure. You will need it to decrypt your .env files.")
	fmt.Println("Consider backing up this key file in a secure location.")
	
	return key, nil
} 