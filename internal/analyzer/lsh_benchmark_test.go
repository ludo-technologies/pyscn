package analyzer

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

// BenchmarkSuite contains comprehensive benchmarks for the LSH implementation

func BenchmarkFeatureExtraction(b *testing.B) {
	extractor := NewASTFeatureExtractor()
	tree := createLSHBenchmarkTree(100) // Tree with 100 nodes
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := extractor.ExtractFeatures(tree)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMinHashComputation(b *testing.B) {
	hasher := NewMinHasher(128)
	features := generateBenchmarkFeatures(1000) // 1000 features
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.ComputeSignature(features)
	}
}

func BenchmarkLSHIndexBuild(b *testing.B) {
	hasher := NewMinHasher(128)
	
	// Pre-generate signatures
	signatures := make(map[string]*MinHashSignature)
	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("fragment_%d", i)
		features := generateBenchmarkFeatures(50)
		signatures[id] = hasher.ComputeSignature(features)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index := NewDefaultLSHIndex()
		err := index.BuildIndex(signatures)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLSHCandidateRetrieval(b *testing.B) {
	// Build index once
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	signatures := make(map[string]*MinHashSignature)
	for i := 0; i < 1000; i++ {
		id := fmt.Sprintf("fragment_%d", i)
		features := generateBenchmarkFeatures(50)
		signatures[id] = hasher.ComputeSignature(features)
	}
	
	err := index.BuildIndex(signatures)
	if err != nil {
		b.Fatal(err)
	}
	
	querySignature := hasher.ComputeSignature(generateBenchmarkFeatures(50))
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.FindCandidates(querySignature)
	}
}

func BenchmarkCloneDetectionScaling(b *testing.B) {
	fragmentCounts := []int{10, 50, 100, 200, 500}
	
	for _, count := range fragmentCounts {
		b.Run(fmt.Sprintf("fragments_%d", count), func(b *testing.B) {
			benchmarkCloneDetectionWithFragmentCount(b, count)
		})
	}
}

func benchmarkCloneDetectionWithFragmentCount(b *testing.B, fragmentCount int) {
	fragments := createBenchmarkFragments(fragmentCount)
	
	b.Run("exhaustive", func(b *testing.B) {
		config := DefaultCloneDetectorConfig()
		config.UseLSH = false
		detector := NewCloneDetector(config)
		ctx := context.Background()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			detector.DetectClonesWithContext(ctx, fragments)
		}
	})
	
	b.Run("lsh", func(b *testing.B) {
		config := DefaultCloneDetectorConfig()
		config.UseLSH = true
		detector := NewCloneDetector(config)
		ctx := context.Background()
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			detector.DetectClonesWithLSH(ctx, fragments)
		}
	})
}

func BenchmarkLSHMemoryUsage(b *testing.B) {
	var m1, m2 runtime.MemStats
	
	fragmentCounts := []int{100, 500, 1000}
	
	for _, count := range fragmentCounts {
		b.Run(fmt.Sprintf("fragments_%d", count), func(b *testing.B) {
			fragments := createBenchmarkFragments(count)
			
			b.Run("exhaustive", func(b *testing.B) {
				config := DefaultCloneDetectorConfig()
				config.UseLSH = false
				detector := NewCloneDetector(config)
				ctx := context.Background()
				
				runtime.GC()
				runtime.ReadMemStats(&m1)
				
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					detector.DetectClonesWithContext(ctx, fragments)
				}
				
				runtime.ReadMemStats(&m2)
				b.ReportMetric(float64(m2.Alloc-m1.Alloc), "bytes/op")
			})
			
			b.Run("lsh", func(b *testing.B) {
				config := DefaultCloneDetectorConfig()
				config.UseLSH = true
				detector := NewCloneDetector(config)
				ctx := context.Background()
				
				runtime.GC()
				runtime.ReadMemStats(&m1)
				
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					detector.DetectClonesWithLSH(ctx, fragments)
				}
				
				runtime.ReadMemStats(&m2)
				b.ReportMetric(float64(m2.Alloc-m1.Alloc), "bytes/op")
			})
		})
	}
}

