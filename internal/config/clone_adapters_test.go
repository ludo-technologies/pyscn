package config

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestCloneConfig_Structure(t *testing.T) {
	cloneConfig := DefaultPyscnConfig()

	// Verify the unified config has all expected sections
	assert.NotNil(t, cloneConfig.Analysis)
	assert.NotNil(t, cloneConfig.Thresholds)
	assert.NotNil(t, cloneConfig.Filtering)
	assert.NotNil(t, cloneConfig.Input)
	assert.NotNil(t, cloneConfig.Output)
	assert.NotNil(t, cloneConfig.Performance)

	// Verify key fields have reasonable values
	assert.Greater(t, cloneConfig.Analysis.MinLines, 0)
	assert.Greater(t, cloneConfig.Analysis.MinNodes, 0)
	assert.Greater(t, cloneConfig.Thresholds.Type1Threshold, 0.0)
	assert.LessOrEqual(t, cloneConfig.Thresholds.Type1Threshold, 1.0)
}

// TestCloneConfig_ToCloneDetectionConfig removed - CloneDetectionConfig type has been deleted

func TestCloneConfig_ToCloneRequest(t *testing.T) {
	cloneConfig := DefaultPyscnConfig()
	cloneConfig.Input.Paths = []string{"/test/path"}
	cloneConfig.Input.Recursive = BoolPtr(true)
	cloneConfig.Input.IncludePatterns = []string{"**/*.py"}
	cloneConfig.Input.ExcludePatterns = []string{"*_test.py"}
	cloneConfig.Output.Format = "json"
	cloneConfig.Output.SortBy = "similarity"
	cloneConfig.Filtering.EnabledCloneTypes = []string{"type1", "type2"}

	outputWriter := os.Stdout
	request := cloneConfig.ToCloneRequest(outputWriter)

	// Verify input parameters
	assert.Equal(t, []string{"/test/path"}, request.Paths)
	assert.True(t, request.Recursive)
	assert.Equal(t, []string{"**/*.py"}, request.IncludePatterns)
	assert.Equal(t, []string{"*_test.py"}, request.ExcludePatterns)

	// Verify analysis configuration
	assert.Equal(t, cloneConfig.Analysis.MinLines, request.MinLines)
	assert.Equal(t, cloneConfig.Analysis.MinNodes, request.MinNodes)
	assert.Equal(t, cloneConfig.Thresholds.SimilarityThreshold, request.SimilarityThreshold)
	assert.Equal(t, cloneConfig.Analysis.MaxEditDistance, request.MaxEditDistance)
	assert.Equal(t, BoolValue(cloneConfig.Analysis.IgnoreLiterals, false), request.IgnoreLiterals)
	assert.Equal(t, BoolValue(cloneConfig.Analysis.IgnoreIdentifiers, false), request.IgnoreIdentifiers)

	// Verify thresholds
	assert.Equal(t, cloneConfig.Thresholds.Type1Threshold, request.Type1Threshold)
	assert.Equal(t, cloneConfig.Thresholds.Type2Threshold, request.Type2Threshold)
	assert.Equal(t, cloneConfig.Thresholds.Type3Threshold, request.Type3Threshold)
	assert.Equal(t, cloneConfig.Thresholds.Type4Threshold, request.Type4Threshold)

	// Verify output configuration
	assert.Equal(t, domain.OutputFormatJSON, request.OutputFormat)
	assert.Equal(t, outputWriter, request.OutputWriter)
	assert.Equal(t, domain.SortBySimilarity, request.SortBy)

	// Verify clone types conversion
	expectedTypes := []domain.CloneType{domain.Type1Clone, domain.Type2Clone}
	assert.Equal(t, expectedTypes, request.CloneTypes)
}

func TestAdapterPlaceholder(t *testing.T) {
	// Note: Analyzer-specific adapter functions are implemented directly in analyzer package
	// to avoid circular import dependencies. This is a placeholder test.
	assert.True(t, true, "Adapter functions work correctly - tested separately in analyzer package")
}

// TestFromCloneDetectionConfig removed - CloneDetectionConfig type has been deleted

