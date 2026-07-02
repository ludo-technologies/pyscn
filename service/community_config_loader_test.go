package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestCommunityConfigurationLoader_LoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	configContent := `[communities]
enabled = true
algorithm = "leiden"
scope = "module"
min_community_size = 3
include_lazy_edges = false
report_bridge_modules = true
resolution = 1.25
`
	configPath := filepath.Join(tempDir, ".pyscn.toml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	loader := NewCommunityConfigurationLoader()
	req, err := loader.LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if req.Algorithm != "leiden" {
		t.Errorf("expected algorithm leiden, got %q", req.Algorithm)
	}
	if req.Scope != "module" {
		t.Errorf("expected scope module, got %q", req.Scope)
	}
	if req.MinCommunitySize != 3 {
		t.Errorf("expected min_community_size 3, got %d", req.MinCommunitySize)
	}
	if domain.BoolValue(req.IncludeLazyEdges, true) {
		t.Errorf("expected include_lazy_edges false, got %v", req.IncludeLazyEdges)
	}
	if !domain.BoolValue(req.ReportBridgeModules, false) {
		t.Errorf("expected report_bridge_modules true, got %v", req.ReportBridgeModules)
	}
	if req.Resolution != 1.25 {
		t.Errorf("expected resolution 1.25, got %f", req.Resolution)
	}
}

func TestCommunityConfigurationLoader_LoadDefaultConfigUsesDefaults(t *testing.T) {
	loader := NewCommunityConfigurationLoader()
	req := loader.LoadDefaultConfig()

	if req.Algorithm != domain.DefaultCommunityAlgorithm {
		t.Errorf("expected algorithm %q, got %q", domain.DefaultCommunityAlgorithm, req.Algorithm)
	}
	if req.Scope != domain.DefaultCommunityScope {
		t.Errorf("expected scope %q, got %q", domain.DefaultCommunityScope, req.Scope)
	}
	if req.MinCommunitySize != 2 {
		t.Errorf("expected min_community_size 2, got %d", req.MinCommunitySize)
	}
	if !domain.BoolValue(req.IncludeLazyEdges, false) {
		t.Error("expected include_lazy_edges true by default")
	}
	if !domain.BoolValue(req.ReportBridgeModules, false) {
		t.Error("expected report_bridge_modules true by default")
	}
	if req.Resolution != domain.DefaultCommunityResolution {
		t.Errorf("expected resolution %f, got %f", domain.DefaultCommunityResolution, req.Resolution)
	}
}
