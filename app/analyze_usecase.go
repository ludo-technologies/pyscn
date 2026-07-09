package app

import (
	"context"
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

	// Clone detection options
	EnableDFA bool // Enable Data Flow Analysis for enhanced Type-4 detection

	ConfigFile string
	Verbose    bool
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

// AnalysisTask represents a single analysis task
type AnalysisTask struct {
	Name    string
	Enabled bool
	Execute func(context.Context) (interface{}, error)
	Result  interface{}
	Error   error
}

// Execute performs comprehensive analysis
func (uc *AnalyzeUseCase) Execute(ctx context.Context, useCaseCfg AnalyzeUseCaseConfig, paths []string) (*domain.AnalyzeResponse, error) {
	startTime := time.Now()

	executionCfg, err := uc.loadExecutionConfig(useCaseCfg.ConfigFile, paths)
	if err != nil {
		return nil, err
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
	files, err := uc.fileReader.CollectPythonFiles(
		paths,
		executionCfg.Recursive,
		executionCfg.IncludePatterns,
		executionCfg.ExcludePatterns,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to collect Python files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Python files found in the specified paths")
	}

	// Estimate per-task durations from file count, then calibrate with actual
	// timings recorded by previous runs on this project (if any)
	estimatedSeconds := uc.estimateTaskSeconds(len(files), useCaseCfg, executionCfg)

	var snapshot *service.ProjectSnapshot
	if uc.needsProjectSnapshot(useCaseCfg) {
		snapshot = service.BuildProjectSnapshotWithOptions(ctx, files, service.ProjectSnapshotOptions{
			IncludeRawMetrics: uc.complexityUseCase != nil && !useCaseCfg.SkipComplexity,
		})
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
	tasks := uc.createAnalysisTasks(useCaseCfg, paths, files, snapshot, executionCfg)

	// Execute tasks in parallel
	var wg sync.WaitGroup
	for _, task := range tasks {
		if !task.Enabled {
			continue
		}

		wg.Add(1)
		go func(t *AnalysisTask) {
			defer wg.Done()
			result, err := t.Execute(ctx)
			t.Result = result
			t.Error = err
			if tracker != nil {
				tracker.TaskCompleted(t.Name)
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

	// Check for errors
	var errors []error
	for _, task := range tasks {
		if task.Enabled && task.Error != nil {
			errors = append(errors, fmt.Errorf("%s: %w", task.Name, task.Error))
		}
	}

	// Build response
	response := uc.buildResponse(tasks, startTime)

	// Return aggregated error if any tasks failed
	if len(errors) > 0 {
		return response, fmt.Errorf("analysis completed with %d error(s): %w", len(errors), errors[0])
	}

	return response, nil
}

func (uc *AnalyzeUseCase) needsProjectSnapshot(config AnalyzeUseCaseConfig) bool {
	return (uc.complexityUseCase != nil && !config.SkipComplexity) ||
		(uc.deadCodeUseCase != nil && !config.SkipDeadCode) ||
		(uc.cboUseCase != nil && !config.SkipCBO) ||
		(uc.lcomUseCase != nil && !config.SkipLCOM)
}

// createAnalysisTasks creates the analysis tasks based on configuration
func (uc *AnalyzeUseCase) createAnalysisTasks(config AnalyzeUseCaseConfig, sourcePaths []string, files []string, snapshot *service.ProjectSnapshot, executionCfg domain.AnalyzeExecutionConfig) []*AnalysisTask {
	tasks := []*AnalysisTask{}

	// Complexity analysis task
	if uc.complexityUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    taskNameComplexity,
			Enabled: !config.SkipComplexity,
			Execute: func(ctx context.Context) (interface{}, error) {
				request := uc.buildComplexityTaskRequest(config, files, executionCfg)
				return uc.complexityUseCase.analyzeSnapshotRequest(ctx, snapshot, request)
			},
		})
	}

	// Dead code analysis task
	if uc.deadCodeUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    taskNameDeadCode,
			Enabled: !config.SkipDeadCode,
			Execute: func(ctx context.Context) (interface{}, error) {
				request := domain.DeadCodeRequest{
					Paths:           files,
					Recursive:       false,
					IncludePatterns: []string{},
					ExcludePatterns: []string{},
					OutputFormat:    domain.OutputFormatJSON,
					OutputWriter:    io.Discard,
					MinSeverity:     config.MinSeverity,
					SortBy:          domain.DeadCodeSortBySeverity,
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
				return uc.deadCodeUseCase.analyzeSnapshotRequest(ctx, snapshot, request)
			},
		})
	}

	// Clone detection task
	if uc.cloneUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    taskNameClones,
			Enabled: !config.SkipClones,
			Execute: func(ctx context.Context) (interface{}, error) {
				request := uc.buildCloneTaskRequest(config, files)
				return uc.cloneUseCase.ExecuteAndReturn(ctx, request)
			},
		})
	}

	// CBO analysis task
	if uc.cboUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    taskNameCBO,
			Enabled: !config.SkipCBO,
			Execute: func(ctx context.Context) (interface{}, error) {
				request := domain.CBORequest{
					Paths:           files,
					Recursive:       nil, // Let config file values take precedence
					IncludePatterns: []string{},
					ExcludePatterns: []string{},
					OutputFormat:    domain.OutputFormatJSON,
					OutputWriter:    io.Discard,
					MinCBO:          config.MinCBO,
					LowThreshold:    domain.DefaultCBOLowThreshold,
					MediumThreshold: domain.DefaultCBOMediumThreshold,
					SortBy:          domain.SortByCoupling,
					ConfigPath:      config.ConfigFile,
					// Boolean options left as nil to allow config file values to take precedence
					ShowZeros:             nil,
					IncludeBuiltins:       nil,
					IncludeImports:        nil,
					GroupNamespaceImports: nil,
				}
				return uc.cboUseCase.analyzeSnapshotRequest(ctx, snapshot, request)
			},
		})
	}

	// LCOM analysis task
	if uc.lcomUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    taskNameLCOM,
			Enabled: !config.SkipLCOM,
			Execute: func(ctx context.Context) (interface{}, error) {
				request := domain.LCOMRequest{
					Paths:           files,
					Recursive:       nil, // Let config file values take precedence
					IncludePatterns: []string{},
					ExcludePatterns: []string{},
					OutputFormat:    domain.OutputFormatJSON,
					OutputWriter:    io.Discard,
					LowThreshold:    0, // Zero: let config file values take precedence via merge
					MediumThreshold: 0, // Zero: let config file values take precedence via merge
					SortBy:          domain.SortByCohesion,
					ConfigPath:      config.ConfigFile,
				}
				return uc.lcomUseCase.analyzeSnapshotRequest(ctx, snapshot, request)
			},
		})
	}

	// System analysis task
	if uc.systemUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    taskNameSystem,
			Enabled: !config.SkipSystem,
			Execute: func(ctx context.Context) (interface{}, error) {
				request := domain.SystemAnalysisRequest{
					Paths:                files,
					Recursive:            nil, // Let config file values take precedence
					IncludePatterns:      []string{},
					ExcludePatterns:      []string{},
					OutputFormat:         domain.OutputFormatJSON,
					OutputWriter:         io.Discard,
					ConfigPath:           config.ConfigFile,
					AnalyzeDependencies:  domain.BoolPtr(executionCfg.SystemAnalyzeDependencies),
					AnalyzeArchitecture:  domain.BoolPtr(executionCfg.SystemAnalyzeArchitecture),
					IncludeStdLib:        nil,
					IncludeThirdParty:    nil,
					FollowRelative:       nil,
					DetectCycles:         nil,
					ValidateArchitecture: nil,
				}
				return uc.systemUseCase.AnalyzeAndReturn(ctx, request)
			},
		})
	}

	// Community detection task.
	if uc.communityUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    taskNameCommunities,
			Enabled: !config.SkipCommunities,
			Execute: func(ctx context.Context) (interface{}, error) {
				request := domain.CommunityAnalysisRequest{
					Paths:           files,
					SourcePaths:     append([]string(nil), sourcePaths...),
					Recursive:       nil,
					IncludePatterns: []string{},
					ExcludePatterns: []string{},
					OutputFormat:    domain.OutputFormatJSON,
					OutputWriter:    io.Discard,
					ConfigPath:      config.ConfigFile,
				}
				return uc.communityUseCase.AnalyzeAndReturn(ctx, request)
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

	return domain.ComplexityRequest{
		Paths:                        files,
		Recursive:                    false,
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
		Enabled:                      domain.BoolPtr(executionCfg.ComplexityEnabled),
		ReportUnchanged:              domain.BoolPtr(executionCfg.ComplexityReportUnchanged),
		ConfigPath:                   config.ConfigFile,
	}
}

