package analyzer

import (
	"sort"
	"strings"
	"time"
	
	"github.com/pyqol/pyqol/internal/parser"
)

// SeverityLevel represents the severity of a dead code finding
type SeverityLevel string

const (
	// SeverityLevelCritical indicates code that is definitely unreachable
	SeverityLevelCritical SeverityLevel = "critical"
	
	// SeverityLevelWarning indicates code that is likely unreachable
	SeverityLevelWarning SeverityLevel = "warning"
	
	// SeverityLevelInfo indicates potential optimization opportunities
	SeverityLevelInfo SeverityLevel = "info"
)

// DeadCodeReason represents the reason why code is considered dead
type DeadCodeReason string

const (
	// ReasonUnreachableAfterReturn indicates code after a return statement
	ReasonUnreachableAfterReturn DeadCodeReason = "unreachable_after_return"
	
	// ReasonUnreachableAfterBreak indicates code after a break statement
	ReasonUnreachableAfterBreak DeadCodeReason = "unreachable_after_break"
	
	// ReasonUnreachableAfterContinue indicates code after a continue statement
	ReasonUnreachableAfterContinue DeadCodeReason = "unreachable_after_continue"
	
	// ReasonUnreachableAfterRaise indicates code after a raise statement
	ReasonUnreachableAfterRaise DeadCodeReason = "unreachable_after_raise"
	
	// ReasonUnreachableBranch indicates an unreachable branch condition
	ReasonUnreachableBranch DeadCodeReason = "unreachable_branch"
	
	// ReasonUnreachableAfterInfiniteLoop indicates code after an infinite loop
	ReasonUnreachableAfterInfiniteLoop DeadCodeReason = "unreachable_after_infinite_loop"
)

// DeadCodeFinding represents a single dead code detection result
type DeadCodeFinding struct {
	// Function information
	FunctionName string `json:"function_name"`
	FilePath     string `json:"file_path"`
	
	// Location information
	StartLine int `json:"start_line"`
	EndLine   int `json:"end_line"`
	
	// Dead code details
	BlockID     string              `json:"block_id"`
	Code        string              `json:"code"`
	Reason      DeadCodeReason      `json:"reason"`
	Severity    SeverityLevel       `json:"severity"`
	Description string              `json:"description"`
	
	// Context information
	Context []string `json:"context,omitempty"`
}

// DeadCodeResult contains the results of dead code analysis for a single CFG
type DeadCodeResult struct {
	// Function information
	FunctionName string `json:"function_name"`
	FilePath     string `json:"file_path"`
	
	// Analysis results
	Findings       []*DeadCodeFinding `json:"findings"`
	TotalBlocks    int                `json:"total_blocks"`
	DeadBlocks     int                `json:"dead_blocks"`
	ReachableRatio float64            `json:"reachable_ratio"`
	
	// Performance metrics
	AnalysisTime time.Duration `json:"analysis_time"`
}

// DeadCodeDetector provides high-level dead code detection functionality
type DeadCodeDetector struct {
	cfg      *CFG
	filePath string // File path for context in findings
}

// NewDeadCodeDetector creates a new dead code detector for the given CFG
func NewDeadCodeDetector(cfg *CFG) *DeadCodeDetector {
	return &DeadCodeDetector{
		cfg:      cfg,
		filePath: "", // Will be set by caller if needed
	}
}

// NewDeadCodeDetectorWithFilePath creates a new dead code detector with file path context
func NewDeadCodeDetectorWithFilePath(cfg *CFG, filePath string) *DeadCodeDetector {
	return &DeadCodeDetector{
		cfg:      cfg,
		filePath: filePath,
	}
}

// Detect performs dead code detection and returns structured findings
func (dcd *DeadCodeDetector) Detect() *DeadCodeResult {
	startTime := time.Now()
	
	result := &DeadCodeResult{
		FunctionName: dcd.getFunctionName(),
		FilePath:     dcd.getFilePath(),
		Findings:     make([]*DeadCodeFinding, 0),
		TotalBlocks:  0,
		DeadBlocks:   0,
		AnalysisTime: time.Since(startTime),
	}
	
	// Handle nil or empty CFG
	if dcd.cfg == nil || dcd.cfg.Blocks == nil {
		return result
	}
	
	result.TotalBlocks = len(dcd.cfg.Blocks)
	
	// Use reachability analyzer to find unreachable blocks
	analyzer := NewReachabilityAnalyzer(dcd.cfg)
	reachResult := analyzer.AnalyzeReachability()
	
	result.ReachableRatio = reachResult.GetReachabilityRatio()
	
	// Convert unreachable blocks to dead code findings
	unreachableWithStatements := reachResult.GetUnreachableBlocksWithStatements()
	result.DeadBlocks = len(unreachableWithStatements)
	
	for _, block := range unreachableWithStatements {
		findings := dcd.analyzeDeadBlock(block)
		result.Findings = append(result.Findings, findings...)
	}
	
	// Sort findings by line number for consistent output
	sort.Slice(result.Findings, func(i, j int) bool {
		return result.Findings[i].StartLine < result.Findings[j].StartLine
	})
	
	result.AnalysisTime = time.Since(startTime)
	return result
}

