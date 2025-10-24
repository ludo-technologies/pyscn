package service

import (
	"reflect"
	"testing"
)

func TestDetectCycles_NoCycle(test *testing.T) {
	graph := NewGraph()
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "c")

	desiredStart := "a"
	cycles := graph.DetectCyclesWithStart(desiredStart)

	if len(cycles) != 0 {
		test.Errorf("expected no cycles, got %v", cycles)
	}
}

func TestDetectCycles_SimpleCycle(test *testing.T) {
	graph := NewGraph()
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "a") // circle

	desiredStart := "a"
	cycles := graph.DetectCyclesWithStart(desiredStart)
	expected := [][]string{{"a", "b", "a"}}

	if !reflect.DeepEqual(cycles, expected) {
		test.Errorf("expected %v, got %v", expected, cycles)
	}
}

func TestDetectCycles_MultipleCycles(test *testing.T) {
	graph := NewGraph()
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "a") // cycle 1
	graph.AddEdge("c", "d")
	graph.AddEdge("d", "e")
	graph.AddEdge("e", "c") // cycle 2

	desiredStart := "a"
	cycles := graph.DetectCyclesWithStart(desiredStart)

	if len(cycles) != 2 {
		test.Errorf("expected 2 cycles, got %v", cycles)
	}
}

func TestDetectCycles_DefaultStart(t *testing.T) {
	graph := NewGraph()
	graph.AddEdge("a", "b")
	graph.AddEdge("b", "a")

	cycles := graph.DetectCycles()
	if len(cycles) != 1 {
		t.Errorf("expected 1 cycle, got %d", len(cycles))
	}
}

func TestRotateCycleToStart(test *testing.T) {
	cycle := []string{"b", "a", "c", "b"}
	expected := []string{"a", "c", "b", "b"}
	rotated := rotateCycleToStart(cycle, "a")
	if !reflect.DeepEqual(rotated, expected) {
		test.Errorf("expected %v got %v", expected, rotated)
	}

	// Return the original slice when the starting point does not exist
	rotated = rotateCycleToStart(cycle, "x")
	if !reflect.DeepEqual(rotated, cycle) {
		test.Errorf("expected original cycle when start not found")
	}
}
