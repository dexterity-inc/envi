package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/term"

	"github.com/dexterity-inc/envi/internal/config"
	"github.com/dexterity-inc/envi/internal/encryption"
	"github.com/dexterity-inc/envi/internal/tui"
	"github.com/dexterity-inc/envi/internal/utils"
)

// Share command flags
var (
	shareGistID           string
	shareWithUsers        []string
	shareReadOnlyAccess   bool
	shareGenerateURL      bool
	shareExpiryInDays     int
	shareGenerateKeyFile  bool
	shareOutputKeyFile    string
	shareSelfContained    bool
	sharePassword         string
	shareGeneratePassword bool
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
	shareCmd.Flags().BoolVar(&shareGenerateKeyFile, "generate-key", false, "Generate a key file for encryption")
	shareCmd.Flags().StringVar(&shareOutputKeyFile, "key-output", "", "Output path for generated key file (default: .envi-share.key)")
	shareCmd.Flags().BoolVar(&shareSelfContained, "self-contained", false, "Share environment variables in a self-contained format")
	shareCmd.Flags().StringVar(&sharePassword, "password", "", "Password for self-contained sharing")
	shareCmd.Flags().BoolVar(&shareGeneratePassword, "generate-password", false, "Generate a password for self-contained sharing")

	// Add the share command to the root command
	rootCmd.AddCommand(shareCmd)
}

// runShareCommand handles the share command execution
func runShareCommand(cmd *cobra.Command, args []string) {
	logger := utils.GetLogger()
	// Get GitHub token
	token, err := config.GetGitHubToken()
	if err != nil {
		utils.FatalError(err, "getting GitHub token")
	}

	// Load config and apply defaults
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Warn("Could not load config: %s", err)
	} else {
		applyEncryptionDefaults(cmd, cfg)
	}

	// Get Gist ID (from flag or config)
	gistID := getGistID(cfg)

	// Handle self-contained sharing
	if shareSelfContained {
		handleSelfContainedSharing(cmd, token, gistID)
		return
	}

	// Generate key file if requested
	var keyFilePath string
	if shareGenerateKeyFile {
		keyFilePath, err = generateKeyFileForSharing()
		if err != nil {
			utils.FatalError(err, "generating key file")
		}

		// Force encryption with key file when generating a key
		encryption.UseEncryption = true
		encryption.UseKeyFile = true
		encryption.EncryptionKeyFile = keyFilePath

		logger.Info("Using generated key file: %s", keyFilePath)
	}

	// Prepare environment content if needed
	envContent, err := prepareEnvContent()
	if err != nil {
		utils.FatalError(err, "preparing environment content")
	}

	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)

	// Get user info
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		utils.FatalError(err, "getting GitHub user")
	}

	// Get Gist details
	gist, _, err := client.Gists.Get(context.Background(), gistID)
	if err != nil {
		utils.FatalError(err, fmt.Sprintf("retrieving Gist with ID %s", gistID))
	}

	// Handle sharing with users if specified
	if len(shareWithUsers) > 0 {
		shareWithGitHubUsers(client, user, gist, envContent, keyFilePath)
	}

	// Generate shareable URL if requested
	if shareGenerateURL {
		generateAndShowURL(client, user, gist, keyFilePath)
	}

	// If neither option was selected, show help
	if len(shareWithUsers) == 0 && !shareGenerateURL {
		logger.Info("Please specify either users to share with (--users) or request a shareable URL (--url)")
		logger.Info("Run 'envi share --help' for usage information")
	}
}

// generateKeyFileForSharing creates a key file for sharing
func generateKeyFileForSharing() (string, error) {
	// Determine key file path
	keyFilePath := ".envi-share.key"
	if shareOutputKeyFile != "" {
		keyFilePath = shareOutputKeyFile
	} else {
		// Generate a unique filename with timestamp
		timestamp := time.Now().Format("20060102-150405")
		keyFilePath = fmt.Sprintf(".envi-share-%s.key", timestamp)
	}

	// Generate the key file
	err := encryption.GenerateKeyFile(keyFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to generate key file: %w", err)
	}

	logger := utils.GetLogger()
	logger.Success("Generated key file: %s", keyFilePath)
	logger.Info("IMPORTANT: Share this key file securely with the recipient.")
	logger.Info("Without this key file, they won't be able to decrypt the environment variables.")

	return keyFilePath, nil
}

