package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestAnalyzeCommandInterface tests the analyze command interface
func TestAnalyzeCommandInterface(t *testing.T) {
	analyzeCmd := NewAnalyzeCommand()
	if analyzeCmd == nil {
		t.Fatal("NewAnalyzeCommand should return a valid command instance")
	}

	cobraCmd := analyzeCmd.CreateCobraCommand()
	if cobraCmd == nil {
		t.Fatal("CreateCobraCommand should return a valid cobra command")
	}

	// Test command name and usage
	if cobraCmd.Use != "analyze [files...]" {
		t.Errorf("Expected command use 'analyze [files...]', got '%s'", cobraCmd.Use)
	}

	if cobraCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	// Test that essential flags are present
	flags := cobraCmd.Flags()
	expectedFlags := []string{"format", "config", "skip-complexity", "skip-deadcode", "skip-clones"}
	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be defined", flagName)
		}
	}
}

// TestCheckCommandInterface tests the check command interface
func TestCheckCommandInterface(t *testing.T) {
	checkCmd := NewCheckCommand()
	if checkCmd == nil {
		t.Fatal("NewCheckCommand should return a valid command instance")
	}

	cobraCmd := checkCmd.CreateCobraCommand()
	if cobraCmd == nil {
		t.Fatal("CreateCobraCommand should return a valid cobra command")
	}

	// Test command name and usage
	if cobraCmd.Use != "check [files...]" {
		t.Errorf("Expected command use 'check [files...]', got '%s'", cobraCmd.Use)
	}

	if cobraCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	// Test CI-friendly flags
	flags := cobraCmd.Flags()
	expectedFlags := []string{"quiet", "max-complexity", "allow-dead-code", "skip-clones"}
	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be defined", flagName)
		}
	}
}

// TestVersionCommandInterface tests the version command interface
func TestVersionCommandInterface(t *testing.T) {
	versionCmd := NewVersionCommand()
	if versionCmd == nil {
		t.Fatal("NewVersionCommand should return a valid command instance")
	}

	cobraCmd := versionCmd.CreateCobraCommand()
	if cobraCmd == nil {
		t.Fatal("CreateCobraCommand should return a valid cobra command")
	}

	// Test command name and usage
	if cobraCmd.Use != "version" {
		t.Errorf("Expected command use 'version', got '%s'", cobraCmd.Use)
	}

	// Test version command execution
	var output bytes.Buffer
	cobraCmd.SetOut(&output)
	cobraCmd.SetErr(&output)
	
	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("Version command should not fail: %v", err)
	}

	result := output.String()
	if result == "" {
		t.Error("Version command should produce output")
	}
}

// TestInitCommandInterface tests the init command interface
func TestInitCommandInterface(t *testing.T) {
	initCmd := NewInitCommand()
	if initCmd == nil {
		t.Fatal("NewInitCommand should return a valid command instance")
	}

	cobraCmd := initCmd.CreateCobraCommand()
	if cobraCmd == nil {
		t.Fatal("CreateCobraCommand should return a valid cobra command")
	}

	// Test command name and usage
	if cobraCmd.Use != "init" {
		t.Errorf("Expected command use 'init', got '%s'", cobraCmd.Use)
	}

	if cobraCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	// Test that flags are present
	flags := cobraCmd.Flags()
	expectedFlags := []string{"force", "config"}
	for _, flagName := range expectedFlags {
		flag := flags.Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected flag '%s' to be defined", flagName)
		}
	}
}

// TestInitCommandExecution tests init command file creation
func TestInitCommandExecution(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".pyqol.yaml")
	
	initCmd := NewInitCommand()
	cobraCmd := initCmd.CreateCobraCommand()
	
	var output bytes.Buffer
	cobraCmd.SetOut(&output)
	cobraCmd.SetErr(&output)
	
	// Set the args to specify the config file location
	cobraCmd.SetArgs([]string{"--config", configFile})
	
	// Test successful creation
	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("Init command should not fail: %v", err)
	}
	
	// Check if file was created
	if _, err := os.Stat(configFile); err != nil {
		t.Errorf("Configuration file should be created: %v", err)
	}
	
	// Check file content
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Should be able to read config file: %v", err)
	}
	
	contentStr := string(content)
	if !strings.Contains(contentStr, "complexity:") {
		t.Error("Config file should contain complexity section")
	}
	if !strings.Contains(contentStr, "dead_code:") {
		t.Error("Config file should contain dead_code section")
	}
	if !strings.Contains(contentStr, "clones:") {
		t.Error("Config file should contain clones section")
	}
}

