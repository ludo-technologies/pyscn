package analyzer

import (
	"cmp"
	"maps"
	"slices"
	"strings"
	"time"

	corecfg "github.com/ludo-technologies/polyscan/core/cfg"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
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
	// Execution-scope information
	FunctionName string                   `json:"function_name"`
	ScopeKind    domain.AnalysisScopeKind `json:"scope_kind"`
	FilePath     string                   `json:"file_path"`

	// Location information
	StartLine int `json:"start_line"`
	EndLine   int `json:"end_line"`

	// Dead code details
	BlockID     string         `json:"block_id"`
	Code        string         `json:"code"`
	Reason      DeadCodeReason `json:"reason"`
	Severity    SeverityLevel  `json:"severity"`
	Description string         `json:"description"`

	// Context information
	Context []string `json:"context,omitempty"`
}

// DeadCodeResult contains the results of dead code analysis for a single CFG
type DeadCodeResult struct {
	// Execution-scope information
	FunctionName string                   `json:"function_name"`
	ScopeKind    domain.AnalysisScopeKind `json:"scope_kind"`
	FilePath     string                   `json:"file_path"`

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
	scope    CFGScope
}

// NewDeadCodeDetector creates a detector for a function CFG.
func NewDeadCodeDetector(cfg *CFG) *DeadCodeDetector {
	return newDeadCodeDetectorForScope(cfg, "", functionCFGScope(cfg))
}

// NewDeadCodeDetectorWithFilePath creates a detector for a function CFG with
// file path context.
func NewDeadCodeDetectorWithFilePath(cfg *CFG, filePath string) *DeadCodeDetector {
	return newDeadCodeDetectorForScope(cfg, filePath, functionCFGScope(cfg))
}

func newDeadCodeDetectorForScope(cfg *CFG, filePath string, scope CFGScope) *DeadCodeDetector {
	return &DeadCodeDetector{
		cfg:      cfg,
		filePath: filePath,
		scope:    scope,
	}
}

