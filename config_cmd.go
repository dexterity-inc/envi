package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	token      string
	clearGistID bool
)

func init() {
	configCmd.Flags().StringVarP(&token, "token", "t", "", "Set your GitHub personal access token")
	configCmd.Flags().BoolVarP(&clearGistID, "clear-gist", "c", false, "Clear the saved Gist ID")
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
			// Update token
			config.GitHubToken = token
			
			if err := SaveConfig(config); err != nil {
				fmt.Printf("Error saving config: %s\n", err)
				return
			}
			
			fmt.Println("GitHub token saved successfully")
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
		if token == "" && !clearGistID {
			// Don't show the full token for security
			tokenStatus := "Not set"
			if config.GitHubToken != "" {
				// Show only first 4 and last 4 characters
				tokenLen := len(config.GitHubToken)
				if tokenLen > 8 {
					tokenStatus = fmt.Sprintf("%s...%s", 
						config.GitHubToken[:4], 
						config.GitHubToken[tokenLen-4:])
				} else {
					tokenStatus = "Set (too short to display safely)"
				}
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
		}
	},
} 