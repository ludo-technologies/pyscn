package analyzer

import (
	"context"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeadCodeDetection(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		expectedDead   int
		deadBlockNames []string
	}{
		{
			name: "UnreachableAfterReturn",
			code: `
def foo():
    x = 1
    return x
    y = 2  # dead code
    z = 3  # dead code
`,
			expectedDead:   1,
			deadBlockNames: []string{"unreachable"},
		},
		{
			name: "UnreachableAfterBreak",
			code: `
def foo():
    while True:
        x = 1
        break
        y = 2  # dead code
    z = 3  # reachable
`,
			expectedDead:   1,
			deadBlockNames: []string{"unreachable"},
		},
		{
			name: "UnreachableAfterContinue",
			code: `
def foo():
    for i in range(10):
        if i > 5:
            continue
            x = 1  # dead code
        y = 2  # reachable
`,
			expectedDead:   1,
			deadBlockNames: []string{"unreachable"},
		},
		{
			name: "DeadBranchInConditional",
			code: `
def foo():
    if False:
        x = 1  # potentially dead (static analysis could detect)
    else:
        y = 2  # reachable
    z = 3  # reachable
`,
			expectedDead: 0, // Without constant folding, both branches are considered reachable
		},
		{
			name: "UnreachableExceptionHandler",
			code: `
def foo():
    try:
        return 1
        raise ValueError()  # dead code
    except ValueError:
        x = 2  # potentially reachable (exception could be raised before return)
    finally:
        y = 3  # always reachable
`,
			expectedDead:   1,
			deadBlockNames: []string{"unreachable"},
		},
		{
			name: "NestedUnreachableCode",
			code: `
def foo():
    if True:
        return 1
        if False:  # dead code
            x = 2  # dead code
        else:
            y = 3  # dead code
    z = 4  # dead code
`,
			expectedDead: 3, // Nested if creates 3 blocks with statements: if condition, then branch, else branch
		},
		{
			name: "UnreachableInLoop",
			code: `
def foo():
    for i in range(10):
        if i == 5:
            return i
            print("never")  # dead code
        else:
            continue
            print("also never")  # dead code
    print("maybe")  # reachable if loop completes
`,
			expectedDead: 2,
		},
		{
			name: "ComplexControlFlow",
			code: `
def complex():
    try:
        for i in range(10):
            if i > 5:
                break
                x = 1  # dead
            elif i < 0:
                return -1
                y = 2  # dead
            else:
                continue
                z = 3  # dead
        return 0
        unreachable = True  # dead
    except:
        pass
    finally:
        cleanup = True  # reachable
`,
			expectedDead: 4, // Fixed elif handling now detects unreachable code after exhaustive if-elif-else
		},
		{
			name: "EmptyFunction",
			code: `
def empty():
    pass
`,
			expectedDead: 0,
		},
		{
			name: "SingleStatementFunction",
			code: `
def single():
    return 42
`,
			expectedDead: 0,
		},
		{
			name: "UnreachableElif",
			code: `
def foo(x):
    if x > 0:
        return "positive"
    elif x < 0:
        return "negative"
    elif x == 0:
        return "zero"
    else:
        return "impossible"  # dead if x is numeric
    print("never")  # dead
`,
			expectedDead: 1, // Detects unreachable code after exhaustive if-elif-else
		},
		{
			name: "ExhaustiveIfElseReturn",
			code: `
def foo(x):
    if x > 0:
        return "positive"
    else:
        return "negative"
    print("unreachable")
`,
			expectedDead:   1,
			deadBlockNames: []string{"unreachable"},
		},
		{
			name: "NestedExhaustiveReturn",
			code: `
def bar(x, y):
    if x > 0:
        if y > 0:
            return "both positive"
        else:
            return "x positive, y not"
    else:
        return "x not positive"
    print("unreachable")
`,
			expectedDead:   1,
			deadBlockNames: []string{"unreachable"},
		},
		{
			name: "MixedTerminators",
			code: `
def baz(x):
    if x > 0:
        return "positive"
    elif x < 0:
        raise ValueError("negative")
    else:
        return "zero"
    print("unreachable")
`,
			expectedDead:   1,
			deadBlockNames: []string{"unreachable"},
		},
		{
			name: "UnreachableInWith",
			code: `
def foo():
    with open("file.txt") as f:
        data = f.read()
        return data
        process(data)  # dead
    cleanup()  # potentially reachable if with raises
`,
			expectedDead: 1,
		},
		{
			name: "UnreachableAfterRaise",
			code: `
def foo():
    if True:
        raise ValueError("error")
        x = 1  # dead
    y = 2  # dead (if condition is always true)
`,
			expectedDead: 1,
		},
		{
			name: "MultipleReturnsWithDead",
			code: `
def foo(x):
    if x > 0:
        return 1
        dead1 = True  # dead
    if x < 0:
        return -1
        dead2 = True  # dead
    return 0
    dead3 = True  # dead
`,
			expectedDead: 3,
		},
		{
			name: "InfiniteLoopWithUnreachable",
			code: `
def foo():
    while True:
        x = 1
        if x == 1:
            continue
            y = 2  # dead
        break
        z = 3  # dead
    after = 4  # reachable via break
`,
			expectedDead: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the code
			p := parser.New()
			ctx := context.Background()
			result, err := p.Parse(ctx, []byte(tt.code))
			require.NoError(t, err, "Failed to parse code")
			ast := result.AST

			// Build all CFGs (module and functions)
			builder := NewCFGBuilder()
			cfgs, err := builder.BuildAll(ast)
			require.NoError(t, err, "Failed to build CFGs")

			// For these tests, we're testing function-level dead code
			// Find the first function CFG (not the module CFG)
			var cfg *CFG
			for name, c := range cfgs {
				if name != domain.ModuleFunctionName {
					cfg = c
					break
				}
			}
			require.NotNil(t, cfg, "No function CFG found")

			// Analyze reachability
			analyzer := NewReachabilityAnalyzer(cfg)
			reachResult := analyzer.AnalyzeReachability()

			// Count dead blocks (excluding entry/exit)
			deadCount := 0
			for _, block := range cfg.Blocks {
				_, isReachable := reachResult.ReachableBlocks[block.ID]
				if !isReachable &&
					block != cfg.Entry &&
					block != cfg.Exit &&
					len(block.Statements) > 0 {
					deadCount++
				}
			}

			// Check dead code count
			assert.Equal(t, tt.expectedDead, deadCount,
				"Expected %d dead blocks, got %d", tt.expectedDead, deadCount)

			// If specific block names are expected, verify them
			if len(tt.deadBlockNames) > 0 {
				var actualDeadNames []string
				for _, block := range cfg.Blocks {
					_, isReachable := reachResult.ReachableBlocks[block.ID]
					if !isReachable &&
						block != cfg.Entry &&
						block != cfg.Exit {
						// Use Label if available, otherwise use ID
						name := block.Label
						if name == "" {
							name = block.ID
						}
						actualDeadNames = append(actualDeadNames, name)
					}
				}

				// Check that expected dead blocks are present
				for _, expectedName := range tt.deadBlockNames {
					found := false
					for _, actualName := range actualDeadNames {
						if strings.Contains(actualName, expectedName) {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected dead block '%s' not found", expectedName)
				}
			}
		})
	}
}

func TestDeadCodeWithStatements(t *testing.T) {
	code := `
def function_with_dead_code():
    x = 1
    if x > 0:
        return x
        y = 2  # This is dead code
        z = 3  # This is also dead code
    return 0
    w = 4  # This is dead code
`

	// Parse the code
	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(code))
	require.NoError(t, err)
	ast := result.AST

	// Build all CFGs and get the function CFG
	builder := NewCFGBuilder()
	cfgs, err := builder.BuildAll(ast)
	require.NoError(t, err)

	// Get the function CFG
	cfg, exists := cfgs["function_with_dead_code"]
	require.True(t, exists, "Failed to find function_with_dead_code CFG")

	// Analyze reachability
	analyzer := NewReachabilityAnalyzer(cfg)
	reachResult := analyzer.AnalyzeReachability()

	// Get unreachable blocks with statements
	unreachableBlocks := reachResult.GetUnreachableBlocksWithStatements()

	// Should have at least 2 unreachable blocks with statements
	assert.GreaterOrEqual(t, len(unreachableBlocks), 2,
		"Expected at least 2 unreachable blocks with statements")

	// Check that HasUnreachableCode returns true
	assert.True(t, reachResult.HasUnreachableCode(),
		"HasUnreachableCode should return true for code with dead blocks")
}

