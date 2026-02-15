package analyzer

import (
	"fmt"
	"sort"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// LCOMResult holds LCOM4 (Lack of Cohesion of Methods) metrics for a class
type LCOMResult struct {
	// Core LCOM4 metric - number of connected components
	LCOM4 int

	// Class information
	ClassName string
	FilePath  string
	StartLine int
	EndLine   int

	// Method statistics
	TotalMethods    int // All methods found in class
	ExcludedMethods int // @staticmethod and @classmethod excluded

	// Instance variable count
	InstanceVariables int // Distinct self.xxx variables

	// Connected component details
	MethodGroups [][]string // Method names grouped by connected component

	// Risk assessment
	RiskLevel string // "low", "medium", "high"
}

// LCOMOptions configures LCOM analysis behavior
type LCOMOptions struct {
	LowThreshold    int // Default: 2 (LCOM4 <= 2 is low risk)
	MediumThreshold int // Default: 5 (LCOM4 3-5 is medium risk)
	ExcludePatterns []string
}

// DefaultLCOMOptions returns default LCOM analysis options
func DefaultLCOMOptions() *LCOMOptions {
	return &LCOMOptions{
		LowThreshold:    domain.DefaultLCOMLowThreshold,
		MediumThreshold: domain.DefaultLCOMMediumThreshold,
		ExcludePatterns: []string{},
	}
}

// LCOMAnalyzer analyzes class cohesion in Python code
type LCOMAnalyzer struct {
	options *LCOMOptions
}

// NewLCOMAnalyzer creates a new LCOM analyzer
func NewLCOMAnalyzer(options *LCOMOptions) *LCOMAnalyzer {
	if options == nil {
		options = DefaultLCOMOptions()
	}
	return &LCOMAnalyzer{
		options: options,
	}
}

// AnalyzeClasses analyzes LCOM4 for all classes in the given AST
func (a *LCOMAnalyzer) AnalyzeClasses(ast *parser.Node, filePath string) ([]*LCOMResult, error) {
	if ast == nil {
		return nil, fmt.Errorf("AST is nil")
	}

	// Collect all class definitions
	classes := a.collectClasses(ast)

	var results []*LCOMResult
	for _, classNode := range classes {
		result, err := a.analyzeClass(classNode, filePath)
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// analyzeClass computes LCOM4 for a single class using connected components
func (a *LCOMAnalyzer) analyzeClass(classNode *parser.Node, filePath string) (*LCOMResult, error) {
	if classNode.Type != parser.NodeClassDef {
		return nil, fmt.Errorf("node is not a class definition")
	}

	result := &LCOMResult{
		ClassName:    classNode.Name,
		FilePath:     filePath,
		StartLine:    classNode.Location.StartLine,
		EndLine:      classNode.Location.EndLine,
		MethodGroups: [][]string{},
	}

	// Step 1: Collect methods and their instance variable accesses
	methods, excluded := a.collectMethods(classNode)
	result.TotalMethods = len(methods) + excluded
	result.ExcludedMethods = excluded

	// Classes with 0 or 1 methods are trivially cohesive
	if len(methods) <= 1 {
		result.LCOM4 = 1
		if len(methods) == 1 {
			for name := range methods {
				result.MethodGroups = [][]string{{name}}
			}
		}
		result.RiskLevel = a.assessRiskLevel(1)
		return result, nil
	}

	// Step 2: Collect all distinct instance variables
	allVars := make(map[string]bool)
	for _, vars := range methods {
		for v := range vars {
			allVars[v] = true
		}
	}
	result.InstanceVariables = len(allVars)

	// Step 3: Compute connected components using Union-Find
	methodNames := make([]string, 0, len(methods))
	for name := range methods {
		methodNames = append(methodNames, name)
	}
	sort.Strings(methodNames) // Deterministic ordering

	// Initialize Union-Find
	parent := make(map[string]string, len(methodNames))
	rank := make(map[string]int, len(methodNames))
	for _, name := range methodNames {
		parent[name] = name
		rank[name] = 0
	}

	var find func(string) string
	find = func(x string) string {
		if parent[x] != x {
			parent[x] = find(parent[x]) // Path compression
		}
		return parent[x]
	}

	union := func(x, y string) {
		rx := find(x)
		ry := find(y)
		if rx == ry {
			return
		}
		// Union by rank
		if rank[rx] < rank[ry] {
			parent[rx] = ry
		} else if rank[rx] > rank[ry] {
			parent[ry] = rx
		} else {
			parent[ry] = rx
			rank[rx]++
		}
	}

	// Build variable -> methods mapping for efficient edge detection
	varToMethods := make(map[string][]string)
	for _, name := range methodNames {
		for v := range methods[name] {
			varToMethods[v] = append(varToMethods[v], name)
		}
	}

	// Union methods that share instance variables
	for _, methodList := range varToMethods {
		for i := 1; i < len(methodList); i++ {
			union(methodList[0], methodList[i])
		}
	}

	// Step 4: Count connected components and build groups
	components := make(map[string][]string) // root -> method names
	for _, name := range methodNames {
		root := find(name)
		components[root] = append(components[root], name)
	}

	result.LCOM4 = len(components)
	for _, group := range components {
		sort.Strings(group)
		result.MethodGroups = append(result.MethodGroups, group)
	}
	// Sort groups for deterministic output
	sort.Slice(result.MethodGroups, func(i, j int) bool {
		return result.MethodGroups[i][0] < result.MethodGroups[j][0]
	})

	result.RiskLevel = a.assessRiskLevel(result.LCOM4)
	return result, nil
}

// collectMethods extracts instance methods and their self.xxx variable accesses from a class.
// Returns a map of methodName -> set of instance variable names, and the count of excluded methods.
func (a *LCOMAnalyzer) collectMethods(classNode *parser.Node) (map[string]map[string]bool, int) {
	methods := make(map[string]map[string]bool)
	excluded := 0

	for _, node := range classNode.Body {
		if node == nil {
			continue
		}

		if node.Type != parser.NodeFunctionDef && node.Type != parser.NodeAsyncFunctionDef {
			continue
		}

		// Check decorators to exclude @classmethod and @staticmethod
		if a.isClassOrStaticMethod(node) {
			excluded++
			continue
		}

		// Extract self.xxx accesses from method body
		vars := make(map[string]bool)
		a.extractInstanceVars(node, vars)
		methods[node.Name] = vars
	}

	return methods, excluded
}

// isClassOrStaticMethod checks if a method is a @classmethod or @staticmethod.
// It first checks decorators, then falls back to parameter inspection since
// some parser implementations may not populate the Decorator field.
func (a *LCOMAnalyzer) isClassOrStaticMethod(funcNode *parser.Node) bool {
	// Check decorators first
	for _, decorator := range funcNode.Decorator {
		if decorator == nil {
			continue
		}
		name := a.getDecoratorName(decorator)
		if name == "classmethod" || name == "staticmethod" {
			return true
		}
	}
	// Fallback: if the first parameter is not "self", it's a classmethod/staticmethod.
	// - @staticmethod: no self/cls parameter (or arbitrary name like "x")
	// - @classmethod: first parameter is "cls"
	if len(funcNode.Args) == 0 {
		return true // No parameters at all → staticmethod
	}
	firstParam := funcNode.Args[0]
	return firstParam.Name != "self"
}

// getDecoratorName extracts the decorator name from a decorator node
func (a *LCOMAnalyzer) getDecoratorName(decorator *parser.Node) string {
	// Check Name field first (used by some paths)
	if decorator.Name != "" {
		return decorator.Name
	}
	// Check Value field (set by buildDecorator)
	if decorator.Value != nil {
		if nameNode, ok := decorator.Value.(*parser.Node); ok {
			if nameNode.Type == parser.NodeName {
				return nameNode.Name
			}
			// For decorator calls like @decorator(args), extract function name
			if nameNode.Type == parser.NodeCall {
				if nameNode.Value != nil {
					if funcNode, ok := nameNode.Value.(*parser.Node); ok && funcNode.Type == parser.NodeName {
						return funcNode.Name
					}
				}
				// Also check Name field on Call node
				if nameNode.Name != "" {
					return nameNode.Name
				}
			}
		}
	}
	return ""
}

// extractInstanceVars walks a method's AST to find all self.xxx attribute accesses
func (a *LCOMAnalyzer) extractInstanceVars(methodNode *parser.Node, vars map[string]bool) {
	a.walkNode(methodNode, func(node *parser.Node) bool {
		if node.Type == parser.NodeAttribute {
			// Check if this is self.xxx
			if a.isSelfAccess(node) && node.Name != "" {
				vars[node.Name] = true
			}
		}
		return true
	})
}

// isSelfAccess checks if an attribute node represents self.xxx access
func (a *LCOMAnalyzer) isSelfAccess(attrNode *parser.Node) bool {
	// The object (self) can be in Value or Left field
	if attrNode.Value != nil {
		if nameNode, ok := attrNode.Value.(*parser.Node); ok {
			if nameNode.Type == parser.NodeName && nameNode.Name == "self" {
				return true
			}
		}
	}
	if attrNode.Left != nil {
		if attrNode.Left.Type == parser.NodeName && attrNode.Left.Name == "self" {
			return true
		}
	}
	return false
}

// assessRiskLevel determines the risk level based on LCOM4 value
func (a *LCOMAnalyzer) assessRiskLevel(lcom4 int) string {
	switch {
	case lcom4 <= a.options.LowThreshold:
		return string(domain.RiskLevelLow)
	case lcom4 <= a.options.MediumThreshold:
		return string(domain.RiskLevelMedium)
	default:
		return string(domain.RiskLevelHigh)
	}
}

// collectClasses collects all class definition nodes from the AST
func (a *LCOMAnalyzer) collectClasses(ast *parser.Node) []*parser.Node {
	var classes []*parser.Node
	a.walkNode(ast, func(node *parser.Node) bool {
		if node.Type == parser.NodeClassDef && node.Name != "" {
			classes = append(classes, node)
		}
		return true
	})
	return classes
}

// walkNode performs a depth-first traversal of the AST
func (a *LCOMAnalyzer) walkNode(node *parser.Node, visitor func(*parser.Node) bool) {
	if node == nil || !visitor(node) {
		return
	}

	for _, child := range node.Children {
		a.walkNode(child, visitor)
	}
	for _, child := range node.Body {
		a.walkNode(child, visitor)
	}
	for _, child := range node.Targets {
		a.walkNode(child, visitor)
	}
	for _, child := range node.Args {
		a.walkNode(child, visitor)
	}
	for _, child := range node.Keywords {
		a.walkNode(child, visitor)
	}
	for _, child := range node.Orelse {
		a.walkNode(child, visitor)
	}
	for _, child := range node.Finalbody {
		a.walkNode(child, visitor)
	}
	for _, child := range node.Handlers {
		a.walkNode(child, visitor)
	}

	if node.Value != nil {
		if valueNode, ok := node.Value.(*parser.Node); ok {
			a.walkNode(valueNode, visitor)
		}
	}
	a.walkNode(node.Left, visitor)
	a.walkNode(node.Right, visitor)
	a.walkNode(node.Test, visitor)
	a.walkNode(node.Iter, visitor)
}

// CalculateLCOM is a convenience function that creates an analyzer with defaults and runs it
func CalculateLCOM(ast *parser.Node, filePath string) ([]*LCOMResult, error) {
	analyzer := NewLCOMAnalyzer(nil)
	return analyzer.AnalyzeClasses(ast, filePath)
}

// CalculateLCOMWithConfig creates an analyzer with custom options and runs it
func CalculateLCOMWithConfig(ast *parser.Node, filePath string, options *LCOMOptions) ([]*LCOMResult, error) {
	analyzer := NewLCOMAnalyzer(options)
	return analyzer.AnalyzeClasses(ast, filePath)
}
