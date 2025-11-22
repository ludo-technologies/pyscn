package config

import (
	"os"
	"path/filepath"
	"testing"
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
	config := DefaultCloneConfig()

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
	config := DefaultCloneConfig()
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
