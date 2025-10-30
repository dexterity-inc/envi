package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name:        "copy simple file",
			content:     "Hello, World!",
			expectError: false,
		},
		{
			name: "copy multi-line file",
			content: `Line 1
Line 2
Line 3`,
			expectError: false,
		},
		{
			name:        "copy empty file",
			content:     "",
			expectError: false,
		},
		{
			name: "copy env file content",
			content: `DB_HOST=localhost
DB_PORT=5432
API_KEY=secret123`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			srcFile := filepath.Join(tmpDir, "source.txt")
			dstFile := filepath.Join(tmpDir, "destination.txt")

			// Create source file
			err := os.WriteFile(srcFile, []byte(tt.content), 0600)
			if err != nil {
				t.Fatalf("Failed to create source file: %v", err)
			}

			// Copy file
			err = copyFile(srcFile, dstFile)

			// Check error
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError {
				// Verify destination file exists
				if _, err := os.Stat(dstFile); os.IsNotExist(err) {
					t.Error("Destination file was not created")
				}

				// Verify content matches
				dstContent, err := os.ReadFile(dstFile)
				if err != nil {
					t.Fatalf("Failed to read destination file: %v", err)
				}

				if string(dstContent) != tt.content {
					t.Errorf("Content mismatch: expected %q, got %q", tt.content, string(dstContent))
				}

				// Verify file permissions
				info, err := os.Stat(dstFile)
				if err != nil {
					t.Fatalf("Failed to stat destination file: %v", err)
				}

				expectedPerm := os.FileMode(0600)
				if info.Mode().Perm() != expectedPerm {
					t.Errorf("Expected permissions %v, got %v", expectedPerm, info.Mode().Perm())
				}
			}
		})
	}
}

func TestCopyFileNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "nonexistent.txt")
	dstFile := filepath.Join(tmpDir, "destination.txt")

	err := copyFile(srcFile, dstFile)
	if err == nil {
		t.Error("Expected error for non-existent source file")
	}
}

func TestCopyFileInvalidDestination(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")

	// Create source file
	err := os.WriteFile(srcFile, []byte("content"), 0600)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Try to copy to invalid destination
	err = copyFile(srcFile, "/invalid/path/destination.txt")
	if err == nil {
		t.Error("Expected error for invalid destination path")
	}
}

func TestSortKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		expected []string
	}{
		{
			name: "alphabetical order",
			input: map[string]string{
				"ZEBRA": "z",
				"APPLE": "a",
				"MANGO": "m",
			},
			expected: []string{"APPLE", "MANGO", "ZEBRA"},
		},
		{
			name: "already sorted",
			input: map[string]string{
				"A": "1",
				"B": "2",
				"C": "3",
			},
			expected: []string{"A", "B", "C"},
		},
		{
			name: "reverse order",
			input: map[string]string{
				"C": "3",
				"B": "2",
				"A": "1",
			},
			expected: []string{"A", "B", "C"},
		},
		{
			name:     "empty map",
			input:    map[string]string{},
			expected: []string{},
		},
		{
			name: "single element",
			input: map[string]string{
				"ONLY": "one",
			},
			expected: []string{"ONLY"},
		},
		{
			name: "env variables",
			input: map[string]string{
				"DB_PORT":    "5432",
				"DB_HOST":    "localhost",
				"API_KEY":    "secret",
				"SECRET_KEY": "topsecret",
			},
			expected: []string{"API_KEY", "DB_HOST", "DB_PORT", "SECRET_KEY"},
		},
		{
			name: "case sensitive sorting",
			input: map[string]string{
				"zebra": "z",
				"APPLE": "a",
				"Mango": "m",
			},
			expected: []string{"APPLE", "Mango", "zebra"},
		},
		{
			name: "numbers in keys",
			input: map[string]string{
				"VAR_3": "three",
				"VAR_1": "one",
				"VAR_2": "two",
			},
			expected: []string{"VAR_1", "VAR_2", "VAR_3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortKeys(tt.input)

			// Check length
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d keys, got %d", len(tt.expected), len(result))
			}

			// Check order
			for i, expected := range tt.expected {
				if i >= len(result) {
					t.Errorf("Missing expected key at index %d: %s", i, expected)
					continue
				}
				if result[i] != expected {
					t.Errorf("At index %d: expected %s, got %s", i, expected, result[i])
				}
			}

			// Verify all keys from input are in output
			for key := range tt.input {
				found := false
				for _, resultKey := range result {
					if resultKey == key {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Key %s from input not found in result", key)
				}
			}
		})
	}
}

func TestSortKeysLarge(t *testing.T) {
	// Test with a larger dataset
	input := make(map[string]string)
	for i := 100; i > 0; i-- {
		key := string(rune('A' + (i % 26)))
		input[key] = "value"
	}

	result := sortKeys(input)

	// Verify result is sorted
	for i := 1; i < len(result); i++ {
		if result[i-1] > result[i] {
			t.Errorf("Result not sorted at index %d: %s > %s", i, result[i-1], result[i])
		}
	}
}

func TestCopyFileLargeContent(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "large.txt")
	dstFile := filepath.Join(tmpDir, "large_copy.txt")

	// Create a large content (1MB)
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte('A' + (i % 26))
	}

	err := os.WriteFile(srcFile, largeContent, 0600)
	if err != nil {
		t.Fatalf("Failed to create large source file: %v", err)
	}

	// Copy the file
	err = copyFile(srcFile, dstFile)
	if err != nil {
		t.Fatalf("Failed to copy large file: %v", err)
	}

	// Verify content
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if len(dstContent) != len(largeContent) {
		t.Errorf("Size mismatch: expected %d, got %d", len(largeContent), len(dstContent))
	}
}

func TestSortKeysPreservesAllKeys(t *testing.T) {
	// Ensure sortKeys doesn't lose any keys
	input := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
		"KEY3": "value3",
		"KEY4": "value4",
		"KEY5": "value5",
	}

	result := sortKeys(input)

	if len(result) != len(input) {
		t.Errorf("Key count mismatch: expected %d, got %d", len(input), len(result))
	}

	// Check each key is present
	for key := range input {
		found := false
		for _, k := range result {
			if k == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Key %s not found in sorted result", key)
		}
	}
}
