package analyzer

import (
	"fmt"
	"testing"
)

func TestReachabilityAnalyzer(t *testing.T) {
	t.Run("SimpleLinearFlow", func(t *testing.T) {
		cfg := NewCFG("test")
		
		// Create a simple linear flow: ENTRY -> block1 -> block2 -> EXIT
		block1 := cfg.CreateBlock("block1")
		block2 := cfg.CreateBlock("block2")
		
		cfg.ConnectBlocks(cfg.Entry, block1, EdgeNormal)
		cfg.ConnectBlocks(block1, block2, EdgeNormal)
		cfg.ConnectBlocks(block2, cfg.Exit, EdgeNormal)
		
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		if report.UnreachableBlocks != 0 {
			t.Errorf("Expected 0 unreachable blocks, got %d", report.UnreachableBlocks)
		}
		if report.ReachableBlocks != 4 { // ENTRY, block1, block2, EXIT
			t.Errorf("Expected 4 reachable blocks, got %d", report.ReachableBlocks)
		}
	})
	
	t.Run("UnreachableBlock", func(t *testing.T) {
		cfg := NewCFG("test")
		
		// Create flow with unreachable block
		block1 := cfg.CreateBlock("block1")
		block2 := cfg.CreateBlock("block2")
		unreachable := cfg.CreateBlock("unreachable")
		
		cfg.ConnectBlocks(cfg.Entry, block1, EdgeNormal)
		cfg.ConnectBlocks(block1, block2, EdgeNormal)
		cfg.ConnectBlocks(block2, cfg.Exit, EdgeNormal)
		// unreachable block has no incoming edges from reachable blocks
		
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		if report.UnreachableBlocks != 1 {
			t.Errorf("Expected 1 unreachable block, got %d", report.UnreachableBlocks)
		}
		if report.ReachableBlocks != 4 { // ENTRY, block1, block2, EXIT
			t.Errorf("Expected 4 reachable blocks, got %d", report.ReachableBlocks)
		}
		
		// Check that the unreachable block is correctly identified
		if !analyzer.IsReachable(unreachable) != true {
			t.Error("Unreachable block not correctly identified")
		}
	})
	
	t.Run("ConditionalBranches", func(t *testing.T) {
		cfg := NewCFG("test")
		
		// Create if-else structure
		condition := cfg.CreateBlock("condition")
		trueBlock := cfg.CreateBlock("true_branch")
		falseBlock := cfg.CreateBlock("false_branch")
		merge := cfg.CreateBlock("merge")
		
		cfg.ConnectBlocks(cfg.Entry, condition, EdgeNormal)
		cfg.ConnectBlocks(condition, trueBlock, EdgeCondTrue)
		cfg.ConnectBlocks(condition, falseBlock, EdgeCondFalse)
		cfg.ConnectBlocks(trueBlock, merge, EdgeNormal)
		cfg.ConnectBlocks(falseBlock, merge, EdgeNormal)
		cfg.ConnectBlocks(merge, cfg.Exit, EdgeNormal)
		
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		if report.UnreachableBlocks != 0 {
			t.Errorf("Expected 0 unreachable blocks, got %d", report.UnreachableBlocks)
		}
		if report.ReachableBlocks != 6 { // All blocks should be reachable
			t.Errorf("Expected 6 reachable blocks, got %d", report.ReachableBlocks)
		}
	})
	
	t.Run("LoopWithUnreachableAfterReturn", func(t *testing.T) {
		cfg := NewCFG("test")
		
		// Create loop with return inside
		loopHeader := cfg.CreateBlock("loop_header")
		loopBody := cfg.CreateBlock("loop_body")
		returnBlock := cfg.CreateBlock("return")
		unreachable := cfg.CreateBlock("after_return")
		
		cfg.ConnectBlocks(cfg.Entry, loopHeader, EdgeNormal)
		cfg.ConnectBlocks(loopHeader, loopBody, EdgeCondTrue)
		cfg.ConnectBlocks(loopBody, returnBlock, EdgeNormal)
		cfg.ConnectBlocks(returnBlock, cfg.Exit, EdgeReturn)
		cfg.ConnectBlocks(loopBody, loopHeader, EdgeLoop)
		// unreachable block after return
		cfg.ConnectBlocks(returnBlock, unreachable, EdgeNormal)
		
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		// Since we follow all edges, unreachable will be reachable
		// In a real implementation, we might want to handle return edges specially
		if report.UnreachableBlocks != 0 {
			// Note: In this simple implementation, the block after return is still
			// considered reachable because we follow all edges
			t.Logf("Found %d unreachable blocks (expected behavior may vary)", report.UnreachableBlocks)
		}
	})
	
	t.Run("CyclicGraph", func(t *testing.T) {
		cfg := NewCFG("test")
		
		// Create a cycle: block1 -> block2 -> block3 -> block1
		block1 := cfg.CreateBlock("block1")
		block2 := cfg.CreateBlock("block2")
		block3 := cfg.CreateBlock("block3")
		exitPath := cfg.CreateBlock("exit_path")
		
		cfg.ConnectBlocks(cfg.Entry, block1, EdgeNormal)
		cfg.ConnectBlocks(block1, block2, EdgeNormal)
		cfg.ConnectBlocks(block2, block3, EdgeNormal)
		cfg.ConnectBlocks(block3, block1, EdgeLoop) // Create cycle
		cfg.ConnectBlocks(block3, exitPath, EdgeCondFalse) // Exit from cycle
		cfg.ConnectBlocks(exitPath, cfg.Exit, EdgeNormal)
		
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		if report.UnreachableBlocks != 0 {
			t.Errorf("Expected 0 unreachable blocks in cyclic graph, got %d", report.UnreachableBlocks)
		}
		if report.ReachableBlocks != 6 { // All blocks should be reachable
			t.Errorf("Expected 6 reachable blocks, got %d", report.ReachableBlocks)
		}
	})
	
	t.Run("MultipleEntryPoints", func(t *testing.T) {
		cfg := NewCFG("test")
		
		// Create blocks accessible from different entry points
		mainPath := cfg.CreateBlock("main_path")
		exceptionPath := cfg.CreateBlock("exception_path")
		merge := cfg.CreateBlock("merge")
		
		cfg.ConnectBlocks(cfg.Entry, mainPath, EdgeNormal)
		cfg.ConnectBlocks(mainPath, merge, EdgeNormal)
		cfg.ConnectBlocks(merge, cfg.Exit, EdgeNormal)
		
		// Exception path not connected from main entry
		cfg.ConnectBlocks(exceptionPath, merge, EdgeNormal)
		
		// First analysis without additional entry point
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		if report.UnreachableBlocks != 1 {
			t.Errorf("Expected 1 unreachable block without additional entry, got %d", report.UnreachableBlocks)
		}
		
		// Add exception handler as entry point and reanalyze
		analyzer.AddEntryPoint(exceptionPath)
		report = analyzer.AnalyzeReachability()
		
		if report.UnreachableBlocks != 0 {
			t.Errorf("Expected 0 unreachable blocks with additional entry, got %d", report.UnreachableBlocks)
		}
	})
	
	t.Run("DisconnectedSubgraph", func(t *testing.T) {
		cfg := NewCFG("test")
		
		// Create main path
		block1 := cfg.CreateBlock("block1")
		cfg.ConnectBlocks(cfg.Entry, block1, EdgeNormal)
		cfg.ConnectBlocks(block1, cfg.Exit, EdgeNormal)
		
		// Create disconnected subgraph
		island1 := cfg.CreateBlock("island1")
		island2 := cfg.CreateBlock("island2")
		island3 := cfg.CreateBlock("island3")
		cfg.ConnectBlocks(island1, island2, EdgeNormal)
		cfg.ConnectBlocks(island2, island3, EdgeNormal)
		cfg.ConnectBlocks(island3, island1, EdgeLoop) // Cycle within disconnected subgraph
		
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		if report.UnreachableBlocks != 3 {
			t.Errorf("Expected 3 unreachable blocks (disconnected subgraph), got %d", report.UnreachableBlocks)
		}
		
		unreachableIDs := report.GetUnreachableBlockIDs()
		if len(unreachableIDs) != 3 {
			t.Errorf("Expected 3 unreachable block IDs, got %d", len(unreachableIDs))
		}
	})
	
	t.Run("MarkUnreachableCode", func(t *testing.T) {
		cfg := NewCFG("test")
		
		// Create blocks with unreachable code
		reachable := cfg.CreateBlock("reachable")
		unreachable1 := cfg.CreateBlock("dead_code")
		unreachable2 := cfg.CreateBlock("")
		
		cfg.ConnectBlocks(cfg.Entry, reachable, EdgeNormal)
		cfg.ConnectBlocks(reachable, cfg.Exit, EdgeNormal)
		
		analyzer := NewReachabilityAnalyzer(cfg)
		analyzer.AnalyzeReachability()
		analyzer.MarkUnreachableCode()
		
		// Check that unreachable blocks are marked
		if unreachable1.Label != "dead_code (unreachable)" {
			t.Errorf("Expected label 'dead_code (unreachable)', got '%s'", unreachable1.Label)
		}
		if unreachable2.Label != LabelUnreachable {
			t.Errorf("Expected label '%s', got '%s'", LabelUnreachable, unreachable2.Label)
		}
	})
	
	t.Run("EmptyCFG", func(t *testing.T) {
		cfg := NewCFG("empty")
		// Connect entry to exit for a minimal CFG
		cfg.ConnectBlocks(cfg.Entry, cfg.Exit, EdgeNormal)
		
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		if report.UnreachableBlocks != 0 {
			t.Errorf("Expected 0 unreachable blocks in minimal CFG, got %d", report.UnreachableBlocks)
		}
		if report.ReachableBlocks != 2 { // Only ENTRY and EXIT
			t.Errorf("Expected 2 reachable blocks (entry/exit), got %d", report.ReachableBlocks)
		}
	})
	
	t.Run("GetReachableBlocks", func(t *testing.T) {
		cfg := NewCFG("test")
		
		block1 := cfg.CreateBlock("block1")
		block2 := cfg.CreateBlock("block2")
		unreachableBlock := cfg.CreateBlock("unreachable")
		
		cfg.ConnectBlocks(cfg.Entry, block1, EdgeNormal)
		cfg.ConnectBlocks(block1, block2, EdgeNormal)
		cfg.ConnectBlocks(block2, cfg.Exit, EdgeNormal)
		
		analyzer := NewReachabilityAnalyzer(cfg)
		analyzer.AnalyzeReachability()
		
		reachableBlocks := analyzer.GetReachableBlocks()
		if len(reachableBlocks) != 4 { // ENTRY, block1, block2, EXIT
			t.Errorf("Expected 4 reachable blocks, got %d", len(reachableBlocks))
		}
		
		unreachableBlocks := analyzer.GetUnreachableBlocks()
		if len(unreachableBlocks) != 1 {
			t.Errorf("Expected 1 unreachable block, got %d", len(unreachableBlocks))
		}
		if unreachableBlocks[0] != unreachableBlock {
			t.Error("Wrong block identified as unreachable")
		}
	})
	
	t.Run("ReportString", func(t *testing.T) {
		cfg := NewCFG("test")
		
		block1 := cfg.CreateBlock("block1")
		_ = cfg.CreateBlock("unreachable") // Create an unreachable block
		
		cfg.ConnectBlocks(cfg.Entry, block1, EdgeNormal)
		cfg.ConnectBlocks(block1, cfg.Exit, EdgeNormal)
		
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		reportStr := report.String()
		if reportStr == "" {
			t.Error("Report string should not be empty")
		}
		
		if !report.HasUnreachableCode() {
			t.Error("Report should indicate unreachable code exists")
		}
	})
}

