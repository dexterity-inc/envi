package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Test all command variables are initialized
func TestCommandVariablesInitialized(t *testing.T) {
	commands := []struct {
		name string
		cmd  interface{}
	}{
		{"validateCmd", validateCmd},
		{"mergeCmd", mergeCmd},
		{"decryptCmd", decryptCmd},
		{"listCmd", listCmd},
		{"completionCmd", completionCmd},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cmd == nil {
				t.Errorf("%s should be initialized", tc.name)
			}
		})
	}
}

// Test all command Use fields
func TestCommandUseFields(t *testing.T) {
	commands := []struct {
		name        string
		cmd         *cobra.Command
		expectedUse string
	}{
		{"root", rootCmd, "envi"},
		{"validate", validateCmd, "validate"},
		{"merge", mergeCmd, "merge"},
		{"decrypt", decryptCmd, "decrypt"},
		{"list", listCmd, "list"},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cmd == nil {
				t.Skip("Command not initialized")
				return
			}
			if !strings.Contains(tc.cmd.Use, tc.expectedUse) {
				t.Errorf("%s command Use should contain %q, got %q", tc.name, tc.expectedUse, tc.cmd.Use)
			}
		})
	}
}

// Test all commands have Short descriptions
func TestCommandShortDescriptions(t *testing.T) {
	commands := []*cobra.Command{
		rootCmd,
		validateCmd,
		mergeCmd,
		decryptCmd,
		listCmd,
		completionCmd,
	}

	for _, cmd := range commands {
		t.Run(cmd.Use, func(t *testing.T) {
			if cmd.Short == "" {
				t.Errorf("Command %s should have Short description", cmd.Use)
			}
			if len(cmd.Short) < 10 {
				t.Errorf("Command %s Short description too brief: %q", cmd.Use, cmd.Short)
			}
		})
	}
}

// Test all commands have Long descriptions
func TestCommandLongDescriptions(t *testing.T) {
	commands := []*cobra.Command{
		rootCmd,
		validateCmd,
		mergeCmd,
		decryptCmd,
		listCmd,
		completionCmd,
	}

	for _, cmd := range commands {
		t.Run(cmd.Use, func(t *testing.T) {
			if cmd.Long == "" {
				t.Errorf("Command %s should have Long description", cmd.Use)
			}
		})
	}
}

// Test validate command flags
func TestValidateCommandFlags(t *testing.T) {
	expectedFlags := []string{"fix", "strict", "required"}
	
	for _, flagName := range expectedFlags {
		t.Run(flagName, func(t *testing.T) {
			flag := validateCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("Validate command should have %s flag", flagName)
			}
		})
	}
}

// Test merge command flags
func TestMergeCommandFlags(t *testing.T) {
	expectedFlags := []string{"files", "gist", "output", "skip-duplicates", "overwrite", "keep-comments", "sort", "backup", "unmask"}
	
	for _, flagName := range expectedFlags {
		t.Run(flagName, func(t *testing.T) {
			flag := mergeCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("Merge command should have %s flag", flagName)
			}
		})
	}
}

// Test decrypt command flags
func TestDecryptCommandFlags(t *testing.T) {
	expectedFlags := []string{"input", "output", "self-contained", "use-key-file", "key-file", "mask"}
	
	for _, flagName := range expectedFlags {
		t.Run(flagName, func(t *testing.T) {
			flag := decryptCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("Decrypt command should have %s flag", flagName)
			}
		})
	}
}

// Test list command flags
func TestListCommandFlags(t *testing.T) {
	expectedFlags := []string{"all", "limit", "format", "urls", "tui"}
	
	for _, flagName := range expectedFlags {
		t.Run(flagName, func(t *testing.T) {
			flag := listCmd.Flags().Lookup(flagName)
			if flag == nil {
				t.Errorf("List command should have %s flag", flagName)
			}
		})
	}
}

// Test all commands have Run functions
func TestCommandsHaveRunFunctions(t *testing.T) {
	commands := []struct {
		name string
		cmd  *cobra.Command
	}{
		{"root", rootCmd},
		{"validate", validateCmd},
		{"merge", mergeCmd},
		{"decrypt", decryptCmd},
		{"list", listCmd},
		{"completion", completionCmd},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			if tc.cmd.Run == nil && tc.cmd.RunE == nil {
				t.Errorf("%s command should have Run or RunE function", tc.name)
			}
		})
	}
}

// Test flag default values for validate command
func TestValidateCommandFlagDefaults(t *testing.T) {
	tests := []struct {
		flagName     string
		expectedType string
	}{
		{"fix", "bool"},
		{"strict", "bool"},
		{"required", "stringSlice"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := validateCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("Flag %s not found", tt.flagName)
			}
			
			if flag.Value.Type() != tt.expectedType {
				t.Errorf("Flag %s should be type %s, got %s", tt.flagName, tt.expectedType, flag.Value.Type())
			}
		})
	}
}

