package analyzer

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

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
	IncludeBuiltins       bool
	IncludeImports        bool
	PublicClassesOnly     bool
	GroupNamespaceImports bool
	ExcludePatterns       []string
	LowThreshold          int // Default: 3 (industry standard)
	MediumThreshold       int // Default: 7 (industry standard)
}

// DefaultCBOOptions returns default CBO analysis options
// Threshold values are sourced from domain/defaults.go
func DefaultCBOOptions() *CBOOptions {
	return &CBOOptions{
		IncludeBuiltins:       false,
		IncludeImports:        true,
		PublicClassesOnly:     false,
		GroupNamespaceImports: true,
		ExcludePatterns:       []string{"test_*", "*_test", "__*__"},
		LowThreshold:          domain.DefaultCBOLowThreshold,    // Industry standard: CBO <= 3 is low risk
		MediumThreshold:       domain.DefaultCBOMediumThreshold, // Industry standard: 3 < CBO <= 7 is medium risk
	}
}

// CBOAnalyzer analyzes class coupling in Python code
type CBOAnalyzer struct {
	options          *CBOOptions
	builtinTypes     map[string]bool
	builtinFunctions map[string]bool
	standardLibs     map[string]bool
	importedNames    map[string]string         // alias -> module.name mapping
	namespaceAliases map[string]string         // module.name -> alias mapping (only for alias imports)
	regexCache       map[string]*regexp.Regexp // pattern -> compiled regex cache

	// futureAnnotations is true when the file being analyzed has
	// `from __future__ import annotations` (PEP 563) at module scope. Under
	// PEP 563 all annotations are evaluated as strings at runtime, so a class
	// referenced only inside a type annotation is never actually imported or
	// resolved and imposes zero coupling cost. File-local, reset per AST. See #628.
	futureAnnotations bool
}

type cboDependencyKind int

const (
	dependencyKindInheritance cboDependencyKind = iota
	dependencyKindTypeHint
	dependencyKindInstantiation
	dependencyKindAttributeAccess
	dependencyKindImport
)

type cboImportMaps struct {
	importedNames    map[string]string
	namespaceAliases map[string]string
}

type cboDependencies struct {
	all               map[string]bool
	inheritance       map[string]bool
	typeHints         map[string]bool
	instantiations    map[string]bool
	attributeAccesses map[string]bool
	imports           map[string]bool
}

func newCBODependencies() *cboDependencies {
	return &cboDependencies{
		all:               make(map[string]bool),
		inheritance:       make(map[string]bool),
		typeHints:         make(map[string]bool),
		instantiations:    make(map[string]bool),
		attributeAccesses: make(map[string]bool),
		imports:           make(map[string]bool),
	}
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
		namespaceAliases: make(map[string]string),
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

	// Import aliases are file-local. Reset the maps on each AST so one analyzed file
	// cannot change dependency identity in the next file.
	a.importedNames = imports.importedNames
	a.namespaceAliases = imports.namespaceAliases
	a.futureAnnotations = a.hasFutureAnnotations(ast)

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

	dependencies := newCBODependencies()

	// References that resolve, per Python scoping, to a class defined within
	// this class are internal helpers, not coupling. See #547.
	resolver := a.newNestedClassResolver(classNode)

	// 1. Analyze inheritance dependencies (a base never resolves to a class
	// nested inside its own body, so no exclusion is needed here).
	a.analyzeInheritance(classNode, dependencies, result)

	// 2. Analyze type hints in class body
	a.analyzeTypeHints(classNode, dependencies, result, resolver)

	// 3. Analyze instantiation and attribute access
	a.analyzeInstantiationAndAccess(classNode, dependencies, result, allClasses, resolver)

	// 4. Calculate final metrics
	result.CouplingCount = len(dependencies.all)
	result.DependentClasses = a.mapToSlice(dependencies.all)
	result.InheritanceDependencies = len(dependencies.inheritance)
	result.TypeHintDependencies = len(dependencies.typeHints)
	result.InstantiationDependencies = len(dependencies.instantiations)
	result.AttributeAccessDependencies = len(dependencies.attributeAccesses)
	result.ImportDependencies = len(dependencies.imports)
	result.RiskLevel = a.assessRiskLevel(result.CouplingCount)

	// 5. Check if class is abstract
	result.IsAbstract = a.isAbstractClass(classNode)

	return result, nil
}

