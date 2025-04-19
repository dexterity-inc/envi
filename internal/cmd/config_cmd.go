package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dexterity-inc/envi/internal/config"
)

// Configuration command variables
var (
	configToken           string
	configClearGistID     bool
	configClearToken      bool
	configForceFileStorage bool
	configEncryptByDefault bool
	configUnmaskByDefault  bool
	configDefaultKeyFile   string
	configUseKeyFileByDefault bool
	configDisableEncryption bool
)

// configCmd is the configuration command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the Envi CLI",
	Long:  `Configure Envi CLI settings including your GitHub token and default Gist ID.`,
	Run:   runConfigCommand,
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

	// Add the config command to the root command
	rootCmd.AddCommand(configCmd)
}

// runConfigCommand handles the config command execution
func runConfigCommand(cmd *cobra.Command, args []string) {
	// Load existing config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %s\n", err)
		return
	}
	
	// Handle token update
	if configToken != "" {
		// Validate token format first
		if !config.IsValidGitHubToken(configToken) {
			fmt.Println("Error: The GitHub token you provided doesn't appear to be valid.")
			fmt.Println("GitHub tokens should be at least 30 characters and follow specific formats.")
			fmt.Println("Please check your token and try again.")
			return
		}
		
		// Decide on storage method based on flags and capabilities
		if configForceFileStorage {
			cfg.GitHubToken = configToken
			cfg.TokenInKeyring = false
			fmt.Println("GitHub token stored in config file as requested.")
			fmt.Println("Warning: This is less secure than system credential storage.")
		} else {
			// Try to store in keyring first
			if err := config.SaveTokenToKeyring(configToken); err != nil {
				fmt.Printf("Error storing token in system credentials: %s\n", err)
				fmt.Println("Would you like to store the token in the config file instead? (y/N)")
				
				// Read user input
				var response string
				fmt.Scanln(&response)
				
				if response == "y" || response == "Y" {
					cfg.GitHubToken = configToken
					cfg.TokenInKeyring = false
					fmt.Println("GitHub token stored in config file.")
					fmt.Println("Warning: This is less secure than system credential storage.")
				} else {
					fmt.Println("Token not saved. You can try again or use environment variables.")
					return
				}
			} else {
				// Clear token from config file if successfully stored in keyring
				if cfg.GitHubToken != "" {
					// Securely wipe the token first
					tempConfig := *cfg
					tempConfig.GitHubToken = ""
					if err := config.SaveConfig(&tempConfig); err != nil {
						fmt.Printf("Warning: Could not securely remove old token from config: %s\n", err)
					}
				}
				
				cfg.GitHubToken = ""
				cfg.TokenInKeyring = true
				fmt.Println("GitHub token securely stored in system credential manager")
			}
		}
		
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("Error saving config: %s\n", err)
			return
		}
	}
	
	// Handle clearing token
	if configClearToken {
		var successful bool = false
		
		// First try to remove from keyring
		if cfg.TokenInKeyring {
			if err := config.DeleteTokenFromKeyring(); err != nil {
				fmt.Printf("Warning: Could not remove token from secure storage: %s\n", err)
			} else {
				fmt.Println("GitHub token removed from secure storage")
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
				fmt.Printf("Warning: Could not securely wipe token: %s\n", err)
			}
			
			cfg.GitHubToken = ""
			fmt.Println("GitHub token removed from config file")
			successful = true
		}
		
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("Error saving config: %s\n", err)
			return
		}
		
		if successful {
			fmt.Println("GitHub token successfully cleared")
		} else {
			fmt.Println("No GitHub token was found to clear")
		}
	}
	
	// Handle clearing gist ID
	if configClearGistID {
		if cfg.LastGistID == "" {
			fmt.Println("No saved Gist ID to clear")
		} else {
			oldID := cfg.LastGistID
			cfg.LastGistID = ""
			
			if err := config.SaveConfig(cfg); err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				return
			}
			
			fmt.Printf("Cleared saved Gist ID: %s\n", oldID)
		}
	}
	
	// Handle encryption settings
	if configEncryptByDefault {
		cfg.EncryptByDefault = true
		cfg.UseMaskedEncryption = false
		fmt.Println("Full encryption will be enabled by default")
		
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("Error saving config: %s\n", err)
			return
		}
	}
	
	// Make masked encryption the default behavior unless full encryption or disable encryption is explicitly requested
	if !configEncryptByDefault && !configDisableEncryption {
		// Only set if the current config doesn't already have masked encryption enabled
		if !cfg.EncryptByDefault || !cfg.UseMaskedEncryption {
			cfg.EncryptByDefault = true
			cfg.UseMaskedEncryption = true
			fmt.Println("Masked encryption enabled by default")
			
			if err := config.SaveConfig(cfg); err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				return
			}
		}
	}
	
	// Handle unmask by default setting
	if configUnmaskByDefault {
		cfg.UnmaskByDefault = true
		fmt.Println("Values will be automatically unmasked when pulling")
		
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("Error saving config: %s\n", err)
			return
		}
	}
	
	if configDisableEncryption {
		cfg.EncryptByDefault = false
		cfg.UseMaskedEncryption = false
		fmt.Println("Encryption has been disabled by default")
		
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("Error saving config: %s\n", err)
			return
		}
	}
	
	if configDefaultKeyFile != "" {
		cfg.DefaultKeyFile = configDefaultKeyFile
		cfg.UseKeyFileByDefault = true
		fmt.Printf("Default encryption key file set to: %s\n", configDefaultKeyFile)
		
		// Check if the key file exists, if not, ask to generate it
		if _, err := os.Stat(configDefaultKeyFile); os.IsNotExist(err) {
			fmt.Printf("Key file %s does not exist. Generate it? (y/n): ", configDefaultKeyFile)
			var response string
			fmt.Scanln(&response)
			
			if strings.ToLower(response) == "y" {
				// TODO: Implement key generation - for now, just output a message
				fmt.Println("Key generation functionality will be implemented in a future version")
				fmt.Println("You can manually create a key file at this location")
			}
		}
	}
	
	if configUseKeyFileByDefault && configDefaultKeyFile == "" {
		cfg.UseKeyFileByDefault = true
		fmt.Println("Key file will be used by default for encryption/decryption")
		
		if cfg.DefaultKeyFile == "" {
			// Set default path
			homeDir, err := os.UserHomeDir()
			if err != nil {
				fmt.Printf("Error getting home directory: %s\n", err)
			} else {
				cfg.DefaultKeyFile = filepath.Join(homeDir, ".envi", ".envi.key")
				fmt.Printf("Default key file set to: %s\n", cfg.DefaultKeyFile)
			}
		}
	}
	
	// If no flags provided, show current configuration
	if !cmd.Flags().Changed("token") && !configClearGistID && !configClearToken && 
	   !configEncryptByDefault && !configUnmaskByDefault && !configDisableEncryption && 
	   configDefaultKeyFile == "" && !configUseKeyFileByDefault && !configForceFileStorage {
		
		// Show current configuration
		showCurrentConfig(cfg)
		return
	}
	
	// Save configuration after all changes
	if err := config.SaveConfig(cfg); err != nil {
		fmt.Printf("Error saving config: %s\n", err)
		return
	}
	
	// Show the updated configuration
	fmt.Println("Configuration updated successfully!")
	showCurrentConfig(cfg)
}

