package analyzer

import (
	"strings"
	"testing"
)

func TestCFGBuilderLoops(t *testing.T) {
	t.Run("SimpleForLoop", func(t *testing.T) {
		source := `
for i in range(10):
    print(i)
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Should have at least entry, exit, header, body, and exit blocks
		if cfg.Size() < 5 {
			t.Errorf("Expected at least 5 blocks, got %d", cfg.Size())
		}
		
		// Check for loop header and body blocks
		hasLoopHeader := false
		hasLoopBody := false
		hasLoopExit := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "loop_header") {
					hasLoopHeader = true
				}
				if strings.Contains(b.Label, "loop_body") {
					hasLoopBody = true
				}
				if strings.Contains(b.Label, "loop_exit") {
					hasLoopExit = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasLoopHeader {
			t.Error("Missing loop header block")
		}
		if !hasLoopBody {
			t.Error("Missing loop body block")
		}
		if !hasLoopExit {
			t.Error("Missing loop exit block")
		}
		
		// Check for loop edges
		hasCondTrue := false
		hasCondFalse := false
		hasLoopBack := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeCondTrue {
					hasCondTrue = true
				}
				if e.Type == EdgeCondFalse {
					hasCondFalse = true
				}
				if e.Type == EdgeLoop {
					hasLoopBack = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if !hasCondTrue {
			t.Error("Missing conditional true edge")
		}
		if !hasCondFalse {
			t.Error("Missing conditional false edge")
		}
		if !hasLoopBack {
			t.Error("Missing loop back edge")
		}
	})
	
	t.Run("ForLoopWithElse", func(t *testing.T) {
		source := `
for i in range(5):
    if i == 3:
        break
    print(i)
else:
    print("completed normally")
print("end")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Check for else block
		hasLoopElse := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "loop_else") {
					hasLoopElse = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasLoopElse {
			t.Error("Missing loop else block")
		}
		
		// Check for break edge
		hasBreakEdge := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeBreak {
					hasBreakEdge = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if !hasBreakEdge {
			t.Error("Missing break edge")
		}
	})
	
	t.Run("SimpleWhileLoop", func(t *testing.T) {
		source := `
count = 0
while count < 10:
    print(count)
    count += 1
print("finished")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Check for while loop structure
		hasLoopHeader := false
		hasLoopBody := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "loop_header") {
					hasLoopHeader = true
				}
				if strings.Contains(b.Label, "loop_body") {
					hasLoopBody = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasLoopHeader {
			t.Error("Missing while loop header block")
		}
		if !hasLoopBody {
			t.Error("Missing while loop body block")
		}
		
		// Check for loop back edge
		hasLoopBack := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeLoop {
					hasLoopBack = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if !hasLoopBack {
			t.Error("Missing loop back edge")
		}
	})
	
	t.Run("WhileLoopWithElse", func(t *testing.T) {
		source := `
i = 0
while i < 3:
    print(i)
    i += 1
else:
    print("loop completed")
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Check for else block
		hasLoopElse := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "loop_else") {
					hasLoopElse = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasLoopElse {
			t.Error("Missing while loop else block")
		}
	})
	
	t.Run("BreakStatement", func(t *testing.T) {
		source := `
for i in range(10):
    if i == 5:
        break
    print(i)
print("after loop")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Check for break edge
		hasBreakEdge := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeBreak {
					hasBreakEdge = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if !hasBreakEdge {
			t.Error("Missing break edge")
		}
		
		// Check for unreachable block after break
		hasUnreachable := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "unreachable") {
					hasUnreachable = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasUnreachable {
			t.Error("Missing unreachable block after break")
		}
	})
	
	t.Run("ContinueStatement", func(t *testing.T) {
		source := `
for i in range(10):
    if i % 2 == 0:
        continue
    print(i)
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Check for continue edge
		hasContinueEdge := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeContinue {
					hasContinueEdge = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if !hasContinueEdge {
			t.Error("Missing continue edge")
		}
	})
	
	t.Run("NestedLoops", func(t *testing.T) {
		source := `
for i in range(3):
    for j in range(3):
        if i == j:
            continue
        if i + j > 3:
            break
        print(i, j)
    print("outer loop iteration", i)
print("all done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Should have complex structure for nested loops
		if cfg.Size() < 10 {
			t.Errorf("Expected at least 10 blocks for nested loops, got %d", cfg.Size())
		}
		
		// Check for multiple loop structures
		loopHeaderCount := 0
		loopBodyCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "loop_header") {
					loopHeaderCount++
				}
				if strings.Contains(b.Label, "loop_body") {
					loopBodyCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if loopHeaderCount < 2 {
			t.Errorf("Expected at least 2 loop headers for nested loops, got %d", loopHeaderCount)
		}
		if loopBodyCount < 2 {
			t.Errorf("Expected at least 2 loop bodies for nested loops, got %d", loopBodyCount)
		}
		
		// Check for both continue and break edges
		hasContinueEdge := false
		hasBreakEdge := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeContinue {
					hasContinueEdge = true
				}
				if e.Type == EdgeBreak {
					hasBreakEdge = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if !hasContinueEdge {
			t.Error("Missing continue edge in nested loops")
		}
		if !hasBreakEdge {
			t.Error("Missing break edge in nested loops")
		}
	})
	
	t.Run("AsyncForLoop", func(t *testing.T) {
		source := `
async def process():
    async for item in async_iterator():
        await process_item(item)
    print("processing complete")
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]
		
		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Check for async for loop structure (should be same as regular for loop)
		hasLoopHeader := false
		hasLoopBody := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "loop_header") {
					hasLoopHeader = true
				}
				if strings.Contains(b.Label, "loop_body") {
					hasLoopBody = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasLoopHeader {
			t.Error("Missing async for loop header block")
		}
		if !hasLoopBody {
			t.Error("Missing async for loop body block")
		}
	})
	
	t.Run("EmptyLoops", func(t *testing.T) {
		source := `
for i in range(5):
    pass

while True:
    break
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Even empty loops should create proper structure
		loopHeaderCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "loop_header") {
					loopHeaderCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if loopHeaderCount < 2 {
			t.Errorf("Expected 2 loop headers, got %d", loopHeaderCount)
		}
	})
}

func TestCFGBuilderLoopEdgeCases(t *testing.T) {
	t.Run("BreakOutsideLoop", func(t *testing.T) {
		source := `
print("before")
break  # This is invalid Python but we should handle it gracefully
print("after")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		
		// Use a logger to capture the error
		// For now, just ensure it doesn't crash
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Should still create a valid CFG
		if cfg.Size() < 2 {
			t.Errorf("Expected at least 2 blocks, got %d", cfg.Size())
		}
	})
	
	t.Run("ContinueOutsideLoop", func(t *testing.T) {
		source := `
print("before")
continue  # This is invalid Python but we should handle it gracefully
print("after")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Should still create a valid CFG
		if cfg.Size() < 2 {
			t.Errorf("Expected at least 2 blocks, got %d", cfg.Size())
		}
	})
	
	t.Run("MultipleBreaksAndContinues", func(t *testing.T) {
		source := `
for i in range(10):
    if i == 2:
        continue
    if i == 5:
        break
    if i == 7:
        continue  # This is unreachable due to break at i==5
    print(i)
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Count break and continue edges
		breakCount := 0
		continueCount := 0
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeBreak {
					breakCount++
				}
				if e.Type == EdgeContinue {
					continueCount++
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if breakCount != 1 {
			t.Errorf("Expected 1 break edge, got %d", breakCount)
		}
		if continueCount < 1 {
			t.Errorf("Expected at least 1 continue edge, got %d", continueCount)
		}
	})
}

