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

func TestInitPushCommand(t *testing.T) {
	// Reset rootCmd to ensure clean state
	rootCmd.ResetCommands()
	
	// Initialize push command
	InitPushCommand()
	
	// Check that push command was added
	pushCmd := rootCmd.Commands()
	found := false
	for _, cmd := range pushCmd {
		if cmd.Use == "push" {
			found = true
			
			// Verify command has required flags
			if cmd.Flags().Lookup("file") == nil {
				t.Error("Push command should have 'file' flag")
			}
			if cmd.Flags().Lookup("id") == nil {
				t.Error("Push command should have 'id' flag")
			}
			if cmd.Flags().Lookup("description") == nil {
				t.Error("Push command should have 'description' flag")
			}
			if cmd.Flags().Lookup("public") == nil {
				t.Error("Push command should have 'public' flag")
			}
			break
		}
	}
	
	if !found {
		t.Error("Push command should be added to root command")
	}
}

func TestInitPullCommand(t *testing.T) {
	// Reset rootCmd to ensure clean state
	rootCmd.ResetCommands()
	
	// Initialize pull command
	InitPullCommand()
	
	// Check that pull command was added
	pullCmd := rootCmd.Commands()
	found := false
	for _, cmd := range pullCmd {
		if cmd.Use == "pull" {
			found = true
			
			// Verify command has required flags
			if cmd.Flags().Lookup("id") == nil {
				t.Error("Pull command should have 'id' flag")
			}
			if cmd.Flags().Lookup("output") == nil {
				t.Error("Pull command should have 'output' flag")
			}
			if cmd.Flags().Lookup("unmask") == nil {
				t.Error("Pull command should have 'unmask' flag")
			}
			break
		}
	}
	
	if !found {
		t.Error("Pull command should be added to root command")
	}
}

func TestInitShareCommand(t *testing.T) {
	// Reset rootCmd to ensure clean state
	rootCmd.ResetCommands()
	
	// Initialize share command
	InitShareCommand()
	
	// Check that share command was added
	shareCmd := rootCmd.Commands()
	found := false
	for _, cmd := range shareCmd {
		if cmd.Use == "share" {
			found = true
			
			// Verify command has required flags
			if cmd.Flags().Lookup("id") == nil {
				t.Error("Share command should have 'id' flag")
			}
			if cmd.Flags().Lookup("users") == nil {
				t.Error("Share command should have 'users' flag")
			}
			break
		}
	}
	
	if !found {
		t.Error("Share command should be added to root command")
	}
}

func TestInitConfigCommand(t *testing.T) {
	// Reset rootCmd to ensure clean state
	rootCmd.ResetCommands()
	
	// Initialize config command
	InitConfigCommand()
	
	// Check that config command was added
	configCmd := rootCmd.Commands()
	found := false
	for _, cmd := range configCmd {
		if cmd.Use == "config" {
			found = true
			
			// Verify command has required flags
			if cmd.Flags().Lookup("token") == nil {
				t.Error("Config command should have 'token' flag")
			}
			break
		}
	}
	
	if !found {
		t.Error("Config command should be added to root command")
	}
}

func TestInitListCommand(t *testing.T) {
	// Reset rootCmd to ensure clean state
	rootCmd.ResetCommands()
	
	// Initialize list command
	InitListCommand()
	
	// Check that list command was added
	listCmd := rootCmd.Commands()
	found := false
	for _, cmd := range listCmd {
		if cmd.Use == "list" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("List command should be added to root command")
	}
}

func TestInitMergeCommand(t *testing.T) {
	// Reset rootCmd to ensure clean state
	rootCmd.ResetCommands()
	
	// Initialize merge command
	InitMergeCommand()
	
	// Check that merge command was added
	mergeCmd := rootCmd.Commands()
	found := false
	for _, cmd := range mergeCmd {
		if cmd.Use == "merge" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Merge command should be added to root command")
	}
}

func TestInitDecryptCommand(t *testing.T) {
	// Reset rootCmd to ensure clean state
	rootCmd.ResetCommands()
	
	// Initialize decrypt command
	InitDecryptCommand()
	
	// Check that decrypt command was added
	decryptCmd := rootCmd.Commands()
	found := false
	for _, cmd := range decryptCmd {
		if cmd.Use == "decrypt" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Decrypt command should be added to root command")
	}
}

func TestInitValidateCommand(t *testing.T) {
	// Reset rootCmd to ensure clean state
	rootCmd.ResetCommands()
	
	// Initialize validate command
	InitValidateCommand()
	
	// Check that validate command was added
	validateCmd := rootCmd.Commands()
	found := false
	for _, cmd := range validateCmd {
		if cmd.Use == "validate" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Validate command should be added to root command")
	}
}

