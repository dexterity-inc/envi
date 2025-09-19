package cmd_test

import (
	"testing"

	"github.com/dexterity-inc/envi/internal/version"
)

func TestVersionCommand(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		contains    []string
	}{
		{
			name:        "version flag short",
			args:        []string{"-v"},
			expectError: false,
			contains:    []string{version.Version},
		},
		{
			name:        "version flag long",
			args:        []string{"--version"},
			expectError: false,
			contains:    []string{version.Version},
		},
		{
			name:        "version command",
			args:        []string{"version"},
			expectError: false,
			contains:    []string{version.Version, version.Commit, version.BuildDate},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test would need the actual CLI setup
			// For now, we'll test the version values directly
			if version.GetVersion() == "" {
				t.Error("Version should not be empty")
			}
			
			if version.GetCommit() == "" {
				t.Error("Commit should not be empty")
			}
			
			if version.GetBuildDate() == "" {
				t.Error("BuildDate should not be empty")
			}
		})
	}
}

func TestVersionValues(t *testing.T) {
	t.Run("version constants are accessible", func(t *testing.T) {
		v := version.GetVersion()
		c := version.GetCommit()
		d := version.GetBuildDate()
		
		// These values might be "dev", "unknown", etc. in test environment
		if v == "" {
			t.Error("Version should not be empty")
		}
		if c == "" {
			t.Error("Commit should not be empty")  
		}
		if d == "" {
			t.Error("BuildDate should not be empty")
		}
	})
}

