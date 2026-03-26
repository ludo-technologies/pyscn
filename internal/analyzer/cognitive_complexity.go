package analyzer

import (
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// CognitiveComplexityResult holds the cognitive complexity score for a function
type CognitiveComplexityResult struct {
	// Total cognitive complexity score
	Total int

	// Function/method information
	FunctionName string
	StartLine    int
	EndLine      int
}

// CalculateCognitiveComplexity computes the cognitive complexity for a function node
// following the SonarSource specification.
//
// Rules:
//   - +1 (base increment) for: if, elif, else, for, while, except, break, continue, goto,
//     ternary (IfExp), and each boolean operator sequence change
//   - +nesting level (nesting increment) for: if, ternary (IfExp), for, while, except,
//     match/case, nested functions/lambdas (structures that increase nesting)
//   - Nesting level increases inside: if, elif, else, for, while, except, with,
//     match/case, lambda, nested function/class definitions
func CalculateCognitiveComplexity(funcNode *parser.Node) *CognitiveComplexityResult {
	if funcNode == nil {
		return &CognitiveComplexityResult{
			Total: 0,
		}
	}

	result := &CognitiveComplexityResult{
		FunctionName: funcNode.Name,
		StartLine:    funcNode.Location.StartLine,
		EndLine:      funcNode.Location.EndLine,
		Total:        0,
	}

	// Traverse the function body with initial nesting depth of 0
	for _, stmt := range funcNode.Body {
		traverseForCognitive(stmt, 0, result)
	}

	return result
}

// traverseForCognitive recursively traverses the AST to calculate cognitive complexity
func traverseForCognitive(node *parser.Node, nestingLevel int, result *CognitiveComplexityResult) {
	if node == nil {
		return
	}

	switch node.Type {
	case parser.NodeIf:
		// +1 base + nesting increment
		result.Total += 1 + nestingLevel
		// Traverse test expression at current nesting (no increment for condition itself)
		traverseForCognitive(node.Test, nestingLevel, result)
		// Traverse body at increased nesting
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel+1, result)
		}
		// Traverse elif/else clauses
		for _, elseNode := range node.Orelse {
			traverseForCognitive(elseNode, nestingLevel, result)
		}
		return

	case parser.NodeElifClause:
		// +1 base (no nesting increment for elif - it's at same structural level as if)
		result.Total += 1
		// Traverse test expression
		traverseForCognitive(node.Test, nestingLevel, result)
		// Traverse body at increased nesting
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel+1, result)
		}
		// Traverse further elif/else
		for _, elseNode := range node.Orelse {
			traverseForCognitive(elseNode, nestingLevel, result)
		}
		return

	case parser.NodeElseClause:
		// +1 base (no nesting increment for else)
		result.Total += 1
		// Traverse body at increased nesting
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel+1, result)
		}
		return

	case parser.NodeFor, parser.NodeAsyncFor:
		// +1 base + nesting increment
		result.Total += 1 + nestingLevel
		// Traverse iterator at current nesting
		traverseForCognitive(node.Iter, nestingLevel, result)
		// Traverse body at increased nesting
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel+1, result)
		}
		// for-else: else gets +1 base
		for _, elseNode := range node.Orelse {
			traverseForCognitive(elseNode, nestingLevel, result)
		}
		return

	case parser.NodeWhile:
		// +1 base + nesting increment
		result.Total += 1 + nestingLevel
		// Traverse test at current nesting
		traverseForCognitive(node.Test, nestingLevel, result)
		// Traverse body at increased nesting
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel+1, result)
		}
		// while-else
		for _, elseNode := range node.Orelse {
			traverseForCognitive(elseNode, nestingLevel, result)
		}
		return

	case parser.NodeTry:
		// try itself does not add to complexity, but increases nesting for body
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel, result)
		}
		// except handlers: +1 base + nesting increment each
		for _, handler := range node.Handlers {
			traverseForCognitive(handler, nestingLevel, result)
		}
		// else clause of try
		for _, elseNode := range node.Orelse {
			traverseForCognitive(elseNode, nestingLevel, result)
		}
		// finally block
		for _, finalNode := range node.Finalbody {
			traverseForCognitive(finalNode, nestingLevel, result)
		}
		return

	case parser.NodeExceptHandler:
		// +1 base + nesting increment
		result.Total += 1 + nestingLevel
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel+1, result)
		}
		return

	case parser.NodeWith, parser.NodeAsyncWith:
		// with does not add complexity but increases nesting for its body
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel+1, result)
		}
		return

	case parser.NodeMatch:
		// match itself: +1 base + nesting increment
		result.Total += 1 + nestingLevel
		// Each case is part of match body
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel+1, result)
		}
		return

	case parser.NodeMatchCase:
		// match_case doesn't add complexity itself (covered by match)
		// Traverse the case body
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel, result)
		}
		return

	case parser.NodeBreak, parser.NodeContinue:
		// +1 base (no nesting increment)
		result.Total += 1
		return

	case parser.NodeBoolOp:
		// Boolean operator sequences: +1 per operator
		// Nested BoolOps with same operator are part of one sequence (no extra increment)
		// But alternating operators each get +1
		countBoolOpComplexity(node, "", result)
		return

	case parser.NodeIfExp:
		// Ternary expression: +1 base + nesting increment
		result.Total += 1 + nestingLevel
		// Traverse sub-expressions (test, body, orelse are in Children/Test)
		traverseForCognitive(node.Test, nestingLevel+1, result)
		for _, child := range node.Body {
			traverseForCognitive(child, nestingLevel+1, result)
		}
		for _, child := range node.Orelse {
			traverseForCognitive(child, nestingLevel+1, result)
		}
		return

	case parser.NodeLambda:
		// Lambda increases nesting but does NOT add base increment
		// (lambda is not a control flow break, only a nesting structure)
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel+1, result)
		}
		for _, child := range node.Children {
			traverseForCognitive(child, nestingLevel+1, result)
		}
		return

	case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
		// Nested function definition: increases nesting (no base increment)
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel+1, result)
		}
		return

	case parser.NodeClassDef:
		// Nested class definition: increases nesting (no base increment)
		for _, bodyNode := range node.Body {
			traverseForCognitive(bodyNode, nestingLevel+1, result)
		}
		return
	}

	// Default: traverse all sub-nodes without changing nesting level
	traverseChildrenForCognitive(node, nestingLevel, result)
}

