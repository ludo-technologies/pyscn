package parser

import (
	"fmt"
	"strconv"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// ASTBuilder converts tree-sitter parse trees to internal AST representation
type ASTBuilder struct {
	source []byte
}

// NewASTBuilder creates a new AST builder
func NewASTBuilder(source []byte) *ASTBuilder {
	return &ASTBuilder{
		source: source,
	}
}

// Helper to extract body from block and set parent references
func (b *ASTBuilder) extractBlockBody(blockNode *Node, parent *Node) []*Node {
	if blockNode == nil {
		return nil
	}
	if blockNode.Type == "block" {
		// Update parent references when extracting from block
		for _, stmt := range blockNode.Body {
			if stmt != nil {
				stmt.Parent = parent
			}
		}
		return blockNode.Body
	}
	// Not a block, return as single element
	blockNode.Parent = parent
	return []*Node{blockNode}
}

// Build converts a tree-sitter tree to internal AST
func (b *ASTBuilder) Build(tree *sitter.Tree) (*Node, error) {
	if tree == nil {
		return nil, fmt.Errorf("tree is nil")
	}

	rootNode := tree.RootNode()
	if rootNode == nil {
		return nil, fmt.Errorf("root node is nil")
	}

	ast := b.buildNode(rootNode)
	return ast, nil
}

// buildNode recursively builds AST nodes from tree-sitter nodes
func (b *ASTBuilder) buildNode(tsNode *sitter.Node) *Node {
	if tsNode == nil {
		return nil
	}

	nodeType := tsNode.Type()

	// Create appropriate AST node based on tree-sitter node type
	switch nodeType {
	case "module":
		return b.buildModule(tsNode)
	case "function_definition":
		return b.buildFunctionDef(tsNode)
	case "class_definition":
		return b.buildClassDef(tsNode)
	case "if_statement":
		return b.buildIfStatement(tsNode)
	case "elif_clause":
		return b.buildElifClause(tsNode)
	case "else_clause":
		return b.buildElseClause(tsNode)
	case "for_statement":
		return b.buildForStatement(tsNode)
	case "while_statement":
		return b.buildWhileStatement(tsNode)
	case "with_statement":
		return b.buildWithStatement(tsNode)
	case "try_statement":
		return b.buildTryStatement(tsNode)
	case "match_statement":
		return b.buildMatchStatement(tsNode)
	case "return_statement":
		return b.buildReturnStatement(tsNode)
	case "delete_statement":
		return b.buildDeleteStatement(tsNode)
	case "raise_statement":
		return b.buildRaiseStatement(tsNode)
	case "assert_statement":
		return b.buildAssertStatement(tsNode)
	case "import_statement":
		return b.buildImportStatement(tsNode)
	case "import_from_statement":
		return b.buildImportFromStatement(tsNode)
	case "global_statement":
		return b.buildGlobalStatement(tsNode)
	case "nonlocal_statement":
		return b.buildNonlocalStatement(tsNode)
	case "expression_statement":
		return b.buildExpressionStatement(tsNode)
	case "assignment":
		return b.buildAssignment(tsNode)
	case "augmented_assignment":
		return b.buildAugmentedAssignment(tsNode)
	case "pass_statement":
		return b.buildPassStatement(tsNode)
	case "break_statement":
		return b.buildBreakStatement(tsNode)
	case "continue_statement":
		return b.buildContinueStatement(tsNode)
	case "decorated_definition":
		return b.buildDecoratedDefinition(tsNode)

	// Expressions
	case "binary_operator":
		return b.buildBinaryOp(tsNode)
	case "unary_operator":
		return b.buildUnaryOp(tsNode)
	case "boolean_operator":
		return b.buildBoolOp(tsNode)
	case "named_expression":
		return b.buildNamedExpr(tsNode)
	case "comparison_operator":
		return b.buildCompare(tsNode)
	case "conditional_expression":
		return b.buildIfExp(tsNode)
	case "lambda":
		return b.buildLambda(tsNode)
	case "call":
		return b.buildCall(tsNode)
	case "attribute":
		return b.buildAttribute(tsNode)
	case "subscript":
		return b.buildSubscript(tsNode)
	case "slice":
		return b.buildSlice(tsNode)
	case "list":
		return b.buildList(tsNode)
	case "tuple":
		return b.buildTuple(tsNode)
	case "dictionary":
		return b.buildDict(tsNode)
	case "set":
		return b.buildSet(tsNode)
	case "list_comprehension":
		return b.buildListComp(tsNode)
	case "dictionary_comprehension":
		return b.buildDictComp(tsNode)
	case "set_comprehension":
		return b.buildSetComp(tsNode)
	case "generator_expression":
		return b.buildGeneratorExp(tsNode)
	case "yield":
		return b.buildYield(tsNode)
	case "yield_from":
		return b.buildYieldFrom(tsNode)
	case "await":
		return b.buildAwait(tsNode)
	case "identifier":
		return b.buildName(tsNode)
	case "integer", "float", "string", "concatenated_string", "true", "false", "none":
		return b.buildConstant(tsNode)
	case "formatted_string", "interpolation":
		return b.buildFormattedString(tsNode)

	// Handle compound statements
	case "block":
		return b.buildBlock(tsNode)

	default:
		// For unhandled types, create a generic node with children
		node := NewNode(NodeType(nodeType))
		node.Location = b.getLocation(tsNode)

		childCount := int(tsNode.ChildCount())
		for i := 0; i < childCount; i++ {
			child := tsNode.Child(i)
			if child != nil && !b.isTrivia(child) {
				if childNode := b.buildNode(child); childNode != nil {
					node.AddChild(childNode)
				}
			}
		}
		return node
	}
}

// buildModule builds a module node
func (b *ASTBuilder) buildModule(tsNode *sitter.Node) *Node {
	node := NewNode(NodeModule)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && !b.isTrivia(child) {
			if stmt := b.buildNode(child); stmt != nil {
				node.AddToBody(stmt)
			}
		}
	}

	return node
}

