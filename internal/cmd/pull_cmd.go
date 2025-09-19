package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
	"golang.org/x/term"

	"github.com/dexterity-inc/envi/internal/config"
	"github.com/dexterity-inc/envi/internal/encryption"
	"github.com/dexterity-inc/envi/internal/security"
	"github.com/dexterity-inc/envi/internal/tui"
	"github.com/dexterity-inc/envi/internal/utils"
)

// Pull command flags
var (
	pullGistID        string
	pullOutput        string
	pullUnmask        bool
	pullForce         bool
	pullSelfContained bool
	pullSharePassword string
)

// pullCmd is the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull .env file from GitHub Gist",
	Long:  `Pull your .env file from a GitHub Gist with optional decryption.`,
	Run:   runPullCommand,
}

// InitPullCommand sets up the pull command and its subcommands
func InitPullCommand() {
	// Initialize the command flags
	pullCmd.Flags().StringVarP(&pullGistID, "id", "i", "", "GitHub Gist ID to pull from")
	pullCmd.Flags().StringVarP(&pullOutput, "output", "o", ".env", "Output file path")
	pullCmd.Flags().BoolVarP(&pullUnmask, "unmask", "u", false, "Decrypt/unmask values when pulling")
	pullCmd.Flags().BoolVarP(&pullForce, "force", "f", false, "Overwrite existing file without confirmation")
	pullCmd.Flags().BoolVarP(&pullSelfContained, "self-contained", "s", false, "Pull a self-contained encrypted share")
	pullCmd.Flags().StringVar(&pullSharePassword, "share-password", "", "Password for decrypting a self-contained encrypted share")

	// Add encryption flags for decryption
	pullCmd.Flags().BoolVar(&encryption.UseKeyFile, "use-key-file", false, "Use key file instead of password")
	pullCmd.Flags().StringVarP(&encryption.EncryptionKeyFile, "key-file", "k", ".envi.key", "Path to encryption key file")
	// Password flag removed for security - passwords should only be entered interactively

	// Add the pull command to the root command
	rootCmd.AddCommand(pullCmd)
}