func BenchmarkLSHParameterTuning(b *testing.B) {
	fragments := createBenchmarkFragments(200)
	ctx := context.Background()
	
	configurations := []struct {
		name   string
		bands  int
		rows   int
		hashes int
	}{
		{"low_precision", 16, 2, 64},
		{"medium_precision", 32, 4, 128},
		{"high_precision", 64, 8, 256},
		{"very_high_precision", 128, 4, 512},
	}
	
	for _, config := range configurations {
		b.Run(config.name, func(b *testing.B) {
			detectorConfig := DefaultCloneDetectorConfig()
			detectorConfig.UseLSH = true
			detectorConfig.LSHBands = config.bands
			detectorConfig.LSHRows = config.rows
			detectorConfig.LSHMinHashCount = config.hashes
			
			detector := NewCloneDetector(detectorConfig)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				detector.DetectClonesWithLSH(ctx, fragments)
			}
		})
	}
}

func BenchmarkFeatureExtractionComplexity(b *testing.B) {
	extractor := NewASTFeatureExtractor()
	
	treeSizes := []int{10, 50, 100, 500}
	
	for _, size := range treeSizes {
		b.Run(fmt.Sprintf("tree_size_%d", size), func(b *testing.B) {
			tree := createLSHBenchmarkTree(size)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := extractor.ExtractFeatures(tree)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkMinHashSizes(b *testing.B) {
	features := generateBenchmarkFeatures(500)
	
	hashCounts := []int{32, 64, 128, 256, 512}
	
	for _, count := range hashCounts {
		b.Run(fmt.Sprintf("hashes_%d", count), func(b *testing.B) {
			hasher := NewMinHasher(count)
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hasher.ComputeSignature(features)
			}
		})
	}
}

func BenchmarkLSHBandConfigurations(b *testing.B) {
	hasher := NewMinHasher(256)
	signature := hasher.ComputeSignature(generateBenchmarkFeatures(100))
	
	configurations := []struct {
		bands int
		rows  int
	}{
		{16, 2},
		{32, 4},
		{64, 4},
		{128, 2},
	}
	
	for _, config := range configurations {
		b.Run(fmt.Sprintf("bands_%d_rows_%d", config.bands, config.rows), func(b *testing.B) {
			lshConfig := LSHConfig{
				Bands: config.bands,
				Rows:  config.rows,
			}
			index := NewLSHIndex(lshConfig)
			
			// Add signature to index
			err := index.AddFragment("test", signature)
			if err != nil {
				b.Skip("Configuration requires too many hashes")
			}
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				index.FindCandidates(signature)
			}
		})
	}
}

func BenchmarkConcurrentLSH(b *testing.B) {
	fragments := createBenchmarkFragments(100)
	config := DefaultCloneDetectorConfig()
	config.UseLSH = true
	detector := NewCloneDetector(config)
	ctx := context.Background()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			detector.DetectClonesWithLSH(ctx, fragments)
		}
	})
}

