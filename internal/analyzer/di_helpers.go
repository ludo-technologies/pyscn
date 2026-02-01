package analyzer

import "github.com/ludo-technologies/pyscn/internal/parser"

// FindClassMethods finds all methods (functions) defined in a class body.
// This is a shared helper for DI anti-pattern detectors.
func FindClassMethods(classNode *parser.Node) []*parser.Node {
	var methods []*parser.Node
	for _, node := range classNode.Body {
		if node != nil && (node.Type == parser.NodeFunctionDef || node.Type == parser.NodeAsyncFunctionDef) {
			methods = append(methods, node)
		}
	}
	return methods
}

// FindInitMethod finds the __init__ method in a class body.
// Returns nil if no __init__ method is found.
// This is a shared helper for DI anti-pattern detectors.
func FindInitMethod(classNode *parser.Node) *parser.Node {
	for _, node := range classNode.Body {
		if node != nil && (node.Type == parser.NodeFunctionDef || node.Type == parser.NodeAsyncFunctionDef) {
			if node.Name == "__init__" {
				return node
			}
		}
	}
	return nil
}
