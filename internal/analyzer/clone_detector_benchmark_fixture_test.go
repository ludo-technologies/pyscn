package analyzer

import "fmt"

// cloneBenchmarkConfig keeps the benchmark fixture in a regime where the
// standard and LSH paths should report the same exact-clone families.
func cloneBenchmarkConfig(useLSH bool) *CloneDetectorConfig {
	config := DefaultCloneDetectorConfig()
	config.MinLines = 2
	config.MinNodes = 3
	config.Type4Threshold = 0.95
	config.GroupingThreshold = 0.95
	config.MaxClonePairs = 5000
	config.UseLSH = useLSH
	config.LSHSimilarityThreshold = 0.50
	config.LSHMinHashCount = 128
	config.LSHBands = 32
	config.LSHRows = 4
	return config
}

func buildCloneBenchmarkFragments(familyCount, copiesPerFamily, noiseCount int) []*CodeFragment {
	extractor := NewASTFeatureExtractor()
	fragments := make([]*CodeFragment, 0, familyCount*copiesPerFamily+noiseCount)

	for family := 0; family < familyCount; family++ {
		complexity := 2 + (family % 4)

		for copyIndex := 0; copyIndex < copiesPerFamily; copyIndex++ {
			tree := buildCloneBenchmarkTree(family)
			PrepareTreeForAPTED(tree)
			features, _ := extractor.ExtractFeatures(tree)
			fragments = append(fragments, &CodeFragment{
				Location: &CodeLocation{
					FilePath:  fmt.Sprintf("family_%02d_copy_%02d.py", family, copyIndex),
					StartLine: 1,
					EndLine:   8,
					StartCol:  1,
					EndCol:    80,
				},
				Content:    fmt.Sprintf("family_%02d", family),
				Hash:       fmt.Sprintf("family_%02d_copy_%02d_hash", family, copyIndex),
				Size:       benchmarkTreeSize(tree),
				LineCount:  8,
				Complexity: complexity,
				TreeNode:   tree,
				Features:   features,
			})
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
	}

	return fragments
}

func buildCloneBenchmarkTree(family int) *TreeNode {
	root := NewTreeNode(1, "FunctionDef")
	params := NewTreeNode(2, "Parameters")
	params.AddChild(NewTreeNode(3, fmt.Sprintf("ArgFamily%d", family%4)))
	params.AddChild(NewTreeNode(4, fmt.Sprintf("ArgFamilyStable%d", family%3)))

	body := NewTreeNode(5, "Body")
	body.AddChild(NewTreeNode(6, "Assign"))

	call := NewTreeNode(7, fmt.Sprintf("CallFamily%d", family))
	call.AddChild(NewTreeNode(8, fmt.Sprintf("ExprFamily%d", family%2)))
	body.AddChild(call)

	branch := NewTreeNode(9, "If")
	branch.AddChild(NewTreeNode(10, fmt.Sprintf("ConditionFamily%d", family%5)))
	branch.AddChild(NewTreeNode(11, "Return"))
	branch.AddChild(NewTreeNode(12, fmt.Sprintf("ReturnFamily%d", family%3)))
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
