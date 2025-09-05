package analyzer

import (
	"fmt"
	"strings"

	"github.com/pyqol/pyqol/internal/parser"
)

// CBOResult holds CBO (Coupling Between Objects) metrics for a class
type CBOResult struct {
	// Core CBO metric
	CouplingCount int

	// Class information
	ClassName string
	FilePath  string
	StartLine int
	EndLine   int

	// Dependency breakdown
	InheritanceDependencies     int
	TypeHintDependencies       int
	InstantiationDependencies  int
	AttributeAccessDependencies int
	ImportDependencies         int

	// Detailed dependency list
	DependentClasses []string

	// Risk assessment
	RiskLevel string // "low", "medium", "high"

	// Additional class metadata
	IsAbstract   bool
	BaseClasses  []string
	Methods      []string
	Attributes   []string
}

// CBOOptions configures CBO analysis behavior
type CBOOptions struct {
	IncludeBuiltins     bool
	IncludeImports      bool
	IncludeThirdParty   bool
	PublicClassesOnly   bool
	ExcludePatterns     []string
	LowThreshold        int // Default: 5
	MediumThreshold     int // Default: 10
}

// DefaultCBOOptions returns default CBO analysis options
func DefaultCBOOptions() *CBOOptions {
	return &CBOOptions{
		IncludeBuiltins:   false,
		IncludeImports:    true,
		IncludeThirdParty: false,
		PublicClassesOnly: false,
		ExcludePatterns:   []string{"test_*", "*_test", "__*__"},
		LowThreshold:      5,
		MediumThreshold:   10,
	}
}

// CBOAnalyzer analyzes class coupling in Python code
type CBOAnalyzer struct {
	options       *CBOOptions
	builtinTypes  map[string]bool
	standardLibs  map[string]bool
	importedNames map[string]string // alias -> module.name mapping
}

// NewCBOAnalyzer creates a new CBO analyzer
func NewCBOAnalyzer(options *CBOOptions) *CBOAnalyzer {
	if options == nil {
		options = DefaultCBOOptions()
	}

	analyzer := &CBOAnalyzer{
		options:       options,
		builtinTypes:  make(map[string]bool),
		standardLibs:  make(map[string]bool),
		importedNames: make(map[string]string),
	}

	analyzer.initializeBuiltinTypes()
	analyzer.initializeStandardLibs()

	return analyzer
}

// AnalyzeClasses analyzes CBO for all classes in the given AST
func (a *CBOAnalyzer) AnalyzeClasses(ast *parser.Node, filePath string) ([]*CBOResult, error) {
	if ast == nil {
		return nil, fmt.Errorf("AST is nil")
	}

	// First pass: collect all imports and class definitions
	classes := a.collectClasses(ast)
	imports := a.collectImports(ast)

	// Update imported names mapping
	for alias, module := range imports {
		a.importedNames[alias] = module
	}

	var results []*CBOResult

	// Second pass: analyze coupling for each class
	for _, classNode := range classes {
		result, err := a.analyzeClass(classNode, filePath, classes)
		if err != nil {
			// Log warning but continue with other classes
			continue
		}

		// Apply filtering
		if a.shouldIncludeClass(result) {
			results = append(results, result)
		}
	}

	return results, nil
}

// analyzeClass analyzes CBO for a single class
func (a *CBOAnalyzer) analyzeClass(classNode *parser.Node, filePath string, allClasses map[string]*parser.Node) (*CBOResult, error) {
	if classNode.Type != parser.NodeClassDef {
		return nil, fmt.Errorf("node is not a class definition")
	}

	result := &CBOResult{
		ClassName:       classNode.Name,
		FilePath:        filePath,
		StartLine:       classNode.Location.StartLine,
		EndLine:         classNode.Location.EndLine,
		DependentClasses: []string{},
		BaseClasses:     []string{},
		Methods:         []string{},
		Attributes:      []string{},
	}

	// Track unique dependencies
	dependencies := make(map[string]bool)

	// 1. Analyze inheritance dependencies
	a.analyzeInheritance(classNode, dependencies, result)

	// 2. Analyze type hints in class body
	a.analyzeTypeHints(classNode, dependencies, result)

	// 3. Analyze instantiation and attribute access
	a.analyzeInstantiationAndAccess(classNode, dependencies, result, allClasses)

	// 4. Calculate final metrics
	result.CouplingCount = len(dependencies)
	result.DependentClasses = a.mapToSlice(dependencies)
	result.RiskLevel = a.assessRiskLevel(result.CouplingCount)

	// 5. Check if class is abstract
	result.IsAbstract = a.isAbstractClass(classNode)

	return result, nil
}