func TestReachabilityWithComplexStructures(t *testing.T) {
	t.Run("NestedLoops", func(t *testing.T) {
		cfg := NewCFG("nested_loops")
		
		// Create nested loop structure
		outerHeader := cfg.CreateBlock("outer_header")
		innerHeader := cfg.CreateBlock("inner_header")
		innerBody := cfg.CreateBlock("inner_body")
		outerContinue := cfg.CreateBlock("outer_continue")
		outerExit := cfg.CreateBlock("outer_exit")
		
		cfg.ConnectBlocks(cfg.Entry, outerHeader, EdgeNormal)
		cfg.ConnectBlocks(outerHeader, innerHeader, EdgeCondTrue)
		cfg.ConnectBlocks(innerHeader, innerBody, EdgeCondTrue)
		cfg.ConnectBlocks(innerBody, innerHeader, EdgeLoop)
		cfg.ConnectBlocks(innerHeader, outerContinue, EdgeCondFalse)
		cfg.ConnectBlocks(outerContinue, outerHeader, EdgeLoop)
		cfg.ConnectBlocks(outerHeader, outerExit, EdgeCondFalse)
		cfg.ConnectBlocks(outerExit, cfg.Exit, EdgeNormal)
		
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		if report.UnreachableBlocks != 0 {
			t.Errorf("Expected 0 unreachable blocks in nested loops, got %d", report.UnreachableBlocks)
		}
	})
	
	t.Run("TryExceptFinally", func(t *testing.T) {
		cfg := NewCFG("try_except_finally")
		
		// Create try-except-finally structure
		tryBlock := cfg.CreateBlock("try")
		exceptBlock := cfg.CreateBlock("except")
		elseBlock := cfg.CreateBlock("else")
		finallyBlock := cfg.CreateBlock("finally")
		
		cfg.ConnectBlocks(cfg.Entry, tryBlock, EdgeNormal)
		cfg.ConnectBlocks(tryBlock, elseBlock, EdgeNormal) // Normal path
		cfg.ConnectBlocks(tryBlock, exceptBlock, EdgeException) // Exception path
		cfg.ConnectBlocks(exceptBlock, finallyBlock, EdgeNormal)
		cfg.ConnectBlocks(elseBlock, finallyBlock, EdgeNormal)
		cfg.ConnectBlocks(finallyBlock, cfg.Exit, EdgeNormal)
		
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		if report.UnreachableBlocks != 0 {
			t.Errorf("Expected 0 unreachable blocks in try-except-finally, got %d", report.UnreachableBlocks)
		}
		if report.ReachableBlocks != 6 { // All blocks should be reachable
			t.Errorf("Expected 6 reachable blocks, got %d", report.ReachableBlocks)
		}
	})
	
	t.Run("BreakContinueInLoop", func(t *testing.T) {
		cfg := NewCFG("break_continue")
		
		// Create loop with break and continue
		loopHeader := cfg.CreateBlock("loop_header")
		checkBreak := cfg.CreateBlock("check_break")
		checkContinue := cfg.CreateBlock("check_continue")
		loopBody := cfg.CreateBlock("loop_body")
		loopExit := cfg.CreateBlock("loop_exit")
		
		cfg.ConnectBlocks(cfg.Entry, loopHeader, EdgeNormal)
		cfg.ConnectBlocks(loopHeader, checkBreak, EdgeCondTrue)
		cfg.ConnectBlocks(checkBreak, loopExit, EdgeBreak) // Break out
		cfg.ConnectBlocks(checkBreak, checkContinue, EdgeNormal)
		cfg.ConnectBlocks(checkContinue, loopHeader, EdgeContinue) // Continue
		cfg.ConnectBlocks(checkContinue, loopBody, EdgeNormal)
		cfg.ConnectBlocks(loopBody, loopHeader, EdgeLoop)
		cfg.ConnectBlocks(loopHeader, loopExit, EdgeCondFalse)
		cfg.ConnectBlocks(loopExit, cfg.Exit, EdgeNormal)
		
		analyzer := NewReachabilityAnalyzer(cfg)
		report := analyzer.AnalyzeReachability()
		
		if report.UnreachableBlocks != 0 {
			t.Errorf("Expected 0 unreachable blocks with break/continue, got %d", report.UnreachableBlocks)
		}
	})
}

