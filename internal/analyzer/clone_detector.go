package analyzer

import (
	"fmt"
	"math"
	"runtime"
	"sort"

	"github.com/pyqol/pyqol/internal/config"
	"github.com/pyqol/pyqol/internal/constants"
	"github.com/pyqol/pyqol/internal/parser"
)

// CloneType represents different types of code clones
type CloneType int

const (
	// Type1Clone - Identical code fragments (except whitespace and comments)
	Type1Clone CloneType = iota + 1
	// Type2Clone - Syntactically identical but with different identifiers/literals
	Type2Clone
	// Type3Clone - Syntactically similar with small modifications
	Type3Clone
	// Type4Clone - Functionally similar but syntactically different
	Type4Clone
)

// String returns string representation of CloneType
func (ct CloneType) String() string {
	switch ct {
	case Type1Clone:
		return "Type-1 (Identical)"
	case Type2Clone:
		return "Type-2 (Renamed)"
	case Type3Clone:
		return "Type-3 (Near-Miss)"
	case Type4Clone:
		return "Type-4 (Semantic)"
	default:
		return "Unknown"
	}
}

// CodeLocation represents a location in source code
type CodeLocation struct {
	FilePath  string
	StartLine int
	EndLine   int
	StartCol  int
	EndCol    int
}

// String returns string representation of CodeLocation
func (cl *CodeLocation) String() string {
	return fmt.Sprintf("%s:%d:%d-%d:%d", cl.FilePath, cl.StartLine, cl.StartCol, cl.EndLine, cl.EndCol)
}

// CodeFragment represents a fragment of code
type CodeFragment struct {
	Location   *CodeLocation
	ASTNode    *parser.Node
	TreeNode   *TreeNode
	Content    string // Original source code content
	Hash       string // Hash for quick comparison
	Size       int    // Number of AST nodes
	LineCount  int    // Number of source lines
	Complexity int    // Cyclomatic complexity (if applicable)
}

// NewCodeFragment creates a new code fragment
func NewCodeFragment(location *CodeLocation, astNode *parser.Node, content string) *CodeFragment {
	return &CodeFragment{
		Location:  location,
		ASTNode:   astNode,
		Content:   content,
		Size:      calculateASTSize(astNode),
		LineCount: location.EndLine - location.StartLine + 1,
	}
}

// calculateASTSize calculates the number of nodes in an AST
func calculateASTSize(node *parser.Node) int {
	if node == nil {
		return 0
	}

	size := 1
	for _, child := range node.Children {
		size += calculateASTSize(child)
	}
	for _, bodyNode := range node.Body {
		size += calculateASTSize(bodyNode)
	}
	for _, orelseNode := range node.Orelse {
		size += calculateASTSize(orelseNode)
	}

	return size
}

// ClonePair represents a pair of similar code fragments
type ClonePair struct {
	Fragment1  *CodeFragment
	Fragment2  *CodeFragment
	Similarity float64   // Similarity score (0.0 to 1.0)
	Distance   float64   // Edit distance
	CloneType  CloneType // Type of clone detected
	Confidence float64   // Confidence in the detection (0.0 to 1.0)
}

// String returns string representation of ClonePair
func (cp *ClonePair) String() string {
	return fmt.Sprintf("%s clone: %s <-> %s (similarity: %.3f)",
		cp.CloneType.String(),
		cp.Fragment1.Location.String(),
		cp.Fragment2.Location.String(),
		cp.Similarity)
}

// CloneGroup represents a group of similar code fragments
type CloneGroup struct {
	ID         int             // Unique identifier for this group
	Fragments  []*CodeFragment // All fragments in this group
	CloneType  CloneType       // Primary type of clones in this group
	Similarity float64         // Average similarity within the group
	Size       int             // Number of fragments
}

// NewCloneGroup creates a new clone group
func NewCloneGroup(id int) *CloneGroup {
	return &CloneGroup{
		ID:        id,
		Fragments: []*CodeFragment{},
	}
}

// AddFragment adds a fragment to the clone group
func (cg *CloneGroup) AddFragment(fragment *CodeFragment) {
	cg.Fragments = append(cg.Fragments, fragment)
	cg.Size = len(cg.Fragments)
}

