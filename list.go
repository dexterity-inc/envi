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
	limitFlag int
	showAllFlag bool
)

func init() {
	listCmd.Flags().IntVarP(&limitFlag, "limit", "l", 10, "Maximum number of Gists to display")
	listCmd.Flags().BoolVarP(&showAllFlag, "all", "a", false, "Show all Gists, not just ones with .env files")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List your GitHub Gists containing .env files",
	Long:  `Display a list of your GitHub Gists that contain .env files.`,
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

		// Get list of gists
		gists, _, err := client.Gists.List(ctx, "", &github.GistListOptions{})
		if err != nil {
			fmt.Printf("Error fetching Gists: %s\n", err)
			os.Exit(1)
		}

		// Filter and display gists
		count := 0
		foundEnv := false
		
		// Get the currently saved Gist ID
		config, _ := LoadConfig()
		
		fmt.Println("╭─────────────────────────────────────────────────────────────────────────────╮")
		fmt.Println("│                                   GISTS                                      │")
		fmt.Println("├─────────────┬─────────────────────────────────────────────────┬─────────────┤")
		fmt.Println("│    ID       │ Description                                      │    Date     │")
		fmt.Println("├─────────────┼─────────────────────────────────────────────────┼─────────────┤")
		
		for _, gist := range gists {
			// Skip if we've hit our limit
			if count >= limitFlag {
				break
			}
			
			// Check if this gist has a .env file
			hasEnvFile := false
			for filename := range gist.Files {
				if filename == ".env" {
					hasEnvFile = true
					foundEnv = true
					break
				}
			}
			
			// Skip if it doesn't have an .env file and we're not showing all
			if !hasEnvFile && !showAllFlag {
				continue
			}
			
			// Format the date
			createdAt := "Unknown"
			if gist.CreatedAt != nil {
				createdAt = gist.CreatedAt.Format("2006-01-02")
			}
			
			// Get description or placeholder
			description := "No description"
			if gist.Description != nil && *gist.Description != "" {
				description = *gist.Description
			}
			
			// Truncate description if too long
			if len(description) > 45 {
				description = description[:42] + "..."
			}
			
			// Mark the current Gist
			currentMark := " "
			if config.LastGistID == *gist.ID {
				currentMark = "*"
			}
			
			// Display gist info
			fmt.Printf("│ %s%8s │ %-45s │ %11s │\n", 
				currentMark, 
				(*gist.ID)[:8], 
				description, 
				createdAt)
			
			count++
		}
		
		fmt.Println("╰─────────────┴─────────────────────────────────────────────────┴─────────────╯")
		
		if count == 0 {
			if showAllFlag {
				fmt.Println("No Gists found. Create one with 'envi push'.")
			} else {
				fmt.Println("No Gists with .env files found. Create one with 'envi push'.")
				fmt.Println("Use --all to show all Gists.")
			}
		} else {
			// Notes section
			fmt.Println("\nNotes:")
			fmt.Println("* Current Gist (saved in config)")
			fmt.Println("\nTo pull a specific Gist:")
			fmt.Println("  envi pull --id <gist_id>")
			fmt.Println("\nTo push to a specific Gist:")
			fmt.Println("  envi push --id <gist_id>")
		}
		
		if !foundEnv && !showAllFlag {
			fmt.Println("\nTip: Use --all to show all Gists, even ones without .env files.")
		}
	},
} 