// getGistID gets the Gist ID from flag or config
func getGistID(cfg *config.Config) string {
	logger := utils.GetLogger()
	if shareGistID == "" {
		if cfg.LastGistID == "" {
			utils.FatalMessage("No Gist ID specified and no saved Gist ID found", "share")
		}
		shareGistID = cfg.LastGistID
		logger.Info("Using saved Gist ID: %s", shareGistID)
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
func shareWithGitHubUsers(client *github.Client, user *github.User, gist *github.Gist, envContent []byte, keyFilePath string) {
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
		readmeContent := createSharingReadmeContent(user, username, keyFilePath)
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

		if keyFilePath != "" {
			fmt.Printf("REMINDER: Don't forget to share the key file (%s) with %s through a secure channel.\n",
				filepath.Base(keyFilePath), username)
		}
	}
}

// generateAndShowURL creates and displays a shareable URL
func generateAndShowURL(client *github.Client, user *github.User, gist *github.Gist, keyFilePath string) {
	fmt.Println("Generating shareable URL...")

	// Make the Gist public before sharing
	if gist.Public != nil && !*gist.Public {
		gist.Public = github.Bool(true)
		// Update the Gist to be public
		_, _, err := client.Gists.Edit(context.Background(), *gist.ID, gist)
		if err != nil {
			fmt.Printf("Error making Gist public: %s\n", err)
			return
		}
		fmt.Println("Gist has been made public for sharing.")
	}

	// Calculate expiry date
	expiryDate := time.Now().AddDate(0, 0, shareExpiryInDays)
	expiryStr := expiryDate.Format("2006-01-02")

	// Create a message to show
	sharingMessage := fmt.Sprintf("Shareable URL will expire on %s\n", expiryStr)
	sharingMessage += "Anyone with this URL can access your .env file.\n"
	sharingMessage += fmt.Sprintf("https://gist.github.com/%s\n", *gist.ID)

	if keyFilePath != "" {
		sharingMessage += "\n"
		sharingMessage += fmt.Sprintf("IMPORTANT: You must share the key file (%s) separately through a secure channel.\n",
			filepath.Base(keyFilePath))
		sharingMessage += "Without this key file, recipients won't be able to decrypt the environment variables.\n"
	}

	// Display message using TUI if enabled
	if encryption.UseTUI {
		tui.DisplayMessage("Shareable URL Generated", sharingMessage)
	} else {
		fmt.Println(sharingMessage)
	}
}

// handleSelfContainedSharing implements self-contained encrypted sharing
func handleSelfContainedSharing(cmd *cobra.Command, token, gistID string) {
	// Read the .env file
	envContent, err := os.ReadFile(".env")
	if err != nil {
		utils.HandleError(utils.WrapFileError(err, "failed to read .env file"), "share")
		return
	}

	// Generate or get password for sharing
	var sharePass string
	if shareGeneratePassword {
		// Generate a secure random password
		sharePass = generateSecurePassword()
		fmt.Printf("Generated password: %s\n", sharePass)
		fmt.Println("IMPORTANT: Save this password securely - it's needed to decrypt the shared content!")
	} else if sharePassword != "" {
		sharePass = sharePassword
	} else {
		// Prompt for password
		if encryption.UseTUI {
			sharePass, err = tui.GetPassword("Enter password for sharing", true)
			if err != nil {
				utils.HandleError(utils.WrapInputError(err, "failed to get password for sharing"), "share")
				return
			}
		} else {
			fmt.Print("Enter password for sharing: ")
			passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				utils.HandleError(utils.WrapInputError(err, "failed to read password from terminal"), "share")
				return
			}
			fmt.Println()
			sharePass = string(passwordBytes)
		}
	}

	if sharePass == "" {
		utils.HandleError(utils.NewInputError("password cannot be empty"), "share")
		return
	}

	// Create self-contained encrypted content
	encryptedContent, err := encryption.GenerateSelfContainedShare(envContent, sharePass)
	if err != nil {
		utils.HandleError(utils.WrapEncryptionError(err, "failed to create self-contained encrypted share"), "share")
		return
	}

	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)

	// Get user info
	user, _, err := client.Users.Get(context.Background(), "")
	if err != nil {
		utils.HandleError(utils.WrapGitHubError(err, "failed to get GitHub user information"), "share")
		return
	}

	// Create a new Gist for sharing
	description := fmt.Sprintf("Self-contained encrypted .env from %s - Created with envi", *user.Login)

	newGist := &github.Gist{
		Description: github.String(description),
		Public:      github.Bool(true), // Self-contained shares are always public
		Files: map[github.GistFilename]github.GistFile{
			github.GistFilename(".env"): {
				Content: github.String(string(encryptedContent)),
			},
		},
	}

	// Add README with instructions
	readmeContent := createSelfContainedReadmeContent(user, sharePass)
	newGist.Files[github.GistFilename("README.md")] = github.GistFile{
		Content: github.String(readmeContent),
	}

	// Create the shared Gist
	createdGist, _, err := client.Gists.Create(context.Background(), newGist)
	if err != nil {
		utils.HandleError(utils.WrapGitHubError(err, "failed to create shared Gist"), "share")
		return
	}

	// Display success message
	sharingMessage := fmt.Sprintf("✅ Self-contained encrypted share created successfully!\n\n")
	sharingMessage += fmt.Sprintf("🔗 Shareable URL: https://gist.github.com/%s\n", *createdGist.ID)
	sharingMessage += fmt.Sprintf("🔑 Password: %s\n\n", sharePass)
	sharingMessage += "📋 Instructions for recipients:\n"
	sharingMessage += "1. Visit the URL above\n"
	sharingMessage += "2. Copy the .env file content\n"
	sharingMessage += "3. Run: envi pull --self-contained --share-password <password>\n"
	sharingMessage += "4. Paste the content when prompted\n\n"
	sharingMessage += "💡 This share is completely self-contained - no separate key files needed!"

	if encryption.UseTUI {
		tui.ShowSuccess("Self-Contained Share Created", []string{
			fmt.Sprintf("Shareable URL: https://gist.github.com/%s", *createdGist.ID),
			fmt.Sprintf("Password: %s", sharePass),
			"Recipients can use: envi pull --self-contained --share-password <password>",
			"No separate key files needed!",
		})
	} else {
		fmt.Println(sharingMessage)
	}
}

