package cmd

import (
	"github.com/spf13/cobra"

	"github.com/dexterity-inc/envi/internal/encryption"
	"github.com/dexterity-inc/envi/internal/utils"
	"github.com/dexterity-inc/envi/internal/version"
)

// Root command definition
var rootCmd = &cobra.Command{
	Use:     "envi",
	Short:   "Manage environment variables with GitHub Gists",
	Long:    `Envi is a secure tool for storing and sharing .env files via GitHub Gists.`,
	Version: version.Version,

	// This will run before the main command execution
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger with TUI setting
		utils.InitLogger(encryption.UseTUI)

		// Check if the version flag was used
		if cmd.Flag("version") != nil && cmd.Flag("version").Changed {
			displayVersion()
			// Call os.Exit(0) to prevent the Run function from executing
			cobra.CheckErr(nil)
		}
	},

	Run: func(cmd *cobra.Command, args []string) {
		// Show help by default when no subcommand is provided
		cmd.Help()
	},
}

// Execute runs the root command and handles errors
func Execute() error {
	// Set up global flags
	rootCmd.PersistentFlags().BoolVar(&encryption.UseTUI, "tui", true, "Use interactive terminal UI")

	// Initialize all commands
	InitConfigCommand()
	InitPushCommand()
	InitPullCommand()
	InitListCommand()
	InitShareCommand()
	InitMergeCommand()
	InitDecryptCommand()
	InitValidateCommand()
	InitCompletionCommand()
	InitVersionCommand()
	InitGistCommand()

	// Initialize command flags
	encryption.InitEncryptionFlags(rootCmd)

	// Run the command
	return rootCmd.Execute()
}
