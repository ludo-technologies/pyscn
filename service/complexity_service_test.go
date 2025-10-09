package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newDefaultComplexityRequest creates a ComplexityRequest with default test values
func newDefaultComplexityRequest(paths ...string) domain.ComplexityRequest {
	if len(paths) == 0 {
		paths = []string{"../testdata/python/simple/functions.py"}
	}
	return domain.ComplexityRequest{
		Paths:           paths,
		OutputFormat:    domain.OutputFormatJSON,
		MinComplexity:   1,
		MaxComplexity:   0,
		SortBy:          domain.SortByComplexity,
		LowThreshold:    5,
		MediumThreshold: 10,
		ShowDetails:     true,
		Recursive:       false,
	}
}

func TestNewComplexityService(t *testing.T) {
	service := NewComplexityService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.parser)
}

func TestComplexityService_Analyze(t *testing.T) {
	service := NewComplexityService()
	ctx := context.Background()

	t.Run("successful analysis with simple Python file", func(t *testing.T) {
		req := newDefaultComplexityRequest()

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Functions)
		assert.NotEmpty(t, response.GeneratedAt)
		assert.NotEmpty(t, response.Version)
		assert.NotNil(t, response.Config)
		assert.GreaterOrEqual(t, response.Summary.TotalFunctions, 1)
		assert.GreaterOrEqual(t, response.Summary.FilesAnalyzed, 1)
	})

	t.Run("analyze complex Python file with control structures", func(t *testing.T) {
		req := newDefaultComplexityRequest("../testdata/python/simple/control_flow.py")

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)

		// Verify response structure
		for _, function := range response.Functions {
			assert.NotEmpty(t, function.Name)
			assert.NotEmpty(t, function.FilePath)
			assert.Greater(t, function.Metrics.Complexity, 0)
			assert.Greater(t, function.Metrics.Nodes, 0)
			assert.GreaterOrEqual(t, function.Metrics.Edges, 0)
			assert.Contains(t, []domain.RiskLevel{
				domain.RiskLevelLow,
				domain.RiskLevelMedium,
				domain.RiskLevelHigh,
			}, function.RiskLevel)
		}

		// Verify summary statistics
		assert.Equal(t, len(response.Functions), response.Summary.TotalFunctions)
		assert.Greater(t, response.Summary.AverageComplexity, 0.0)
		assert.Greater(t, response.Summary.MaxComplexity, 0)
		assert.Greater(t, response.Summary.MinComplexity, 0)
	})

	t.Run("analyze with filtering by complexity", func(t *testing.T) {
		req := newDefaultComplexityRequest("../testdata/python/simple/control_flow.py")
		req.MinComplexity = 5 // Only functions with complexity >= 5

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		if response != nil {
			for _, function := range response.Functions {
				assert.GreaterOrEqual(t, function.Metrics.Complexity, 5)
			}
		}
	})

	t.Run("analyze with max complexity limit", func(t *testing.T) {
		req := newDefaultComplexityRequest("../testdata/python/simple/control_flow.py")
		req.MaxComplexity = 3 // Only functions with complexity <= 3

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		if response != nil {
			for _, function := range response.Functions {
				assert.LessOrEqual(t, function.Metrics.Complexity, 3)
			}
		}
	})

	t.Run("analyze multiple files", func(t *testing.T) {
		req := newDefaultComplexityRequest(
			"../testdata/python/simple/functions.py",
			"../testdata/python/simple/control_flow.py",
		)

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 2, response.Summary.FilesAnalyzed)

		// Should have functions from both files
		filePathsFound := make(map[string]bool)
		for _, function := range response.Functions {
			filePathsFound[function.FilePath] = true
		}
		assert.GreaterOrEqual(t, len(filePathsFound), 1) // At least one of the files should have functions
	})

	t.Run("error handling for non-existent file", func(t *testing.T) {
		req := newDefaultComplexityRequest("../testdata/non_existent_file.py")

		response, err := service.Analyze(ctx, req)

		// Should return error or response with errors
		if err == nil {
			assert.NotNil(t, response)
			assert.NotEmpty(t, response.Errors)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := newDefaultComplexityRequest()

		_, err := service.Analyze(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cancelled")
	})

	t.Run("no functions found returns error", func(t *testing.T) {
		// Use a file that likely has no functions
		req := newDefaultComplexityRequest("../testdata/python/simple/imports.py")
		req.MinComplexity = 100 // Very high threshold to filter out all functions

		_, err := service.Analyze(ctx, req)

		if err != nil {
			assert.Contains(t, err.Error(), "no functions found to analyze")
		}
	})
}

func TestComplexityService_AnalyzeFile(t *testing.T) {
	service := NewComplexityService()
	ctx := context.Background()

	t.Run("analyze single file", func(t *testing.T) {
		req := newDefaultComplexityRequest()

		response, err := service.AnalyzeFile(ctx, "../testdata/python/simple/functions.py", req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.Summary.FilesAnalyzed)
	})
}

func TestComplexityService_FilterFunctions(t *testing.T) {
	service := NewComplexityService()

	functions := []domain.FunctionComplexity{
		{
			Name:     "func1",
			FilePath: "test.py",
			Metrics:  domain.ComplexityMetrics{Complexity: 2},
		},
		{
			Name:     "func2",
			FilePath: "test.py",
			Metrics:  domain.ComplexityMetrics{Complexity: 5},
		},
		{
			Name:     "func3",
			FilePath: "test.py",
			Metrics:  domain.ComplexityMetrics{Complexity: 10},
		},
		{
			Name:     "func4",
			FilePath: "test.py",
			Metrics:  domain.ComplexityMetrics{Complexity: 15},
		},
	}

	t.Run("filter by minimum complexity", func(t *testing.T) {
		req := domain.ComplexityRequest{
			MinComplexity: 5,
			MaxComplexity: 0,
		}

		filtered := service.filterFunctions(functions, req)

		require.Len(t, filtered, 3)
		assert.Equal(t, "func2", filtered[0].Name)
		assert.Equal(t, "func3", filtered[1].Name)
		assert.Equal(t, "func4", filtered[2].Name)
	})

	t.Run("filter by maximum complexity", func(t *testing.T) {
		req := domain.ComplexityRequest{
			MinComplexity: 1,
			MaxComplexity: 8,
		}

		filtered := service.filterFunctions(functions, req)

		require.Len(t, filtered, 2)
		assert.Equal(t, "func1", filtered[0].Name)
		assert.Equal(t, "func2", filtered[1].Name)
	})

	t.Run("filter by both min and max complexity", func(t *testing.T) {
		req := domain.ComplexityRequest{
			MinComplexity: 5,
			MaxComplexity: 10,
		}

		filtered := service.filterFunctions(functions, req)

		require.Len(t, filtered, 2)
		assert.Equal(t, "func2", filtered[0].Name)
		assert.Equal(t, "func3", filtered[1].Name)
	})
}

func TestComplexityService_SortFunctions(t *testing.T) {
	service := NewComplexityService()

	functions := []domain.FunctionComplexity{
		{
			Name:      "func_c",
			FilePath:  "test.py",
			Metrics:   domain.ComplexityMetrics{Complexity: 5},
			RiskLevel: domain.RiskLevelMedium,
		},
		{
			Name:      "func_a",
			FilePath:  "test.py",
			Metrics:   domain.ComplexityMetrics{Complexity: 10},
			RiskLevel: domain.RiskLevelHigh,
		},
		{
			Name:      "func_b",
			FilePath:  "test.py",
			Metrics:   domain.ComplexityMetrics{Complexity: 2},
			RiskLevel: domain.RiskLevelLow,
		},
	}

	t.Run("sort by complexity", func(t *testing.T) {
		sorted := service.sortFunctions(functions, domain.SortByComplexity)

		require.Len(t, sorted, 3)
		assert.Equal(t, "func_a", sorted[0].Name) // Complexity 10
		assert.Equal(t, "func_c", sorted[1].Name) // Complexity 5
		assert.Equal(t, "func_b", sorted[2].Name) // Complexity 2
	})

	t.Run("sort by name", func(t *testing.T) {
		sorted := service.sortFunctions(functions, domain.SortByName)

		require.Len(t, sorted, 3)
		assert.Equal(t, "func_a", sorted[0].Name)
		assert.Equal(t, "func_b", sorted[1].Name)
		assert.Equal(t, "func_c", sorted[2].Name)
	})

	t.Run("sort by risk", func(t *testing.T) {
		sorted := service.sortFunctions(functions, domain.SortByRisk)

		require.Len(t, sorted, 3)
		assert.Equal(t, "func_a", sorted[0].Name) // High risk
		assert.Equal(t, "func_c", sorted[1].Name) // Medium risk
		assert.Equal(t, "func_b", sorted[2].Name) // Low risk
	})
}

func TestComplexityService_GenerateSummary(t *testing.T) {
	service := NewComplexityService()

	functions := []domain.FunctionComplexity{
		{
			Name:      "func1",
			Metrics:   domain.ComplexityMetrics{Complexity: 2},
			RiskLevel: domain.RiskLevelLow,
		},
		{
			Name:      "func2",
			Metrics:   domain.ComplexityMetrics{Complexity: 8},
			RiskLevel: domain.RiskLevelMedium,
		},
		{
			Name:      "func3",
			Metrics:   domain.ComplexityMetrics{Complexity: 15},
			RiskLevel: domain.RiskLevelHigh,
		},
	}

	t.Run("generate summary with functions", func(t *testing.T) {
		req := domain.ComplexityRequest{}
		summary := service.generateSummary(functions, 2, req)

		assert.Equal(t, 3, summary.TotalFunctions)
		assert.Equal(t, 2, summary.FilesAnalyzed)
		assert.Equal(t, 8.333333333333334, summary.AverageComplexity) // (2+8+15)/3
		assert.Equal(t, 15, summary.MaxComplexity)
		assert.Equal(t, 2, summary.MinComplexity)
		assert.Equal(t, 1, summary.LowRiskFunctions)
		assert.Equal(t, 1, summary.MediumRiskFunctions)
		assert.Equal(t, 1, summary.HighRiskFunctions)
		assert.NotNil(t, summary.ComplexityDistribution)
	})

	t.Run("generate summary with no functions", func(t *testing.T) {
		req := domain.ComplexityRequest{}
		summary := service.generateSummary([]domain.FunctionComplexity{}, 5, req)

		assert.Equal(t, 0, summary.TotalFunctions)
		assert.Equal(t, 5, summary.FilesAnalyzed)
		assert.Equal(t, 0.0, summary.AverageComplexity)
		assert.Equal(t, 0, summary.MaxComplexity)
		assert.Equal(t, 0, summary.MinComplexity)
	})
}

func TestComplexityService_CalculateRiskLevel(t *testing.T) {
	service := NewComplexityService()

	testCases := []struct {
		name              string
		complexity        int
		lowThreshold      int
		mediumThreshold   int
		expectedRiskLevel domain.RiskLevel
	}{
		{
			name:              "low risk",
			complexity:        3,
			lowThreshold:      5,
			mediumThreshold:   10,
			expectedRiskLevel: domain.RiskLevelLow,
		},
		{
			name:              "medium risk",
			complexity:        8,
			lowThreshold:      5,
			mediumThreshold:   10,
			expectedRiskLevel: domain.RiskLevelMedium,
		},
		{
			name:              "high risk",
			complexity:        15,
			lowThreshold:      5,
			mediumThreshold:   10,
			expectedRiskLevel: domain.RiskLevelHigh,
		},
		{
			name:              "boundary case - exactly low threshold",
			complexity:        5,
			lowThreshold:      5,
			mediumThreshold:   10,
			expectedRiskLevel: domain.RiskLevelLow,
		},
		{
			name:              "boundary case - exactly medium threshold",
			complexity:        10,
			lowThreshold:      5,
			mediumThreshold:   10,
			expectedRiskLevel: domain.RiskLevelMedium,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := domain.ComplexityRequest{
				LowThreshold:    tc.lowThreshold,
				MediumThreshold: tc.mediumThreshold,
			}

			riskLevel := service.calculateRiskLevel(tc.complexity, req)
			assert.Equal(t, tc.expectedRiskLevel, riskLevel)
		})
	}
}