// DetectInFunction analyzes a single CFG and returns findings
func DetectInFunction(cfg *CFG) *DeadCodeResult {
	detector := NewDeadCodeDetector(cfg)
	return detector.Detect()
}

// DetectInFunctionWithFilePath analyzes a single CFG with file path context
func DetectInFunctionWithFilePath(cfg *CFG, filePath string) *DeadCodeResult {
	detector := NewDeadCodeDetectorWithFilePath(cfg, filePath)
	return detector.Detect()
}

// DetectInFile analyzes multiple CFGs from a file and returns combined findings
func DetectInFile(cfgs map[string]*CFG, filePath string) []*DeadCodeResult {
	var results []*DeadCodeResult
	
	for functionName, cfg := range cfgs {
		// Skip the main module CFG for now, focus on functions
		if functionName == "__main__" {
			continue
		}
		
		// Use the file path-aware constructor for accurate reporting
		detector := NewDeadCodeDetectorWithFilePath(cfg, filePath)
		result := detector.Detect()
		result.FunctionName = functionName
		// FilePath is already set by the detector
		
		// Only include results that have findings
		if len(result.Findings) > 0 || result.DeadBlocks > 0 {
			results = append(results, result)
		}
	}
	
	return results
}

// analyzeDeadBlock analyzes a dead block and creates appropriate findings
func (dcd *DeadCodeDetector) analyzeDeadBlock(block *BasicBlock) []*DeadCodeFinding {
	var findings []*DeadCodeFinding
	
	if block == nil || len(block.Statements) == 0 {
		return findings
	}
	
	// Determine the reason for the dead code
	reason, severity := dcd.determineDeadCodeReason(block)
	
	// Create a finding for this dead block
	finding := &DeadCodeFinding{
		FunctionName: dcd.getFunctionName(),
		FilePath:     dcd.getFilePath(),
		StartLine:    dcd.getBlockStartLine(block),
		EndLine:      dcd.getBlockEndLine(block),
		BlockID:      block.ID,
		Code:         dcd.getBlockCode(block),
		Reason:       reason,
		Severity:     severity,
		Description:  dcd.generateDescription(reason, block),
		Context:      dcd.getBlockContext(block),
	}
	
	findings = append(findings, finding)
	
	return findings
}

// determineDeadCodeReason analyzes the block to determine why it's dead
func (dcd *DeadCodeDetector) determineDeadCodeReason(block *BasicBlock) (DeadCodeReason, SeverityLevel) {
	// Check direct predecessors for control flow patterns
	reason := ReasonUnreachableBranch // default
	severity := SeverityLevelWarning  // default
	
	// Analyze control flow patterns by checking predecessors
	if terminatorReason, terminatorSeverity := dcd.findTerminatorInPredecessors(block); terminatorReason != "" {
		reason = terminatorReason
		severity = terminatorSeverity
	}
	
	return reason, severity
}

