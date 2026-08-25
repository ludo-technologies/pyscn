package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"sync"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
)

// AnalyzeUseCaseConfig holds configuration for the analyze use case
type AnalyzeUseCaseConfig struct {
	SkipComplexity  bool
	SkipDeadCode    bool
	SkipClones      bool
	SkipCBO         bool
	SkipLCOM        bool
	SkipSystem      bool
	SkipCommunities bool

	// SelectAnalysesUsed is true when --select was provided on the CLI.
	SelectAnalysesUsed bool
	// SkipCommunitiesExplicit is true when --skip-communities was provided.
	SkipCommunitiesExplicit bool

	MinComplexity   int
	MinSeverity     domain.DeadCodeSeverity
	CloneSimilarity float64
	MinCBO          int

	// Complexity thresholds (0 = unset, use config file or default)
	LowThreshold                 int
	MediumThreshold              int
	CognitiveComplexityThreshold int
	NestingDepthThreshold        int

	FunctionSLOCWarnThreshold     int
	FunctionSLOCCriticalThreshold int

	// Clone detection options
	EnableDFA bool // Enable Data Flow Analysis for enhanced Type-4 detection

	ConfigFile string
	Verbose    bool
}

// AnalyzeRequestOverrides contains request-scoped values that take precedence
// over the resolved project configuration.
type AnalyzeRequestOverrides struct {
	Recursive                 *bool
	ComplexityEnabled         *bool
	DeadCodeEnabled           *bool
	SystemEnabled             *bool
	SystemAnalyzeDependencies *bool
	SystemAnalyzeArchitecture *bool
	ModuleGraph               *domain.ModuleGraphOptions
}

// AnalyzeUseCase orchestrates comprehensive analysis
type AnalyzeUseCase struct {
	complexityUseCase *ComplexityUseCase
	deadCodeUseCase   *DeadCodeUseCase
	cloneUseCase      *CloneUseCase
	cboUseCase        *CBOUseCase
	lcomUseCase       *LCOMUseCase
	systemUseCase     *SystemAnalysisUseCase
	communityUseCase  *CommunityUseCase

	fileReader       domain.FileReader
	configLoader     domain.AnalyzeConfigurationLoader
	formatter        domain.AnalyzeOutputFormatter
	progressManager  domain.ProgressManager
	parallelExecutor domain.ParallelExecutor
	errorCategorizer domain.ErrorCategorizer
}

// AnalyzeUseCaseBuilder builds an AnalyzeUseCase
type AnalyzeUseCaseBuilder struct {
	complexityUseCase *ComplexityUseCase
	deadCodeUseCase   *DeadCodeUseCase
	cloneUseCase      *CloneUseCase
	cboUseCase        *CBOUseCase
	lcomUseCase       *LCOMUseCase
	systemUseCase     *SystemAnalysisUseCase
	communityUseCase  *CommunityUseCase

	fileReader       domain.FileReader
	configLoader     domain.AnalyzeConfigurationLoader
	formatter        domain.AnalyzeOutputFormatter
	progressManager  domain.ProgressManager
	parallelExecutor domain.ParallelExecutor
	errorCategorizer domain.ErrorCategorizer
}

// NewAnalyzeUseCaseBuilder creates a new builder
func NewAnalyzeUseCaseBuilder() *AnalyzeUseCaseBuilder {
	return &AnalyzeUseCaseBuilder{}
}

// WithComplexityUseCase sets the complexity use case
func (b *AnalyzeUseCaseBuilder) WithComplexityUseCase(uc *ComplexityUseCase) *AnalyzeUseCaseBuilder {
	b.complexityUseCase = uc
	return b
}

// WithDeadCodeUseCase sets the dead code use case
func (b *AnalyzeUseCaseBuilder) WithDeadCodeUseCase(uc *DeadCodeUseCase) *AnalyzeUseCaseBuilder {
	b.deadCodeUseCase = uc
	return b
}

// WithCloneUseCase sets the clone use case
func (b *AnalyzeUseCaseBuilder) WithCloneUseCase(uc *CloneUseCase) *AnalyzeUseCaseBuilder {
	b.cloneUseCase = uc
	return b
}

// WithCBOUseCase sets the CBO use case
func (b *AnalyzeUseCaseBuilder) WithCBOUseCase(uc *CBOUseCase) *AnalyzeUseCaseBuilder {
	b.cboUseCase = uc
	return b
}

// WithLCOMUseCase sets the LCOM use case
func (b *AnalyzeUseCaseBuilder) WithLCOMUseCase(uc *LCOMUseCase) *AnalyzeUseCaseBuilder {
	b.lcomUseCase = uc
	return b
}

// WithSystemUseCase sets the system analysis use case
func (b *AnalyzeUseCaseBuilder) WithSystemUseCase(uc *SystemAnalysisUseCase) *AnalyzeUseCaseBuilder {
	b.systemUseCase = uc
	return b
}

// WithCommunityUseCase sets the community analysis use case
func (b *AnalyzeUseCaseBuilder) WithCommunityUseCase(uc *CommunityUseCase) *AnalyzeUseCaseBuilder {
	b.communityUseCase = uc
	return b
}

// WithFileReader sets the file reader
func (b *AnalyzeUseCaseBuilder) WithFileReader(fr domain.FileReader) *AnalyzeUseCaseBuilder {
	b.fileReader = fr
	return b
}

// WithConfigLoader sets the analyze configuration loader.
func (b *AnalyzeUseCaseBuilder) WithConfigLoader(cl domain.AnalyzeConfigurationLoader) *AnalyzeUseCaseBuilder {
	b.configLoader = cl
	return b
}

// WithFormatter sets the formatter
func (b *AnalyzeUseCaseBuilder) WithFormatter(f domain.AnalyzeOutputFormatter) *AnalyzeUseCaseBuilder {
	b.formatter = f
	return b
}

