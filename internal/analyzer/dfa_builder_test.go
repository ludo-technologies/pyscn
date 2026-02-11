package analyzer

import (
	"context"
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

		if chainA != nil {
			assert.Len(t, chainA.Defs, 1)
			assert.Equal(t, DefKindParameter, chainA.Defs[0].Kind)
		}

		if chainB != nil {
			assert.Len(t, chainB.Defs, 1)
			assert.Equal(t, DefKindParameter, chainB.Defs[0].Kind)
		}
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
		if chainObj != nil {
			// obj.attr should create a UseKindAttribute
			hasAttrUse := false
			for _, use := range chainObj.Uses {
				if use.Kind == UseKindAttribute {
					hasAttrUse = true
					break
				}
			}
			assert.True(t, hasAttrUse, "Should have attribute access use")
		}
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
		if chainF != nil {
			// f(10) should create a UseKindCall
			hasCallUse := false
			for _, use := range chainF.Uses {
				if use.Kind == UseKindCall {
					hasCallUse = true
					break
				}
			}
			assert.True(t, hasCallUse, "Should have function call use")
		}
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

		simWithDFA := analyzerWithDFA.ComputeSimilarity(f1, f2, nil)
		simWithoutDFA := analyzerWithoutDFA.ComputeSimilarity(f1, f2, nil)

		// Both should produce valid similarity scores
		assert.Greater(t, simWithDFA, 0.0)
		assert.Greater(t, simWithoutDFA, 0.0)
		assert.LessOrEqual(t, simWithDFA, 1.0)
		assert.LessOrEqual(t, simWithoutDFA, 1.0)
	})
}
