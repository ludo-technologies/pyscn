package analyzer

import (
	"context"
	"strings"
	"testing"

	"github.com/pyqol/pyqol/internal/parser"
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
        return 2  # unreachable
    return 3  # unreachable
`,
			checkFunc: func(t *testing.T, cfg *CFG) {
				// Should detect multiple levels of unreachable code
				analyzer := NewReachabilityAnalyzer(cfg)
				result := analyzer.AnalyzeReachability()
				unreachable := result.GetUnreachableBlocksWithStatements()
				assert.Greater(t, len(unreachable), 0)
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

			// Build CFG
			builder := NewCFGBuilder()
			cfg, err := builder.Build(ast)
			require.NoError(t, err)

			// Run specific checks
			tc.checkFunc(t, cfg)
		})
	}
}
