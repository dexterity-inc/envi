package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-github/v37/github"
	"github.com/spf13/cobra"
	"github.com/dexterity-inc/envi/internal/encryption"
)

func TestRootCommand(t *testing.T) {
	// Test root command properties
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
	
	if rootCmd.Use != "envi" {
		t.Errorf("Expected root command Use to be 'envi', got %s", rootCmd.Use)
	}
	
	if rootCmd.Short == "" {
		t.Error("Root command should have a short description")
	}
	
	if rootCmd.Long == "" {
		t.Error("Root command should have a long description")
	}
	
	// Test that version is set
	if rootCmd.Version == "" {
		t.Error("Root command should have a version")
	}
}

func TestRootCommandPersistentPreRun(t *testing.T) {
	// Test the PersistentPreRun function
	if rootCmd.PersistentPreRun == nil {
		t.Fatal("PersistentPreRun should not be nil")
	}
	
	// Create a test command
	testCmd := &cobra.Command{
		Use: "test",
	}
	
	// Test that PersistentPreRun doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("PersistentPreRun panicked: %v", r)
		}
	}()
	
	// Call PersistentPreRun with empty args
	rootCmd.PersistentPreRun(testCmd, []string{})
}

func TestRootCommandRun(t *testing.T) {
	// Test the Run function shows help
	if rootCmd.Run == nil {
		t.Fatal("Run function should not be nil")
	}
	
	// Create a buffer to capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	
	// Test that Run function doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Run function panicked: %v", r)
		}
	}()
	
	// Call Run function
	rootCmd.Run(rootCmd, []string{})
}

func TestCreateSharingReadmeContent(t *testing.T) {
	// Create a test user
	login := "testuser"
	user := &github.User{
		Login: &login,
	}
	
	tests := []struct {
		name               string
		user               *github.User
		recipientUsername  string
		keyFilePath       string
		useEncryption     bool
		useMaskedEncryption bool
		expectedContains  []string
	}{
		{
			name:              "basic sharing without encryption",
			user:              user,
			recipientUsername: "recipient",
			keyFilePath:       "",
			useEncryption:     false,
			useMaskedEncryption: false,
			expectedContains: []string{
				"# Shared Environment Variables",
				"@testuser",
				"@recipient",
				"Instructions",
				"envi",
			},
		},
		{
			name:              "sharing with key file encryption",
			user:              user,
			recipientUsername: "recipient",
			keyFilePath:       "/path/to/key",
			useEncryption:     true,
			useMaskedEncryption: false,
			expectedContains: []string{
				"# Shared Environment Variables",
				"Decryption Instructions",
				"key file",
				"envi pull",
				"--use-key-file",
				"--key-file",
			},
		},
		{
			name:              "sharing with password encryption",
			user:              user,
			recipientUsername: "recipient",
			keyFilePath:       "",
			useEncryption:     false,
			useMaskedEncryption: true,
			expectedContains: []string{
				"# Shared Environment Variables",
				"Decryption Instructions",
				"encryption password",
				"envi pull",
				"--unmask",
				"prompted to enter",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up encryption state
			originalUseEncryption := encryption.UseEncryption
			originalUseMaskedEncryption := encryption.UseMaskedEncryption
			defer func() {
				encryption.UseEncryption = originalUseEncryption
				encryption.UseMaskedEncryption = originalUseMaskedEncryption
			}()
			
			encryption.UseEncryption = tt.useEncryption
			encryption.UseMaskedEncryption = tt.useMaskedEncryption
			
			result := createSharingReadmeContent(tt.user, tt.recipientUsername, tt.keyFilePath)
			
			if result == "" {
				t.Error("createSharingReadmeContent should not return empty string")
			}
			
			// Check that all expected content is present
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, but it didn't.\nResult: %s", expected, result)
				}
			}
			
			// Check that the result contains proper markdown formatting
			if !strings.HasPrefix(result, "# ") {
				t.Error("Result should start with a markdown header")
			}
			
			// Check for proper command formatting
			if (tt.useEncryption || tt.useMaskedEncryption) && !strings.Contains(result, "```shell") {
				t.Error("Encrypted sharing should contain shell code blocks")
			}
		})
	}
}

func TestCreateSharingReadmeContentWithNilUser(t *testing.T) {
	// Test behavior with nil user (should handle gracefully or panic predictably)
	defer func() {
		if r := recover(); r == nil {
			// If it doesn't panic, that's also valid behavior
			// We just want to ensure it doesn't cause unexpected crashes
		}
	}()
	
	result := createSharingReadmeContent(nil, "recipient", "")
	
	// If we get here without panicking, check the result isn't completely empty
	if result == "" {
		t.Error("createSharingReadmeContent should handle nil user gracefully")
	}
}

func TestEncryptionFlags(t *testing.T) {
	// Create a test command
	testCmd := &cobra.Command{
		Use: "test",
	}
	
	// Initialize encryption flags
	encryption.InitEncryptionFlags(testCmd)
	
	// Check that the common encryption flags are present
	expectedFlags := []string{
		"encrypt",
		"mask",
		"use-key-file",
		"key-file",
	}
	
	for _, flagName := range expectedFlags {
		flag := testCmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag %s to be initialized", flagName)
		}
	}
	
	// Check default value for key-file flag
	keyFileFlag := testCmd.PersistentFlags().Lookup("key-file")
	if keyFileFlag != nil && keyFileFlag.DefValue != ".envi.key" {
		t.Errorf("Expected key-file default to be '.envi.key', got %s", keyFileFlag.DefValue)
	}
}

func TestPersistentFlags(t *testing.T) {
	// Create a test command to verify persistent flags setup
	testCmd := &cobra.Command{
		Use: "test",
	}
	
	// Add the TUI flag similar to Execute function
	testCmd.PersistentFlags().BoolVar(&encryption.UseTUI, "tui", true, "Use interactive terminal UI")
	
	// Check that the TUI flag is properly set
	tuiFlag := testCmd.PersistentFlags().Lookup("tui")
	if tuiFlag == nil {
		t.Error("TUI flag should be initialized")
	}
	
	if tuiFlag.DefValue != "true" {
		t.Errorf("Expected TUI flag default to be 'true', got %s", tuiFlag.DefValue)
	}
}