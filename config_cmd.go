package main

import (
	"fmt"
	"os"
	"strings"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	token      string
	clearGistID bool
	clearToken  bool
	forceFileStorage bool
	encryptByDefault bool
	maskByDefault bool
	defaultKeyFile string
	useKeyFileByDefault bool
	disableEncryption bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure the Envi CLI",
	Long:  `Configure Envi CLI settings including your GitHub token and default Gist ID.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load existing config
		config, err := LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %s\n", err)
			return
		}
		
		// Handle token update
		if token != "" {
			// Validate token format first
			if !isValidGitHubToken(token) {
				fmt.Println("Error: The GitHub token you provided doesn't appear to be valid.")
				fmt.Println("GitHub tokens should be at least 30 characters and follow specific formats.")
				fmt.Println("Please check your token and try again.")
				return
			}
			
			// Decide on storage method based on flags and capabilities
			if forceFileStorage {
				config.GitHubToken = token
				config.TokenInKeyring = false
				fmt.Println("GitHub token stored in config file as requested.")
				fmt.Println("Warning: This is less secure than system credential storage.")
			} else {
				// Try to store in keyring first
				if err := SaveTokenToKeyring(token); err != nil {
					fmt.Printf("Error storing token in system credentials: %s\n", err)
					fmt.Println("Would you like to store the token in the config file instead? (y/N)")
					
					// Read user input
					var response string
					fmt.Scanln(&response)
					
					if response == "y" || response == "Y" {
						config.GitHubToken = token
						config.TokenInKeyring = false
						fmt.Println("GitHub token stored in config file.")
						fmt.Println("Warning: This is less secure than system credential storage.")
					} else {
						fmt.Println("Token not saved. You can try again or use environment variables.")
						return
					}
				} else {
					// Clear token from config file if successfully stored in keyring
					if config.GitHubToken != "" {
						// Securely wipe the token first
						tempConfig := *config
						tempConfig.GitHubToken = ""
						if err := SaveConfig(&tempConfig); err != nil {
							fmt.Printf("Warning: Could not securely remove old token from config: %s\n", err)
						}
					}
					
					config.GitHubToken = ""
					config.TokenInKeyring = true
					fmt.Println("GitHub token securely stored in system credential manager")
				}
			}
			
			if err := SaveConfig(config); err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				return
			}
		}
		
		// Handle clearing token
		if clearToken {
			var successful bool = false
			
			// First try to remove from keyring
			if config.TokenInKeyring {
				if err := DeleteTokenFromKeyring(); err != nil {
					fmt.Printf("Warning: Could not remove token from secure storage: %s\n", err)
				} else {
					fmt.Println("GitHub token removed from secure storage")
					successful = true
				}
				config.TokenInKeyring = false
			}
			
			// Also clear from config file
			if config.GitHubToken != "" {
				// First overwrite with zeros
				tempConfig := *config
				tempConfig.GitHubToken = ""
				if err := SaveConfig(&tempConfig); err != nil {
					fmt.Printf("Warning: Could not securely wipe token: %s\n", err)
				}
				
				config.GitHubToken = ""
				fmt.Println("GitHub token removed from config file")
				successful = true
			}
			
			if err := SaveConfig(config); err != nil {
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
		if clearGistID {
			if config.LastGistID == "" {
				fmt.Println("No saved Gist ID to clear")
			} else {
				oldID := config.LastGistID
				config.LastGistID = ""
				
				if err := SaveConfig(config); err != nil {
					fmt.Printf("Error saving config: %s\n", err)
					return
				}
				
				fmt.Printf("Cleared saved Gist ID: %s\n", oldID)
			}
		}
		
		// Handle encryption settings
		if encryptByDefault {
			config.EncryptByDefault = true
			
			// If mask-by-default is also set, use masked encryption
			if maskByDefault {
				config.UseMaskedEncryption = true
				fmt.Println("Masked encryption will be enabled by default (variable names visible, values encrypted)")
			} else {
				config.UseMaskedEncryption = false
				fmt.Println("Full encryption will be enabled by default")
			}
			
			if err := SaveConfig(config); err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				return
			}
		} else if maskByDefault && !encryptByDefault {
			config.EncryptByDefault = true
			config.UseMaskedEncryption = true
			fmt.Println("Masked encryption will be enabled by default (variable names visible, values encrypted)")
			
			if err := SaveConfig(config); err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				return
			}
		}
		
		if disableEncryption {
			config.EncryptByDefault = false
			config.UseMaskedEncryption = false
			fmt.Println("Encryption has been disabled by default")
			
			if err := SaveConfig(config); err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				return
			}
		}
		
		if defaultKeyFile != "" {
			config.DefaultKeyFile = defaultKeyFile
			config.UseKeyFileByDefault = true
			fmt.Printf("Default encryption key file set to: %s\n", defaultKeyFile)
			
			// Check if the key file exists, if not, ask to generate it
			if _, err := os.Stat(defaultKeyFile); os.IsNotExist(err) {
				fmt.Printf("Key file %s does not exist. Generate it? (y/n): ", defaultKeyFile)
				var response string
				fmt.Scanln(&response)
				if response == "y" || response == "Y" {
					// Save current key file setting
					originalKeyFile := encryptionKeyFile
					encryptionKeyFile = defaultKeyFile
					
					// Generate the key
					_, err := generateAndSaveKey()
					if err != nil {
						fmt.Printf("Error generating key file: %s\n", err)
					}
					
					// Restore original setting
					encryptionKeyFile = originalKeyFile
				} else {
					fmt.Println("Key file not generated. You'll need to create it before using encryption.")
				}
			}
			
			if err := SaveConfig(config); err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				return
			}
		}
		
		if useKeyFileByDefault {
			config.UseKeyFileByDefault = true
			fmt.Println("Encryption will use key file by default instead of password")
			
			if err := SaveConfig(config); err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				return
			}
		}
		
		// If no flags provided, show the current config
		if token == "" && !clearGistID && !clearToken && !encryptByDefault && 
		   !disableEncryption && defaultKeyFile == "" && !useKeyFileByDefault &&
		   !maskByDefault {
			showCurrentConfig(config)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}

// Initialize config command flags
func initConfigFlags() {
	configCmd.Flags().StringVarP(&token, "token", "t", "", "Set your GitHub personal access token")
	configCmd.Flags().BoolVarP(&clearGistID, "clear-gist", "c", false, "Clear the saved Gist ID")
	configCmd.Flags().BoolVar(&clearToken, "clear-token", false, "Clear the saved GitHub token")
	configCmd.Flags().BoolVar(&forceFileStorage, "force-file-storage", false, "Force token storage in config file (less secure)")
	
	// Encryption configuration flags
	configCmd.Flags().BoolVar(&encryptByDefault, "encrypt-by-default", false, "Enable encryption by default for all push operations")
	configCmd.Flags().BoolVar(&maskByDefault, "mask-by-default", false, "Use masked encryption by default (encrypt values only, keeping variable names visible)")
	configCmd.Flags().BoolVar(&disableEncryption, "disable-encryption", false, "Disable encryption by default")
	configCmd.Flags().StringVar(&defaultKeyFile, "default-key-file", "", "Set the default encryption key file path")
	configCmd.Flags().BoolVar(&useKeyFileByDefault, "use-key-file", false, "Use key file by default instead of password for encryption")
}

// showCurrentConfig displays the current configuration settings
func showCurrentConfig(config *Config) {
	// Try to get token status
	var tokenStatus string
	
	// Check environment variable first
	envToken := os.Getenv("GITHUB_TOKEN")
	if envToken != "" {
		if isValidGitHubToken(envToken) {
			tokenStatus = "Set via GITHUB_TOKEN environment variable (currently active)"
		} else {
			tokenStatus = "Warning: GITHUB_TOKEN environment variable contains invalid token format"
		}
	} else if config.TokenInKeyring {
		// Test if we can access the token (don't display it)
		_, err := GetTokenFromKeyring()
		if err == nil {
			tokenStatus = "Securely stored in system credential manager"
		} else {
			tokenStatus = fmt.Sprintf("Failed to access token in secure storage: %s", err)
		}
	} else if config.GitHubToken != "" {
		// Show only first 4 and last 4 characters for config file storage
		tokenLen := len(config.GitHubToken)
		if tokenLen > 8 {
			tokenStatus = fmt.Sprintf("%s...%s (stored in config file)", 
				config.GitHubToken[:4], 
				config.GitHubToken[tokenLen-4:])
		} else {
			tokenStatus = "Set in config file (too short to display safely)"
		}
		
		// Validate token format
		if !isValidGitHubToken(config.GitHubToken) {
			tokenStatus += " - Warning: Token format appears invalid!"
		}
	} else {
		tokenStatus = "Not set"
	}
	
	fmt.Printf("GitHub Token: %s\n", tokenStatus)
	
	// Show saved gist ID
	gistStatus := "Not set"
	if config.LastGistID != "" {
		gistStatus = config.LastGistID
	}
	fmt.Printf("Default Gist ID: %s\n", gistStatus)
	
	// Show encryption settings
	fmt.Println("\nEncryption Settings:")
	if config.EncryptByDefault {
		if config.UseMaskedEncryption {
			fmt.Println("  ✓ Masked encryption enabled by default (variable names visible, values encrypted)")
		} else {
			fmt.Println("  ✓ Full encryption enabled by default (entire file encrypted)")
		}
	} else {
		fmt.Println("  • Encryption disabled by default")
	}
	
	if config.UseKeyFileByDefault {
		fmt.Println("  ✓ Using key file for encryption (more secure)")
		if config.DefaultKeyFile != "" {
			fmt.Printf("  • Default key file: %s\n", config.DefaultKeyFile)
			
			// Check if the key file exists
			keyFilePath := config.DefaultKeyFile
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
	
	if config.LastGistID != "" {
		fmt.Println("\nTo use the saved Gist ID:")
		fmt.Println("  envi push              # will prompt to use the saved ID")
		fmt.Println("  envi push --id " + config.LastGistID + "  # will update directly")
		fmt.Println("\nTo clear the saved Gist ID:")
		fmt.Println("  envi config --clear-gist")
	}
	
	fmt.Println("\nSecurity Information:")
	if envToken != "" {
		fmt.Println("  ✓ Using environment variable for token (highest security)")
		fmt.Println("  • This takes precedence over stored tokens")
	}
	
	if config.TokenInKeyring {
		fmt.Println("  ✓ Persistent token is stored in your system's secure credential manager")
		fmt.Println("  ✓ To remove the token: envi config --clear-token")
	} else if config.GitHubToken != "" {
		fmt.Println("  ! Your token is stored in the config file")
		fmt.Println("  ! For better security, consider recreating your token and")
		fmt.Println("    running envi config --token <new-token> on a supported system")
		fmt.Println("  ✓ To remove the token: envi config --clear-token")
		
		// Check permissions
		configPath, err := ConfigPath()
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
	fmt.Println("  envi push --encrypt                        # Fully encrypt when pushing")
	fmt.Println("  envi push --mask                           # Mask values only (variable names visible)")
	fmt.Println("  envi push --encrypt --use-key-file         # Encrypt with key file")
	fmt.Println("  envi push --encrypt --generate-key         # Generate and use key file")
	fmt.Println("  envi pull --decrypt                        # Auto-decrypt when pulling")
	fmt.Println("  envi pull --unmask                         # Unmask values when pulling")
	
	fmt.Println("\nTo configure encryption defaults:")
	fmt.Println("  envi config --encrypt-by-default           # Always use full encryption")
	fmt.Println("  envi config --mask-by-default              # Always use masked encryption (values only)")
	fmt.Println("  envi config --disable-encryption           # Don't encrypt by default")
	fmt.Println("  envi config --default-key-file ~/.envi.key # Set default key file")
	fmt.Println("  envi config --use-key-file                 # Use key file by default")
} 