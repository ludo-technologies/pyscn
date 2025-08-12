package analyzer

import (
	"strings"
	"testing"
)

func TestCFGBuilderConditionalFlow(t *testing.T) {
	t.Run("SimpleIf", func(t *testing.T) {
		source := `
if x > 0:
    print("positive")
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should have at least entry, exit, then block, merge block
		if cfg.Size() < 4 {
			t.Errorf("Expected at least 4 blocks, got %d", cfg.Size())
		}

		// Check for true and false edges
		hasCondTrue := false
		hasCondFalse := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeCondTrue {
					hasCondTrue = true
				}
				if e.Type == EdgeCondFalse {
					hasCondFalse = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})

		if !hasCondTrue {
			t.Error("Missing true branch edge")
		}
		if !hasCondFalse {
			t.Error("Missing false branch edge")
		}

		// Check for merge block
		hasMerge := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "if_merge") {
					hasMerge = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasMerge {
			t.Error("Missing merge block")
		}
	})

	t.Run("IfElse", func(t *testing.T) {
		source := `
if x > 0:
    print("positive")
else:
    print("non-positive")
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Check for then and else blocks
		hasThen := false
		hasElse := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "if_then") {
					hasThen = true
				}
				if strings.Contains(b.Label, "if_else") {
					hasElse = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasThen {
			t.Error("Missing then block")
		}
		if !hasElse {
			t.Error("Missing else block")
		}
	})

	t.Run("IfElifElse", func(t *testing.T) {
		source := `
if x > 10:
    print("greater than 10")
elif x == 10:
    print("equal to 10")
else:
    print("less than 10")
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should have blocks for branches and merges
		if cfg.Size() < 5 {
			t.Errorf("Expected at least 5 blocks for if/elif/else, got %d", cfg.Size())
		}

		// Count conditional edges
		condTrueCount := 0
		condFalseCount := 0
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeCondTrue {
					condTrueCount++
				}
				if e.Type == EdgeCondFalse {
					condFalseCount++
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})

		// Should have at least one true and false edge
		if condTrueCount < 1 {
			t.Errorf("Expected at least 1 true edge, got %d", condTrueCount)
		}
		if condFalseCount < 1 {
			t.Errorf("Expected at least 1 false edge, got %d", condFalseCount)
		}
	})

	t.Run("NestedIf", func(t *testing.T) {
		source := `
if x > 5:
    if x < 15:
        print("between 5 and 15")
    else:
        print("15 or greater")
else:
    print("5 or less")
print("done")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should have many blocks for nested structure
		if cfg.Size() < 8 {
			t.Errorf("Expected at least 8 blocks for nested if, got %d", cfg.Size())
		}

		// Should have multiple merge blocks
		mergeCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "merge") {
					mergeCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if mergeCount < 2 {
			t.Errorf("Expected at least 2 merge blocks for nested if, got %d", mergeCount)
		}
	})

	t.Run("IfWithReturn", func(t *testing.T) {
		source := `
def check(x):
    if x > 0:
        return True
    return False
`
		ast := parseSource(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Check that then branch connects to exit
		hasReturnFromThen := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "if_then") {
					// Check if this block has a return edge to exit
					for _, edge := range b.Successors {
						if edge.Type == EdgeReturn && edge.To == cfg.Exit {
							hasReturnFromThen = true
						}
					}
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasReturnFromThen {
			t.Error("Then branch should have return edge to exit")
		}
	})

	t.Run("MultipleElif", func(t *testing.T) {
		source := `
if x > 100:
    print("huge")
elif x > 50:
    print("large")
elif x > 10:
    print("medium")
elif x > 0:
    print("small")
else:
    print("non-positive")
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should have multiple blocks for the complex structure
		if cfg.Size() < 5 {
			t.Errorf("Expected at least 5 blocks for multiple elif, got %d", cfg.Size())
		}

		// Check that we have conditional edges
		hasCondEdges := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeCondTrue || e.Type == EdgeCondFalse {
					hasCondEdges = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})

		if !hasCondEdges {
			t.Error("Missing conditional edges")
		}
	})

	t.Run("ConditionalExpression", func(t *testing.T) {
		source := `
result = "positive" if x > 0 else "non-positive"
print(result)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// For now, conditional expressions are treated as simple expressions
		// Just verify the CFG builds successfully
		if cfg.Entry == nil || cfg.Exit == nil {
			t.Error("CFG missing entry or exit")
		}
	})

	t.Run("EmptyBranches", func(t *testing.T) {
		source := `
if x > 0:
    pass
else:
    pass
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should still create blocks for empty branches
		hasThen := false
		hasElse := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "if_then") {
					hasThen = true
				}
				if strings.Contains(b.Label, "if_else") {
					hasElse = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasThen {
			t.Error("Missing then block for empty branch")
		}
		if !hasElse {
			t.Error("Missing else block for empty branch")
		}
	})
}

func TestCFGBuilderComplexConditionals(t *testing.T) {
	source := `
def complex_conditionals(x, y):
    if x > 0:
        if y > 0:
            print("both positive")
        elif y == 0:
            print("x positive, y zero")
        else:
            print("x positive, y negative")
    elif x == 0:
        if y != 0:
            print("x zero, y non-zero")
        else:
            return "origin"
    else:
        if y > 0:
            return "x negative, y positive"
        print("x negative")
    
    if x + y > 100:
        return "sum large"
    elif x + y < -100:
        return "sum small"
    
    return "normal"
`

	ast := parseSource(t, source)
	funcNode := ast.Body[0]

	builder := NewCFGBuilder()
	cfg, err := builder.Build(funcNode)

	if err != nil {
		t.Fatalf("Failed to build CFG: %v", err)
	}

	// Verify complex structure has many blocks
	if cfg.Size() < 10 {
		t.Errorf("Expected at least 10 blocks for complex conditionals, got %d", cfg.Size())
	}

	// Count different edge types
	edgeCounts := make(map[EdgeType]int)
	cfg.Walk(&testVisitor{
		onEdge: func(e *Edge) bool {
			edgeCounts[e.Type]++
			return true
		},
		onBlock: func(b *BasicBlock) bool { return true },
	})

	// Check for presence of different edge types
	if edgeCounts[EdgeCondTrue] < 2 {
		t.Errorf("Expected multiple true edges, got %d", edgeCounts[EdgeCondTrue])
	}
	if edgeCounts[EdgeCondFalse] < 2 {
		t.Errorf("Expected multiple false edges, got %d", edgeCounts[EdgeCondFalse])
	}
	if edgeCounts[EdgeReturn] < 1 {
		t.Errorf("Expected at least one return edge, got %d", edgeCounts[EdgeReturn])
	}
}
