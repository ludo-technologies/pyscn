package analyzer

import (
	"strings"
	"testing"
)

func containsAll(set []string, elems ...string) bool {
	m := make(map[string]struct{}, len(set))
	for _, s := range set {
		m[s] = struct{}{}
	}
	for _, e := range elems {
		if _, ok := m[e]; !ok {
			return false
		}
	}
	return true
}

func TestDepGraph_SCCAndCycles(t *testing.T) {
	g := NewDepGraph()
	g.AddEdge("A", "B")
	g.AddEdge("B", "C")
	g.AddEdge("C", "A") // cycle A-B-C
	g.AddEdge("D", "E") // acyclic edge

	cycles := g.Cycles()
	if len(cycles) < 1 {
		t.Fatalf("expected at least one cycle, got %d", len(cycles))
	}

	foundABC := false
	for _, c := range cycles {
		if len(c) == 3 && containsAll(c, "A", "B", "C") {
			foundABC = true
		}
	}
	if !foundABC {
		t.Fatalf("expected cycle [A,B,C] in cycles: %#v", cycles)
	}
}

func TestDepGraph_SelfLoopNotCycle(t *testing.T) {
    g := NewDepGraph()
    g.AddEdge("Z", "Z") // self-loop should NOT be treated as a cycle
    cycles := g.Cycles()
    if len(cycles) != 0 {
        t.Fatalf("expected no cycles for self-loop, got %#v", cycles)
    }
}

func TestDepGraph_DOTHighlightsCycles(t *testing.T) {
	g := NewDepGraph()
	g.AddEdge("A", "B")
	g.AddEdge("B", "C")
	g.AddEdge("C", "A") // cycle

	dot := g.ToDOT()
	// Node fill for cycle members
	if !strings.Contains(dot, "\"A\" [style=filled, fillcolor=\"#ffe6e6\"];\n") {
		t.Fatalf("expected A node to be highlighted in DOT, got:\n%s", dot)
	}
	// Edge color red for cycle edges
	if !strings.Contains(dot, "\"A\" -> \"B\" [color=red];") {
		t.Fatalf("expected A->B edge to be red in DOT, got:\n%s", dot)
	}
}
