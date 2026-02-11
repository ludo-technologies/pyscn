package analyzer

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRecursionDepthLimits tests that our recursion depth limits prevent stack overflow
func TestRecursionDepthLimits(t *testing.T) {
	// Create a deep linear tree that would cause stack overflow without depth limits
	root := NewTreeNode(0, "root")
	current := root

	// Create a tree with depth > 1000 to test our limits
	const deepDepth = 1200
	for i := 1; i < deepDepth; i++ {
		child := NewTreeNode(i, "node")
		current.AddChild(child)
		current = child
	}

	t.Run("Size with depth limit", func(t *testing.T) {
		// Should not crash and should return a reasonable value
		size := root.Size()
		assert.Greater(t, size, 0, "Size should be positive")
		assert.LessOrEqual(t, size, 1001, "Size should be limited by depth limit")
	})

	t.Run("Height with depth limit", func(t *testing.T) {
		// Should not crash and should return a reasonable value
		height := root.Height()
		assert.Greater(t, height, 0, "Height should be positive")
		assert.LessOrEqual(t, height, 1000, "Height should be limited by depth limit")
	})

	t.Run("GetSubtreeNodes with depth limit", func(t *testing.T) {
		// Should not crash and should return a reasonable number of nodes
		nodes := GetSubtreeNodes(root)
		assert.Greater(t, len(nodes), 0, "Should return some nodes")
		assert.LessOrEqual(t, len(nodes), 1001, "Node count should be limited by depth limit")
	})

	t.Run("APTED operations with deep trees", func(t *testing.T) {
		// Create another deep tree
		root2 := NewTreeNode(0, "root2")
		current2 := root2
		for i := 1; i < 800; i++ {
			child := NewTreeNode(i, "node2")
			current2.AddChild(child)
			current2 = child
		}

		// APTED operations should not crash
		analyzer := NewAPTEDAnalyzer(NewDefaultCostModel())

		distance := analyzer.ComputeDistanceTrees(root, root2)
		assert.GreaterOrEqual(t, distance, 0.0, "Distance should be non-negative")

		similarity := analyzer.ComputeSimilarityTrees(root, root2, nil)
		assert.GreaterOrEqual(t, similarity, 0.0, "Similarity should be non-negative")
		assert.LessOrEqual(t, similarity, 1.0, "Similarity should not exceed 1.0")
	})
}

// TestSizeWithDepthLimit tests the SizeWithDepthLimit method directly
func TestSizeWithDepthLimit(t *testing.T) {
	tests := []struct {
		name     string
		depth    int
		maxDepth int
		expected int
	}{
		{"shallow tree within limit", 3, 10, 4}, // root + 3 children in chain
		{"deep tree at limit", 5, 5, 6},         // root + 5 children in chain
		{"deep tree exceeds limit", 10, 3, 4},   // limited to depth 3
		{"zero depth limit", 5, 0, 1},           // should return 1 (treat as leaf)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a linear tree of specified depth
			root := NewTreeNode(0, "root")
			current := root
			for i := 1; i <= tt.depth; i++ {
				child := NewTreeNode(i, "node")
				current.AddChild(child)
				current = child
			}

			result := root.SizeWithDepthLimit(tt.maxDepth)
			assert.Equal(t, tt.expected, result, "Size with depth limit should match expected")
		})
	}
}

// TestHeightWithDepthLimit tests the HeightWithDepthLimit method directly
func TestHeightWithDepthLimit(t *testing.T) {
	tests := []struct {
		name     string
		depth    int
		maxDepth int
		expected int
	}{
		{"shallow tree within limit", 3, 10, 3},
		{"deep tree at limit", 5, 5, 5},
		{"deep tree exceeds limit", 10, 3, 3}, // limited to depth 3
		{"zero depth limit", 5, 0, 0},         // should return 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a linear tree of specified depth
			root := NewTreeNode(0, "root")
			current := root
			for i := 1; i <= tt.depth; i++ {
				child := NewTreeNode(i, "node")
				current.AddChild(child)
				current = child
			}

			result := root.HeightWithDepthLimit(tt.maxDepth)
			assert.Equal(t, tt.expected, result, "Height with depth limit should match expected")
		})
	}
}

// TestGetSubtreeNodesWithDepthLimit tests the GetSubtreeNodesWithDepthLimit function
func TestGetSubtreeNodesWithDepthLimit(t *testing.T) {
	// Create a tree with multiple levels
	root := NewTreeNode(0, "root")
	child1 := NewTreeNode(1, "child1")
	child2 := NewTreeNode(2, "child2")
	grandchild := NewTreeNode(3, "grandchild")

	root.AddChild(child1)
	root.AddChild(child2)
	child1.AddChild(grandchild)

	tests := []struct {
		name         string
		maxDepth     int
		expectedSize int
	}{
		{"full depth", 10, 4},   // All nodes
		{"limited depth", 2, 3}, // root + 2 children (grandchild excluded)
		{"minimal depth", 1, 1}, // just root
		{"zero depth", 0, 0},    // no nodes
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodes := GetSubtreeNodesWithDepthLimit(root, tt.maxDepth)
			assert.Equal(t, tt.expectedSize, len(nodes), "Should return expected number of nodes")
		})
	}
}

// TestAPTEDRecursionProtection tests that APTED methods handle deep recursion safely
func TestAPTEDRecursionProtection(t *testing.T) {
	// Create two moderately deep trees to test the protection
	tree1 := createDeepLinearTree(500, "tree1")
	tree2 := createDeepLinearTree(400, "tree2")

	analyzer := NewAPTEDAnalyzer(NewDefaultCostModel())

	t.Run("compute distance with deep trees", func(t *testing.T) {
		// This should not panic or crash
		distance := analyzer.ComputeDistanceTrees(tree1, tree2)
		assert.GreaterOrEqual(t, distance, 0.0, "Distance should be non-negative")

		// Should complete in reasonable time (not hang)
		require.NotEqual(t, math.Inf(1), distance, "Distance should be finite")
	})

	t.Run("compute similarity with deep trees", func(t *testing.T) {
		// This should not panic or crash
		similarity := analyzer.ComputeSimilarityTrees(tree1, tree2, nil)
		assert.GreaterOrEqual(t, similarity, 0.0, "Similarity should be non-negative")
		assert.LessOrEqual(t, similarity, 1.0, "Similarity should not exceed 1.0")
	})
}

// Helper function to create a deep linear tree
func createDeepLinearTree(depth int, prefix string) *TreeNode {
	root := NewTreeNode(0, prefix+"_root")
	current := root

	for i := 1; i < depth; i++ {
		child := NewTreeNode(i, prefix+"_node")
		current.AddChild(child)
		current = child
	}

	return root
}

// Benchmark to ensure performance is reasonable with depth limits
func BenchmarkRecursionProtection(b *testing.B) {
	tree := createDeepLinearTree(1000, "bench")

	b.Run("Size", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = tree.Size()
		}
	})

	b.Run("Height", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = tree.Height()
		}
	})

	b.Run("GetSubtreeNodes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = GetSubtreeNodes(tree)
		}
	})
}
