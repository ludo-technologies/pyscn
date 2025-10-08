package analyzer

import (
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// NestingDepthResult holds the maximum nesting depth and related metadata for a function
type NestingDepthResult struct {
	// Maximum nesting depth found in the function
	MaxDepth int

	// Function/method information
	FunctionName string
	StartLine    int
	EndLine      int

	// Location of deepest nesting (line number)
	DeepestNestingLine int
}

// CalculateMaxNestingDepth computes the maximum nesting depth for a function node by traversing its AST and tracking depth through nested control structures
func CalculateMaxNestingDepth(funcNode *parser.Node) *NestingDepthResult {
	if funcNode == nil {
		return &NestingDepthResult{
			MaxDepth: 0,
		}
	}

	result := &NestingDepthResult{
		FunctionName: funcNode.Name,
		StartLine:    funcNode.Location.StartLine,
		EndLine:      funcNode.Location.EndLine,
		MaxDepth:     0,
	}

	// Start traversal from function body statements with initial depth of 0 We traverse the body directly to avoid counting the function itself as nesting
	for _, stmt := range funcNode.Body {
		traverseForNesting(stmt, 0, result)
	}

	return result
}

// traverseForNesting recursively traverses the AST to find maximum nesting depth
func traverseForNesting(node *parser.Node, currentDepth int, result *NestingDepthResult) {
	if node == nil {
		return
	}

	// Check if current node increases nesting level
	newDepth := currentDepth
	if isNestingNode(node) {
		newDepth = currentDepth + 1

		// Update max depth if this is deeper
		if newDepth > result.MaxDepth {
			result.MaxDepth = newDepth
			result.DeepestNestingLine = node.Location.StartLine
		}
	}

	// Traverse body statements (for compound statements)
	for _, bodyNode := range node.Body {
		traverseForNesting(bodyNode, newDepth, result)
	}

	// Traverse else/elif clauses
	for _, elseNode := range node.Orelse {
		traverseForNesting(elseNode, newDepth, result)
	}

	// Traverse exception handlers (for try statements)
	for _, handler := range node.Handlers {
		traverseForNesting(handler, newDepth, result)
	}

	// Traverse finally block
	for _, finalNode := range node.Finalbody {
		traverseForNesting(finalNode, newDepth, result)
	}

	// Traverse regular children (for expressions and other nodes)
	for _, child := range node.Children {
		traverseForNesting(child, newDepth, result)
	}

	// Traverse conditional test expressions (but don't increase depth)
	if node.Test != nil {
		traverseForNesting(node.Test, currentDepth, result)
	}

	// Traverse iterator expressions
	if node.Iter != nil {
		traverseForNesting(node.Iter, currentDepth, result)
	}

	// Handle comprehensions (list/dict/set comprehensions, generator expressions)
	// These also introduce nesting
	if isComprehensionNode(node) {
		// Traverse comprehension elements
		for _, arg := range node.Args {
			traverseForNesting(arg, newDepth, result)
		}
	}
}

// isNestingNode determines if a node type increases nesting depth
func isNestingNode(node *parser.Node) bool {
	if node == nil {
		return false
	}

	switch node.Type {
	case parser.NodeIf, parser.NodeFor, parser.NodeAsyncFor, parser.NodeWhile, parser.NodeWith, parser.NodeAsyncWith, parser.NodeTry, parser.NodeExceptHandler, parser.NodeMatch, parser.NodeMatchCase, parser.NodeElifClause,
		parser.NodeFunctionDef, parser.NodeAsyncFunctionDef, parser.NodeClassDef, parser.NodeLambda, parser.NodeListComp, parser.NodeSetComp, parser.NodeDictComp, parser.NodeGeneratorExp:
		// These nodes increase nesting depth
		return true
	case parser.NodeElseClause:
		// else doesn't add nesting, it's at same level
		return false
	default:
		return false
	}
}

// isComprehensionNode checks if a node is a comprehension type
func isComprehensionNode(node *parser.Node) bool {
	if node == nil {
		return false
	}

	switch node.Type {
	case parser.NodeListComp, parser.NodeSetComp, parser.NodeDictComp, parser.NodeGeneratorExp:
		return true
	default:
		return false
	}
}