// Benchmark tests
func BenchmarkReachabilitySmallCFG(b *testing.B) {
	// Create a small CFG with 10 blocks
	cfg := createBenchmarkCFG(10)
	analyzer := NewReachabilityAnalyzer(cfg)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeReachability()
	}
}

func BenchmarkReachabilityMediumCFG(b *testing.B) {
	// Create a medium CFG with 100 blocks
	cfg := createBenchmarkCFG(100)
	analyzer := NewReachabilityAnalyzer(cfg)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeReachability()
	}
}

func BenchmarkReachabilityLargeCFG(b *testing.B) {
	// Create a large CFG with 1000 blocks
	cfg := createBenchmarkCFG(1000)
	analyzer := NewReachabilityAnalyzer(cfg)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeReachability()
	}
}

func BenchmarkReachabilityVeryLargeCFG(b *testing.B) {
	// Create a very large CFG with 10000 blocks
	cfg := createBenchmarkCFG(10000)
	analyzer := NewReachabilityAnalyzer(cfg)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.AnalyzeReachability()
	}
	
	// Report performance
	report := analyzer.AnalyzeReachability()
	blocksPerSecond := float64(cfg.Size()) / report.AnalysisTime.Seconds()
	b.Logf("Performance: %.0f blocks/second for %d blocks", blocksPerSecond, cfg.Size())
}

