package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

// Config stores application configuration
type Config struct {
	GitHubToken         string               `yaml:"github_token,omitempty"`
	LastGistID          string               `yaml:"last_gist_id,omitempty"`
	TokenInKeyring      bool                 `yaml:"token_in_keyring"`
	EncryptByDefault    bool                 `yaml:"encrypt_by_default"`
	UseMaskedEncryption bool                 `yaml:"use_masked_encryption"`
	UnmaskByDefault     bool                 `yaml:"unmask_by_default"`
	DefaultKeyFile      string               `yaml:"default_key_file,omitempty"`
	UseKeyFileByDefault bool                 `yaml:"use_key_file_by_default"`
	GistHistory         map[string]*GistInfo `yaml:"gist_history,omitempty"`
	Projects            map[string]*Project  `yaml:"projects,omitempty"`
}

// GistInfo stores enhanced metadata about a gist
type GistInfo struct {
	ID          string   `yaml:"id"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	ProjectName string   `yaml:"project_name,omitempty"`
	Environment string   `yaml:"environment,omitempty"`
	CreatedAt   string   `yaml:"created_at"`
	UpdatedAt   string   `yaml:"updated_at"`
	Tags        []string `yaml:"tags,omitempty"`
	IsEncrypted bool     `yaml:"is_encrypted"`
	IsPublic    bool     `yaml:"is_public"`
	UsageCount  int      `yaml:"usage_count"`
	LastUsed    string   `yaml:"last_used,omitempty"`
	FileCount   int      `yaml:"file_count"`
	URL         string   `yaml:"url"`
}

// Project stores information about a project
type Project struct {
	Name         string   `yaml:"name"`
	Path         string   `yaml:"path"`
	Environments []string `yaml:"environments,omitempty"`
	CreatedAt    string   `yaml:"created_at"`
	LastUsed     string   `yaml:"last_used,omitempty"`
	GistIDs      []string `yaml:"gist_ids,omitempty"`
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

// AddGistToHistory adds or updates a gist in the history
func (c *Config) AddGistToHistory(gistInfo *GistInfo) {
	if c.GistHistory == nil {
		c.GistHistory = make(map[string]*GistInfo)
	}
	c.GistHistory[gistInfo.ID] = gistInfo
}

// GetGistInfo retrieves gist information from history
func (c *Config) GetGistInfo(gistID string) (*GistInfo, bool) {
	if c.GistHistory == nil {
		return nil, false
	}
	gistInfo, exists := c.GistHistory[gistID]
	return gistInfo, exists
}

// UpdateGistUsage updates the usage statistics for a gist
func (c *Config) UpdateGistUsage(gistID string) {
	if c.GistHistory == nil {
		return
	}

	if gistInfo, exists := c.GistHistory[gistID]; exists {
		gistInfo.UsageCount++
		gistInfo.LastUsed = time.Now().Format("2006-01-02 15:04:05")
	}
}

// AddProject adds or updates a project in the configuration
func (c *Config) AddProject(project *Project) {
	if c.Projects == nil {
		c.Projects = make(map[string]*Project)
	}
	c.Projects[project.Name] = project
}

// GetProject retrieves project information
func (c *Config) GetProject(projectName string) (*Project, bool) {
	if c.Projects == nil {
		return nil, false
	}
	project, exists := c.Projects[projectName]
	return project, exists
}

// GetProjectByPath retrieves project information by path
func (c *Config) GetProjectByPath(path string) (*Project, bool) {
	if c.Projects == nil {
		return nil, false
	}

	for _, project := range c.Projects {
		if project.Path == path {
			return project, true
		}
	}
	return nil, false
}

// GenerateGistName generates an enhanced name for a gist
func GenerateGistName(envFile, projectName, environment string) string {
	// Get the base filename without extension
	baseName := strings.TrimSuffix(filepath.Base(envFile), filepath.Ext(envFile))

	// If it's just "env", use a more descriptive name
	if baseName == "env" {
		baseName = "environment"
	}

	// Build the name components
	var nameParts []string

	// Add project name if available
	if projectName != "" {
		nameParts = append(nameParts, projectName)
	}

	// Add environment type if available
	if environment != "" {
		nameParts = append(nameParts, environment)
	} else if baseName != "environment" {
		// Try to extract environment from filename
		if strings.Contains(baseName, ".") {
			parts := strings.Split(baseName, ".")
			if len(parts) > 1 {
				envPart := parts[len(parts)-1]
				if envPart != "env" && envPart != "environment" {
					nameParts = append(nameParts, envPart)
				}
			}
		}
	}

	// Add timestamp for uniqueness
	timestamp := time.Now().Format("2006-01-02")
	nameParts = append(nameParts, timestamp)

	// Join all parts
	if len(nameParts) > 0 {
		return strings.Join(nameParts, " - ")
	}

	// Fallback
	return fmt.Sprintf("Environment Variables - %s", timestamp)
}

// GenerateGistDescription generates an enhanced description for a gist
func GenerateGistDescription(envFile, projectName, environment string, isEncrypted bool) string {
	var descParts []string

	// Add project context
	if projectName != "" {
		descParts = append(descParts, fmt.Sprintf("Project: %s", projectName))
	}

	// Add environment context
	if environment != "" {
		descParts = append(descParts, fmt.Sprintf("Environment: %s", environment))
	}

	// Add file context
	descParts = append(descParts, fmt.Sprintf("File: %s", filepath.Base(envFile)))

	// Add encryption status
	if isEncrypted {
		descParts = append(descParts, "Status: Encrypted")
	}

	// Add creation timestamp
	descParts = append(descParts, fmt.Sprintf("Created: %s", time.Now().Format("2006-01-02 15:04:05")))

	// Add tool attribution
	descParts = append(descParts, "Tool: envi")

	return strings.Join(descParts, " | ")
}

// GetEnvironmentFromFilename extracts environment type from filename
func GetEnvironmentFromFilename(filename string) string {
	baseName := filepath.Base(filename)

	// Handle common patterns
	if strings.Contains(baseName, ".env.") {
		parts := strings.Split(baseName, ".")
		if len(parts) >= 3 {
			return parts[2] // .env.environment -> environment
		}
	}

	// Handle other patterns
	if strings.Contains(baseName, "production") || strings.Contains(baseName, "prod") {
		return "production"
	}
	if strings.Contains(baseName, "development") || strings.Contains(baseName, "dev") {
		return "development"
	}
	if strings.Contains(baseName, "staging") || strings.Contains(baseName, "stage") {
		return "staging"
	}
	if strings.Contains(baseName, "test") || strings.Contains(baseName, "testing") {
		return "testing"
	}

	return ""
}
