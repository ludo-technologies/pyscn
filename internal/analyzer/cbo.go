package analyzer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
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
	TypeHintDependencies        int
	InstantiationDependencies   int
	AttributeAccessDependencies int
	ImportDependencies          int

	// Detailed dependency list
	DependentClasses []string

	// Risk assessment
	RiskLevel string // "low", "medium", "high"

	// Additional class metadata
	IsAbstract  bool
	BaseClasses []string
	Methods     []string
	Attributes  []string
}

// CBOOptions configures CBO analysis behavior
type CBOOptions struct {
	IncludeBuiltins   bool
	IncludeImports    bool
	PublicClassesOnly bool
	ExcludePatterns   []string
	LowThreshold      int // Default: 3 (industry standard)
	MediumThreshold   int // Default: 7 (industry standard)
}

// DefaultCBOOptions returns default CBO analysis options
// Threshold values are sourced from domain/defaults.go
func DefaultCBOOptions() *CBOOptions {
	return &CBOOptions{
		IncludeBuiltins:   false,
		IncludeImports:    true,
		PublicClassesOnly: false,
		ExcludePatterns:   []string{"test_*", "*_test", "__*__"},
		LowThreshold:      domain.DefaultCBOLowThreshold,    // Industry standard: CBO <= 3 is low risk
		MediumThreshold:   domain.DefaultCBOMediumThreshold, // Industry standard: 3 < CBO <= 7 is medium risk
	}
}

// CBOAnalyzer analyzes class coupling in Python code
type CBOAnalyzer struct {
	options          *CBOOptions
	builtinTypes     map[string]bool
	builtinFunctions map[string]bool
	standardLibs     map[string]bool
	importedNames    map[string]string         // alias -> module.name mapping
	regexCache       map[string]*regexp.Regexp // pattern -> compiled regex cache
}

// NewCBOAnalyzer creates a new CBO analyzer
func NewCBOAnalyzer(options *CBOOptions) *CBOAnalyzer {
	if options == nil {
		options = DefaultCBOOptions()
	}

	analyzer := &CBOAnalyzer{
		options:          options,
		builtinTypes:     make(map[string]bool),
		builtinFunctions: make(map[string]bool),
		standardLibs:     make(map[string]bool),
		importedNames:    make(map[string]string),
		regexCache:       make(map[string]*regexp.Regexp),
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
		ClassName:        classNode.Name,
		FilePath:         filePath,
		StartLine:        classNode.Location.StartLine,
		EndLine:          classNode.Location.EndLine,
		DependentClasses: []string{},
		BaseClasses:      []string{},
		Methods:          []string{},
		Attributes:       []string{},
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
				// Check if this is an imported dependency
				if a.isImportedDependency(baseName) {
					result.ImportDependencies++
				} else {
					result.InheritanceDependencies++
				}
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
			// Look for type annotation node - could be Name, Subscript, Attribute, etc.
			for _, child := range node.Children {
				if child != nil && a.isTypeAnnotation(child) {
					a.extractTypeAnnotationDependencies(child, dependencies, result)
				}
			}
		case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
			// Method with type annotations
			a.analyzeMethodTypeHints(node, dependencies, result)
		}
		return true
	})
}

// isTypeAnnotation checks if a node represents a type annotation
func (a *CBOAnalyzer) isTypeAnnotation(node *parser.Node) bool {
	return node.Type == parser.NodeName ||
		node.Type == parser.NodeSubscript ||
		node.Type == parser.NodeAttribute ||
		node.Type == parser.NodeTypeNode ||
		node.Type == parser.NodeGenericType ||
		node.Type == parser.NodeTypeParameter
}

// extractTypeAnnotationDependencies extracts class dependencies from type annotations
func (a *CBOAnalyzer) extractTypeAnnotationDependencies(node *parser.Node, dependencies map[string]bool, result *CBOResult) {
	switch node.Type {
	case parser.NodeName:
		// Simple type: User
		if node.Name != "" && a.shouldIncludeDependency(node.Name) {
			dependencies[node.Name] = true
			// Check if this is an imported dependency
			if a.isImportedDependency(node.Name) {
				result.ImportDependencies++
			} else {
				result.TypeHintDependencies++
			}
		}
	case parser.NodeSubscript:
		// Generic type: List[User], Dict[str, User]
		// For generics, we want to extract the type parameters, not the container
		if node.Right != nil {
			a.extractTypeAnnotationDependencies(node.Right, dependencies, result)
		} else if len(node.Children) > 1 {
			a.extractTypeAnnotationDependencies(node.Children[1], dependencies, result)
		}
	case parser.NodeAttribute:
		// Module.Type: typing.List, mymodule.MyClass
		typeName := a.extractClassName(node)
		if typeName != "" && a.shouldIncludeDependency(typeName) {
			dependencies[typeName] = true
			// Check if this is an imported dependency
			if a.isImportedDependency(typeName) {
				result.ImportDependencies++
			} else {
				result.TypeHintDependencies++
			}
		}
	case parser.NodeTypeNode:
		// Tree-sitter 'type' node - recurse into children
		for _, child := range node.Children {
			if child != nil && a.isTypeAnnotation(child) {
				a.extractTypeAnnotationDependencies(child, dependencies, result)
			}
		}
	case parser.NodeGenericType:
		// Tree-sitter generic_type node (e.g., List[User])
		// Look for type_parameter children to get the actual types we depend on
		for _, child := range node.Children {
			if child != nil && child.Type == parser.NodeTypeParameter {
				a.extractTypeAnnotationDependencies(child, dependencies, result)
			}
		}
	case parser.NodeTypeParameter:
		// Tree-sitter type_parameter node - recurse into children
		for _, child := range node.Children {
			if child != nil && a.isTypeAnnotation(child) {
				a.extractTypeAnnotationDependencies(child, dependencies, result)
			}
		}
	}
}

