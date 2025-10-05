package constants

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
	// added or removed statements. Medium-high similarity (≥80%).
	DefaultType3CloneThreshold = 0.80

	// DefaultType4CloneThreshold represents the similarity threshold for Type-4 clones.
	// Type-4 clones are syntactically different but functionally similar fragments.
	// They perform the same computation but through different syntactic variants.
	// Medium similarity (≥70%).
	DefaultType4CloneThreshold = 0.75
)

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
