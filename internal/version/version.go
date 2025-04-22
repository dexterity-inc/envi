package version

// Version is the current version of the application
// This value is set by the build process using ldflags
var Version = "dev"

// Commit is the git commit SHA used to build the application
// This value is set by the build process using ldflags
var Commit = "unknown"

// BuildDate is the date when the application was built
// This value is set by the build process using ldflags
var BuildDate = "unknown"

// GetVersion returns the current version
func GetVersion() string {
	return Version
}

// GetCommit returns the git commit SHA
func GetCommit() string {
	return Commit
}

// GetBuildDate returns the build date
func GetBuildDate() string {
	return BuildDate
} 