// TestInitCommandFileExists tests init command behavior when file already exists
func TestInitCommandFileExists(t *testing.T) {
	// Create temporary directory with existing config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".pyqol.yaml")
	
	// Create existing file
	err := os.WriteFile(configFile, []byte("existing config"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	initCmd := NewInitCommand()
	cobraCmd := initCmd.CreateCobraCommand()
	
	var output bytes.Buffer
	cobraCmd.SetOut(&output)
	cobraCmd.SetErr(&output)
	
	// Should fail without --force
	cobraCmd.SetArgs([]string{"--config", configFile})
	err = cobraCmd.Execute()
	if err == nil {
		t.Error("Init command should fail when file exists without --force")
	}
	
	// Should succeed with --force
	output.Reset()
	cobraCmd.SetArgs([]string{"--config", configFile, "--force"})
	err = cobraCmd.Execute()
	if err != nil {
		t.Errorf("Init command should succeed with --force: %v", err)
	}
	
	// Check that file was overwritten
	content, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Should be able to read config file: %v", err)
	}
	
	if strings.Contains(string(content), "existing config") {
		t.Error("File should be overwritten with --force")
	}
}

// TestVersionCommandShortFlag tests version command --short flag
func TestVersionCommandShortFlag(t *testing.T) {
	versionCmd := NewVersionCommand()
	cobraCmd := versionCmd.CreateCobraCommand()
	
	var output bytes.Buffer
	cobraCmd.SetOut(&output)
	cobraCmd.SetErr(&output)
	
	// Test with --short flag
	cobraCmd.SetArgs([]string{"--short"})
	
	err := cobraCmd.Execute()
	if err != nil {
		t.Fatalf("Version command with --short should not fail: %v", err)
	}
	
	result := strings.TrimSpace(output.String())
	
	if result == "" {
		t.Error("Short version should not be empty")
	}
	
	// Test without --short flag (full version)
	output.Reset()
	cobraCmd.SetArgs([]string{})
	
	err = cobraCmd.Execute()
	if err != nil {
		t.Fatalf("Version command should not fail: %v", err)
	}
	
	fullResult := strings.TrimSpace(output.String())
	if fullResult == "" {
		t.Error("Full version should not be empty")
	}
}

// TestAnalyzeCommandValidation tests analyze command input validation
func TestAnalyzeCommandValidation(t *testing.T) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analyzeCmd := NewAnalyzeCommand()
			cobraCmd := analyzeCmd.CreateCobraCommand()
			
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

// TestCheckCommandValidation tests check command input validation
func TestCheckCommandValidation(t *testing.T) {
	// Check command should default to current directory if no args
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	
	// This shouldn't fail validation (though analysis might fail)
	var output bytes.Buffer
	cobraCmd.SetOut(&output)
	cobraCmd.SetErr(&output)
	cobraCmd.SetArgs([]string{}) // No args - should default to "."
	
	// We can't easily test full execution without proper setup,
	// but we can test that validation passes
	if cobraCmd.Args != nil {
		err := cobraCmd.Args(cobraCmd, []string{})
		if err != nil {
			t.Errorf("Check command should accept empty args: %v", err)
		}
	}
}

// TestCommandHelpOutput tests that help output is comprehensive
func TestCommandHelpOutput(t *testing.T) {
	commands := []struct {
		name    string
		command func() *cobra.Command
	}{
		{"analyze", func() *cobra.Command { return NewAnalyzeCmd() }},
		{"check", func() *cobra.Command { return NewCheckCmd() }},
		{"version", func() *cobra.Command { return NewVersionCmd() }},
		{"init", func() *cobra.Command { return NewInitCmd() }},
	}

	for _, cmd := range commands {
		t.Run(cmd.name, func(t *testing.T) {
			cobraCmd := cmd.command()
			
			// Test help output
			var output bytes.Buffer
			cobraCmd.SetOut(&output)
			cobraCmd.SetArgs([]string{"--help"})
			
			err := cobraCmd.Execute()
			if err != nil {
				t.Fatalf("Help command should not fail: %v", err)
			}
			
			helpOutput := output.String()
			
			// Check that help contains essential elements
			if !strings.Contains(helpOutput, "Usage:") {
				t.Error("Help should contain Usage section")
			}
			
			if !strings.Contains(helpOutput, "Examples:") {
				t.Error("Help should contain Examples section")
			}
			
			if !strings.Contains(helpOutput, "Flags:") {
				t.Error("Help should contain Flags section")
			}
		})
	}
}