func TestDeadCodeInExceptionHandling(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		expectDead  bool
		description string
	}{
		{
			name: "DeadAfterReturnInTry",
			code: `
def foo():
    try:
        x = 1
        return x
        y = 2  # dead
    except:
        z = 3
    finally:
        cleanup()
`,
			expectDead:  true,
			description: "Code after return in try block should be dead",
		},
		{
			name: "DeadAfterRaiseInTry",
			code: `
def foo():
    try:
        raise ValueError()
        x = 1  # dead
    except ValueError:
        y = 2
`,
			expectDead:  true,
			description: "Code after raise should be dead",
		},
		{
			name: "DeadInUnreachableExcept",
			code: `
def foo():
    try:
        return 1
    except ValueError:
        x = 2  # potentially reachable (conservative analysis)
    except TypeError:
        y = 3  # potentially reachable (conservative analysis)
`,
			expectDead:  false,
			description: "Exception handlers are conservatively considered reachable",
		},
		{
			name: "DeadAfterReturnInExcept",
			code: `
def foo():
    try:
        raise ValueError()
    except ValueError:
        return 1
        x = 2  # dead
`,
			expectDead:  true,
			description: "Code after return in except block should be dead",
		},
		{
			name: "FinallyAlwaysReachable",
			code: `
def foo():
    try:
        return 1
    finally:
        x = 2  # reachable
`,
			expectDead:  false,
			description: "Finally block should always be reachable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the code
			p := parser.New()
			ctx := context.Background()
			result, err := p.Parse(ctx, []byte(tt.code))
			require.NoError(t, err, "Failed to parse code: %s", tt.description)
			ast := result.AST

			// Build all CFGs (module and functions)
			builder := NewCFGBuilder()
			cfgs, err := builder.BuildAll(ast)
			require.NoError(t, err, "Failed to build CFGs")

			// For these tests, we're testing function-level dead code
			// Find the first function CFG (not the module CFG)
			var cfg *CFG
			for name, c := range cfgs {
				if name != domain.ModuleFunctionName {
					cfg = c
					break
				}
			}
			require.NotNil(t, cfg, "No function CFG found")

			// Analyze reachability
			analyzer := NewReachabilityAnalyzer(cfg)
			reachResult := analyzer.AnalyzeReachability()

			// Check if there's dead code
			hasDead := reachResult.HasUnreachableCode()
			assert.Equal(t, tt.expectDead, hasDead, tt.description)
		})
	}
}

func TestDeadCodeEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		code       string
		shouldPass bool
	}{
		{
			name:       "EmptyModule",
			code:       ``,
			shouldPass: true,
		},
		{
			name: "OnlyComments",
			code: `
# This is a comment
# Another comment
`,
			shouldPass: true,
		},
		{
			name: "DeeplyNestedDead",
			code: `
def deep():
    if True:
        if True:
            if True:
                return 1
                if True:  # dead
                    if True:  # dead
                        x = 1  # dead
`,
			shouldPass: true,
		},
		{
			name: "AsyncFunctionDead",
			code: `
async def async_func():
    await something()
    return 1
    x = 2  # dead
`,
			shouldPass: true,
		},
		{
			name: "GeneratorDead",
			code: `
def gen():
    yield 1
    return
    yield 2  # dead
`,
			shouldPass: true,
		},
		{
			name: "LambdaDead",
			code: `
lambda x: (1, 2, 3 if False else 4)
`,
			shouldPass: true,
		},
		{
			name: "ClassMethodDead",
			code: `
class MyClass:
    def method(self):
        return self.value
        self.other = 1  # dead
`,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the code
			p := parser.New()
			ctx := context.Background()
			result, err := p.Parse(ctx, []byte(tt.code))

			if !tt.shouldPass {
				assert.Error(t, err, "Expected parsing to fail")
				return
			}

			require.NoError(t, err, "Failed to parse code")
			ast := result.AST

			// Build all CFGs (module and functions)
			builder := NewCFGBuilder()
			cfgs, err := builder.BuildAll(ast)
			require.NoError(t, err, "Failed to build CFGs")

			// For these tests, we're testing function-level dead code
			// Find the first function CFG (not the module CFG)
			var cfg *CFG
			for name, c := range cfgs {
				if name != domain.ModuleFunctionName {
					cfg = c
					break
				}
			}

			// If no function CFG found, use the module CFG (for empty/comments-only tests)
			if cfg == nil {
				cfg = cfgs[domain.ModuleFunctionName]
				require.NotNil(t, cfg, "No CFG found at all")
			}

			// Analyze reachability
			analyzer := NewReachabilityAnalyzer(cfg)
			reachResult := analyzer.AnalyzeReachability()

			// Just ensure analysis completes without panic
			_ = reachResult.HasUnreachableCode()
		})
	}
}

