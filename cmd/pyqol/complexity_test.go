package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestComplexityCommandInterface tests the basic command interface
func TestComplexityCommandInterface(t *testing.T) {
	// Test command creation and basic structure
	complexityCmd := NewComplexityCommand()
	if complexityCmd == nil {
		t.Fatal("NewComplexityCommand should return a valid command instance")
	}

	cobraCmd := complexityCmd.CreateCobraCommand()
	if cobraCmd == nil {
		t.Fatal("CreateCobraCommand should return a valid cobra command")
	}

	// Test command name and usage
	if cobraCmd.Use != "complexity [files...]" {
		t.Errorf("Expected command use 'complexity [files...]', got '%s'", cobraCmd.Use)
	}

	if cobraCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	// Test that flags are properly configured
	flags := cobraCmd.Flags()
	
	expectedFlags := []string{"format", "min", "max", "sort", "details", "config", "low-threshold", "medium-threshold"}
	for _, flagName := range expectedFlags {
		if !flags.HasFlags() {
			t.Error("Command should have flags defined")
			break
		}
		
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be defined", flagName)
		}
	}
}

// TestComplexityCommandValidation tests input validation without file analysis
func TestComplexityCommandValidation(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{
			name:        "No files provided",
			args:        []string{},
			expectError: true,
		},
		{
			name:        "Non-existent file",
			args:        []string{"/nonexistent/file.py"},
			expectError: true,
		},
		{
			name:        "Empty directory",
			args:        []string{"/tmp"},
			expectError: true, // Should fail because no Python files found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			complexityCmd := NewComplexityCommand()
			cobraCmd := complexityCmd.CreateCobraCommand()
			
			var output bytes.Buffer
			cobraCmd.SetOut(&output)
			cobraCmd.SetErr(&output)
			cobraCmd.SetArgs(tt.args)
			
			err := cobraCmd.Execute()
			
			if tt.expectError && err == nil {
				t.Error("Expected validation error but none occurred")
			} else if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestComplexityCommandFlags tests flag parsing and validation
func TestComplexityCommandFlags(t *testing.T) {
	// Create a temporary directory with a Python file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.py")
	
	err := os.WriteFile(testFile, []byte("def test(): pass"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	flagTests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "Valid format flag",
			args:    []string{"--format", "json", tempDir},
			wantErr: false, // May still fail due to file discovery, but flag should parse
		},
		{
			name:    "Invalid format flag",
			args:    []string{"--format", "invalid", tempDir},
			wantErr: true,
		},
		{
			name:    "Valid min complexity",
			args:    []string{"--min", "1", tempDir},
			wantErr: false,
		},
		{
			name:    "Invalid min complexity",
			args:    []string{"--min", "-1", tempDir},
			wantErr: true,
		},
		{
			name:    "Valid sort option",
			args:    []string{"--sort", "name", tempDir},
			wantErr: false,
		},
		{
			name:    "Invalid sort option",
			args:    []string{"--sort", "invalid", tempDir},
			wantErr: true,
		},
		{
			name:    "Invalid threshold combination",
			args:    []string{"--low-threshold", "10", "--medium-threshold", "5", tempDir},
			wantErr: true,
		},
	}

	for _, tt := range flagTests {
		t.Run(tt.name, func(t *testing.T) {
			complexityCmd := NewComplexityCommand()
			cobraCmd := complexityCmd.CreateCobraCommand()
			
			var output bytes.Buffer
			cobraCmd.SetOut(&output)
			cobraCmd.SetErr(&output)
			cobraCmd.SetArgs(tt.args)
			
			err := cobraCmd.Execute()
			
			// We expect either validation error OR file discovery error
			// Both are acceptable for these flag validation tests
			if !tt.wantErr && err != nil {
				// Check if it's a file discovery error (acceptable) vs validation error
				errMsg := err.Error()
				if !strings.Contains(errMsg, "no Python files found") && 
				   !strings.Contains(errMsg, "file not found") {
					t.Errorf("Unexpected validation error: %v", err)
				}
			} else if tt.wantErr && err == nil {
				t.Error("Expected validation error but none occurred")
			}
		})
	}
}

// TestComplexityCommandDefaults tests default values
func TestComplexityCommandDefaults(t *testing.T) {
	cmd := NewComplexityCommand()
	
	if cmd.outputFormat != "text" {
		t.Errorf("Expected default outputFormat to be 'text', got '%s'", cmd.outputFormat)
	}
	
	if cmd.minComplexity != 1 {
		t.Errorf("Expected default minComplexity to be 1, got %d", cmd.minComplexity)
	}
	
	if cmd.maxComplexity != 0 {
		t.Errorf("Expected default maxComplexity to be 0, got %d", cmd.maxComplexity)
	}
	
	if cmd.sortBy != "complexity" {
		t.Errorf("Expected default sortBy to be 'complexity', got '%s'", cmd.sortBy)
	}
	
	if cmd.lowThreshold != 9 {
		t.Errorf("Expected default lowThreshold to be 9, got %d", cmd.lowThreshold)
	}
	
	if cmd.mediumThreshold != 19 {
		t.Errorf("Expected default mediumThreshold to be 19, got %d", cmd.mediumThreshold)
	}
}

// TestComplexityCommandHelp tests help output
func TestComplexityCommandHelp(t *testing.T) {
	complexityCmd := NewComplexityCommand()
	cobraCmd := complexityCmd.CreateCobraCommand()
	
	var output bytes.Buffer
	cobraCmd.SetOut(&output)
	cobraCmd.SetArgs([]string{"--help"})
	
	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("Help command should not return error: %v", err)
	}
	
	helpOutput := output.String()
	
	// Check that help contains key information
	expectedContent := []string{
		"complexity",
		"Python files",
		"--format",
		"--min",
		"--sort",
	}
	
	for _, content := range expectedContent {
		if !strings.Contains(helpOutput, content) {
			t.Errorf("Help output should contain '%s'", content)
		}
	}
}