// runPullCommand handles the pull command execution
func runPullCommand(cmd *cobra.Command, args []string) {
	logger := utils.GetLogger()

	// Validate output file path for security
	if err := security.ValidateOutputPath(pullOutput); err != nil {
		utils.Error("Invalid output file path: %s", err)
		utils.Fatal("Security validation failed")
	}

	// Validate key file path if using key file
	if encryption.UseKeyFile {
		if err := security.ValidateKeyFilePath(encryption.EncryptionKeyFile); err != nil {
			utils.Error("Invalid key file path: %s", err)
			utils.Fatal("Security validation failed")
		}
	}

	// Handle self-contained pull
	if pullSelfContained {
		handleSelfContainedPull(cmd)
		return
	}

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
		// Apply config defaults
		if !cmd.Flags().Changed("unmask") && cfg != nil && cfg.UnmaskByDefault {
			pullUnmask = true
			logger.Info("Using default setting: Automatically unmasking values")
		}

		if !cmd.Flags().Changed("use-key-file") && cfg.UseKeyFileByDefault {
			encryption.UseKeyFile = true
			logger.Info("Using default setting: Using key file for decryption")
		}

		if !cmd.Flags().Changed("key-file") && cfg.DefaultKeyFile != "" {
			encryption.EncryptionKeyFile = cfg.DefaultKeyFile
			logger.Info("Using default key file: %s", encryption.EncryptionKeyFile)
		}
	}

	// Get Gist ID (from flag or config)
	if pullGistID == "" && cfg != nil && cfg.LastGistID != "" {
		useLastID, err := utils.Confirm(
			"Use saved Gist?",
			fmt.Sprintf("Would you like to pull from your last used Gist (%s)?", cfg.LastGistID),
		)
		if err != nil {
			utils.FatalError(err, "getting confirmation")
		}

		if useLastID {
			pullGistID = cfg.LastGistID
			logger.Info("Using saved Gist ID: %s", pullGistID)
		}
	}

	// Check if Gist ID is provided
	if pullGistID == "" {
		utils.FatalMessage("No Gist ID specified and no saved Gist ID found", "pull")
	}

	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)

	// Get Gist
	gist, _, err := client.Gists.Get(cmd.Context(), pullGistID)
	if err != nil {
		utils.FatalError(err, fmt.Sprintf("retrieving Gist with ID %s", pullGistID))
	}

	// Find .env file in Gist
	var envFile *github.GistFile
	for filename, file := range gist.Files {
		if string(filename) == ".env" {
			envFile = &file
			break
		}
	}

	if envFile == nil {
		utils.FatalMessage("No .env file found in this Gist", "pull")
	}

	// Get content
	envContent := []byte(*envFile.Content)

	// Check if content is encrypted and needs decryption
	isEncrypted := encryption.IsEncrypted(envContent)
	isMasked := encryption.IsMasked(envContent)

	if (isEncrypted || isMasked) && pullUnmask {
		fmt.Println("Detected encrypted content. Attempting to decrypt...")

		var decryptedContent []byte
		var err error

		if isEncrypted {
			decryptedContent, err = encryption.DecryptContent(envContent)
		} else if isMasked {
			decryptedContent, err = encryption.UnmaskEnvContent(envContent)
		}

		if err != nil {
			fmt.Println("Error decrypting content. Please check the encryption key or password and try again.")
			os.Exit(1)
		}

		envContent = decryptedContent
		fmt.Println("Successfully decrypted content!")
	} else if (isEncrypted || isMasked) && !pullUnmask {
		fmt.Println("Note: Content is encrypted/masked but --unmask flag was not specified.")
		fmt.Println("The file will be saved in its encrypted form.")
		fmt.Println("To decrypt, run 'envi pull --id " + pullGistID + " --unmask'")
	}

	// Check if output file already exists
	if _, err := os.Stat(pullOutput); err == nil && !pullForce {
		var overwrite bool

		if encryption.UseTUI {
			overwrite, err = tui.Confirm(
				"Overwrite file?",
				fmt.Sprintf("The file %s already exists. Overwrite?", pullOutput),
			)
			if err != nil {
				utils.HandleError(utils.WrapInputError(err, "failed to get confirmation for file overwrite"), "pull")
				return
			}
		} else {
			fmt.Printf("The file %s already exists. Overwrite? (y/N)", pullOutput)
			var response string
			fmt.Scanln(&response)
			overwrite = strings.ToLower(response) == "y"
		}

		if !overwrite {
			fmt.Println("Operation canceled.")
			os.Exit(0)
		}
	}

	// Write the file
	err = os.WriteFile(pullOutput, envContent, 0600)
	if err != nil {
		utils.HandleError(utils.WrapFileError(err, fmt.Sprintf("failed to write file %s", pullOutput)), "pull")
		return
	}

	fmt.Printf("Successfully pulled .env file to %s\n", pullOutput)

	// Save the Gist ID to config for future use
	if cfg != nil {
		cfg.LastGistID = pullGistID

		// Update usage statistics
		cfg.UpdateGistUsage(pullGistID)

		if err := config.SaveConfig(cfg); err != nil {
			utils.Warn("Could not save Gist ID to config: %v", err)
		}
	}
}