// CloneDetectorConfig holds configuration for clone detection
// DEPRECATED: Use config.CloneConfig with adapter functions instead
type CloneDetectorConfig struct {
	// Minimum number of lines for a code fragment to be considered
	MinLines int

	// Minimum number of AST nodes for a code fragment
	MinNodes int

	// Similarity thresholds for different clone types
	Type1Threshold float64 // Usually > constants.DefaultType1CloneThreshold
	Type2Threshold float64 // Usually > constants.DefaultType2CloneThreshold
	Type3Threshold float64 // Usually > constants.DefaultType3CloneThreshold
	Type4Threshold float64 // Usually > constants.DefaultType4CloneThreshold

	// Maximum edit distance allowed
	MaxEditDistance float64

	// Whether to ignore differences in literals
	IgnoreLiterals bool

	// Whether to ignore differences in identifiers
	IgnoreIdentifiers bool

	// Cost model to use for APTED
	CostModelType string // "default", "python", "weighted"
}

// DefaultCloneDetectorConfig returns default configuration
func DefaultCloneDetectorConfig() *CloneDetectorConfig {
	return &CloneDetectorConfig{
		MinLines:          5,
		MinNodes:          10,
		Type1Threshold:    constants.DefaultType1CloneThreshold,
		Type2Threshold:    constants.DefaultType2CloneThreshold,
		Type3Threshold:    constants.DefaultType3CloneThreshold,
		Type4Threshold:    constants.DefaultType4CloneThreshold,
		MaxEditDistance:   50.0,
		IgnoreLiterals:    false,
		IgnoreIdentifiers: false,
		CostModelType:     "python",
	}
}

// CloneDetector detects code clones using APTED algorithm
type CloneDetector struct {
	config      *CloneDetectorConfig
	analyzer    *APTEDAnalyzer
	converter   *TreeConverter
	fragments   []*CodeFragment
	clonePairs  []*ClonePair
	cloneGroups []*CloneGroup
}

// NewCloneDetectorFromConfig creates a new clone detector from unified config
func NewCloneDetectorFromConfig(cloneConfig *config.CloneConfig) *CloneDetector {
	// Convert unified config to legacy config directly
	legacyConfig := &CloneDetectorConfig{
		MinLines:          cloneConfig.Analysis.MinLines,
		MinNodes:          cloneConfig.Analysis.MinNodes,
		Type1Threshold:    cloneConfig.Thresholds.Type1Threshold,
		Type2Threshold:    cloneConfig.Thresholds.Type2Threshold,
		Type3Threshold:    cloneConfig.Thresholds.Type3Threshold,
		Type4Threshold:    cloneConfig.Thresholds.Type4Threshold,
		MaxEditDistance:   cloneConfig.Analysis.MaxEditDistance,
		IgnoreLiterals:    cloneConfig.Analysis.IgnoreLiterals,
		IgnoreIdentifiers: cloneConfig.Analysis.IgnoreIdentifiers,
		CostModelType:     cloneConfig.Analysis.CostModelType,
	}
	return NewCloneDetector(legacyConfig)
}

// NewCloneDetector creates a new clone detector with the given configuration
// DEPRECATED: Use NewCloneDetectorFromConfig with unified config.CloneConfig instead
func NewCloneDetector(config *CloneDetectorConfig) *CloneDetector {
	// Create appropriate cost model based on configuration
	var costModel CostModel
	switch config.CostModelType {
	case "default":
		costModel = NewDefaultCostModel()
	case "python":
		costModel = NewPythonCostModelWithConfig(config.IgnoreLiterals, config.IgnoreIdentifiers)
	case "weighted":
		baseCostModel := NewPythonCostModelWithConfig(config.IgnoreLiterals, config.IgnoreIdentifiers)
		costModel = NewWeightedCostModel(1.0, 1.0, 0.8, baseCostModel)
	default:
		costModel = NewPythonCostModel()
	}

	analyzer := NewAPTEDAnalyzer(costModel)

	return &CloneDetector{
		config:      config,
		analyzer:    analyzer,
		converter:   NewTreeConverter(),
		fragments:   []*CodeFragment{},
		clonePairs:  []*ClonePair{},
		cloneGroups: []*CloneGroup{},
	}
}

