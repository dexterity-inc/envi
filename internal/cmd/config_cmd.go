package cmd

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dexterity-inc/envi/internal/config"
	"github.com/dexterity-inc/envi/internal/utils"
)

// Configuration command variables
var (
	configToken               string
	configClearGistID         bool
	configClearToken          bool
	configForceFileStorage    bool
	configEncryptByDefault    bool
	configUnmaskByDefault     bool
	configDefaultKeyFile      string
	configUseKeyFileByDefault bool
	configDisableEncryption   bool
	configGenerateKey         bool
	configKeyFile             string
)

// configCmd is the configuration command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the Envi CLI",
	Long: `Configure Envi CLI settings including your GitHub token and default Gist ID.

GitHub Token Permissions Required:
  • gist (Full control of private gists)
  • repo (Full control of private repositories) - for project detection

To create a GitHub Personal Access Token:
  1. Go to GitHub.com → Settings → Developer settings → Personal access tokens
  2. Click "Generate new token (classic)"
  3. Give it a descriptive name (e.g., "Envi CLI")
  4. Set expiration as needed
  5. Select the required scopes:
     - gist (Full control of private gists)
     - repo (Full control of private repositories)
  6. Click "Generate token"
  7. Copy the token and use: envi config --token YOUR_TOKEN

The token will be securely stored in your system's credential manager.`,
	Run: runConfigCommand,
}

// InitConfigCommand sets up the config command and its subcommands
func InitConfigCommand() {
	// Initialize the command flags
	configCmd.Flags().StringVarP(&configToken, "token", "t", "", "Set your GitHub personal access token")
	configCmd.Flags().BoolVarP(&configClearGistID, "clear-gist", "c", false, "Clear the saved Gist ID")
	configCmd.Flags().BoolVar(&configClearToken, "clear-token", false, "Remove the GitHub token from secure storage")
	configCmd.Flags().BoolVar(&configForceFileStorage, "force-file-storage", false, "Force token storage in file instead of system credential manager (not recommended)")
	configCmd.Flags().BoolVar(&configEncryptByDefault, "encrypt-by-default", false, "Enable full encryption by default (entire file encrypted)")
	configCmd.Flags().BoolVar(&configUnmaskByDefault, "unmask-by-default", false, "Automatically unmask/decrypt values when pulling (otherwise they remain encrypted)")
	configCmd.Flags().StringVar(&configDefaultKeyFile, "default-key-file", "", "Set the default encryption key file path")
	configCmd.Flags().BoolVar(&configUseKeyFileByDefault, "use-key-file", false, "Use key file by default instead of password for encryption")
	configCmd.Flags().BoolVar(&configDisableEncryption, "disable-encryption", false, "Disable encryption by default")
	configCmd.Flags().BoolVar(&configGenerateKey, "generate-key", false, "Generate a new encryption key file")
	configCmd.Flags().StringVar(&configKeyFile, "key-file", "", "Specify the encryption key file")

	// Add the config command to the root command
	rootCmd.AddCommand(configCmd)
}

// runConfigCommand handles the config command execution
func runConfigCommand(cmd *cobra.Command, args []string) {
	// Load existing config
	cfg, err := config.LoadConfig()
	if err != nil {
		utils.Error("Error loading config: %s", err)
		utils.Fatal("Failed to load configuration")
	}

	// Handle token update
	if configToken != "" {
		handleTokenUpdate(cfg)
	}

	// Handle clearing token
	if configClearToken {
		handleTokenClear(cfg)
	}

	// Handle clearing Gist ID
	if configClearGistID {
		handleGistIDClear(cfg)
	}

	// Handle encryption settings
	if configEncryptByDefault || configUnmaskByDefault || configDefaultKeyFile != "" || configUseKeyFileByDefault {
		handleEncryptionSettings(cfg)
	}

	// Handle key generation
	if configGenerateKey {
		handleKeyGeneration()
	}

	// Show current configuration
	showCurrentConfig(cfg)
}

