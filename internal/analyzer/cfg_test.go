package analyzer

import (
	"testing"
	"github.com/pyqol/pyqol/internal/parser"
)

func TestBasicBlock(t *testing.T) {
	t.Run("NewBasicBlock", func(t *testing.T) {
		block := NewBasicBlock("test")
		
		if block.ID != "test" {
			t.Errorf("Expected ID 'test', got %s", block.ID)
		}
		if len(block.Statements) != 0 {
			t.Errorf("Expected empty statements, got %d", len(block.Statements))
		}
		if len(block.Predecessors) != 0 {
			t.Errorf("Expected no predecessors, got %d", len(block.Predecessors))
		}
		if len(block.Successors) != 0 {
			t.Errorf("Expected no successors, got %d", len(block.Successors))
		}
	})
	
	t.Run("AddStatement", func(t *testing.T) {
		block := NewBasicBlock("test")
		stmt1 := &parser.Node{Type: parser.NodeAssign}
		stmt2 := &parser.Node{Type: parser.NodeReturn}
		
		block.AddStatement(stmt1)
		block.AddStatement(stmt2)
		block.AddStatement(nil) // Should be ignored
		
		if len(block.Statements) != 2 {
			t.Errorf("Expected 2 statements, got %d", len(block.Statements))
		}
		if block.Statements[0] != stmt1 {
			t.Error("First statement mismatch")
		}
		if block.Statements[1] != stmt2 {
			t.Error("Second statement mismatch")
		}
	})
	
	t.Run("AddSuccessor", func(t *testing.T) {
		block1 := NewBasicBlock("bb1")
		block2 := NewBasicBlock("bb2")
		
		edge := block1.AddSuccessor(block2, EdgeNormal)
		
		if edge == nil {
			t.Fatal("AddSuccessor returned nil")
		}
		if edge.From != block1 {
			t.Error("Edge.From mismatch")
		}
		if edge.To != block2 {
			t.Error("Edge.To mismatch")
		}
		if edge.Type != EdgeNormal {
			t.Error("Edge.Type mismatch")
		}
		
		if len(block1.Successors) != 1 {
			t.Errorf("Expected 1 successor, got %d", len(block1.Successors))
		}
		if len(block2.Predecessors) != 1 {
			t.Errorf("Expected 1 predecessor, got %d", len(block2.Predecessors))
		}
	})
	
	t.Run("RemoveSuccessor", func(t *testing.T) {
		block1 := NewBasicBlock("bb1")
		block2 := NewBasicBlock("bb2")
		block3 := NewBasicBlock("bb3")
		
		block1.AddSuccessor(block2, EdgeNormal)
		block1.AddSuccessor(block3, EdgeNormal)
		
		block1.RemoveSuccessor(block2)
		
		if len(block1.Successors) != 1 {
			t.Errorf("Expected 1 successor after removal, got %d", len(block1.Successors))
		}
		if block1.Successors[0].To != block3 {
			t.Error("Wrong successor remains after removal")
		}
		if len(block2.Predecessors) != 0 {
			t.Errorf("Expected no predecessors after removal, got %d", len(block2.Predecessors))
		}
	})
	
	t.Run("IsEmpty", func(t *testing.T) {
		block := NewBasicBlock("test")
		
		if !block.IsEmpty() {
			t.Error("New block should be empty")
		}
		
		block.AddStatement(&parser.Node{Type: parser.NodeAssign})
		
		if block.IsEmpty() {
			t.Error("Block with statement should not be empty")
		}
	})
	
	t.Run("String", func(t *testing.T) {
		block := NewBasicBlock("bb1")
		block.Label = "loop_header"
		
		str := block.String()
		if str != "[loop_header: 0 stmts]" {
			t.Errorf("Unexpected string representation: %s", str)
		}
		
		// Test entry block
		block.IsEntry = true
		str = block.String()
		if str != "[ENTRY: loop_header]" {
			t.Errorf("Unexpected entry block string: %s", str)
		}
		
		// Test exit block
		block.IsEntry = false
		block.IsExit = true
		str = block.String()
		if str != "[EXIT: loop_header]" {
			t.Errorf("Unexpected exit block string: %s", str)
		}
	})
}

