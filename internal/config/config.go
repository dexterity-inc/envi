package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

// Config stores application configuration
type Config struct {
	GitHubToken         string `yaml:"github_token,omitempty"`
	LastGistID          string `yaml:"last_gist_id,omitempty"`
	TokenInKeyring      bool   `yaml:"token_in_keyring"`
	EncryptByDefault    bool   `yaml:"encrypt_by_default"`
	UseMaskedEncryption bool   `yaml:"use_masked_encryption"`
	UnmaskByDefault     bool   `yaml:"unmask_by_default"`
	DefaultKeyFile      string `yaml:"default_key_file,omitempty"`
	UseKeyFileByDefault bool   `yaml:"use_key_file_by_default"`
}

const (
	// App constants for keyring
	applicationName = "envi-cli"
	tokenUsername   = "github-token"
	
	// Default file permissions for config
	configFilePerms = 0600
)

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, ".envi")
	configPath := filepath.Join(configDir, "config.yaml")
	
	return configPath, nil
}

// EnsureConfigDir ensures the config directory exists
func EnsureConfigDir() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, ".envi")
	
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("error creating config directory: %w", err)
		}
	}
	
	return nil
}

// LoadConfig loads the configuration from disk
func LoadConfig() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	
	// Create default config if no file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config
		defaultConfig := &Config{
			EncryptByDefault:    true,
			UseMaskedEncryption: true,
		}
		
		// Ensure the config directory exists
		if err := EnsureConfigDir(); err != nil {
			return nil, err
		}
		
		// Save default config
		if err := SaveConfig(defaultConfig); err != nil {
			return nil, err
		}
		
		return defaultConfig, nil
	}
	
	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}
	
	// Unmarshal the YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}
	
	// Verify file permissions
	verifyConfigPermissions(configPath)
	
	return &config, nil
}

// SaveConfig saves the configuration to disk
func SaveConfig(config *Config) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}
	
	// Ensure the config directory exists
	if err := EnsureConfigDir(); err != nil {
		return err
	}
	
	// Marshal the YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error encoding config: %w", err)
	}
	
	// Write the file with secure permissions
	if err := os.WriteFile(configPath, data, configFilePerms); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}
	
	return nil
}

// GetGitHubToken fetches the GitHub token, trying environment variable, then keyring, then config file
func GetGitHubToken() (string, error) {
	// First try environment variable
	envToken := os.Getenv("GITHUB_TOKEN")
	if envToken != "" {
		if !IsValidGitHubToken(envToken) {
			return "", errors.New("GitHub token from environment variable has invalid format")
		}
		return envToken, nil
	}
	
	// Load config
	config, err := LoadConfig()
	if err != nil {
		return "", fmt.Errorf("error loading config: %w", err)
	}
	
	// Try keyring if configured
	if config.TokenInKeyring {
		token, err := GetTokenFromKeyring()
		if err == nil {
			return token, nil
		}
	}
	
	// Try token from config file
	if config.GitHubToken != "" {
		if !IsValidGitHubToken(config.GitHubToken) {
			return "", errors.New("GitHub token in config file has invalid format")
		}
		return config.GitHubToken, nil
	}
	
	return "", errors.New("no GitHub token found. Use 'envi config --token YOUR_TOKEN' to set one")
}

// SaveTokenToKeyring saves the GitHub token to the system keyring
func SaveTokenToKeyring(token string) error {
	return keyring.Set(applicationName, tokenUsername, token)
}

// GetTokenFromKeyring retrieves the GitHub token from the system keyring
func GetTokenFromKeyring() (string, error) {
	return keyring.Get(applicationName, tokenUsername)
}

// DeleteTokenFromKeyring removes the GitHub token from the system keyring
func DeleteTokenFromKeyring() error {
	return keyring.Delete(applicationName, tokenUsername)
}

// IsValidGitHubToken checks if a token is a valid GitHub PAT format
func IsValidGitHubToken(token string) bool {
	// GitHub Personal Access Tokens are at least 40 characters
	if len(token) < 30 {
		return false
	}
	
	// Matches the format of GitHub tokens
	// Classic PATs: ghp_*
	// Fine-grained PATs: github_pat_*
	// OAuth tokens: gho_*
	// User-to-server tokens: ghu_*
	// Server-to-server tokens: ghs_*
	validPrefixes := []string{"ghp_", "github_pat_", "gho_", "ghu_", "ghs_"}
	
	// Also allow the old format tokens that are just hex
	hexRegex := regexp.MustCompile(`^[a-f0-9]{40}$`)
	
	// Check if it has a valid prefix
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(token, prefix) {
			return true
		}
	}
	
	// Check if it's a valid old-style token
	return hexRegex.MatchString(token)
}

// verifyConfigPermissions checks and warns about insecure file permissions
func verifyConfigPermissions(configPath string) {
	info, err := os.Stat(configPath)
	if err != nil {
		return // Ignore errors here
	}
	
	// Check if permissions are too open
	if info.Mode().Perm() != configFilePerms {
		fmt.Printf("Warning: Config file has insecure permissions: %o\n", info.Mode().Perm())
		fmt.Printf("Run 'chmod 600 %s' to fix\n", configPath)
	}
} 