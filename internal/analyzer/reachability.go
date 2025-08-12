package analyzer

import (
	"time"
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
	cfg *CFG
}

// NewReachabilityAnalyzer creates a new reachability analyzer for the given CFG
func NewReachabilityAnalyzer(cfg *CFG) *ReachabilityAnalyzer {
	return &ReachabilityAnalyzer{
		cfg: cfg,
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

	// Use the existing CFG Walk method to identify reachable blocks
	visitor := &reachabilityVisitor{
		reachableBlocks: result.ReachableBlocks,
	}

	ra.cfg.Walk(visitor)

	// Identify unreachable blocks by comparing against all blocks
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

