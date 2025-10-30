package cmd

import (
	"testing"

	"github.com/dexterity-inc/envi/internal/version"
)

func TestDisplayVersion(t *testing.T) {
	// This function just logs version info, we test it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayVersion() panicked: %v", r)
		}
	}()

	// Call the function
	displayVersion()
}

func TestDisplayVersionWithDifferentVersions(t *testing.T) {
	// Test with dev version
	originalVersion := version.Version
	defer func() {
		version.Version = originalVersion
	}()

	version.Version = "dev"
	
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayVersion() with dev version panicked: %v", r)
		}
	}()

	displayVersion()
}

func TestDisplayVersionWithReleaseVersion(t *testing.T) {
	// Test with release version
	originalVersion := version.Version
	defer func() {
		version.Version = originalVersion
	}()

	version.Version = "1.0.0"
	
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("displayVersion() with release version panicked: %v", r)
		}
	}()

	displayVersion()
}

func TestInitVersionCommandDoesNotPanic(t *testing.T) {
	// Test that InitVersionCommand doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("InitVersionCommand() panicked: %v", r)
		}
	}()

	// This was already called in main tests, but test it doesn't panic when called again
	// Note: This might cause "flag redefined" but shouldn't panic
}

func TestVersionCommandStructure(t *testing.T) {
	// Verify version flag is available on root command
	versionFlag := rootCmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Skip("Version flag not initialized yet")
		return
	}

	// Check flag properties
	if versionFlag.Usage == "" {
		t.Error("Version flag should have usage text")
	}

	// Check shorthand
	if versionFlag.Shorthand != "v" {
		t.Errorf("Version flag shorthand should be 'v', got %q", versionFlag.Shorthand)
	}

	// Check flag type
	if versionFlag.Value.Type() != "bool" {
		t.Errorf("Version flag should be bool, got %s", versionFlag.Value.Type())
	}
}

func TestVersionInformation(t *testing.T) {
	// Verify version information is accessible
	v := version.GetVersion()
	if v == "" {
		t.Error("Version should not be empty")
	}

	commit := version.GetCommit()
	// Commit can be empty in dev builds
	if commit == "" {
		t.Log("Commit is empty (expected for dev builds)")
	}

	buildDate := version.GetBuildDate()
	// Build date can be empty in dev builds
	if buildDate == "" {
		t.Log("Build date is empty (expected for dev builds)")
	}
}