// analyzeInheritance analyzes inheritance-based coupling
func (a *CBOAnalyzer) analyzeInheritance(classNode *parser.Node, dependencies *cboDependencies, result *CBOResult) {
	// Analyze base classes from Bases field
	for _, baseNode := range classNode.Bases {
		if baseNode != nil {
			baseName := a.extractClassName(baseNode)
			if baseName != "" {
				result.BaseClasses = append(result.BaseClasses, baseName)
			}
			if baseName != "" && a.shouldIncludeDependencyForClass(baseName, result.ClassName) {
				a.addDependency(dependencies, baseName, dependencyKindInheritance)
			}
		}
	}
}

// analyzeTypeHints analyzes type annotation dependencies
func (a *CBOAnalyzer) analyzeTypeHints(classNode *parser.Node, dependencies *cboDependencies, result *CBOResult, resolver *nestedClassResolver) {
	// Walk through class body looking for type annotations
	a.walkNode(classNode, func(node *parser.Node) bool {
		switch node.Type {
		case parser.NodeAnnAssign:
			// Variable with type annotation: x: SomeType = value. A class-body
			// or method-body annotation resolves at its own lexical scope, so
			// no signature scope is forced here.
			for _, child := range node.Children {
				if child != nil && a.isTypeAnnotation(child) {
					a.extractTypeAnnotationDependencies(child, dependencies, result, resolver, nil)
				}
			}
		case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
			// Method with type annotations
			a.analyzeMethodTypeHints(node, dependencies, result, resolver)
		}
		return true
	})
}

// isTypeAnnotation checks if a node represents a type annotation
func (a *CBOAnalyzer) isTypeAnnotation(node *parser.Node) bool {
	return node.Type == parser.NodeName ||
		node.Type == parser.NodeSubscript ||
		node.Type == parser.NodeAttribute ||
		node.Type == parser.NodeTuple ||
		node.Type == parser.NodeList ||
		node.Type == parser.NodeTypeNode ||
		node.Type == parser.NodeGenericType ||
		node.Type == parser.NodeTypeParameter ||
		node.Type == parser.NodeBinOp // Union type: X | Y (Python 3.10+)
}

// extractTypeAnnotationDependencies extracts class dependencies from type
// annotations. sigScope is the Python evaluation scope for the annotation: nil
// for a class-body or method-body annotation (resolved at the reference's own
// lexical scope), or the method's defining scope for a method *signature*
// annotation, which Python evaluates while the class body executes (see #547).
func (a *CBOAnalyzer) extractTypeAnnotationDependencies(node *parser.Node, dependencies *cboDependencies, result *CBOResult, resolver *nestedClassResolver, sigScope *parser.Node) {
	switch node.Type {
	case parser.NodeName:
		// Simple type: User
		if node.Name != "" && a.shouldIncludeDependencyForClass(node.Name, result.ClassName) && !resolver.annotationExcluded(node, node.Name, sigScope) {
			a.addDependency(dependencies, node.Name, dependencyKindTypeHint)
		}
	case parser.NodeSubscript:
		// Generic type: List[User], Dict[str, User]
		// For generics, we want to extract the type parameters, not the container
		for _, typeArg := range a.typeArgumentNodes(node) {
			a.extractTypeAnnotationDependencies(typeArg, dependencies, result, resolver, sigScope)
		}
	case parser.NodeAttribute:
		// Module.Type: typing.List, mymodule.MyClass
		typeName := a.extractClassName(node)
		if typeName != "" && a.shouldIncludeDependencyForClass(typeName, result.ClassName) && !resolver.annotationExcluded(node, typeName, sigScope) {
			a.addDependency(dependencies, typeName, dependencyKindTypeHint)
		}
	case parser.NodeTuple, parser.NodeList:
		// Multi-argument generic parameters: dict[str, User], tuple[A, B].
		for _, child := range node.Children {
			if child != nil && a.isTypeAnnotation(child) {
				a.extractTypeAnnotationDependencies(child, dependencies, result, resolver, sigScope)
			}
		}
	case parser.NodeTypeNode:
		// Tree-sitter 'type' node - recurse into children
		for _, child := range node.Children {
			if child != nil && a.isTypeAnnotation(child) {
				a.extractTypeAnnotationDependencies(child, dependencies, result, resolver, sigScope)
			}
		}
	case parser.NodeGenericType:
		// Tree-sitter generic_type node (e.g., List[User])
		// Look for type_parameter children to get the actual types we depend on
		for _, child := range node.Children {
			if child != nil && child.Type == parser.NodeTypeParameter {
				a.extractTypeAnnotationDependencies(child, dependencies, result, resolver, sigScope)
			}
		}
	case parser.NodeTypeParameter:
		// Tree-sitter type_parameter node - recurse into children
		for _, child := range node.Children {
			if child != nil && a.isTypeAnnotation(child) {
				a.extractTypeAnnotationDependencies(child, dependencies, result, resolver, sigScope)
			}
		}
	case parser.NodeBinOp:
		// Union type using | operator (Python 3.10+): Context | None, str | int
		if node.Op == "|" {
			// Extract dependencies from both sides of the union
			if node.Left != nil {
				a.extractTypeAnnotationDependencies(node.Left, dependencies, result, resolver, sigScope)
			}
			if node.Right != nil {
				a.extractTypeAnnotationDependencies(node.Right, dependencies, result, resolver, sigScope)
			}
		}
	}
}