func TestDeadCodeMetrics(t *testing.T) {
	code := `
def complex_function():
    x = 1
    if x > 0:
        y = 2
        if y > 1:
            return x + y
            dead1 = 1  # dead
        dead2 = 2  # potentially reachable
    else:
        z = 3
        return z
        dead3 = 3  # dead
    return 0
    dead4 = 4  # dead
    dead5 = 5  # dead
`

	// Parse the code
	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(code))
	require.NoError(t, err)
	ast := result.AST

	// Build all CFGs to get the function CFG
	builder := NewCFGBuilder()
	cfgs, err := builder.BuildAll(ast)
	require.NoError(t, err)

	// Get the complex_function CFG
	cfg, exists := cfgs["complex_function"]
	require.True(t, exists, "Failed to find complex_function CFG")

	// Analyze reachability
	analyzer := NewReachabilityAnalyzer(cfg)
	reachResult := analyzer.AnalyzeReachability()

	// Get reachability ratio
	ratio := reachResult.GetReachabilityRatio()
	assert.Greater(t, ratio, 0.0, "Reachability ratio should be greater than 0")
	assert.LessOrEqual(t, ratio, 1.0, "Reachability ratio should be at most 1")

	// Get unreachable blocks
	unreachableBlocks := reachResult.GetUnreachableBlocksWithStatements()
	assert.Greater(t, len(unreachableBlocks), 0, "Should have unreachable blocks")

	// Verify metrics consistency
	totalBlocks := len(cfg.Blocks)
	reachableCount := len(reachResult.ReachableBlocks)
	expectedRatio := float64(reachableCount) / float64(totalBlocks)
	assert.InDelta(t, expectedRatio, ratio, 0.01, "Ratio calculation should be consistent")
}

// TestDeadCodeSkipsTrailingSemicolonNoise ensures the dead-code detector does
// not emit `unreachable_branch` findings for the empty statement produced by a
// trailing `;` after a terminating statement (raise/return). Regression test
// for https://github.com/ludo-technologies/pyscn/issues/499.
func TestDeadCodeSkipsTrailingSemicolonNoise(t *testing.T) {
	code := `
def f(x):
    if x is None:
        raise ValueError("nope");
    return x


def g(y):
    return y;
`

	p := parser.New()
	ctx := context.Background()
	parseResult, err := p.Parse(ctx, []byte(code))
	require.NoError(t, err)

	builder := NewCFGBuilder()
	cfgs, err := builder.BuildAll(parseResult.AST)
	require.NoError(t, err)

	for _, fnName := range []string{"f", "g"} {
		cfg, ok := cfgs[fnName]
		require.True(t, ok, "expected CFG for %s", fnName)

		result := DetectInFunction(cfg)
		require.NotNil(t, result)

		for _, finding := range result.Findings {
			assert.NotEqual(t, ";", finding.Code,
				"function %s: dead-code finding for a bare ';' (trailing-semicolon empty statement) should be filtered out; got %+v",
				fnName, finding)
		}
	}
}

