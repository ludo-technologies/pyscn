package analyzer

import (
	"context"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/require"
)

// firstFunctionFragment parses src and returns a CodeFragment wrapping the first
// FunctionDef node. Used to exercise the semantic analyzer on small, hand-written
// snippets without going through the full clone detector pipeline.
func firstFunctionFragment(t *testing.T, src string) *CodeFragment {
	t.Helper()
	p := parser.New()
	result, err := p.Parse(context.Background(), []byte(src))
	require.NoError(t, err)
	require.NotNil(t, result.AST)

	var fn *parser.Node
	result.AST.Walk(func(n *parser.Node) bool {
		if fn != nil {
			return false
		}
		if n.Type == parser.NodeFunctionDef {
			fn = n
			return false
		}
		return true
	})
	require.NotNil(t, fn, "no FunctionDef found in source")

	return &CodeFragment{
		ASTNode: fn,
		Location: &CodeLocation{
			StartLine: 1,
			EndLine:   1,
		},
	}
}

// TestSemanticMinCFGBlocks_GateBlocksLinearFragments verifies that the default
// minimum-block gate suppresses Type-4 similarity for pairs of trivially small,
// fully linear functions. These produce saturated similarity (~1.0) under the
// raw CFG comparison and are the dominant FP shape observed in cross-repo audits.
func TestSemanticMinCFGBlocks_GateBlocksLinearFragments(t *testing.T) {
	f1 := firstFunctionFragment(t, `def foo(x):
    return x + 1
`)
	f2 := firstFunctionFragment(t, `def bar(y):
    return y * 2
`)

	analyzer := NewSemanticSimilarityAnalyzer()

	// With the default gate (DefaultSemanticMinCFGBlocks), these tiny linear
	// fragments are below the threshold and must score 0.0.
	require.Equal(t, 0.0, analyzer.ComputeSimilarity(f1, f2),
		"expected gated similarity to be 0.0 for sub-threshold CFGs")

	// Sanity-check the gate is the reason: disabling the gate should restore
	// the (saturated) similarity, demonstrating the gate is what cut it.
	analyzer.SetMinCFGBlocks(0)
	require.Greater(t, analyzer.ComputeSimilarity(f1, f2), 0.0,
		"expected non-zero similarity once gate is disabled")
}

// TestSemanticMinCFGBlocks_GatePassesBranchingFragments verifies that fragments
// with at least one branch (≥ DefaultSemanticMinCFGBlocks blocks) clear the gate.
func TestSemanticMinCFGBlocks_GatePassesBranchingFragments(t *testing.T) {
	f1 := firstFunctionFragment(t, `def foo(xs):
    total = 0
    for x in xs:
        if x > 0:
            total = total + x
    return total
`)
	f2 := firstFunctionFragment(t, `def bar(items):
    acc = 0
    for it in items:
        if it > 0:
            acc = acc + it
    return acc
`)

	analyzer := NewSemanticSimilarityAnalyzer()
	require.Greater(t, analyzer.ComputeSimilarity(f1, f2), 0.0,
		"fragments with branching/looping CFGs must clear the min-block gate")
}

// TestSemanticMinCFGBlocks_DefaultMatchesDomainConstant guards against the
// default drifting silently away from the domain-level constant.
func TestSemanticMinCFGBlocks_DefaultMatchesDomainConstant(t *testing.T) {
	a := NewSemanticSimilarityAnalyzer()
	require.Equal(t, domain.DefaultSemanticMinCFGBlocks, a.minCFGBlocks)

	b := NewSemanticSimilarityAnalyzerWithDFA()
	require.Equal(t, domain.DefaultSemanticMinCFGBlocks, b.minCFGBlocks)
}
