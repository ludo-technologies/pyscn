package analyzer

import (
	"strings"
	"testing"
)

func TestCFGBuilderExceptionHandling(t *testing.T) {
	t.Run("SimpleTryExcept", func(t *testing.T) {
		source := `
try:
    risky_operation()
except ValueError:
    print("ValueError occurred")
print("after try")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Check for try and except blocks
		hasTryBlock := false
		hasExceptBlock := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "try_block") {
					hasTryBlock = true
				}
				if strings.Contains(b.Label, "except_block") {
					hasExceptBlock = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasTryBlock {
			t.Error("Missing try block")
		}
		if !hasExceptBlock {
			t.Error("Missing except block")
		}
		
		// Check for exception edge
		hasExceptionEdge := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeException {
					hasExceptionEdge = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if !hasExceptionEdge {
			t.Error("Missing exception edge")
		}
	})
	
	t.Run("TryExceptElse", func(t *testing.T) {
		source := `
try:
    result = compute_value()
except ValueError:
    result = None
else:
    print("No exception occurred")
print("result:", result)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Check for try, except, and else blocks
		hasTryBlock := false
		hasExceptBlock := false
		hasTryElseBlock := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "try_block") {
					hasTryBlock = true
				}
				if strings.Contains(b.Label, "except_block") {
					hasExceptBlock = true
				}
				if strings.Contains(b.Label, "try_else") {
					hasTryElseBlock = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasTryBlock {
			t.Error("Missing try block")
		}
		if !hasExceptBlock {
			t.Error("Missing except block")
		}
		if !hasTryElseBlock {
			t.Error("Missing try else block")
		}
	})
	
	t.Run("TryExceptFinally", func(t *testing.T) {
		source := `
try:
    file = open("test.txt")
    content = file.read()
except IOError:
    content = "default"
finally:
    file.close()
print("content:", content)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Check for finally block
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
		
		if !hasFinallyBlock {
			t.Error("Missing finally block")
		}
		
		// Finally should be reachable from both try and except paths
		finallyBlockReachable := 0
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.To != nil && strings.Contains(e.To.Label, "finally_block") {
					finallyBlockReachable++
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if finallyBlockReachable < 2 {
			t.Errorf("Finally block should be reachable from multiple paths, got %d", finallyBlockReachable)
		}
	})
	
	t.Run("TryExceptElseFinally", func(t *testing.T) {
		source := `
try:
    value = risky_calculation()
except ValueError:
    value = 0
except TypeError:
    value = -1
else:
    print("Success!")
finally:
    print("Cleanup")
print("value:", value)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Count except blocks (should have 2)
		exceptBlockCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "except_block") {
					exceptBlockCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if exceptBlockCount != 2 {
			t.Errorf("Expected 2 except blocks, got %d", exceptBlockCount)
		}
		
		// Check for all required blocks
		hasTryBlock := false
		hasTryElseBlock := false
		hasFinallyBlock := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "try_block") {
					hasTryBlock = true
				}
				if strings.Contains(b.Label, "try_else") {
					hasTryElseBlock = true
				}
				if strings.Contains(b.Label, "finally_block") {
					hasFinallyBlock = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasTryBlock {
			t.Error("Missing try block")
		}
		if !hasTryElseBlock {
			t.Error("Missing try else block")
		}
		if !hasFinallyBlock {
			t.Error("Missing finally block")
		}
	})
	
	t.Run("MultipleExceptHandlers", func(t *testing.T) {
		source := `
try:
    dangerous_operation()
except ValueError as ve:
    print("Value error:", ve)
except TypeError as te:
    print("Type error:", te)
except Exception as e:
    print("General error:", e)
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Should have 3 except blocks
		exceptBlockCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "except_block") {
					exceptBlockCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if exceptBlockCount != 3 {
			t.Errorf("Expected 3 except blocks, got %d", exceptBlockCount)
		}
		
		// Should have multiple exception edges from try to handlers
		exceptionEdgeCount := 0
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeException {
					exceptionEdgeCount++
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if exceptionEdgeCount < 3 {
			t.Errorf("Expected at least 3 exception edges, got %d", exceptionEdgeCount)
		}
	})
	
	t.Run("RaiseStatement", func(t *testing.T) {
		source := `
def validate(x):
    if x < 0:
        raise ValueError("Negative value not allowed")
    return x * 2

try:
    result = validate(-5)
except ValueError:
    result = 0
print("result:", result)
`
		ast := parseSource(t, source)
		
		// Build CFG for the function
		funcNode := ast.Body[0]
		builder := NewCFGBuilder()
		funcCfg, err := builder.Build(funcNode)
		
		if err != nil {
			t.Fatalf("Failed to build function CFG: %v", err)
		}
		
		// Check for exception edge from raise
		hasRaiseExceptionEdge := false
		funcCfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeException {
					hasRaiseExceptionEdge = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if !hasRaiseExceptionEdge {
			t.Error("Missing exception edge from raise statement")
		}
		
		// Check for unreachable block after raise
		hasUnreachable := false
		funcCfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "unreachable") {
					hasUnreachable = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasUnreachable {
			t.Error("Missing unreachable block after raise")
		}
		
		// Build CFG for the main module with try/except
		builder2 := NewCFGBuilder()
		mainCfg, err := builder2.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build main CFG: %v", err)
		}
		
		// Main should have try/except structure
		hasTryBlock := false
		hasExceptBlock := false
		mainCfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "try_block") {
					hasTryBlock = true
				}
				if strings.Contains(b.Label, "except_block") {
					hasExceptBlock = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasTryBlock {
			t.Error("Missing try block in main")
		}
		if !hasExceptBlock {
			t.Error("Missing except block in main")
		}
	})
	
	t.Run("NestedTryBlocks", func(t *testing.T) {
		source := `
try:
    print("outer try")
    try:
        print("inner try")
        risky_operation()
    except ValueError:
        print("inner except")
    print("between inner and outer")
except Exception:
    print("outer except")
finally:
    print("outer finally")
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Should have multiple try blocks
		tryBlockCount := 0
		exceptBlockCount := 0
		finallyBlockCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "try_block") {
					tryBlockCount++
				}
				if strings.Contains(b.Label, "except_block") {
					exceptBlockCount++
				}
				if strings.Contains(b.Label, "finally_block") {
					finallyBlockCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if tryBlockCount < 2 {
			t.Errorf("Expected at least 2 try blocks for nested try, got %d", tryBlockCount)
		}
		if exceptBlockCount < 2 {
			t.Errorf("Expected at least 2 except blocks for nested try, got %d", exceptBlockCount)
		}
		if finallyBlockCount != 1 {
			t.Errorf("Expected 1 finally block, got %d", finallyBlockCount)
		}
	})
	
	t.Run("EmptyTryBlocks", func(t *testing.T) {
		source := `
try:
    pass
except:
    pass
else:
    pass
finally:
    pass
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Even empty blocks should be created
		hasTryBlock := false
		hasExceptBlock := false
		hasTryElseBlock := false
		hasFinallyBlock := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "try_block") {
					hasTryBlock = true
				}
				if strings.Contains(b.Label, "except_block") {
					hasExceptBlock = true
				}
				if strings.Contains(b.Label, "try_else") {
					hasTryElseBlock = true
				}
				if strings.Contains(b.Label, "finally_block") {
					hasFinallyBlock = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})
		
		if !hasTryBlock {
			t.Error("Missing try block")
		}
		if !hasExceptBlock {
			t.Error("Missing except block")
		}
		if !hasTryElseBlock {
			t.Error("Missing try else block")
		}
		if !hasFinallyBlock {
			t.Error("Missing finally block")
		}
	})
	
	t.Run("RaiseWithinTryExcept", func(t *testing.T) {
		source := `
try:
    if condition:
        raise RuntimeError("Something went wrong")
    print("normal path")
except RuntimeError:
    print("caught runtime error")
except:
    print("caught other error")
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Should have exception edges both from raise and from try to except
		exceptionEdgeCount := 0
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeException {
					exceptionEdgeCount++
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		// Should have at least 4 exception edges:
		// 1. try -> except_block_1
		// 2. try -> except_block_2  
		// 3. raise -> except_block_1
		// 4. raise -> except_block_2
		if exceptionEdgeCount < 4 {
			t.Errorf("Expected at least 4 exception edges, got %d", exceptionEdgeCount)
		}
	})
}

func TestCFGBuilderExceptionEdgeCases(t *testing.T) {
	t.Run("RaiseOutsideTry", func(t *testing.T) {
		source := `
print("before")
raise ValueError("error")
print("after")  # unreachable
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Should connect raise to exit (unhandled exception)
		hasExceptionToExit := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeException && e.To == cfg.Exit {
					hasExceptionToExit = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})
		
		if !hasExceptionToExit {
			t.Error("Missing exception edge to exit for unhandled exception")
		}
		
		// Should create unreachable block
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
			t.Error("Missing unreachable block after raise")
		}
	})
	
	t.Run("ReturnInTryFinally", func(t *testing.T) {
		source := `
def test_return():
    try:
        return "from try"
    finally:
        print("finally executed")
    print("unreachable")
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]
		
		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}
		
		// Should have return edge and finally block
		hasReturnEdge := false
		hasFinallyBlock := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeReturn {
					hasReturnEdge = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "finally_block") {
					hasFinallyBlock = true
				}
				return true
			},
		})
		
		if !hasReturnEdge {
			t.Error("Missing return edge")
		}
		if !hasFinallyBlock {
			t.Error("Missing finally block")
		}
	})
}