func (uc *AnalyzeUseCase) buildCloneTaskRequest(config AnalyzeUseCaseConfig, files []string) domain.CloneRequest {
	// Sparse request: zero values mean "not set" and are filled from the
	// config file (or defaults) during MergeConfig inside the use case.
	return domain.CloneRequest{
		Paths:               files,
		OutputFormat:        domain.OutputFormatJSON,
		OutputWriter:        io.Discard,
		SimilarityThreshold: config.CloneSimilarity,
		ConfigPath:          config.ConfigFile,
	}
}

// buildResponse builds the analyze response from task results
func (uc *AnalyzeUseCase) buildResponse(tasks []*AnalysisTask, startTime time.Time) *domain.AnalyzeResponse {
	response := &domain.AnalyzeResponse{
		GeneratedAt: time.Now(),
		Duration:    time.Since(startTime).Milliseconds(),
	}

	// Collect results from tasks
	for _, task := range tasks {
		if !task.Enabled {
			continue
		}

		switch result := task.Result.(type) {
		case *domain.ComplexityResponse:
			response.Summary.ComplexityEnabled = true
			if result != nil {
				response.Complexity = result
			}
		case *domain.DeadCodeResponse:
			response.Summary.DeadCodeEnabled = true
			if result != nil {
				response.DeadCode = result
			}
		case *domain.CloneResponse:
			response.Summary.CloneEnabled = true
			if result != nil {
				response.Clone = result
			}
		case *domain.CBOResponse:
			response.Summary.CBOEnabled = true
			if result != nil {
				response.CBO = result
			}
		case *domain.LCOMResponse:
			response.Summary.LCOMEnabled = true
			if result != nil {
				response.LCOM = result
			}
		case *domain.SystemAnalysisResponse:
			response.Summary.DepsEnabled = true
			if result != nil {
				response.System = result
				if result.ArchitectureAnalysis != nil {
					response.Summary.ArchEnabled = true
				}
			}
		case *domain.CommunityAnalysisResult:
			response.Summary.CommunitiesEnabled = true
			if result != nil {
				response.Communities = result
			}
		case nil:
			uc.markSummaryForTask(&response.Summary, task.Name)
		default:
			uc.markSummaryForTask(&response.Summary, task.Name)
		}
	}

	// Calculate summary statistics
	uc.calculateSummary(&response.Summary, response)

	// Generate actionable suggestions from analysis results
	response.Suggestions = domain.GenerateSuggestions(response)

	return response
}

