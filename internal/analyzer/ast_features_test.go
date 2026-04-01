package analyzer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewASTFeatureExtractor(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	assert.NotNil(t, extractor)
	// Default values
	assert.Equal(t, 3, extractor.maxSubtreeHeight)
	assert.Equal(t, 4, extractor.kGramSize)
	assert.True(t, extractor.includeTypes)
	assert.False(t, extractor.includeLiterals)
}

func TestWithOptions(t *testing.T) {
	extractor := NewASTFeatureExtractor().WithOptions(5, 3, false, true)

	assert.Equal(t, 5, extractor.maxSubtreeHeight)
	assert.Equal(t, 3, extractor.kGramSize)
	assert.False(t, extractor.includeTypes)
	assert.True(t, extractor.includeLiterals)
}

func TestExtractFeatures_NilTree(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	features, err := extractor.ExtractFeatures(nil)

	assert.NoError(t, err)
	assert.Empty(t, features)
}

func TestExtractFeatures_SingleNode(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	node := NewTreeNode(1, "FunctionDef")

	features, err := extractor.ExtractFeatures(node)

	assert.NoError(t, err)
	assert.NotEmpty(t, features)
	// Should contain type feature
	assert.Contains(t, features, "type:FunctionDef")
}

func TestExtractFeatures_Deterministic(t *testing.T) {
	// Same tree should produce same features
	extractor := NewASTFeatureExtractor()

	tree1 := NewTreeNode(0, "FunctionDef")
	tree2 := NewTreeNode(0, "FunctionDef")

	f1, _ := extractor.ExtractFeatures(tree1)
	f2, _ := extractor.ExtractFeatures(tree2)

	assert.Equal(t, f1, f2)
}

func TestExtractFeatures_Patterns(t *testing.T) {
	extractor := NewASTFeatureExtractor()

	// Create a tree with various structural elements
	// FunctionDef
	//   Arguments
	//   If
	//     Compare
	//     Return
	root := NewTreeNode(1, "FunctionDef")
	args := NewTreeNode(2, "Arguments")
	ifNode := NewTreeNode(3, "If")
	compare := NewTreeNode(4, "Compare")
	ret := NewTreeNode(5, "Return")

	root.AddChild(args)
	root.AddChild(ifNode)
	ifNode.AddChild(compare)
	ifNode.AddChild(ret)

	features, err := extractor.ExtractFeatures(root)
	assert.NoError(t, err)

	// Check for pattern tokens
	expectedPatterns := []string{
		"pattern:FunctionDef",
		"pattern:If",
		"pattern:Compare",
		"pattern:Return",
	}

	for _, p := range expectedPatterns {
		assert.Contains(t, features, p)
	}
}

func TestExtractFeatures_Config(t *testing.T) {
	t.Run("ExcludeTypes", func(t *testing.T) {
		extractor := NewASTFeatureExtractor().WithOptions(3, 4, false, false)
		node := NewTreeNode(1, "FunctionDef")
		features, err := extractor.ExtractFeatures(node)
		assert.NoError(t, err)

		// Should NOT contain type feature
		for _, f := range features {
			assert.False(t, strings.HasPrefix(f, "type:"), "Should not contain type features")
		}
	})

	t.Run("IncludeLiterals", func(t *testing.T) {
		extractor := NewASTFeatureExtractor().WithOptions(3, 4, true, true)
		node := NewTreeNode(1, "Constant(42)")
		features, err := extractor.ExtractFeatures(node)
		assert.NoError(t, err)

		// Should contain literal in type
		assert.Contains(t, features, "type:Constant(42)")
	})
}