// WithProgressManager sets the progress manager
func (b *AnalyzeUseCaseBuilder) WithProgressManager(pm domain.ProgressManager) *AnalyzeUseCaseBuilder {
	b.progressManager = pm
	return b
}

// WithParallelExecutor sets the parallel executor
func (b *AnalyzeUseCaseBuilder) WithParallelExecutor(pe domain.ParallelExecutor) *AnalyzeUseCaseBuilder {
	b.parallelExecutor = pe
	return b
}

// WithErrorCategorizer sets the error categorizer
func (b *AnalyzeUseCaseBuilder) WithErrorCategorizer(ec domain.ErrorCategorizer) *AnalyzeUseCaseBuilder {
	b.errorCategorizer = ec
	return b
}

// Build creates the AnalyzeUseCase
func (b *AnalyzeUseCaseBuilder) Build() (*AnalyzeUseCase, error) {
	if b.fileReader == nil {
		return nil, fmt.Errorf("file reader is required")
	}
	if err := b.validateAggregateCollaborators(); err != nil {
		return nil, fmt.Errorf("validate aggregate collaborators: %w", err)
	}
	if b.configLoader == nil {
		b.configLoader = service.NewAnalyzeConfigurationLoader()
	}
	if b.formatter == nil {
		b.formatter = service.NewAnalyzeFormatter()
	}
	if b.progressManager == nil {
		b.progressManager = service.NewProgressManager()
	}
	if b.parallelExecutor == nil {
		b.parallelExecutor = service.NewParallelExecutor()
	}
	if b.errorCategorizer == nil {
		b.errorCategorizer = service.NewErrorCategorizer()
	}

	return &AnalyzeUseCase{
		complexityUseCase: b.complexityUseCase,
		deadCodeUseCase:   b.deadCodeUseCase,
		cloneUseCase:      b.cloneUseCase,
		cboUseCase:        b.cboUseCase,
		lcomUseCase:       b.lcomUseCase,
		systemUseCase:     b.systemUseCase,
		communityUseCase:  b.communityUseCase,
		fileReader:        b.fileReader,
		configLoader:      b.configLoader,
		formatter:         b.formatter,
		progressManager:   b.progressManager,
		parallelExecutor:  b.parallelExecutor,
		errorCategorizer:  b.errorCategorizer,
	}, nil
}

func (b *AnalyzeUseCaseBuilder) validateAggregateCollaborators() error {
	if b.complexityUseCase != nil && b.complexityUseCase.snapshot == nil {
		return fmt.Errorf("complexity use case requires a snapshot collaborator")
	}
	if b.deadCodeUseCase != nil && b.deadCodeUseCase.snapshot == nil {
		return fmt.Errorf("dead-code use case requires a snapshot collaborator")
	}
	if b.cloneUseCase != nil && b.cloneUseCase.snapshot == nil {
		return fmt.Errorf("clone use case requires a snapshot collaborator")
	}
	if b.cboUseCase != nil && b.cboUseCase.snapshot == nil {
		return fmt.Errorf("cbo use case requires a snapshot collaborator")
	}
	if b.lcomUseCase != nil && b.lcomUseCase.snapshot == nil {
		return fmt.Errorf("lcom use case requires a snapshot collaborator")
	}
	if b.systemUseCase != nil && b.systemUseCase.graphService == nil {
		return fmt.Errorf("system use case requires a graph collaborator")
	}
	if b.communityUseCase != nil && b.communityUseCase.graphService == nil {
		return fmt.Errorf("community use case requires a graph collaborator")
	}
	return nil
}

// Task names used both for display and as keys for progress estimation
const (
	taskNameComplexity  = "Complexity Analysis"
	taskNameDeadCode    = "Dead Code Detection"
	taskNameClones      = "Clone Detection"
	taskNameCBO         = "Class Coupling (CBO)"
	taskNameLCOM        = "Class Cohesion (LCOM)"
	taskNameSystem      = "System Analysis"
	taskNameCommunities = "Community Detection"
)

// analysisTask represents a single aggregate-owned analysis task.
type analysisTask struct {
	Name    string
	Kind    domain.AnalysisKind
	Enabled bool
	Execute func(context.Context) (analysisTaskResult, error)
	Result  analysisTaskResult
	Error   error
}

type analysisTaskResult interface {
	analysisKind() domain.AnalysisKind
	analysisFailures() []domain.AnalysisFailure
	applyTo(*domain.AnalyzeResponse)
}

type complexityTaskResult struct{ response *domain.ComplexityResponse }

func (complexityTaskResult) analysisKind() domain.AnalysisKind { return domain.AnalysisKindComplexity }
func (r complexityTaskResult) analysisFailures() []domain.AnalysisFailure {
	return r.response.AnalysisFailures()
}
func (r complexityTaskResult) applyTo(response *domain.AnalyzeResponse) {
	response.Complexity = r.response
}

type deadCodeTaskResult struct{ response *domain.DeadCodeResponse }

func (deadCodeTaskResult) analysisKind() domain.AnalysisKind { return domain.AnalysisKindDeadCode }
func (r deadCodeTaskResult) analysisFailures() []domain.AnalysisFailure {
	return r.response.AnalysisFailures()
}
func (r deadCodeTaskResult) applyTo(response *domain.AnalyzeResponse) { response.DeadCode = r.response }

type cloneTaskResult struct{ response *domain.CloneResponse }

func (cloneTaskResult) analysisKind() domain.AnalysisKind { return domain.AnalysisKindClones }
func (r cloneTaskResult) analysisFailures() []domain.AnalysisFailure {
	return r.response.AnalysisFailures()
}
func (r cloneTaskResult) applyTo(response *domain.AnalyzeResponse) { response.Clone = r.response }

type cboTaskResult struct{ response *domain.CBOResponse }

func (cboTaskResult) analysisKind() domain.AnalysisKind { return domain.AnalysisKindCBO }
func (r cboTaskResult) analysisFailures() []domain.AnalysisFailure {
	return r.response.AnalysisFailures()
}
func (r cboTaskResult) applyTo(response *domain.AnalyzeResponse) { response.CBO = r.response }

