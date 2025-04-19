package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"

	"github.com/dexterity-inc/envi/internal/config"
	"github.com/dexterity-inc/envi/internal/encryption"
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
		fmt.Println("Error: You must specify either local files to merge (--files) or a Gist ID to merge with (--gist)")
		fmt.Println("Run 'envi merge --help' for usage information")
		os.Exit(1)
	}

	// Create backup if output file exists
	if _, err := os.Stat(mergeOutput); err == nil && mergeCreateBackup {
		backupFile := fmt.Sprintf("%s.bak.%s", mergeOutput, time.Now().Format("20060102150405"))
		err := copyFile(mergeOutput, backupFile)
		if err != nil {
			fmt.Printf("Warning: Could not create backup file: %s\n", err)
		} else {
			fmt.Printf("Created backup of existing file at %s\n", backupFile)
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
		fmt.Printf("Fetching Gist with ID: %s\n", mergeGistID)
		
		// Get GitHub token
		token, err := config.GetGitHubToken()
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			os.Exit(1)
		}
		
		// Create GitHub client
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(cmd.Context(), ts)
		client := github.NewClient(tc)
		
		// Get Gist
		gist, _, err := client.Gists.Get(cmd.Context(), mergeGistID)
		if err != nil {
			fmt.Printf("Error retrieving Gist with ID %s: %s\n", mergeGistID, err)
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
		remoteContent = []byte(*envFile.Content)
		
		// Check if content is encrypted and needs decryption
		isEncrypted := encryption.IsEncrypted(remoteContent)
		isMasked := encryption.IsMasked(remoteContent)
		
		if (isEncrypted || isMasked) && mergeUnmask {
			fmt.Println("Detected encrypted content. Attempting to decrypt...")
			
			var decryptedContent []byte
			var err error
			
			if isEncrypted {
				decryptedContent, err = encryption.DecryptContent(remoteContent)
			} else if isMasked {
				decryptedContent, err = encryption.UnmaskEnvContent(remoteContent)
			}
			
			if err != nil {
				fmt.Printf("Error decrypting content: %s\n", err)
				os.Exit(1)
			}
			
			remoteContent = decryptedContent
			fmt.Println("Successfully decrypted remote content!")
		} else if (isEncrypted || isMasked) && !mergeUnmask {
			fmt.Println("Warning: Remote content is encrypted/masked but --unmask flag not specified.")
			fmt.Println("Merging encrypted content - this may not be what you want.")
		}
		
		// Save to a temporary file
		tempFile := ".env.remote.tmp"
		if err := os.WriteFile(tempFile, remoteContent, 0600); err != nil {
			fmt.Printf("Error writing temporary file: %s\n", err)
			os.Exit(1)
		}
		defer os.Remove(tempFile) // Clean up temporary file
		
		// Add to files to process
		filesToProcess = append(filesToProcess, tempFile)
		fmt.Println("Remote .env file added to merge")
	}

	// Verify all local files exist
	for _, file := range filesToProcess {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("Error: .env file not found at %s\n", file)
			os.Exit(1)
		}
	}

	// Process each file
	for _, file := range filesToProcess {
		fmt.Printf("Processing file: %s\n", file)
		
		// Open file
		f, err := os.Open(file)
		if err != nil {
			fmt.Printf("Error opening file %s: %s\n", file, err)
			os.Exit(1)
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
			
			// Handle environment variables (KEY=value)
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				value := parts[1]
				
				// Check for duplicates
				_, exists := variables[key]
				if exists {
					// Handling duplicates differently based on whether this is from Gist
					isRemoteFile := file == ".env.remote.tmp"
					
					if mergeOverwrite && isRemoteFile {
						// If we're overwriting and this is the remote file, it takes precedence
						fmt.Printf("Overwriting with remote value for variable: %s\n", key)
						variables[key] = value
					} else if mergeSkipDuplicates && !isRemoteFile {
						// If we're skipping duplicates and this is a local file, it takes precedence
						fmt.Printf("Keeping local value for duplicate variable: %s\n", key)
					} else if !mergeSkipDuplicates && !mergeOverwrite {
						fmt.Printf("Warning: Duplicate variable found: %s\n", key)
						fmt.Printf("  Local value: %s\n", variables[key])
						fmt.Printf("  Remote value: %s\n", value)
						fmt.Println("Use --overwrite to prefer remote values or --skip-duplicates to prefer local values")
					}
				} else {
					variables[key] = value
					variableOrder = append(variableOrder, key)
				}
			}
		}
		
		f.Close()
		
		// Check for scanner errors
		if err := scanner.Err(); err != nil {
			fmt.Printf("Error reading file %s: %s\n", file, err)
			os.Exit(1)
		}
	}

	// Create output file
	outFile, err := os.Create(mergeOutput)
	if err != nil {
		fmt.Printf("Error creating output file: %s\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	// Write merged content
	writer := bufio.NewWriter(outFile)
	
	// Add a header comment
	fmt.Fprintf(writer, "# .env file created by envi merge\n")
	fmt.Fprintf(writer, "# Created on %s\n", time.Now().Format("2006-01-02 15:04:05"))
	
	if mergeGistID != "" {
		fmt.Fprintf(writer, "# Merged local .env with remote Gist: %s\n", mergeGistID)
	} else {
		fmt.Fprintf(writer, "# Merged from %d files: %s\n", len(filesToProcess), strings.Join(filesToProcess, ", "))
	}
	fmt.Fprintln(writer, "")
	
	// Write comments if keeping them
	if mergeKeepComments && len(comments) > 0 {
		fmt.Fprintf(writer, "# Merged comments from source files:\n")
		for _, comment := range comments {
			fmt.Fprintln(writer, comment)
		}
		fmt.Fprintln(writer, "")
	}
	
	// Write variables
	if mergeSort {
		// Sort variables alphabetically
		sortedKeys := sortKeys(variables)
		for _, key := range sortedKeys {
			fmt.Fprintf(writer, "%s=%s\n", key, variables[key])
		}
	} else {
		// Use original order
		for _, key := range variableOrder {
			fmt.Fprintf(writer, "%s=%s\n", key, variables[key])
		}
	}
	
	writer.Flush()
	
	fmt.Printf("Successfully merged .env files into %s\n", mergeOutput)
	fmt.Printf("Merged %d variables\n", len(variables))
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	// Read source file
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	
	// Create destination file
	return os.WriteFile(dst, sourceData, 0600)
}

// sortKeys returns the keys of a map in alphabetical order
func sortKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	
	// Simple bubble sort for alphabetical ordering
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	
	return keys
} 