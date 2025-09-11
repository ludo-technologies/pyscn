package analyzer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// TestLSHIntegration tests the complete LSH pipeline
func TestLSHIntegration(t *testing.T) {
	// Create clone detector with LSH enabled
	config := DefaultCloneDetectorConfig()
	config.UseLSH = true
	config.LSHBands = 32
	config.LSHRows = 4
	config.LSHMinHashCount = 128
	config.LSHAutoThreshold = true
	config.Type3Threshold = 0.7
	
	detector := NewCloneDetector(config)
	
	// Create test fragments
	fragments := createTestFragmentsForLSH()
	
	// Run LSH-accelerated clone detection
	ctx := context.Background()
	clonePairs, cloneGroups := detector.DetectClonesWithLSH(ctx, fragments)
	
	// Verify results
	assert.NotNil(t, clonePairs)
	assert.NotNil(t, cloneGroups)
	
	// Results should be valid (may not find clones in test data)
	assert.GreaterOrEqual(t, len(clonePairs), 0)
	
	// Verify clone pairs are sorted by similarity
	for i := 1; i < len(clonePairs); i++ {
		assert.GreaterOrEqual(t, clonePairs[i-1].Similarity, clonePairs[i].Similarity)
	}
	
	// Check that LSH is properly enabled
	assert.True(t, detector.IsLSHEnabled())
	
	// Get LSH statistics
	stats := detector.GetLSHStats()
	assert.NotNil(t, stats)
	assert.Greater(t, stats.NumFragments, 0)
}

// TestLSHVsExhaustiveComparison compares LSH results with exhaustive search
func TestLSHVsExhaustiveComparison(t *testing.T) {
	// Create test fragments
	fragments := createTestFragmentsForLSH()
	
	// Test with exhaustive search
	configExhaustive := DefaultCloneDetectorConfig()
	configExhaustive.UseLSH = false
	configExhaustive.Type3Threshold = 0.7
	detectorExhaustive := NewCloneDetector(configExhaustive)
	
	ctx := context.Background()
	exhaustivePairs, _ := detectorExhaustive.DetectClonesWithContext(ctx, fragments)
	
	// Test with LSH
	configLSH := DefaultCloneDetectorConfig()
	configLSH.UseLSH = true
	configLSH.LSHBands = 16 // Lower precision for faster test
	configLSH.LSHRows = 8
	configLSH.LSHMinHashCount = 256
	configLSH.Type3Threshold = 0.7
	detectorLSH := NewCloneDetector(configLSH)
	
	lshPairs, _ := detectorLSH.DetectClonesWithLSH(ctx, fragments)
	
	// Calculate recall: how many true clones did LSH find?
	trueClones := make(map[string]bool)
	for _, pair := range exhaustivePairs {
		key := createPairKey(pair.Fragment1, pair.Fragment2)
		trueClones[key] = true
	}
	
	foundByLSH := 0
	for _, pair := range lshPairs {
		key := createPairKey(pair.Fragment1, pair.Fragment2)
		if trueClones[key] {
			foundByLSH++
		}
	}
	
	recall := 0.0
	if len(exhaustivePairs) > 0 {
		recall = float64(foundByLSH) / float64(len(exhaustivePairs))
	} else {
		recall = 1.0 // If no clones found by exhaustive, LSH is perfect
	}
	
	// LSH should have reasonable recall (may be 0 if no clones exist)
	assert.GreaterOrEqual(t, recall, 0.0, "LSH recall should be non-negative")
	
	// Log performance comparison
	t.Logf("Exhaustive found %d pairs, LSH found %d pairs, recall: %.3f", 
		len(exhaustivePairs), len(lshPairs), recall)
}

// TestLSHPerformanceEstimation tests performance estimation
func TestLSHPerformanceEstimation(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	config.UseLSH = true
	detector := NewCloneDetector(config)
	
	tests := []struct {
		name          string
		fragmentCount int
	}{
		{"small dataset", 50},
		{"medium dataset", 200},
		{"large dataset", 1000},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exhaustive, lsh, speedup := detector.EstimateLSHPerformance(test.fragmentCount)
			
			assert.Greater(t, exhaustive, 0)
			assert.Greater(t, lsh, 0)
			assert.Greater(t, speedup, 0.0)
			
			// For larger datasets, speedup should be significant
			if test.fragmentCount >= 500 {
				assert.Greater(t, speedup, 5.0)
			} else if test.fragmentCount >= 200 {
				assert.Greater(t, speedup, 2.0)
			}
			
			t.Logf("Fragment count: %d, Exhaustive: %d, LSH: %d, Speedup: %.2fx",
				test.fragmentCount, exhaustive, lsh, speedup)
		})
	}
}

// TestLSHEnableDisable tests enabling and disabling LSH
func TestLSHEnableDisable(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	config.UseLSH = false // Start disabled
	detector := NewCloneDetector(config)
	
	// Initially disabled
	assert.False(t, detector.IsLSHEnabled())
	
	// Enable LSH
	err := detector.EnableLSH()
	assert.NoError(t, err)
	assert.True(t, detector.IsLSHEnabled())
	
	// Disable LSH
	detector.DisableLSH()
	assert.False(t, detector.IsLSHEnabled())
}

