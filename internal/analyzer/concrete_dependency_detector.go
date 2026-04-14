package analyzer

import (
	"fmt"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// ConcreteDependencyDetector detects concrete dependency anti-patterns
type ConcreteDependencyDetector struct {
	abstractPrefixes []string
	abstractSuffixes []string
}

// NewConcreteDependencyDetector creates a new concrete dependency detector
func NewConcreteDependencyDetector() *ConcreteDependencyDetector {
	return &ConcreteDependencyDetector{
		// Common abstract class/interface naming patterns
		abstractPrefixes: []string{"Abstract", "Base", "I"},
		abstractSuffixes: []string{"Interface", "Protocol", "ABC", "Base", "Abstract"},
	}
}

// Analyze detects concrete dependencies in the given AST
func (d *ConcreteDependencyDetector) Analyze(ast *parser.Node, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	// Detect type hints with concrete classes in constructors
	typeHintFindings := d.detectConcreteTypeHints(ast, filePath)
	findings = append(findings, typeHintFindings...)

	// Detect direct instantiation in constructors
	instantiationFindings := d.detectDirectInstantiation(ast, filePath)
	findings = append(findings, instantiationFindings...)

	return findings
}

// detectConcreteTypeHints detects type hints with concrete classes in constructor parameters
func (d *ConcreteDependencyDetector) detectConcreteTypeHints(ast *parser.Node, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	classes := ast.FindByType(parser.NodeClassDef)

	for _, class := range classes {
		initMethod := FindInitMethod(class)
		if initMethod == nil {
			continue
		}

		// Analyze parameter type hints
		for _, arg := range initMethod.Args {
			if arg == nil || arg.Type != parser.NodeArg {
				continue
			}

			// Skip self/cls
			if arg.Name == "self" || arg.Name == "cls" {
				continue
			}

			// Check type annotation
			typeName := d.extractTypeHintName(arg)
			if typeName != "" && d.isConcreteType(typeName) {
				finding := domain.DIAntipatternFinding{
					Type:       domain.DIAntipatternConcreteDependency,
					Subtype:    string(domain.ConcreteDepTypeHint),
					Severity:   domain.DIAntipatternSeverityInfo,
					ClassName:  class.Name,
					MethodName: "__init__",
					Location: domain.SourceLocation{
						FilePath:  filePath,
						StartLine: arg.Location.StartLine,
						EndLine:   arg.Location.EndLine,
						StartCol:  arg.Location.StartCol,
						EndCol:    arg.Location.EndCol,
					},
					Description: fmt.Sprintf("Parameter '%s' has concrete type hint '%s'", arg.Name, typeName),
					Suggestion:  "Consider using an abstract type (Protocol, ABC, or interface) instead of a concrete class",
					Details: map[string]interface{}{
						"parameter_name": arg.Name,
						"type_name":      typeName,
					},
				}
				findings = append(findings, finding)
			}
		}
	}

	return findings
}

// detectDirectInstantiation detects direct class instantiation in constructors
func (d *ConcreteDependencyDetector) detectDirectInstantiation(ast *parser.Node, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	classes := ast.FindByType(parser.NodeClassDef)

	for _, class := range classes {
		initMethod := FindInitMethod(class)
		if initMethod == nil {
			continue
		}

		// Find all Call nodes that instantiate classes
		instantiations := d.findInstantiations(initMethod, class.Name)

		for _, inst := range instantiations {
			finding := domain.DIAntipatternFinding{
				Type:       domain.DIAntipatternConcreteDependency,
				Subtype:    string(domain.ConcreteDepInstantiation),
				Severity:   domain.DIAntipatternSeverityWarning,
				ClassName:  class.Name,
				MethodName: "__init__",
				Location: domain.SourceLocation{
					FilePath:  filePath,
					StartLine: inst.location.StartLine,
					EndLine:   inst.location.EndLine,
					StartCol:  inst.location.StartCol,
					EndCol:    inst.location.EndCol,
				},
				Description: fmt.Sprintf("Directly instantiates concrete class '%s' in constructor", inst.className),
				Suggestion:  "Inject the dependency as a parameter instead of creating it in the constructor",
				Details: map[string]interface{}{
					"instantiated_class": inst.className,
				},
			}
			findings = append(findings, finding)
		}
	}

	return findings
}

// instantiationInfo holds information about a class instantiation
type instantiationInfo struct {
	className string
	location  parser.Location
}

// findInstantiations finds all class instantiations in a function
func (d *ConcreteDependencyDetector) findInstantiations(funcNode *parser.Node, ownerClassName string) []instantiationInfo {
	var instantiations []instantiationInfo

	// Use WalkDeep to traverse including Value field
	funcNode.WalkDeep(func(node *parser.Node) bool {
		// Only check Assign nodes - the Call is in the Value field
		if node.Type == parser.NodeAssign {
			if node.Value != nil {
				if valueNode, ok := node.Value.(*parser.Node); ok {
					if valueNode.Type == parser.NodeCall {
						className := d.extractCalleeClassName(valueNode)
						if className != "" && d.isClassInstantiation(className, ownerClassName) {
							instantiations = append(instantiations, instantiationInfo{
								className: className,
								location:  valueNode.Location,
							})
						}
					}
				}
			}
		}
		return true
	})

	return instantiations
}

// extractCalleeClassName extracts the class name from a Call node
func (d *ConcreteDependencyDetector) extractCalleeClassName(callNode *parser.Node) string {
	// Check Value field first - this is where tree-sitter stores the callee
	if callNode.Value != nil {
		if valueNode, ok := callNode.Value.(*parser.Node); ok {
			if valueNode.Type == parser.NodeName {
				return valueNode.Name
			}
			if valueNode.Type == parser.NodeAttribute {
				return d.extractAttributeClassName(valueNode)
			}
		}
	}

	// Check direct Name child (fallback)
	for _, child := range callNode.Children {
		if child != nil && child.Type == parser.NodeName && child.Name != "" {
			return child.Name
		}
	}

	// Check Left field
	if callNode.Left != nil {
		if callNode.Left.Type == parser.NodeName {
			return callNode.Left.Name
		}
		// Handle module.Class() pattern
		if callNode.Left.Type == parser.NodeAttribute {
			return d.extractAttributeClassName(callNode.Left)
		}
	}

	return ""
}

// extractAttributeClassName extracts class name from an Attribute node
func (d *ConcreteDependencyDetector) extractAttributeClassName(attrNode *parser.Node) string {
	return ExtractAttributeName(attrNode)
}

// isClassInstantiation checks if a call is likely a class instantiation
func (d *ConcreteDependencyDetector) isClassInstantiation(name string, ownerClassName string) bool {
	// Skip if it's a call to self's own class
	if name == ownerClassName {
		return false
	}

	// Skip if it looks like a function call (lowercase first letter)
	if name != "" && name[0] >= 'a' && name[0] <= 'z' {
		return false
	}

	// Skip if it's a builtin type
	if d.isBuiltinType(name) {
		return false
	}

	// Skip if it looks abstract
	if !d.isConcreteType(name) {
		return false
	}

	return true
}

// isBuiltinType checks if a name is a built-in type
func (d *ConcreteDependencyDetector) isBuiltinType(name string) bool {
	builtins := map[string]bool{
		"bool": true, "int": true, "float": true, "complex": true,
		"str": true, "bytes": true, "bytearray": true,
		"list": true, "tuple": true, "range": true,
		"dict": true, "set": true, "frozenset": true,
		"object": true, "type": true, "super": true,
		"Exception": true, "BaseException": true, "ValueError": true,
		"TypeError": true, "KeyError": true, "IndexError": true,
		"AttributeError": true, "RuntimeError": true,
	}
	return builtins[name]
}

// extractTypeHintName extracts the type name from an argument's type annotation
func (d *ConcreteDependencyDetector) extractTypeHintName(arg *parser.Node) string {
	// Check children for type annotation (tree-sitter AST structure)
	for _, child := range arg.Children {
		if child == nil {
			continue
		}

		// Recursively look for type annotations
		if d.isTypeAnnotation(child) {
			typeName := d.extractTypeNameRecursive(child)
			if typeName != "" {
				return typeName
			}
		}
	}

	// Fall back to the raw type annotation text when the AST shape is unsupported.
	if arg.Value != nil {
		if typeStr, ok := arg.Value.(string); ok && typeStr != "" {
			return d.extractTypeNameFromString(typeStr)
		}
	}

	return ""
}

// isTypeAnnotation checks if a node represents a type annotation
func (d *ConcreteDependencyDetector) isTypeAnnotation(node *parser.Node) bool {
	return node.Type == parser.NodeName ||
		node.Type == parser.NodeBinOp ||
		node.Type == parser.NodeSubscript ||
		node.Type == parser.NodeAttribute ||
		node.Type == parser.NodeTypeNode ||
		node.Type == parser.NodeGenericType ||
		node.Type == parser.NodeTypeParameter
}

// extractTypeNameRecursive recursively extracts type name from type annotation nodes
func (d *ConcreteDependencyDetector) extractTypeNameRecursive(node *parser.Node) string {
	if node == nil {
		return ""
	}

	switch node.Type {
	case parser.NodeName:
		if d.isGenericTypeName(node.Name) || node.Name == "None" {
			return ""
		}
		return node.Name
	case parser.NodeAttribute:
		typeName := d.extractAttributeClassName(node)
		if d.isGenericTypeName(typeName) || typeName == "None" {
			return ""
		}
		return typeName
	case parser.NodeSubscript:
		return d.extractGenericInnerType(node)
	case parser.NodeBinOp:
		if node.Op == "|" {
			if left := d.extractTypeNameRecursive(node.Left); left != "" {
				return left
			}
			return d.extractTypeNameRecursive(node.Right)
		}
	case parser.NodeTypeNode, parser.NodeGenericType, parser.NodeTypeParameter:
		// Recursively check children
		for _, child := range node.Children {
			if child != nil && d.isTypeAnnotation(child) {
				typeName := d.extractTypeNameRecursive(child)
				if typeName != "" {
					return typeName
				}
			}
		}
	}

	return ""
}

// extractGenericInnerType extracts the inner type from a generic type
func (d *ConcreteDependencyDetector) extractGenericInnerType(subscriptNode *parser.Node) string {
	for _, child := range subscriptNode.Children {
		if child == nil {
			continue
		}
		if typeName := d.extractTypeNameRecursive(child); typeName != "" {
			return typeName
		}
	}

	if valueNode, ok := subscriptNode.Value.(*parser.Node); ok {
		return d.extractTypeNameRecursive(valueNode)
	}

	return ""
}

func (d *ConcreteDependencyDetector) extractTypeNameFromString(typeName string) string {
	typeName = strings.TrimSpace(typeName)
	if typeName == "" {
		return ""
	}

	if strings.Contains(typeName, "|") {
		parts := strings.Split(typeName, "|")
		for _, part := range parts {
			if extracted := d.extractTypeNameFromString(part); extracted != "" {
				return extracted
			}
		}
		return ""
	}

	if open := strings.Index(typeName, "["); open >= 0 && strings.HasSuffix(typeName, "]") {
		outer := strings.TrimSpace(typeName[:open])
		inner := strings.TrimSpace(typeName[open+1 : len(typeName)-1])
		if d.isGenericTypeName(d.lastTypeSegment(outer)) {
			for _, part := range strings.Split(inner, ",") {
				if extracted := d.extractTypeNameFromString(part); extracted != "" {
					return extracted
				}
			}
			return ""
		}
	}

	return d.lastTypeSegment(typeName)
}

func (d *ConcreteDependencyDetector) lastTypeSegment(typeName string) string {
	typeName = strings.TrimSpace(typeName)
	if typeName == "" {
		return ""
	}

	if idx := strings.LastIndex(typeName, "."); idx >= 0 {
		return strings.TrimSpace(typeName[idx+1:])
	}

	return typeName
}

func (d *ConcreteDependencyDetector) isGenericTypeName(typeName string) bool {
	switch typeName {
	case "Optional", "List", "Dict", "Set", "Tuple", "Union", "Callable", "Type", "Any", "Sequence", "Mapping", "Iterable":
		return true
	default:
		return false
	}
}

// isConcreteType checks if a type name is likely a concrete class
func (d *ConcreteDependencyDetector) isConcreteType(typeName string) bool {
	// Empty or builtin types are not concrete dependencies
	if typeName == "" || d.isBuiltinType(typeName) {
		return false
	}

	// Check for abstract naming conventions
	for _, prefix := range d.abstractPrefixes {
		if strings.HasPrefix(typeName, prefix) {
			return false
		}
	}

	for _, suffix := range d.abstractSuffixes {
		if strings.HasSuffix(typeName, suffix) {
			return false
		}
	}

	return !d.isGenericTypeName(typeName)
}