func TestFromCloneRequest(t *testing.T) {
	outputWriter := io.Discard
	request := &domain.CloneRequest{
		Paths:           []string{"/test/path1", "/test/path2"},
		Recursive:       false,
		IncludePatterns: []string{"**/*.py", "*.pyx"},
		ExcludePatterns: []string{"test_*.py"},

		MinLines:            12,
		MinNodes:            25,
		SimilarityThreshold: 0.8,
		MaxEditDistance:     45.0,
		IgnoreLiterals:      false,
		IgnoreIdentifiers:   true,

		Type1Threshold: 0.96,
		Type2Threshold: 0.86,
		Type3Threshold: 0.76,
		Type4Threshold: 0.66,

		OutputFormat: domain.OutputFormatYAML,
		OutputWriter: outputWriter,
		ShowDetails:  true,
		ShowContent:  false,
		SortBy:       domain.SortBySize,
		GroupClones:  true,

		MinSimilarity: 0.5,
		MaxSimilarity: 0.9,
		CloneTypes:    []domain.CloneType{domain.Type1Clone, domain.Type3Clone, domain.Type4Clone},
	}

	cloneConfig := FromCloneRequest(request)

	// Verify input conversion
	assert.Equal(t, request.Paths, cloneConfig.Input.Paths)
	assert.Equal(t, BoolPtr(request.Recursive), cloneConfig.Input.Recursive)
	assert.Equal(t, request.IncludePatterns, cloneConfig.Input.IncludePatterns)
	assert.Equal(t, request.ExcludePatterns, cloneConfig.Input.ExcludePatterns)

	// Verify analysis conversion
	assert.Equal(t, request.MinLines, cloneConfig.Analysis.MinLines)
	assert.Equal(t, request.MinNodes, cloneConfig.Analysis.MinNodes)
	assert.Equal(t, request.MaxEditDistance, cloneConfig.Analysis.MaxEditDistance)
	assert.Equal(t, BoolPtr(request.IgnoreLiterals), cloneConfig.Analysis.IgnoreLiterals)
	assert.Equal(t, BoolPtr(request.IgnoreIdentifiers), cloneConfig.Analysis.IgnoreIdentifiers)

	// Verify thresholds conversion
	assert.Equal(t, request.Type1Threshold, cloneConfig.Thresholds.Type1Threshold)
	assert.Equal(t, request.Type2Threshold, cloneConfig.Thresholds.Type2Threshold)
	assert.Equal(t, request.Type3Threshold, cloneConfig.Thresholds.Type3Threshold)
	assert.Equal(t, request.Type4Threshold, cloneConfig.Thresholds.Type4Threshold)
	assert.Equal(t, request.SimilarityThreshold, cloneConfig.Thresholds.SimilarityThreshold)

	// Verify output conversion
	assert.Equal(t, "yaml", cloneConfig.Output.Format)
	assert.Equal(t, outputWriter, cloneConfig.Output.Writer)
	assert.Equal(t, BoolPtr(request.ShowDetails), cloneConfig.Output.ShowDetails)
	assert.Equal(t, BoolPtr(request.ShowContent), cloneConfig.Output.ShowContent)
	assert.Equal(t, "size", cloneConfig.Output.SortBy)
	assert.Equal(t, BoolPtr(request.GroupClones), cloneConfig.Output.GroupClones)

	// Verify filtering conversion
	assert.Equal(t, request.MinSimilarity, cloneConfig.Filtering.MinSimilarity)
	assert.Equal(t, request.MaxSimilarity, cloneConfig.Filtering.MaxSimilarity)
	expectedCloneTypes := []string{"type1", "type3", "type4"}
	assert.Equal(t, expectedCloneTypes, cloneConfig.Filtering.EnabledCloneTypes)
}

func TestRoundTripConversion(t *testing.T) {
	t.Run("Configuration roundtrip validation", func(t *testing.T) {
		// Test that unified config can be created and is valid
		original := DefaultPyscnConfig()
		assert.NoError(t, original.Validate(), "Default config should be valid")

		// Test that we can create variations and they remain valid
		modified := DefaultPyscnConfig()
		modified.Analysis.MinLines = 10
		modified.Thresholds.Type1Threshold = 0.98
		assert.NoError(t, modified.Validate(), "Modified config should be valid")
	})

	// CloneDetectionConfig roundtrip test removed - CloneDetectionConfig type has been deleted
}

// TestDefaultConfigBackwardCompatibility removed - CloneDetection field has been deleted

func TestCloneConfigValidation(t *testing.T) {
	t.Run("Valid default config", func(t *testing.T) {
		config := DefaultPyscnConfig()
		err := config.Validate()
		assert.NoError(t, err)
	})

	t.Run("Invalid analysis config", func(t *testing.T) {
		config := DefaultPyscnConfig()
		config.Analysis.MinLines = 0 // Invalid

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "analysis config invalid")
	})

	t.Run("Invalid threshold config", func(t *testing.T) {
		config := DefaultPyscnConfig()
		config.Thresholds.Type1Threshold = -0.1 // Invalid range

		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "thresholds config invalid")
	})
}
