package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

// TestSystemAnalysisConfigurationLoader_MergeConfigOverrideEqualsDefault verifies
// that an override value which happens to equal the domain default (here the
// text output format) still takes precedence over the base config. Previously a
// `!= domain.OutputFormatText` guard silently dropped such overrides.
func TestSystemAnalysisConfigurationLoader_MergeConfigOverrideEqualsDefault(t *testing.T) {
	loader := NewSystemAnalysisConfigurationLoader()

	base := loader.LoadDefaultConfig("")
	base.OutputFormat = domain.OutputFormatJSON

	override := &domain.SystemAnalysisRequest{
		OutputFormat: domain.OutputFormatText,
	}

	merged := loader.MergeConfig(base, override)

	if merged.OutputFormat != domain.OutputFormatText {
		t.Errorf("expected output format %q to override base, got %q", domain.OutputFormatText, merged.OutputFormat)
	}
}

func TestSystemAnalysisConfigurationLoaderLoadsConfiguredProjectRoot(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, ".pyscn.toml")
	if err := os.WriteFile(configPath, []byte("project_root = \"src\"\n"), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	request, err := NewSystemAnalysisConfigurationLoader().LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	want := filepath.Join(root, "src")
	if request.ProjectRoot != want {
		t.Fatalf("expected project root %q, got %q", want, request.ProjectRoot)
	}
}

func TestSystemAnalysisConfigurationLoaderDiscoversProjectRootFromTargetPath(t *testing.T) {
	root := t.TempDir()
	targetDir := filepath.Join(root, "src")
	targetPath := filepath.Join(targetDir, "consumer.py")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("failed to create target directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".pyscn.toml"), []byte("project_root = \"src\"\n"), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("value = 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write target: %v", err)
	}

	request := NewSystemAnalysisConfigurationLoader().LoadDefaultConfig(targetPath)
	if request.ProjectRoot != targetDir {
		t.Errorf("expected project root %q, got %q", targetDir, request.ProjectRoot)
	}
}

// TestSystemAnalysisConfigurationLoader_MergeConfigZeroValueKeepsBase verifies
// that a zero-valued override ("no CLI flags set") preserves all base values.
func TestSystemAnalysisConfigurationLoader_MergeConfigZeroValueKeepsBase(t *testing.T) {
	loader := NewSystemAnalysisConfigurationLoader()

	base := loader.LoadDefaultConfig("")
	base.OutputFormat = domain.OutputFormatJSON
	base.MinCohesion = 0.9
	base.MaxResponsibilities = 7
	base.IncludeStdLib = domain.BoolPtr(true)

	override := &domain.SystemAnalysisRequest{}

	merged := loader.MergeConfig(base, override)

	if merged.OutputFormat != domain.OutputFormatJSON {
		t.Errorf("expected output format %q preserved, got %q", domain.OutputFormatJSON, merged.OutputFormat)
	}
	if merged.MinCohesion != 0.9 {
		t.Errorf("expected min cohesion 0.9 preserved, got %f", merged.MinCohesion)
	}
	if merged.MaxResponsibilities != 7 {
		t.Errorf("expected max responsibilities 7 preserved, got %d", merged.MaxResponsibilities)
	}
	if !domain.BoolValue(merged.IncludeStdLib, false) {
		t.Errorf("expected include_stdlib true preserved, got %v", merged.IncludeStdLib)
	}
}

func TestSystemAnalysisConfigurationLoader_MergesExplicitProjectRoot(t *testing.T) {
	loader := NewSystemAnalysisConfigurationLoader()
	base := loader.LoadDefaultConfig("")
	base.ProjectRoot = "inferred"

	merged := loader.MergeConfig(base, &domain.SystemAnalysisRequest{ProjectRoot: "explicit"})
	if merged.ProjectRoot != "explicit" {
		t.Fatalf("expected explicit project root to override base, got %q", merged.ProjectRoot)
	}
}
