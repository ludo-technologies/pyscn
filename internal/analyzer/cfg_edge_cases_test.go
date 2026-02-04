package analyzer

import (
	"context"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCFGEdgeCases tests various edge cases and boundary conditions
func TestCFGEdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		code        string
		expectError bool
		description string
	}{
		{
			name:        "NilInput",
			code:        "",
			expectError: false,
			description: "Empty input should be handled gracefully",
		},
		{
			name:        "OnlyWhitespace",
			code:        "   \n\t  \n   ",
			expectError: false,
			description: "Whitespace-only input should be handled",
		},
		{
			name: "OnlyComments",
			code: `
# This is a comment
# Another comment
# Third comment
`,
			expectError: false,
			description: "Comment-only files should be handled",
		},
		{
			name:        "SinglePass",
			code:        "pass",
			expectError: false,
			description: "Single pass statement should work",
		},
		{
			name:        "SingleReturn",
			code:        "return 42",
			expectError: false,
			description: "Single return statement should work",
		},
		{
			name: "DeeplyNestedStructures",
			code: `
def deeply_nested():
    if True:
        if True:
            if True:
                if True:
                    if True:
                        if True:
                            if True:
                                if True:
                                    if True:
                                        if True:
                                            return "deep"
                                        else:
                                            return "else10"
                                    else:
                                        return "else9"
                                else:
                                    return "else8"
                            else:
                                return "else7"
                        else:
                            return "else6"
                    else:
                        return "else5"
                else:
                    return "else4"
            else:
                return "else3"
        else:
            return "else2"
    else:
        return "else1"
`,
			expectError: false,
			description: "Deeply nested structures should not cause stack overflow",
		},
		{
			name:        "LargeNumberOfVariables",
			code:        generateLargeVariableCode(100),
			expectError: false,
			description: "Large number of variables should be handled",
		},
		{
			name:        "ManySequentialStatements",
			code:        generateSequentialStatements(100),
			expectError: false,
			description: "Many sequential statements should be handled",
		},
		{
			name: "EmptyFunction",
			code: `
def empty_function():
    pass
`,
			expectError: false,
			description: "Empty function should be handled",
		},
		{
			name: "EmptyClass",
			code: `
class EmptyClass:
    pass
`,
			expectError: false,
			description: "Empty class should be handled",
		},
		{
			name: "MultipleEmptyFunctions",
			code: `
def func1():
    pass

def func2():
    pass

def func3():
    pass
`,
			expectError: false,
			description: "Multiple empty functions should be handled",
		},
		{
			name: "RecursiveFunction",
			code: `
def factorial(n):
    if n <= 1:
        return 1
    else:
        return n * factorial(n - 1)
`,
			expectError: false,
			description: "Recursive function should be handled",
		},
		{
			name: "ComplexExpressions",
			code: `
def complex_expr():
    result = (a + b * c - d / e) ** f % g & h | i ^ j << k >> l
    return result and (x or y) and not z
`,
			expectError: false,
			description: "Complex expressions should be handled",
		},
		{
			name: "UnicodeIdentifiers",
			code: `
def tëst_ünicøde():
    αβγ = 42
    return αβγ
`,
			expectError: false,
			description: "Unicode identifiers should be handled",
		},
		{
			name: "VeryLongFunctionName",
			code: `
def this_is_a_very_long_function_name_that_exceeds_normal_length_limits_and_should_still_work():
    return "long_name"
`,
			expectError: false,
			description: "Very long function names should be handled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the code
			p := parser.New()
			ctx := context.Background()
			result, err := p.Parse(ctx, []byte(tc.code))

			if tc.expectError {
				assert.Error(t, err, tc.description)
				return
			}

			require.NoError(t, err, "Parse failed: %s", tc.description)
			ast := result.AST

			// Build CFG
			builder := NewCFGBuilder()
			cfg, err := builder.Build(ast)

			if tc.expectError {
				assert.Error(t, err, tc.description)
				return
			}

			require.NoError(t, err, "CFG build failed: %s", tc.description)

			// Basic validation
			assert.NotNil(t, cfg, "CFG should not be nil")
			assert.NotNil(t, cfg.Entry, "Entry block should exist")
			assert.NotNil(t, cfg.Exit, "Exit block should exist")
			assert.GreaterOrEqual(t, cfg.Size(), 2, "CFG should have at least entry and exit")

			// Test reachability analysis
			analyzer := NewReachabilityAnalyzer(cfg)
			reachResult := analyzer.AnalyzeReachability()

			// Entry should always be reachable
			_, entryReachable := reachResult.ReachableBlocks[cfg.Entry.ID]
			assert.True(t, entryReachable, "Entry should be reachable")

			// Reachability ratio should be valid
			ratio := reachResult.GetReachabilityRatio()
			assert.GreaterOrEqual(t, ratio, 0.0, "Ratio should be >= 0")
			assert.LessOrEqual(t, ratio, 1.0, "Ratio should be <= 1")
		})
	}
}

