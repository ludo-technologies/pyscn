package analyzer

import (
	"strings"
	"testing"
)

func TestCFGBuilderComprehensions(t *testing.T) {
	t.Run("SimpleListComprehension", func(t *testing.T) {
		source := `
result = [x for x in range(10)]
print(result)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should have at least entry, exit, init, header, body, append blocks
		if cfg.Size() < 6 {
			t.Errorf("Expected at least 6 blocks for simple comprehension, got %d", cfg.Size())
		}

		// Check for comprehension-specific blocks
		hasCompInit := false
		hasCompHeader := false
		hasCompBody := false
		hasCompAppend := false
		hasCompExit := false

		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "comp_init") {
					hasCompInit = true
				}
				if strings.Contains(b.Label, "comp_header") {
					hasCompHeader = true
				}
				if strings.Contains(b.Label, "comp_body") {
					hasCompBody = true
				}
				if strings.Contains(b.Label, "comp_append") {
					hasCompAppend = true
				}
				if strings.Contains(b.Label, "comp_exit") {
					hasCompExit = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasCompInit {
			t.Error("Missing comprehension init block")
		}
		if !hasCompHeader {
			t.Error("Missing comprehension header block")
		}
		if !hasCompBody {
			t.Error("Missing comprehension body block")
		}
		if !hasCompAppend {
			t.Error("Missing comprehension append block")
		}
		if !hasCompExit {
			t.Error("Missing comprehension exit block")
		}

		// Check for loop edges
		hasCondTrue := false
		hasCondFalse := false
		hasLoopBack := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeCondTrue && strings.Contains(e.From.Label, "comp_header") {
					hasCondTrue = true
				}
				if e.Type == EdgeCondFalse && strings.Contains(e.From.Label, "comp_header") {
					hasCondFalse = true
				}
				if e.Type == EdgeLoop && strings.Contains(e.To.Label, "comp_header") {
					hasLoopBack = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})

		if !hasCondTrue {
			t.Error("Missing conditional true edge from comprehension header")
		}
		if !hasCondFalse {
			t.Error("Missing conditional false edge from comprehension header")
		}
		if !hasLoopBack {
			t.Error("Missing loop back edge to comprehension header")
		}
	})

	t.Run("FilteredListComprehension", func(t *testing.T) {
		source := `
evens = [x for x in range(10) if x % 2 == 0]
print(evens)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should have additional filter block
		if cfg.Size() < 7 {
			t.Errorf("Expected at least 7 blocks for filtered comprehension, got %d", cfg.Size())
		}

		// Check for filter block
		hasCompFilter := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "comp_filter") {
					hasCompFilter = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasCompFilter {
			t.Error("Missing comprehension filter block")
		}

		// Check filter edges
		hasFilterTrue := false
		hasFilterFalse := false
		cfg.Walk(&testVisitor{
			onEdge: func(e *Edge) bool {
				if e.Type == EdgeCondTrue && strings.Contains(e.From.Label, "comp_filter") {
					hasFilterTrue = true
				}
				if e.Type == EdgeCondFalse && strings.Contains(e.From.Label, "comp_filter") {
					hasFilterFalse = true
				}
				return true
			},
			onBlock: func(b *BasicBlock) bool { return true },
		})

		if !hasFilterTrue {
			t.Error("Missing conditional true edge from filter block")
		}
		if !hasFilterFalse {
			t.Error("Missing conditional false edge from filter block")
		}
	})

	t.Run("NestedComprehension", func(t *testing.T) {
		source := `
matrix = [[i*j for j in range(3)] for i in range(3)]
print(matrix)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Note: The inner comprehension is treated as the value expression of the outer comprehension
		// So we primarily see the outer comprehension's CFG structure
		// This is a design decision - we could recursively process inner comprehensions if needed
		if cfg.Size() < 6 {
			t.Errorf("Expected at least 6 blocks for nested comprehension, got %d", cfg.Size())
		}

		// Should have at least one comprehension header for the outer comprehension
		headerCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "comp_header") {
					headerCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if headerCount < 1 {
			t.Errorf("Expected at least 1 comprehension header for nested comprehension, got %d", headerCount)
		}

		// The test could be enhanced to recursively process inner comprehensions
		// For now, we're treating the inner comprehension as an opaque expression
	})

	t.Run("DictionaryComprehension", func(t *testing.T) {
		source := `
squares = {x: x**2 for x in range(5)}
print(squares)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Dictionary comprehension should have similar structure to list comprehension
		if cfg.Size() < 6 {
			t.Errorf("Expected at least 6 blocks for dict comprehension, got %d", cfg.Size())
		}

		// Verify basic comprehension structure exists
		hasCompHeader := false
		hasCompBody := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "comp_header") {
					hasCompHeader = true
				}
				if strings.Contains(b.Label, "comp_body") {
					hasCompBody = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasCompHeader {
			t.Error("Missing comprehension header for dict comprehension")
		}
		if !hasCompBody {
			t.Error("Missing comprehension body for dict comprehension")
		}
	})

	t.Run("SetComprehension", func(t *testing.T) {
		source := `
unique = {x % 5 for x in range(20)}
print(unique)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Set comprehension should have similar structure to list comprehension
		if cfg.Size() < 6 {
			t.Errorf("Expected at least 6 blocks for set comprehension, got %d", cfg.Size())
		}

		// Verify basic comprehension structure exists
		hasCompHeader := false
		hasCompBody := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "comp_header") {
					hasCompHeader = true
				}
				if strings.Contains(b.Label, "comp_body") {
					hasCompBody = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasCompHeader {
			t.Error("Missing comprehension header for set comprehension")
		}
		if !hasCompBody {
			t.Error("Missing comprehension body for set comprehension")
		}
	})

	t.Run("GeneratorExpression", func(t *testing.T) {
		source := `
gen = (x**2 for x in range(10))
result = list(gen)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Generator expression should have similar structure
		if cfg.Size() < 6 {
			t.Errorf("Expected at least 6 blocks for generator expression, got %d", cfg.Size())
		}

		hasCompHeader := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "comp_header") {
					hasCompHeader = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasCompHeader {
			t.Error("Missing comprehension header for generator expression")
		}
	})

	t.Run("MultipleForClauses", func(t *testing.T) {
		source := `
pairs = [(x, y) for x in range(3) for y in range(3)]
print(pairs)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Multiple for clauses should create nested loop structure
		if cfg.Size() < 8 {
			t.Errorf("Expected at least 8 blocks for multiple for clauses, got %d", cfg.Size())
		}

		// Count comprehension headers (should have 2 for two for clauses)
		headerCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "comp_header") {
					headerCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if headerCount < 2 {
			t.Errorf("Expected at least 2 headers for multiple for clauses, got %d", headerCount)
		}
	})

	t.Run("MultipleConditions", func(t *testing.T) {
		source := `
filtered = [x for x in range(100) if x % 2 == 0 if x % 3 == 0]
print(filtered)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Multiple conditions should create multiple filter blocks
		if cfg.Size() < 7 {
			t.Errorf("Expected at least 7 blocks for multiple conditions, got %d", cfg.Size())
		}

		// Check for filter blocks
		filterCount := 0
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "comp_filter") {
					filterCount++
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		// Note: Current implementation may combine conditions into one filter
		// This is implementation-specific and can be adjusted
		if filterCount < 1 {
			t.Error("Expected at least one filter block for multiple conditions")
		}
	})

	t.Run("ComprehensionInFunction", func(t *testing.T) {
		source := `
def process_data():
    result = [x * 2 for x in range(5)]
    return result

data = process_data()
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfgs, err := builder.BuildAll(ast)

		if err != nil {
			t.Fatalf("Failed to build CFGs: %v", err)
		}

		// Should have CFGs for main and process_data function
		if len(cfgs) < 2 {
			t.Errorf("Expected at least 2 CFGs, got %d", len(cfgs))
		}

		// Check the function CFG
		funcCFG, exists := cfgs["process_data"]
		if !exists {
			t.Fatal("Missing CFG for process_data function")
		}

		// Function with comprehension should have comprehension blocks
		hasCompHeader := false
		funcCFG.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "comp_header") {
					hasCompHeader = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasCompHeader {
			t.Error("Missing comprehension header in function CFG")
		}
	})

	t.Run("ComprehensionWithComplexExpression", func(t *testing.T) {
		source := `
result = [x if x % 2 == 0 else -x for x in range(10)]
print(result)
`
		ast := parseSource(t, source)
		builder := NewCFGBuilder()
		cfg, err := builder.Build(ast)

		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		// Should still have standard comprehension structure
		if cfg.Size() < 6 {
			t.Errorf("Expected at least 6 blocks for comprehension with conditional expression, got %d", cfg.Size())
		}

		// Check for comprehension blocks
		hasCompHeader := false
		hasCompBody := false
		cfg.Walk(&testVisitor{
			onBlock: func(b *BasicBlock) bool {
				if strings.Contains(b.Label, "comp_header") {
					hasCompHeader = true
				}
				if strings.Contains(b.Label, "comp_body") {
					hasCompBody = true
				}
				return true
			},
			onEdge: func(e *Edge) bool { return true },
		})

		if !hasCompHeader {
			t.Error("Missing comprehension header")
		}
		if !hasCompBody {
			t.Error("Missing comprehension body")
		}
	})
}
