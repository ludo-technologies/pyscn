package analyzer

import (
	"fmt"
	"testing"

	"github.com/pyqol/pyqol/internal/constants"
	"github.com/pyqol/pyqol/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestCloneDetector_Creation(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	assert.NotNil(t, detector, "Detector should be created")
	assert.Equal(t, config, detector.config, "Config should be set correctly")
	assert.NotNil(t, detector.analyzer, "APTED analyzer should be initialized")
	assert.NotNil(t, detector.converter, "Tree converter should be initialized")
}

func TestCloneDetectorConfig_Defaults(t *testing.T) {
	config := DefaultCloneDetectorConfig()

	assert.Equal(t, 5, config.MinLines, "Default min lines should be 5")
	assert.Equal(t, 10, config.MinNodes, "Default min nodes should be 10")
	assert.Equal(t, constants.DefaultType1CloneThreshold, config.Type1Threshold, "Default Type-1 threshold should match constant")
	assert.Equal(t, constants.DefaultType2CloneThreshold, config.Type2Threshold, "Default Type-2 threshold should match constant")
	assert.Equal(t, constants.DefaultType3CloneThreshold, config.Type3Threshold, "Default Type-3 threshold should match constant")
	assert.Equal(t, constants.DefaultType4CloneThreshold, config.Type4Threshold, "Default Type-4 threshold should match constant")
	assert.Equal(t, 50.0, config.MaxEditDistance, "Default max edit distance should be 50.0")
	assert.False(t, config.IgnoreLiterals, "Default ignore literals should be false")
	assert.False(t, config.IgnoreIdentifiers, "Default ignore identifiers should be false")
	assert.Equal(t, "python", config.CostModelType, "Default cost model should be python")
}

