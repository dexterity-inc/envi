package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	token      string
	clearGistID bool
	clearToken  bool
)

func init() {
	configCmd.Flags().StringVarP(&token, "token", "t", "", "Set your GitHub personal access token")
	configCmd.Flags().BoolVarP(&clearGistID, "clear-gist", "c", false, "Clear the saved Gist ID")
	configCmd.Flags().BoolVar(&clearToken, "clear-token", false, "Clear the saved GitHub token")
	rootCmd.AddCommand(configCmd)
}

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
			// Store token in secure keyring
			if err := SaveTokenToKeyring(token); err != nil {
				fmt.Printf("Error storing token in secure storage: %s\n", err)
				fmt.Println("Falling back to config file storage (less secure)")
				
				// Fallback to config file
				config.GitHubToken = token
				config.TokenInKeyring = false
			} else {
				// Clear token from config file if successfully stored in keyring
				config.GitHubToken = ""
				config.TokenInKeyring = true
				fmt.Println("GitHub token securely stored in system credential manager")
			}
			
			if err := SaveConfig(config); err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				return
			}
		}
		
		// Handle clearing token
		if clearToken {
			// First try to remove from keyring
			if config.TokenInKeyring {
				if err := DeleteTokenFromKeyring(); err != nil {
					fmt.Printf("Warning: Could not remove token from secure storage: %s\n", err)
				} else {
					fmt.Println("GitHub token removed from secure storage")
				}
				config.TokenInKeyring = false
			}
			
			// Also clear from config file
			if config.GitHubToken != "" {
				config.GitHubToken = ""
				fmt.Println("GitHub token removed from config file")
			}
			
			if err := SaveConfig(config); err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				return
			}
			
			if !config.TokenInKeyring && config.GitHubToken == "" {
				fmt.Println("GitHub token successfully cleared")
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
		
		// If no flags provided, show the current config
		if token == "" && !clearGistID && !clearToken {
			showCurrentConfig(config)
		}
	},
}

// showCurrentConfig displays the current configuration settings
func showCurrentConfig(config *Config) {
	// Try to get token status
	var tokenStatus string
	
	if config.TokenInKeyring {
		// Test if we can access the token (don't display it)
		_, err := GetTokenFromKeyring()
		if err == nil {
			tokenStatus = "Securely stored in system credential manager"
		} else {
			tokenStatus = "Failed to access token in secure storage"
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
	
	if config.LastGistID != "" {
		fmt.Println("\nTo use the saved Gist ID:")
		fmt.Println("  envi push              # will prompt to use the saved ID")
		fmt.Println("  envi push --id " + config.LastGistID + "  # will update directly")
		fmt.Println("\nTo clear the saved Gist ID:")
		fmt.Println("  envi config --clear-gist")
	}
	
	fmt.Println("\nSecurity Information:")
	if config.TokenInKeyring {
		fmt.Println("  ✓ Your token is stored in your system's secure credential manager")
		fmt.Println("  ✓ To remove the token: envi config --clear-token")
	} else if config.GitHubToken != "" {
		fmt.Println("  ! Your token is stored in the config file")
		fmt.Println("  ! For better security, consider recreating your token and")
		fmt.Println("    running envi config --token <new-token> on a supported system")
		fmt.Println("  ✓ To remove the token: envi config --clear-token")
	}
} 