// analyzeMethodTypeHints analyzes type hints in method signatures
func (a *CBOAnalyzer) analyzeMethodTypeHints(methodNode *parser.Node, dependencies map[string]bool, result *CBOResult) {
	if methodNode.Name != "" {
		result.Methods = append(result.Methods, methodNode.Name)
	}

	// Analyze parameter types
	for _, arg := range methodNode.Args {
		if arg != nil && arg.Type == parser.NodeArg {
			// Look for type annotation in argument children
			for _, child := range arg.Children {
				if child != nil && a.isTypeAnnotation(child) {
					a.extractTypeAnnotationDependencies(child, dependencies, result)
				}
			}
		}
	}

	// Analyze return type annotation
	// Look for return type annotations in function children
	for _, child := range methodNode.Children {
		if child != nil && a.isTypeAnnotation(child) {
			a.extractTypeAnnotationDependencies(child, dependencies, result)
		}
	}
}

// analyzeInstantiationAndAccess analyzes object instantiation and attribute access
func (a *CBOAnalyzer) analyzeInstantiationAndAccess(classNode *parser.Node, dependencies map[string]bool, result *CBOResult, allClasses map[string]*parser.Node) {
	a.walkNode(classNode, func(node *parser.Node) bool {
		switch node.Type {
		case parser.NodeAssign:
			// Assignment that might contain class instantiation: self.logger = Logger()
			// Use structural AST analysis instead of string parsing
			if node.Value != nil {
				if valueNode, ok := node.Value.(*parser.Node); ok {
					if valueNode.Type == parser.NodeCall {
						className := a.extractClassNameFromCallNode(valueNode)
						if className != "" && a.shouldIncludeDependency(className) {
							// Check if this is an imported dependency
							if a.isImportedDependency(className) {
								dependencies[className] = true
								result.ImportDependencies++
							} else if _, isClass := allClasses[className]; isClass {
								// Known local class (instantiation)
								dependencies[className] = true
								result.InstantiationDependencies++
							} else if a.options.IncludeBuiltins && a.builtinTypes[className] {
								// Builtin type (only if explicitly enabled)
								dependencies[className] = true
								result.InstantiationDependencies++
							}
							// Note: function calls are NOT added to dependencies
						}
					}
				}
			}
		case parser.NodeCall:
			// Function/class call - could be instantiation
			// Use structural AST analysis instead of string parsing
			className := a.extractClassNameFromCallNode(node)
			if className != "" && a.shouldIncludeDependency(className) {
				// Check if this is an imported dependency
				if a.isImportedDependency(className) {
					dependencies[className] = true
					result.ImportDependencies++
				} else if _, isClass := allClasses[className]; isClass {
					// Known local class (instantiation)
					dependencies[className] = true
					result.InstantiationDependencies++
				} else if a.options.IncludeBuiltins && a.builtinTypes[className] {
					// Builtin type (only if explicitly enabled)
					dependencies[className] = true
					result.InstantiationDependencies++
				}
				// Note: function calls are NOT added to dependencies
			}
		case parser.NodeAttribute:
			// Attribute access: obj.method() or obj.attr
			if objNode := node.Left; objNode != nil {
				objType := a.inferObjectType(objNode)
				if objType != "" && a.shouldIncludeDependency(objType) {
					dependencies[objType] = true
					// Check if this is an imported dependency
					if a.isImportedDependency(objType) {
						result.ImportDependencies++
					} else {
						result.AttributeAccessDependencies++
					}
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
	case parser.NodeSubscript:
		// Handle generic types like List[User], Dict[str, User]
		// For subscripts, the type parameter is typically in Right field or Children
		if node.Right != nil {
			return a.extractClassName(node.Right)
		}
		// Fallback to checking children for the subscript content
		if len(node.Children) > 1 {
			return a.extractClassName(node.Children[1])
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

	// Skip built-in types if not included, but always skip built-in functions
	if a.isBuiltinFunction(className) {
		return false // Always exclude built-in functions regardless of IncludeBuiltins
	}
	if !a.options.IncludeBuiltins && a.builtinTypes[className] {
		return false
	}

	// Skip standard library if not included
	if !a.options.IncludeImports {
		// Check both direct match and root module for qualified names like json.JSONDecoder
		if a.standardLibs[className] {
			return false
		}
		// Check root module for fully qualified names
		if strings.Contains(className, ".") {
			rootModule := strings.SplitN(className, ".", 2)[0]
			if a.standardLibs[rootModule] {
				return false
			}
		}
	}

	return true
}

// isBuiltinFunction checks if a name is a built-in function
func (a *CBOAnalyzer) isBuiltinFunction(name string) bool {
	return a.builtinFunctions[name]
}

// extractClassNameFromCallNode extracts class name from Call node using structural AST analysis
func (a *CBOAnalyzer) extractClassNameFromCallNode(callNode *parser.Node) string {
	if callNode == nil || callNode.Type != parser.NodeCall {
		return ""
	}

	// Method 1: Check direct Name nodes in immediate children (most common case)
	// This handles simple calls like Logger(), MyClass()
	for _, child := range callNode.Children {
		if child != nil && child.Type == parser.NodeName && child.Name != "" {
			return child.Name
		}
	}

	// Method 2: Check Left field for function being called
	if callNode.Left != nil {
		switch callNode.Left.Type {
		case parser.NodeName:
			// Simple function/class call: Logger()
			return callNode.Left.Name
		case parser.NodeAttribute:
			// Attribute access: module.Class()
			return a.extractClassNameFromAttribute(callNode.Left)
		}
	}

	// Method 3: Check Value field if it's a Node with Name type
	if callNode.Value != nil {
		if valueNode, ok := callNode.Value.(*parser.Node); ok {
			if valueNode.Type == parser.NodeName && valueNode.Name != "" {
				return valueNode.Name
			}
		}
	}

	return ""
}

// extractClassNameFromAttribute extracts class name from Attribute node using direct AST access
func (a *CBOAnalyzer) extractClassNameFromAttribute(attrNode *parser.Node) string {
	if attrNode == nil || attrNode.Type != parser.NodeAttribute {
		return ""
	}

	// For module.Class pattern, we want the rightmost name (Class)
	// but for full qualification, we might want module.Class

	// Get the rightmost part (the actual class name)
	if attrNode.Right != nil && attrNode.Right.Type == parser.NodeName {
		rightName := attrNode.Right.Name

		// Get the left part (module name) if needed
		if attrNode.Left != nil && attrNode.Left.Type == parser.NodeName {
			leftName := attrNode.Left.Name
			// Return full qualification for better accuracy
			return leftName + "." + rightName
		}

		// Return just the class name if no module prefix
		return rightName
	}

	return ""
}

// isImportedDependency checks if a dependency comes from imports
func (a *CBOAnalyzer) isImportedDependency(className string) bool {
	// Check if the class name or its parts exist in importedNames
	if _, exists := a.importedNames[className]; exists {
		return true
	}

	// Check for qualified names like module.Class
	if strings.Contains(className, ".") {
		parts := strings.SplitN(className, ".", 2)
		if len(parts) == 2 {
			moduleName := parts[0]
			// Check if the module part is imported
			if _, exists := a.importedNames[moduleName]; exists {
				return true
			}
		}
	}

	return false
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

	for _, child := range node.Args {
		a.walkNode(child, visitor)
	}

	// Also traverse Value field if it contains a Node
	if node.Value != nil {
		if valueNode, ok := node.Value.(*parser.Node); ok {
			a.walkNode(valueNode, visitor)
		}
	}
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
		// Check cache first
		regex, exists := a.regexCache[pattern]
		if !exists {
			// Convert glob pattern to regex pattern and cache it
			regexPattern := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), "\\*", ".*") + "$"
			var err error
			regex, err = regexp.Compile(regexPattern)
			if err != nil {
				// Fallback to exact match if regex compilation fails
				return str == pattern
			}
			a.regexCache[pattern] = regex
		}
		return regex.MatchString(str)
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

// initializeBuiltinTypes initializes the built-in types and functions set
func (a *CBOAnalyzer) initializeBuiltinTypes() {
	// Built-in types (can be dependencies when IncludeBuiltins=true)
	builtinTypes := []string{
		"bool", "int", "float", "complex", "str", "bytes", "bytearray",
		"list", "tuple", "range", "dict", "set", "frozenset",
		"object", "type", "super", "property", "classmethod", "staticmethod",
		"Exception", "BaseException", "ValueError", "TypeError", "KeyError",
		"IndexError", "AttributeError", "NameError", "RuntimeError",
		"memoryview", "slice",
	}

	// Built-in functions (never counted as dependencies)
	builtinFunctions := []string{
		"print", "len", "max", "min", "sum", "abs", "round", "pow",
		"sorted", "reversed", "enumerate", "zip", "map", "filter",
		"any", "all", "iter", "next", "chr", "ord", "bin", "hex", "oct",
		"hash", "id", "repr", "ascii", "format", "callable", "hasattr",
		"getattr", "setattr", "delattr", "dir", "vars", "locals", "globals",
		"isinstance", "issubclass", "open", "input", "eval", "exec", "compile",
	}

	for _, builtin := range builtinTypes {
		a.builtinTypes[builtin] = true
	}

	for _, builtin := range builtinFunctions {
		a.builtinFunctions[builtin] = true
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
