package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

func TestNewASTFeatureExtractor(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	assert.Equal(t, 3, extractor.maxSubtreeHeight)
	assert.Equal(t, 4, extractor.kGramSize)
	assert.True(t, extractor.includeTypes)
	assert.False(t, extractor.includeLiterals)
	assert.True(t, extractor.includeStructure)
}

func TestNewASTFeatureExtractorWithConfig(t *testing.T) {
	extractor := NewASTFeatureExtractorWithConfig(5, 3, false, true, false)
	
	assert.Equal(t, 5, extractor.maxSubtreeHeight)
	assert.Equal(t, 3, extractor.kGramSize)
	assert.False(t, extractor.includeTypes)
	assert.True(t, extractor.includeLiterals)
	assert.False(t, extractor.includeStructure)
}

func TestExtractFeatures_NilTree(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	features, err := extractor.ExtractFeatures(nil)
	
	assert.NoError(t, err)
	assert.Empty(t, features)
}

func TestExtractFeatures_SimpleTree(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	// Create a simple tree: root with two children
	root := &TreeNode{
		ID:    0,
		Label: "FunctionDef",
		Children: []*TreeNode{
			{ID: 1, Label: "Name", Children: []*TreeNode{}},
			{ID: 2, Label: "Block", Children: []*TreeNode{
				{ID: 3, Label: "Return", Children: []*TreeNode{}},
			}},
		},
	}
	
	features, err := extractor.ExtractFeatures(root)
	
	require.NoError(t, err)
	assert.NotEmpty(t, features)
	
	// Check that we have different types of features
	hasSubtreeFeature := false
	hasKgramFeature := false
	hasPatternFeature := false
	hasTypeCountFeature := false
	
	for _, feature := range features {
		if len(feature) > 8 && feature[:8] == "subtree:" {
			hasSubtreeFeature = true
		}
		if len(feature) > 6 && feature[:6] == "kgram:" {
			hasKgramFeature = true
		}
		if len(feature) > 8 && feature[:8] == "pattern:" {
			hasPatternFeature = true
		}
		if len(feature) > 11 && feature[:11] == "type_count:" {
			hasTypeCountFeature = true
		}
	}
	
	assert.True(t, hasSubtreeFeature, "Should have subtree features")
	assert.True(t, hasKgramFeature, "Should have k-gram features")
	assert.True(t, hasPatternFeature, "Should have pattern features")
	assert.True(t, hasTypeCountFeature, "Should have type count features")
}

func TestExtractSubtreeHashes(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	// Create a tree with depth > 1
	root := &TreeNode{
		ID:    0,
		Label: "FunctionDef",
		Children: []*TreeNode{
			{ID: 1, Label: "Name", Children: []*TreeNode{}},
			{ID: 2, Label: "Block", Children: []*TreeNode{
				{ID: 3, Label: "Return", Children: []*TreeNode{}},
			}},
		},
	}
	
	hashes := extractor.ExtractSubtreeHashes(root, 2)
	
	assert.NotEmpty(t, hashes)
	
	// Check that hashes are unique (no duplicates)
	hashSet := make(map[string]bool)
	for _, hash := range hashes {
		assert.False(t, hashSet[hash], "Hash should be unique: %s", hash)
		hashSet[hash] = true
	}
}

func TestExtractSubtreeHashes_MaxHeightZero(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	root := &TreeNode{ID: 0, Label: "Test"}
	
	hashes := extractor.ExtractSubtreeHashes(root, 0)
	
	assert.Empty(t, hashes)
}

func TestExtractNodeSequences(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	// Create a linear tree for predictable sequences
	root := &TreeNode{
		ID:    0,
		Label: "FunctionDef",
		Children: []*TreeNode{
			{ID: 1, Label: "Name", Children: []*TreeNode{
				{ID: 2, Label: "Identifier", Children: []*TreeNode{}},
			}},
			{ID: 3, Label: "Block", Children: []*TreeNode{}},
		},
	}
	
	sequences := extractor.ExtractNodeSequences(root, 3)
	
	assert.NotEmpty(t, sequences)
	
	// Check that sequences contain expected patterns
	found := false
	for _, seq := range sequences {
		if seq == "functiondef_name_identifier" {
			found = true
			break
		}
	}
	assert.True(t, found, "Should contain expected sequence")
}

func TestExtractNodeSequences_KTooLarge(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	root := &TreeNode{ID: 0, Label: "Test"}
	
	sequences := extractor.ExtractNodeSequences(root, 10) // k larger than tree
	
	assert.Empty(t, sequences)
}

