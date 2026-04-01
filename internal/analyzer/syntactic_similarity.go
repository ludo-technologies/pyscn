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
func (s *SyntacticSimilarityAnalyzer) ComputeSimilarity(f1, f2 *CodeFragment) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}

	// Use pre-computed features if available (avoids redundant tree traversal)
	if len(f1.Features) > 0 && len(f2.Features) > 0 {
		return jaccardSimilarity(f1.Features, f2.Features)
	}

	// Get or build tree nodes for fragments
	tree1 := s.getTreeNode(f1)
	tree2 := s.getTreeNode(f2)

	if tree1 == nil || tree2 == nil {
		return 0.0
	}

	// Extract normalized feature sets
	features1, err1 := s.extractor.ExtractFeatures(tree1)
	features2, err2 := s.extractor.ExtractFeatures(tree2)

	if err1 != nil || err2 != nil {
		return 0.0
	}

	// Compute Jaccard similarity
	return jaccardSimilarity(features1, features2)
}

// ComputeDistance computes the syntactic distance between two code fragments.
// Returns 1 - similarity, so distance ranges from 0 (identical) to 1 (completely different).
// Returns 0.0 for nil inputs (no distance can be computed).
func (s *SyntacticSimilarityAnalyzer) ComputeDistance(f1, f2 *CodeFragment) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}
	return 1.0 - s.ComputeSimilarity(f1, f2)
}

// jaccardSimilarity computes the Jaccard coefficient between two string slices.
// Jaccard(A, B) = |A ∩ B| / |A ∪ B|
// If both slices are sorted (as produced by ASTFeatureExtractor.ExtractFeatures),
// uses an O(n+m) merge-join without hash map allocation.
func jaccardSimilarity(set1, set2 []string) float64 {
	if len(set1) == 0 && len(set2) == 0 {
		return 1.0 // Both empty = identical
	}
	if len(set1) == 0 || len(set2) == 0 {
		return 0.0 // One empty = no similarity
	}

	// Merge-join on sorted slices: count unique elements and intersection
	i, j := 0, 0
	intersection := 0
	union := 0

	for i < len(set1) && j < len(set2) {
		if set1[i] == set2[j] {
			intersection++
			union++
			// Skip duplicates in both
			val := set1[i]
			for i < len(set1) && set1[i] == val {
				i++
			}
			for j < len(set2) && set2[j] == val {
				j++
			}
		} else if set1[i] < set2[j] {
			union++
			val := set1[i]
			for i < len(set1) && set1[i] == val {
				i++
			}
		} else {
			union++
			val := set2[j]
			for j < len(set2) && set2[j] == val {
				j++
			}
		}
	}
	// Count remaining unique elements
	for i < len(set1) {
		union++
		val := set1[i]
		for i < len(set1) && set1[i] == val {
			i++
		}
	}
	for j < len(set2) {
		union++
		val := set2[j]
		for j < len(set2) && set2[j] == val {
			j++
		}
	}

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
