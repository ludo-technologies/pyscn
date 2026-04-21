package analyzer

import (
	"context"
	"fmt"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

func BenchmarkCloneDetector_DetectClones(b *testing.B) {
	datasets := []struct {
		name            string
		familyCount     int
		copiesPerFamily int
		noiseCount      int
	}{
		{name: "ExactCloneFamilyDataset_72", familyCount: 6, copiesPerFamily: 8, noiseCount: 24},
		{name: "ExactCloneFamilyDataset_144", familyCount: 12, copiesPerFamily: 8, noiseCount: 48},
	}

	for _, dataset := range datasets {
		fragments := buildCloneBenchmarkFragments(dataset.familyCount, dataset.copiesPerFamily, dataset.noiseCount)
		b.Run(dataset.name, func(b *testing.B) {
			benchmarkCloneDetectionMode(b, "standard", cloneBenchmarkConfig(false), fragments)
			benchmarkCloneDetectionMode(b, "lsh", cloneBenchmarkConfig(true), fragments)
		})
	}
}

func benchmarkCloneDetectionMode(b *testing.B, name string, config *CloneDetectorConfig, fragments []*CodeFragment) {
	b.Run(name, func(b *testing.B) {
		b.ReportAllocs()
		ctx := context.Background()
		for i := 0; i < b.N; i++ {
			detector := NewCloneDetector(config)

			var pairs []*ClonePair
			var groups []*CloneGroup
			if config.UseLSH {
				pairs, groups = detector.DetectClonesWithLSH(ctx, fragments)
			} else {
				pairs, groups = detector.DetectClones(fragments)
			}

			if len(pairs) == 0 || len(groups) == 0 {
				b.Fatalf("expected clone detection benchmark dataset to produce pairs and groups")
			}
		}
	})
}

func BenchmarkCloneDetector_ExtractFragments(b *testing.B) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	astNodes := make([]*parser.Node, 100)
	for i := 0; i < 100; i++ {
		astNodes[i] = &parser.Node{
			Type:     parser.NodeFunctionDef,
			Name:     fmt.Sprintf("function_%d", i),
			Location: parser.Location{StartLine: i * 10, EndLine: (i * 10) + 8},
			Children: []*parser.Node{
				{Type: parser.NodeName, Name: fmt.Sprintf("param_%d", i)},
			},
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fragments := detector.ExtractFragments(astNodes, "/benchmark.py")
		_ = fragments
	}
}

func BenchmarkCloneDetector_CompareFragments(b *testing.B) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	fragment1 := &CodeFragment{
		Location:  &CodeLocation{FilePath: "/test1.py"},
		Size:      20,
		LineCount: 10,
	}
	fragment2 := &CodeFragment{
		Location:  &CodeLocation{FilePath: "/test2.py"},
		Size:      18,
		LineCount: 9,
	}

	tree1 := NewTreeNode(1, "FunctionDef")
	tree1.AddChild(NewTreeNode(2, "Body"))
	fragment1.TreeNode = tree1
	PrepareTreeForAPTED(tree1)

	tree2 := NewTreeNode(1, "FunctionDef")
	tree2.AddChild(NewTreeNode(2, "Body"))
	fragment2.TreeNode = tree2
	PrepareTreeForAPTED(tree2)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pair := detector.compareFragments(fragment1, fragment2)
		_ = pair
	}
}

func BenchmarkCloneDetectionMemory(b *testing.B) {
	datasets := []struct {
		name  string
		count int
	}{
		{name: "SmallDataset", count: 100},
		{name: "LargeDataset", count: 200},
	}

	for _, dataset := range datasets {
		fragments := createTestFragments(dataset.count, "bench")
		b.Run(dataset.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				detector := NewCloneDetector(cloneBenchmarkConfig(false))
				detector.fragments = fragments
				detector.clonePairs = nil
				detector.DetectClones(detector.fragments)
			}
		})
	}
}
