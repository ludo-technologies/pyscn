package analyzer

// SyntacticSimilarityAnalyzer computes syntactic similarity using APTED with
// identifier and literal normalization. This is used for Type-2 clone detection
// (syntactically identical but with different identifiers/literals).
type SyntacticSimilarityAnalyzer struct {
	analyzer  *APTEDAnalyzer
	converter *TreeConverter
}

// NewSyntacticSimilarityAnalyzer creates a new syntactic similarity analyzer
// using a Python cost model that ignores identifier and literal differences.
func NewSyntacticSimilarityAnalyzer() *SyntacticSimilarityAnalyzer {
	// Use PythonCostModel with IgnoreLiterals=true, IgnoreIdentifiers=true
	// This makes the comparison focus on structure, treating all identifiers
	// and literals as equivalent.
	costModel := NewPythonCostModelWithConfig(true, true)
	return &SyntacticSimilarityAnalyzer{
		analyzer:  NewAPTEDAnalyzer(costModel),
		converter: NewTreeConverter(),
	}
}

// NewSyntacticSimilarityAnalyzerWithOptions creates a syntactic similarity analyzer
// with configurable normalization options.
func NewSyntacticSimilarityAnalyzerWithOptions(ignoreLiterals, ignoreIdentifiers bool) *SyntacticSimilarityAnalyzer {
	costModel := NewPythonCostModelWithConfig(ignoreLiterals, ignoreIdentifiers)
	return &SyntacticSimilarityAnalyzer{
		analyzer:  NewAPTEDAnalyzer(costModel),
		converter: NewTreeConverter(),
	}
}

// ComputeSimilarity computes the syntactic similarity between two code fragments.
// It ignores differences in identifier names and literal values, focusing only
// on the structural syntax pattern.
func (s *SyntacticSimilarityAnalyzer) ComputeSimilarity(f1, f2 *CodeFragment) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}

	// Get or build tree nodes for fragments
	tree1 := s.getTreeNode(f1)
	tree2 := s.getTreeNode(f2)

	if tree1 == nil || tree2 == nil {
		return 0.0
	}

	// Compute similarity using APTED with normalization
	return s.analyzer.ComputeSimilarity(tree1, tree2)
}

// ComputeDistance computes the syntactic edit distance between two code fragments.
func (s *SyntacticSimilarityAnalyzer) ComputeDistance(f1, f2 *CodeFragment) float64 {
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
func (s *SyntacticSimilarityAnalyzer) getTreeNode(f *CodeFragment) *TreeNode {
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
func (s *SyntacticSimilarityAnalyzer) GetName() string {
	return "syntactic"
}

// GetAnalyzer returns the underlying APTED analyzer (for advanced usage)
func (s *SyntacticSimilarityAnalyzer) GetAnalyzer() *APTEDAnalyzer {
	return s.analyzer
}
