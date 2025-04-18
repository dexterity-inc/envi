package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	// Flag to specify gist ID for updates
	updateGistID string
	
	// Flag to save the last used gist ID
	saveGistID bool
)

func init() {
	pushCmd.Flags().StringVarP(&updateGistID, "id", "i", "", "GitHub Gist ID to update (if not provided, a new Gist will be created)")
	pushCmd.Flags().BoolVarP(&saveGistID, "save", "s", true, "Save the Gist ID to config for future updates (default: true)")
	rootCmd.AddCommand(pushCmd)
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push .env file to GitHub Gist",
	Long:  `Push the .env file from the current directory to a GitHub Gist. Can update an existing Gist or create a new one.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Get GitHub token
		token, err := GetGitHubToken()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Read .env file
		envContent, err := os.ReadFile(".env")
		if err != nil {
			fmt.Printf("Error reading .env file: %s\n", err)
			os.Exit(1)
		}

		// Create GitHub client
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)

		// If no gist ID provided, check config for last used ID
		if updateGistID == "" {
			config, err := LoadConfig()
			if err == nil && config.LastGistID != "" {
				// Ask user if they want to use the saved gist ID
				fmt.Printf("Found saved Gist ID: %s\n", config.LastGistID)
				fmt.Print("Do you want to update this Gist? (y/n): ")
				var response string
				fmt.Scanln(&response)
				if response == "y" || response == "Y" {
					updateGistID = config.LastGistID
				}
			}
		}

		// Create or update gist
		var gist *github.Gist
		filename := ".env"
		description := "Environment variables - Managed with envi CLI"
		
		if updateGistID != "" {
			// Try to update an existing gist
			existingGist, _, err := client.Gists.Get(ctx, updateGistID)
			if err != nil {
				fmt.Printf("Error fetching Gist with ID %s: %s\n", updateGistID, err)
				fmt.Println("Creating a new Gist instead...")
			} else {
				// Update the existing gist
				existingGist.Files = map[github.GistFilename]github.GistFile{
					github.GistFilename(filename): {
						Content: github.String(string(envContent)),
					},
				}
				
				gist, _, err = client.Gists.Edit(ctx, updateGistID, existingGist)
				if err != nil {
					fmt.Printf("Error updating Gist: %s\n", err)
					os.Exit(1)
				}
				
				fmt.Printf("Successfully updated Gist with ID: %s\n", *gist.ID)
				if gist.HTMLURL != nil {
					fmt.Printf("URL: %s\n", *gist.HTMLURL)
				}
				
				// Save gist ID to config if requested
				if saveGistID {
					saveGistIDToConfig(*gist.ID)
				}
				
				return
			}
		}
		
		// Create a new gist
		public := false
		newGist := &github.Gist{
			Description: &description,
			Public:      &public,
			Files: map[github.GistFilename]github.GistFile{
				github.GistFilename(filename): {
					Content: github.String(string(envContent)),
				},
			},
		}

		gist, _, err = client.Gists.Create(ctx, newGist)
		if err != nil {
			fmt.Printf("Error creating Gist: %s\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created new Gist\n")
		fmt.Printf("Gist ID: %s\n", *gist.ID)
		if gist.HTMLURL != nil {
			fmt.Printf("URL: %s\n", *gist.HTMLURL)
		}
		
		// Save gist ID to config if requested
		if saveGistID {
			saveGistIDToConfig(*gist.ID)
		}
	},
}

// saveGistIDToConfig saves the gist ID to the config file
func saveGistIDToConfig(gistID string) {
	config, err := LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not save Gist ID to config: %s\n", err)
		return
	}
	
	config.LastGistID = gistID
	
	if err := SaveConfig(config); err != nil {
		fmt.Printf("Warning: Could not save Gist ID to config: %s\n", err)
		return
	}
	
	fmt.Println("Gist ID saved for future updates. Use 'envi push' without '--id' flag to update this Gist.")
} 