// handleTokenUpdate handles updating the GitHub token
func handleTokenUpdate(cfg *config.Config) {
	// Validate token format first
	if !config.IsValidGitHubToken(configToken) {
		utils.Error("The GitHub token you provided doesn't appear to be valid.")
		utils.Info("GitHub tokens should be at least 30 characters and follow specific formats.")
		utils.Info("Please check your token and try again.")
		utils.Fatal("Invalid token format")
	}

	// Decide on storage method based on flags and capabilities
	if configForceFileStorage {
		cfg.GitHubToken = configToken
		cfg.TokenInKeyring = false
		utils.Info("GitHub token stored in config file as requested.")
		utils.Warn("This is less secure than system credential storage.")
		utils.Info("")
		utils.Info("📋 Required GitHub Token Permissions:")
		utils.Info("   • gist (Full control of private gists)")
		utils.Info("   • repo (Full control of private repositories)")
		utils.Info("")
		utils.Info("🔗 If you need to update permissions, visit:")
		utils.Info("   https://github.com/settings/tokens")
	} else {
		// Try to store in keyring first
		if err := config.SaveTokenToKeyring(configToken); err != nil {
			utils.Error("Error storing token in system credentials: %s", err)
			utils.Info("Would you like to store the token in the config file instead? (y/N)")

			// Read user input
			var response string
			_, err := fmt.Scanln(&response)
			if err != nil {
				utils.Error("Error reading user input: %s", err)
				utils.Fatal("Failed to read user input")
			}

			if response == "y" || response == "Y" {
				cfg.GitHubToken = configToken
				cfg.TokenInKeyring = false
				utils.Info("GitHub token stored in config file.")
				utils.Warn("This is less secure than system credential storage.")
				utils.Info("")
				utils.Info("📋 Required GitHub Token Permissions:")
				utils.Info("   • gist (Full control of private gists)")
				utils.Info("   • repo (Full control of private repositories)")
				utils.Info("")
				utils.Info("🔗 If you need to update permissions, visit:")
				utils.Info("   https://github.com/settings/tokens")
			} else {
				utils.Info("Token not saved. You can try again or use environment variables.")
				return
			}
		} else {
			// Clear token from config file if successfully stored in keyring
			if cfg.GitHubToken != "" {
				// Securely wipe the token first
				tempConfig := *cfg
				tempConfig.GitHubToken = ""
				if err := config.SaveConfig(&tempConfig); err != nil {
					utils.Warn("Could not securely remove old token from config: %s", err)
				}
			}

			cfg.GitHubToken = ""
			cfg.TokenInKeyring = true
			utils.Success("GitHub token securely stored in system credential manager")
			utils.Info("")
			utils.Info("📋 Required GitHub Token Permissions:")
			utils.Info("   • gist (Full control of private gists)")
			utils.Info("   • repo (Full control of private repositories)")
			utils.Info("")
			utils.Info("🔗 If you need to update permissions, visit:")
			utils.Info("   https://github.com/settings/tokens")
		}
	}

	if err := config.SaveConfig(cfg); err != nil {
		utils.Error("Error saving config: %s", err)
		utils.Fatal("Failed to save config")
	}
}

// handleTokenClear handles clearing the GitHub token
func handleTokenClear(cfg *config.Config) {
	var successful bool = false

	// First try to remove from keyring
	if cfg.TokenInKeyring {
		if err := config.DeleteTokenFromKeyring(); err != nil {
			utils.Warn("Could not remove token from secure storage: %s", err)
		} else {
			utils.Info("GitHub token removed from secure storage")
			successful = true
		}
		cfg.TokenInKeyring = false
	}

	// Also clear from config file
	if cfg.GitHubToken != "" {
		// First overwrite with zeros
		tempConfig := *cfg
		tempConfig.GitHubToken = ""
		if err := config.SaveConfig(&tempConfig); err != nil {
			utils.Warn("Could not securely wipe token: %s", err)
		}

		cfg.GitHubToken = ""
		utils.Info("GitHub token removed from config file")
		successful = true
	}

	if err := config.SaveConfig(cfg); err != nil {
		utils.Error("Error saving config: %s", err)
		utils.Fatal("Failed to save config")
	}

	if successful {
		utils.Success("GitHub token successfully cleared")
	} else {
		utils.Info("No GitHub token was found to clear")
	}
}

