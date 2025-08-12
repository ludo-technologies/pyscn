package analyzer

import (
	"context"
	"github.com/pyqol/pyqol/internal/parser"
	"strings"
	"testing"
	"time"
)

func TestReachabilityAnalyzer(t *testing.T) {
	t.Run("EmptyCFG", func(t *testing.T) {
		cfg := NewCFG("empty")
		// Connect entry to exit for a minimal valid CFG
		cfg.Entry.AddSuccessor(cfg.Exit, EdgeNormal)

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		if result.TotalBlocks != 2 { // Entry and exit blocks
			t.Errorf("Expected 2 blocks (entry/exit), got %d", result.TotalBlocks)
		}

		if result.ReachableCount != 2 {
			t.Errorf("Expected 2 reachable blocks, got %d", result.ReachableCount)
		}

		if result.UnreachableCount != 0 {
			t.Errorf("Expected 0 unreachable blocks, got %d", result.UnreachableCount)
		}
	})

	t.Run("LinearFlow", func(t *testing.T) {
		cfg := NewCFG("linear")

		// Create linear flow: entry -> block1 -> block2 -> exit
		block1 := cfg.CreateBlock("block1")
		block2 := cfg.CreateBlock("block2")

		cfg.Entry.AddSuccessor(block1, EdgeNormal)
		block1.AddSuccessor(block2, EdgeNormal)
		block2.AddSuccessor(cfg.Exit, EdgeNormal)

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		if result.TotalBlocks != 4 {
			t.Errorf("Expected 4 blocks, got %d", result.TotalBlocks)
		}

		if result.ReachableCount != 4 {
			t.Errorf("Expected 4 reachable blocks, got %d", result.ReachableCount)
		}

		if result.UnreachableCount != 0 {
			t.Errorf("Expected 0 unreachable blocks, got %d", result.UnreachableCount)
		}

		// All blocks should be reachable
		expectedReachable := []string{"bb0", "bb1", "bb2", "bb3"} // entry, exit, block1, block2
		for _, id := range expectedReachable {
			if _, exists := result.ReachableBlocks[id]; !exists {
				t.Errorf("Block %s should be reachable", id)
			}
		}
	})

	t.Run("UnreachableCode", func(t *testing.T) {
		cfg := NewCFG("unreachable")

		// Create flow with unreachable code: entry -> block1 -> exit
		// block2 and block3 are disconnected (unreachable)
		block1 := cfg.CreateBlock("reachable")
		block2 := cfg.CreateBlock("unreachable1")
		block3 := cfg.CreateBlock("unreachable2")

		// Add statements to make blocks meaningful
		dummyNode := &parser.Node{Type: "ExpressionStatement"}
		block1.AddStatement(dummyNode)
		block2.AddStatement(dummyNode)
		block3.AddStatement(dummyNode)

		// Connect only reachable path
		cfg.Entry.AddSuccessor(block1, EdgeNormal)
		block1.AddSuccessor(cfg.Exit, EdgeNormal)

		// Leave block2 and block3 disconnected
		block2.AddSuccessor(block3, EdgeNormal)

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		if result.TotalBlocks != 5 {
			t.Errorf("Expected 5 blocks, got %d", result.TotalBlocks)
		}

		if result.ReachableCount != 3 {
			t.Errorf("Expected 3 reachable blocks, got %d", result.ReachableCount)
		}

		if result.UnreachableCount != 2 {
			t.Errorf("Expected 2 unreachable blocks, got %d", result.UnreachableCount)
		}

		// Check specific blocks
		if _, exists := result.ReachableBlocks[block1.ID]; !exists {
			t.Error("block1 should be reachable")
		}

		if _, exists := result.UnreachableBlocks[block2.ID]; !exists {
			t.Error("block2 should be unreachable")
		}

		if _, exists := result.UnreachableBlocks[block3.ID]; !exists {
			t.Error("block3 should be unreachable")
		}

		// Test utility methods
		if !result.HasUnreachableCode() {
			t.Error("Should detect unreachable code")
		}

		unreachableWithStmts := result.GetUnreachableBlocksWithStatements()
		if len(unreachableWithStmts) != 2 {
			t.Errorf("Expected 2 unreachable blocks with statements, got %d", len(unreachableWithStmts))
		}

		ratio := result.GetReachabilityRatio()
		expected := 3.0 / 5.0 // 3 reachable out of 5 total
		if ratio != expected {
			t.Errorf("Expected reachability ratio %f, got %f", expected, ratio)
		}
	})

	t.Run("ConditionalFlow", func(t *testing.T) {
		cfg := NewCFG("conditional")

		// Create conditional flow: entry -> condition -> (true_branch | false_branch) -> merge -> exit
		condition := cfg.CreateBlock("condition")
		trueBranch := cfg.CreateBlock("true_branch")
		falseBranch := cfg.CreateBlock("false_branch")
		merge := cfg.CreateBlock("merge")

		cfg.Entry.AddSuccessor(condition, EdgeNormal)
		condition.AddSuccessor(trueBranch, EdgeCondTrue)
		condition.AddSuccessor(falseBranch, EdgeCondFalse)
		trueBranch.AddSuccessor(merge, EdgeNormal)
		falseBranch.AddSuccessor(merge, EdgeNormal)
		merge.AddSuccessor(cfg.Exit, EdgeNormal)

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		if result.ReachableCount != 6 {
			t.Errorf("Expected 6 reachable blocks, got %d", result.ReachableCount)
		}

		if result.UnreachableCount != 0 {
			t.Errorf("Expected 0 unreachable blocks, got %d", result.UnreachableCount)
		}
	})

	t.Run("LoopFlow", func(t *testing.T) {
		cfg := NewCFG("loop")

		// Create loop flow: entry -> header -> (body -> header | exit)
		header := cfg.CreateBlock("loop_header")
		body := cfg.CreateBlock("loop_body")

		cfg.Entry.AddSuccessor(header, EdgeNormal)
		header.AddSuccessor(body, EdgeCondTrue)
		header.AddSuccessor(cfg.Exit, EdgeCondFalse)
		body.AddSuccessor(header, EdgeLoop)

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		if result.ReachableCount != 4 {
			t.Errorf("Expected 4 reachable blocks, got %d", result.ReachableCount)
		}

		if result.UnreachableCount != 0 {
			t.Errorf("Expected 0 unreachable blocks, got %d", result.UnreachableCount)
		}
	})

	t.Run("ComplexFlow", func(t *testing.T) {
		// Create a complex CFG with mixed reachable and unreachable blocks
		cfg := NewCFG("complex")

		// Main reachable path
		mainBlock := cfg.CreateBlock("main")
		conditionalBlock := cfg.CreateBlock("conditional")
		truePath := cfg.CreateBlock("true_path")
		falsePath := cfg.CreateBlock("false_path")
		mergeBlock := cfg.CreateBlock("merge")

		// Unreachable isolated blocks
		isolatedBlock1 := cfg.CreateBlock("isolated1")
		isolatedBlock2 := cfg.CreateBlock("isolated2")

		// Add statements to isolated blocks
		dummyNode := &parser.Node{Type: "ExpressionStatement"}
		isolatedBlock1.AddStatement(dummyNode)
		isolatedBlock2.AddStatement(dummyNode)

		// Connect main reachable path
		cfg.Entry.AddSuccessor(mainBlock, EdgeNormal)
		mainBlock.AddSuccessor(conditionalBlock, EdgeNormal)
		conditionalBlock.AddSuccessor(truePath, EdgeCondTrue)
		conditionalBlock.AddSuccessor(falsePath, EdgeCondFalse)
		truePath.AddSuccessor(mergeBlock, EdgeNormal)
		falsePath.AddSuccessor(mergeBlock, EdgeNormal)
		mergeBlock.AddSuccessor(cfg.Exit, EdgeNormal)

		// Leave isolated blocks disconnected
		isolatedBlock1.AddSuccessor(isolatedBlock2, EdgeNormal)

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		if result.TotalBlocks != 9 {
			t.Errorf("Expected 9 blocks, got %d", result.TotalBlocks)
		}

		if result.ReachableCount != 7 {
			t.Errorf("Expected 7 reachable blocks, got %d", result.ReachableCount)
		}

		if result.UnreachableCount != 2 {
			t.Errorf("Expected 2 unreachable blocks, got %d", result.UnreachableCount)
		}

		// Check that isolated blocks are unreachable
		if _, exists := result.UnreachableBlocks[isolatedBlock1.ID]; !exists {
			t.Error("isolatedBlock1 should be unreachable")
		}

		if _, exists := result.UnreachableBlocks[isolatedBlock2.ID]; !exists {
			t.Error("isolatedBlock2 should be unreachable")
		}
	})

	t.Run("AnalyzeReachabilityFrom", func(t *testing.T) {
		cfg := NewCFG("custom_start")

		block1 := cfg.CreateBlock("block1")
		block2 := cfg.CreateBlock("block2")
		block3 := cfg.CreateBlock("block3")

		// Create path: entry -> block1 -> block2
		// block3 is only reachable from block2
		cfg.Entry.AddSuccessor(block1, EdgeNormal)
		block1.AddSuccessor(block2, EdgeNormal)
		block2.AddSuccessor(block3, EdgeNormal)
		block3.AddSuccessor(cfg.Exit, EdgeNormal)

		analyzer := NewReachabilityAnalyzer(cfg)

		// Analyze from block2 - should only reach block3 and exit
		result := analyzer.AnalyzeReachabilityFrom(block2)

		if result.ReachableCount != 3 {
			t.Errorf("Expected 3 reachable blocks from block2, got %d", result.ReachableCount)
		}

		// Should reach block2, block3, and exit
		if _, exists := result.ReachableBlocks[block2.ID]; !exists {
			t.Error("block2 should be reachable from itself")
		}

		if _, exists := result.ReachableBlocks[block3.ID]; !exists {
			t.Error("block3 should be reachable from block2")
		}

		if _, exists := result.ReachableBlocks[cfg.Exit.ID]; !exists {
			t.Error("exit should be reachable from block2")
		}

		// Entry and block1 should be unreachable from block2
		if _, exists := result.UnreachableBlocks[cfg.Entry.ID]; !exists {
			t.Error("entry should be unreachable from block2")
		}

		if _, exists := result.UnreachableBlocks[block1.ID]; !exists {
			t.Error("block1 should be unreachable from block2")
		}
	})

	t.Run("EmptyBlocksNotConsideredUnreachableCode", func(t *testing.T) {
		cfg := NewCFG("empty_blocks")

		reachableBlock := cfg.CreateBlock("reachable")
		unreachableBlock := cfg.CreateBlock("unreachable")

		// Add statement only to reachable block
		dummyNode := &parser.Node{Type: "ExpressionStatement"}
		reachableBlock.AddStatement(dummyNode)
		// unreachableBlock remains empty
		_ = unreachableBlock // Suppress unused variable warning

		cfg.Entry.AddSuccessor(reachableBlock, EdgeNormal)
		reachableBlock.AddSuccessor(cfg.Exit, EdgeNormal)
		// unreachableBlock is disconnected

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		if result.UnreachableCount != 1 {
			t.Errorf("Expected 1 unreachable block, got %d", result.UnreachableCount)
		}

		// Empty unreachable blocks should not be considered "unreachable code"
		unreachableWithStmts := result.GetUnreachableBlocksWithStatements()
		if len(unreachableWithStmts) != 0 {
			t.Errorf("Expected 0 unreachable blocks with statements, got %d", len(unreachableWithStmts))
		}

		if result.HasUnreachableCode() {
			t.Error("Should not report unreachable code for empty blocks")
		}
	})

	t.Run("PerformanceTracking", func(t *testing.T) {
		cfg := NewCFG("performance")

		// Create a simple chain of blocks
		prev := cfg.Entry
		for i := 0; i < 100; i++ {
			block := cfg.CreateBlock("block")
			prev.AddSuccessor(block, EdgeNormal)
			prev = block
		}
		prev.AddSuccessor(cfg.Exit, EdgeNormal)

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		// Should complete quickly
		if result.AnalysisTime > time.Millisecond*100 {
			t.Errorf("Analysis took too long: %v", result.AnalysisTime)
		}

		if result.ReachableCount != 102 { // entry + exit + 100 blocks
			t.Errorf("Expected 102 reachable blocks, got %d", result.ReachableCount)
		}
	})
}

