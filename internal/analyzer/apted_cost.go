package analyzer

import (
	"strings"
)

// CostModel defines the interface for calculating edit operation costs
type CostModel interface {
	// Insert returns the cost of inserting a node
	Insert(node *TreeNode) float64

	// Delete returns the cost of deleting a node
	Delete(node *TreeNode) float64

	// Rename returns the cost of renaming node1 to node2
	Rename(node1, node2 *TreeNode) float64
}

// DefaultCostModel implements a uniform cost model where all operations cost 1.0
type DefaultCostModel struct{}

// NewDefaultCostModel creates a new default cost model
func NewDefaultCostModel() *DefaultCostModel {
	return &DefaultCostModel{}
}

// Insert returns the cost of inserting a node (always 1.0)
func (c *DefaultCostModel) Insert(node *TreeNode) float64 {
	return 1.0
}

// Delete returns the cost of deleting a node (always 1.0)
func (c *DefaultCostModel) Delete(node *TreeNode) float64 {
	return 1.0
}

// Rename returns the cost of renaming node1 to node2
func (c *DefaultCostModel) Rename(node1, node2 *TreeNode) float64 {
	if node1 == nil || node2 == nil {
		return 1.0
	}

	// If labels are identical, no cost for rename
	if node1.Label == node2.Label {
		return 0.0
	}

	return 1.0
}

// PythonCostModel implements a Python-aware cost model with different costs for different node types
type PythonCostModel struct {
	// Base costs for different operations
	BaseInsertCost float64
	BaseDeleteCost float64
	BaseRenameCost float64

	// Whether to ignore differences in literal values
	IgnoreLiterals bool

	// Whether to ignore differences in identifier names
	IgnoreIdentifiers bool

	// Whether to reduce weight for boilerplate nodes (type annotations, decorators, Field() calls)
	ReduceBoilerplateWeight bool

	// Multiplier for boilerplate nodes (default: 0.1)
	BoilerplateMultiplier float64
}

// NewPythonCostModel creates a new Python-aware cost model with default settings
func NewPythonCostModel() *PythonCostModel {
	return &PythonCostModel{
		BaseInsertCost:          1.0,
		BaseDeleteCost:          1.0,
		BaseRenameCost:          1.0,
		IgnoreLiterals:          false,
		IgnoreIdentifiers:       false,
		ReduceBoilerplateWeight: true, // Enable by default to reduce false positives
		BoilerplateMultiplier:   0.1,
	}
}

// NewPythonCostModelWithConfig creates a Python cost model with custom configuration
func NewPythonCostModelWithConfig(ignoreLiterals, ignoreIdentifiers bool) *PythonCostModel {
	return &PythonCostModel{
		BaseInsertCost:          1.0,
		BaseDeleteCost:          1.0,
		BaseRenameCost:          1.0,
		IgnoreLiterals:          ignoreLiterals,
		IgnoreIdentifiers:       ignoreIdentifiers,
		ReduceBoilerplateWeight: true, // Enable by default to reduce false positives
		BoilerplateMultiplier:   0.1,
	}
}

// NewPythonCostModelWithBoilerplateConfig creates a Python cost model with full configuration
func NewPythonCostModelWithBoilerplateConfig(ignoreLiterals, ignoreIdentifiers, reduceBoilerplate bool, boilerplateMultiplier float64) *PythonCostModel {
	if boilerplateMultiplier <= 0 {
		boilerplateMultiplier = 0.1
	}
	return &PythonCostModel{
		BaseInsertCost:          1.0,
		BaseDeleteCost:          1.0,
		BaseRenameCost:          1.0,
		IgnoreLiterals:          ignoreLiterals,
		IgnoreIdentifiers:       ignoreIdentifiers,
		ReduceBoilerplateWeight: reduceBoilerplate,
		BoilerplateMultiplier:   boilerplateMultiplier,
	}
}

// Insert returns the cost of inserting a node
func (c *PythonCostModel) Insert(node *TreeNode) float64 {
	if node == nil {
		return c.BaseInsertCost
	}

	// Different costs based on node type
	multiplier := c.getNodeTypeMultiplier(node.Label)
	return c.BaseInsertCost * multiplier
}

