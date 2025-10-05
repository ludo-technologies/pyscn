package app

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/ludo-technologies/pyscn/service"
)

// AnalyzeUseCaseConfig holds configuration for the analyze use case
type AnalyzeUseCaseConfig struct {
	SkipComplexity bool
	SkipDeadCode   bool
	SkipClones     bool
	SkipCBO        bool
	SkipSystem     bool

	MinComplexity   int
	MinSeverity     domain.DeadCodeSeverity
	CloneSimilarity float64
	MinCBO          int

	ConfigFile string
	Verbose    bool
}

// AnalyzeUseCase orchestrates comprehensive analysis
type AnalyzeUseCase struct {
	complexityUseCase *ComplexityUseCase
	deadCodeUseCase   *DeadCodeUseCase
	cloneUseCase      *CloneUseCase
	cboUseCase        *CBOUseCase
	systemUseCase     *SystemAnalysisUseCase

	fileReader       domain.FileReader
	formatter        *service.AnalyzeFormatter
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
	systemUseCase     *SystemAnalysisUseCase

	fileReader       domain.FileReader
	formatter        *service.AnalyzeFormatter
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

// WithFormatter sets the formatter
func (b *AnalyzeUseCaseBuilder) WithFormatter(f *service.AnalyzeFormatter) *AnalyzeUseCaseBuilder {
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
		systemUseCase:     b.systemUseCase,
		fileReader:        b.fileReader,
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
func (uc *AnalyzeUseCase) Execute(ctx context.Context, config AnalyzeUseCaseConfig, paths []string) (*domain.AnalyzeResponse, error) {
	startTime := time.Now()

	// Load configuration to get file patterns
	includePatterns, excludePatterns, patternErr := uc.getFilePatterns(config.ConfigFile, paths)
	if patternErr != nil {
		return nil, patternErr
	}

	// Validate and collect files using configured patterns
	files, err := uc.fileReader.CollectPythonFiles(
		paths,
		true, // recursive
		includePatterns,
		excludePatterns,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to collect Python files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no Python files found in the specified paths")
	}

	// Calculate estimated time based on file count and enabled analyses
	estimatedTime := uc.calculateEstimatedTime(len(files), config)

	// Start unified progress tracking with time-based estimation
	var progressDone chan struct{}
	if uc.progressManager != nil {
		uc.progressManager.Initialize(100) // 100% based progress
		progressDone = uc.startTimeBasedProgressUpdater(estimatedTime)
	}

	// Create analysis tasks
	tasks := uc.createAnalysisTasks(config, files)

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
func (uc *AnalyzeUseCase) createAnalysisTasks(config AnalyzeUseCaseConfig, files []string) []*AnalysisTask {
	tasks := []*AnalysisTask{}

	// Complexity analysis task
	if uc.complexityUseCase != nil {
		tasks = append(tasks, &AnalysisTask{
			Name:    "Complexity Analysis",
			Enabled: !config.SkipComplexity,
			Execute: func(ctx context.Context) (interface{}, error) {
				request := domain.ComplexityRequest{
					Paths:           files,
					OutputFormat:    domain.OutputFormatJSON,
					OutputWriter:    io.Discard,
					MinComplexity:   config.MinComplexity,
					LowThreshold:    9,
					MediumThreshold: 19,
					SortBy:          domain.SortByComplexity,
					ConfigPath:      config.ConfigFile,
				}
				return uc.complexityUseCase.AnalyzeAndReturn(ctx, request)
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
					Paths:        files,
					OutputFormat: domain.OutputFormatJSON,
					OutputWriter: io.Discard,
					MinSeverity:  config.MinSeverity,
					SortBy:       domain.DeadCodeSortBySeverity,
					ConfigPath:   config.ConfigFile,
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
				// Start with defaults to ensure all required fields are populated
				defaultReq := domain.DefaultCloneRequest()
				request := domain.CloneRequest{
					Paths:               files,
					OutputFormat:        domain.OutputFormatJSON,
					OutputWriter:        io.Discard,
					MinLines:            defaultReq.MinLines,
					MinNodes:            defaultReq.MinNodes,
					SimilarityThreshold: config.CloneSimilarity,
					Type1Threshold:      defaultReq.Type1Threshold,
					Type2Threshold:      defaultReq.Type2Threshold,
					Type3Threshold:      defaultReq.Type3Threshold,
					Type4Threshold:      defaultReq.Type4Threshold,
					GroupClones:         true, // Enable clone grouping (default behavior)
					ConfigPath:          config.ConfigFile,
				}
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
					OutputFormat:    domain.OutputFormatJSON,
					OutputWriter:    io.Discard,
					MinCBO:          config.MinCBO,
					LowThreshold:    5,
					MediumThreshold: 10,
					SortBy:          domain.SortByCoupling,
				}
				return uc.cboUseCase.AnalyzeAndReturn(ctx, request)
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
					Paths:               files,
					OutputFormat:        domain.OutputFormatJSON,
					OutputWriter:        io.Discard,
					AnalyzeDependencies: true,
					AnalyzeArchitecture: true,
					ConfigPath:          config.ConfigFile,
				}
				return uc.systemUseCase.AnalyzeAndReturn(ctx, request)
			},
		})
	}

	return tasks
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
	}

	// Clone statistics
	if response.Clone != nil {
		summary.TotalClones = response.Clone.Statistics.TotalClones
		summary.ClonePairs = response.Clone.Statistics.TotalClonePairs
		summary.CloneGroups = response.Clone.Statistics.TotalCloneGroups

		// Calculate code duplication percentage
		if response.Clone.Statistics.LinesAnalyzed > 0 {
			duplicatedLineSet := make(map[string]map[int]bool)

			for _, pair := range response.Clone.ClonePairs {
				// Track unique lines in Clone1
				if pair.Clone1 != nil && pair.Clone1.Location != nil {
					filePath := pair.Clone1.Location.FilePath
					if duplicatedLineSet[filePath] == nil {
						duplicatedLineSet[filePath] = make(map[int]bool)
					}
					for line := pair.Clone1.Location.StartLine; line <= pair.Clone1.Location.EndLine; line++ {
						duplicatedLineSet[filePath][line] = true
					}
				}
				// Track unique lines in Clone2
				if pair.Clone2 != nil && pair.Clone2.Location != nil {
					filePath := pair.Clone2.Location.FilePath
					if duplicatedLineSet[filePath] == nil {
						duplicatedLineSet[filePath] = make(map[int]bool)
					}
					for line := pair.Clone2.Location.StartLine; line <= pair.Clone2.Location.EndLine; line++ {
						duplicatedLineSet[filePath][line] = true
					}
				}
			}

			// Count unique duplicated lines
			duplicatedLines := 0
			for _, lines := range duplicatedLineSet {
				duplicatedLines += len(lines)
			}

			summary.CodeDuplication = (float64(duplicatedLines) / float64(response.Clone.Statistics.LinesAnalyzed)) * 100
		}
	}

	// CBO statistics
	if response.CBO != nil {
		summary.CBOClasses = response.CBO.Summary.TotalClasses
		summary.HighCouplingClasses = response.CBO.Summary.HighRiskClasses
		summary.MediumCouplingClasses = response.CBO.Summary.MediumRiskClasses
		summary.AverageCoupling = response.CBO.Summary.AverageCBO
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

// getFilePatterns loads file patterns from configuration or returns defaults
func (uc *AnalyzeUseCase) getFilePatterns(configPath string, paths []string) ([]string, []string, error) {
	// Default patterns
	defaultInclude := []string{"*.py", "*.pyi"}
	defaultExclude := []string{"test_*.py", "*_test.py"}

	// Try to load configuration
	targetPath := ""
	if len(paths) > 0 {
		targetPath = paths[0]
	}

	cfg, err := config.LoadConfigWithTarget(configPath, targetPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load configuration for pattern resolution: %w", err)
	}
	if cfg == nil {
		return defaultInclude, defaultExclude, nil
	}

	// Use configured patterns if available
	includePatterns := cfg.Analysis.IncludePatterns
	excludePatterns := cfg.Analysis.ExcludePatterns

	// Fall back to defaults if not specified
	if len(includePatterns) == 0 {
		includePatterns = defaultInclude
	}
	if len(excludePatterns) == 0 {
		excludePatterns = defaultExclude
	}

	return includePatterns, excludePatterns, nil
}

// calculateEstimatedTime estimates the total analysis time based on file count and enabled analyses
func (uc *AnalyzeUseCase) calculateEstimatedTime(fileCount int, config AnalyzeUseCaseConfig) float64 {
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
	if !config.SkipSystem {
		totalTime += 0.02 * n // System: ~0.02s per file (slightly heavier)
	}

	// Clone detection (quadratic - the dominant factor)
	if !config.SkipClones {
		// More conservative estimate: O(n²) is the bottleneck
		// For 207 files: 0.01 * 207² ≈ 428 seconds
		totalTime += 0.01 * n * n
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
