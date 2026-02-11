package analyzer

import (
	"math"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// SemanticSimilarityAnalyzer computes semantic similarity using CFG (Control Flow Graph)
// and optionally DFA (Data Flow Analysis) feature comparison. This is used for Type-4
// clone detection (functionally similar code with different syntax).
type SemanticSimilarityAnalyzer struct {
	enableDFA        bool    // Enable DFA-based analysis
	cfgFeatureWeight float64 // Weight for CFG features (default: 0.6)
	dfaFeatureWeight float64 // Weight for DFA features (default: 0.4)
}

// NewSemanticSimilarityAnalyzer creates a new semantic similarity analyzer
func NewSemanticSimilarityAnalyzer() *SemanticSimilarityAnalyzer {
	return &SemanticSimilarityAnalyzer{
		enableDFA:        false,
		cfgFeatureWeight: domain.DefaultCFGFeatureWeight,
		dfaFeatureWeight: domain.DefaultDFAFeatureWeight,
	}
}

// NewSemanticSimilarityAnalyzerWithDFA creates a new analyzer with DFA enabled
func NewSemanticSimilarityAnalyzerWithDFA() *SemanticSimilarityAnalyzer {
	return &SemanticSimilarityAnalyzer{
		enableDFA:        true,
		cfgFeatureWeight: domain.DefaultCFGFeatureWeight,
		dfaFeatureWeight: domain.DefaultDFAFeatureWeight,
	}
}

// SetEnableDFA enables or disables DFA analysis
func (s *SemanticSimilarityAnalyzer) SetEnableDFA(enable bool) {
	s.enableDFA = enable
}

// SetWeights sets the CFG and DFA feature weights
func (s *SemanticSimilarityAnalyzer) SetWeights(cfgWeight, dfaWeight float64) {
	s.cfgFeatureWeight = cfgWeight
	s.dfaFeatureWeight = dfaWeight
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

// ComputeDistance computes the semantic distance (1 - similarity) between two code fragments.
func (s *SemanticSimilarityAnalyzer) ComputeDistance(f1, f2 *CodeFragment) float64 {
	if f1 == nil && f2 == nil {
		return 0.0
	}
	if f1 == nil || f2 == nil {
		return 1.0
	}
	return 1.0 - s.ComputeSimilarity(f1, f2, nil)
}

// ComputeSimilarity computes the semantic similarity between two code fragments
// by comparing their CFG structures and optionally DFA features.
func (s *SemanticSimilarityAnalyzer) ComputeSimilarity(f1, f2 *CodeFragment, _ *TFIDFCalculator) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}

	// Build CFGs for both fragments
	cfg1, err1 := s.buildCFGFromFragment(f1)
	cfg2, err2 := s.buildCFGFromFragment(f2)

	if err1 != nil || err2 != nil || cfg1 == nil || cfg2 == nil {
		return 0.0
	}

	// Extract CFG features from both CFGs
	cfgFeatures1 := s.extractCFGFeatures(cfg1)
	cfgFeatures2 := s.extractCFGFeatures(cfg2)

	// Compare CFG features to compute similarity
	cfgSimilarity := s.compareCFGFeatures(cfgFeatures1, cfgFeatures2)

	// If DFA is not enabled, return CFG similarity only
	if !s.enableDFA {
		return cfgSimilarity
	}

	// Build DFA info for both CFGs
	dfaBuilder := NewDFABuilder()
	dfaInfo1, _ := dfaBuilder.Build(cfg1)
	dfaInfo2, _ := dfaBuilder.Build(cfg2)

	// Extract DFA features
	dfaFeatures1 := ExtractDFAFeatures(dfaInfo1)
	dfaFeatures2 := ExtractDFAFeatures(dfaInfo2)

	// Compare DFA features
	dfaSimilarity := s.compareDFAFeatures(dfaFeatures1, dfaFeatures2)

	// Combine CFG and DFA similarities with configured weights
	return s.cfgFeatureWeight*cfgSimilarity + s.dfaFeatureWeight*dfaSimilarity
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

// compareCFGFeatures compares two CFG feature sets and returns a similarity score.
//
// Weight rationale (based on clone detection research and empirical observations):
//   - Cyclomatic complexity (0.25): Primary indicator of control flow complexity.
//     McCabe's metric directly measures decision complexity, making it the most
//     reliable single indicator of semantic equivalence.
//   - Edge type distribution (0.25): Captures the pattern of control flow (conditionals,
//     loops, exceptions). Similar distributions suggest similar behavioral patterns.
//   - Block count (0.20): Basic structural size indicator. Similar block counts
//     suggest similar decomposition of logic.
//   - Edge count (0.15): Complements block count; together they describe graph density.
//   - Branching factor (0.10): Average successors per block; indicates complexity
//     of individual decisions.
//   - Loop/conditional structure (0.05): Fine-grained structure; lower weight as
//     it's partially captured by edge type distribution.
//
// These weights can be adjusted via configuration in future versions if needed.
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

	// 1. Block count similarity (weight: 0.20)
	blockSim := s.computeCountSimilarity(f1.BlockCount, f2.BlockCount)
	weights = append(weights, 0.20)
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

	// 5. Branching factor similarity (weight: 0.10)
	branchSim := s.computeFloatSimilarity(f1.BranchingFactor, f2.BranchingFactor)
	weights = append(weights, 0.10)
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

// compareDFAFeatures compares two DFA feature sets and returns a similarity score.
//
// Weight rationale (based on data flow analysis research):
//   - Pair count (0.25): Total def-use pairs indicate data flow complexity
//   - Chain length (0.20): Uses per definition shows variable reuse patterns
//   - Cross-block pairs (0.20): Data dependencies across control flow structure
//   - Def kind distribution (0.20): Coding patterns in variable definitions
//   - Use kind distribution (0.15): Variable access patterns
func (s *SemanticSimilarityAnalyzer) compareDFAFeatures(f1, f2 *DFAFeatures) float64 {
	if f1 == nil || f2 == nil {
		return 0.0
	}

	// Handle edge cases
	if f1.TotalDefs == 0 && f2.TotalDefs == 0 {
		return 1.0 // Both empty
	}
	if f1.TotalDefs == 0 || f2.TotalDefs == 0 {
		return 0.0 // One empty
	}

	// Calculate individual similarity components
	var weights []float64
	var similarities []float64

	// 1. Total pairs similarity (weight from constants)
	pairSim := s.computeCountSimilarity(f1.TotalPairs, f2.TotalPairs)
	weights = append(weights, domain.DefaultDFAPairCountWeight)
	similarities = append(similarities, pairSim)

	// 2. Average chain length similarity
	chainSim := s.computeFloatSimilarity(f1.AvgChainLength, f2.AvgChainLength)
	weights = append(weights, domain.DefaultDFAChainLengthWeight)
	similarities = append(similarities, chainSim)

	// 3. Cross-block pairs ratio similarity
	crossBlockRatio1 := 0.0
	crossBlockRatio2 := 0.0
	if f1.TotalPairs > 0 {
		crossBlockRatio1 = float64(f1.CrossBlockPairs) / float64(f1.TotalPairs)
	}
	if f2.TotalPairs > 0 {
		crossBlockRatio2 = float64(f2.CrossBlockPairs) / float64(f2.TotalPairs)
	}
	crossBlockSim := s.computeFloatSimilarity(crossBlockRatio1, crossBlockRatio2)
	weights = append(weights, domain.DefaultDFACrossBlockWeight)
	similarities = append(similarities, crossBlockSim)

	// 4. Definition kind distribution similarity
	defKindSim := s.compareDefUseKindDistributions(f1.DefKindCounts, f2.DefKindCounts)
	weights = append(weights, domain.DefaultDFADefKindWeight)
	similarities = append(similarities, defKindSim)

	// 5. Use kind distribution similarity
	useKindSim := s.compareDefUseKindDistributions(f1.UseKindCounts, f2.UseKindCounts)
	weights = append(weights, domain.DefaultDFAUseKindWeight)
	similarities = append(similarities, useKindSim)

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

// compareDefUseKindDistributions compares two DefUseKind distributions using cosine similarity
func (s *SemanticSimilarityAnalyzer) compareDefUseKindDistributions(dist1, dist2 map[DefUseKind]int) float64 {
	if len(dist1) == 0 && len(dist2) == 0 {
		return 1.0
	}
	if len(dist1) == 0 || len(dist2) == 0 {
		return 0.0
	}

	// Get all kinds
	allKinds := make(map[DefUseKind]bool)
	for k := range dist1 {
		allKinds[k] = true
	}
	for k := range dist2 {
		allKinds[k] = true
	}

	// Compute cosine similarity
	var dotProduct float64
	var norm1 float64
	var norm2 float64

	for k := range allKinds {
		v1 := float64(dist1[k])
		v2 := float64(dist2[k])

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

// BuildDFA builds DFA info from a CFG (exposed for testing)
func (s *SemanticSimilarityAnalyzer) BuildDFA(cfg *CFG) (*DFAInfo, error) {
	builder := NewDFABuilder()
	return builder.Build(cfg)
}

// ExtractDFAFeatures extracts DFA features from DFA info (exposed for testing)
func (s *SemanticSimilarityAnalyzer) ExtractDFAFeaturesFromInfo(info *DFAInfo) *DFAFeatures {
	return ExtractDFAFeatures(info)
}

// IsDFAEnabled returns whether DFA analysis is enabled
func (s *SemanticSimilarityAnalyzer) IsDFAEnabled() bool {
	return s.enableDFA
}