// Delete returns the cost of deleting a node
func (c *PythonCostModel) Delete(node *TreeNode) float64 {
	if node == nil {
		return c.BaseDeleteCost
	}

	// Different costs based on node type
	multiplier := c.getNodeTypeMultiplier(node.Label)
	return c.BaseDeleteCost * multiplier
}

// Rename returns the cost of renaming node1 to node2
func (c *PythonCostModel) Rename(node1, node2 *TreeNode) float64 {
	if node1 == nil || node2 == nil {
		return c.BaseRenameCost
	}

	// If labels are identical, no cost
	if node1.Label == node2.Label {
		return 0.0
	}

	// Apply ignore patterns
	if c.shouldIgnoreDifference(node1.Label, node2.Label) {
		return 0.0
	}

	// Check if both nodes are of similar types
	similarity := c.calculateLabelSimilarity(node1.Label, node2.Label)

	// Scale rename cost based on similarity
	return c.BaseRenameCost * (1.0 - similarity)
}

// getNodeTypeMultiplier returns a cost multiplier based on the node type
func (c *PythonCostModel) getNodeTypeMultiplier(label string) float64 {
	// Boilerplate nodes (type annotations, decorators, Field() calls) get very low weight
	// This reduces false positives for framework patterns like dataclasses and Pydantic
	// Uses the shared IsBoilerplateLabel function to avoid duplication
	if c.ReduceBoilerplateWeight && IsBoilerplateLabel(label) {
		return c.BoilerplateMultiplier
	}

	// Structural nodes are more expensive to modify
	if c.isStructuralNode(label) {
		return 1.5
	}

	// Control flow nodes are expensive
	if c.isControlFlowNode(label) {
		return 1.3
	}

	// Expression nodes are less expensive
	if c.isExpressionNode(label) {
		return 0.8
	}

	// Literals and identifiers can be very cheap if configured to ignore
	if c.isLiteralNode(label) && c.IgnoreLiterals {
		return 0.1
	}

	if c.isIdentifierNode(label) && c.IgnoreIdentifiers {
		return 0.2
	}

	return 1.0 // Default multiplier
}

// isStructuralNode checks if a node represents a structural element
func (c *PythonCostModel) isStructuralNode(label string) bool {
	structuralNodes := []string{
		"FunctionDef", "AsyncFunctionDef", "ClassDef", "Module",
		"Arguments", "Arg", "Decorator",
	}

	for _, nodeType := range structuralNodes {
		if strings.HasPrefix(label, nodeType) {
			return true
		}
	}

	return false
}

// isControlFlowNode checks if a node represents a control flow element
func (c *PythonCostModel) isControlFlowNode(label string) bool {
	controlFlowNodes := []string{
		"If", "For", "While", "Try", "With", "Match",
		"AsyncFor", "AsyncWith", "ExceptHandler",
		"Break", "Continue", "Return", "Raise",
	}

	for _, nodeType := range controlFlowNodes {
		if strings.HasPrefix(label, nodeType) {
			return true
		}
	}

	return false
}

// isExpressionNode checks if a node represents an expression
func (c *PythonCostModel) isExpressionNode(label string) bool {
	expressionNodes := []string{
		"BinOp", "UnaryOp", "BoolOp", "Compare", "Call",
		"Attribute", "Subscript", "List", "Tuple", "Dict", "Set",
		"Lambda", "IfExp", "ListComp", "SetComp", "DictComp", "GeneratorExp",
	}

	for _, nodeType := range expressionNodes {
		if strings.HasPrefix(label, nodeType) {
			return true
		}
	}

	return false
}

// isLiteralNode checks if a node represents a literal value
func (c *PythonCostModel) isLiteralNode(label string) bool {
	return strings.HasPrefix(label, "Constant(")
}

// isIdentifierNode checks if a node represents an identifier
func (c *PythonCostModel) isIdentifierNode(label string) bool {
	return strings.HasPrefix(label, "Name(")
}

