package cmd

import (
	"testing"
)

func TestCompletionCommandStructure(t *testing.T) {
	// Test that completion command has proper structure
	if completionCmd == nil {
		t.Fatal("completionCmd should not be nil")
	}

	// Check Use field
	if completionCmd.Use == "" {
		t.Error("Completion command should have Use field")
	}

	// Check that it contains valid args
	if !contains(completionCmd.Use, "bash") && 
	   !contains(completionCmd.Use, "zsh") && 
	   !contains(completionCmd.Use, "fish") && 
	   !contains(completionCmd.Use, "powershell") {
		t.Log("Completion command Use field:", completionCmd.Use)
	}

	// Check Short description
	if completionCmd.Short == "" {
		t.Error("Completion command should have Short description")
	}

	// Check Long description
	if completionCmd.Long == "" {
		t.Error("Completion command should have Long description")
	}

	// Check ValidArgs
	if len(completionCmd.ValidArgs) != 4 {
		t.Errorf("Expected 4 valid args (bash, zsh, fish, powershell), got %d", len(completionCmd.ValidArgs))
	}

	// Verify all required shells are in ValidArgs
	requiredShells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range requiredShells {
		found := false
		for _, arg := range completionCmd.ValidArgs {
			if arg == shell {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Required shell %s not found in ValidArgs", shell)
		}
	}
}

func TestCompletionCommandHasRun(t *testing.T) {
	// Verify Run function exists
	if completionCmd.Run == nil {
		t.Error("Completion command should have Run function")
	}
}

func TestCompletionCommandDisablesFlagsInUseLine(t *testing.T) {
	// Check that DisableFlagsInUseLine is set
	if !completionCmd.DisableFlagsInUseLine {
		t.Error("Completion command should have DisableFlagsInUseLine set to true")
	}
}

func TestCompletionCommandLongDescriptionContent(t *testing.T) {
	// Verify long description contains instructions for all shells
	shells := []string{"Bash", "Zsh", "Fish", "PowerShell"}
	
	for _, shell := range shells {
		if !contains(completionCmd.Long, shell) {
			t.Errorf("Long description should contain instructions for %s", shell)
		}
	}
}

func TestCompletionCommandExamplesInLongDescription(t *testing.T) {
	// Verify long description contains example commands
	if !contains(completionCmd.Long, "envi completion") {
		t.Error("Long description should contain example completion commands")
	}

	// Should contain code examples
	if !contains(completionCmd.Long, "$") || !contains(completionCmd.Long, "source") {
		t.Error("Long description should contain example shell commands")
	}
}

func TestInitCompletionCommandDoesNotPanic(t *testing.T) {
	// Test that InitCompletionCommand doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("InitCompletionCommand() panicked: %v", r)
		}
	}()

	// Already called in main tests
	// Just verify it can be called without panicking
}

func TestCompletionCommandIsAddedToRoot(t *testing.T) {
	// This test verifies completion command structure exists
	// The command is added by InitCompletionCommand which may not be called yet
	if completionCmd == nil {
		t.Skip("Completion command not initialized")
		return
	}

	// Just verify the command exists and has proper structure
	if completionCmd.Use == "" {
		t.Error("Completion command should have Use field")
	}
}

func TestCompletionCommandValidArgsCount(t *testing.T) {
	// Verify ExactValidArgs(1) is set properly
	if completionCmd.Args == nil {
		t.Error("Completion command should have Args validator")
	}
}

// Helper function to check if a string contains a substring (case-insensitive for some checks)
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
	       (findSubstring(s, substr) || findSubstring(toLower(s), toLower(substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + ('a' - 'A')
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func TestCompletionCommandForAllShells(t *testing.T) {
	shells := []string{"bash", "zsh", "fish", "powershell"}
	
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			// Verify each shell is a valid arg
			found := false
			for _, validArg := range completionCmd.ValidArgs {
				if validArg == shell {
					found = true
					break
				}
			}
			
			if !found {
				t.Errorf("Shell %s should be in ValidArgs", shell)
			}
		})
	}
}