// findTerminatorInPredecessors efficiently finds terminator statements in control flow predecessors
func (dcd *DeadCodeDetector) findTerminatorInPredecessors(block *BasicBlock) (DeadCodeReason, SeverityLevel) {
	if block == nil {
		return "", SeverityLevelWarning
	}

	// First, check all blocks in the CFG for terminators that precede this block
	// This handles cases where CFG edges might not be perfectly set up
	blockStartLine := dcd.getBlockStartLine(block)
	
	for _, otherBlock := range dcd.cfg.Blocks {
		if otherBlock == nil || otherBlock == block {
			continue
		}
		
		otherEndLine := dcd.getBlockEndLine(otherBlock)
		
		// Check if the other block ends before this block starts (sequential in source)
		if otherEndLine < blockStartLine && (blockStartLine - otherEndLine) <= 5 {
			if dcd.blockContainsReturn(otherBlock) {
				return ReasonUnreachableAfterReturn, SeverityLevelCritical
			}
			if dcd.blockContainsBreak(otherBlock) {
				return ReasonUnreachableAfterBreak, SeverityLevelCritical
			}
			if dcd.blockContainsContinue(otherBlock) {
				return ReasonUnreachableAfterContinue, SeverityLevelCritical
			}
			if dcd.blockContainsRaise(otherBlock) {
				return ReasonUnreachableAfterRaise, SeverityLevelCritical
			}
		}
	}

	// Secondary check: use CFG edges if available
	for _, predEdge := range block.Predecessors {
		if predEdge == nil || predEdge.From == nil {
			continue
		}

		predBlock := predEdge.From
		
		// Check for terminator statements in predecessor block
		if dcd.blockContainsReturn(predBlock) {
			if dcd.isSequentiallyAfter(predBlock, block) {
				return ReasonUnreachableAfterReturn, SeverityLevelCritical
			}
		}
		if dcd.blockContainsBreak(predBlock) {
			if dcd.isSequentiallyAfter(predBlock, block) {
				return ReasonUnreachableAfterBreak, SeverityLevelCritical
			}
		}
		if dcd.blockContainsContinue(predBlock) {
			if dcd.isSequentiallyAfter(predBlock, block) {
				return ReasonUnreachableAfterContinue, SeverityLevelCritical
			}
		}
		if dcd.blockContainsRaise(predBlock) {
			if dcd.isSequentiallyAfter(predBlock, block) {
				return ReasonUnreachableAfterRaise, SeverityLevelCritical
			}
		}
	}
	
	return "", SeverityLevelWarning
}


// blockContainsReturn checks if a block contains a return statement
func (dcd *DeadCodeDetector) blockContainsReturn(block *BasicBlock) bool {
	for _, stmt := range block.Statements {
		if stmt.Type == parser.NodeReturn {
			return true
		}
	}
	return false
}

// blockContainsBreak checks if a block contains a break statement
func (dcd *DeadCodeDetector) blockContainsBreak(block *BasicBlock) bool {
	for _, stmt := range block.Statements {
		if stmt.Type == parser.NodeBreak {
			return true
		}
	}
	return false
}

// blockContainsContinue checks if a block contains a continue statement
func (dcd *DeadCodeDetector) blockContainsContinue(block *BasicBlock) bool {
	for _, stmt := range block.Statements {
		if stmt.Type == parser.NodeContinue {
			return true
		}
	}
	return false
}

// blockContainsRaise checks if a block contains a raise statement
func (dcd *DeadCodeDetector) blockContainsRaise(block *BasicBlock) bool {
	for _, stmt := range block.Statements {
		if stmt.Type == parser.NodeRaise {
			return true
		}
	}
	return false
}

// isSequentiallyAfter checks if successor block comes sequentially after predecessor
// This uses both CFG edge analysis and line number heuristics for accurate detection
func (dcd *DeadCodeDetector) isSequentiallyAfter(predecessor, successor *BasicBlock) bool {
	if predecessor == nil || successor == nil {
		return false
	}

	// Primary check: line numbers (for dead code after return/break/continue/raise)
	predEnd := dcd.getBlockEndLine(predecessor)
	succStart := dcd.getBlockStartLine(successor)
	
	// If successor comes immediately after predecessor in source code
	if predEnd < succStart && (succStart - predEnd) <= 10 { // Allow reasonable gap
		return true
	}

	// Secondary check: CFG edge analysis for complex control flow
	for _, succEdge := range predecessor.Successors {
		if succEdge != nil && succEdge.To == successor {
			// Consider normal sequential flow
			if succEdge.Type == EdgeNormal {
				return true
			}
			// Also consider cases where the terminator forces this flow
			if succEdge.Type == EdgeReturn || succEdge.Type == EdgeBreak || 
			   succEdge.Type == EdgeContinue {
				return true
			}
		}
	}
	
	return false
}

// Helper methods for extracting information from blocks

// getFunctionName extracts the function name from the CFG
func (dcd *DeadCodeDetector) getFunctionName() string {
	if dcd.cfg == nil || dcd.cfg.Name == "" {
		return "unknown"
	}
	return dcd.cfg.Name
}

// getFilePath extracts the file path from the detector context
func (dcd *DeadCodeDetector) getFilePath() string {
	if dcd.filePath != "" {
		return dcd.filePath
	}
	// Fallback for backward compatibility
	return "unknown"
}

// getBlockStartLine gets the starting line number of a block
func (dcd *DeadCodeDetector) getBlockStartLine(block *BasicBlock) int {
	if block == nil || len(block.Statements) == 0 {
		return 0
	}
	return block.Statements[0].Location.StartLine
}

