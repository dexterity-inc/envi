package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/dexterity-inc/envi/internal/config"
	"github.com/dexterity-inc/envi/internal/tui"
	"github.com/dexterity-inc/envi/internal/utils"
)

// Gist command flags
var (
	gistShowHistory  bool
	gistShowProjects bool
	gistOpen         string
	gistStats        bool
)

// gistCmd is the gist management command
var gistCmd = &cobra.Command{
	Use:   "gist",
	Short: "Enhanced gist management",
	Long:  `Manage your gists with enhanced features including interactive TUI, project tracking, and metadata.`,
	Run:   runGistCommand,
}

// InitGistCommand sets up the gist command and its subcommands
func InitGistCommand() {
	// Initialize the command flags
	gistCmd.Flags().BoolVarP(&gistShowHistory, "history", "i", false, "Show gist history with enhanced metadata")
	gistCmd.Flags().BoolVarP(&gistShowProjects, "projects", "p", false, "Show project information")
	gistCmd.Flags().StringVarP(&gistOpen, "open", "o", "", "Open a specific gist by ID")
	gistCmd.Flags().BoolVarP(&gistStats, "stats", "s", false, "Show gist usage statistics")

	// Add the gist command to the root command
	rootCmd.AddCommand(gistCmd)
}

// runGistCommand handles the gist command execution
func runGistCommand(cmd *cobra.Command, args []string) {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		utils.Error("Error loading config: %s", err)
		utils.Fatal("Failed to load configuration")
	}

	// Handle different subcommands
	if gistOpen != "" {
		openSpecificGist(cmd, gistOpen)
		return
	}

	if gistShowProjects {
		showProjects(cfg)
		return
	}

	if gistStats {
		showGistStats(cfg)
		return
	}

	// Default: show interactive gist manager
	showInteractiveGistManager(cmd, cfg)
}

// openSpecificGist opens a specific gist by ID
func openSpecificGist(cmd *cobra.Command, gistID string) {
	// Get GitHub token
	token, err := config.GetGitHubToken()
	if err != nil {
		utils.Error("Error: %s", err)
		utils.Fatal("Failed to get GitHub token")
	}

	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)

	// Get the gist
	gist, _, err := client.Gists.Get(cmd.Context(), gistID)
	if err != nil {
		utils.Error("Error retrieving Gist with ID %s: %s", gistID, err)
		utils.Fatal("Failed to retrieve Gist")
	}

	// Display gist information
	utils.Info("Gist ID: %s", *gist.ID)
	if gist.Description != nil {
		utils.Info("Description: %s", *gist.Description)
	}
	utils.Info("URL: https://gist.github.com/%s", *gist.ID)
	utils.Info("Created: %s", gist.CreatedAt.Format("2006-01-02 15:04:05"))
	utils.Info("Updated: %s", gist.UpdatedAt.Format("2006-01-02 15:04:05"))
	utils.Info("Public: %t", gist.Public != nil && *gist.Public)
	utils.Info("Files: %d", len(gist.Files))

	// List files
	utils.Info("")
	utils.Info("Files:")
	for filename := range gist.Files {
		utils.Info("  - %s", filename)
	}

	// Open in browser
	utils.Info("")
	utils.Info("Opening gist in browser...")
	openURL(fmt.Sprintf("https://gist.github.com/%s", gistID))
}

// showProjects displays project information
func showProjects(cfg *config.Config) {
	if cfg.Projects == nil || len(cfg.Projects) == 0 {
		utils.Info("No projects found in configuration.")
		return
	}

	utils.Info("📁 Projects:")
	utils.Info("============")

	for name, project := range cfg.Projects {
		utils.Info("")
		utils.Info("🏗️  %s", name)
		utils.Info("   Path: %s", project.Path)
		utils.Info("   Created: %s", project.CreatedAt)
		utils.Info("   Last Used: %s", project.LastUsed)

		if len(project.Environments) > 0 {
			utils.Info("   Environments: %s", strings.Join(project.Environments, ", "))
		}

		if len(project.GistIDs) > 0 {
			utils.Info("   Gists: %d", len(project.GistIDs))
			for _, gistID := range project.GistIDs {
				utils.Info("     - %s (https://gist.github.com/%s)", gistID, gistID)
			}
		}
	}
}

