package analyzer

import (
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

func TestPythonCFGClassifier(t *testing.T) {
	classifier := pythonCFGClassifier{}
	tests := []struct {
		name  string
		node  *parser.Node
		match func(any) bool
	}{
		{name: "return", node: &parser.Node{Type: parser.NodeReturn}, match: classifier.IsReturn},
		{name: "break", node: &parser.Node{Type: parser.NodeBreak}, match: classifier.IsBreak},
		{name: "continue", node: &parser.Node{Type: parser.NodeContinue}, match: classifier.IsContinue},
		{name: "raise", node: &parser.Node{Type: parser.NodeRaise}, match: classifier.IsThrow},
		{name: "pass", node: &parser.Node{Type: parser.NodePass}, match: classifier.IsNoOp},
		{name: "separator", node: &parser.Node{Type: parser.NodeType(";")}, match: classifier.IsNoOp},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !test.match(test.node) {
				t.Fatalf("classifier did not match %s", test.node.Type)
			}
		})
	}
}

func TestCoreReachabilityPreservesExceptionFlowAfterRaise(t *testing.T) {
	graph := NewCFG("raise_flow")
	raiseBlock := graph.CreateBlock("raise")
	normalBlock := graph.CreateBlock("normal_fallthrough")
	handlerBlock := graph.CreateBlock("handler")
	raiseBlock.AddStatement(&parser.Node{Type: parser.NodeRaise})
	normalBlock.AddStatement(&parser.Node{Type: parser.NodeExpr})
	handlerBlock.AddStatement(&parser.Node{Type: parser.NodePass})

	graph.ConnectBlocks(graph.Entry, raiseBlock, EdgeNormal)
	graph.ConnectBlocks(raiseBlock, normalBlock, EdgeNormal)
	graph.ConnectBlocks(raiseBlock, handlerBlock, EdgeException)
	graph.ConnectBlocks(handlerBlock, graph.Exit, EdgeNormal)

	result := NewReachabilityAnalyzer(graph).AnalyzeReachability()
	if _, ok := result.ReachableBlocks[handlerBlock.ID]; !ok {
		t.Fatal("exception handler should remain reachable after raise")
	}
	if _, ok := result.UnreachableBlocks[normalBlock.ID]; !ok {
		t.Fatal("normal fallthrough after raise should be unreachable")
	}
}

func TestDeadCodeSuppressesUnreachablePassBlock(t *testing.T) {
	graph := NewCFG("pass_only")
	passBlock := graph.CreateBlock("unreachable_pass")
	passBlock.AddStatement(&parser.Node{Type: parser.NodePass})
	graph.ConnectBlocks(graph.Entry, graph.Exit, EdgeNormal)

	detector := NewDeadCodeDetector(graph)
	result := detector.Detect()
	if result.DeadBlocks != 0 || len(result.Findings) != 0 {
		t.Fatalf("unreachable pass should not be reported: %+v", result)
	}
	if detector.HasDeadCode() {
		t.Fatal("unreachable pass should not make HasDeadCode true")
	}
}
