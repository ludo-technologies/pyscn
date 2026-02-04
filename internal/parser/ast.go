package parser

import "fmt"

// NodeType represents the type of AST node
type NodeType string

// Python AST node types
const (
	// Module and structure
	NodeModule      NodeType = "Module"
	NodeInteractive NodeType = "Interactive"
	NodeExpression  NodeType = "Expression"
	NodeSuite       NodeType = "Suite"

	// Statements
	NodeFunctionDef      NodeType = "FunctionDef"
	NodeAsyncFunctionDef NodeType = "AsyncFunctionDef"
	NodeClassDef         NodeType = "ClassDef"
	NodeReturn           NodeType = "Return"
	NodeDelete           NodeType = "Delete"
	NodeAssign           NodeType = "Assign"
	NodeAugAssign        NodeType = "AugAssign"
	NodeAnnAssign        NodeType = "AnnAssign"
	NodeFor              NodeType = "For"
	NodeAsyncFor         NodeType = "AsyncFor"
	NodeWhile            NodeType = "While"
	NodeIf               NodeType = "If"
	NodeWith             NodeType = "With"
	NodeAsyncWith        NodeType = "AsyncWith"
	NodeMatch            NodeType = "Match"
	NodeRaise            NodeType = "Raise"
	NodeTry              NodeType = "Try"
	NodeAssert           NodeType = "Assert"
	NodeImport           NodeType = "Import"
	NodeImportFrom       NodeType = "ImportFrom"
	NodeGlobal           NodeType = "Global"
	NodeNonlocal         NodeType = "Nonlocal"
	NodeExpr             NodeType = "Expr"
	NodePass             NodeType = "Pass"
	NodeBreak            NodeType = "Break"
	NodeContinue         NodeType = "Continue"

	// Expressions
	NodeBoolOp         NodeType = "BoolOp"
	NodeNamedExpr      NodeType = "NamedExpr"
	NodeBinOp          NodeType = "BinOp"
	NodeUnaryOp        NodeType = "UnaryOp"
	NodeLambda         NodeType = "Lambda"
	NodeIfExp          NodeType = "IfExp"
	NodeDict           NodeType = "Dict"
	NodeSet            NodeType = "Set"
	NodeListComp       NodeType = "ListComp"
	NodeSetComp        NodeType = "SetComp"
	NodeDictComp       NodeType = "DictComp"
	NodeGeneratorExp   NodeType = "GeneratorExp"
	NodeAwait          NodeType = "Await"
	NodeYield          NodeType = "Yield"
	NodeYieldFrom      NodeType = "YieldFrom"
	NodeCompare        NodeType = "Compare"
	NodeCall           NodeType = "Call"
	NodeFormattedValue NodeType = "FormattedValue"
	NodeJoinedStr      NodeType = "JoinedStr"
	NodeConstant       NodeType = "Constant"
	NodeAttribute      NodeType = "Attribute"
	NodeSubscript      NodeType = "Subscript"
	NodeStarred        NodeType = "Starred"
	NodeName           NodeType = "Name"
	NodeList           NodeType = "List"
	NodeTuple          NodeType = "Tuple"
	NodeSlice          NodeType = "Slice"

	// Patterns (for match statements)
	NodeMatchValue     NodeType = "MatchValue"
	NodeMatchSingleton NodeType = "MatchSingleton"
	NodeMatchSequence  NodeType = "MatchSequence"
	NodeMatchMapping   NodeType = "MatchMapping"
	NodeMatchClass     NodeType = "MatchClass"
	NodeMatchStar      NodeType = "MatchStar"
	NodeMatchAs        NodeType = "MatchAs"
	NodeMatchOr        NodeType = "MatchOr"

	// Other
	NodeAlias         NodeType = "Alias"
	NodeExceptHandler NodeType = "ExceptHandler"
	NodeArguments     NodeType = "Arguments"
	NodeArg           NodeType = "Arg"
	NodeKeyword       NodeType = "Keyword"
	NodeComprehension NodeType = "Comprehension"
	NodeDecorator     NodeType = "Decorator"
	NodeWithItem      NodeType = "WithItem"
	NodeMatchCase     NodeType = "MatchCase"
	NodeElseClause    NodeType = "else_clause" // Structural marker from parser
	NodeElifClause    NodeType = "elif_clause" // Structural marker from parser
	NodeBlock         NodeType = "block"       // Block of statements from parser

	// Tree-sitter specific nodes
	NodeGenericType   NodeType = "generic_type"
	NodeTypeParameter NodeType = "type_parameter"
	NodeTypeNode      NodeType = "type"
)

// Location represents the position of a node in the source code
type Location struct {
	File      string
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
}

