package analyzer

import (
	"math"
	"sort"
)

// APTEDAnalyzer implements the APTED (All Path Tree Edit Distance) algorithm
// Based on Pawlik & Augsten's optimal O(n² log n) algorithm
type APTEDAnalyzer struct {
	costModel CostModel
	cache     map[string]float64 // Memoization cache for subproblems
}

// NewAPTEDAnalyzer creates a new APTED analyzer with the given cost model
func NewAPTEDAnalyzer(costModel CostModel) *APTEDAnalyzer {
	return &APTEDAnalyzer{
		costModel: costModel,
		cache:     make(map[string]float64),
	}
}

// ComputeDistance computes the tree edit distance between two trees
func (a *APTEDAnalyzer) ComputeDistance(tree1, tree2 *TreeNode) float64 {
	// Handle edge cases
	if tree1 == nil && tree2 == nil {
		return 0.0
	}
	if tree1 == nil {
		return a.computeInsertCost(tree2)
	}
	if tree2 == nil {
		return a.computeDeleteCost(tree1)
	}

	// Use optimized version for large trees
	size1, size2 := tree1.Size(), tree2.Size()
	if size1 > 500 || size2 > 500 {
		return a.computeDistanceOptimized(tree1, tree2)
	}

	// Clear cache for new computation
	a.cache = make(map[string]float64)

	// Prepare both trees for APTED
	keyRoots1 := PrepareTreeForAPTED(tree1)
	keyRoots2 := PrepareTreeForAPTED(tree2)

	// Sort key roots in descending order
	sort.Sort(sort.Reverse(sort.IntSlice(keyRoots1)))
	sort.Sort(sort.Reverse(sort.IntSlice(keyRoots2)))

	// Compute distance using APTED algorithm
	return a.apted(tree1, tree2, keyRoots1, keyRoots2)
}

// computeDistanceOptimized uses an optimized version with early termination and pruning
func (a *APTEDAnalyzer) computeDistanceOptimized(tree1, tree2 *TreeNode) float64 {
	// Early termination based on size difference
	size1, size2 := tree1.Size(), tree2.Size()
	sizeDiff := math.Abs(float64(size1 - size2))
	
	// If size difference is too large, return early upper bound
	maxDistance := math.Max(float64(size1), float64(size2))
	if sizeDiff > maxDistance * 0.8 {
		return sizeDiff // Conservative estimate
	}

	// Use a simplified dynamic programming approach for very large trees
	if size1 > 2000 || size2 > 2000 {
		return a.computeApproximateDistance(tree1, tree2)
	}

	// Clear cache and use standard algorithm with optimizations
	a.cache = make(map[string]float64)
	
	// Prepare both trees for APTED
	keyRoots1 := PrepareTreeForAPTED(tree1)
	keyRoots2 := PrepareTreeForAPTED(tree2)

	// Limit key roots to reduce computation
	if len(keyRoots1) > 100 {
		keyRoots1 = keyRoots1[:100]
	}
	if len(keyRoots2) > 100 {
		keyRoots2 = keyRoots2[:100]
	}

	// Sort key roots in descending order
	sort.Sort(sort.Reverse(sort.IntSlice(keyRoots1)))
	sort.Sort(sort.Reverse(sort.IntSlice(keyRoots2)))

	// Compute distance using optimized APTED algorithm
	return a.aptedOptimized(tree1, tree2, keyRoots1, keyRoots2, maxDistance * 0.5)
}

// computeApproximateDistance computes an approximate distance for very large trees
func (a *APTEDAnalyzer) computeApproximateDistance(tree1, tree2 *TreeNode) float64 {
	// Use a simplified structural similarity measure
	depth1, depth2 := tree1.Height(), tree2.Height()
	size1, size2 := tree1.Size(), tree2.Size()
	
	// Compute structural differences
	depthDiff := math.Abs(float64(depth1 - depth2))
	sizeDiff := math.Abs(float64(size1 - size2))
	
	// Simple heuristic based on structural properties
	return (depthDiff * 2.0) + (sizeDiff * 0.5)
}

// aptedOptimized implements the optimized APTED algorithm with early termination
func (a *APTEDAnalyzer) aptedOptimized(tree1, tree2 *TreeNode, keyRoots1, keyRoots2 []int, maxDistance float64) float64 {
	// Get all nodes in post-order
	nodes1 := a.getPostOrderNodes(tree1)
	nodes2 := a.getPostOrderNodes(tree2)

	size1 := len(nodes1)
	size2 := len(nodes2)

	// Initialize distance matrix
	td := make([][]float64, size1+1)
	for i := range td {
		td[i] = make([]float64, size2+1)
	}

	// Main APTED loop with early termination
	for _, i := range keyRoots1 {
		for _, j := range keyRoots2 {
			a.computeForestDistanceOptimized(nodes1, nodes2, i, j, td, maxDistance)
			
			// Early termination if distance exceeds threshold
			if td[size1][size2] > maxDistance {
				return td[size1][size2]
			}
		}
	}

	return td[size1][size2]
}

