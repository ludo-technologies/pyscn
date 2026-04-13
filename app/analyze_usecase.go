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
	SkipComplexity bool
	SkipDeadCode   bool
	SkipClones     bool
	SkipCBO        bool
	SkipLCOM       bool
	SkipSystem     bool

	MinComplexity   int
	MinSeverity     domain.DeadCodeSeverity
	CloneSimilarity float64
	MinCBO          int

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
		fileReader:        b.fileReader,
		configLoader:      b.configLoader,
		formatter:         b.formatter,
		progressManager:   b.progressManager,
		parallelExecutor:  b.parallelExecutor,
		errorCategorizer:  b.errorCategorizer,
	}, nil
}

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

	// Calculate estimated time based on file count and enabled analyses
	estimatedTime := uc.calculateEstimatedTime(len(files), useCaseCfg, executionCfg)

	// Start unified progress tracking with time-based estimation
	var progressDone chan struct{}
	if uc.progressManager != nil {
		uc.progressManager.Initialize(100) // 100% based progress
		progressDone = uc.startTimeBasedProgressUpdater(estimatedTime)
	}

	// Create analysis tasks
	tasks := uc.createAnalysisTasks(useCaseCfg, files, executionCfg)

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

// createAnalysisTasks creates the analysis tasks based on configuration
func (uc *AnalyzeUseCase) createAnalysisTasks(config AnalyzeUseCaseConfig, files []string, executionCfg domain.AnalyzeExecutionConfig) []*AnalysisTask {
	tasks := []*AnalysisTask{}

	// Complexity analysis task
	if uc.complexityUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    "Complexity Analysis",
			Enabled: !config.SkipComplexity,
			Execute: func(ctx context.Context) (interface{}, error) {
				request := uc.buildComplexityTaskRequest(config, files, executionCfg)
				return uc.complexityUseCase.analyzeResolvedRequest(ctx, request)
			},
		})
	}

	// Dead code analysis task
	if uc.deadCodeUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    "Dead Code Detection",
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
				return uc.deadCodeUseCase.AnalyzeAndReturn(ctx, request)
			},
		})
	}

	// Clone detection task
	if uc.cloneUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    "Clone Detection",
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
			Name:    "Class Coupling (CBO)",
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
					ShowZeros:       nil,
					IncludeBuiltins: nil,
					IncludeImports:  nil,
				}
				return uc.cboUseCase.AnalyzeAndReturn(ctx, request)
			},
		})
	}

	// LCOM analysis task
	if uc.lcomUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    "Class Cohesion (LCOM)",
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
				return uc.lcomUseCase.AnalyzeAndReturn(ctx, request)
			},
		})
	}

	// System analysis task
	if uc.systemUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    "System Analysis",
			Enabled: !config.SkipSystem,
			Execute: func(ctx context.Context) (interface{}, error) {
				request := domain.SystemAnalysisRequest{
					Paths:           files,
					Recursive:       nil, // Let config file values take precedence
					IncludePatterns: []string{},
					ExcludePatterns: []string{},
					OutputFormat:    domain.OutputFormatJSON,
					OutputWriter:    io.Discard,
					ConfigPath:      config.ConfigFile,
					// Boolean options left as nil to allow config file values to take precedence
					AnalyzeDependencies:  nil,
					AnalyzeArchitecture:  nil,
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

	return tasks
}

func (uc *AnalyzeUseCase) buildComplexityTaskRequest(config AnalyzeUseCaseConfig, files []string, executionCfg domain.AnalyzeExecutionConfig) domain.ComplexityRequest {
	minComplexity := config.MinComplexity
	if minComplexity <= 0 {
		minComplexity = executionCfg.ComplexityMinComplexity
	}

	return domain.ComplexityRequest{
		Paths:           files,
		Recursive:       false,
		IncludePatterns: []string{},
		ExcludePatterns: []string{},
		OutputFormat:    domain.OutputFormatJSON,
		OutputWriter:    io.Discard,
		MinComplexity:   minComplexity,
		MaxComplexity:   executionCfg.ComplexityMaxComplexity,
		SortBy:          domain.SortByComplexity,
		LowThreshold:    executionCfg.ComplexityLowThreshold,
		MediumThreshold: executionCfg.ComplexityMediumThreshold,
		Enabled:         domain.BoolPtr(executionCfg.ComplexityEnabled),
		ReportUnchanged: domain.BoolPtr(executionCfg.ComplexityReportUnchanged),
		ConfigPath:      config.ConfigFile,
	}
}

func (uc *AnalyzeUseCase) buildCloneTaskRequest(config AnalyzeUseCaseConfig, files []string) domain.CloneRequest {
	request := *domain.DefaultCloneRequest()
	request.Paths = files
	request.OutputFormat = domain.OutputFormatJSON
	request.OutputWriter = io.Discard
	request.SimilarityThreshold = config.CloneSimilarity
	request.ConfigPath = config.ConfigFile

	return request
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

		// Calculate code duplication based on K-Core clone groups
		// K-Core groups represent true duplication clusters where each fragment
		// is similar to at least k other fragments (default k=2)
		// This filters out false positives from structural similarity
		totalLines := response.Clone.Statistics.LinesAnalyzed
		groupCount := response.Clone.Statistics.TotalCloneGroups

		if totalLines > 0 && groupCount > 0 {
			// Calculate group density: groups per 1000 lines of code
			// This normalizes for project size
			linesInThousands := float64(totalLines) / domain.GroupDensityLinesUnit
			if linesInThousands < domain.GroupDensityMinLines {
				linesInThousands = domain.GroupDensityMinLines
			}
			groupDensity := float64(groupCount) / linesInThousands

			// Convert density to percentage for penalty calculation
			// 0.5 groups/1000 lines = 10% duplication (max penalty)
			// This makes the scoring stricter for duplicate code clusters
			summary.CodeDuplication = math.Min(domain.DuplicationThresholdHigh, groupDensity*domain.GroupDensityCoefficient)
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

// calculateEstimatedTime estimates the total analysis time based on file count and enabled analyses
func (uc *AnalyzeUseCase) calculateEstimatedTime(fileCount int, config AnalyzeUseCaseConfig, executionCfg domain.AnalyzeExecutionConfig) float64 {
	n := float64(fileCount)
	totalTime := 0.0

	// Linear analyses (fast)
	if !config.SkipComplexity {
		totalTime += 0.01 * n // Complexity: ~0.01s per file
	}
	if !config.SkipDeadCode {
		totalTime += 0.01 * n // Dead Code: ~0.01s per file
	}
	if !config.SkipCBO {
		totalTime += 0.01 * n // CBO: ~0.01s per file
	}
	if !config.SkipLCOM {
		totalTime += 0.01 * n // LCOM: ~0.01s per file
	}
	if !config.SkipSystem {
		totalTime += 0.02 * n // System: ~0.02s per file (slightly heavier)
	}

	// Clone detection - account for LSH configuration
	if !config.SkipClones {
		// Estimate fragment count (empirical average: ~5.0 fragments per file)
		estimatedFragments := n * 5.0

		// Determine LSH usage using centralized logic
		useLSH := domain.ShouldUseLSH(executionCfg.CloneLSHEnabled, int(estimatedFragments), executionCfg.CloneLSHAutoThreshold)

		if useLSH {
			// LSH enabled: Near-linear O(n^1.1) complexity
			// LSH candidate filtering significantly reduces the number of APTED comparisons
			// Exponent 1.1 and coefficient 0.01 are empirically tuned
			// Note: Actual performance varies by environment and code characteristics
			totalTime += 0.01 * math.Pow(estimatedFragments, 1.1)
		} else {
			// LSH disabled: Quadratic O(n²) complexity - full pairwise comparison
			// All fragment pairs are compared via expensive APTED tree edit distance
			// This becomes the bottleneck for codebases with many fragments
			// Coefficient 0.001 accounts for fragment count estimation
			totalTime += 0.001 * estimatedFragments * estimatedFragments
		}
	}

	// Minimum time to avoid division by zero
	if totalTime < 0.1 {
		totalTime = 0.1
	}

	return totalTime
}

// startTimeBasedProgressUpdater starts a background goroutine that updates progress based on elapsed time
func (uc *AnalyzeUseCase) startTimeBasedProgressUpdater(estimatedTime float64) chan struct{} {
	done := make(chan struct{})
	startTime := time.Now()

	// Start progress bar
	uc.progressManager.Start()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				elapsed := time.Since(startTime).Seconds()
				// Progress increases with elapsed time, but caps at 99%
				// (we'll set it to 100% when tasks actually complete)
				progress := int((elapsed / estimatedTime) * 100)
				if progress > 99 {
					progress = 99
				}
				uc.progressManager.Update(progress, 100)

			case <-done:
				return
			}
		}
	}()

	return done
}