// Test flag shorthand definitions
func TestFlagShorthands(t *testing.T) {
	tests := []struct {
		cmd       *cobra.Command
		flagName  string
		shorthand string
	}{
		{validateCmd, "strict", "s"},
		{mergeCmd, "files", "f"},
		{mergeCmd, "gist", "g"},
		{mergeCmd, "output", "o"},
		{mergeCmd, "skip-duplicates", "s"},
		{mergeCmd, "overwrite", "w"},
		{mergeCmd, "keep-comments", "c"},
		{decryptCmd, "input", "i"},
		{decryptCmd, "output", "o"},
		{decryptCmd, "key-file", "k"},
		{decryptCmd, "mask", "m"},
		{listCmd, "all", "a"},
		{listCmd, "limit", "l"},
		{listCmd, "format", "f"},
		{listCmd, "urls", "u"},
		{listCmd, "tui", "t"},
	}

	for _, tt := range tests {
		t.Run(tt.cmd.Use+"_"+tt.flagName, func(t *testing.T) {
			flag := tt.cmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("Flag %s not found on %s command", tt.flagName, tt.cmd.Use)
			}
			
			if flag.Shorthand != tt.shorthand {
				t.Errorf("Flag %s shorthand should be %q, got %q", tt.flagName, tt.shorthand, flag.Shorthand)
			}
		})
	}
}

// Test list command format flag values
func TestListCommandFormatFlagDefault(t *testing.T) {
	formatFlag := listCmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Fatal("format flag not found")
	}

	// Default should be "table"
	if formatFlag.DefValue != "table" {
		t.Errorf("format flag default should be 'table', got %q", formatFlag.DefValue)
	}
}

// Test list command limit flag default
func TestListCommandLimitFlagDefault(t *testing.T) {
	limitFlag := listCmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Fatal("limit flag not found")
	}

	// Default should be "10"
	if limitFlag.DefValue != "10" {
		t.Errorf("limit flag default should be '10', got %q", limitFlag.DefValue)
	}
}

// Test merge command boolean flag defaults
func TestMergeCommandBooleanFlagDefaults(t *testing.T) {
	tests := []struct {
		flagName     string
		expectedDefault string
	}{
		{"keep-comments", "true"},
		{"backup", "true"},
		{"sort", "false"},
		{"skip-duplicates", "false"},
		{"overwrite", "false"},
		{"unmask", "false"},
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			flag := mergeCmd.Flags().Lookup(tt.flagName)
			if flag == nil {
				t.Fatalf("Flag %s not found", tt.flagName)
			}

			if flag.DefValue != tt.expectedDefault {
				t.Errorf("Flag %s default should be %q, got %q", tt.flagName, tt.expectedDefault, flag.DefValue)
			}
		})
	}
}

// Test decrypt command output file default
func TestDecryptCommandOutputDefault(t *testing.T) {
	outputFlag := decryptCmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}

	// Default should be ".env"
	if outputFlag.DefValue != ".env" {
		t.Errorf("output flag default should be '.env', got %q", outputFlag.DefValue)
	}
}

// Test merge command output file default
func TestMergeCommandOutputDefault(t *testing.T) {
	outputFlag := mergeCmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Fatal("output flag not found")
	}

	// Default should be ".env"
	if outputFlag.DefValue != ".env" {
		t.Errorf("output flag default should be '.env', got %q", outputFlag.DefValue)
	}
}

// Test decrypt key file default
func TestDecryptKeyFileDefault(t *testing.T) {
	keyFileFlag := decryptCmd.Flags().Lookup("key-file")
	if keyFileFlag == nil {
		t.Fatal("key-file flag not found")
	}

	// Default should be ".envi.key"
	if keyFileFlag.DefValue != ".envi.key" {
		t.Errorf("key-file flag default should be '.envi.key', got %q", keyFileFlag.DefValue)
	}
}

// Test command help text quality
func TestCommandHelpTextQuality(t *testing.T) {
	commands := []*cobra.Command{
		validateCmd,
		mergeCmd,
		decryptCmd,
		listCmd,
	}

	for _, cmd := range commands {
		t.Run(cmd.Use, func(t *testing.T) {
			// Short should be concise
			if len(cmd.Short) > 100 {
				t.Errorf("Command %s Short description too long: %d chars", cmd.Use, len(cmd.Short))
			}

			// Long should be informative
			if len(cmd.Long) < len(cmd.Short) {
				t.Errorf("Command %s Long description should be longer than Short", cmd.Use)
			}

			// Should not contain placeholders
			if strings.Contains(cmd.Short, "TODO") || strings.Contains(cmd.Long, "TODO") {
				t.Errorf("Command %s contains TODO placeholders", cmd.Use)
			}
		})
	}
}

// Test all command flags have usage text
func TestAllFlagsHaveUsage(t *testing.T) {
	commands := []*cobra.Command{
		validateCmd,
		mergeCmd,
		decryptCmd,
		listCmd,
	}

	for _, cmd := range commands {
		t.Run(cmd.Use, func(t *testing.T) {
			cmd.Flags().VisitAll(func(flag *pflag.Flag) {
				if flag.Usage == "" {
					t.Errorf("Flag %s on command %s should have usage text", flag.Name, cmd.Use)
				}
			})
		})
	}
}

// Test root command version
func TestRootCommandVersion(t *testing.T) {
	if rootCmd.Version == "" {
		t.Error("Root command should have version set")
	}
}

// Test root command has persistent pre-run hook
func TestRootCommandHasPersistentPreRunHook(t *testing.T) {
	if rootCmd.PersistentPreRun == nil {
		t.Error("Root command should have PersistentPreRun")
	}
}