type lcomTaskResult struct{ response *domain.LCOMResponse }

func (lcomTaskResult) analysisKind() domain.AnalysisKind { return domain.AnalysisKindLCOM }
func (r lcomTaskResult) analysisFailures() []domain.AnalysisFailure {
	return r.response.AnalysisFailures()
}
func (r lcomTaskResult) applyTo(response *domain.AnalyzeResponse) { response.LCOM = r.response }

type systemTaskResult struct {
	response *domain.SystemAnalysisResponse
}

func (systemTaskResult) analysisKind() domain.AnalysisKind { return domain.AnalysisKindSystem }
func (r systemTaskResult) analysisFailures() []domain.AnalysisFailure {
	return r.response.AnalysisFailures()
}
func (r systemTaskResult) applyTo(response *domain.AnalyzeResponse) {
	response.System = r.response
	if r.response != nil && r.response.ArchitectureAnalysis != nil {
		response.Summary.ArchEnabled = true
	}
}

type communityTaskResult struct {
	response *domain.CommunityAnalysisResult
}

func (communityTaskResult) analysisKind() domain.AnalysisKind {
	return domain.AnalysisKindCommunities
}
func (r communityTaskResult) analysisFailures() []domain.AnalysisFailure {
	return r.response.AnalysisFailures()
}
func (r communityTaskResult) applyTo(response *domain.AnalyzeResponse) {
	response.Communities = r.response
}

type analysisRunError struct {
	failures []domain.AnalysisFailure
	causes   []error
}

func (e *analysisRunError) Error() string {
	return fmt.Sprintf("analysis completed with %d failure(s): %s", len(e.failures), e.failures[0].Message)
}

func (e *analysisRunError) Unwrap() []error {
	return e.causes
}

func (e *analysisRunError) AnalysisFailures() []domain.AnalysisFailure {
	return append([]domain.AnalysisFailure(nil), e.failures...)
}

// ProjectAnalysisResult owns the canonical snapshot and the analyses derived
// from it for callers that need to run additional snapshot-aware policies.
type ProjectAnalysisResult struct {
	Response *domain.AnalyzeResponse
	Snapshot *service.ProjectSnapshot
}

// Execute performs comprehensive analysis
func (uc *AnalyzeUseCase) Execute(ctx context.Context, useCaseCfg AnalyzeUseCaseConfig, paths []string) (*domain.AnalyzeResponse, error) {
	result, err := uc.executeProject(ctx, useCaseCfg, paths, AnalyzeRequestOverrides{})
	if result == nil {
		if err != nil {
			return nil, fmt.Errorf("execute project analysis: %w", err)
		}
		return nil, nil
	}
	if err != nil {
		return result.Response, fmt.Errorf("execute project analysis: %w", err)
	}
	return result.Response, nil
}

// ExecuteWithOverrides performs comprehensive analysis with request-scoped
// overrides applied after project configuration is resolved.
func (uc *AnalyzeUseCase) ExecuteWithOverrides(ctx context.Context, useCaseCfg AnalyzeUseCaseConfig, paths []string, overrides AnalyzeRequestOverrides) (*domain.AnalyzeResponse, error) {
	result, err := uc.executeProject(ctx, useCaseCfg, paths, overrides)
	if result == nil {
		if err != nil {
			return nil, fmt.Errorf("execute project analysis with overrides: %w", err)
		}
		return nil, nil
	}
	if err != nil {
		return result.Response, fmt.Errorf("execute project analysis with overrides: %w", err)
	}
	return result.Response, nil
}

// ExecuteProjectWithOverrides returns the response and the sealed snapshot
// that produced it.
func (uc *AnalyzeUseCase) ExecuteProjectWithOverrides(ctx context.Context, useCaseCfg AnalyzeUseCaseConfig, paths []string, overrides AnalyzeRequestOverrides) (*ProjectAnalysisResult, error) {
	return uc.executeProject(ctx, useCaseCfg, paths, overrides)
}

