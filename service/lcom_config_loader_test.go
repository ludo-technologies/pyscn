package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLCOMConfigurationLoader(t *testing.T) {
	loader := NewLCOMConfigurationLoader()
	assert.NotNil(t, loader)
}

func TestLCOMConfigurationLoader_LoadDefaultConfig(t *testing.T) {
	loader := NewLCOMConfigurationLoader()
	config := loader.LoadDefaultConfig()

	require.NotNil(t, config)
	assert.Greater(t, config.LowThreshold, 0)
	assert.Greater(t, config.MediumThreshold, config.LowThreshold)
	assert.Equal(t, domain.SortByCohesion, config.SortBy)
	require.NotNil(t, config.ShowDetails)
	assert.False(t, *config.ShowDetails)
}

func TestLCOMConfigurationLoader_MergeConfig(t *testing.T) {
	loader := NewLCOMConfigurationLoader()

	t.Run("nil base returns override", func(t *testing.T) {
		override := &domain.LCOMRequest{LowThreshold: 3}
		result := loader.MergeConfig(nil, override)
		assert.Equal(t, 3, result.LowThreshold)
	})

	t.Run("nil override returns base", func(t *testing.T) {
		base := &domain.LCOMRequest{LowThreshold: 3}
		result := loader.MergeConfig(base, nil)
		assert.Equal(t, 3, result.LowThreshold)
	})

	t.Run("override takes precedence", func(t *testing.T) {
		base := &domain.LCOMRequest{
			LowThreshold:    2,
			MediumThreshold: 5,
			OutputFormat:    domain.OutputFormatText,
		}
		override := &domain.LCOMRequest{
			LowThreshold: 3,
			OutputFormat: domain.OutputFormatJSON,
		}
		result := loader.MergeConfig(base, override)
		assert.Equal(t, 3, result.LowThreshold)
		assert.Equal(t, domain.OutputFormatJSON, result.OutputFormat)
		// MediumThreshold not overridden, so base is used
		assert.Equal(t, 5, result.MediumThreshold)
	})

	t.Run("paths override when non-empty", func(t *testing.T) {
		base := &domain.LCOMRequest{
			Paths: []string{"base.py"},
		}
		override := &domain.LCOMRequest{
			Paths: []string{"override.py"},
		}
		result := loader.MergeConfig(base, override)
		assert.Equal(t, []string{"override.py"}, result.Paths)
	})

	t.Run("empty override paths keep base", func(t *testing.T) {
		base := &domain.LCOMRequest{
			Paths: []string{"base.py"},
		}
		override := &domain.LCOMRequest{}
		result := loader.MergeConfig(base, override)
		assert.Equal(t, []string{"base.py"}, result.Paths)
	})

	// Regression: an explicit override equal to a default value must still win
	// over the base. Zero-value fields mean "not set" and keep the base.
	t.Run("override matching default value wins", func(t *testing.T) {
		base := &domain.LCOMRequest{
			SortBy:       domain.SortByName,
			LowThreshold: 5,
		}
		override := &domain.LCOMRequest{
			SortBy:       domain.SortByCohesion,
			LowThreshold: domain.DefaultLCOMLowThreshold,
		}
		result := loader.MergeConfig(base, override)
		assert.Equal(t, domain.SortByCohesion, result.SortBy,
			"explicit SortBy override matching default should win over base")
		assert.Equal(t, domain.DefaultLCOMLowThreshold, result.LowThreshold,
			"explicit LowThreshold override matching default should win over base")
	})

	t.Run("zero override preserves base", func(t *testing.T) {
		base := &domain.LCOMRequest{
			SortBy:          domain.SortByName,
			LowThreshold:    5,
			MediumThreshold: 9,
			ShowDetails:     domain.BoolPtr(true),
		}
		override := &domain.LCOMRequest{}
		result := loader.MergeConfig(base, override)
		assert.Equal(t, domain.SortByName, result.SortBy)
		assert.Equal(t, 5, result.LowThreshold)
		assert.Equal(t, 9, result.MediumThreshold)
		assert.True(t, domain.BoolValue(result.ShowDetails, false))
	})

	t.Run("explicit false show details wins", func(t *testing.T) {
		base := &domain.LCOMRequest{ShowDetails: domain.BoolPtr(true)}
		override := &domain.LCOMRequest{ShowDetails: domain.BoolPtr(false)}
		result := loader.MergeConfig(base, override)
		assert.False(t, domain.BoolValue(result.ShowDetails, true))
	})
}
