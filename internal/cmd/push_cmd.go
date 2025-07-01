package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/dexterity-inc/envi/internal/config"
	"github.com/dexterity-inc/envi/internal/encryption"
	"github.com/dexterity-inc/envi/internal/utils"
)

// Push command flags
var (
	pushEnvFile      string
	pushGistID       string
	pushDescription  string
	pushPublic       bool
	pushForce        bool
	pushAutoGenerate bool
	pushAutoDesc     bool
)

// pushCmd is the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push .env file to GitHub Gist",
	Long:  `Push your .env file to a GitHub Gist for secure storage and sharing.`,
	Run:   runPushCommand,
}

// InitPushCommand sets up the push command and its subcommands
func InitPushCommand() {
	// Initialize the command flags
	pushCmd.Flags().StringVarP(&pushEnvFile, "file", "f", ".env", "Path to .env file")
	pushCmd.Flags().StringVarP(&pushGistID, "id", "i", "", "Gist ID to update (if not specified, creates new Gist)")
	pushCmd.Flags().StringVarP(&pushDescription, "description", "d", "", "Description for the Gist")
	pushCmd.Flags().BoolVarP(&pushPublic, "public", "p", false, "Make Gist public")
	pushCmd.Flags().BoolVar(&pushForce, "force", false, "Force push without confirmation")
	pushCmd.Flags().BoolVar(&pushAutoGenerate, "auto", false, "Auto-generate .env file if it doesn't exist")
	pushCmd.Flags().BoolVar(&pushAutoDesc, "auto-desc", false, "Auto-generate description from .env file")

	// Add the push command to the root command
	rootCmd.AddCommand(pushCmd)
}

// generateDescription automatically creates a description based on the project
func generateDescription(envFile string) string {
	// Try to get Git repository name
	projectName := getProjectName()

	// Get environment type from filename
	environment := config.GetEnvironmentFromFilename(envFile)

	// Use enhanced description generation
	return config.GenerateGistDescription(envFile, projectName, environment, encryption.UseEncryption)
}

// getProjectName tries to get the project name from git or directory name
func getProjectName() string {
	// Try to get git repository name
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()

	if err == nil && len(output) > 0 {
		// Parse Git remote URL to extract repo name
		repoURL := strings.TrimSpace(string(output))

		// Handle SSH URL format (git@github.com:user/repo.git)
		if strings.HasPrefix(repoURL, "git@") {
			parts := strings.Split(repoURL, ":")
			if len(parts) > 1 {
				repoPath := parts[1]
				// Remove .git suffix if present
				repoPath = strings.TrimSuffix(repoPath, ".git")
				// Get the last part (repo name)
				pathParts := strings.Split(repoPath, "/")
				if len(pathParts) > 0 {
					return pathParts[len(pathParts)-1]
				}
			}
		}

		// Handle HTTPS URL format (https://github.com/user/repo.git)
		if strings.HasPrefix(repoURL, "http") {
			parts := strings.Split(repoURL, "/")
			if len(parts) > 0 {
				repoName := parts[len(parts)-1]
				// Remove .git suffix if present
				return strings.TrimSuffix(repoName, ".git")
			}
		}
	}

	// If git failed, use current directory name
	pwd, err := os.Getwd()
	if err == nil {
		return filepath.Base(pwd)
	}

	// If all else fails, return empty string
	return ""
}

