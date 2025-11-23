package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestCBOConfigurationLoader_LoadConfig(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create .pyscn.toml with CBO settings
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
	loader := NewCBOConfigurationLoader()
	req, err := loader.LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify CBO settings were loaded and converted correctly
	if req.LowThreshold != 5 {
		t.Errorf("Expected LowThreshold 5, got %d", req.LowThreshold)
	}
	if req.MediumThreshold != 10 {
		t.Errorf("Expected MediumThreshold 10, got %d", req.MediumThreshold)
	}
	if req.MinCBO != 2 {
		t.Errorf("Expected MinCBO 2, got %d", req.MinCBO)
	}
	if req.MaxCBO != 20 {
		t.Errorf("Expected MaxCBO 20, got %d", req.MaxCBO)
	}
	if !domain.BoolValue(req.ShowZeros, false) {
		t.Errorf("Expected ShowZeros true, got %v", req.ShowZeros)
	}
	if !domain.BoolValue(req.IncludeBuiltins, false) {
		t.Errorf("Expected IncludeBuiltins true, got %v", req.IncludeBuiltins)
	}
	if domain.BoolValue(req.IncludeImports, true) {
		t.Errorf("Expected IncludeImports false, got %v", req.IncludeImports)
	}
}

func TestCBOConfigurationLoader_LoadDefaultConfig(t *testing.T) {
	loader := NewCBOConfigurationLoader()
	req := loader.LoadDefaultConfig()

	if req == nil {
		t.Fatal("Expected non-nil default config")
	}

	// Verify default values
	if req.LowThreshold != 3 {
		t.Errorf("Expected default LowThreshold 3, got %d", req.LowThreshold)
	}
	if req.MediumThreshold != 7 {
		t.Errorf("Expected default MediumThreshold 7, got %d", req.MediumThreshold)
	}
	if req.MinCBO != 0 {
		t.Errorf("Expected default MinCBO 0, got %d", req.MinCBO)
	}
	if req.MaxCBO != 0 {
		t.Errorf("Expected default MaxCBO 0 (no limit), got %d", req.MaxCBO)
	}
	if domain.BoolValue(req.ShowZeros, false) {
		t.Errorf("Expected default ShowZeros false, got %v", req.ShowZeros)
	}
	if domain.BoolValue(req.IncludeBuiltins, false) {
		t.Errorf("Expected default IncludeBuiltins false, got %v", req.IncludeBuiltins)
	}
	if !domain.BoolValue(req.IncludeImports, true) {
		t.Errorf("Expected default IncludeImports true, got %v", req.IncludeImports)
	}
}

func TestCBOConfigurationLoader_MergeConfig(t *testing.T) {
	loader := NewCBOConfigurationLoader()

	tests := []struct {
		name     string
		base     *domain.CBORequest
		override *domain.CBORequest
		expected *domain.CBORequest
	}{
		{
			name: "override thresholds",
			base: &domain.CBORequest{
				LowThreshold:    3,
				MediumThreshold: 7,
				MinCBO:          0,
				MaxCBO:          0,
			},
			override: &domain.CBORequest{
				LowThreshold:    5,
				MediumThreshold: 10,
				MinCBO:          2,
				MaxCBO:          20,
			},
			expected: &domain.CBORequest{
				LowThreshold:    5,
				MediumThreshold: 10,
				MinCBO:          2,
				MaxCBO:          20,
			},
		},
		{
			name: "override boolean flags",
			base: &domain.CBORequest{
				ShowZeros:       domain.BoolPtr(false),
				IncludeBuiltins: domain.BoolPtr(false),
				IncludeImports:  domain.BoolPtr(true),
			},
			override: &domain.CBORequest{
				ShowZeros:       domain.BoolPtr(true),
				IncludeBuiltins: domain.BoolPtr(true),
				IncludeImports:  domain.BoolPtr(false),
			},
			expected: &domain.CBORequest{
				ShowZeros:       domain.BoolPtr(true),
				IncludeBuiltins: domain.BoolPtr(true),
				IncludeImports:  domain.BoolPtr(false),
			},
		},
		{
			name: "paths always from override",
			base: &domain.CBORequest{
				Paths: []string{"/base/path"},
			},
			override: &domain.CBORequest{
				Paths: []string{"/override/path"},
			},
			expected: &domain.CBORequest{
				Paths: []string{"/override/path"},
			},
		},
		{
			name:     "nil base returns override",
			base:     nil,
			override: &domain.CBORequest{LowThreshold: 5},
			expected: &domain.CBORequest{LowThreshold: 5},
		},
		{
			name:     "nil override returns base",
			base:     &domain.CBORequest{LowThreshold: 3},
			override: nil,
			expected: &domain.CBORequest{LowThreshold: 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loader.MergeConfig(tt.base, tt.override)

			if tt.base == nil || tt.override == nil {
				// For nil cases, just check the result matches expected
				if result.LowThreshold != tt.expected.LowThreshold {
					t.Errorf("Expected LowThreshold %d, got %d", tt.expected.LowThreshold, result.LowThreshold)
				}
				return
			}

			// Check each field that was set in expected
			if tt.expected.LowThreshold != 0 && result.LowThreshold != tt.expected.LowThreshold {
				t.Errorf("Expected LowThreshold %d, got %d", tt.expected.LowThreshold, result.LowThreshold)
			}
			if tt.expected.MediumThreshold != 0 && result.MediumThreshold != tt.expected.MediumThreshold {
				t.Errorf("Expected MediumThreshold %d, got %d", tt.expected.MediumThreshold, result.MediumThreshold)
			}
			if tt.expected.MinCBO != 0 && result.MinCBO != tt.expected.MinCBO {
				t.Errorf("Expected MinCBO %d, got %d", tt.expected.MinCBO, result.MinCBO)
			}
			if tt.expected.MaxCBO != 0 && result.MaxCBO != tt.expected.MaxCBO {
				t.Errorf("Expected MaxCBO %d, got %d", tt.expected.MaxCBO, result.MaxCBO)
			}
			if len(tt.expected.Paths) > 0 {
				if len(result.Paths) != len(tt.expected.Paths) {
					t.Errorf("Expected %d paths, got %d", len(tt.expected.Paths), len(result.Paths))
				}
			}
		})
	}
}

func TestCBOConfigurationLoader_FindDefaultConfigFile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	loader := NewCBOConfigurationLoader()

	// Test: No config file
	result := loader.FindDefaultConfigFile()
	// Should return empty string if no config found in current directory
	// (This is expected behavior, not an error)
	t.Logf("No config file found (expected): %s", result)

	// Test: Create .pyscn.toml and find it
	configPath := filepath.Join(tempDir, ".pyscn.toml")
	if err := os.WriteFile(configPath, []byte("[cbo]\n"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	result = loader.FindDefaultConfigFile()
	if result == "" {
		t.Error("Expected to find .pyscn.toml")
	}
	t.Logf("Found config file: %s", result)
}