// showGistStats displays gist usage statistics
func showGistStats(cfg *config.Config) {
	if cfg.GistHistory == nil || len(cfg.GistHistory) == 0 {
		utils.Info("No gist history found.")
		return
	}

	utils.Info("📊 Gist Statistics:")
	utils.Info("===================")

	var totalUsage int
	var encryptedCount int
	var publicCount int
	var recentCount int

	// Calculate statistics
	for _, gist := range cfg.GistHistory {
		totalUsage += gist.UsageCount
		if gist.IsEncrypted {
			encryptedCount++
		}
		if gist.IsPublic {
			publicCount++
		}

		// Check if used in last 7 days
		if gist.LastUsed != "" {
			if lastUsed, err := time.Parse("2006-01-02 15:04:05", gist.LastUsed); err == nil {
				if lastUsed.After(time.Now().AddDate(0, 0, -7)) {
					recentCount++
				}
			}
		}
	}

	utils.Info("Total Gists: %d", len(cfg.GistHistory))
	utils.Info("Total Usage: %d", totalUsage)
	utils.Info("Encrypted: %d", encryptedCount)
	utils.Info("Public: %d", publicCount)
	utils.Info("Recently Used (7 days): %d", recentCount)

	// Show most used gists
	utils.Info("")
	utils.Info("🔥 Most Used Gists:")
	var sortedGists []*config.GistInfo
	for _, gist := range cfg.GistHistory {
		sortedGists = append(sortedGists, gist)
	}

	// Sort by usage count (simple bubble sort)
	for i := 0; i < len(sortedGists)-1; i++ {
		for j := i + 1; j < len(sortedGists); j++ {
			if sortedGists[i].UsageCount < sortedGists[j].UsageCount {
				sortedGists[i], sortedGists[j] = sortedGists[j], sortedGists[i]
			}
		}
	}

	// Show top 5 most used gists
	topCount := 5
	if len(sortedGists) < topCount {
		topCount = len(sortedGists)
	}

	for i := 0; i < topCount; i++ {
		gist := sortedGists[i]
		utils.Info("  %d. %s - %d uses", i+1, gist.ID, gist.UsageCount)
		if gist.Description != "" {
			utils.Info("     Description: %s", gist.Description)
		}
	}
}

// showInteractiveGistManager shows the interactive TUI gist manager
func showInteractiveGistManager(cmd *cobra.Command, cfg *config.Config) {
	// Get GitHub token
	token, err := config.GetGitHubToken()
	if err != nil {
		utils.Error("Error: %s", err)
		utils.Fatal("Failed to get GitHub token")
	}

	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)

	// Get user's Gists
	var allGists []*github.Gist
	page := 1
	perPage := 100 // Get more gists for better management

	for {
		opts := &github.GistListOptions{
			ListOptions: github.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}

		gists, resp, err := client.Gists.List(cmd.Context(), "", opts)
		if err != nil {
			utils.Error("Error fetching Gists: %s", err)
			utils.Fatal("Failed to fetch Gists")
		}

		allGists = append(allGists, gists...)

		if resp.NextPage == 0 {
			break
		}

		page = resp.NextPage
	}

	// Convert to enhanced format
	var gistInfos []*config.GistInfo

	for _, gist := range allGists {
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

	// Show interactive TUI (no extra info message)
	if err := tui.ShowGistManager(gistInfos); err != nil {
		utils.Error("Error in interactive gist manager: %s", err)
		utils.Fatal("Failed to start interactive gist manager")
	}
}

// openURL opens a URL in the default browser
func openURL(url string) {
	// This is a simplified implementation
	// In a real implementation, you'd use a cross-platform library
	// like github.com/pkg/browser or similar
	utils.Info("Would open URL: %s", url)
	utils.Info("Please open the URL manually in your browser")
}
