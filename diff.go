package main

import (
	"context"
	"fmt"
	"os"
	"sort"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	diffGistID string
	showValuesFlag bool
)

func init() {
	diffCmd.Flags().StringVarP(&diffGistID, "id", "i", "", "GitHub Gist ID to compare with (if not specified, uses the saved ID)")
	diffCmd.Flags().BoolVarP(&showValuesFlag, "values", "v", false, "Show values in the diff output (may expose sensitive data)")
	rootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare local .env with a Gist",
	Long:  `Compare the local .env file with a remote .env file stored in a GitHub Gist.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if .env exists locally
		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			fmt.Println("Error: Local .env file not found")
			os.Exit(1)
		}

		// If no Gist ID provided, check config for last used ID
		if diffGistID == "" {
			config, err := LoadConfig()
			if err != nil {
				fmt.Printf("Error loading config: %s\n", err)
				os.Exit(1)
			}
			
			if config.LastGistID == "" {
				fmt.Println("Error: No Gist ID specified and no saved Gist ID found")
				fmt.Println("Use 'envi diff --id GIST_ID' or first push an .env file with 'envi push'")
				os.Exit(1)
			}
			
			// Ask user if they want to use the saved gist ID
			fmt.Printf("Found saved Gist ID: %s\n", config.LastGistID)
			fmt.Print("Do you want to compare with this Gist? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Comparison canceled")
				return
			}
			
			diffGistID = config.LastGistID
		}

		// Parse local .env file
		localVars, err := parseEnvFile(".env")
		if err != nil {
			fmt.Printf("Error parsing local .env file: %s\n", err)
			os.Exit(1)
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
		gist, _, err := client.Gists.Get(ctx, diffGistID)
		if err != nil {
			fmt.Printf("Error fetching Gist: %s\n", err)
			os.Exit(1)
		}

		// Extract .env file content
		var remoteEnvContent string
		found := false
		for filename, file := range gist.Files {
			if filename == ".env" {
				remoteEnvContent = *file.Content
				found = true
				break
			}
		}

		if !found {
			fmt.Println("Error: No .env file found in the specified Gist")
			os.Exit(1)
		}

		// Write remote content to temp file for parsing
		tempFile := ".env.remote.tmp"
		err = os.WriteFile(tempFile, []byte(remoteEnvContent), 0600)
		if err != nil {
			fmt.Printf("Error creating temp file: %s\n", err)
			os.Exit(1)
		}
		defer os.Remove(tempFile) // Clean up temp file when done

		// Parse remote .env file
		remoteVars, err := parseEnvFile(tempFile)
		if err != nil {
			fmt.Printf("Error parsing remote .env file: %s\n", err)
			os.Exit(1)
		}

		// Compare variables
		var onlyLocal, onlyRemote, different []string
		
		// Find variables only in local or with different values
		for key, localVal := range localVars {
			if remoteVal, exists := remoteVars[key]; exists {
				if localVal != remoteVal {
					different = append(different, key)
				}
			} else {
				onlyLocal = append(onlyLocal, key)
			}
		}
		
		// Find variables only in remote
		for key := range remoteVars {
			if _, exists := localVars[key]; !exists {
				onlyRemote = append(onlyRemote, key)
			}
		}
		
		// Sort the slices for consistent output
		sort.Strings(onlyLocal)
		sort.Strings(onlyRemote)
		sort.Strings(different)
		
		// Display results
		fmt.Println("\n=== Comparing local .env with Gist ===")
		
		if len(onlyLocal) == 0 && len(onlyRemote) == 0 && len(different) == 0 {
			fmt.Println("\n✅ Files are identical")
			return
		}
		
		// Variables only in local file
		if len(onlyLocal) > 0 {
			fmt.Printf("\n🟢 Variables only in local .env (%d):\n", len(onlyLocal))
			for _, key := range onlyLocal {
				if showValuesFlag {
					fmt.Printf("   %s=%s\n", key, localVars[key])
				} else {
					fmt.Printf("   %s\n", key)
				}
			}
		}
		
		// Variables only in remote file
		if len(onlyRemote) > 0 {
			fmt.Printf("\n🔴 Variables only in remote .env (%d):\n", len(onlyRemote))
			for _, key := range onlyRemote {
				if showValuesFlag {
					fmt.Printf("   %s=%s\n", key, remoteVars[key])
				} else {
					fmt.Printf("   %s\n", key)
				}
			}
		}
		
		// Variables with different values
		if len(different) > 0 {
			fmt.Printf("\n🟡 Variables with different values (%d):\n", len(different))
			for _, key := range different {
				if showValuesFlag {
					fmt.Printf("   %s:\n", key)
					fmt.Printf("     Local:  %s\n", localVars[key])
					fmt.Printf("     Remote: %s\n", remoteVars[key])
				} else {
					fmt.Printf("   %s\n", key)
				}
			}
			
			if !showValuesFlag {
				fmt.Println("\nTip: Use --values flag to see the actual values (may expose sensitive data)")
			}
		}
		
		// Show summary
		fmt.Printf("\nSummary: %d only in local, %d only in remote, %d different\n", 
			len(onlyLocal), len(onlyRemote), len(different))
	},
} 