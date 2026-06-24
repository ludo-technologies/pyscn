package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestLoadCommunitiesFromPyscnToml(t *testing.T) {
	tempDir := t.TempDir()

	configContent := `[communities]
enabled = true
algorithm = "leiden"
scope = "module"
min_community_size = 3
include_lazy_edges = false
report_bridge_modules = false
resolution = 0.75
`
	configPath := filepath.Join(tempDir, ".pyscn.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewTomlConfigLoader()
	cfg, err := loader.LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !domain.BoolValue(cfg.CommunitiesEnabled, false) {
		t.Errorf("Expected enabled true, got %v", cfg.CommunitiesEnabled)
	}
	if cfg.CommunitiesAlgorithm != "leiden" {
		t.Errorf("Expected algorithm leiden, got %q", cfg.CommunitiesAlgorithm)
	}
	if cfg.CommunitiesScope != "module" {
		t.Errorf("Expected scope module, got %q", cfg.CommunitiesScope)
	}
	if cfg.CommunitiesMinCommunitySize != 3 {
		t.Errorf("Expected min_community_size 3, got %d", cfg.CommunitiesMinCommunitySize)
	}
	if domain.BoolValue(cfg.CommunitiesIncludeLazyEdges, true) {
		t.Errorf("Expected include_lazy_edges false, got %v", cfg.CommunitiesIncludeLazyEdges)
	}
	if domain.BoolValue(cfg.CommunitiesReportBridgeModules, true) {
		t.Errorf("Expected report_bridge_modules false, got %v", cfg.CommunitiesReportBridgeModules)
	}
	if cfg.CommunitiesResolution != 0.75 {
		t.Errorf("Expected resolution 0.75, got %f", cfg.CommunitiesResolution)
	}
}

func TestLoadCommunitiesFromPyscnTomlPartial(t *testing.T) {
	tempDir := t.TempDir()

	configContent := `[communities]
enabled = true
min_community_size = 4
`
	configPath := filepath.Join(tempDir, ".pyscn.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	loader := NewTomlConfigLoader()
	cfg, err := loader.LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if !domain.BoolValue(cfg.CommunitiesEnabled, false) {
		t.Errorf("Expected enabled true, got %v", cfg.CommunitiesEnabled)
	}
	if cfg.CommunitiesMinCommunitySize != 4 {
		t.Errorf("Expected min_community_size 4, got %d", cfg.CommunitiesMinCommunitySize)
	}
	if cfg.CommunitiesAlgorithm != domain.DefaultCommunityAlgorithm {
		t.Errorf("Expected default algorithm %q, got %q", domain.DefaultCommunityAlgorithm, cfg.CommunitiesAlgorithm)
	}
	if cfg.CommunitiesScope != domain.DefaultCommunityScope {
		t.Errorf("Expected default scope %q, got %q", domain.DefaultCommunityScope, cfg.CommunitiesScope)
	}
	if !domain.BoolValue(cfg.CommunitiesIncludeLazyEdges, false) {
		t.Errorf("Expected default include_lazy_edges true, got %v", cfg.CommunitiesIncludeLazyEdges)
	}
	if !domain.BoolValue(cfg.CommunitiesReportBridgeModules, false) {
		t.Errorf("Expected default report_bridge_modules true, got %v", cfg.CommunitiesReportBridgeModules)
	}
	if cfg.CommunitiesResolution != domain.DefaultCommunityResolution {
		t.Errorf("Expected default resolution %f, got %f", domain.DefaultCommunityResolution, cfg.CommunitiesResolution)
	}
}

func TestMergeCommunitiesSection(t *testing.T) {
	cfg := DefaultPyscnConfig()

	communities := CommunitiesTomlConfig{
		Enabled:             boolPtr(true),
		Algorithm:           "leiden",
		Scope:               "module",
		MinCommunitySize:    intPtr(5),
		IncludeLazyEdges:    boolPtr(false),
		ReportBridgeModules: boolPtr(false),
		Resolution:          float64Ptr(1.5),
	}

	mergeCommunitiesSection(cfg, &communities)

	if !domain.BoolValue(cfg.CommunitiesEnabled, false) {
		t.Errorf("Expected enabled true, got %v", cfg.CommunitiesEnabled)
	}
	if cfg.CommunitiesAlgorithm != "leiden" {
		t.Errorf("Expected algorithm leiden, got %q", cfg.CommunitiesAlgorithm)
	}
	if cfg.CommunitiesScope != "module" {
		t.Errorf("Expected scope module, got %q", cfg.CommunitiesScope)
	}
	if cfg.CommunitiesMinCommunitySize != 5 {
		t.Errorf("Expected min_community_size 5, got %d", cfg.CommunitiesMinCommunitySize)
	}
	if domain.BoolValue(cfg.CommunitiesIncludeLazyEdges, true) {
		t.Errorf("Expected include_lazy_edges false, got %v", cfg.CommunitiesIncludeLazyEdges)
	}
	if domain.BoolValue(cfg.CommunitiesReportBridgeModules, true) {
		t.Errorf("Expected report_bridge_modules false, got %v", cfg.CommunitiesReportBridgeModules)
	}
	if cfg.CommunitiesResolution != 1.5 {
		t.Errorf("Expected resolution 1.5, got %f", cfg.CommunitiesResolution)
	}
}

func TestMergeCommunitiesSectionNilValues(t *testing.T) {
	cfg := DefaultPyscnConfig()
	originalAlgorithm := cfg.CommunitiesAlgorithm

	communities := CommunitiesTomlConfig{}
	mergeCommunitiesSection(cfg, &communities)

	if cfg.CommunitiesAlgorithm != originalAlgorithm {
		t.Errorf("Expected defaults to remain, got algorithm %q", cfg.CommunitiesAlgorithm)
	}
	if domain.BoolValue(cfg.CommunitiesEnabled, true) {
		t.Error("Expected communities to remain disabled by default")
	}
}

func TestDefaultPyscnConfigCommunitiesDefaults(t *testing.T) {
	cfg := DefaultPyscnConfig()

	if domain.BoolValue(cfg.CommunitiesEnabled, true) {
		t.Error("Expected communities disabled by default")
	}
	if cfg.CommunitiesAlgorithm != domain.DefaultCommunityAlgorithm {
		t.Errorf("Expected algorithm %q, got %q", domain.DefaultCommunityAlgorithm, cfg.CommunitiesAlgorithm)
	}
	if cfg.CommunitiesScope != domain.DefaultCommunityScope {
		t.Errorf("Expected scope %q, got %q", domain.DefaultCommunityScope, cfg.CommunitiesScope)
	}
	if cfg.CommunitiesMinCommunitySize != 2 {
		t.Errorf("Expected min_community_size 2, got %d", cfg.CommunitiesMinCommunitySize)
	}
	if !domain.BoolValue(cfg.CommunitiesIncludeLazyEdges, false) {
		t.Error("Expected include_lazy_edges true by default")
	}
	if !domain.BoolValue(cfg.CommunitiesReportBridgeModules, false) {
		t.Error("Expected report_bridge_modules true by default")
	}
	if cfg.CommunitiesResolution != domain.DefaultCommunityResolution {
		t.Errorf("Expected resolution %f, got %f", domain.DefaultCommunityResolution, cfg.CommunitiesResolution)
	}
}

func float64Ptr(val float64) *float64 {
	return &val
}
