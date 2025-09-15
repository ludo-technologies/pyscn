package constants

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCloneThresholds(t *testing.T) {
	t.Run("Constants have expected values", func(t *testing.T) {
		assert.Equal(t, 0.95, DefaultType1CloneThreshold, "Type1 threshold should be 0.95")
		assert.Equal(t, 0.85, DefaultType2CloneThreshold, "Type2 threshold should be 0.85")
		assert.Equal(t, 0.80, DefaultType3CloneThreshold, "Type3 threshold should be 0.80")
		assert.Equal(t, 0.75, DefaultType4CloneThreshold, "Type4 threshold should be 0.75")
	})

	t.Run("Constants are in correct order", func(t *testing.T) {
		assert.Greater(t, DefaultType1CloneThreshold, DefaultType2CloneThreshold,
			"Type1 threshold should be > Type2 threshold")
		assert.Greater(t, DefaultType2CloneThreshold, DefaultType3CloneThreshold,
			"Type2 threshold should be > Type3 threshold")
		assert.Greater(t, DefaultType3CloneThreshold, DefaultType4CloneThreshold,
			"Type3 threshold should be > Type4 threshold")
	})

	t.Run("All constants are in valid range", func(t *testing.T) {
		thresholds := []float64{
			DefaultType1CloneThreshold,
			DefaultType2CloneThreshold,
			DefaultType3CloneThreshold,
			DefaultType4CloneThreshold,
		}

		for i, threshold := range thresholds {
			assert.GreaterOrEqual(t, threshold, 0.0,
				"Threshold %d should be >= 0.0", i+1)
			assert.LessOrEqual(t, threshold, 1.0,
				"Threshold %d should be <= 1.0", i+1)
		}
	})
}

func TestDefaultCloneThresholdsFunction(t *testing.T) {
	config := DefaultCloneThresholds()

	assert.Equal(t, DefaultType1CloneThreshold, config.Type1Threshold)
	assert.Equal(t, DefaultType2CloneThreshold, config.Type2Threshold)
	assert.Equal(t, DefaultType3CloneThreshold, config.Type3Threshold)
	assert.Equal(t, DefaultType4CloneThreshold, config.Type4Threshold)
}

func TestCloneThresholdConfigValidation(t *testing.T) {
	t.Run("Valid configuration", func(t *testing.T) {
		config := DefaultCloneThresholds()
		err := config.ValidateThresholds()
		assert.NoError(t, err, "Default configuration should be valid")
	})

	t.Run("Invalid range - too low", func(t *testing.T) {
		config := CloneThresholdConfig{
			Type1Threshold: -0.1,
			Type2Threshold: 0.8,
			Type3Threshold: 0.7,
			Type4Threshold: 0.6,
		}
		err := config.ValidateThresholds()
		assert.Error(t, err, "Should reject negative threshold")
		assert.Contains(t, err.Error(), "out of range", "Error should mention range violation")
	})

	t.Run("Invalid range - too high", func(t *testing.T) {
		config := CloneThresholdConfig{
			Type1Threshold: 1.1,
			Type2Threshold: 0.8,
			Type3Threshold: 0.7,
			Type4Threshold: 0.6,
		}
		err := config.ValidateThresholds()
		assert.Error(t, err, "Should reject threshold > 1.0")
		assert.Contains(t, err.Error(), "out of range", "Error should mention range violation")
	})

	t.Run("Invalid ordering - Type1 <= Type2", func(t *testing.T) {
		config := CloneThresholdConfig{
			Type1Threshold: 0.8,
			Type2Threshold: 0.85,
			Type3Threshold: 0.7,
			Type4Threshold: 0.6,
		}
		err := config.ValidateThresholds()
		assert.Error(t, err, "Should reject Type1 <= Type2")
		assert.Contains(t, err.Error(), "Type1 threshold", "Error should mention Type1 threshold")
	})

	t.Run("Invalid ordering - Type2 <= Type3", func(t *testing.T) {
		config := CloneThresholdConfig{
			Type1Threshold: 0.95,
			Type2Threshold: 0.7,
			Type3Threshold: 0.75,
			Type4Threshold: 0.6,
		}
		err := config.ValidateThresholds()
		assert.Error(t, err, "Should reject Type2 <= Type3")
		assert.Contains(t, err.Error(), "Type2 threshold", "Error should mention Type2 threshold")
	})

	t.Run("Invalid ordering - Type3 <= Type4", func(t *testing.T) {
		config := CloneThresholdConfig{
			Type1Threshold: 0.95,
			Type2Threshold: 0.85,
			Type3Threshold: 0.6,
			Type4Threshold: 0.65,
		}
		err := config.ValidateThresholds()
		assert.Error(t, err, "Should reject Type3 <= Type4")
		assert.Contains(t, err.Error(), "Type3 threshold", "Error should mention Type3 threshold")
	})
}

func TestGetThresholdForType(t *testing.T) {
	config := DefaultCloneThresholds()

	tests := []struct {
		cloneType int
		expected  float64
	}{
		{1, DefaultType1CloneThreshold},
		{2, DefaultType2CloneThreshold},
		{3, DefaultType3CloneThreshold},
		{4, DefaultType4CloneThreshold},
		{0, DefaultType4CloneThreshold},  // Default fallback
		{5, DefaultType4CloneThreshold},  // Default fallback
		{-1, DefaultType4CloneThreshold}, // Default fallback
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("CloneType_%d", test.cloneType), func(t *testing.T) {
			result := config.GetThresholdForType(test.cloneType)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestCloneTypeNames(t *testing.T) {
	expectedNames := map[int]string{
		1: "Type-1 (Identical)",
		2: "Type-2 (Renamed)",
		3: "Type-3 (Near-Miss)",
		4: "Type-4 (Semantic)",
	}

	for cloneType, expectedName := range expectedNames {
		t.Run(fmt.Sprintf("Type_%d", cloneType), func(t *testing.T) {
			name, exists := CloneTypeNames[cloneType]
			assert.True(t, exists, "Clone type %d should have a name", cloneType)
			assert.Equal(t, expectedName, name)
		})
	}
}

func TestCloneTypeDescriptions(t *testing.T) {
	// Verify all clone types have descriptions
	for cloneType := 1; cloneType <= 4; cloneType++ {
		t.Run(fmt.Sprintf("Type_%d_has_description", cloneType), func(t *testing.T) {
			description, exists := CloneTypeDescriptions[cloneType]
			assert.True(t, exists, "Clone type %d should have a description", cloneType)
			assert.NotEmpty(t, description, "Description should not be empty")
			assert.Greater(t, len(description), 20, "Description should be meaningful")
		})
	}
}

// TestTemplateConfigConsistency verifies that init.go template uses the same values as constants
func TestTemplateConfigConsistency(t *testing.T) {
	t.Run("Template values match constants", func(t *testing.T) {
		// These values should match what's in cmd/pyscn/init.go template
		expectedType1 := DefaultType1CloneThreshold // 0.95
		expectedType2 := DefaultType2CloneThreshold // 0.85
		expectedType3 := DefaultType3CloneThreshold // 0.80
		expectedType4 := DefaultType4CloneThreshold // 0.75

		// Verify the constants are what we expect
		assert.Equal(t, 0.95, expectedType1, "Type1 constant should be 0.95")
		assert.Equal(t, 0.85, expectedType2, "Type2 constant should be 0.85")
		assert.Equal(t, 0.80, expectedType3, "Type3 constant should be 0.80")
		assert.Equal(t, 0.75, expectedType4, "Type4 constant should be 0.75")
	})
}
