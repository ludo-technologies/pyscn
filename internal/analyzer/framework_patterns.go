package analyzer

import (
	"strings"
)

// IsBoilerplateLabel checks if a tree node label represents boilerplate code.
// Boilerplate includes type annotations, decorators, and type hint related nodes.
// This is the single source of truth for boilerplate detection, used by both
// the cost model and any other components that need to identify boilerplate.
func IsBoilerplateLabel(label string) bool {
	// Type annotations (AnnAssign nodes)
	if strings.HasPrefix(label, "AnnAssign") {
		return true
	}

	// Decorator nodes
	if strings.HasPrefix(label, "Decorator") {
		return true
	}

	// Type hint related nodes
	labelLower := strings.ToLower(label)
	typeHintPatterns := []string{
		"generic_type",
		"type_parameter",
	}
	for _, pattern := range typeHintPatterns {
		if strings.Contains(labelLower, pattern) {
			return true
		}
	}

	// Check for common field factory patterns in the label
	// These appear in dataclasses, Pydantic, and attrs
	fieldPatterns := []string{
		"Field(", "field(", "Factory(", "attrib(", "attr.ib(",
	}
	for _, pattern := range fieldPatterns {
		if strings.Contains(label, pattern) {
			return true
		}
	}

	return false
}