// analyzeInheritance analyzes inheritance-based coupling
func (a *CBOAnalyzer) analyzeInheritance(classNode *parser.Node, dependencies map[string]bool, result *CBOResult) {
	// Analyze base classes from Bases field
	for _, baseNode := range classNode.Bases {
		if baseNode != nil {
			baseName := a.extractClassName(baseNode)
			if baseName != "" && a.shouldIncludeDependency(baseName) {
				dependencies[baseName] = true
				result.BaseClasses = append(result.BaseClasses, baseName)
				result.InheritanceDependencies++
			}
		}
	}
}

// analyzeTypeHints analyzes type annotation dependencies
func (a *CBOAnalyzer) analyzeTypeHints(classNode *parser.Node, dependencies map[string]bool, result *CBOResult) {
	// Walk through class body looking for type annotations
	a.walkNode(classNode, func(node *parser.Node) bool {
		switch node.Type {
		case parser.NodeAnnAssign:
			// Variable with type annotation: x: SomeType = value
			if annotation := a.findChildByType(node, parser.NodeName); annotation != nil {
				typeName := a.extractClassName(annotation)
				if typeName != "" && a.shouldIncludeDependency(typeName) {
					dependencies[typeName] = true
					result.TypeHintDependencies++
				}
			}
		case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
			// Method with type annotations
			a.analyzeMethodTypeHints(node, dependencies, result)
		}
		return true
	})
}

// analyzeMethodTypeHints analyzes type hints in method signatures
func (a *CBOAnalyzer) analyzeMethodTypeHints(methodNode *parser.Node, dependencies map[string]bool, result *CBOResult) {
	if methodNode.Name != "" {
		result.Methods = append(result.Methods, methodNode.Name)
	}

	// Analyze parameter types
	for _, arg := range methodNode.Args {
		if arg != nil && arg.Type == parser.NodeArg {
			// Look for type annotation in argument
			if typeNode := a.findChildByType(arg, parser.NodeName); typeNode != nil {
				typeName := a.extractClassName(typeNode)
				if typeName != "" && a.shouldIncludeDependency(typeName) {
					dependencies[typeName] = true
					result.TypeHintDependencies++
				}
			}
		}
	}

	// Analyze return type annotation
	// This would need tree-sitter specific parsing for return type annotations
}

// analyzeInstantiationAndAccess analyzes object instantiation and attribute access
func (a *CBOAnalyzer) analyzeInstantiationAndAccess(classNode *parser.Node, dependencies map[string]bool, result *CBOResult, allClasses map[string]*parser.Node) {
	a.walkNode(classNode, func(node *parser.Node) bool {
		switch node.Type {
		case parser.NodeCall:
			// Function/class call - could be instantiation
			if funcNode := node.Left; funcNode != nil {
				className := a.extractClassName(funcNode)
				if className != "" && a.shouldIncludeDependency(className) {
					// Check if it's a known class (instantiation)
					if _, isClass := allClasses[className]; isClass {
						dependencies[className] = true
						result.InstantiationDependencies++
					}
				}
			}
		case parser.NodeAttribute:
			// Attribute access: obj.method() or obj.attr
			if objNode := node.Left; objNode != nil {
				objType := a.inferObjectType(objNode)
				if objType != "" && a.shouldIncludeDependency(objType) {
					dependencies[objType] = true
					result.AttributeAccessDependencies++
				}
			}
		}
		return true
	})
}