// getBlockEndLine gets the ending line number of a block
func (dcd *DeadCodeDetector) getBlockEndLine(block *BasicBlock) int {
	if block == nil || len(block.Statements) == 0 {
		return 0
	}
	return block.Statements[len(block.Statements)-1].Location.EndLine
}

// getBlockCode extracts the code from a block
func (dcd *DeadCodeDetector) getBlockCode(block *BasicBlock) string {
	if block == nil || len(block.Statements) == 0 {
		return ""
	}
	
	var codes []string
	for _, stmt := range block.Statements {
		// Create a simple representation of the statement
		nodeDesc := dcd.getNodeDescription(stmt)
		codes = append(codes, strings.TrimSpace(nodeDesc))
	}
	
	return strings.Join(codes, "\n")
}

// getNodeDescription creates a simple description of a parser node
func (dcd *DeadCodeDetector) getNodeDescription(node *parser.Node) string {
	if node == nil {
		return ""
	}
	
	switch node.Type {
	case parser.NodeReturn:
		return "return"
	case parser.NodeBreak:
		return "break"
	case parser.NodeContinue:
		return "continue"
	case parser.NodeRaise:
		return "raise"
	case parser.NodeAssign:
		if node.Name != "" {
			return node.Name + " = ..."
		}
		return "assignment"
	case parser.NodeExpr:
		return "expression"
	case parser.NodeIf:
		return "if"
	case parser.NodeFor:
		return "for"
	case parser.NodeWhile:
		return "while"
	case parser.NodeTry:
		return "try"
	case parser.NodePass:
		return "pass"
	default:
		return string(node.Type)
	}
}

// getBlockContext provides context around the dead code
func (dcd *DeadCodeDetector) getBlockContext(block *BasicBlock) []string {
	// For now, return empty context
	// This can be enhanced to show surrounding code
	return []string{}
}

// generateDescription creates a human-readable description of the dead code
func (dcd *DeadCodeDetector) generateDescription(reason DeadCodeReason, block *BasicBlock) string {
	switch reason {
	case ReasonUnreachableAfterReturn:
		return "Code appears after a return statement and will never be executed"
	case ReasonUnreachableAfterBreak:
		return "Code appears after a break statement and will never be executed"
	case ReasonUnreachableAfterContinue:
		return "Code appears after a continue statement and will never be executed"
	case ReasonUnreachableAfterRaise:
		return "Code appears after a raise statement and will never be executed"
	case ReasonUnreachableBranch:
		return "Code in this branch is unreachable under normal execution flow"
	case ReasonUnreachableAfterInfiniteLoop:
		return "Code appears after an infinite loop and will never be executed"
	default:
		return "Code is unreachable and will never be executed"
	}
}

// HasDeadCode checks if the CFG contains any dead code
func (dcd *DeadCodeDetector) HasDeadCode() bool {
	if dcd.cfg == nil {
		return false
	}
	
	analyzer := NewReachabilityAnalyzer(dcd.cfg)
	result := analyzer.AnalyzeReachability()
	return result.HasUnreachableCode()
}

// GetDeadCodeRatio returns the ratio of dead blocks to total blocks
func (dcd *DeadCodeDetector) GetDeadCodeRatio() float64 {
	result := dcd.Detect()
	if result.TotalBlocks == 0 {
		return 0.0
	}
	return float64(result.DeadBlocks) / float64(result.TotalBlocks)
}

// FilterFindingsBySeverity filters findings by minimum severity level
func FilterFindingsBySeverity(findings []*DeadCodeFinding, minSeverity SeverityLevel) []*DeadCodeFinding {
	severityOrder := map[SeverityLevel]int{
		SeverityLevelInfo:     1,
		SeverityLevelWarning:  2,
		SeverityLevelCritical: 3,
	}
	
	minLevel := severityOrder[minSeverity]
	var filtered []*DeadCodeFinding
	
	for _, finding := range findings {
		if severityOrder[finding.Severity] >= minLevel {
			filtered = append(filtered, finding)
		}
	}
	
	return filtered
}

// GroupFindingsByReason groups findings by their reason
func GroupFindingsByReason(findings []*DeadCodeFinding) map[DeadCodeReason][]*DeadCodeFinding {
	groups := make(map[DeadCodeReason][]*DeadCodeFinding)
	
	for _, finding := range findings {
		groups[finding.Reason] = append(groups[finding.Reason], finding)
	}
	
	return groups
}