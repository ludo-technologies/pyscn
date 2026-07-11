package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
)

func TestDeadCodeConfigurationLoader_MergeConfig_NilGuards(t *testing.T) {
	loader := NewDeadCodeConfigurationLoader()

	t.Run("nil base returns override", func(t *testing.T) {
		override := &domain.DeadCodeRequest{MinSeverity: domain.DeadCodeSeverityCritical}
		result := loader.MergeConfig(nil, override)
		assert.Equal(t, domain.DeadCodeSeverityCritical, result.MinSeverity)
	})

	t.Run("nil override returns base", func(t *testing.T) {
		base := &domain.DeadCodeRequest{MinSeverity: domain.DeadCodeSeverityCritical}
		result := loader.MergeConfig(base, nil)
		assert.Equal(t, domain.DeadCodeSeverityCritical, result.MinSeverity)
	})
}

// TestDeadCodeConfigurationLoader_MergeConfig_OverrideMatchesDefault is a
// regression test: explicit MinSeverity/SortBy overrides equal to their
// defaults must still win over the base. The old guards
// (!= domain.DeadCodeSeverityWarning, != domain.DeadCodeSortBySeverity)
// dropped them.
func TestDeadCodeConfigurationLoader_MergeConfig_OverrideMatchesDefault(t *testing.T) {
	loader := NewDeadCodeConfigurationLoader()

	base := &domain.DeadCodeRequest{
		MinSeverity: domain.DeadCodeSeverityCritical,
		SortBy:      domain.DeadCodeSortByFile,
	}
	override := &domain.DeadCodeRequest{
		MinSeverity: domain.DeadCodeSeverityWarning,
		SortBy:      domain.DeadCodeSortBySeverity,
	}

	merged := loader.MergeConfig(base, override)
	assert.Equal(t, domain.DeadCodeSeverityWarning, merged.MinSeverity,
		"explicit MinSeverity override matching default should win over base")
	assert.Equal(t, domain.DeadCodeSortBySeverity, merged.SortBy,
		"explicit SortBy override matching default should win over base")
}

// TestDeadCodeConfigurationLoader_MergeConfig_ZeroOverridePreservesBase verifies
// that unset (zero-value) override fields leave the base values intact.
func TestDeadCodeConfigurationLoader_MergeConfig_ZeroOverridePreservesBase(t *testing.T) {
	loader := NewDeadCodeConfigurationLoader()

	base := &domain.DeadCodeRequest{
		MinSeverity:  domain.DeadCodeSeverityCritical,
		SortBy:       domain.DeadCodeSortByFile,
		ContextLines: 5,
		ShowContext:  domain.BoolPtr(true),
	}
	override := &domain.DeadCodeRequest{}

	merged := loader.MergeConfig(base, override)
	assert.Equal(t, domain.DeadCodeSeverityCritical, merged.MinSeverity)
	assert.Equal(t, domain.DeadCodeSortByFile, merged.SortBy)
	assert.Equal(t, 5, merged.ContextLines)
	assert.True(t, domain.BoolValue(merged.ShowContext, false))
}

// TestDeadCodeConfigurationLoader_MergeConfig_ExplicitFalseBoolWins verifies
// that a non-nil *bool override (including explicit false) takes precedence
// over the base value.
func TestDeadCodeConfigurationLoader_MergeConfig_ExplicitFalseBoolWins(t *testing.T) {
	loader := NewDeadCodeConfigurationLoader()

	base := &domain.DeadCodeRequest{
		DetectAfterReturn: domain.BoolPtr(true),
	}
	override := &domain.DeadCodeRequest{
		DetectAfterReturn: domain.BoolPtr(false),
	}

	merged := loader.MergeConfig(base, override)
	assert.False(t, domain.BoolValue(merged.DetectAfterReturn, true),
		"explicit false *bool override should win over base")
}