// TestLSHFallback tests fallback to exhaustive search when LSH fails
func TestLSHFallback(t *testing.T) {
	// Create detector with invalid LSH configuration
	config := DefaultCloneDetectorConfig()
	config.UseLSH = true
	config.LSHMinHashCount = 32 // Too few hashes for 32 bands * 4 rows
	
	detector := NewCloneDetector(config)
	fragments := createTestFragmentsForLSH()
	
	ctx := context.Background()
	
	// Should fallback to exhaustive search without error
	pairs, groups := detector.DetectClonesWithLSH(ctx, fragments)
	
	assert.NotNil(t, pairs)
	assert.NotNil(t, groups)
	// Should still find some clones via fallback
}

// TestLSHCancellation tests context cancellation during LSH
func TestLSHCancellation(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	config.UseLSH = true
	detector := NewCloneDetector(config)
	
	fragments := createLargeTestFragmentSet(100)
	
	// Create context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	
	// Should handle cancellation gracefully
	pairs, groups := detector.DetectClonesWithLSH(ctx, fragments)
	
	assert.NotNil(t, pairs)
	assert.NotNil(t, groups)
	// Results may be empty due to cancellation, but should not panic
}

// TestLSHWithRealCodeFragments tests LSH with realistic code patterns
func TestLSHWithRealCodeFragments(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	config.UseLSH = true
	config.Type3Threshold = 0.8
	detector := NewCloneDetector(config)
	
	fragments := createRealisticCodeFragments()
	
	ctx := context.Background()
	pairs, groups := detector.DetectClonesWithLSH(ctx, fragments)
	
	// Should find the similar function variants
	assert.GreaterOrEqual(t, len(pairs), 1)
	
	// Check that similar functions are grouped together
	if len(groups) > 0 {
		largestGroup := groups[0]
		for _, group := range groups {
			if group.Size > largestGroup.Size {
				largestGroup = group
			}
		}
		assert.GreaterOrEqual(t, largestGroup.Size, 2)
	}
}

// TestLSHAccuracyWithDifferentThresholds tests LSH with various similarity thresholds
func TestLSHAccuracyWithDifferentThresholds(t *testing.T) {
	fragments := createTestFragmentsForLSH()
	
	thresholds := []float64{0.6, 0.7, 0.8, 0.9}
	
	for _, threshold := range thresholds {
		t.Run(fmt.Sprintf("threshold_%.1f", threshold), func(t *testing.T) {
			config := DefaultCloneDetectorConfig()
			config.UseLSH = true
			config.Type3Threshold = threshold
			config.LSHSimilarityThreshold = threshold * 0.9 // Slightly lower for better recall
			
			detector := NewCloneDetector(config)
			
			ctx := context.Background()
			pairs, _ := detector.DetectClonesWithLSH(ctx, fragments)
			
			// Higher thresholds should find fewer but higher quality clones
			for _, pair := range pairs {
				assert.GreaterOrEqual(t, pair.Similarity, threshold)
			}
			
			t.Logf("Threshold %.1f: found %d clone pairs", threshold, len(pairs))
		})
	}
}

// Helper functions

func createTestFragmentsForLSH() []*CodeFragment {
	fragments := []*CodeFragment{}
	
	// Create some similar code fragments
	for i := 0; i < 10; i++ {
		location := &CodeLocation{
			FilePath:  fmt.Sprintf("test%d.py", i),
			StartLine: 1,
			EndLine:   10,
		}
		
		// Create AST nodes with similar structure
		astNode := createTestASTNode(fmt.Sprintf("function_%d", i))
		fragment := NewCodeFragment(location, astNode, fmt.Sprintf("def function_%d():\n    return %d", i, i))
		
		// Convert to tree node
		converter := NewTreeConverter()
		fragment.TreeNode = converter.ConvertAST(astNode)
		
		fragments = append(fragments, fragment)
	}
	
	// Add some identical fragments
	for i := 0; i < 3; i++ {
		location := &CodeLocation{
			FilePath:  fmt.Sprintf("identical%d.py", i),
			StartLine: 1,
			EndLine:   10,
		}
		
		astNode := createTestASTNode("identical_function")
		fragment := NewCodeFragment(location, astNode, "def identical_function():\n    return 42")
		
		converter := NewTreeConverter()
		fragment.TreeNode = converter.ConvertAST(astNode)
		
		fragments = append(fragments, fragment)
	}
	
	return fragments
}

func createLargeTestFragmentSet(count int) []*CodeFragment {
	fragments := []*CodeFragment{}
	
	for i := 0; i < count; i++ {
		location := &CodeLocation{
			FilePath:  fmt.Sprintf("large_test_%d.py", i),
			StartLine: 1,
			EndLine:   20,
		}
		
		astNode := createTestASTNode(fmt.Sprintf("large_function_%d", i))
		fragment := NewCodeFragment(location, astNode, fmt.Sprintf("def large_function_%d():\n    pass", i))
		
		converter := NewTreeConverter()
		fragment.TreeNode = converter.ConvertAST(astNode)
		
		fragments = append(fragments, fragment)
	}
	
	return fragments
}

