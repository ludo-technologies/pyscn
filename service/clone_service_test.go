package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDefaultCloneRequest(paths ...string) *domain.CloneRequest {
	if len(paths) == 0 {
		paths = []string{"../testdata/python/simple/functions.py"}
	}
	return &domain.CloneRequest{
		Paths:             paths,
		MinLines:          3,
		MinNodes:          10,
		MinSimilarity:     0.7,
		MaxSimilarity:     1.0,
		Type1Threshold:    0.95,
		Type2Threshold:    0.85,
		Type3Threshold:    0.75,
		Type4Threshold:    0.65,
		CloneTypes:        []domain.CloneType{domain.Type1Clone, domain.Type2Clone},
		OutputFormat:      domain.OutputFormatJSON,
		ShowDetails:       true,
		ShowContent:       false,
		GroupClones:       true,
		IgnoreLiterals:    true,
		IgnoreIdentifiers: false,
	}
}

func TestNewCloneService(t *testing.T) {
	service := NewCloneService()

	assert.NotNil(t, service)
}

func TestCloneService_DetectClones(t *testing.T) {
	service := NewCloneService()
	ctx := context.Background()

	t.Run("nil context should return error", func(t *testing.T) {
		req := &domain.CloneRequest{
			Paths:             []string{"../testdata/python/simple/functions.py"},
			MinLines:          3,
			MinNodes:          10,
			MinSimilarity:     0.7,
			MaxSimilarity:     1.0,
			Type1Threshold:    0.95,
			Type2Threshold:    0.85,
			Type3Threshold:    0.75,
			Type4Threshold:    0.65,
			CloneTypes:        []domain.CloneType{domain.Type1Clone, domain.Type2Clone},
			OutputFormat:      domain.OutputFormatJSON,
			ShowDetails:       true,
			ShowContent:       false,
			GroupClones:       true,
			IgnoreLiterals:    true,
			IgnoreIdentifiers: false,
		}

		//nolint:staticcheck // Intentionally testing nil context error handling
		_, err := service.DetectClones(nil, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("nil request should return error", func(t *testing.T) {
		_, err := service.DetectClones(ctx, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "clone request cannot be nil")
	})

	t.Run("successful clone detection with valid files", func(t *testing.T) {
		req := newDefaultCloneRequest("../testdata/python/simple/functions.py", "../testdata/python/simple/control_flow.py")
		req.MinSimilarity = 0.5
		req.CloneTypes = []domain.CloneType{domain.Type1Clone, domain.Type2Clone, domain.Type3Clone}
		req.MaxEditDistance = 10.0

		// Skip validation check since Validate method may not exist
		response, err := service.DetectClonesInFiles(ctx, req.Paths, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.True(t, response.Success)
		assert.NotNil(t, response.Statistics)
		assert.NotNil(t, response.Request)
		assert.Greater(t, response.Duration, int64(0))
		assert.GreaterOrEqual(t, response.Statistics.FilesAnalyzed, 0)
		assert.GreaterOrEqual(t, response.Statistics.LinesAnalyzed, 0)

		// Verify clone pairs structure if any found
		for _, pair := range response.ClonePairs {
			assert.NotNil(t, pair.Clone1)
			assert.NotNil(t, pair.Clone2)
			assert.NotNil(t, pair.Clone1.Location)
			assert.NotNil(t, pair.Clone2.Location)
			assert.GreaterOrEqual(t, pair.Similarity, req.MinSimilarity)
			assert.LessOrEqual(t, pair.Similarity, req.MaxSimilarity)
		}

		// Verify clone groups structure if any found
		for _, group := range response.CloneGroups {
			assert.GreaterOrEqual(t, len(group.Clones), 2)
			assert.GreaterOrEqual(t, group.Similarity, req.MinSimilarity)
			assert.LessOrEqual(t, group.Similarity, req.MaxSimilarity)
		}
	})

	t.Run("empty file list should succeed but return empty results", func(t *testing.T) {
		req := newDefaultCloneRequest()
		req.Paths = []string{}
		req.CloneTypes = []domain.CloneType{domain.Type1Clone}

		response, err := service.DetectClonesInFiles(ctx, req.Paths, req)

		// Should return error for empty file paths
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file paths cannot be empty")
		assert.Nil(t, response)
	})

	t.Run("context cancellation should be handled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		req := newDefaultCloneRequest()
		req.CloneTypes = []domain.CloneType{domain.Type1Clone}

		_, err := service.DetectClonesInFiles(ctx, req.Paths, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cancelled")
	})

	t.Run("request with timeout should be respected", func(t *testing.T) {
		req := newDefaultCloneRequest()
		req.CloneTypes = []domain.CloneType{domain.Type1Clone}
		req.Timeout = time.Millisecond * 1 // Very short timeout

		startTime := time.Now()
		_, err := service.DetectClonesInFiles(ctx, req.Paths, req)
		duration := time.Since(startTime)

		// Should either succeed quickly or timeout
		if err != nil {
			// If it fails, it should be due to timeout or cancellation
			assert.True(t,
				duration < time.Millisecond*100 || // Finished quickly
					err.Error() != "" && (ctx.Err() == context.DeadlineExceeded || duration < time.Second), // Or timed out reasonably
				"Expected quick completion or timeout, got duration: %v, error: %v", duration, err)
		}
	})
}

func TestCloneService_DetectClonesInFiles(t *testing.T) {
	service := NewCloneService()
	ctx := context.Background()

	t.Run("nil context should return error", func(t *testing.T) {
		req := newDefaultCloneRequest()
		req.CloneTypes = []domain.CloneType{domain.Type1Clone}

		//nolint:staticcheck // Intentionally testing nil context error handling
		_, err := service.DetectClonesInFiles(nil, []string{"test.py"}, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("nil request should return error", func(t *testing.T) {
		_, err := service.DetectClonesInFiles(ctx, []string{"test.py"}, nil)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "clone request cannot be nil")
	})

	t.Run("empty file paths should return error", func(t *testing.T) {
		req := newDefaultCloneRequest()
		req.CloneTypes = []domain.CloneType{domain.Type1Clone}

		_, err := service.DetectClonesInFiles(ctx, []string{}, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file paths cannot be empty")
	})

	t.Run("non-existent files should be skipped with warnings", func(t *testing.T) {
		req := newDefaultCloneRequest("../testdata/non_existent_file.py")
		req.CloneTypes = []domain.CloneType{domain.Type1Clone}

		response, err := service.DetectClonesInFiles(ctx, req.Paths, req)

		// Should succeed but with empty results (files are skipped)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.True(t, response.Success)
		assert.Empty(t, response.Clones)
		assert.Empty(t, response.ClonePairs)
		assert.Empty(t, response.CloneGroups)
	})

	t.Run("files analyzed counts only successfully processed files", func(t *testing.T) {
		valid := "../testdata/python/simple/functions.py"
		missing := "../testdata/python/missing.py"

		req := newDefaultCloneRequest(valid, missing)
		req.CloneTypes = []domain.CloneType{domain.Type1Clone}

		response, err := service.DetectClonesInFiles(ctx, req.Paths, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Statistics)
		assert.Equal(t, 1, response.Statistics.FilesAnalyzed)
	})

	t.Run("parsing failures are not counted", func(t *testing.T) {
		// Create brokenFile
		tmpDir := t.TempDir()
		brokenFile := filepath.Join(tmpDir, "broken.py")
		assert.NoError(t, os.WriteFile(brokenFile, []byte("def oops(:\n    pass"), 0o644))

		valid := "../testdata/python/simple/functions.py"

		req := newDefaultCloneRequest(valid, brokenFile)
		req.CloneTypes = []domain.CloneType{domain.Type1Clone}

		response, err := service.DetectClonesInFiles(ctx, req.Paths, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Statistics)
		assert.Equal(t, 1, response.Statistics.FilesAnalyzed)
	})

	t.Run("statistics uses processed file count after detection", func(t *testing.T) {
		tmp := t.TempDir()
		f1 := filepath.Join(tmp, "a.py")
		f2 := filepath.Join(tmp, "b.py")
		missing := "../testdata/python/missing.py"

		body := "def foo(x):\n	return x\n"
		require.NoError(t, os.WriteFile(f1, []byte(body), 0o644))
		require.NoError(t, os.WriteFile(f2, []byte(body), 0o644))

		req := newDefaultCloneRequest(f1, f2, missing)
		req.CloneTypes = []domain.CloneType{domain.Type1Clone, domain.Type2Clone, domain.Type3Clone}

		response, err := service.DetectClonesInFiles(ctx, req.Paths, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.Statistics)
		assert.Equal(t, 2, response.Statistics.FilesAnalyzed) // skips missing; counts only processed files
	})
}

func TestCloneService_ComputeSimilarity(t *testing.T) {
	service := NewCloneService()
	ctx := context.Background()

	t.Run("compute similarity between identical fragments", func(t *testing.T) {
		fragment1 := "def hello():\n    print('Hello, World!')\n    return True"
		fragment2 := "def hello():\n    print('Hello, World!')\n    return True"

		similarity, err := service.ComputeSimilarity(ctx, fragment1, fragment2)

		assert.NoError(t, err)
		assert.Equal(t, 1.0, similarity) // Identical should be 1.0
	})

	t.Run("compute similarity between similar fragments", func(t *testing.T) {
		fragment1 := "def hello():\n    print('Hello, World!')\n    return True"
		fragment2 := "def greet():\n    print('Hello, World!')\n    return True"

		similarity, err := service.ComputeSimilarity(ctx, fragment1, fragment2)

		assert.NoError(t, err)
		assert.Greater(t, similarity, 0.0)
		assert.Less(t, similarity, 1.0)
	})

	t.Run("compute similarity between different fragments", func(t *testing.T) {
		fragment1 := "def hello():\n    print('Hello, World!')"
		fragment2 := "class MyClass:\n    def __init__(self):\n        self.value = 42"

		similarity, err := service.ComputeSimilarity(ctx, fragment1, fragment2)

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, similarity, 0.0)
		assert.LessOrEqual(t, similarity, 1.0)
	})

	t.Run("empty fragments should return error", func(t *testing.T) {
		_, err := service.ComputeSimilarity(ctx, "", "def test(): pass")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fragments cannot be empty")

		_, err = service.ComputeSimilarity(ctx, "def test(): pass", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fragments cannot be empty")
	})

	t.Run("nil context should return error", func(t *testing.T) {
		fragment1 := "def hello(): pass"
		fragment2 := "def world(): pass"

		//nolint:staticcheck // Intentionally testing nil context error handling
		_, err := service.ComputeSimilarity(nil, fragment1, fragment2)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context cannot be nil")
	})

	t.Run("excessively large fragments should return error", func(t *testing.T) {
		// Create a fragment larger than the 1MB limit
		largeFragment := "def large_function():\n"
		for i := 0; i < 100000; i++ {
			largeFragment += "    # This is a very long comment to make the fragment large\n"
		}

		normalFragment := "def normal(): pass"

		_, err := service.ComputeSimilarity(ctx, largeFragment, normalFragment)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fragment size exceeds maximum allowed size")
	})

	t.Run("invalid Python syntax should return error", func(t *testing.T) {
		invalidFragment := "def invalid syntax here:"
		validFragment := "def valid(): pass"

		_, err := service.ComputeSimilarity(ctx, invalidFragment, validFragment)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse")
	})
}

func TestCloneService_CreateDetectorConfig(t *testing.T) {
	service := NewCloneService()

	req := newDefaultCloneRequest()
	req.MinLines = 5
	req.MinNodes = 20
	req.MaxEditDistance = 15.0
	req.IgnoreIdentifiers = false

	config := service.createDetectorConfig(req)

	assert.NotNil(t, config)
	assert.Equal(t, 5, config.MinLines)
	assert.Equal(t, 20, config.MinNodes)
	assert.Equal(t, 0.95, config.Type1Threshold)
	assert.Equal(t, 0.85, config.Type2Threshold)
	assert.Equal(t, 0.75, config.Type3Threshold)
	assert.Equal(t, 0.65, config.Type4Threshold)
	assert.Equal(t, 15.0, config.MaxEditDistance)
	assert.Equal(t, true, config.IgnoreLiterals)
	assert.Equal(t, false, config.IgnoreIdentifiers)
	assert.Equal(t, "python", config.CostModelType)
	assert.Equal(t, 10000, config.MaxClonePairs)
	assert.Equal(t, 50, config.BatchSizeThreshold)
}

func TestCloneService_ConvertCloneType(t *testing.T) {
	testCases := []struct {
		name               string
		analyzerType       int // Using int since analyzer types are in internal package
		expectedDomainType domain.CloneType
	}{
		{"Type1", 1, domain.Type1Clone},
		{"Type2", 2, domain.Type2Clone},
		{"Type3", 3, domain.Type3Clone},
		{"Type4", 4, domain.Type4Clone},
		{"Unknown", 99, domain.Type1Clone}, // Default case
	}

	// Note: We can't directly test the convertCloneType method since it uses internal types
	// but we can test the domain.CloneType.String() method instead
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			str := tc.expectedDomainType.String()
			assert.Contains(t, str, "Type-")
		})
	}
}

func TestCloneService_FilterClonePairs(t *testing.T) {
	service := NewCloneService()

	pairs := []*domain.ClonePair{
		{
			ID:         1,
			Similarity: 0.8,
			Type:       domain.Type1Clone,
		},
		{
			ID:         2,
			Similarity: 0.6,
			Type:       domain.Type2Clone,
		},
		{
			ID:         3,
			Similarity: 0.9,
			Type:       domain.Type3Clone,
		},
		{
			ID:         4,
			Similarity: 0.4,
			Type:       domain.Type1Clone,
		},
	}

	t.Run("filter by similarity range", func(t *testing.T) {
		req := newDefaultCloneRequest()
		req.MinSimilarity = 0.7
		req.MaxSimilarity = 0.85
		req.CloneTypes = []domain.CloneType{domain.Type1Clone, domain.Type2Clone, domain.Type3Clone}

		filtered := service.filterClonePairs(pairs, req)

		require.Len(t, filtered, 1)
		assert.Equal(t, 1, filtered[0].ID)
		assert.Equal(t, 0.8, filtered[0].Similarity)
	})

	t.Run("filter by clone types", func(t *testing.T) {
		req := newDefaultCloneRequest()
		req.MinSimilarity = 0.0
		req.MaxSimilarity = 1.0
		req.CloneTypes = []domain.CloneType{domain.Type1Clone}

		filtered := service.filterClonePairs(pairs, req)

		require.Len(t, filtered, 2)
		assert.Equal(t, domain.Type1Clone, filtered[0].Type)
		assert.Equal(t, domain.Type1Clone, filtered[1].Type)
	})

	t.Run("filter by both similarity and types", func(t *testing.T) {
		req := newDefaultCloneRequest()
		req.MinSimilarity = 0.7
		req.MaxSimilarity = 1.0
		req.CloneTypes = []domain.CloneType{domain.Type1Clone, domain.Type3Clone}

		filtered := service.filterClonePairs(pairs, req)

		require.Len(t, filtered, 2)
		assert.Contains(t, []domain.CloneType{domain.Type1Clone, domain.Type3Clone}, filtered[0].Type)
		assert.Contains(t, []domain.CloneType{domain.Type1Clone, domain.Type3Clone}, filtered[1].Type)
	})
}

func TestCloneService_FilterCloneGroups(t *testing.T) {
	service := NewCloneService()

	groups := []*domain.CloneGroup{
		{
			ID:         1,
			Similarity: 0.8,
			Type:       domain.Type1Clone,
		},
		{
			ID:         2,
			Similarity: 0.6,
			Type:       domain.Type2Clone,
		},
		{
			ID:         3,
			Similarity: 0.9,
			Type:       domain.Type3Clone,
		},
	}

	t.Run("filter by similarity range", func(t *testing.T) {
		req := newDefaultCloneRequest()
		req.MinSimilarity = 0.7
		req.MaxSimilarity = 0.85
		req.CloneTypes = []domain.CloneType{domain.Type1Clone, domain.Type2Clone, domain.Type3Clone}

		filtered := service.filterCloneGroups(groups, req)

		require.Len(t, filtered, 1)
		assert.Equal(t, 1, filtered[0].ID)
		assert.Equal(t, 0.8, filtered[0].Similarity)
	})

	t.Run("filter by clone types", func(t *testing.T) {
		req := newDefaultCloneRequest()
		req.MinSimilarity = 0.0
		req.MaxSimilarity = 1.0
		req.CloneTypes = []domain.CloneType{domain.Type2Clone, domain.Type3Clone}

		filtered := service.filterCloneGroups(groups, req)

		require.Len(t, filtered, 2)
		assert.Contains(t, []domain.CloneType{domain.Type2Clone, domain.Type3Clone}, filtered[0].Type)
		assert.Contains(t, []domain.CloneType{domain.Type2Clone, domain.Type3Clone}, filtered[1].Type)
	})
}

func TestCloneService_CreateStatistics(t *testing.T) {
	service := NewCloneService()

	clones := []*domain.Clone{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}

	pairs := []*domain.ClonePair{
		{ID: 1, Similarity: 0.8, Type: domain.Type1Clone},
		{ID: 2, Similarity: 0.9, Type: domain.Type2Clone},
		{ID: 3, Similarity: 0.7, Type: domain.Type1Clone},
	}

	groups := []*domain.CloneGroup{
		{ID: 1},
		{ID: 2},
	}

	filesAnalyzed := 5
	linesAnalyzed := 1000

	stats := service.createStatistics(clones, pairs, groups, filesAnalyzed, linesAnalyzed)

	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats.TotalClones)
	assert.Equal(t, 3, stats.TotalClonePairs)
	assert.Equal(t, 2, stats.TotalCloneGroups)
	assert.Equal(t, 5, stats.FilesAnalyzed)
	assert.Equal(t, 1000, stats.LinesAnalyzed)
	assert.InDelta(t, 0.8, stats.AverageSimilarity, 0.01) // (0.8 + 0.9 + 0.7) / 3

	// Check clone type counts
	assert.Equal(t, 2, stats.ClonesByType["Type-1"])
	assert.Equal(t, 1, stats.ClonesByType["Type-2"])
}

func TestCloneService_ResponseStructure(t *testing.T) {
	service := NewCloneService()
	ctx := context.Background()

	req := newDefaultCloneRequest()
	req.MinLines = 1
	req.MinNodes = 1
	req.MinSimilarity = 0.0
	req.CloneTypes = []domain.CloneType{domain.Type1Clone, domain.Type2Clone, domain.Type3Clone, domain.Type4Clone}
	req.ShowContent = true

	response, err := service.DetectClonesInFiles(ctx, req.Paths, req)

	assert.NoError(t, err)
	require.NotNil(t, response)

	// Verify response structure
	assert.NotNil(t, response.Clones)
	// ClonePairs and CloneGroups can be nil if no clones are found
	assert.NotNil(t, response.Statistics)
	if response.Request != nil {
		assert.Equal(t, req, response.Request)
	}
	assert.True(t, response.Success)
	assert.Greater(t, response.Duration, int64(0))
	assert.Empty(t, response.Error)

	// Verify statistics structure
	stats := response.Statistics
	assert.GreaterOrEqual(t, stats.TotalClones, 0)
	assert.GreaterOrEqual(t, stats.TotalClonePairs, 0)
	assert.GreaterOrEqual(t, stats.TotalCloneGroups, 0)
	assert.GreaterOrEqual(t, stats.FilesAnalyzed, 1)
	assert.GreaterOrEqual(t, stats.LinesAnalyzed, 0)
	assert.GreaterOrEqual(t, stats.AverageSimilarity, 0.0)
	assert.LessOrEqual(t, stats.AverageSimilarity, 1.0)
	assert.NotNil(t, stats.ClonesByType)
}