func TestExtractFeatures_Comprehensive(t *testing.T) {
	tests := []struct {
		name            string
		tree            *TreeNode
		config          func(*ASTFeatureExtractor)
		expectedFeats   []string
		unexpectedFeats []string
	}{
		{
			name:          "Empty Tree",
			tree:          nil,
			expectedFeats: []string{},
		},
		{
			name: "Simple Function",
			tree: func() *TreeNode {
				n := NewTreeNode(1, "FunctionDef")
				n.AddChild(NewTreeNode(2, "Pass"))
				return n
			}(),
			expectedFeats: []string{
				"type:FunctionDef",
				"type:Pass",
				"pattern:FunctionDef",
				"typedist:FunctionDef:1",
				"typedist:Pass:1",
			},
		},
		{
			name: "KGrams",
			tree: func() *TreeNode {
				// A -> B -> C -> D
				a := NewTreeNode(1, "A")
				b := NewTreeNode(2, "B")
				c := NewTreeNode(3, "C")
				d := NewTreeNode(4, "D")
				a.AddChild(b)
				b.AddChild(c)
				c.AddChild(d)
				return a
			}(),
			config: func(e *ASTFeatureExtractor) {
				e.WithOptions(3, 3, true, false) // k=3
			},
			expectedFeats: []string{
				"kgram:A:B:C",
				"kgram:B:C:D",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewASTFeatureExtractor()
			if tt.config != nil {
				tt.config(extractor)
			}

			features, err := extractor.ExtractFeatures(tt.tree)
			assert.NoError(t, err)

			if tt.tree == nil {
				assert.Empty(t, features)
				return
			}

			for _, exp := range tt.expectedFeats {
				assert.Contains(t, features, exp)
			}

			for _, unexp := range tt.unexpectedFeats {
				assert.NotContains(t, features, unexp)
			}
		})
	}
}

func TestExtractSubtreeHashes_NilTree(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	hashes := extractor.ExtractSubtreeHashes(nil, 3)
	assert.Empty(t, hashes)
}

func TestExtractSubtreeHashes_SingleNode(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	node := NewTreeNode(1, "Test")

	hashes := extractor.ExtractSubtreeHashes(node, 3)

	assert.Len(t, hashes, 1)
	assert.True(t, strings.HasPrefix(hashes[0], "sub:0:"))
}

func TestExtractSubtreeHashes_MaxHeight(t *testing.T) {
	extractor := NewASTFeatureExtractor()

	// Build tree of height 4
	root := NewTreeNode(1, "root")
	child := NewTreeNode(2, "child")
	grandchild := NewTreeNode(3, "grandchild")
	greatgrand := NewTreeNode(4, "greatgrand")

	root.AddChild(child)
	child.AddChild(grandchild)
	grandchild.AddChild(greatgrand)

	// With maxHeight=2, should not include root's hash
	hashes := extractor.ExtractSubtreeHashes(root, 2)

	// Count hashes by height
	heightCounts := make(map[string]int)
	for _, h := range hashes {
		parts := strings.Split(h, ":")
		heightCounts[parts[1]]++
	}

	assert.Greater(t, heightCounts["0"], 0) // Leaves
	assert.Greater(t, heightCounts["1"], 0) // Height 1
	assert.Greater(t, heightCounts["2"], 0) // Height 2
	_, has3 := heightCounts["3"]
	assert.False(t, has3) // No height 3+ hashes
}

func TestExtractSubtreeHashes_OrderSensitivity(t *testing.T) {
	extractor := NewASTFeatureExtractor()

	// Tree 1: Root -> [A, B]
	root1 := NewTreeNode(1, "Root")
	root1.AddChild(NewTreeNode(2, "A"))
	root1.AddChild(NewTreeNode(3, "B"))

	// Tree 2: Root -> [B, A]
	root2 := NewTreeNode(4, "Root")
	root2.AddChild(NewTreeNode(5, "B"))
	root2.AddChild(NewTreeNode(6, "A"))

	hashes1 := extractor.ExtractSubtreeHashes(root1, 2)
	hashes2 := extractor.ExtractSubtreeHashes(root2, 2)

	// Find root hashes (height 1)
	var h1, h2 string
	for _, h := range hashes1 {
		if strings.HasPrefix(h, "sub:1:") {
			h1 = h
			break
		}
	}
	for _, h := range hashes2 {
		if strings.HasPrefix(h, "sub:1:") {
			h2 = h
			break
		}
	}

	assert.NotEmpty(t, h1)
	assert.NotEmpty(t, h2)
	assert.NotEqual(t, h1, h2, "Hashes should be different for different child order")
}

