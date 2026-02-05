package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestLoadComplexityFromPyscnToml(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create .pyscn.toml with complexity settings
	configContent := `[complexity]
low_threshold = 5
medium_threshold = 7
max_complexity = 9
min_complexity = 3
`
	configPath := filepath.Join(tempDir, ".pyscn.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	loader := NewTomlConfigLoader()
	config, err := loader.LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify complexity settings were loaded
	if config.ComplexityLowThreshold != 5 {
		t.Errorf("Expected low_threshold 5, got %d", config.ComplexityLowThreshold)
	}
	if config.ComplexityMediumThreshold != 7 {
		t.Errorf("Expected medium_threshold 7, got %d", config.ComplexityMediumThreshold)
	}
	if config.ComplexityMaxComplexity != 9 {
		t.Errorf("Expected max_complexity 9, got %d", config.ComplexityMaxComplexity)
	}
	if config.ComplexityMinComplexity != 3 {
		t.Errorf("Expected min_complexity 3, got %d", config.ComplexityMinComplexity)
	}
}

func TestLoadComplexityFromPyscnTomlPartial(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create .pyscn.toml with only some complexity settings
	configContent := `[complexity]
low_threshold = 4
medium_threshold = 6
`
	configPath := filepath.Join(tempDir, ".pyscn.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	loader := NewTomlConfigLoader()
	config, err := loader.LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify specified settings were loaded
	if config.ComplexityLowThreshold != 4 {
		t.Errorf("Expected low_threshold 4, got %d", config.ComplexityLowThreshold)
	}
	if config.ComplexityMediumThreshold != 6 {
		t.Errorf("Expected medium_threshold 6, got %d", config.ComplexityMediumThreshold)
	}

	// Verify unspecified settings use defaults
	if config.ComplexityMaxComplexity != DefaultMaxComplexityLimit {
		t.Errorf("Expected default max_complexity %d, got %d", DefaultMaxComplexityLimit, config.ComplexityMaxComplexity)
	}
	if config.ComplexityMinComplexity != DefaultMinComplexityFilter {
		t.Errorf("Expected default min_complexity %d, got %d", DefaultMinComplexityFilter, config.ComplexityMinComplexity)
	}
}

func TestMergeComplexitySection(t *testing.T) {
	// Create a default config
	config := DefaultPyscnConfig()

	// Create complexity settings
	complexity := ComplexityTomlConfig{
		LowThreshold:    intPtr(3),
		MediumThreshold: intPtr(5),
		MaxComplexity:   intPtr(10),
		MinComplexity:   intPtr(2),
	}

	// Merge complexity settings
	mergeComplexitySection(config, &complexity)

	// Verify settings were merged
	if config.ComplexityLowThreshold != 3 {
		t.Errorf("Expected low_threshold 3, got %d", config.ComplexityLowThreshold)
	}
	if config.ComplexityMediumThreshold != 5 {
		t.Errorf("Expected medium_threshold 5, got %d", config.ComplexityMediumThreshold)
	}
	if config.ComplexityMaxComplexity != 10 {
		t.Errorf("Expected max_complexity 10, got %d", config.ComplexityMaxComplexity)
	}
	if config.ComplexityMinComplexity != 2 {
		t.Errorf("Expected min_complexity 2, got %d", config.ComplexityMinComplexity)
	}
}

func TestMergeComplexitySectionNilValues(t *testing.T) {
	// Create a default config
	config := DefaultPyscnConfig()
	originalLow := config.ComplexityLowThreshold

	// Create complexity settings with nil values
	complexity := ComplexityTomlConfig{
		LowThreshold:    nil,
		MediumThreshold: nil,
		MaxComplexity:   nil,
		MinComplexity:   nil,
	}

	// Merge complexity settings
	mergeComplexitySection(config, &complexity)

	// Verify defaults were not changed
	if config.ComplexityLowThreshold != originalLow {
		t.Errorf("Expected defaults to remain, got %d", config.ComplexityLowThreshold)
	}
}

// Helper function to create int pointer
func intPtr(val int) *int {
	return &val
}

func TestLoadDeadCodeFromPyscnToml(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create .pyscn.toml with dead_code settings
	configContent := `[dead_code]
min_severity = "info"
show_context = true
context_lines = 5
sort_by = "line"
detect_after_return = false
detect_after_break = false
`
	configPath := filepath.Join(tempDir, ".pyscn.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	loader := NewTomlConfigLoader()
	config, err := loader.LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Debug: print values
	t.Logf("Loaded PyscnConfig:")
	t.Logf("  DeadCodeMinSeverity: %s", config.DeadCodeMinSeverity)
	t.Logf("  DeadCodeShowContext: %v", config.DeadCodeShowContext)
	t.Logf("  DeadCodeContextLines: %d", config.DeadCodeContextLines)
	t.Logf("  DeadCodeSortBy: %s", config.DeadCodeSortBy)

	// Verify dead_code settings were loaded
	if config.DeadCodeMinSeverity != "info" {
		t.Errorf("Expected min_severity 'info', got %s", config.DeadCodeMinSeverity)
	}
	if !domain.BoolValue(config.DeadCodeShowContext, false) {
		t.Errorf("Expected show_context true, got %v", config.DeadCodeShowContext)
	}
	if config.DeadCodeContextLines != 5 {
		t.Errorf("Expected context_lines 5, got %d", config.DeadCodeContextLines)
	}
	if config.DeadCodeSortBy != "line" {
		t.Errorf("Expected sort_by 'line', got %s", config.DeadCodeSortBy)
	}
	if domain.BoolValue(config.DeadCodeDetectAfterReturn, true) {
		t.Errorf("Expected detect_after_return false, got %v", config.DeadCodeDetectAfterReturn)
	}
	if domain.BoolValue(config.DeadCodeDetectAfterBreak, true) {
		t.Errorf("Expected detect_after_break false, got %v", config.DeadCodeDetectAfterBreak)
	}
}

func TestLoadConfig_DirectPyprojectPathIgnoresSiblingPyscn(t *testing.T) {
	tempDir := t.TempDir()

	pyscnContent := `[analysis]
include_patterns = ["from-pyscn"]
`
	pyprojectContent := `[tool.pyscn.analysis]
include_patterns = ["from-pyproject"]
`

	if err := os.WriteFile(filepath.Join(tempDir, ".pyscn.toml"), []byte(pyscnContent), 0644); err != nil {
		t.Fatalf("Failed to write .pyscn.toml: %v", err)
	}
	pyprojectPath := filepath.Join(tempDir, "pyproject.toml")
	if err := os.WriteFile(pyprojectPath, []byte(pyprojectContent), 0644); err != nil {
		t.Fatalf("Failed to write pyproject.toml: %v", err)
	}

	loader := NewTomlConfigLoader()
	cfg, err := loader.LoadConfig(pyprojectPath)
	if err != nil {
		t.Fatalf("Failed to load explicit pyproject.toml: %v", err)
	}

	if len(cfg.AnalysisIncludePatterns) != 1 || cfg.AnalysisIncludePatterns[0] != "from-pyproject" {
		t.Fatalf("Expected include pattern from explicit pyproject.toml, got %v", cfg.AnalysisIncludePatterns)
	}
}

func TestFindConfigFileFromPath_DetectsQuotedPyscnSection(t *testing.T) {
	tempDir := t.TempDir()
	pyprojectPath := filepath.Join(tempDir, "pyproject.toml")
	content := `[tool."pyscn".analysis]
include_patterns = ["quoted"]
`
	if err := os.WriteFile(pyprojectPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write pyproject.toml: %v", err)
	}

	loader := NewTomlConfigLoader()
	found := loader.FindConfigFileFromPath(tempDir)
	if found != pyprojectPath {
		t.Fatalf("Expected %s, got %s", pyprojectPath, found)
	}
}

func TestResolveConfigPath_MissingTomlReturnsError(t *testing.T) {
	loader := NewTomlConfigLoader()

	_, err := loader.ResolveConfigPath("/nonexistent/.pyscn.toml", "")
	if err == nil {
		t.Fatal("Expected error for missing explicit TOML file")
	}
	if !strings.Contains(err.Error(), "config file not found") {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestResolveConfigPath_MissingNonTomlPathReturnsError(t *testing.T) {
	loader := NewTomlConfigLoader()

	_, err := loader.ResolveConfigPath("/nonexistent/config-dir", "")
	if err == nil {
		t.Fatal("Expected error for missing explicit config path")
	}
	if !strings.Contains(err.Error(), "config file not found") {
		t.Fatalf("Unexpected error: %v", err)
	}
}
