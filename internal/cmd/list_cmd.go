package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/dexterity-inc/envi/internal/config"
)

// List command flags
var (
	listAll       bool
	listLimit     int
	listFormat    string
	listShowURLs  bool
)

// listCmd is the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List your .env file Gists",
	Long:  `List all your GitHub Gists containing .env files.`,
	Run:   runListCommand,
}

// InitListCommand sets up the list command and its subcommands
func InitListCommand() {
	// Initialize the command flags
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "Show all Gists, not just those with .env files")
	listCmd.Flags().IntVarP(&listLimit, "limit", "l", 10, "Limit number of Gists to show")
	listCmd.Flags().StringVarP(&listFormat, "format", "f", "table", "Output format (table, json)")
	listCmd.Flags().BoolVarP(&listShowURLs, "urls", "u", false, "Show Gist URLs in output")

	// Add the list command to the root command
	rootCmd.AddCommand(listCmd)
}

// runListCommand handles the list command execution
func runListCommand(cmd *cobra.Command, args []string) {
	// Get GitHub token
	token, err := config.GetGitHubToken()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	
	// Load config to get last used Gist ID
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not load config: %s\n", err)
	}
	
	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)
	
	// Get user's Gists
	var allGists []*github.Gist
	page := 1
	perPage := 30 // GitHub's default per page
	
	for {
		opts := &github.GistListOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}
		
		gists, resp, err := client.Gists.List(cmd.Context(), "", opts)
		if err != nil {
			fmt.Printf("Error fetching Gists: %s\n", err)
			os.Exit(1)
		}
		
		allGists = append(allGists, gists...)
		
		if resp.NextPage == 0 || len(allGists) >= listLimit {
			break
		}
		
		page = resp.NextPage
	}
	
	// Filter Gists if needed
	var filteredGists []*github.Gist
	for _, gist := range allGists {
		if len(filteredGists) >= listLimit {
			break
		}
		
		// Check if this Gist has an .env file
		hasEnvFile := false
		for filename := range gist.Files {
			if string(filename) == ".env" {
				hasEnvFile = true
				break
			}
		}
		
		if listAll || hasEnvFile {
			filteredGists = append(filteredGists, gist)
		}
	}
	
	// Display Gists
	if len(filteredGists) == 0 {
		fmt.Println("No Gists found")
		if !listAll {
			fmt.Println("Try using --all to show all your Gists, not just those with .env files")
		}
		return
	}
	
	// Print output in requested format
	if listFormat == "json" {
		// Simple JSON output (in a real app, you'd use json.Marshal)
		fmt.Println("[")
		for i, gist := range filteredGists {
			fmt.Printf("  {\n    \"id\": \"%s\",\n", *gist.ID)
			
			// Description
			desc := "No description"
			if gist.Description != nil && *gist.Description != "" {
				desc = *gist.Description
			}
			fmt.Printf("    \"description\": \"%s\",\n", desc)
			
			// Created date
			fmt.Printf("    \"created_at\": \"%s\",\n", gist.CreatedAt.Format(time.RFC3339))
			
			// URL
			if listShowURLs {
				fmt.Printf("    \"url\": \"https://gist.github.com/%s\",\n", *gist.ID)
			}
			
			// Files
			fmt.Printf("    \"files\": [")
			fileCount := 0
			for filename := range gist.Files {
				if fileCount > 0 {
					fmt.Print(", ")
				}
				fmt.Printf("\"%s\"", filename)
				fileCount++
			}
			fmt.Print("]\n")
			
			// Current Gist indicator
			if cfg != nil && cfg.LastGistID == *gist.ID {
				fmt.Print("    \"current\": true\n")
			} else {
				fmt.Print("    \"current\": false\n")
			}
			
			if i < len(filteredGists)-1 {
				fmt.Println("  },")
			} else {
				fmt.Println("  }")
			}
		}
		fmt.Println("]")
	} else {
		// Table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		
		// Print header
		if listShowURLs {
			fmt.Fprintln(w, "ID\tDESCRIPTION\tFILES\tCREATED\tURL\t")
		} else {
			fmt.Fprintln(w, "ID\tDESCRIPTION\tFILES\tCREATED\t")
		}
		
		// Print each Gist
		for _, gist := range filteredGists {
			// Get description
			desc := "No description"
			if gist.Description != nil && *gist.Description != "" {
				desc = *gist.Description
				if len(desc) > 40 {
					desc = desc[:37] + "..."
				}
			}
			
			// Format created time
			createdTime := "Unknown"
			if gist.CreatedAt != nil {
				createdTime = gist.CreatedAt.Format("2006-01-02")
			}
			
			// Build file list
			var fileList []string
			for filename := range gist.Files {
				fileList = append(fileList, string(filename))
			}
			filesStr := strings.Join(fileList, ", ")
			if len(filesStr) > 30 {
				filesStr = filesStr[:27] + "..."
			}
			
			// Highlight current Gist
			idStr := *gist.ID
			if cfg != nil && cfg.LastGistID == *gist.ID {
				idStr = idStr + " *"
			}
			
			// Print row
			if listShowURLs {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\thttps://gist.github.com/%s\t\n",
					idStr, desc, filesStr, createdTime, *gist.ID)
			} else {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t\n",
					idStr, desc, filesStr, createdTime)
			}
		}
		
		w.Flush()
		fmt.Println("\n* = current Gist")
	}
} 