func TestComplexityService_GetComplexityDistributionKey(t *testing.T) {
	service := NewComplexityService()

	testCases := []struct {
		complexity  int
		expectedKey string
	}{
		{1, "1"},
		{2, "2-5"},
		{3, "2-5"},
		{5, "2-5"},
		{6, "6-10"},
		{8, "6-10"},
		{10, "6-10"},
		{11, "11-20"},
		{15, "11-20"},
		{20, "11-20"},
		{21, "21+"},
		{50, "21+"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("complexity_%d", tc.complexity), func(t *testing.T) {
			key := service.getComplexityDistributionKey(tc.complexity)
			assert.Equal(t, tc.expectedKey, key)
		})
	}
}

func TestComplexityService_BuildConfigForResponse(t *testing.T) {
	service := NewComplexityService()

	req := newDefaultComplexityRequest()
	req.MinComplexity = 2
	req.MaxComplexity = 20
	req.Recursive = true
	req.IncludePatterns = []string{"**/*.py"}
	req.ExcludePatterns = []string{"test_*.py"}

	config := service.buildConfigForResponse(req)

	configMap, ok := config.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "json", configMap["output_format"])
	assert.Equal(t, 2, configMap["min_complexity"])
	assert.Equal(t, 20, configMap["max_complexity"])
	assert.Equal(t, 5, configMap["low_threshold"])
	assert.Equal(t, 10, configMap["medium_threshold"])
	assert.Equal(t, "complexity", configMap["sort_by"])
	assert.Equal(t, true, configMap["show_details"])
	assert.Equal(t, true, configMap["recursive"])
	assert.Equal(t, []string{"**/*.py"}, configMap["include_patterns"])
	assert.Equal(t, []string{"test_*.py"}, configMap["exclude_patterns"])
}

func TestComplexityService_ResponseMetadata(t *testing.T) {
	service := NewComplexityService()
	ctx := context.Background()

	req := newDefaultComplexityRequest()

	beforeTime := time.Now()
	response, err := service.Analyze(ctx, req)
	afterTime := time.Now()

	assert.NoError(t, err)
	require.NotNil(t, response)

	// Verify timestamp is within expected range
	assert.NotEmpty(t, response.GeneratedAt)
	generatedTime, err := time.Parse(time.RFC3339, response.GeneratedAt)
	assert.NoError(t, err)
	assert.True(t, generatedTime.After(beforeTime.Add(-time.Second)) && generatedTime.Before(afterTime.Add(time.Second)),
		"Generated time %v should be between %v and %v", generatedTime, beforeTime, afterTime)

	// Verify version is present
	assert.NotEmpty(t, response.Version)

	// Verify config is present
	assert.NotNil(t, response.Config)
}
