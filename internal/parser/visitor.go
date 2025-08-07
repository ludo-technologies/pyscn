package parser

import (
	"fmt"
	"io"
	"strings"
)

// Visitor defines the interface for visiting AST nodes
type Visitor interface {
	// Visit is called for each node in the AST
	// Return false to stop traversal
	Visit(node *Node) bool
}

// Accept implements the visitor pattern for AST nodes
func (n *Node) Accept(visitor Visitor) {
	if n == nil {
		return
	}
	
	if !visitor.Visit(n) {
		return
	}
	
	// Visit all children
	for _, child := range n.GetChildren() {
		if child != nil {
			child.Accept(visitor)
		}
	}
}

// FuncVisitor is a visitor that uses a function
type FuncVisitor struct {
	fn func(*Node) bool
}

// NewFuncVisitor creates a visitor from a function
func NewFuncVisitor(fn func(*Node) bool) *FuncVisitor {
	return &FuncVisitor{fn: fn}
}

// Visit implements the Visitor interface
func (v *FuncVisitor) Visit(node *Node) bool {
	return v.fn(node)
}

// CollectorVisitor collects nodes matching a predicate
type CollectorVisitor struct {
	predicate func(*Node) bool
	nodes     []*Node
}

// NewCollectorVisitor creates a visitor that collects matching nodes
func NewCollectorVisitor(predicate func(*Node) bool) *CollectorVisitor {
	return &CollectorVisitor{
		predicate: predicate,
		nodes:     []*Node{},
	}
}

// Visit implements the Visitor interface
func (v *CollectorVisitor) Visit(node *Node) bool {
	if v.predicate(node) {
		v.nodes = append(v.nodes, node)
	}
	return true
}

// GetNodes returns the collected nodes
func (v *CollectorVisitor) GetNodes() []*Node {
	return v.nodes
}

// PrinterVisitor prints the AST structure
type PrinterVisitor struct {
	writer io.Writer
	indent int
	prefix string
}

// NewPrinterVisitor creates a visitor that prints the AST
func NewPrinterVisitor(w io.Writer) *PrinterVisitor {
	return &PrinterVisitor{
		writer: w,
		indent: 0,
		prefix: "  ",
	}
}

// Visit implements the Visitor interface
func (v *PrinterVisitor) Visit(node *Node) bool {
	// Print indentation
	fmt.Fprint(v.writer, strings.Repeat(v.prefix, v.indent))
	
	// Print node information
	switch {
	case node.Name != "":
		fmt.Fprintf(v.writer, "%s: %s\n", node.Type, node.Name)
	case node.Value != nil:
		fmt.Fprintf(v.writer, "%s: %v\n", node.Type, node.Value)
	case node.Op != "":
		fmt.Fprintf(v.writer, "%s: %s\n", node.Type, node.Op)
	default:
		fmt.Fprintf(v.writer, "%s\n", node.Type)
	}
	
	// Increase indent for children
	v.indent++
	
	// Visit children manually to control indentation
	for _, child := range node.GetChildren() {
		if child != nil {
			child.Accept(v)
		}
	}
	
	// Restore indent
	v.indent--
	
	return false // We handle children manually
}

// TransformVisitor transforms nodes in the AST
type TransformVisitor struct {
	transformer func(*Node) *Node
}

// NewTransformVisitor creates a visitor that transforms nodes
func NewTransformVisitor(transformer func(*Node) *Node) *TransformVisitor {
	return &TransformVisitor{
		transformer: transformer,
	}
}

// Visit implements the Visitor interface
func (v *TransformVisitor) Visit(node *Node) bool {
	// Transform current node
	transformed := v.transformer(node)
	
	// If transformation returned a different node, replace fields
	if transformed != node && transformed != nil {
		node.Type = transformed.Type
		node.Value = transformed.Value
		node.Name = transformed.Name
		node.Op = transformed.Op
		// Note: Not replacing children/body to avoid breaking structure
	}
	
	return true
}

