package analyzer

import (
	"fmt"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// ServiceLocatorDetector detects service locator anti-pattern
type ServiceLocatorDetector struct {
	locatorMethods []string
}

// NewServiceLocatorDetector creates a new service locator detector
func NewServiceLocatorDetector() *ServiceLocatorDetector {
	return &ServiceLocatorDetector{
		locatorMethods: domain.ServiceLocatorMethodNames(),
	}
}

// Analyze detects service locator pattern in the given AST
func (d *ServiceLocatorDetector) Analyze(ast *parser.Node, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	classes := ast.FindByType(parser.NodeClassDef)

	for _, class := range classes {
		classFindings := d.analyzeClass(class, filePath)
		findings = append(findings, classFindings...)
	}

	return findings
}

// analyzeClass analyzes a single class for service locator pattern
func (d *ServiceLocatorDetector) analyzeClass(classNode *parser.Node, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	// Find all methods in the class
	methods := d.findMethods(classNode)

	for _, method := range methods {
		methodFindings := d.analyzeMethod(method, classNode.Name, filePath)
		findings = append(findings, methodFindings...)
	}

	return findings
}

// analyzeMethod analyzes a method for service locator pattern usage
func (d *ServiceLocatorDetector) analyzeMethod(methodNode *parser.Node, className string, filePath string) []domain.DIAntipatternFinding {
	var findings []domain.DIAntipatternFinding

	// Use custom walk function that includes Value field
	d.walkNode(methodNode, func(node *parser.Node) bool {
		// Check Assign nodes - the Call is in the Value field
		if node.Type == parser.NodeAssign {
			if node.Value != nil {
				if valueNode, ok := node.Value.(*parser.Node); ok {
					if valueNode.Type == parser.NodeCall {
						if locatorInfo := d.isServiceLocatorCall(valueNode); locatorInfo != nil {
							finding := d.createFinding(className, methodNode.Name, filePath, valueNode, locatorInfo)
							findings = append(findings, finding)
						}
					}
				}
			}
		}
		// Check Return nodes - the Call is in the Value field
		if node.Type == parser.NodeReturn {
			if node.Value != nil {
				if valueNode, ok := node.Value.(*parser.Node); ok {
					if valueNode.Type == parser.NodeCall {
						if locatorInfo := d.isServiceLocatorCall(valueNode); locatorInfo != nil {
							finding := d.createFinding(className, methodNode.Name, filePath, valueNode, locatorInfo)
							findings = append(findings, finding)
						}
					}
				}
			}
		}
		return true
	})

	return findings
}

// createFinding creates a DIAntipatternFinding for service locator pattern
func (d *ServiceLocatorDetector) createFinding(className, methodName, filePath string, node *parser.Node, locatorInfo *serviceLocatorInfo) domain.DIAntipatternFinding {
	return domain.DIAntipatternFinding{
		Type:       domain.DIAntipatternServiceLocator,
		Severity:   domain.DIAntipatternSeverityWarning,
		ClassName:  className,
		MethodName: methodName,
		Location: domain.SourceLocation{
			FilePath:  filePath,
			StartLine: node.Location.StartLine,
			EndLine:   node.Location.EndLine,
			StartCol:  node.Location.StartCol,
			EndCol:    node.Location.EndCol,
		},
		Description: fmt.Sprintf("Uses service locator pattern via '%s'", locatorInfo.methodName),
		Suggestion:  "Inject the dependency as a constructor parameter instead of using service locator",
		Details: map[string]interface{}{
			"locator_method": locatorInfo.methodName,
			"container_name": locatorInfo.containerName,
		},
	}
}

// walkNode recursively walks AST nodes including the Value field
func (d *ServiceLocatorDetector) walkNode(node *parser.Node, visitor func(*parser.Node) bool) {
	if node == nil || !visitor(node) {
		return
	}

	for _, child := range node.Children {
		d.walkNode(child, visitor)
	}

	for _, child := range node.Body {
		d.walkNode(child, visitor)
	}

	for _, child := range node.Args {
		d.walkNode(child, visitor)
	}

	// Also traverse Value field if it contains a Node
	if node.Value != nil {
		if valueNode, ok := node.Value.(*parser.Node); ok {
			d.walkNode(valueNode, visitor)
		}
	}
}

// serviceLocatorInfo holds information about a service locator call
type serviceLocatorInfo struct {
	methodName    string
	containerName string
}

// isServiceLocatorCall checks if a call is a service locator pattern
func (d *ServiceLocatorDetector) isServiceLocatorCall(callNode *parser.Node) *serviceLocatorInfo {
	// Check Value field first - tree-sitter stores the callee there
	if callNode.Value != nil {
		if valueNode, ok := callNode.Value.(*parser.Node); ok {
			// Check for attribute-style calls: container.get_service(), locator.resolve()
			if valueNode.Type == parser.NodeAttribute {
				methodName := d.extractMethodName(valueNode)
				containerName := d.extractContainerName(valueNode)

				if d.isLocatorMethodName(methodName) {
					return &serviceLocatorInfo{
						methodName:    methodName,
						containerName: containerName,
					}
				}

				// Check for common container patterns
				fullCall := d.buildFullCallName(valueNode)
				if d.isKnownLocatorPattern(fullCall) {
					return &serviceLocatorInfo{
						methodName:    fullCall,
						containerName: containerName,
					}
				}
			}

			// Check for direct function calls: get_service(), resolve()
			if valueNode.Type == parser.NodeName {
				if d.isLocatorMethodName(valueNode.Name) {
					return &serviceLocatorInfo{
						methodName:    valueNode.Name,
						containerName: "",
					}
				}
			}
		}
	}

	// Fallback: Check Left field
	if callNode.Left != nil && callNode.Left.Type == parser.NodeAttribute {
		methodName := d.extractMethodName(callNode.Left)
		containerName := d.extractContainerName(callNode.Left)

		if d.isLocatorMethodName(methodName) {
			return &serviceLocatorInfo{
				methodName:    methodName,
				containerName: containerName,
			}
		}

		// Check for common container patterns
		fullCall := d.buildFullCallName(callNode.Left)
		if d.isKnownLocatorPattern(fullCall) {
			return &serviceLocatorInfo{
				methodName:    fullCall,
				containerName: containerName,
			}
		}
	}

	// Check for direct function calls via Children
	directName := d.extractDirectCallName(callNode)
	if d.isLocatorMethodName(directName) {
		return &serviceLocatorInfo{
			methodName:    directName,
			containerName: "",
		}
	}

	return nil
}

// extractMethodName extracts the method name from an attribute node
func (d *ServiceLocatorDetector) extractMethodName(attrNode *parser.Node) string {
	// For Attribute nodes, the method name is stored in Name field
	if attrNode.Name != "" {
		return attrNode.Name
	}
	// Fallback: check Right field
	if attrNode.Right != nil && attrNode.Right.Type == parser.NodeName {
		return attrNode.Right.Name
	}
	return ""
}

// extractContainerName extracts the container/object name from an attribute node
func (d *ServiceLocatorDetector) extractContainerName(attrNode *parser.Node) string {
	// For Attribute nodes, the object is stored in Value field
	if attrNode.Value != nil {
		if valueNode, ok := attrNode.Value.(*parser.Node); ok {
			if valueNode.Type == parser.NodeName {
				return valueNode.Name
			}
			// Handle chained attributes: self.container.get
			if valueNode.Type == parser.NodeAttribute {
				return d.extractAttributeChain(valueNode)
			}
		}
	}
	// Fallback: check Left field
	if attrNode.Left != nil {
		if attrNode.Left.Type == parser.NodeName {
			return attrNode.Left.Name
		}
		if attrNode.Left.Type == parser.NodeAttribute {
			return d.extractAttributeChain(attrNode.Left)
		}
	}
	return ""
}

// extractAttributeChain builds a string from a chain of attributes
func (d *ServiceLocatorDetector) extractAttributeChain(attrNode *parser.Node) string {
	if attrNode == nil {
		return ""
	}

	if attrNode.Type == parser.NodeName {
		return attrNode.Name
	}

	if attrNode.Type == parser.NodeAttribute {
		// Get left part from Value field first
		var left string
		if attrNode.Value != nil {
			if valueNode, ok := attrNode.Value.(*parser.Node); ok {
				left = d.extractAttributeChain(valueNode)
			}
		}
		// Fallback to Left field
		if left == "" && attrNode.Left != nil {
			left = d.extractAttributeChain(attrNode.Left)
		}

		// Get right part (method name) from Name field first
		right := attrNode.Name
		if right == "" && attrNode.Right != nil && attrNode.Right.Type == parser.NodeName {
			right = attrNode.Right.Name
		}

		if left != "" && right != "" {
			return left + "." + right
		}
		if left != "" {
			return left
		}
		return right
	}

	return ""
}

// buildFullCallName builds the full method call name (e.g., "container.get")
func (d *ServiceLocatorDetector) buildFullCallName(attrNode *parser.Node) string {
	containerName := d.extractContainerName(attrNode)
	methodName := d.extractMethodName(attrNode)

	if containerName != "" && methodName != "" {
		return containerName + "." + methodName
	}
	return methodName
}

// extractDirectCallName extracts name from a direct function call
func (d *ServiceLocatorDetector) extractDirectCallName(callNode *parser.Node) string {
	// Check children for Name nodes
	for _, child := range callNode.Children {
		if child != nil && child.Type == parser.NodeName && child.Name != "" {
			return child.Name
		}
	}

	// Check Left field
	if callNode.Left != nil && callNode.Left.Type == parser.NodeName {
		return callNode.Left.Name
	}

	return ""
}

// isLocatorMethodName checks if a method name matches service locator patterns
func (d *ServiceLocatorDetector) isLocatorMethodName(name string) bool {
	if name == "" {
		return false
	}

	// Normalize to lowercase for comparison
	nameLower := strings.ToLower(name)

	for _, pattern := range d.locatorMethods {
		patternLower := strings.ToLower(pattern)

		// Exact match
		if nameLower == patternLower {
			return true
		}

		// Handle patterns with dots (e.g., "container.get")
		if strings.Contains(pattern, ".") {
			parts := strings.Split(patternLower, ".")
			if len(parts) == 2 && parts[1] == nameLower {
				return true
			}
		}
	}

	return false
}

// isKnownLocatorPattern checks if a full call matches known locator patterns
func (d *ServiceLocatorDetector) isKnownLocatorPattern(fullCall string) bool {
	if fullCall == "" {
		return false
	}

	fullCallLower := strings.ToLower(fullCall)

	for _, pattern := range d.locatorMethods {
		patternLower := strings.ToLower(pattern)

		// Check for pattern match (e.g., "container.get" matches "container.get")
		if strings.HasSuffix(fullCallLower, patternLower) {
			return true
		}
	}

	// Additional common patterns
	commonPatterns := []string{
		"container.get",
		"container.resolve",
		"locator.get",
		"locator.resolve",
		"registry.get",
		"ioc.resolve",
		"injector.get",
		"services.get",
	}

	for _, pattern := range commonPatterns {
		if strings.HasSuffix(fullCallLower, pattern) {
			return true
		}
	}

	return false
}

// findMethods finds all methods in a class
func (d *ServiceLocatorDetector) findMethods(classNode *parser.Node) []*parser.Node {
	var methods []*parser.Node
	for _, node := range classNode.Body {
		if node != nil && (node.Type == parser.NodeFunctionDef || node.Type == parser.NodeAsyncFunctionDef) {
			methods = append(methods, node)
		}
	}
	return methods
}
