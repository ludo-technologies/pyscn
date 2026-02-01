package analyzer

import (
	"slices"
	"strings"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

// FrameworkPattern represents a detected Python framework pattern
type FrameworkPattern int

const (
	// PatternNone indicates no framework pattern was detected
	PatternNone FrameworkPattern = iota
	// PatternDataclass indicates a @dataclass decorated class
	PatternDataclass
	// PatternPydanticModel indicates a Pydantic BaseModel class
	PatternPydanticModel
	// PatternNamedTuple indicates a typing.NamedTuple subclass
	PatternNamedTuple
	// PatternTypedDict indicates a typing.TypedDict subclass
	PatternTypedDict
	// PatternAttrs indicates an @attrs/@attr.s decorated class
	PatternAttrs
)

// String returns the string representation of the framework pattern
func (fp FrameworkPattern) String() string {
	switch fp {
	case PatternDataclass:
		return "dataclass"
	case PatternPydanticModel:
		return "pydantic"
	case PatternNamedTuple:
		return "namedtuple"
	case PatternTypedDict:
		return "typeddict"
	case PatternAttrs:
		return "attrs"
	default:
		return "none"
	}
}

// PatternDetector detects framework-specific patterns in Python AST nodes
type PatternDetector struct {
	// EnableDataclassDetection enables detection of @dataclass patterns
	EnableDataclassDetection bool
	// EnablePydanticDetection enables detection of Pydantic BaseModel patterns
	EnablePydanticDetection bool
	// EnableNamedTupleDetection enables detection of NamedTuple patterns
	EnableNamedTupleDetection bool
	// EnableTypedDictDetection enables detection of TypedDict patterns
	EnableTypedDictDetection bool
	// EnableAttrsDetection enables detection of attrs patterns
	EnableAttrsDetection bool
}

// NewPatternDetector creates a new pattern detector with all detections enabled
func NewPatternDetector() *PatternDetector {
	return &PatternDetector{
		EnableDataclassDetection:  true,
		EnablePydanticDetection:   true,
		EnableNamedTupleDetection: true,
		EnableTypedDictDetection:  true,
		EnableAttrsDetection:      true,
	}
}

// DetectPattern analyzes an AST node and returns the detected framework pattern
func (pd *PatternDetector) DetectPattern(node *parser.Node) FrameworkPattern {
	if node == nil {
		return PatternNone
	}

	// Only check class definitions
	if node.Type != parser.NodeClassDef {
		return PatternNone
	}

	// Check decorators first
	if pd.EnableDataclassDetection && pd.hasDataclassDecorator(node) {
		return PatternDataclass
	}

	if pd.EnableAttrsDetection && pd.hasAttrsDecorator(node) {
		return PatternAttrs
	}

	// Check base classes
	if pd.EnablePydanticDetection && pd.hasPydanticBase(node) {
		return PatternPydanticModel
	}

	if pd.EnableNamedTupleDetection && pd.hasNamedTupleBase(node) {
		return PatternNamedTuple
	}

	if pd.EnableTypedDictDetection && pd.hasTypedDictBase(node) {
		return PatternTypedDict
	}

	return PatternNone
}

// IsBoilerplateNode checks if a node is boilerplate (type annotation, Field(), etc.)
func (pd *PatternDetector) IsBoilerplateNode(node *parser.Node) bool {
	if node == nil {
		return false
	}

	// Type annotations (AnnAssign nodes)
	if node.Type == parser.NodeAnnAssign {
		return true
	}

	// Decorator nodes
	if node.Type == parser.NodeDecorator {
		return true
	}

	// Field() calls (Pydantic, dataclasses, attrs)
	if pd.isFieldCall(node) {
		return true
	}

	// Type hint nodes within annotations
	if pd.isTypeHintNode(node) {
		return true
	}

	return false
}

// IsBoilerplateLabel checks if a tree node label represents boilerplate
func (pd *PatternDetector) IsBoilerplateLabel(label string) bool {
	// Check for AnnAssign (annotated assignment - type hints)
	if strings.HasPrefix(label, "AnnAssign") {
		return true
	}

	// Check for Decorator
	if strings.HasPrefix(label, "Decorator") {
		return true
	}

	// Check for common type annotation patterns
	boilerplatePatterns := []string{
		"type", // type hints
		"generic_type",
		"type_parameter",
	}

	labelLower := strings.ToLower(label)
	for _, pattern := range boilerplatePatterns {
		if strings.Contains(labelLower, pattern) {
			return true
		}
	}

	return false
}

// CountBoilerplateNodes counts the number of boilerplate nodes in a tree
func (pd *PatternDetector) CountBoilerplateNodes(node *parser.Node) (boilerplate, total int) {
	if node == nil {
		return 0, 0
	}

	total = 1
	if pd.IsBoilerplateNode(node) {
		boilerplate = 1
	}

	// Count children
	for _, child := range node.Children {
		b, t := pd.CountBoilerplateNodes(child)
		boilerplate += b
		total += t
	}

	// Count body nodes
	for _, bodyNode := range node.Body {
		b, t := pd.CountBoilerplateNodes(bodyNode)
		boilerplate += b
		total += t
	}

	// Count orelse nodes
	for _, orelseNode := range node.Orelse {
		b, t := pd.CountBoilerplateNodes(orelseNode)
		boilerplate += b
		total += t
	}

	// Count decorators
	for _, decorator := range node.Decorator {
		b, t := pd.CountBoilerplateNodes(decorator)
		boilerplate += b
		total += t
	}

	return boilerplate, total
}

// CalculateSemanticContentRatio calculates the ratio of non-boilerplate nodes
// Returns a value between 0.0 (all boilerplate) and 1.0 (no boilerplate)
func (pd *PatternDetector) CalculateSemanticContentRatio(node *parser.Node) float64 {
	boilerplate, total := pd.CountBoilerplateNodes(node)
	if total == 0 {
		return 1.0 // Empty node is considered semantic
	}

	semantic := total - boilerplate
	return float64(semantic) / float64(total)
}

// hasDataclassDecorator checks if a class has @dataclass decorator
func (pd *PatternDetector) hasDataclassDecorator(node *parser.Node) bool {
	for _, decorator := range node.Decorator {
		name := pd.getDecoratorName(decorator)
		if name == "dataclass" || name == "dataclasses.dataclass" {
			return true
		}
	}
	return false
}

// hasAttrsDecorator checks if a class has @attrs/@attr.s decorator
func (pd *PatternDetector) hasAttrsDecorator(node *parser.Node) bool {
	attrsNames := []string{"attrs", "attr.s", "attr.attrs", "define", "attr.define", "frozen", "attr.frozen"}
	for _, decorator := range node.Decorator {
		name := pd.getDecoratorName(decorator)
		if slices.Contains(attrsNames, name) {
			return true
		}
	}
	return false
}

// hasPydanticBase checks if a class inherits from Pydantic BaseModel
func (pd *PatternDetector) hasPydanticBase(node *parser.Node) bool {
	pydanticBases := []string{"BaseModel", "pydantic.BaseModel", "BaseSettings", "pydantic.BaseSettings"}
	for _, base := range node.Bases {
		name := pd.getBaseName(base)
		if slices.Contains(pydanticBases, name) {
			return true
		}
	}
	return false
}

// hasNamedTupleBase checks if a class inherits from NamedTuple
func (pd *PatternDetector) hasNamedTupleBase(node *parser.Node) bool {
	for _, base := range node.Bases {
		name := pd.getBaseName(base)
		if name == "NamedTuple" || name == "typing.NamedTuple" {
			return true
		}
	}
	return false
}

// hasTypedDictBase checks if a class inherits from TypedDict
func (pd *PatternDetector) hasTypedDictBase(node *parser.Node) bool {
	for _, base := range node.Bases {
		name := pd.getBaseName(base)
		if name == "TypedDict" || name == "typing.TypedDict" || name == "typing_extensions.TypedDict" {
			return true
		}
	}
	return false
}

// getDecoratorName extracts the name from a decorator node
func (pd *PatternDetector) getDecoratorName(decorator *parser.Node) string {
	if decorator == nil {
		return ""
	}

	// Direct name (e.g., @dataclass)
	if decorator.Type == parser.NodeName {
		return decorator.Name
	}

	// Attribute access (e.g., @dataclasses.dataclass)
	if decorator.Type == parser.NodeAttribute {
		return pd.getAttributeFullName(decorator)
	}

	// Call (e.g., @dataclass())
	if decorator.Type == parser.NodeCall {
		// Get the function being called
		for _, child := range decorator.Children {
			if child.Type == parser.NodeName {
				return child.Name
			}
			if child.Type == parser.NodeAttribute {
				return pd.getAttributeFullName(child)
			}
		}
	}

	// Decorator wrapper
	if decorator.Type == parser.NodeDecorator {
		for _, child := range decorator.Children {
			name := pd.getDecoratorName(child)
			if name != "" {
				return name
			}
		}
	}

	return ""
}

// getAttributeFullName gets the full dotted name from an Attribute node
func (pd *PatternDetector) getAttributeFullName(attr *parser.Node) string {
	if attr == nil || attr.Type != parser.NodeAttribute {
		return ""
	}

	// Attribute nodes have a value (the object) and attr (the attribute name)
	attrName := attr.Name
	if attrName == "" {
		// Try to get from children or Value field
		if attr.Value != nil {
			if str, ok := attr.Value.(string); ok {
				attrName = str
			}
		}
	}

	// Get the object part
	for _, child := range attr.Children {
		if child.Type == parser.NodeName {
			return child.Name + "." + attrName
		}
		if child.Type == parser.NodeAttribute {
			return pd.getAttributeFullName(child) + "." + attrName
		}
	}

	return attrName
}

// getBaseName extracts the name from a base class node
func (pd *PatternDetector) getBaseName(base *parser.Node) string {
	if base == nil {
		return ""
	}

	if base.Type == parser.NodeName {
		return base.Name
	}

	if base.Type == parser.NodeAttribute {
		return pd.getAttributeFullName(base)
	}

	return ""
}

// isFieldCall checks if a node is a Field() call (Pydantic, dataclasses, attrs)
func (pd *PatternDetector) isFieldCall(node *parser.Node) bool {
	if node == nil || node.Type != parser.NodeCall {
		return false
	}

	fieldNames := []string{"Field", "field", "Factory", "attrib", "attr.ib"}
	attrFieldNames := []string{
		"dataclasses.field",
		"pydantic.Field",
		"attr.ib",
		"attr.attrib",
		"attrs.field",
	}

	// Check if the call is to Field, field, or Factory
	for _, child := range node.Children {
		if child.Type == parser.NodeName {
			if slices.Contains(fieldNames, child.Name) {
				return true
			}
		}
		if child.Type == parser.NodeAttribute {
			fullName := pd.getAttributeFullName(child)
			if slices.Contains(attrFieldNames, fullName) {
				return true
			}
		}
	}

	return false
}

// isTypeHintNode checks if a node is part of type hint syntax
func (pd *PatternDetector) isTypeHintNode(node *parser.Node) bool {
	if node == nil {
		return false
	}

	// Generic type nodes
	if node.Type == parser.NodeGenericType || node.Type == parser.NodeTypeParameter || node.Type == parser.NodeTypeNode {
		return true
	}

	typeNames := []string{
		"List", "Dict", "Set", "Tuple", "Optional", "Union",
		"Callable", "Type", "Sequence", "Mapping", "Iterable",
		"Iterator", "Generator", "Coroutine", "AsyncIterator",
		"AsyncGenerator", "Awaitable", "ContextManager",
		"AsyncContextManager", "Pattern", "Match", "IO",
		"TextIO", "BinaryIO", "Any", "NoReturn", "ClassVar",
		"Final", "Literal", "TypeVar", "Generic", "Protocol",
		"Annotated", "TypeAlias", "TypeGuard", "Never",
	}

	// Subscript used for generic types (e.g., List[int], Optional[str])
	if node.Type == parser.NodeSubscript {
		// Check if it's a type annotation context
		for _, child := range node.Children {
			if child.Type == parser.NodeName {
				if slices.Contains(typeNames, child.Name) {
					return true
				}
			}
		}
	}

	return false
}

// CountBoilerplateInTreeNode counts boilerplate nodes in a TreeNode
func (pd *PatternDetector) CountBoilerplateInTreeNode(node *TreeNode) (boilerplate, total int) {
	if node == nil {
		return 0, 0
	}

	total = 1
	if pd.IsBoilerplateLabel(node.Label) {
		boilerplate = 1
	}

	// Recursively count children
	for _, child := range node.Children {
		b, t := pd.CountBoilerplateInTreeNode(child)
		boilerplate += b
		total += t
	}

	return boilerplate, total
}

// CalculateSemanticContentRatioFromTree calculates semantic content ratio from TreeNode
func (pd *PatternDetector) CalculateSemanticContentRatioFromTree(node *TreeNode) float64 {
	boilerplate, total := pd.CountBoilerplateInTreeNode(node)
	if total == 0 {
		return 1.0
	}

	semantic := total - boilerplate
	return float64(semantic) / float64(total)
}
