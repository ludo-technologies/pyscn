package analyzer

import (
	"fmt"
	"github.com/pyqol/pyqol/internal/parser"
	"math/rand"
	"testing"
)

// BenchmarkReachabilityAnalysis benchmarks the reachability analysis performance
func BenchmarkReachabilityAnalysis(b *testing.B) {
	benchmarkSizes := []int{10, 50, 100, 500, 1000, 2500, 5000}

	for _, size := range benchmarkSizes {
		b.Run(fmt.Sprintf("LinearChain_%d", size), func(b *testing.B) {
			cfg := createLinearCFG(size)
			analyzer := NewReachabilityAnalyzer(cfg)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := analyzer.AnalyzeReachability()
				if result.ReachableCount != size+2 { // +2 for entry and exit
					b.Errorf("Expected %d reachable blocks, got %d", size+2, result.ReachableCount)
				}
			}
		})

		b.Run(fmt.Sprintf("BinaryTree_%d", size), func(b *testing.B) {
			cfg := createBinaryTreeCFG(size)
			analyzer := NewReachabilityAnalyzer(cfg)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := analyzer.AnalyzeReachability()
				// All blocks in binary tree should be reachable
				if result.UnreachableCount != 0 {
					b.Errorf("Expected 0 unreachable blocks, got %d", result.UnreachableCount)
				}
			}
		})

		b.Run(fmt.Sprintf("ComplexGraph_%d", size), func(b *testing.B) {
			cfg := createComplexCFG(size)
			analyzer := NewReachabilityAnalyzer(cfg)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := analyzer.AnalyzeReachability()
				// Should complete without errors
				if result.TotalBlocks <= 0 {
					b.Error("Expected positive total blocks")
				}
			}
		})

		b.Run(fmt.Sprintf("WithUnreachable_%d", size), func(b *testing.B) {
			cfg := createCFGWithUnreachableBlocks(size)
			analyzer := NewReachabilityAnalyzer(cfg)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := analyzer.AnalyzeReachability()
				// Should have some unreachable blocks
				if result.UnreachableCount == 0 {
					b.Error("Expected some unreachable blocks")
				}
			}
		})
	}
}

// BenchmarkLargeCFG benchmarks very large CFG analysis
func BenchmarkLargeCFG(b *testing.B) {
	sizes := []int{10000, 25000, 50000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Linear_%d", size), func(b *testing.B) {
			cfg := createLinearCFG(size)
			analyzer := NewReachabilityAnalyzer(cfg)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := analyzer.AnalyzeReachability()

				// Validate performance target: >10k blocks/second
				blocksPerSecond := float64(result.TotalBlocks) / result.AnalysisTime.Seconds()
				if blocksPerSecond < 10000 {
					b.Logf("Performance: %.0f blocks/second (target: >10,000)", blocksPerSecond)
				}
			}
		})
	}
}

// BenchmarkMemoryUsage benchmarks memory efficiency
func BenchmarkMemoryUsage(b *testing.B) {
	sizes := []int{1000, 5000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("MemoryEfficiency_%d", size), func(b *testing.B) {
			cfg := createComplexCFG(size)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				analyzer := NewReachabilityAnalyzer(cfg)
				result := analyzer.AnalyzeReachability()

				// Force evaluation to ensure all data structures are populated
				_ = result.GetUnreachableBlocksWithStatements()
				_ = result.GetReachabilityRatio()
				_ = result.HasUnreachableCode()
			}
		})
	}
}

// BenchmarkAnalyzeFrom benchmarks analysis from specific starting blocks
func BenchmarkAnalyzeFrom(b *testing.B) {
	size := 1000
	cfg := createComplexCFG(size)
	analyzer := NewReachabilityAnalyzer(cfg)

	// Get a random block from the middle of the CFG
	var testBlock *BasicBlock
	count := 0
	for _, block := range cfg.Blocks {
		if count == size/2 {
			testBlock = block
			break
		}
		count++
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := analyzer.AnalyzeReachabilityFrom(testBlock)
		if result.TotalBlocks != len(cfg.Blocks) {
			b.Error("Total blocks mismatch")
		}
	}
}

// Helper functions to create test CFGs

// createLinearCFG creates a linear chain of blocks: entry -> b1 -> b2 -> ... -> bn -> exit
func createLinearCFG(size int) *CFG {
	cfg := NewCFG("linear_benchmark")
	cfg.Entry.AddSuccessor(cfg.Exit, EdgeNormal)

	if size == 0 {
		return cfg
	}

	// Create linear chain
	prev := cfg.Entry
	for i := 0; i < size; i++ {
		block := cfg.CreateBlock(fmt.Sprintf("block_%d", i))
		// Add a dummy statement to make it meaningful
		dummyNode := &parser.Node{Type: "ExpressionStatement"}
		block.AddStatement(dummyNode)

		// Remove direct entry->exit edge first time
		if i == 0 {
			prev.Successors = []*Edge{} // Remove direct edge to exit
			cfg.Exit.Predecessors = []*Edge{}
		}

		prev.AddSuccessor(block, EdgeNormal)
		prev = block
	}

	// Connect last block to exit
	prev.AddSuccessor(cfg.Exit, EdgeNormal)

	return cfg
}