// buildFunctionDef builds a function definition node
func (b *ASTBuilder) buildFunctionDef(tsNode *sitter.Node) *Node {
	node := NewNode(NodeFunctionDef)
	node.Location = b.getLocation(tsNode)

	// Check if it's async
	if b.hasChildOfType(tsNode, "async") {
		node.Type = NodeAsyncFunctionDef
	}

	// Get function name
	if nameNode := b.getChildByFieldName(tsNode, "name"); nameNode != nil {
		node.Name = b.getNodeText(nameNode)
	}

	// Get parameters
	if paramsNode := b.getChildByFieldName(tsNode, "parameters"); paramsNode != nil {
		node.Args = b.buildParameters(paramsNode)
		// Set parent for args
		for _, arg := range node.Args {
			if arg != nil {
				arg.Parent = node
			}
		}
	}

	// Get body
	if bodyNode := b.getChildByFieldName(tsNode, "body"); bodyNode != nil {
		if body := b.buildNode(bodyNode); body != nil {
			node.Body = b.extractBlockBody(body, node)
		}
	}

	// Get return type annotation if present
	if returnType := b.getChildByFieldName(tsNode, "return_type"); returnType != nil {
		// Store return type in Value field
		node.Value = b.getNodeText(returnType)
	}

	return node
}

// buildClassDef builds a class definition node
func (b *ASTBuilder) buildClassDef(tsNode *sitter.Node) *Node {
	node := NewNode(NodeClassDef)
	node.Location = b.getLocation(tsNode)

	// Get class name
	if nameNode := b.getChildByFieldName(tsNode, "name"); nameNode != nil {
		node.Name = b.getNodeText(nameNode)
	}

	// Get base classes
	if superclasses := b.getChildByFieldName(tsNode, "superclasses"); superclasses != nil {
		node.Bases = b.buildArgumentList(superclasses)
	}

	// Get body
	if bodyNode := b.getChildByFieldName(tsNode, "body"); bodyNode != nil {
		if body := b.buildNode(bodyNode); body != nil {
			node.Body = b.extractBlockBody(body, node)
		}
	}

	return node
}

// buildIfStatement builds an if statement node
func (b *ASTBuilder) buildIfStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeIf)
	node.Location = b.getLocation(tsNode)

	// Get condition
	if condition := b.getChildByFieldName(tsNode, "condition"); condition != nil {
		node.Test = b.buildNode(condition)
	}

	// Get consequence
	if consequence := b.getChildByFieldName(tsNode, "consequence"); consequence != nil {
		if body := b.buildNode(consequence); body != nil {
			node.Body = b.extractBlockBody(body, node)
		}
	}

	// Get alternatives (else/elif) - there may be multiple with the same field name
	// Tree-sitter can have both elif_clause and else_clause as "alternative"
	var elifNode *Node
	var elseNode *Node

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && tsNode.FieldNameForChild(i) == "alternative" {
			alt := b.buildNode(child)
			if alt != nil {
				if alt.Type == NodeIf || alt.Type == NodeElifClause {
					elifNode = alt
				} else if alt.Type == NodeElseClause {
					elseNode = alt
				}
			}
		}
	}

	// If we have an elif, attach the else to it
	if elifNode != nil {
		if elseNode != nil {
			// Attach else to the elif chain
			b.attachElseToElifChain(elifNode, elseNode)
		}
		node.Orelse = []*Node{elifNode}
	} else if elseNode != nil {
		// Just an else clause, no elif
		node.Orelse = b.extractBlockBody(elseNode, node)
	}

	return node
}

// attachElseToElifChain attaches an else clause to the end of an elif chain
func (b *ASTBuilder) attachElseToElifChain(elifNode *Node, elseNode *Node) {
	current := elifNode
	// Find the last elif in the chain
	for len(current.Orelse) > 0 && current.Orelse[0].Type == NodeElifClause {
		current = current.Orelse[0]
	}
	// Attach the else clause
	current.Orelse = b.extractBlockBody(elseNode, current)
}

// buildElifClause builds an elif clause node (similar to if statement)
func (b *ASTBuilder) buildElifClause(tsNode *sitter.Node) *Node {
	node := NewNode(NodeElifClause)
	node.Location = b.getLocation(tsNode)

	// Get condition
	if condition := b.getChildByFieldName(tsNode, "condition"); condition != nil {
		node.Test = b.buildNode(condition)
	}

	// Get consequence (body)
	if consequence := b.getChildByFieldName(tsNode, "consequence"); consequence != nil {
		if body := b.buildNode(consequence); body != nil {
			node.Body = b.extractBlockBody(body, node)
		}
	}

	// Note: elif_clause in tree-sitter doesn't have an alternative field
	// The else clause is handled at the parent if_statement level

	return node
}

// buildElseClause builds an else clause node
func (b *ASTBuilder) buildElseClause(tsNode *sitter.Node) *Node {
	node := NewNode(NodeElseClause)
	node.Location = b.getLocation(tsNode)

	// Get body
	if body := b.getChildByFieldName(tsNode, "body"); body != nil {
		if bodyNode := b.buildNode(body); bodyNode != nil {
			node.Body = b.extractBlockBody(bodyNode, node)
		}
	}

	return node
}