func TestNormalizeLabel(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	tests := []struct {
		input    string
		expected string
	}{
		{"FunctionDef", "functiondef"},
		{"NodeFunctionDef", "functiondef"},
		{"FunctionDefNode", "functiondef"},
		{"UPPERCASE", "uppercase"},
		{"MixedCase", "mixedcase"},
	}
	
	for _, test := range tests {
		result := extractor.normalizeLabel(test.input)
		assert.Equal(t, test.expected, result, "Input: %s", test.input)
	}
}

func TestComputeSubtreeHash_SameTrees(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	// Create two identical trees
	tree1 := &TreeNode{
		ID:    0,
		Label: "FunctionDef",
		Children: []*TreeNode{
			{ID: 1, Label: "Name", Children: []*TreeNode{}},
		},
	}
	
	tree2 := &TreeNode{
		ID:    10, // Different ID
		Label: "FunctionDef",
		Children: []*TreeNode{
			{ID: 11, Label: "Name", Children: []*TreeNode{}},
		},
	}
	
	hash1 := extractor.computeSubtreeHash(tree1, 3)
	hash2 := extractor.computeSubtreeHash(tree2, 3)
	
	assert.Equal(t, hash1, hash2, "Identical trees should have same hash")
}

func TestComputeSubtreeHash_DifferentTrees(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	tree1 := &TreeNode{
		ID:    0,
		Label: "FunctionDef",
		Children: []*TreeNode{
			{ID: 1, Label: "Name", Children: []*TreeNode{}},
		},
	}
	
	tree2 := &TreeNode{
		ID:    0,
		Label: "ClassDef", // Different label
		Children: []*TreeNode{
			{ID: 1, Label: "Name", Children: []*TreeNode{}},
		},
	}
	
	hash1 := extractor.computeSubtreeHash(tree1, 3)
	hash2 := extractor.computeSubtreeHash(tree2, 3)
	
	assert.NotEqual(t, hash1, hash2, "Different trees should have different hashes")
}

func TestExtractStructuralPatterns(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	// Create a tree with control structures
	root := &TreeNode{
		ID:    0,
		Label: "FunctionDef",
		Children: []*TreeNode{
			{ID: 1, Label: "If", Children: []*TreeNode{
				{ID: 2, Label: "For", Children: []*TreeNode{}},
			}},
			{ID: 3, Label: "While", Children: []*TreeNode{}},
		},
	}
	
	patterns := extractor.extractStructuralPatterns(root)
	
	assert.NotEmpty(t, patterns)
	
	// Check for control structure patterns
	hasControlPattern := false
	hasDepthPattern := false
	hasBranchingPattern := false
	
	for _, pattern := range patterns {
		if len(pattern) > 8 && pattern[:8] == "control:" {
			hasControlPattern = true
		}
		if len(pattern) > 6 && pattern[:6] == "depth:" {
			hasDepthPattern = true
		}
		if len(pattern) > 13 && pattern[:13] == "avg_branching" {
			hasBranchingPattern = true
		}
	}
	
	assert.True(t, hasControlPattern, "Should have control structure patterns")
	assert.True(t, hasDepthPattern, "Should have depth patterns")
	assert.True(t, hasBranchingPattern, "Should have branching patterns")
}

func TestCountControlStructures(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	root := &TreeNode{
		ID:    0,
		Label: "FunctionDef",
		Children: []*TreeNode{
			{ID: 1, Label: "If", Children: []*TreeNode{}},
			{ID: 2, Label: "For", Children: []*TreeNode{}},
			{ID: 3, Label: "If", Children: []*TreeNode{}}, // Another if
		},
	}
	
	counts := extractor.countControlStructures(root)
	
	assert.Equal(t, 1, counts["function"])
	assert.Equal(t, 2, counts["if"])
	assert.Equal(t, 1, counts["for"])
	assert.Equal(t, 0, counts["while"]) // Should be 0 for missing structures
}

func TestComputeAverageBranchingFactor(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	// Create a tree where:
	// Root has 2 children
	// First child has 1 child
	// Second child has 0 children
	// Third level has 0 children
	// Total: 4 nodes, 3 children total -> avg = 3/4 = 0.75
	root := &TreeNode{
		ID:    0,
		Label: "Root",
		Children: []*TreeNode{
			{ID: 1, Label: "Child1", Children: []*TreeNode{
				{ID: 3, Label: "Grandchild", Children: []*TreeNode{}},
			}},
			{ID: 2, Label: "Child2", Children: []*TreeNode{}},
		},
	}
	
	avgBranching := extractor.computeAverageBranchingFactor(root)
	
	expected := 3.0 / 4.0 // 3 total children across 4 nodes
	assert.InDelta(t, expected, avgBranching, 0.001)
}

