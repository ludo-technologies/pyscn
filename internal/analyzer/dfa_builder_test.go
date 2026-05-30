package analyzer

import (
	"context"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseSourceForDFA parses Python source code and returns the AST
func parseSourceForDFA(t *testing.T, source string) *parser.Node {
	p := parser.New()
	ctx := context.Background()

	result, err := p.Parse(ctx, []byte(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	return result.AST
}

// buildCFGForDFA builds a CFG from source code
func buildCFGForDFA(t *testing.T, source string) *CFG {
	ast := parseSourceForDFA(t, source)
	builder := NewCFGBuilder()
	cfg, err := builder.Build(ast)
	if err != nil {
		t.Fatalf("Failed to build CFG: %v", err)
	}
	return cfg
}

func assertChainHasUseKind(t *testing.T, chain *DefUseChain, kind DefUseKind) {
	t.Helper()

	for _, use := range chain.Uses {
		if use.Kind == kind {
			return
		}
	}
	t.Fatalf("Expected %s to have use kind %s, got %v", chain.Variable, kind, chain.Uses)
}

func requireDFAChain(t *testing.T, info *DFAInfo, variable string) *DefUseChain {
	t.Helper()

	chain := info.Chains[variable]
	require.NotNil(t, chain, "%q should be in variable chains", variable)
	return chain
}

func assertUsesOnlyInBlockLabel(t *testing.T, chain *DefUseChain, wantUses int, label string) {
	t.Helper()

	require.Len(t, chain.Uses, wantUses, "unexpected uses for %q", chain.Variable)
	for _, use := range chain.Uses {
		require.NotNil(t, use.Block, "use for %q has nil block", chain.Variable)
		assert.Truef(t, strings.HasPrefix(use.Block.Label, label),
			"use for %q is in block %q, want prefix %q", chain.Variable, use.Block.Label, label)
	}
}

func TestDFABuilder(t *testing.T) {
	t.Run("Build_NilCFG", func(t *testing.T) {
		builder := NewDFABuilder()
		info, err := builder.Build(nil)

		assert.NoError(t, err)
		assert.Nil(t, info)
	})

	t.Run("Build_SimpleAssignment", func(t *testing.T) {
		source := `
x = 10
y = x
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)
		require.NotNil(t, info)

		// Should have 2 definitions (x and y)
		assert.Equal(t, 2, info.TotalDefs())

		// Should have 1 use (x in y = x)
		assert.Equal(t, 1, info.TotalUses())

		// Check x chain
		chainX := info.Chains["x"]
		require.NotNil(t, chainX)
		assert.Len(t, chainX.Defs, 1)
		assert.Equal(t, DefKindAssign, chainX.Defs[0].Kind)

		// Check y chain
		chainY := info.Chains["y"]
		require.NotNil(t, chainY)
		assert.Len(t, chainY.Defs, 1)
	})

	t.Run("Build_DefUsePairInSameBlock", func(t *testing.T) {
		source := `
x = 10
y = x + 1
z = y * 2
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		// x defined at pos 0, used at pos 1 -> 1 pair
		chainX := info.Chains["x"]
		require.NotNil(t, chainX)
		assert.Len(t, chainX.Pairs, 1)
		assert.False(t, chainX.Pairs[0].IsCrossBlock())

		// y defined at pos 1, used at pos 2 -> 1 pair
		chainY := info.Chains["y"]
		require.NotNil(t, chainY)
		assert.Len(t, chainY.Pairs, 1)
	})

	t.Run("Build_MultipleAssignment", func(t *testing.T) {
		source := `
a, b = 1, 2
c = a + b
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		// Should have 3 definitions (a, b, c)
		assert.Equal(t, 3, info.TotalDefs())

		// a and b should be in the chains
		assert.Contains(t, info.Chains, "a")
		assert.Contains(t, info.Chains, "b")
		assert.Contains(t, info.Chains, "c")
	})

	t.Run("Build_AugmentedAssignment", func(t *testing.T) {
		source := `
x = 10
x += 5
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		chainX := info.Chains["x"]
		require.NotNil(t, chainX)

		// Should have 2 defs (original + augmented)
		assert.Len(t, chainX.Defs, 2)

		// Find the augmented def
		hasAugmented := false
		for _, def := range chainX.Defs {
			if def.Kind == DefKindAugmented {
				hasAugmented = true
				break
			}
		}
		assert.True(t, hasAugmented, "Should have augmented assignment def")

		// x += 5 also uses x, so there should be uses
		assert.GreaterOrEqual(t, len(chainX.Uses), 1)
	})

	t.Run("Build_ForLoop", func(t *testing.T) {
		source := `
for i in range(10):
    print(i)
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		chainI := info.Chains["i"]
		require.NotNil(t, chainI)

		// i should be defined as for target
		assert.Len(t, chainI.Defs, 1)
		assert.Equal(t, DefKindForTarget, chainI.Defs[0].Kind)

		// i should be used in print(i)
		assert.GreaterOrEqual(t, len(chainI.Uses), 1)
	})

	t.Run("Build_FunctionParameters", func(t *testing.T) {
		source := `
def add(a, b):
    return a + b
`
		ast := parseSourceForDFA(t, source)
		funcNode := ast.Body[0]

		cfgBuilder := NewCFGBuilder()
		cfg, err := cfgBuilder.Build(funcNode)
		require.NoError(t, err)

		dfaBuilder := NewDFABuilder()
		info, err := dfaBuilder.Build(cfg)
		require.NoError(t, err)

		// a and b should be defined as parameters
		chainA := info.Chains["a"]
		chainB := info.Chains["b"]

		require.NotNil(t, chainA)
		assert.Len(t, chainA.Defs, 1)
		assert.Equal(t, DefKindParameter, chainA.Defs[0].Kind)

		require.NotNil(t, chainB)
		assert.Len(t, chainB.Defs, 1)
		assert.Equal(t, DefKindParameter, chainB.Defs[0].Kind)
	})

	t.Run("Build_Import", func(t *testing.T) {
		source := `
import os
x = os.path
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		chainOS := info.Chains["os"]
		if chainOS != nil {
			assert.Len(t, chainOS.Defs, 1)
			assert.Equal(t, DefKindImport, chainOS.Defs[0].Kind)
		}
	})

	t.Run("Build_AttributeAccess", func(t *testing.T) {
		source := `
obj = SomeClass()
value = obj.attr
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		chainObj := info.Chains["obj"]
		require.NotNil(t, chainObj)
		assertChainHasUseKind(t, chainObj, UseKindAttribute)
	})

	t.Run("Build_FunctionCall", func(t *testing.T) {
		source := `
def f(x):
    return x

result = f(10)
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		chainF := info.Chains["f"]
		require.NotNil(t, chainF)
		assertChainHasUseKind(t, chainF, UseKindCall)
	})

	t.Run("Build_SubscriptExpression", func(t *testing.T) {
		source := `
items = [1, 2, 3]
idx = 1
value = items[idx]
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		chainItems := info.Chains["items"]
		require.NotNil(t, chainItems)
		assertChainHasUseKind(t, chainItems, UseKindSubscript)

		chainIdx := info.Chains["idx"]
		require.NotNil(t, chainIdx)
		assertChainHasUseKind(t, chainIdx, UseKindRead)
	})

	t.Run("Build_ControlFlowCrossBlock", func(t *testing.T) {
		source := `
x = 10
if x > 5:
    y = x
else:
    y = 0
z = y
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		// x is defined in one block and potentially used in another
		chainX := info.Chains["x"]
		require.NotNil(t, chainX)
		assert.GreaterOrEqual(t, len(chainX.Uses), 1)

		// y is defined in conditional blocks and used after
		chainY := info.Chains["y"]
		if chainY != nil {
			assert.GreaterOrEqual(t, len(chainY.Defs), 1)
		}
	})

	t.Run("Build_WithStatement_ExtractsAliasTarget", func(t *testing.T) {
		source := `
with open(path) as f:
    data = f.read()
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		// 'f' should be extracted as a DefKindWithTarget definition
		chainF := info.Chains["f"]
		require.NotNil(t, chainF, "'f' should be in variable chains")
		assert.Len(t, chainF.Defs, 1)
		assert.Equal(t, DefKindWithTarget, chainF.Defs[0].Kind)

		// Note: attribute access like f.read() may not link the base identifier
		// due to how extractUsesFromExpression handles identifiers in attribute chains.
		// Follow-up issue to track: link attribute access bases to their definitions.
	})

	t.Run("Build_WithStatement_MultipleItems", func(t *testing.T) {
		source := `
with open(src) as inp, open(dst) as out:
    out.write(inp.read())
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		// Both 'inp' and 'out' should be extracted as DefKindWithTarget
		chainInp := info.Chains["inp"]
		require.NotNil(t, chainInp, "'inp' should be in variable chains")
		assert.Len(t, chainInp.Defs, 1)
		assert.Equal(t, DefKindWithTarget, chainInp.Defs[0].Kind)

		chainOut := info.Chains["out"]
		require.NotNil(t, chainOut, "'out' should be in variable chains")
		assert.Len(t, chainOut.Defs, 1)
		assert.Equal(t, DefKindWithTarget, chainOut.Defs[0].Kind)
	})

	t.Run("Build_AsyncWithStatement_ExtractsAliasTarget", func(t *testing.T) {
		source := `
async def fetch(session, url):
    async with session.get(url) as response:
        return response
`
		ast := parseSourceForDFA(t, source)
		cfgs, err := NewCFGBuilder().BuildAll(ast)
		require.NoError(t, err)

		cfg, ok := cfgs["fetch"]
		require.True(t, ok, "expected CFG for async function 'fetch'")

		info, err := NewDFABuilder().Build(cfg)
		require.NoError(t, err)

		chainResponse := info.Chains["response"]
		require.NotNil(t, chainResponse, "'response' should be in variable chains")
		assert.Len(t, chainResponse.Defs, 1)
		assert.Equal(t, DefKindWithTarget, chainResponse.Defs[0].Kind)
	})

	t.Run("Build_WithStatement_TupleAliasUnpacks", func(t *testing.T) {
		source := `
with cm() as (a, b):
    use(a, b)
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		for _, name := range []string{"a", "b"} {
			chain := info.Chains[name]
			require.NotNil(t, chain, "%q should be in variable chains", name)
			assert.Len(t, chain.Defs, 1)
			assert.Equal(t, DefKindWithTarget, chain.Defs[0].Kind)
		}
	})

	t.Run("Build_WithStatement_ListAliasWithStarredUnpacks", func(t *testing.T) {
		source := `
with cm() as [a, *rest]:
    use(a, rest)
`
		cfg := buildCFGForDFA(t, source)
		builder := NewDFABuilder()
		info, err := builder.Build(cfg)

		require.NoError(t, err)

		for _, name := range []string{"a", "rest"} {
			chain := info.Chains[name]
			require.NotNil(t, chain, "%q should be in variable chains", name)
			assert.Len(t, chain.Defs, 1)
			assert.Equal(t, DefKindWithTarget, chain.Defs[0].Kind)
		}
	})

	t.Run("Build_IfStatement_UsesBodyOnlyInBodyBlock", func(t *testing.T) {
		source := `
cond = True
body_value = 1
if cond:
    sink(body_value)
`
		info, err := NewDFABuilder().Build(buildCFGForDFA(t, source))
		require.NoError(t, err)

		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "cond"), 1, LabelEntry)
		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "body_value"), 1, LabelIfThen)
	})

	t.Run("Build_ForStatement_UsesIterableWithoutReadingTargetOrBodyInHeader", func(t *testing.T) {
		source := `
items = [1, 2, 3]
body_value = 1
for item in items:
    sink(body_value)
`
		info, err := NewDFABuilder().Build(buildCFGForDFA(t, source))
		require.NoError(t, err)

		item := requireDFAChain(t, info, "item")
		require.Len(t, item.Defs, 1)
		assert.Equal(t, DefKindForTarget, item.Defs[0].Kind)
		assert.Empty(t, item.Uses, "loop targets should not be phantom reads")

		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "items"), 1, LabelLoopHeader)
		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "body_value"), 1, LabelLoopBody)
	})

	t.Run("Build_WhileStatement_UsesConditionWithoutReadingBodyInHeader", func(t *testing.T) {
		source := `
cond = True
body_value = 1
while cond:
    sink(body_value)
    cond = False
`
		info, err := NewDFABuilder().Build(buildCFGForDFA(t, source))
		require.NoError(t, err)

		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "body_value"), 1, LabelLoopBody)
		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "cond"), 1, LabelLoopHeader)
	})

	t.Run("Build_WithStatement_UsesContextExprWithoutReadingBodyInSetup", func(t *testing.T) {
		source := `
path = "data.txt"
body_value = 1
with open(path) as handle:
    sink(body_value)
`
		info, err := NewDFABuilder().Build(buildCFGForDFA(t, source))
		require.NoError(t, err)

		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "path"), 1, LabelWithSetup)
		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "body_value"), 1, LabelWithBody)

		handle := requireDFAChain(t, info, "handle")
		require.Len(t, handle.Defs, 1)
		assert.Equal(t, DefKindWithTarget, handle.Defs[0].Kind)
		assert.Empty(t, handle.Uses, "with targets should not be phantom reads")
	})

	t.Run("Build_MatchStatement_UsesSubjectWithoutDoubleCountingCaseBody", func(t *testing.T) {
		source := `
subject = 1
body_value = 1
match subject:
    case _:
        sink(body_value)
`
		info, err := NewDFABuilder().Build(buildCFGForDFA(t, source))
		require.NoError(t, err)

		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "subject"), 1, LabelMatchEval)
		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "body_value"), 1, LabelMatchCase)
	})

	t.Run("Build_ExceptHandler_DoesNotDoubleCountHandlerBody", func(t *testing.T) {
		source := `
body_value = 1
try:
    raise ValueError()
except ValueError as exc:
    sink(body_value)
`
		info, err := NewDFABuilder().Build(buildCFGForDFA(t, source))
		require.NoError(t, err)

		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "body_value"), 1, LabelExceptBlock)
	})

	t.Run("Build_FunctionDef_DoesNotAttributeBodyUsesToOuterBlock", func(t *testing.T) {
		source := `
outer_value = 1
def nested():
    sink(outer_value)
`
		info, err := NewDFABuilder().Build(buildCFGForDFA(t, source))
		require.NoError(t, err)

		chain := requireDFAChain(t, info, "outer_value")
		require.Len(t, chain.Defs, 1)
		assert.Empty(t, chain.Uses, "function body uses belong to the nested function CFG")
	})

	t.Run("Build_FunctionDef_UsesDefaultArgumentsInOuterBlock", func(t *testing.T) {
		source := `
default_value = 1
def nested(arg=default_value):
    return arg
`
		info, err := NewDFABuilder().Build(buildCFGForDFA(t, source))
		require.NoError(t, err)

		chain := requireDFAChain(t, info, "default_value")
		require.Len(t, chain.Defs, 1)
		assertUsesOnlyInBlockLabel(t, chain, 1, LabelEntry)
	})

	t.Run("Build_NestedFunctionCFG_RetainsBodyUses", func(t *testing.T) {
		source := `
outer_value = 1
def nested():
    sink(outer_value)
`
		ast := parseSourceForDFA(t, source)
		cfgs, err := NewCFGBuilder().BuildAll(ast)
		require.NoError(t, err)

		nestedCFG, ok := cfgs["nested"]
		require.True(t, ok, "expected nested function CFG")

		info, err := NewDFABuilder().Build(nestedCFG)
		require.NoError(t, err)

		assertUsesOnlyInBlockLabel(t, requireDFAChain(t, info, "outer_value"), 1, LabelFunctionBody)
	})
}

func TestDFAFeatureComparison(t *testing.T) {
	t.Run("CompareDFAFeatures_Identical", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzerWithDFA()

		f1 := &DFAFeatures{
			TotalDefs:       5,
			TotalUses:       10,
			TotalPairs:      8,
			UniqueVariables: 3,
			AvgChainLength:  2.0,
			CrossBlockPairs: 2,
			IntraBlockPairs: 6,
			DefKindCounts:   map[DefUseKind]int{DefKindAssign: 4, DefKindParameter: 1},
			UseKindCounts:   map[DefUseKind]int{UseKindRead: 8, UseKindCall: 2},
		}
		f2 := &DFAFeatures{
			TotalDefs:       5,
			TotalUses:       10,
			TotalPairs:      8,
			UniqueVariables: 3,
			AvgChainLength:  2.0,
			CrossBlockPairs: 2,
			IntraBlockPairs: 6,
			DefKindCounts:   map[DefUseKind]int{DefKindAssign: 4, DefKindParameter: 1},
			UseKindCounts:   map[DefUseKind]int{UseKindRead: 8, UseKindCall: 2},
		}

		similarity := analyzer.compareDFAFeatures(f1, f2)
		assert.InDelta(t, 1.0, similarity, 0.001, "Identical features should have similarity 1.0")
	})

	t.Run("CompareDFAFeatures_Different", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzerWithDFA()

		f1 := &DFAFeatures{
			TotalDefs:       5,
			TotalUses:       10,
			TotalPairs:      8,
			UniqueVariables: 3,
			AvgChainLength:  2.0,
			CrossBlockPairs: 2,
			IntraBlockPairs: 6,
			DefKindCounts:   map[DefUseKind]int{DefKindAssign: 4, DefKindParameter: 1},
			UseKindCounts:   map[DefUseKind]int{UseKindRead: 8, UseKindCall: 2},
		}
		f2 := &DFAFeatures{
			TotalDefs:       10,
			TotalUses:       20,
			TotalPairs:      15,
			UniqueVariables: 6,
			AvgChainLength:  1.5,
			CrossBlockPairs: 5,
			IntraBlockPairs: 10,
			DefKindCounts:   map[DefUseKind]int{DefKindAssign: 8, DefKindForTarget: 2},
			UseKindCounts:   map[DefUseKind]int{UseKindRead: 15, UseKindAttribute: 5},
		}

		similarity := analyzer.compareDFAFeatures(f1, f2)
		assert.Greater(t, similarity, 0.0, "Different features should have some similarity")
		assert.Less(t, similarity, 1.0, "Different features should not be identical")
	})

	t.Run("CompareDFAFeatures_NilFeatures", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzerWithDFA()

		f1 := &DFAFeatures{TotalDefs: 5}
		similarity := analyzer.compareDFAFeatures(nil, f1)
		assert.Equal(t, 0.0, similarity)

		similarity = analyzer.compareDFAFeatures(f1, nil)
		assert.Equal(t, 0.0, similarity)

		similarity = analyzer.compareDFAFeatures(nil, nil)
		assert.Equal(t, 0.0, similarity)
	})

	t.Run("CompareDFAFeatures_EmptyFeatures", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzerWithDFA()

		f1 := &DFAFeatures{TotalDefs: 0, DefKindCounts: map[DefUseKind]int{}, UseKindCounts: map[DefUseKind]int{}}
		f2 := &DFAFeatures{TotalDefs: 0, DefKindCounts: map[DefUseKind]int{}, UseKindCounts: map[DefUseKind]int{}}

		similarity := analyzer.compareDFAFeatures(f1, f2)
		assert.Equal(t, 1.0, similarity, "Both empty should have similarity 1.0")
	})
}

func TestSemanticSimilarityWithDFA(t *testing.T) {
	t.Run("DFAEnabled", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzerWithDFA()
		assert.True(t, analyzer.IsDFAEnabled())
	})

	t.Run("DFADisabled", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzer()
		assert.False(t, analyzer.IsDFAEnabled())
	})

	t.Run("SetEnableDFA", func(t *testing.T) {
		analyzer := NewSemanticSimilarityAnalyzer()
		assert.False(t, analyzer.IsDFAEnabled())

		analyzer.SetEnableDFA(true)
		assert.True(t, analyzer.IsDFAEnabled())

		analyzer.SetEnableDFA(false)
		assert.False(t, analyzer.IsDFAEnabled())
	})

	t.Run("ComputeSimilarityWithDFA", func(t *testing.T) {
		// Two similar functions with same data flow pattern
		source1 := `
def calc(a, b):
    result = a + b
    return result
`
		source2 := `
def compute(x, y):
    sum = x + y
    return sum
`
		ast1 := parseSourceForDFA(t, source1)
		ast2 := parseSourceForDFA(t, source2)

		f1 := &CodeFragment{ASTNode: ast1.Body[0]}
		f2 := &CodeFragment{ASTNode: ast2.Body[0]}

		analyzerWithDFA := NewSemanticSimilarityAnalyzerWithDFA()
		analyzerWithoutDFA := NewSemanticSimilarityAnalyzer()

		// This test exercises the DFA/non-DFA code paths on small linear fragments;
		// disable the V(G) gate so both paths can produce a score (the gate itself
		// is exercised in semantic_min_size_test.go).
		analyzerWithDFA.SetMinCyclomaticComplexity(0)
		analyzerWithoutDFA.SetMinCyclomaticComplexity(0)

		simWithDFA := analyzerWithDFA.ComputeSimilarity(f1, f2)
		simWithoutDFA := analyzerWithoutDFA.ComputeSimilarity(f1, f2)

		// Both should produce valid similarity scores
		assert.Greater(t, simWithDFA, 0.0)
		assert.Greater(t, simWithoutDFA, 0.0)
		assert.LessOrEqual(t, simWithDFA, 1.0)
		assert.LessOrEqual(t, simWithoutDFA, 1.0)
	})
}
