package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	
	"github.com/zalando/go-keyring"
)

const (
	// Service name for keyring
	keyringServiceName = "envi-cli"
	// Account name for token in keyring
	keyringTokenAccount = "github-token"
)

type Config struct {
	// GitHubToken is kept for backward compatibility but will be migrated to secure storage
	GitHubToken string `json:"github_token,omitempty"`
	// LastGistID stores the ID of the last used Gist
	LastGistID  string `json:"last_gist_id,omitempty"`
	// Flag to indicate if token is stored in keyring
	TokenInKeyring bool `json:"token_in_keyring,omitempty"`
	// Flag to enable encryption by default
	EncryptByDefault bool `json:"encrypt_by_default,omitempty"`
	// Flag to use masked encryption by default
	UseMaskedEncryption bool `json:"use_masked_encryption,omitempty"`
	// Default key file path
	DefaultKeyFile string `json:"default_key_file,omitempty"`
	// Flag to use key file by default instead of password
	UseKeyFileByDefault bool `json:"use_key_file_by_default,omitempty"`
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %v", err)
	}

	// Create the .envi directory if it doesn't exist
	configDir := filepath.Join(homeDir, ".envi")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory: %v", err)
	}

	return filepath.Join(configDir, "config.json"), nil
}

// verifyConfigPermissions checks and fixes file permissions if needed
func verifyConfigPermissions(configPath string) error {
	info, err := os.Stat(configPath)
	if err != nil {
		return err
	}
	
	// Check if permissions are too open (not 0600)
	if info.Mode().Perm() != 0600 {
		fmt.Fprintf(os.Stderr, "Warning: Config file has insecure permissions: %o\n", info.Mode().Perm())
		// Fix permissions
		if err := os.Chmod(configPath, 0600); err != nil {
			return fmt.Errorf("failed to fix config file permissions: %v", err)
		}
		fmt.Fprintf(os.Stderr, "Fixed config file permissions to 0600\n")
	}
	return nil
}

// secureWipeFile securely overwrites sensitive data in a file
func secureWipeFile(filePath string) error {
	// Get file info to determine size
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	
	// Open file for writing
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Overwrite with zeros
	zeroBytes := make([]byte, 1024)
	remaining := info.Size()
	
	for remaining > 0 {
		writeSize := int64(len(zeroBytes))
		if remaining < writeSize {
			writeSize = remaining
		}
		
		_, err := file.Write(zeroBytes[:writeSize])
		if err != nil {
			return err
		}
		
		remaining -= writeSize
	}
	
	// Sync to ensure it's written to disk
	return file.Sync()
}

// LoadConfig reads the configuration from disk
func LoadConfig() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	// If config doesn't exist, return default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	// Verify and fix file permissions if needed
	if err := verifyConfigPermissions(configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not verify config file permissions: %v\n", err)
		// Continue anyway, not fatal
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Migrate token to keyring if it exists in the config file
	if config.GitHubToken != "" && !config.TokenInKeyring {
		// Validate token before migrating
		if isValidGitHubToken(config.GitHubToken) {
			err := SaveTokenToKeyring(config.GitHubToken)
			if err == nil {
				// If successfully migrated to keyring, update config
				// Securely wipe the old token from the config file
				tempConfig := config
				tempConfig.GitHubToken = strings.Repeat("0", len(config.GitHubToken)) // Overwrite with zeros
				if err := SaveConfig(&tempConfig); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to securely wipe token during migration: %v\n", err)
				}
				
				// Now update config to remove token and set flag
				config.GitHubToken = ""
				config.TokenInKeyring = true
				if err := SaveConfig(&config); err != nil {
					// Non-fatal, just log the error
					fmt.Fprintf(os.Stderr, "Warning: Failed to update config after token migration: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "Successfully migrated GitHub token to secure storage\n")
				}
			} else {
				fmt.Fprintf(os.Stderr, "Warning: Could not migrate token to secure storage: %v\n", err)
				fmt.Fprintf(os.Stderr, "Your token will remain in the config file. For better security, consider using a system with keyring support.\n")
			}
		} else {
			fmt.Fprintf(os.Stderr, "Warning: Stored GitHub token has invalid format and will not be migrated\n")
		}
	}

	return &config, nil
}