// Detect performs dead code detection and returns structured findings
func (dcd *DeadCodeDetector) Detect() *DeadCodeResult {
	startTime := time.Now()

	result := &DeadCodeResult{
		FunctionName: dcd.getFunctionName(),
		ScopeKind:    dcd.scope.Kind,
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

	classifier := pythonCFGClassifier{}
	reachResult := corecfg.AnalyzeReachability(dcd.cfg, corecfg.ReachabilityConfig{Classifier: classifier})
	if result.TotalBlocks > 0 {
		result.ReachableRatio = float64(reachResult.ReachableCount) / float64(result.TotalBlocks)
	}
	coreResult := corecfg.DetectDeadCode(dcd.cfg, corecfg.DeadCodeConfig{Classifier: classifier})
	regionReasons, regionSeeds := deadRegionReasons(dcd.cfg, reachResult.Reachable)
	liveLines := reachableStatementLines(dcd.cfg, reachResult.Reachable)

	reportedBlocks := make(map[string]bool)
	for _, coreFinding := range coreResult.Findings {
		if reportedBlocks[coreFinding.BlockID] {
			continue
		}
		block := dcd.cfg.GetBlock(coreFinding.BlockID)
		findings := dcd.analyzeCoreDeadBlock(block, coreFinding.Reason, regionReasons)
		result.Findings = append(result.Findings, findings...)
		if len(findings) > 0 {
			reportedBlocks[coreFinding.BlockID] = true
			result.DeadBlocks++
		}
	}

	// Merge overlapping/contiguous findings that share a reason. A compound
	// statement (e.g. `if`) spans its body, so the body's own block produces a
	// finding whose line range is nested inside the `if` finding's range. Left
	// as-is, the same source line is reported—and tallied—more than once.
	//
	// Each dead region is collapsed first, because line adjacency is the wrong
	// test inside one: its blocks can be separated by blank or comment lines that
	// carry no reachable code. Only gaps free of reachable code are crossed, so an
	// unrelated live branch the CFG routes around keeps its own lines.
	result.Findings = mergeFindingsByRegion(result.Findings, regionSeeds, liveLines)

	// Line adjacency is the right test between regions, which is what this join
	// collapses into non-overlapping findings.
	result.Findings = dropNestedFindings(mergeContiguousFindings(result.Findings))

	result.AnalysisTime = time.Since(startTime)
	return result
}

// DetectInFunction analyzes a single CFG and returns findings
func DetectInFunction(cfg *CFG) *DeadCodeResult {
	return detectInScope(cfg, "", functionCFGScope(cfg))
}

// DetectInFunctionWithFilePath analyzes a single CFG with file path context
func DetectInFunctionWithFilePath(cfg *CFG, filePath string) *DeadCodeResult {
	return detectInScope(cfg, filePath, functionCFGScope(cfg))
}

func functionCFGScope(cfg *CFG) CFGScope {
	name := "unknown"
	if cfg != nil && cfg.Name != "" {
		name = cfg.Name
	}
	return CFGScope{Kind: domain.AnalysisScopeFunction, Name: name}
}

// DetectInScopeWithFilePath analyzes one explicitly owned execution scope.
func DetectInScopeWithFilePath(scopedCFG ScopedCFG, filePath string) *DeadCodeResult {
	return detectInScope(scopedCFG.Graph, filePath, scopedCFG.Scope)
}

func detectInScope(cfg *CFG, filePath string, scope CFGScope) *DeadCodeResult {
	return newDeadCodeDetectorForScope(cfg, filePath, scope).Detect()
}

// DetectInFile analyzes multiple CFGs from a file and returns combined findings
func DetectInFile(cfgs ControlFlowGraphs, filePath string) []*DeadCodeResult {
	var results []*DeadCodeResult

	for _, scopedCFG := range cfgs {
		if scopedCFG.Scope.Kind != domain.AnalysisScopeFunction && scopedCFG.Scope.Kind != domain.AnalysisScopeClass {
			continue
		}
		result := DetectInScopeWithFilePath(scopedCFG, filePath)

		// Only include results that have findings
		if len(result.Findings) > 0 || result.DeadBlocks > 0 {
			results = append(results, result)
		}
	}

	return results
}

func (dcd *DeadCodeDetector) analyzeCoreDeadBlock(block *BasicBlock, coreReason string, regionReasons map[string]DeadCodeReason) []*DeadCodeFinding {
	var findings []*DeadCodeFinding

	if block == nil || len(block.Statements) == 0 {
		return findings
	}

	// Skip blocks whose only "statements" are empty separators (a bare `;`).
	// In Python a trailing semicolon (`raise X;` or `return y;`) parses as the
	// terminating statement followed by an empty statement. That empty statement
	// is technically unreachable, but reporting it as `unreachable_branch` with
	// `code: ";"` and a `0-0` column range is noise — there's nothing for the
	// user to act on beyond a stylistic trailing semicolon.
	if isOnlyNoOpStatements(block) {
		return findings
	}

	// Skip the unreachable bare `yield` that only exists to make the enclosing
	// function a generator:
	//
	//	async def run(self) -> AsyncGenerator[Event, None]:
	//	    raise NotImplementedError()
	//	    yield
	//
	// Python decides generator-ness syntactically, so removing that `yield`
	// turns the function into a plain coroutine and breaks its contract. There
	// is nothing for the user to act on. Only the sole `yield` of a scope earns
	// the exemption — dead code in a function that already yields is real.
	if isGeneratorMarkerBlock(block) && countScopeYields(cfgSourceNode(dcd.cfg)) == 1 {
		return findings
	}

	reason, severity := ReasonUnreachableBranch, SeverityLevelWarning
	if regionReason, ok := regionReasons[block.ID]; ok {
		reason, severity = regionReason, SeverityLevelCritical
	}
	switch coreReason {
	case "after_return":
		reason, severity = ReasonUnreachableAfterReturn, SeverityLevelCritical
	case "after_break":
		reason, severity = ReasonUnreachableAfterBreak, SeverityLevelCritical
	case "after_continue":
		reason, severity = ReasonUnreachableAfterContinue, SeverityLevelCritical
	case "after_throw":
		reason, severity = ReasonUnreachableAfterRaise, SeverityLevelCritical
	}

	startLine, endLine := dcd.getBlockStartLine(block), dcd.getBlockEndLine(block)
	if try := enclosingDeadTry(block); try != nil {
		startLine = min(startLine, try.Location.StartLine)
		endLine = max(endLine, try.Location.EndLine)
	}

	// Create a finding for this dead block
	finding := &DeadCodeFinding{
		FunctionName: dcd.getFunctionName(),
		ScopeKind:    dcd.scope.Kind,
		FilePath:     dcd.getFilePath(),
		StartLine:    startLine,
		EndLine:      endLine,
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

// Helper methods for extracting information from blocks

// getFunctionName extracts the function name from the CFG
func (dcd *DeadCodeDetector) getFunctionName() string {
	if dcd.scope.Name != "" {
		return dcd.scope.Name
	}
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
	node := mustPythonNode(block.Statements[0])
	return node.Location.StartLine
}

// getBlockEndLine gets the ending line number of a block
func (dcd *DeadCodeDetector) getBlockEndLine(block *BasicBlock) int {
	if block == nil || len(block.Statements) == 0 {
		return 0
	}
	node := mustPythonNode(block.Statements[len(block.Statements)-1])
	return node.Location.EndLine
}

// getBlockCode extracts the code from a block
func (dcd *DeadCodeDetector) getBlockCode(block *BasicBlock) string {
	if block == nil || len(block.Statements) == 0 {
		return ""
	}

	var codes []string
	for _, value := range block.Statements {
		stmt := mustPythonNode(value)
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
	return dcd.Detect().DeadBlocks > 0
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

// mergeContiguousFindings collapses findings whose line ranges overlap or are
// directly adjacent (no reachable line between them) and that share the same
// reason into a single finding. Findings must be pre-sorted by StartLine (then
// EndLine). This removes the overlapping/duplicate ranges that arise because a
// compound statement's finding spans its body while the body's block emits its
// own nested finding.
func mergeContiguousFindings(findings []*DeadCodeFinding) []*DeadCodeFinding {
	lineFindings := make([]*corecfg.LineFinding, 0, len(findings))
	origins := make(map[*corecfg.LineFinding]*DeadCodeFinding, len(findings))
	for _, finding := range findings {
		lineFinding := &corecfg.LineFinding{
			StartLine:   finding.StartLine,
			EndLine:     finding.EndLine,
			Reason:      string(finding.Reason),
			Severity:    toCoreSeverity(finding.Severity),
			Description: finding.Description,
			Code:        finding.Code,
		}
		lineFindings = append(lineFindings, lineFinding)
		origins[lineFinding] = finding
	}
	corecfg.SortLineFindings(lineFindings)
	mergedLines := corecfg.MergeContiguousFindings(lineFindings)
	merged := make([]*DeadCodeFinding, 0, len(mergedLines))
	for _, lineFinding := range mergedLines {
		finding := origins[lineFinding]
		finding.StartLine = lineFinding.StartLine
		finding.EndLine = lineFinding.EndLine
		finding.Severity = fromCoreSeverity(lineFinding.Severity)
		finding.Description = lineFinding.Description
		finding.Code = lineFinding.Code
		merged = append(merged, finding)
	}
	return merged
}

// unreachableLabelReasons maps the label the CFG builder gives the block after
// each terminator to the reason reported for the code in that block.
var unreachableLabelReasons = map[string]DeadCodeReason{
	LabelUnreachableAfterReturn:   ReasonUnreachableAfterReturn,
	LabelUnreachableAfterRaise:    ReasonUnreachableAfterRaise,
	LabelUnreachableAfterBreak:    ReasonUnreachableAfterBreak,
	LabelUnreachableAfterContinue: ReasonUnreachableAfterContinue,
}

// reachableStatementLines collects every source line carrying a reachable
// statement. A dead region is a graph notion, not a line range: the blocks of a
// single region can be separated in the source by an unrelated live branch the
// CFG routes around, and merging across it would swallow code the analyzer
// still reports as reachable.
func reachableStatementLines(cfg *CFG, reachable map[string]bool) map[int]bool {
	lines := make(map[int]bool)
	for id, ok := range reachable {
		if !ok {
			continue
		}
		block := cfg.GetBlock(id)
		if block == nil {
			continue
		}
		for _, value := range block.Statements {
			node := mustPythonNode(value)
			location := node.Location
			endLine := location.EndLine
			if len(node.Body) > 0 {
				// Compound nodes include their entire body, but this block only
				// executes the header. Body statements have their own CFG blocks.
				// Keep multiline headers and headers with a same-line body live.
				endLine = min(endLine, max(location.StartLine, node.Body[0].Location.StartLine-1))
			}
			for line := location.StartLine; line <= endLine; line++ {
				lines[line] = true
			}
		}
	}
	return lines
}

// deadRegionReasons maps each unreachable block to the terminator that made
// its region dead. The block following a terminator is labeled with it, and
// every unreachable block reachable from that block inherits the reason, so a
// dead region reports one reason however far it extends. Blocks are visited in
// creation (source) order: a terminator inside an already dead region, whose
// following block flows back into that region, does not override its reason.
//
// The second return value maps each block to the seed of the region it belongs
// to, so callers can group findings describing the same region.
func deadRegionReasons(cfg *CFG, reachable map[string]bool) (map[string]DeadCodeReason, map[string]string) {
	blocks := slices.Collect(maps.Values(cfg.Blocks))
	// Block IDs are "bb<N>" in creation order; comparing length first sorts N numerically.
	slices.SortFunc(blocks, func(a, b *BasicBlock) int {
		return cmp.Or(cmp.Compare(len(a.ID), len(b.ID)), cmp.Compare(a.ID, b.ID))
	})

	reasons := make(map[string]DeadCodeReason)
	seeds := make(map[string]string)
	var mark func(block *BasicBlock, reason DeadCodeReason, seed string)
	mark = func(block *BasicBlock, reason DeadCodeReason, seed string) {
		if reachable[block.ID] {
			return
		}
		if _, marked := reasons[block.ID]; marked {
			return
		}
		reasons[block.ID] = reason
		seeds[block.ID] = seed
		for _, edge := range block.Successors {
			mark(edge.To, reason, seed)
		}
	}
	for _, block := range blocks {
		// createBlock suffixes the label with "_<counter>".
		label := block.Label[:max(strings.LastIndexByte(block.Label, '_'), 0)]
		if reason, ok := unreachableLabelReasons[label]; ok {
			mark(block, reason, block.ID)
		}
	}
	return reasons, seeds
}

// mergeFindingsByRegion collapses each dead region's findings into as few
// findings as the source allows. Because a region is a graph notion rather than
// a line range, a group is only joined across gaps carrying no reachable
// statement; that still reunites a region split by blank or comment lines while
// leaving live code the CFG routes around untouched.
//
// Findings that belong to no region (synthetic ones, and blocks the classifier
// never placed in a region) pass through untouched, so their identity survives.
//
// Must run before mergeContiguousFindings, which merges by line adjacency and
// would otherwise split the very ranges this combines.
func mergeFindingsByRegion(findings []*DeadCodeFinding, regionSeeds map[string]string, liveLines map[int]bool) []*DeadCodeFinding {
	grouped := make(map[string][]*DeadCodeFinding)
	kept := make([]*DeadCodeFinding, 0, len(findings))
	for _, finding := range findings {
		seed := regionSeeds[finding.BlockID]
		if seed == "" {
			kept = append(kept, finding)
			continue
		}
		grouped[seed] = append(grouped[seed], finding)
	}

	out := make([]*DeadCodeFinding, 0, len(kept)+len(grouped))
	out = append(out, kept...)
	for _, group := range grouped {
		out = append(out, mergeRegionGroup(group, liveLines)...)
	}
	return out
}

// mergeRegionGroup collapses one region's findings, but only as far as the
// source permits: a reachable statement between two findings ends the run,
// because the analyzer must keep reporting that statement's own range.
func mergeRegionGroup(group []*DeadCodeFinding, liveLines map[int]bool) []*DeadCodeFinding {
	// Block order is CFG order, not source order.
	slices.SortStableFunc(group, func(a, b *DeadCodeFinding) int {
		return cmp.Or(cmp.Compare(a.StartLine, b.StartLine), cmp.Compare(b.EndLine, a.EndLine))
	})

	merged := make([]*DeadCodeFinding, 0, len(group))
	for _, finding := range group {
		last := len(merged) - 1
		if last >= 0 && canJoinRegion(merged[last], finding, liveLines) {
			merged[last] = joinRegion(merged[last], finding)
			continue
		}
		merged = append(merged, finding)
	}
	return merged
}

// canJoinRegion reports whether two findings of one dead region may share a
// range. Overlapping or touching findings always may; otherwise the gap between
// them has to be free of reachable code.
func canJoinRegion(lo, hi *DeadCodeFinding, liveLines map[int]bool) bool {
	if hi.StartLine <= lo.EndLine {
		return true
	}
	for line := lo.EndLine + 1; line < hi.StartLine; line++ {
		if liveLines[line] {
			return false
		}
	}
	return true
}

// joinRegion widens one finding to absorb another of the same region, keeping
// the highest severity and the union of the code snippets.
func joinRegion(lo, hi *DeadCodeFinding) *DeadCodeFinding {
	lo.StartLine = min(lo.StartLine, hi.StartLine)
	lo.EndLine = max(lo.EndLine, hi.EndLine)
	if hi.Severity > lo.Severity {
		lo.Severity = hi.Severity
		lo.Description = hi.Description
	}
	lo.Code = joinSnippets(lo.Code, hi.Code)
	return lo
}

// joinSnippets concatenates two snippets, keeping the union of the
// statements they quote.
//
// No de-duplication happens here: the statements of two distinct blocks are
// distinct, so neighbouring statements that share a line - or that are spelled
// identically (`return` in both arms of an if/else) - must both survive. The
// shared-boundary case is already handled downstream, by
// mergeContiguousFindings.
func joinSnippets(lo, hi string) string {
	switch {
	case lo == "":
		return hi
	case hi == "":
		return lo
	default:
		return lo + "\n" + hi
	}
}

// dropNestedFindings removes findings whose line range lies inside another
// finding's range. Merging only joins findings that share a reason, so a block
// inside a dead region that got a different reason (e.g. an except handler
// entered only from a dead inner finally) would otherwise re-report lines the
// enclosing finding already covers.
func dropNestedFindings(findings []*DeadCodeFinding) []*DeadCodeFinding {
	slices.SortStableFunc(findings, func(a, b *DeadCodeFinding) int {
		return cmp.Or(cmp.Compare(a.StartLine, b.StartLine), cmp.Compare(b.EndLine, a.EndLine))
	})
	kept := findings[:0]
	maxEnd := 0
	for _, finding := range findings {
		if len(kept) > 0 && finding.EndLine <= maxEnd {
			continue
		}
		kept = append(kept, finding)
		maxEnd = finding.EndLine
	}
	return kept
}

// enclosingDeadTry returns the outermost try statement whose body starts with
// the block's first statement, or nil. The `try:` header is not a CFG
// statement, but the first body statement runs exactly when the try is
// entered, so that statement being dead means the whole try statement (header,
// handlers, else and finally) is dead.
func enclosingDeadTry(block *BasicBlock) *parser.Node {
	var try *parser.Node
	node := mustPythonNode(block.Statements[0])
	for node.Parent != nil && node.Parent.Type == parser.NodeTry &&
		len(node.Parent.Body) > 0 && node.Parent.Body[0] == node {
		try = node.Parent
		node = try
	}
	return try
}

// isOnlyNoOpStatements reports whether every statement is pass or a bare `;`.
func isOnlyNoOpStatements(block *BasicBlock) bool {
	if block == nil || len(block.Statements) == 0 {
		return false
	}
	classifier := pythonCFGClassifier{}
	for _, stmt := range block.Statements {
		if !classifier.IsNoOp(stmt) {
			return false
		}
	}
	return true
}

// isGeneratorMarkerBlock reports whether the block is nothing but a bare
// `yield`, the statement Python requires to make a function a generator. A
// `yield` carrying a value is excluded: it produces something no caller can
// ever receive, so it is genuine dead code.
func isGeneratorMarkerBlock(block *BasicBlock) bool {
	if block == nil || len(block.Statements) != 1 {
		return false
	}
	node, ok := pythonNode(block.Statements[0])
	if !ok || node.Type != parser.NodeYield {
		return false
	}
	value, ok := node.Value.(*parser.Node)
	return !ok || value == nil
}

// countScopeYields counts the yield expressions owned by one execution scope,
// i.e. the yields that decide whether that scope is a generator.
//
// A nested def/lambda/class suite is its own scope and is not counted, but the
// header Python evaluates at definition time — decorators, parameter defaults
// and annotations, class bases — runs in the enclosing scope, so `def inner(x=(yield 1))`
// makes the *enclosing* function a generator. By the same rule the root's own
// header belongs to whatever scope defines the root, not to the root itself.
func countScopeYields(root *parser.Node) int {
	if root == nil {
		return 0
	}

	count := 0
	var visit, visitChild func(node *parser.Node)

	visit = func(node *parser.Node) {
		if node == nil {
			return
		}
		if node.Type == parser.NodeYield || node.Type == parser.NodeYieldFrom {
			count++
		}
		for _, child := range parser.OrderedChildren(node, nil) {
			visitChild(child)
		}
	}

	visitChild = func(node *parser.Node) {
		if definesOwnScope(node) {
			// The suite runs elsewhere; only its header runs in this scope.
			for _, header := range definitionTimeExpressions(node) {
				visit(header)
			}
			return
		}
		visit(node)
	}

	if definesOwnScope(root) {
		// Skip root's own header: the defining scope evaluated it.
		for _, stmt := range root.Body {
			visitChild(stmt)
		}
		return count
	}
	visit(root)
	return count
}

// definesOwnScope reports whether the node's suite runs in a scope of its own.
func definesOwnScope(node *parser.Node) bool {
	if node == nil {
		return false
	}
	switch node.Type {
	case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef, parser.NodeClassDef, parser.NodeLambda:
		return true
	default:
		return false
	}
}

// definitionTimeExpressions returns the parts of a def/lambda/class header that
// Python evaluates in the enclosing scope when the definition is executed.
func definitionTimeExpressions(node *parser.Node) []*parser.Node {
	expressions := make([]*parser.Node, 0, len(node.Decorator)+len(node.Bases)+2*len(node.Args)+1)
	expressions = append(expressions, node.Decorator...)
	expressions = append(expressions, node.Bases...)
	for _, arg := range node.Args {
		if arg == nil {
			continue
		}
		if defaultValue, ok := arg.Value.(*parser.Node); ok {
			expressions = append(expressions, defaultValue)
		}
		expressions = append(expressions, arg.Right) // parameter annotation
	}
	return append(expressions, node.Right) // return annotation
}

func toCoreSeverity(severity SeverityLevel) corecfg.DeadCodeSeverity {
	switch severity {
	case SeverityLevelCritical:
		return corecfg.SeverityCritical
	case SeverityLevelWarning:
		return corecfg.SeverityWarning
	default:
		return corecfg.SeverityInfo
	}
}

func fromCoreSeverity(severity corecfg.DeadCodeSeverity) SeverityLevel {
	switch severity {
	case corecfg.SeverityCritical:
		return SeverityLevelCritical
	case corecfg.SeverityWarning:
		return SeverityLevelWarning
	default:
		return SeverityLevelInfo
	}
}
