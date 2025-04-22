package cmd

import (
	"fmt"
	"runtime"

	"github.com/dexterity-inc/envi/internal/version"
)

// displayVersion prints the version information
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

// InitVersionCommand initializes the version flags
func InitVersionCommand() {
	// We don't need a separate version command as Cobra already provides
	// --version flag. This function is kept for consistency in command initialization.
	
	// Add a custom -v short flag for version
	rootCmd.Flags().BoolP("version", "v", false, "Display version information")
	
	// Override the default version template to use our custom version display
	rootCmd.SetVersionTemplate(`{{.Name}} version {{.Version}}
`)
} 