// apted implements the main APTED algorithm
func (a *APTEDAnalyzer) apted(tree1, tree2 *TreeNode, keyRoots1, keyRoots2 []int) float64 {
	// Get all nodes in post-order
	nodes1 := a.getPostOrderNodes(tree1)
	nodes2 := a.getPostOrderNodes(tree2)

	size1 := len(nodes1)
	size2 := len(nodes2)

	// Initialize distance matrix
	td := make([][]float64, size1+1)
	for i := range td {
		td[i] = make([]float64, size2+1)
	}

	// Main APTED loop
	for _, i := range keyRoots1 {
		for _, j := range keyRoots2 {
			a.computeForestDistance(nodes1, nodes2, i, j, td)
		}
	}

	return td[size1][size2]
}

// computeForestDistanceOptimized computes the distance between two forests with early termination
func (a *APTEDAnalyzer) computeForestDistanceOptimized(nodes1, nodes2 []*TreeNode, i, j int, td [][]float64, maxDistance float64) {
	// Get left-most leaves for the subtrees
	lml_i := nodes1[i].LeftMostLeaf
	lml_j := nodes2[j].LeftMostLeaf

	// Initialize forest distance matrix
	fd := make([][]float64, i+2)
	for k := range fd {
		fd[k] = make([]float64, j+2)
	}

	// Base cases for forest distance
	for x := lml_i; x <= i; x++ {
		fd[x+1][lml_j] = fd[x][lml_j] + a.costModel.Delete(nodes1[x])
		// Early termination check
		if fd[x+1][lml_j] > maxDistance {
			return
		}
	}

	for y := lml_j; y <= j; y++ {
		fd[lml_i][y+1] = fd[lml_i][y] + a.costModel.Insert(nodes2[y])
		// Early termination check
		if fd[lml_i][y+1] > maxDistance {
			return
		}
	}

	// Main computation
	for x := lml_i; x <= i; x++ {
		for y := lml_j; y <= j; y++ {
			lml_x := nodes1[x].LeftMostLeaf
			lml_y := nodes2[y].LeftMostLeaf

			if lml_x == lml_i && lml_y == lml_j {
				// Both nodes are at the leftmost leaf of their respective forests
				deleteCost := fd[x][y+1] + a.costModel.Delete(nodes1[x])
				insertCost := fd[x+1][y] + a.costModel.Insert(nodes2[y])
				renameCost := fd[x][y] + a.costModel.Rename(nodes1[x], nodes2[y])

				fd[x+1][y+1] = math.Min(deleteCost, math.Min(insertCost, renameCost))
				td[x+1][y+1] = fd[x+1][y+1]
			} else {
				// At least one node is not at the leftmost leaf
				deleteCost := fd[x][y+1] + a.costModel.Delete(nodes1[x])
				insertCost := fd[x+1][y] + a.costModel.Insert(nodes2[y])

				var subtreeCost float64
				if lml_x == lml_i {
					subtreeCost = fd[lml_i][y] + td[x+1][lml_y]
				} else if lml_y == lml_j {
					subtreeCost = fd[x][lml_j] + td[lml_x][y+1]
				} else {
					subtreeCost = fd[lml_i][lml_j] + td[lml_x][lml_y]
				}

				fd[x+1][y+1] = math.Min(deleteCost, math.Min(insertCost, subtreeCost))
			}

			// Early termination check
			if fd[x+1][y+1] > maxDistance {
				return
			}
		}
	}
}