// runPushCommand handles the push command execution
func runPushCommand(cmd *cobra.Command, args []string) {
	logger := utils.GetLogger()

	// Get GitHub token
	token, err := config.GetGitHubToken()
	if err != nil {
		utils.FatalError(err, "getting GitHub token")
	}

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Warn("Could not load config: %s", err)
	} else {
		// Apply encryption defaults if not explicitly set
		applyEncryptionDefaults(cmd, cfg)
	}

	// Check if .env file exists
	if _, err := os.Stat(pushEnvFile); os.IsNotExist(err) {
		if pushAutoGenerate {
			// Create a sample .env file
			logger.Info("No .env file found. Creating a sample at %s", pushEnvFile)
			sampleContent := "# Sample .env file created by envi\n" +
				"# Replace these with your actual environment variables\n\n" +
				"DB_HOST=localhost\n" +
				"DB_PORT=5432\n" +
				"DB_USER=username\n" +
				"DB_PASSWORD=password\n" +
				"API_KEY=your_api_key_here\n"

			if err := os.WriteFile(pushEnvFile, []byte(sampleContent), utils.EnvFilePerms); err != nil {
				utils.FatalError(err, "creating sample .env file")
			}
		} else {
			utils.FatalMessage(fmt.Sprintf(".env file not found at %s", pushEnvFile), "push")
		}
	}

	// Generate auto description if enabled and no description specified manually
	if pushAutoDesc && !cmd.Flags().Changed("description") {
		pushDescription = generateDescription(pushEnvFile)
		logger.Info("Using auto-generated description: %s", pushDescription)
	}

	// Read .env file
	envContent, err := os.ReadFile(pushEnvFile)
	if err != nil {
		utils.FatalError(err, "reading .env file")
	}

	// Handle encryption options
	if encryption.UseEncryption && encryption.UseMaskedEncryption {
		logger.Warn("Both --encrypt and --mask flags specified. Using --mask (masked encryption).")
		encryption.UseEncryption = false
	}

	if encryption.UseEncryption {
		logger.Info("Encrypting .env file...")
		encryptedContent, err := encryption.EncryptContent(envContent)
		if err != nil {
			utils.FatalError(err, "encrypting .env file")
		}
		envContent = encryptedContent
		logger.Success("Encryption successful.")
	} else if encryption.UseMaskedEncryption {
		logger.Info("Masking values in .env file...")
		maskedContent, err := encryption.MaskEnvContent(envContent)
		if err != nil {
			utils.FatalError(err, "masking .env file")
		}
		envContent = maskedContent
		logger.Success("Value masking successful. Variable names remain visible.")
	}

	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)

	// Get Gist ID (from flag or config)
	if pushGistID == "" && cfg != nil && cfg.LastGistID != "" {
		useLastID, err := utils.Confirm("Use saved Gist?", fmt.Sprintf("Would you like to update your last used Gist (%s)?", cfg.LastGistID))
		if err != nil {
			utils.FatalError(err, "getting confirmation")
		}

		if useLastID {
			pushGistID = cfg.LastGistID
			logger.Info("Using saved Gist ID: %s", pushGistID)
		}
	}

	// Create or update Gist
	var gist *github.Gist
	var gistID string

	if pushGistID == "" {
		// Create new Gist
		logger.Info("Creating new Gist...")
		gist, err = createNewGist(client, envContent, pushDescription, pushPublic)
		if err != nil {
			utils.FatalError(err, "creating new Gist")
		}
		gistID = *gist.ID
		logger.Success("Gist created successfully!")
	} else {
		// Update existing Gist
		logger.Info("Updating existing Gist...")
		gist, err = updateExistingGist(client, pushGistID, envContent, pushDescription, pushPublic)
		if err != nil {
			utils.FatalError(err, "updating existing Gist")
		}
		gistID = *gist.ID
		logger.Success("Gist updated successfully!")
	}

	// Save Gist ID to config
	if cfg != nil {
		cfg.LastGistID = gistID

		// Add to gist history
		gistInfo := &config.GistInfo{
			ID:          gistID,
			Name:        pushDescription,
			Description: pushDescription,
			CreatedAt:   gist.CreatedAt.Format(utils.TimeFormatShort),
			UpdatedAt:   gist.UpdatedAt.Format(utils.TimeFormatShort),
			IsEncrypted: encryption.UseEncryption || encryption.UseMaskedEncryption,
			IsPublic:    pushPublic,
			FileCount:   len(gist.Files),
			URL:         fmt.Sprintf("https://gist.github.com/%s", gistID),
		}
		cfg.AddGistToHistory(gistInfo)

		if err := config.SaveConfig(cfg); err != nil {
			logger.Warn("Could not save Gist ID to config: %s", err)
		}
	}

	// Display results
	logger.Info("Gist URL: https://gist.github.com/%s", gistID)
	if pushPublic {
		logger.Info("Gist is public and can be accessed by anyone with the URL")
	} else {
		logger.Info("Gist is private and only accessible to you")
	}

	if encryption.UseEncryption || encryption.UseMaskedEncryption {
		logger.Info("Content is encrypted - keep your encryption key/password secure!")
	}
}