func (uc *AnalyzeUseCase) executeProject(ctx context.Context, useCaseCfg AnalyzeUseCaseConfig, paths []string, overrides AnalyzeRequestOverrides) (*ProjectAnalysisResult, error) {
	startTime := time.Now()

	executionCfg, err := uc.loadExecutionConfig(useCaseCfg.ConfigFile, paths)
	if err != nil {
		return nil, err
	}
	if overrides.Recursive != nil {
		executionCfg.Recursive = *overrides.Recursive
	}
	if overrides.ComplexityEnabled != nil {
		executionCfg.ComplexityEnabled = *overrides.ComplexityEnabled
	}
	if overrides.DeadCodeEnabled != nil {
		executionCfg.DeadCodeEnabled = *overrides.DeadCodeEnabled
	}
	if overrides.SystemEnabled != nil {
		executionCfg.SystemEnabled = *overrides.SystemEnabled
	}
	if overrides.SystemAnalyzeDependencies != nil {
		executionCfg.SystemAnalyzeDependencies = *overrides.SystemAnalyzeDependencies
	}
	if overrides.SystemAnalyzeArchitecture != nil {
		executionCfg.SystemAnalyzeArchitecture = *overrides.SystemAnalyzeArchitecture
	}
	if overrides.ModuleGraph != nil {
		executionCfg.ModuleGraph = *overrides.ModuleGraph
	}
	useCaseCfg.ConfigFile = executionCfg.ConfigPath

	if !executionCfg.ComplexityEnabled {
		useCaseCfg.SkipComplexity = true
	}
	if !executionCfg.DeadCodeEnabled {
		useCaseCfg.SkipDeadCode = true
	}
	if !executionCfg.SystemEnabled {
		useCaseCfg.SkipSystem = true
	}

	if !useCaseCfg.SelectAnalysesUsed && executionCfg.CommunitiesEnabledExplicit {
		useCaseCfg.SkipCommunities = !executionCfg.CommunitiesEnabled
	}
	if useCaseCfg.SkipCommunitiesExplicit {
		useCaseCfg.SkipCommunities = true
	}

	// Validate and collect files using configured patterns
	analysisFiles, err := uc.fileReader.CollectPythonFiles(
		paths,
		executionCfg.Recursive,
		executionCfg.IncludePatterns,
		executionCfg.ExcludePatterns,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to collect Python files: %w", err)
	}

	moduleFiles := analysisFiles
	if uc.needsModuleGraph(useCaseCfg) {
		moduleSelection := executionCfg.PythonFileSelection().ForModules()
		moduleFiles, err = uc.fileReader.CollectPythonFiles(
			paths,
			executionCfg.Recursive,
			moduleSelection.IncludePatterns,
			moduleSelection.ExcludePatterns,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to collect Python modules: %w", err)
		}
	}

	if len(analysisFiles) == 0 && len(moduleFiles) == 0 {
		return nil, fmt.Errorf("no Python files found in the specified paths")
	}
	analysisFiles, _, err = prepareAnalysisPaths(analysisFiles)
	if err != nil {
		return nil, fmt.Errorf("prepare analysis paths: %w", err)
	}
	moduleFiles, _, err = prepareAnalysisPaths(moduleFiles)
	if err != nil {
		return nil, fmt.Errorf("prepare module paths: %w", err)
	}
	allFiles, pathIndex, err := prepareAnalysisPaths(append(append([]string(nil), analysisFiles...), moduleFiles...))
	if err != nil {
		return nil, fmt.Errorf("prepare analysis paths: %w", err)
	}
	// Estimate per-task durations from file count, then calibrate with actual
	// timings recorded by previous runs on this project (if any)
	estimatedSeconds := uc.estimateTaskSeconds(len(allFiles), useCaseCfg, executionCfg)

	snapshot := service.BuildAnalysisProjectSnapshot(ctx, analysisFiles, moduleFiles, service.ProjectSnapshotOptions{
		IncludeRawMetrics: uc.complexityUseCase != nil && !useCaseCfg.SkipComplexity,
		ProjectRoot:       service.FindProjectRoot(paths),
	})

	var moduleGraph *service.ProjectModuleGraph
	var moduleGraphErr error
	if uc.needsModuleGraph(useCaseCfg) {
		moduleGraph, moduleGraphErr = snapshot.BuildDependencyGraph(ctx, &executionCfg.ModuleGraph)
	}

	// Start unified progress tracking; task completions feed back into the
	// estimate so the bar recalibrates to actual machine/codebase speed
	var tracker *analysisProgressTracker
	var progressDone chan struct{}
	if uc.progressManager != nil {
		tracker = newAnalysisProgressTracker(applyTimingFactors(estimatedSeconds, service.LoadAnalysisTimingFactors()))
		uc.progressManager.Initialize(100) // 100% based progress
		progressDone = uc.startProgressUpdater(tracker)
	}

	// Create analysis tasks
	tasks := uc.createAnalysisTasks(useCaseCfg, paths, analysisFiles, snapshot, moduleGraph, moduleGraphErr, executionCfg)

	// Execute tasks in parallel
	var wg sync.WaitGroup
	for _, task := range tasks {
		if !task.Enabled {
			continue
		}

		wg.Add(1)
		go func(t *analysisTask) {
			defer wg.Done()
			result, err := t.Execute(ctx)
			t.Result = result
			t.Error = err
			if tracker != nil {
				tracker.TaskCompleted(t.Kind)
			}
		}(task)
	}

	// Wait for all tasks to complete
	wg.Wait()

	// Stop progress updater and ensure progress bar reaches 100%
	if progressDone != nil {
		close(progressDone)
		uc.progressManager.Update(100, 100)
		uc.progressManager.Complete(true)
	}

	// Persist observed timings to improve the initial estimate of future runs
	if tracker != nil {
		service.UpdateAnalysisTimingFactors(estimatedSeconds, tracker.CompletedDurations())
	}

	// Build response
	coverage := snapshot.Coverage()
	if uc.needsModuleGraph(useCaseCfg) {
		coverage = snapshot.ModuleCoverage()
	}
	response, err := uc.buildResponse(tasks, startTime, pathIndex, coverage)
	result := &ProjectAnalysisResult{Response: response, Snapshot: snapshot}
	if err != nil {
		return result, fmt.Errorf("build analysis response: %w", err)
	}

	if len(response.Failures) > 0 {
		return result, newAnalysisRunError(tasks, response.Failures)
	}

	return result, nil
}

func newAnalysisRunError(tasks []*analysisTask, failures []domain.AnalysisFailure) error {
	causes := make([]error, 0, len(tasks))
	for _, task := range tasks {
		if task.Enabled && task.Error != nil {
			causes = append(causes, fmt.Errorf("%s: %w", task.Name, task.Error))
		}
	}
	for _, failure := range failures {
		cause := errors.Unwrap(failure)
		if cause == nil || errorListContains(causes, cause) {
			continue
		}
		causes = append(causes, failure)
	}
	return &analysisRunError{
		failures: append([]domain.AnalysisFailure(nil), failures...),
		causes:   causes,
	}
}

func errorListContains(errorsToCheck []error, target error) bool {
	for _, candidate := range errorsToCheck {
		if errors.Is(candidate, target) {
			return true
		}
	}
	return false
}

// createAnalysisTasks creates the analysis tasks based on configuration
func (uc *AnalyzeUseCase) needsModuleGraph(config AnalyzeUseCaseConfig) bool {
	return (uc.systemUseCase != nil && !config.SkipSystem) ||
		(uc.communityUseCase != nil && !config.SkipCommunities)
}