// showCurrentConfig displays the current configuration settings
func showCurrentConfig(cfg *config.Config) {
	// Try to get token status
	var tokenStatus string
	
	// Check environment variable first
	envToken := os.Getenv("GITHUB_TOKEN")
	if envToken != "" {
		if config.IsValidGitHubToken(envToken) {
			tokenStatus = "Set via GITHUB_TOKEN environment variable (currently active)"
		} else {
			tokenStatus = "Warning: GITHUB_TOKEN environment variable contains invalid token format"
		}
	} else if cfg.TokenInKeyring {
		// Test if we can access the token (don't display it)
		_, err := config.GetTokenFromKeyring()
		if err == nil {
			tokenStatus = "Securely stored in system credential manager"
		} else {
			tokenStatus = fmt.Sprintf("Failed to access token in secure storage: %s", err)
		}
	} else if cfg.GitHubToken != "" {
		// Show only first 4 and last 4 characters for config file storage
		tokenLen := len(cfg.GitHubToken)
		if tokenLen > 8 {
			tokenStatus = fmt.Sprintf("%s...%s (stored in config file)", 
				cfg.GitHubToken[:4], 
				cfg.GitHubToken[tokenLen-4:])
		} else {
			tokenStatus = "Set in config file (too short to display safely)"
		}
		
		// Validate token format
		if !config.IsValidGitHubToken(cfg.GitHubToken) {
			tokenStatus += " - Warning: Token format appears invalid!"
		}
	} else {
		tokenStatus = "Not set"
	}
	
	fmt.Printf("GitHub Token: %s\n", tokenStatus)
	
	// Show saved gist ID
	gistStatus := "Not set"
	if cfg.LastGistID != "" {
		gistStatus = cfg.LastGistID
	}
	fmt.Printf("Default Gist ID: %s\n", gistStatus)
	
	// Show encryption settings
	fmt.Println("\nEncryption Settings:")
	if cfg.EncryptByDefault {
		if cfg.UseMaskedEncryption {
			fmt.Println("  ✓ Masked encryption enabled by default (variable names visible, values encrypted)")
			fmt.Println("    This is the default behavior")
		} else {
			fmt.Println("  ✓ Full encryption enabled by default (entire file encrypted)")
		}
	} else {
		fmt.Println("  • Encryption disabled by default")
		fmt.Println("    To enable masked encryption (recommended), run 'envi config' with no flags")
	}

	// Show unmask by default setting
	if cfg.UnmaskByDefault {
		fmt.Println("  ✓ Values will be automatically unmasked when pulling")
	} else {
		fmt.Println("  • Values will remain masked when pulling unless --unmask is specified")
	}
	
	if cfg.UseKeyFileByDefault {
		fmt.Println("  ✓ Using key file for encryption (more secure)")
		if cfg.DefaultKeyFile != "" {
			fmt.Printf("  • Default key file: %s\n", cfg.DefaultKeyFile)
			
			// Check if the key file exists
			keyFilePath := cfg.DefaultKeyFile
			if strings.HasPrefix(keyFilePath, "~/") {
				homeDir, err := os.UserHomeDir()
				if err == nil {
					keyFilePath = filepath.Join(homeDir, keyFilePath[2:])
				}
			}
			
			if _, err := os.Stat(keyFilePath); err == nil {
				fmt.Println("  ✓ Key file exists")
			} else {
				fmt.Println("  ! Key file does not exist - run 'envi config --default-key-file' again to create it")
			}
		} else {
			fmt.Println("  ! No default key file set - use 'envi config --default-key-file PATH' to set")
		}
	} else {
		fmt.Println("  • Using password-based encryption")
	}
	
	if cfg.LastGistID != "" {
		fmt.Println("\nTo use the saved Gist ID:")
		fmt.Println("  envi push              # will prompt to use the saved ID")
		fmt.Println("  envi push --id " + cfg.LastGistID + "  # will update directly")
		fmt.Println("\nTo clear the saved Gist ID:")
		fmt.Println("  envi config --clear-gist")
	}
	
	fmt.Println("\nSecurity Information:")
	if envToken != "" {
		fmt.Println("  ✓ Using environment variable for token (highest security)")
		fmt.Println("  • This takes precedence over stored tokens")
	}
	
	if cfg.TokenInKeyring {
		fmt.Println("  ✓ Persistent token is stored in your system's secure credential manager")
		fmt.Println("  ✓ To remove the token: envi config --clear-token")
	} else if cfg.GitHubToken != "" {
		fmt.Println("  ! Your token is stored in the config file")
		fmt.Println("  ! For better security, consider recreating your token and")
		fmt.Println("    running envi config --token <new-token> on a supported system")
		fmt.Println("  ✓ To remove the token: envi config --clear-token")
		
		// Check permissions
		configPath, err := config.ConfigPath()
		if err == nil {
			if info, err := os.Stat(configPath); err == nil {
				if info.Mode().Perm() != 0600 {
					fmt.Printf("  ! Warning: Config file has insecure permissions: %o\n", info.Mode().Perm())
					fmt.Println("    Run 'chmod 600 " + configPath + "' to fix")
				}
			}
		}
	}
	
	fmt.Println("\nEncryption Commands:")
	fmt.Println("  envi push                               # Push with masked encryption (default)")
	fmt.Println("  envi push --encrypt                     # Use full encryption (entire file)")
	fmt.Println("  envi push --encrypt --use-key-file      # Encrypt with key file")
	fmt.Println("  envi pull                               # Pull masked values (values remain encrypted)")
	fmt.Println("  envi pull --unmask                      # Unmask encrypted values when pulling")
	
	fmt.Println("\nTo configure encryption defaults:")
	fmt.Println("  envi config                             # Enable masked encryption (default)")
	fmt.Println("  envi config --encrypt-by-default        # Always use full encryption")
	fmt.Println("  envi config --unmask-by-default         # Always unmask values when pulling")
	fmt.Println("  envi config --disable-encryption        # Don't encrypt by default")
	fmt.Println("  envi config --default-key-file ~/.envi.key # Set default key file")
	fmt.Println("  envi config --use-key-file              # Use key file by default")
} 