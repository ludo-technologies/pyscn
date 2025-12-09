package domain

import (
	"testing"
)

// TestDefaultValueConsistency ensures all default values are properly defined
// and maintain expected relationships
func TestDefaultValueConsistency(t *testing.T) {
	t.Run("Clone thresholds are properly ordered", func(t *testing.T) {
		// Type1 > Type2 > Type3 > Type4 (higher is more restrictive)
		if DefaultType1CloneThreshold <= DefaultType2CloneThreshold {
			t.Errorf("Type1 threshold (%.2f) should be > Type2 threshold (%.2f)",
				DefaultType1CloneThreshold, DefaultType2CloneThreshold)
		}
		if DefaultType2CloneThreshold <= DefaultType3CloneThreshold {
			t.Errorf("Type2 threshold (%.2f) should be > Type3 threshold (%.2f)",
				DefaultType2CloneThreshold, DefaultType3CloneThreshold)
		}
		if DefaultType3CloneThreshold <= DefaultType4CloneThreshold {
			t.Errorf("Type3 threshold (%.2f) should be > Type4 threshold (%.2f)",
				DefaultType3CloneThreshold, DefaultType4CloneThreshold)
		}
	})

	t.Run("Clone thresholds are within valid range", func(t *testing.T) {
		thresholds := []struct {
			name  string
			value float64
		}{
			{"Type1", DefaultType1CloneThreshold},
			{"Type2", DefaultType2CloneThreshold},
			{"Type3", DefaultType3CloneThreshold},
			{"Type4", DefaultType4CloneThreshold},
			{"Similarity", DefaultCloneSimilarityThreshold},
			{"Grouping", DefaultCloneGroupingThreshold},
			{"LSHSimilarity", DefaultLSHSimilarityThreshold},
		}

		for _, th := range thresholds {
			if th.value < 0.0 || th.value > 1.0 {
				t.Errorf("%s threshold (%.2f) is outside valid range [0.0, 1.0]", th.name, th.value)
			}
		}
	})

	t.Run("Complexity thresholds are properly ordered", func(t *testing.T) {
		if DefaultComplexityLowThreshold >= DefaultComplexityMediumThreshold {
			t.Errorf("Low complexity threshold (%d) should be < medium threshold (%d)",
				DefaultComplexityLowThreshold, DefaultComplexityMediumThreshold)
		}
	})

	t.Run("CBO thresholds are properly ordered", func(t *testing.T) {
		if DefaultCBOLowThreshold >= DefaultCBOMediumThreshold {
			t.Errorf("CBO low threshold (%d) should be < medium threshold (%d)",
				DefaultCBOLowThreshold, DefaultCBOMediumThreshold)
		}
	})

	t.Run("DFA weights sum to 1.0", func(t *testing.T) {
		sum := DefaultDFAPairCountWeight +
			DefaultDFAChainLengthWeight +
			DefaultDFACrossBlockWeight +
			DefaultDFADefKindWeight +
			DefaultDFAUseKindWeight

		if sum < 0.99 || sum > 1.01 {
			t.Errorf("DFA weights should sum to 1.0, got %.2f", sum)
		}
	})

	t.Run("CFG/DFA feature weights sum to 1.0", func(t *testing.T) {
		sum := DefaultCFGFeatureWeight + DefaultDFAFeatureWeight
		if sum < 0.99 || sum > 1.01 {
			t.Errorf("CFG/DFA weights should sum to 1.0, got %.2f", sum)
		}
	})

	t.Run("Performance defaults are positive", func(t *testing.T) {
		if DefaultMaxMemoryMB <= 0 {
			t.Errorf("MaxMemoryMB (%d) should be > 0", DefaultMaxMemoryMB)
		}
		if DefaultBatchSize <= 0 {
			t.Errorf("BatchSize (%d) should be > 0", DefaultBatchSize)
		}
		if DefaultMaxGoroutines <= 0 {
			t.Errorf("MaxGoroutines (%d) should be > 0", DefaultMaxGoroutines)
		}
		if DefaultTimeoutSeconds <= 0 {
			t.Errorf("TimeoutSeconds (%d) should be > 0", DefaultTimeoutSeconds)
		}
	})

	t.Run("LSH parameters are valid", func(t *testing.T) {
		if DefaultLSHAutoThreshold <= 0 {
			t.Errorf("LSHAutoThreshold (%d) should be > 0", DefaultLSHAutoThreshold)
		}
		if DefaultLSHBands <= 0 {
			t.Errorf("LSHBands (%d) should be > 0", DefaultLSHBands)
		}
		if DefaultLSHRows <= 0 {
			t.Errorf("LSHRows (%d) should be > 0", DefaultLSHRows)
		}
		if DefaultLSHHashes <= 0 {
			t.Errorf("LSHHashes (%d) should be > 0", DefaultLSHHashes)
		}
	})

	t.Run("Clone analysis defaults are positive", func(t *testing.T) {
		if DefaultCloneMinLines <= 0 {
			t.Errorf("CloneMinLines (%d) should be > 0", DefaultCloneMinLines)
		}
		if DefaultCloneMinNodes <= 0 {
			t.Errorf("CloneMinNodes (%d) should be > 0", DefaultCloneMinNodes)
		}
		if DefaultCloneMaxEditDistance <= 0 {
			t.Errorf("CloneMaxEditDistance (%.2f) should be > 0", DefaultCloneMaxEditDistance)
		}
	})

	t.Run("Grouping threshold equals Type3 threshold", func(t *testing.T) {
		if DefaultCloneGroupingThreshold != DefaultType3CloneThreshold {
			t.Errorf("Grouping threshold (%.2f) should equal Type3 threshold (%.2f)",
				DefaultCloneGroupingThreshold, DefaultType3CloneThreshold)
		}
	})
}

// TestExpectedDefaultValues verifies that default values match expected industry standards
func TestExpectedDefaultValues(t *testing.T) {
	t.Run("Clone type thresholds match academic standards", func(t *testing.T) {
		// Based on Roy & Cordy (2007) and Bellon et al. (2007)
		if DefaultType1CloneThreshold != 0.98 {
			t.Errorf("Type1 threshold should be 0.98, got %.2f", DefaultType1CloneThreshold)
		}
		if DefaultType2CloneThreshold != 0.95 {
			t.Errorf("Type2 threshold should be 0.95, got %.2f", DefaultType2CloneThreshold)
		}
		if DefaultType3CloneThreshold != 0.80 {
			t.Errorf("Type3 threshold should be 0.80, got %.2f", DefaultType3CloneThreshold)
		}
		if DefaultType4CloneThreshold != 0.75 {
			t.Errorf("Type4 threshold should be 0.75, got %.2f", DefaultType4CloneThreshold)
		}
	})

	t.Run("CBO thresholds match industry standards", func(t *testing.T) {
		// Based on Chidamber & Kemerer (1994)
		if DefaultCBOLowThreshold != 3 {
			t.Errorf("CBO low threshold should be 3, got %d", DefaultCBOLowThreshold)
		}
		if DefaultCBOMediumThreshold != 7 {
			t.Errorf("CBO medium threshold should be 7, got %d", DefaultCBOMediumThreshold)
		}
	})

	t.Run("Complexity thresholds match McCabe standards", func(t *testing.T) {
		// Based on McCabe (1976)
		if DefaultComplexityLowThreshold != 9 {
			t.Errorf("Complexity low threshold should be 9, got %d", DefaultComplexityLowThreshold)
		}
		if DefaultComplexityMediumThreshold != 19 {
			t.Errorf("Complexity medium threshold should be 19, got %d", DefaultComplexityMediumThreshold)
		}
	})
}