// ExtractFragments extracts code fragments from AST nodes
func (cd *CloneDetector) ExtractFragments(astNodes []*parser.Node, filePath string) []*CodeFragment {
	var fragments []*CodeFragment

	for _, node := range astNodes {
		cd.extractFragmentsRecursive(node, filePath, &fragments)
	}

	return fragments
}

// extractFragmentsRecursive recursively extracts fragments from AST
func (cd *CloneDetector) extractFragmentsRecursive(node *parser.Node, filePath string, fragments *[]*CodeFragment) {
	if node == nil {
		return
	}

	// Check if this node should be considered as a fragment
	if cd.isFragmentCandidate(node) {
		location := &CodeLocation{
			FilePath:  filePath,
			StartLine: node.Location.StartLine,
			EndLine:   node.Location.EndLine,
			StartCol:  node.Location.StartCol,
			EndCol:    node.Location.EndCol,
		}

		fragment := NewCodeFragment(location, node, "")

		// Filter fragments based on configuration
		if cd.shouldIncludeFragment(fragment) {
			*fragments = append(*fragments, fragment)
		}
	}

	// Recursively process children
	for _, child := range node.Children {
		cd.extractFragmentsRecursive(child, filePath, fragments)
	}

	for _, bodyNode := range node.Body {
		cd.extractFragmentsRecursive(bodyNode, filePath, fragments)
	}

	for _, orelseNode := range node.Orelse {
		cd.extractFragmentsRecursive(orelseNode, filePath, fragments)
	}
}

// isFragmentCandidate checks if a node should be considered as a fragment candidate
func (cd *CloneDetector) isFragmentCandidate(node *parser.Node) bool {
	// Consider functions, classes, and compound statements as fragment candidates
	candidateTypes := []parser.NodeType{
		parser.NodeFunctionDef,
		parser.NodeAsyncFunctionDef,
		parser.NodeClassDef,
		parser.NodeFor,
		parser.NodeAsyncFor,
		parser.NodeWhile,
		parser.NodeIf,
		parser.NodeTry,
		parser.NodeWith,
		parser.NodeAsyncWith,
	}

	for _, candidateType := range candidateTypes {
		if node.Type == candidateType {
			return true
		}
	}

	return false
}

// shouldIncludeFragment determines if a fragment should be included in analysis
func (cd *CloneDetector) shouldIncludeFragment(fragment *CodeFragment) bool {
	// Check minimum size requirements
	if fragment.Size < cd.config.MinNodes {
		return false
	}

	if fragment.LineCount < cd.config.MinLines {
		return false
	}

	return true
}

// DetectClones detects clones in the given code fragments
func (cd *CloneDetector) DetectClones(fragments []*CodeFragment) ([]*ClonePair, []*CloneGroup) {
	cd.fragments = fragments
	cd.clonePairs = []*ClonePair{}
	cd.cloneGroups = []*CloneGroup{}

	// Convert AST fragments to tree nodes
	cd.prepareFragments()

	// Detect clone pairs
	cd.detectClonePairs()

	// Group related clones
	cd.groupClones()

	return cd.clonePairs, cd.cloneGroups
}

// prepareFragments converts AST fragments to tree nodes
func (cd *CloneDetector) prepareFragments() {
	for _, fragment := range cd.fragments {
		if fragment.ASTNode != nil {
			fragment.TreeNode = cd.converter.ConvertAST(fragment.ASTNode)
			if fragment.TreeNode != nil {
				PrepareTreeForAPTED(fragment.TreeNode)
			}
		}
	}
}

// detectClonePairs detects pairs of similar code fragments with memory management
func (cd *CloneDetector) detectClonePairs() {
	n := len(cd.fragments)

	// Memory management constants
	const (
		maxClonePairs = 10000             // Maximum pairs to keep in memory
		batchSize     = 1000              // Process fragments in batches
		memoryLimit   = 100 * 1024 * 1024 // 100MB memory limit
	)

	// Early return for small datasets
	if n <= 1 {
		return
	}

	// Estimate memory usage and use batching if needed
	estimatedPairs := (n * (n - 1)) / 2
	if estimatedPairs > maxClonePairs || n > batchSize {
		cd.detectClonePairsWithBatching(maxClonePairs, batchSize)
	} else {
		cd.detectClonePairsStandard()
	}

	// Sort and limit final results
	cd.limitAndSortClonePairs(maxClonePairs)
}

