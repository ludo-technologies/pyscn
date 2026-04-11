package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCloneConfigurationLoader_MergeConfig(t *testing.T) {
	loader := NewCloneConfigurationLoader()

	t.Run("preserves config values when request still uses defaults", func(t *testing.T) {
		base := &domain.CloneRequest{
			Recursive:           false,
			IncludePatterns:     []string{"pkg/**/*.py"},
			ExcludePatterns:     []string{"tests/**/*.py"},
			MinLines:            12,
			SimilarityThreshold: 0.9,
			ShowDetails:         true,
			ShowContent:         true,
			GroupClones:         false,
		}

		merged := loader.MergeConfig(base, domain.DefaultCloneRequest())
		require.NotNil(t, merged)

		assert.False(t, merged.Recursive)
		assert.Equal(t, []string{"pkg/**/*.py"}, merged.IncludePatterns)
		assert.Equal(t, []string{"tests/**/*.py"}, merged.ExcludePatterns)
		assert.Equal(t, 12, merged.MinLines)
		assert.Equal(t, 0.9, merged.SimilarityThreshold)
		assert.True(t, merged.ShowDetails)
		assert.True(t, merged.ShowContent)
		assert.False(t, merged.GroupClones)
	})

	t.Run("overrides config with non-default request values", func(t *testing.T) {
		base := domain.DefaultCloneRequest()
		override := domain.DefaultCloneRequest()
		override.Recursive = false
		override.IncludePatterns = []string{"src/**/*.py"}
		override.ExcludePatterns = []string{"vendor/**/*.py"}
		override.MinLines = 7
		override.SimilarityThreshold = 0.72
		override.ShowContent = true
		override.GroupClones = false
		override.CloneTypes = []domain.CloneType{domain.Type1Clone, domain.Type4Clone}

		merged := loader.MergeConfig(base, override)
		require.NotNil(t, merged)

		assert.False(t, merged.Recursive)
		assert.Equal(t, []string{"src/**/*.py"}, merged.IncludePatterns)
		assert.Equal(t, []string{"vendor/**/*.py"}, merged.ExcludePatterns)
		assert.Equal(t, 7, merged.MinLines)
		assert.Equal(t, 0.72, merged.SimilarityThreshold)
		assert.True(t, merged.ShowContent)
		assert.False(t, merged.GroupClones)
		assert.Equal(t, []domain.CloneType{domain.Type1Clone, domain.Type4Clone}, merged.CloneTypes)
	})
}

func TestCloneConfigurationLoaderWithFlags_MergeConfig(t *testing.T) {
	loader := NewCloneConfigurationLoaderWithFlags(map[string]bool{
		"show-content": true,
		"similarity":   true,
	})

	base := &domain.CloneRequest{
		ShowContent:         true,
		SimilarityThreshold: 0.9,
	}
	override := domain.DefaultCloneRequest()

	merged := loader.MergeConfig(base, override)
	require.NotNil(t, merged)

	assert.False(t, merged.ShowContent)
	assert.Equal(t, domain.DefaultCloneSimilarityThreshold, merged.SimilarityThreshold)
}
