package app

import (
	"context"
	"os"
	"path/filepath"
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

func TestAnalyzeUseCase_Execute_DisablesComplexityFromConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".pyscn.toml")
	configContent := `[complexity]
enabled = false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	config := AnalyzeUseCaseConfig{
		ConfigFile:      configPath,
		SkipComplexity:  false,
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      true,
		MinComplexity:   1,
		MinSeverity:     domain.DeadCodeSeverityWarning,
		CloneSimilarity: 0.8,
	}

	builder := NewAnalyzeUseCaseBuilder()
	builder.WithFileReader(service.NewFileReader())
	builder.WithFormatter(service.NewAnalyzeFormatter())
	builder.WithProgressManager(service.NewProgressManager())
	builder.WithParallelExecutor(service.NewParallelExecutor())
	builder.WithErrorCategorizer(service.NewErrorCategorizer())
	builder.WithComplexityUseCase(NewComplexityUseCase(
		service.NewComplexityService(),
		service.NewFileReader(),
		service.NewOutputFormatter(),
		service.NewConfigurationLoader(),
	))

	useCase, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build AnalyzeUseCase: %v", err)
	}

	response, err := useCase.Execute(context.Background(), config, []string{"../testdata/python/simple"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if response.Summary.ComplexityEnabled {
		t.Errorf("Expected complexity to be disabled, got %v", response.Summary.ComplexityEnabled)
	}
	if response.Complexity != nil {
		t.Errorf("Expected no complexity response, got %+v", response.Complexity)
	}
}

func TestAnalyzeUseCase_LoadExecutionConfig(t *testing.T) {
	useCase := &AnalyzeUseCase{}

	t.Run("uses analyze defaults without config file", func(t *testing.T) {
		executionCfg, err := useCase.loadExecutionConfig("")
		if err != nil {
			t.Fatalf("loadExecutionConfig returned error: %v", err)
		}

		if !executionCfg.complexityEnabled {
			t.Error("Expected complexity to be enabled by default")
		}
		if !executionCfg.reportUnchanged {
			t.Error("Expected report_unchanged to be true by default")
		}
		if executionCfg.complexityLowThreshold != domain.DefaultComplexityLowThreshold {
			t.Errorf("Expected low threshold %d, got %d", domain.DefaultComplexityLowThreshold, executionCfg.complexityLowThreshold)
		}
		if executionCfg.complexityMediumThreshold != domain.DefaultComplexityMediumThreshold {
			t.Errorf("Expected medium threshold %d, got %d", domain.DefaultComplexityMediumThreshold, executionCfg.complexityMediumThreshold)
		}
		if executionCfg.complexityMaxComplexity != domain.DefaultComplexityMaxLimit {
			t.Errorf("Expected max complexity %d, got %d", domain.DefaultComplexityMaxLimit, executionCfg.complexityMaxComplexity)
		}
		if len(executionCfg.includePatterns) != 2 || executionCfg.includePatterns[1] != "*.pyi" {
			t.Errorf("Expected default include patterns to include .pyi files, got %v", executionCfg.includePatterns)
		}
		defaultCloneReq := domain.DefaultCloneRequest()
		if executionCfg.lshEnabled != defaultCloneReq.LSHEnabled {
			t.Errorf("Expected default LSH enabled %q, got %q", defaultCloneReq.LSHEnabled, executionCfg.lshEnabled)
		}
		if executionCfg.lshThreshold != defaultCloneReq.LSHAutoThreshold {
			t.Errorf("Expected default LSH threshold %d, got %d", defaultCloneReq.LSHAutoThreshold, executionCfg.lshThreshold)
		}
	})

	t.Run("uses resolved config values when config file exists", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".pyscn.toml")
		configContent := `[analysis]
include_patterns = ["pkg/**/*.py"]
exclude_patterns = ["tests/**/*.py"]
recursive = false

[complexity]
enabled = false
report_unchanged = false
low_threshold = 3
medium_threshold = 7
max_complexity = 11

[clones]
lsh_enabled = "true"
lsh_auto_threshold = 123
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		executionCfg, err := useCase.loadExecutionConfig(configPath)
		if err != nil {
			t.Fatalf("loadExecutionConfig returned error: %v", err)
		}

		if executionCfg.complexityEnabled {
			t.Error("Expected complexity to be disabled")
		}
		if executionCfg.reportUnchanged {
			t.Error("Expected report_unchanged to be false")
		}
		if executionCfg.complexityLowThreshold != 3 {
			t.Errorf("Expected low threshold 3, got %d", executionCfg.complexityLowThreshold)
		}
		if executionCfg.complexityMediumThreshold != 7 {
			t.Errorf("Expected medium threshold 7, got %d", executionCfg.complexityMediumThreshold)
		}
		if executionCfg.complexityMaxComplexity != 11 {
			t.Errorf("Expected max complexity 11, got %d", executionCfg.complexityMaxComplexity)
		}
		if executionCfg.recursive {
			t.Error("Expected recursive to be false")
		}
		if len(executionCfg.includePatterns) != 1 || executionCfg.includePatterns[0] != "pkg/**/*.py" {
			t.Errorf("Expected custom include patterns, got %v", executionCfg.includePatterns)
		}
		if len(executionCfg.excludePatterns) != 1 || executionCfg.excludePatterns[0] != "tests/**/*.py" {
			t.Errorf("Expected custom exclude patterns, got %v", executionCfg.excludePatterns)
		}
		if executionCfg.lshEnabled != "true" {
			t.Errorf("Expected LSH enabled to be %q, got %q", "true", executionCfg.lshEnabled)
		}
		if executionCfg.lshThreshold != 123 {
			t.Errorf("Expected LSH threshold 123, got %d", executionCfg.lshThreshold)
		}
	})
}
