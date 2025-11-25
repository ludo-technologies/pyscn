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

func newDefaultCBORequest(paths ...string) domain.CBORequest {
	if len(paths) == 0 {
		paths = []string{"../testdata/python/complex/decorators.py"}
	}
	return domain.CBORequest{
		Paths:           paths,
		OutputFormat:    domain.OutputFormatJSON,
		MinCBO:          0,
		MaxCBO:          0,
		SortBy:          domain.SortByCoupling,
		ShowZeros:       domain.BoolPtr(true),
		LowThreshold:    5,
		MediumThreshold: 10,
		ShowDetails:     true,
		Recursive:       domain.BoolPtr(false),
		IncludeBuiltins: domain.BoolPtr(false),
		IncludeImports:  domain.BoolPtr(true),
	}
}

func TestNewCBOService(t *testing.T) {
	service := NewCBOService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.parser)
}

func TestCBOService_Analyze(t *testing.T) {
	service := NewCBOService()
	ctx := context.Background()

	t.Run("successful analysis with Python file containing classes", func(t *testing.T) {
		req := newDefaultCBORequest("../testdata/python/complex/decorators.py")

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.GeneratedAt)
		assert.NotEmpty(t, response.Version)
		assert.NotNil(t, response.Config)
		assert.GreaterOrEqual(t, response.Summary.FilesAnalyzed, 1)

		// Verify response structure
		for _, class := range response.Classes {
			assert.NotEmpty(t, class.Name)
			assert.NotEmpty(t, class.FilePath)
			assert.GreaterOrEqual(t, class.Metrics.CouplingCount, 0)
			assert.Contains(t, []domain.RiskLevel{
				domain.RiskLevelLow,
				domain.RiskLevelMedium,
				domain.RiskLevelHigh,
			}, class.RiskLevel)
		}
	})

	t.Run("analyze file with no classes should return empty result", func(t *testing.T) {
		req := newDefaultCBORequest("../testdata/python/simple/functions.py")

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Empty(t, response.Classes)
		assert.Contains(t, response.Warnings, "No classes found to analyze")
		assert.Equal(t, 0, response.Summary.TotalClasses)
		assert.Equal(t, 1, response.Summary.FilesAnalyzed)
	})

	t.Run("analyze with filtering by CBO count", func(t *testing.T) {
		req := newDefaultCBORequest("../testdata/python/complex/exceptions.py")
		req.MinCBO = 1 // Only classes with CBO >= 1
		req.ShowZeros = domain.BoolPtr(false)

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		if response != nil {
			for _, class := range response.Classes {
				assert.GreaterOrEqual(t, class.Metrics.CouplingCount, 1)
			}
		}
	})

	t.Run("analyze with max CBO limit", func(t *testing.T) {
		req := newDefaultCBORequest("../testdata/python/complex/exceptions.py")
		req.MaxCBO = 5 // Only classes with CBO <= 5

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		if response != nil {
			for _, class := range response.Classes {
				assert.LessOrEqual(t, class.Metrics.CouplingCount, 5)
			}
		}
	})

	t.Run("analyze multiple files", func(t *testing.T) {
		req := newDefaultCBORequest(
			"../testdata/python/complex/decorators.py",
			"../testdata/python/complex/exceptions.py",
		)

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 2, response.Summary.FilesAnalyzed)
	})

	t.Run("analyze with ShowZeros=false filters out zero CBO classes", func(t *testing.T) {
		req := newDefaultCBORequest("../testdata/python/complex/decorators.py")
		req.ShowZeros = domain.BoolPtr(false) // Filter out zero CBO classes

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		if response != nil {
			for _, class := range response.Classes {
				assert.Greater(t, class.Metrics.CouplingCount, 0)
			}
		}
	})

	t.Run("error handling for non-existent file", func(t *testing.T) {
		req := newDefaultCBORequest("../testdata/non_existent_file.py")

		response, err := service.Analyze(ctx, req)

		// Should return valid response with errors
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Errors)
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := newDefaultCBORequest("../testdata/python/complex/decorators.py")

		_, err := service.Analyze(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cancelled")
	})

	t.Run("analyze with include builtins enabled", func(t *testing.T) {
		req := newDefaultCBORequest("../testdata/python/complex/exceptions.py")
		req.IncludeBuiltins = domain.BoolPtr(true) // Include built-in dependencies

		response, err := service.Analyze(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		// When including builtins, CBO counts may be higher
	})
}

