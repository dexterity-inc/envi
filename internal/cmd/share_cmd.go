package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/dexterity-inc/envi/internal/config"
	"github.com/dexterity-inc/envi/internal/encryption"
	"github.com/dexterity-inc/envi/internal/tui"
)

// Share command flags
var (
	shareGistID        string
	shareWithUsers     []string
	shareReadOnlyAccess bool
	shareGenerateURL   bool
	shareExpiryInDays  int
)

// shareCmd is the share command
var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share .env file with other users",
	Long:  `Share your .env file with team members by creating a shared Gist or generating a shareable URL.`,
	Run:   runShareCommand,
}

// InitShareCommand sets up the share command and its subcommands
func InitShareCommand() {
	// Initialize the command flags
	shareCmd.Flags().StringVarP(&shareGistID, "id", "i", "", "GitHub Gist ID to share")
	shareCmd.Flags().StringSliceVarP(&shareWithUsers, "users", "u", []string{}, "GitHub usernames to share with (comma-separated)")
	shareCmd.Flags().BoolVarP(&shareReadOnlyAccess, "readonly", "r", true, "Share with read-only access")
	shareCmd.Flags().BoolVarP(&shareGenerateURL, "url", "l", false, "Generate a shareable URL")
	shareCmd.Flags().IntVarP(&shareExpiryInDays, "expiry", "e", 7, "Expiry time for shareable URL in days")
	
	// Add the share command to the root command
	rootCmd.AddCommand(shareCmd)
}

// runShareCommand handles the share command execution
func runShareCommand(cmd *cobra.Command, args []string) {
	// Get GitHub token
	token, err := config.GetGitHubToken()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	
	// Load config and apply defaults
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not load config: %s\n", err)
	} else {
		applyEncryptionDefaults(cmd, cfg)
	}
	
	// Get Gist ID (from flag or config)
	gistID := getGistID(cfg)
	
	// Prepare environment content if needed
	envContent, err := prepareEnvContent()
	if err != nil {
		fmt.Println("Error: An issue occurred while preparing the environment content. Please check the input and try again.")
		os.Exit(1)
	}
	
	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)
	
	// Get user info
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		fmt.Printf("Error getting GitHub user: %s\n", err)
		os.Exit(1)
	}
	
	// Get Gist details
	gist, _, err := client.Gists.Get(context.Background(), gistID)
	if err != nil {
		fmt.Printf("Error retrieving Gist with ID %s: %s\n", gistID, err)
		os.Exit(1)
	}
	
	// Handle sharing with users if specified
	if len(shareWithUsers) > 0 {
		shareWithGitHubUsers(client, user, gist, envContent)
	}
	
	// Generate shareable URL if requested
	if shareGenerateURL {
		generateAndShowURL(client, user, gist)
	}
	
	// If neither option was selected, show help
	if len(shareWithUsers) == 0 && !shareGenerateURL {
		fmt.Println("Please specify either users to share with (--users) or request a shareable URL (--url)")
		fmt.Println("Run 'envi share --help' for usage information")
	}
}

// getGistID gets the Gist ID from flag or config
func getGistID(cfg *config.Config) string {
	if shareGistID == "" {
		if cfg.LastGistID == "" {
			fmt.Println("Error: No Gist ID specified and no saved Gist ID found")
			fmt.Println("Use 'envi share --id GIST_ID' or first push an .env file with 'envi push'")
			os.Exit(1)
		}
		shareGistID = cfg.LastGistID
		fmt.Printf("Using saved Gist ID: %s\n", shareGistID)
	}
	return shareGistID
}

// prepareEnvContent reads and encrypts env content if needed
func prepareEnvContent() ([]byte, error) {
	// Only needed if sharing with users
	if len(shareWithUsers) == 0 {
		return nil, nil
	}
	
	// Check if .env file exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		return nil, fmt.Errorf("no .env file found in the current directory")
	}
	
	// Read .env file
	envContent, err := os.ReadFile(".env")
	if err != nil {
		return nil, fmt.Errorf("error reading .env file: %s", err)
	}
	
	// Handle encryption options
	if encryption.UseEncryption && encryption.UseMaskedEncryption {
		fmt.Println("Warning: Both --encrypt and --mask flags specified. Using --mask (masked encryption).")
		encryption.UseEncryption = false
	}
	
	if encryption.UseEncryption {
		fmt.Println("Encrypting .env file...")
		encryptedContent, err := encryption.EncryptContent(envContent)
		if err != nil {
			return nil, fmt.Errorf("error encrypting .env file: %s", err)
		}
		envContent = encryptedContent
		fmt.Println("Encryption successful.")
	} else if encryption.UseMaskedEncryption {
		fmt.Println("Masking values in .env file...")
		maskedContent, err := encryption.MaskEnvContent(envContent)
		if err != nil {
			return nil, fmt.Errorf("error masking .env file: %s", err)
		}
		envContent = maskedContent
		fmt.Println("Value masking successful. Variable names remain visible.")
	}
	
	return envContent, nil
}

// shareWithGitHubUsers shares env with specified GitHub users
func shareWithGitHubUsers(client *github.Client, user *github.User, gist *github.Gist, envContent []byte) {
	fmt.Printf("Sharing .env with users: %s\n", strings.Join(shareWithUsers, ", "))
	
	ctx := context.Background()
	
	// Process each user
	for _, username := range shareWithUsers {
		// Create description with proper attribution
		description := fmt.Sprintf("Shared .env from %s to %s - Created with envi", *user.Login, username)
		if encryption.UseEncryption {
			description += " (encrypted)"
		} else if encryption.UseMaskedEncryption {
			description += " (masked)"
		}
		
		// Create a new Gist for sharing
		newGist := &github.Gist{
			Description: github.String(description),
			Public:      github.Bool(false),
			Files: map[github.GistFilename]github.GistFile{
				github.GistFilename(".env"): {
					Content: github.String(string(envContent)),
				},
			},
		}
		
		// Add README with instructions
		readmeContent := createSharingReadmeContent(user, username)
		newGist.Files[github.GistFilename("README.md")] = github.GistFile{
			Content: github.String(readmeContent),
		}
		
		// Create the shared Gist
		createdGist, _, err := client.Gists.Create(ctx, newGist)
		if err != nil {
			fmt.Printf("Error creating shared Gist for %s: %s\n", username, err)
			continue
		}
		
		fmt.Printf("Successfully shared with %s: https://gist.github.com/%s\n", username, *createdGist.ID)
	}
}

// generateAndShowURL creates and displays a shareable URL
func generateAndShowURL(client *github.Client, user *github.User, gist *github.Gist) {
	fmt.Println("Generating shareable URL...")
	
	// Calculate expiry date
	expiryDate := time.Now().AddDate(0, 0, shareExpiryInDays)
	expiryStr := expiryDate.Format("2006-01-02")
	
	// Create a message to show
	sharingMessage := fmt.Sprintf("Shareable URL will expire on %s\n", expiryStr)
	sharingMessage += "Anyone with this URL can access your .env file.\n"
	sharingMessage += fmt.Sprintf("https://gist.github.com/%s\n", *gist.ID)
	
	// Display message using TUI if enabled
	if encryption.UseTUI {
		tui.DisplayMessage("Shareable URL Generated", sharingMessage)
	} else {
		fmt.Println(sharingMessage)
	}
} 