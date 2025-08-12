package analyzer

import (
	"fmt"
	"time"
)

// ReachabilityAnalyzer performs reachability analysis on a CFG
type ReachabilityAnalyzer struct {
	cfg              *CFG
	reachableBlocks  map[string]bool
	processedBlocks  map[string]bool
	entryPoints      []*BasicBlock
}

// ReachabilityReport contains the results of reachability analysis
type ReachabilityReport struct {
	TotalBlocks       int
	ReachableBlocks   int
	UnreachableBlocks int
	UnreachableList   []*BasicBlock
	AnalysisTime      time.Duration
}

// NewReachabilityAnalyzer creates a new reachability analyzer for a CFG
func NewReachabilityAnalyzer(cfg *CFG) *ReachabilityAnalyzer {
	return &ReachabilityAnalyzer{
		cfg:             cfg,
		reachableBlocks: make(map[string]bool),
		processedBlocks: make(map[string]bool),
		entryPoints:     []*BasicBlock{cfg.Entry},
	}
}

// AddEntryPoint adds an additional entry point for analysis
// This is useful for exception handlers or other special blocks
func (ra *ReachabilityAnalyzer) AddEntryPoint(block *BasicBlock) {
	if block != nil {
		ra.entryPoints = append(ra.entryPoints, block)
	}
}

// AnalyzeReachability performs the reachability analysis
func (ra *ReachabilityAnalyzer) AnalyzeReachability() *ReachabilityReport {
	startTime := time.Now()
	
	// Clear previous analysis
	ra.reachableBlocks = make(map[string]bool)
	ra.processedBlocks = make(map[string]bool)
	
	// Perform DFS from each entry point
	for _, entry := range ra.entryPoints {
		ra.markReachableFromBlock(entry)
	}
	
	// Build the report
	report := &ReachabilityReport{
		TotalBlocks:     len(ra.cfg.Blocks),
		ReachableBlocks: len(ra.reachableBlocks),
		AnalysisTime:    time.Since(startTime),
	}
	
	// Identify unreachable blocks
	for id, block := range ra.cfg.Blocks {
		if !ra.reachableBlocks[id] {
			report.UnreachableList = append(report.UnreachableList, block)
		}
	}
	
	report.UnreachableBlocks = len(report.UnreachableList)
	
	return report
}

// markReachableFromBlock performs DFS to mark all reachable blocks
func (ra *ReachabilityAnalyzer) markReachableFromBlock(block *BasicBlock) {
	if block == nil || ra.processedBlocks[block.ID] {
		return
	}
	
	// Mark as processed to avoid cycles
	ra.processedBlocks[block.ID] = true
	
	// Mark as reachable
	ra.reachableBlocks[block.ID] = true
	
	// Process all successors
	for _, edge := range block.Successors {
		ra.markReachableFromBlock(edge.To)
	}
}

// IsReachable checks if a specific block is reachable
func (ra *ReachabilityAnalyzer) IsReachable(block *BasicBlock) bool {
	if block == nil {
		return false
	}
	return ra.reachableBlocks[block.ID]
}

// GetUnreachableBlocks returns all unreachable blocks
func (ra *ReachabilityAnalyzer) GetUnreachableBlocks() []*BasicBlock {
	var unreachable []*BasicBlock
	
	for id, block := range ra.cfg.Blocks {
		if !ra.reachableBlocks[id] {
			unreachable = append(unreachable, block)
		}
	}
	
	return unreachable
}

// GetReachableBlocks returns all reachable blocks
func (ra *ReachabilityAnalyzer) GetReachableBlocks() []*BasicBlock {
	var reachable []*BasicBlock
	
	for id, block := range ra.cfg.Blocks {
		if ra.reachableBlocks[id] {
			reachable = append(reachable, block)
		}
	}
	
	return reachable
}

// MarkUnreachableCode identifies and marks unreachable code patterns
func (ra *ReachabilityAnalyzer) MarkUnreachableCode() {
	for _, block := range ra.GetUnreachableBlocks() {
		// Set a label to indicate unreachable code
		if block.Label == "" {
			block.Label = LabelUnreachable
		} else {
			block.Label = fmt.Sprintf("%s (unreachable)", block.Label)
		}
	}
}

// String returns a summary of the reachability analysis
func (report *ReachabilityReport) String() string {
	return fmt.Sprintf(
		"Reachability Report:\n"+
			"  Total blocks: %d\n"+
			"  Reachable: %d\n"+
			"  Unreachable: %d\n"+
			"  Analysis time: %v",
		report.TotalBlocks,
		report.ReachableBlocks,
		report.UnreachableBlocks,
		report.AnalysisTime,
	)
}

// HasUnreachableCode returns true if any unreachable blocks were found
func (report *ReachabilityReport) HasUnreachableCode() bool {
	return report.UnreachableBlocks > 0
}

// GetUnreachableBlockIDs returns a list of unreachable block IDs
func (report *ReachabilityReport) GetUnreachableBlockIDs() []string {
	ids := make([]string, 0, len(report.UnreachableList))
	for _, block := range report.UnreachableList {
		ids = append(ids, block.ID)
	}
	return ids
}