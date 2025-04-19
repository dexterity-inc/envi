package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// Validate command flags
var (
	validateFix         bool
	validateStrict      bool
	validateRequired    []string
)

// validateCmd is the validation command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate .env file against .env.example",
	Long:  `Compare your project's .env file with .env.example to identify missing variables.`,
	Run:   runValidateCommand,
}

// InitValidateCommand sets up the validate command and its subcommands
func InitValidateCommand() {
	// Initialize the command flags
	validateCmd.Flags().BoolVar(&validateFix, "fix", false, "Fix missing variables by adding them to .env file")
	validateCmd.Flags().BoolVarP(&validateStrict, "strict", "s", false, "Use strict validation (no empty values)")
	validateCmd.Flags().StringSliceVar(&validateRequired, "required", []string{}, "Required variables (comma-separated)")

	// Add the validate command to the root command
	rootCmd.AddCommand(validateCmd)
}

// runValidateCommand handles the validate command execution
func runValidateCommand(cmd *cobra.Command, args []string) {
	envFile := ".env"
	exampleFile := ".env.example"

	// Check if .env.example file exists
	if _, err := os.Stat(exampleFile); os.IsNotExist(err) {
		fmt.Printf("Error: %s file not found\n", exampleFile)
		fmt.Println("An example environment file is required for validation")
		os.Exit(1)
	}

	// Check if .env file exists
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		fmt.Printf("Error: %s file not found\n", envFile)
		fmt.Println("Create a .env file first or copy from .env.example")
		os.Exit(1)
	}

	// Parse the current .env file
	currentVars, currentComments, err := parseEnvFile(envFile)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", envFile, err)
		os.Exit(1)
	}

	// Parse the reference .env.example file
	referenceVars, _, err := parseEnvFile(exampleFile)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", exampleFile, err)
		os.Exit(1)
	}

	// Find missing variables
	missingVars := make(map[string]string)
	for key, value := range referenceVars {
		if _, exists := currentVars[key]; !exists {
			missingVars[key] = value
		}
	}

	// Check for extra variables in .env that aren't in .env.example
	extraVars := make([]string, 0)
	for key := range currentVars {
		if _, exists := referenceVars[key]; !exists {
			extraVars = append(extraVars, key)
		}
	}

	// Report results
	if len(missingVars) == 0 && len(extraVars) == 0 {
		fmt.Println("✅ Validation successful: .env contains all variables from .env.example")
		fmt.Printf("Found %d environment variables\n", len(currentVars))
		checkStrictAndRequired(currentVars)
		return
	}

	// Report missing variables
	if len(missingVars) > 0 {
		fmt.Printf("❌ Found %d missing variables in .env:\n", len(missingVars))
		for key, value := range missingVars {
			fmt.Printf("  %s=%s\n", key, value)
		}

		// Fix missing variables if requested
		if validateFix {
			err := addMissingVars(envFile, missingVars, currentVars, currentComments)
			if err != nil {
				fmt.Printf("Error fixing .env file: %s\n", err)
				os.Exit(1)
			}
			fmt.Printf("✅ Added %d missing variables to .env\n", len(missingVars))
			
			// Recalculate current vars
			currentVars, _, _ = parseEnvFile(envFile)
		} else {
			fmt.Println("Run 'envi validate --fix' to add these missing variables to your .env file")
		}
	}

	// Report extra variables
	if len(extraVars) > 0 {
		fmt.Printf("⚠️  Found %d extra variables in .env that are not in .env.example:\n", len(extraVars))
		for _, key := range extraVars {
			fmt.Printf("  %s=%s\n", key, currentVars[key])
		}
		fmt.Println("You may want to add these to .env.example if they are needed")
	}

	// Check strict validation and required variables
	checkStrictAndRequired(currentVars)
}

// checkStrictAndRequired validates strict mode and required variables
func checkStrictAndRequired(vars map[string]string) {
	// Check for strict validation errors (empty values)
	hasStrictErrors := false
	if validateStrict {
		for key, value := range vars {
			if value == "" {
				if !hasStrictErrors {
					fmt.Println("\n❌ Strict validation errors:")
					hasStrictErrors = true
				}
				fmt.Printf("  Empty value for variable: %s\n", key)
			}
		}
		if !hasStrictErrors {
			fmt.Println("✅ All variables have values (strict validation passed)")
		}
	}

	// Check for required variables
	hasMissingRequired := false
	if len(validateRequired) > 0 {
		for _, requiredVar := range validateRequired {
			if _, found := vars[requiredVar]; !found {
				if !hasMissingRequired {
					fmt.Println("\n❌ Missing required variables:")
					hasMissingRequired = true
				}
				fmt.Printf("  %s\n", requiredVar)
			}
		}
		if !hasMissingRequired {
			fmt.Println("✅ All required variables are present")
		}
	}
}

// parseEnvFile reads an .env file and returns a map of variables and a slice of comments
func parseEnvFile(filename string) (map[string]string, []string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	variables := make(map[string]string)
	comments := []string{}
	envVarRegex := regexp.MustCompile(`^([A-Za-z0-9_]+)=(.*)$`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Skip empty lines
		if trimmedLine == "" {
			continue
		}

		// Handle comments
		if strings.HasPrefix(trimmedLine, "#") {
			comments = append(comments, line)
			continue
		}

		// Handle environment variables
		if envVarRegex.MatchString(line) {
			matches := envVarRegex.FindStringSubmatch(line)
			varName := matches[1]
			varValue := matches[2]
			variables[varName] = varValue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return variables, comments, nil
}

// addMissingVars adds missing variables to the .env file
func addMissingVars(filename string, missingVars, currentVars map[string]string, comments []string) error {
	// Create a backup of the original file
	backupFile := filename + ".bak"
	err := copyFile(filename, backupFile)
	if err != nil {
		return fmt.Errorf("could not create backup: %w", err)
	}
	fmt.Printf("Created backup at %s\n", backupFile)

	// Open file for writing
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// First write all existing content
	existingFile, err := os.Open(backupFile)
	if err != nil {
		return err
	}
	defer existingFile.Close()

	scanner := bufio.NewScanner(existingFile)
	for scanner.Scan() {
		fmt.Fprintln(writer, scanner.Text())
	}
	
	// Add a separator for new variables
	if len(missingVars) > 0 {
		fmt.Fprintln(writer, "")
		fmt.Fprintln(writer, "# Added missing variables from .env.example")
		
		// Add missing variables
		for key, value := range missingVars {
			fmt.Fprintf(writer, "%s=%s\n", key, value)
		}
	}

	return writer.Flush()
} 