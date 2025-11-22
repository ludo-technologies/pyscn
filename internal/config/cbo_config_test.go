package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCBOFromPyscnToml(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create .pyscn.toml with cbo settings
	configContent := `[cbo]
low_threshold = 5
medium_threshold = 10
min_cbo = 2
max_cbo = 20
show_zeros = true
include_builtins = true
include_imports = false
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
	t.Logf("  CboLowThreshold: %d", config.CboLowThreshold)
	t.Logf("  CboMediumThreshold: %d", config.CboMediumThreshold)
	t.Logf("  CboMinCbo: %d", config.CboMinCbo)
	t.Logf("  CboMaxCbo: %d", config.CboMaxCbo)
	t.Logf("  CboShowZeros: %v", config.CboShowZeros)
	t.Logf("  CboIncludeBuiltins: %v", config.CboIncludeBuiltins)
	t.Logf("  CboIncludeImports: %v", config.CboIncludeImports)

	// Verify cbo settings were loaded
	if config.CboLowThreshold != 5 {
		t.Errorf("Expected low_threshold 5, got %d", config.CboLowThreshold)
	}
	if config.CboMediumThreshold != 10 {
		t.Errorf("Expected medium_threshold 10, got %d", config.CboMediumThreshold)
	}
	if config.CboMinCbo != 2 {
		t.Errorf("Expected min_cbo 2, got %d", config.CboMinCbo)
	}
	if config.CboMaxCbo != 20 {
		t.Errorf("Expected max_cbo 20, got %d", config.CboMaxCbo)
	}
	if !config.CboShowZeros {
		t.Errorf("Expected show_zeros true, got %v", config.CboShowZeros)
	}
	if !config.CboIncludeBuiltins {
		t.Errorf("Expected include_builtins true, got %v", config.CboIncludeBuiltins)
	}
	if config.CboIncludeImports {
		t.Errorf("Expected include_imports false, got %v", config.CboIncludeImports)
	}
}

func TestLoadCBOFromPyscnTomlPartial(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create .pyscn.toml with only some cbo settings
	configContent := `[cbo]
low_threshold = 4
medium_threshold = 8
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
	if config.CboLowThreshold != 4 {
		t.Errorf("Expected low_threshold 4, got %d", config.CboLowThreshold)
	}
	if config.CboMediumThreshold != 8 {
		t.Errorf("Expected medium_threshold 8, got %d", config.CboMediumThreshold)
	}

	// Verify unspecified settings use defaults
	if config.CboMinCbo != 0 {
		t.Errorf("Expected default min_cbo 0, got %d", config.CboMinCbo)
	}
	if config.CboMaxCbo != 0 {
		t.Errorf("Expected default max_cbo 0, got %d", config.CboMaxCbo)
	}
	if config.CboShowZeros {
		t.Errorf("Expected default show_zeros false, got %v", config.CboShowZeros)
	}
	if config.CboIncludeBuiltins {
		t.Errorf("Expected default include_builtins false, got %v", config.CboIncludeBuiltins)
	}
	if !config.CboIncludeImports {
		t.Errorf("Expected default include_imports true, got %v", config.CboIncludeImports)
	}
}

func TestMergeCboSection(t *testing.T) {
	// Create a default config
	config := DefaultPyscnConfig()

	// Create cbo settings
	cbo := CboTomlConfig{
		LowThreshold:    intPtr(4),
		MediumThreshold: intPtr(9),
		MinCbo:          intPtr(1),
		MaxCbo:          intPtr(15),
		ShowZeros:       boolPtr(true),
		IncludeBuiltins: boolPtr(true),
		IncludeImports:  boolPtr(false),
	}

	// Merge cbo settings
	mergeCboSection(config, &cbo)

	// Verify settings were merged
	if config.CboLowThreshold != 4 {
		t.Errorf("Expected low_threshold 4, got %d", config.CboLowThreshold)
	}
	if config.CboMediumThreshold != 9 {
		t.Errorf("Expected medium_threshold 9, got %d", config.CboMediumThreshold)
	}
	if config.CboMinCbo != 1 {
		t.Errorf("Expected min_cbo 1, got %d", config.CboMinCbo)
	}
	if config.CboMaxCbo != 15 {
		t.Errorf("Expected max_cbo 15, got %d", config.CboMaxCbo)
	}
	if !config.CboShowZeros {
		t.Errorf("Expected show_zeros true, got %v", config.CboShowZeros)
	}
	if !config.CboIncludeBuiltins {
		t.Errorf("Expected include_builtins true, got %v", config.CboIncludeBuiltins)
	}
	if config.CboIncludeImports {
		t.Errorf("Expected include_imports false, got %v", config.CboIncludeImports)
	}
}

func TestMergeCboSectionNilValues(t *testing.T) {
	// Create a default config
	config := DefaultPyscnConfig()
	originalLow := config.CboLowThreshold

	// Create cbo settings with nil values
	cbo := CboTomlConfig{
		LowThreshold:    nil,
		MediumThreshold: nil,
		MinCbo:          nil,
		MaxCbo:          nil,
		ShowZeros:       nil,
		IncludeBuiltins: nil,
		IncludeImports:  nil,
	}

	// Merge cbo settings
	mergeCboSection(config, &cbo)

	// Verify defaults were not changed
	if config.CboLowThreshold != originalLow {
		t.Errorf("Expected defaults to remain, got %d", config.CboLowThreshold)
	}
}

// Helper function to create bool pointer
func boolPtr(val bool) *bool {
	return &val
}