// buildForStatement builds a for loop node
func (b *ASTBuilder) buildForStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeFor)
	node.Location = b.getLocation(tsNode)

	// Check if it's async
	if b.hasChildOfType(tsNode, "async") {
		node.Type = NodeAsyncFor
	}

	// Get target variable
	if left := b.getChildByFieldName(tsNode, "left"); left != nil {
		target := b.buildNode(left)
		if target != nil {
			node.Targets = []*Node{target}
		}
	}

	// Get iterator
	if right := b.getChildByFieldName(tsNode, "right"); right != nil {
		node.Iter = b.buildNode(right)
	}

	// Get body
	if bodyNode := b.getChildByFieldName(tsNode, "body"); bodyNode != nil {
		if body := b.buildNode(bodyNode); body != nil {
			node.Body = b.extractBlockBody(body, node)
		}
	}

	// Get else clause if present
	if alternative := b.getChildByFieldName(tsNode, "alternative"); alternative != nil {
		if alt := b.buildNode(alternative); alt != nil {
			node.Orelse = b.extractBlockBody(alt, node)
		}
	}

	return node
}

// buildWhileStatement builds a while loop node
func (b *ASTBuilder) buildWhileStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeWhile)
	node.Location = b.getLocation(tsNode)

	// Get condition
	if condition := b.getChildByFieldName(tsNode, "condition"); condition != nil {
		node.Test = b.buildNode(condition)
	}

	// Get body
	if bodyNode := b.getChildByFieldName(tsNode, "body"); bodyNode != nil {
		if body := b.buildNode(bodyNode); body != nil {
			node.Body = b.extractBlockBody(body, node)
		}
	}

	// Get else clause if present
	if alternative := b.getChildByFieldName(tsNode, "alternative"); alternative != nil {
		if alt := b.buildNode(alternative); alt != nil {
			node.Orelse = b.extractBlockBody(alt, node)
		}
	}

	return node
}

// buildWithStatement builds a with statement node
func (b *ASTBuilder) buildWithStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeWith)
	node.Location = b.getLocation(tsNode)

	// Check if it's async
	if b.hasChildOfType(tsNode, "async") {
		node.Type = NodeAsyncWith
	}

	// Get with items
	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() == "with_clause" {
			if withItem := b.buildWithItem(child); withItem != nil {
				node.AddChild(withItem)
			}
		}
	}

	// Get body
	if bodyNode := b.getChildByFieldName(tsNode, "body"); bodyNode != nil {
		if body := b.buildNode(bodyNode); body != nil {
			node.Body = b.extractBlockBody(body, node)
		}
	}

	return node
}

// buildTryStatement builds a try statement node
func (b *ASTBuilder) buildTryStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeTry)
	node.Location = b.getLocation(tsNode)

	// Get body
	if bodyNode := b.getChildByFieldName(tsNode, "body"); bodyNode != nil {
		if body := b.buildNode(bodyNode); body != nil {
			node.Body = b.extractBlockBody(body, node)
		}
	}

	// Get except handlers
	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() == "except_clause" {
			if handler := b.buildExceptHandler(child); handler != nil {
				node.Handlers = append(node.Handlers, handler)
			}
		}
	}

	// Get else clause
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() == "else_clause" {
			if bodyNode := b.getChildByFieldName(child, "body"); bodyNode != nil {
				if body := b.buildNode(bodyNode); body != nil {
					node.Orelse = b.extractBlockBody(body, node)
				}
			}
		}
	}

	// Get finally clause
	// Note: finally_clause doesn't have a "body" field, we need to find the block child directly
	// This is similar to how except_clause is handled in buildExceptHandler
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() == "finally_clause" {
			// Extract block directly from finally_clause children
			finallyChildCount := int(child.ChildCount())
			for j := 0; j < finallyChildCount; j++ {
				finallyChild := child.Child(j)
				if finallyChild != nil && finallyChild.Type() == "block" {
					if body := b.buildNode(finallyChild); body != nil {
						node.Finalbody = b.extractBlockBody(body, node)
					}
					break
				}
			}
		}
	}

	return node
}

// buildMatchStatement builds a match statement node (Python 3.10+)
func (b *ASTBuilder) buildMatchStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeMatch)
	node.Location = b.getLocation(tsNode)

	// Get subject
	if subject := b.getChildByFieldName(tsNode, "subject"); subject != nil {
		node.Test = b.buildNode(subject)
	}

	// Get cases
	if body := b.getChildByFieldName(tsNode, "body"); body != nil {
		childCount := int(body.ChildCount())
		for i := 0; i < childCount; i++ {
			child := body.Child(i)
			if child != nil && child.Type() == "case_clause" {
				if caseNode := b.buildMatchCase(child); caseNode != nil {
					node.AddToBody(caseNode)
				}
			}
		}
	}

	return node
}

// Helper method implementations...

// buildReturnStatement builds a return statement node
func (b *ASTBuilder) buildReturnStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeReturn)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "return" {
			node.Value = b.buildNode(child)
			break
		}
	}

	return node
}

// buildDeleteStatement builds a delete statement node
func (b *ASTBuilder) buildDeleteStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeDelete)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "del" {
			if target := b.buildNode(child); target != nil {
				node.Targets = append(node.Targets, target)
			}
		}
	}

	return node
}

// buildRaiseStatement builds a raise statement node
func (b *ASTBuilder) buildRaiseStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeRaise)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "raise" {
			// First non-raise child is the exception
			if node.Value == nil {
				node.Value = b.buildNode(child)
			} else {
				// Second is the cause (raise ... from ...)
				node.AddChild(b.buildNode(child))
			}
		}
	}

	return node
}

// buildAssertStatement builds an assert statement node
func (b *ASTBuilder) buildAssertStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeAssert)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	argCount := 0
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "assert" && child.Type() != "," {
			if argCount == 0 {
				node.Test = b.buildNode(child)
			} else {
				node.Value = b.buildNode(child)
			}
			argCount++
		}
	}

	return node
}

