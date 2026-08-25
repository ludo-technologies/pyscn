package analyzer

import (
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// NestingDepthResult holds maximum nesting depth and metadata for an execution scope.
type NestingDepthResult struct {
	// Maximum nesting depth found in the function
	MaxDepth int

	// Historical function/method metadata fields.
	FunctionName string
	StartLine    int
	EndLine      int

	// Location of deepest nesting (line number)
	DeepestNestingLine int
}

// CalculateMaxNestingDepth traverses one owned execution scope and tracks depth
// through its nested control structures.
func CalculateMaxNestingDepth(scopeNode *parser.Node) *NestingDepthResult {
	if scopeNode == nil {
		return &NestingDepthResult{
			MaxDepth: 0,
		}
	}

	result := &NestingDepthResult{
		FunctionName: scopeNode.Name,
		StartLine:    scopeNode.Location.StartLine,
		EndLine:      scopeNode.Location.EndLine,
		MaxDepth:     0,
	}

	// Traverse the body directly so the owning scope is not counted as nesting.
	for _, stmt := range scopeNode.Body {
		traverseForNesting(stmt, 0, result)
	}

	return result
}

// traverseForNesting recursively traverses the AST to find maximum nesting depth
func traverseForNesting(node *parser.Node, currentDepth int, result *NestingDepthResult) {
	if node == nil {
		return
	}
	if node.Type == parser.NodeFunctionDef || node.Type == parser.NodeAsyncFunctionDef || node.Type == parser.NodeClassDef {
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
		parser.NodeLambda, parser.NodeListComp, parser.NodeSetComp, parser.NodeDictComp, parser.NodeGeneratorExp:
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
