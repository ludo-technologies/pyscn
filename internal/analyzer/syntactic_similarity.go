package analyzer

// SyntacticSimilarityAnalyzer computes syntactic similarity using normalized
// AST hash comparison with Jaccard coefficient. This is used for Type-2 clone
// detection (syntactically identical but with different identifiers/literals).
//
// Unlike the previous APTED-based approach which measured tree edit distance,
// this implementation compares sets of normalized node hashes. This eliminates
// false positives from structurally similar but semantically different code,
// as only nodes with identical normalized structure contribute to similarity.
type SyntacticSimilarityAnalyzer struct {
	extractor *ASTFeatureExtractor
	converter *TreeConverter
}

// NewSyntacticSimilarityAnalyzer creates a new syntactic similarity analyzer
// using normalized AST hash comparison that ignores identifier and literal differences.
func NewSyntacticSimilarityAnalyzer() *SyntacticSimilarityAnalyzer {
	// Use ASTFeatureExtractor with includeLiterals=false to normalize
	// identifiers and literals, focusing only on structural patterns.
	extractor := NewASTFeatureExtractor().WithOptions(
		3,     // maxSubtreeHeight
		4,     // kGramSize
		true,  // includeTypes
		false, // includeLiterals - ignore literal values for Type-2
	)
	return &SyntacticSimilarityAnalyzer{
		extractor: extractor,
		converter: NewTreeConverter(),
	}
}

// NewSyntacticSimilarityAnalyzerWithOptions creates a syntactic similarity analyzer
// with configurable normalization options.
func NewSyntacticSimilarityAnalyzerWithOptions(ignoreLiterals, ignoreIdentifiers bool) *SyntacticSimilarityAnalyzer {
	// Both ignoreLiterals and ignoreIdentifiers map to includeLiterals=false
	// since our normalization strips both when includeLiterals is false.
	includeLiterals := !(ignoreLiterals || ignoreIdentifiers)
	extractor := NewASTFeatureExtractor().WithOptions(
		3,    // maxSubtreeHeight
		4,    // kGramSize
		true, // includeTypes
		includeLiterals,
	)
	return &SyntacticSimilarityAnalyzer{
		extractor: extractor,
		converter: NewTreeConverter(),
	}
}

// ComputeSimilarity computes the syntactic similarity between two code fragments
// using Jaccard coefficient of normalized AST hash sets.
// It ignores differences in identifier names and literal values, focusing only
// on the structural syntax pattern.
func (s *SyntacticSimilarityAnalyzer) ComputeSimilarity(f1, f2 *CodeFragment, calc *TFIDFCalculator) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}

	// Get or build tree nodes for fragments
	tree1 := s.getTreeNode(f1)
	tree2 := s.getTreeNode(f2)

	if tree1 == nil || tree2 == nil {
		return 0.0
	}

	// Extract normalized feature sets
	features1, _ := s.extractor.ExtractFeatures(tree1)
	features2, _ := s.extractor.ExtractFeatures(tree2)

	// Use TF-IDF + Cosine if calculator is provided and feature is enabled
	if calc != nil {
		v1 := calc.ToWeightedVector(features1)
		v2 := calc.ToWeightedVector(features2)
		return CosineSimilarity(v1, v2)
	}

	return jaccardSimilarity(features1, features2)
}

// ComputeDistance computes the syntactic distance between two code fragments.
// Returns 1 - similarity, so distance ranges from 0 (identical) to 1 (completely different).
// Returns 0.0 for nil inputs (no distance can be computed).
func (s *SyntacticSimilarityAnalyzer) ComputeDistance(f1, f2 *CodeFragment) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}
	return 1.0 - s.ComputeSimilarity(f1, f2, nil)
}

// jaccardSimilarity computes the Jaccard coefficient between two string sets.
// Jaccard(A, B) = |A ∩ B| / |A ∪ B|
func jaccardSimilarity(set1, set2 []string) float64 {
	if len(set1) == 0 && len(set2) == 0 {
		return 1.0 // Both empty = identical
	}
	if len(set1) == 0 || len(set2) == 0 {
		return 0.0 // One empty = no similarity
	}

	// Build hash set for set1
	s1 := make(map[string]struct{}, len(set1))
	for _, f := range set1 {
		s1[f] = struct{}{}
	}

	// Build hash set for set2
	s2 := make(map[string]struct{}, len(set2))
	for _, f := range set2 {
		s2[f] = struct{}{}
	}

	// Count intersection
	intersection := 0
	for f := range s1 {
		if _, ok := s2[f]; ok {
			intersection++
		}
	}

	// Union size = |A| + |B| - |A ∩ B|
	union := len(s1) + len(s2) - intersection

	if union == 0 {
		return 1.0
	}

	return float64(intersection) / float64(union)
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

// GetExtractor returns the underlying feature extractor (for advanced usage)
func (s *SyntacticSimilarityAnalyzer) GetExtractor() *ASTFeatureExtractor {
	return s.extractor
}
