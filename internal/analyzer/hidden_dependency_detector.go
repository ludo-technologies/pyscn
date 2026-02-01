package analyzer

import (
	"fmt"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// HiddenDependencyDetector detects hidden dependency anti-patterns
type HiddenDependencyDetector struct {
	moduleVariables map[string]bool // Track module-level variable names
}

// NewHiddenDependencyDetector creates a new hidden dependency detector
func NewHiddenDependencyDetector() *HiddenDependencyDetector {
	return &HiddenDependencyDetector{
		moduleVariables: make(map[string]bool),
	}
}

// Analyze detects hidden dependencies in the given AST
func (d *HiddenDependencyDetector) Analyze(ast *parser.Node, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	// First pass: collect module-level variables
	d.collectModuleVariables(ast)

	// Detect global statement usage
	globalFindings := d.detectGlobalStatements(ast, filePath)
	findings = append(findings, globalFindings...)

	// Detect singleton pattern via _instance
	singletonFindings := d.detectSingletonPattern(ast, filePath)
	findings = append(findings, singletonFindings...)

	// Detect module variable access in classes
	moduleVarFindings := d.detectModuleVariableAccess(ast, filePath)
	findings = append(findings, moduleVarFindings...)

	return findings
}

// collectModuleVariables collects module-level variable names
func (d *HiddenDependencyDetector) collectModuleVariables(ast *parser.Node) {
	// Module-level variables are direct children of the module (not inside classes/functions)
	for _, node := range ast.Body {
		if node == nil {
			continue
		}

		// Skip class and function definitions
		if node.Type == parser.NodeClassDef || node.Type == parser.NodeFunctionDef || node.Type == parser.NodeAsyncFunctionDef {
			continue
		}

		// Handle assignments
		if node.Type == parser.NodeAssign {
			for _, target := range node.Targets {
				if target != nil && target.Type == parser.NodeName {
					// Skip private/dunder names for now (they're often intentional)
					if !strings.HasPrefix(target.Name, "__") {
						d.moduleVariables[target.Name] = true
					}
				}
			}
		}

		// Handle annotated assignments
		if node.Type == parser.NodeAnnAssign {
			for _, target := range node.Targets {
				if target != nil && target.Type == parser.NodeName {
					if !strings.HasPrefix(target.Name, "__") {
						d.moduleVariables[target.Name] = true
					}
				}
			}
			// Also check Name field for simple annotations
			if node.Name != "" && !strings.HasPrefix(node.Name, "__") {
				d.moduleVariables[node.Name] = true
			}
		}
	}
}

// detectGlobalStatements detects use of global statement inside classes
func (d *HiddenDependencyDetector) detectGlobalStatements(ast *parser.Node, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	classes := ast.FindByType(parser.NodeClassDef)

	for _, class := range classes {
		// Find global statements inside the class
		globalNodes := class.FindByType(parser.NodeGlobal)

		for _, globalNode := range globalNodes {
			// Get the names from the global statement
			globalNames := d.getGlobalNames(globalNode)
			namesStr := strings.Join(globalNames, ", ")
			if namesStr == "" {
				namesStr = "unknown"
			}

			// Find the containing method
			methodName := d.findContainingMethodName(globalNode, class)

			finding := domain.DIAntipatternFinding{
				Type:       domain.DIAntipatternHiddenDependency,
				Subtype:    string(domain.HiddenDepGlobal),
				Severity:   domain.DIAntipatternSeverityError,
				ClassName:  class.Name,
				MethodName: methodName,
				Location: domain.SourceLocation{
					FilePath:  filePath,
					StartLine: globalNode.Location.StartLine,
					EndLine:   globalNode.Location.EndLine,
					StartCol:  globalNode.Location.StartCol,
					EndCol:    globalNode.Location.EndCol,
				},
				Description: fmt.Sprintf("Uses global statement to access '%s'", namesStr),
				Suggestion:  "Inject the dependency as a constructor parameter instead of using global state",
				Details: map[string]interface{}{
					"global_names": globalNames,
				},
			}
			findings = append(findings, finding)
		}
	}

	return findings
}

// detectSingletonPattern detects singleton pattern via _instance class variable
func (d *HiddenDependencyDetector) detectSingletonPattern(ast *parser.Node, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	classes := ast.FindByType(parser.NodeClassDef)

	for _, class := range classes {
		// Check for _instance class variable
		if d.hasSingletonPattern(class) {
			finding := domain.DIAntipatternFinding{
				Type:      domain.DIAntipatternHiddenDependency,
				Subtype:   string(domain.HiddenDepSingleton),
				Severity:  domain.DIAntipatternSeverityWarning,
				ClassName: class.Name,
				Location: domain.SourceLocation{
					FilePath:  filePath,
					StartLine: class.Location.StartLine,
					EndLine:   class.Location.EndLine,
					StartCol:  class.Location.StartCol,
					EndCol:    class.Location.EndCol,
				},
				Description: "Class implements singleton pattern using '_instance' class variable",
				Suggestion:  "Consider using dependency injection to provide a single instance instead of singleton pattern",
				Details: map[string]interface{}{
					"pattern": "singleton",
				},
			}
			findings = append(findings, finding)
		}
	}

	return findings
}

// detectModuleVariableAccess detects direct access to module-level variables in methods
func (d *HiddenDependencyDetector) detectModuleVariableAccess(ast *parser.Node, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	if len(d.moduleVariables) == 0 {
		return findings
	}

	classes := ast.FindByType(parser.NodeClassDef)

	for _, class := range classes {
		// Find all methods in the class
		methods := d.findMethods(class)

		for _, method := range methods {
			// Find all name references in the method
			accessedVars := d.findModuleVariableAccesses(method, class.Name)

			for varName, location := range accessedVars {
				finding := domain.DIAntipatternFinding{
					Type:       domain.DIAntipatternHiddenDependency,
					Subtype:    string(domain.HiddenDepModuleVariable),
					Severity:   domain.DIAntipatternSeverityWarning,
					ClassName:  class.Name,
					MethodName: method.Name,
					Location: domain.SourceLocation{
						FilePath:  filePath,
						StartLine: location.StartLine,
						EndLine:   location.EndLine,
						StartCol:  location.StartCol,
						EndCol:    location.EndCol,
					},
					Description: fmt.Sprintf("Directly accesses module-level variable '%s'", varName),
					Suggestion:  "Inject the dependency as a constructor parameter instead of accessing module-level state",
					Details: map[string]interface{}{
						"variable_name": varName,
					},
				}
				findings = append(findings, finding)
			}
		}
	}

	return findings
}

// getGlobalNames extracts variable names from a global statement
func (d *HiddenDependencyDetector) getGlobalNames(globalNode *parser.Node) []string {
	var names []string

	// Check Names field
	if len(globalNode.Names) > 0 {
		names = append(names, globalNode.Names...)
	}

	// Also check children for Name nodes
	for _, child := range globalNode.Children {
		if child != nil && child.Type == parser.NodeName && child.Name != "" {
			names = append(names, child.Name)
		}
	}

	return names
}

// findContainingMethodName finds the method containing a node
func (d *HiddenDependencyDetector) findContainingMethodName(node *parser.Node, class *parser.Node) string {
	for _, method := range d.findMethods(class) {
		if d.nodeContains(method, node) {
			return method.Name
		}
	}
	return ""
}

// nodeContains checks if parent contains child
func (d *HiddenDependencyDetector) nodeContains(parent, child *parser.Node) bool {
	if parent.Location.StartLine > child.Location.StartLine {
		return false
	}
	if parent.Location.EndLine < child.Location.EndLine {
		return false
	}
	return true
}

// hasSingletonPattern checks if a class has singleton pattern
func (d *HiddenDependencyDetector) hasSingletonPattern(class *parser.Node) bool {
	// Look for _instance class variable
	for _, node := range class.Body {
		if node == nil {
			continue
		}

		// Check assignments at class level
		if node.Type == parser.NodeAssign {
			for _, target := range node.Targets {
				if target != nil && target.Type == parser.NodeName {
					if target.Name == "_instance" || target.Name == "__instance" {
						return true
					}
				}
			}
		}

		// Check annotated assignments
		if node.Type == parser.NodeAnnAssign {
			if node.Name == "_instance" || node.Name == "__instance" {
				return true
			}
		}
	}

	// Also check for cls._instance pattern in methods
	methods := d.findMethods(class)
	for _, method := range methods {
		if d.hasClsInstanceAccess(method) {
			return true
		}
	}

	return false
}

// hasClsInstanceAccess checks if a method accesses cls._instance
func (d *HiddenDependencyDetector) hasClsInstanceAccess(method *parser.Node) bool {
	found := false
	method.Walk(func(node *parser.Node) bool {
		if node.Type == parser.NodeAttribute {
			// Check for cls._instance or ClassName._instance pattern
			if node.Left != nil && node.Left.Type == parser.NodeName {
				if node.Right != nil && node.Right.Type == parser.NodeName {
					if node.Right.Name == "_instance" || node.Right.Name == "__instance" {
						found = true
						return false
					}
				}
			}
		}
		return true
	})
	return found
}

// findMethods finds all methods in a class
func (d *HiddenDependencyDetector) findMethods(class *parser.Node) []*parser.Node {
	var methods []*parser.Node
	for _, node := range class.Body {
		if node != nil && (node.Type == parser.NodeFunctionDef || node.Type == parser.NodeAsyncFunctionDef) {
			methods = append(methods, node)
		}
	}
	return methods
}

// findModuleVariableAccesses finds all module variable accesses in a method
func (d *HiddenDependencyDetector) findModuleVariableAccesses(method *parser.Node, className string) map[string]parser.Location {
	accesses := make(map[string]parser.Location)

	// Collect local variables defined in the method
	localVars := d.collectLocalVariables(method)

	method.Walk(func(node *parser.Node) bool {
		if node.Type == parser.NodeName && node.Name != "" {
			// Skip if it's a local variable
			if localVars[node.Name] {
				return true
			}

			// Skip common built-ins and special names
			if d.isBuiltinOrSpecial(node.Name) {
				return true
			}

			// Skip self/cls
			if node.Name == "self" || node.Name == "cls" || node.Name == className {
				return true
			}

			// Check if it's a module-level variable
			if d.moduleVariables[node.Name] {
				// Only record the first occurrence
				if _, exists := accesses[node.Name]; !exists {
					accesses[node.Name] = node.Location
				}
			}
		}
		return true
	})

	return accesses
}

// collectLocalVariables collects all local variable names in a function
func (d *HiddenDependencyDetector) collectLocalVariables(funcNode *parser.Node) map[string]bool {
	locals := make(map[string]bool)

	// Add parameters
	for _, arg := range funcNode.Args {
		if arg != nil && arg.Type == parser.NodeArg {
			locals[arg.Name] = true
		}
	}

	// Add assigned variables
	funcNode.Walk(func(node *parser.Node) bool {
		if node.Type == parser.NodeAssign {
			for _, target := range node.Targets {
				if target != nil && target.Type == parser.NodeName {
					locals[target.Name] = true
				}
			}
		}
		if node.Type == parser.NodeAnnAssign {
			if node.Name != "" {
				locals[node.Name] = true
			}
		}
		// For loop variables
		if node.Type == parser.NodeFor || node.Type == parser.NodeAsyncFor {
			for _, target := range node.Targets {
				if target != nil && target.Type == parser.NodeName {
					locals[target.Name] = true
				}
			}
		}
		return true
	})

	return locals
}

// isBuiltinOrSpecial checks if a name is a built-in or special name
func (d *HiddenDependencyDetector) isBuiltinOrSpecial(name string) bool {
	builtins := map[string]bool{
		"True": true, "False": true, "None": true,
		"print": true, "len": true, "range": true, "str": true, "int": true, "float": true,
		"list": true, "dict": true, "set": true, "tuple": true, "bool": true,
		"type": true, "super": true, "object": true, "isinstance": true, "issubclass": true,
		"getattr": true, "setattr": true, "hasattr": true, "delattr": true,
		"open": true, "input": true, "map": true, "filter": true, "zip": true,
		"enumerate": true, "sorted": true, "reversed": true, "sum": true, "min": true, "max": true,
		"abs": true, "round": true, "any": true, "all": true, "next": true, "iter": true,
		"Exception": true, "BaseException": true, "ValueError": true, "TypeError": true,
		"KeyError": true, "IndexError": true, "AttributeError": true, "RuntimeError": true,
	}
	return builtins[name]
}
