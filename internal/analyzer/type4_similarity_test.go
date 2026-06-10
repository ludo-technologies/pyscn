package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType4SimilarityScores(t *testing.T) {
	config := DefaultCloneDetectorConfig()
	config.EnableDFAAnalysis = true
	config.Type4Threshold = 0.70
	detector := NewCloneDetector(config)
	_, functions := loadType4FunctionFragments(t, detector)
	analyzer := NewSemanticSimilarityAnalyzerWithDFA()

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
		first := functions[expected[0]]
		second := functions[expected[1]]
		require.NotNil(t, first, "missing fixture function %s", expected[0])
		require.NotNil(t, second, "missing fixture function %s", expected[1])

		similarity := analyzer.ComputeSimilarity(first, second)
		assert.GreaterOrEqual(t, similarity, config.Type4Threshold, "%s <-> %s", expected[0], expected[1])
	}

	negativePairs := [][2]string{
		{"sum_iterative.py::sum_numbers", "find_max_b.py::find_maximum"},
		{"sum_iterative.py::sum_range", "find_max_b.py::find_min_max"},
		{"find_max_a.py::find_maximum", "find_max_b.py::find_min_max"},
		{"field_converter_a.py::delimited_list2param", "field_converter_b.py::uploadfield2properties"},
	}
	for _, negative := range negativePairs {
		first := functions[negative[0]]
		second := functions[negative[1]]
		require.NotNil(t, first, "missing fixture function %s", negative[0])
		require.NotNil(t, second, "missing fixture function %s", negative[1])

		similarity := analyzer.ComputeSimilarity(first, second)
		assert.Less(t, similarity, config.Type4Threshold, "%s <-> %s", negative[0], negative[1])
	}
}