// createReadmeContent creates a helpful README for the Gist
func createReadmeContent(fullEncryption, maskedEncryption bool) string {
	readme := "# Environment Variables\n\n" +
		"This .env file was created with [envi](https://github.com/dexterity-inc/envi).\n\n"

	if fullEncryption {
		readme += "## Encryption Notice\n\n" +
			"This .env file is fully encrypted and requires decryption to use.\n\n" +
			"To decrypt and use this file, install envi and run:\n\n" +
			"```shell\n" +
			"envi pull --id <gist-id> --unmask\n" +
			"```\n\n" +
			"You will need the encryption password or key file that was used to encrypt this file.\n"
	} else if maskedEncryption {
		readme += "## Encryption Notice\n\n" +
			"The values in this .env file are masked (encrypted). The variable names are visible, but the values need to be unmasked to use.\n\n" +
			"To unmask and use this file, install envi and run:\n\n" +
			"```shell\n" +
			"envi pull --id <gist-id> --unmask\n" +
			"```\n\n" +
			"You will need the encryption password or key file that was used to mask the values.\n"
	}

	readme += "\n## Install envi\n\n" +
		"```shell\n" +
		"# macOS/Linux\n" +
		"brew tap dexterity-inc/tap\n" +
		"brew install envi\n" +
		"\n" +
		"# Or download directly\n" +
		"curl -sSL https://github.com/dexterity-inc/envi/releases/latest/download/envi-$(uname -s)-$(uname -m) -o /usr/local/bin/envi\n" +
		"chmod +x /usr/local/bin/envi\n" +
		"\n" +
		"# Windows (via Scoop)\n" +
		"scoop bucket add dexterity-inc https://github.com/dexterity-inc/scoop-bucket\n" +
		"scoop install envi\n" +
		"```\n"

	return readme
}

// applyEncryptionDefaults applies encryption defaults from config
func applyEncryptionDefaults(cmd *cobra.Command, cfg *config.Config) {
	// Apply config defaults if not explicitly set
	if !cmd.Flags().Changed("encrypt") && !cmd.Flags().Changed("mask") && cfg.EncryptByDefault {
		if cfg.UseMaskedEncryption {
			encryption.UseMaskedEncryption = true
			fmt.Println("Using default setting: Masking values in .env file (variable names visible)")
		} else {
			encryption.UseEncryption = true
			fmt.Println("Using default setting: Fully encrypting .env file")
		}
	}

	if !cmd.Flags().Changed("use-key-file") && cfg.UseKeyFileByDefault {
		encryption.UseKeyFile = true
		fmt.Println("Using default setting: Using key file for encryption")
	}

	if !cmd.Flags().Changed("key-file") && cfg.DefaultKeyFile != "" {
		encryption.EncryptionKeyFile = cfg.DefaultKeyFile
		fmt.Printf("Using default key file: %s\n", encryption.EncryptionKeyFile)
	}
}

// createNewGist creates a new GitHub Gist
func createNewGist(client *github.Client, content []byte, description string, public bool) (*github.Gist, error) {
	// Create the Gist files
	files := map[github.GistFilename]github.GistFile{
		".env": {
			Content: github.String(string(content)),
		},
	}

	// Add README if content is encrypted
	if encryption.UseEncryption || encryption.UseMaskedEncryption {
		readmeContent := createReadmeContent(encryption.UseEncryption, encryption.UseMaskedEncryption)
		files["README.md"] = github.GistFile{
			Content: github.String(readmeContent),
		}
	}

	// Create the Gist
	gist := &github.Gist{
		Description: github.String(description),
		Public:      github.Bool(public),
		Files:       files,
	}

	// Create the Gist via GitHub API
	createdGist, _, err := client.Gists.Create(context.Background(), gist)
	if err != nil {
		return nil, fmt.Errorf("failed to create Gist: %w", err)
	}

	return createdGist, nil
}

// updateExistingGist updates an existing GitHub Gist
func updateExistingGist(client *github.Client, gistID string, content []byte, description string, public bool) (*github.Gist, error) {
	// Get the existing Gist first
	existingGist, _, err := client.Gists.Get(context.Background(), gistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing Gist: %w", err)
	}

	// Update the .env file content
	existingGist.Files[".env"] = github.GistFile{
		Content: github.String(string(content)),
	}

	// Update README if content is encrypted
	if encryption.UseEncryption || encryption.UseMaskedEncryption {
		readmeContent := createReadmeContent(encryption.UseEncryption, encryption.UseMaskedEncryption)
		existingGist.Files["README.md"] = github.GistFile{
			Content: github.String(readmeContent),
		}
	}

	// Update description and public status if provided
	if description != "" {
		existingGist.Description = github.String(description)
	}
	existingGist.Public = github.Bool(public)

	// Update the Gist via GitHub API
	updatedGist, _, err := client.Gists.Edit(context.Background(), gistID, existingGist)
	if err != nil {
		return nil, fmt.Errorf("failed to update Gist: %w", err)
	}

	return updatedGist, nil
}