// detectClonePairsStandard uses the standard O(n²) approach for small datasets
func (cd *CloneDetector) detectClonePairsStandard() {
	n := len(cd.fragments)

	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			fragment1 := cd.fragments[i]
			fragment2 := cd.fragments[j]

			// Skip if both fragments are from the same location
			if cd.isSameLocation(fragment1.Location, fragment2.Location) {
				continue
			}

			// Compute similarity
			pair := cd.compareFragments(fragment1, fragment2)
			if pair != nil && cd.isSignificantClone(pair) {
				cd.clonePairs = append(cd.clonePairs, pair)
			}
		}
	}
}

// detectClonePairsWithBatching processes fragments in batches to manage memory
func (cd *CloneDetector) detectClonePairsWithBatching(maxPairs, batchSize int) {
	n := len(cd.fragments)

	// Priority queue to keep only the best pairs
	topPairs := make([]*ClonePair, 0, maxPairs)
	minSimilarity := cd.config.Type4Threshold // Use the lowest threshold as minimum

	// Process in batches to limit memory usage
	for batchStart := 0; batchStart < n; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > n {
			batchEnd = n
		}

		// Process current batch against all previous and current fragments
		for i := batchStart; i < batchEnd; i++ {
			// Compare with fragments in current batch
			for j := i + 1; j < batchEnd; j++ {
				if pair := cd.tryCreateClonePair(i, j, minSimilarity); pair != nil {
					topPairs = cd.addPairWithLimit(topPairs, pair, maxPairs)
					// Update minimum similarity threshold
					if len(topPairs) >= maxPairs {
						minSimilarity = topPairs[len(topPairs)-1].Similarity
					}
				}
			}

			// Compare with all previous fragments
			for j := 0; j < batchStart; j++ {
				if pair := cd.tryCreateClonePair(i, j, minSimilarity); pair != nil {
					topPairs = cd.addPairWithLimit(topPairs, pair, maxPairs)
					// Update minimum similarity threshold
					if len(topPairs) >= maxPairs {
						minSimilarity = topPairs[len(topPairs)-1].Similarity
					}
				}
			}
		}

		// Periodic garbage collection hint for large batches
		if batchStart%5000 == 0 {
			// Force garbage collection to prevent memory buildup
			runtime.GC()
		}
	}

	// Replace clone pairs with the best ones found
	cd.clonePairs = topPairs
}

// compareFragments compares two fragments and returns a clone pair if similar
func (cd *CloneDetector) compareFragments(fragment1, fragment2 *CodeFragment) *ClonePair {
	if fragment1.TreeNode == nil || fragment2.TreeNode == nil {
		return nil
	}

	// Compute edit distance and similarity
	distance := cd.analyzer.ComputeDistance(fragment1.TreeNode, fragment2.TreeNode)
	similarity := cd.analyzer.ComputeSimilarity(fragment1.TreeNode, fragment2.TreeNode)

	// Determine clone type based on similarity
	cloneType := cd.classifyCloneType(similarity, distance)
	if cloneType == 0 {
		return nil // Not a significant clone
	}

	// Calculate confidence based on fragment size and similarity
	confidence := cd.calculateConfidence(fragment1, fragment2, similarity)

	return &ClonePair{
		Fragment1:  fragment1,
		Fragment2:  fragment2,
		Similarity: similarity,
		Distance:   distance,
		CloneType:  cloneType,
		Confidence: confidence,
	}
}

// classifyCloneType classifies the type of clone based on similarity
func (cd *CloneDetector) classifyCloneType(similarity, distance float64) CloneType {
	if similarity >= cd.config.Type1Threshold {
		return Type1Clone
	} else if similarity >= cd.config.Type2Threshold {
		return Type2Clone
	} else if similarity >= cd.config.Type3Threshold {
		return Type3Clone
	} else if similarity >= cd.config.Type4Threshold {
		return Type4Clone
	}

	return 0 // Not a clone
}

