package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/dexterity-inc/envi/internal/config"
	"github.com/dexterity-inc/envi/internal/encryption"
	"github.com/dexterity-inc/envi/internal/tui"
)

// Pull command flags
var (
	pullGistID      string
	pullOutput      string
	pullUnmask      bool
	pullForce       bool
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
	
	// Add encryption flags for decryption
	pullCmd.Flags().BoolVar(&encryption.UseKeyFile, "use-key-file", false, "Use key file instead of password")
	pullCmd.Flags().StringVarP(&encryption.EncryptionKeyFile, "key-file", "k", ".envi.key", "Path to encryption key file")
	pullCmd.Flags().StringVarP(&encryption.EncryptionPassword, "password", "p", "", "Encryption password (not recommended)")

	// Add the pull command to the root command
	rootCmd.AddCommand(pullCmd)
}

// runPullCommand handles the pull command execution
func runPullCommand(cmd *cobra.Command, args []string) {
	// Get GitHub token
	token, err := config.GetGitHubToken()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Warning: Could not load config: %s\n", err)
	} else {
		// Apply config defaults
		if !cmd.Flags().Changed("unmask") && cfg != nil && cfg.UnmaskByDefault {
			pullUnmask = true
			fmt.Println("Using default setting: Automatically unmasking values")
		}
		
		if !cmd.Flags().Changed("use-key-file") && cfg.UseKeyFileByDefault {
			encryption.UseKeyFile = true
			fmt.Println("Using default setting: Using key file for decryption")
		}
		
		if !cmd.Flags().Changed("key-file") && cfg.DefaultKeyFile != "" {
			encryption.EncryptionKeyFile = cfg.DefaultKeyFile
			fmt.Printf("Using default key file: %s\n", encryption.EncryptionKeyFile)
		}
	}
	
	// Get Gist ID (from flag or config)
	if pullGistID == "" && cfg != nil && cfg.LastGistID != "" {
		if encryption.UseTUI {
			// Ask user if they want to use the last Gist ID
			useLastID, err := tui.Confirm(
				"Use saved Gist?",
				fmt.Sprintf("Would you like to pull from your last used Gist (%s)?", cfg.LastGistID),
			)
			if err != nil {
				fmt.Printf("Error getting confirmation: %s\n", err)
				os.Exit(1)
			}
			
			if useLastID {
				pullGistID = cfg.LastGistID
				fmt.Printf("Using saved Gist ID: %s\n", pullGistID)
			}
		} else {
			// Ask in plain terminal mode
			fmt.Printf("Found saved Gist ID: %s\n", cfg.LastGistID)
			fmt.Println("Would you like to pull from this Gist? (y/N)")
			
			var response string
			fmt.Scanln(&response)
			
			if strings.ToLower(response) == "y" {
				pullGistID = cfg.LastGistID
				fmt.Printf("Using saved Gist ID: %s\n", pullGistID)
			}
		}
	}
	
	// Check if Gist ID is provided
	if pullGistID == "" {
		fmt.Println("Error: No Gist ID specified and no saved Gist ID found")
		fmt.Println("Use 'envi pull --id GIST_ID' or first push an .env file with 'envi push'")
		os.Exit(1)
	}
	
	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)
	
	// Get Gist
	gist, _, err := client.Gists.Get(cmd.Context(), pullGistID)
	if err != nil {
		fmt.Printf("Error retrieving Gist with ID %s: %s\n", pullGistID, err)
		os.Exit(1)
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
		fmt.Println("Error: No .env file found in this Gist")
		os.Exit(1)
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
				fmt.Printf("Error getting confirmation: %s\n", err)
				os.Exit(1)
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
	
	// Write content to file
	if err := ioutil.WriteFile(pullOutput, envContent, 0600); err != nil {
		fmt.Printf("Error writing to %s: %s\n", pullOutput, err)
		os.Exit(1)
	}
	
	fmt.Printf("Successfully pulled .env file to %s\n", pullOutput)
	
	// Save Gist ID in config if it's not already saved
	if cfg != nil && cfg.LastGistID != pullGistID {
		cfg.LastGistID = pullGistID
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("Warning: Could not save Gist ID to config: %s\n", err)
		} else {
			fmt.Println("Saved Gist ID for future use")
		}
	}
} 