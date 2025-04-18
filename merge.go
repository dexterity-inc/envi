package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	mergeGistID    string
	preferLocalFlag bool
	preferRemoteFlag bool
	mergeBackupFlag bool
	mergePushFlag bool
)

func init() {
	mergeCmd.Flags().StringVarP(&mergeGistID, "id", "i", "", "GitHub Gist ID to merge with (if not specified, uses the saved ID)")
	mergeCmd.Flags().BoolVarP(&preferLocalFlag, "local", "l", false, "Prefer local values when there are conflicts")
	mergeCmd.Flags().BoolVarP(&preferRemoteFlag, "remote", "r", false, "Prefer remote values when there are conflicts")
	mergeCmd.Flags().BoolVarP(&mergeBackupFlag, "backup", "b", true, "Create a backup of the local .env file before merging (default: true)")
	mergeCmd.Flags().BoolVarP(&mergePushFlag, "push", "p", false, "Push the merged result back to the Gist")
	rootCmd.AddCommand(mergeCmd)
}

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge local .env with a Gist",
	Long:  `Merge the local .env file with a remote .env file stored in a GitHub Gist.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Validate flags
		if preferLocalFlag && preferRemoteFlag {
			fmt.Println("Error: Cannot specify both --local and --remote flags")
			os.Exit(1)
		}

		// Check if .env exists locally
		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			fmt.Println("Error: Local .env file not found")
			os.Exit(1)
		}

		// If no Gist ID provided, check config for last used ID
		if mergeGistID == "" {
			config, err := LoadConfig()
			if err != nil {
				fmt.Printf("Error loading config: %s\n", err)
				os.Exit(1)
			}
			
			if config.LastGistID == "" {
				fmt.Println("Error: No Gist ID specified and no saved Gist ID found")
				fmt.Println("Use 'envi merge --id GIST_ID' or first push an .env file with 'envi push'")
				os.Exit(1)
			}
			
			// Ask user if they want to use the saved gist ID
			fmt.Printf("Found saved Gist ID: %s\n", config.LastGistID)
			fmt.Print("Do you want to merge with this Gist? (y/n): ")
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Merge canceled")
				return
			}
			
			mergeGistID = config.LastGistID
		}

		// Create backup if requested
		if mergeBackupFlag {
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
		}

		// Parse local .env file
		localVars, comments, err := parseEnvFileWithComments(".env")
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
		gist, _, err := client.Gists.Get(ctx, mergeGistID)
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
		remoteVars, remoteComments, err := parseEnvFileWithComments(tempFile)
		if err != nil {
			fmt.Printf("Error parsing remote .env file: %s\n", err)
			os.Exit(1)
		}

		// Identify and merge variables
		mergedVars := make(map[string]string)
		var onlyLocal, onlyRemote, conflicts []string
		
		// First, add all remote variables
		for key, value := range remoteVars {
			mergedVars[key] = value
		}
		
		// Then, process local variables
		for key, localValue := range localVars {
			if remoteValue, exists := remoteVars[key]; exists {
				// Variable exists in both, check for conflicts
				if localValue != remoteValue {
					conflicts = append(conflicts, key)
					
					// Resolve conflict based on flags
					if preferLocalFlag {
						mergedVars[key] = localValue
					} else if preferRemoteFlag {
						mergedVars[key] = remoteValue
					} else {
						// Interactive resolution
						fmt.Printf("\nConflict for variable '%s':\n", key)
						fmt.Printf("  1. Local:  %s\n", localValue)
						fmt.Printf("  2. Remote: %s\n", remoteValue)
						fmt.Print("Choose value (1/2): ")
						
						var choice string
						fmt.Scanln(&choice)
						
						if choice == "1" {
							mergedVars[key] = localValue
						} else if choice == "2" {
							mergedVars[key] = remoteValue
						} else {
							fmt.Println("Invalid choice, using local value by default")
							mergedVars[key] = localValue
						}
					}
				}
				// Otherwise, already have the remote value
			} else {
				// Variable only in local
				onlyLocal = append(onlyLocal, key)
				mergedVars[key] = localValue
			}
		}
		
		// Find variables only in remote
		for key := range remoteVars {
			if _, exists := localVars[key]; !exists {
				onlyRemote = append(onlyRemote, key)
			}
		}
		
		// Merge comments
		allComments := make(map[string][]string)
		for line, comments := range remoteComments {
			allComments[line] = comments
		}
		for line, comments := range comments {
			// If this line already has comments from remote, append
			if existingComments, exists := allComments[line]; exists {
				allComments[line] = append(existingComments, comments...)
			} else {
				allComments[line] = comments
			}
		}
		
		// Sort keys for ordered output
		var sortedKeys []string
		for key := range mergedVars {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)
		
		// Generate the merged file content
		var mergedContent strings.Builder
		
		// Add a header
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		mergedContent.WriteString(fmt.Sprintf("# Merged by envi on %s\n", timestamp))
		mergedContent.WriteString(fmt.Sprintf("# Merged local .env with Gist ID: %s\n\n", mergeGistID))
		
		// Add variables with their comments
		for _, key := range sortedKeys {
			// Add any comments for this key
			if keyComments, exists := allComments[key]; exists {
				for _, comment := range keyComments {
					mergedContent.WriteString(comment + "\n")
				}
			}
			
			mergedContent.WriteString(fmt.Sprintf("%s=%s\n", key, mergedVars[key]))
			
			// Add newline after each variable group
			mergedContent.WriteString("\n")
		}
		
		// Write the merged content to .env
		err = os.WriteFile(".env", []byte(mergedContent.String()), 0600)
		if err != nil {
			fmt.Printf("Error writing merged .env file: %s\n", err)
			os.Exit(1)
		}
		
		// Display merge summary
		fmt.Println("\n=== Merge Summary ===")
		fmt.Printf("Variables from remote: %d\n", len(remoteVars))
		fmt.Printf("Variables from local: %d\n", len(localVars))
		fmt.Printf("Variables in merged result: %d\n", len(mergedVars))
		fmt.Printf("  - Only in local: %d\n", len(onlyLocal))
		fmt.Printf("  - Only in remote: %d\n", len(onlyRemote))
		fmt.Printf("  - Conflicts resolved: %d\n", len(conflicts))
		
		fmt.Println("\n✅ Successfully merged .env file")
		
		// Push the merged result if requested
		if mergePushFlag {
			fmt.Println("\nPushing merged result back to Gist...")
			
			// Update the existing gist
			existingGist, _, err := client.Gists.Get(ctx, mergeGistID)
			if err != nil {
				fmt.Printf("Error fetching Gist for update: %s\n", err)
				os.Exit(1)
			}
			
			filename := ".env"
			mergedContent := mergedContent.String()
			
			existingGist.Files = map[github.GistFilename]github.GistFile{
				github.GistFilename(filename): {
					Content: github.String(mergedContent),
				},
			}
			
			// Update description to indicate merge
			description := "Environment variables - Merged with envi CLI"
			existingGist.Description = &description
			
			_, _, err = client.Gists.Edit(ctx, mergeGistID, existingGist)
			if err != nil {
				fmt.Printf("Error updating Gist with merged content: %s\n", err)
				os.Exit(1)
			}
			
			fmt.Println("✅ Successfully pushed merged result to Gist")
		}
	},
}

// parseEnvFileWithComments reads an env file and returns variables and comments
func parseEnvFileWithComments(filePath string) (map[string]string, map[string][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	vars := make(map[string]string)
	comments := make(map[string][]string)
	
	var currentComments []string
	var lastKey string
	
	scanner := bufio.NewScanner(file)
	lineRegex := `^\s*([\w\.]+)\s*=\s*(.*)?\s*$`
	
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		
		// Handle comments
		if line == "" {
			// Empty line, reset current comments unless we just processed a key
			if lastKey != "" {
				lastKey = ""
			} else {
				currentComments = nil
			}
			continue
		} else if strings.HasPrefix(line, "#") {
			// Store comment
			currentComments = append(currentComments, line)
			continue
		}
		
		// Extract key-value pairs
		matches := regexp.MustCompile(lineRegex).FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			value := matches[2]
			vars[key] = value
			
			// Associate comments with this key
			if len(currentComments) > 0 {
				comments[key] = currentComments
				currentComments = nil
			}
			
			lastKey = key
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}
	
	return vars, comments, nil
} 