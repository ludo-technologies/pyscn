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

// TestBreakContinueRaiseWithFinally tests that break/continue/raise statements
// correctly route through finally blocks before reaching their targets.
// This is required by Python semantics - finally blocks always execute.
func TestBreakContinueRaiseWithFinally(t *testing.T) {
	t.Run("BreakWithFinally", func(t *testing.T) {
		source := `
def test():
    for i in range(5):
        try:
            if i == 2:
                break
        finally:
            print("cleanup")
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Find the block containing break statement
		var breakBlock *BasicBlock
		var finallyBlock *BasicBlock
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					finallyBlock = b
				}
				for _, stmt := range b.Statements {
					if stmt.Type == "Break" {
						breakBlock = b
					}
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if breakBlock == nil {
			t.Fatal("Break block not found")
		}
		if finallyBlock == nil {
			t.Fatal("Finally block not found")
		}

		// Check that break routes to finally block, not directly to loop exit
		foundBreakToFinally := false
		for _, edge := range breakBlock.Successors {
			if edge.Type == EdgeBreak && edge.To == finallyBlock {
				foundBreakToFinally = true
				t.Logf("Break correctly routes to finally block")
			}
		}

		if !foundBreakToFinally {
			t.Error("Break does not route through finally block")
			for _, edge := range breakBlock.Successors {
				t.Logf("  Break edge: %s -> %s (type: %v)", breakBlock.Label, edge.To.Label, edge.Type)
			}
		}
	})

	t.Run("ContinueWithFinally", func(t *testing.T) {
		source := `
def test():
    for i in range(5):
        try:
            if i == 2:
                continue
        finally:
            print("cleanup")
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		var continueBlock *BasicBlock
		var finallyBlock *BasicBlock
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					finallyBlock = b
				}
				for _, stmt := range b.Statements {
					if stmt.Type == "Continue" {
						continueBlock = b
					}
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if continueBlock == nil {
			t.Fatal("Continue block not found")
		}
		if finallyBlock == nil {
			t.Fatal("Finally block not found")
		}

		foundContinueToFinally := false
		for _, edge := range continueBlock.Successors {
			if edge.Type == EdgeContinue && edge.To == finallyBlock {
				foundContinueToFinally = true
				t.Logf("Continue correctly routes to finally block")
			}
		}

		if !foundContinueToFinally {
			t.Error("Continue does not route through finally block")
			for _, edge := range continueBlock.Successors {
				t.Logf("  Continue edge: %s -> %s (type: %v)", continueBlock.Label, edge.To.Label, edge.Type)
			}
		}
	})

	t.Run("RaiseWithFinally", func(t *testing.T) {
		source := `
def test():
    try:
        raise ValueError("error")
    finally:
        print("cleanup")
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		var raiseBlock *BasicBlock
		var finallyBlock *BasicBlock
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					finallyBlock = b
				}
				for _, stmt := range b.Statements {
					if stmt.Type == "Raise" {
						raiseBlock = b
					}
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if raiseBlock == nil {
			t.Fatal("Raise block not found")
		}
		if finallyBlock == nil {
			t.Fatal("Finally block not found")
		}

		foundRaiseToFinally := false
		for _, edge := range raiseBlock.Successors {
			if edge.Type == EdgeException && edge.To == finallyBlock {
				foundRaiseToFinally = true
				t.Logf("Raise correctly routes to finally block")
			}
		}

		if !foundRaiseToFinally {
			t.Error("Raise does not route through finally block")
			for _, edge := range raiseBlock.Successors {
				t.Logf("  Raise edge: %s -> %s (type: %v)", raiseBlock.Label, edge.To.Label, edge.Type)
			}
		}
	})

	t.Run("NestedFinallyWithBreak", func(t *testing.T) {
		// Tests the expected CFG: break → inner_finally → outer_finally → loop_exit
		source := `
def test():
    for i in range(5):
        try:
            try:
                if i == 2:
                    break
            finally:
                print("inner finally")
        finally:
            print("outer finally")
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		var breakBlock *BasicBlock
		finallyBlocks := []*BasicBlock{}

		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					finallyBlocks = append(finallyBlocks, b)
				}
				for _, stmt := range b.Statements {
					if stmt.Type == "Break" {
						breakBlock = b
					}
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if breakBlock == nil {
			t.Fatal("Break block not found")
		}
		if len(finallyBlocks) != 2 {
			t.Fatalf("Expected 2 finally blocks, got %d", len(finallyBlocks))
		}

		// Break should route to the innermost finally
		foundBreakToFinally := false
		var innerFinally *BasicBlock
		for _, edge := range breakBlock.Successors {
			if edge.Type == EdgeBreak {
				for _, fb := range finallyBlocks {
					if edge.To == fb {
						foundBreakToFinally = true
						innerFinally = fb
						t.Logf("Break routes to finally block: %s", fb.Label)
						break
					}
				}
			}
		}

		if !foundBreakToFinally {
			t.Error("Break does not route to any finally block")
		}

		// Inner finally should route to outer finally or loop exit via any edge type
		// (CFG only tracks reachability, not specific edge types for propagation)
		if innerFinally != nil {
			var outerFinally *BasicBlock
			for _, fb := range finallyBlocks {
				if fb != innerFinally {
					outerFinally = fb
					break
				}
			}

			hasConnectionToOuter := false
			for _, edge := range innerFinally.Successors {
				if edge.To == outerFinally {
					hasConnectionToOuter = true
					t.Logf("Inner finally connects to outer finally via %v edge", edge.Type)
				}
			}
			if !hasConnectionToOuter && outerFinally != nil {
				t.Error("Inner finally does not connect to outer finally")
			}

			// Check outer finally connects to loop exit
			if outerFinally != nil {
				hasLoopExitConnection := false
				for _, edge := range outerFinally.Successors {
					if strings.Contains(edge.To.Label, "for_exit") || strings.Contains(edge.To.Label, "loop_exit") {
						hasLoopExitConnection = true
						t.Logf("Outer finally connects to loop exit via %v edge", edge.Type)
					}
				}
				if !hasLoopExitConnection {
					t.Logf("Outer finally successors:")
					for _, edge := range outerFinally.Successors {
						t.Logf("  -> %s (type: %v)", edge.To.Label, edge.Type)
					}
				}
			}
		}
	})

	t.Run("BreakInsideFinally", func(t *testing.T) {
		// BUG REPRODUCTION: break inside finally should route to loop exit,
		// NOT back to the finally block itself (which would create a loop)
		source := `
def test():
    for i in range(5):
        try:
            pass
        finally:
            if i == 2:
                break  # This should go to loop exit, NOT to finally_block
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		var breakBlock *BasicBlock
		var finallyBlock *BasicBlock
		var loopExitBlock *BasicBlock

		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					finallyBlock = b
				}
				if strings.Contains(b.Label, "for_exit") || strings.Contains(b.Label, "loop_exit") {
					loopExitBlock = b
				}
				for _, stmt := range b.Statements {
					if stmt.Type == "Break" {
						breakBlock = b
					}
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if breakBlock == nil {
			t.Fatal("Break block not found")
		}
		if finallyBlock == nil {
			t.Fatal("Finally block not found")
		}
		if loopExitBlock == nil {
			t.Fatal("Loop exit block not found")
		}

		// Check that break does NOT route to its own finally block (bug)
		for _, edge := range breakBlock.Successors {
			if edge.Type == EdgeBreak && edge.To == finallyBlock {
				t.Errorf("BUG: Break inside finally incorrectly routes back to finally block (creates loop)")
				t.Logf("  Break block: %s", breakBlock.Label)
				t.Logf("  Incorrectly routes to: %s", finallyBlock.Label)
			}
		}

		// Check that break routes to loop exit (correct behavior)
		foundBreakToLoopExit := false
		for _, edge := range breakBlock.Successors {
			if edge.Type == EdgeBreak && edge.To == loopExitBlock {
				foundBreakToLoopExit = true
				t.Logf("Break correctly routes to loop exit")
			}
		}

		if !foundBreakToLoopExit {
			t.Error("Break inside finally should route directly to loop exit")
			t.Logf("Break block successors:")
			for _, edge := range breakBlock.Successors {
				t.Logf("  -> %s (type: %v)", edge.To.Label, edge.Type)
			}
		}
	})

	t.Run("ContinueInsideFinally", func(t *testing.T) {
		// continue inside finally should route to loop header, NOT back to finally
		source := `
def test():
    for i in range(5):
        try:
            pass
        finally:
            if i == 2:
                continue  # This should go to loop header, NOT to finally_block
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		var continueBlock *BasicBlock
		var finallyBlock *BasicBlock
		var loopHeaderBlock *BasicBlock

		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					finallyBlock = b
				}
				if strings.Contains(b.Label, "loop_header") {
					loopHeaderBlock = b
				}
				for _, stmt := range b.Statements {
					if stmt.Type == "Continue" {
						continueBlock = b
					}
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if continueBlock == nil {
			t.Fatal("Continue block not found")
		}
		if finallyBlock == nil {
			t.Fatal("Finally block not found")
		}
		if loopHeaderBlock == nil {
			t.Fatal("Loop header block not found")
		}

		// Check that continue does NOT route to its own finally block (bug)
		for _, edge := range continueBlock.Successors {
			if edge.Type == EdgeContinue && edge.To == finallyBlock {
				t.Errorf("BUG: Continue inside finally incorrectly routes back to finally block")
			}
		}

		// Check that continue routes to loop header (correct behavior)
		foundContinueToLoopHeader := false
		for _, edge := range continueBlock.Successors {
			if edge.Type == EdgeContinue && edge.To == loopHeaderBlock {
				foundContinueToLoopHeader = true
				t.Logf("Continue correctly routes to loop header")
			}
		}

		if !foundContinueToLoopHeader {
			t.Error("Continue inside finally should route directly to loop header")
		}
	})

	t.Run("RaiseInsideFinally", func(t *testing.T) {
		// raise inside finally should propagate to exit, NOT back to finally
		source := `
def test():
    try:
        pass
    finally:
        if True:
            raise ValueError("error")  # Should go to exit, NOT to finally_block
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		var raiseBlock *BasicBlock
		var finallyBlock *BasicBlock

		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					finallyBlock = b
				}
				for _, stmt := range b.Statements {
					if stmt.Type == "Raise" {
						raiseBlock = b
					}
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if raiseBlock == nil {
			t.Fatal("Raise block not found")
		}
		if finallyBlock == nil {
			t.Fatal("Finally block not found")
		}

		// Check that raise does NOT route to its own finally block (bug)
		for _, edge := range raiseBlock.Successors {
			if edge.Type == EdgeException && edge.To == finallyBlock {
				t.Errorf("BUG: Raise inside finally incorrectly routes back to finally block")
			}
		}

		// Check that raise routes to exit (correct behavior for unhandled exception)
		foundRaiseToExit := false
		for _, edge := range raiseBlock.Successors {
			if edge.Type == EdgeException && edge.To == cfg.Exit {
				foundRaiseToExit = true
				t.Logf("Raise correctly routes to exit")
			}
		}

		if !foundRaiseToExit {
			t.Error("Raise inside finally should route to exit (unhandled exception)")
		}
	})
}