// computeForestDistance computes the distance between two forests
func (a *APTEDAnalyzer) computeForestDistance(nodes1, nodes2 []*TreeNode, i, j int, td [][]float64) {
	// Get left-most leaves for the subtrees
	lml_i := nodes1[i].LeftMostLeaf
	lml_j := nodes2[j].LeftMostLeaf

	// Initialize forest distance matrix
	fd := make([][]float64, i+2)
	for k := range fd {
		fd[k] = make([]float64, j+2)
	}

	// Base cases for forest distance
	for x := lml_i; x <= i; x++ {
		fd[x+1][lml_j] = fd[x][lml_j] + a.costModel.Delete(nodes1[x])
	}

	for y := lml_j; y <= j; y++ {
		fd[lml_i][y+1] = fd[lml_i][y] + a.costModel.Insert(nodes2[y])
	}

	// Main computation
	for x := lml_i; x <= i; x++ {
		for y := lml_j; y <= j; y++ {
			lml_x := nodes1[x].LeftMostLeaf
			lml_y := nodes2[y].LeftMostLeaf

			if lml_x == lml_i && lml_y == lml_j {
				// Both nodes are at the leftmost leaf of their respective forests
				deleteCost := fd[x][y+1] + a.costModel.Delete(nodes1[x])
				insertCost := fd[x+1][y] + a.costModel.Insert(nodes2[y])
				renameCost := fd[x][y] + a.costModel.Rename(nodes1[x], nodes2[y])

				fd[x+1][y+1] = math.Min(deleteCost, math.Min(insertCost, renameCost))
				td[x+1][y+1] = fd[x+1][y+1]
			} else {
				// At least one node is not at the leftmost leaf
				deleteCost := fd[x][y+1] + a.costModel.Delete(nodes1[x])
				insertCost := fd[x+1][y] + a.costModel.Insert(nodes2[y])

				var subtreeCost float64
				if lml_x == lml_i {
					subtreeCost = fd[lml_i][y] + td[x+1][lml_y]
				} else if lml_y == lml_j {
					subtreeCost = fd[x][lml_j] + td[lml_x][y+1]
				} else {
					subtreeCost = fd[lml_i][lml_j] + td[lml_x][lml_y]
				}

				fd[x+1][y+1] = math.Min(deleteCost, math.Min(insertCost, subtreeCost))
			}
		}
	}
}

// getPostOrderNodes returns all nodes in post-order traversal
func (a *APTEDAnalyzer) getPostOrderNodes(root *TreeNode) []*TreeNode {
	if root == nil {
		return []*TreeNode{}
	}

	var nodes []*TreeNode
	a.postOrderTraversal(root, &nodes)
	
	// Sort by post-order ID to ensure correct ordering
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].PostOrderID < nodes[j].PostOrderID
	})

	return nodes
}

// postOrderTraversal performs post-order traversal
func (a *APTEDAnalyzer) postOrderTraversal(node *TreeNode, nodes *[]*TreeNode) {
	if node == nil {
		return
	}

	// Visit children first
	for _, child := range node.Children {
		a.postOrderTraversal(child, nodes)
	}

	// Then visit this node
	*nodes = append(*nodes, node)
}

// computeInsertCost computes the cost of inserting an entire subtree
func (a *APTEDAnalyzer) computeInsertCost(root *TreeNode) float64 {
	if root == nil {
		return 0.0
	}

	cost := a.costModel.Insert(root)
	for _, child := range root.Children {
		cost += a.computeInsertCost(child)
	}

	return cost
}

// computeDeleteCost computes the cost of deleting an entire subtree
func (a *APTEDAnalyzer) computeDeleteCost(root *TreeNode) float64 {
	if root == nil {
		return 0.0
	}

	cost := a.costModel.Delete(root)
	for _, child := range root.Children {
		cost += a.computeDeleteCost(child)
	}

	return cost
}

// ComputeSimilarity computes similarity score between two trees (0.0 to 1.0)
func (a *APTEDAnalyzer) ComputeSimilarity(tree1, tree2 *TreeNode) float64 {
	distance := a.ComputeDistance(tree1, tree2)
	
	// Normalize by the maximum possible distance
	maxSize := float64(math.Max(float64(tree1.Size()), float64(tree2.Size())))
	if maxSize == 0 {
		return 1.0
	}

	return 1.0 - (distance / maxSize)
}

// TreeEditResult holds the result of tree edit distance computation
type TreeEditResult struct {
	Distance   float64
	Similarity float64
	Tree1Size  int
	Tree2Size  int
	Operations int // Estimated number of edit operations
}

// ComputeDetailedDistance computes detailed tree edit distance information
func (a *APTEDAnalyzer) ComputeDetailedDistance(tree1, tree2 *TreeNode) *TreeEditResult {
	distance := a.ComputeDistance(tree1, tree2)
	similarity := a.ComputeSimilarity(tree1, tree2)

	var size1, size2 int
	if tree1 != nil {
		size1 = tree1.Size()
	}
	if tree2 != nil {
		size2 = tree2.Size()
	}

	return &TreeEditResult{
		Distance:   distance,
		Similarity: similarity,
		Tree1Size:  size1,
		Tree2Size:  size2,
		Operations: int(distance), // Approximate number of operations
	}
}

