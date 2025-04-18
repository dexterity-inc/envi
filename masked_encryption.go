package main

import (
	"bufio"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// MaskEnvContent processes a .env file and encrypts only the values while preserving keys
func MaskEnvContent(content []byte) ([]byte, error) {
	var encKey []byte
	var err error
	
	// Get encryption key based on user preference
	if useKeyFile {
		encKey, err = getKeyFromFile()
		if err != nil {
			if autoGenKey {
				encKey, err = generateAndSaveKey()
				if err != nil {
					return nil, fmt.Errorf("failed to generate encryption key: %v", err)
				}
				fmt.Println("Generated new encryption key and saved to file")
			} else {
				return nil, fmt.Errorf("failed to get encryption key from file: %v", err)
			}
		}
	} else {
		encKey, err = getKeyFromPassword()
		if err != nil {
			return nil, fmt.Errorf("failed to get encryption key from password: %v", err)
		}
	}

	// Parse the .env file line by line
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	
	var maskedLines []string
	lineNum := 0
	
	// Add header to indicate this file uses masked encryption
	maskedLines = append(maskedLines, "# ENVI_MASKED_ENCRYPTION_V1")
	
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		
		// Skip empty lines or comments
		if line == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			maskedLines = append(maskedLines, line)
			continue
		}
		
		// Find the key-value separator
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			// Not a key=value line, keep as is
			maskedLines = append(maskedLines, line)
			continue
		}
		
		keyName := parts[0]
		value := parts[1]
		
		// Skip encryption for empty values
		if strings.TrimSpace(value) == "" {
			maskedLines = append(maskedLines, line)
			continue
		}
		
		// Encrypt the value
		encryptedValue, err := encryptValue(value, encKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt value at line %d: %v", lineNum, err)
		}
		
		// Create masked line with format: KEY=ENVI_MASKED[base64_encrypted_value]
		maskedLine := fmt.Sprintf("%s=ENVI_MASKED[%s]", keyName, encryptedValue)
		maskedLines = append(maskedLines, maskedLine)
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading .env file: %v", err)
	}
	
	return []byte(strings.Join(maskedLines, "\n")), nil
}

// UnmaskEnvContent decrypts masked values in a .env file while preserving the structure
func UnmaskEnvContent(content []byte) ([]byte, error) {
	contentStr := string(content)
	
	// Check if the content is masked encrypted
	if !strings.Contains(contentStr, "# ENVI_MASKED_ENCRYPTION_V1") {
		return nil, fmt.Errorf("content is not masked encrypted or uses an unsupported format")
	}
	
	var encKey []byte
	var err error
	
	// Get decryption key based on user preference
	if useKeyFile {
		encKey, err = getKeyFromFile()
		if err != nil {
			return nil, fmt.Errorf("failed to get decryption key from file: %v", err)
		}
	} else {
		encKey, err = getKeyFromPassword()
		if err != nil {
			return nil, fmt.Errorf("failed to get decryption key from password: %v", err)
		}
	}
	
	// Parse the masked .env file line by line
	scanner := bufio.NewScanner(strings.NewReader(contentStr))
	
	var unmaskedLines []string
	lineNum := 0
	
	// Regular expression to match masked values
	maskedValueRegex := regexp.MustCompile(`=ENVI_MASKED\[(.*?)\]`)
	
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		
		// Skip the mask indicator line
		if strings.TrimSpace(line) == "# ENVI_MASKED_ENCRYPTION_V1" {
			continue
		}
		
		// Check if line contains a masked value
		matches := maskedValueRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			// Extract masked value part
			encodedValue := matches[1]
			
			// Split to get the key
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				unmaskedLines = append(unmaskedLines, line) // Keep as is if malformed
				continue
			}
			
			keyName := parts[0]
			
			// Decrypt the value
			decryptedValue, err := decryptValue(encodedValue, encKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt value at line %d: %v", lineNum, err)
			}
			
			// Create unmasked line
			unmaskedLine := fmt.Sprintf("%s=%s", keyName, decryptedValue)
			unmaskedLines = append(unmaskedLines, unmaskedLine)
		} else {
			// Not a masked line, keep as is
			unmaskedLines = append(unmaskedLines, line)
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading masked .env file: %v", err)
	}
	
	return []byte(strings.Join(unmaskedLines, "\n")), nil
}

// IsMaskedEncrypted checks if content uses masked encryption
func IsMaskedEncrypted(content []byte) bool {
	return strings.Contains(string(content), "# ENVI_MASKED_ENCRYPTION_V1")
}

// encryptValue encrypts a single value
func encryptValue(value string, key []byte) (string, error) {
	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}
	
	// Create a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}
	
	// Encrypt the value
	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)
	
	// Base64 encode the result
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptValue decrypts a single value
func decryptValue(encodedValue string, key []byte) (string, error) {
	// Decode base64
	ciphertext, err := base64.StdEncoding.DecodeString(encodedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted value: %v", err)
	}
	
	// Create cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %v", err)
	}
	
	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %v", err)
	}
	
	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	
	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %v", err)
	}
	
	return string(plaintext), nil
} 