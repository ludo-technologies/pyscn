package analyzer

import (
	"fmt"
	"sort"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// LCOMResult holds LCOM4 (Lack of Cohesion of Methods) metrics for a class
type LCOMResult struct {
	// Core LCOM4 metric - number of connected components
	LCOM4 int

	// Class information
	ClassName string
	FilePath  string
	StartLine int
	EndLine   int

	// Method statistics
	TotalMethods    int // All methods found in class
	ExcludedMethods int // @staticmethod and @classmethod excluded

	// Instance variable count
	InstanceVariables int // Distinct self.xxx variables

	// Connected component details
	MethodGroups [][]string // Method names grouped by connected component

	// Risk assessment
	RiskLevel string // "low", "medium", "high"
}

// LCOMOptions configures LCOM analysis behavior
type LCOMOptions struct {
	LowThreshold    int // Default: 2 (LCOM4 <= 2 is low risk)
	MediumThreshold int // Default: 5 (LCOM4 3-5 is medium risk)
	ExcludePatterns []string
}

// DefaultLCOMOptions returns default LCOM analysis options
func DefaultLCOMOptions() *LCOMOptions {
	return &LCOMOptions{
		LowThreshold:    domain.DefaultLCOMLowThreshold,
		MediumThreshold: domain.DefaultLCOMMediumThreshold,
		ExcludePatterns: []string{},
	}
}

// LCOMAnalyzer analyzes class cohesion in Python code
type LCOMAnalyzer struct {
	options *LCOMOptions
}

// NewLCOMAnalyzer creates a new LCOM analyzer
func NewLCOMAnalyzer(options *LCOMOptions) *LCOMAnalyzer {
	if options == nil {
		options = DefaultLCOMOptions()
	}
	return &LCOMAnalyzer{
		options: options,
	}
}

// AnalyzeClasses analyzes LCOM4 for all classes in the given AST
func (a *LCOMAnalyzer) AnalyzeClasses(ast *parser.Node, filePath string) ([]*LCOMResult, error) {
	if ast == nil {
		return nil, fmt.Errorf("AST is nil")
	}

	// Collect all class definitions
	classes := a.collectClasses(ast)
	ctypesFields := a.collectCtypesFields(ast, classes)

	var results []*LCOMResult
	for _, classNode := range classes {
		result, err := a.analyzeClass(classNode, filePath, ctypesFields[classNode])
		if err != nil {
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// analyzeClass computes LCOM4 for a single class using connected components
func (a *LCOMAnalyzer) analyzeClass(classNode *parser.Node, filePath string, declaredFields map[string]bool) (*LCOMResult, error) {
	if classNode.Type != parser.NodeClassDef {
		return nil, fmt.Errorf("node is not a class definition")
	}

	result := &LCOMResult{
		ClassName:    classNode.Name,
		FilePath:     filePath,
		StartLine:    classNode.Location.StartLine,
		EndLine:      classNode.Location.EndLine,
		MethodGroups: [][]string{},
	}

	// Step 1: Collect methods, their instance variable accesses, and intra-class calls
	methods, excluded, methodCalls := a.collectMethods(classNode, declaredFields)
	result.TotalMethods = len(methods) + excluded
	result.ExcludedMethods = excluded

	// Step 2: Collect all distinct instance variables
	allVars := make(map[string]bool)
	for _, vars := range methods {
		for v := range vars {
			allVars[v] = true
		}
	}
	for v := range declaredFields {
		allVars[v] = true
	}
	result.InstanceVariables = len(allVars)

	// Classes with 0 or 1 methods are trivially cohesive
	if len(methods) <= 1 {
		result.LCOM4 = 1
		if len(methods) == 1 {
			for name := range methods {
				result.MethodGroups = [][]string{{name}}
			}
		}
		result.RiskLevel = a.assessRiskLevel(1)
		return result, nil
	}

	// Step 3: Compute connected components using Union-Find
	methodNames := make([]string, 0, len(methods))
	for name := range methods {
		methodNames = append(methodNames, name)
	}
	sort.Strings(methodNames) // Deterministic ordering

	// Initialize Union-Find
	parent := make(map[string]string, len(methodNames))
	rank := make(map[string]int, len(methodNames))
	for _, name := range methodNames {
		parent[name] = name
		rank[name] = 0
	}

	var find func(string) string
	find = func(x string) string {
		if parent[x] != x {
			parent[x] = find(parent[x]) // Path compression
		}
		return parent[x]
	}

	union := func(x, y string) {
		rx := find(x)
		ry := find(y)
		if rx == ry {
			return
		}
		// Union by rank
		if rank[rx] < rank[ry] {
			parent[rx] = ry
		} else if rank[rx] > rank[ry] {
			parent[ry] = rx
		} else {
			parent[ry] = rx
			rank[rx]++
		}
	}

	// Build variable -> methods mapping for efficient edge detection
	varToMethods := make(map[string][]string)
	for _, name := range methodNames {
		for v := range methods[name] {
			varToMethods[v] = append(varToMethods[v], name)
		}
	}

	// Union methods that share instance variables
	for _, methodList := range varToMethods {
		for i := 1; i < len(methodList); i++ {
			union(methodList[0], methodList[i])
		}
	}

	// Union methods connected by intra-class method calls (self.xxx())
	for caller, callees := range methodCalls {
		for callee := range callees {
			if _, exists := methods[callee]; exists {
				union(caller, callee)
			}
		}
	}

	// Step 4: Count connected components and build groups
	components := make(map[string][]string) // root -> method names
	for _, name := range methodNames {
		root := find(name)
		components[root] = append(components[root], name)
	}

	result.LCOM4 = len(components)
	for _, group := range components {
		sort.Strings(group)
		result.MethodGroups = append(result.MethodGroups, group)
	}
	// Sort groups for deterministic output
	sort.Slice(result.MethodGroups, func(i, j int) bool {
		return result.MethodGroups[i][0] < result.MethodGroups[j][0]
	})

	result.RiskLevel = a.assessRiskLevel(result.LCOM4)
	return result, nil
}

// collectMethods extracts instance methods and their self.xxx variable accesses from a class.
// Returns:
//   - methods: methodName -> set of instance variable names
//   - excluded: count of excluded methods (@classmethod/@staticmethod)
//   - calls: methodName -> set of called method names (self.xxx() intra-class calls)
func (a *LCOMAnalyzer) collectMethods(classNode *parser.Node, declaredFields map[string]bool) (map[string]map[string]bool, int, map[string]map[string]bool) {
	methods := make(map[string]map[string]bool)
	calls := make(map[string]map[string]bool)
	excluded := 0

	for _, node := range classNode.Body {
		if node == nil {
			continue
		}

		if node.Type != parser.NodeFunctionDef && node.Type != parser.NodeAsyncFunctionDef {
			continue
		}

		// Check decorators to exclude @classmethod and @staticmethod
		if a.isClassOrStaticMethod(node) {
			excluded++
			continue
		}

		// Extract self.xxx accesses from method body
		vars := make(map[string]bool)
		a.extractInstanceVars(node, vars)
		if len(declaredFields) > 0 && a.hasBareSelfAccess(node) {
			for field := range declaredFields {
				vars[field] = true
			}
		}
		methods[node.Name] = vars

		// Extract self.xxx() method calls
		methodCalls := make(map[string]bool)
		a.extractMethodCalls(node, methodCalls)
		calls[node.Name] = methodCalls
	}

	return methods, excluded, calls
}

// isClassOrStaticMethod checks if a method has a @classmethod or @staticmethod decorator.
func (a *LCOMAnalyzer) isClassOrStaticMethod(funcNode *parser.Node) bool {
	for _, decorator := range funcNode.Decorator {
		if decorator == nil {
			continue
		}
		name := a.getDecoratorName(decorator)
		if name == "classmethod" || name == "staticmethod" {
			return true
		}
	}
	return false
}

// getDecoratorName extracts the decorator name from a decorator node
func (a *LCOMAnalyzer) getDecoratorName(decorator *parser.Node) string {
	// Check Name field first (used by some paths)
	if decorator.Name != "" {
		return decorator.Name
	}
	// Check Value field (set by buildDecorator)
	if decorator.Value != nil {
		if nameNode, ok := decorator.Value.(*parser.Node); ok {
			if nameNode.Type == parser.NodeName {
				return nameNode.Name
			}
			// For decorator calls like @decorator(args), extract function name
			if nameNode.Type == parser.NodeCall {
				if nameNode.Value != nil {
					if funcNode, ok := nameNode.Value.(*parser.Node); ok && funcNode.Type == parser.NodeName {
						return funcNode.Name
					}
				}
				// Also check Name field on Call node
				if nameNode.Name != "" {
					return nameNode.Name
				}
			}
		}
	}
	return ""
}

// extractMethodCalls walks a method's AST to find all self.xxx() method call targets
func (a *LCOMAnalyzer) extractMethodCalls(methodNode *parser.Node, calls map[string]bool) {
	methodNode.WalkDeep(func(node *parser.Node) bool {
		if node.Type == parser.NodeCall && node.Value != nil {
			if attrNode, ok := node.Value.(*parser.Node); ok {
				if attrNode.Type == parser.NodeAttribute && a.isSelfAccess(attrNode) && attrNode.Name != "" {
					calls[attrNode.Name] = true
				}
			}
		}
		return true
	})
}

// extractInstanceVars walks a method's AST to find all self.xxx attribute accesses
func (a *LCOMAnalyzer) extractInstanceVars(methodNode *parser.Node, vars map[string]bool) {
	methodNode.WalkDeep(func(node *parser.Node) bool {
		if node.Type == parser.NodeAttribute {
			// Check if this is self.xxx
			if a.isSelfAccess(node) && node.Name != "" {
				vars[node.Name] = true
			}
		}
		return true
	})
}

// hasBareSelfAccess reports whether a method uses self as a value rather than
// only as the base object in self.attr. ctypes.Structure methods commonly pass
// self to C functions that read or mutate fields declared in _fields_.
func (a *LCOMAnalyzer) hasBareSelfAccess(node *parser.Node) bool {
	if node == nil {
		return false
	}

	switch node.Type {
	case parser.NodeFunctionDef, parser.NodeAsyncFunctionDef:
		for _, bodyNode := range node.Body {
			if a.hasBareSelfAccess(bodyNode) {
				return true
			}
		}
		return false
	case parser.NodeName:
		return node.Name == "self"
	case parser.NodeAttribute:
		if valueNode := nodeValue(node); valueNode != nil {
			if valueNode.Type == parser.NodeName && valueNode.Name == "self" {
				return false
			}
			if a.hasBareSelfAccess(valueNode) {
				return true
			}
		}
		for _, child := range parser.OrderedChildren(node, nil) {
			if child != nodeValue(node) && a.hasBareSelfAccess(child) {
				return true
			}
		}
		return false
	case parser.NodeCall:
		if valueNode := nodeValue(node); valueNode != nil && a.hasBareSelfAccess(valueNode) {
			return true
		}
		for _, arg := range node.Args {
			if a.hasBareSelfAccess(arg) {
				return true
			}
		}
		for _, keyword := range node.Keywords {
			if a.hasBareSelfAccess(keyword) {
				return true
			}
		}
		return false
	}

	for _, child := range parser.OrderedChildren(node, nil) {
		if a.hasBareSelfAccess(child) {
			return true
		}
	}
	return false
}

// isSelfAccess checks if an attribute node represents self.xxx access
func (a *LCOMAnalyzer) isSelfAccess(attrNode *parser.Node) bool {
	// The object (self) can be in Value or Left field
	if attrNode.Value != nil {
		if nameNode, ok := attrNode.Value.(*parser.Node); ok {
			if nameNode.Type == parser.NodeName && nameNode.Name == "self" {
				return true
			}
		}
	}
	if attrNode.Left != nil {
		if attrNode.Left.Type == parser.NodeName && attrNode.Left.Name == "self" {
			return true
		}
	}
	return false
}

// assessRiskLevel determines the risk level based on LCOM4 value
func (a *LCOMAnalyzer) assessRiskLevel(lcom4 int) string {
	switch {
	case lcom4 <= a.options.LowThreshold:
		return string(domain.RiskLevelLow)
	case lcom4 <= a.options.MediumThreshold:
		return string(domain.RiskLevelMedium)
	default:
		return string(domain.RiskLevelHigh)
	}
}

// collectClasses collects all class definition nodes from the AST
func (a *LCOMAnalyzer) collectClasses(ast *parser.Node) []*parser.Node {
	var classes []*parser.Node
	ast.WalkDeep(func(node *parser.Node) bool {
		if node.Type == parser.NodeClassDef && node.Name != "" {
			classes = append(classes, node)
		}
		return true
	})
	return classes
}

func (a *LCOMAnalyzer) collectCtypesFields(ast *parser.Node, classes []*parser.Node) map[*parser.Node]map[string]bool {
	fieldsByClass := make(map[*parser.Node]map[string]bool)
	// Map class name -> node, with nil marking names that resolve ambiguously
	// (e.g. two ctypes classes share a name across scopes). Ambiguous names are
	// skipped for external `<Name>._fields_ = ...` assignments since we cannot
	// safely attribute them to a specific class.
	ctypesByName := make(map[string]*parser.Node, len(classes))
	var ctypesClasses []*parser.Node
	for _, classNode := range classes {
		if !a.isCtypesStructureClass(classNode) {
			continue
		}
		ctypesClasses = append(ctypesClasses, classNode)
		if existing, ok := ctypesByName[classNode.Name]; ok {
			if existing != nil {
				ctypesByName[classNode.Name] = nil
			}
			continue
		}
		ctypesByName[classNode.Name] = classNode
	}
	if len(ctypesClasses) == 0 {
		return fieldsByClass
	}

	for _, classNode := range ctypesClasses {
		for _, bodyNode := range classNode.Body {
			if bodyNode == nil {
				continue
			}
			if !isAssignmentNode(bodyNode) || !a.assignmentTargetsName(bodyNode, "_fields_") {
				continue
			}
			a.addFields(fieldsByClass, classNode, a.extractCtypesFieldNames(bodyNode.Value))
		}
	}

	ast.WalkDeep(func(node *parser.Node) bool {
		if node == nil || !isAssignmentNode(node) {
			return true
		}
		for _, target := range node.Targets {
			className := a.ctypesFieldsAssignmentClass(target)
			if className == "" {
				continue
			}
			classNode, ok := ctypesByName[className]
			if !ok || classNode == nil {
				continue
			}
			a.addFields(fieldsByClass, classNode, a.extractCtypesFieldNames(node.Value))
		}
		return true
	})

	return fieldsByClass
}

// ctypesStructureBases lists the ctypes base classes whose subclasses declare
// fields via the _fields_ class attribute (often assigned outside the class
// body to allow self-referential pointer types).
var ctypesStructureBases = map[string]bool{
	"Structure":             true,
	"Union":                 true,
	"BigEndianStructure":    true,
	"LittleEndianStructure": true,
	"BigEndianUnion":        true,
	"LittleEndianUnion":     true,
}

func (a *LCOMAnalyzer) isCtypesStructureClass(classNode *parser.Node) bool {
	if classNode == nil || classNode.Type != parser.NodeClassDef {
		return false
	}
	for _, base := range classNode.Bases {
		if base == nil {
			continue
		}
		switch base.Type {
		case parser.NodeName, parser.NodeAttribute:
			if ctypesStructureBases[base.Name] {
				return true
			}
		}
	}
	return false
}

func (a *LCOMAnalyzer) assignmentTargetsName(assignNode *parser.Node, name string) bool {
	for _, target := range assignNode.Targets {
		if target != nil && target.Type == parser.NodeName && target.Name == name {
			return true
		}
	}
	return false
}

func (a *LCOMAnalyzer) ctypesFieldsAssignmentClass(target *parser.Node) string {
	if target == nil || target.Type != parser.NodeAttribute || target.Name != "_fields_" {
		return ""
	}
	valueNode := nodeValue(target)
	if valueNode == nil || valueNode.Type != parser.NodeName {
		return ""
	}
	return valueNode.Name
}

func (a *LCOMAnalyzer) extractCtypesFieldNames(value interface{}) []string {
	valueNode, ok := value.(*parser.Node)
	if !ok || valueNode == nil {
		return nil
	}

	var fields []string
	for _, fieldNode := range valueNode.Children {
		if fieldNode == nil || fieldNode.Type != parser.NodeTuple || len(fieldNode.Children) == 0 {
			continue
		}
		nameNode := fieldNode.Children[0]
		if nameNode == nil || nameNode.Type != parser.NodeConstant {
			continue
		}
		fieldName, ok := nameNode.Value.(string)
		if !ok || fieldName == "" {
			continue
		}
		fields = append(fields, fieldName)
	}
	return fields
}

func (a *LCOMAnalyzer) addFields(fieldsByClass map[*parser.Node]map[string]bool, classNode *parser.Node, fields []string) {
	if len(fields) == 0 || classNode == nil {
		return
	}
	if fieldsByClass[classNode] == nil {
		fieldsByClass[classNode] = make(map[string]bool, len(fields))
	}
	for _, field := range fields {
		fieldsByClass[classNode][field] = true
	}
}

func isAssignmentNode(node *parser.Node) bool {
	return node != nil && (node.Type == parser.NodeAssign || node.Type == parser.NodeAnnAssign)
}

// CalculateLCOM is a convenience function that creates an analyzer with defaults and runs it
func CalculateLCOM(ast *parser.Node, filePath string) ([]*LCOMResult, error) {
	analyzer := NewLCOMAnalyzer(nil)
	return analyzer.AnalyzeClasses(ast, filePath)
}

// CalculateLCOMWithConfig creates an analyzer with custom options and runs it
func CalculateLCOMWithConfig(ast *parser.Node, filePath string, options *LCOMOptions) ([]*LCOMResult, error) {
	analyzer := NewLCOMAnalyzer(options)
	return analyzer.AnalyzeClasses(ast, filePath)
}