// buildImportStatement builds an import statement node
func (b *ASTBuilder) buildImportStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeImport)
	node.Location = b.getLocation(tsNode)

	// Tree-sitter uses "name" field for import names
	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		fieldName := tsNode.FieldNameForChild(i)
		child := tsNode.Child(i)

		if fieldName == "name" && child != nil {
			if child.Type() == "dotted_name" {
				// Simple import
				node.Names = append(node.Names, b.getNodeText(child))
			} else if child.Type() == "aliased_import" {
				// Import with alias
				if alias := b.buildAlias(child); alias != nil {
					node.AddChild(alias)
					// Also add the original name to Names
					if nameChild := b.getChildByFieldName(child, "name"); nameChild != nil {
						node.Names = append(node.Names, b.getNodeText(nameChild))
					}
				}
			}
		}
	}

	return node
}

// buildImportFromStatement builds an import from statement node
func (b *ASTBuilder) buildImportFromStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeImportFrom)
	node.Location = b.getLocation(tsNode)

	// Count leading dots for relative imports
	text := b.getNodeText(tsNode)
	if strings.HasPrefix(text, "from") {
		afterFrom := text[4:]
		trimmed := strings.TrimLeft(afterFrom, " ")
		node.Level = len(trimmed) - len(strings.TrimLeft(trimmed, "."))
	}

	// Get module name
	if moduleNode := b.getChildByFieldName(tsNode, "module_name"); moduleNode != nil {
		// Handle relative imports
		if moduleNode.Type() == "relative_import" {
			// Count dots in import_prefix
			for i := 0; i < int(moduleNode.ChildCount()); i++ {
				child := moduleNode.Child(i)
				if child != nil && child.Type() == "import_prefix" {
					dots := b.getNodeText(child)
					node.Level = len(dots)
				} else if child != nil && child.Type() == "dotted_name" {
					node.Module = b.getNodeText(child)
				}
			}
		} else {
			node.Module = b.getNodeText(moduleNode)
		}
	}

	// Get imported names - tree-sitter uses "name" field directly
	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		fieldName := tsNode.FieldNameForChild(i)
		child := tsNode.Child(i)

		if fieldName == "name" && child != nil {
			// Handle each imported name
			if child.Type() == "dotted_name" || child.Type() == "identifier" {
				node.Names = append(node.Names, b.getNodeText(child))
			} else if child.Type() == "aliased_import" {
				// Handle aliased imports - extract the original name
				if nameChild := b.getChildByFieldName(child, "name"); nameChild != nil {
					node.Names = append(node.Names, b.getNodeText(nameChild))
				}
				// Also build alias for additional info
				if alias := b.buildAlias(child); alias != nil {
					node.AddChild(alias)
				}
			}
		} else if child != nil && child.Type() == "wildcard_import" {
			// Handle wildcard imports (from module import *)
			node.Names = append(node.Names, "*")
		}
	}

	return node
}

// buildGlobalStatement builds a global statement node
func (b *ASTBuilder) buildGlobalStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeGlobal)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() == "identifier" {
			node.Names = append(node.Names, b.getNodeText(child))
		}
	}

	return node
}

// buildNonlocalStatement builds a nonlocal statement node
func (b *ASTBuilder) buildNonlocalStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeNonlocal)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() == "identifier" {
			node.Names = append(node.Names, b.getNodeText(child))
		}
	}

	return node
}

// / buildExpressionStatement builds an expression statement node
func (b *ASTBuilder) buildExpressionStatement(tsNode *sitter.Node) *Node {
	// We always create the container NodeExpr
	node := NewNode(NodeExpr)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && !b.isTrivia(child) {
			// We create the child (Walrus, Function call, etc.)
			expressionNode := b.buildNode(child)

			if expressionNode != nil {
				// We attach it to the container
				node.Value = expressionNode
				node.AddChild(expressionNode)
				return node // We return the container, not the child
			}
		}
	}
	return node
}

// buildAssignment builds an assignment node
func (b *ASTBuilder) buildAssignment(tsNode *sitter.Node) *Node {
	node := NewNode(NodeAssign)
	node.Location = b.getLocation(tsNode)

	// Get left-hand side (targets)
	if left := b.getChildByFieldName(tsNode, "left"); left != nil {
		if target := b.buildNode(left); target != nil {
			node.Targets = []*Node{target}
		}
	}

	// Get right-hand side (value)
	if right := b.getChildByFieldName(tsNode, "right"); right != nil {
		node.Value = b.buildNode(right)
	}

	// Check if it's an annotated assignment
	if typeNode := b.getChildByFieldName(tsNode, "type"); typeNode != nil {
		node.Type = NodeAnnAssign
		// Store type annotation in the first child
		node.AddChild(b.buildNode(typeNode))
	}

	return node
}

// buildAugmentedAssignment builds an augmented assignment node
func (b *ASTBuilder) buildAugmentedAssignment(tsNode *sitter.Node) *Node {
	node := NewNode(NodeAugAssign)
	node.Location = b.getLocation(tsNode)

	// Get target
	if left := b.getChildByFieldName(tsNode, "left"); left != nil {
		if target := b.buildNode(left); target != nil {
			node.Targets = []*Node{target}
		}
	}

	// Get operator
	if operator := b.getChildByFieldName(tsNode, "operator"); operator != nil {
		node.Op = strings.TrimSuffix(b.getNodeText(operator), "=")
	}

	// Get value
	if right := b.getChildByFieldName(tsNode, "right"); right != nil {
		node.Value = b.buildNode(right)
	}

	return node
}

