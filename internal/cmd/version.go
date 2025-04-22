package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/dexterity-inc/envi/internal/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long:  `Display the version number and build information for Envi CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		displayVersion()
	},
}

func displayVersion() {
	fmt.Printf("Envi CLI v%s\n", version.GetVersion())
	
	// Only show commit and build date if not using the dev version
	if version.GetVersion() != "dev" {
		fmt.Printf("- Commit: %s\n", version.GetCommit())
		fmt.Printf("- Build Date: %s\n", version.GetBuildDate())
	}
	
	fmt.Printf("- Go version: %s\n", runtime.Version())
	fmt.Printf("- OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}

// InitVersionCommand initializes the version command
func InitVersionCommand() {
	rootCmd.AddCommand(versionCmd)
} 