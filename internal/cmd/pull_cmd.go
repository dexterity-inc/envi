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

var (
	pullGistID        string
	pullOutput        string
	pullUnmask        bool
	pullForce         bool
	pullSelfContained bool
)

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull .env file from GitHub Gist",
	Long:  `Pull your .env file from a GitHub Gist with optional decryption.`,
	Run:   runPullCommand,
}

func InitPullCommand() {
	pullCmd.Flags().StringVarP(&pullGistID, "id", "i", "", "GitHub Gist ID to pull from")
	pullCmd.Flags().StringVarP(&pullOutput, "output", "o", ".env", "Output file path")
	pullCmd.Flags().BoolVarP(&pullUnmask, "unmask", "u", false, "Decrypt/unmask values when pulling")
	pullCmd.Flags().BoolVarP(&pullForce, "force", "f", false, "Overwrite existing file without confirmation")
	pullCmd.Flags().BoolVarP(&pullSelfContained, "self-contained", "s", false, "Pull a self-contained encrypted share")

	pullCmd.Flags().BoolVar(&encryption.UseKeyFile, "use-key-file", false, "Use key file instead of password")
	pullCmd.Flags().StringVarP(&encryption.EncryptionKeyFile, "key-file", "k", ".envi.key", "Path to encryption key file")

	rootCmd.AddCommand(pullCmd)
}

func runPullCommand(cmd *cobra.Command, args []string) {
	logger := utils.GetLogger()

	if err := security.ValidateOutputPath(pullOutput); err != nil {
		utils.Error("Invalid output file path: %s", err)
		utils.Fatal("Security validation failed")
	}

	if encryption.UseKeyFile {
		if err := security.ValidateKeyFilePath(encryption.EncryptionKeyFile); err != nil {
			utils.Error("Invalid key file path: %s", err)
			utils.Fatal("Security validation failed")
		}
	}

	if pullSelfContained {
		handleSelfContainedPull(cmd)
		return
	}

	token, err := config.GetGitHubToken()
	if err != nil {
		utils.FatalError(err, "getting GitHub token")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Warn("Could not load config: %s", err)
	} else {
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

	if pullGistID == "" {
		utils.Fatal("No Gist ID specified and no saved Gist ID found")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)

	gist, _, err := client.Gists.Get(cmd.Context(), pullGistID)
	if err != nil {
		utils.FatalError(err, fmt.Sprintf("retrieving Gist with ID %s", pullGistID))
	}

	var envFile *github.GistFile
	for filename, file := range gist.Files {
		if string(filename) == ".env" {
			envFile = &file
			break
		}
	}

	if envFile == nil {
		utils.Fatal("No .env file found in this Gist")
	}

	envContent := []byte(*envFile.Content)
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

	err = os.WriteFile(pullOutput, envContent, 0600)
	if err != nil {
		utils.HandleError(utils.WrapFileError(err, fmt.Sprintf("failed to write file %s", pullOutput)), "pull")
		return
	}

	fmt.Printf("Successfully pulled .env file to %s\n", pullOutput)

	if cfg != nil {
		cfg.LastGistID = pullGistID
		cfg.UpdateGistUsage(pullGistID)

		if err := config.SaveConfig(cfg); err != nil {
			utils.Warn("Could not save Gist ID to config: %v", err)
		}
	}
}

func handleSelfContainedPull(cmd *cobra.Command) {
	var sharePassword string

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

	if sharePassword == "" {
		utils.HandleError(utils.NewInputError("password cannot be empty"), "pull")
		return
	}

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
