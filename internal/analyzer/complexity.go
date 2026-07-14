package analyzer

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// ComplexityResult holds cyclomatic complexity metrics for a function or method
type ComplexityResult struct {
	// McCabe cyclomatic complexity
	Complexity int

	// Raw CFG metrics
	Edges               int
	Nodes               int
	ConnectedComponents int

	// Function/method information
	FunctionName string
	StartLine    int
	StartCol     int
	EndLine      int

	// Nesting depth
	NestingDepth        int
	CognitiveComplexity int

	// Decision points breakdown
	IfStatements      int
	LoopStatements    int
	ExceptionHandlers int
	SwitchCases       int

	// SLOC is the source lines of code for this function.
	SLOC int

	// Risk assessment based on complexity thresholds
	RiskLevel string // "low", "medium", "high"
}

// Interface methods for reporter compatibility

func (cr *ComplexityResult) GetComplexity() int {
	return cr.Complexity
}

func (cr *ComplexityResult) GetFunctionName() string {
	return cr.FunctionName
}

func (cr *ComplexityResult) GetRiskLevel() string {
	return cr.RiskLevel
}

func (cr *ComplexityResult) GetDetailedMetrics() map[string]int {
	return map[string]int{
		"nodes":                cr.Nodes,
		"edges":                cr.Edges,
		"cognitive_complexity": cr.CognitiveComplexity,
		"if_statements":        cr.IfStatements,
		"loop_statements":      cr.LoopStatements,
		"exception_handlers":   cr.ExceptionHandlers,
		"switch_cases":         cr.SwitchCases,
	}
}

// String returns a human-readable representation of the complexity result
func (cr *ComplexityResult) String() string {
	return fmt.Sprintf("Function: %s, Complexity: %d, Risk: %s",
		cr.FunctionName, cr.Complexity, cr.RiskLevel)
}

// complexityVisitor implements CFGVisitor to count edges and nodes
type complexityVisitor struct {
	edgeCount         int
	nodeCount         int
	decisionPoints    map[*BasicBlock]int // Track decision points per block
	loopStatements    int
	exceptionHandlers int
	switchCases       int
}

// VisitBlock counts nodes and analyzes decision points
func (cv *complexityVisitor) VisitBlock(block *BasicBlock) bool {
	if block == nil {
		return true
	}

	// Count all blocks except entry/exit for accurate complexity
	if !block.IsEntry && !block.IsExit {
		cv.nodeCount++
	}

	return true
}

// VisitEdge counts edges and categorizes decision points
func (cv *complexityVisitor) VisitEdge(edge *Edge) bool {
	if edge == nil {
		return true
	}

	cv.edgeCount++

	// Count decision points accurately by source block
	// A decision point is a block with multiple outgoing edges
	if edge.From != nil {
		if cv.decisionPoints == nil {
			cv.decisionPoints = make(map[*BasicBlock]int)
		}

		switch edge.Type {
		case EdgeCondTrue, EdgeCondFalse:
			// Mark this block as having conditional edges
			// We only count the block once, regardless of number of edges
			cv.decisionPoints[edge.From] = 1
		case EdgeLoop:
			cv.loopStatements++
		case EdgeException:
			cv.exceptionHandlers++
		}
	}

	return true
}

// CalculateComplexity computes McCabe cyclomatic complexity for a CFG using default thresholds
func CalculateComplexity(cfg *CFG) *ComplexityResult {
	defaultConfig := config.DefaultConfig()
	return CalculateComplexityWithConfig(cfg, &defaultConfig.Complexity)
}

// CalculateComplexityWithConfig computes McCabe cyclomatic complexity using provided configuration
func CalculateComplexityWithConfig(cfg *CFG, complexityConfig *config.ComplexityConfig) *ComplexityResult {
	if cfg == nil {
		return &ComplexityResult{
			Complexity: 0,
			RiskLevel:  "low",
		}
	}

	visitor := &complexityVisitor{
		decisionPoints: make(map[*BasicBlock]int),
	}
	cfg.Walk(visitor)

	// Primary method: count decision points + 1
	// This is more reliable for CFGs with entry/exit nodes
	astMetrics, hasASTMetrics := calculateASTComplexityMetrics(complexitySourceNode(cfg))
	reportedMetrics := resolveComplexityMetrics(visitor, astMetrics, hasASTMetrics)
	decisionPoints := countDecisionPoints(visitor, reportedMetrics, hasASTMetrics)
	complexity := decisionPoints + 1

	// Ensure minimum complexity of 1 for any function
	if complexity < 1 {
		complexity = 1
	}

	// Calculate nesting depth and location from the original AST node if available.
	// The same node feeds CalculateCognitiveComplexity above via complexitySourceNode.
	nestingDepth := 0
	startLine := 0
	startCol := 0
	endLine := 0
	if sourceNode := complexitySourceNode(cfg); sourceNode != nil {
		nestingDepth = CalculateMaxNestingDepth(sourceNode).MaxDepth

		startLine = sourceNode.Location.StartLine
		startCol = sourceNode.Location.StartCol
		endLine = sourceNode.Location.EndLine
	}

	result := &ComplexityResult{
		Complexity:          complexity,
		Edges:               visitor.edgeCount,
		Nodes:               visitor.nodeCount,
		ConnectedComponents: 1,
		FunctionName:        cfg.Name,
		StartLine:           startLine,
		StartCol:            startCol,
		EndLine:             endLine,
		NestingDepth:        nestingDepth,
		CognitiveComplexity: reportedMetrics.CognitiveComplexity,
		IfStatements:        reportedMetrics.IfStatements,
		LoopStatements:      reportedMetrics.LoopStatements,
		ExceptionHandlers:   reportedMetrics.ExceptionHandlers,
		SwitchCases:         reportedMetrics.SwitchCases,
		RiskLevel:           complexityConfig.AssessRiskLevel(complexity, reportedMetrics.CognitiveComplexity, nestingDepth),
	}

	return result
}

