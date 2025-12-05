package analyzer

import (
	"math"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

// SemanticSimilarityAnalyzer computes semantic similarity using CFG (Control Flow Graph)
// feature comparison. This is used for Type-4 clone detection (functionally similar
// code with different syntax).
type SemanticSimilarityAnalyzer struct {
	cfgBuilder *CFGBuilder
}

// NewSemanticSimilarityAnalyzer creates a new semantic similarity analyzer
func NewSemanticSimilarityAnalyzer() *SemanticSimilarityAnalyzer {
	return &SemanticSimilarityAnalyzer{
		cfgBuilder: NewCFGBuilder(),
	}
}

// CFGFeatures captures key structural properties of a control flow graph
type CFGFeatures struct {
	BlockCount       int              // Number of basic blocks
	EdgeCount        int              // Number of edges
	EdgeTypeCounts   map[EdgeType]int // Distribution of edge types
	CyclomaticNumber int              // Cyclomatic complexity: V(G) = E - N + 2P
	BranchingFactor  float64          // Average number of successors per block
	LoopEdgeCount    int              // Number of loop back-edges
	ConditionalCount int              // Number of conditional branches
}

// ComputeSimilarity computes the semantic similarity between two code fragments
// by comparing their CFG structures.
func (s *SemanticSimilarityAnalyzer) ComputeSimilarity(f1, f2 *CodeFragment) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}

	// Build CFGs for both fragments
	cfg1, err1 := s.buildCFGFromFragment(f1)
	cfg2, err2 := s.buildCFGFromFragment(f2)

	if err1 != nil || err2 != nil || cfg1 == nil || cfg2 == nil {
		return 0.0
	}

	// Extract features from both CFGs
	features1 := s.extractCFGFeatures(cfg1)
	features2 := s.extractCFGFeatures(cfg2)

	// Compare features to compute similarity
	return s.compareCFGFeatures(features1, features2)
}

// buildCFGFromFragment builds a CFG from a code fragment
func (s *SemanticSimilarityAnalyzer) buildCFGFromFragment(f *CodeFragment) (*CFG, error) {
	if f.ASTNode == nil {
		return nil, nil
	}

	// Create a fresh builder for each fragment to avoid state issues
	builder := NewCFGBuilder()
	return builder.Build(f.ASTNode)
}

// extractCFGFeatures extracts structural features from a CFG
func (s *SemanticSimilarityAnalyzer) extractCFGFeatures(cfg *CFG) *CFGFeatures {
	features := &CFGFeatures{
		EdgeTypeCounts: make(map[EdgeType]int),
	}

	if cfg == nil {
		return features
	}

	// Count blocks
	features.BlockCount = cfg.Size()

	// Walk through all blocks to extract edge information
	totalSuccessors := 0
	for _, block := range cfg.Blocks {
		for _, edge := range block.Successors {
			features.EdgeCount++
			features.EdgeTypeCounts[edge.Type]++

			// Count specific edge types
			if edge.Type == EdgeLoop {
				features.LoopEdgeCount++
			}
			if edge.Type == EdgeCondTrue || edge.Type == EdgeCondFalse {
				features.ConditionalCount++
			}
		}
		totalSuccessors += len(block.Successors)
	}

	// Calculate branching factor
	if features.BlockCount > 0 {
		features.BranchingFactor = float64(totalSuccessors) / float64(features.BlockCount)
	}

	// Calculate cyclomatic complexity: V(G) = E - N + 2P
	// For a single connected component, P = 1
	features.CyclomaticNumber = features.EdgeCount - features.BlockCount + 2

	return features
}