// traverseChildrenForCognitive traverses all children of a node
func traverseChildrenForCognitive(node *parser.Node, nestingLevel int, result *CognitiveComplexityResult) {
	for _, child := range node.Body {
		traverseForCognitive(child, nestingLevel, result)
	}
	for _, child := range node.Orelse {
		traverseForCognitive(child, nestingLevel, result)
	}
	for _, child := range node.Handlers {
		traverseForCognitive(child, nestingLevel, result)
	}
	for _, child := range node.Finalbody {
		traverseForCognitive(child, nestingLevel, result)
	}
	for _, child := range node.Children {
		traverseForCognitive(child, nestingLevel, result)
	}
	if node.Test != nil {
		traverseForCognitive(node.Test, nestingLevel, result)
	}
	if node.Iter != nil {
		traverseForCognitive(node.Iter, nestingLevel, result)
	}
	if node.Left != nil {
		traverseForCognitive(node.Left, nestingLevel, result)
	}
	if node.Right != nil {
		traverseForCognitive(node.Right, nestingLevel, result)
	}
	for _, arg := range node.Args {
		traverseForCognitive(arg, nestingLevel, result)
	}
	for _, kw := range node.Keywords {
		traverseForCognitive(kw, nestingLevel, result)
	}
	for _, target := range node.Targets {
		traverseForCognitive(target, nestingLevel, result)
	}
}

// countBoolOpComplexity counts complexity from boolean operator sequences.
// A sequence of the same operator (a and b and c) counts as +1.
// Alternating operators (a and b or c) counts as +2.
// Nested BoolOps are handled by checking if children are also BoolOps.
//
// Note: The parser's buildBoolOp stores operands only in Children (via AddChild),
// not in Left/Right. We only traverse Children to avoid double-counting.
func countBoolOpComplexity(node *parser.Node, parentOp string, result *CognitiveComplexityResult) {
	if node == nil || node.Type != parser.NodeBoolOp {
		return
	}

	currentOp := node.Op

	// If the operator changes from parent, or there is no parent, add +1
	if currentOp != parentOp {
		result.Total += 1
	}

	// Traverse Children only (parser stores BoolOp operands exclusively in Children)
	for _, child := range node.Children {
		if child != nil && child.Type == parser.NodeBoolOp {
			countBoolOpComplexity(child, currentOp, result)
		} else {
			// Non-BoolOp children are regular expressions, traverse normally
			traverseForCognitive(child, 0, result)
		}
	}
}
