package analyzer

import (
	"fmt"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// TreeNode represents a node in the ordered tree for APTED algorithm
type TreeNode struct {
	// Unique identifier for this node
	ID int

	// Label for the node (typically the node type or value)
	Label string

	// Tree structure
	Children []*TreeNode
	Parent   *TreeNode

	// APTED-specific fields for optimization
	PostOrderID  int  // Post-order traversal position
	LeftMostLeaf int  // Left-most leaf descendant
	KeyRoot      bool // Whether this node is a key root

	// Optional metadata from original AST
	OriginalNode *parser.Node
}

// NewTreeNode creates a new tree node with the given ID and label
func NewTreeNode(id int, label string) *TreeNode {
	return &TreeNode{
		ID:       id,
		Label:    label,
		Children: []*TreeNode{},
	}
}

// AddChild adds a child node to this node
func (t *TreeNode) AddChild(child *TreeNode) {
	if child != nil {
		child.Parent = t
		t.Children = append(t.Children, child)
	}
}

// IsLeaf returns true if this node has no children
func (t *TreeNode) IsLeaf() bool {
	return len(t.Children) == 0
}

// Size returns the size of the subtree rooted at this node
func (t *TreeNode) Size() int {
	return t.SizeWithDepthLimit(1000) // Default recursion limit
}

// SizeWithDepthLimit returns the size with maximum recursion depth limit
func (t *TreeNode) SizeWithDepthLimit(maxDepth int) int {
	if maxDepth <= 0 {
		return 1 // Return 1 to avoid infinite loops, treat as leaf
	}

	size := 1
	for _, child := range t.Children {
		size += child.SizeWithDepthLimit(maxDepth - 1)
	}
	return size
}

// Height returns the height of the subtree rooted at this node
func (t *TreeNode) Height() int {
	return t.HeightWithDepthLimit(1000) // Default recursion limit
}

// HeightWithDepthLimit returns the height with maximum recursion depth limit
func (t *TreeNode) HeightWithDepthLimit(maxDepth int) int {
	if maxDepth <= 0 {
		return 0 // Treat as leaf when depth limit reached
	}

	if t.IsLeaf() {
		return 0
	}

	maxHeight := 0
	for _, child := range t.Children {
		if h := child.HeightWithDepthLimit(maxDepth - 1); h > maxHeight {
			maxHeight = h
		}
	}
	return maxHeight + 1
}

// String returns a string representation of the node
func (t *TreeNode) String() string {
	return fmt.Sprintf("Node{ID: %d, Label: %s, Children: %d}", t.ID, t.Label, len(t.Children))
}

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
// A docstring is the first Expr(Constant(str)) in a function/class/module body.
func (tc *TreeConverter) isDocstring(node *parser.Node, positionInBody int) bool {
	if !tc.skipDocstrings {
		return false
	}

	// Must be the first statement (position 0)
	if positionInBody != 0 {
		return false
	}

	// Must be an Expr node
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
func (tc *TreeConverter) ConvertAST(astNode *parser.Node) *TreeNode {
	if astNode == nil {
		return nil
	}

	// Create tree node with simplified label
	label := tc.getNodeLabel(astNode)
	treeNode := NewTreeNode(tc.nextID, label)
	tc.nextID++

	// Store reference to original AST node
	treeNode.OriginalNode = astNode

	// Convert children recursively
	for _, child := range astNode.Children {
		if childNode := tc.ConvertAST(child); childNode != nil {
			treeNode.AddChild(childNode)
		}
	}

	// Also convert body, orelse, and other AST-specific children
	// Skip docstrings if configured (first Expr(Constant(str)) in body)
	canHaveDocstring := tc.canNodeHaveDocstring(astNode.Type)
	for i, bodyNode := range astNode.Body {
		if canHaveDocstring && tc.isDocstring(bodyNode, i) {
			continue
		}
		if childNode := tc.ConvertAST(bodyNode); childNode != nil {
			treeNode.AddChild(childNode)
		}
	}

	for _, orelseNode := range astNode.Orelse {
		if childNode := tc.ConvertAST(orelseNode); childNode != nil {
			treeNode.AddChild(childNode)
		}
	}

	for _, finalbodyNode := range astNode.Finalbody {
		if childNode := tc.ConvertAST(finalbodyNode); childNode != nil {
			treeNode.AddChild(childNode)
		}
	}

	for _, handlerNode := range astNode.Handlers {
		if childNode := tc.ConvertAST(handlerNode); childNode != nil {
			treeNode.AddChild(childNode)
		}
	}

	return treeNode
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
	case parser.NodeBinOp, parser.NodeUnaryOp, parser.NodeBoolOp:
		if astNode.Op != "" {
			label = fmt.Sprintf("%s(%s)", astNode.Type, astNode.Op)
		}
	}

	return label
}

// PostOrderTraversal performs post-order traversal and assigns post-order IDs
func PostOrderTraversal(root *TreeNode) {
	if root == nil {
		return
	}

	postOrderID := 0
	postOrderTraversalRecursive(root, &postOrderID)
}

// postOrderTraversalRecursive recursively performs post-order traversal
func postOrderTraversalRecursive(node *TreeNode, postOrderID *int) {
	if node == nil {
		return
	}

	// Visit children first
	for _, child := range node.Children {
		postOrderTraversalRecursive(child, postOrderID)
	}

	// Then visit this node
	node.PostOrderID = *postOrderID
	*postOrderID++
}

// ComputeLeftMostLeaves computes left-most leaf descendants for all nodes
func ComputeLeftMostLeaves(root *TreeNode) {
	if root == nil {
		return
	}
	computeLeftMostLeavesRecursive(root)
}

// computeLeftMostLeavesRecursive recursively computes left-most leaf descendants
func computeLeftMostLeavesRecursive(node *TreeNode) int {
	if node.IsLeaf() || len(node.Children) == 0 {
		node.LeftMostLeaf = node.PostOrderID
		return node.LeftMostLeaf
	}

	// Get left-most leaf from first child
	leftMostLeaf := computeLeftMostLeavesRecursive(node.Children[0])
	node.LeftMostLeaf = leftMostLeaf

	// Process remaining children
	for i := 1; i < len(node.Children); i++ {
		computeLeftMostLeavesRecursive(node.Children[i])
	}

	return leftMostLeaf
}

// ComputeKeyRoots identifies key roots for path decomposition
func ComputeKeyRoots(root *TreeNode) []int {
	if root == nil {
		return []int{}
	}

	keyRoots := []int{}
	visited := make(map[int]bool)

	computeKeyRootsRecursive(root, &keyRoots, visited)

	return keyRoots
}

// computeKeyRootsRecursive recursively identifies key roots
func computeKeyRootsRecursive(node *TreeNode, keyRoots *[]int, visited map[int]bool) {
	if node == nil {
		return
	}

	// A node is a key root if its left-most leaf hasn't been visited
	if !visited[node.LeftMostLeaf] {
		node.KeyRoot = true
		*keyRoots = append(*keyRoots, node.PostOrderID)
		visited[node.LeftMostLeaf] = true
	}

	// Process children
	for _, child := range node.Children {
		computeKeyRootsRecursive(child, keyRoots, visited)
	}
}

// PrepareTreeForAPTED prepares a tree for APTED algorithm by computing all necessary indices
func PrepareTreeForAPTED(root *TreeNode) []int {
	if root == nil {
		return []int{}
	}

	// Step 1: Assign post-order IDs
	PostOrderTraversal(root)

	// Step 2: Compute left-most leaf descendants
	ComputeLeftMostLeaves(root)

	// Step 3: Identify key roots
	keyRoots := ComputeKeyRoots(root)

	return keyRoots
}

// GetNodeByPostOrderID finds a node by its post-order ID
func GetNodeByPostOrderID(root *TreeNode, postOrderID int) *TreeNode {
	if root == nil {
		return nil
	}

	if root.PostOrderID == postOrderID {
		return root
	}

	for _, child := range root.Children {
		if node := GetNodeByPostOrderID(child, postOrderID); node != nil {
			return node
		}
	}

	return nil
}

// GetSubtreeNodes returns all nodes in the subtree rooted at the given node
func GetSubtreeNodes(root *TreeNode) []*TreeNode {
	return GetSubtreeNodesWithDepthLimit(root, 1000) // Default recursion limit
}

// GetSubtreeNodesWithDepthLimit returns all nodes with maximum recursion depth limit
func GetSubtreeNodesWithDepthLimit(root *TreeNode, maxDepth int) []*TreeNode {
	if root == nil || maxDepth <= 0 {
		return []*TreeNode{}
	}

	nodes := []*TreeNode{root}
	for _, child := range root.Children {
		nodes = append(nodes, GetSubtreeNodesWithDepthLimit(child, maxDepth-1)...)
	}

	return nodes
}
