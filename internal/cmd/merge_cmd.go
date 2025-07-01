package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/dexterity-inc/envi/internal/config"
	"github.com/dexterity-inc/envi/internal/encryption"
	"github.com/dexterity-inc/envi/internal/utils"
)

// Merge command flags
var (
	mergeFiles          []string
	mergeOutput         string
	mergeGistID         string
	mergeSkipDuplicates bool
	mergeOverwrite      bool
	mergeKeepComments   bool
	mergeSort           bool
	mergeCreateBackup   bool
	mergeUnmask         bool
)

// mergeCmd is the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Merge multiple .env files",
	Long:  `Merge multiple .env files or merge with a remote Gist .env file.`,
	Run:   runMergeCommand,
}

// InitMergeCommand sets up the merge command and its subcommands
func InitMergeCommand() {
	// Initialize the command flags
	mergeCmd.Flags().StringSliceVarP(&mergeFiles, "files", "f", []string{}, "Paths to local .env files to merge (comma-separated)")
	mergeCmd.Flags().StringVarP(&mergeGistID, "gist", "g", "", "GitHub Gist ID to merge with (will fetch remote .env)")
	mergeCmd.Flags().StringVarP(&mergeOutput, "output", "o", ".env", "Output file path")
	mergeCmd.Flags().BoolVarP(&mergeSkipDuplicates, "skip-duplicates", "s", false, "Skip duplicates (local file takes precedence)")
	mergeCmd.Flags().BoolVarP(&mergeOverwrite, "overwrite", "w", false, "Overwrite duplicates (remote file takes precedence)")
	mergeCmd.Flags().BoolVarP(&mergeKeepComments, "keep-comments", "c", true, "Keep comments from all files")
	mergeCmd.Flags().BoolVar(&mergeSort, "sort", false, "Sort variables alphabetically")
	mergeCmd.Flags().BoolVar(&mergeCreateBackup, "backup", true, "Create backup of output file if it exists")
	mergeCmd.Flags().BoolVar(&mergeUnmask, "unmask", false, "Unmask/decrypt values from remote Gist when merging")

	// Add the merge command to the root command
	rootCmd.AddCommand(mergeCmd)
}