// calculateConfidence calculates confidence in clone detection
func (cd *CloneDetector) calculateConfidence(fragment1, fragment2 *CodeFragment, similarity float64) float64 {
	// Base confidence on similarity
	confidence := similarity

	// Increase confidence for larger fragments
	avgSize := float64(fragment1.Size+fragment2.Size) / 2.0
	sizeBonus := math.Min(avgSize/100.0, 0.2) // Up to 20% bonus for large fragments
	confidence += sizeBonus

	// Increase confidence if both fragments have similar complexity
	if fragment1.Complexity > 0 && fragment2.Complexity > 0 {
		complexityRatio := float64(math.Min(float64(fragment1.Complexity), float64(fragment2.Complexity))) /
			float64(math.Max(float64(fragment1.Complexity), float64(fragment2.Complexity)))
		confidence += complexityRatio * 0.1 // Up to 10% bonus for similar complexity
	}

	// Cap confidence at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// isSignificantClone determines if a clone pair is significant enough to report
func (cd *CloneDetector) isSignificantClone(pair *ClonePair) bool {
	// Check minimum similarity threshold
	if pair.Similarity < cd.config.Type4Threshold {
		return false
	}

	// Check maximum distance threshold
	if pair.Distance > cd.config.MaxEditDistance {
		return false
	}

	// Additional filtering based on fragment characteristics
	minSize := math.Min(float64(pair.Fragment1.Size), float64(pair.Fragment2.Size))
	return minSize >= float64(cd.config.MinNodes)
}

// groupClones groups related clone pairs into clone groups
func (cd *CloneDetector) groupClones() {
	// Simple clustering based on shared fragments
	fragmentToGroup := make(map[*CodeFragment]*CloneGroup)
	groupID := 0

	for _, pair := range cd.clonePairs {
		group1 := fragmentToGroup[pair.Fragment1]
		group2 := fragmentToGroup[pair.Fragment2]

		if group1 == nil && group2 == nil {
			// Create new group
			newGroup := NewCloneGroup(groupID)
			groupID++
			newGroup.AddFragment(pair.Fragment1)
			newGroup.AddFragment(pair.Fragment2)
			newGroup.CloneType = pair.CloneType
			newGroup.Similarity = pair.Similarity

			fragmentToGroup[pair.Fragment1] = newGroup
			fragmentToGroup[pair.Fragment2] = newGroup
			cd.cloneGroups = append(cd.cloneGroups, newGroup)
		} else if group1 != nil && group2 == nil {
			// Add fragment2 to existing group1
			group1.AddFragment(pair.Fragment2)
			fragmentToGroup[pair.Fragment2] = group1
		} else if group1 == nil && group2 != nil {
			// Add fragment1 to existing group2
			group2.AddFragment(pair.Fragment1)
			fragmentToGroup[pair.Fragment1] = group2
		} else if group1 != group2 {
			// Merge two groups (simple approach: add all fragments from group2 to group1)
			for _, fragment := range group2.Fragments {
				if fragmentToGroup[fragment] == group2 {
					fragmentToGroup[fragment] = group1
				}
			}
			group1.Fragments = append(group1.Fragments, group2.Fragments...)
			group1.Size = len(group1.Fragments)

			// Remove group2 from clone groups
			for i, group := range cd.cloneGroups {
				if group == group2 {
					cd.cloneGroups = append(cd.cloneGroups[:i], cd.cloneGroups[i+1:]...)
					break
				}
			}
		}
	}

	// Calculate average similarity for each group
	for _, group := range cd.cloneGroups {
		cd.calculateGroupSimilarity(group)
	}
}

// calculateGroupSimilarity calculates average similarity within a clone group
func (cd *CloneDetector) calculateGroupSimilarity(group *CloneGroup) {
	if group.Size < 2 {
		group.Similarity = 1.0
		return
	}

	totalSimilarity := 0.0
	pairCount := 0

	// Calculate all pairwise similarities
	for i := 0; i < group.Size; i++ {
		for j := i + 1; j < group.Size; j++ {
			fragment1 := group.Fragments[i]
			fragment2 := group.Fragments[j]

			if fragment1.TreeNode != nil && fragment2.TreeNode != nil {
				similarity := cd.analyzer.ComputeSimilarity(fragment1.TreeNode, fragment2.TreeNode)
				totalSimilarity += similarity
				pairCount++
			}
		}
	}

	if pairCount > 0 {
		group.Similarity = totalSimilarity / float64(pairCount)
	} else {
		group.Similarity = 0.0
	}
}

// isSameLocation checks if two locations refer to the same code
func (cd *CloneDetector) isSameLocation(loc1, loc2 *CodeLocation) bool {
	return loc1.FilePath == loc2.FilePath &&
		loc1.StartLine == loc2.StartLine &&
		loc1.EndLine == loc2.EndLine
}

// GetStatistics returns clone detection statistics
func (cd *CloneDetector) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	stats["total_fragments"] = len(cd.fragments)
	stats["total_clone_pairs"] = len(cd.clonePairs)
	stats["total_clone_groups"] = len(cd.cloneGroups)

	// Count by clone type
	typeCounts := make(map[string]int)
	for _, pair := range cd.clonePairs {
		typeCounts[pair.CloneType.String()]++
	}
	stats["clone_types"] = typeCounts

	// Average similarity
	if len(cd.clonePairs) > 0 {
		totalSim := 0.0
		for _, pair := range cd.clonePairs {
			totalSim += pair.Similarity
		}
		stats["average_similarity"] = totalSim / float64(len(cd.clonePairs))
	}

	return stats
}