// TestCFGErrorRecovery tests error recovery and graceful degradation
func TestCFGErrorRecovery(t *testing.T) {
	testCases := []struct {
		name        string
		setupFunc   func() (*CFGBuilder, error)
		description string
	}{
		{
			name: "BuilderWithNilAST",
			setupFunc: func() (*CFGBuilder, error) {
				builder := NewCFGBuilder()
				_, err := builder.Build(nil)
				return builder, err
			},
			description: "Building CFG with nil AST should return error",
		},
		{
			name: "MultipleBuilds",
			setupFunc: func() (*CFGBuilder, error) {
				builder := NewCFGBuilder()

				// Parse simple code
				p := parser.New()
				ctx := context.Background()
				result1, err := p.Parse(ctx, []byte("x = 1"))
				if err != nil {
					return builder, err
				}
				ast1 := result1.AST

				// Build first CFG
				_, err = builder.Build(ast1)
				if err != nil {
					return builder, err
				}

				// Build second CFG with same builder
				result2, err := p.Parse(ctx, []byte("y = 2"))
				if err != nil {
					return builder, err
				}
				ast2 := result2.AST

				_, err = builder.Build(ast2)
				return builder, err
			},
			description: "Multiple builds with same builder should work",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			builder, err := tc.setupFunc()

			// For error recovery tests, we expect some to fail gracefully
			if tc.name == "BuilderWithNilAST" {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}

			assert.NotNil(t, builder, "Builder should not be nil")
		})
	}
}

// TestCFGMemoryEfficiency tests memory usage patterns
func TestCFGMemoryEfficiency(t *testing.T) {
	// Test that CFG doesn't hold onto unnecessary references
	code := `
def memory_test():
    x = 1
    y = 2
    if x < y:
        z = x + y
        return z
    else:
        w = x - y
        return w
`

	// Parse the code
	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(code))
	require.NoError(t, err)
	ast := result.AST

	// Build CFG
	builder := NewCFGBuilder()
	cfg, err := builder.Build(ast)
	require.NoError(t, err)

	// Verify CFG structure is reasonable
	assert.LessOrEqual(t, cfg.Size(), 20, "CFG should not have excessive blocks")

	// Verify no circular references in basic structure
	visited := make(map[*BasicBlock]bool)
	var checkCircular func(*BasicBlock, map[*BasicBlock]bool) bool
	checkCircular = func(block *BasicBlock, path map[*BasicBlock]bool) bool {
		if path[block] {
			return true // Found cycle
		}
		if visited[block] {
			return false // Already checked this path
		}

		visited[block] = true
		path[block] = true

		for _, edge := range block.Successors {
			if checkCircular(edge.To, path) {
				return true
			}
		}

		delete(path, block)
		return false
	}

	// Note: Cycles are expected in CFGs (loops), so this test is more about
	// ensuring we don't have infinite recursion in our data structures
	checkCircular(cfg.Entry, make(map[*BasicBlock]bool))
}

// TestCFGConcurrency tests thread safety of CFG operations
func TestCFGConcurrency(t *testing.T) {
	code := `
def concurrent_test():
    for i in range(10):
        if i % 2 == 0:
            print("even")
        else:
            print("odd")
    return "done"
`

	// Parse the code once
	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(code))
	require.NoError(t, err)
	ast := result.AST

	// Build CFG
	builder := NewCFGBuilder()
	cfg, err := builder.Build(ast)
	require.NoError(t, err)

	// Test concurrent reachability analysis
	// Note: This is a basic test - proper concurrency testing would need more sophisticated setup
	analyzer1 := NewReachabilityAnalyzer(cfg)
	analyzer2 := NewReachabilityAnalyzer(cfg)

	result1 := analyzer1.AnalyzeReachability()
	result2 := analyzer2.AnalyzeReachability()

	// Results should be the same
	assert.Equal(t, len(result1.ReachableBlocks), len(result2.ReachableBlocks),
		"Concurrent analysis should produce same results")
}