// Node represents an AST node
type Node struct {
	Type     NodeType
	Value    interface{} // Can hold various values depending on node type
	Children []*Node
	Location Location
	Parent   *Node

	// Additional fields for specific node types
	Name      string   // For function/class definitions, variables
	Targets   []*Node  // For assignments
	Body      []*Node  // For compound statements
	Orelse    []*Node  // For if/for/while/try statements
	Finalbody []*Node  // For try statements
	Handlers  []*Node  // For try statements
	Test      *Node    // For if/while statements
	Iter      *Node    // For for loops
	Args      []*Node  // For function calls
	Keywords  []*Node  // For function calls
	Decorator []*Node  // For decorated functions/classes
	Bases     []*Node  // For class definitions
	Left      *Node    // For binary operations
	Right     *Node    // For binary operations
	Op        string   // For operations
	Module    string   // For imports
	Names     []string // For imports
	Level     int      // For relative imports
}

// NewNode creates a new AST node
func NewNode(nodeType NodeType) *Node {
	return &Node{
		Type:     nodeType,
		Children: []*Node{},
		Body:     []*Node{},
		Orelse:   []*Node{},
		Args:     []*Node{},
		Keywords: []*Node{},
		Names:    []string{},
	}
}

// AddChild adds a child node
func (n *Node) AddChild(child *Node) {
	if child != nil {
		child.Parent = n
		n.Children = append(n.Children, child)
	}
}

// AddToBody adds a node to the body
func (n *Node) AddToBody(node *Node) {
	if node != nil {
		node.Parent = n
		n.Body = append(n.Body, node)
	}
}

// GetChildren returns all child nodes
func (n *Node) GetChildren() []*Node {
	allChildren := []*Node{}
	allChildren = append(allChildren, n.Children...)
	allChildren = append(allChildren, n.Body...)
	allChildren = append(allChildren, n.Orelse...)
	allChildren = append(allChildren, n.Finalbody...)
	allChildren = append(allChildren, n.Handlers...)

	if n.Test != nil {
		allChildren = append(allChildren, n.Test)
	}
	if n.Iter != nil {
		allChildren = append(allChildren, n.Iter)
	}
	if n.Left != nil {
		allChildren = append(allChildren, n.Left)
	}
	if n.Right != nil {
		allChildren = append(allChildren, n.Right)
	}

	allChildren = append(allChildren, n.Targets...)
	allChildren = append(allChildren, n.Args...)
	allChildren = append(allChildren, n.Keywords...)
	allChildren = append(allChildren, n.Decorator...)
	allChildren = append(allChildren, n.Bases...)

	return allChildren
}

// IsStatement returns true if the node is a statement
func (n *Node) IsStatement() bool {
	switch n.Type {
	case NodeFunctionDef, NodeAsyncFunctionDef, NodeClassDef,
		NodeReturn, NodeDelete, NodeAssign, NodeAugAssign, NodeAnnAssign,
		NodeFor, NodeAsyncFor, NodeWhile, NodeIf, NodeWith, NodeAsyncWith,
		NodeMatch, NodeRaise, NodeTry, NodeAssert, NodeImport, NodeImportFrom,
		NodeGlobal, NodeNonlocal, NodeExpr, NodePass, NodeBreak, NodeContinue:
		return true
	default:
		return false
	}
}

// IsExpression returns true if the node is an expression
func (n *Node) IsExpression() bool {
	switch n.Type {
	case NodeBoolOp, NodeNamedExpr, NodeBinOp, NodeUnaryOp, NodeLambda,
		NodeIfExp, NodeDict, NodeSet, NodeListComp, NodeSetComp, NodeDictComp,
		NodeGeneratorExp, NodeAwait, NodeYield, NodeYieldFrom, NodeCompare,
		NodeCall, NodeFormattedValue, NodeJoinedStr, NodeConstant,
		NodeAttribute, NodeSubscript, NodeStarred, NodeName, NodeList,
		NodeTuple, NodeSlice:
		return true
	default:
		return false
	}
}

// IsControlFlow returns true if the node represents control flow
func (n *Node) IsControlFlow() bool {
	switch n.Type {
	case NodeIf, NodeFor, NodeAsyncFor, NodeWhile, NodeWith, NodeAsyncWith,
		NodeMatch, NodeTry, NodeBreak, NodeContinue, NodeReturn, NodeRaise:
		return true
	default:
		return false
	}
}

// String returns a string representation of the node
func (n *Node) String() string {
	if n.Name != "" {
		return fmt.Sprintf("%s(%s)", n.Type, n.Name)
	}
	if n.Value != nil {
		return fmt.Sprintf("%s(%v)", n.Type, n.Value)
	}
	return string(n.Type)
}

// Walk traverses the AST using depth-first search
func (n *Node) Walk(visitor func(*Node) bool) {
	if !visitor(n) {
		return
	}

	for _, child := range n.GetChildren() {
		if child != nil {
			child.Walk(visitor)
		}
	}
}

