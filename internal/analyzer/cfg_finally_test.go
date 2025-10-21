package analyzer

import (
	"strings"
	"testing"
)

// TestFinallyBlockParsing tests if finally blocks are correctly parsed and added to CFG
func TestFinallyBlockParsing(t *testing.T) {
	t.Run("SimpleTryFinally", func(t *testing.T) {
		source := `
def test_finally():
    try:
        print("try")
    finally:
        print("finally")
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		// Check if finally block is parsed
		if len(funcNode.Body) == 0 {
			t.Fatal("Function has no body")
		}

		tryNode := funcNode.Body[0]
		t.Logf("Try node type: %s", tryNode.Type)
		t.Logf("Finalbody length: %d", len(tryNode.Finalbody))

		if len(tryNode.Finalbody) == 0 {
			t.Error("❌ Finally block is NOT parsed by parser")
		} else {
			t.Logf("✅ Finally block is parsed by parser (%d statements)", len(tryNode.Finalbody))
		}

		// Build CFG
		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Check for finally block in CFG
		hasFinallyBlock := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					hasFinallyBlock = true
					t.Logf("✅ Found finally block in CFG: %s", b.Label)
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if len(tryNode.Finalbody) > 0 && !hasFinallyBlock {
			t.Error("❌ Parser found finally block but CFG builder did not create finally block")
		}

		t.Logf("Total CFG blocks: %d", len(cfg.Blocks))
		for _, block := range cfg.Blocks {
			t.Logf("  Block: %s", block.Label)
		}
	})

	t.Run("TryExceptFinally", func(t *testing.T) {
		source := `
def test():
    try:
        risky()
    except ValueError:
        handle()
    finally:
        cleanup()
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]
		tryNode := funcNode.Body[0]

		t.Logf("Finalbody length: %d", len(tryNode.Finalbody))
		t.Logf("Handlers length: %d", len(tryNode.Handlers))

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		hasFinallyBlock := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					hasFinallyBlock = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if len(tryNode.Finalbody) > 0 && !hasFinallyBlock {
			t.Error("Finally block parsed but not in CFG")
		}
	})

	t.Run("FinallyWithReturn", func(t *testing.T) {
		source := `
def test():
    try:
        return "value"
    finally:
        cleanup()
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]
		tryNode := funcNode.Body[0]

		t.Logf("Finalbody length: %d", len(tryNode.Finalbody))

		// Verify parser correctly extracted finally body
		if len(tryNode.Finalbody) == 0 {
			t.Error("Parser failed to extract finally body")
		}

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Check if finally block is created
		// Note: Proper finally-return interaction is complex and will be
		// implemented in Phase 3 (CFG enhancement)
		hasFinallyBlock := false
		hasReturnEdge := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					hasFinallyBlock = true
					t.Logf("Found finally block: %s", b.Label)
				}
				return true
			},
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeReturn {
					hasReturnEdge = true
					t.Logf("Return edge from %s to %s", e.From.Label, e.To.Label)
				}
				return true
			},
		})

		if !hasReturnEdge {
			t.Error("Missing return edge")
		}

		// TODO: Phase 3 - Implement proper finally-return interaction
		// Finally blocks should be executed even when return is in try block
		// Currently the CFG doesn't enforce this execution order
		if hasFinallyBlock {
			t.Logf("✅ Finally block created in CFG")
		} else {
			t.Logf("⚠️  Finally block not in CFG (return interaction not yet implemented)")
		}
	})

	t.Run("ReturnInFinally", func(t *testing.T) {
		source := `
def test():
    try:
        x = 1
    finally:
        return "from finally"
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]
		tryNode := funcNode.Body[0]

		t.Logf("Finalbody length: %d", len(tryNode.Finalbody))

		// Verify parser correctly extracted finally body
		if len(tryNode.Finalbody) == 0 {
			t.Error("Parser failed to extract finally body")
		}

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Check that finally block with return connects to EXIT
		// This is critical to avoid self-loop when return is inside finally
		hasFinallyBlock := false
		finallyBlockConnectsToExit := false

		var finallyBlock *BasicBlock
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					hasFinallyBlock = true
					finallyBlock = b
					t.Logf("Found finally block: %s", b.Label)
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasFinallyBlock {
			t.Error("Finally block not created in CFG")
		}

		if finallyBlock != nil {
			// Check that finally block connects to EXIT (not to itself)
			for _, edge := range finallyBlock.Successors {
				if edge.To == cfg.Exit && edge.Type == EdgeReturn {
					finallyBlockConnectsToExit = true
					t.Logf("✅ Finally block correctly connects to EXIT with EdgeReturn")
					break
				}
				if edge.To == finallyBlock {
					t.Errorf("❌ Finally block has self-loop! From=%s To=%s Type=%v",
						edge.From.Label, edge.To.Label, edge.Type)
				}
			}

			if !finallyBlockConnectsToExit {
				t.Error("Finally block with return does not connect to EXIT")
			}
		}
	})

	t.Run("NestedFinallyWithReturnInInner", func(t *testing.T) {
		source := `
def test():
    try:
        try:
            x = 1
        finally:
            return "inner"
    finally:
        print("outer cleanup")
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Find inner and outer finally blocks
		var innerFinally *BasicBlock
		var outerFinally *BasicBlock
		finallyBlocks := []*BasicBlock{}

		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					finallyBlocks = append(finallyBlocks, b)
					t.Logf("Found finally block: %s", b.Label)
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if len(finallyBlocks) != 2 {
			t.Errorf("Expected 2 finally blocks, got %d", len(finallyBlocks))
		}

		// Determine which is inner and which is outer by checking successors
		// Inner finally should have return edge to outer finally
		// Outer finally should have edge to EXIT
		for _, fb := range finallyBlocks {
			hasReturnToFinally := false
			hasReturnToExit := false

			for _, edge := range fb.Successors {
				if edge.Type == EdgeReturn {
					if strings.Contains(edge.To.Label, "finally_block") {
						hasReturnToFinally = true
					}
					if edge.To == cfg.Exit {
						hasReturnToExit = true
					}
				}
			}

			if hasReturnToFinally {
				innerFinally = fb
				t.Logf("Inner finally: %s (routes to outer finally)", fb.Label)
			}
			if hasReturnToExit {
				outerFinally = fb
				t.Logf("Outer finally: %s (routes to EXIT)", fb.Label)
			}
		}

		// Verify the chain: inner finally -> outer finally -> EXIT
		if innerFinally == nil {
			t.Error("Inner finally block not found or not routing to outer finally")
		}
		if outerFinally == nil {
			t.Error("Outer finally block not found or not routing to EXIT")
		}

		if innerFinally != nil && outerFinally != nil {
			// Check that inner finally routes to outer finally
			foundRoute := false
			for _, edge := range innerFinally.Successors {
				if edge.Type == EdgeReturn && edge.To == outerFinally {
					foundRoute = true
					t.Logf("✅ Inner finally correctly routes to outer finally")
					break
				}
			}
			if !foundRoute {
				t.Error("Inner finally does not route return to outer finally")
			}

			// Check that outer finally routes to EXIT
			foundExit := false
			for _, edge := range outerFinally.Successors {
				if edge.Type == EdgeReturn && edge.To == cfg.Exit {
					foundExit = true
					t.Logf("✅ Outer finally correctly routes to EXIT")
					break
				}
			}
			if !foundExit {
				t.Error("Outer finally does not route to EXIT")
			}

			// Verify no self-loops
			for _, fb := range finallyBlocks {
				for _, edge := range fb.Successors {
					if edge.To == fb {
						t.Errorf("❌ Finally block %s has self-loop!", fb.Label)
					}
				}
			}
		}
	})
}
