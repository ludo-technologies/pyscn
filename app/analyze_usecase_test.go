package app

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

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
	complexityUseCase := NewSnapshotComplexityUseCase(
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

func TestAnalyzeUseCaseBuilderRejectsStandaloneAnalyzerCollaborator(t *testing.T) {
	standaloneComplexity := NewComplexityUseCase(
		service.NewComplexityService(),
		service.NewFileReader(),
		service.NewOutputFormatter(),
		service.NewConfigurationLoader(),
	)

	_, err := NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithComplexityUseCase(standaloneComplexity).
		Build()
	if err == nil || !strings.Contains(err.Error(), "complexity use case requires a snapshot collaborator") {
		t.Fatalf("expected aggregate collaborator validation, got %v", err)
	}
}

func TestAnalyzeUseCase_Execute_ReportsProjectCoverageWithoutComplexity(t *testing.T) {
	projectDir := t.TempDir()
	validPath := filepath.Join(projectDir, "valid.py")
	brokenPath := filepath.Join(projectDir, "broken.py")
	if err := os.WriteFile(validPath, []byte("def valid():\n    return 1\n"), 0o644); err != nil {
		t.Fatalf("write valid Python source: %v", err)
	}
	if err := os.WriteFile(brokenPath, []byte("def broken(:\n    pass\n"), 0o644); err != nil {
		t.Fatalf("write broken Python source: %v", err)
	}

	deadCodeUseCase := NewSnapshotDeadCodeUseCase(
		service.NewDeadCodeService(),
		service.NewFileReader(),
		service.NewDeadCodeFormatter(),
		service.NewDeadCodeConfigurationLoader(),
	)
	useCase, err := NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithDeadCodeUseCase(deadCodeUseCase).
		Build()
	if err != nil {
		t.Fatalf("build analyze use case: %v", err)
	}

	response, err := useCase.Execute(context.Background(), AnalyzeUseCaseConfig{
		SkipComplexity:  true,
		SkipDeadCode:    false,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      true,
		SkipCommunities: true,
	}, []string{projectDir})
	if err != nil {
		t.Fatalf("execute dead-code-only analysis: %v", err)
	}

	if response.Summary.TotalFiles != 2 || response.Summary.AnalyzedFiles != 1 || response.Summary.SkippedFiles != 1 {
		t.Fatalf("expected coverage 2 total / 1 analyzed / 1 skipped, got %+v", response.Summary)
	}
	if response.Summary.HealthScore >= 100 || response.Summary.Grade == "A" {
		t.Fatalf("incomplete analysis must not receive a perfect grade, got %d/%s", response.Summary.HealthScore, response.Summary.Grade)
	}
	if len(response.Diagnostics) != 1 {
		t.Fatalf("expected one project diagnostic, got %+v", response.Diagnostics)
	}
	diagnostic := response.Diagnostics[0]
	if diagnostic.Code != domain.DiagnosticCodeParse || filepath.Base(diagnostic.FilePath) != "broken.py" {
		t.Fatalf("expected typed parse diagnostic for broken.py, got %+v", diagnostic)
	}
}

func TestAnalyzeUseCaseBuildResponsePreservesEveryTypedFailure(t *testing.T) {
	failures := []domain.AnalysisFailure{
		{Analysis: domain.AnalysisKindDeadCode, Code: domain.AnalysisFailureCodeExecution, FilePath: "a.py", Message: "first"},
		{Analysis: domain.AnalysisKindDeadCode, Code: domain.AnalysisFailureCodeExecution, FilePath: "b.py", Message: "second"},
	}
	tasks := []*AnalysisTask{{
		Name:    taskNameDeadCode,
		Kind:    domain.AnalysisKindDeadCode,
		Enabled: true,
		Result:  &domain.DeadCodeResponse{Failures: failures},
	}}

	response, err := (&AnalyzeUseCase{}).buildResponse(tasks, time.Now(), analysisPathIndex{reportedByIdentity: map[string]string{}}, domain.AnalysisCoverage{})
	if err != nil {
		t.Fatalf("build response: %v", err)
	}
	if !reflect.DeepEqual(response.Failures, failures) {
		t.Fatalf("expected lossless typed failures, got %+v", response.Failures)
	}
}

func TestAnalyzeUseCaseExecutePreservesAnalyzerErrorIdentity(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "source.py")
	if err := os.WriteFile(sourcePath, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	sentinel := errors.New("complexity backend unavailable")
	complexityUseCase := NewSnapshotComplexityUseCase(
		failingComplexityService{err: sentinel},
		service.NewFileReader(),
		service.NewOutputFormatter(),
		service.NewConfigurationLoader(),
	)
	useCase, err := NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithComplexityUseCase(complexityUseCase).
		Build()
	if err != nil {
		t.Fatalf("build analyze use case: %v", err)
	}

	response, err := useCase.Execute(context.Background(), AnalyzeUseCaseConfig{
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      true,
		SkipCommunities: true,
	}, []string{sourcePath})
	if err == nil {
		t.Fatal("expected aggregate analysis error")
	}
	if response == nil || len(response.Failures) != 1 {
		t.Fatalf("expected typed partial response, got %+v", response)
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected aggregate error to preserve analyzer identity, got %v", err)
	}
}

type failingComplexityService struct {
	err error
}

func (s failingComplexityService) Analyze(context.Context, domain.ComplexityRequest) (*domain.ComplexityResponse, error) {
	return nil, s.err
}

func (s failingComplexityService) AnalyzeFile(context.Context, string, domain.ComplexityRequest) (*domain.ComplexityResponse, error) {
	return nil, s.err
}

func (s failingComplexityService) AnalyzeSnapshot(context.Context, *service.ProjectSnapshot, domain.ComplexityRequest) (*domain.ComplexityResponse, error) {
	return nil, s.err
}

func TestAnalyzeUseCase_Execute_SystemGraphExcludesUnparsedFiles(t *testing.T) {
	projectDir := t.TempDir()
	validPath := filepath.Join(projectDir, "valid.py")
	brokenPath := filepath.Join(projectDir, "broken.py")
	if err := os.WriteFile(validPath, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("write valid Python source: %v", err)
	}
	if err := os.WriteFile(brokenPath, []byte("def broken(:\n"), 0o644); err != nil {
		t.Fatalf("write broken Python source: %v", err)
	}

	systemUseCase, err := NewSystemAnalysisUseCaseBuilder().
		WithGraphService(service.NewSystemAnalysisService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewSystemAnalysisFormatter()).
		WithConfigLoader(service.NewSystemAnalysisConfigurationLoader()).
		Build()
	if err != nil {
		t.Fatalf("build system analysis use case: %v", err)
	}
	useCase, err := NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithConfigLoader(service.NewAnalyzeConfigurationLoader()).
		WithSystemUseCase(systemUseCase).
		Build()
	if err != nil {
		t.Fatalf("build analyze use case: %v", err)
	}

	response, err := useCase.Execute(context.Background(), AnalyzeUseCaseConfig{
		SkipComplexity:  true,
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      false,
		SkipCommunities: true,
	}, []string{projectDir})
	if err != nil {
		t.Fatalf("execute system-only analysis: %v", err)
	}
	if response.System == nil {
		t.Fatal("expected system analysis response")
	}
	if response.System.Summary.TotalModules != 1 {
		t.Fatalf("expected only parsed modules in system graph, got %d", response.System.Summary.TotalModules)
	}
}

func TestAnalyzeUseCase_ExecuteWithOverridesHonorsExplicitDependencyGate(t *testing.T) {
	projectDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectDir, "a.py"), []byte("import b\n"), 0o644); err != nil {
		t.Fatalf("write a.py: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "b.py"), []byte("import a\n"), 0o644); err != nil {
		t.Fatalf("write b.py: %v", err)
	}
	configPath := filepath.Join(projectDir, ".pyscn.toml")
	if err := os.WriteFile(configPath, []byte("[system_analysis]\nenabled = false\n\n[dependencies]\nenabled = false\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	useCase := newModuleQualityAnalyzeUseCase(t)
	response, err := useCase.ExecuteWithOverrides(context.Background(), AnalyzeUseCaseConfig{
		ConfigFile:              configPath,
		SkipComplexity:          true,
		SkipDeadCode:            true,
		SkipClones:              true,
		SkipCBO:                 true,
		SkipLCOM:                true,
		SkipCommunities:         true,
		SkipCommunitiesExplicit: true,
		SelectAnalysesUsed:      true,
	}, []string{projectDir}, AnalyzeRequestOverrides{
		SystemEnabled:             domain.BoolPtr(true),
		SystemAnalyzeDependencies: domain.BoolPtr(true),
		SystemAnalyzeArchitecture: domain.BoolPtr(false),
	})
	if err != nil {
		t.Fatalf("execute dependency gate: %v", err)
	}
	if response.System == nil || response.System.DependencyAnalysis == nil || response.System.DependencyAnalysis.CircularDependencies == nil || !response.System.DependencyAnalysis.CircularDependencies.HasCircularDependencies {
		t.Fatalf("expected circular dependency result, got %+v", response.System)
	}
	if response.System.ArchitectureAnalysis != nil || response.Summary.ArchEnabled {
		t.Fatalf("dependency-only execution must not run architecture analysis: %+v", response.Summary)
	}
}

func newModuleQualityAnalyzeUseCase(t *testing.T) *AnalyzeUseCase {
	t.Helper()

	systemUseCase, err := NewSystemAnalysisUseCaseBuilder().
		WithGraphService(service.NewSystemAnalysisService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewSystemAnalysisFormatter()).
		WithConfigLoader(service.NewSystemAnalysisConfigurationLoader()).
		Build()
	if err != nil {
		t.Fatalf("build system analysis use case: %v", err)
	}

	useCase, err := NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewAnalyzeFormatter()).
		WithProgressManager(service.NewProgressManager()).
		WithParallelExecutor(service.NewParallelExecutor()).
		WithErrorCategorizer(service.NewErrorCategorizer()).
		WithSystemUseCase(systemUseCase).
		WithComplexityUseCase(NewSnapshotComplexityUseCase(
			service.NewComplexityService(),
			service.NewFileReader(),
			service.NewOutputFormatter(),
			service.NewConfigurationLoader(),
		)).
		Build()
	if err != nil {
		t.Fatalf("build analyze use case: %v", err)
	}

	return useCase
}

func TestAnalyzeUseCase_Execute_PublishesModuleQuality(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)
	sourcePath := "hotspot.py"
	source := `def hotspot(value):
	if value > 10:
		return value
	return 0
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("write Python source: %v", err)
	}
	useCase := newModuleQualityAnalyzeUseCase(t)

	response, err := useCase.Execute(context.Background(), AnalyzeUseCaseConfig{
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      false,
		SkipCommunities: true,
		MinComplexity:   5,
	}, []string{sourcePath, "./hotspot.py"})
	if err != nil {
		t.Fatalf("execute analysis: %v", err)
	}

	if len(response.ModuleQuality) != 1 {
		t.Fatalf("expected 1 module-quality entry, got %d", len(response.ModuleQuality))
	}
	module := response.ModuleQuality[0]
	if module.FilePath != sourcePath {
		t.Errorf("expected module path %q, got %q", sourcePath, module.FilePath)
	}
	if module.ModuleName == "" || module.LinesOfCode == 0 || module.FunctionCount == 0 {
		t.Errorf("expected system metadata to join the relative module path, got %+v", module)
	}
	if module.AnalyzedFunctionCount <= len(response.Complexity.Functions) {
		t.Errorf("expected module rollup to retain filtered complexity records: got %d records and %d visible results",
			module.AnalyzedFunctionCount, len(response.Complexity.Functions))
	}

	systemOnly, err := useCase.Execute(context.Background(), AnalyzeUseCaseConfig{
		SkipComplexity:  true,
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipCommunities: true,
	}, []string{sourcePath})
	if err != nil {
		t.Fatalf("execute system-only analysis: %v", err)
	}
	if len(systemOnly.ModuleQuality) != 1 || systemOnly.ModuleQuality[0].FilePath != sourcePath {
		t.Fatalf("expected system-only module path %q, got %+v", sourcePath, systemOnly.ModuleQuality)
	}
}

func TestAnalyzeUseCase_Execute_PublishesDirectoryComplexity(t *testing.T) {
	projectRoot := t.TempDir()
	pkgDir := filepath.Join(projectRoot, "pkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("create package directory: %v", err)
	}
	sourcePath := filepath.Join(pkgDir, "hotspot.py")
	source := `def hotspot(value):
	if value > 10:
		return value
	return 0
`
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("write Python source: %v", err)
	}

	response, err := newModuleQualityAnalyzeUseCase(t).Execute(context.Background(), AnalyzeUseCaseConfig{
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      true,
		SkipCommunities: true,
		MinComplexity:   1,
	}, []string{projectRoot})
	if err != nil {
		t.Fatalf("execute analysis: %v", err)
	}
	if response.Complexity == nil {
		t.Fatal("expected complexity response")
	}
	if len(response.Complexity.ByDirectory) != 1 {
		t.Fatalf("expected one directory rollup, got %+v", response.Complexity.ByDirectory)
	}
	rollup := response.Complexity.ByDirectory[0]
	if rollup.DirectoryPath != "pkg" || rollup.FunctionCount != len(response.Complexity.Functions) {
		t.Fatalf("expected rollup to reconcile with reported functions, got %+v", rollup)
	}
}

func TestPrepareAnalysisPaths_PreservesFirstReportedPath(t *testing.T) {
	projectRoot := t.TempDir()
	t.Chdir(projectRoot)

	absPath := filepath.Join(projectRoot, "pkg", "hot.py")
	paths, pathIndex, err := prepareAnalysisPaths([]string{"pkg/hot.py", "pkg/../pkg/hot.py", absPath})
	if err != nil {
		t.Fatalf("prepare analysis paths: %v", err)
	}
	want := "pkg/hot.py"
	if len(paths) != 1 || paths[0] != want {
		t.Fatalf("expected first reported path %q, got %v", want, paths)
	}
	reported, err := pathIndex.reportedPath(absPath)
	if err != nil {
		t.Fatalf("resolve reported path: %v", err)
	}
	if reported != want {
		t.Fatalf("expected indexed path %q, got %q", want, reported)
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
	builder.WithComplexityUseCase(NewSnapshotComplexityUseCase(
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

func TestAnalyzeUseCase_Execute_DisablesAnalyzersFromConfig(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "sample.py")
	if err := os.WriteFile(sourcePath, []byte("def sample():\n    return 1\n"), 0644); err != nil {
		t.Fatalf("Failed to write source file: %v", err)
	}

	configPath := filepath.Join(tempDir, ".pyscn.toml")
	configContent := `[dead_code]
enabled = false

[system_analysis]
enabled = false

[dependencies]
enabled = false

[architecture]
enabled = false
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	useCase := &AnalyzeUseCase{
		deadCodeUseCase: &DeadCodeUseCase{},
		systemUseCase:   &SystemAnalysisUseCase{},
		fileReader:      service.NewFileReader(),
		configLoader:    service.NewAnalyzeConfigurationLoader(),
	}

	response, err := useCase.Execute(context.Background(), AnalyzeUseCaseConfig{
		ConfigFile:      configPath,
		MinSeverity:     domain.DeadCodeSeverityWarning,
		CloneSimilarity: 0.8,
	}, []string{tempDir})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if response.Summary.DeadCodeEnabled {
		t.Errorf("Expected dead code to be disabled, got %v", response.Summary.DeadCodeEnabled)
	}
	if response.Summary.DepsEnabled {
		t.Errorf("Expected system dependencies to be disabled, got %v", response.Summary.DepsEnabled)
	}
	if response.Summary.ArchEnabled {
		t.Errorf("Expected architecture to be disabled, got %v", response.Summary.ArchEnabled)
	}
	if response.DeadCode != nil {
		t.Errorf("Expected no dead code response, got %+v", response.DeadCode)
	}
	if response.System != nil {
		t.Errorf("Expected no system response, got %+v", response.System)
	}
}

