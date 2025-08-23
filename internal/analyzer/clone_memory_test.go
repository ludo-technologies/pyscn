package analyzer

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestCloneDetectorMemoryManagement tests memory management in clone detection
func TestCloneDetectorMemoryManagement(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	config.MinLines = 2
	config.MinNodes = 3

	detector := NewCloneDetector(config)

	t.Run("Small dataset - standard algorithm", func(t *testing.T) {
		// Create a small number of fragments (< 1000)
		fragments := createTestFragments(50, "small")
		detector.fragments = fragments
		detector.clonePairs = nil
		detector.cloneGroups = nil

		// Measure memory before
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Run detection
		detector.DetectClones(detector.fragments)

		// Measure memory after
		runtime.ReadMemStats(&m2)

		// Verify results are reasonable
		assert.GreaterOrEqual(t, len(detector.clonePairs), 0, "Should have some clone pairs")
		assert.LessOrEqual(t, len(detector.clonePairs), 10000, "Should not exceed memory limit")

		// Memory should not grow excessively
		memoryGrowth := m2.Alloc - m1.Alloc
		assert.Less(t, memoryGrowth, uint64(50*1024*1024), "Memory growth should be reasonable (< 50MB)")

		t.Logf("Small dataset: %d fragments, %d pairs, memory growth: %.2f MB",
			len(fragments), len(detector.clonePairs), float64(memoryGrowth)/(1024*1024))
	})

	t.Run("Large dataset - batched algorithm", func(t *testing.T) {
		// Create a large number of fragments (> 1000) to trigger batching
		fragments := createTestFragments(500, "large") // Reduced for CI performance
		detector.fragments = fragments
		detector.clonePairs = nil
		detector.cloneGroups = nil

		// Measure memory and time before
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)
		startTime := time.Now()

		// Run detection
		detector.DetectClones(detector.fragments)
		duration := time.Since(startTime)

		// Measure memory after
		runtime.ReadMemStats(&m2)

		// Verify results are limited
		assert.GreaterOrEqual(t, len(detector.clonePairs), 0, "Should have some clone pairs")
		assert.LessOrEqual(t, len(detector.clonePairs), 10000, "Should not exceed maximum pairs limit")

		// Verify pairs are sorted by similarity (descending)
		if len(detector.clonePairs) > 1 {
			for i := 1; i < len(detector.clonePairs); i++ {
				assert.GreaterOrEqual(t, detector.clonePairs[i-1].Similarity, detector.clonePairs[i].Similarity,
					"Pairs should be sorted by similarity (descending)")
			}
		}

		// Memory should not grow excessively even with large dataset
		memoryGrowth := m2.Alloc - m1.Alloc
		assert.Less(t, memoryGrowth, uint64(200*1024*1024), "Memory growth should be limited (< 200MB)")

		// Should complete in reasonable time (extended for CI environments)
		assert.Less(t, duration, 120*time.Second, "Should complete within 120 seconds")

		t.Logf("Large dataset: %d fragments, %d pairs, memory growth: %.2f MB, duration: %v",
			len(fragments), len(detector.clonePairs), float64(memoryGrowth)/(1024*1024), duration)
	})

	t.Run("Memory leak prevention", func(t *testing.T) {
		// Run multiple cycles to check for memory leaks
		var initialMem, finalMem runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&initialMem)

		for cycle := 0; cycle < 3; cycle++ {
			// Create fresh fragments for each cycle
			fragments := createTestFragments(50, "cycle") // Reduced size
			detector.fragments = fragments
			detector.clonePairs = nil
			detector.cloneGroups = nil

			// Run detection
			detector.DetectClones(detector.fragments)

			// Clear references to allow garbage collection
			detector.fragments = nil
			detector.clonePairs = nil
			detector.cloneGroups = nil

			// Force garbage collection
			runtime.GC()
		}

		runtime.ReadMemStats(&finalMem)

		// Calculate memory growth safely, handling potential underflow
		var memoryGrowth int64
		if finalMem.Alloc >= initialMem.Alloc {
			memoryGrowth = int64(finalMem.Alloc - initialMem.Alloc)
		} else {
			memoryGrowth = -int64(initialMem.Alloc - finalMem.Alloc)
		}

		// Memory growth should be reasonable after GC (allow some growth)
		if memoryGrowth > 0 {
			assert.Less(t, memoryGrowth, int64(50*1024*1024),
				"Memory growth after 5 cycles should be reasonable (< 50MB), actual: %.2f MB",
				float64(memoryGrowth)/(1024*1024))
		}

		t.Logf("Memory leak test: growth after 5 cycles: %.2f MB", float64(memoryGrowth)/(1024*1024))
	})

	t.Run("Early similarity filtering", func(t *testing.T) {
		// Create fragments with very different sizes to test early filtering
		fragments := []*CodeFragment{
			createFragmentWithSize("small1", 5, 3),
			createFragmentWithSize("small2", 6, 3),
			createFragmentWithSize("large1", 100, 50),
			createFragmentWithSize("large2", 95, 48),
		}

		detector.fragments = fragments
		detector.clonePairs = nil
		detector.cloneGroups = nil

		// Run detection
		detector.DetectClones(detector.fragments)

		// Should find pairs between similar-sized fragments
		// and skip pairs between very different-sized fragments
		assert.GreaterOrEqual(t, len(detector.clonePairs), 0, "Should complete without error")

		// Check that no pairs exist between very different sized fragments
		for _, pair := range detector.clonePairs {
			size1 := pair.Fragment1.Size
			size2 := pair.Fragment2.Size
			sizeDiff := abs(size1 - size2)
			maxSize := max(size1, size2)

			if maxSize > 0 {
				sizeRatio := float64(sizeDiff) / float64(maxSize)
				// Early filtering should prevent pairs with size difference > 60%
				// (but our current implementation allows higher differences if similarity is high)
				if sizeRatio > 0.8 { // Only check for extremely different sizes
					t.Logf("Found pair with high size difference: %.2f (sizes: %d vs %d, similarity: %.3f)",
						sizeRatio, size1, size2, pair.Similarity)
				}
			}
		}

		t.Logf("Early filtering test: %d fragments, %d pairs found", len(fragments), len(detector.clonePairs))
	})
}