func TestDeadCodeNoOverlappingFindings(t *testing.T) {
	// A compound statement (if) spans its body, so the body's own block used to
	// emit a finding whose range was nested inside the if's finding range,
	// double-counting the same source line. The whole region after `raise` is
	// one contiguous dead region and should produce non-overlapping findings.
	code := `
def do_flip(self, show_widgets=True):
    raise NotImplementedError
    self.widget_tooltip = None
    if show_widgets:
        self.render_widgets()
    if self.notifications:
        self.render_notifications()
`

	p := parser.New()
	ctx := context.Background()
	parseResult, err := p.Parse(ctx, []byte(code))
	require.NoError(t, err)

	builder := NewCFGBuilder()
	cfgs, err := builder.BuildAll(parseResult.AST)
	require.NoError(t, err)

	cfg, ok := cfgs["do_flip"]
	require.True(t, ok, "expected CFG for do_flip")

	result := DetectInFunction(cfg)
	require.NotNil(t, result)
	require.NotEmpty(t, result.Findings)

	// Findings must not have overlapping line ranges: each finding's start line
	// must be strictly after the previous finding's end line.
	for i := 1; i < len(result.Findings); i++ {
		prev := result.Findings[i-1]
		cur := result.Findings[i]
		assert.Greater(t, cur.StartLine, prev.EndLine,
			"findings overlap: %+v then %+v", prev, cur)
	}

	// The contiguous dead region collapses to a single finding.
	assert.Len(t, result.Findings, 1, "contiguous dead region should be one finding")
}

func TestMergeContiguousFindings(t *testing.T) {
	mk := func(start, end int, reason DeadCodeReason, sev SeverityLevel) *DeadCodeFinding {
		return &DeadCodeFinding{StartLine: start, EndLine: end, Reason: reason, Severity: sev, Code: "x"}
	}

	t.Run("merges overlapping same-reason findings", func(t *testing.T) {
		in := []*DeadCodeFinding{
			mk(4, 6, ReasonUnreachableAfterRaise, SeverityLevelWarning),
			mk(6, 6, ReasonUnreachableAfterRaise, SeverityLevelCritical),
		}
		out := mergeContiguousFindings(in)
		require.Len(t, out, 1)
		assert.Equal(t, 4, out[0].StartLine)
		assert.Equal(t, 6, out[0].EndLine)
		assert.Equal(t, SeverityLevelCritical, out[0].Severity, "keeps highest severity")
	})

	t.Run("merges adjacent same-reason findings", func(t *testing.T) {
		in := []*DeadCodeFinding{
			mk(4, 6, ReasonUnreachableAfterRaise, SeverityLevelCritical),
			mk(7, 8, ReasonUnreachableAfterRaise, SeverityLevelCritical),
		}
		out := mergeContiguousFindings(in)
		require.Len(t, out, 1)
		assert.Equal(t, 8, out[0].EndLine)
	})

	t.Run("does not merge across a gap", func(t *testing.T) {
		in := []*DeadCodeFinding{
			mk(4, 4, ReasonUnreachableAfterRaise, SeverityLevelCritical),
			mk(6, 6, ReasonUnreachableAfterRaise, SeverityLevelCritical),
		}
		out := mergeContiguousFindings(in)
		assert.Len(t, out, 2)
	})

	t.Run("does not merge different reasons", func(t *testing.T) {
		in := []*DeadCodeFinding{
			mk(4, 6, ReasonUnreachableAfterRaise, SeverityLevelCritical),
			mk(6, 6, ReasonUnreachableBranch, SeverityLevelWarning),
		}
		out := mergeContiguousFindings(in)
		assert.Len(t, out, 2)
	})
}

func TestIsOnlyNoOpStatements(t *testing.T) {
	assert.False(t, isOnlyNoOpStatements(nil), "nil block")
	assert.False(t, isOnlyNoOpStatements(&BasicBlock{}), "empty block")

	semi := &parser.Node{Type: parser.NodeType(";")}
	ret := &parser.Node{Type: parser.NodeReturn}

	pass := &parser.Node{Type: parser.NodePass}

	assert.True(t, isOnlyNoOpStatements(&BasicBlock{Statements: []any{semi}}),
		"block with only a separator")
	assert.True(t, isOnlyNoOpStatements(&BasicBlock{Statements: []any{semi, semi}}),
		"block with multiple separators")
	assert.True(t, isOnlyNoOpStatements(&BasicBlock{Statements: []any{pass}}),
		"block with only pass")
	assert.False(t, isOnlyNoOpStatements(&BasicBlock{Statements: []any{ret}}),
		"block with a real statement")
	assert.False(t, isOnlyNoOpStatements(&BasicBlock{Statements: []any{semi, ret}}),
		"block mixing separators and a real statement")
}