// WalkDeep traverses the AST including the Value field when it contains a *Node.
// This is necessary because tree-sitter stores some child nodes in the Value field
// (e.g., Call nodes store the callee in Value, Assign nodes store the RHS in Value).
func (n *Node) WalkDeep(visitor func(*Node) bool) {
	visited := make(map[*Node]bool)
	n.walkDeep(visitor, visited)
}

func (n *Node) walkDeep(visitor func(*Node) bool, visited map[*Node]bool) {
	if n == nil {
		return
	}
	if visited[n] {
		return
	}
	visited[n] = true

	if !visitor(n) {
		return
	}

	// Traverse all standard AST children (Body/Orelse/Handlers/etc.).
	for _, child := range n.GetChildren() {
		child.walkDeep(visitor, visited)
	}

	// Traverse Value field if it contains a Node (tree-sitter-specific storage).
	if valueNode, ok := n.Value.(*Node); ok {
		valueNode.walkDeep(visitor, visited)
	}
}

// Find finds all nodes matching a predicate
func (n *Node) Find(predicate func(*Node) bool) []*Node {
	var results []*Node
	n.Walk(func(node *Node) bool {
		if predicate(node) {
			results = append(results, node)
		}
		return true
	})
	return results
}

// FindByType finds all nodes of a specific type
func (n *Node) FindByType(nodeType NodeType) []*Node {
	return n.Find(func(node *Node) bool {
		return node.Type == nodeType
	})
}

// GetParentOfType finds the nearest parent of a specific type
func (n *Node) GetParentOfType(nodeType NodeType) *Node {
	current := n.Parent
	for current != nil {
		if current.Type == nodeType {
			return current
		}
		current = current.Parent
	}
	return nil
}

// Copy creates a deep copy of the node
func (n *Node) Copy() *Node {
	if n == nil {
		return nil
	}

	copied := &Node{
		Type:     n.Type,
		Value:    n.Value,
		Location: n.Location,
		Name:     n.Name,
		Op:       n.Op,
		Module:   n.Module,
		Names:    append([]string{}, n.Names...),
		Level:    n.Level,
	}

	// Copy children
	for _, child := range n.Children {
		if child != nil {
			copiedChild := child.Copy()
			copiedChild.Parent = copied
			copied.Children = append(copied.Children, copiedChild)
		}
	}

	// Copy body
	for _, node := range n.Body {
		if node != nil {
			copiedNode := node.Copy()
			copiedNode.Parent = copied
			copied.Body = append(copied.Body, copiedNode)
		}
	}

	// Copy other fields
	for _, node := range n.Orelse {
		if node != nil {
			copiedNode := node.Copy()
			copiedNode.Parent = copied
			copied.Orelse = append(copied.Orelse, copiedNode)
		}
	}

	for _, node := range n.Finalbody {
		if node != nil {
			copiedNode := node.Copy()
			copiedNode.Parent = copied
			copied.Finalbody = append(copied.Finalbody, copiedNode)
		}
	}

	for _, node := range n.Handlers {
		if node != nil {
			copiedNode := node.Copy()
			copiedNode.Parent = copied
			copied.Handlers = append(copied.Handlers, copiedNode)
		}
	}

	if n.Test != nil {
		copied.Test = n.Test.Copy()
		copied.Test.Parent = copied
	}

	if n.Iter != nil {
		copied.Iter = n.Iter.Copy()
		copied.Iter.Parent = copied
	}

	if n.Left != nil {
		copied.Left = n.Left.Copy()
		copied.Left.Parent = copied
	}

	if n.Right != nil {
		copied.Right = n.Right.Copy()
		copied.Right.Parent = copied
	}

	for _, node := range n.Targets {
		if node != nil {
			copiedNode := node.Copy()
			copiedNode.Parent = copied
			copied.Targets = append(copied.Targets, copiedNode)
		}
	}

	for _, node := range n.Args {
		if node != nil {
			copiedNode := node.Copy()
			copiedNode.Parent = copied
			copied.Args = append(copied.Args, copiedNode)
		}
	}

	for _, node := range n.Keywords {
		if node != nil {
			copiedNode := node.Copy()
			copiedNode.Parent = copied
			copied.Keywords = append(copied.Keywords, copiedNode)
		}
	}

	for _, node := range n.Decorator {
		if node != nil {
			copiedNode := node.Copy()
			copiedNode.Parent = copied
			copied.Decorator = append(copied.Decorator, copiedNode)
		}
	}

	for _, node := range n.Bases {
		if node != nil {
			copiedNode := node.Copy()
			copiedNode.Parent = copied
			copied.Bases = append(copied.Bases, copiedNode)
		}
	}

	return copied
}
