package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dexterity-inc/envi/internal/utils"
)

// Validate command flags
var (
	validateFix      bool
	validateStrict   bool
	validateRequired []string
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
		utils.Error("%s file not found", exampleFile)
		utils.Info("An example environment file is required for validation")
		utils.Fatal("Missing example file")
	}

	// Check if .env file exists
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		utils.Error("%s file not found", envFile)
		utils.Info("Create a .env file first or copy from .env.example")
		utils.Fatal("Missing .env file")
	}

	// Parse the current .env file
	currentVars, currentComments, err := parseEnvFile(envFile)
	if err != nil {
		utils.Error("Error reading %s: %s", envFile, err)
		utils.Fatal("Failed to read .env file")
	}

	// Parse the reference .env.example file
	referenceVars, _, err := parseEnvFile(exampleFile)
	if err != nil {
		utils.Error("Error reading %s: %s", exampleFile, err)
		utils.Fatal("Failed to read .env.example file")
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
		utils.Success("Validation successful: .env contains all variables from .env.example")
		utils.Info("Found %d environment variables", len(currentVars))
		checkStrictAndRequired(currentVars)
		return
	}

	// Report missing variables
	if len(missingVars) > 0 {
		utils.Error("Found %d missing variables in .env:", len(missingVars))
		for key, value := range missingVars {
			utils.Info("  %s=%s", key, value)
		}

		// Fix missing variables if requested
		if validateFix {
			err := addMissingVars(envFile, missingVars, currentVars, currentComments)
			if err != nil {
				utils.Error("Error fixing .env file: %s", err)
				utils.Fatal("Failed to fix .env file")
			}
			utils.Success("Added %d missing variables to .env", len(missingVars))

			// Recalculate current vars
			currentVars, _, _ = parseEnvFile(envFile)
		} else {
			utils.Info("Run 'envi validate --fix' to add these missing variables to your .env file")
		}
	}

	// Report extra variables
	if len(extraVars) > 0 {
		utils.Warn("Found %d extra variables in .env that are not in .env.example:", len(extraVars))
		for _, key := range extraVars {
			utils.Info("  %s=%s", key, currentVars[key])
		}
		utils.Info("You may want to add these to .env.example if they are needed")
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
					utils.Error("Strict validation errors:")
					hasStrictErrors = true
				}
				utils.Info("  Empty value for variable: %s", key)
			}
		}
		if !hasStrictErrors {
			utils.Success("All variables have values (strict validation passed)")
		}
	}

	// Check for required variables
	hasMissingRequired := false
	if len(validateRequired) > 0 {
		for _, requiredVar := range validateRequired {
			if _, found := vars[requiredVar]; !found {
				if !hasMissingRequired {
					utils.Error("Missing required variables:")
					hasMissingRequired = true
				}
				utils.Info("  %s", requiredVar)
			}
		}
		if !hasMissingRequired {
			utils.Success("All required variables are present")
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

			// Remove quotes if present
			if len(varValue) >= 2 && (varValue[0] == '"' && varValue[len(varValue)-1] == '"' ||
				varValue[0] == '\'' && varValue[len(varValue)-1] == '\'') {
				varValue = varValue[1 : len(varValue)-1]
			}

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
	// Read the current file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Create new content with missing variables
	var newContent strings.Builder

	// Write existing content
	newContent.Write(content)

	// Add a separator if the file doesn't end with a newline
	if len(content) > 0 && content[len(content)-1] != '\n' {
		newContent.WriteString("\n")
	}

	// Add missing variables
	if len(missingVars) > 0 {
		newContent.WriteString("\n# Added by envi validate --fix\n")
		for key, value := range missingVars {
			newContent.WriteString(fmt.Sprintf("%s=%s\n", key, value))
		}
	}

	// Write back to file
	return os.WriteFile(filename, []byte(newContent.String()), 0600)
}