// runMergeCommand handles the merge command execution
func runMergeCommand(cmd *cobra.Command, args []string) {
	// Check if we're merging with a Gist or local files
	if mergeGistID == "" && len(mergeFiles) == 0 {
		utils.Error("You must specify either local files to merge (--files) or a Gist ID to merge with (--gist)")
		utils.Info("Run 'envi merge --help' for usage information")
		utils.Fatal("Missing input files")
	}

	// Create backup if output file exists
	if _, err := os.Stat(mergeOutput); err == nil && mergeCreateBackup {
		backupFile := fmt.Sprintf("%s.bak.%s", mergeOutput, time.Now().Format("20060102150405"))
		err := copyFile(mergeOutput, backupFile)
		if err != nil {
			utils.Warn("Could not create backup file: %s", err)
		} else {
			utils.Info("Created backup of existing file at %s", backupFile)
		}
	}

	// Variables to store merged content
	variables := make(map[string]string)
	comments := []string{}
	variableOrder := []string{} // To preserve order if not sorting
	filesToProcess := mergeFiles

	// If merging with a Gist, fetch the remote .env file
	var remoteContent []byte
	if mergeGistID != "" {
		utils.Info("Fetching Gist with ID: %s", mergeGistID)

		// Get GitHub token
		token, err := config.GetGitHubToken()
		if err != nil {
			utils.Error("Error: %s", err)
			utils.Fatal("Failed to get GitHub token")
		}

		// Create GitHub client
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(cmd.Context(), ts)
		client := github.NewClient(tc)

		// Get Gist
		gist, _, err := client.Gists.Get(cmd.Context(), mergeGistID)
		if err != nil {
			utils.Error("Error retrieving Gist with ID %s: %s", mergeGistID, err)
			utils.Fatal("Failed to retrieve Gist")
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
			utils.Error("No .env file found in this Gist")
			utils.Fatal("Missing .env file in Gist")
		}

		// Get content
		remoteContent = []byte(*envFile.Content)

		// Check if content is encrypted and needs decryption
		isEncrypted := encryption.IsEncrypted(remoteContent)
		isMasked := encryption.IsMasked(remoteContent)

		if (isEncrypted || isMasked) && mergeUnmask {
			utils.Info("Detected encrypted content. Attempting to decrypt...")

			var decryptedContent []byte
			var err error

			if isEncrypted {
				decryptedContent, err = encryption.DecryptContent(remoteContent)
			} else if isMasked {
				decryptedContent, err = encryption.UnmaskEnvContent(remoteContent)
			}

			if err != nil {
				utils.Error("Error decrypting content. Please check your encryption settings and try again.")
				utils.Fatal("Decryption failed")
			}

			remoteContent = decryptedContent
			utils.Success("Successfully decrypted remote content!")
		} else if (isEncrypted || isMasked) && !mergeUnmask {
			utils.Warn("Remote content is encrypted/masked but --unmask flag not specified.")
			utils.Info("Merging encrypted content - this may not be what you want.")
		}

		// Save to a temporary file
		tempFile := ".env.remote.tmp"
		if err := os.WriteFile(tempFile, remoteContent, 0600); err != nil {
			utils.Error("Error writing temporary file: %s", err)
			utils.Fatal("Failed to write temporary file")
		}
		defer os.Remove(tempFile) // Clean up temporary file

		// Add to files to process
		filesToProcess = append(filesToProcess, tempFile)
		utils.Info("Remote .env file added to merge")
	}

	// Verify all local files exist
	for _, file := range filesToProcess {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			utils.Error(".env file not found at %s", file)
			utils.Fatal("File not found")
		}
	}

	// Process each file
	for _, file := range filesToProcess {
		utils.Info("Processing file: %s", file)

		// Open file
		f, err := os.Open(file)
		if err != nil {
			utils.Error("Error opening file %s: %s", file, err)
			utils.Fatal("Failed to open file")
		}

		// Read file line by line
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			trimmedLine := strings.TrimSpace(line)

			// Handle empty lines
			if trimmedLine == "" {
				continue
			}

			// Handle comments
			if strings.HasPrefix(trimmedLine, "#") {
				if mergeKeepComments {
					comments = append(comments, line)
				}
				continue
			}

			// Handle environment variables
			if strings.Contains(trimmedLine, "=") {
				parts := strings.SplitN(trimmedLine, "=", 2)
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Handle duplicate keys based on flags
				if existingValue, exists := variables[key]; exists {
					if mergeSkipDuplicates {
						utils.Info("Skipping duplicate key: %s (keeping existing value)", key)
						continue
					} else if mergeOverwrite {
						utils.Info("Overwriting duplicate key: %s (new value: %s)", key, value)
						variables[key] = value
					} else {
						utils.Warn("Duplicate key found: %s", key)
						utils.Info("  Existing: %s", existingValue)
						utils.Info("  New: %s", value)
						utils.Info("  Using new value (use --skip-duplicates to keep existing)")
						variables[key] = value
					}
				} else {
					variables[key] = value
					variableOrder = append(variableOrder, key)
				}
			}
		}

		if err := scanner.Err(); err != nil {
			utils.Error("Error reading file %s: %s", file, err)
			utils.Fatal("Failed to read file")
		}

		f.Close()
	}

	// Write merged content to output file
	output, err := os.Create(mergeOutput)
	if err != nil {
		utils.Error("Error creating output file: %s", err)
		utils.Fatal("Failed to create output file")
	}
	defer output.Close()

	// Write comments first
	for _, comment := range comments {
		output.WriteString(comment + "\n")
	}

	// Write variables
	var keysToWrite []string
	if mergeSort {
		keysToWrite = sortKeys(variables)
	} else {
		keysToWrite = variableOrder
	}

	for _, key := range keysToWrite {
		output.WriteString(fmt.Sprintf("%s=%s\n", key, variables[key]))
	}

	utils.Success("Successfully merged %d files into %s", len(filesToProcess), mergeOutput)
	utils.Info("Total variables: %d", len(variables))
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0600)
}

// sortKeys returns a sorted slice of map keys
func sortKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	// Use Go's efficient built-in sort algorithm instead of bubble sort
	sort.Strings(keys)

	return keys
}