// buildPassStatement builds a pass statement node
func (b *ASTBuilder) buildPassStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodePass)
	node.Location = b.getLocation(tsNode)
	return node
}

// buildBreakStatement builds a break statement node
func (b *ASTBuilder) buildBreakStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeBreak)
	node.Location = b.getLocation(tsNode)
	return node
}

// buildContinueStatement builds a continue statement node
func (b *ASTBuilder) buildContinueStatement(tsNode *sitter.Node) *Node {
	node := NewNode(NodeContinue)
	node.Location = b.getLocation(tsNode)
	return node
}

// buildDecoratedDefinition builds a decorated function or class
func (b *ASTBuilder) buildDecoratedDefinition(tsNode *sitter.Node) *Node {
	var defNode *Node

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil {
			switch child.Type() {
			case "decorator":
				dec := b.buildDecorator(child)
				if dec != nil && defNode != nil {
					defNode.Decorator = append(defNode.Decorator, dec)
				}
			case "function_definition", "class_definition":
				defNode = b.buildNode(child)
			}
		}
	}

	return defNode
}

// Expression builders...

// buildBinaryOp builds a binary operation node
func (b *ASTBuilder) buildBinaryOp(tsNode *sitter.Node) *Node {
	node := NewNode(NodeBinOp)
	node.Location = b.getLocation(tsNode)

	if left := b.getChildByFieldName(tsNode, "left"); left != nil {
		node.Left = b.buildNode(left)
	}

	if operator := b.getChildByFieldName(tsNode, "operator"); operator != nil {
		node.Op = b.getNodeText(operator)
	}

	if right := b.getChildByFieldName(tsNode, "right"); right != nil {
		node.Right = b.buildNode(right)
	}

	return node
}

// buildUnaryOp builds a unary operation node
func (b *ASTBuilder) buildUnaryOp(tsNode *sitter.Node) *Node {
	node := NewNode(NodeUnaryOp)
	node.Location = b.getLocation(tsNode)

	if operator := b.getChildByFieldName(tsNode, "operator"); operator != nil {
		node.Op = b.getNodeText(operator)
	}

	if operand := b.getChildByFieldName(tsNode, "operand"); operand != nil {
		node.Value = b.buildNode(operand)
	}

	return node
}

// buildNamedExpr builds a named expression node (walrus operator)
func (b *ASTBuilder) buildNamedExpr(tsNode *sitter.Node) *Node {
	node := NewNode(NodeNamedExpr)
	node.Location = b.getLocation(tsNode)

	// Get name (target)
	if name := b.getChildByFieldName(tsNode, "name"); name != nil {
		node.AddChild(b.buildNode(name))
	}

	// Get value
	if value := b.getChildByFieldName(tsNode, "value"); value != nil {
		node.Value = b.buildNode(value)
	}

	return node
}

// buildBoolOp builds a boolean operation node
func (b *ASTBuilder) buildBoolOp(tsNode *sitter.Node) *Node {
	node := NewNode(NodeBoolOp)
	node.Location = b.getLocation(tsNode)

	if operator := b.getChildByFieldName(tsNode, "operator"); operator != nil {
		node.Op = b.getNodeText(operator)
	}

	if left := b.getChildByFieldName(tsNode, "left"); left != nil {
		node.AddChild(b.buildNode(left))
	}

	if right := b.getChildByFieldName(tsNode, "right"); right != nil {
		node.AddChild(b.buildNode(right))
	}

	return node
}

// buildCompare builds a comparison operation node
func (b *ASTBuilder) buildCompare(tsNode *sitter.Node) *Node {
	node := NewNode(NodeCompare)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && !b.isTrivia(child) {
			if i == 0 {
				node.Left = b.buildNode(child)
			} else if b.isComparisonOperator(child) {
				node.Op = b.getNodeText(child)
			} else {
				node.AddChild(b.buildNode(child))
			}
		}
	}

	return node
}

// buildIfExp builds a conditional expression node
func (b *ASTBuilder) buildIfExp(tsNode *sitter.Node) *Node {
	node := NewNode(NodeIfExp)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	stage := 0 // 0: body, 1: test, 2: orelse

	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil {
			if child.Type() == "if" {
				stage = 1
			} else if child.Type() == "else" {
				stage = 2
			} else if !b.isTrivia(child) {
				built := b.buildNode(child)
				switch stage {
				case 0:
					node.AddToBody(built)
				case 1:
					node.Test = built
				case 2:
					node.Orelse = []*Node{built}
				}
			}
		}
	}

	return node
}

// buildLambda builds a lambda expression node
func (b *ASTBuilder) buildLambda(tsNode *sitter.Node) *Node {
	node := NewNode(NodeLambda)
	node.Location = b.getLocation(tsNode)

	if params := b.getChildByFieldName(tsNode, "parameters"); params != nil {
		node.Args = b.buildParameters(params)
	}

	if body := b.getChildByFieldName(tsNode, "body"); body != nil {
		node.AddToBody(b.buildNode(body))
	}

	return node
}

// buildCall builds a function call node
func (b *ASTBuilder) buildCall(tsNode *sitter.Node) *Node {
	node := NewNode(NodeCall)
	node.Location = b.getLocation(tsNode)

	if function := b.getChildByFieldName(tsNode, "function"); function != nil {
		node.Value = b.buildNode(function)
	}

	if arguments := b.getChildByFieldName(tsNode, "arguments"); arguments != nil {
		node.Args, node.Keywords = b.buildCallArguments(arguments)
	}

	return node
}

