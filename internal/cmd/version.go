package cmd

import (
	"runtime"

	"github.com/dexterity-inc/envi/internal/utils"
	"github.com/dexterity-inc/envi/internal/version"
)

// displayVersion prints the version information
func displayVersion() {
	utils.Info("Envi CLI v%s", version.GetVersion())

	// Only show commit and build date if not using the dev version
	if version.GetVersion() != "dev" {
		utils.Info("- Commit: %s", version.GetCommit())
		utils.Info("- Build Date: %s", version.GetBuildDate())
	}

	utils.Info("- Go version: %s", runtime.Version())
	utils.Info("- OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
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