func TestCBOService_AnalyzeFile(t *testing.T) {
	service := NewCBOService()
	ctx := context.Background()

	t.Run("analyze single file", func(t *testing.T) {
		req := domain.CBORequest{
			OutputFormat:    domain.OutputFormatJSON,
			MinCBO:          0,
			MaxCBO:          0,
			SortBy:          domain.SortByCoupling,
			ShowZeros:       domain.BoolPtr(true),
			LowThreshold:    5,
			MediumThreshold: 10,
			ShowDetails:     true,
			Recursive:       domain.BoolPtr(false),
			IncludeBuiltins: domain.BoolPtr(false),
			IncludeImports:  domain.BoolPtr(true),
		}

		response, err := service.AnalyzeFile(ctx, "../testdata/python/complex/decorators.py", req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.Summary.FilesAnalyzed)
	})
}

func TestCBOService_FilterClasses(t *testing.T) {
	service := NewCBOService()

	classes := []domain.ClassCoupling{
		{
			Name:     "ClassA",
			FilePath: "test.py",
			Metrics:  domain.CBOMetrics{CouplingCount: 0},
		},
		{
			Name:     "ClassB",
			FilePath: "test.py",
			Metrics:  domain.CBOMetrics{CouplingCount: 3},
		},
		{
			Name:     "ClassC",
			FilePath: "test.py",
			Metrics:  domain.CBOMetrics{CouplingCount: 8},
		},
		{
			Name:     "ClassD",
			FilePath: "test.py",
			Metrics:  domain.CBOMetrics{CouplingCount: 12},
		},
	}

	t.Run("filter by minimum CBO", func(t *testing.T) {
		req := domain.CBORequest{
			MinCBO:    3,
			MaxCBO:    0,
			ShowZeros: domain.BoolPtr(true),
		}

		filtered := service.filterClasses(classes, req)

		require.Len(t, filtered, 3)
		assert.Equal(t, "ClassB", filtered[0].Name)
		assert.Equal(t, "ClassC", filtered[1].Name)
		assert.Equal(t, "ClassD", filtered[2].Name)
	})

	t.Run("filter by maximum CBO", func(t *testing.T) {
		req := domain.CBORequest{
			MinCBO:    0,
			MaxCBO:    5,
			ShowZeros: domain.BoolPtr(true),
		}

		filtered := service.filterClasses(classes, req)

		require.Len(t, filtered, 2)
		assert.Equal(t, "ClassA", filtered[0].Name)
		assert.Equal(t, "ClassB", filtered[1].Name)
	})

	t.Run("filter by both min and max CBO", func(t *testing.T) {
		req := domain.CBORequest{
			MinCBO:    3,
			MaxCBO:    8,
			ShowZeros: domain.BoolPtr(true),
		}

		filtered := service.filterClasses(classes, req)

		require.Len(t, filtered, 2)
		assert.Equal(t, "ClassB", filtered[0].Name)
		assert.Equal(t, "ClassC", filtered[1].Name)
	})

	t.Run("filter out zeros when ShowZeros=false", func(t *testing.T) {
		req := domain.CBORequest{
			MinCBO:    0,
			MaxCBO:    0,
			ShowZeros: domain.BoolPtr(false),
		}

		filtered := service.filterClasses(classes, req)

		require.Len(t, filtered, 3)
		assert.Equal(t, "ClassB", filtered[0].Name)
		assert.Equal(t, "ClassC", filtered[1].Name)
		assert.Equal(t, "ClassD", filtered[2].Name)
	})
}