// Helper methods

// collectClasses collects all class definitions from AST
func (a *CBOAnalyzer) collectClasses(ast *parser.Node) map[string]*parser.Node {
	classes := make(map[string]*parser.Node)
	
	a.walkNode(ast, func(node *parser.Node) bool {
		if node.Type == parser.NodeClassDef && node.Name != "" {
			classes[node.Name] = node
		}
		return true
	})
	
	return classes
}

// collectImports collects import statements and their aliases
func (a *CBOAnalyzer) collectImports(ast *parser.Node) map[string]string {
	imports := make(map[string]string)
	
	a.walkNode(ast, func(node *parser.Node) bool {
		switch node.Type {
		case parser.NodeImport:
			// import module as alias
			for _, child := range node.Children {
				if child.Type == parser.NodeAlias {
					module := child.Name
					alias := child.Name // Default to module name
					if child.Value != nil {
						if aliasStr, ok := child.Value.(string); ok {
							alias = aliasStr
						}
					}
					imports[alias] = module
				}
			}
		case parser.NodeImportFrom:
			// from module import name as alias
			module := node.Module
			for _, child := range node.Children {
				if child.Type == parser.NodeAlias {
					name := child.Name
					alias := name // Default to imported name
					if child.Value != nil {
						if aliasStr, ok := child.Value.(string); ok {
							alias = aliasStr
						}
					}
					imports[alias] = module + "." + name
				}
			}
		}
		return true
	})
	
	return imports
}

// extractClassName extracts class name from a node
func (a *CBOAnalyzer) extractClassName(node *parser.Node) string {
	if node == nil {
		return ""
	}
	
	switch node.Type {
	case parser.NodeName:
		return node.Name
	case parser.NodeAttribute:
		// Handle module.ClassName
		if node.Left != nil && node.Right != nil {
			left := a.extractClassName(node.Left)
			right := a.extractClassName(node.Right)
			if left != "" && right != "" {
				return left + "." + right
			}
		}
	}
	
	return ""
}

// shouldIncludeDependency checks if a dependency should be included
func (a *CBOAnalyzer) shouldIncludeDependency(className string) bool {
	// Skip self-references
	if className == "" {
		return false
	}
	
	// Check exclude patterns
	for _, pattern := range a.options.ExcludePatterns {
		if a.matchesPattern(className, pattern) {
			return false
		}
	}
	
	// Skip built-in types if not included
	if !a.options.IncludeBuiltins && a.builtinTypes[className] {
		return false
	}
	
	// Skip standard library if not included
	if !a.options.IncludeImports && a.standardLibs[className] {
		return false
	}
	
	return true
}

// shouldIncludeClass checks if a class should be included in results
func (a *CBOAnalyzer) shouldIncludeClass(result *CBOResult) bool {
	if a.options.PublicClassesOnly && strings.HasPrefix(result.ClassName, "_") {
		return false
	}
	
	for _, pattern := range a.options.ExcludePatterns {
		if a.matchesPattern(result.ClassName, pattern) {
			return false
		}
	}
	
	return true
}

// assessRiskLevel determines risk level based on CBO count
func (a *CBOAnalyzer) assessRiskLevel(cbo int) string {
	if cbo <= a.options.LowThreshold {
		return "low"
	} else if cbo <= a.options.MediumThreshold {
		return "medium"
	}
	return "high"
}

// walkNode recursively walks AST nodes
func (a *CBOAnalyzer) walkNode(node *parser.Node, visitor func(*parser.Node) bool) {
	if node == nil || !visitor(node) {
		return
	}
	
	for _, child := range node.Children {
		a.walkNode(child, visitor)
	}
	
	for _, child := range node.Body {
		a.walkNode(child, visitor)
	}
}

// findChildByType finds first child node of specified type
func (a *CBOAnalyzer) findChildByType(node *parser.Node, nodeType parser.NodeType) *parser.Node {
	if node == nil {
		return nil
	}
	
	for _, child := range node.Children {
		if child.Type == nodeType {
			return child
		}
		if found := a.findChildByType(child, nodeType); found != nil {
			return found
		}
	}
	
	return nil
}

