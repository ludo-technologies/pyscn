package analyzer

import (
	"time"

	corecfg "github.com/ludo-technologies/polyscan/core/cfg"
)

// ReachabilityResult preserves pyscn's enriched projection of core reachability.
type ReachabilityResult struct {
	ReachableBlocks   map[string]*BasicBlock
	UnreachableBlocks map[string]*BasicBlock
	TotalBlocks       int
	ReachableCount    int
	UnreachableCount  int
	AnalysisTime      time.Duration
}

// ReachabilityAnalyzer performs reachability analysis on CFGs.
type ReachabilityAnalyzer struct {
	cfg *CFG
}

// NewReachabilityAnalyzer creates a new reachability analyzer for the given CFG.
func NewReachabilityAnalyzer(cfg *CFG) *ReachabilityAnalyzer {
	return &ReachabilityAnalyzer{cfg: cfg}
}

// AnalyzeReachability performs reachability analysis starting from the entry block.
func (ra *ReachabilityAnalyzer) AnalyzeReachability() *ReachabilityResult {
	startTime := time.Now()
	result := newReachabilityResult(ra.cfg)
	if ra.cfg == nil || ra.cfg.Entry == nil || ra.cfg.Blocks == nil {
		result.AnalysisTime = time.Since(startTime)
		return result
	}

	coreResult := corecfg.AnalyzeReachability(ra.cfg, corecfg.ReachabilityConfig{
		Classifier: pythonCFGClassifier{},
	})
	for id, block := range ra.cfg.Blocks {
		if coreResult.Reachable[id] {
			result.ReachableBlocks[id] = block
		} else {
			result.UnreachableBlocks[id] = block
		}
	}
	result.ReachableCount = coreResult.ReachableCount
	result.UnreachableCount = coreResult.UnreachableCount
	result.AnalysisTime = time.Since(startTime)
	return result
}

// AnalyzeReachabilityFrom retains pyscn's explicit-start structural traversal.
func (ra *ReachabilityAnalyzer) AnalyzeReachabilityFrom(startBlock *BasicBlock) *ReachabilityResult {
	startTime := time.Now()
	result := newReachabilityResult(ra.cfg)
	if ra.cfg == nil || ra.cfg.Blocks == nil || startBlock == nil {
		result.AnalysisTime = time.Since(startTime)
		return result
	}

	visited := make(map[string]bool)
	var visit func(*BasicBlock)
	visit = func(block *BasicBlock) {
		if block == nil || visited[block.ID] {
			return
		}
		visited[block.ID] = true
		result.ReachableBlocks[block.ID] = block
		for _, edge := range block.Successors {
			visit(edge.To)
		}
	}
	visit(startBlock)
	for id, block := range ra.cfg.Blocks {
		if !visited[id] {
			result.UnreachableBlocks[id] = block
		}
	}
	result.ReachableCount = len(result.ReachableBlocks)
	result.UnreachableCount = len(result.UnreachableBlocks)
	result.AnalysisTime = time.Since(startTime)
	return result
}

func newReachabilityResult(graph *CFG) *ReachabilityResult {
	result := &ReachabilityResult{
		ReachableBlocks:   make(map[string]*BasicBlock),
		UnreachableBlocks: make(map[string]*BasicBlock),
	}
	if graph != nil && graph.Blocks != nil {
		result.TotalBlocks = len(graph.Blocks)
	}
	return result
}

// GetUnreachableBlocksWithStatements returns unreachable non-empty blocks.
func (result *ReachabilityResult) GetUnreachableBlocksWithStatements() map[string]*BasicBlock {
	blocks := make(map[string]*BasicBlock)
	for id, block := range result.UnreachableBlocks {
		if !block.IsEmpty() {
			blocks[id] = block
		}
	}
	return blocks
}

// GetReachabilityRatio returns the ratio of reachable blocks to total blocks.
func (result *ReachabilityResult) GetReachabilityRatio() float64 {
	if result.TotalBlocks == 0 {
		return 1
	}
	return float64(result.ReachableCount) / float64(result.TotalBlocks)
}

// HasUnreachableCode reports whether an unreachable block contains statements.
func (result *ReachabilityResult) HasUnreachableCode() bool {
	return len(result.GetUnreachableBlocksWithStatements()) > 0
}

// reachabilityVisitor remains for traversal benchmarks and compatibility tests.
type reachabilityVisitor struct {
	reachableBlocks map[string]*BasicBlock
}

func (rv *reachabilityVisitor) VisitBlock(block *BasicBlock) bool {
	if block != nil {
		rv.reachableBlocks[block.ID] = block
	}
	return true
}

func (rv *reachabilityVisitor) VisitEdge(_ *Edge) bool { return true }

func (ra *ReachabilityAnalyzer) blockContainsReturn(block *BasicBlock) bool {
	if block == nil {
		return false
	}
	classifier := pythonCFGClassifier{}
	for _, statement := range block.Statements {
		if classifier.IsReturn(statement) {
			return true
		}
	}
	return false
}