// tryCreateClonePair attempts to create a clone pair if it meets similarity threshold
func (cd *CloneDetector) tryCreateClonePair(i, j int, minSimilarity float64) *ClonePair {
	fragment1 := cd.fragments[i]
	fragment2 := cd.fragments[j]

	// Skip if both fragments are from the same location
	if cd.isSameLocation(fragment1.Location, fragment2.Location) {
		return nil
	}

	// Quick similarity check before expensive computation
	if fragment1.TreeNode == nil || fragment2.TreeNode == nil {
		return nil
	}

	// Early similarity estimation (fast)
	sizeDiff := float64(abs(fragment1.Size - fragment2.Size))
	maxSize := float64(max(fragment1.Size, fragment2.Size))
	if maxSize > 0 && (sizeDiff/maxSize) > (1.0-minSimilarity) {
		// Size difference too large for minimum similarity
		return nil
	}

	// Full similarity computation
	pair := cd.compareFragments(fragment1, fragment2)
	if pair != nil && cd.isSignificantClone(pair) && pair.Similarity >= minSimilarity {
		return pair
	}
	return nil
}

// addPairWithLimit adds a pair to the collection while maintaining size limit
func (cd *CloneDetector) addPairWithLimit(pairs []*ClonePair, newPair *ClonePair, maxPairs int) []*ClonePair {
	// If under limit, just add
	if len(pairs) < maxPairs {
		pairs = append(pairs, newPair)
		// Keep sorted by similarity (descending)
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Similarity > pairs[j].Similarity
		})
		return pairs
	}

	// If at limit, check if new pair is better than the worst
	if newPair.Similarity > pairs[len(pairs)-1].Similarity {
		// Replace worst pair
		pairs[len(pairs)-1] = newPair
		// Re-sort to maintain order
		sort.Slice(pairs, func(i, j int) bool {
			return pairs[i].Similarity > pairs[j].Similarity
		})
	}

	return pairs
}

// limitAndSortClonePairs ensures final results are sorted and limited
func (cd *CloneDetector) limitAndSortClonePairs(maxPairs int) {
	// Sort clone pairs by similarity (descending)
	sort.Slice(cd.clonePairs, func(i, j int) bool {
		return cd.clonePairs[i].Similarity > cd.clonePairs[j].Similarity
	})

	// Limit the number of pairs to prevent memory issues
	if len(cd.clonePairs) > maxPairs {
		cd.clonePairs = cd.clonePairs[:maxPairs]
	}
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
