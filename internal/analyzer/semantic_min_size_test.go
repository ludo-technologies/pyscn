package analyzer

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/require"
)

func firstNodeOfType(t *testing.T, src string, nt parser.NodeType) *parser.Node {
	t.Helper()
	nodes := parseSourceForDFA(t, src).FindByType(nt)
	require.NotEmpty(t, nodes, "no node of type %v found in source", nt)
	return nodes[0]
}

func fragmentFor(t *testing.T, src string, nt parser.NodeType) *CodeFragment {
	t.Helper()
	return &CodeFragment{
		ASTNode:  firstNodeOfType(t, src, nt),
		Location: &CodeLocation{StartLine: 1, EndLine: 1},
	}
}

// TestSemanticMinCyclomatic_GateBlocksLinearFunctions verifies that the default
// gate suppresses Type-4 similarity for pairs of trivially small, fully linear
// functions (V(G)=1). These produce saturated similarity under the raw CFG
// comparison and are the dominant FP shape observed in cross-repo audits.
func TestSemanticMinCyclomatic_GateBlocksLinearFunctions(t *testing.T) {
	f1 := fragmentFor(t, "def foo(x):\n    return x + 1\n", parser.NodeFunctionDef)
	f2 := fragmentFor(t, "def bar(y):\n    return y * 2\n", parser.NodeFunctionDef)

	analyzer := NewSemanticSimilarityAnalyzer()

	require.Equal(t, 0.0, analyzer.ComputeSimilarity(f1, f2),
		"expected gated similarity to be 0.0 for V(G)=1 fragments")

	// Sanity-check the gate is the reason: disabling restores the saturated score.
	analyzer.SetMinCyclomaticComplexity(0)
	require.Greater(t, analyzer.ComputeSimilarity(f1, f2), 0.0,
		"expected non-zero similarity once gate is disabled")
}

// TestSemanticMinCyclomatic_GatePassesBranchingFunctions verifies that functions
// with at least one branch/loop (V(G)>=2) clear the gate.
func TestSemanticMinCyclomatic_GatePassesBranchingFunctions(t *testing.T) {
	f1 := fragmentFor(t,
		"def foo(xs):\n    total = 0\n    for x in xs:\n        if x > 0:\n            total = total + x\n    return total\n",
		parser.NodeFunctionDef)
	f2 := fragmentFor(t,
		"def bar(items):\n    acc = 0\n    for it in items:\n        if it > 0:\n            acc = acc + it\n    return acc\n",
		parser.NodeFunctionDef)

	analyzer := NewSemanticSimilarityAnalyzer()
	require.Greater(t, analyzer.ComputeSimilarity(f1, f2), 0.0,
		"fragments with branching/looping CFGs must clear the min-cyclomatic gate")
}

// TestSemanticMinCyclomatic_GatePassesStandaloneControlFlow verifies that
// standalone control-flow statement fragments (extracted by the clone detector
// from NodeIf / NodeFor / NodeWhile / NodeTry) are admitted by the gate even
// though their CFGs are slightly smaller than the equivalent function bodies.
// This guards against the BlockCount-based gate's regression of rejecting
// `if x: y` (which has V(G)=2 but fewer than 5 blocks).
func TestSemanticMinCyclomatic_GatePassesStandaloneControlFlow(t *testing.T) {
	cases := []struct {
		name string
		src  string
		nt   parser.NodeType
	}{
		{"if_no_else", "if ready:\n    value = 1\n", parser.NodeIf},
		{"for_stmt", "for x in xs:\n    total = total + x\n", parser.NodeFor},
		{"while_stmt", "i = 0\nwhile i < n:\n    i = i + 1\n", parser.NodeWhile},
		{"try_except", "try:\n    do_it()\nexcept Exception:\n    pass\n", parser.NodeTry},
	}

	analyzer := NewSemanticSimilarityAnalyzer()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f1 := fragmentFor(t, c.src, c.nt)
			f2 := fragmentFor(t, c.src, c.nt)
			require.Greater(t, analyzer.ComputeSimilarity(f1, f2), 0.0,
				"standalone control-flow fragment must clear the gate (V(G) should be >= 2)")
		})
	}
}

// TestSemanticMinCyclomatic_DefaultMatchesDomainConstant guards against the
// default drifting silently away from the domain-level constant.
func TestSemanticMinCyclomatic_DefaultMatchesDomainConstant(t *testing.T) {
	a := NewSemanticSimilarityAnalyzer()
	require.Equal(t, domain.DefaultSemanticMinCyclomaticComplexity, a.minCyclomatic)

	b := NewSemanticSimilarityAnalyzerWithDFA()
	require.Equal(t, domain.DefaultSemanticMinCyclomaticComplexity, b.minCyclomatic)
}
