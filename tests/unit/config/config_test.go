package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dexterity-inc/envi/internal/config"
	"github.com/dexterity-inc/envi/tests/helpers"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name           string
		setupFunc      func(testConfig *helpers.TestConfig) error
		expectedConfig *config.Config
		expectError    bool
	}{
		{
			name: "load default config when file doesn't exist",
			setupFunc: func(testConfig *helpers.TestConfig) error {
				return nil
			},
			expectedConfig: &config.Config{
				EncryptByDefault:    true,
				UseMaskedEncryption: true,
			},
			expectError: false,
		},
		{
			name: "load existing config file",
			setupFunc: func(testConfig *helpers.TestConfig) error {
				configContent := `github_token: "test_token"
last_gist_id: "test_gist"
token_in_keyring: true`
				helpers.CreateTestConfigFile(t, testConfig.ConfigDir, []byte(configContent))
				return nil
			},
			expectedConfig: &config.Config{
				GitHubToken:    "test_token",
				LastGistID:     "test_gist",
				TokenInKeyring: true,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testConfig := helpers.SetupTestEnvironment(t)
			defer testConfig.Cleanup()

			originalHome := os.Getenv("HOME")
			os.Setenv("HOME", testConfig.TempDir)
			defer os.Setenv("HOME", originalHome)

			if err := tt.setupFunc(testConfig); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			loadedConfig, err := config.LoadConfig()

			if tt.expectError {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if loadedConfig.GitHubToken != tt.expectedConfig.GitHubToken {
				t.Errorf("GitHubToken mismatch: expected %s, got %s", 
					tt.expectedConfig.GitHubToken, loadedConfig.GitHubToken)
			}
		})
	}
}

func TestConfigPath(t *testing.T) {
	testConfig := helpers.SetupTestEnvironment(t)
	defer testConfig.Cleanup()

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", testConfig.TempDir)
	defer os.Setenv("HOME", originalHome)

	path, err := config.ConfigPath()
	if err != nil {
		t.Fatalf("Failed to get config path: %v", err)
	}

	expectedPath := filepath.Join(testConfig.TempDir, ".envi", "config.yaml")
	if path != expectedPath {
		t.Errorf("Config path mismatch: expected %s, got %s", expectedPath, path)
	}
}