// handleGistIDClear handles clearing the Gist ID
func handleGistIDClear(cfg *config.Config) {
	if cfg.LastGistID != "" {
		cfg.LastGistID = ""
		if err := config.SaveConfig(cfg); err != nil {
			utils.Error("Error saving config: %s", err)
			utils.Fatal("Failed to save config")
		}
		utils.Success("Default Gist ID cleared")
	} else {
		utils.Info("No default Gist ID was set to clear")
	}
}

// handleEncryptionSettings handles encryption-related configuration
func handleEncryptionSettings(cfg *config.Config) {
	if configEncryptByDefault {
		cfg.EncryptByDefault = true
		utils.Info("Full encryption enabled by default")
	}

	if configUnmaskByDefault {
		cfg.UnmaskByDefault = true
		utils.Info("Automatic unmasking enabled by default")
	}

	if configDefaultKeyFile != "" {
		cfg.DefaultKeyFile = configDefaultKeyFile
		utils.Info("Default key file set to: %s", configDefaultKeyFile)
	}

	if configUseKeyFileByDefault {
		cfg.UseKeyFileByDefault = true
		utils.Info("Key file usage enabled by default")
	}

	if err := config.SaveConfig(cfg); err != nil {
		utils.Error("Error saving config: %s", err)
		utils.Fatal("Failed to save config")
	}
}

// handleKeyGeneration handles generating a new encryption key
func handleKeyGeneration() {
	keyFile := configKeyFile
	if keyFile == "" {
		keyFile = utils.DefaultKeyFile
	}

	// Generate the key (32 random bytes)
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		utils.Error("Error generating encryption key: %s", err)
		utils.Fatal("Failed to generate key")
	}

	// Save the key to file
	if err := os.WriteFile(keyFile, key, 0600); err != nil {
		utils.Error("Error saving key file: %s", err)
		utils.Fatal("Failed to save key file")
	}

	utils.Success("Encryption key generated and saved to: %s", keyFile)
	utils.Info("Keep this file secure and don't share it with others")
}

// showCurrentConfig displays the current configuration
func showCurrentConfig(cfg *config.Config) {
	utils.Info("Current Configuration:")
	utils.Info("=====================")

	// Token status
	if cfg.TokenInKeyring {
		utils.Info("GitHub Token: [Securely stored in system credentials]")
	} else if cfg.GitHubToken != "" {
		utils.Info("GitHub Token: [Stored in config file]")
	} else {
		utils.Info("GitHub Token: [Not configured]")
	}

	// Gist ID
	if cfg.LastGistID != "" {
		utils.Info("Default Gist ID: %s", cfg.LastGistID)
	} else {
		utils.Info("Default Gist ID: [Not set]")
	}

	// Encryption settings
	utils.Info("Encryption Settings:")
	utils.Info("  - Encrypt by default: %t", cfg.EncryptByDefault)
	utils.Info("  - Masked encryption by default: %t", cfg.UseMaskedEncryption)
	utils.Info("  - Unmask by default: %t", cfg.UnmaskByDefault)
	utils.Info("  - Use key file by default: %t", cfg.UseKeyFileByDefault)

	if cfg.DefaultKeyFile != "" {
		utils.Info("  - Default key file: %s", cfg.DefaultKeyFile)
	}

	// Project information
	if cfg.Projects != nil && len(cfg.Projects) > 0 {
		utils.Info("Projects: %d configured", len(cfg.Projects))
		for name, project := range cfg.Projects {
			utils.Info("  - %s: %s", name, project.Path)
		}
	} else {
		utils.Info("Projects: [None configured]")
	}

	// Gist history
	if cfg.GistHistory != nil && len(cfg.GistHistory) > 0 {
		utils.Info("Gist History: %d entries", len(cfg.GistHistory))
	} else {
		utils.Info("Gist History: [Empty]")
	}
}
