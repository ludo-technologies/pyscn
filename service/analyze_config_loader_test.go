package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestAnalyzeConfigurationLoader_LoadAnalyzeExecutionConfig(t *testing.T) {
	loader := NewAnalyzeConfigurationLoader()

	t.Run("uses analyze defaults when no config file is found", func(t *testing.T) {
		cfg, err := loader.LoadAnalyzeExecutionConfig("", t.TempDir())
		if err != nil {
			t.Fatalf("LoadAnalyzeExecutionConfig returned error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected non-nil execution config")
		}

		if cfg.ConfigPath != "" {
			t.Errorf("expected empty config path, got %q", cfg.ConfigPath)
		}
		if len(cfg.IncludePatterns) != 2 || cfg.IncludePatterns[1] != "*.pyi" {
			t.Errorf("expected default include patterns to include .pyi files, got %v", cfg.IncludePatterns)
		}
		if !cfg.ComplexityEnabled {
			t.Error("expected complexity enabled by default")
		}
		if !cfg.ComplexityReportUnchanged {
			t.Error("expected report_unchanged enabled by default")
		}
		defaultCloneReq := domain.DefaultCloneRequest()
		if cfg.CloneLSHEnabled != defaultCloneReq.LSHEnabled {
			t.Errorf("expected default LSH enabled %q, got %q", defaultCloneReq.LSHEnabled, cfg.CloneLSHEnabled)
		}
		if cfg.CloneLSHAutoThreshold != defaultCloneReq.LSHAutoThreshold {
			t.Errorf("expected default LSH threshold %d, got %d", defaultCloneReq.LSHAutoThreshold, cfg.CloneLSHAutoThreshold)
		}
	})

	t.Run("resolves config from target path and loads analyze settings", func(t *testing.T) {
		projectDir := t.TempDir()
		targetDir := filepath.Join(projectDir, "pkg")
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			t.Fatalf("failed to create target dir: %v", err)
		}

		configPath := filepath.Join(projectDir, ".pyscn.toml")
		configContent := `[analysis]
include_patterns = ["pkg/**/*.py"]
exclude_patterns = ["tests/**/*.py"]
recursive = false

[complexity]
enabled = false
report_unchanged = false
low_threshold = 3
medium_threshold = 7
max_complexity = 11

[output]
min_complexity = 9

[clones]
lsh_enabled = "true"
lsh_auto_threshold = 123
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to write config file: %v", err)
		}

		cfg, err := loader.LoadAnalyzeExecutionConfig("", targetDir)
		if err != nil {
			t.Fatalf("LoadAnalyzeExecutionConfig returned error: %v", err)
		}
		if cfg == nil {
			t.Fatal("expected non-nil execution config")
		}

		if cfg.ConfigPath != configPath {
			t.Errorf("expected config path %q, got %q", configPath, cfg.ConfigPath)
		}
		if cfg.Recursive {
			t.Error("expected recursive false")
		}
		if len(cfg.IncludePatterns) != 1 || cfg.IncludePatterns[0] != "pkg/**/*.py" {
			t.Errorf("expected custom include patterns, got %v", cfg.IncludePatterns)
		}
		if len(cfg.ExcludePatterns) != 1 || cfg.ExcludePatterns[0] != "tests/**/*.py" {
			t.Errorf("expected custom exclude patterns, got %v", cfg.ExcludePatterns)
		}
		if cfg.ComplexityEnabled {
			t.Error("expected complexity disabled")
		}
		if cfg.ComplexityReportUnchanged {
			t.Error("expected report_unchanged false")
		}
		if cfg.ComplexityLowThreshold != 3 {
			t.Errorf("expected low threshold 3, got %d", cfg.ComplexityLowThreshold)
		}
		if cfg.ComplexityMediumThreshold != 7 {
			t.Errorf("expected medium threshold 7, got %d", cfg.ComplexityMediumThreshold)
		}
		if cfg.ComplexityMaxComplexity != 11 {
			t.Errorf("expected max complexity 11, got %d", cfg.ComplexityMaxComplexity)
		}
		if cfg.ComplexityMinComplexity != 9 {
			t.Errorf("expected min complexity 9, got %d", cfg.ComplexityMinComplexity)
		}
		if cfg.CloneLSHEnabled != "true" {
			t.Errorf("expected LSH enabled true, got %q", cfg.CloneLSHEnabled)
		}
		if cfg.CloneLSHAutoThreshold != 123 {
			t.Errorf("expected LSH threshold 123, got %d", cfg.CloneLSHAutoThreshold)
		}
	})
}
