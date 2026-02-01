package analyzer

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// ConstructorAnalyzer detects constructor over-injection anti-pattern
type ConstructorAnalyzer struct {
	threshold int
}

// NewConstructorAnalyzer creates a new constructor analyzer
func NewConstructorAnalyzer(threshold int) *ConstructorAnalyzer {
	if threshold <= 0 {
		threshold = domain.DefaultDIConstructorParamThreshold
	}
	return &ConstructorAnalyzer{
		threshold: threshold,
	}
}

// Analyze detects constructor over-injection in the given AST
func (a *ConstructorAnalyzer) Analyze(ast *parser.Node, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	// Find all class definitions
	classes := ast.FindByType(parser.NodeClassDef)

	for _, class := range classes {
		classFindings := a.analyzeClass(class, filePath)
		findings = append(findings, classFindings...)
	}

	return findings
}

// analyzeClass analyzes a single class for constructor over-injection
func (a *ConstructorAnalyzer) analyzeClass(classNode *parser.Node, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	// Find __init__ method
	initMethod := a.findInitMethod(classNode)
	if initMethod == nil {
		return findings
	}

	// Count parameters (excluding self)
	paramCount := a.countParameters(initMethod)
	if paramCount > a.threshold {
		finding := domain.DIAntipatternFinding{
			Type:       domain.DIAntipatternConstructorOverInjection,
			Severity:   domain.DIAntipatternSeverityWarning,
			ClassName:  classNode.Name,
			MethodName: "__init__",
			Location: domain.SourceLocation{
				FilePath:  filePath,
				StartLine: initMethod.Location.StartLine,
				EndLine:   initMethod.Location.EndLine,
				StartCol:  initMethod.Location.StartCol,
				EndCol:    initMethod.Location.EndCol,
			},
			Description: fmt.Sprintf("Constructor has %d parameters (threshold: %d)", paramCount, a.threshold),
			Suggestion:  "Consider using a builder pattern, parameter object, or splitting responsibilities",
			Details: map[string]interface{}{
				"parameter_count": paramCount,
				"threshold":       a.threshold,
				"parameters":      a.getParameterNames(initMethod),
			},
		}
		findings = append(findings, finding)
	}

	return findings
}

// findInitMethod finds the __init__ method in a class
func (a *ConstructorAnalyzer) findInitMethod(classNode *parser.Node) *parser.Node {
	// Check body for __init__ method
	for _, node := range classNode.Body {
		if node != nil && (node.Type == parser.NodeFunctionDef || node.Type == parser.NodeAsyncFunctionDef) {
			if node.Name == "__init__" {
				return node
			}
		}
	}
	return nil
}

// countParameters counts the number of parameters excluding 'self'
func (a *ConstructorAnalyzer) countParameters(funcNode *parser.Node) int {
	count := 0
	for _, arg := range funcNode.Args {
		if arg != nil && arg.Type == parser.NodeArg {
			// Skip 'self' parameter
			if arg.Name != "self" && arg.Name != "cls" {
				count++
			}
		}
	}
	return count
}

// getParameterNames returns the names of all parameters excluding 'self'
func (a *ConstructorAnalyzer) getParameterNames(funcNode *parser.Node) []string {
	var params []string
	for _, arg := range funcNode.Args {
		if arg != nil && arg.Type == parser.NodeArg {
			if arg.Name != "self" && arg.Name != "cls" {
				params = append(params, arg.Name)
			}
		}
	}
	return params
}
