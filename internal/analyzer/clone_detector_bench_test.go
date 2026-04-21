package analyzer

import (
	"context"
	"fmt"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

func BenchmarkCloneDetector_DetectClones(b *testing.B) {
	datasets := []struct {
		name              string
		familyCount       int
		variantsPerFamily int
		noiseCount        int
	}{
		{name: "FamilyDataset_72", familyCount: 6, variantsPerFamily: 8, noiseCount: 24},
		{name: "FamilyDataset_144", familyCount: 12, variantsPerFamily: 8, noiseCount: 48},
	}

	for _, dataset := range datasets {
		fragments := buildCloneBenchmarkFragments(dataset.familyCount, dataset.variantsPerFamily, dataset.noiseCount)
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

func cloneBenchmarkConfig(useLSH bool) *CloneDetectorConfig {
	config := DefaultCloneDetectorConfig()
	config.MinLines = 2
	config.MinNodes = 3
	config.Type4Threshold = 0.60
	config.GroupingThreshold = 0.70
	config.MaxClonePairs = 5000
	config.UseLSH = useLSH
	config.LSHSimilarityThreshold = 0.35
	config.LSHMinHashCount = 128
	config.LSHBands = 32
	config.LSHRows = 4
	return config
}

func buildCloneBenchmarkFragments(familyCount, variantsPerFamily, noiseCount int) []*CodeFragment {
	extractor := NewASTFeatureExtractor()
	fragments := make([]*CodeFragment, 0, familyCount*variantsPerFamily+noiseCount)
	fragmentIndex := 0

	for family := 0; family < familyCount; family++ {
		for variant := 0; variant < variantsPerFamily; variant++ {
			tree := buildCloneBenchmarkTree(family, variant)
			PrepareTreeForAPTED(tree)
			features, _ := extractor.ExtractFeatures(tree)

			lineCount := 6 + (variant % 3)
			fragments = append(fragments, &CodeFragment{
				Location: &CodeLocation{
					FilePath:  fmt.Sprintf("family_%02d_variant_%02d.py", family, variant),
					StartLine: 1,
					EndLine:   lineCount,
					StartCol:  1,
					EndCol:    80,
				},
				Content:    fmt.Sprintf("family_%02d_variant_%02d", family, variant),
				Hash:       fmt.Sprintf("family_%02d_variant_%02d_hash", family, variant),
				Size:       benchmarkTreeSize(tree),
				LineCount:  lineCount,
				Complexity: 2 + (family+variant)%4,
				TreeNode:   tree,
				Features:   features,
			})
			fragmentIndex++
		}
	}

	for noise := 0; noise < noiseCount; noise++ {
		tree := buildCloneBenchmarkNoiseTree(familyCount + noise)
		PrepareTreeForAPTED(tree)
		features, _ := extractor.ExtractFeatures(tree)

		lineCount := 5 + (noise % 4)
		fragments = append(fragments, &CodeFragment{
			Location: &CodeLocation{
				FilePath:  fmt.Sprintf("noise_%02d.py", noise),
				StartLine: 1,
				EndLine:   lineCount,
				StartCol:  1,
				EndCol:    80,
			},
			Content:    fmt.Sprintf("noise_%02d", noise),
			Hash:       fmt.Sprintf("noise_%02d_hash", noise),
			Size:       benchmarkTreeSize(tree),
			LineCount:  lineCount,
			Complexity: 1 + (noise % 5),
			TreeNode:   tree,
			Features:   features,
		})
		fragmentIndex++
	}

	return fragments
}

func buildCloneBenchmarkTree(family, variant int) *TreeNode {
	root := NewTreeNode(1, "FunctionDef")
	params := NewTreeNode(2, "Parameters")
	params.AddChild(NewTreeNode(3, fmt.Sprintf("ArgFamily%d", family%4)))
	params.AddChild(NewTreeNode(4, fmt.Sprintf("ArgVariant%d", variant%3)))

	body := NewTreeNode(5, "Body")
	body.AddChild(NewTreeNode(6, "Assign"))

	call := NewTreeNode(7, fmt.Sprintf("CallFamily%d", family))
	call.AddChild(NewTreeNode(8, fmt.Sprintf("ExprVariant%d", variant%2)))
	body.AddChild(call)

	branch := NewTreeNode(9, "If")
	branch.AddChild(NewTreeNode(10, fmt.Sprintf("ConditionFamily%d", family%5)))
	branch.AddChild(NewTreeNode(11, "Return"))
	branch.AddChild(NewTreeNode(12, fmt.Sprintf("ReturnVariant%d", variant%3)))
	body.AddChild(branch)

	root.AddChild(params)
	root.AddChild(body)
	return root
}

func buildCloneBenchmarkNoiseTree(seed int) *TreeNode {
	rootLabels := []string{"ClassDef", "While", "Try", "With", "Match"}
	childLabels := []string{"Assign", "Raise", "Yield", "Await", "Delete"}

	root := NewTreeNode(1, rootLabels[seed%len(rootLabels)])
	left := NewTreeNode(2, childLabels[seed%len(childLabels)])
	left.AddChild(NewTreeNode(3, fmt.Sprintf("NoiseLeaf%d", seed%7)))

	right := NewTreeNode(4, childLabels[(seed+2)%len(childLabels)])
	right.AddChild(NewTreeNode(5, fmt.Sprintf("NoiseBranch%d", seed%11)))

	root.AddChild(left)
	root.AddChild(right)
	return root
}

func benchmarkTreeSize(root *TreeNode) int {
	if root == nil {
		return 0
	}

	size := 1
	for _, child := range root.Children {
		size += benchmarkTreeSize(child)
	}
	return size
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
		{name: "LargeDataset", count: 500},
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