// StatisticsVisitor collects statistics about the AST
type StatisticsVisitor struct {
	NodeCounts map[NodeType]int
	TotalNodes int
	MaxDepth   int
	curDepth   int
}

// NewStatisticsVisitor creates a visitor that collects statistics
func NewStatisticsVisitor() *StatisticsVisitor {
	return &StatisticsVisitor{
		NodeCounts: make(map[NodeType]int),
		TotalNodes: 0,
		MaxDepth:   0,
		curDepth:   0,
	}
}

// Visit implements the Visitor interface
func (v *StatisticsVisitor) Visit(node *Node) bool {
	v.TotalNodes++
	v.NodeCounts[node.Type]++
	
	v.curDepth++
	if v.curDepth > v.MaxDepth {
		v.MaxDepth = v.curDepth
	}
	
	// Visit children
	for _, child := range node.GetChildren() {
		if child != nil {
			child.Accept(v)
		}
	}
	
	v.curDepth--
	return false // We handle children manually
}

// PathVisitor tracks the path to each visited node
type PathVisitor struct {
	path    []*Node
	visitor func(node *Node, path []*Node) bool
}

// NewPathVisitor creates a visitor that tracks the path to each node
func NewPathVisitor(visitor func(node *Node, path []*Node) bool) *PathVisitor {
	return &PathVisitor{
		path:    []*Node{},
		visitor: visitor,
	}
}

// Visit implements the Visitor interface
func (v *PathVisitor) Visit(node *Node) bool {
	// Add current node to path
	v.path = append(v.path, node)
	
	// Call the visitor function with current path
	shouldContinue := v.visitor(node, v.path)
	
	if shouldContinue {
		// Visit children
		for _, child := range node.GetChildren() {
			if child != nil {
				child.Accept(v)
			}
		}
	}
	
	// Remove current node from path
	v.path = v.path[:len(v.path)-1]
	
	return false // We handle children manually
}

// DepthFirstVisitor performs depth-first traversal
type DepthFirstVisitor struct {
	preOrder  func(*Node) bool
	postOrder func(*Node)
}

// NewDepthFirstVisitor creates a depth-first visitor
func NewDepthFirstVisitor(preOrder func(*Node) bool, postOrder func(*Node)) *DepthFirstVisitor {
	return &DepthFirstVisitor{
		preOrder:  preOrder,
		postOrder: postOrder,
	}
}

// Visit implements the Visitor interface
func (v *DepthFirstVisitor) Visit(node *Node) bool {
	// Pre-order visit
	if v.preOrder != nil {
		if !v.preOrder(node) {
			return false
		}
	}
	
	// Visit children
	for _, child := range node.GetChildren() {
		if child != nil {
			child.Accept(v)
		}
	}
	
	// Post-order visit
	if v.postOrder != nil {
		v.postOrder(node)
	}
	
	return false // We handle children manually
}

// ValidatorVisitor validates the AST structure
type ValidatorVisitor struct {
	errors []string
}

// NewValidatorVisitor creates a visitor that validates the AST
func NewValidatorVisitor() *ValidatorVisitor {
	return &ValidatorVisitor{
		errors: []string{},
	}
}

