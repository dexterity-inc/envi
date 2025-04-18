package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available Gists",
	Long:  `List available Gists containing .env files from your GitHub account.`,
	Run: func(cmd *cobra.Command, args []string) {
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

		// Get user's Gists
		gists, _, err := client.Gists.List(ctx, "", nil)
		if err != nil {
			fmt.Printf("Error fetching Gists: %s\n", err)
			os.Exit(1)
		}

		// Filter Gists containing .env files
		var envGists []*github.Gist
		for _, gist := range gists {
			for filename := range gist.Files {
				if filename == ".env" {
					envGists = append(envGists, gist)
					break
				}
			}
		}

		if len(envGists) == 0 {
			fmt.Println("No Gists containing .env files found.")
			return
		}

		// Load config to see if there's a last used Gist
		config, _ := LoadConfig()
		
		// Print the list of Gists
		fmt.Printf("Found %d Gists containing .env files:\n\n", len(envGists))
		
		for i, gist := range envGists {
			// Highlight if this is the last used Gist
			prefix := fmt.Sprintf("%d)", i+1)
			if config != nil && config.LastGistID == *gist.ID {
				prefix = color.GreenString("* " + prefix)
			} else {
				prefix = "  " + prefix
			}
			
			// Format description
			description := "No description"
			if gist.Description != nil && *gist.Description != "" {
				description = *gist.Description
			}
			
			// Format updated time
			updated := "Unknown"
			if gist.UpdatedAt != nil {
				updated = gist.UpdatedAt.Format(time.RFC3339)
			}
			
			fmt.Printf("%s %s (ID: %s, Updated: %s)\n", 
				prefix, description, *gist.ID, updated)
		}
		
		// Note about the highlighted Gist
		if config != nil && config.LastGistID != "" {
			fmt.Printf("\n* = Last used Gist\n")
		}
	},
} 