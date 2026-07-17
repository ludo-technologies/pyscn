package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
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
	expectedFlags := []string{"html", "json", "csv", "yaml", "config", "skip-complexity", "skip-deadcode", "skip-clones"}
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
	expectedFlags := []string{"quiet", "max-complexity", "allow-dead-code", "skip-clones", "allow-circular-deps", "max-cycles"}
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
	configFile := filepath.Join(tempDir, ".pyscn.toml")

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

	// Check for top-level sections
	if !strings.Contains(contentStr, "[output]") {
		t.Error("Config file should contain [output] section")
	}
	if !strings.Contains(contentStr, "[analysis]") {
		t.Error("Config file should contain [analysis] section")
	}
	if !strings.Contains(contentStr, "[complexity]") {
		t.Error("Config file should contain [complexity] section")
	}
	if !strings.Contains(contentStr, "[dead_code]") {
		t.Error("Config file should contain [dead_code] section")
	}
	if !strings.Contains(contentStr, "[clones]") {
		t.Error("Config file should contain [clones] section")
	}

	// Check for key settings
	if !strings.Contains(contentStr, "min_lines") {
		t.Error("Config file should contain min_lines setting")
	}
	if !strings.Contains(contentStr, "type1_threshold") {
		t.Error("Config file should contain type1_threshold setting")
	}
	if !strings.Contains(contentStr, "lsh_enabled") {
		t.Error("Config file should contain lsh_enabled setting")
	}
	if !strings.Contains(contentStr, "include_patterns") {
		t.Error("Config file should contain include_patterns setting")
	}
}

