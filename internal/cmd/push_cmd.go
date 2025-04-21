package cmd

import (
	"fmt"
	"os"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/dexterity-inc/envi/internal/config"
	"github.com/dexterity-inc/envi/internal/encryption"
	"github.com/dexterity-inc/envi/internal/tui"
)

// Push command flags
var (
	pushGistID        string
	pushDescription   string
	pushPublic        bool
	pushEnvFile       string
	pushAutoGenerate  bool
)

// pushCmd is the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push .env file to GitHub Gist",
	Long:  `Push your .env file to a new or existing GitHub Gist with optional encryption.`,
	Run:   runPushCommand,
}

// InitPushCommand sets up the push command and its subcommands
func InitPushCommand() {
	// Initialize the command flags
	pushCmd.Flags().StringVarP(&pushGistID, "id", "i", "", "GitHub Gist ID to update (leave blank for new Gist)")
	pushCmd.Flags().StringVarP(&pushDescription, "description", "d", "Environment variables created with envi", "Description for the Gist")
	pushCmd.Flags().BoolVarP(&pushPublic, "public", "p", false, "Make the Gist public (default private)")
	pushCmd.Flags().StringVarP(&pushEnvFile, "file", "f", ".env", "Path to the .env file")
	pushCmd.Flags().BoolVarP(&pushAutoGenerate, "auto", "a", false, "Auto-generate a sample .env file if none exists")
	
	// Add the push command to the root command
	rootCmd.AddCommand(pushCmd)
}

// runPushCommand handles the push command execution
func runPushCommand(cmd *cobra.Command, args []string) {
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
		// Apply encryption defaults if not explicitly set
		applyEncryptionDefaults(cmd, cfg)
	}
	
	// Check if .env file exists
	if _, err := os.Stat(pushEnvFile); os.IsNotExist(err) {
		if pushAutoGenerate {
			// Create a sample .env file
			fmt.Printf("No .env file found. Creating a sample at %s\n", pushEnvFile)
			sampleContent := "# Sample .env file created by envi\n" +
				"# Replace these with your actual environment variables\n\n" +
				"DB_HOST=localhost\n" +
				"DB_PORT=5432\n" +
				"DB_USER=username\n" +
				"DB_PASSWORD=password\n" +
				"API_KEY=your_api_key_here\n"
			
			if err := os.WriteFile(pushEnvFile, []byte(sampleContent), 0600); err != nil {
				fmt.Printf("Error creating sample .env file: %s\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Error: .env file not found at %s\n", pushEnvFile)
			fmt.Println("Create the file first or use --auto to generate a sample")
			os.Exit(1)
		}
	}
	
	// Read .env file
	envContent, err := os.ReadFile(pushEnvFile)
	if err != nil {
		fmt.Printf("Error reading .env file: %s\n", err)
		os.Exit(1)
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
			fmt.Printf("Error encrypting .env file: %s\n", err)
			os.Exit(1)
		}
		envContent = encryptedContent
		fmt.Println("Encryption successful.")
	} else if encryption.UseMaskedEncryption {
		fmt.Println("Masking values in .env file...")
		maskedContent, err := encryption.MaskEnvContent(envContent)
		if err != nil {
			fmt.Println("Error masking .env file. Please check the input and try again.")
			os.Exit(1)
		}
		envContent = maskedContent
		fmt.Println("Value masking successful. Variable names remain visible.")
	}
	
	// Create GitHub client
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(cmd.Context(), ts)
	client := github.NewClient(tc)
	
	// Get Gist ID (from flag or config)
	if pushGistID == "" && cfg != nil && cfg.LastGistID != "" {
		useLastID, err := tui.Confirm("Use saved Gist?", fmt.Sprintf("Would you like to update your last used Gist (%s)?", cfg.LastGistID))
		if err != nil {
			fmt.Printf("Error getting confirmation: %s\n", err)
			os.Exit(1)
		}
		
		if useLastID {
			pushGistID = cfg.LastGistID
			fmt.Printf("Using saved Gist ID: %s\n", pushGistID)
		}
	}
	
	// Create or update Gist
	if pushGistID == "" {
		// Create new Gist
		newGist := &github.Gist{
			Description: github.String(pushDescription),
			Public:      github.Bool(pushPublic),
			Files: map[github.GistFilename]github.GistFile{
				github.GistFilename(".env"): {
					Content: github.String(string(envContent)),
				},
			},
		}
		
		// Add README with instructions if encrypted
		if encryption.UseEncryption || encryption.UseMaskedEncryption {
			readmeContent := createReadmeContent(encryption.UseEncryption, encryption.UseMaskedEncryption)
			newGist.Files[github.GistFilename("README.md")] = github.GistFile{
				Content: github.String(readmeContent),
			}
		}
		
		// Create the Gist
		gist, _, err := client.Gists.Create(cmd.Context(), newGist)
		if err != nil {
			fmt.Printf("Error creating Gist: %s\n", err)
			os.Exit(1)
		}
		
		// Save Gist ID in config
		if cfg != nil {
			cfg.LastGistID = *gist.ID
			if err := config.SaveConfig(cfg); err != nil {
				fmt.Printf("Warning: Could not save Gist ID to config: %s\n", err)
			}
		}
		
		fmt.Println("Successfully pushed .env to GitHub Gist!")
		fmt.Printf("Gist URL: https://gist.github.com/%s\n", *gist.ID)
		fmt.Printf("Gist ID: %s (saved for future use)\n", *gist.ID)
	} else {
		// Update existing Gist
		// First, get the current Gist to preserve other files
		gist, _, err := client.Gists.Get(cmd.Context(), pushGistID)
		if err != nil {
			fmt.Printf("Error retrieving Gist with ID %s: %s\n", pushGistID, err)
			os.Exit(1)
		}
		
		// Update the Gist
		gist.Files = map[github.GistFilename]github.GistFile{
			github.GistFilename(".env"): {
				Content: github.String(string(envContent)),
			},
		}
		
		// Add README with instructions if encrypted
		if encryption.UseEncryption || encryption.UseMaskedEncryption {
			readmeContent := createReadmeContent(encryption.UseEncryption, encryption.UseMaskedEncryption)
			gist.Files[github.GistFilename("README.md")] = github.GistFile{
				Content: github.String(readmeContent),
			}
		}
		
		// Update Gist description if provided
		if pushDescription != "Environment variables created with envi" && cmd.Flags().Changed("description") {
			gist.Description = github.String(pushDescription)
		}
		
		// Update the Gist
		_, _, err = client.Gists.Edit(cmd.Context(), pushGistID, gist)
		if err != nil {
			fmt.Printf("Error updating Gist: %s\n", err)
			os.Exit(1)
		}
		
		fmt.Println("Successfully updated .env in GitHub Gist!")
		fmt.Printf("Gist URL: https://gist.github.com/%s\n", pushGistID)
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
		"```\n\n" +
		"Learn more at https://github.com/dexterity-inc/envi\n"
	
	return readme
}

// applyEncryptionDefaults applies default encryption settings from config
func applyEncryptionDefaults(cmd *cobra.Command, cfg *config.Config) {
	// Apply default encryption settings if not explicitly set by flags
	if !cmd.Flags().Changed("encrypt") && !cmd.Flags().Changed("mask") && cfg.EncryptByDefault {
		if cfg.UseMaskedEncryption {
			encryption.UseMaskedEncryption = true
			encryption.UseEncryption = false
			fmt.Println("Using default setting: Masked encryption enabled")
		} else {
			encryption.UseEncryption = true
			encryption.UseMaskedEncryption = false
			fmt.Println("Using default setting: Full encryption enabled")
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