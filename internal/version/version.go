package version

// Version is the current version of the application
// This value is set by the build process using ldflags
var Version = "dev"

// GetVersion returns the current version
func GetVersion() string {
	return Version
} 