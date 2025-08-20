package config

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pyqol/pyqol/domain"
	"github.com/pyqol/pyqol/internal/constants"
)

func TestCloneConfig_ToCloneDetectorConfig(t *testing.T) {
	cloneConfig := DefaultCloneConfig()
	
	detectorConfig := cloneConfig.ToCloneDetectorConfig()
	
	// Verify all fields are properly mapped
	assert.Equal(t, cloneConfig.Analysis.MinLines, detectorConfig.MinLines)
	assert.Equal(t, cloneConfig.Analysis.MinNodes, detectorConfig.MinNodes)
	assert.Equal(t, cloneConfig.Thresholds.Type1Threshold, detectorConfig.Type1Threshold)
	assert.Equal(t, cloneConfig.Thresholds.Type2Threshold, detectorConfig.Type2Threshold)
	assert.Equal(t, cloneConfig.Thresholds.Type3Threshold, detectorConfig.Type3Threshold)
	assert.Equal(t, cloneConfig.Thresholds.Type4Threshold, detectorConfig.Type4Threshold)
	assert.Equal(t, cloneConfig.Analysis.MaxEditDistance, detectorConfig.MaxEditDistance)
	assert.Equal(t, cloneConfig.Analysis.IgnoreLiterals, detectorConfig.IgnoreLiterals)
	assert.Equal(t, cloneConfig.Analysis.IgnoreIdentifiers, detectorConfig.IgnoreIdentifiers)
	assert.Equal(t, cloneConfig.Analysis.CostModelType, detectorConfig.CostModelType)
}

func TestCloneConfig_ToCloneDetectionConfig(t *testing.T) {
	cloneConfig := DefaultCloneConfig()
	
	detectionConfig := cloneConfig.ToCloneDetectionConfig()
	
	// Verify all fields are properly mapped
	assert.True(t, detectionConfig.Enabled) // Should be enabled by default
	assert.Equal(t, cloneConfig.Analysis.MinLines, detectionConfig.MinLines)
	assert.Equal(t, cloneConfig.Analysis.MinNodes, detectionConfig.MinNodes)
	assert.Equal(t, cloneConfig.Thresholds.Type1Threshold, detectionConfig.Type1Threshold)
	assert.Equal(t, cloneConfig.Thresholds.Type2Threshold, detectionConfig.Type2Threshold)
	assert.Equal(t, cloneConfig.Thresholds.Type3Threshold, detectionConfig.Type3Threshold)
	assert.Equal(t, cloneConfig.Thresholds.Type4Threshold, detectionConfig.Type4Threshold)
	assert.Equal(t, cloneConfig.Thresholds.SimilarityThreshold, detectionConfig.SimilarityThreshold)
	assert.Equal(t, cloneConfig.Analysis.MaxEditDistance, detectionConfig.MaxEditDistance)
	assert.Equal(t, cloneConfig.Analysis.CostModelType, detectionConfig.CostModelType)
	assert.Equal(t, cloneConfig.Analysis.IgnoreLiterals, detectionConfig.IgnoreLiterals)
	assert.Equal(t, cloneConfig.Analysis.IgnoreIdentifiers, detectionConfig.IgnoreIdentifiers)
	assert.Equal(t, cloneConfig.Output.ShowContent, detectionConfig.ShowContent)
	assert.Equal(t, cloneConfig.Output.GroupClones, detectionConfig.GroupClones)
}

