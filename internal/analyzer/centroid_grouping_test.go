package analyzer

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
)

func TestCentroidGroupingThresholds(t *testing.T) {
	t.Run("Default thresholds are set correctly", func(t *testing.T) {
		grouping := NewCentroidGrouping(0.7)

		assert.Equal(t, domain.DefaultType1CloneThreshold, grouping.type1Threshold, "Type1 threshold should use domain.DefaultType1CloneThreshold")
		assert.Equal(t, domain.DefaultType2CloneThreshold, grouping.type2Threshold, "Type2 threshold should use domain.DefaultType2CloneThreshold")
		assert.Equal(t, domain.DefaultType3CloneThreshold, grouping.type3Threshold, "Type3 threshold should use domain.DefaultType3CloneThreshold")
		assert.Equal(t, domain.DefaultType4CloneThreshold, grouping.type4Threshold, "Type4 threshold should use domain.DefaultType4CloneThreshold")
	})

	t.Run("SetThresholds updates values correctly", func(t *testing.T) {
		grouping := NewCentroidGrouping(0.7)

		// Set custom thresholds
		grouping.SetThresholds(0.98, 0.88, 0.78, 0.68)

		assert.Equal(t, 0.98, grouping.type1Threshold, "Type1 threshold should be updated")
		assert.Equal(t, 0.88, grouping.type2Threshold, "Type2 threshold should be updated")
		assert.Equal(t, 0.78, grouping.type3Threshold, "Type3 threshold should be updated")
		assert.Equal(t, 0.68, grouping.type4Threshold, "Type4 threshold should be updated")
	})

	t.Run("Clone type classification uses configured thresholds", func(t *testing.T) {
		grouping := NewCentroidGrouping(0.7)

		// Set custom thresholds for testing
		grouping.SetThresholds(0.90, 0.80, 0.70, 0.60)

		// Create a test group with controlled similarity
		testCases := []struct {
			similarity   float64
			expectedType CloneType
			description  string
		}{
			{0.95, Type1Clone, "High similarity should be Type1"},
			{0.85, Type2Clone, "Medium-high similarity should be Type2"},
			{0.75, Type3Clone, "Medium similarity should be Type3"},
			{0.65, Type4Clone, "Medium-low similarity should be Type4"},
			{0.50, Type4Clone, "Low similarity should be Type4 (default)"},
		}

		for _, tc := range testCases {
			t.Run(tc.description, func(t *testing.T) {
				// Create a test group and directly test classification
				group := &CloneGroup{}

				// Directly classify based on similarity since we can't create proper tree nodes
				grouping.classifyGroupBySimilarity(group, tc.similarity)
				assert.Equal(t, tc.expectedType, group.CloneType, tc.description)
			})
		}
	})
}

func TestCentroidGroupingConsistency(t *testing.T) {
	t.Run("Thresholds should be in descending order", func(t *testing.T) {
		grouping := NewCentroidGrouping(0.7)

		assert.Greater(t, grouping.type1Threshold, grouping.type2Threshold,
			"Type1 threshold should be > Type2 threshold")
		assert.Greater(t, grouping.type2Threshold, grouping.type3Threshold,
			"Type2 threshold should be > Type3 threshold")
		assert.Greater(t, grouping.type3Threshold, grouping.type4Threshold,
			"Type3 threshold should be > Type4 threshold")
	})

	t.Run("GetName returns correct value", func(t *testing.T) {
		grouping := NewCentroidGrouping(0.7)
		assert.Equal(t, "Centroid-based", grouping.GetName())
	})
}