// TestInitCommandFileExists tests init command behavior when file already exists
func TestInitCommandFileExists(t *testing.T) {
	// Create temporary directory with existing config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, ".pyscn.toml")

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

func TestAnalyzeCommandSelectValidation(t *testing.T) {
	analyzeCmd := NewAnalyzeCommand()
	analyzeCmd.selectAnalyses = []string{"complexity", "deps", "communities"}
	if err := analyzeCmd.validateSelectedAnalyses(); err != nil {
		t.Fatalf("expected valid selected analyses, got %v", err)
	}

	analyzeCmd.selectAnalyses = []string{"complexity", "nope"}
	if err := analyzeCmd.validateSelectedAnalyses(); err == nil {
		t.Fatal("expected invalid selected analysis to fail")
	}
}

func TestAnalyzeCommandSelectCommunitiesEnablesAnalysis(t *testing.T) {
	analyzeCmd := NewAnalyzeCommand()
	analyzeCmd.selectAnalyses = []string{"communities"}

	config := analyzeCmd.createUseCaseConfig()
	if config.SkipCommunities {
		t.Fatal("expected --select communities to enable community analysis")
	}
}

func TestAnalyzeCommandDefaultEnablesCommunities(t *testing.T) {
	analyzeCmd := NewAnalyzeCommand()

	config := analyzeCmd.createUseCaseConfig()
	if config.SkipCommunities {
		t.Fatal("expected default analyze to enable community analysis")
	}
}

func TestAnalyzeCommandSelectWithoutCommunitiesSkipsAnalysis(t *testing.T) {
	analyzeCmd := NewAnalyzeCommand()
	analyzeCmd.selectAnalyses = []string{"complexity", "deps"}

	config := analyzeCmd.createUseCaseConfig()
	if !config.SkipCommunities {
		t.Fatal("expected --select without communities to disable community analysis")
	}
}

func TestAnalyzeCommandSelectDepsCommunities(t *testing.T) {
	analyzeCmd := NewAnalyzeCommand()
	analyzeCmd.selectAnalyses = []string{"deps", "communities"}

	config := analyzeCmd.createUseCaseConfig()
	if config.SkipSystem {
		t.Fatal("expected --select deps to enable system analysis")
	}
	if config.SkipCommunities {
		t.Fatal("expected --select communities to enable community analysis")
	}
}

func TestAnalyzeCommandShouldWriteStandaloneCommunityJSON(t *testing.T) {
	analyzeCmd := NewAnalyzeCommand()
	analyzeCmd.json = true
	analyzeCmd.selectAnalyses = []string{"communities"}

	response := &domain.AnalyzeResponse{
		Communities: &domain.CommunityAnalysisResult{TotalCommunities: 2},
	}
	if !analyzeCmd.shouldWriteStandaloneCommunityJSON(response) {
		t.Fatal("expected standalone community JSON for --json --select communities")
	}

	analyzeCmd.selectAnalyses = []string{"deps", "communities"}
	if analyzeCmd.shouldWriteStandaloneCommunityJSON(response) {
		t.Fatal("expected unified analyze JSON when multiple analyses are selected")
	}

	analyzeCmd.selectAnalyses = []string{"communities"}
	analyzeCmd.json = false
	if analyzeCmd.shouldWriteStandaloneCommunityJSON(response) {
		t.Fatal("expected unified analyze output for non-JSON formats")
	}

	if analyzeCmd.shouldWriteStandaloneCommunityJSON(nil) {
		t.Fatal("expected false when response is nil")
	}
	if analyzeCmd.shouldWriteStandaloneCommunityJSON(&domain.AnalyzeResponse{}) {
		t.Fatal("expected false when community analysis is nil")
	}
}

func TestAnalyzeCommandSkipCommunitiesOverridesSelect(t *testing.T) {
	analyzeCmd := NewAnalyzeCommand()
	analyzeCmd.selectAnalyses = []string{"communities"}
	analyzeCmd.skipCommunities = true

	config := analyzeCmd.createUseCaseConfig()
	if !config.SkipCommunities {
		t.Fatal("expected --skip-communities to disable community analysis")
	}
	if !config.SkipCommunitiesExplicit {
		t.Fatal("expected skip-communities explicit flag to be set")
	}
}

func TestAnalyzeCommandSelectComposesWithSkipFlags(t *testing.T) {
	analyzeCmd := NewAnalyzeCommand()
	analyzeCmd.selectAnalyses = []string{"complexity", "clones"}
	analyzeCmd.skipClones = true

	config := analyzeCmd.createUseCaseConfig()
	if config.SkipComplexity {
		t.Fatal("expected selected complexity to remain enabled")
	}
	if !config.SkipClones {
		t.Fatal("expected skip-clones to disable selected clones")
	}
	if !config.SkipDeadCode {
		t.Fatal("expected unselected dead code to be disabled")
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

// TestCheckCircularDependencies tests circular dependency detection
func TestCheckCircularDependencies(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()

	var stdout, stderr bytes.Buffer
	cobraCmd.SetOut(&stdout)
	cobraCmd.SetErr(&stderr)

	// Test with circular dependencies
	cobraCmd.SetArgs([]string{"--select", "deps", filepath.Join("..", "..", "testdata", "python", "circular_deps_test")})

	err := cobraCmd.Execute()
	output := stdout.String() + stderr.String()

	if err == nil {
		t.Fatalf("Expected error because of circular dependency, got none, output: %s", output)
	}

	if !strings.Contains(output, "circular dependency detected") {
		t.Errorf("Expected circular dependency warning, got: %s", output)
	}
}

// TestCheckCircularDependenciesWithAllowFlag tests --allow-circular-deps flag
func TestCheckCircularDependenciesWithAllowFlag(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()

	var stdout, stderr bytes.Buffer
	cobraCmd.SetOut(&stdout)
	cobraCmd.SetErr(&stderr)

	// Test with --allow-circular-deps flag
	cobraCmd.SetArgs([]string{"--select", "deps", "--allow-circular-deps", filepath.Join("..", "..", "testdata", "python", "circular_deps_test")})

	err := cobraCmd.Execute()
	output := stdout.String() + stderr.String()

	if err != nil {
		t.Fatalf("Expected no error with --allow-circular-deps flag, got: %v, output: %s", err, output)
	}

	if !strings.Contains(output, "circular dependency") {
		t.Errorf("Expected circular dependency warning, got: %s", output)
	}

	if !strings.Contains(output, "Code quality check passed") {
		t.Errorf("Expected check to pass with --allow-circular-deps, got: %s", output)
	}
}

// TestCheckCircularDependenciesWithMaxCycles tests --max-cycles flag
func TestCheckCircularDependenciesWithMaxCycles(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()

	var stdout, stderr bytes.Buffer
	cobraCmd.SetOut(&stdout)
	cobraCmd.SetErr(&stderr)

	// Test with --max-cycles 1 (should pass)
	cobraCmd.SetArgs([]string{"--select", "deps", "--max-cycles", "1", filepath.Join("..", "..", "testdata", "python", "circular_deps_test")})

	err := cobraCmd.Execute()
	output := stdout.String() + stderr.String()

	if err != nil {
		t.Fatalf("Expected no error with --max-cycles 1, got: %v, output: %s", err, output)
	}

	if !strings.Contains(output, "within allowed limit") {
		t.Errorf("Expected 'within allowed limit' message, got: %s", output)
	}
}

func TestCheckCircularDependenciesUsesConfigPatterns(t *testing.T) {
	tempDir := t.TempDir()
	fixtureDir := filepath.Join("..", "..", "testdata", "python", "circular_deps_test")
	files := []string{"main.py", "player.py", "physics.py"}

	for _, name := range files {
		content, err := os.ReadFile(filepath.Join(fixtureDir, name))
		if err != nil {
			t.Fatalf("Failed to read fixture %s: %v", name, err)
		}
		if err := os.WriteFile(filepath.Join(tempDir, name), content, 0644); err != nil {
			t.Fatalf("Failed to write fixture %s: %v", name, err)
		}
	}

	configPath := filepath.Join(tempDir, ".pyscn.toml")
	configContent := `[analysis]
exclude_patterns = ["**/physics.py"]
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()

	var stdout, stderr bytes.Buffer
	cobraCmd.SetOut(&stdout)
	cobraCmd.SetErr(&stderr)
	cobraCmd.SetArgs([]string{"--select", "deps", "--config", configPath, tempDir})

	err := cobraCmd.Execute()
	output := stdout.String() + stderr.String()

	if err != nil {
		t.Fatalf("Expected deps check to respect config exclude pattern, got: %v, output: %s", err, output)
	}

	if strings.Contains(output, "circular dependency detected") {
		t.Fatalf("Expected no circular dependency output when physics.py is excluded, got: %s", output)
	}
}

func TestCheckMockdataUsesExplicitConfigPath(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	nestedDir := filepath.Join(projectDir, "nested")
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(projectDir, "main.py"), []byte("value = 42\n"), 0o644); err != nil {
		t.Fatalf("failed to write root source file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nestedDir, "data.py"), []byte("email = \"test@example.com\"\n"), 0o644); err != nil {
		t.Fatalf("failed to write nested source file: %v", err)
	}

	// The auto-discovered config enables recursion, while the explicitly selected
	// config disables it. The nested finding must therefore not be analyzed.
	if err := os.WriteFile(filepath.Join(projectDir, ".pyscn.toml"), []byte("[analysis]\nrecursive = true\n"), 0o644); err != nil {
		t.Fatalf("failed to write auto-discovered config: %v", err)
	}
	explicitConfigPath := filepath.Join(configDir, "explicit.toml")
	if err := os.WriteFile(explicitConfigPath, []byte("[analysis]\nrecursive = false\n"), 0o644); err != nil {
		t.Fatalf("failed to write explicit config: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Errorf("failed to restore current directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(projectDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()

	var stdout, stderr bytes.Buffer
	cobraCmd.SetOut(&stdout)
	cobraCmd.SetErr(&stderr)
	cobraCmd.SetArgs([]string{"--select", "mockdata", "--config", explicitConfigPath, projectDir})

	err = cobraCmd.Execute()
	if err != nil {
		output := stdout.String() + stderr.String()
		t.Fatalf("expected mockdata check to use the explicit non-recursive config, got: %v, output: %s", err, output)
	}
}

func TestCheckComplexityUsesConfigExcludePatterns(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	configDir := filepath.Join(tempDir, "config")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatalf("failed to create project directory: %v", err)
	}
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}

	complexSource := `def complex_function(value):
    result = 0
    if value > 0: result += 1
    if value > 1: result += 1
    if value > 2: result += 1
    if value > 3: result += 1
    if value > 4: result += 1
    if value > 5: result += 1
    if value > 6: result += 1
    if value > 7: result += 1
    if value > 8: result += 1
    if value > 9: result += 1
    if value > 10: result += 1
    return result
`
	if err := os.WriteFile(filepath.Join(projectDir, "complex.py"), []byte(complexSource), 0o644); err != nil {
		t.Fatalf("failed to write complex source file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "safe.py"), []byte("def safe():\n    return 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write safe source file: %v", err)
	}

	baseline := NewCheckCommand().CreateCobraCommand()
	baseline.SetOut(&bytes.Buffer{})
	baseline.SetErr(&bytes.Buffer{})
	baseline.SetArgs([]string{"--select", "complexity", projectDir})
	if err := baseline.Execute(); err == nil {
		t.Fatal("expected the unfiltered complexity fixture to fail the quality gate")
	}

	configPath := filepath.Join(configDir, "explicit.toml")
	if err := os.WriteFile(configPath, []byte("[analysis]\nexclude_patterns = [\"**/complex.py\"]\n"), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	checkCmd := NewCheckCommand().CreateCobraCommand()
	var output bytes.Buffer
	checkCmd.SetOut(&output)
	checkCmd.SetErr(&output)
	checkCmd.SetArgs([]string{"--select", "complexity", "--config", configPath, projectDir})
	if err := checkCmd.Execute(); err != nil {
		t.Fatalf("expected complexity check to honor config exclude_patterns, got: %v, output: %s", err, output.String())
	}
}

func TestCheckDeadCodeUsesConfigBooleanOverrides(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "dead.py")
	source := "def sample():\n    return 1\n    print('unreachable')\n"
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	baseline := NewCheckCommand().CreateCobraCommand()
	baseline.SetOut(&bytes.Buffer{})
	baseline.SetErr(&bytes.Buffer{})
	baseline.SetArgs([]string{"--select", "deadcode", sourcePath})
	if err := baseline.Execute(); err == nil {
		t.Fatal("expected the unreachable-code fixture to fail the quality gate")
	}

	configPath := filepath.Join(tempDir, "explicit.toml")
	config := "[dead_code]\ndetect_after_return = false\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	checkCmd := NewCheckCommand().CreateCobraCommand()
	var output bytes.Buffer
	checkCmd.SetOut(&output)
	checkCmd.SetErr(&output)
	checkCmd.SetArgs([]string{"--select", "deadcode", "--config", configPath, sourcePath})
	if err := checkCmd.Execute(); err != nil {
		t.Fatalf("expected dead-code check to honor detect_after_return=false, got: %v, output: %s", err, output.String())
	}
}

func TestCheckMockdataUsesConfiguredKeyword(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "data.py")
	if err := os.WriteFile(sourcePath, []byte("productionfixture = 42\n"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	baseline := NewCheckCommand().CreateCobraCommand()
	baseline.SetOut(&bytes.Buffer{})
	baseline.SetErr(&bytes.Buffer{})
	baseline.SetArgs([]string{"--select", "mockdata", sourcePath})
	if err := baseline.Execute(); err != nil {
		t.Fatalf("expected the custom keyword fixture to pass with defaults, got: %v", err)
	}

	configPath := filepath.Join(tempDir, "explicit.toml")
	config := "[mock_data]\nkeywords = [\"productionfixture\"]\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	checkCmd := NewCheckCommand().CreateCobraCommand()
	var output bytes.Buffer
	checkCmd.SetOut(&output)
	checkCmd.SetErr(&output)
	checkCmd.SetArgs([]string{"--select", "mockdata", "--config", configPath, sourcePath})
	if err := checkCmd.Execute(); err == nil {
		t.Fatalf("expected configured mockdata keyword to fail the quality gate, output: %s", output.String())
	}
	if !strings.Contains(output.String(), "productionfixture") {
		t.Fatalf("expected custom keyword finding in output, got: %s", output.String())
	}
}

func TestCheckMockdataUsesConfiguredIgnorePatterns(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "ignored.py")
	if err := os.WriteFile(sourcePath, []byte("email = \"test@example.com\"\n"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	baseline := NewCheckCommand().CreateCobraCommand()
	baseline.SetOut(&bytes.Buffer{})
	baseline.SetErr(&bytes.Buffer{})
	baseline.SetArgs([]string{"--select", "mockdata", sourcePath})
	if err := baseline.Execute(); err == nil {
		t.Fatal("expected the unignored mockdata fixture to fail the quality gate")
	}

	configPath := filepath.Join(tempDir, "explicit.toml")
	config := "[mock_data]\nignore_patterns = ['ignored\\.py$']\n"
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	checkCmd := NewCheckCommand().CreateCobraCommand()
	var output bytes.Buffer
	checkCmd.SetOut(&output)
	checkCmd.SetErr(&output)
	checkCmd.SetArgs([]string{"--select", "mockdata", "--config", configPath, sourcePath})
	if err := checkCmd.Execute(); err != nil {
		t.Fatalf("expected mockdata check to honor ignore_patterns, got: %v, output: %s", err, output.String())
	}
}

// TestCheckNoDepsAnalysis tests that deps analysis is opt-in by default
func TestCheckNoDepsAnalysis(t *testing.T) {
	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()

	var stdout, stderr bytes.Buffer
	cobraCmd.SetOut(&stdout)
	cobraCmd.SetErr(&stderr)

	// Run check without --select (should not include deps by default)
	cobraCmd.SetArgs([]string{filepath.Join("..", "..", "testdata", "python", "circular_deps_test")})

	_ = cobraCmd.Execute()
	output := stdout.String() + stderr.String()

	// Should not fail on circular deps (since deps analysis is not run)
	if strings.Contains(output, "circular dependency") {
		t.Errorf("Deps analysis should not run by default, but got: %s", output)
	}
}

func TestCheckDIRespectsConfigThreshold(t *testing.T) {
	tempDir := t.TempDir()

	sourcePath := filepath.Join(tempDir, "service.py")
	source := `class Service:
    def __init__(self, a, b, c, d, e, f):
        pass
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	configPath := filepath.Join(tempDir, ".pyscn.toml")
	config := `[di]
constructor_param_threshold = 10
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()

	var stdout, stderr bytes.Buffer
	cobraCmd.SetOut(&stdout)
	cobraCmd.SetErr(&stderr)
	cobraCmd.SetArgs([]string{"--select", "di", "--config", configPath, tempDir})

	err := cobraCmd.Execute()
	output := stdout.String() + stderr.String()

	if err != nil {
		t.Fatalf("expected no error when DI threshold comes from config, got: %v, output: %s", err, output)
	}
}

func TestCheckDILoadsDefaultConfigFromParentDirectory(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	subDir := filepath.Join(projectDir, "pkg")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}

	configPath := filepath.Join(projectDir, ".pyscn.toml")
	config := `[di]
constructor_param_threshold = 10
`
	if err := os.WriteFile(configPath, []byte(config), 0o644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	sourcePath := filepath.Join(subDir, "service.py")
	source := `class Service:
    def __init__(self, a, b, c, d, e, f):
        pass
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Errorf("failed to restore current directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()

	var stdout, stderr bytes.Buffer
	cobraCmd.SetOut(&stdout)
	cobraCmd.SetErr(&stderr)
	cobraCmd.SetArgs([]string{"--select", "di", "."})

	err = cobraCmd.Execute()
	output := stdout.String() + stderr.String()

	if err != nil {
		t.Fatalf("expected parent config to be discovered automatically, got: %v, output: %s", err, output)
	}
	if strings.Contains(output, "constructor_over_injection") {
		t.Fatalf("expected parent config threshold to suppress the finding, output: %s", output)
	}
}

func TestCheckFailsOnMalformedDiscoveredPyscnConfig(t *testing.T) {
	tempDir := t.TempDir()
	projectDir := filepath.Join(tempDir, "project")
	subDir := filepath.Join(projectDir, "pkg")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create test directories: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, ".pyscn.toml"), []byte("[analysis\nrecursive = false\n"), 0o644); err != nil {
		t.Fatalf("failed to write malformed config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "sample.py"), []byte("def sample():\n    return 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Errorf("failed to restore current directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	cobraCmd.SetArgs([]string{"--select", "complexity", "."})

	err = cobraCmd.Execute()
	if err == nil {
		t.Fatal("expected malformed discovered .pyscn.toml to fail the check")
	}
	if !strings.Contains(err.Error(), "failed to load configuration") {
		t.Fatalf("expected configuration load error, got: %v", err)
	}
}

func TestCheckIgnoresMalformedPyprojectWithoutPyscnSection(t *testing.T) {
	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "pyproject.toml"), []byte("[project\nname = 'broken'\n"), 0o644); err != nil {
		t.Fatalf("failed to write unrelated pyproject.toml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "sample.py"), []byte("def sample():\n    return 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Errorf("failed to restore current directory: %v", chdirErr)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to change directory: %v", err)
	}

	checkCmd := NewCheckCommand()
	cobraCmd := checkCmd.CreateCobraCommand()
	cobraCmd.SetArgs([]string{"--select", "complexity", "."})

	if err := cobraCmd.Execute(); err != nil {
		t.Fatalf("unrelated pyproject.toml should not affect check: %v", err)
	}
}

func TestCountDIAntipatternIssuesFailsOnAnalysisErrors(t *testing.T) {
	checkCmd := NewCheckCommand()
	response := &domain.DIAntipatternResponse{
		Errors: []string{"[broken.py] Parse error: syntax errors found in source code"},
	}

	var stderr bytes.Buffer
	_, err := checkCmd.countDIAntipatternIssues(&stderr, response)
	if err == nil {
		t.Fatal("expected DI analysis errors to fail the check")
	}
	if !strings.Contains(err.Error(), "Parse error") {
		t.Fatalf("expected parse error to be preserved, got: %v", err)
	}
}

// TestAnalyzeCommandThresholdFlags verifies that complexity threshold flags
// on the analyze command are mapped into AnalyzeUseCaseConfig. This is the CLI
// counterpart to the MergeConfig fix for issue #553.
func TestAnalyzeCommandThresholdFlags(t *testing.T) {
	t.Run("default values are zero (unset)", func(t *testing.T) {
		analyzeCmd := NewAnalyzeCommand()
		config := analyzeCmd.createUseCaseConfig()

		if config.LowThreshold != 0 {
			t.Errorf("expected LowThreshold 0, got %d", config.LowThreshold)
		}
		if config.MediumThreshold != 0 {
			t.Errorf("expected MediumThreshold 0, got %d", config.MediumThreshold)
		}
		if config.CognitiveComplexityThreshold != 0 {
			t.Errorf("expected CognitiveComplexityThreshold 0, got %d", config.CognitiveComplexityThreshold)
		}
		if config.NestingDepthThreshold != 0 {
			t.Errorf("expected NestingDepthThreshold 0, got %d", config.NestingDepthThreshold)
		}
	})

	t.Run("explicit flag values are mapped", func(t *testing.T) {
		analyzeCmd := NewAnalyzeCommand()
		analyzeCmd.lowThreshold = 9
		analyzeCmd.mediumThreshold = 19
		analyzeCmd.cognitiveComplexityThreshold = 25
		analyzeCmd.nestingDepthThreshold = 7

		config := analyzeCmd.createUseCaseConfig()

		if config.LowThreshold != 9 {
			t.Errorf("expected LowThreshold 9, got %d", config.LowThreshold)
		}
		if config.MediumThreshold != 19 {
			t.Errorf("expected MediumThreshold 19, got %d", config.MediumThreshold)
		}
		if config.CognitiveComplexityThreshold != 25 {
			t.Errorf("expected CognitiveComplexityThreshold 25, got %d", config.CognitiveComplexityThreshold)
		}
		if config.NestingDepthThreshold != 7 {
			t.Errorf("expected NestingDepthThreshold 7, got %d", config.NestingDepthThreshold)
		}
	})

	t.Run("flags are registered on cobra command", func(t *testing.T) {
		analyzeCmd := NewAnalyzeCommand()
		cobraCmd := analyzeCmd.CreateCobraCommand()

		expectedFlags := []string{
			"low-threshold",
			"medium-threshold",
			"cognitive-complexity-threshold",
			"nesting-depth-threshold",
		}
		for _, name := range expectedFlags {
			if cobraCmd.Flags().Lookup(name) == nil {
				t.Errorf("expected flag --%s to be registered", name)
			}
		}
	})
}