// Visit implements the Visitor interface
func (v *ValidatorVisitor) Visit(node *Node) bool {
	// Validate node structure
	switch node.Type {
	case NodeFunctionDef, NodeAsyncFunctionDef:
		if node.Name == "" {
			v.errors = append(v.errors, fmt.Sprintf("Function at %+v missing name", node.Location))
		}
		if len(node.Body) == 0 {
			v.errors = append(v.errors, fmt.Sprintf("Function '%s' at %+v has empty body", node.Name, node.Location))
		}
		
	case NodeClassDef:
		if node.Name == "" {
			v.errors = append(v.errors, fmt.Sprintf("Class at %+v missing name", node.Location))
		}
		
	case NodeIf:
		if node.Test == nil {
			v.errors = append(v.errors, fmt.Sprintf("If statement at %+v missing condition", node.Location))
		}
		if len(node.Body) == 0 {
			v.errors = append(v.errors, fmt.Sprintf("If statement at %+v has empty body", node.Location))
		}
		
	case NodeFor, NodeAsyncFor:
		if len(node.Targets) == 0 {
			v.errors = append(v.errors, fmt.Sprintf("For loop at %+v missing target", node.Location))
		}
		if node.Iter == nil {
			v.errors = append(v.errors, fmt.Sprintf("For loop at %+v missing iterator", node.Location))
		}
		
	case NodeWhile:
		if node.Test == nil {
			v.errors = append(v.errors, fmt.Sprintf("While loop at %+v missing condition", node.Location))
		}
		
	case NodeBinOp:
		if node.Left == nil || node.Right == nil {
			v.errors = append(v.errors, fmt.Sprintf("Binary operation at %+v missing operand", node.Location))
		}
		if node.Op == "" {
			v.errors = append(v.errors, fmt.Sprintf("Binary operation at %+v missing operator", node.Location))
		}
		
	case NodeUnaryOp:
		if node.Value == nil {
			v.errors = append(v.errors, fmt.Sprintf("Unary operation at %+v missing operand", node.Location))
		}
		if node.Op == "" {
			v.errors = append(v.errors, fmt.Sprintf("Unary operation at %+v missing operator", node.Location))
		}
	}
	
	// Check parent-child relationships
	for _, child := range node.GetChildren() {
		if child != nil && child.Parent != node {
			// Only report if parent is completely wrong, not just nil
			if child.Parent != nil {
				v.errors = append(v.errors, fmt.Sprintf("Node %s at %+v has incorrect parent (expected %s, got %s)", 
					child.Type, child.Location, node.Type, child.Parent.Type))
			}
		}
	}
	
	return true
}

// GetErrors returns validation errors
func (v *ValidatorVisitor) GetErrors() []string {
	return v.errors
}

// IsValid returns true if no errors were found
func (v *ValidatorVisitor) IsValid() bool {
	return len(v.errors) == 0
}

// SimplifierVisitor simplifies the AST by removing unnecessary nodes
type SimplifierVisitor struct {
	simplified bool
}

// NewSimplifierVisitor creates a visitor that simplifies the AST
func NewSimplifierVisitor() *SimplifierVisitor {
	return &SimplifierVisitor{}
}

// Visit implements the Visitor interface
func (v *SimplifierVisitor) Visit(node *Node) bool {
	// Simplify constant expressions
	if node.Type == NodeBinOp && node.Op == "+" {
		if isConstant(node.Left) && isConstant(node.Right) {
			// Simplify constant addition
			if leftVal, leftOk := getConstantValue(node.Left); leftOk {
				if rightVal, rightOk := getConstantValue(node.Right); rightOk {
					node.Type = NodeConstant
					node.Value = addConstants(leftVal, rightVal)
					node.Left = nil
					node.Right = nil
					node.Op = ""
					v.simplified = true
				}
			}
		}
	}
	
	// Remove pass statements from bodies
	if len(node.Body) > 0 {
		filtered := []*Node{}
		for _, stmt := range node.Body {
			if stmt.Type != NodePass {
				filtered = append(filtered, stmt)
			}
		}
		if len(filtered) != len(node.Body) {
			node.Body = filtered
			v.simplified = true
		}
	}
	
	return true
}

// WasSimplified returns true if any simplification was performed
func (v *SimplifierVisitor) WasSimplified() bool {
	return v.simplified
}

// Helper functions for SimplifierVisitor
func isConstant(node *Node) bool {
	return node != nil && node.Type == NodeConstant
}

func getConstantValue(node *Node) (interface{}, bool) {
	if node == nil || node.Type != NodeConstant {
		return nil, false
	}
	return node.Value, true
}

func addConstants(left, right interface{}) interface{} {
	switch l := left.(type) {
	case int64:
		if r, ok := right.(int64); ok {
			return l + r
		}
	case float64:
		if r, ok := right.(float64); ok {
			return l + r
		}
	case string:
		if r, ok := right.(string); ok {
			return l + r
		}
	}
	return nil
}