func TestAnalyzeUseCase_LoadExecutionConfig(t *testing.T) {
	useCase := &AnalyzeUseCase{configLoader: service.NewAnalyzeConfigurationLoader()}

	t.Run("uses analyze defaults without config file", func(t *testing.T) {
		executionCfg, err := useCase.loadExecutionConfig("", []string{t.TempDir()})
		if err != nil {
			t.Fatalf("loadExecutionConfig returned error: %v", err)
		}

		if !executionCfg.ComplexityEnabled {
			t.Error("Expected complexity to be enabled by default")
		}
		if !executionCfg.ComplexityReportUnchanged {
			t.Error("Expected report_unchanged to be true by default")
		}
		if !executionCfg.DeadCodeEnabled {
			t.Error("Expected dead code to be enabled by default")
		}
		if !executionCfg.SystemEnabled {
			t.Error("Expected system analysis to be enabled by default")
		}
		if !executionCfg.SystemAnalyzeDependencies {
			t.Error("Expected dependency analysis to be enabled by default")
		}
		if !executionCfg.SystemAnalyzeArchitecture {
			t.Error("Expected architecture analysis to be enabled by default")
		}
		if executionCfg.ComplexityLowThreshold != domain.DefaultComplexityLowThreshold {
			t.Errorf("Expected low threshold %d, got %d", domain.DefaultComplexityLowThreshold, executionCfg.ComplexityLowThreshold)
		}
		if executionCfg.ComplexityMediumThreshold != domain.DefaultComplexityMediumThreshold {
			t.Errorf("Expected medium threshold %d, got %d", domain.DefaultComplexityMediumThreshold, executionCfg.ComplexityMediumThreshold)
		}
		if executionCfg.ComplexityMaxComplexity != domain.DefaultComplexityMaxLimit {
			t.Errorf("Expected max complexity %d, got %d", domain.DefaultComplexityMaxLimit, executionCfg.ComplexityMaxComplexity)
		}
		if executionCfg.ComplexityMinComplexity != domain.DefaultComplexityMinFilter {
			t.Errorf("Expected min complexity %d, got %d", domain.DefaultComplexityMinFilter, executionCfg.ComplexityMinComplexity)
		}
		if len(executionCfg.IncludePatterns) != 1 || executionCfg.IncludePatterns[0] != "**/*.py" {
			t.Errorf("Expected default include patterns to include runtime Python files, got %v", executionCfg.IncludePatterns)
		}
		defaultCloneReq := domain.DefaultCloneRequest()
		if executionCfg.CloneLSHEnabled != defaultCloneReq.LSHEnabled {
			t.Errorf("Expected default LSH enabled %q, got %q", defaultCloneReq.LSHEnabled, executionCfg.CloneLSHEnabled)
		}
		if executionCfg.CloneLSHAutoThreshold != defaultCloneReq.LSHAutoThreshold {
			t.Errorf("Expected default LSH threshold %d, got %d", defaultCloneReq.LSHAutoThreshold, executionCfg.CloneLSHAutoThreshold)
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

[dead_code]
enabled = false

[system_analysis]
enabled = true
enable_dependencies = false
enable_architecture = true

[dependencies]
enabled = true

[output]
min_complexity = 9

[clones]
lsh_enabled = "true"
lsh_auto_threshold = 123
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		executionCfg, err := useCase.loadExecutionConfig(configPath, []string{tempDir})
		if err != nil {
			t.Fatalf("loadExecutionConfig returned error: %v", err)
		}

		if executionCfg.ComplexityEnabled {
			t.Error("Expected complexity to be disabled")
		}
		if executionCfg.ComplexityReportUnchanged {
			t.Error("Expected report_unchanged to be false")
		}
		if executionCfg.ComplexityLowThreshold != 3 {
			t.Errorf("Expected low threshold 3, got %d", executionCfg.ComplexityLowThreshold)
		}
		if executionCfg.ComplexityMediumThreshold != 7 {
			t.Errorf("Expected medium threshold 7, got %d", executionCfg.ComplexityMediumThreshold)
		}
		if executionCfg.ComplexityMaxComplexity != 11 {
			t.Errorf("Expected max complexity 11, got %d", executionCfg.ComplexityMaxComplexity)
		}
		if executionCfg.ComplexityMinComplexity != 9 {
			t.Errorf("Expected min complexity 9, got %d", executionCfg.ComplexityMinComplexity)
		}
		if executionCfg.DeadCodeEnabled {
			t.Error("Expected dead code to be disabled")
		}
		if !executionCfg.SystemEnabled {
			t.Error("Expected system analysis to be enabled")
		}
		if !executionCfg.SystemAnalyzeDependencies {
			t.Error("Expected dependencies to be enabled through dependencies section")
		}
		if !executionCfg.SystemAnalyzeArchitecture {
			t.Error("Expected architecture to be enabled through system analysis section")
		}
		if executionCfg.Recursive {
			t.Error("Expected recursive to be false")
		}
		if len(executionCfg.IncludePatterns) != 1 || executionCfg.IncludePatterns[0] != "pkg/**/*.py" {
			t.Errorf("Expected custom include patterns, got %v", executionCfg.IncludePatterns)
		}
		if len(executionCfg.ExcludePatterns) != 1 || executionCfg.ExcludePatterns[0] != "tests/**/*.py" {
			t.Errorf("Expected custom exclude patterns, got %v", executionCfg.ExcludePatterns)
		}
		if executionCfg.CloneLSHEnabled != "true" {
			t.Errorf("Expected LSH enabled to be %q, got %q", "true", executionCfg.CloneLSHEnabled)
		}
		if executionCfg.CloneLSHAutoThreshold != 123 {
			t.Errorf("Expected LSH threshold 123, got %d", executionCfg.CloneLSHAutoThreshold)
		}
	})

	t.Run("keeps system defaults when config omits system sections", func(t *testing.T) {
		tempDir := t.TempDir()
		configPath := filepath.Join(tempDir, ".pyscn.toml")
		configContent := `[complexity]
enabled = false
`
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		executionCfg, err := useCase.loadExecutionConfig(configPath, []string{tempDir})
		if err != nil {
			t.Fatalf("loadExecutionConfig returned error: %v", err)
		}

		if !executionCfg.SystemEnabled {
			t.Error("Expected system analysis to remain enabled")
		}
		if !executionCfg.SystemAnalyzeDependencies {
			t.Error("Expected dependency analysis to remain enabled")
		}
		if !executionCfg.SystemAnalyzeArchitecture {
			t.Error("Expected architecture analysis to remain enabled")
		}
	})
}

func TestAnalyzeUseCase_buildCloneTaskRequest_PropagatesExecutionConfig(t *testing.T) {
	useCase := &AnalyzeUseCase{}
	config := AnalyzeUseCaseConfig{
		CloneSimilarity: 0.8,
		ConfigFile:      "/tmp/.pyscn.toml",
	}
	files := []string{"a.py", "b.py"}

	request := useCase.buildCloneTaskRequest(config, files, domain.AnalyzeExecutionConfig{Recursive: true})

	// Values already resolved by the unified analyze configuration are explicit;
	// unrelated fields stay sparse so MergeConfig can fill them as usual.
	if len(request.Paths) != len(files) || request.Paths[0] != files[0] || request.Paths[1] != files[1] {
		t.Fatalf("expected clone task paths %v, got %v", files, request.Paths)
	}
	if request.OutputFormat != domain.OutputFormatJSON {
		t.Fatalf("expected JSON output format, got %q", request.OutputFormat)
	}
	if request.OutputWriter == nil {
		t.Fatal("expected clone task output writer to be set")
	}
	if request.SimilarityThreshold != config.CloneSimilarity {
		t.Fatalf("expected similarity threshold %.2f, got %.2f", config.CloneSimilarity, request.SimilarityThreshold)
	}
	if request.ConfigPath != config.ConfigFile {
		t.Fatalf("expected config path %q, got %q", config.ConfigFile, request.ConfigPath)
	}
	if !domain.BoolValue(request.Recursive, false) {
		t.Fatal("expected effective recursive setting to be propagated")
	}
	if request.MaxSimilarity != 0 {
		t.Fatalf("expected max similarity to be unset (0), got %.2f", request.MaxSimilarity)
	}
	if request.GroupMode != "" {
		t.Fatalf("expected group mode to be unset, got %q", request.GroupMode)
	}
	if request.GroupThreshold != 0 {
		t.Fatalf("expected group threshold to be unset (0), got %.2f", request.GroupThreshold)
	}
	if request.KCoreK != 0 {
		t.Fatalf("expected k-core to be unset (0), got %d", request.KCoreK)
	}
}

func TestAnalyzeUseCase_Execute_PreservesCloneConfigDefaults(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".pyscn.toml")
	configContent := `[clones]
show_content = true
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	builder := NewAnalyzeUseCaseBuilder()
	builder.WithFileReader(service.NewFileReader())
	builder.WithFormatter(service.NewAnalyzeFormatter())
	builder.WithProgressManager(service.NewProgressManager())
	builder.WithParallelExecutor(service.NewParallelExecutor())
	builder.WithErrorCategorizer(service.NewErrorCategorizer())

	cloneUseCase, err := NewCloneUseCaseBuilder().
		WithSnapshotService(service.NewCloneService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewCloneOutputFormatter()).
		WithConfigLoader(service.NewCloneConfigurationLoader()).
		Build()
	if err != nil {
		t.Fatalf("failed to build clone use case: %v", err)
	}
	builder.WithCloneUseCase(cloneUseCase)

	useCase, err := builder.Build()
	if err != nil {
		t.Fatalf("failed to build analyze use case: %v", err)
	}

	response, err := useCase.Execute(context.Background(), AnalyzeUseCaseConfig{
		SkipComplexity:  true,
		SkipDeadCode:    true,
		SkipClones:      false,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      true,
		CloneSimilarity: domain.DefaultCloneSimilarityThreshold,
		ConfigFile:      configPath,
	}, []string{"../testdata/python/frameworks/actual_clones.py"})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if response.Clone == nil {
		t.Fatal("expected clone response")
	}
	if response.Clone.Request == nil {
		t.Fatal("expected clone request in response")
	}
	if response.Clone.Statistics == nil {
		t.Fatal("expected clone statistics")
	}
	if response.Clone.Statistics.TotalClonePairs == 0 {
		t.Fatalf("expected at least one clone pair, got %+v", response.Clone.Statistics)
	}

	defaultReq := domain.DefaultCloneRequest()
	if response.Clone.Request.MaxSimilarity != defaultReq.MaxSimilarity {
		t.Fatalf("expected max similarity %.2f, got %.2f", defaultReq.MaxSimilarity, response.Clone.Request.MaxSimilarity)
	}
	if response.Clone.Request.GroupMode != defaultReq.GroupMode {
		t.Fatalf("expected group mode %q, got %q", defaultReq.GroupMode, response.Clone.Request.GroupMode)
	}
	if response.Clone.Request.GroupThreshold != defaultReq.GroupThreshold {
		t.Fatalf("expected group threshold %.2f, got %.2f", defaultReq.GroupThreshold, response.Clone.Request.GroupThreshold)
	}
	if response.Clone.Request.KCoreK != defaultReq.KCoreK {
		t.Fatalf("expected k-core %d, got %d", defaultReq.KCoreK, response.Clone.Request.KCoreK)
	}
	if !domain.BoolValue(response.Clone.Request.ShowContent, false) {
		t.Fatal("expected show_content from config to be preserved")
	}
}

// TestBuildComplexityTaskRequest_ThresholdOverrides verifies that CLI flag
// values (> 0) in AnalyzeUseCaseConfig take precedence over execution config
// values, and that zero (unset) falls back to execution config. This is the
// runtime counterpart to the MergeConfig fix for issue #553.
func TestBuildComplexityTaskRequest_ThresholdOverrides(t *testing.T) {
	uc := &AnalyzeUseCase{}

	executionCfg := domain.AnalyzeExecutionConfig{
		ComplexityLowThreshold:       10,
		ComplexityMediumThreshold:    20,
		CognitiveComplexityThreshold: 30,
		NestingDepthThreshold:        11,
	}

	t.Run("CLI flags override execution config", func(t *testing.T) {
		config := AnalyzeUseCaseConfig{
			LowThreshold:                 9,
			MediumThreshold:              19,
			CognitiveComplexityThreshold: 25,
			NestingDepthThreshold:        7,
		}
		req := uc.buildComplexityTaskRequest(config, []string{"test.py"}, executionCfg)

		if req.LowThreshold != 9 {
			t.Errorf("LowThreshold: expected 9 (CLI), got %d", req.LowThreshold)
		}
		if req.MediumThreshold != 19 {
			t.Errorf("MediumThreshold: expected 19 (CLI), got %d", req.MediumThreshold)
		}
		if req.CognitiveComplexityThreshold != 25 {
			t.Errorf("CognitiveComplexityThreshold: expected 25 (CLI), got %d", req.CognitiveComplexityThreshold)
		}
		if req.NestingDepthThreshold != 7 {
			t.Errorf("NestingDepthThreshold: expected 7 (CLI), got %d", req.NestingDepthThreshold)
		}
	})

	t.Run("zero flags fall back to execution config", func(t *testing.T) {
		config := AnalyzeUseCaseConfig{}
		req := uc.buildComplexityTaskRequest(config, []string{"test.py"}, executionCfg)

		if req.LowThreshold != 10 {
			t.Errorf("LowThreshold: expected 10 (exec), got %d", req.LowThreshold)
		}
		if req.MediumThreshold != 20 {
			t.Errorf("MediumThreshold: expected 20 (exec), got %d", req.MediumThreshold)
		}
		if req.CognitiveComplexityThreshold != 30 {
			t.Errorf("CognitiveComplexityThreshold: expected 30 (exec), got %d", req.CognitiveComplexityThreshold)
		}
		if req.NestingDepthThreshold != 11 {
			t.Errorf("NestingDepthThreshold: expected 11 (exec), got %d", req.NestingDepthThreshold)
		}
	})

	t.Run("partial override only affects set flags", func(t *testing.T) {
		config := AnalyzeUseCaseConfig{
			CognitiveComplexityThreshold: 25,
		}
		req := uc.buildComplexityTaskRequest(config, []string{"test.py"}, executionCfg)

		if req.LowThreshold != 10 {
			t.Errorf("LowThreshold: expected 10 (exec), got %d", req.LowThreshold)
		}
		if req.CognitiveComplexityThreshold != 25 {
			t.Errorf("CognitiveComplexityThreshold: expected 25 (CLI), got %d", req.CognitiveComplexityThreshold)
		}
		if req.NestingDepthThreshold != 11 {
			t.Errorf("NestingDepthThreshold: expected 11 (exec), got %d", req.NestingDepthThreshold)
		}
	})
}

func TestBuildComplexityTaskRequest_UsesExecutionConfigBooleans(t *testing.T) {
	uc := &AnalyzeUseCase{}
	executionCfg := domain.AnalyzeExecutionConfig{
		Recursive:   false,
		ShowDetails: true,
	}

	req := uc.buildComplexityTaskRequest(AnalyzeUseCaseConfig{}, []string{"test.py"}, executionCfg)

	if req.Recursive == nil || *req.Recursive {
		t.Error("Recursive: expected explicit false from execution config")
	}
	if req.ShowDetails == nil || !*req.ShowDetails {
		t.Error("ShowDetails: expected explicit true from execution config")
	}
}
