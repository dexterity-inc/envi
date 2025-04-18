package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	// Flag to display full Gist IDs
	showFullID bool
	// Flag to display custom format
	showCustomFormat string
	// Flag to sort by different criteria
	sortBy string
)

func init() {
	listCmd.Flags().BoolVar(&showFullID, "full-id", false, "Display full Gist IDs (default: false)")
	listCmd.Flags().StringVar(&showCustomFormat, "format", "", "Custom format for displaying gists (e.g. 'id,desc,date')")
	listCmd.Flags().StringVar(&sortBy, "sort", "date", "Sort gists by: 'date', 'name', or 'id'")
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
		
		// Sort gists based on user preference
		sortGists(envGists, sortBy)
		
		// Print the list of Gists
		fmt.Printf("Found %d Gists containing .env files:\n\n", len(envGists))
		
		// Create color formatters
		boldWhite := color.New(color.FgHiWhite, color.Bold).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		
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
			
			// Extract project name and date if available in the description
			projectName := extractProjectName(description)
			date := extractDate(description)
			
			// Detect encryption status
			encStatus := ""
			if strings.Contains(description, "(encrypted)") {
				encStatus = yellow(" [encrypted]")
			} else if strings.Contains(description, "(masked)") {
				encStatus = yellow(" [masked]")
			}
			
			// Format updated time
			updated := "Unknown"
			if gist.UpdatedAt != nil {
				updated = formatTime(*gist.UpdatedAt)
			}
			
			// Format Gist ID - shortened by default
			gistID := *gist.ID
			if !showFullID && len(gistID) > 10 {
				gistID = gistID[:7] + "..." // Show first 7 chars with ellipsis
			}
			
			// Display gist info based on format
			if showCustomFormat != "" {
				// Custom format handling
				switch showCustomFormat {
				case "id,desc":
					fmt.Printf("%s %s | %s%s\n", prefix, cyan(gistID), description, encStatus)
				case "id,date":
					fmt.Printf("%s %s | Updated: %s%s\n", prefix, cyan(gistID), updated, encStatus)
				case "desc,date":
					fmt.Printf("%s %s | Updated: %s%s\n", prefix, description, updated, encStatus)
				case "id":
					fmt.Printf("%s %s%s\n", prefix, cyan(gistID), encStatus)
				case "desc":
					fmt.Printf("%s %s%s\n", prefix, description, encStatus)
				default:
					// Default custom format
					fmt.Printf("%s %s | %s | Updated: %s%s\n", prefix, cyan(gistID), description, updated, encStatus)
				}
			} else {
				// Enhanced standard format with project name and date emphasized
				if projectName != "" && date != "" {
					fmt.Printf("%s %s | Project: %s | Created: %s | ID: %s%s\n", 
						prefix, boldWhite(projectName), green(date), updated, cyan(gistID), encStatus)
				} else {
					fmt.Printf("%s %s (ID: %s, Updated: %s)%s\n", 
						prefix, description, cyan(gistID), updated, encStatus)
				}
			}
		}
		
		// Note about the highlighted Gist
		if config != nil && config.LastGistID != "" {
			fmt.Printf("\n* = %s\n", green("Last used Gist"))
		}
		
		// Help messages
		if !showFullID {
			fmt.Printf("\nTip: Use %s to see full Gist IDs\n", cyan("envi list --full-id"))
		}
		fmt.Printf("Tip: Use %s to customize the display format\n", cyan("envi list --format=\"id,desc,date\""))
		fmt.Printf("Tip: Use %s to sort by different criteria\n", cyan("envi list --sort=name"))
	},
}

// sortGists sorts the gists based on the specified criteria
func sortGists(gists []*github.Gist, by string) {
	switch by {
	case "date":
		// Sort by date (most recent first)
		for i := 0; i < len(gists)-1; i++ {
			for j := i + 1; j < len(gists); j++ {
				if gists[i].UpdatedAt != nil && gists[j].UpdatedAt != nil {
					if gists[i].UpdatedAt.Before(*gists[j].UpdatedAt) {
						gists[i], gists[j] = gists[j], gists[i]
					}
				}
			}
		}
	case "name":
		// Sort by project name from description
		for i := 0; i < len(gists)-1; i++ {
			for j := i + 1; j < len(gists); j++ {
				nameI := getDescriptionOrEmpty(gists[i])
				nameJ := getDescriptionOrEmpty(gists[j])
				if nameI > nameJ {
					gists[i], gists[j] = gists[j], gists[i]
				}
			}
		}
	case "id":
		// Sort by Gist ID
		for i := 0; i < len(gists)-1; i++ {
			for j := i + 1; j < len(gists); j++ {
				if *gists[i].ID > *gists[j].ID {
					gists[i], gists[j] = gists[j], gists[i]
				}
			}
		}
	}
}

// getDescriptionOrEmpty safely gets the description or returns empty string
func getDescriptionOrEmpty(gist *github.Gist) string {
	if gist.Description != nil {
		return *gist.Description
	}
	return ""
}

// formatTime formats a time.Time in a more human-readable format
func formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	
	if diff < time.Hour {
		return fmt.Sprintf("%d min ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	} else if diff < 48*time.Hour {
		return "yesterday"
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	}
	
	return t.Format("Jan 2, 2006")
}

// extractProjectName tries to extract project name from a description
func extractProjectName(desc string) string {
	if strings.Contains(desc, "Environment variables for ") {
		parts := strings.Split(desc, "Environment variables for ")
		if len(parts) > 1 {
			projectParts := strings.Split(parts[1], " (")
			if len(projectParts) > 0 {
				return projectParts[0]
			}
		}
	}
	return ""
}

// extractDate tries to extract date from a description
func extractDate(desc string) string {
	if strings.Contains(desc, " (") && strings.Contains(desc, ")") {
		start := strings.Index(desc, " (") + 2
		end := strings.Index(desc, ")")
		if start < end && end-start > 5 { // Ensure we have at least a year
			dateCandidate := desc[start:end]
			// Check if it looks like a date (simple check)
			if strings.Contains(dateCandidate, "-") && len(dateCandidate) >= 8 {
				return dateCandidate
			}
		}
	}
	return ""
} 