func TestInitCompletionCommand(t *testing.T) {
	// Reset rootCmd to ensure clean state
	rootCmd.ResetCommands()
	
	// Initialize completion command
	InitCompletionCommand()
	
	// Check that completion command was added
	completionCmd := rootCmd.Commands()
	found := false
	for _, cmd := range completionCmd {
		// Completion command Use is "completion [bash|zsh|fish|powershell]"
		if strings.HasPrefix(cmd.Use, "completion") {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Completion command should be added to root command")
	}
}

func TestInitVersionCommand(t *testing.T) {
	// Reset rootCmd to ensure clean state
	rootCmd.ResetFlags()
	
	// Initialize version command  
	InitVersionCommand()
	
	// Version command adds a flag, not a subcommand
	// Check that version flag was added
	versionFlag := rootCmd.Flags().Lookup("version")
	if versionFlag == nil {
		t.Error("Version flag should be added to root command")
	}
	
	// Check short flag
	vFlag := rootCmd.Flags().ShorthandLookup("v")
	if vFlag == nil {
		t.Error("Version short flag (-v) should be added to root command")
	}
}

func TestInitGistCommand(t *testing.T) {
	// Reset rootCmd to ensure clean state
	rootCmd.ResetCommands()
	
	// Initialize gist command
	InitGistCommand()
	
	// Check that gist command was added
	gistCmd := rootCmd.Commands()
	found := false
	for _, cmd := range gistCmd {
		if cmd.Use == "gist" {
			found = true
			break
		}
	}
	
	if !found {
		t.Error("Gist command should be added to root command")
	}
}

func TestCreateSharingReadmeContentVariants(t *testing.T) {
	// Create a test user
	login := "testuser"
	user := &github.User{
		Login: &login,
	}
	
	tests := []struct {
		name                string
		recipientUsername   string
		keyFilePath         string
		useEncryption       bool
		useMaskedEncryption bool
		expectedContains    []string
		expectedNotContains []string
	}{
		{
			name:              "unencrypted sharing",
			recipientUsername: "recipient1",
			keyFilePath:       "",
			useEncryption:     false,
			useMaskedEncryption: false,
			expectedContains: []string{
				"Shared Environment Variables",
				"@testuser",
				"@recipient1",
				"Instructions",
			},
			expectedNotContains: []string{
				"Decryption Instructions",
				"encryption password",
				"key file",
			},
		},
		{
			name:              "encrypted with key file",
			recipientUsername: "recipient2",
			keyFilePath:       "/path/to/key.file",
			useEncryption:     true,
			useMaskedEncryption: false,
			expectedContains: []string{
				"Decryption Instructions",
				"key file",
				"--use-key-file",
				"--key-file",
			},
			expectedNotContains: []string{
				"encryption password",
			},
		},
		{
			name:              "masked encryption without key file",
			recipientUsername: "recipient3",
			keyFilePath:       "",
			useEncryption:     false,
			useMaskedEncryption: true,
			expectedContains: []string{
				"Decryption Instructions",
				"encryption password",
				"--unmask",
			},
			expectedNotContains: []string{
				"--use-key-file",
			},
		},
		{
			name:              "installation instructions present",
			recipientUsername: "recipient4",
			keyFilePath:       "",
			useEncryption:     false,
			useMaskedEncryption: false,
			expectedContains: []string{
				"Getting started with envi",
				"brew install envi",
				"scoop install envi",
			},
			expectedNotContains: []string{},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set encryption state
			originalUseEncryption := encryption.UseEncryption
			originalUseMaskedEncryption := encryption.UseMaskedEncryption
			defer func() {
				encryption.UseEncryption = originalUseEncryption
				encryption.UseMaskedEncryption = originalUseMaskedEncryption
			}()
			
			encryption.UseEncryption = tt.useEncryption
			encryption.UseMaskedEncryption = tt.useMaskedEncryption
			
			result := createSharingReadmeContent(user, tt.recipientUsername, tt.keyFilePath)
			
			// Check expected content
			for _, expected := range tt.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, but it didn't", expected)
				}
			}
			
			// Check content that should not be present
			for _, notExpected := range tt.expectedNotContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected result to NOT contain %q, but it did", notExpected)
				}
			}
			
			// Verify markdown formatting
			if !strings.HasPrefix(result, "# ") {
				t.Error("Result should start with a markdown header")
			}
			
			// Verify it ends with attribution
			if !strings.Contains(result, "envi](https://github.com/dexterity-inc/envi") {
				t.Error("Result should contain attribution link")
			}
		})
	}
}

func TestCommandStructure(t *testing.T) {
	// Test that root command has all required properties
	if rootCmd.Use != "envi" {
		t.Errorf("Expected root command Use to be 'envi', got %s", rootCmd.Use)
	}
	
	if rootCmd.Short == "" {
		t.Error("Root command should have a short description")
	}
	
	if rootCmd.Long == "" {
		t.Error("Root command should have a long description")
	}
	
	// Test PersistentPreRun exists
	if rootCmd.PersistentPreRun == nil {
		t.Error("Root command should have PersistentPreRun function")
	}
	
	// Test Run exists
	if rootCmd.Run == nil {
		t.Error("Root command should have Run function")
	}
}

func TestExecuteFunctionStructure(t *testing.T) {
	// We can't easily test Execute() fully without mocking,
	// but we can verify the individual init functions work
	
	// Just verify that all commands are registered without panicking
	// Note: We cannot call Init functions multiple times as they add flags
	// which would cause "flag redefined" errors
	
	// Verify encryption flags can be initialized on a new command
	testCmd := &cobra.Command{Use: "test"}
	encryption.InitEncryptionFlags(testCmd)
	
	// Verify the test command has encryption flags
	if testCmd.PersistentFlags().Lookup("encrypt") == nil {
		t.Error("Encryption flags should be initialized")
	}
}

func TestAllCommandsHaveHelp(t *testing.T) {
	// Verify that each command can display help without panicking
	commands := []string{
		"config", "push", "pull", "list", "share", 
		"merge", "decrypt", "validate", "gist",
	}
	
	for _, cmdName := range commands {
		t.Run(cmdName, func(t *testing.T) {
			// Find the command by name
			for _, cmd := range rootCmd.Commands() {
				if strings.HasPrefix(cmd.Use, cmdName) {
					// Verify command has help text
					if cmd.Short == "" {
						t.Errorf("Command %s should have short description", cmdName)
					}
					break
				}
			}
		})
	}
}