package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

// TestMockDataConfigurationLoader_MergeConfigOverrideEqualsDefault verifies that
// override values equal to the domain defaults still take precedence over the
// base config.
func TestMockDataConfigurationLoader_MergeConfigOverrideEqualsDefault(t *testing.T) {
	loader := NewMockDataConfigurationLoader()

	base := domain.DefaultMockDataRequest()
	base.MinSeverity = domain.MockDataSeverityError
	base.SortBy = domain.MockDataSortByLine

	override := &domain.MockDataRequest{
		MinSeverity: domain.MockDataSeverityWarning,
		SortBy:      domain.MockDataSortBySeverity,
	}

	merged := loader.MergeConfig(base, override)

	if merged.MinSeverity != domain.MockDataSeverityWarning {
		t.Errorf("expected min severity %q, got %q", domain.MockDataSeverityWarning, merged.MinSeverity)
	}
	if merged.SortBy != domain.MockDataSortBySeverity {
		t.Errorf("expected sort by %q, got %q", domain.MockDataSortBySeverity, merged.SortBy)
	}
}

// TestMockDataConfigurationLoader_MergeConfigZeroValueKeepsBase verifies that a
// zero-valued override preserves all base values.
func TestMockDataConfigurationLoader_MergeConfigZeroValueKeepsBase(t *testing.T) {
	loader := NewMockDataConfigurationLoader()

	base := domain.DefaultMockDataRequest()
	base.MinSeverity = domain.MockDataSeverityError
	base.SortBy = domain.MockDataSortByLine
	base.IgnoreTests = domain.BoolPtr(true)
	base.Recursive = domain.BoolPtr(false)
	base.Keywords = []string{"mock", "stub"}

	override := &domain.MockDataRequest{}

	merged := loader.MergeConfig(base, override)

	if merged.MinSeverity != domain.MockDataSeverityError {
		t.Errorf("expected min severity %q preserved, got %q", domain.MockDataSeverityError, merged.MinSeverity)
	}
	if merged.SortBy != domain.MockDataSortByLine {
		t.Errorf("expected sort by %q preserved, got %q", domain.MockDataSortByLine, merged.SortBy)
	}
	if !domain.BoolValue(merged.IgnoreTests, false) {
		t.Errorf("expected ignore_tests true preserved, got %v", merged.IgnoreTests)
	}
	if domain.BoolValue(merged.Recursive, true) {
		t.Errorf("expected recursive false preserved, got %v", merged.Recursive)
	}
	if len(merged.Keywords) != 2 {
		t.Errorf("expected keywords preserved, got %v", merged.Keywords)
	}
}

func TestMockDataConfigurationLoader_MergeConfigExplicitFalseRecursiveWins(t *testing.T) {
	loader := NewMockDataConfigurationLoader()
	base := &domain.MockDataRequest{Recursive: domain.BoolPtr(true)}
	override := &domain.MockDataRequest{Recursive: domain.BoolPtr(false)}

	merged := loader.MergeConfig(base, override)
	if domain.BoolValue(merged.Recursive, true) {
		t.Errorf("expected explicit recursive=false override, got %v", merged.Recursive)
	}
}
