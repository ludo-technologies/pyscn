package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadComplexityFromPyprojectToml(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create pyproject.toml with complexity settings
	configContent := `[tool.pyscn.complexity]
low_threshold = 4
medium_threshold = 6
max_complexity = 10
min_complexity = 2
`
	configPath := filepath.Join(tempDir, "pyproject.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	config, err := LoadPyprojectConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify complexity settings were loaded
	if config.ComplexityLowThreshold != 4 {
		t.Errorf("Expected low_threshold 4, got %d", config.ComplexityLowThreshold)
	}
	if config.ComplexityMediumThreshold != 6 {
		t.Errorf("Expected medium_threshold 6, got %d", config.ComplexityMediumThreshold)
	}
	if config.ComplexityMaxComplexity != 10 {
		t.Errorf("Expected max_complexity 10, got %d", config.ComplexityMaxComplexity)
	}
	if config.ComplexityMinComplexity != 2 {
		t.Errorf("Expected min_complexity 2, got %d", config.ComplexityMinComplexity)
	}
}

func TestLoadComplexityAndClonesFromPyprojectToml(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create pyproject.toml with both complexity and clones settings
	configContent := `[tool.pyscn.complexity]
low_threshold = 3
medium_threshold = 5

[tool.pyscn.clones]
min_lines = 10
min_nodes = 20
`
	configPath := filepath.Join(tempDir, "pyproject.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load config
	config, err := LoadPyprojectConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify complexity settings were loaded
	if config.ComplexityLowThreshold != 3 {
		t.Errorf("Expected low_threshold 3, got %d", config.ComplexityLowThreshold)
	}
	if config.ComplexityMediumThreshold != 5 {
		t.Errorf("Expected medium_threshold 5, got %d", config.ComplexityMediumThreshold)
	}

	// Verify clones settings were also loaded
	if config.Analysis.MinLines != 10 {
		t.Errorf("Expected min_lines 10, got %d", config.Analysis.MinLines)
	}
	if config.Analysis.MinNodes != 20 {
		t.Errorf("Expected min_nodes 20, got %d", config.Analysis.MinNodes)
	}
}

func TestLoadArchitectureLayersAndRulesFromPyprojectToml(t *testing.T) {
	tempDir := t.TempDir()

	configContent := `[tool.pyscn.architecture]
enabled = true

[[tool.pyscn.architecture.layers]]
name = "api"
description = "API layer"
packages = ["myapp.api"]

[[tool.pyscn.architecture.layers]]
name = "domain"
description = "Domain layer"
packages = ["myapp.domain"]

[[tool.pyscn.architecture.rules]]
from = "api"
allow = ["domain"]
deny = ["infrastructure"]
`
	configPath := filepath.Join(tempDir, "pyproject.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config, err := LoadPyprojectConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if len(config.ArchitectureLayers) != 2 {
		t.Fatalf("Expected 2 layers, got %d", len(config.ArchitectureLayers))
	}
	if config.ArchitectureLayers[0].Name != "api" {
		t.Errorf("Expected layer[0].Name 'api', got %q", config.ArchitectureLayers[0].Name)
	}
	if config.ArchitectureLayers[1].Name != "domain" {
		t.Errorf("Expected layer[1].Name 'domain', got %q", config.ArchitectureLayers[1].Name)
	}
	if len(config.ArchitectureLayers[0].Packages) != 1 || config.ArchitectureLayers[0].Packages[0] != "myapp.api" {
		t.Errorf("Expected layer[0].Packages ['myapp.api'], got %v", config.ArchitectureLayers[0].Packages)
	}

	if len(config.ArchitectureRules) != 1 {
		t.Fatalf("Expected 1 rule, got %d", len(config.ArchitectureRules))
	}
	if config.ArchitectureRules[0].From != "api" {
		t.Errorf("Expected rule[0].From 'api', got %q", config.ArchitectureRules[0].From)
	}
	if len(config.ArchitectureRules[0].Allow) != 1 || config.ArchitectureRules[0].Allow[0] != "domain" {
		t.Errorf("Expected rule[0].Allow ['domain'], got %v", config.ArchitectureRules[0].Allow)
	}
	if len(config.ArchitectureRules[0].Deny) != 1 || config.ArchitectureRules[0].Deny[0] != "infrastructure" {
		t.Errorf("Expected rule[0].Deny ['infrastructure'], got %v", config.ArchitectureRules[0].Deny)
	}
}

func TestPyprojectTomlPriority(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create both .pyscn.toml and pyproject.toml
	// .pyscn.toml should have priority
	pyscnContent := `[complexity]
low_threshold = 5
`
	pyprojectContent := `[tool.pyscn.complexity]
low_threshold = 10
`

	pyscnPath := filepath.Join(tempDir, ".pyscn.toml")
	if err := os.WriteFile(pyscnPath, []byte(pyscnContent), 0644); err != nil {
		t.Fatalf("Failed to write .pyscn.toml: %v", err)
	}

	pyprojectPath := filepath.Join(tempDir, "pyproject.toml")
	if err := os.WriteFile(pyprojectPath, []byte(pyprojectContent), 0644); err != nil {
		t.Fatalf("Failed to write pyproject.toml: %v", err)
	}

	// Load config using TOML loader (which checks .pyscn.toml first)
	loader := NewTomlConfigLoader()
	config, err := loader.LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify .pyscn.toml was used (priority)
	if config.ComplexityLowThreshold != 5 {
		t.Errorf("Expected .pyscn.toml value 5, got %d (pyproject.toml should be ignored)", config.ComplexityLowThreshold)
	}
}
