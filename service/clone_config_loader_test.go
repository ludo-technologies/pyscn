package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneConfigurationLoader_MergeConfig(t *testing.T) {
	loader := NewCloneConfigurationLoader()

	t.Run("preserves config values when request leaves fields unset", func(t *testing.T) {
		base := &domain.CloneRequest{
			Recursive:           domain.BoolPtr(false),
			IncludePatterns:     []string{"pkg/**/*.py"},
			ExcludePatterns:     []string{"tests/**/*.py"},
			MinLines:            12,
			SimilarityThreshold: 0.9,
			ShowDetails:         domain.BoolPtr(true),
			ShowContent:         domain.BoolPtr(true),
			GroupClones:         domain.BoolPtr(false),
		}

		// A sparse override (nil pointers, zero values) means "not set", so the
		// base config values must be preserved.
		merged := loader.MergeConfig(base, &domain.CloneRequest{})
		require.NotNil(t, merged)

		assert.False(t, domain.BoolValue(merged.Recursive, true))
		assert.Equal(t, []string{"pkg/**/*.py"}, merged.IncludePatterns)
		assert.Equal(t, []string{"tests/**/*.py"}, merged.ExcludePatterns)
		assert.Equal(t, 12, merged.MinLines)
		assert.Equal(t, 0.9, merged.SimilarityThreshold)
		assert.True(t, domain.BoolValue(merged.ShowDetails, false))
		assert.True(t, domain.BoolValue(merged.ShowContent, false))
		assert.False(t, domain.BoolValue(merged.GroupClones, true))
	})

	t.Run("overrides config with non-default request values", func(t *testing.T) {
		base := domain.DefaultCloneRequest()
		override := domain.DefaultCloneRequest()
		override.Recursive = domain.BoolPtr(false)
		override.IncludePatterns = []string{"src/**/*.py"}
		override.ExcludePatterns = []string{"vendor/**/*.py"}
		override.MinLines = 7
		override.SimilarityThreshold = 0.72
		override.ShowContent = domain.BoolPtr(true)
		override.GroupClones = domain.BoolPtr(false)
		override.CloneTypes = []domain.CloneType{domain.Type1Clone, domain.Type4Clone}

		merged := loader.MergeConfig(base, override)
		require.NotNil(t, merged)

		assert.False(t, domain.BoolValue(merged.Recursive, true))
		assert.Equal(t, []string{"src/**/*.py"}, merged.IncludePatterns)
		assert.Equal(t, []string{"vendor/**/*.py"}, merged.ExcludePatterns)
		assert.Equal(t, 7, merged.MinLines)
		assert.Equal(t, 0.72, merged.SimilarityThreshold)
		assert.True(t, domain.BoolValue(merged.ShowContent, false))
		assert.False(t, domain.BoolValue(merged.GroupClones, true))
		assert.Equal(t, []domain.CloneType{domain.Type1Clone, domain.Type4Clone}, merged.CloneTypes)
	})
}

func TestCloneConfigurationLoader_GetDefaultCloneConfig_LoadsGroupingDefaults(t *testing.T) {
	loader := NewCloneConfigurationLoader()
	configPath := filepath.Join(t.TempDir(), ".pyscn.toml")
	configContent := `[clones]
show_content = true
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	req, err := loader.LoadCloneConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, req)

	assert.Equal(t, "connected", req.GroupMode)
	assert.Equal(t, domain.DefaultCloneGroupingThreshold, req.GroupThreshold)
	assert.Equal(t, 2, req.KCoreK)
	assert.True(t, domain.BoolValue(req.SkipDocstrings, false))
}

func TestCloneConfigurationLoader_LoadCloneConfig_HonorsSkipDocstrings(t *testing.T) {
	loader := NewCloneConfigurationLoader()
	configPath := filepath.Join(t.TempDir(), ".pyscn.toml")
	configContent := `[clones]
skip_docstrings = false
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	req, err := loader.LoadCloneConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, req)

	assert.False(t, domain.BoolValue(req.SkipDocstrings, true))
}