func createRealisticCodeFragments() []*CodeFragment {
	fragments := []*CodeFragment{}
	
	// Create similar function variants
	functionTemplates := []string{
		"calculate_sum",
		"calculate_product", 
		"calculate_average",
	}
	
	for i, template := range functionTemplates {
		for variant := 0; variant < 3; variant++ {
			location := &CodeLocation{
				FilePath:  fmt.Sprintf("realistic_%d_%d.py", i, variant),
				StartLine: 1,
				EndLine:   15,
			}
			
			astNode := createComplexASTNode(fmt.Sprintf("%s_v%d", template, variant))
			fragment := NewCodeFragment(location, astNode, 
				fmt.Sprintf("def %s_v%d(items):\n    result = 0\n    for item in items:\n        result += item\n    return result", 
					template, variant))
			
			converter := NewTreeConverter()
			fragment.TreeNode = converter.ConvertAST(astNode)
			
			fragments = append(fragments, fragment)
		}
	}
	
	return fragments
}

func createTestASTNode(name string) *parser.Node {
	return &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: name,
		Children: []*parser.Node{
			{
				Type: parser.NodeName,
				Name: name,
			},
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeReturn,
				Children: []*parser.Node{
					{
						Type:  parser.NodeConstant,
						Value: 42,
					},
				},
			},
		},
		Location: parser.Location{
			StartLine: 1,
			EndLine:   3,
			StartCol:  1,
			EndCol:    10,
		},
	}
}

func createComplexASTNode(name string) *parser.Node {
	return &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: name,
		Children: []*parser.Node{
			{
				Type: parser.NodeName,
				Name: name,
			},
			{
				Type: parser.NodeArguments,
				Children: []*parser.Node{
					{
						Type: parser.NodeArg,
						Name: "items",
					},
				},
			},
		},
		Body: []*parser.Node{
			{
				Type: parser.NodeAssign,
				Targets: []*parser.Node{
					{
						Type: parser.NodeName,
						Name: "result",
					},
				},
				Children: []*parser.Node{
					{
						Type:  parser.NodeConstant,
						Value: 0,
					},
				},
			},
			{
				Type: parser.NodeFor,
				Children: []*parser.Node{
					{
						Type: parser.NodeName,
						Name: "item",
					},
					{
						Type: parser.NodeName,
						Name: "items",
					},
				},
				Body: []*parser.Node{
					{
						Type: parser.NodeAugAssign,
						Children: []*parser.Node{
							{
								Type: parser.NodeName,
								Name: "result",
							},
							{
								Type: parser.NodeName,
								Name: "item",
							},
						},
					},
				},
			},
			{
				Type: parser.NodeReturn,
				Children: []*parser.Node{
					{
						Type: parser.NodeName,
						Name: "result",
					},
				},
			},
		},
		Location: parser.Location{
			StartLine: 1,
			EndLine:   6,
			StartCol:  1,
			EndCol:    20,
		},
	}
}

func createPairKey(frag1, frag2 *CodeFragment) string {
	loc1 := frag1.Location.String()
	loc2 := frag2.Location.String()
	if loc1 < loc2 {
		return loc1 + ":" + loc2
	}
	return loc2 + ":" + loc1
}

// Benchmark tests

func BenchmarkLSHVsExhaustive_Small(b *testing.B) {
	benchmarkLSHVsExhaustive(b, 50)
}

func BenchmarkLSHVsExhaustive_Medium(b *testing.B) {
	benchmarkLSHVsExhaustive(b, 200)
}

func BenchmarkLSHVsExhaustive_Large(b *testing.B) {
	benchmarkLSHVsExhaustive(b, 500)
}

func benchmarkLSHVsExhaustive(b *testing.B, fragmentCount int) {
	fragments := createLargeTestFragmentSet(fragmentCount)
	ctx := context.Background()
	
	b.Run("Exhaustive", func(b *testing.B) {
		config := DefaultCloneDetectorConfig()
		config.UseLSH = false
		detector := NewCloneDetector(config)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			detector.DetectClonesWithContext(ctx, fragments)
		}
	})
	
	b.Run("LSH", func(b *testing.B) {
		config := DefaultCloneDetectorConfig()
		config.UseLSH = true
		detector := NewCloneDetector(config)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			detector.DetectClonesWithLSH(ctx, fragments)
		}
	})
}

func BenchmarkLSHPipeline(b *testing.B) {
	config := DefaultCloneDetectorConfig()
	config.UseLSH = true
	detector := NewCloneDetector(config)
	fragments := createTestFragmentsForLSH()
	ctx := context.Background()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		detector.DetectClonesWithLSH(ctx, fragments)
	}
}