func cloneModuleGraph(graph *service.ProjectModuleGraph, buildErr error) (*service.ProjectModuleGraph, error) {
	if buildErr != nil {
		return nil, fmt.Errorf("build module graph: %w", buildErr)
	}
	if graph == nil {
		return nil, fmt.Errorf("module graph is required")
	}
	return graph.Clone(), nil
}

func (uc *AnalyzeUseCase) createAnalysisTasks(config AnalyzeUseCaseConfig, sourcePaths []string, files []string, snapshot *service.ProjectSnapshot, moduleGraph *service.ProjectModuleGraph, moduleGraphErr error, executionCfg domain.AnalyzeExecutionConfig) []*analysisTask {
	tasks := []*analysisTask{}

	// Complexity analysis task
	if uc.complexityUseCase != nil {
		tasks = append(tasks, &analysisTask{
			Name:    taskNameComplexity,
			Kind:    domain.AnalysisKindComplexity,
			Enabled: !config.SkipComplexity,
			Execute: func(ctx context.Context) (analysisTaskResult, error) {
				request := uc.buildComplexityTaskRequest(config, files, executionCfg)
				projectRoot, err := complexityDirectoryRoot(sourcePaths, files)
				if err != nil {
					return nil, domain.NewInvalidInputError("invalid complexity analysis scope", err)
				}
				response, err := uc.complexityUseCase.analyzeSnapshotRequest(ctx, snapshot, request, projectRoot)
				return complexityTaskResult{response: response}, err
			},
		})
	}

	// Dead code analysis task
	if uc.deadCodeUseCase != nil {
		tasks = append(tasks, &analysisTask{
			Name:    taskNameDeadCode,
			Kind:    domain.AnalysisKindDeadCode,
			Enabled: !config.SkipDeadCode,
			Execute: func(ctx context.Context) (analysisTaskResult, error) {
				request := domain.DeadCodeRequest{
					Paths:           files,
					Recursive:       domain.BoolPtr(executionCfg.Recursive),
					IncludePatterns: []string{},
					ExcludePatterns: []string{},
					OutputFormat:    domain.OutputFormatJSON,
					OutputWriter:    io.Discard,
					MinSeverity:     config.MinSeverity,
					SortBy:          "", // Zero: let config file values take precedence via merge
					ConfigPath:      config.ConfigFile,
					// Detection options left as nil to allow config file values to take precedence
					// If not set in config, defaults from DefaultDeadCodeRequest() will be used
					ShowContext:               nil,
					ContextLines:              0, // 0 = use config file or default value
					DetectAfterReturn:         nil,
					DetectAfterBreak:          nil,
					DetectAfterContinue:       nil,
					DetectAfterRaise:          nil,
					DetectUnreachableBranches: nil,
				}
				response, err := uc.deadCodeUseCase.analyzeSnapshotRequest(ctx, snapshot, request)
				return deadCodeTaskResult{response: response}, err
			},
		})
	}

	// Clone detection task
	if uc.cloneUseCase != nil {
		tasks = append(tasks, &analysisTask{
			Name:    taskNameClones,
			Kind:    domain.AnalysisKindClones,
			Enabled: !config.SkipClones,
			Execute: func(ctx context.Context) (analysisTaskResult, error) {
				request := uc.buildCloneTaskRequest(config, files, executionCfg)
				response, err := uc.cloneUseCase.analyzeSnapshotRequest(ctx, snapshot, request)
				return cloneTaskResult{response: response}, err
			},
		})
	}

	// CBO analysis task
	if uc.cboUseCase != nil {
		tasks = append(tasks, &analysisTask{
			Name:    taskNameCBO,
			Kind:    domain.AnalysisKindCBO,
			Enabled: !config.SkipCBO,
			Execute: func(ctx context.Context) (analysisTaskResult, error) {
				request := domain.CBORequest{
					Paths:           files,
					Recursive:       domain.BoolPtr(executionCfg.Recursive),
					IncludePatterns: []string{},
					ExcludePatterns: []string{},
					OutputFormat:    domain.OutputFormatJSON,
					OutputWriter:    io.Discard,
					MinCBO:          config.MinCBO,
					LowThreshold:    0, // Zero: let config file values take precedence via merge
					MediumThreshold: 0, // Zero: let config file values take precedence via merge
					SortBy:          domain.SortByCoupling,
					ConfigPath:      config.ConfigFile,
					// Boolean options left as nil to allow config file values to take precedence
					ShowZeros:             nil,
					IncludeBuiltins:       nil,
					IncludeImports:        nil,
					GroupNamespaceImports: nil,
				}
				response, err := uc.cboUseCase.analyzeSnapshotRequest(ctx, snapshot, request)
				return cboTaskResult{response: response}, err
			},
		})
	}

	// LCOM analysis task
	if uc.lcomUseCase != nil {
		tasks = append(tasks, &analysisTask{
			Name:    taskNameLCOM,
			Kind:    domain.AnalysisKindLCOM,
			Enabled: !config.SkipLCOM,
			Execute: func(ctx context.Context) (analysisTaskResult, error) {
				request := domain.LCOMRequest{
					Paths:           files,
					Recursive:       domain.BoolPtr(executionCfg.Recursive),
					IncludePatterns: []string{},
					ExcludePatterns: []string{},
					OutputFormat:    domain.OutputFormatJSON,
					OutputWriter:    io.Discard,
					LowThreshold:    0, // Zero: let config file values take precedence via merge
					MediumThreshold: 0, // Zero: let config file values take precedence via merge
					SortBy:          domain.SortByCohesion,
					ConfigPath:      config.ConfigFile,
				}
				response, err := uc.lcomUseCase.analyzeSnapshotRequest(ctx, snapshot, request)
				return lcomTaskResult{response: response}, err
			},
		})
	}

	// System analysis task
	if uc.systemUseCase != nil {
		tasks = append(tasks, &analysisTask{
			Name:    taskNameSystem,
			Kind:    domain.AnalysisKindSystem,
			Enabled: !config.SkipSystem,
			Execute: func(ctx context.Context) (analysisTaskResult, error) {
				ownedGraph, err := cloneModuleGraph(moduleGraph, moduleGraphErr)
				if err != nil {
					return nil, fmt.Errorf("prepare system analysis graph: %w", err)
				}
				request := domain.SystemAnalysisRequest{
					Paths:                files,
					Recursive:            domain.BoolPtr(executionCfg.Recursive),
					IncludePatterns:      []string{},
					ExcludePatterns:      []string{},
					OutputFormat:         domain.OutputFormatJSON,
					OutputWriter:         io.Discard,
					ConfigPath:           config.ConfigFile,
					AnalyzeDependencies:  domain.BoolPtr(executionCfg.SystemAnalyzeDependencies),
					AnalyzeArchitecture:  domain.BoolPtr(executionCfg.SystemAnalyzeArchitecture),
					IncludeStdLib:        domain.BoolPtr(executionCfg.ModuleGraph.IncludeStdLib),
					IncludeThirdParty:    domain.BoolPtr(executionCfg.ModuleGraph.IncludeThirdParty),
					FollowRelative:       domain.BoolPtr(executionCfg.ModuleGraph.FollowRelative),
					DetectCycles:         nil,
					ValidateArchitecture: nil,
				}
				response, err := uc.systemUseCase.analyzeGraphRequest(ctx, ownedGraph, request)
				return systemTaskResult{response: response}, err
			},
		})
	}

	// Community detection task.
	if uc.communityUseCase != nil {
		tasks = append(tasks, &analysisTask{
			Name:    taskNameCommunities,
			Kind:    domain.AnalysisKindCommunities,
			Enabled: !config.SkipCommunities,
			Execute: func(ctx context.Context) (analysisTaskResult, error) {
				ownedGraph, err := cloneModuleGraph(moduleGraph, moduleGraphErr)
				if err != nil {
					return nil, fmt.Errorf("prepare community analysis graph: %w", err)
				}
				request := domain.CommunityAnalysisRequest{
					Paths:             files,
					SourcePaths:       append([]string(nil), sourcePaths...),
					Recursive:         domain.BoolPtr(executionCfg.Recursive),
					IncludePatterns:   []string{},
					ExcludePatterns:   []string{},
					OutputFormat:      domain.OutputFormatJSON,
					OutputWriter:      io.Discard,
					ConfigPath:        config.ConfigFile,
					IncludeStdLib:     domain.BoolPtr(executionCfg.ModuleGraph.IncludeStdLib),
					IncludeThirdParty: domain.BoolPtr(executionCfg.ModuleGraph.IncludeThirdParty),
					FollowRelative:    domain.BoolPtr(executionCfg.ModuleGraph.FollowRelative),
				}
				response, err := uc.communityUseCase.analyzeGraphRequest(ctx, ownedGraph, request)
				return communityTaskResult{response: response}, err
			},
		})
	}

	return tasks
}

