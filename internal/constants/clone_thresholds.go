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

// DFA (Data Flow Analysis) Feature Weights for semantic similarity comparison.
// These weights determine how each DFA metric contributes to the overall similarity score.
const (
	// DefaultDFAPairCountWeight is the weight for total def-use pair count similarity.
	// Higher pair count often indicates more complex data flow, which is a key semantic indicator.
	DefaultDFAPairCountWeight = 0.25

	// DefaultDFAChainLengthWeight is the weight for average chain length similarity.
	// Chain length (uses per definition) indicates variable reuse patterns.
	DefaultDFAChainLengthWeight = 0.20

	// DefaultDFACrossBlockWeight is the weight for cross-block pair ratio similarity.
	// Cross-block pairs indicate data dependencies across control flow, a key structural pattern.
	DefaultDFACrossBlockWeight = 0.20

	// DefaultDFADefKindWeight is the weight for definition kind distribution similarity.
	// Different definition kinds (assign, param, loop) indicate different coding patterns.
	DefaultDFADefKindWeight = 0.20

	// DefaultDFAUseKindWeight is the weight for use kind distribution similarity.
	// How variables are used (read, call, attribute) reflects access patterns.
	DefaultDFAUseKindWeight = 0.15
)

// Combined Semantic Feature Weights for Type-4 clone detection.
// When DFA analysis is enabled, both CFG and DFA features contribute to similarity.
const (
	// DefaultCFGFeatureWeight is the weight for CFG-based features in semantic similarity.
	// CFG captures control flow structure (branches, loops, etc.).
	DefaultCFGFeatureWeight = 0.60

	// DefaultDFAFeatureWeight is the weight for DFA-based features in semantic similarity.
	// DFA captures data flow patterns (variable definitions and uses).
	DefaultDFAFeatureWeight = 0.40
)