// buildAttribute builds an attribute access node
func (b *ASTBuilder) buildAttribute(tsNode *sitter.Node) *Node {
	node := NewNode(NodeAttribute)
	node.Location = b.getLocation(tsNode)

	if object := b.getChildByFieldName(tsNode, "object"); object != nil {
		node.Value = b.buildNode(object)
	}

	if attr := b.getChildByFieldName(tsNode, "attribute"); attr != nil {
		node.Name = b.getNodeText(attr)
	}

	return node
}

// buildSubscript builds a subscript node
func (b *ASTBuilder) buildSubscript(tsNode *sitter.Node) *Node {
	node := NewNode(NodeSubscript)
	node.Location = b.getLocation(tsNode)

	if value := b.getChildByFieldName(tsNode, "object"); value != nil {
		node.Value = b.buildNode(value)
	}

	if subscript := b.getChildByFieldName(tsNode, "subscript"); subscript != nil {
		node.AddChild(b.buildNode(subscript))
	}

	return node
}

// buildSlice builds a slice node
func (b *ASTBuilder) buildSlice(tsNode *sitter.Node) *Node {
	node := NewNode(NodeSlice)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	sliceArgs := []*Node{nil, nil, nil} // lower, upper, step
	argIndex := 0

	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil {
			if child.Type() == ":" {
				argIndex++
			} else if !b.isTrivia(child) && argIndex < 3 {
				sliceArgs[argIndex] = b.buildNode(child)
			}
		}
	}

	// Store slice components
	if sliceArgs[0] != nil {
		node.AddChild(sliceArgs[0]) // lower
	}
	if sliceArgs[1] != nil {
		node.AddChild(sliceArgs[1]) // upper
	}
	if sliceArgs[2] != nil {
		node.AddChild(sliceArgs[2]) // step
	}

	return node
}

// Collection builders...

// buildList builds a list node
func (b *ASTBuilder) buildList(tsNode *sitter.Node) *Node {
	node := NewNode(NodeList)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "[" && child.Type() != "]" && child.Type() != "," {
			node.AddChild(b.buildNode(child))
		}
	}

	return node
}

// buildTuple builds a tuple node
func (b *ASTBuilder) buildTuple(tsNode *sitter.Node) *Node {
	node := NewNode(NodeTuple)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "(" && child.Type() != ")" && child.Type() != "," {
			node.AddChild(b.buildNode(child))
		}
	}

	return node
}

// buildDict builds a dictionary node
func (b *ASTBuilder) buildDict(tsNode *sitter.Node) *Node {
	node := NewNode(NodeDict)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() == "pair" {
			if key := b.getChildByFieldName(child, "key"); key != nil {
				node.AddChild(b.buildNode(key))
			}
			if value := b.getChildByFieldName(child, "value"); value != nil {
				node.AddChild(b.buildNode(value))
			}
		}
	}

	return node
}

// buildSet builds a set node
func (b *ASTBuilder) buildSet(tsNode *sitter.Node) *Node {
	node := NewNode(NodeSet)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "{" && child.Type() != "}" && child.Type() != "," {
			node.AddChild(b.buildNode(child))
		}
	}

	return node
}

// Comprehension builders...

// buildListComp builds a list comprehension node
func (b *ASTBuilder) buildListComp(tsNode *sitter.Node) *Node {
	node := NewNode(NodeListComp)
	node.Location = b.getLocation(tsNode)
	b.buildComprehension(tsNode, node)
	return node
}

// buildDictComp builds a dictionary comprehension node
func (b *ASTBuilder) buildDictComp(tsNode *sitter.Node) *Node {
	node := NewNode(NodeDictComp)
	node.Location = b.getLocation(tsNode)
	b.buildComprehension(tsNode, node)
	return node
}

// buildSetComp builds a set comprehension node
func (b *ASTBuilder) buildSetComp(tsNode *sitter.Node) *Node {
	node := NewNode(NodeSetComp)
	node.Location = b.getLocation(tsNode)
	b.buildComprehension(tsNode, node)
	return node
}

// buildGeneratorExp builds a generator expression node
func (b *ASTBuilder) buildGeneratorExp(tsNode *sitter.Node) *Node {
	node := NewNode(NodeGeneratorExp)
	node.Location = b.getLocation(tsNode)
	b.buildComprehension(tsNode, node)
	return node
}

// buildComprehension extracts comprehension parts
func (b *ASTBuilder) buildComprehension(tsNode *sitter.Node, node *Node) {
	// Extract the body (the expression being generated)
	childCount := int(tsNode.ChildCount())
	if childCount > 0 {
		firstChild := tsNode.Child(0)
		// Skip opening bracket/brace
		if firstChild != nil && firstChild.Type() != "[" && firstChild.Type() != "{" && firstChild.Type() != "(" {
			node.Value = b.buildNode(firstChild)
		} else if childCount > 1 {
			// Body is likely the second child after opening bracket
			bodyChild := tsNode.Child(1)
			if bodyChild != nil && bodyChild.Type() != "for_in_clause" {
				node.Value = b.buildNode(bodyChild)
			}
		}
	}

	// Process for_in_clauses and if_clauses
	var currentComp *Node
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "for_in_clause" {
			// Create new comprehension node for each for clause
			comp := NewNode(NodeComprehension)

			// Extract target variable(s)
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild != nil && subChild.Type() == "identifier" {
					// First identifier after "for" is the target
					comp.Targets = []*Node{b.buildNode(subChild)}
					break
				}
			}

			// Extract iterator expression (after "in")
			foundIn := false
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild != nil {
					if subChild.Type() == "in" {
						foundIn = true
					} else if foundIn && subChild.Type() != "identifier" {
						comp.Iter = b.buildNode(subChild)
						break
					}
				}
			}

			node.AddChild(comp)
			currentComp = comp
		} else if child.Type() == "if_clause" && currentComp != nil {
			// if_clause follows the for_in_clause it applies to
			// Extract the condition expression
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild != nil && subChild.Type() != "if" {
					currentComp.Test = b.buildNode(subChild)
					break
				}
			}
		}
	}
}

