package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType4CloneDetection(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	config.EnableDFAAnalysis = true
	config.Type4Threshold = 0.70

	detector := NewCloneDetector(config)
	allFragments, functionFragments := loadType4FunctionFragments(t, detector)

	pairs, groups := detector.DetectClones(allFragments)
	require.NotEmpty(t, pairs)
	require.NotEmpty(t, groups)

	pairByID := make(map[string]*ClonePair, len(pairs))
	for _, pair := range pairs {
		pairByID[type4PairID(type4FragmentID(pair.Fragment1), type4FragmentID(pair.Fragment2))] = pair
	}

	expectedPairs := [][2]string{
		{"sum_iterative.py::sum_numbers", "sum_recursive.py::sum_numbers"},
		{"sum_iterative.py::sum_range", "sum_recursive.py::sum_range"},
		{"find_max_a.py::find_maximum", "find_max_b.py::find_maximum"},
		{"find_max_a.py::find_min_max", "find_max_b.py::find_min_max"},
		{"filter_transform_a.py::filter_and_double", "filter_transform_b.py::filter_and_double"},
		{"filter_transform_a.py::process_data", "filter_transform_b.py::process_data"},
		{"filter_transform_a.py::count_matching", "filter_transform_b.py::count_matching"},
	}
	for _, expected := range expectedPairs {
		pair := pairByID[type4PairID(expected[0], expected[1])]
		require.NotNil(t, pair, "missing Type-4 pair %s <-> %s", expected[0], expected[1])
		assert.Equal(t, Type4Clone, pair.CloneType)
		assert.GreaterOrEqual(t, pair.Similarity, config.Type4Threshold)
	}

	negativePairs := [][2]string{
		{"sum_iterative.py::sum_numbers", "find_max_b.py::find_maximum"},
		{"sum_iterative.py::sum_range", "find_max_b.py::find_min_max"},
		{"find_max_a.py::find_maximum", "find_max_b.py::find_min_max"},
	}
	for _, negative := range negativePairs {
		assert.Nil(t, pairByID[type4PairID(negative[0], negative[1])], "unexpected Type-4 pair %s <-> %s", negative[0], negative[1])
	}

	for _, expected := range expectedPairs {
		first := functionFragments[expected[0]]
		second := functionFragments[expected[1]]
		require.NotNil(t, first, "missing fixture function %s", expected[0])
		require.NotNil(t, second, "missing fixture function %s", expected[1])

		pair := detector.compareFragments(first, second)
		require.NotNil(t, pair, "direct comparison missed %s <-> %s", expected[0], expected[1])
		assert.Equal(t, Type4Clone, pair.CloneType)
	}
}

func loadType4FunctionFragments(t *testing.T, detector *CloneDetector) ([]*CodeFragment, map[string]*CodeFragment) {
	t.Helper()

	testDir := "../../testdata/python/clones/type4"
	files, err := filepath.Glob(filepath.Join(testDir, "*.py"))
	require.NoError(t, err)
	sort.Strings(files)

	p := parser.New()
	ctx := context.Background()
	functionFragments := make(map[string]*CodeFragment)
	var allFragments []*CodeFragment

	for _, file := range files {
		content, err := os.ReadFile(file)
		require.NoError(t, err)

		result, err := p.Parse(ctx, content)
		require.NoError(t, err)

		fragments := detector.ExtractFragmentsWithSource([]*parser.Node{result.AST}, file, content)
		allFragments = append(allFragments, fragments...)
		for _, fragment := range fragments {
			if fragment.ASTNode != nil && fragment.ASTNode.Type == parser.NodeFunctionDef {
				functionFragments[type4FragmentID(fragment)] = fragment
			}
		}
	}

	require.NotEmpty(t, allFragments)
	return allFragments, functionFragments
}

func type4FragmentID(fragment *CodeFragment) string {
	if fragment == nil || fragment.Location == nil || fragment.ASTNode == nil {
		return ""
	}
	return filepath.Base(fragment.Location.FilePath) + "::" + fragment.ASTNode.Name
}

func type4PairID(first, second string) string {
	if first > second {
		first, second = second, first
	}
	return first + "|" + second
}