// countDecisionPoints counts the number of decision points in the CFG
func countDecisionPoints(visitor *complexityVisitor, metrics astComplexityMetrics, hasASTMetrics bool) int {
	// Decision points are nodes that have multiple outgoing edges
	// For McCabe complexity, each decision point adds 1 to complexity

	// Count actual conditional decisions (unique blocks with conditional edges)
	conditionalDecisions := len(visitor.decisionPoints)
	if hasASTMetrics && metrics.MatchStatements > 0 {
		conditionalDecisions -= metrics.MatchStatements
		if conditionalDecisions < 0 {
			conditionalDecisions = 0
		}
	}

	return conditionalDecisions + metrics.ExceptionHandlers + metrics.SwitchCases
}

type astComplexityMetrics struct {
	CognitiveComplexity int
	IfStatements        int
	LoopStatements      int
	ExceptionHandlers   int
	MatchStatements     int
	SwitchCases         int
}

func complexitySourceNode(cfg *CFG) *parser.Node {
	if cfg == nil {
		return nil
	}
	if cfg.FunctionNode != nil {
		return cfg.FunctionNode
	}
	return cfg.ModuleNode
}

func resolveComplexityMetrics(visitor *complexityVisitor, astMetrics astComplexityMetrics, hasASTMetrics bool) astComplexityMetrics {
	if hasASTMetrics {
		return astMetrics
	}
	return astComplexityMetrics{
		IfStatements:      len(visitor.decisionPoints),
		LoopStatements:    visitor.loopStatements,
		ExceptionHandlers: visitor.exceptionHandlers,
		SwitchCases:       visitor.switchCases,
	}
}

func calculateASTComplexityMetrics(root *parser.Node) (astComplexityMetrics, bool) {
	if root == nil {
		return astComplexityMetrics{}, false
	}

	metrics := astComplexityMetrics{}
	metrics.CognitiveComplexity = CalculateCognitiveComplexity(root).Total
	for _, stmt := range root.Body {
		collectASTStatementMetrics(stmt, &metrics)
	}
	return metrics, true
}

func collectASTStatementMetrics(node *parser.Node, metrics *astComplexityMetrics) {
	if node == nil {
		return
	}

	switch node.Type {
	case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef, parser.NodeClassDef:
		return
	case parser.NodeMatch:
		if len(node.Body) > 0 {
			metrics.MatchStatements++
		}
	case parser.NodeIf, parser.NodeElifClause:
		metrics.IfStatements++
	case parser.NodeFor, parser.NodeAsyncFor, parser.NodeWhile:
		metrics.LoopStatements++
	case parser.NodeExceptHandler:
		metrics.ExceptionHandlers++
	case parser.NodeMatchCase:
		metrics.SwitchCases++
	}

	for _, child := range node.GetChildren() {
		collectASTStatementMetrics(child, metrics)
	}
}

// CalculateFileComplexity calculates complexity for all functions in a collection of CFGs
func CalculateFileComplexity(cfgs []*CFG) []*ComplexityResult {
	defaultConfig := config.DefaultConfig()
	return CalculateFileComplexityWithConfig(cfgs, &defaultConfig.Complexity)
}

// CalculateFileComplexityWithConfig calculates complexity using provided configuration
func CalculateFileComplexityWithConfig(cfgs []*CFG, complexityConfig *config.ComplexityConfig) []*ComplexityResult {
	results := make([]*ComplexityResult, 0, len(cfgs))

	for _, cfg := range cfgs {
		if cfg != nil {
			result := CalculateComplexityWithConfig(cfg, complexityConfig)

			// Only include results that should be reported according to config
			if complexityConfig.ShouldReport(result.Complexity) {
				results = append(results, result)
			}
		}
	}

	return results
}

// AggregateComplexity calculates aggregate metrics for multiple functions
type AggregateComplexity struct {
	TotalFunctions    int
	AverageComplexity float64
	MaxComplexity     int
	MinComplexity     int
	HighRiskCount     int
	MediumRiskCount   int
	LowRiskCount      int
}

// CalculateAggregateComplexity computes aggregate complexity metrics
func CalculateAggregateComplexity(results []*ComplexityResult) *AggregateComplexity {
	if len(results) == 0 {
		return &AggregateComplexity{}
	}

	agg := &AggregateComplexity{
		TotalFunctions: len(results),
		MinComplexity:  results[0].Complexity,
		MaxComplexity:  results[0].Complexity,
	}

	totalComplexity := 0

	for _, result := range results {
		totalComplexity += result.Complexity

		if result.Complexity > agg.MaxComplexity {
			agg.MaxComplexity = result.Complexity
		}
		if result.Complexity < agg.MinComplexity {
			agg.MinComplexity = result.Complexity
		}

		switch result.RiskLevel {
		case "high":
			agg.HighRiskCount++
		case "medium":
			agg.MediumRiskCount++
		case "low":
			agg.LowRiskCount++
		}
	}

	agg.AverageComplexity = float64(totalComplexity) / float64(len(results))

	return agg
}
