package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "envi",
	Short: "Environment variable management using GitHub Gists",
	Long:  `Envi is a CLI tool to push and pull .env files to/from GitHub Gists for secure storage and sharing.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If no subcommand is provided, print help
		cmd.Help()
	},
}

func Execute() {
	// First, ensure all commands are added to rootCmd from their respective init functions
	
	// Then, initialize all flags after all commands are registered
	initEncryptionFlags()  // Initialize encryption flags for push and pull commands
	initConfigFlags()      // Initialize config command flags
	
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	Execute()
} 