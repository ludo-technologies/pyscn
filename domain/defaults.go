package domain

// CloneThresholds defines the standard similarity thresholds for different types of code clones.
// These values are based on research in clone detection and represent industry standards.
//
// References:
// - Roy, C. K., & Cordy, J. R. (2007). A survey on software clone detection research
// - Bellon, S., et al. (2007). Comparison and evaluation of clone detection tools
const (
	// DefaultType1CloneThreshold represents the similarity threshold for Type-1 clones.
	// Type-1 clones are identical code fragments except for variations in whitespace,
	// layout and comments. They should have very high similarity (≥98%).
	DefaultType1CloneThreshold = 0.98

	// DefaultType2CloneThreshold represents the similarity threshold for Type-2 clones.
	// Type-2 clones are syntactically identical fragments except for variations in
	// identifiers, literals, types, layout and comments. High similarity (≥95%).
	DefaultType2CloneThreshold = 0.95

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

// ============================================================================
// Complexity Analysis Defaults
// ============================================================================

// Complexity thresholds based on McCabe cyclomatic complexity.
// Reference: McCabe, T.J. (1976). A Complexity Measure
const (
	// DefaultComplexityLowThreshold is the upper bound for low-risk complexity.
	// Functions with complexity <= 9 are considered simple and maintainable.
	DefaultComplexityLowThreshold = 9

	// DefaultComplexityMediumThreshold is the upper bound for medium-risk complexity.
	// Functions with complexity 10-19 require attention but are acceptable.
	DefaultComplexityMediumThreshold = 19

	// DefaultComplexityMinFilter is the minimum complexity to include in reports.
	// Default 1 means all functions are included.
	DefaultComplexityMinFilter = 1

	// DefaultComplexityMaxLimit is the maximum complexity limit for enforcement.
	// 0 means no limit is enforced.
	DefaultComplexityMaxLimit = 0
)

// ============================================================================
// CBO (Coupling Between Objects) Defaults
// ============================================================================

// CBO thresholds based on Chidamber & Kemerer metrics suite.
// Reference: Chidamber, S.R. & Kemerer, C.F. (1994). A Metrics Suite for OOD
const (
	// DefaultCBOLowThreshold is the upper bound for low-risk coupling.
	// Classes with CBO <= 3 are well-encapsulated.
	DefaultCBOLowThreshold = 3

	// DefaultCBOMediumThreshold is the upper bound for medium-risk coupling.
	// Classes with CBO 4-7 may need refactoring consideration.
	DefaultCBOMediumThreshold = 7
)

// ============================================================================
// Dead Code Detection Defaults
// ============================================================================

const (
	// DefaultDeadCodeMinSeverity is the minimum severity level for dead code reports.
	// Options: "info", "warning", "critical"
	DefaultDeadCodeMinSeverity = "warning"

	// DefaultDeadCodeContextLines is the number of context lines shown around dead code.
	DefaultDeadCodeContextLines = 3

	// DefaultDeadCodeSortBy is the default sort order for dead code results.
	// Options: "severity", "file", "line"
	DefaultDeadCodeSortBy = "severity"
)

// ============================================================================
// Clone Analysis Defaults
// ============================================================================

const (
	// DefaultCloneMinLines is the minimum number of lines for a code fragment to be considered.
	DefaultCloneMinLines = 5

	// DefaultCloneMinNodes is the minimum number of AST nodes for a code fragment.
	DefaultCloneMinNodes = 10

	// DefaultCloneMaxEditDistance is the maximum tree edit distance for clone comparison.
	DefaultCloneMaxEditDistance = 50.0

	// DefaultCloneSimilarityThreshold is the general similarity threshold for clone detection.
	DefaultCloneSimilarityThreshold = 0.9

	// DefaultCloneGroupingThreshold is the threshold for grouping related clones.
	// Uses Type-3 threshold as default for grouping similar code.
	DefaultCloneGroupingThreshold = DefaultType3CloneThreshold
)

// ============================================================================
// LSH (Locality-Sensitive Hashing) Acceleration Defaults
// ============================================================================

const (
	// DefaultLSHAutoThreshold is the file count threshold for automatic LSH activation.
	// When file count exceeds this, LSH acceleration is automatically enabled.
	DefaultLSHAutoThreshold = 500

	// DefaultLSHSimilarityThreshold is the minimum similarity for LSH candidate filtering.
	DefaultLSHSimilarityThreshold = 0.50

	// DefaultLSHBands is the number of bands in the LSH algorithm.
	DefaultLSHBands = 32

	// DefaultLSHRows is the number of rows per band in the LSH algorithm.
	DefaultLSHRows = 4

	// DefaultLSHHashes is the total number of hash functions used.
	DefaultLSHHashes = 128
)

// ============================================================================
// Performance Defaults
// ============================================================================

const (
	// DefaultMaxMemoryMB is the maximum memory usage in megabytes for batch processing.
	DefaultMaxMemoryMB = 100

	// DefaultBatchSize is the default batch size for processing files.
	DefaultBatchSize = 100

	// DefaultMaxGoroutines is the default number of concurrent goroutines.
	DefaultMaxGoroutines = 4

	// DefaultTimeoutSeconds is the default timeout in seconds for analysis operations.
	DefaultTimeoutSeconds = 300
)

// ============================================================================
// DI Anti-pattern Detection Defaults
// ============================================================================

const (
	// DefaultDIConstructorParamThreshold is the maximum allowed constructor parameters.
	// Classes with more than 5 parameters in __init__ are flagged as constructor over-injection.
	// Reference: Martin, R.C. (2008). Clean Code - recommends max 3-4, we use 5 as threshold.
	DefaultDIConstructorParamThreshold = 5

	// DefaultDIMinSeverity is the minimum severity level for DI anti-pattern reports.
	// Options: "info", "warning", "error"
	DefaultDIMinSeverity = "warning"
)

// ServiceLocatorMethodNames returns the method names that indicate service locator pattern
func ServiceLocatorMethodNames() []string {
	return []string{
		"get_service",
		"resolve",
		"get_instance",
		"locate",
		"get_component",
		"get_dependency",
		"get_bean",
		"inject",
		"container.get",
		"locator.get",
		"registry.get",
		"ioc.resolve",
	}
}

// ============================================================================
// Mock Data Detection Defaults
// ============================================================================

const (
	// DefaultMockDataMinSeverity is the minimum severity level for mock data reports.
	// Options: "info", "warning", "error"
	DefaultMockDataMinSeverity = "warning"

	// DefaultMockDataSortBy is the default sort order for mock data results.
	// Options: "severity", "file", "line", "type"
	DefaultMockDataSortBy = "severity"

	// DefaultMockDataIgnoreTests determines whether test files are ignored by default.
	DefaultMockDataIgnoreTests = true
)

// DefaultMockDataKeywords returns the default keywords used to detect mock data.
// These are common identifiers used in placeholder/mock data.
func DefaultMockDataKeywords() []string {
	return []string{
		"mock", "fake", "dummy", "test", "sample", "example",
		"placeholder", "stub", "fixture", "temp", "tmp",
		"foo", "bar", "baz", "qux", "lorem", "ipsum",
	}
}

// DefaultMockDataDomains returns the default domains that indicate mock data.
// These are reserved domains per RFC 2606 and common test domains.
func DefaultMockDataDomains() []string {
	return []string{
		"example.com", "example.org", "example.net",
		"test.com", "test.org", "test.net",
		"localhost", "invalid",
		"foo.com", "bar.com",
	}
}

// DefaultMockDataTestPatterns returns the default patterns for test files to ignore.
func DefaultMockDataTestPatterns() []string {
	return []string{
		"test_*.py",
		"*_test.py",
		"tests/",
		"test/",
		"testing/",
		"__tests__/",
		"conftest.py",
	}
}