func TestCloneType_String(t *testing.T) {
	tests := []struct {
		cloneType CloneType
		expected  string
	}{
		{Type1Clone, "Type-1 (Identical)"},
		{Type2Clone, "Type-2 (Renamed)"},
		{Type3Clone, "Type-3 (Near-Miss)"},
		{Type4Clone, "Type-4 (Semantic)"},
		{CloneType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.cloneType.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCodeLocation_String(t *testing.T) {
	location := &CodeLocation{
		FilePath:  "/path/to/file.py",
		StartLine: 10,
		EndLine:   20,
		StartCol:  5,
		EndCol:    15,
	}

	expected := "/path/to/file.py:10:5-20:15"
	result := location.String()
	assert.Equal(t, expected, result)
}

func TestCodeFragment_Creation(t *testing.T) {
	location := &CodeLocation{
		FilePath:  "/test/file.py",
		StartLine: 1,
		EndLine:   10,
		StartCol:  1,
		EndCol:    20,
	}

	// Create a simple mock AST node
	astNode := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "test_function",
	}

	content := "def test_function():\n    pass"
	fragment := NewCodeFragment(location, astNode, content)

	assert.NotNil(t, fragment, "Fragment should be created")
	assert.Equal(t, location, fragment.Location, "Location should be set")
	assert.Equal(t, astNode, fragment.ASTNode, "AST node should be set")
	assert.Equal(t, content, fragment.Content, "Content should be set")
	assert.Equal(t, 10, fragment.LineCount, "Line count should be calculated correctly")
	assert.Greater(t, fragment.Size, 0, "Size should be calculated")
}

func TestCalculateASTSize(t *testing.T) {
	// Test with nil node
	size := calculateASTSize(nil)
	assert.Equal(t, 0, size, "Nil node should have size 0")

	// Test with single node
	node := &parser.Node{
		Type: parser.NodeName,
		Name: "variable",
	}
	size = calculateASTSize(node)
	assert.Equal(t, 1, size, "Single node should have size 1")

	// Test with node with children
	parent := &parser.Node{
		Type: parser.NodeFunctionDef,
		Name: "test_func",
		Children: []*parser.Node{
			{Type: parser.NodeName, Name: "param1"},
			{Type: parser.NodeName, Name: "param2"},
		},
	}
	size = calculateASTSize(parent)
	assert.Equal(t, 3, size, "Node with 2 children should have size 3")
}

func TestClonePair_String(t *testing.T) {
	fragment1 := &CodeFragment{
		Location: &CodeLocation{
			FilePath:  "/test1.py",
			StartLine: 1,
			EndLine:   5,
		},
	}

	fragment2 := &CodeFragment{
		Location: &CodeLocation{
			FilePath:  "/test2.py",
			StartLine: 10,
			EndLine:   14,
		},
	}

	pair := &ClonePair{
		Fragment1:  fragment1,
		Fragment2:  fragment2,
		Similarity: 0.85,
		CloneType:  Type2Clone,
	}

	result := pair.String()
	expected := "Type-2 (Renamed) clone: /test1.py:1:0-5:0 <-> /test2.py:10:0-14:0 (similarity: 0.850)"
	assert.Equal(t, expected, result)
}

func TestCloneGroup_Operations(t *testing.T) {
	group := NewCloneGroup(1)
	assert.Equal(t, 1, group.ID, "Group ID should be set")
	assert.Equal(t, 0, group.Size, "Initial size should be 0")
	assert.Empty(t, group.Fragments, "Initial fragments should be empty")

	// Add fragment
	fragment := &CodeFragment{
		Location: &CodeLocation{FilePath: "/test.py"},
		Size:     10,
	}
	group.AddFragment(fragment)

	assert.Equal(t, 1, group.Size, "Size should be updated after adding fragment")
	assert.Len(t, group.Fragments, 1, "Fragments should contain added fragment")
	assert.Equal(t, fragment, group.Fragments[0], "Fragment should be stored correctly")
}

func TestCloneDetector_IsFragmentCandidate(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	tests := []struct {
		name     string
		nodeType parser.NodeType
		expected bool
	}{
		{"function definition", parser.NodeFunctionDef, true},
		{"async function definition", parser.NodeAsyncFunctionDef, true},
		{"class definition", parser.NodeClassDef, true},
		{"for loop", parser.NodeFor, true},
		{"async for loop", parser.NodeAsyncFor, true},
		{"while loop", parser.NodeWhile, true},
		{"if statement", parser.NodeIf, true},
		{"try statement", parser.NodeTry, true},
		{"with statement", parser.NodeWith, true},
		{"async with statement", parser.NodeAsyncWith, true},
		{"simple expression", parser.NodeName, false},
		{"constant", parser.NodeConstant, false},
		{"binary operation", parser.NodeBinOp, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := &parser.Node{Type: tt.nodeType}
			result := detector.isFragmentCandidate(node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCloneDetector_ShouldIncludeFragment(t *testing.T) {
	config := &CloneDetectorConfig{
		MinLines: 5,
		MinNodes: 10,
	}
	detector := NewCloneDetector(config)

	tests := []struct {
		name      string
		size      int
		lineCount int
		expected  bool
	}{
		{"meets both requirements", 15, 8, true},
		{"too few nodes", 8, 8, false},
		{"too few lines", 15, 3, false},
		{"meets neither requirement", 5, 3, false},
		{"exactly at thresholds", 10, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fragment := &CodeFragment{
				Size:      tt.size,
				LineCount: tt.lineCount,
			}
			result := detector.shouldIncludeFragment(fragment)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCloneDetector_ClassifyCloneType(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	tests := []struct {
		name       string
		similarity float64
		expected   CloneType
	}{
		{"very high similarity", 0.96, Type1Clone},
		{"high similarity", 0.90, Type2Clone},
		{"medium similarity", 0.75, Type3Clone},
		{"low similarity", 0.65, Type4Clone},
		{"very low similarity", 0.50, CloneType(0)}, // Not a clone
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.classifyCloneType(tt.similarity, 0.0)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCloneDetector_IsSignificantClone(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	fragment1 := &CodeFragment{Size: 15, LineCount: 8}
	fragment2 := &CodeFragment{Size: 12, LineCount: 7}

	tests := []struct {
		name       string
		similarity float64
		distance   float64
		expected   bool
	}{
		{"high similarity, low distance", 0.85, 10.0, true},
		{"low similarity", 0.40, 10.0, false},
		{"high distance", 0.85, 100.0, false},
		{"threshold values", 0.60, 50.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pair := &ClonePair{
				Fragment1:  fragment1,
				Fragment2:  fragment2,
				Similarity: tt.similarity,
				Distance:   tt.distance,
			}
			result := detector.isSignificantClone(pair)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCloneDetector_IsSameLocation(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	loc1 := &CodeLocation{
		FilePath:  "/test.py",
		StartLine: 10,
		EndLine:   20,
	}

	loc2 := &CodeLocation{
		FilePath:  "/test.py",
		StartLine: 10,
		EndLine:   20,
	}

	loc3 := &CodeLocation{
		FilePath:  "/other.py",
		StartLine: 10,
		EndLine:   20,
	}

	assert.True(t, detector.isSameLocation(loc1, loc2), "Identical locations should be same")
	assert.False(t, detector.isSameLocation(loc1, loc3), "Different file paths should not be same")
}

func TestCloneDetector_CalculateConfidence(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	fragment1 := &CodeFragment{
		Size:       50,
		Complexity: 5,
	}

	fragment2 := &CodeFragment{
		Size:       45,
		Complexity: 5,
	}

	confidence := detector.calculateConfidence(fragment1, fragment2, 0.8)

	assert.GreaterOrEqual(t, confidence, 0.8, "Confidence should be at least the similarity")
	assert.LessOrEqual(t, confidence, 1.0, "Confidence should not exceed 1.0")
}

func TestCloneDetector_GetStatistics(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	// Add some mock data
	detector.fragments = make([]*CodeFragment, 10)
	detector.clonePairs = []*ClonePair{
		{CloneType: Type1Clone, Similarity: 0.95},
		{CloneType: Type2Clone, Similarity: 0.85},
		{CloneType: Type1Clone, Similarity: 0.98},
	}
	detector.cloneGroups = make([]*CloneGroup, 2)

	stats := detector.GetStatistics()

	assert.Equal(t, 10, stats["total_fragments"])
	assert.Equal(t, 3, stats["total_clone_pairs"])
	assert.Equal(t, 2, stats["total_clone_groups"])

	typeStats := stats["clone_types"].(map[string]int)
	assert.Equal(t, 2, typeStats["Type-1 (Identical)"])
	assert.Equal(t, 1, typeStats["Type-2 (Renamed)"])

	avgSimilarity := stats["average_similarity"].(float64)
	expected := (0.95 + 0.85 + 0.98) / 3
	assert.InDelta(t, expected, avgSimilarity, 0.001)
}

// Integration test with mock AST nodes
func TestCloneDetector_ExtractFragments_Integration(t *testing.T) {
	config := &CloneDetectorConfig{
		MinLines: 1,
		MinNodes: 1,
	}
	detector := NewCloneDetector(config)

	// Create mock AST nodes representing functions
	function1 := &parser.Node{
		Type:     parser.NodeFunctionDef,
		Name:     "test_function_1",
		Location: parser.Location{StartLine: 1, EndLine: 10},
		Children: []*parser.Node{
			{Type: parser.NodeName, Name: "param1"},
		},
	}

	function2 := &parser.Node{
		Type:     parser.NodeFunctionDef,
		Name:     "test_function_2",
		Location: parser.Location{StartLine: 15, EndLine: 25},
		Children: []*parser.Node{
			{Type: parser.NodeName, Name: "param2"},
		},
	}

	astNodes := []*parser.Node{function1, function2}
	fragments := detector.ExtractFragments(astNodes, "/test.py")

	assert.Len(t, fragments, 2, "Should extract 2 fragments")

	for i, fragment := range fragments {
		assert.NotNil(t, fragment.Location, "Fragment %d should have location", i)
		assert.Equal(t, "/test.py", fragment.Location.FilePath, "Fragment %d should have correct file path", i)
		assert.Greater(t, fragment.Size, 0, "Fragment %d should have positive size", i)
		assert.Greater(t, fragment.LineCount, 0, "Fragment %d should have positive line count", i)
	}
}

// Benchmark tests
func BenchmarkCloneDetector_ExtractFragments(b *testing.B) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	// Create benchmark AST with many function nodes
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fragments := detector.ExtractFragments(astNodes, "/benchmark.py")
		_ = fragments // Prevent compiler optimization
	}
}

func BenchmarkCloneDetector_CompareFragments(b *testing.B) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	// Create two similar fragments
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

	// Create simple tree nodes for comparison
	tree1 := NewTreeNode(1, "FunctionDef")
	tree1.AddChild(NewTreeNode(2, "Body"))
	fragment1.TreeNode = tree1
	PrepareTreeForAPTED(tree1)

	tree2 := NewTreeNode(1, "FunctionDef")
	tree2.AddChild(NewTreeNode(2, "Body"))
	fragment2.TreeNode = tree2
	PrepareTreeForAPTED(tree2)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pair := detector.compareFragments(fragment1, fragment2)
		_ = pair // Prevent compiler optimization
	}
}

// Test error conditions and edge cases
func TestCloneDetector_EdgeCases(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	detector := NewCloneDetector(config)

	// Test with empty fragments
	emptyFragments := []*CodeFragment{}
	pairs, groups := detector.DetectClones(emptyFragments)
	assert.Empty(t, pairs, "Should return empty pairs for empty input")
	assert.Empty(t, groups, "Should return empty groups for empty input")

	// Test with single fragment
	singleFragment := []*CodeFragment{
		{
			Location:  &CodeLocation{FilePath: "/test.py"},
			TreeNode:  NewTreeNode(1, "Test"),
			Size:      10,
			LineCount: 5,
		},
	}
	pairs, groups = detector.DetectClones(singleFragment)
	assert.Empty(t, pairs, "Should return empty pairs for single fragment")
	assert.Empty(t, groups, "Should return empty groups for single fragment")

	// Test with fragments that have no tree nodes
	fragmentsWithoutTrees := []*CodeFragment{
		{Location: &CodeLocation{FilePath: "/test1.py"}, TreeNode: nil},
		{Location: &CodeLocation{FilePath: "/test2.py"}, TreeNode: nil},
	}
	pairs, groups = detector.DetectClones(fragmentsWithoutTrees)
	assert.Empty(t, pairs, "Should handle fragments without tree nodes gracefully")
	assert.Empty(t, groups, "Should handle fragments without tree nodes gracefully")
}