// OptimizedAPTEDAnalyzer extends APTEDAnalyzer with performance optimizations
type OptimizedAPTEDAnalyzer struct {
	*APTEDAnalyzer
	maxDistance     float64 // Early termination threshold
	enableEarlyStop bool    // Whether to enable early stopping
}

// NewOptimizedAPTEDAnalyzer creates an optimized APTED analyzer
func NewOptimizedAPTEDAnalyzer(costModel CostModel, maxDistance float64) *OptimizedAPTEDAnalyzer {
	return &OptimizedAPTEDAnalyzer{
		APTEDAnalyzer:   NewAPTEDAnalyzer(costModel),
		maxDistance:     maxDistance,
		enableEarlyStop: maxDistance > 0,
	}
}

// ComputeDistance computes tree edit distance with early stopping optimization
func (a *OptimizedAPTEDAnalyzer) ComputeDistance(tree1, tree2 *TreeNode) float64 {
	// Quick size-based early termination
	if a.enableEarlyStop {
		sizeDiff := math.Abs(float64(tree1.Size() - tree2.Size()))
		if sizeDiff > a.maxDistance {
			return a.maxDistance + 1.0 // Indicates distance exceeds threshold
		}
	}

	// Use parent implementation
	distance := a.APTEDAnalyzer.ComputeDistance(tree1, tree2)

	// Early termination check
	if a.enableEarlyStop && distance > a.maxDistance {
		return a.maxDistance + 1.0
	}

	return distance
}

// BatchComputeDistances computes distances between multiple tree pairs efficiently
func (a *APTEDAnalyzer) BatchComputeDistances(pairs [][2]*TreeNode) []float64 {
	distances := make([]float64, len(pairs))

	for i, pair := range pairs {
		distances[i] = a.ComputeDistance(pair[0], pair[1])
	}

	return distances
}

// ClusterResult represents the result of tree clustering
type ClusterResult struct {
	Groups     [][]int     // Groups of tree indices that are similar
	Distances  [][]float64 // Distance matrix between all trees
	Threshold  float64     // Similarity threshold used
}

// ClusterSimilarTrees clusters trees based on similarity threshold
func (a *APTEDAnalyzer) ClusterSimilarTrees(trees []*TreeNode, similarityThreshold float64) *ClusterResult {
	// Input validation
	if len(trees) == 0 {
		return &ClusterResult{
			Groups:    [][]int{},
			Distances: [][]float64{},
			Threshold: similarityThreshold,
		}
	}

	// Filter out nil trees
	validTrees := make([]*TreeNode, 0, len(trees))
	originalIndices := make([]int, 0, len(trees))
	for i, tree := range trees {
		if tree != nil {
			validTrees = append(validTrees, tree)
			originalIndices = append(originalIndices, i)
		}
	}

	if len(validTrees) == 0 {
		return &ClusterResult{
			Groups:    [][]int{},
			Distances: [][]float64{},
			Threshold: similarityThreshold,
		}
	}

	if len(validTrees) == 1 {
		// Single tree case
		return &ClusterResult{
			Groups:    [][]int{{originalIndices[0]}},
			Distances: [][]float64{{0.0}},
			Threshold: similarityThreshold,
		}
	}

	n := len(validTrees)
	distances := make([][]float64, n)
	
	// Initialize distance matrix with proper allocation
	for i := 0; i < n; i++ {
		distances[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if i == j {
				distances[i][j] = 0.0
			} else {
				// Initialize with a high value
				distances[i][j] = math.Inf(1)
			}
		}
	}

	// Compute distance matrix with robust error handling
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if validTrees[i] != nil && validTrees[j] != nil {
				dist := a.ComputeDistance(validTrees[i], validTrees[j])
				distances[i][j] = dist
				distances[j][i] = dist
			}
		}
	}

	// Simple clustering based on threshold
	visited := make([]bool, n)
	var groups [][]int

	for i := 0; i < n; i++ {
		if visited[i] {
			continue
		}

		// Start a new cluster with original indices
		cluster := []int{originalIndices[i]}
		visited[i] = true

		// Add similar trees to this cluster
		for j := i + 1; j < n; j++ {
			if !visited[j] && distances[i][j] != math.Inf(1) {
				maxSize := math.Max(float64(validTrees[i].Size()), float64(validTrees[j].Size()))
				if maxSize > 0 {
					similarity := 1.0 - distances[i][j]/maxSize
					if similarity >= similarityThreshold {
						cluster = append(cluster, originalIndices[j])
						visited[j] = true
					}
				}
			}
		}

		groups = append(groups, cluster)
	}

	return &ClusterResult{
		Groups:    groups,
		Distances: distances,
		Threshold: similarityThreshold,
	}
}