func (uc *AnalyzeUseCase) buildComplexityTaskRequest(config AnalyzeUseCaseConfig, files []string, executionCfg domain.AnalyzeExecutionConfig) domain.ComplexityRequest {
	minComplexity := config.MinComplexity
	if minComplexity <= 0 {
		minComplexity = executionCfg.ComplexityMinComplexity
	}

	// CLI flag values take precedence over config file when explicitly set (> 0).
	// Otherwise fall back to execution config (from config file or defaults).
	lowThreshold := executionCfg.ComplexityLowThreshold
	if config.LowThreshold > 0 {
		lowThreshold = config.LowThreshold
	}
	mediumThreshold := executionCfg.ComplexityMediumThreshold
	if config.MediumThreshold > 0 {
		mediumThreshold = config.MediumThreshold
	}
	cognitiveThreshold := executionCfg.CognitiveComplexityThreshold
	if config.CognitiveComplexityThreshold > 0 {
		cognitiveThreshold = config.CognitiveComplexityThreshold
	}
	nestingThreshold := executionCfg.NestingDepthThreshold
	if config.NestingDepthThreshold > 0 {
		nestingThreshold = config.NestingDepthThreshold
	}
	slocWarnThreshold := executionCfg.FunctionSLOCWarnThreshold
	if config.FunctionSLOCWarnThreshold > 0 {
		slocWarnThreshold = config.FunctionSLOCWarnThreshold
	}
	slocCriticalThreshold := executionCfg.FunctionSLOCCriticalThreshold
	if config.FunctionSLOCCriticalThreshold > 0 {
		slocCriticalThreshold = config.FunctionSLOCCriticalThreshold
	}

	return domain.ComplexityRequest{
		Paths:                        files,
		Recursive:                    domain.BoolPtr(executionCfg.Recursive),
		ShowDetails:                  domain.BoolPtr(executionCfg.ShowDetails),
		IncludePatterns:              []string{},
		ExcludePatterns:              []string{},
		OutputFormat:                 domain.OutputFormatJSON,
		OutputWriter:                 io.Discard,
		MinComplexity:                minComplexity,
		MaxComplexity:                executionCfg.ComplexityMaxComplexity,
		SortBy:                       domain.SortByComplexity,
		LowThreshold:                 lowThreshold,
		MediumThreshold:              mediumThreshold,
		CognitiveComplexityThreshold: cognitiveThreshold,
		NestingDepthThreshold:        nestingThreshold,

		FunctionSLOCWarnThreshold:     slocWarnThreshold,
		FunctionSLOCCriticalThreshold: slocCriticalThreshold,

		Enabled:         domain.BoolPtr(executionCfg.ComplexityEnabled),
		ReportUnchanged: domain.BoolPtr(executionCfg.ComplexityReportUnchanged),
		ConfigPath:      config.ConfigFile,
	}
}