// annotationExcluded reports whether a type-annotation reference resolves to a
// nested class and so must not be counted as coupling. When sigScope is set
// (a method signature annotation) the name is resolved in that fixed scope;
// otherwise it is resolved at the reference's own lexical scope.
func (r *nestedClassResolver) annotationExcluded(ref *parser.Node, name string, sigScope *parser.Node) bool {
	if sigScope != nil {
		return r.excludesInScope(name, sigScope)
	}
	return r.excludes(ref, name)
}

// analyzeMethodTypeHints analyzes type hints in method signatures
func (a *CBOAnalyzer) analyzeMethodTypeHints(methodNode *parser.Node, dependencies *cboDependencies, result *CBOResult, resolver *nestedClassResolver) {
	if methodNode.Name != "" {
		result.Methods = append(result.Methods, methodNode.Name)
	}

	// Parameter and return annotations are evaluated in the scope that *defines*
	// the method (the class body for a normal method), not in the method body,
	// so resolve their names there rather than at the annotation's lexical
	// position under the FunctionDef (see #547).
	sigScope := resolver.signatureScope(methodNode)

	// Analyze parameter types
	for _, arg := range methodNode.Args {
		if arg != nil && arg.Type == parser.NodeArg {
			if arg.Right != nil && a.isTypeAnnotation(arg.Right) {
				a.extractTypeAnnotationDependencies(arg.Right, dependencies, result, resolver, sigScope)
			}
		}
	}

	// Analyze return type annotation
	// Return type is stored in Right field (not Children) to avoid DFA interference
	if methodNode.Right != nil && a.isTypeAnnotation(methodNode.Right) {
		a.extractTypeAnnotationDependencies(methodNode.Right, dependencies, result, resolver, sigScope)
	}
}