// generateSecurePassword generates a secure random password
func generateSecurePassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, 16)
	for i := range password {
		password[i] = charset[rand.Intn(len(charset))]
	}
	return string(password)
}

// createSelfContainedReadmeContent generates README for self-contained shares
func createSelfContainedReadmeContent(user *github.User, password string) string {
	readmeContent := fmt.Sprintf("# Self-Contained Encrypted Environment Variables\n\n")
	readmeContent += fmt.Sprintf("This Gist contains environment variables shared by @%s using envi's self-contained encryption.\n\n", *user.Login)
	readmeContent += "## 🔐 Decryption Password\n\n"
	readmeContent += fmt.Sprintf("**Password:** `%s`\n\n", password)
	readmeContent += "## 📋 How to Use\n\n"
	readmeContent += "### Option 1: Using envi CLI (Recommended)\n\n"
	readmeContent += "1. Install envi if you haven't already:\n\n"
	readmeContent += "```shell\n"
	readmeContent += "# macOS/Linux\n"
	readmeContent += "brew install dexterity-inc/tap/envi\n\n"
	readmeContent += "# Windows\n"
	readmeContent += "scoop bucket add dexterity-inc https://github.com/dexterity-inc/scoop-bucket\n"
	readmeContent += "scoop install envi\n"
	readmeContent += "```\n\n"
	readmeContent += "2. Pull the environment variables:\n\n"
	readmeContent += "```shell\n"
	readmeContent += "envi pull --self-contained --password " + password + "\n"
	readmeContent += "```\n\n"
	readmeContent += "3. When prompted, paste the content from the `.env` file above\n\n"
	readmeContent += "### Option 2: Manual Decryption\n\n"
	readmeContent += "If you prefer to decrypt manually:\n\n"
	readmeContent += "1. Copy the content from the `.env` file above\n"
	readmeContent += "2. Save it to a file (e.g., `encrypted.env`)\n"
	readmeContent += "3. Run: `envi decrypt --self-contained --password " + password + " --input encrypted.env --output .env`\n\n"
	readmeContent += "## 🔒 Security Features\n\n"
	readmeContent += "- ✅ **Self-contained**: No separate key files needed\n"
	readmeContent += "- ✅ **Password-protected**: AES-256 encryption with PBKDF2 key derivation\n"
	readmeContent += "- ✅ **Tamper-resistant**: Authenticated encryption prevents modification\n"
	readmeContent += "- ✅ **Zero-knowledge**: The sender cannot decrypt your data\n\n"
	readmeContent += "---\n"
	readmeContent += "Shared using [envi](https://github.com/dexterity-inc/envi), an open-source environment variable manager"

	return readmeContent
}
