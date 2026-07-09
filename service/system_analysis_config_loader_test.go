package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

// TestSystemAnalysisConfigurationLoader_MergeConfigOverrideEqualsDefault verifies
// that an override value which happens to equal the domain default (here the
// text output format) still takes precedence over the base config. Previously a
// `!= domain.OutputFormatText` guard silently dropped such overrides.
func TestSystemAnalysisConfigurationLoader_MergeConfigOverrideEqualsDefault(t *testing.T) {
	loader := NewSystemAnalysisConfigurationLoader()

	base := loader.LoadDefaultConfig()
	base.OutputFormat = domain.OutputFormatJSON

	override := &domain.SystemAnalysisRequest{
		OutputFormat: domain.OutputFormatText,
	}

	merged := loader.MergeConfig(base, override)

	if merged.OutputFormat != domain.OutputFormatText {
		t.Errorf("expected output format %q to override base, got %q", domain.OutputFormatText, merged.OutputFormat)
	}
}

// TestSystemAnalysisConfigurationLoader_MergeConfigZeroValueKeepsBase verifies
// that a zero-valued override ("no CLI flags set") preserves all base values.
func TestSystemAnalysisConfigurationLoader_MergeConfigZeroValueKeepsBase(t *testing.T) {
	loader := NewSystemAnalysisConfigurationLoader()

	base := loader.LoadDefaultConfig()
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