func (uc *AnalyzeUseCase) buildCloneTaskRequest(config AnalyzeUseCaseConfig, files []string, executionCfg domain.AnalyzeExecutionConfig) domain.CloneRequest {
	// Sparse request: zero values mean "not set" and are filled from the
	// config file (or defaults) during MergeConfig inside the use case.
	return domain.CloneRequest{
		Paths:               files,
		Recursive:           domain.BoolPtr(executionCfg.Recursive),
		OutputFormat:        domain.OutputFormatJSON,
		OutputWriter:        io.Discard,
		SimilarityThreshold: config.CloneSimilarity,
		ConfigPath:          config.ConfigFile,
	}
}

// buildResponse builds the analyze response from task results
func (uc *AnalyzeUseCase) buildResponse(tasks []*analysisTask, startTime time.Time, pathIndex analysisPathIndex, coverage domain.AnalysisCoverage) (*domain.AnalyzeResponse, error) {
	response := &domain.AnalyzeResponse{
		GeneratedAt: time.Now(),
		Duration:    time.Since(startTime).Milliseconds(),
		Diagnostics: coverage.Diagnostics,
	}
	response.Summary.TotalFiles = coverage.TotalFiles
	response.Summary.AnalyzedFiles = coverage.AnalyzedFiles
	response.Summary.SkippedFiles = coverage.SkippedFiles

	// Collect results from tasks
	for _, task := range tasks {
		if !task.Enabled {
			continue
		}
		if err := markAnalysisEnabled(&response.Summary, task.Kind); err != nil {
			return response, err
		}
		if task.Error != nil {
			response.Failures = append(response.Failures, domain.NewAnalysisFailure(
				task.Kind,
				domain.AnalysisFailureCodeExecution,
				"",
				task.Error.Error(),
				task.Error,
			))
		}
		if task.Result != nil {
			if resultKind := task.Result.analysisKind(); resultKind != task.Kind {
				return response, fmt.Errorf("analysis task %s returned %s result", task.Kind, resultKind)
			}
			response.Failures = append(response.Failures, task.Result.analysisFailures()...)
			task.Result.applyTo(response)
		}
	}

	moduleQuality, err := aggregateModuleQuality(response, pathIndex)
	if err != nil {
		return response, fmt.Errorf("assemble module quality: %w", err)
	}
	response.ModuleQuality = moduleQuality

	// Calculate summary statistics
	uc.calculateSummary(&response.Summary, response)

	// Generate actionable suggestions from analysis results
	response.Suggestions = domain.GenerateSuggestions(response)

	return response, nil
}

// markAnalysisEnabled records an attempted analysis from its typed identity.
func markAnalysisEnabled(summary *domain.AnalyzeSummary, kind domain.AnalysisKind) error {
	switch kind {
	case domain.AnalysisKindComplexity:
		summary.ComplexityEnabled = true
	case domain.AnalysisKindDeadCode:
		summary.DeadCodeEnabled = true
	case domain.AnalysisKindClones:
		summary.CloneEnabled = true
	case domain.AnalysisKindCBO:
		summary.CBOEnabled = true
	case domain.AnalysisKindLCOM:
		summary.LCOMEnabled = true
	case domain.AnalysisKindSystem:
		summary.DepsEnabled = true
	case domain.AnalysisKindCommunities:
		summary.CommunitiesEnabled = true
	default:
		return fmt.Errorf("unsupported analysis kind %q", kind)
	}
	return nil
}

// calculateSummary calculates the summary statistics
func (uc *AnalyzeUseCase) calculateSummary(summary *domain.AnalyzeSummary, response *domain.AnalyzeResponse) {
	// Complexity statistics
	if response.Complexity != nil {
		summary.TotalFunctions = response.Complexity.Summary.TotalFunctions
		summary.TotalClassScopes = response.Complexity.Summary.TotalClassScopes
		summary.MaxClassComplexity = response.Complexity.Summary.MaxClassComplexity
		summary.MaxClassCognitiveComplexity = response.Complexity.Summary.MaxClassCognitiveComplexity
		summary.MaxClassNestingDepth = response.Complexity.Summary.MaxClassNestingDepth
		summary.HighComplexityClassScopeCount = response.Complexity.Summary.HighRiskClassScopes
		summary.FunctionsParsed = response.Complexity.Summary.FunctionsParsed
		summary.AverageComplexity = response.Complexity.Summary.AverageComplexity
		summary.AverageCognitiveComplexity = response.Complexity.Summary.AverageCognitiveComplexity
		summary.AverageNestingDepth = response.Complexity.Summary.AverageNestingDepth
		summary.HighComplexityCount = response.Complexity.Summary.HighRiskFunctions
	}

	// Dead code statistics
	if response.DeadCode != nil {
		summary.DeadCodeCount = response.DeadCode.Summary.TotalFindings
		summary.CriticalDeadCode = response.DeadCode.Summary.CriticalFindings
		summary.WarningDeadCode = response.DeadCode.Summary.WarningFindings
		summary.InfoDeadCode = response.DeadCode.Summary.InfoFindings
	}

	// Clone statistics
	if response.Clone != nil {
		summary.TotalClones = response.Clone.Statistics.TotalClones
		summary.ClonePairs = response.Clone.Statistics.TotalClonePairs
		summary.CloneGroups = response.Clone.Statistics.TotalCloneGroups

		// Calculate code duplication based on fragment ratio
		// Measures what proportion of all code fragments are involved in duplication
		totalFragments := response.Clone.Statistics.TotalFragments
		totalClones := response.Clone.Statistics.TotalClones

		if totalFragments > 0 && totalClones > 0 {
			summary.CodeDuplication = math.Min(domain.DuplicationThresholdHigh, float64(totalClones)/float64(totalFragments)*100)
		}
	}

	// CBO statistics
	if response.CBO != nil {
		summary.CBOClasses = response.CBO.Summary.TotalClasses
		summary.HighCouplingClasses = response.CBO.Summary.HighRiskClasses
		summary.MediumCouplingClasses = response.CBO.Summary.MediumRiskClasses
		summary.AverageCoupling = response.CBO.Summary.AverageCBO
	}

	// LCOM statistics
	if response.LCOM != nil {
		summary.LCOMClasses = response.LCOM.Summary.TotalClasses
		summary.HighLCOMClasses = response.LCOM.Summary.HighRiskClasses
		summary.MediumLCOMClasses = response.LCOM.Summary.MediumRiskClasses
		summary.AverageLCOM = response.LCOM.Summary.AverageLCOM
	}

	// System analysis statistics
	if response.System != nil {
		if response.System.DependencyAnalysis != nil {
			da := response.System.DependencyAnalysis
			summary.DepsTotalModules = da.TotalModules
			summary.DepsMaxDepth = da.MaxDepth
			if da.CircularDependencies != nil {
				summary.DepsModulesInCycles = da.CircularDependencies.TotalModulesInCycles
			}
			if da.CouplingAnalysis != nil {
				summary.DepsMainSequenceDeviation = da.CouplingAnalysis.MainSequenceDeviation
			}
		}
		if response.System.ArchitectureAnalysis != nil {
			aa := response.System.ArchitectureAnalysis
			summary.ArchCompliance = aa.ComplianceScore
		}
	}

	// Community detection statistics (feed the community risk score / health penalty)
	if response.Communities != nil {
		c := response.Communities
		summary.CommunityCount = c.TotalCommunities
		summary.CommunityModularity = c.Modularity
		// Use the analysis bridge count, not the emitted list, so the health
		// penalty is independent of whether bridge modules are reported.
		summary.CommunityBridgeModules = c.BridgeModuleCount
		internalEdges, crossEdges := 0, 0
		for i := range c.Communities {
			internalEdges += c.Communities[i].InternalEdges
			crossEdges += c.Communities[i].OutgoingCrossCommunityEdges
		}
		summary.CommunityInternalEdges = internalEdges
		summary.CommunityCrossEdges = crossEdges
		summary.CommunityPackageAlignment = c.PackageAlignmentScore
		summary.CommunityLayerAlignment = c.LayerAlignmentScore
	}

	// Calculate health score with error handling
	if err := summary.CalculateHealthScore(); err != nil {
		// Log warning
		log.Printf("WARNING: Failed to calculate health score: %v", err)

		// Fallback processing: calculate simple score
		summary.HealthScore = summary.CalculateFallbackScore()
		summary.Grade = domain.GetGradeFromScore(summary.HealthScore)
	}
}