func TestReachabilityWithRealCFG(t *testing.T) {
	t.Run("SimpleFunction", func(t *testing.T) {
		source := `
def simple_function(x):
    if x > 0:
        return x * 2
    else:
        return x * -1
`
		ast := parseSourceForReachability(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		// Should have a reasonable number of reachable blocks
		if result.ReachableCount < 4 { // At least entry, condition, branches, exit
			t.Errorf("Expected at least 4 reachable blocks, got %d", result.ReachableCount)
		}

		// The function may have some empty unreachable blocks - this is normal CFGBuilder behavior
		// The key is that there should be no unreachable code (blocks with statements)
		if result.HasUnreachableCode() {
			t.Error("Simple function should not have unreachable code with statements")
		}
	})

	t.Run("FunctionWithUnreachableCode", func(t *testing.T) {
		source := `
def function_with_dead_code(x):
    if x > 0:
        return x * 2
    print("This should be reachable")
    return x * -1
`
		ast := parseSourceForReachability(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		// This function has all code reachable, but CFGBuilder may create empty helper blocks
		// The key is that there should be no unreachable code with actual statements
		if result.HasUnreachableCode() {
			t.Error("Function should not have unreachable code with statements")
		}
	})

	t.Run("FunctionWithRealDeadCode", func(t *testing.T) {
		source := `
def function_with_real_dead_code(x):
    return x * 2
    print("This is unreachable")  # Should be detected as unreachable
    x += 1
`
		ast := parseSourceForReachability(t, source)
		funcNode := ast.Body[0]

		builder := NewCFGBuilder()
		cfg, err := builder.Build(funcNode)
		if err != nil {
			t.Fatalf("Failed to build CFG: %v", err)
		}

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		// Should detect unreachable code after return
		if result.UnreachableCount == 0 {
			t.Error("Expected to find unreachable blocks after return statement")
		}

		if !result.HasUnreachableCode() {
			t.Error("Should detect unreachable code after return")
		}

		unreachableWithStmts := result.GetUnreachableBlocksWithStatements()
		if len(unreachableWithStmts) == 0 {
			t.Error("Should find unreachable blocks with statements")
		}
	})
}

func TestReachabilityEdgeCases(t *testing.T) {
	t.Run("NilCFG", func(t *testing.T) {
		analyzer := NewReachabilityAnalyzer(nil)
		// Should not panic
		result := analyzer.AnalyzeReachability()

		if result.TotalBlocks != 0 {
			t.Errorf("Expected 0 blocks for nil CFG, got %d", result.TotalBlocks)
		}
	})

	t.Run("CFGWithNilEntry", func(t *testing.T) {
		cfg := &CFG{
			Entry:  nil,
			Exit:   nil,
			Blocks: make(map[string]*BasicBlock),
		}

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		if result.ReachableCount != 0 {
			t.Errorf("Expected 0 reachable blocks with nil entry, got %d", result.ReachableCount)
		}
	})

	t.Run("AnalyzeFromNilBlock", func(t *testing.T) {
		cfg := NewCFG("test")
		analyzer := NewReachabilityAnalyzer(cfg)

		result := analyzer.AnalyzeReachabilityFrom(nil)

		if result.ReachableCount != 0 {
			t.Errorf("Expected 0 reachable blocks from nil, got %d", result.ReachableCount)
		}
	})

	t.Run("SingleBlockCFG", func(t *testing.T) {
		cfg := NewCFG("single")
		// Just entry and exit, connected
		cfg.Entry.AddSuccessor(cfg.Exit, EdgeNormal)

		analyzer := NewReachabilityAnalyzer(cfg)
		result := analyzer.AnalyzeReachability()

		if result.ReachableCount != 2 {
			t.Errorf("Expected 2 reachable blocks, got %d", result.ReachableCount)
		}

		if result.UnreachableCount != 0 {
			t.Errorf("Expected 0 unreachable blocks, got %d", result.UnreachableCount)
		}

		if result.GetReachabilityRatio() != 1.0 {
			t.Errorf("Expected 100%% reachability, got %f", result.GetReachabilityRatio())
		}
	})
}

// Helper function to parse source code for testing (reachability specific)
func parseSourceForReachability(t *testing.T, source string) *parser.Node {
	p := parser.New()
	ctx := context.Background()

	result, err := p.Parse(ctx, []byte(source))
	if err != nil {
		t.Fatalf("Failed to parse source: %v", err)
	}

	return result.AST
}

// Helper to compare strings in error messages
func containsString(slice []string, str string) bool {
	for _, s := range slice {
		if strings.Contains(s, str) {
			return true
		}
	}
	return false
}