func BenchmarkLSHIndexUpdate(b *testing.B) {
	index := NewDefaultLSHIndex()
	hasher := NewMinHasher(128)
	
	// Pre-populate index
	for i := 0; i < 500; i++ {
		signature := hasher.ComputeSignature(generateBenchmarkFeatures(50))
		if err := index.AddFragment(fmt.Sprintf("frag_%d", i), signature); err != nil {
			b.Fatalf("Failed to add fragment: %v", err)
		}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signature := hasher.ComputeSignature(generateBenchmarkFeatures(50))
		err := index.AddFragment(fmt.Sprintf("new_frag_%d", i), signature)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSignatureSimilarity(b *testing.B) {
	hasher := NewMinHasher(128)
	
	sig1 := hasher.ComputeSignature(generateBenchmarkFeatures(100))
	sig2 := hasher.ComputeSignature(generateBenchmarkFeatures(100))
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.EstimateJaccardSimilarity(sig1, sig2)
	}
}

// Helper functions for benchmarks

func createLSHBenchmarkTree(nodeCount int) *TreeNode {
	if nodeCount <= 0 {
		return &TreeNode{ID: 0, Label: "EmptyNode"}
	}
	
	root := &TreeNode{
		ID:    0,
		Label: "FunctionDef",
	}
	
	// Create a balanced binary tree-like structure
	nodes := []*TreeNode{root}
	nodeID := 1
	
	for len(nodes) > 0 && nodeID < nodeCount {
		current := nodes[0]
		nodes = nodes[1:]
		
		// Add children (2 per node to create binary tree)
		for child := 0; child < 2 && nodeID < nodeCount; child++ {
			childNode := &TreeNode{
				ID:     nodeID,
				Label:  fmt.Sprintf("Node_%d", nodeID),
				Parent: current,
			}
			current.Children = append(current.Children, childNode)
			nodes = append(nodes, childNode)
			nodeID++
		}
	}
	
	return root
}

func generateBenchmarkFeatures(count int) []string {
	features := make([]string, count)
	for i := 0; i < count; i++ {
		features[i] = fmt.Sprintf("feature_%d", i)
	}
	return features
}

func createBenchmarkFragments(count int) []*CodeFragment {
	fragments := make([]*CodeFragment, count)
	converter := NewTreeConverter()
	
	for i := 0; i < count; i++ {
		location := &CodeLocation{
			FilePath:  fmt.Sprintf("bench_%d.py", i),
			StartLine: 1,
			EndLine:   10,
		}
		
		astNode := &parser.Node{
			Type: parser.NodeFunctionDef,
			Name: fmt.Sprintf("bench_function_%d", i),
			Body: []*parser.Node{
				{
					Type: parser.NodeReturn,
					Children: []*parser.Node{
						{
							Type:  parser.NodeConstant,
							Value: i,
						},
					},
				},
			},
			Location: parser.Location{
				StartLine: 1,
				EndLine:   3,
			},
		}
		
		fragment := NewCodeFragment(location, astNode, fmt.Sprintf("def bench_function_%d():\n    return %d", i, i))
		fragment.TreeNode = converter.ConvertAST(astNode)
		
		fragments[i] = fragment
	}
	
	return fragments
}

// Performance comparison benchmark
func BenchmarkPerformanceComparison(b *testing.B) {
	fragmentCounts := []int{50, 100, 200}
	
	for _, count := range fragmentCounts {
		fragments := createBenchmarkFragments(count)
		
		b.Run(fmt.Sprintf("count_%d", count), func(b *testing.B) {
			// Measure exhaustive search
			b.Run("exhaustive", func(b *testing.B) {
				config := DefaultCloneDetectorConfig()
				config.UseLSH = false
				detector := NewCloneDetector(config)
				ctx := context.Background()
				
				start := time.Now()
				b.ResetTimer()
				
				for i := 0; i < b.N; i++ {
					detector.DetectClonesWithContext(ctx, fragments)
				}
				
				elapsed := time.Since(start)
				b.ReportMetric(float64(elapsed.Nanoseconds())/float64(b.N), "ns/op")
			})
			
			// Measure LSH search
			b.Run("lsh", func(b *testing.B) {
				config := DefaultCloneDetectorConfig()
				config.UseLSH = true
				detector := NewCloneDetector(config)
				ctx := context.Background()
				
				start := time.Now()
				b.ResetTimer()
				
				for i := 0; i < b.N; i++ {
					detector.DetectClonesWithLSH(ctx, fragments)
				}
				
				elapsed := time.Since(start)
				b.ReportMetric(float64(elapsed.Nanoseconds())/float64(b.N), "ns/op")
			})
		})
	}
}