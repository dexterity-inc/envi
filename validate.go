package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// Command flags
	fixFlag bool
	examplePath string
)

func init() {
	validateCmd.Flags().BoolVarP(&fixFlag, "fix", "f", false, "Automatically add missing environment variables from .env.example to .env")
	validateCmd.Flags().StringVarP(&examplePath, "example", "e", ".env.example", "Path to the example env file")
	rootCmd.AddCommand(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate .env file against .env.example",
	Long:  `Compare your .env file with .env.example to find missing variables.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if example file exists
		if _, err := os.Stat(examplePath); os.IsNotExist(err) {
			fmt.Printf("Error: Example file %s not found\n", examplePath)
			os.Exit(1)
		}

		// Check if .env file exists
		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			fmt.Println("Error: .env file not found in current directory")
			os.Exit(1)
		}

		// Parse both files
		exampleVars, err := parseEnvFile(examplePath)
		if err != nil {
			fmt.Printf("Error parsing %s: %s\n", examplePath, err)
			os.Exit(1)
		}

		envVars, err := parseEnvFile(".env")
		if err != nil {
			fmt.Printf("Error parsing .env: %s\n", err)
			os.Exit(1)
		}

		// Find missing variables
		var missingVars []string
		var missingVarsWithValues []string

		for key, value := range exampleVars {
			if _, exists := envVars[key]; !exists {
				missingVars = append(missingVars, key)
				missingVarsWithValues = append(missingVarsWithValues, fmt.Sprintf("%s=%s", key, value))
			}
		}

		// Display results
		if len(missingVars) == 0 {
			fmt.Println("✅ Validation successful: All variables from .env.example exist in .env")
			return
		}

		fmt.Printf("❌ Validation failed: %d missing variable(s) in .env\n", len(missingVars))
		for _, key := range missingVars {
			fmt.Printf("  - %s\n", key)
		}

		// Fix if requested
		if fixFlag && len(missingVars) > 0 {
			err := addMissingVars(".env", missingVarsWithValues)
			if err != nil {
				fmt.Printf("Error fixing .env file: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Added %d missing variable(s) to .env\n", len(missingVars))
		} else if len(missingVars) > 0 {
			fmt.Println("\nRun with --fix flag to add missing variables to your .env file")
		}
	},
}

// parseEnvFile reads an env file and returns a map of key-value pairs
func parseEnvFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineRegex := regexp.MustCompile(`^\s*([\w\.]+)\s*=\s*(.*)?\s*$`)

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract key-value pairs
		matches := lineRegex.FindStringSubmatch(line)
		if len(matches) == 3 {
			key := matches[1]
			value := matches[2]
			vars[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return vars, nil
}

// addMissingVars appends missing variables to the .env file
func addMissingVars(envPath string, missingVars []string) error {
	// Read the file content first
	content, err := os.ReadFile(envPath)
	if err != nil {
		return err
	}
	
	// Check if content ends with newline
	endsWithNewline := len(content) > 0 && content[len(content)-1] == '\n'
	
	// Open file for writing (truncate existing content)
	file, err := os.OpenFile(envPath, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Write original content back
	if _, err := file.Write(content); err != nil {
		return err
	}
	
	// Add a newline if needed
	if !endsWithNewline {
		if _, err := file.WriteString("\n"); err != nil {
			return err
		}
	}
	
	// Add a comment to indicate added variables
	if _, err := file.WriteString("\n# Added automatically by envi validate\n"); err != nil {
		return err
	}
	
	// Append each missing variable
	for _, varLine := range missingVars {
		if _, err := file.WriteString(varLine + "\n"); err != nil {
			return err
		}
	}
	
	return nil
} 