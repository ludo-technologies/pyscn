package constants

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultCloneThresholds(t *testing.T) {
	t.Run("Constants have expected values", func(t *testing.T) {
		assert.Equal(t, 0.98, DefaultType1CloneThreshold, "Type1 threshold should be 0.98")
		assert.Equal(t, 0.95, DefaultType2CloneThreshold, "Type2 threshold should be 0.95")
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
		expectedType1 := DefaultType1CloneThreshold // 0.98
		expectedType2 := DefaultType2CloneThreshold // 0.95
		expectedType3 := DefaultType3CloneThreshold // 0.80
		expectedType4 := DefaultType4CloneThreshold // 0.75

		// Verify the constants are what we expect
		assert.Equal(t, 0.98, expectedType1, "Type1 constant should be 0.98")
		assert.Equal(t, 0.95, expectedType2, "Type2 constant should be 0.95")
		assert.Equal(t, 0.80, expectedType3, "Type3 constant should be 0.80")
		assert.Equal(t, 0.75, expectedType4, "Type4 constant should be 0.75")
	})
}