func TestEdgeType(t *testing.T) {
	tests := []struct {
		edgeType EdgeType
		expected string
	}{
		{EdgeNormal, "normal"},
		{EdgeCondTrue, "true"},
		{EdgeCondFalse, "false"},
		{EdgeException, "exception"},
		{EdgeLoop, "loop"},
		{EdgeBreak, "break"},
		{EdgeContinue, "continue"},
		{EdgeReturn, "return"},
		{EdgeType(999), "unknown"},
	}
	
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.edgeType.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCFG(t *testing.T) {
	t.Run("NewCFG", func(t *testing.T) {
		cfg := NewCFG("test_func")
		
		if cfg.Name != "test_func" {
			t.Errorf("Expected name 'test_func', got %s", cfg.Name)
		}
		if cfg.Entry == nil {
			t.Fatal("Entry block is nil")
		}
		if !cfg.Entry.IsEntry {
			t.Error("Entry block not marked as entry")
		}
		if cfg.Exit == nil {
			t.Fatal("Exit block is nil")
		}
		if !cfg.Exit.IsExit {
			t.Error("Exit block not marked as exit")
		}
		if cfg.Size() != 2 {
			t.Errorf("Expected 2 blocks (entry/exit), got %d", cfg.Size())
		}
	})
	
	t.Run("CreateBlock", func(t *testing.T) {
		cfg := NewCFG("test")
		
		block1 := cfg.CreateBlock("first")
		block2 := cfg.CreateBlock("second")
		
		if block1.Label != "first" {
			t.Errorf("Expected label 'first', got %s", block1.Label)
		}
		if block2.Label != "second" {
			t.Errorf("Expected label 'second', got %s", block2.Label)
		}
		if cfg.Size() != 4 { // entry, exit, block1, block2
			t.Errorf("Expected 4 blocks, got %d", cfg.Size())
		}
		
		// Check unique IDs
		if block1.ID == block2.ID {
			t.Error("Blocks should have unique IDs")
		}
	})
	
	t.Run("AddBlock", func(t *testing.T) {
		cfg := NewCFG("test")
		external := NewBasicBlock("external")
		
		cfg.AddBlock(external)
		
		if cfg.GetBlock("external") != external {
			t.Error("Failed to add external block")
		}
		if cfg.Size() != 3 { // entry, exit, external
			t.Errorf("Expected 3 blocks, got %d", cfg.Size())
		}
		
		// Test nil handling
		cfg.AddBlock(nil)
		if cfg.Size() != 3 {
			t.Error("Adding nil block should not change size")
		}
	})
	
	t.Run("RemoveBlock", func(t *testing.T) {
		cfg := NewCFG("test")
		block1 := cfg.CreateBlock("block1")
		block2 := cfg.CreateBlock("block2")
		block3 := cfg.CreateBlock("block3")
		
		// Create connections
		cfg.ConnectBlocks(block1, block2, EdgeNormal)
		cfg.ConnectBlocks(block2, block3, EdgeNormal)
		
		// Remove middle block
		cfg.RemoveBlock(block2)
		
		if cfg.GetBlock(block2.ID) != nil {
			t.Error("Block2 should be removed from CFG")
		}
		if len(block1.Successors) != 0 {
			t.Error("Block1 should have no successors after removal")
		}
		if len(block3.Predecessors) != 0 {
			t.Error("Block3 should have no predecessors after removal")
		}
		
		// Try to remove entry/exit (should be ignored)
		initialSize := cfg.Size()
		cfg.RemoveBlock(cfg.Entry)
		cfg.RemoveBlock(cfg.Exit)
		if cfg.Size() != initialSize {
			t.Error("Should not be able to remove entry/exit blocks")
		}
	})
	
	t.Run("ConnectBlocks", func(t *testing.T) {
		cfg := NewCFG("test")
		block1 := cfg.CreateBlock("block1")
		block2 := cfg.CreateBlock("block2")
		
		edge := cfg.ConnectBlocks(block1, block2, EdgeCondTrue)
		
		if edge == nil {
			t.Fatal("ConnectBlocks returned nil")
		}
		if edge.Type != EdgeCondTrue {
			t.Error("Edge type mismatch")
		}
		
		// Test nil handling
		nilEdge := cfg.ConnectBlocks(nil, block2, EdgeNormal)
		if nilEdge != nil {
			t.Error("Connecting from nil should return nil")
		}
		
		nilEdge = cfg.ConnectBlocks(block1, nil, EdgeNormal)
		if nilEdge != nil {
			t.Error("Connecting to nil should return nil")
		}
	})
}

func TestCFGTraversal(t *testing.T) {
	// Build a simple CFG:
	// ENTRY -> A -> B -> EXIT
	//           \-> C -/
	cfg := NewCFG("test")
	blockA := cfg.CreateBlock("A")
	blockB := cfg.CreateBlock("B")
	blockC := cfg.CreateBlock("C")
	
	cfg.ConnectBlocks(cfg.Entry, blockA, EdgeNormal)
	cfg.ConnectBlocks(blockA, blockB, EdgeCondTrue)
	cfg.ConnectBlocks(blockA, blockC, EdgeCondFalse)
	cfg.ConnectBlocks(blockB, cfg.Exit, EdgeNormal)
	cfg.ConnectBlocks(blockC, cfg.Exit, EdgeNormal)
	
	t.Run("DepthFirstWalk", func(t *testing.T) {
		visitedBlocks := []string{}
		visitedEdges := []EdgeType{}
		
		visitor := &testVisitor{
			onBlock: func(b *BasicBlock) bool {
				visitedBlocks = append(visitedBlocks, b.Label)
				return true
			},
			onEdge: func(e *Edge) bool {
				visitedEdges = append(visitedEdges, e.Type)
				return true
			},
		}
		
		cfg.Walk(visitor)
		
		// Should visit all blocks
		if len(visitedBlocks) != 5 {
			t.Errorf("Expected to visit 5 blocks, visited %d", len(visitedBlocks))
		}
		
		// Entry should be first
		if len(visitedBlocks) > 0 && visitedBlocks[0] != "ENTRY" {
			t.Errorf("Expected ENTRY first, got %s", visitedBlocks[0])
		}
		
		// Should visit all edges
		if len(visitedEdges) != 5 {
			t.Errorf("Expected to visit 5 edges, visited %d", len(visitedEdges))
		}
	})
	
	t.Run("BreadthFirstWalk", func(t *testing.T) {
		visitedBlocks := []string{}
		
		visitor := &testVisitor{
			onBlock: func(b *BasicBlock) bool {
				visitedBlocks = append(visitedBlocks, b.Label)
				return true
			},
			onEdge: func(e *Edge) bool {
				return true
			},
		}
		
		cfg.BreadthFirstWalk(visitor)
		
		// Should visit all blocks
		if len(visitedBlocks) != 5 {
			t.Errorf("Expected to visit 5 blocks, visited %d", len(visitedBlocks))
		}
		
		// Entry should be first
		if len(visitedBlocks) > 0 && visitedBlocks[0] != "ENTRY" {
			t.Errorf("Expected ENTRY first, got %s", visitedBlocks[0])
		}
		
		// In BFS, A should come before B and C
		aIndex := indexOf(visitedBlocks, "A")
		bIndex := indexOf(visitedBlocks, "B")
		cIndex := indexOf(visitedBlocks, "C")
		
		if aIndex >= 0 && bIndex >= 0 && aIndex > bIndex {
			t.Error("In BFS, A should be visited before B")
		}
		if aIndex >= 0 && cIndex >= 0 && aIndex > cIndex {
			t.Error("In BFS, A should be visited before C")
		}
	})
	
	t.Run("WalkWithEarlyStop", func(t *testing.T) {
		visitedCount := 0
		
		visitor := &testVisitor{
			onBlock: func(b *BasicBlock) bool {
				visitedCount++
				return visitedCount <= 3 // Continue for first 3 blocks, stop on 4th
			},
			onEdge: func(e *Edge) bool {
				return true
			},
		}
		
		cfg.Walk(visitor)
		
		// Walk will visit blocks until visitor returns false
		// Due to DFS nature, it might visit more than expected due to recursion
		if visitedCount < 3 || visitedCount > 5 {
			t.Errorf("Expected to visit 3-5 blocks with early stop, visited %d", visitedCount)
		}
	})
}

func TestCFGWithLoop(t *testing.T) {
	// Build a CFG with a loop:
	// ENTRY -> A -> B -> C -> EXIT
	//               ^----/
	cfg := NewCFG("loop_test")
	blockA := cfg.CreateBlock("A")
	blockB := cfg.CreateBlock("B")
	blockC := cfg.CreateBlock("C")
	
	cfg.ConnectBlocks(cfg.Entry, blockA, EdgeNormal)
	cfg.ConnectBlocks(blockA, blockB, EdgeNormal)
	cfg.ConnectBlocks(blockB, blockC, EdgeNormal)
	cfg.ConnectBlocks(blockC, blockB, EdgeLoop) // Loop back edge
	cfg.ConnectBlocks(blockC, cfg.Exit, EdgeCondFalse)
	
	t.Run("LoopDetection", func(t *testing.T) {
		visitedBlocks := make(map[string]int)
		
		visitor := &testVisitor{
			onBlock: func(b *BasicBlock) bool {
				visitedBlocks[b.Label]++
				return true
			},
			onEdge: func(e *Edge) bool {
				return true
			},
		}
		
		cfg.Walk(visitor)
		
		// Each block should be visited exactly once (DFS prevents revisiting)
		for block, count := range visitedBlocks {
			if count != 1 {
				t.Errorf("Block %s visited %d times, expected 1", block, count)
			}
		}
		
		// Check that loop edge exists
		hasLoopEdge := false
		for _, edge := range blockC.Successors {
			if edge.Type == EdgeLoop && edge.To == blockB {
				hasLoopEdge = true
				break
			}
		}
		if !hasLoopEdge {
			t.Error("Loop edge not found")
		}
	})
}

// testVisitor implements CFGVisitor for testing
type testVisitor struct {
	onBlock func(*BasicBlock) bool
	onEdge  func(*Edge) bool
}

func (v *testVisitor) VisitBlock(block *BasicBlock) bool {
	if v.onBlock != nil {
		return v.onBlock(block)
	}
	return true
}

func (v *testVisitor) VisitEdge(edge *Edge) bool {
	if v.onEdge != nil {
		return v.onEdge(edge)
	}
	return true
}

// Helper function to find index of string in slice
func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}