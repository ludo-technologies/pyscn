package service

import (
	"context"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDefaultLCOMRequest(paths ...string) domain.LCOMRequest {
	if len(paths) == 0 {
		paths = []string{"../testdata/python/simple/lcom_classes.py"}
	}
	return domain.LCOMRequest{
		Paths:           paths,
		OutputFormat:    domain.OutputFormatJSON,
		OutputWriter:    nil,
		MinLCOM:         0,
		MaxLCOM:         0,
		SortBy:          domain.SortByCohesion,
		LowThreshold:    2,
		MediumThreshold: 5,
		Recursive:       domain.BoolPtr(false),
	}
}

func TestNewLCOMService(t *testing.T) {
	svc := NewLCOMService()
	assert.NotNil(t, svc)
	assert.NotNil(t, svc.parser)
}

func TestLCOMService_Analyze(t *testing.T) {
	svc := NewLCOMService()
	ctx := context.Background()

	t.Run("successful analysis with LCOM test fixtures", func(t *testing.T) {
		req := newDefaultLCOMRequest("../testdata/python/simple/lcom_classes.py")

		response, err := svc.Analyze(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.GeneratedAt)
		assert.NotEmpty(t, response.Version)
		assert.NotNil(t, response.Config)
		assert.GreaterOrEqual(t, response.Summary.FilesAnalyzed, 1)

		// Verify we found multiple classes
		assert.Greater(t, len(response.Classes), 0)

		// Verify response structure
		for _, class := range response.Classes {
			assert.NotEmpty(t, class.Name)
			assert.NotEmpty(t, class.FilePath)
			assert.GreaterOrEqual(t, class.Metrics.LCOM4, 1)
			assert.Contains(t, []domain.RiskLevel{
				domain.RiskLevelLow,
				domain.RiskLevelMedium,
				domain.RiskLevelHigh,
			}, class.RiskLevel)
		}
	})

	t.Run("analyze file with no classes returns warning", func(t *testing.T) {
		req := newDefaultLCOMRequest("../testdata/python/simple/functions.py")

		response, err := svc.Analyze(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Empty(t, response.Classes)
		assert.Contains(t, response.Warnings, "No classes found to analyze")
	})

	t.Run("non-existent file returns error", func(t *testing.T) {
		req := newDefaultLCOMRequest("/nonexistent/file.py")

		response, err := svc.Analyze(ctx, req)

		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Errors)
	})

	t.Run("filter by MinLCOM", func(t *testing.T) {
		req := newDefaultLCOMRequest("../testdata/python/simple/lcom_classes.py")
		req.MinLCOM = 2

		response, err := svc.Analyze(ctx, req)

		require.NoError(t, err)
		for _, class := range response.Classes {
			assert.GreaterOrEqual(t, class.Metrics.LCOM4, 2)
		}
	})

	t.Run("filter by MaxLCOM", func(t *testing.T) {
		req := newDefaultLCOMRequest("../testdata/python/simple/lcom_classes.py")
		req.MaxLCOM = 1

		response, err := svc.Analyze(ctx, req)

		require.NoError(t, err)
		for _, class := range response.Classes {
			assert.LessOrEqual(t, class.Metrics.LCOM4, 1)
		}
	})

	t.Run("sort by cohesion (LCOM4 descending)", func(t *testing.T) {
		req := newDefaultLCOMRequest("../testdata/python/simple/lcom_classes.py")
		req.SortBy = domain.SortByCohesion

		response, err := svc.Analyze(ctx, req)

		require.NoError(t, err)
		if len(response.Classes) > 1 {
			for i := 0; i < len(response.Classes)-1; i++ {
				assert.GreaterOrEqual(t, response.Classes[i].Metrics.LCOM4, response.Classes[i+1].Metrics.LCOM4)
			}
		}
	})

	t.Run("sort by name", func(t *testing.T) {
		req := newDefaultLCOMRequest("../testdata/python/simple/lcom_classes.py")
		req.SortBy = domain.SortByName

		response, err := svc.Analyze(ctx, req)

		require.NoError(t, err)
		if len(response.Classes) > 1 {
			for i := 0; i < len(response.Classes)-1; i++ {
				assert.LessOrEqual(t, response.Classes[i].Name, response.Classes[i+1].Name)
			}
		}
	})

	t.Run("cancelled context returns error", func(t *testing.T) {
		cancelCtx, cancel := context.WithCancel(ctx)
		cancel()

		req := newDefaultLCOMRequest("../testdata/python/simple/lcom_classes.py")

		_, err := svc.Analyze(cancelCtx, req)
		assert.Error(t, err)
	})
}

func TestLCOMService_Summary(t *testing.T) {
	svc := NewLCOMService()
	ctx := context.Background()

	req := newDefaultLCOMRequest("../testdata/python/simple/lcom_classes.py")
	response, err := svc.Analyze(ctx, req)

	require.NoError(t, err)
	assert.NotNil(t, response)

	summary := response.Summary
	assert.Greater(t, summary.TotalClasses, 0)
	assert.GreaterOrEqual(t, summary.AverageLCOM, 0.0)
	assert.GreaterOrEqual(t, summary.MinLCOM, 0)
	assert.GreaterOrEqual(t, summary.MaxLCOM, summary.MinLCOM)
	assert.Equal(t, summary.TotalClasses, summary.LowRiskClasses+summary.MediumRiskClasses+summary.HighRiskClasses)
	assert.NotEmpty(t, summary.LCOMDistribution)
	assert.NotEmpty(t, summary.LeastCohesiveClasses)
}