// Yield and await builders...

// buildYield builds a yield expression node
func (b *ASTBuilder) buildYield(tsNode *sitter.Node) *Node {
	node := NewNode(NodeYield)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "yield" {
			node.Value = b.buildNode(child)
			break
		}
	}

	return node
}

// buildYieldFrom builds a yield from expression node
func (b *ASTBuilder) buildYieldFrom(tsNode *sitter.Node) *Node {
	node := NewNode(NodeYieldFrom)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "yield" && child.Type() != "from" {
			node.Value = b.buildNode(child)
			break
		}
	}

	return node
}

// buildAwait builds an await expression node
func (b *ASTBuilder) buildAwait(tsNode *sitter.Node) *Node {
	node := NewNode(NodeAwait)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "await" {
			node.Value = b.buildNode(child)
			break
		}
	}

	return node
}

// Literal builders...

// buildName builds an identifier/name node
func (b *ASTBuilder) buildName(tsNode *sitter.Node) *Node {
	node := NewNode(NodeName)
	node.Location = b.getLocation(tsNode)
	node.Name = b.getNodeText(tsNode)
	return node
}

// buildConstant builds a constant value node
func (b *ASTBuilder) buildConstant(tsNode *sitter.Node) *Node {
	node := NewNode(NodeConstant)
	node.Location = b.getLocation(tsNode)

	text := b.getNodeText(tsNode)
	nodeType := tsNode.Type()

	switch nodeType {
	case "integer":
		if val, err := strconv.ParseInt(text, 0, 64); err == nil {
			node.Value = val
		} else {
			node.Value = text
		}
	case "float":
		if val, err := strconv.ParseFloat(text, 64); err == nil {
			node.Value = val
		} else {
			node.Value = text
		}
	case "string", "concatenated_string":
		// Remove quotes
		if len(text) >= 2 {
			node.Value = text[1 : len(text)-1]
		} else {
			node.Value = text
		}
	case "true":
		node.Value = true
	case "false":
		node.Value = false
	case "none":
		node.Value = nil
	default:
		node.Value = text
	}

	return node
}

// buildFormattedString builds a formatted string (f-string) node
func (b *ASTBuilder) buildFormattedString(tsNode *sitter.Node) *Node {
	node := NewNode(NodeJoinedStr)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil {
			switch child.Type() {
			case "interpolation":
				// Extract the expression inside the interpolation
				exprCount := int(child.ChildCount())
				for j := 0; j < exprCount; j++ {
					exprChild := child.Child(j)
					if exprChild != nil && exprChild.Type() != "{" && exprChild.Type() != "}" {
						fmtValue := NewNode(NodeFormattedValue)
						fmtValue.Value = b.buildNode(exprChild)
						node.AddChild(fmtValue)
					}
				}
			case "string_content":
				// Regular string content
				strNode := NewNode(NodeConstant)
				strNode.Value = b.getNodeText(child)
				node.AddChild(strNode)
			}
		}
	}

	return node
}

// buildBlock builds a block of statements
func (b *ASTBuilder) buildBlock(tsNode *sitter.Node) *Node {
	node := NewNode("block")
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && !b.isTrivia(child) {
			if stmt := b.buildNode(child); stmt != nil {
				node.AddToBody(stmt)
			}
		}
	}

	return node
}

// Helper methods...

// buildParameters builds function parameters
func (b *ASTBuilder) buildParameters(tsNode *sitter.Node) []*Node {
	var params []*Node

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil {
			switch child.Type() {
			case "identifier":
				arg := NewNode(NodeArg)
				arg.Name = b.getNodeText(child)
				params = append(params, arg)
			case "default_parameter":
				arg := NewNode(NodeArg)
				if nameNode := b.getChildByFieldName(child, "name"); nameNode != nil {
					arg.Name = b.getNodeText(nameNode)
				}
				if valueNode := b.getChildByFieldName(child, "value"); valueNode != nil {
					arg.Value = b.buildNode(valueNode)
				}
				params = append(params, arg)
			case "typed_parameter", "typed_default_parameter":
				arg := NewNode(NodeArg)
				if nameNode := b.getChildByFieldName(child, "name"); nameNode != nil {
					arg.Name = b.getNodeText(nameNode)
				}
				// Store type annotation in Value field for now
				if typeNode := b.getChildByFieldName(child, "type"); typeNode != nil {
					arg.Value = b.getNodeText(typeNode)
				}
				params = append(params, arg)
			case "list_splat_pattern":
				arg := NewNode(NodeArg)
				arg.Name = "*" + b.getNodeTextExcluding(child, "*")
				params = append(params, arg)
			case "dictionary_splat_pattern":
				arg := NewNode(NodeArg)
				arg.Name = "**" + b.getNodeTextExcluding(child, "**")
				params = append(params, arg)
			}
		}
	}

	return params
}

// buildArgumentList builds a list of arguments (for base classes, etc.)
func (b *ASTBuilder) buildArgumentList(tsNode *sitter.Node) []*Node {
	var args []*Node

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "(" && child.Type() != ")" && child.Type() != "," {
			args = append(args, b.buildNode(child))
		}
	}

	return args
}