// createBinaryTreeCFG creates a binary tree structure
func createBinaryTreeCFG(maxBlocks int) *CFG {
	cfg := NewCFG("binary_tree_benchmark")

	blocks := []*BasicBlock{cfg.Entry}
	created := 1

	for len(blocks) > 0 && created < maxBlocks {
		current := blocks[0]
		blocks = blocks[1:]

		// Create left and right children
		if created < maxBlocks {
			left := cfg.CreateBlock(fmt.Sprintf("left_%d", created))
			dummyNode := &parser.Node{Type: "ExpressionStatement"}
			left.AddStatement(dummyNode)
			current.AddSuccessor(left, EdgeCondTrue)
			blocks = append(blocks, left)
			created++
		}

		if created < maxBlocks {
			right := cfg.CreateBlock(fmt.Sprintf("right_%d", created))
			dummyNode := &parser.Node{Type: "ExpressionStatement"}
			right.AddStatement(dummyNode)
			current.AddSuccessor(right, EdgeCondFalse)
			blocks = append(blocks, right)
			created++
		}
	}

	// Connect leaf nodes to exit
	for _, block := range cfg.Blocks {
		if len(block.Successors) == 0 && !block.IsExit {
			block.AddSuccessor(cfg.Exit, EdgeNormal)
		}
	}

	return cfg
}

// createComplexCFG creates a complex CFG with loops, conditions, and multiple paths
func createComplexCFG(size int) *CFG {
	cfg := NewCFG("complex_benchmark")

	if size <= 0 {
		cfg.Entry.AddSuccessor(cfg.Exit, EdgeNormal)
		return cfg
	}

	rng := rand.New(rand.NewSource(42)) // Consistent random generation
	blocks := make([]*BasicBlock, 0, size)

	// Create blocks
	for i := 0; i < size; i++ {
		block := cfg.CreateBlock(fmt.Sprintf("block_%d", i))
		dummyNode := &parser.Node{Type: "ExpressionStatement"}
		block.AddStatement(dummyNode)
		blocks = append(blocks, block)
	}

	// Connect entry to first block
	if len(blocks) > 0 {
		cfg.Entry.AddSuccessor(blocks[0], EdgeNormal)
	}

	// Create complex connections
	for i, block := range blocks {
		numSuccessors := rng.Intn(3) + 1 // 1-3 successors

		for j := 0; j < numSuccessors && j < 3; j++ {
			var target *BasicBlock

			if i == len(blocks)-1 {
				// Last block connects to exit
				target = cfg.Exit
			} else {
				// Random forward connection (avoid creating too many cycles)
				maxTarget := min(i+10, len(blocks)-1)
				if maxTarget > i {
					targetIdx := i + 1 + rng.Intn(maxTarget-i)
					target = blocks[targetIdx]
				} else {
					target = cfg.Exit
				}
			}

			// Add edge with varied types
			edgeType := EdgeNormal
			switch j {
			case 1:
				edgeType = EdgeCondTrue
			case 2:
				edgeType = EdgeCondFalse
			}

			// Check if edge already exists
			exists := false
			for _, existingEdge := range block.Successors {
				if existingEdge.To == target {
					exists = true
					break
				}
			}

			if !exists {
				block.AddSuccessor(target, edgeType)
			}
		}

		// Occasionally create a loop back
		if i > 5 && rng.Float64() < 0.1 {
			backTarget := blocks[rng.Intn(i)]
			// Check if back edge already exists
			exists := false
			for _, existingEdge := range block.Successors {
				if existingEdge.To == backTarget {
					exists = true
					break
				}
			}
			if !exists {
				block.AddSuccessor(backTarget, EdgeLoop)
			}
		}
	}

	return cfg
}

// createCFGWithUnreachableBlocks creates a CFG with intentionally unreachable blocks
func createCFGWithUnreachableBlocks(size int) *CFG {
	cfg := NewCFG("unreachable_benchmark")

	if size <= 0 {
		cfg.Entry.AddSuccessor(cfg.Exit, EdgeNormal)
		return cfg
	}

	// Create reachable chain (70% of blocks)
	reachableSize := (size * 7) / 10
	unreachableSize := size - reachableSize

	// Build reachable part
	prev := cfg.Entry
	for i := 0; i < reachableSize; i++ {
		block := cfg.CreateBlock(fmt.Sprintf("reachable_%d", i))
		dummyNode := &parser.Node{Type: "ExpressionStatement"}
		block.AddStatement(dummyNode)
		prev.AddSuccessor(block, EdgeNormal)
		prev = block
	}
	prev.AddSuccessor(cfg.Exit, EdgeNormal)

	// Build unreachable part (disconnected)
	var unreachablePrev *BasicBlock
	for i := 0; i < unreachableSize; i++ {
		block := cfg.CreateBlock(fmt.Sprintf("unreachable_%d", i))
		dummyNode := &parser.Node{Type: "ExpressionStatement"}
		block.AddStatement(dummyNode)

		if unreachablePrev != nil {
			unreachablePrev.AddSuccessor(block, EdgeNormal)
		}
		unreachablePrev = block
	}

	return cfg
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
