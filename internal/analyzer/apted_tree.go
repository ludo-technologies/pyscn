package analyzer

import (
	"fmt"
	"strings"

	coreapted "github.com/ludo-technologies/polyscan/core/apted"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// TreeConverter converts parser AST nodes to APTED tree nodes
type TreeConverter struct {
	nextID         int
	skipDocstrings bool
}

// NewTreeConverter creates a new tree converter with default settings (no docstring skipping)
func NewTreeConverter() *TreeConverter {
	return &TreeConverter{nextID: 0, skipDocstrings: false}
}

// NewTreeConverterWithConfig creates a tree converter with configuration
func NewTreeConverterWithConfig(skipDocstrings bool) *TreeConverter {
	return &TreeConverter{nextID: 0, skipDocstrings: skipDocstrings}
}

// isDocstring checks if the given node is a docstring at the given position in the body.
// A docstring is the first string constant in a function/class/module body.
// The parser returns NodeConstant directly in the Body (not wrapped in NodeExpr).
func (tc *TreeConverter) isDocstring(node *parser.Node, positionInBody int) bool {
	if !tc.skipDocstrings {
		return false
	}

	// Must be the first statement (position 0)
	if positionInBody != 0 {
		return false
	}

	// Case 1: Direct Constant node (actual parser output)
	// The parser's buildExpressionStatement returns the child node directly
	if node.Type == parser.NodeConstant {
		if node.Value == nil {
			return false
		}
		_, isString := node.Value.(string)
		return isString
	}

	// Case 2: Expr wrapping Constant (for backward compatibility with manual AST construction)
	if node.Type != parser.NodeExpr {
		return false
	}

	// Must have exactly one child which is a Constant
	if len(node.Children) != 1 {
		return false
	}

	child := node.Children[0]
	if child.Type != parser.NodeConstant {
		return false
	}

	// The constant must be a string
	if child.Value == nil {
		return false
	}

	// Check if value is a string type
	_, isString := child.Value.(string)
	return isString
}

// ConvertAST converts a parser AST node to an APTED tree
func (tc *TreeConverter) ConvertAST(astNode *parser.Node) *coreapted.TreeNode {
	if astNode == nil {
		return nil
	}

	// Create tree node with simplified label
	label := tc.getNodeLabel(astNode)
	treeNode := coreapted.NewTreeNode(tc.nextID, label)
	tc.nextID++

	// Store reference to original AST node
	treeNode.OriginalNode = astNode

	for _, child := range parser.OrderedChildren(astNode, tc.shouldSkipBodyNode) {
		if childNode := tc.ConvertAST(child); childNode != nil {
			treeNode.AddChild(childNode)
		}
	}

	return treeNode
}

func (tc *TreeConverter) shouldSkipBodyNode(parent *parser.Node, bodyNode *parser.Node, bodyIndex int) bool {
	return tc.canNodeHaveDocstring(parent.Type) && tc.isDocstring(bodyNode, bodyIndex)
}

// canNodeHaveDocstring checks if a node type can have a docstring
func (tc *TreeConverter) canNodeHaveDocstring(nodeType parser.NodeType) bool {
	switch nodeType {
	case parser.NodeModule, parser.NodeClassDef, parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
		return true
	default:
		return false
	}
}

// getNodeLabel extracts a meaningful label from the AST node
func (tc *TreeConverter) getNodeLabel(astNode *parser.Node) string {
	// Use the node type as the primary label
	label := string(astNode.Type)

	// For some node types, include additional information
	switch astNode.Type {
	case parser.NodeName:
		if astNode.Name != "" {
			label = fmt.Sprintf("Name(%s)", astNode.Name)
		}
	case parser.NodeConstant:
		if astNode.Value != nil {
			label = fmt.Sprintf("Constant(%v)", astNode.Value)
		}
	case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
		if astNode.Name != "" {
			label = fmt.Sprintf("FunctionDef(%s)", astNode.Name)
		}
	case parser.NodeClassDef:
		if astNode.Name != "" {
			label = fmt.Sprintf("ClassDef(%s)", astNode.Name)
		}
	case parser.NodeAttribute:
		if astNode.Name != "" {
			label = fmt.Sprintf("Attribute(%s)", astNode.Name)
		}
	case parser.NodeKeyword:
		if astNode.Name != "" {
			label = fmt.Sprintf("Keyword(%s)", astNode.Name)
		}
	case parser.NodeArg:
		if astNode.Name != "" {
			label = fmt.Sprintf("Arg(%s)", astNode.Name)
		}
	case parser.NodeAlias:
		if astNode.Name != "" {
			if alias, ok := astNode.Value.(string); ok && alias != "" {
				label = fmt.Sprintf("Alias(%s as %s)", astNode.Name, alias)
			} else {
				label = fmt.Sprintf("Alias(%s)", astNode.Name)
			}
		}
	case parser.NodeWithItem, parser.NodeExceptHandler:
		if astNode.Name != "" {
			label = fmt.Sprintf("%s(%s)", astNode.Type, astNode.Name)
		}
	case parser.NodeGlobal, parser.NodeNonlocal:
		if len(astNode.Names) > 0 {
			label = fmt.Sprintf("%s(%s)", astNode.Type, strings.Join(astNode.Names, ","))
		}
	case parser.NodeImport:
		if len(astNode.Names) > 0 {
			label = fmt.Sprintf("Import(%s)", strings.Join(astNode.Names, ","))
		}
	case parser.NodeImportFrom:
		module := astNode.Module
		if astNode.Level > 0 {
			module = strings.Repeat(".", astNode.Level) + module
		}
		if module != "" || len(astNode.Names) > 0 {
			label = fmt.Sprintf("ImportFrom(%s:%s)", module, strings.Join(astNode.Names, ","))
		}
	case parser.NodeBinOp, parser.NodeUnaryOp, parser.NodeBoolOp, parser.NodeCompare, parser.NodeAugAssign:
		if astNode.Op != "" {
			label = fmt.Sprintf("%s(%s)", astNode.Type, astNode.Op)
		}
	}

	return label
}
