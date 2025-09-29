package version

import "testing"

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("GetVersion() should not return empty string")
	}
}

func TestGetCommit(t *testing.T) {
	commit := GetCommit()
	if commit == "" {
		t.Error("GetCommit() should not return empty string")
	}
}

func TestGetBuildDate(t *testing.T) {
	buildDate := GetBuildDate()
	if buildDate == "" {
		t.Error("GetBuildDate() should not return empty string")
	}
}

func TestVersionConstants(t *testing.T) {
	if Version == "" {
		t.Error("Version constant should not be empty")
	}
	
	if Commit == "" {
		t.Error("Commit constant should not be empty")
	}
	
	if BuildDate == "" {
		t.Error("BuildDate constant should not be empty")
	}
}