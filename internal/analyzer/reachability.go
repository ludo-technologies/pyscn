package analyzer

import (
	"time"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

// ReachabilityResult contains the results of reachability analysis
type ReachabilityResult struct {
	// ReachableBlocks contains blocks that can be reached from entry
	ReachableBlocks map[string]*BasicBlock

	// UnreachableBlocks contains blocks that cannot be reached from entry
	UnreachableBlocks map[string]*BasicBlock

	// TotalBlocks is the total number of blocks analyzed
	TotalBlocks int

	// ReachableCount is the number of reachable blocks
	ReachableCount int

	// UnreachableCount is the number of unreachable blocks
	UnreachableCount int

	// AnalysisTime is the time taken to perform the analysis
	AnalysisTime time.Duration
}

// ReachabilityAnalyzer performs reachability analysis on CFGs
type ReachabilityAnalyzer struct {
	cfg                    *CFG
	allPathsReturnCache    map[string]bool // cached results of allSuccessorsReturn
	allPathsReturnComputed map[string]bool // tracks which blocks have been computed
}

// NewReachabilityAnalyzer creates a new reachability analyzer for the given CFG
func NewReachabilityAnalyzer(cfg *CFG) *ReachabilityAnalyzer {
	return &ReachabilityAnalyzer{
		cfg:                    cfg,
		allPathsReturnCache:    make(map[string]bool),
		allPathsReturnComputed: make(map[string]bool),
	}
}

// AnalyzeReachability performs reachability analysis starting from the entry block
func (ra *ReachabilityAnalyzer) AnalyzeReachability() *ReachabilityResult {
	startTime := time.Now()

	result := &ReachabilityResult{
		ReachableBlocks:   make(map[string]*BasicBlock),
		UnreachableBlocks: make(map[string]*BasicBlock),
		TotalBlocks:       0,
	}

	// Handle nil CFG or empty CFG
	if ra.cfg == nil || ra.cfg.Entry == nil || ra.cfg.Blocks == nil {
		result.AnalysisTime = time.Since(startTime)
		return result
	}

	result.TotalBlocks = len(ra.cfg.Blocks)

	// Handle empty blocks map
	if len(ra.cfg.Blocks) == 0 {
		result.AnalysisTime = time.Since(startTime)
		return result
	}

	// Perform enhanced reachability analysis that includes all-paths-return detection
	ra.performEnhancedReachabilityAnalysis(result)

	// Update counts
	result.ReachableCount = len(result.ReachableBlocks)
	result.UnreachableCount = len(result.UnreachableBlocks)
	result.AnalysisTime = time.Since(startTime)

	return result
}

// AnalyzeReachabilityFrom performs reachability analysis from a specific starting block
func (ra *ReachabilityAnalyzer) AnalyzeReachabilityFrom(startBlock *BasicBlock) *ReachabilityResult {
	startTime := time.Now()

	result := &ReachabilityResult{
		ReachableBlocks:   make(map[string]*BasicBlock),
		UnreachableBlocks: make(map[string]*BasicBlock),
		TotalBlocks:       0,
	}

	// Handle nil CFG or nil start block
	if ra.cfg == nil || ra.cfg.Blocks == nil || startBlock == nil {
		result.AnalysisTime = time.Since(startTime)
		return result
	}

	result.TotalBlocks = len(ra.cfg.Blocks)

	// Handle empty blocks
	if len(ra.cfg.Blocks) == 0 {
		result.AnalysisTime = time.Since(startTime)
		return result
	}

	// Perform DFS traversal from the start block
	visited := make(map[string]bool)
	ra.traverseFrom(startBlock, visited, result.ReachableBlocks)

	// Identify unreachable blocks
	for id, block := range ra.cfg.Blocks {
		if _, isReachable := result.ReachableBlocks[id]; !isReachable {
			result.UnreachableBlocks[id] = block
		}
	}

	// Update counts
	result.ReachableCount = len(result.ReachableBlocks)
	result.UnreachableCount = len(result.UnreachableBlocks)
	result.AnalysisTime = time.Since(startTime)

	return result
}

// traverseFrom performs DFS traversal from a given block
func (ra *ReachabilityAnalyzer) traverseFrom(block *BasicBlock, visited map[string]bool, reachable map[string]*BasicBlock) {
	if block == nil || visited[block.ID] {
		return
	}

	visited[block.ID] = true
	reachable[block.ID] = block

	// Visit all successors
	for _, edge := range block.Successors {
		ra.traverseFrom(edge.To, visited, reachable)
	}
}

// GetUnreachableBlocksWithStatements returns unreachable blocks that contain statements
func (result *ReachabilityResult) GetUnreachableBlocksWithStatements() map[string]*BasicBlock {
	blocksWithStatements := make(map[string]*BasicBlock)

	for id, block := range result.UnreachableBlocks {
		if !block.IsEmpty() {
			blocksWithStatements[id] = block
		}
	}

	return blocksWithStatements
}

// GetReachabilityRatio returns the ratio of reachable blocks to total blocks
func (result *ReachabilityResult) GetReachabilityRatio() float64 {
	if result.TotalBlocks == 0 {
		return 1.0
	}
	return float64(result.ReachableCount) / float64(result.TotalBlocks)
}

// HasUnreachableCode returns true if there are unreachable blocks with statements
func (result *ReachabilityResult) HasUnreachableCode() bool {
	for _, block := range result.UnreachableBlocks {
		if !block.IsEmpty() {
			return true
		}
	}
	return false
}

// reachabilityVisitor implements CFGVisitor for reachability analysis
type reachabilityVisitor struct {
	reachableBlocks map[string]*BasicBlock
}

// VisitBlock marks a block as reachable
func (rv *reachabilityVisitor) VisitBlock(block *BasicBlock) bool {
	if block != nil {
		rv.reachableBlocks[block.ID] = block
	}
	return true // Continue traversal
}

// VisitEdge continues traversal through edges
func (rv *reachabilityVisitor) VisitEdge(edge *Edge) bool {
	return true // Continue traversal
}

// performEnhancedReachabilityAnalysis performs reachability analysis with all-paths-return detection
func (ra *ReachabilityAnalyzer) performEnhancedReachabilityAnalysis(result *ReachabilityResult) {
	// First, perform standard reachability analysis from entry
	visitor := &reachabilityVisitor{
		reachableBlocks: result.ReachableBlocks,
	}
	ra.cfg.Walk(visitor)

	// Then, apply all-paths-return analysis to potentially mark more blocks as unreachable
	ra.detectAllPathsReturnUnreachability(result)

	// Finally, identify unreachable blocks by comparing against all blocks
	for id, block := range ra.cfg.Blocks {
		if _, isReachable := result.ReachableBlocks[id]; !isReachable {
			result.UnreachableBlocks[id] = block
		}
	}
}

// detectAllPathsReturnUnreachability detects blocks unreachable due to all-paths-return scenarios
func (ra *ReachabilityAnalyzer) detectAllPathsReturnUnreachability(result *ReachabilityResult) {
	// Track blocks that have all paths leading to returns
	allPathsReturnBlocks := make(map[string]bool)

	// First pass: identify blocks where all successors eventually return
	for _, block := range ra.cfg.Blocks {
		if ra.allSuccessorsReturn(block, make(map[string]bool)) {
			allPathsReturnBlocks[block.ID] = true
		}
	}

	// Second pass: mark successors of all-paths-return blocks as unreachable
	// unless they are also part of an all-paths-return scenario
	for blockID := range allPathsReturnBlocks {
		block := ra.cfg.Blocks[blockID]
		ra.markSuccessorsUnreachableAfterReturn(block, result, make(map[string]bool))
	}
}

// allSuccessorsReturn checks if all execution paths from this block lead to returns
func (ra *ReachabilityAnalyzer) allSuccessorsReturn(block *BasicBlock, visited map[string]bool) bool {
	if block == nil {
		return false
	}

	// Check cache first (memoization)
	if ra.allPathsReturnComputed[block.ID] {
		return ra.allPathsReturnCache[block.ID]
	}

	// Avoid infinite recursion (cycle detection)
	if visited[block.ID] {
		return false
	}
	visited[block.ID] = true
	defer func() { visited[block.ID] = false }()

	// If this block contains a return statement, it leads to return
	if ra.blockContainsReturn(block) {
		ra.allPathsReturnCache[block.ID] = true
		ra.allPathsReturnComputed[block.ID] = true
		return true
	}

	// If this is the exit block, it doesn't count as a return
	if block == ra.cfg.Exit {
		ra.allPathsReturnCache[block.ID] = false
		ra.allPathsReturnComputed[block.ID] = true
		return false
	}

	// If no successors, it's not a return path
	if len(block.Successors) == 0 {
		ra.allPathsReturnCache[block.ID] = false
		ra.allPathsReturnComputed[block.ID] = true
		return false
	}

	// All successors must lead to returns
	for _, edge := range block.Successors {
		// Skip return edges to exit as they represent actual returns
		if edge.Type == EdgeReturn && edge.To == ra.cfg.Exit {
			continue
		}

		if !ra.allSuccessorsReturn(edge.To, visited) {
			ra.allPathsReturnCache[block.ID] = false
			ra.allPathsReturnComputed[block.ID] = true
			return false
		}
	}

	ra.allPathsReturnCache[block.ID] = true
	ra.allPathsReturnComputed[block.ID] = true
	return true
}

// markSuccessorsUnreachableAfterReturn marks blocks unreachable after all-paths-return blocks
func (ra *ReachabilityAnalyzer) markSuccessorsUnreachableAfterReturn(block *BasicBlock, result *ReachabilityResult, visited map[string]bool) {
	if block == nil || visited[block.ID] {
		return
	}
	visited[block.ID] = true

	// If this block contains a return, its fall-through successors are unreachable
	if ra.blockContainsReturn(block) {
		for _, edge := range block.Successors {
			// Only normal edges after returns are unreachable (not return edges to exit)
			if edge.Type == EdgeNormal {
				// Remove from reachable blocks if it was marked as such
				delete(result.ReachableBlocks, edge.To.ID)

				// Recursively mark successors as unreachable
				ra.markSuccessorsUnreachableAfterReturn(edge.To, result, copyVisited(visited))
			}
		}
	}
}

// blockContainsReturn checks if a block contains a return statement
func (ra *ReachabilityAnalyzer) blockContainsReturn(block *BasicBlock) bool {
	if block == nil {
		return false
	}

	for _, stmt := range block.Statements {
		if stmt != nil && stmt.Type == parser.NodeReturn {
			return true
		}
	}

	return false
}

// copyVisited creates a copy of the visited map to avoid sharing state between recursive calls
func copyVisited(visited map[string]bool) map[string]bool {
	copy := make(map[string]bool)
	for k, v := range visited {
		copy[k] = v
	}
	return copy
}