func TestCBOService_SortClasses(t *testing.T) {
	service := NewCBOService()

	classes := []domain.ClassCoupling{
		{
			Name:      "ClassC",
			FilePath:  "file2.py",
			StartLine: 20,
			Metrics:   domain.CBOMetrics{CouplingCount: 5},
			RiskLevel: domain.RiskLevelMedium,
		},
		{
			Name:      "ClassA",
			FilePath:  "file1.py",
			StartLine: 10,
			Metrics:   domain.CBOMetrics{CouplingCount: 10},
			RiskLevel: domain.RiskLevelHigh,
		},
		{
			Name:      "ClassB",
			FilePath:  "file1.py",
			StartLine: 5,
			Metrics:   domain.CBOMetrics{CouplingCount: 2},
			RiskLevel: domain.RiskLevelLow,
		},
	}

	t.Run("sort by coupling count", func(t *testing.T) {
		sorted := service.sortClasses(classes, domain.SortByCoupling)

		require.Len(t, sorted, 3)
		assert.Equal(t, "ClassA", sorted[0].Name) // CBO 10
		assert.Equal(t, "ClassC", sorted[1].Name) // CBO 5
		assert.Equal(t, "ClassB", sorted[2].Name) // CBO 2
	})

	t.Run("sort by name", func(t *testing.T) {
		sorted := service.sortClasses(classes, domain.SortByName)

		require.Len(t, sorted, 3)
		assert.Equal(t, "ClassA", sorted[0].Name)
		assert.Equal(t, "ClassB", sorted[1].Name)
		assert.Equal(t, "ClassC", sorted[2].Name)
	})

	t.Run("sort by risk", func(t *testing.T) {
		sorted := service.sortClasses(classes, domain.SortByRisk)

		require.Len(t, sorted, 3)
		assert.Equal(t, "ClassA", sorted[0].Name) // High risk
		assert.Equal(t, "ClassC", sorted[1].Name) // Medium risk
		assert.Equal(t, "ClassB", sorted[2].Name) // Low risk
	})

	t.Run("sort by location", func(t *testing.T) {
		sorted := service.sortClasses(classes, domain.SortByLocation)

		require.Len(t, sorted, 3)
		assert.Equal(t, "ClassB", sorted[0].Name) // file1.py:5
		assert.Equal(t, "ClassA", sorted[1].Name) // file1.py:10
		assert.Equal(t, "ClassC", sorted[2].Name) // file2.py:20
	})

	t.Run("sort by default (coupling)", func(t *testing.T) {
		sorted := service.sortClasses(classes, domain.SortCriteria("unknown"))

		require.Len(t, sorted, 3)
		assert.Equal(t, "ClassA", sorted[0].Name) // CBO 10
		assert.Equal(t, "ClassC", sorted[1].Name) // CBO 5
		assert.Equal(t, "ClassB", sorted[2].Name) // CBO 2
	})
}

func TestCBOService_GenerateSummary(t *testing.T) {
	service := NewCBOService()

	classes := []domain.ClassCoupling{
		{
			Name:      "ClassA",
			Metrics:   domain.CBOMetrics{CouplingCount: 2},
			RiskLevel: domain.RiskLevelLow,
		},
		{
			Name:      "ClassB",
			Metrics:   domain.CBOMetrics{CouplingCount: 8},
			RiskLevel: domain.RiskLevelMedium,
		},
		{
			Name:      "ClassC",
			Metrics:   domain.CBOMetrics{CouplingCount: 15},
			RiskLevel: domain.RiskLevelHigh,
		},
	}

	t.Run("generate summary with classes", func(t *testing.T) {
		req := domain.CBORequest{}
		summary := service.generateSummary(classes, 2, req)

		assert.Equal(t, 3, summary.TotalClasses)
		assert.Equal(t, 3, summary.ClassesAnalyzed)
		assert.Equal(t, 2, summary.FilesAnalyzed)
		assert.Equal(t, 8.333333333333334, summary.AverageCBO) // (2+8+15)/3
		assert.Equal(t, 15, summary.MaxCBO)
		assert.Equal(t, 2, summary.MinCBO)
		assert.Equal(t, 1, summary.LowRiskClasses)
		assert.Equal(t, 1, summary.MediumRiskClasses)
		assert.Equal(t, 1, summary.HighRiskClasses)
		assert.NotNil(t, summary.CBODistribution)
		assert.Len(t, summary.MostCoupledClasses, 3) // All classes since <= 10
	})

	t.Run("generate summary with no classes", func(t *testing.T) {
		req := domain.CBORequest{}
		summary := service.generateSummary([]domain.ClassCoupling{}, 5, req)

		assert.Equal(t, 0, summary.TotalClasses)
		assert.Equal(t, 0, summary.ClassesAnalyzed)
		assert.Equal(t, 5, summary.FilesAnalyzed)
		assert.Equal(t, 0.0, summary.AverageCBO)
		assert.Equal(t, 0, summary.MaxCBO)
		assert.Equal(t, 0, summary.MinCBO)
	})

	t.Run("generate summary limits most coupled classes to 10", func(t *testing.T) {
		// Create 15 classes
		manyClasses := make([]domain.ClassCoupling, 15)
		for i := 0; i < 15; i++ {
			manyClasses[i] = domain.ClassCoupling{
				Name:      fmt.Sprintf("Class%d", i),
				Metrics:   domain.CBOMetrics{CouplingCount: i + 1},
				RiskLevel: domain.RiskLevelLow,
			}
		}

		req := domain.CBORequest{}
		summary := service.generateSummary(manyClasses, 1, req)

		assert.Equal(t, 15, summary.TotalClasses)
		assert.Len(t, summary.MostCoupledClasses, 10) // Should be limited to 10
	})
}

