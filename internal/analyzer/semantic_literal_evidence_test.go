package analyzer

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/require"
)

// Regression fixtures for issue #482: two apispec-style field converters from
// flask-smorest. Identical control-flow template (isinstance check + version
// branch filling a dict), but the field types and spec keys are entirely
// different, so they are not semantic clones.
const delimitedListConverterSrc = `def delimited_list2param(self, field, **kwargs):
    """Document DelimitedList field as a query parameter."""
    ret = {}
    if isinstance(field, DelimitedList):
        if self.openapi_version.major < 3:
            ret["collectionFormat"] = "csv"
        else:
            ret["explode"] = False
            ret["style"] = "form"
    return ret
`

const uploadFieldConverterSrc = `def uploadfield2properties(self, field, **kwargs):
    """Document Upload field properties in the API spec."""
    ret = {}
    if isinstance(field, Upload):
        if self.openapi_version.major < 3:
            ret["type"] = "file"
        else:
            ret["type"] = "string"
            ret["format"] = field.format
    return ret
`

// TestLiteralEvidence_DisjointLiteralsSuppressType4 verifies that fragments
// whose string-literal vocabularies are completely disjoint score below the
// default Type-4 threshold even when their CFGs match exactly (issue #482).
func TestLiteralEvidence_DisjointLiteralsSuppressType4(t *testing.T) {
	f1 := fragmentFor(t, delimitedListConverterSrc, parser.NodeFunctionDef)
	f2 := fragmentFor(t, uploadFieldConverterSrc, parser.NodeFunctionDef)

	for name, analyzer := range map[string]*SemanticSimilarityAnalyzer{
		"cfg_only": NewSemanticSimilarityAnalyzer(),
		"with_dfa": NewSemanticSimilarityAnalyzerWithDFA(),
	} {
		t.Run(name, func(t *testing.T) {
			similarity := analyzer.ComputeSimilarity(f1, f2)
			require.Less(t, similarity, domain.DefaultType4CloneThreshold,
				"disjoint literal vocabularies must suppress Type-4 classification")
			require.Greater(t, similarity, 0.0,
				"penalty should reduce, not zero out, the similarity")
		})
	}
}

// TestLiteralEvidence_StandaloneIfBlocks covers the exact fragment shape
// reported in issue #482: the standalone `if isinstance(...)` blocks rather
// than the enclosing functions.
func TestLiteralEvidence_StandaloneIfBlocks(t *testing.T) {
	f1 := fragmentFor(t, delimitedListConverterSrc, parser.NodeIf)
	f2 := fragmentFor(t, uploadFieldConverterSrc, parser.NodeIf)

	analyzer := NewSemanticSimilarityAnalyzerWithDFA()
	similarity := analyzer.ComputeSimilarity(f1, f2)
	require.Less(t, similarity, domain.DefaultType4CloneThreshold,
		"isinstance blocks with disjoint spec keys must not reach the Type-4 threshold")
}

// TestLiteralEvidence_SharedLiteralsKeepScore verifies that a single shared
// literal disables the penalty: overlap means the fragments emit at least
// partially the same vocabulary.
func TestLiteralEvidence_SharedLiteralsKeepScore(t *testing.T) {
	src1 := `def converter_a(self, field, **kwargs):
    ret = {}
    if isinstance(field, Upload):
        if self.openapi_version.major < 3:
            ret["type"] = "file"
        else:
            ret["type"] = "string"
            ret["format"] = "binary"
    return ret
`
	// Same vocabulary, slightly different wiring.
	src2 := `def converter_b(self, field, **kwargs):
    ret = {}
    if isinstance(field, Upload):
        if self.openapi_version.major < 3:
            ret["format"] = "binary"
        else:
            ret["type"] = "string"
            ret["format"] = "binary"
    return ret
`
	f1 := fragmentFor(t, src1, parser.NodeFunctionDef)
	f2 := fragmentFor(t, src2, parser.NodeFunctionDef)

	analyzer := NewSemanticSimilarityAnalyzer()
	similarity := analyzer.ComputeSimilarity(f1, f2)
	require.GreaterOrEqual(t, similarity, domain.DefaultType4CloneThreshold,
		"overlapping literal vocabularies must not be penalized")
}

// TestLiteralEvidence_DocstringsAreNotEvidence verifies that differing
// docstrings alone never trigger the penalty: bare string statements are
// excluded from literal evidence.
func TestLiteralEvidence_DocstringsAreNotEvidence(t *testing.T) {
	src1 := `def sum_numbers(numbers):
    """Calculate sum using a for loop over the input."""
    "stray string statement"
    total = 0
    for num in numbers:
        total = total + num
    return total
`
	src2 := `def add_all(values):
    """Recursively accumulate every value."""
    "another stray string"
    acc = 0
    for v in values:
        acc = acc + v
    return acc
`
	f1 := fragmentFor(t, src1, parser.NodeFunctionDef)
	f2 := fragmentFor(t, src2, parser.NodeFunctionDef)

	signals1 := extractSemanticSignals(f1.ASTNode)
	signals2 := extractSemanticSignals(f2.ASTNode)
	require.Empty(t, signals1.stringLiterals, "docstrings and bare string statements must be ignored")
	require.Empty(t, signals2.stringLiterals, "docstrings and bare string statements must be ignored")

	analyzer := NewSemanticSimilarityAnalyzer()
	similarity := analyzer.ComputeSimilarity(f1, f2)
	require.GreaterOrEqual(t, similarity, domain.DefaultType4CloneThreshold,
		"equivalent loops must stay above the threshold despite differing docstrings")
}

// TestLiteralEvidence_InsufficientLiteralsNoPenalty verifies the evidence
// minimum: one literal per side is not enough to conclude anything.
func TestLiteralEvidence_InsufficientLiteralsNoPenalty(t *testing.T) {
	src1 := `def log_positive(items):
    count = 0
    for item in items:
        if item > 0:
            count = count + 1
    return "found: " + str(count)
`
	src2 := `def tally_above_zero(values):
    n = 0
    for v in values:
        if v > 0:
            n = n + 1
    return "total: " + str(n)
`
	f1 := fragmentFor(t, src1, parser.NodeFunctionDef)
	f2 := fragmentFor(t, src2, parser.NodeFunctionDef)

	signals1 := extractSemanticSignals(f1.ASTNode)
	signals2 := extractSemanticSignals(f2.ASTNode)
	require.Len(t, signals1.stringLiterals, 1)
	require.Len(t, signals2.stringLiterals, 1)

	analyzer := NewSemanticSimilarityAnalyzer()
	similarity := analyzer.ComputeSimilarity(f1, f2)
	require.GreaterOrEqual(t, similarity, domain.DefaultType4CloneThreshold,
		"a single differing literal per side must not suppress an otherwise equivalent pair")
}