func TestExtractNodeTypeDistribution(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	root := &TreeNode{
		ID:    0,
		Label: "FunctionDef",
		Children: []*TreeNode{
			{ID: 1, Label: "Name", Children: []*TreeNode{}},
			{ID: 2, Label: "Name", Children: []*TreeNode{}}, // Duplicate
			{ID: 3, Label: "Block", Children: []*TreeNode{}},
		},
	}
	
	distribution := extractor.extractNodeTypeDistribution(root)
	
	assert.Equal(t, 1, distribution["functiondef"])
	assert.Equal(t, 2, distribution["name"])
	assert.Equal(t, 1, distribution["block"])
}

func TestExtractLiterals_WithOriginalNode(t *testing.T) {
	extractor := NewASTFeatureExtractorWithConfig(3, 4, true, true, true)
	
	// Create a tree node with original parser node
	originalNode := &parser.Node{
		Type:  parser.NodeConstant,
		Value: "test_string",
	}
	
	root := &TreeNode{
		ID:           0,
		Label:        "Constant",
		OriginalNode: originalNode,
		Children:     []*TreeNode{},
	}
	
	literals := extractor.extractLiterals(root)
	
	assert.NotEmpty(t, literals)
}

func TestExtractLiteralsRecursive_NoOriginalNode(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	root := &TreeNode{
		ID:           0,
		Label:        "Test",
		OriginalNode: nil, // No original node
		Children:     []*TreeNode{},
	}
	
	literals := extractor.extractLiterals(root)
	
	assert.Empty(t, literals)
}

func TestNormalizeLiteral(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "string_literal"},
		{`'world'`, "string_literal"},
		{"42", "int_literal"},
		{"3.14", "float_literal"},
		{"true", "true"},
		{"false", "false"},
		{"None", "none"},
	}
	
	for _, test := range tests {
		result := extractor.normalizeLiteral(test.input)
		assert.Equal(t, test.expected, result, "Input: %s", test.input)
	}
}

func TestExtractFeatures_DisabledComponents(t *testing.T) {
	// Create extractor with all components disabled
	extractor := NewASTFeatureExtractorWithConfig(3, 4, false, false, false)
	
	root := &TreeNode{
		ID:    0,
		Label: "Test",
		Children: []*TreeNode{
			{ID: 1, Label: "Child", Children: []*TreeNode{}},
		},
	}
	
	features, err := extractor.ExtractFeatures(root)
	
	require.NoError(t, err)
	
	// Should have no features since all components are disabled
	assert.Empty(t, features)
}

func TestExtractFeatures_Consistency(t *testing.T) {
	extractor := NewASTFeatureExtractor()
	
	root := &TreeNode{
		ID:    0,
		Label: "FunctionDef",
		Children: []*TreeNode{
			{ID: 1, Label: "Name", Children: []*TreeNode{}},
			{ID: 2, Label: "Block", Children: []*TreeNode{}},
		},
	}
	
	// Extract features multiple times
	features1, err1 := extractor.ExtractFeatures(root)
	features2, err2 := extractor.ExtractFeatures(root)
	
	require.NoError(t, err1)
	require.NoError(t, err2)
	
	// Results should be consistent
	assert.Equal(t, features1, features2)
}

func BenchmarkExtractFeatures(b *testing.B) {
	extractor := NewASTFeatureExtractor()
	
	// Create a moderately complex tree
	root := &TreeNode{
		ID:    0,
		Label: "FunctionDef",
		Children: []*TreeNode{
			{ID: 1, Label: "Name", Children: []*TreeNode{}},
			{ID: 2, Label: "Args", Children: []*TreeNode{
				{ID: 3, Label: "Arg", Children: []*TreeNode{}},
				{ID: 4, Label: "Arg", Children: []*TreeNode{}},
			}},
			{ID: 5, Label: "Block", Children: []*TreeNode{
				{ID: 6, Label: "If", Children: []*TreeNode{
					{ID: 7, Label: "Compare", Children: []*TreeNode{}},
					{ID: 8, Label: "Block", Children: []*TreeNode{
						{ID: 9, Label: "Return", Children: []*TreeNode{}},
					}},
				}},
				{ID: 10, Label: "For", Children: []*TreeNode{
					{ID: 11, Label: "Target", Children: []*TreeNode{}},
					{ID: 12, Label: "Iter", Children: []*TreeNode{}},
					{ID: 13, Label: "Block", Children: []*TreeNode{}},
				}},
			}},
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.ExtractFeatures(root)
		if err != nil {
			b.Fatal(err)
		}
	}
}