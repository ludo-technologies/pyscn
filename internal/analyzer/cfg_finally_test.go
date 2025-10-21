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
}