func (uc *AnalyzeUseCase) loadExecutionConfig(configPath string, paths []string) (domain.AnalyzeExecutionConfig, error) {
	targetPath := ""
	if len(paths) > 0 {
		targetPath = paths[0]
	}

	return uc.configLoader.LoadAnalyzeExecutionConfig(configPath, targetPath)
}

// estimateTaskSeconds estimates the duration of each enabled analysis task in
// seconds, keyed by typed analysis identity. The formulas capture how each analysis scales
// with file count; absolute accuracy comes from calibration against actual
// timings (see applyTimingFactors and UpdateAnalysisTimingFactors).
func (uc *AnalyzeUseCase) estimateTaskSeconds(fileCount int, config AnalyzeUseCaseConfig, executionCfg domain.AnalyzeExecutionConfig) map[domain.AnalysisKind]float64 {
	n := float64(fileCount)
	estimates := map[domain.AnalysisKind]float64{}

	// Linear analyses (fast)
	if uc.complexityUseCase != nil && !config.SkipComplexity {
		estimates[domain.AnalysisKindComplexity] = 0.01 * n // Complexity: ~0.01s per file
	}
	if uc.deadCodeUseCase != nil && !config.SkipDeadCode {
		estimates[domain.AnalysisKindDeadCode] = 0.01 * n // Dead Code: ~0.01s per file
	}
	if uc.cboUseCase != nil && !config.SkipCBO {
		estimates[domain.AnalysisKindCBO] = 0.01 * n // CBO: ~0.01s per file
	}
	if uc.lcomUseCase != nil && !config.SkipLCOM {
		estimates[domain.AnalysisKindLCOM] = 0.01 * n // LCOM: ~0.01s per file
	}
	if uc.systemUseCase != nil && !config.SkipSystem {
		estimates[domain.AnalysisKindSystem] = 0.02 * n // System: ~0.02s per file (slightly heavier)
	}
	if uc.communityUseCase != nil && !config.SkipCommunities {
		estimates[domain.AnalysisKindCommunities] = 0.02 * n
	}

	// Clone detection - account for LSH configuration
	if uc.cloneUseCase != nil && !config.SkipClones {
		// Estimate fragment count (empirical average: ~5.0 fragments per file)
		estimatedFragments := n * 5.0

		// Determine LSH usage using centralized logic.
		useLSH := domain.ShouldUseLSHWithPairEstimate(
			executionCfg.CloneLSHEnabled,
			int(estimatedFragments),
			executionCfg.CloneLSHAutoThreshold,
			domain.DefaultLSHAutoPairThreshold,
		)

		if useLSH {
			// LSH enabled: Near-linear O(n^1.1) complexity
			// LSH candidate filtering significantly reduces the number of APTED comparisons
			estimates[domain.AnalysisKindClones] = 0.01 * math.Pow(estimatedFragments, 1.1)
		} else {
			// LSH disabled: Quadratic O(n²) complexity - full pairwise comparison
			// All fragment pairs are compared via expensive APTED tree edit distance
			estimates[domain.AnalysisKindClones] = 0.001 * estimatedFragments * estimatedFragments
		}
	}

	return estimates
}

// startProgressUpdater starts a background goroutine that periodically renders
// the tracker's current progress estimate
func (uc *AnalyzeUseCase) startProgressUpdater(tracker *analysisProgressTracker) chan struct{} {
	done := make(chan struct{})

	// Start progress bar
	uc.progressManager.Start()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				uc.progressManager.Update(tracker.Percent(), 100)

			case <-done:
				return
			}
		}
	}()

	return done
}