// compareCFGFeatures compares two CFG feature sets and returns a similarity score
func (s *SemanticSimilarityAnalyzer) compareCFGFeatures(f1, f2 *CFGFeatures) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}

	// Handle edge cases
	if f1.BlockCount == 0 && f2.BlockCount == 0 {
		return 1.0 // Both empty
	}
	if f1.BlockCount == 0 || f2.BlockCount == 0 {
		return 0.0 // One empty
	}

	// Calculate individual similarity components
	var weights []float64
	var similarities []float64

	// 1. Block count similarity (weight: 0.2)
	blockSim := s.computeCountSimilarity(f1.BlockCount, f2.BlockCount)
	weights = append(weights, 0.2)
	similarities = append(similarities, blockSim)

	// 2. Edge count similarity (weight: 0.15)
	edgeSim := s.computeCountSimilarity(f1.EdgeCount, f2.EdgeCount)
	weights = append(weights, 0.15)
	similarities = append(similarities, edgeSim)

	// 3. Cyclomatic complexity similarity (weight: 0.25)
	ccSim := s.computeCountSimilarity(f1.CyclomaticNumber, f2.CyclomaticNumber)
	weights = append(weights, 0.25)
	similarities = append(similarities, ccSim)

	// 4. Edge type distribution similarity (weight: 0.25)
	edgeTypeSim := s.compareEdgeDistributions(f1.EdgeTypeCounts, f2.EdgeTypeCounts)
	weights = append(weights, 0.25)
	similarities = append(similarities, edgeTypeSim)

	// 5. Branching factor similarity (weight: 0.1)
	branchSim := s.computeFloatSimilarity(f1.BranchingFactor, f2.BranchingFactor)
	weights = append(weights, 0.1)
	similarities = append(similarities, branchSim)

	// 6. Loop/conditional structure similarity (weight: 0.05)
	loopSim := s.computeCountSimilarity(f1.LoopEdgeCount, f2.LoopEdgeCount)
	condSim := s.computeCountSimilarity(f1.ConditionalCount, f2.ConditionalCount)
	structureSim := (loopSim + condSim) / 2.0
	weights = append(weights, 0.05)
	similarities = append(similarities, structureSim)

	// Compute weighted average
	var totalWeight float64
	var weightedSum float64
	for i := range weights {
		totalWeight += weights[i]
		weightedSum += weights[i] * similarities[i]
	}

	if totalWeight == 0 {
		return 0.0
	}

	similarity := weightedSum / totalWeight

	// Clamp to [0, 1]
	return math.Max(0.0, math.Min(1.0, similarity))
}

// computeCountSimilarity computes similarity between two integer counts
func (s *SemanticSimilarityAnalyzer) computeCountSimilarity(a, b int) float64 {
	if a == 0 && b == 0 {
		return 1.0
	}

	maxVal := math.Max(float64(a), float64(b))
	if maxVal == 0 {
		return 1.0
	}

	diff := math.Abs(float64(a - b))
	return 1.0 - (diff / maxVal)
}

// computeFloatSimilarity computes similarity between two float values
func (s *SemanticSimilarityAnalyzer) computeFloatSimilarity(a, b float64) float64 {
	if a == 0 && b == 0 {
		return 1.0
	}

	maxVal := math.Max(a, b)
	if maxVal == 0 {
		return 1.0
	}

	diff := math.Abs(a - b)
	return 1.0 - (diff / maxVal)
}

// compareEdgeDistributions compares two edge type distributions using cosine similarity
func (s *SemanticSimilarityAnalyzer) compareEdgeDistributions(dist1, dist2 map[EdgeType]int) float64 {
	if len(dist1) == 0 && len(dist2) == 0 {
		return 1.0
	}
	if len(dist1) == 0 || len(dist2) == 0 {
		return 0.0
	}

	// Get all edge types
	allTypes := make(map[EdgeType]bool)
	for t := range dist1 {
		allTypes[t] = true
	}
	for t := range dist2 {
		allTypes[t] = true
	}

	// Compute cosine similarity
	var dotProduct float64
	var norm1 float64
	var norm2 float64

	for t := range allTypes {
		v1 := float64(dist1[t])
		v2 := float64(dist2[t])

		dotProduct += v1 * v2
		norm1 += v1 * v1
		norm2 += v2 * v2
	}

	norm1 = math.Sqrt(norm1)
	norm2 = math.Sqrt(norm2)

	if norm1 == 0 || norm2 == 0 {
		return 0.0
	}

	return dotProduct / (norm1 * norm2)
}

// GetName returns the name of this analyzer
func (s *SemanticSimilarityAnalyzer) GetName() string {
	return "semantic"
}

// BuildCFG builds a CFG from a parser.Node (exposed for testing)
func (s *SemanticSimilarityAnalyzer) BuildCFG(node *parser.Node) (*CFG, error) {
	builder := NewCFGBuilder()
	return builder.Build(node)
}

// ExtractFeatures extracts CFG features (exposed for testing)
func (s *SemanticSimilarityAnalyzer) ExtractFeatures(cfg *CFG) *CFGFeatures {
	return s.extractCFGFeatures(cfg)
}
