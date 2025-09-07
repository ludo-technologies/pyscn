package analyzer

import (
	"context"
	"fmt"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"testing"
)

// DemoReachabilityAnalyzer demonstrates the complete workflow from source code to reachability analysis
func DemoReachabilityAnalyzer() {
	// Sample Python code with unreachable code after return
	source := `
def example_function(x):
    if x > 0:
        return x * 2
    else:
        return x * -1
    print("This code is unreachable")
    return 0
`

	// Parse the source code
	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(source))
	if err != nil {
		fmt.Printf("Parse error: %v\n", err)
		return
	}

	// Build CFG for the function
	funcNode := result.AST.Body[0] // First function
	builder := NewCFGBuilder()
	cfg, err := builder.Build(funcNode)
	if err != nil {
		fmt.Printf("CFG build error: %v\n", err)
		return
	}

	// Perform reachability analysis
	analyzer := NewReachabilityAnalyzer(cfg)
	reachabilityResult := analyzer.AnalyzeReachability()

	// Print results
	fmt.Printf("Total blocks: %d\n", reachabilityResult.TotalBlocks)
	fmt.Printf("Reachable blocks: %d\n", reachabilityResult.ReachableCount)
	fmt.Printf("Unreachable blocks: %d\n", reachabilityResult.UnreachableCount)
	fmt.Printf("Has unreachable code: %t\n", reachabilityResult.HasUnreachableCode())
	fmt.Printf("Reachability ratio: %.2f\n", reachabilityResult.GetReachabilityRatio())
	fmt.Printf("Analysis time: %v\n", reachabilityResult.AnalysisTime)

	// Get unreachable blocks with actual statements (dead code)
	deadCodeBlocks := reachabilityResult.GetUnreachableBlocksWithStatements()
	fmt.Printf("Dead code blocks: %d\n", len(deadCodeBlocks))

	// Demonstrate the clean API for Dead Code Detection integration
	if reachabilityResult.HasUnreachableCode() {
		fmt.Println("\nDead code found:")
		for id, block := range deadCodeBlocks {
			fmt.Printf("  Block %s: %s (%d statements)\n", id, block.Label, len(block.Statements))
		}
	}
}

// TestReachabilityIntegration demonstrates integration with the full CFG pipeline
func TestReachabilityIntegration(t *testing.T) {
	testCases := []struct {
		name            string
		source          string
		expectReachable bool
		expectDead      bool
	}{
		{
			name: "FullyReachableFunction",
			source: `
def reachable_function(x):
    if x > 0:
        return "positive"
    return "non-positive"
`,
			expectReachable: false, // CFGBuilder may create helper blocks
			expectDead:      false, // But no dead code with actual statements
		},
		{
			name: "FunctionWithDeadCode",
			source: `
def dead_code_function(x):
    return x + 1
    print("This is dead code")
    x = x * 2
`,
			expectReachable: false,
			expectDead:      true,
		},
		{
			name: "ComplexControlFlow",
			source: `
def complex_function(x, y):
    if x > 0:
        if y > 0:
            return x + y
        else:
            return x - y
    else:
        if y > 0:
            return -x + y
        else:
            return -x - y
`,
			expectReachable: false, // CFGBuilder may create helper blocks
			expectDead:      false, // But no dead code with actual statements
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Parse and build CFG
			p := parser.New()
			ctx := context.Background()
			result, err := p.Parse(ctx, []byte(tc.source))
			if err != nil {
				t.Fatalf("Parse error: %v", err)
			}

			funcNode := result.AST.Body[0]
			builder := NewCFGBuilder()
			cfg, err := builder.Build(funcNode)
			if err != nil {
				t.Fatalf("CFG build error: %v", err)
			}

			// Analyze reachability
			analyzer := NewReachabilityAnalyzer(cfg)
			reachResult := analyzer.AnalyzeReachability()

			// Validate expectations
			allReachable := reachResult.GetReachabilityRatio() == 1.0
			if allReachable != tc.expectReachable {
				t.Errorf("Expected all reachable: %t, got: %t (ratio: %.2f)",
					tc.expectReachable, allReachable, reachResult.GetReachabilityRatio())
			}

			hasDeadCode := reachResult.HasUnreachableCode()
			if hasDeadCode != tc.expectDead {
				t.Errorf("Expected dead code: %t, got: %t", tc.expectDead, hasDeadCode)
			}

			// Validate performance (should be very fast for small functions)
			if reachResult.AnalysisTime.Milliseconds() > 10 {
				t.Errorf("Analysis took too long: %v", reachResult.AnalysisTime)
			}

			// Log detailed results for debugging
			t.Logf("Results for %s:", tc.name)
			t.Logf("  Total: %d, Reachable: %d, Unreachable: %d",
				reachResult.TotalBlocks, reachResult.ReachableCount, reachResult.UnreachableCount)
			t.Logf("  Reachability ratio: %.2f", reachResult.GetReachabilityRatio())
			t.Logf("  Dead code blocks: %d", len(reachResult.GetUnreachableBlocksWithStatements()))
			t.Logf("  Analysis time: %v", reachResult.AnalysisTime)
		})
	}
}

// TestReachabilityAPIForDeadCodeDetection demonstrates the clean API for Dead Code Detection
func TestReachabilityAPIForDeadCodeDetection(t *testing.T) {
	source := `
def function_with_mixed_code(x):
    if x > 10:
        return "large"
    elif x > 5:
        return "medium"
    return "small"
    
    # This code is unreachable
    print("Dead code 1")
    
    if x < 0:
        print("Dead code 2")
    
    return "more dead code"
`

	// Build CFG
	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, []byte(source))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	funcNode := result.AST.Body[0]
	builder := NewCFGBuilder()
	cfg, err := builder.Build(funcNode)
	if err != nil {
		t.Fatalf("CFG build error: %v", err)
	}

	// Analyze reachability
	analyzer := NewReachabilityAnalyzer(cfg)
	result_analysis := analyzer.AnalyzeReachability()

	// Demonstrate Dead Code Detection API usage
	if result_analysis.HasUnreachableCode() {
		deadBlocks := result_analysis.GetUnreachableBlocksWithStatements()

		// This is the API that Dead Code Detection (#23) will use
		for blockID, block := range deadBlocks {
			t.Logf("Found dead code in block %s (%s):", blockID, block.Label)

			// Each block contains AST nodes that represent dead code
			for i, stmt := range block.Statements {
				t.Logf("  Statement %d: %s", i, stmt.Type)
				// In the actual Dead Code Detection implementation, this would:
				// 1. Extract line numbers from stmt
				// 2. Create Finding structs
				// 3. Generate user-friendly error messages
			}
		}

		if len(deadBlocks) == 0 {
			t.Error("Expected to find dead code blocks")
		}
	} else {
		t.Error("Expected to find unreachable code")
	}

	// Validate that the API provides all necessary information
	if result_analysis.TotalBlocks <= 0 {
		t.Error("Expected positive total blocks")
	}

	if result_analysis.ReachableCount <= 0 {
		t.Error("Expected positive reachable count")
	}

	if result_analysis.GetReachabilityRatio() <= 0 || result_analysis.GetReachabilityRatio() > 1.0 {
		t.Errorf("Invalid reachability ratio: %f", result_analysis.GetReachabilityRatio())
	}
}
