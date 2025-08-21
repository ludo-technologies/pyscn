package constants

import "fmt"

// CloneThresholds defines the standard similarity thresholds for different types of code clones.
// These values are based on research in clone detection and represent industry standards.
//
// References:
// - Roy, C. K., & Cordy, J. R. (2007). A survey on software clone detection research
// - Bellon, S., et al. (2007). Comparison and evaluation of clone detection tools
const (
	// DefaultType1CloneThreshold represents the similarity threshold for Type-1 clones.
	// Type-1 clones are identical code fragments except for variations in whitespace,
	// layout and comments. They should have very high similarity (≥95%).
	DefaultType1CloneThreshold = 0.95

	// DefaultType2CloneThreshold represents the similarity threshold for Type-2 clones.
	// Type-2 clones are syntactically identical fragments except for variations in
	// identifiers, literals, types, layout and comments. High similarity (≥85%).
	DefaultType2CloneThreshold = 0.85

	// DefaultType3CloneThreshold represents the similarity threshold for Type-3 clones.
	// Type-3 clones are copied fragments with further modifications such as changed,
	// added or removed statements. Medium-high similarity (≥70%).
	DefaultType3CloneThreshold = 0.70

	// DefaultType4CloneThreshold represents the similarity threshold for Type-4 clones.
	// Type-4 clones are syntactically different but functionally similar fragments.
	// They perform the same computation but through different syntactic variants.
	// Medium similarity (≥60%).
	DefaultType4CloneThreshold = 0.60
)

// CloneThresholdConfig holds all clone detection threshold values
type CloneThresholdConfig struct {
	Type1Threshold float64
	Type2Threshold float64
	Type3Threshold float64
	Type4Threshold float64
}

// DefaultCloneThresholds returns the default clone detection thresholds
func DefaultCloneThresholds() CloneThresholdConfig {
	return CloneThresholdConfig{
		Type1Threshold: DefaultType1CloneThreshold,
		Type2Threshold: DefaultType2CloneThreshold,
		Type3Threshold: DefaultType3CloneThreshold,
		Type4Threshold: DefaultType4CloneThreshold,
	}
}

// ValidateThresholds validates that clone thresholds are in correct order and range
func (c *CloneThresholdConfig) ValidateThresholds() error {
	// Check range
	thresholds := []float64{c.Type1Threshold, c.Type2Threshold, c.Type3Threshold, c.Type4Threshold}
	for i, threshold := range thresholds {
		if threshold < 0.0 || threshold > 1.0 {
			return fmt.Errorf("threshold %d is out of range [0.0, 1.0]: %f", i+1, threshold)
		}
	}

	// Check ordering: Type1 > Type2 > Type3 > Type4
	if c.Type1Threshold <= c.Type2Threshold {
		return fmt.Errorf("Type1 threshold (%.3f) should be > Type2 threshold (%.3f)", c.Type1Threshold, c.Type2Threshold)
	}
	if c.Type2Threshold <= c.Type3Threshold {
		return fmt.Errorf("Type2 threshold (%.3f) should be > Type3 threshold (%.3f)", c.Type2Threshold, c.Type3Threshold)
	}
	if c.Type3Threshold <= c.Type4Threshold {
		return fmt.Errorf("Type3 threshold (%.3f) should be > Type4 threshold (%.3f)", c.Type3Threshold, c.Type4Threshold)
	}

	return nil
}

// GetThresholdForType returns the threshold for a specific clone type
func (c *CloneThresholdConfig) GetThresholdForType(cloneType int) float64 {
	switch cloneType {
	case 1:
		return c.Type1Threshold
	case 2:
		return c.Type2Threshold
	case 3:
		return c.Type3Threshold
	case 4:
		return c.Type4Threshold
	default:
		return c.Type4Threshold // Default to most permissive
	}
}

// CloneTypeNames provides human-readable names for clone types
var CloneTypeNames = map[int]string{
	1: "Type-1 (Identical)",
	2: "Type-2 (Renamed)",
	3: "Type-3 (Near-Miss)",
	4: "Type-4 (Semantic)",
}

// CloneTypeDescriptions provides detailed descriptions for each clone type
var CloneTypeDescriptions = map[int]string{
	1: "Identical code fragments except for whitespace, layout and comments",
	2: "Syntactically identical except for variations in identifiers, literals and types",
	3: "Copied fragments with further modifications (changed, added or removed statements)",
	4: "Syntactically different but functionally similar code fragments",
}