// handleSelfContainedPull handles pulling self-contained encrypted shares
func handleSelfContainedPull(cmd *cobra.Command) {
	// Get the share password
	var sharePassword string
	if pullSharePassword != "" {
		sharePassword = pullSharePassword
	} else {
		// Prompt for password
		if encryption.UseTUI {
			var err error
			sharePassword, err = tui.GetPassword("Enter password for self-contained share", false)
			if err != nil {
				utils.HandleError(utils.WrapInputError(err, "failed to get password for self-contained share"), "pull")
				return
			}
		} else {
			fmt.Print("Enter password for self-contained share: ")
			passwordBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				utils.HandleError(utils.WrapInputError(err, "failed to read password from terminal"), "pull")
				return
			}
			fmt.Println()
			sharePassword = string(passwordBytes)
		}
	}

	if sharePassword == "" {
		utils.HandleError(utils.NewInputError("password cannot be empty"), "pull")
		return
	}

	// Get the encrypted content
	var encryptedContent []byte
	var err error

	if encryption.UseTUI {
		// Use TUI to get content
		textContent, err := tui.GetText("Self-Contained Share", "Encrypted Content",
			"Paste the encrypted content from the shared Gist", "",
			"Paste the content from the .env file in the shared Gist", true)
		if err != nil {
			utils.HandleError(utils.WrapInputError(err, "failed to get encrypted content"), "pull")
			return
		}
		encryptedContent = []byte(textContent)
	} else {
		// Use terminal input
		fmt.Println("Paste the encrypted content from the shared Gist (press Enter when done):")
		var inputLines []string
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				break
			}
			inputLines = append(inputLines, line)
		}
		encryptedContent = []byte(strings.Join(inputLines, "\n"))
	}

	if len(encryptedContent) == 0 {
		utils.HandleError(utils.NewInputError("no encrypted content provided"), "pull")
		return
	}

	// Check if it's a self-contained share
	if !encryption.IsSelfContainedShare(encryptedContent) {
		if encryption.UseTUI {
			tui.ShowError(utils.NewValidationError("not a self-contained encrypted share"), "pull")
		} else {
			fmt.Println("Error: The provided content is not a self-contained encrypted share")
			fmt.Println("Make sure you copied the entire content from the .env file in the shared Gist")
			fmt.Printf("Content starts with: %s\n", string(encryptedContent[:min(50, len(encryptedContent))]))
		}
		utils.HandleError(utils.NewValidationError("content is not a self-contained encrypted share"), "pull")
		return
	}

	// Decrypt the content
	fmt.Println("Decrypting self-contained share...")
	decryptedContent, err := encryption.DecryptSelfContainedShare(encryptedContent, sharePassword)
	if err != nil {
		if encryption.UseTUI {
			tui.ShowError(err, "decrypt")
		} else {
			fmt.Println("Error decrypting content. Please check that:")
			fmt.Println("1. The password is correct")
			fmt.Println("2. The encrypted content was copied completely")
			fmt.Println("3. The content hasn't been modified")
		}
		utils.HandleError(utils.WrapEncryptionError(err, "failed to decrypt self-contained share"), "pull")
		return
	}

	// Check if output file already exists
	if _, err := os.Stat(pullOutput); err == nil && !pullForce {
		var overwrite bool

		if encryption.UseTUI {
			overwrite, err = tui.Confirm(
				"Overwrite file?",
				fmt.Sprintf("The file %s already exists. Overwrite?", pullOutput),
			)
			if err != nil {
				utils.HandleError(utils.WrapInputError(err, "failed to get confirmation for file overwrite"), "pull")
				return
			}
		} else {
			fmt.Printf("The file %s already exists. Overwrite? (y/N): ", pullOutput)
			var response string
			fmt.Scanln(&response)
			overwrite = strings.ToLower(response) == "y"
		}

		if !overwrite {
			fmt.Println("Operation canceled.")
			os.Exit(0)
		}
	}

	// Write the decrypted content
	err = os.WriteFile(pullOutput, decryptedContent, 0600)
	if err != nil {
		utils.HandleError(utils.WrapFileError(err, fmt.Sprintf("failed to write decrypted content to %s", pullOutput)), "pull")
		return
	}

	if encryption.UseTUI {
		tui.ShowSuccess("Successfully Decrypted", []string{
			fmt.Sprintf("Environment variables saved to: %s", pullOutput),
			"Content was decrypted using self-contained encryption method",
		})
	} else {
		fmt.Printf("✅ Successfully decrypted and saved environment variables to %s\n", pullOutput)
		fmt.Println("🔒 The content was decrypted using the self-contained encryption method")
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