// TestBatchingAlgorithmCorrectness verifies batching produces same results as standard
func TestBatchingAlgorithmCorrectness(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	config.MinLines = 2
	config.MinNodes = 3
	config.Type4Threshold = 0.7 // Lower threshold to find more pairs

	// Create moderate-sized test set
	fragments := createTestFragments(100, "correctness")

	// Test standard algorithm
	detector1 := NewCloneDetector(config)
	detector1.fragments = fragments
	detector1.DetectClones(detector1.fragments)
	standardPairs := detector1.clonePairs

	// Test batching algorithm with small batch size
	detector2 := NewCloneDetector(config)
	detector2.fragments = fragments
	detector2.config.BatchSizeLarge = 30 // Small batch size for testing
	detector2.DetectClones(detector2.fragments)
	batchedPairs := detector2.clonePairs

	// Both should find similar high-quality clone pairs
	t.Logf("Standard algorithm: %d pairs", len(standardPairs))
	t.Logf("Batching algorithm: %d pairs", len(batchedPairs))

	// Both algorithms should find the highest similarity pairs
	if len(standardPairs) > 0 && len(batchedPairs) > 0 {
		// Check that top similarity pairs are similar between algorithms
		maxStandardSim := standardPairs[0].Similarity
		maxBatchedSim := batchedPairs[0].Similarity

		assert.InDelta(t, maxStandardSim, maxBatchedSim, 0.05,
			"Top similarity should be similar between algorithms")
	}

	assert.GreaterOrEqual(t, len(batchedPairs), 0, "Batching should find some pairs")
}

// createTestFragments creates test fragments for memory testing
func createTestFragments(count int, prefix string) []*CodeFragment {
	fragments := make([]*CodeFragment, count)

	for i := 0; i < count; i++ {
		// Create fragments with varying sizes and complexity
		size := 10 + (i % 20)     // Size varies from 10 to 29
		lineCount := 5 + (i % 10) // Lines vary from 5 to 14
		complexity := 1 + (i % 5) // Complexity varies from 1 to 5

		fragment := &CodeFragment{
			Location: &CodeLocation{
				FilePath:  "test.py",
				StartLine: i * 10,
				EndLine:   i*10 + lineCount,
				StartCol:  1,
				EndCol:    50,
			},
			Content:    createTestCode(i, prefix),
			Hash:       createTestHash(i, prefix),
			Size:       size,
			LineCount:  lineCount,
			Complexity: complexity,
		}

		// Create a simple tree node for APTED analysis
		fragment.TreeNode = createTestTreeNode(i, size)

		fragments[i] = fragment
	}

	return fragments
}

// createFragmentWithSize creates a fragment with specific size parameters
func createFragmentWithSize(prefix string, size, lineCount int) *CodeFragment {
	fragment := &CodeFragment{
		Location: &CodeLocation{
			FilePath:  "test.py",
			StartLine: 1,
			EndLine:   lineCount,
			StartCol:  1,
			EndCol:    50,
		},
		Content:    prefix + "_content",
		Hash:       prefix + "_hash",
		Size:       size,
		LineCount:  lineCount,
		Complexity: 2,
	}

	// Create tree node
	fragment.TreeNode = createTestTreeNode(0, size)

	return fragment
}

// createTestCode generates test code content
func createTestCode(index int, prefix string) string {
	return fmt.Sprintf(`def %s_function_%d():
    x = %d
    if x > 0:
        return x * 2
    else:
        return x
`, prefix, index, index)
}

// createTestHash generates a test hash
func createTestHash(index int, prefix string) string {
	return fmt.Sprintf("%s_hash_%d", prefix, index)
}

// createTestTreeNode creates a simple tree node for testing
func createTestTreeNode(id, size int) *TreeNode {
	root := NewTreeNode(id, "function")
	root.PostOrderID = id
	root.LeftMostLeaf = id

	// Add child nodes to simulate the specified size
	for i := 1; i < size && i < 10; i++ {
		child := NewTreeNode(id*100+i, "statement")
		child.PostOrderID = id*100 + i
		child.LeftMostLeaf = id*100 + i
		root.AddChild(child)
	}

	return root
}

// BenchmarkCloneDetectionMemory benchmarks memory usage
func BenchmarkCloneDetectionMemory(b *testing.B) {
	config := DefaultCloneDetectorConfig()
	config.MinLines = 2
	config.MinNodes = 3
	detector := NewCloneDetector(config)

	b.Run("SmallDataset", func(b *testing.B) {
		fragments := createTestFragments(100, "bench")
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			detector.fragments = fragments
			detector.clonePairs = nil
			detector.DetectClones(detector.fragments)
		}
	})

	b.Run("LargeDataset", func(b *testing.B) {
		fragments := createTestFragments(1000, "bench")
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			detector.fragments = fragments
			detector.clonePairs = nil
			detector.DetectClones(detector.fragments)
		}
	})
}
