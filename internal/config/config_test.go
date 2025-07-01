package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestConfigPath(t *testing.T) {
	// Test that ConfigPath returns a valid path
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error = %v", err)
	}

	if path == "" {
		t.Error("ConfigPath() returned empty path")
	}

	// Verify path contains expected components
	if !containsPath(path, ".envi") || !containsPath(path, "config.yaml") {
		t.Errorf("ConfigPath() = %v, expected to contain .envi and config.yaml", path)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	// Create temporary home directory for testing
	tempHome, err := os.MkdirTemp("", "config_test_home")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Set HOME environment variable temporarily
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Test EnsureConfigDir
	err = EnsureConfigDir()
	if err != nil {
		t.Fatalf("EnsureConfigDir() error = %v", err)
	}

	// Verify directory was created
	configDir := filepath.Join(tempHome, ".envi")
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		t.Error("Config directory was not created")
	}

	// Test that calling it again doesn't error
	err = EnsureConfigDir()
	if err != nil {
		t.Errorf("EnsureConfigDir() second call error = %v", err)
	}
}

func TestLoadConfig_NewConfig(t *testing.T) {
	// Create temporary home directory
	tempHome, err := os.MkdirTemp("", "config_test_new")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Set HOME environment variable temporarily
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Test loading config when no file exists (should create default)
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Verify default values
	if !cfg.EncryptByDefault {
		t.Error("Default config should have EncryptByDefault = true")
	}

	if !cfg.UseMaskedEncryption {
		t.Error("Default config should have UseMaskedEncryption = true")
	}

	// Verify config file was created
	configPath := filepath.Join(tempHome, ".envi", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Config file was not created")
	}
}

func TestSaveConfig(t *testing.T) {
	// Create temporary home directory
	tempHome, err := os.MkdirTemp("", "config_test_save")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}
	defer os.RemoveAll(tempHome)

	// Set HOME environment variable temporarily
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	// Create test config
	cfg := &Config{
		GitHubToken:         "test_token",
		EncryptByDefault:    false,
		UseMaskedEncryption: false,
		LastGistID:          "test_gist_id",
	}

	// Save config
	err = SaveConfig(cfg)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// Load config back and verify
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() after save error = %v", err)
	}

	if loadedConfig.GitHubToken != cfg.GitHubToken {
		t.Errorf("GitHubToken = %v, want %v", loadedConfig.GitHubToken, cfg.GitHubToken)
	}

	if loadedConfig.EncryptByDefault != cfg.EncryptByDefault {
		t.Errorf("EncryptByDefault = %v, want %v", loadedConfig.EncryptByDefault, cfg.EncryptByDefault)
	}

	if loadedConfig.UseMaskedEncryption != cfg.UseMaskedEncryption {
		t.Errorf("UseMaskedEncryption = %v, want %v", loadedConfig.UseMaskedEncryption, cfg.UseMaskedEncryption)
	}

	if loadedConfig.LastGistID != cfg.LastGistID {
		t.Errorf("LastGistID = %v, want %v", loadedConfig.LastGistID, cfg.LastGistID)
	}
}

func TestIsValidGitHubToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  bool
	}{
		{
			name:  "valid classic PAT",
			token: "ghp_1234567890abcdef1234567890abcdef12345678",
			want:  true,
		},
		{
			name:  "valid fine-grained PAT",
			token: "github_pat_12345678901234567890123456789012345678901234567890",
			want:  true,
		},
		{
			name:  "valid OAuth token",
			token: "gho_1234567890abcdef1234567890abcdef12345678",
			want:  true,
		},
		{
			name:  "valid user-to-server token",
			token: "ghu_1234567890abcdef1234567890abcdef12345678",
			want:  true,
		},
		{
			name:  "valid server-to-server token",
			token: "ghs_1234567890abcdef1234567890abcdef12345678",
			want:  true,
		},
		{
			name:  "valid old-style hex token",
			token: "1234567890abcdef1234567890abcdef12345678",
			want:  true,
		},
		{
			name:  "too short token",
			token: "ghp_short",
			want:  false,
		},
		{
			name:  "invalid prefix",
			token: "invalid_1234567890abcdef1234567890abcdef12345678",
			want:  false,
		},
		{
			name:  "empty token",
			token: "",
			want:  false,
		},
		{
			name:  "invalid hex token",
			token: "ghz_1234567890abcdef1234567890abcdef12345678",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidGitHubToken(tt.token)
			if got != tt.want {
				t.Errorf("IsValidGitHubToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_GistHistory(t *testing.T) {
	cfg := &Config{
		GistHistory: make(map[string]*GistInfo),
	}

	// Test adding gist info
	gistInfo := &GistInfo{
		ID:          "test_gist_id",
		Name:        "test_gist",
		Description: "Test gist description",
		CreatedAt:   time.Now().Format(time.RFC3339),
		UpdatedAt:   time.Now().Format(time.RFC3339),
	}

	cfg.AddGistToHistory(gistInfo)

	// Verify gist was added
	retrieved, exists := cfg.GetGistInfo("test_gist_id")
	if !exists {
		t.Error("Gist should exist in history")
	}

	if retrieved.ID != gistInfo.ID {
		t.Errorf("Gist ID = %v, want %v", retrieved.ID, gistInfo.ID)
	}

	if retrieved.Name != gistInfo.Name {
		t.Errorf("Gist Name = %v, want %v", retrieved.Name, gistInfo.Name)
	}

	// Test updating usage
	cfg.UpdateGistUsage("test_gist_id")
	updated, _ := cfg.GetGistInfo("test_gist_id")
	if updated.UsageCount != 1 {
		t.Errorf("Usage count = %v, want 1", updated.UsageCount)
	}
}

func TestConfig_Projects(t *testing.T) {
	cfg := &Config{
		Projects: make(map[string]*Project),
	}

	// Test adding project
	project := &Project{
		Name:      "test_project",
		Path:      "/path/to/test/project",
		CreatedAt: time.Now().Format(time.RFC3339),
	}

	cfg.AddProject(project)

	// Verify project was added
	retrieved, exists := cfg.GetProject("test_project")
	if !exists {
		t.Error("Project should exist")
	}

	if retrieved.Name != project.Name {
		t.Errorf("Project Name = %v, want %v", retrieved.Name, project.Name)
	}

	if retrieved.Path != project.Path {
		t.Errorf("Project Path = %v, want %v", retrieved.Path, project.Path)
	}

	// Test getting project by path
	retrieved2, exists2 := cfg.GetProjectByPath("/path/to/test/project")
	if !exists2 {
		t.Error("Project should be found by path")
	}

	if retrieved2.Name != project.Name {
		t.Errorf("Project Name = %v, want %v", retrieved2.Name, project.Name)
	}
}

func TestGenerateGistName(t *testing.T) {
	tests := []struct {
		name        string
		envFile     string
		projectName string
		environment string
		wantContain string
	}{
		{
			name:        "with project and environment",
			envFile:     ".env.production",
			projectName: "myapp",
			environment: "production",
			wantContain: "myapp",
		},
		{
			name:        "with project only",
			envFile:     ".env",
			projectName: "myapp",
			environment: "",
			wantContain: "myapp",
		},
		{
			name:        "without project",
			envFile:     ".env.staging",
			projectName: "",
			environment: "staging",
			wantContain: "staging",
		},
		{
			name:        "just env file",
			envFile:     ".env",
			projectName: "",
			environment: "",
			wantContain: "202", // Should contain current year
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateGistName(tt.envFile, tt.projectName, tt.environment)
			if !containsStr(got, tt.wantContain) {
				t.Errorf("GenerateGistName() = %v, want to contain %v", got, tt.wantContain)
			}
		})
	}
}

func TestGenerateGistDescription(t *testing.T) {
	tests := []struct {
		name        string
		envFile     string
		projectName string
		environment string
		isEncrypted bool
		wantContain []string
	}{
		{
			name:        "encrypted with all details",
			envFile:     ".env.production",
			projectName: "myapp",
			environment: "production",
			isEncrypted: true,
			wantContain: []string{"myapp", "production", "Encrypted"},
		},
		{
			name:        "unencrypted simple",
			envFile:     ".env",
			projectName: "",
			environment: "",
			isEncrypted: false,
			wantContain: []string{".env", "Tool: envi"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateGistDescription(tt.envFile, tt.projectName, tt.environment, tt.isEncrypted)
			for _, contain := range tt.wantContain {
				if !containsStr(got, contain) {
					t.Errorf("GenerateGistDescription() = %v, should contain %v", got, contain)
				}
			}
		})
	}
}

func TestGetEnvironmentFromFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{
			name:     "production env",
			filename: ".env.production",
			want:     "production",
		},
		{
			name:     "staging env",
			filename: ".env.staging",
			want:     "staging",
		},
		{
			name:     "development env",
			filename: ".env.development",
			want:     "development",
		},
		{
			name:     "base env file",
			filename: ".env",
			want:     "",
		},
		{
			name:     "env with path",
			filename: "/path/to/.env.test",
			want:     "test",
		},
		{
			name:     "complex filename",
			filename: "config/.env.production.local",
			want:     "production",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEnvironmentFromFilename(tt.filename)
			if got != tt.want {
				t.Errorf("GetEnvironmentFromFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper functions for testing
func containsPath(path, substr string) bool {
	return len(path) >= len(substr) &&
		(containsStr(path, substr) || containsStr(filepath.Base(path), substr))
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) &&
		(len(substr) == 0 ||
			(len(s) > 0 &&
				(s[:len(substr)] == substr ||
					(len(s) > len(substr) && containsStr(s[1:], substr)))))
}

// Benchmark tests
func BenchmarkGenerateGistName(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateGistName(".env.production", "MyProject", "production")
	}
}

func BenchmarkGenerateGistDescription(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateGistDescription(".env.production", "MyProject", "production", true)
	}
}

func BenchmarkGetEnvironmentFromFilename(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetEnvironmentFromFilename(".env.production")
	}
}

func BenchmarkIsValidGitHubToken(b *testing.B) {
	token := "ghp_1234567890abcdef1234567890abcdef12345678"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValidGitHubToken(token)
	}
}
