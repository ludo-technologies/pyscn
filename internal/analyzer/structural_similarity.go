package analyzer

// StructuralSimilarityAnalyzer computes structural similarity using APTED tree edit distance.
// This is used for Type-3 clone detection (near-miss clones with modifications).
type StructuralSimilarityAnalyzer struct {
	analyzer  *APTEDAnalyzer
	converter *TreeConverter
}

// NewStructuralSimilarityAnalyzer creates a new structural similarity analyzer
// using the standard Python cost model (no normalization).
func NewStructuralSimilarityAnalyzer() *StructuralSimilarityAnalyzer {
	costModel := NewPythonCostModel()
	return &StructuralSimilarityAnalyzer{
		analyzer:  NewAPTEDAnalyzer(costModel),
		converter: NewTreeConverter(),
	}
}

// NewStructuralSimilarityAnalyzerWithCostModel creates a structural similarity analyzer
// with a custom cost model.
func NewStructuralSimilarityAnalyzerWithCostModel(costModel CostModel) *StructuralSimilarityAnalyzer {
	return &StructuralSimilarityAnalyzer{
		analyzer:  NewAPTEDAnalyzer(costModel),
		converter: NewTreeConverter(),
	}
}

// ComputeSimilarity computes the structural similarity between two code fragments
// using APTED tree edit distance.
func (s *StructuralSimilarityAnalyzer) ComputeSimilarity(f1, f2 *CodeFragment, _ *TFIDFCalculator) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}

	// Get or build tree nodes for fragments
	tree1 := s.getTreeNode(f1)
	tree2 := s.getTreeNode(f2)

	if tree1 == nil || tree2 == nil {
		return 0.0
	}

	// Compute similarity using APTED
	return s.analyzer.ComputeSimilarity(tree1, tree2, nil)
}

// ComputeDistance computes the edit distance between two code fragments.
// This is useful for additional metrics beyond similarity.
func (s *StructuralSimilarityAnalyzer) ComputeDistance(f1, f2 *CodeFragment) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}

	tree1 := s.getTreeNode(f1)
	tree2 := s.getTreeNode(f2)

	if tree1 == nil || tree2 == nil {
		return 0.0
	}

	return s.analyzer.ComputeDistance(tree1, tree2)
}

// getTreeNode retrieves or builds the tree node for a code fragment
func (s *StructuralSimilarityAnalyzer) getTreeNode(f *CodeFragment) *TreeNode {
	// Use cached TreeNode if available
	if f.TreeNode != nil {
		return f.TreeNode
	}

	// Build from AST node if available
	if f.ASTNode != nil {
		tree := s.converter.ConvertAST(f.ASTNode)
		if tree != nil {
			PrepareTreeForAPTED(tree)
		}
		return tree
	}

	return nil
}

// GetName returns the name of this analyzer
func (s *StructuralSimilarityAnalyzer) GetName() string {
	return "structural"
}

// GetAnalyzer returns the underlying APTED analyzer (for advanced usage)
func (s *StructuralSimilarityAnalyzer) GetAnalyzer() *APTEDAnalyzer {
	return s.analyzer
}