// Helper functions for generating test code

func generateLargeVariableCode(count int) string {
	var builder strings.Builder
	builder.WriteString("def large_vars():\n")

	for i := 0; i < count; i++ {
		builder.WriteString("    var")
		builder.WriteString(strings.Repeat("_", i%10))
		builder.WriteString(" = ")
		builder.WriteString(string(rune('0' + (i % 10))))
		builder.WriteString("\n")
	}

	builder.WriteString("    return var0\n")
	return builder.String()
}

func generateSequentialStatements(count int) string {
	var builder strings.Builder
	builder.WriteString("def sequential():\n")

	for i := 0; i < count; i++ {
		builder.WriteString("    x = ")
		builder.WriteString(string(rune('0' + (i % 10))))
		builder.WriteString("\n")
	}

	builder.WriteString("    return x\n")
	return builder.String()
}

// TestCFGBoundaryConditions tests specific boundary conditions
func TestCFGBoundaryConditions(t *testing.T) {
	testCases := []struct {
		name      string
		code      string
		checkFunc func(t *testing.T, cfg *CFG)
	}{
		{
			name: "SingleNode",
			code: "pass",
			checkFunc: func(t *testing.T, cfg *CFG) {
				// Should have entry and exit at minimum
				assert.GreaterOrEqual(t, cfg.Size(), 2)
			},
		},
		{
			name: "NoControlFlow",
			code: `
x = 1
y = 2
z = x + y
`,
			checkFunc: func(t *testing.T, cfg *CFG) {
				// Linear code should have minimal blocks
				assert.LessOrEqual(t, cfg.Size(), 5)
			},
		},
		{
			name: "OnlyReturns",
			code: `
def only_returns():
    return 1
    return 2  # unreachable
    return 3  # unreachable
`,
			checkFunc: func(t *testing.T, cfg *CFG) {
				// Should detect unreachable returns
				analyzer := NewReachabilityAnalyzer(cfg)
				result := analyzer.AnalyzeReachability()
				assert.True(t, result.HasUnreachableCode())
			},
		},
		{
			name: "NestedReturns",
			code: `
def nested_returns():
    if True:
        if True:
            return 1
            x = 1  # unreachable
        return 2  # reachable (CFG doesn't know True is always true)
        y = 2  # unreachable
    return 3  # reachable (CFG doesn't know True is always true)
    z = 3  # unreachable
`,
			checkFunc: func(t *testing.T, cfg *CFG) {
				// Should detect unreachable code after return statements
				analyzer := NewReachabilityAnalyzer(cfg)
				result := analyzer.AnalyzeReachability()
				unreachable := result.GetUnreachableBlocksWithStatements()
				assert.GreaterOrEqual(t, len(unreachable), 3) // At least 3 unreachable blocks with statements
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse the code
			p := parser.New()
			ctx := context.Background()
			result, err := p.Parse(ctx, []byte(tc.code))
			require.NoError(t, err)
			ast := result.AST

			// Build all CFGs
			builder := NewCFGBuilder()
			cfgs, err := builder.BuildAll(ast)
			require.NoError(t, err)

			// Get appropriate CFG based on test case
			var cfg *CFG
			if tc.name == "OnlyReturns" || tc.name == "NestedReturns" {
				// These tests define functions, get the function CFG
				for name, c := range cfgs {
					if name != "__main__" {
						cfg = c
						break
					}
				}
				require.NotNil(t, cfg, "Failed to find function CFG")
			} else {
				// Module-level tests
				cfg = cfgs["__main__"]
				require.NotNil(t, cfg, "Failed to find main CFG")
			}

			// Run specific checks
			tc.checkFunc(t, cfg)
		})
	}
}

// TestSimpleWalrusStatement tests the detection of the walrus operator
func TestSimpleWalrusStatement(t *testing.T) {
	code := `
		def simple_walrus():
			(x := 42)
			return x
	`
	funcName := "simple_walrus"

	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(code))
	require.NoError(t, err)

	builder := NewCFGBuilder()
	cfgs, err := builder.BuildAll(result.AST)
	require.NoError(t, err)

	cfg, exists := cfgs[funcName]
	require.True(t, exists)

	foundExpr := false
	foundWalrus := false

	// Searching for the block wich contain the instruction
	for _, block := range cfg.Blocks {
		for _, stmt := range block.Statements {
			// First verification, is it a NodeExpr (the container) ?
			if stmt.Type == parser.NodeExpr {
				foundExpr = true
				// Second verification, does it contain the Walrus ?
				stmt.Walk(func(n *parser.Node) bool {
					if n.Type == parser.NodeNamedExpr {
						foundWalrus = true
						return false
					}
					return true
				})
			}
		}
	}

	assert.True(t, foundExpr, "Should have detected a NodeExpr")
	assert.True(t, foundWalrus, "The NodeExpr should contain a NodeNamedExpr")
}

