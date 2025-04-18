package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	// Flag to specify which Gist to pull from
	pullGistID string
	
	// Flag to force overwrite of existing .env file
	forceOverwrite bool
	
	// Flag to create a backup of the existing .env file
	backupFlag bool
)

func init() {
	pullCmd.Flags().StringVarP(&pullGistID, "id", "i", "", "GitHub Gist ID to pull from")
	pullCmd.Flags().BoolVarP(&forceOverwrite, "force", "f", false, "Force overwrite of existing .env file")
	pullCmd.Flags().BoolVarP(&backupFlag, "backup", "b", true, "Create a backup of existing .env file (default: true)")
	// Encryption flags now defined in encryption.go
	rootCmd.AddCommand(pullCmd)
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull .env file from GitHub Gist",
	Long:  `Pull .env file from a GitHub Gist and save it to the current directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config for encryption settings and last used Gist ID
		config, err := LoadConfig()
		if err != nil {
			fmt.Printf("Error loading config: %s\n", err)
			os.Exit(1)
		}
		
		// Apply default encryption settings if not explicitly set by flags
		if !cmd.Flags().Changed("decrypt") && !cmd.Flags().Changed("unmask") && config.EncryptByDefault {
			if config.UseMaskedEncryption {
				useMaskedEncryption = true
				fmt.Println("Using default setting: Unmasking enabled")
			} else {
				useEncryption = true
				fmt.Println("Using default setting: Decryption enabled")
			}
		}
		
		if !cmd.Flags().Changed("use-key-file") && config.UseKeyFileByDefault {
			useKeyFile = true
			fmt.Println("Using default setting: Using key file for decryption")
		}
		
		if !cmd.Flags().Changed("key-file") && config.DefaultKeyFile != "" {
			encryptionKeyFile = config.DefaultKeyFile
			fmt.Printf("Using default key file: %s\n", encryptionKeyFile)
		}
		
		// Check if .env exists and handle backups/confirmations
		if _, err := os.Stat(".env"); err == nil {
			if backupFlag {
				// Create backup with timestamp
				timestamp := time.Now().Format("20060102-150405")
				backupPath := fmt.Sprintf(".env.backup.%s", timestamp)
				
				envContent, err := os.ReadFile(".env")
				if err != nil {
					fmt.Printf("Error reading .env for backup: %s\n", err)
					os.Exit(1)
				}
				
				err = os.WriteFile(backupPath, envContent, 0600)
				if err != nil {
					fmt.Printf("Error creating backup file: %s\n", err)
					os.Exit(1)
				}
				
				fmt.Printf("Created backup at: %s\n", backupPath)
			} else if !forceOverwrite {
				// Ask for confirmation before overwriting
				fmt.Print("An .env file already exists. Overwrite? (y/n): ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Pull canceled")
					return
				}
			}
		}
		
		// If no Gist ID provided, check config for last used ID
		if pullGistID == "" {
			if config.LastGistID == "" {
				fmt.Println("Error: No Gist ID specified and no saved Gist ID found")
				fmt.Println("Use 'envi pull --id GIST_ID' or first push an .env file with 'envi push'")
				os.Exit(1)
			}
			
			// Ask user if they want to use the saved gist ID
			fmt.Printf("Found saved Gist ID: %s\n", config.LastGistID)
			fmt.Print("Do you want to pull from this Gist? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Pull canceled")
				return
			}
			
			pullGistID = config.LastGistID
		}

		// Get GitHub token
		token, err := GetGitHubToken()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Create GitHub client
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)

		// Get Gist from GitHub
		gist, _, err := client.Gists.Get(ctx, pullGistID)
		if err != nil {
			fmt.Printf("Error fetching Gist: %s\n", err)
			os.Exit(1)
		}

		// Extract .env file content
		var envContent string
		found := false
		for filename, file := range gist.Files {
			if filename == ".env" {
				envContent = *file.Content
				found = true
				break
			}
		}

		if !found {
			fmt.Println("Error: No .env file found in the specified Gist")
			os.Exit(1)
		}

		// Check if content is encrypted or masked
		isEncrypted := strings.HasPrefix(envContent, "ENVI_ENCRYPTED_V1")
		isMasked := strings.Contains(envContent, "# ENVI_MASKED_ENCRYPTION_V1")
		
		// Auto-detect encryption type and use appropriate flag
		if isEncrypted && isMasked {
			fmt.Println("Warning: Content appears to be both fully encrypted and masked. This is unexpected.")
			fmt.Println("Will attempt full decryption first.")
			isEncrypted = true
			isMasked = false
		} else if isEncrypted {
			fmt.Println("Detected fully encrypted .env file")
			
			if !useEncryption && !useMaskedEncryption {
				fmt.Print("The .env file is encrypted. Decrypt it? (y/n): ")
				var response string
				fmt.Scanln(&response)
				if response == "y" || response == "Y" {
					useEncryption = true
				} else {
					fmt.Println("Warning: Writing encrypted content to .env file without decryption.")
				}
			} else if useMaskedEncryption {
				fmt.Println("Warning: Content is fully encrypted but --unmask flag was specified.")
				fmt.Print("Switch to full decryption instead? (y/n): ")
				var response string
				fmt.Scanln(&response)
				if response == "y" || response == "Y" {
					useEncryption = true
					useMaskedEncryption = false
				} else {
					fmt.Println("Will attempt to unmask anyway, but this will likely fail.")
				}
			}
		} else if isMasked {
			fmt.Println("Detected masked .env file (values are encrypted but variable names visible)")
			
			if !useEncryption && !useMaskedEncryption {
				fmt.Print("The .env file has masked values. Unmask them? (y/n): ")
				var response string
				fmt.Scanln(&response)
				if response == "y" || response == "Y" {
					useMaskedEncryption = true
				} else {
					fmt.Println("Warning: Writing masked content to .env file without unmasking.")
				}
			} else if useEncryption {
				fmt.Println("Warning: Content has masked values but --decrypt flag was specified.")
				fmt.Print("Switch to unmasking instead? (y/n): ")
				var response string
				fmt.Scanln(&response)
				if response == "y" || response == "Y" {
					useEncryption = false
					useMaskedEncryption = true
				} else {
					fmt.Println("Will attempt to decrypt anyway, but this will likely fail.")
				}
			}
		} else if useEncryption || useMaskedEncryption {
			fmt.Println("Warning: --decrypt or --unmask flag specified but the content doesn't appear to be encrypted.")
			fmt.Print("Continue anyway? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Pull canceled")
				return
			}
			useEncryption = false
			useMaskedEncryption = false
		}
		
		// Process content based on encryption type
		if isEncrypted && useEncryption {
			fmt.Println("Decrypting .env file...")
			decryptedContent, err := DecryptContent([]byte(envContent))
			if err != nil {
				fmt.Printf("Error decrypting .env file: %s\n", err)
				os.Exit(1)
			}
			envContent = string(decryptedContent)
			fmt.Println("Decryption successful.")
		} else if isMasked && useMaskedEncryption {
			fmt.Println("Unmasking values in .env file...")
			unmaskedContent, err := UnmaskEnvContent([]byte(envContent))
			if err != nil {
				fmt.Printf("Error unmasking .env file: %s\n", err)
				os.Exit(1)
			}
			envContent = string(unmaskedContent)
			fmt.Println("Unmasking successful.")
		}

		// Write content to .env file
		err = os.WriteFile(".env", []byte(envContent), 0600)
		if err != nil {
			fmt.Printf("Error writing .env file: %s\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully pulled .env file from Gist")
		
		// Save Gist ID for future use if it's not already saved
		config, err = LoadConfig()
		if err == nil && config.LastGistID != pullGistID {
			config.LastGistID = pullGistID
			if saveErr := SaveConfig(config); saveErr == nil {
				fmt.Println("Saved Gist ID for future use")
			}
		}
	},
} 