// SaveConfig writes the configuration to disk
func SaveConfig(config *Config) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode config: %v", err)
	}

	// Ensure file permissions are restricted to the current user
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// isValidGitHubToken checks if the provided token has valid format
func isValidGitHubToken(token string) bool {
	// Basic length check - all GitHub tokens should be at least 30 characters
	if len(token) < 30 {
		return false
	}
	
	// Check for newer personal access token (PAT) formats with prefixes
	// Fine-grained tokens: github_pat_
	// Classic tokens with scope prefixes: ghp_, gho_, ghu_, ghs_
	if strings.HasPrefix(token, "github_pat_") ||
	   strings.HasPrefix(token, "ghp_") || 
	   strings.HasPrefix(token, "gho_") || 
	   strings.HasPrefix(token, "ghu_") || 
	   strings.HasPrefix(token, "ghs_") {
		return true
	}
	
	// Check classic token format (40 char hex)
	hexPattern := regexp.MustCompile(`^[0-9a-f]{40}$`)
	if hexPattern.MatchString(token) {
		return true
	}

	// Support for other potential formats that meet length requirements
	if len(token) >= 36 {
		return true
	}
	
	return false
}

// SaveTokenToKeyring stores the GitHub token in the system's secure credential store
func SaveTokenToKeyring(token string) error {
	// Validate token format before storing
	if !isValidGitHubToken(token) {
		return fmt.Errorf("invalid GitHub token format")
	}
	
	err := keyring.Set(keyringServiceName, keyringTokenAccount, token)
	if err != nil {
		// Provide more informative error based on platform
		switch os.Getenv("OS") {
		case "Windows_NT":
			return fmt.Errorf("failed to store token in Windows Credential Manager: %v", err)
		case "Darwin":
			return fmt.Errorf("failed to store token in macOS Keychain: %v", err)
		default:
			return fmt.Errorf("failed to store token in secure storage (libsecret/kwallet): %v", err)
		}
	}
	return nil
}

// GetTokenFromKeyring retrieves the GitHub token from the system's secure credential store
func GetTokenFromKeyring() (string, error) {
	token, err := keyring.Get(keyringServiceName, keyringTokenAccount)
	if err != nil {
		// Provide more informative error based on platform
		var errMsg string
		switch os.Getenv("OS") {
		case "Windows_NT":
			errMsg = fmt.Sprintf("failed to access Windows Credential Manager: %v", err)
		case "Darwin":
			errMsg = fmt.Sprintf("failed to access macOS Keychain: %v", err)
		default:
			errMsg = fmt.Sprintf("failed to access secure storage (libsecret/kwallet): %v", err)
		}
		return "", fmt.Errorf(errMsg)
	}
	
	// Validate token after retrieving
	if !isValidGitHubToken(token) {
		return "", fmt.Errorf("retrieved token has invalid format, may be corrupted")
	}
	
	return token, nil
}

// DeleteTokenFromKeyring removes the GitHub token from the system's secure credential store
func DeleteTokenFromKeyring() error {
	err := keyring.Delete(keyringServiceName, keyringTokenAccount)
	if err != nil {
		// Provide more informative error based on platform
		switch os.Getenv("OS") {
		case "Windows_NT":
			return fmt.Errorf("failed to delete token from Windows Credential Manager: %v", err)
		case "Darwin":
			return fmt.Errorf("failed to delete token from macOS Keychain: %v", err)
		default:
			return fmt.Errorf("failed to delete token from secure storage (libsecret/kwallet): %v", err)
		}
	}
	return nil
}

// GetGitHubToken retrieves the GitHub token from various sources in order of preference:
// 1. Environment variable
// 2. System keyring
// 3. Config file (legacy)
func GetGitHubToken() (string, error) {
	// First check environment variable (highest priority)
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		// Validate token from environment
		if !isValidGitHubToken(token) {
			return "", fmt.Errorf("GITHUB_TOKEN environment variable contains an invalid token format")
		}
		return token, nil
	}

	// Load config to check if token is in keyring
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}

	// Try to get from keyring if config indicates it's there
	if config.TokenInKeyring {
		token, err := GetTokenFromKeyring()
		if err == nil && token != "" {
			return token, nil
		}
		// If there was an error, log it and fall back to config file check
		fmt.Fprintf(os.Stderr, "Warning: Could not retrieve token from secure storage: %v\n", err)
		fmt.Fprintf(os.Stderr, "Falling back to config file check...\n")
	}

	// Legacy: check config file as fallback
	if config.GitHubToken != "" {
		// Validate token from config
		if !isValidGitHubToken(config.GitHubToken) {
			return "", fmt.Errorf("stored GitHub token has invalid format")
		}
		return config.GitHubToken, nil
	}

	return "", fmt.Errorf("GitHub token not found. Set GITHUB_TOKEN environment variable or run 'envi config --token your-token'")
} 