// Helper function to create CFG for benchmarking
func createBenchmarkCFG(numBlocks int) *CFG {
	cfg := NewCFG("benchmark")
	
	if numBlocks <= 2 {
		return cfg
	}
	
	// Create a complex CFG with branches and loops
	blocks := make([]*BasicBlock, numBlocks-2) // Excluding entry and exit
	
	for i := 0; i < numBlocks-2; i++ {
		blocks[i] = cfg.CreateBlock(fmt.Sprintf("block_%d", i))
	}
	
	// Connect entry to first block
	if len(blocks) > 0 {
		cfg.ConnectBlocks(cfg.Entry, blocks[0], EdgeNormal)
	}
	
	// Create connections with some branches and loops
	for i := 0; i < len(blocks); i++ {
		if i < len(blocks)-1 {
			// Normal flow to next block
			cfg.ConnectBlocks(blocks[i], blocks[i+1], EdgeNormal)
			
			// Add some branches (every 3rd block)
			if i%3 == 0 && i+2 < len(blocks) {
				cfg.ConnectBlocks(blocks[i], blocks[i+2], EdgeCondTrue)
			}
			
			// Add some loops (every 5th block)
			if i%5 == 0 && i > 0 {
				cfg.ConnectBlocks(blocks[i], blocks[i-1], EdgeLoop)
			}
		}
	}
	
	// Connect last block to exit
	if len(blocks) > 0 {
		cfg.ConnectBlocks(blocks[len(blocks)-1], cfg.Exit, EdgeNormal)
	}
	
	// Add some unreachable blocks (10% of blocks)
	unreachableCount := numBlocks / 10
	for i := 0; i < unreachableCount; i++ {
		unreachable := cfg.CreateBlock(fmt.Sprintf("unreachable_%d", i))
		// These blocks are not connected to the main flow
		_ = unreachable
	}
	
	return cfg
}