// markSummaryForTask ensures the summary reflects analyses that attempted to run
func (uc *AnalyzeUseCase) markSummaryForTask(summary *domain.AnalyzeSummary, taskName string) {
	switch taskName {
	case "Complexity Analysis":
		summary.ComplexityEnabled = true
	case "Dead Code Detection":
		summary.DeadCodeEnabled = true
	case "Clone Detection":
		summary.CloneEnabled = true
	case "Class Coupling (CBO)":
		summary.CBOEnabled = true
	case "Class Cohesion (LCOM)":
		summary.LCOMEnabled = true
	case "System Analysis":
		summary.DepsEnabled = true
	case taskNameCommunities:
		summary.CommunitiesEnabled = true
	}
}

// calculateSummary calculates the summary statistics
func (uc *AnalyzeUseCase) calculateSummary(summary *domain.AnalyzeSummary, response *domain.AnalyzeResponse) {
	// Complexity statistics
	if response.Complexity != nil {
		summary.TotalFiles = response.Complexity.Summary.FilesAnalyzed
		summary.AnalyzedFiles = response.Complexity.Summary.FilesAnalyzed
		summary.TotalFunctions = len(response.Complexity.Functions)
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
// seconds, keyed by task name. The formulas capture how each analysis scales
// with file count; absolute accuracy comes from calibration against actual
// timings (see applyTimingFactors and UpdateAnalysisTimingFactors).
func (uc *AnalyzeUseCase) estimateTaskSeconds(fileCount int, config AnalyzeUseCaseConfig, executionCfg domain.AnalyzeExecutionConfig) map[string]float64 {
	n := float64(fileCount)
	estimates := map[string]float64{}

	// Linear analyses (fast)
	if uc.complexityUseCase != nil && !config.SkipComplexity {
		estimates[taskNameComplexity] = 0.01 * n // Complexity: ~0.01s per file
	}
	if uc.deadCodeUseCase != nil && !config.SkipDeadCode {
		estimates[taskNameDeadCode] = 0.01 * n // Dead Code: ~0.01s per file
	}
	if uc.cboUseCase != nil && !config.SkipCBO {
		estimates[taskNameCBO] = 0.01 * n // CBO: ~0.01s per file
	}
	if uc.lcomUseCase != nil && !config.SkipLCOM {
		estimates[taskNameLCOM] = 0.01 * n // LCOM: ~0.01s per file
	}
	if uc.systemUseCase != nil && !config.SkipSystem {
		estimates[taskNameSystem] = 0.02 * n // System: ~0.02s per file (slightly heavier)
	}
	if uc.communityUseCase != nil && !config.SkipCommunities {
		estimates[taskNameCommunities] = 0.02 * n
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
			estimates[taskNameClones] = 0.01 * math.Pow(estimatedFragments, 1.1)
		} else {
			// LSH disabled: Quadratic O(n²) complexity - full pairwise comparison
			// All fragment pairs are compared via expensive APTED tree edit distance
			estimates[taskNameClones] = 0.001 * estimatedFragments * estimatedFragments
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