// buildCallArguments builds call arguments and keywords
func (b *ASTBuilder) buildCallArguments(tsNode *sitter.Node) ([]*Node, []*Node) {
	var args []*Node
	var keywords []*Node

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil {
			switch child.Type() {
			case "keyword_argument":
				kw := NewNode(NodeKeyword)
				if nameNode := b.getChildByFieldName(child, "name"); nameNode != nil {
					kw.Name = b.getNodeText(nameNode)
				}
				if valueNode := b.getChildByFieldName(child, "value"); valueNode != nil {
					kw.Value = b.buildNode(valueNode)
				}
				keywords = append(keywords, kw)
			case "(", ")", ",":
				// Skip syntax elements
			default:
				args = append(args, b.buildNode(child))
			}
		}
	}

	return args, keywords
}

// buildWithItem builds a with item node
func (b *ASTBuilder) buildWithItem(tsNode *sitter.Node) *Node {
	node := NewNode(NodeWithItem)
	node.Location = b.getLocation(tsNode)

	if item := b.getChildByFieldName(tsNode, "item"); item != nil {
		node.Value = b.buildNode(item)
	}

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() == "as_pattern" {
			if alias := b.getChildByFieldName(child, "alias"); alias != nil {
				node.Name = b.getNodeText(alias)
			}
		}
	}

	return node
}

// buildExceptHandler builds an except handler node
func (b *ASTBuilder) buildExceptHandler(tsNode *sitter.Node) *Node {
	node := NewNode(NodeExceptHandler)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil {
			switch child.Type() {
			case "as_pattern":
				// Exception type and optional name
				if exType := b.getChildByFieldName(child, "type"); exType != nil {
					node.Value = b.buildNode(exType)
				}
				if alias := b.getChildByFieldName(child, "alias"); alias != nil {
					node.Name = b.getNodeText(alias)
				}
			case "block":
				// Handler body
				if body := b.buildNode(child); body != nil {
					node.Body = b.extractBlockBody(body, node)
				}
			default:
				if child.Type() != "except" && child.Type() != ":" {
					// Exception type without alias
					node.Value = b.buildNode(child)
				}
			}
		}
	}

	return node
}

// buildMatchCase builds a match case node
func (b *ASTBuilder) buildMatchCase(tsNode *sitter.Node) *Node {
	node := NewNode(NodeMatchCase)
	node.Location = b.getLocation(tsNode)

	if pattern := b.getChildByFieldName(tsNode, "pattern"); pattern != nil {
		node.Test = b.buildNode(pattern)
	}

	if guard := b.getChildByFieldName(tsNode, "guard"); guard != nil {
		// Store guard in Value field
		node.Value = b.buildNode(guard)
	}

	if consequence := b.getChildByFieldName(tsNode, "consequence"); consequence != nil {
		if body := b.buildNode(consequence); body != nil {
			node.Body = b.extractBlockBody(body, node)
		}
	}

	return node
}

// buildDecorator builds a decorator node
func (b *ASTBuilder) buildDecorator(tsNode *sitter.Node) *Node {
	node := NewNode(NodeDecorator)
	node.Location = b.getLocation(tsNode)

	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() != "@" {
			node.Value = b.buildNode(child)
			break
		}
	}

	return node
}

// buildAlias builds an import alias node
func (b *ASTBuilder) buildAlias(tsNode *sitter.Node) *Node {
	node := NewNode(NodeAlias)
	node.Location = b.getLocation(tsNode)

	if tsNode.Type() == "aliased_import" {
		if nameNode := b.getChildByFieldName(tsNode, "name"); nameNode != nil {
			node.Name = b.getNodeText(nameNode)
		}
		if aliasNode := b.getChildByFieldName(tsNode, "alias"); aliasNode != nil {
			node.Value = b.getNodeText(aliasNode)
		}
	} else {
		node.Name = b.getNodeText(tsNode)
	}

	return node
}

// Utility methods...

// getLocation extracts location information from a tree-sitter node
func (b *ASTBuilder) getLocation(tsNode *sitter.Node) Location {
	startPoint := tsNode.StartPoint()
	endPoint := tsNode.EndPoint()

	return Location{
		StartLine: int(startPoint.Row) + 1,
		StartCol:  int(startPoint.Column),
		EndLine:   int(endPoint.Row) + 1,
		EndCol:    int(endPoint.Column),
	}
}

// getNodeText gets the text content of a node
func (b *ASTBuilder) getNodeText(tsNode *sitter.Node) string {
	return tsNode.Content(b.source)
}

// getNodeTextExcluding gets node text excluding certain prefixes
func (b *ASTBuilder) getNodeTextExcluding(tsNode *sitter.Node, exclude string) string {
	text := b.getNodeText(tsNode)
	return strings.TrimPrefix(text, exclude)
}

// getChildByFieldName gets a child node by field name
func (b *ASTBuilder) getChildByFieldName(tsNode *sitter.Node, fieldName string) *sitter.Node {
	return tsNode.ChildByFieldName(fieldName)
}

// hasChildOfType checks if a node has a child of a specific type
func (b *ASTBuilder) hasChildOfType(tsNode *sitter.Node, childType string) bool {
	childCount := int(tsNode.ChildCount())
	for i := 0; i < childCount; i++ {
		child := tsNode.Child(i)
		if child != nil && child.Type() == childType {
			return true
		}
	}
	return false
}

// isTrivia checks if a node is trivia (comments, whitespace)
func (b *ASTBuilder) isTrivia(tsNode *sitter.Node) bool {
	nodeType := tsNode.Type()
	return nodeType == "comment" || nodeType == "line_continuation"
}

// isComparisonOperator checks if a node is a comparison operator
func (b *ASTBuilder) isComparisonOperator(tsNode *sitter.Node) bool {
	text := b.getNodeText(tsNode)
	operators := []string{"<", ">", "==", ">=", "<=", "!=", "in", "not in", "is", "is not"}
	for _, op := range operators {
		if text == op {
			return true
		}
	}
	return false
}
