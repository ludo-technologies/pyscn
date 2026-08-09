package analyzer

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

// ParsedModule is a read-only view of source and syntax captured for module analysis.
// It borrows the source and AST for the duration of analysis; callers retain
// ownership and must not mutate either after construction.
type ParsedModule struct {
	path   string
	source []byte
	ast    *parser.Node
}

// NewParsedModule creates a validated module-analysis input.
func NewParsedModule(path string, source []byte, ast *parser.Node) (ParsedModule, error) {
	if path == "" {
		return ParsedModule{}, fmt.Errorf("parsed module path cannot be empty")
	}
	if ast == nil {
		return ParsedModule{}, fmt.Errorf("parsed module AST cannot be nil")
	}
	return ParsedModule{path: path, source: source, ast: ast}, nil
}