// inferObjectType tries to infer the type of an object from context
func (a *CBOAnalyzer) inferObjectType(node *parser.Node) string {
	// Simple heuristic - try to extract type from variable name or method call
	if node.Type == parser.NodeName {
		// Look up in imported names
		if fullName, exists := a.importedNames[node.Name]; exists {
			return fullName
		}
		return node.Name
	}
	
	return ""
}

// isAbstractClass checks if a class is abstract (has @abstractmethod decorators)
func (a *CBOAnalyzer) isAbstractClass(classNode *parser.Node) bool {
	hasAbstractMethod := false
	
	a.walkNode(classNode, func(node *parser.Node) bool {
		if node.Type == parser.NodeFunctionDef || node.Type == parser.NodeAsyncFunctionDef {
			for _, decorator := range node.Decorator {
				if decorator != nil && strings.Contains(decorator.Name, "abstractmethod") {
					hasAbstractMethod = true
					return false
				}
			}
		}
		return true
	})
	
	return hasAbstractMethod
}

// matchesPattern checks if a string matches a simple pattern (with * wildcards)
func (a *CBOAnalyzer) matchesPattern(str, pattern string) bool {
	if pattern == "*" {
		return true
	}
	
	if strings.Contains(pattern, "*") {
		// Simple wildcard matching
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			prefix, suffix := parts[0], parts[1]
			return strings.HasPrefix(str, prefix) && strings.HasSuffix(str, suffix)
		}
	}
	
	return str == pattern
}

// mapToSlice converts map keys to slice
func (a *CBOAnalyzer) mapToSlice(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for key := range m {
		result = append(result, key)
	}
	return result
}

// initializeBuiltinTypes initializes the built-in types set
func (a *CBOAnalyzer) initializeBuiltinTypes() {
	builtins := []string{
		"bool", "int", "float", "complex", "str", "bytes", "bytearray",
		"list", "tuple", "range", "dict", "set", "frozenset",
		"object", "type", "super", "property", "classmethod", "staticmethod",
		"Exception", "BaseException", "ValueError", "TypeError", "KeyError",
		"IndexError", "AttributeError", "NameError", "RuntimeError",
	}
	
	for _, builtin := range builtins {
		a.builtinTypes[builtin] = true
	}
}

// initializeStandardLibs initializes common standard library modules
func (a *CBOAnalyzer) initializeStandardLibs() {
	stdlibs := []string{
		"os", "sys", "re", "json", "datetime", "collections", "itertools",
		"functools", "operator", "math", "random", "string", "io", "pathlib",
		"unittest", "logging", "argparse", "configparser", "urllib", "http",
	}
	
	for _, stdlib := range stdlibs {
		a.standardLibs[stdlib] = true
	}
}

// CalculateCBO is a convenience function for calculating CBO with default config
func CalculateCBO(ast *parser.Node, filePath string) ([]*CBOResult, error) {
	analyzer := NewCBOAnalyzer(DefaultCBOOptions())
	return analyzer.AnalyzeClasses(ast, filePath)
}

// CalculateCBOWithConfig calculates CBO with custom configuration
func CalculateCBOWithConfig(ast *parser.Node, filePath string, options *CBOOptions) ([]*CBOResult, error) {
	analyzer := NewCBOAnalyzer(options)
	return analyzer.AnalyzeClasses(ast, filePath)
}

// CalculateFilesCBO calculates CBO for multiple files
func CalculateFilesCBO(asts map[string]*parser.Node, options *CBOOptions) (map[string][]*CBOResult, error) {
	results := make(map[string][]*CBOResult)
	analyzer := NewCBOAnalyzer(options)
	
	for filePath, ast := range asts {
		fileResults, err := analyzer.AnalyzeClasses(ast, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze file %s: %w", filePath, err)
		}
		results[filePath] = fileResults
	}
	
	return results, nil
}