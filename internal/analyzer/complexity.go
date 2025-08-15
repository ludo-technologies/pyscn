package analyzer

import (
	"fmt"
	"github.com/pyqol/pyqol/internal/config"
)

// ComplexityResult holds cyclomatic complexity metrics for a function or method
type ComplexityResult struct {
	// McCabe cyclomatic complexity
	Complexity int

	// Raw CFG metrics
	Edges int
	Nodes int
	ConnectedComponents int

	// Function/method information
	FunctionName string
	StartLine    int
	EndLine      int

	// Decision points breakdown
	IfStatements      int
	LoopStatements    int
	ExceptionHandlers int
	SwitchCases       int

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

func (cr *ComplexityResult) GetNodes() int {
	return cr.Nodes
}

func (cr *ComplexityResult) GetEdges() int {
	return cr.Edges
}

func (cr *ComplexityResult) GetIfStatements() int {
	return cr.IfStatements
}

func (cr *ComplexityResult) GetLoopStatements() int {
	return cr.LoopStatements
}

func (cr *ComplexityResult) GetExceptionHandlers() int {
	return cr.ExceptionHandlers
}

func (cr *ComplexityResult) GetSwitchCases() int {
	return cr.SwitchCases
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
	ifStatements      int
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

	// Count decision points by edge type
	switch edge.Type {
	case EdgeCondTrue, EdgeCondFalse:
		cv.ifStatements++
	case EdgeLoop:
		cv.loopStatements++
	case EdgeException:
		cv.exceptionHandlers++
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

	visitor := &complexityVisitor{}
	cfg.Walk(visitor)

	// Primary method: count decision points + 1
	// This is more reliable for CFGs with entry/exit nodes
	decisionPoints := countDecisionPoints(visitor)
	complexity := decisionPoints + 1

	// Ensure minimum complexity of 1 for any function
	if complexity < 1 {
		complexity = 1
	}

	result := &ComplexityResult{
		Complexity:          complexity,
		Edges:              visitor.edgeCount,
		Nodes:              visitor.nodeCount,
		ConnectedComponents: 1,
		FunctionName:       cfg.Name,
		IfStatements:       visitor.ifStatements / 2, // Divide by 2 since true/false edges are counted separately
		LoopStatements:     visitor.loopStatements,
		ExceptionHandlers:  visitor.exceptionHandlers,
		SwitchCases:        visitor.switchCases,
		RiskLevel:          complexityConfig.AssessRiskLevel(complexity),
	}

	return result
}

// countDecisionPoints counts the number of decision points in the CFG
func countDecisionPoints(visitor *complexityVisitor) int {
	// Decision points are nodes that have multiple outgoing edges
	// For McCabe complexity, each decision point adds 1 to complexity
	
	// Count conditional pairs (if/else) as one decision point each
	conditionalDecisions := visitor.ifStatements / 2
	
	// Add other decision types
	// Note: loops without conditions are just jumps, not decisions
	// But loops with conditions are already counted in conditionals
	return conditionalDecisions + visitor.exceptionHandlers + visitor.switchCases
}

// assessRiskLevel determines risk level based on default complexity thresholds
// Deprecated: Use config.ComplexityConfig.AssessRiskLevel instead
func assessRiskLevel(complexity int) string {
	if complexity <= 9 {
		return "low"
	} else if complexity <= 19 {
		return "medium"
	}
	return "high"
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
	TotalFunctions   int
	AverageComplexity float64
	MaxComplexity    int
	MinComplexity    int
	HighRiskCount    int
	MediumRiskCount  int
	LowRiskCount     int
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