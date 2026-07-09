package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

// TestDIAntipatternConfigurationLoader_MergeConfigOverrideEqualsDefault verifies
// that override values equal to the domain defaults still take precedence over
// the base config.
func TestDIAntipatternConfigurationLoader_MergeConfigOverrideEqualsDefault(t *testing.T) {
	loader := NewDIAntipatternConfigurationLoader()

	base := domain.DefaultDIAntipatternRequest()
	base.MinSeverity = domain.DIAntipatternSeverityError
	base.ConstructorParamThreshold = 10

	override := &domain.DIAntipatternRequest{
		MinSeverity:               domain.DIAntipatternSeverityWarning,
		ConstructorParamThreshold: domain.DefaultDIConstructorParamThreshold,
	}

	merged := loader.MergeConfig(base, override)

	if merged.MinSeverity != domain.DIAntipatternSeverityWarning {
		t.Errorf("expected min severity %q, got %q", domain.DIAntipatternSeverityWarning, merged.MinSeverity)
	}
	if merged.ConstructorParamThreshold != domain.DefaultDIConstructorParamThreshold {
		t.Errorf("expected threshold %d, got %d", domain.DefaultDIConstructorParamThreshold, merged.ConstructorParamThreshold)
	}
}

// TestDIAntipatternConfigurationLoader_MergeConfigZeroValueKeepsBase verifies
// that a zero-valued override preserves all base values.
func TestDIAntipatternConfigurationLoader_MergeConfigZeroValueKeepsBase(t *testing.T) {
	loader := NewDIAntipatternConfigurationLoader()

	base := domain.DefaultDIAntipatternRequest()
	base.MinSeverity = domain.DIAntipatternSeverityError
	base.ConstructorParamThreshold = 10
	base.Recursive = domain.BoolPtr(false)

	override := &domain.DIAntipatternRequest{}

	merged := loader.MergeConfig(base, override)

	if merged.MinSeverity != domain.DIAntipatternSeverityError {
		t.Errorf("expected min severity %q preserved, got %q", domain.DIAntipatternSeverityError, merged.MinSeverity)
	}
	if merged.ConstructorParamThreshold != 10 {
		t.Errorf("expected threshold 10 preserved, got %d", merged.ConstructorParamThreshold)
	}
	if domain.BoolValue(merged.Recursive, true) {
		t.Errorf("expected recursive false preserved, got %v", merged.Recursive)
	}
}
