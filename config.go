package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	GitHubToken string `json:"github_token"`
	LastGistID  string `json:"last_gist_id,omitempty"`
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

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// GetGitHubToken retrieves the GitHub token from config or environment
func GetGitHubToken() (string, error) {
	// First check environment variable
	token := os.Getenv("GITHUB_TOKEN")
	if token != "" {
		return token, nil
	}

	// If not in environment, check config file
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}

	if config.GitHubToken == "" {
		return "", fmt.Errorf("GitHub token not found. Set GITHUB_TOKEN environment variable or run 'envi config --token your-token'")
	}

	return config.GitHubToken, nil
} 