func TestExtractSubtreeHashes_Literals(t *testing.T) {
	// Default: includeLiterals = false
	extractor := NewASTFeatureExtractor()

	n1 := NewTreeNode(1, "Constant(1)")
	n2 := NewTreeNode(2, "Constant(2)")

	h1 := extractor.ExtractSubtreeHashes(n1, 1)
	h2 := extractor.ExtractSubtreeHashes(n2, 1)

	assert.Equal(t, h1, h2, "Hashes should be identical when literals are ignored")

	// includeLiterals = true
	extractor = NewASTFeatureExtractor().WithOptions(3, 4, true, true)
	h3 := extractor.ExtractSubtreeHashes(n1, 1)
	h4 := extractor.ExtractSubtreeHashes(n2, 1)

	assert.NotEqual(t, h3, h4, "Hashes should be different when literals are included")
}

func TestExtractNodeSequences_NilTree(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	seqs := extractor.ExtractNodeSequences(nil, 4)
	assert.Empty(t, seqs)
}

func TestExtractNodeSequences_KTooLarge(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	node := NewTreeNode(1, "single")

	// Tree has only 1 node, k=4 should return empty
	seqs := extractor.ExtractNodeSequences(node, 4)
	assert.Empty(t, seqs)
}

func TestExtractNodeSequences_ValidKGrams(t *testing.T) {
	extractor := NewASTFeatureExtractor()

	// Create tree: A -> B -> C
	a := NewTreeNode(1, "A")
	b := NewTreeNode(2, "B")
	c := NewTreeNode(3, "C")
	a.AddChild(b)
	b.AddChild(c)

	// Pre-order: A, B, C
	// With k=2: ["A:B", "B:C"]
	seqs := extractor.ExtractNodeSequences(a, 2)

	assert.Len(t, seqs, 2)
	assert.Contains(t, seqs, "A:B")
	assert.Contains(t, seqs, "B:C")
}

func TestExtractNodeSequences_KEqualsOne(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	node := NewTreeNode(1, "test")

	// k=1 should return empty (as per implementation)
	seqs := extractor.ExtractNodeSequences(node, 1)
	assert.Empty(t, seqs)
}

func TestCanonicalLabel_WithPayload(t *testing.T) {
	extractor := NewASTFeatureExtractor() // includeLiterals=false

	// Should strip payload
	assert.Equal(t, "FunctionDef", extractor.canonicalLabel("FunctionDef(myFunc)"))
	assert.Equal(t, "Constant", extractor.canonicalLabel("Constant(42)"))
}

func TestCanonicalLabel_WithLiterals(t *testing.T) {
	extractor := NewASTFeatureExtractor().WithOptions(0, 0, true, true)

	// Should keep payload
	assert.Equal(t, "FunctionDef(myFunc)", extractor.canonicalLabel("FunctionDef(myFunc)"))
}

func TestPreorderLabels_NilTree(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	labels := extractor.preorderLabels(nil)
	assert.Equal(t, labels, []string{})
}

func TestPreorderLabels_Traversal(t *testing.T) {
	extractor := NewASTFeatureExtractor()

	// Tree:
	//      A
	//     / \
	//    B   C
	//   /
	//  D
	root := NewTreeNode(1, "A")
	b := NewTreeNode(2, "B")
	c := NewTreeNode(3, "C")
	d := NewTreeNode(4, "D")

	root.AddChild(b)
	root.AddChild(c)
	b.AddChild(d)

	// Preorder: A, B, D, C
	labels := extractor.preorderLabels(root)
	expected := []string{"A", "B", "D", "C"}

	assert.Equal(t, expected, labels)
}

func TestPreorderLabels_WithLiterals(t *testing.T) {
	extractor := NewASTFeatureExtractor().WithOptions(3, 4, true, true)

	node := NewTreeNode(1, "Constant(42)")
	labels := extractor.preorderLabels(node)

	assert.Equal(t, []string{"Constant(42)"}, labels)
}

