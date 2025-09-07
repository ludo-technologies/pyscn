package service

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ludo-technologies/pyscn/domain"
)

// TestFileReader_Basic tests basic FileReader functionality
func TestFileReader_Basic(t *testing.T) {
	reader := NewFileReader()
	
	t.Run("NewFileReader creates instance", func(t *testing.T) {
		assert.NotNil(t, reader)
	})
	
	t.Run("IsValidPythonFile recognizes .py files", func(t *testing.T) {
		assert.True(t, reader.IsValidPythonFile("test.py"))
		assert.True(t, reader.IsValidPythonFile("module.pyi"))
		assert.False(t, reader.IsValidPythonFile("test.txt"))
		assert.False(t, reader.IsValidPythonFile("README.md"))
	})
	
	t.Run("FileExists handles non-existent files", func(t *testing.T) {
		exists, err := reader.FileExists("/path/that/does/not/exist")
		assert.NoError(t, err)
		assert.False(t, exists)
	})
}

// TestOutputFormatter_Basic tests basic OutputFormatter functionality
func TestOutputFormatter_Basic(t *testing.T) {
	formatter := NewOutputFormatter()
	
	t.Run("NewOutputFormatter creates instance", func(t *testing.T) {
		assert.NotNil(t, formatter)
	})
	
	t.Run("Format handles unsupported format", func(t *testing.T) {
		response := &domain.ComplexityResponse{
			Functions: []domain.FunctionComplexity{},
			Summary: domain.ComplexitySummary{
				TotalFunctions: 0,
			},
		}
		
		_, err := formatter.Format(response, domain.OutputFormat("unsupported"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported format")
	})
}

// TestComplexityService_Basic tests basic ComplexityService functionality
func TestComplexityService_Basic(t *testing.T) {
	service := NewComplexityService()
	
	t.Run("NewComplexityService creates instance", func(t *testing.T) {
		assert.NotNil(t, service)
		assert.NotNil(t, service.parser)
	})
	
	// Test sorting function
	t.Run("sortFunctions handles empty slice", func(t *testing.T) {
		var functions []domain.FunctionComplexity
		result := service.sortFunctions(functions, domain.SortByComplexity)
		assert.Equal(t, 0, len(result))
	})
	
	// Test filtering function  
	t.Run("filterFunctions handles empty slice", func(t *testing.T) {
		var functions []domain.FunctionComplexity
		req := domain.ComplexityRequest{MinComplexity: 1, MaxComplexity: 10}
		result := service.filterFunctions(functions, req)
		assert.Equal(t, 0, len(result))
	})
	
	// Test summary generation
	t.Run("generateSummary handles empty data", func(t *testing.T) {
		var functions []domain.FunctionComplexity
		req := domain.ComplexityRequest{LowThreshold: 3, MediumThreshold: 7}
		summary := service.generateSummary(functions, 0, req)
		
		assert.Equal(t, 0, summary.TotalFunctions)
		assert.Equal(t, 0.0, summary.AverageComplexity)
	})
}

// TestDeadCodeService_Basic tests basic DeadCodeService functionality
func TestDeadCodeService_Basic(t *testing.T) {
	service := NewDeadCodeService()
	
	t.Run("NewDeadCodeService creates instance", func(t *testing.T) {
		assert.NotNil(t, service)
		assert.NotNil(t, service.parser)
	})
	
	// Test sorting function
	t.Run("sortFiles handles empty slice", func(t *testing.T) {
		var files []domain.FileDeadCode
		result := service.sortFiles(files, domain.DeadCodeSortByFile)
		assert.Equal(t, 0, len(result))
	})
	
	// Test filtering function
	t.Run("filterFiles handles empty slice", func(t *testing.T) {
		var files []domain.FileDeadCode
		req := domain.DeadCodeRequest{} // Use default values
		result := service.filterFiles(files, req)
		assert.Equal(t, 0, len(result))
	})
	
	// Test summary generation
	t.Run("generateSummary handles empty data", func(t *testing.T) {
		var files []domain.FileDeadCode
		req := domain.DeadCodeRequest{}
		summary := service.generateSummary(files, 0, req)
		
		assert.Equal(t, 0, summary.TotalFindings)
	})
}

// TestCloneService_Basic tests basic CloneService functionality
func TestCloneService_Basic(t *testing.T) {
	service := NewCloneService()
	
	t.Run("NewCloneService creates instance", func(t *testing.T) {
		assert.NotNil(t, service)
	})
	
	// Test filtering function
	t.Run("filterClonePairs handles empty slice", func(t *testing.T) {
		var pairs []*domain.ClonePair
		req := &domain.CloneRequest{
			MinLines: 3,
			MinNodes: 5,
			SimilarityThreshold: 0.8,
		}
		result := service.filterClonePairs(pairs, req)
		assert.Equal(t, 0, len(result))
	})
	
	// Test statistics creation
	t.Run("createStatistics handles empty data", func(t *testing.T) {
		var clones []*domain.Clone
		var pairs []*domain.ClonePair
		var groups []*domain.CloneGroup
		
		stats := service.createStatistics(clones, pairs, groups, 0, 0)
		
		assert.Equal(t, 0, stats.TotalClones)
		assert.Equal(t, 0, stats.TotalClonePairs)
		assert.Equal(t, 0, stats.TotalCloneGroups)
		assert.Equal(t, 0, stats.FilesAnalyzed)
		assert.Equal(t, 0, stats.LinesAnalyzed)
	})
	
	// Test basic validation
	t.Run("ComputeSimilarity validates inputs", func(t *testing.T) {
		ctx := context.Background()
		
		// Test empty fragment1
		_, err := service.ComputeSimilarity(ctx, "", "print('hello')")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fragments cannot be empty")
		
		// Test empty fragment2
		_, err = service.ComputeSimilarity(ctx, "print('hello')", "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fragments cannot be empty")
		
		// Test context functionality (implementation may handle context.TODO())
		_, _ = service.ComputeSimilarity(context.TODO(), "print('hello')", "print('world')")
		// This may succeed or fail depending on implementation - we're testing the method exists
		
		// Test large fragments
		largeFragment := strings.Repeat("x", 2*1024*1024) // 2MB
		_, err = service.ComputeSimilarity(ctx, largeFragment, "print('hello')")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "fragment size exceeds maximum allowed size")
	})
	
	// Test DetectClones validation
	t.Run("DetectClones validates inputs", func(t *testing.T) {
		ctx := context.Background()
		
		// Test DetectClones functionality
		_, _ = service.DetectClones(context.TODO(), &domain.CloneRequest{})
		// This may succeed or fail depending on implementation - we're testing the method exists
		
		// Test nil request
		_, err := service.DetectClones(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "clone request cannot be nil")
	})
}

// TestServiceIntegration_Basic tests basic service integration
func TestServiceIntegration_Basic(t *testing.T) {
	t.Run("All services can be created", func(t *testing.T) {
		complexityService := NewComplexityService()
		deadCodeService := NewDeadCodeService()
		cloneService := NewCloneService()
		fileReader := NewFileReader()
		outputFormatter := NewOutputFormatter()
		
		assert.NotNil(t, complexityService)
		assert.NotNil(t, deadCodeService)
		assert.NotNil(t, cloneService)
		assert.NotNil(t, fileReader)
		assert.NotNil(t, outputFormatter)
	})
}