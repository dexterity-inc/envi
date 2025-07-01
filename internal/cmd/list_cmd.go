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
	"github.com/dexterity-inc/envi/internal/tui"
	"github.com/dexterity-inc/envi/internal/utils"
)

// List command flags
var (
	listAll      bool
	listLimit    int
	listFormat   string
	listShowURLs bool
	listTUI      bool
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
	listCmd.Flags().BoolVarP(&listTUI, "tui", "t", false, "Enable the enhanced TUI gist manager")

	// Add the list command to the root command
	rootCmd.AddCommand(listCmd)
}

// runListCommand handles the list command execution
func runListCommand(cmd *cobra.Command, args []string) {
	logger := utils.GetLogger()

	// Get GitHub token
	token, err := config.GetGitHubToken()
	if err != nil {
		utils.FatalError(err, "getting GitHub token")
	}

	// Load config to get last used Gist ID and gist history
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Warn("Could not load config: %s", err)
	}

	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)

	// Get user's Gists
	var allGists []*github.Gist
	page := 1
	perPage := 30 // GitHub's default per page

	// Pre-allocate slice with estimated capacity to reduce memory allocations
	// Estimate based on limit or reasonable default
	estimatedCapacity := listLimit
	if estimatedCapacity > 300 { // Don't over-allocate for very large limits
		estimatedCapacity = 300
	}
	allGists = make([]*github.Gist, 0, estimatedCapacity)

	for {
		opts := &github.GistListOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}

		gists, resp, err := client.Gists.List(cmd.Context(), "", opts)
		if err != nil {
			utils.FatalError(err, "fetching Gists")
		}

		allGists = append(allGists, gists...)

		if resp.NextPage == 0 || len(allGists) >= listLimit {
			break
		}

		page = resp.NextPage
	}

	// Filter Gists with pre-allocated capacity
	filteredGists := make([]*github.Gist, 0, len(allGists))
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

	// If TUI is requested, use the enhanced gist manager
	if listTUI {
		// Convert GitHub gists to config.GistInfo format with pre-allocated capacity
		gistInfos := make([]*config.GistInfo, 0, len(filteredGists))

		for _, gist := range filteredGists {
			// Try to get existing metadata from config
			var gistInfo *config.GistInfo
			if cfg != nil {
				if existing, exists := cfg.GetGistInfo(*gist.ID); exists {
					gistInfo = existing
				}
			}

			// If no existing metadata, create basic info
			if gistInfo == nil {
				desc := "No description"
				if gist.Description != nil && *gist.Description != "" {
					desc = *gist.Description
				}

				gistInfo = &config.GistInfo{
					ID:          *gist.ID,
					Name:        desc,
					Description: desc,
					CreatedAt:   gist.CreatedAt.Format("2006-01-02 15:04:05"),
					UpdatedAt:   gist.UpdatedAt.Format("2006-01-02 15:04:05"),
					IsPublic:    gist.Public != nil && *gist.Public,
					FileCount:   len(gist.Files),
					URL:         fmt.Sprintf("https://gist.github.com/%s", *gist.ID),
				}
			}

			gistInfos = append(gistInfos, gistInfo)
		}

		// Show the enhanced TUI gist manager
		if err := tui.ShowGistManager(gistInfos); err != nil {
			utils.FatalError(err, "showing TUI")
		}
		return
	}

	// Display Gists in traditional format
	if len(filteredGists) == 0 {
		logger.Info("No Gists found")
		if !listAll {
			logger.Info("Try using --all to show all your Gists, not just those with .env files")
		}
		return
	}

	// Print output in requested format
	if listFormat == "json" {
		// Enhanced JSON output with metadata
		fmt.Println("[")
		for i, gist := range filteredGists {
			fmt.Printf("  {\n    \"id\": \"%s\",\n", *gist.ID)

			// Description
			desc := "No description"
			if gist.Description != nil && *gist.Description != "" {
				desc = *gist.Description
			}
			fmt.Printf("    \"description\": \"%s\",\n", desc)

			// Enhanced metadata from config
			if cfg != nil {
				if gistInfo, exists := cfg.GetGistInfo(*gist.ID); exists {
					fmt.Printf("    \"name\": \"%s\",\n", gistInfo.Name)
					fmt.Printf("    \"project_name\": \"%s\",\n", gistInfo.ProjectName)
					fmt.Printf("    \"environment\": \"%s\",\n", gistInfo.Environment)
					fmt.Printf("    \"usage_count\": %d,\n", gistInfo.UsageCount)
					fmt.Printf("    \"is_encrypted\": %t,\n", gistInfo.IsEncrypted)
					fmt.Printf("    \"is_public\": %t,\n", gistInfo.IsPublic)
					if gistInfo.LastUsed != "" {
						fmt.Printf("    \"last_used\": \"%s\",\n", gistInfo.LastUsed)
					}
				}
			}

			// Created date
			fmt.Printf("    \"created_at\": \"%s\",\n", gist.CreatedAt.Format(time.RFC3339))

			// URL
			fmt.Printf("    \"url\": \"https://gist.github.com/%s\",\n", *gist.ID)

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
		// Enhanced table format
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

		// Print header with enhanced columns
		fmt.Fprintln(w, "ID\tNAME\tPROJECT\tENV\tUSAGE\tFILES\tCREATED\tURL\t")

		// Print each Gist
		for _, gist := range filteredGists {
			// Get enhanced metadata
			var name, project, env, usage string
			if cfg != nil {
				if gistInfo, exists := cfg.GetGistInfo(*gist.ID); exists {
					name = gistInfo.Name
					if len(name) > 30 {
						name = name[:27] + "..."
					}
					project = gistInfo.ProjectName
					if len(project) > 15 {
						project = project[:12] + "..."
					}
					env = gistInfo.Environment
					if len(env) > 10 {
						env = env[:7] + "..."
					}
					usage = fmt.Sprintf("%d", gistInfo.UsageCount)
				}
			}

			// Fallback values
			if name == "" {
				desc := "No description"
				if gist.Description != nil && *gist.Description != "" {
					desc = *gist.Description
				}
				if len(desc) > 30 {
					desc = desc[:27] + "..."
				}
				name = desc
			}
			if project == "" {
				project = "-"
			}
			if env == "" {
				env = "-"
			}
			if usage == "" {
				usage = "0"
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
			if len(filesStr) > 20 {
				filesStr = filesStr[:17] + "..."
			}

			// Highlight current Gist
			idStr := *gist.ID
			if cfg != nil && cfg.LastGistID == *gist.ID {
				idStr = idStr + " *"
			}

			// Always print the URL
			urlStr := fmt.Sprintf("https://gist.github.com/%s", *gist.ID)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t\n",
				idStr, name, project, env, usage, filesStr, createdTime, urlStr)
		}

		w.Flush()
		fmt.Println("\n* = current Gist")
	}
}