// TestWalrusOperatorInConditional tests support for the walrus operator (:=) in if conditions
func TestWalrusOperatorInConditional(t *testing.T) {
	code := `
		def walrus_if():
			items = [1, 2, 3]
			if (n := len(items)) > 2:
				return n
			return 0
	`
	funcName := "walrus_if"

	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(code))
	require.NoError(t, err)

	builder := NewCFGBuilder()
	cfgs, err := builder.BuildAll(result.AST)
	require.NoError(t, err)

	cfg, exists := cfgs[funcName]
	require.True(t, exists, "Function CFG should exist")
	assert.NotNil(t, cfg)

	// Verify structure
	assert.NotNil(t, cfg.Entry)
	assert.NotNil(t, cfg.Exit)

	// Find the if block
	var ifBlock *BasicBlock
	for _, block := range cfg.Blocks {
		// In the builder, the if condition usually ends a block
		// We look for the block that has CondTrue and CondFalse edges
		hasTrue := false
		hasFalse := false
		for _, edge := range block.Successors {
			if edge.Type == EdgeCondTrue {
				hasTrue = true
			}
			if edge.Type == EdgeCondFalse {
				hasFalse = true
			}
		}
		if hasTrue && hasFalse {
			ifBlock = block
			break
		}
	}
	require.NotNil(t, ifBlock, "Should find a block with conditional edges")

	// Verify the walrus operator is processed
	foundWalrus := false
	for _, stmt := range ifBlock.Statements {
		stmt.Walk(func(n *parser.Node) bool {
			if n.Type == parser.NodeNamedExpr {
				foundWalrus = true
				return false
			}
			return true
		})
		if foundWalrus {
			break
		}
	}
	assert.True(t, foundWalrus, "Should find named expression (walrus) in the conditional block")
}

// TestWalrusOperatorInWhile tests support for the walrus operator (:=) in while loops
func TestWalrusOperatorInWhile(t *testing.T) {
	code := `
		def walrus_while():
			data = [1, 2, 3, 0]
			it = iter(data)
			res = []
			while (x := next(it)) != 0:
				res.append(x)
			return res
	`
	funcName := "walrus_while"

	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(code))
	require.NoError(t, err)

	builder := NewCFGBuilder()
	cfgs, err := builder.BuildAll(result.AST)
	require.NoError(t, err)

	cfg, exists := cfgs[funcName]
	require.True(t, exists, "Function CFG should exist")
	assert.NotNil(t, cfg)

	// Find loop header
	var loopHeader *BasicBlock
	for _, block := range cfg.Blocks {
		if strings.HasPrefix(block.Label, LabelLoopHeader) {
			loopHeader = block
			break
		}
	}

	require.NotNil(t, loopHeader, "Should find loop header")

	// The while condition contains the walrus operator.
	// It should be present in the loop header block statements.
	foundWalrus := false
	for _, stmt := range loopHeader.Statements {
		stmt.Walk(func(n *parser.Node) bool {
			if n.Type == parser.NodeNamedExpr {
				foundWalrus = true
				return false
			}
			return true
		})
		if foundWalrus {
			break
		}
	}
	assert.True(t, foundWalrus, "Loop header should contain walrus operator")
}

