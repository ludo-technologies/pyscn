package analyzer

import (
	"fmt"

	corecfg "github.com/ludo-technologies/polyscan/core/cfg"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

type pythonCFGClassifier struct{}

var (
	_ corecfg.StatementClassifier = pythonCFGClassifier{}
	_ corecfg.NoOpClassifier      = pythonCFGClassifier{}
)

func pythonNode(value any) (*parser.Node, bool) {
	node, ok := value.(*parser.Node)
	return node, ok && node != nil
}

func mustPythonNode(value any) *parser.Node {
	node, ok := pythonNode(value)
	if !ok {
		panic(fmt.Sprintf("CFG value has type %T, want *parser.Node", value))
	}
	return node
}

func (pythonCFGClassifier) IsReturn(value any) bool {
	node, ok := pythonNode(value)
	return ok && node.Type == parser.NodeReturn
}

func (pythonCFGClassifier) IsBreak(value any) bool {
	node, ok := pythonNode(value)
	return ok && node.Type == parser.NodeBreak
}

func (pythonCFGClassifier) IsContinue(value any) bool {
	node, ok := pythonNode(value)
	return ok && node.Type == parser.NodeContinue
}

func (pythonCFGClassifier) IsThrow(value any) bool {
	node, ok := pythonNode(value)
	return ok && node.Type == parser.NodeRaise
}

func (pythonCFGClassifier) IsNoOp(value any) bool {
	node, ok := pythonNode(value)
	return ok && (node.Type == parser.NodePass || string(node.Type) == ";")
}
