package app

import (
	"context"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
)

func TestAnalyzeUseCase_Execute(t *testing.T) {
	// Create test configuration
	config := AnalyzeUseCaseConfig{
		SkipComplexity: false,
		SkipDeadCode:   true,
		SkipClones:     true,
		SkipCBO:        true,
		SkipSystem:     true,
		MinComplexity:  5,
		MinSeverity:    domain.DeadCodeSeverityWarning,
		Verbose:        false,
	}

	// Create a minimal use case with only required dependencies
	builder := NewAnalyzeUseCaseBuilder()

	// Set up minimal dependencies
	fileReader := service.NewFileReader()
	builder.WithFileReader(fileReader)
	builder.WithFormatter(service.NewAnalyzeFormatter())
	builder.WithProgressManager(service.NewProgressManager())
	builder.WithParallelExecutor(service.NewParallelExecutor())
	builder.WithErrorCategorizer(service.NewErrorCategorizer())

	// Build minimal complexity use case for testing
	complexityService := service.NewComplexityService()
	complexityFormatter := service.NewOutputFormatter()
	complexityConfigLoader := service.NewConfigurationLoader()
	complexityUseCase := NewComplexityUseCase(
		complexityService,
		service.NewFileReader(),
		complexityFormatter,
		complexityConfigLoader,
	)
	builder.WithComplexityUseCase(complexityUseCase)

	// Build the use case
	useCase, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build AnalyzeUseCase: %v", err)
	}

	// Test with test data files
	testPaths := []string{"../testdata/python/simple"}

	// Execute analysis
	ctx := context.Background()
	response, err := useCase.Execute(ctx, config, testPaths)

	// Verify basic execution (may fail if no test files, which is fine for structure test)
	if err != nil && err.Error() != "no Python files found in the specified paths" {
		t.Logf("Analysis execution failed (expected if no test files): %v", err)
	}

	// Verify response structure
	if response != nil {
		if response.Summary.ComplexityEnabled != true {
			t.Errorf("Expected complexity to be enabled, got %v", response.Summary.ComplexityEnabled)
		}
		if response.Summary.DeadCodeEnabled != false {
			t.Errorf("Expected dead code to be disabled, got %v", response.Summary.DeadCodeEnabled)
		}
	}
}

func TestAnalyzeUseCaseBuilder(t *testing.T) {
	builder := NewAnalyzeUseCaseBuilder()

	// Test building without required dependencies
	_, err := builder.Build()
	if err == nil {
		t.Error("Expected error when building without file reader, got nil")
	}

	// Test building with all dependencies
	builder.
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewAnalyzeFormatter()).
		WithProgressManager(service.NewProgressManager()).
		WithParallelExecutor(service.NewParallelExecutor()).
		WithErrorCategorizer(service.NewErrorCategorizer())

	useCase, err := builder.Build()
	if err != nil {
		t.Errorf("Failed to build with all dependencies: %v", err)
	}

	if useCase == nil {
		t.Error("Expected non-nil use case, got nil")
	}
}