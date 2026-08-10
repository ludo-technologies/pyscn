package service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
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
		ShowDetails:     domain.BoolPtr(true),
		Recursive:       domain.BoolPtr(false),
	}
}

func findFunctionComplexity(functions []domain.FunctionComplexity, name string) *domain.FunctionComplexity {
	for i := range functions {
		if functions[i].Name == name {
			return &functions[i]
		}
	}
	return nil
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
		assert.NotEmpty(t, response.RawMetrics)
		assert.NotNil(t, response.RawMetricsSummary)
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

	t.Run("report_unchanged false filters complexity one functions", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := tempDir + "/mixed.py"
		content := []byte("def unchanged():\n    return 1\n\n\ndef branch(x):\n    if x:\n        return 1\n    return 0\n")
		err := os.WriteFile(filePath, content, 0644)
		require.NoError(t, err)

		req := newDefaultComplexityRequest(filePath)
		req.ReportUnchanged = domain.BoolPtr(false)

		response, err := service.Analyze(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		require.Len(t, response.Functions, 1)
		assert.Equal(t, "branch", response.Functions[0].Name)
		assert.Equal(t, len(response.AnalyzedFunctions), response.Summary.TotalFunctions)
		assert.Equal(t, response.Summary.TotalFunctions, response.Summary.FunctionsParsed)
	})

	t.Run("module rollups ignore presentation filters", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := tempDir + "/mixed.py"
		content := []byte("def unchanged():\n    return 1\n\n\ndef branch(x):\n    if x:\n        return 1\n    return 0\n")
		require.NoError(t, os.WriteFile(filePath, content, 0644))

		baselineRequest := newDefaultComplexityRequest(filePath)
		baselineRequest.MinComplexity = 1
		baselineRequest.ReportUnchanged = domain.BoolPtr(true)
		baseline, err := service.Analyze(ctx, baselineRequest)
		require.NoError(t, err)

		filteredRequest := newDefaultComplexityRequest(filePath)
		filteredRequest.MinComplexity = 5
		filteredRequest.ReportUnchanged = domain.BoolPtr(false)
		filtered, err := service.Analyze(ctx, filteredRequest)
		require.NoError(t, err)

		assert.NotEqual(t, baseline.Functions, filtered.Functions)
		assert.Equal(t, baseline.ModuleRollups, filtered.ModuleRollups)
		assert.Greater(t, baseline.ModuleRollups[filePath].AnalyzedFunctionCount, len(filtered.Functions))
	})

	t.Run("summary aggregates ignore min_complexity across entrypoints", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := tempDir + "/mixed.py"
		content := []byte(`def trivial():
    return 1


def small(flag):
    if flag:
        return 1
    return 0


def branchy(x):
    if x > 3:
        return 4
    elif x > 2:
        return 3
    elif x > 1:
        return 2
    elif x > 0:
        return 1
    return 0
`)
		require.NoError(t, os.WriteFile(filePath, content, 0644))

		baselineReq := newDefaultComplexityRequest(filePath)
		baselineReq.MinComplexity = 1
		baselineReq.ReportUnchanged = domain.BoolPtr(true)
		filteredReq := newDefaultComplexityRequest(filePath)
		filteredReq.MinComplexity = 5
		filteredReq.ReportUnchanged = domain.BoolPtr(true)

		baseline, err := service.Analyze(ctx, baselineReq)
		require.NoError(t, err)
		filtered, err := service.Analyze(ctx, filteredReq)
		require.NoError(t, err)

		require.NotEmpty(t, filtered.Functions, "at least one function must survive min_complexity=5")
		require.Less(t, len(filtered.Functions), len(baseline.Functions), "min_complexity=5 must drop the trivial functions")
		assert.Equal(t, len(baseline.AnalyzedFunctions), baseline.Summary.TotalFunctions)
		assert.Equal(t, baseline.Summary.TotalFunctions, baseline.Summary.FunctionsParsed)
		assert.Equal(t, baseline.Summary, filtered.Summary, "raising min_complexity must not change summary counts or aggregates")

		snapshot := BuildProjectSnapshot(ctx, []string{filePath})
		baselineSnap, err := service.AnalyzeSnapshot(ctx, snapshot, baselineReq)
		require.NoError(t, err)
		filteredSnap, err := service.AnalyzeSnapshot(ctx, snapshot, filteredReq)
		require.NoError(t, err)

		assert.Equal(t, baseline.Summary, baselineSnap.Summary, "snapshot path must produce the same summary as Analyze")
		assert.Equal(t, baselineSnap.Summary, filteredSnap.Summary, "raising min_complexity must not change snapshot summary counts or aggregates")
	})

	t.Run("public metrics use literal statement counts", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := tempDir + "/literal_counts.py"
		content := []byte(`def debug_no_ifs():
    values = []
    for var in [1, 2, 3]:
        values.append(var)
    for factor in [1, 2, 3]:
        for var in [1, 2, 3]:
            values.append(var * factor)
    for factor in [1, 2, 3]:
        for var in [1, 2, 3]:
            values.append(factor - var)
    return values

def handle(value):
    try:
        risky(value)
    except ValueError:
        if value == 1:
            return 1
        elif value == 2:
            return 2
    except TypeError:
        return 3
`)
		err := os.WriteFile(filePath, content, 0644)
		require.NoError(t, err)

		req := newDefaultComplexityRequest(filePath)

		response, err := service.Analyze(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		noIfs := findFunctionComplexity(response.Functions, "debug_no_ifs")
		require.NotNil(t, noIfs)
		assert.Equal(t, 0, noIfs.Metrics.IfStatements)
		assert.Equal(t, 5, noIfs.Metrics.LoopStatements)
		assert.Equal(t, 0, noIfs.Metrics.ExceptionHandlers)

		handler := findFunctionComplexity(response.Functions, "handle")
		require.NotNil(t, handler)
		assert.Equal(t, 2, handler.Metrics.IfStatements)
		assert.Equal(t, 0, handler.Metrics.LoopStatements)
		assert.Equal(t, 2, handler.Metrics.ExceptionHandlers)
		assert.Equal(t, 5, handler.Metrics.CognitiveComplexity)
	})

	t.Run("module-level complexity includes cognitive and nesting metrics", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := tempDir + "/module_control_flow.py"
		content := []byte(`for i in range(10):
    if i % 2 == 0:
        for j in range(i):
            if j > 0:
                print(i, j)
            else:
                print("zero")
    elif i % 3 == 0:
        print("three")
    else:
        print("other")
`)
		err := os.WriteFile(filePath, content, 0644)
		require.NoError(t, err)

		req := newDefaultComplexityRequest(filePath)

		response, err := service.Analyze(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		var module *domain.FunctionComplexity
		for i := range response.Functions {
			if response.Functions[i].Name == domain.ModuleFunctionName {
				module = &response.Functions[i]
				break
			}
		}
		require.NotNil(t, module, "expected module-level complexity result")
		assert.Greater(t, module.Metrics.Complexity, 1)
		assert.Greater(t, module.Metrics.IfStatements, 0)
		assert.Greater(t, module.Metrics.LoopStatements, 0)
		assert.Greater(t, module.Metrics.CognitiveComplexity, 0)
		assert.Greater(t, module.Metrics.NestingDepth, 0)
	})

	t.Run("enabled false suppresses complexity results", func(t *testing.T) {
		req := newDefaultComplexityRequest("../testdata/python/simple/functions.py")
		req.Enabled = domain.BoolPtr(false)

		response, err := service.Analyze(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Empty(t, response.Functions)
		assert.NotEmpty(t, response.RawMetrics)
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
		require.NotNil(t, response.RawMetricsSummary)
		assert.Len(t, response.RawMetrics, 2)
		assert.Equal(t, 2, response.RawMetricsSummary.FilesAnalyzed)

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

	t.Run("comment only files still return raw metrics", func(t *testing.T) {
		req := newDefaultComplexityRequest("../testdata/python/edge_cases/syntax_errors.py")

		response, err := service.Analyze(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Len(t, response.RawMetrics, 1)
		require.NotNil(t, response.RawMetricsSummary)
		assert.Equal(t, 1, response.RawMetricsSummary.FilesAnalyzed)
		assert.NotZero(t, response.RawMetrics[0].TotalLines)
		assert.Equal(t, 0, response.RawMetrics[0].LLOC)
	})

	t.Run("parse errors still return raw metrics", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := tempDir + "/invalid.py"
		err := os.WriteFile(filePath, []byte("def broken(:\n    pass\n"), 0644)
		require.NoError(t, err)

		req := newDefaultComplexityRequest(filePath)

		response, err := service.Analyze(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Empty(t, response.Functions)
		assert.Len(t, response.RawMetrics, 1)
		require.NotNil(t, response.RawMetricsSummary)
		assert.Equal(t, 0, response.RawMetrics[0].LLOC)
		assert.NotEmpty(t, response.Errors)
		assert.Equal(t, 0, response.Summary.FilesAnalyzed)
		assert.Equal(t, 1, response.Summary.TotalFiles)
		assert.Equal(t, 1, response.Summary.SkippedFiles)
	})

	t.Run("raw metrics include parse failures without inflating analyzed file count", func(t *testing.T) {
		tempDir := t.TempDir()
		invalidPath := tempDir + "/invalid.py"
		err := os.WriteFile(invalidPath, []byte("def broken(:\n    pass\n"), 0644)
		require.NoError(t, err)

		validPath := "../testdata/python/simple/functions.py"
		req := newDefaultComplexityRequest(validPath, invalidPath)

		response, err := service.Analyze(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		require.NotNil(t, response.RawMetricsSummary)
		assert.NotEmpty(t, response.Functions)
		assert.NotEmpty(t, response.Errors)
		assert.Equal(t, 1, response.Summary.FilesAnalyzed)
		// The shortfall must be visible from the summary alone: a consumer
		// should not have to cross-reference Errors against its own input
		// list to notice that half the request was dropped (issue #690).
		assert.Equal(t, 2, response.Summary.TotalFiles)
		assert.Equal(t, 1, response.Summary.SkippedFiles)
		assert.Equal(t, 2, response.RawMetricsSummary.FilesAnalyzed)
		assert.Len(t, response.RawMetrics, 2)

		rawMetricPaths := make(map[string]bool, len(response.RawMetrics))
		for _, metrics := range response.RawMetrics {
			rawMetricPaths[metrics.FilePath] = true
		}
		assert.True(t, rawMetricPaths[validPath])
		assert.True(t, rawMetricPaths[invalidPath])
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

		filtered, _ := service.filterFunctions(functions, req)

		require.Len(t, filtered, 3)
		assert.Equal(t, "func2", filtered[0].Name)
		assert.Equal(t, "func3", filtered[1].Name)
		assert.Equal(t, "func4", filtered[2].Name)
	})

	t.Run("MaxComplexity does not filter functions", func(t *testing.T) {
		// MaxComplexity is used only as a warning threshold, not for filtering
		req := domain.ComplexityRequest{
			MinComplexity: 1,
			MaxComplexity: 8, // This should NOT filter out functions
		}

		filtered, _ := service.filterFunctions(functions, req)

		// All 4 functions should be returned (MaxComplexity doesn't filter)
		require.Len(t, filtered, 4)
		assert.Equal(t, "func1", filtered[0].Name)
		assert.Equal(t, "func2", filtered[1].Name)
		assert.Equal(t, "func3", filtered[2].Name)
		assert.Equal(t, "func4", filtered[3].Name)
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
			Name: "func1",
			Metrics: domain.ComplexityMetrics{
				Complexity:          2,
				CognitiveComplexity: 3,
				NestingDepth:        1,
			},
			RiskLevel: domain.RiskLevelLow,
		},
		{
			Name: "func2",
			Metrics: domain.ComplexityMetrics{
				Complexity:          8,
				CognitiveComplexity: 12,
				NestingDepth:        2,
			},
			RiskLevel: domain.RiskLevelMedium,
		},
		{
			Name: "func3",
			Metrics: domain.ComplexityMetrics{
				Complexity:          15,
				CognitiveComplexity: 30,
				NestingDepth:        6,
			},
			RiskLevel: domain.RiskLevelHigh,
		},
	}

	t.Run("generate summary with functions", func(t *testing.T) {
		summary := service.generateSummary(functions, complexitySummaryCounts{
			filesAnalyzed: 2,
		})

		assert.Equal(t, 3, summary.TotalFunctions)
		assert.Equal(t, 3, summary.FunctionsParsed)
		assert.Equal(t, 2, summary.FilesAnalyzed)
		assert.Equal(t, 8.333333333333334, summary.AverageComplexity) // (2+8+15)/3
		assert.Equal(t, 15.0, summary.AverageCognitiveComplexity)
		assert.Equal(t, 3.0, summary.AverageNestingDepth)
		assert.Equal(t, 15, summary.MaxComplexity)
		assert.Equal(t, 2, summary.MinComplexity)
		assert.Equal(t, 1, summary.LowRiskFunctions)
		assert.Equal(t, 1, summary.MediumRiskFunctions)
		assert.Equal(t, 1, summary.HighRiskFunctions)
		assert.NotNil(t, summary.ComplexityDistribution)
	})

	t.Run("generate summary with no functions", func(t *testing.T) {
		summary := service.generateSummary([]domain.FunctionComplexity{}, complexitySummaryCounts{
			filesAnalyzed: 5,
		})

		assert.Equal(t, 0, summary.TotalFunctions)
		assert.Equal(t, 0, summary.FunctionsParsed)
		assert.Equal(t, 5, summary.FilesAnalyzed)
		assert.Equal(t, 0.0, summary.AverageComplexity)
		assert.Equal(t, 0, summary.MaxComplexity)
		assert.Equal(t, 0, summary.MinComplexity)
	})

	t.Run("summary counts use complete population when min_complexity drops functions", func(t *testing.T) {
		allFunctions := []domain.FunctionComplexity{
			{Name: "trivial", Metrics: domain.ComplexityMetrics{Complexity: 1}},
			{Name: "trivial2", Metrics: domain.ComplexityMetrics{Complexity: 2}},
			{Name: "complex", Metrics: domain.ComplexityMetrics{Complexity: 8}},
		}

		req := domain.ComplexityRequest{MinComplexity: 5}
		filtered, _ := service.filterFunctions(allFunctions, req)
		summary := service.generateSummary(allFunctions, complexitySummaryCounts{
			filesAnalyzed: 1,
		})

		assert.Len(t, filtered, 1, "the presentation filter still limits the reported list")
		assert.Equal(t, 3, summary.TotalFunctions, "summary count describes the population used for its aggregates")
		assert.Equal(t, 3, summary.FunctionsParsed, "all parsed functions contribute to summary aggregates")
		assert.InDelta(t, 11.0/3.0, summary.AverageComplexity, 1e-9, "average is over all parsed functions, not the filtered set")
		assert.Equal(t, 1, summary.MinComplexity, "min reflects the complete population")
	})
}

func TestComplexityService_CalculateRiskLevel(t *testing.T) {
	service := NewComplexityService()

	testCases := []struct {
		name                string
		complexity          int
		cognitiveComplexity int
		nestingDepth        int
		lowThreshold        int
		mediumThreshold     int
		expectedRiskLevel   domain.RiskLevel
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
		{
			name:                "high risk from cognitive complexity",
			complexity:          3,
			cognitiveComplexity: domain.DefaultCognitiveComplexityThreshold + 1,
			lowThreshold:        5,
			mediumThreshold:     10,
			expectedRiskLevel:   domain.RiskLevelHigh,
		},
		{
			name:              "high risk from nesting depth",
			complexity:        3,
			nestingDepth:      domain.DefaultNestingDepthThreshold + 1,
			lowThreshold:      5,
			mediumThreshold:   10,
			expectedRiskLevel: domain.RiskLevelHigh,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := domain.ComplexityRequest{
				LowThreshold:    tc.lowThreshold,
				MediumThreshold: tc.mediumThreshold,
			}

			riskLevel := service.calculateRiskLevel(tc.complexity, tc.cognitiveComplexity, tc.nestingDepth, req)
			assert.Equal(t, tc.expectedRiskLevel, riskLevel)
		})
	}
}

func TestComplexityService_MetricThresholdWarnings(t *testing.T) {
	service := NewComplexityService()
	result := &analyzer.ComplexityResult{
		FunctionName:        "dense",
		StartLine:           12,
		StartCol:            4,
		CognitiveComplexity: 30,
		NestingDepth:        8,
	}
	req := domain.ComplexityRequest{
		CognitiveComplexityThreshold: 25,
		NestingDepthThreshold:        7,
	}

	warnings := service.metricThresholdWarnings("sample.py", "dense", result, req)

	require.Len(t, warnings, 2)
	assert.Contains(t, warnings[0], "sample.py:12:5")
	assert.Contains(t, warnings[0], "dense cognitive complexity too high (30 > 25)")
	assert.Contains(t, warnings[1], "dense nesting depth too high (8 > 7)")
}

func TestComplexityService_FunctionSLOC(t *testing.T) {
	service := NewComplexityService()
	ctx := context.Background()

	source := `"""Module docstring."""


def build_table():
    """Build a table."""
    rows = []
`
	for i := 0; i < 60; i++ {
		source += fmt.Sprintf("    rows.append(%d)\n", i)
	}
	source += `    return rows


def short():
    return 1
`

	path := fmt.Sprintf("%s/long_function.py", t.TempDir())
	require.NoError(t, os.WriteFile(path, []byte(source), 0o644))

	req := newDefaultComplexityRequest(path)
	response, err := service.Analyze(ctx, req)
	require.NoError(t, err)

	longFunction := findFunctionComplexity(response.Functions, "build_table")
	require.NotNil(t, longFunction)
	// def + docstring is excluded + rows = [] + 60 appends + return
	assert.Equal(t, 63, longFunction.Metrics.SLOC)

	shortFunction := findFunctionComplexity(response.Functions, "short")
	require.NotNil(t, shortFunction)
	assert.Equal(t, 2, shortFunction.Metrics.SLOC)

	// The function stays visible as long even though its complexity is 1.
	assert.Equal(t, 1, longFunction.Metrics.Complexity)
	assert.True(t, longFunction.ExceedsSLOC(domain.DefaultFunctionSLOCWarnThreshold))
	assert.False(t, shortFunction.ExceedsSLOC(domain.DefaultFunctionSLOCWarnThreshold))

	// Module scope spans the whole file but must never count as a long function.
	moduleScope := findFunctionComplexity(response.Functions, domain.ModuleFunctionName)
	require.NotNil(t, moduleScope)
	assert.Greater(t, moduleScope.Metrics.SLOC, 63)
	assert.False(t, moduleScope.ExceedsSLOC(domain.DefaultFunctionSLOCWarnThreshold))
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
	req.Recursive = domain.BoolPtr(true)
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
	assert.Equal(t, true, configMap["enabled"])
	assert.Equal(t, true, configMap["report_unchanged"])
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