// analyzeInstantiationAndAccess analyzes object instantiation and attribute access
func (a *CBOAnalyzer) analyzeInstantiationAndAccess(classNode *parser.Node, dependencies *cboDependencies, result *CBOResult, allClasses map[string]*parser.Node, resolver *nestedClassResolver) {
	// A reference that resolves to a class defined within this class (a local
	// helper) is internal implementation, not coupling. Resolution is
	// scope-aware, so a same-named class reached from a different scope still
	// counts (see #547).
	include := func(ref *parser.Node, dep string) bool {
		return a.shouldIncludeDependencyForClass(dep, result.ClassName) && !resolver.excludes(ref, dep)
	}

	a.walkNode(classNode, func(node *parser.Node) bool {
		switch node.Type {
		case parser.NodeAssign:
			// Assignment that might contain class instantiation: self.logger = Logger()
			// Use structural AST analysis instead of string parsing
			if node.Value != nil {
				if valueNode, ok := node.Value.(*parser.Node); ok {
					if valueNode.Type == parser.NodeCall {
						className := a.extractClassNameFromCallNode(valueNode)
						if dep := a.callDependencyName(className, allClasses, false); dep != "" && include(node, dep) {
							a.addDependency(dependencies, dep, dependencyKindInstantiation)
						}
						// Note: function calls are NOT added to dependencies
					}
				}
			}
		case parser.NodeCall:
			// Function/class call - could be instantiation
			// Use structural AST analysis instead of string parsing
			className := a.extractClassNameFromCallNode(node)
			if dep := a.callDependencyName(className, allClasses, false); dep != "" && include(node, dep) {
				a.addDependency(dependencies, dep, dependencyKindInstantiation)
			}
			// Note: function calls are NOT added to dependencies
		case parser.NodeAttribute:
			// Attribute access: obj.method() or obj.attr
			if a.shouldSkipAttributeAccess(node) {
				return true
			}

			// Handle imported namespace members (e.g. cst.Arg): if the dotted
			// name resolves to a class reference, count it as coupling. The same
			// heuristic used for calls avoids counting functions like os.getcwd.
			attrName := a.extractClassName(node)
			if dep := a.callDependencyName(attrName, allClasses, true); dep != "" && include(node, dep) {
				a.addDependency(dependencies, dep, dependencyKindAttributeAccess)
				break
			}

			objNode := node.Left
			if objNode == nil {
				if valueNode, ok := node.Value.(*parser.Node); ok {
					objNode = valueNode
				}
			}
			if objNode != nil {
				objType := a.inferObjectType(objNode)
				if a.isAttributeReceiverDependency(objType, allClasses) && include(node, objType) {
					a.addDependency(dependencies, objType, dependencyKindAttributeAccess)
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

// nestedClassDefs maps each class name to the class-definition nodes for
// classes defined within classNode's subtree (its body, methods and nested
// functions), excluding classNode itself.
func (a *CBOAnalyzer) nestedClassDefs(classNode *parser.Node) map[string][]*parser.Node {
	defs := make(map[string][]*parser.Node)
	a.walkNode(classNode, func(node *parser.Node) bool {
		if node != classNode && node.Type == parser.NodeClassDef && node.Name != "" {
			defs[node.Name] = append(defs[node.Name], node)
		}
		return true
	})
	return defs
}

// buildParentMap records each node's parent across classNode's subtree so a
// reference's lexical scope can be resolved without relying on
// parser-populated Parent pointers.
func (a *CBOAnalyzer) buildParentMap(classNode *parser.Node) map[*parser.Node]*parser.Node {
	parents := make(map[*parser.Node]*parser.Node)
	a.walkNode(classNode, func(node *parser.Node) bool {
		for _, child := range node.GetChildren() {
			if child != nil {
				parents[child] = node
			}
		}
		return true
	})
	return parents
}

// enclosingScope returns the nearest enclosing function of node, or classNode
// if node is in the class body. Python class scope does not enclose its
// methods, so the class body and a method are distinct scopes.
func enclosingScope(node, classNode *parser.Node, parents map[*parser.Node]*parser.Node) *parser.Node {
	for p := parents[node]; p != nil; p = parents[p] {
		if p == classNode {
			return classNode
		}
		if p.Type == parser.NodeFunctionDef || p.Type == parser.NodeAsyncFunctionDef {
			return p
		}
	}
	return classNode
}

// isWithin reports whether node lies in ancestor's subtree.
func isWithin(node, ancestor *parser.Node, parents map[*parser.Node]*parser.Node) bool {
	for p := parents[node]; p != nil; p = parents[p] {
		if p == ancestor {
			return true
		}
	}
	return false
}

// resolvesToNestedClass reports whether a reference to name at ref resolves,
// per Python scoping, to a class defined within classNode rather than to an
// outer or top-level class. A function-local class is visible only within
// its enclosing function's subtree; a class defined directly in classNode's
// body is visible only in the class body, not in its methods. Only such
// references are internal coupling and should be excluded (see #547); a
// same-named class reached from a different scope is real coupling.
func resolvesToNestedClass(ref *parser.Node, name string, classNode *parser.Node, nestedDefs map[string][]*parser.Node, parents map[*parser.Node]*parser.Node) bool {
	return resolvesInScope(enclosingScope(ref, classNode, parents), name, classNode, nestedDefs, parents)
}

// resolvesInScope reports whether a reference whose Python evaluation scope is
// evalScope resolves to a class defined within classNode (internal coupling to
// exclude) rather than to an outer or top-level class. evalScope is where the
// name is actually looked up: a method body resolves in the method's own scope,
// but a method signature annotation resolves in the method's defining scope.
func resolvesInScope(evalScope *parser.Node, name string, classNode *parser.Node, nestedDefs map[string][]*parser.Node, parents map[*parser.Node]*parser.Node) bool {
	for _, def := range nestedDefs[name] {
		owner := enclosingScope(def, classNode, parents)
		if owner == classNode {
			// Defined in the class body: visible only to class-body-scope refs.
			// Python class scope does not enclose its methods, so a method body
			// cannot see it; a signature annotation (evaluated in the class body)
			// can.
			if evalScope == classNode {
				return true
			}
		} else if evalScope == owner || isWithin(evalScope, owner, parents) {
			// Defined in a function: visible throughout that function's subtree.
			return true
		}
	}
	return false
}

// nestedClassResolver decides, scope-aware, whether a reference is to a class
// defined within the analyzed class (internal coupling to exclude, see #547).
// It is built once per class and shared across the dependency passes.
type nestedClassResolver struct {
	classNode  *parser.Node
	nestedDefs map[string][]*parser.Node
	parents    map[*parser.Node]*parser.Node
}

func (a *CBOAnalyzer) newNestedClassResolver(classNode *parser.Node) *nestedClassResolver {
	nestedDefs := a.nestedClassDefs(classNode)
	var parents map[*parser.Node]*parser.Node
	if len(nestedDefs) > 0 {
		// Only needed when there is something to exclude.
		parents = a.buildParentMap(classNode)
	}
	return &nestedClassResolver{classNode: classNode, nestedDefs: nestedDefs, parents: parents}
}

// excludes reports whether a reference to name at ref resolves to a class
// nested within the analyzed class and so must not be counted as coupling.
// The reference is resolved at its own lexical scope.
func (r *nestedClassResolver) excludes(ref *parser.Node, name string) bool {
	if len(r.nestedDefs) == 0 {
		return false
	}
	return resolvesToNestedClass(ref, name, r.classNode, r.nestedDefs, r.parents)
}

// excludesInScope is like excludes but resolves name in a caller-supplied
// evaluation scope rather than at a reference's lexical position. It is used
// for method signature annotations, which Python evaluates in the method's
// defining scope (see signatureScope).
func (r *nestedClassResolver) excludesInScope(name string, evalScope *parser.Node) bool {
	if len(r.nestedDefs) == 0 {
		return false
	}
	return resolvesInScope(evalScope, name, r.classNode, r.nestedDefs, r.parents)
}

// signatureScope returns the scope in which methodNode's signature annotations
// are evaluated: the scope that defines the method (the class body for a
// normal method), not the method body.
func (r *nestedClassResolver) signatureScope(methodNode *parser.Node) *parser.Node {
	return enclosingScope(methodNode, r.classNode, r.parents)
}

// collectImports collects import statements and their aliases.
// It returns both the alias -> module.name map and a reverse map of
// module.name -> alias for namespace aliases created by "import ... as ...".
func (a *CBOAnalyzer) collectImports(ast *parser.Node) cboImportMaps {
	imports := cboImportMaps{
		importedNames:    make(map[string]string),
		namespaceAliases: make(map[string]string),
	}

	a.walkNode(ast, func(node *parser.Node) bool {
		switch node.Type {
		case parser.NodeImport:
			// import module as alias
			aliasedNames := make(map[string]bool)
			for _, child := range node.Children {
				if child.Type == parser.NodeAlias {
					module := child.Name
					alias := importBindingName(module)
					if child.Value != nil {
						if aliasStr, ok := child.Value.(string); ok {
							alias = aliasStr
							aliasedNames[module] = true
						}
					}
					imports.importedNames[alias] = module
					// Only true aliases (different from the module binding) form a
					// namespace that should be collapsed.
					if alias != module {
						imports.namespaceAliases[module] = alias
					}
				}
			}
			for _, name := range node.Names {
				if !aliasedNames[name] {
					imports.importedNames[importBindingName(name)] = name
				}
			}
		case parser.NodeImportFrom:
			// from module import name as alias
			module := node.Module
			aliasedNames := make(map[string]bool)
			for _, child := range node.Children {
				if child.Type == parser.NodeAlias {
					name := child.Name
					alias := name // Default to imported name
					if child.Value != nil {
						if aliasStr, ok := child.Value.(string); ok {
							alias = aliasStr
							aliasedNames[name] = true
						}
					}
					imports.importedNames[alias] = module + "." + name
				}
			}
			for _, name := range node.Names {
				if !aliasedNames[name] {
					imports.importedNames[name] = module + "." + name
				}
			}
		}
		return true
	})

	return imports
}

// hasFutureAnnotations reports whether the module contains
// `from __future__ import annotations` (PEP 563). When present, every
// annotation in the file is evaluated as a string at runtime rather than
// resolved eagerly, so annotation-only references carry no import cost.
//
// Tree-sitter's Python grammar parses `from __future__ import ...` as its
// own "future_import_statement" node rather than the general
// import_from_statement, and pyscn's AST builder has no dedicated case for
// it, so it falls through to a generic node whose Type is the raw grammar
// name and whose children are the raw parse-tree tokens. We detect it by
// that raw type and look for an "annotations" name among its children
// rather than relying on node.Module/node.Names, which only ImportFrom
// nodes populate.
func (a *CBOAnalyzer) hasFutureAnnotations(ast *parser.Node) bool {
	found := false
	a.walkNode(ast, func(node *parser.Node) bool {
		if found {
			return false
		}
		if node.Type == parser.NodeImportFrom && node.Module == "__future__" {
			for _, name := range node.Names {
				if name == "annotations" {
					found = true
					return false
				}
			}
		}
		if string(node.Type) == "future_import_statement" && importsAnnotationsName(node) {
			found = true
			return false
		}
		return true
	})
	return found
}

// importsAnnotationsName reports whether an "annotations" Name token appears
// anywhere among node's descendants. Used to inspect the raw children of a
// future_import_statement node (see hasFutureAnnotations).
func importsAnnotationsName(node *parser.Node) bool {
	if node == nil {
		return false
	}
	if node.Type == parser.NodeName && node.Name == "annotations" {
		return true
	}
	for _, child := range node.Children {
		if importsAnnotationsName(child) {
			return true
		}
	}
	return false
}

func importBindingName(module string) string {
	if strings.Contains(module, ".") {
		return strings.SplitN(module, ".", 2)[0]
	}
	return module
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
		left := a.extractClassName(node.Left)
		if left == "" {
			if valueNode, ok := node.Value.(*parser.Node); ok {
				left = a.extractClassName(valueNode)
			}
		}

		right := a.extractClassName(node.Right)
		if right == "" {
			right = node.Name
		}

		if left != "" && right != "" {
			return left + "." + right
		}
		if right != "" {
			return right
		}
	case parser.NodeSubscript:
		// Handle generic types like List[User], Dict[str, User]
		if valueNode, ok := node.Value.(*parser.Node); ok {
			return a.extractClassName(valueNode)
		}
	case parser.NodeTuple, parser.NodeList:
		for _, child := range node.Children {
			if name := a.extractClassName(child); name != "" {
				return name
			}
		}
	}

	return ""
}

func (a *CBOAnalyzer) typeArgumentNodes(node *parser.Node) []*parser.Node {
	if node == nil || node.Type != parser.NodeSubscript {
		return nil
	}
	if node.Right != nil {
		return []*parser.Node{node.Right}
	}
	return node.Children
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

	return true
}

// callDependencyName returns the portion of a called dotted name that refers
// to a class, or "" when the reference does not read as class coupling. The
// class part is found by trying dotted prefixes from longest to shortest, so
// class-method calls couple to the class itself: Path.cwd() -> Path,
// datetime.datetime.now() -> datetime.datetime. Locally defined classes are
// matched structurally; imported names carry no type information in
// single-file analysis, so they only count when a prefix reads as a class —
// otherwise function calls like os.getcwd() or suppress() would be reported
// as class coupling.
func (a *CBOAnalyzer) callDependencyName(className string, allClasses map[string]*parser.Node, suppressUppercaseLeaf bool) string {
	if className == "" {
		return ""
	}
	imported := a.isImportedDependency(className)
	parts := strings.Split(className, ".")

	// Enum members and other class constants are attribute reads on the imported
	// class, not distinct classes. Prefer the imported class binding before the
	// generic dotted-name heuristic has a chance to treat ALL_CAPS members as
	// CapWords-style class references. For non-call attribute reads, module
	// constants such as re.DOTALL and subprocess.PIPE are not classes and should
	// not count as coupling. Keep call targets such as uuid.UUID() or httpx.URL()
	// eligible because they are constructor-shaped references.
	if imported && len(parts) > 1 {
		if binding := parts[0]; a.isImportedDependency(binding) && a.looksLikeClassReference(binding) {
			return binding
		}
		if suppressUppercaseLeaf && isUppercaseConstant(parts[len(parts)-1]) {
			return ""
		}
	}

	for end := len(parts); end > 0; end-- {
		prefix := strings.Join(parts[:end], ".")
		if _, isClass := allClasses[prefix]; isClass {
			return prefix
		}
		if imported && a.looksLikeClassReference(prefix) {
			return prefix
		}
	}
	return ""
}

// knownLowercaseClasses lists well-known standard library classes that do not
// follow the PEP 8 CapWords convention and would otherwise be rejected by the
// capitalization heuristic in looksLikeClassReference.
var knownLowercaseClasses = map[string]bool{
	"datetime.datetime":       true,
	"datetime.date":           true,
	"datetime.time":           true,
	"datetime.timedelta":      true,
	"datetime.timezone":       true,
	"collections.defaultdict": true,
	"collections.deque":       true,
	"array.array":             true,
	"socket.socket":           true,
	"threading.local":         true,
}

// resolveImportedName expands an alias or module-qualified name to the full
// imported path (e.g. "FilePath" -> "pathlib.Path", "os.getcwd" -> "os.getcwd").
func (a *CBOAnalyzer) resolveImportedName(name string) string {
	if full, exists := a.importedNames[name]; exists {
		return full
	}
	if strings.Contains(name, ".") {
		parts := strings.SplitN(name, ".", 2)
		if full, exists := a.importedNames[parts[0]]; exists {
			return full + "." + parts[1]
		}
	}
	return name
}

// looksLikeClassReference applies the PEP 8 naming convention to decide
// whether an imported name refers to a class (CapWords) rather than a
// function (snake_case). Aliases resolve to their original imported name so
// the convention is judged on the name the defining module chose.
func (a *CBOAnalyzer) looksLikeClassReference(name string) bool {
	resolved := a.resolveImportedName(name)
	if knownLowercaseClasses[resolved] {
		return true
	}
	leaf := strings.TrimLeft(dependencyLeafName(resolved), "_")
	for _, r := range leaf {
		// Only a clearly lower-case leading letter marks a function; uncased
		// scripts keep counting as classes to stay conservative.
		return !unicode.IsLower(r)
	}
	return false
}

func (a *CBOAnalyzer) isAttributeReceiverDependency(className string, allClasses map[string]*parser.Node) bool {
	if className == "" || className == "self" || className == "cls" {
		return false
	}
	if _, isClass := allClasses[className]; isClass {
		return true
	}
	// Same reasoning as callDependencyName: an imported receiver is only class
	// coupling when the name reads as a class; plain modules (os, json, ...)
	// are not classes.
	if a.isImportedDependency(className) {
		return a.looksLikeClassReference(className)
	}
	return a.options.IncludeBuiltins && a.builtinTypes[className]
}

func (a *CBOAnalyzer) shouldSkipAttributeAccess(node *parser.Node) bool {
	if node == nil {
		return true
	}
	if node.Parent != nil && node.Parent.Type == parser.NodeCall {
		return true
	}
	if a.isWithinTypeAnnotation(node) || a.isWithinClassBase(node) {
		return true
	}
	return false
}

func (a *CBOAnalyzer) isWithinTypeAnnotation(node *parser.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Type {
		case parser.NodeAnnAssign:
			for _, child := range current.Children {
				if child != nil && a.isTypeAnnotation(child) && nodeDescendsFrom(node, child) {
					return true
				}
			}
			return false
		case parser.NodeArg:
			return current.Right != nil && nodeDescendsFrom(node, current.Right)
		case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
			return current.Right != nil && nodeDescendsFrom(node, current.Right)
		case parser.NodeTypeNode, parser.NodeGenericType, parser.NodeTypeParameter:
			return true
		}
	}
	return false
}

func (a *CBOAnalyzer) isWithinClassBase(node *parser.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Type != parser.NodeClassDef {
			continue
		}
		for _, base := range current.Bases {
			if nodeDescendsFrom(node, base) {
				return true
			}
		}
		return false
	}
	return false
}

func nodeDescendsFrom(node, ancestor *parser.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}

func (a *CBOAnalyzer) shouldIncludeDependencyForClass(className, ownerClass string) bool {
	if !a.shouldIncludeDependency(className) {
		return false
	}
	if ownerClass != "" && a.sameDependencyIdentity(className, ownerClass) {
		return false
	}
	if a.isTypeSystemDependency(className) {
		return false
	}
	if !a.options.IncludeImports && a.isStandardLibraryDependency(className) {
		return false
	}
	return true
}

func (a *CBOAnalyzer) addDependency(dependencies *cboDependencies, className string, kind cboDependencyKind) {
	if dependencies == nil || className == "" {
		return
	}

	// Under PEP 563 (`from __future__ import annotations`), annotations are
	// stored as strings and never evaluated at runtime. A reference that
	// appears only inside a type annotation therefore has zero import cost
	// and should not be counted as coupling. See #628.
	if kind == dependencyKindTypeHint && a.futureAnnotations {
		return
	}

	className = a.canonicalDependencyName(className)

	dependencies.all[className] = true

	if a.isImportedDependency(className) {
		dependencies.imports[className] = true
		return
	}

	switch kind {
	case dependencyKindInheritance:
		dependencies.inheritance[className] = true
	case dependencyKindTypeHint:
		dependencies.typeHints[className] = true
	case dependencyKindInstantiation:
		dependencies.instantiations[className] = true
	case dependencyKindAttributeAccess:
		dependencies.attributeAccesses[className] = true
	case dependencyKindImport:
		dependencies.imports[className] = true
	}
}

// canonicalDependencyName collapses references through a namespace alias to the
// alias itself when GroupNamespaceImports is enabled. This prevents a single
// "import libcst as cst" from inflating CBO by one edge per cst.Member usage.
func (a *CBOAnalyzer) canonicalDependencyName(className string) string {
	if !a.options.GroupNamespaceImports {
		return className
	}
	if className == "" {
		return className
	}

	// If the name is a direct reference to an aliased module, prefer the alias.
	if alias, exists := a.namespaceAliases[className]; exists {
		return alias
	}

	// If the name is alias.Member and the leading part is a real namespace
	// alias (created by "import ... as ..."), collapse to the alias.
	if strings.Contains(className, ".") {
		parts := strings.SplitN(className, ".", 2)
		alias := parts[0]
		if module, exists := a.importedNames[alias]; exists {
			if canonicalAlias, aliased := a.namespaceAliases[module]; aliased && canonicalAlias == alias {
				return alias
			}
		}
	}

	return className
}

func (a *CBOAnalyzer) sameDependencyIdentity(className, ownerClass string) bool {
	if className == ownerClass {
		return true
	}
	if imported, exists := a.importedNames[className]; exists {
		return dependencyLeafName(imported) == ownerClass
	}
	return false
}

func dependencyLeafName(className string) string {
	if strings.Contains(className, ".") {
		parts := strings.Split(className, ".")
		return parts[len(parts)-1]
	}
	return className
}

func isUppercaseConstant(name string) bool {
	if name == "" {
		return false
	}
	hasLetter := false
	for _, r := range name {
		if unicode.IsLetter(r) {
			hasLetter = true
			if unicode.IsLower(r) {
				return false
			}
		}
	}
	return hasLetter
}

func (a *CBOAnalyzer) isTypeSystemDependency(className string) bool {
	if className == "" {
		return false
	}

	if imported, exists := a.importedNames[className]; exists {
		return a.typeSystemRoot(imported)
	}
	if strings.Contains(className, ".") {
		return a.typeSystemRoot(className)
	}

	return false
}

func (a *CBOAnalyzer) isStandardLibraryDependency(className string) bool {
	if className == "" {
		return false
	}

	if imported, exists := a.importedNames[className]; exists {
		return a.standardLibraryRoot(imported)
	}
	if strings.Contains(className, ".") {
		return a.standardLibraryRoot(className)
	}

	return false
}

func (a *CBOAnalyzer) standardLibraryRoot(className string) bool {
	root := className
	if strings.Contains(root, ".") {
		root = strings.SplitN(root, ".", 2)[0]
	}
	return a.standardLibs[root]
}

func (a *CBOAnalyzer) typeSystemRoot(className string) bool {
	root := className
	if strings.Contains(root, ".") {
		root = strings.SplitN(root, ".", 2)[0]
	}

	switch root {
	case "abc", "typing", "typing_extensions":
		return true
	default:
		return false
	}
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
			switch valueNode.Type {
			case parser.NodeName:
				return valueNode.Name
			case parser.NodeAttribute:
				return a.extractClassNameFromAttribute(valueNode)
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

	rightName := attrNode.Name
	if attrNode.Right != nil && attrNode.Right.Type == parser.NodeName {
		rightName = attrNode.Right.Name
	}
	if rightName == "" {
		return ""
	}

	leftName := a.extractClassName(attrNode.Left)
	if leftName == "" {
		if valueNode, ok := attrNode.Value.(*parser.Node); ok {
			leftName = a.extractClassName(valueNode)
		}
	}
	if leftName != "" {
		return leftName + "." + rightName
	}

	return rightName
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
	if node == nil {
		return
	}
	node.Walk(visitor)
}

// inferObjectType tries to infer the type of an object from context.
// We return the local binding name so that namespace aliases like
// "import libcst as cst" collapse consistently in addDependency.
func (a *CBOAnalyzer) inferObjectType(node *parser.Node) string {
	// Simple heuristic - try to extract type from variable name or method call
	if node.Type == parser.NodeName {
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
	sort.Strings(result)
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
		// Cython primitive type qualifiers
		"cython.int", "cython.float", "cython.long", "cython.uint",
		"cython.ulong", "cython.ulonglong", "cython.ushort", "cython.double",
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
		"abc", "typing", "typing_extensions",
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