// shouldIgnoreDifference determines if the difference between two labels should be ignored
func (c *PythonCostModel) shouldIgnoreDifference(label1, label2 string) bool {
	// Ignore literal differences if configured
	if c.IgnoreLiterals && c.isLiteralNode(label1) && c.isLiteralNode(label2) {
		return true
	}

	// Ignore identifier differences if configured
	if c.IgnoreIdentifiers && c.isIdentifierNode(label1) && c.isIdentifierNode(label2) {
		return true
	}

	return false
}

// calculateLabelSimilarity calculates similarity between two node labels.
// Returns a value in [0, 1] that reduces the rename cost.
// Lower values = higher rename cost = stricter clone detection.
func (c *PythonCostModel) calculateLabelSimilarity(label1, label2 string) float64 {
	// Extract base node types (remove parenthetical content)
	baseType1 := c.extractBaseNodeType(label1)
	baseType2 := c.extractBaseNodeType(label2)

	// If base types are identical, moderate similarity
	// This ensures renaming Name(foo) -> Name(bar) still has meaningful cost (0.7)
	if baseType1 == baseType2 {
		return 0.3
	}

	// Check for related node types (e.g., For/AsyncFor)
	if c.areRelatedNodeTypes(baseType1, baseType2) {
		return 0.2
	}

	// Check for same category (e.g., both control flow)
	if c.areSameCategory(baseType1, baseType2) {
		return 0.1
	}

	return 0.0 // No similarity - full rename cost
}

// extractBaseNodeType extracts the base node type from a label
func (c *PythonCostModel) extractBaseNodeType(label string) string {
	if idx := strings.Index(label, "("); idx != -1 {
		return label[:idx]
	}
	return label
}

// areRelatedNodeTypes checks if two node types are related
func (c *PythonCostModel) areRelatedNodeTypes(type1, type2 string) bool {
	relatedPairs := [][2]string{
		{"FunctionDef", "AsyncFunctionDef"},
		{"For", "AsyncFor"},
		{"With", "AsyncWith"},
		{"BinOp", "UnaryOp"},
		{"List", "Tuple"},
		{"ListComp", "GeneratorExp"},
		{"If", "IfExp"},
	}

	for _, pair := range relatedPairs {
		if (type1 == pair[0] && type2 == pair[1]) || (type1 == pair[1] && type2 == pair[0]) {
			return true
		}
	}

	return false
}

// areSameCategory checks if two node types belong to the same category
func (c *PythonCostModel) areSameCategory(type1, type2 string) bool {
	if c.isStructuralNode(type1) && c.isStructuralNode(type2) {
		return true
	}

	if c.isControlFlowNode(type1) && c.isControlFlowNode(type2) {
		return true
	}

	if c.isExpressionNode(type1) && c.isExpressionNode(type2) {
		return true
	}

	return false
}

// WeightedCostModel allows custom weights for different operation types
type WeightedCostModel struct {
	InsertWeight  float64
	DeleteWeight  float64
	RenameWeight  float64
	BaseCostModel CostModel
}

// NewWeightedCostModel creates a new weighted cost model
func NewWeightedCostModel(insertWeight, deleteWeight, renameWeight float64, baseCostModel CostModel) *WeightedCostModel {
	return &WeightedCostModel{
		InsertWeight:  insertWeight,
		DeleteWeight:  deleteWeight,
		RenameWeight:  renameWeight,
		BaseCostModel: baseCostModel,
	}
}

// Insert returns the weighted cost of inserting a node
func (c *WeightedCostModel) Insert(node *TreeNode) float64 {
	return c.InsertWeight * c.BaseCostModel.Insert(node)
}

// Delete returns the weighted cost of deleting a node
func (c *WeightedCostModel) Delete(node *TreeNode) float64 {
	return c.DeleteWeight * c.BaseCostModel.Delete(node)
}

// Rename returns the weighted cost of renaming node1 to node2
func (c *WeightedCostModel) Rename(node1, node2 *TreeNode) float64 {
	return c.RenameWeight * c.BaseCostModel.Rename(node1, node2)
}