func TestCloneConfig_ToCloneRequest(t *testing.T) {
	cloneConfig := DefaultCloneConfig()
	cloneConfig.Input.Paths = []string{"/test/path"}
	cloneConfig.Input.Recursive = true
	cloneConfig.Input.IncludePatterns = []string{"*.py"}
	cloneConfig.Input.ExcludePatterns = []string{"*_test.py"}
	cloneConfig.Output.Format = "json"
	cloneConfig.Output.SortBy = "similarity"
	cloneConfig.Filtering.EnabledCloneTypes = []string{"type1", "type2"}
	
	outputWriter := os.Stdout
	request := cloneConfig.ToCloneRequest(outputWriter)
	
	// Verify input parameters
	assert.Equal(t, []string{"/test/path"}, request.Paths)
	assert.True(t, request.Recursive)
	assert.Equal(t, []string{"*.py"}, request.IncludePatterns)
	assert.Equal(t, []string{"*_test.py"}, request.ExcludePatterns)
	
	// Verify analysis configuration
	assert.Equal(t, cloneConfig.Analysis.MinLines, request.MinLines)
	assert.Equal(t, cloneConfig.Analysis.MinNodes, request.MinNodes)
	assert.Equal(t, cloneConfig.Thresholds.SimilarityThreshold, request.SimilarityThreshold)
	assert.Equal(t, cloneConfig.Analysis.MaxEditDistance, request.MaxEditDistance)
	assert.Equal(t, cloneConfig.Analysis.IgnoreLiterals, request.IgnoreLiterals)
	assert.Equal(t, cloneConfig.Analysis.IgnoreIdentifiers, request.IgnoreIdentifiers)
	
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

func TestFromCloneDetectorConfig(t *testing.T) {
	analyzerConfig := CloneDetectorConfig{
		MinLines:          10,
		MinNodes:          20,
		Type1Threshold:    0.98,
		Type2Threshold:    0.88,
		Type3Threshold:    0.78,
		Type4Threshold:    0.68,
		MaxEditDistance:   60.0,
		IgnoreLiterals:    true,
		IgnoreIdentifiers: true,
		CostModelType:     "weighted",
	}
	
	cloneConfig := FromCloneDetectorConfig(analyzerConfig)
	
	// Verify conversion
	assert.Equal(t, analyzerConfig.MinLines, cloneConfig.Analysis.MinLines)
	assert.Equal(t, analyzerConfig.MinNodes, cloneConfig.Analysis.MinNodes)
	assert.Equal(t, analyzerConfig.Type1Threshold, cloneConfig.Thresholds.Type1Threshold)
	assert.Equal(t, analyzerConfig.Type2Threshold, cloneConfig.Thresholds.Type2Threshold)
	assert.Equal(t, analyzerConfig.Type3Threshold, cloneConfig.Thresholds.Type3Threshold)
	assert.Equal(t, analyzerConfig.Type4Threshold, cloneConfig.Thresholds.Type4Threshold)
	assert.Equal(t, analyzerConfig.MaxEditDistance, cloneConfig.Analysis.MaxEditDistance)
	assert.Equal(t, analyzerConfig.IgnoreLiterals, cloneConfig.Analysis.IgnoreLiterals)
	assert.Equal(t, analyzerConfig.IgnoreIdentifiers, cloneConfig.Analysis.IgnoreIdentifiers)
	assert.Equal(t, analyzerConfig.CostModelType, cloneConfig.Analysis.CostModelType)
}

func TestFromCloneDetectionConfig(t *testing.T) {
	detectionConfig := CloneDetectionConfig{
		Enabled:             true,
		MinLines:            8,
		MinNodes:            15,
		Type1Threshold:      0.97,
		Type2Threshold:      0.87,
		Type3Threshold:      0.77,
		Type4Threshold:      0.67,
		SimilarityThreshold: 0.75,
		MaxEditDistance:     40.0,
		CostModelType:       "default",
		IgnoreLiterals:      true,
		IgnoreIdentifiers:   false,
		ShowContent:         true,
		GroupClones:         false,
	}
	
	cloneConfig := FromCloneDetectionConfig(detectionConfig)
	
	// Verify conversion
	assert.Equal(t, detectionConfig.MinLines, cloneConfig.Analysis.MinLines)
	assert.Equal(t, detectionConfig.MinNodes, cloneConfig.Analysis.MinNodes)
	assert.Equal(t, detectionConfig.Type1Threshold, cloneConfig.Thresholds.Type1Threshold)
	assert.Equal(t, detectionConfig.Type2Threshold, cloneConfig.Thresholds.Type2Threshold)
	assert.Equal(t, detectionConfig.Type3Threshold, cloneConfig.Thresholds.Type3Threshold)
	assert.Equal(t, detectionConfig.Type4Threshold, cloneConfig.Thresholds.Type4Threshold)
	assert.Equal(t, detectionConfig.SimilarityThreshold, cloneConfig.Thresholds.SimilarityThreshold)
	assert.Equal(t, detectionConfig.MaxEditDistance, cloneConfig.Analysis.MaxEditDistance)
	assert.Equal(t, detectionConfig.CostModelType, cloneConfig.Analysis.CostModelType)
	assert.Equal(t, detectionConfig.IgnoreLiterals, cloneConfig.Analysis.IgnoreLiterals)
	assert.Equal(t, detectionConfig.IgnoreIdentifiers, cloneConfig.Analysis.IgnoreIdentifiers)
	assert.Equal(t, detectionConfig.ShowContent, cloneConfig.Output.ShowContent)
	assert.Equal(t, detectionConfig.GroupClones, cloneConfig.Output.GroupClones)
}

func TestFromCloneRequest(t *testing.T) {
	outputWriter := io.Discard
	request := &domain.CloneRequest{
		Paths:           []string{"/test/path1", "/test/path2"},
		Recursive:       false,
		IncludePatterns: []string{"*.py", "*.pyx"},
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
	assert.Equal(t, request.Recursive, cloneConfig.Input.Recursive)
	assert.Equal(t, request.IncludePatterns, cloneConfig.Input.IncludePatterns)
	assert.Equal(t, request.ExcludePatterns, cloneConfig.Input.ExcludePatterns)
	
	// Verify analysis conversion
	assert.Equal(t, request.MinLines, cloneConfig.Analysis.MinLines)
	assert.Equal(t, request.MinNodes, cloneConfig.Analysis.MinNodes)
	assert.Equal(t, request.MaxEditDistance, cloneConfig.Analysis.MaxEditDistance)
	assert.Equal(t, request.IgnoreLiterals, cloneConfig.Analysis.IgnoreLiterals)
	assert.Equal(t, request.IgnoreIdentifiers, cloneConfig.Analysis.IgnoreIdentifiers)
	
	// Verify thresholds conversion
	assert.Equal(t, request.Type1Threshold, cloneConfig.Thresholds.Type1Threshold)
	assert.Equal(t, request.Type2Threshold, cloneConfig.Thresholds.Type2Threshold)
	assert.Equal(t, request.Type3Threshold, cloneConfig.Thresholds.Type3Threshold)
	assert.Equal(t, request.Type4Threshold, cloneConfig.Thresholds.Type4Threshold)
	assert.Equal(t, request.SimilarityThreshold, cloneConfig.Thresholds.SimilarityThreshold)
	
	// Verify output conversion
	assert.Equal(t, "yaml", cloneConfig.Output.Format)
	assert.Equal(t, outputWriter, cloneConfig.Output.Writer)
	assert.Equal(t, request.ShowDetails, cloneConfig.Output.ShowDetails)
	assert.Equal(t, request.ShowContent, cloneConfig.Output.ShowContent)
	assert.Equal(t, "size", cloneConfig.Output.SortBy)
	assert.Equal(t, request.GroupClones, cloneConfig.Output.GroupClones)
	
	// Verify filtering conversion
	assert.Equal(t, request.MinSimilarity, cloneConfig.Filtering.MinSimilarity)
	assert.Equal(t, request.MaxSimilarity, cloneConfig.Filtering.MaxSimilarity)
	expectedCloneTypes := []string{"type1", "type3", "type4"}
	assert.Equal(t, expectedCloneTypes, cloneConfig.Filtering.EnabledCloneTypes)
}

func TestRoundTripConversion(t *testing.T) {
	t.Run("CloneDetectorConfig roundtrip", func(t *testing.T) {
		original := CloneDetectorConfig{
			MinLines:          7,
			MinNodes:          14,
			Type1Threshold:    constants.DefaultType1CloneThreshold,
			Type2Threshold:    constants.DefaultType2CloneThreshold,
			Type3Threshold:    constants.DefaultType3CloneThreshold,
			Type4Threshold:    constants.DefaultType4CloneThreshold,
			MaxEditDistance:   30.0,
			IgnoreLiterals:    false,
			IgnoreIdentifiers: true,
			CostModelType:     "python",
		}
		
		// Convert to unified config and back
		unified := FromCloneDetectorConfig(original)
		roundtrip := unified.ToCloneDetectorConfig()
		
		assert.Equal(t, original, roundtrip)
	})
	
	t.Run("CloneDetectionConfig roundtrip", func(t *testing.T) {
		original := CloneDetectionConfig{
			Enabled:             true,
			MinLines:            6,
			MinNodes:            12,
			Type1Threshold:      0.99,
			Type2Threshold:      0.89,
			Type3Threshold:      0.79,
			Type4Threshold:      0.69,
			SimilarityThreshold: 0.65,
			MaxEditDistance:     25.0,
			CostModelType:       "weighted",
			IgnoreLiterals:      true,
			IgnoreIdentifiers:   true,
			ShowContent:         false,
			GroupClones:         true,
			SortBy:              "similarity",
			MinSimilarity:       0.1,
			MaxSimilarity:       0.95,
			CloneTypes:          []string{"type2", "type4"},
		}
		
		// Convert to unified config and back
		unified := FromCloneDetectionConfig(original)
		roundtrip := unified.ToCloneDetectionConfig()
		
		assert.Equal(t, original, roundtrip)
	})
}

func TestDefaultConfigBackwardCompatibility(t *testing.T) {
	config := DefaultConfig()
	
	// Verify that both old and new config fields are populated
	require.NotNil(t, config.Clones)
	require.NotNil(t, config.CloneDetection)
	
	// Verify they contain the same values
	assert.Equal(t, config.Clones.Analysis.MinLines, config.CloneDetection.MinLines)
	assert.Equal(t, config.Clones.Analysis.MinNodes, config.CloneDetection.MinNodes)
	assert.Equal(t, config.Clones.Thresholds.Type1Threshold, config.CloneDetection.Type1Threshold)
	assert.Equal(t, config.Clones.Thresholds.Type2Threshold, config.CloneDetection.Type2Threshold)
	assert.Equal(t, config.Clones.Thresholds.Type3Threshold, config.CloneDetection.Type3Threshold)
	assert.Equal(t, config.Clones.Thresholds.Type4Threshold, config.CloneDetection.Type4Threshold)
}

func TestCloneConfigValidation(t *testing.T) {
	t.Run("Valid default config", func(t *testing.T) {
		config := DefaultCloneConfig()
		err := config.Validate()
		assert.NoError(t, err)
	})
	
	t.Run("Invalid analysis config", func(t *testing.T) {
		config := DefaultCloneConfig()
		config.Analysis.MinLines = 0 // Invalid
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "analysis config invalid")
	})
	
	t.Run("Invalid threshold config", func(t *testing.T) {
		config := DefaultCloneConfig()
		config.Thresholds.Type1Threshold = -0.1 // Invalid range
		
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "thresholds config invalid")
	})
}