func TestPreorderLabels_WithoutLiterals(t *testing.T) {

	node := NewTreeNode(1, "Constant(42)")

	extractor := NewASTFeatureExtractor()

	labels := extractor.preorderLabels(node)
	assert.Equal(t, []string{"Constant"}, labels)
}

func TestBinCount(t *testing.T) {
	extractor := NewASTFeatureExtractor()

	tests := []struct {
		count    int
		expected string
	}{
		{0, "1"},
		{1, "1"},
		{2, "2-3"},
		{3, "2-3"},
		{4, "4-7"},
		{7, "4-7"},
		{8, "8-15"},
		{15, "8-15"},
		{16, "16+"},
		{100, "16+"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("count_%d", tt.count), func(t *testing.T) {
			result := extractor.binCount(tt.count)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractPatterns_NilTree(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	patterns := extractor.extractPatterns(nil)

	assert.Equal(t, patterns, []string{})
}

func TestExtractPatterns_ComplexTree(t *testing.T) {
	extractor := NewASTFeatureExtractor()

	// Construct a tree with specific patterns:
	// FunctionDef
	//   For
	//     If
	//       Return
	//   Assign
	//     Call
	root := NewTreeNode(1, "FunctionDef")
	forNode := NewTreeNode(2, "For")
	ifNode := NewTreeNode(3, "If")
	returnNode := NewTreeNode(4, "Return")
	assignNode := NewTreeNode(5, "Assign")
	callNode := NewTreeNode(6, "Call")

	root.AddChild(forNode)
	root.AddChild(assignNode)
	forNode.AddChild(ifNode)
	ifNode.AddChild(returnNode)
	assignNode.AddChild(callNode)

	patterns := extractor.extractPatterns(root)

	// Expected patterns (sorted alphabetically as per implementation)
	expected := []string{
		"Assign",
		"Call",
		"For",
		"FunctionDef",
		"If",
		"Return",
	}

	assert.Equal(t, expected, patterns)
}

func TestExtractPatterns_NoPatterns(t *testing.T) {
	extractor := NewASTFeatureExtractor()

	// Tree with nodes that are not in the pattern list (e.g. Module, Expr, Name)
	root := NewTreeNode(1, "Module")
	child := NewTreeNode(2, "Expr")
	child2 := NewTreeNode(3, "Name")
	root.AddChild(child)
	child.AddChild(child2)

	patterns := extractor.extractPatterns(root)
	assert.Empty(t, patterns)
}

func TestExtractPatterns_MixPatterns(t *testing.T) {
	extractor := NewASTFeatureExtractor()

	// Construct a tree with specific patterns:
	// FunctionDef
	//   For
	//     Module
	//     If
	//       Expr
	//       Return
	//   Name
	//   Assign
	//     Call

	//Real patterns declaration
	root := NewTreeNode(1, "FunctionDef")
	forNode := NewTreeNode(2, "For")
	ifNode := NewTreeNode(3, "If")
	returnNode := NewTreeNode(4, "Return")
	assignNode := NewTreeNode(5, "Assign")
	callNode := NewTreeNode(6, "Call")

	//Fake patternes declaration
	moduleNode := NewTreeNode(7, "Module")
	exprNode := NewTreeNode(8, "Expr")
	nameNode := NewTreeNode(9, "Name")

	root.AddChild(forNode)
	root.AddChild(assignNode)
	root.AddChild(nameNode)

	forNode.AddChild(ifNode)
	forNode.AddChild(moduleNode)

	ifNode.AddChild(returnNode)
	ifNode.AddChild(exprNode)

	assignNode.AddChild(callNode)

	patterns := extractor.extractPatterns(root)

	// Expected patterns (sorted alphabetically as per implementation)
	expected := []string{
		"Assign",
		"Call",
		"For",
		"FunctionDef",
		"If",
		"Return",
	}

	assert.Equal(t, expected, patterns)
}