// TestWalrusOperatorInComprehension tests support for the walrus operator (:=) in comprehensions
func TestWalrusOperatorInComprehension(t *testing.T) {
	code := `
		def walrus_comp():
			data = [1, 2, 3]
			return [y for x in data if (y := x * 2) > 2]
	`
	funcName := "walrus_comp"

	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(code))
	require.NoError(t, err)

	builder := NewCFGBuilder()
	cfgs, err := builder.BuildAll(result.AST)
	require.NoError(t, err)

	cfg, exists := cfgs[funcName]
	require.True(t, exists, "Function CFG should exist")
	assert.NotNil(t, cfg)

	// Verify comprehension structure and walrus operator location
	hasCompInit := false
	hasCompHeader := false
	hasCompFilter := false
	hasCompBody := false
	hasCompAppend := false
	hasCompExit := false
	foundWalrusInFilter := false

	cfg.Walk(&testVisitor{
		onBlock: func(b *BasicBlock) bool {
			if strings.Contains(b.Label, "comp_init") {
				hasCompInit = true
			}
			if strings.Contains(b.Label, "comp_header") {
				hasCompHeader = true
			}
			if strings.Contains(b.Label, "comp_filter") {
				hasCompFilter = true
				// The walrus operator (y := x * 2) is in the filter condition
				for _, stmt := range b.Statements {
					stmt.Walk(func(n *parser.Node) bool {
						if n.Type == parser.NodeNamedExpr {
							foundWalrusInFilter = true
							return false
						}
						return true
					})
				}
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

	assert.True(t, hasCompInit, "Should have comp_init block")
	assert.True(t, hasCompHeader, "Should have comp_header block")
	assert.True(t, hasCompFilter, "Should have comp_filter block")
	assert.True(t, hasCompBody, "Should have comp_body block")
	assert.True(t, hasCompAppend, "Should have comp_append block")
	assert.True(t, hasCompExit, "Should have comp_exit block")
	assert.True(t, foundWalrusInFilter, "Walrus operator should be in comp_filter block")

	// Verify edges
	hasHeaderToBody := false
	hasBodyToFilter := false
	hasFilterToAppend := false
	hasFilterToHeader := false
	hasAppendToHeader := false
	hasHeaderToExit := false

	cfg.Walk(&testVisitor{
		onEdge: func(e *Edge) bool {
			if strings.Contains(e.From.Label, "comp_header") && strings.Contains(e.To.Label, "comp_body") && e.Type == EdgeCondTrue {
				hasHeaderToBody = true
			}
			if strings.Contains(e.From.Label, "comp_body") && strings.Contains(e.To.Label, "comp_filter") && e.Type == EdgeNormal {
				hasBodyToFilter = true
			}
			if strings.Contains(e.From.Label, "comp_filter") && strings.Contains(e.To.Label, "comp_append") && e.Type == EdgeCondTrue {
				hasFilterToAppend = true
			}
			if strings.Contains(e.From.Label, "comp_filter") && strings.Contains(e.To.Label, "comp_header") && e.Type == EdgeCondFalse {
				hasFilterToHeader = true
			}
			if strings.Contains(e.From.Label, "comp_append") && strings.Contains(e.To.Label, "comp_header") && e.Type == EdgeLoop {
				hasAppendToHeader = true
			}
			if strings.Contains(e.From.Label, "comp_header") && strings.Contains(e.To.Label, "comp_exit") && e.Type == EdgeCondFalse {
				hasHeaderToExit = true
			}
			return true
		},
		onBlock: func(b *BasicBlock) bool { return true },
	})

	assert.True(t, hasHeaderToBody, "Missing EdgeCondTrue from header to body")
	assert.True(t, hasBodyToFilter, "Missing EdgeNormal from body to filter")
	assert.True(t, hasFilterToAppend, "Missing EdgeCondTrue from filter to append")
	assert.True(t, hasFilterToHeader, "Missing EdgeCondFalse from filter back to header")
	assert.True(t, hasAppendToHeader, "Missing EdgeLoop from append back to header")
	assert.True(t, hasHeaderToExit, "Missing EdgeCondFalse from header to exit")
}

// TestWalrusOperatorWithComprehension tests support for the walrus operator (:=) assigning a comprehension
func TestWalrusOperatorWithComprehension(t *testing.T) {
	code := `
		def walrus_with_comp():
			(x := [i for i in range(10)])
			return x
	`
	funcName := "walrus_with_comp"

	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(code))
	require.NoError(t, err)

	builder := NewCFGBuilder()
	cfgs, err := builder.BuildAll(result.AST)
	require.NoError(t, err)

	cfg, exists := cfgs[funcName]
	require.True(t, exists, "Function CFG should exist")
	assert.NotNil(t, cfg)

	// This one assigns the result of a comprehension to x.
	// (x := [...])
	// This is an expression statement where the expression is a named expr.

	foundWalrus := false
	for _, block := range cfg.Blocks {
		for _, stmt := range block.Statements {
			stmt.Walk(func(n *parser.Node) bool {
				if n.Type == parser.NodeNamedExpr {
					foundWalrus = true
					return false
				}
				return true
			})
			if foundWalrus {
				break
			}
		}
		if foundWalrus {
			break
		}
	}
	assert.True(t, foundWalrus, "Should find walrus operator")
}
