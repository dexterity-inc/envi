package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	
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
		err := SaveTokenToKeyring(config.GitHubToken)
		if err == nil {
			// If successfully migrated to keyring, update config
			config.GitHubToken = ""
			config.TokenInKeyring = true
			if err := SaveConfig(&config); err != nil {
				// Non-fatal, just log the error
				fmt.Fprintf(os.Stderr, "Warning: Failed to update config after token migration: %v\n", err)
			}
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

// SaveTokenToKeyring stores the GitHub token in the system's secure credential store
func SaveTokenToKeyring(token string) error {
	err := keyring.Set(keyringServiceName, keyringTokenAccount, token)
	if err != nil {
		return fmt.Errorf("failed to store token in secure storage: %v", err)
	}
	return nil
}

// GetTokenFromKeyring retrieves the GitHub token from the system's secure credential store
func GetTokenFromKeyring() (string, error) {
	token, err := keyring.Get(keyringServiceName, keyringTokenAccount)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve token from secure storage: %v", err)
	}
	return token, nil
}

// DeleteTokenFromKeyring removes the GitHub token from the system's secure credential store
func DeleteTokenFromKeyring() error {
	err := keyring.Delete(keyringServiceName, keyringTokenAccount)
	if err != nil {
		return fmt.Errorf("failed to delete token from secure storage: %v", err)
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
		// If there was an error, fall back to config file check
	}

	// Legacy: check config file as fallback
	if config.GitHubToken != "" {
		return config.GitHubToken, nil
	}

	return "", fmt.Errorf("GitHub token not found. Set GITHUB_TOKEN environment variable or run 'envi config --token your-token'")
} 