func TestCBOService_GetCBORange(t *testing.T) {
	service := NewCBOService()

	testCases := []struct {
		cbo         int
		expectedKey string
	}{
		{0, "0"},
		{1, "1-5"},
		{3, "1-5"},
		{5, "1-5"},
		{6, "6-10"},
		{8, "6-10"},
		{10, "6-10"},
		{11, "11-20"},
		{15, "11-20"},
		{20, "11-20"},
		{21, "21-50"},
		{30, "21-50"},
		{50, "21-50"},
		{51, "50+"},
		{100, "50+"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("cbo_%d", tc.cbo), func(t *testing.T) {
			key := service.getCBORange(tc.cbo)
			assert.Equal(t, tc.expectedKey, key)
		})
	}
}

func TestCBOService_BuildCBOOptions(t *testing.T) {
	service := NewCBOService()

	req := domain.CBORequest{
		IncludeBuiltins: domain.BoolPtr(true),
		IncludeImports:  domain.BoolPtr(false),
		ExcludePatterns: []string{"test_*.py"},
		LowThreshold:    3,
		MediumThreshold: 8,
	}

	options := service.buildCBOOptions(req)

	assert.NotNil(t, options)
	assert.Equal(t, true, options.IncludeBuiltins)
	assert.Equal(t, false, options.IncludeImports)
	assert.Equal(t, false, options.PublicClassesOnly)
	assert.Equal(t, []string{"test_*.py"}, options.ExcludePatterns)
	assert.Equal(t, 3, options.LowThreshold)
	assert.Equal(t, 8, options.MediumThreshold)
}

func TestCBOService_BuildConfigForResponse(t *testing.T) {
	service := NewCBOService()

	req := domain.CBORequest{
		MinCBO:          1,
		MaxCBO:          20,
		ShowZeros:       domain.BoolPtr(false),
		LowThreshold:    5,
		MediumThreshold: 10,
		IncludeBuiltins: domain.BoolPtr(true),
		IncludeImports:  domain.BoolPtr(false),
		OutputFormat:    domain.OutputFormatJSON,
		SortBy:          domain.SortByCoupling,
	}

	config := service.buildConfigForResponse(req)

	configMap, ok := config.(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, 1, configMap["minCBO"])
	assert.Equal(t, 20, configMap["maxCBO"])
	assert.Equal(t, false, configMap["showZeros"])
	assert.Equal(t, 5, configMap["lowThreshold"])
	assert.Equal(t, 10, configMap["mediumThreshold"])
	assert.Equal(t, true, configMap["includeBuiltins"])
	assert.Equal(t, false, configMap["includeImports"])
	assert.Equal(t, domain.OutputFormatJSON, configMap["outputFormat"])
	assert.Equal(t, domain.SortByCoupling, configMap["sortBy"])
}

func TestCBOService_ResponseMetadata(t *testing.T) {
	service := NewCBOService()
	ctx := context.Background()

	req := newDefaultCBORequest("../testdata/python/complex/decorators.py")

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
