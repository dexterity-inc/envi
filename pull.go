package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	pullGistID    string
	backupFlag    bool
	forceOverwrite bool
)

func init() {
	pullCmd.Flags().StringVarP(&pullGistID, "id", "i", "", "GitHub Gist ID to pull from (if not specified, uses the saved ID)")
	pullCmd.Flags().BoolVarP(&backupFlag, "backup", "b", false, "Create a backup of the existing .env file before overwriting")
	pullCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "Force overwrite without confirmation if .env file exists")
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull .env file from GitHub Gist",
	Long:  `Pull .env file from a GitHub Gist and save it to the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if .env exists and handle backups/confirmations
		if _, err := os.Stat(".env"); err == nil {
			if backupFlag {
				// Create backup with timestamp
				timestamp := time.Now().Format("20060102-150405")
				backupPath := fmt.Sprintf(".env.backup.%s", timestamp)
				
				envContent, err := os.ReadFile(".env")
				if err != nil {
					fmt.Printf("Error reading .env for backup: %s\n", err)
					os.Exit(1)
				}
				
				err = os.WriteFile(backupPath, envContent, 0600)
				if err != nil {
					fmt.Printf("Error creating backup file: %s\n", err)
					os.Exit(1)
				}
				
				fmt.Printf("Created backup at: %s\n", backupPath)
			} else if !forceOverwrite {
				// Ask for confirmation before overwriting
				fmt.Print("An .env file already exists. Overwrite? (y/n): ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Pull canceled")
					return
				}
			}
		}
		
		// If no Gist ID provided, check config for last used ID
		if pullGistID == "" {
			config, err := LoadConfig()
			if err != nil {
				fmt.Printf("Error loading config: %s\n", err)
				os.Exit(1)
			}
			
			if config.LastGistID == "" {
				fmt.Println("Error: No Gist ID specified and no saved Gist ID found")
				fmt.Println("Use 'envi pull --id GIST_ID' or first push an .env file with 'envi push'")
				os.Exit(1)
			}
			
			// Ask user if they want to use the saved gist ID
			fmt.Printf("Found saved Gist ID: %s\n", config.LastGistID)
			fmt.Print("Do you want to pull from this Gist? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Pull canceled")
				return
			}
			
			pullGistID = config.LastGistID
		}

		// Get GitHub token
		token, err := GetGitHubToken()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Create GitHub client
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)

		// Get Gist from GitHub
		gist, _, err := client.Gists.Get(ctx, pullGistID)
		if err != nil {
			fmt.Printf("Error fetching Gist: %s\n", err)
			os.Exit(1)
		}

		// Extract .env file content
		var envContent string
		found := false
		for filename, file := range gist.Files {
			if filename == ".env" {
				envContent = *file.Content
				found = true
				break
			}
		}

		if !found {
			fmt.Println("Error: No .env file found in the specified Gist")
			os.Exit(1)
		}

		// Write content to .env file
		err = os.WriteFile(".env", []byte(envContent), 0600)
		if err != nil {
			fmt.Printf("Error writing .env file: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully pulled .env file from Gist")
		
		// Save Gist ID for future use if it's not already saved
		config, err := LoadConfig()
		if err == nil && config.LastGistID != pullGistID {
			config.LastGistID = pullGistID
			if err := SaveConfig(config); err == nil {
				fmt.Println("Saved Gist ID for future use")
			}
		}
	},
} 