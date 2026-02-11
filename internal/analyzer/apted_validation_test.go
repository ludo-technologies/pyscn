package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAPTEDAnalyzer_NilChecks(t *testing.T) {
	costModel := NewPythonCostModel()
	analyzer := NewAPTEDAnalyzer(costModel)

	t.Run("ComputeDistance with nil trees", func(t *testing.T) {
		// Both nil
		distance := analyzer.ComputeDistanceTrees(nil, nil)
		assert.Equal(t, 0.0, distance, "Distance between two nil trees should be 0")

		// Create a test tree for comparison
		tree := NewTreeNode(1, "test")
		tree.AddChild(NewTreeNode(2, "child"))

		// First nil
		distance = analyzer.ComputeDistanceTrees(nil, tree)
		assert.Greater(t, distance, 0.0, "Distance from nil to tree should be > 0")

		// Second nil
		distance = analyzer.ComputeDistanceTrees(tree, nil)
		assert.Greater(t, distance, 0.0, "Distance from tree to nil should be > 0")
	})

	t.Run("ComputeSimilarity with nil trees", func(t *testing.T) {
		// Both nil
		similarity := analyzer.ComputeSimilarityTrees(nil, nil, nil)
		assert.Equal(t, 1.0, similarity, "Similarity between two nil trees should be 1.0")

		// Create a test tree for comparison
		tree := NewTreeNode(1, "test")
		tree.AddChild(NewTreeNode(2, "child"))

		// First nil
		similarity = analyzer.ComputeSimilarityTrees(nil, tree, nil)
		assert.Equal(t, 0.0, similarity, "Similarity from nil to tree should be 0.0")

		// Second nil
		similarity = analyzer.ComputeSimilarityTrees(tree, nil, nil)
		assert.Equal(t, 0.0, similarity, "Similarity from tree to nil should be 0.0")
	})

	t.Run("computeApproximateDistance with nil trees", func(t *testing.T) {
		// Both nil
		distance := analyzer.computeApproximateDistance(nil, nil)
		assert.Equal(t, 0.0, distance, "Approximate distance between two nil trees should be 0")

		// Create a test tree for comparison
		tree := NewTreeNode(1, "test")
		tree.AddChild(NewTreeNode(2, "child"))

		// First nil
		distance = analyzer.computeApproximateDistance(nil, tree)
		assert.Equal(t, float64(tree.Size()), distance, "Approximate distance from nil should equal tree size")

		// Second nil
		distance = analyzer.computeApproximateDistance(tree, nil)
		assert.Equal(t, float64(tree.Size()), distance, "Approximate distance to nil should equal tree size")
	})

	t.Run("Valid trees", func(t *testing.T) {
		tree1 := NewTreeNode(1, "function")
		tree1.AddChild(NewTreeNode(2, "param"))
		tree1.AddChild(NewTreeNode(3, "body"))

		tree2 := NewTreeNode(4, "function")
		tree2.AddChild(NewTreeNode(5, "param"))
		tree2.AddChild(NewTreeNode(6, "body"))

		// Should work normally
		distance := analyzer.ComputeDistanceTrees(tree1, tree2)
		assert.GreaterOrEqual(t, distance, 0.0, "Distance should be non-negative")

		similarity := analyzer.ComputeSimilarityTrees(tree1, tree2, nil)
		assert.GreaterOrEqual(t, similarity, 0.0, "Similarity should be >= 0")
		assert.LessOrEqual(t, similarity, 1.0, "Similarity should be <= 1")
	})
}
