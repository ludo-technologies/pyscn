package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

// AnalysisStatus represents the status of an individual analysis
type AnalysisStatus struct {
	Name      string
	Enabled   bool
	Started   bool
	Completed bool
	Success   bool
	Error     error
	Duration  time.Duration
	StartTime time.Time
}

// AnalysisResult aggregates all analysis results
type AnalysisResult struct {
	Complexity *AnalysisStatus
	DeadCode   *AnalysisStatus
	Clones     *AnalysisStatus
	Overall    struct {
		StartTime    time.Time
		EndTime      time.Time
		TotalTime    time.Duration
		SuccessCount int
		FailureCount int
		SkippedCount int
	}
}

// AnalyzeCommand represents the comprehensive analysis command
type AnalyzeCommand struct {
	// Output configuration
	outputFormat string
	configFile   string
	verbose      bool

	// Analysis selection
	skipComplexity bool
	skipDeadCode   bool
	skipClones     bool

	// Quick filters
	minComplexity   int
	minSeverity     string
	cloneSimilarity float64
}

// NewAnalyzeCommand creates a new analyze command
func NewAnalyzeCommand() *AnalyzeCommand {
	return &AnalyzeCommand{
		outputFormat:    "text",
		configFile:      "",
		verbose:         false,
		skipComplexity:  false,
		skipDeadCode:    false,
		skipClones:      false,
		minComplexity:   5,
		minSeverity:     "warning",
		cloneSimilarity: 0.8,
	}
}

// NewAnalysisResult creates a new analysis result tracker
func NewAnalysisResult() *AnalysisResult {
	return &AnalysisResult{
		Complexity: &AnalysisStatus{Name: "Complexity Analysis", Enabled: false},
		DeadCode:   &AnalysisStatus{Name: "Dead Code Detection", Enabled: false},
		Clones:     &AnalysisStatus{Name: "Clone Detection", Enabled: false},
	}
}

// CreateCobraCommand creates the cobra command for comprehensive analysis
func (c *AnalyzeCommand) CreateCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze [files...]",
		Short: "Run comprehensive analysis on Python files",
		Long: `Run comprehensive analysis including complexity, dead code detection, and clone detection.

This command performs all available static analyses on Python code:
• Cyclomatic complexity analysis
• Dead code detection using CFG analysis
• Code clone detection using APTED algorithm

The analyses run concurrently for optimal performance. Results are combined
and presented in a unified format.

Examples:
  # Analyze current directory
  pyqol analyze .

  # Analyze specific files with JSON output
  pyqol analyze --format json src/myfile.py

  # Skip clone detection, focus on complexity and dead code
  pyqol analyze --skip-clones src/

  # Quick analysis with higher thresholds
  pyqol analyze --min-complexity 10 --min-severity critical src/`,
		Args: cobra.MinimumNArgs(1),
		RunE: c.runAnalyze,
	}

	// Output flags
	cmd.Flags().StringVarP(&c.outputFormat, "format", "f", "text", "Output format (text, json, yaml, csv)")
	cmd.Flags().StringVarP(&c.configFile, "config", "c", "", "Configuration file path")

	// Analysis selection flags
	cmd.Flags().BoolVar(&c.skipComplexity, "skip-complexity", false, "Skip complexity analysis")
	cmd.Flags().BoolVar(&c.skipDeadCode, "skip-deadcode", false, "Skip dead code detection")
	cmd.Flags().BoolVar(&c.skipClones, "skip-clones", false, "Skip clone detection")

	// Quick filter flags
	cmd.Flags().IntVar(&c.minComplexity, "min-complexity", 5, "Minimum complexity to report")
	cmd.Flags().StringVar(&c.minSeverity, "min-severity", "warning", "Minimum dead code severity (critical, warning, info)")
	cmd.Flags().Float64Var(&c.cloneSimilarity, "clone-threshold", 0.8, "Minimum similarity for clone detection (0.0-1.0)")

	return cmd
}

// runAnalyze executes the comprehensive analysis
func (c *AnalyzeCommand) runAnalyze(cmd *cobra.Command, args []string) error {
	// Get verbose flag from parent command
	if cmd.Parent() != nil {
		c.verbose, _ = cmd.Parent().Flags().GetBool("verbose")
	}

	// Initialize analysis result tracking
	result := NewAnalysisResult()
	result.Overall.StartTime = time.Now()

	// Configure which analyses to run
	result.Complexity.Enabled = !c.skipComplexity
	result.DeadCode.Enabled = !c.skipDeadCode
	result.Clones.Enabled = !c.skipClones

	// Print analysis plan
	c.printAnalysisPlan(cmd, result)

	// Run analyses concurrently with status tracking
	var wg sync.WaitGroup
	statusMutex := &sync.Mutex{}

	// Start real-time status monitoring if verbose
	var statusDone chan bool
	if c.verbose {
		statusDone = make(chan bool)
		go c.monitorAnalysisProgress(cmd, result, statusMutex, statusDone)
	}

	// Run complexity analysis
	if result.Complexity.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.runAnalysisWithStatus(result.Complexity, statusMutex, func() error {
				return c.runComplexityAnalysis(cmd, args)
			})
		}()
	}

	// Run dead code analysis
	if result.DeadCode.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.runAnalysisWithStatus(result.DeadCode, statusMutex, func() error {
				return c.runDeadCodeAnalysis(cmd, args)
			})
		}()
	}

	// Run clone analysis
	if result.Clones.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c.runAnalysisWithStatus(result.Clones, statusMutex, func() error {
				return c.runCloneAnalysis(cmd, args)
			})
		}()
	}

	// Wait for all analyses to complete
	wg.Wait()

	// Stop status monitoring
	if c.verbose {
		close(statusDone)
	}
	result.Overall.EndTime = time.Now()
	result.Overall.TotalTime = result.Overall.EndTime.Sub(result.Overall.StartTime)

	// Calculate overall statistics
	c.calculateOverallStats(result)

	// Print comprehensive status report
	c.printStatusReport(cmd, result)

	// Determine exit status based on results
	if result.Overall.FailureCount > 0 {
		return fmt.Errorf("analysis completed with %d failure(s) out of %d enabled analysis type(s)",
			result.Overall.FailureCount, result.Overall.SuccessCount+result.Overall.FailureCount)
	}

	return nil
}

// ErrorCategory represents the category of an analysis error
type ErrorCategory string

const (
	ErrorCategoryInput      ErrorCategory = "Input Error"
	ErrorCategoryConfig     ErrorCategory = "Configuration Error"
	ErrorCategoryProcessing ErrorCategory = "Processing Error"
	ErrorCategoryOutput     ErrorCategory = "Output Error"
	ErrorCategoryTimeout    ErrorCategory = "Timeout Error"
	ErrorCategoryUnknown    ErrorCategory = "Unknown Error"
)

// CategorizedError wraps an error with category information
type CategorizedError struct {
	Category ErrorCategory
	Message  string
	Original error
}

func (e *CategorizedError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Category, e.Message)
}

// categorizeError attempts to categorize an error based on its content
func categorizeError(err error) *CategorizedError {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Check for common error patterns
	if containsAny(errMsg, []string{"invalid input", "no files found", "path", "directory"}) {
		return &CategorizedError{
			Category: ErrorCategoryInput,
			Message:  "Failed to process input files or directories",
			Original: err,
		}
	}
	if containsAny(errMsg, []string{"config", "configuration", "invalid format"}) {
		return &CategorizedError{
			Category: ErrorCategoryConfig,
			Message:  "Configuration file or settings error",
			Original: err,
		}
	}
	if containsAny(errMsg, []string{"timeout", "deadline", "context canceled"}) {
		return &CategorizedError{
			Category: ErrorCategoryTimeout,
			Message:  "Analysis timed out",
			Original: err,
		}
	}
	if containsAny(errMsg, []string{"write", "output", "format"}) {
		return &CategorizedError{
			Category: ErrorCategoryOutput,
			Message:  "Failed to generate or write output",
			Original: err,
		}
	}
	if containsAny(errMsg, []string{"parse", "syntax", "analysis", "process"}) {
		return &CategorizedError{
			Category: ErrorCategoryProcessing,
			Message:  "Error during code analysis processing",
			Original: err,
		}
	}

	return &CategorizedError{
		Category: ErrorCategoryUnknown,
		Message:  errMsg,
		Original: err,
	}
}

// runAnalysisWithStatus wraps an analysis function with status tracking and error categorization
func (c *AnalyzeCommand) runAnalysisWithStatus(status *AnalysisStatus, mutex *sync.Mutex, analysisFunc func() error) {
	mutex.Lock()
	status.Started = true
	status.StartTime = time.Now()
	mutex.Unlock()

	err := analysisFunc()

	mutex.Lock()
	status.Completed = true
	status.Duration = time.Since(status.StartTime)
	if err != nil {
		status.Success = false
		status.Error = categorizeError(err)
	} else {
		status.Success = true
	}
	mutex.Unlock()
}

// printAnalysisPlan prints the planned analyses
func (c *AnalyzeCommand) printAnalysisPlan(cmd *cobra.Command, result *AnalysisResult) {
	if c.verbose {
		fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Analysis Plan:\n")
		if result.Complexity.Enabled {
			fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ Complexity Analysis (threshold: >%d)\n", c.minComplexity)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "  ⏭ Complexity Analysis (skipped)\n")
		}
		if result.DeadCode.Enabled {
			fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ Dead Code Detection (minimum severity: %s)\n", c.minSeverity)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "  ⏭ Dead Code Detection (skipped)\n")
		}
		if result.Clones.Enabled {
			fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ Clone Detection (similarity: %.1f)\n", c.cloneSimilarity)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "  ⏭ Clone Detection (skipped)\n")
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "\n")
	}
}

// calculateOverallStats calculates overall statistics from individual analysis statuses
func (c *AnalyzeCommand) calculateOverallStats(result *AnalysisResult) {
	analyses := []*AnalysisStatus{result.Complexity, result.DeadCode, result.Clones}

	for _, analysis := range analyses {
		if analysis.Enabled {
			if analysis.Success {
				result.Overall.SuccessCount++
			} else {
				result.Overall.FailureCount++
			}
		} else {
			result.Overall.SkippedCount++
		}
	}
}

// printStatusReport prints a comprehensive status report
func (c *AnalyzeCommand) printStatusReport(cmd *cobra.Command, result *AnalysisResult) {
	fmt.Fprintf(cmd.ErrOrStderr(), "\n📊 Analysis Summary:\n")
	fmt.Fprintf(cmd.ErrOrStderr(), "Total time: %v\n", result.Overall.TotalTime.Round(time.Millisecond))
	fmt.Fprintf(cmd.ErrOrStderr(), "\n")

	// Print individual analysis status
	analyses := []*AnalysisStatus{result.Complexity, result.DeadCode, result.Clones}
	for _, analysis := range analyses {
		if analysis.Enabled {
			if analysis.Success {
				fmt.Fprintf(cmd.ErrOrStderr(), "✅ %s: SUCCESS (%v)\n",
					analysis.Name, analysis.Duration.Round(time.Millisecond))
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "❌ %s: FAILED (%v)\n",
					analysis.Name, analysis.Duration.Round(time.Millisecond))
				if c.verbose && analysis.Error != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "   Error: %v\n", analysis.Error)
				}
			}
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "⏭ %s: SKIPPED\n", analysis.Name)
		}
	}

	// Print overall statistics
	fmt.Fprintf(cmd.ErrOrStderr(), "\n")
	if result.Overall.FailureCount > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Result: %d succeeded, %d failed, %d skipped\n",
			result.Overall.SuccessCount, result.Overall.FailureCount, result.Overall.SkippedCount)

		// Print categorized error summary
		errorCategories := make(map[ErrorCategory][]string)
		for _, analysis := range analyses {
			if analysis.Enabled && !analysis.Success && analysis.Error != nil {
				if categorizedErr, ok := analysis.Error.(*CategorizedError); ok {
					errorCategories[categorizedErr.Category] = append(
						errorCategories[categorizedErr.Category],
						fmt.Sprintf("%s: %s", analysis.Name, categorizedErr.Message))
				} else {
					errorCategories[ErrorCategoryUnknown] = append(
						errorCategories[ErrorCategoryUnknown],
						fmt.Sprintf("%s: %v", analysis.Name, analysis.Error))
				}
			}
		}

		// Print errors grouped by category
		fmt.Fprintf(cmd.ErrOrStderr(), "\nError Summary by Category:\n")
		for category, errors := range errorCategories {
			fmt.Fprintf(cmd.ErrOrStderr(), "  🔴 %s (%d):\n", category, len(errors))
			for _, err := range errors {
				fmt.Fprintf(cmd.ErrOrStderr(), "    • %s\n", err)
			}
		}

		// Print recovery suggestions
		c.printRecoverySuggestions(cmd, errorCategories)

	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "✅ Result: All %d enabled analysis type(s) completed successfully\n",
			result.Overall.SuccessCount)
	}
}

// printRecoverySuggestions prints helpful suggestions based on error categories
func (c *AnalyzeCommand) printRecoverySuggestions(cmd *cobra.Command, errorCategories map[ErrorCategory][]string) {
	if len(errorCategories) == 0 {
		return
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "\n💡 Recovery Suggestions:\n")

	for category := range errorCategories {
		switch category {
		case ErrorCategoryInput:
			fmt.Fprintf(cmd.ErrOrStderr(), "  📁 Input Issues: Check that files/directories exist and contain Python files\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "     Try: pyqol analyze . --verbose to see detailed file discovery\n")
		case ErrorCategoryConfig:
			fmt.Fprintf(cmd.ErrOrStderr(), "  ⚙️  Config Issues: Verify configuration file format and values\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "     Try: pyqol init to generate a valid config file\n")
		case ErrorCategoryTimeout:
			fmt.Fprintf(cmd.ErrOrStderr(), "  ⏰ Timeout Issues: Consider analyzing smaller file sets or increasing timeout\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "     Try: Analyze specific files instead of entire directories\n")
		case ErrorCategoryOutput:
			fmt.Fprintf(cmd.ErrOrStderr(), "  📤 Output Issues: Check write permissions and output format validity\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "     Try: Use --format text or check file system permissions\n")
		case ErrorCategoryProcessing:
			fmt.Fprintf(cmd.ErrOrStderr(), "  🔧 Processing Issues: Some files may have syntax errors or be corrupted\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "     Try: Run individual analysis types to isolate the problem\n")
		case ErrorCategoryUnknown:
			fmt.Fprintf(cmd.ErrOrStderr(), "  ❓ Unknown Issues: Run with --verbose for detailed error information\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "     Try: pyqol analyze . --verbose or check GitHub issues\n")
		}
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "\n  📖 For more help: pyqol --help or visit the documentation\n")
}

// monitorAnalysisProgress provides real-time status updates during analysis
func (c *AnalyzeCommand) monitorAnalysisProgress(cmd *cobra.Command, result *AnalysisResult, mutex *sync.Mutex, done chan bool) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			mutex.Lock()
			analyses := []*AnalysisStatus{result.Complexity, result.DeadCode, result.Clones}

			var running []string

			for _, analysis := range analyses {
				if analysis.Enabled {
					if analysis.Started && !analysis.Completed {
						elapsed := time.Since(analysis.StartTime)
						running = append(running, fmt.Sprintf("%s (%v)", analysis.Name, elapsed.Round(time.Second)))
					}
				}
			}
			mutex.Unlock()

			if len(running) > 0 {
				fmt.Fprintf(cmd.ErrOrStderr(), "🔄 Running: %s\n", joinStrings(running, ", "))
			}
		}
	}
}

// joinStrings joins a slice of strings with a delimiter
func joinStrings(strs []string, delimiter string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += delimiter + strs[i]
	}
	return result
}

// containsAny checks if a string contains any of the given substrings
func containsAny(str string, substrings []string) bool {
	for _, substr := range substrings {
		if len(str) >= len(substr) {
			for i := 0; i <= len(str)-len(substr); i++ {
				if str[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// runComplexityAnalysis runs complexity analysis with configured parameters
func (c *AnalyzeCommand) runComplexityAnalysis(cmd *cobra.Command, args []string) error {
	complexityCmd := NewComplexityCommand()

	// Configure complexity command with analyze parameters
	complexityCmd.outputFormat = c.outputFormat
	complexityCmd.minComplexity = c.minComplexity
	complexityCmd.configFile = c.configFile
	complexityCmd.verbose = c.verbose

	// Build complexity request
	request, err := complexityCmd.buildComplexityRequest(cmd, args)
	if err != nil {
		return err
	}

	// Create use case
	useCase, err := complexityCmd.createComplexityUseCase(cmd)
	if err != nil {
		return err
	}

	// Execute analysis
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	return useCase.Execute(ctx, request)
}

// runDeadCodeAnalysis runs dead code analysis with configured parameters
func (c *AnalyzeCommand) runDeadCodeAnalysis(cmd *cobra.Command, args []string) error {
	deadCodeCmd := NewDeadCodeCommand()

	// Configure dead code command with analyze parameters
	deadCodeCmd.outputFormat = c.outputFormat
	deadCodeCmd.minSeverity = c.minSeverity
	deadCodeCmd.configFile = c.configFile
	deadCodeCmd.verbose = c.verbose

	// Build dead code request
	request, err := deadCodeCmd.buildDeadCodeRequest(cmd, args)
	if err != nil {
		return err
	}

	// Create use case
	useCase, err := deadCodeCmd.createDeadCodeUseCase(cmd)
	if err != nil {
		return err
	}

	// Execute analysis
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	return useCase.Execute(ctx, request)
}

// runCloneAnalysis runs clone detection with configured parameters
func (c *AnalyzeCommand) runCloneAnalysis(cmd *cobra.Command, args []string) error {
	cloneCmd := NewCloneCommand()

	// Configure clone command with analyze parameters
	cloneCmd.outputFormat = c.outputFormat
	cloneCmd.similarityThreshold = c.cloneSimilarity
	cloneCmd.configFile = c.configFile
	cloneCmd.verbose = c.verbose

	// Create clone request
	request, err := cloneCmd.createCloneRequest(args)
	if err != nil {
		return err
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return fmt.Errorf("invalid clone request: %w", err)
	}

	// Create use case
	useCase, err := cloneCmd.createCloneUseCase(cmd)
	if err != nil {
		return err
	}

	// Execute analysis
	ctx := context.Background()
	return useCase.Execute(ctx, *request)
}

// NewAnalyzeCmd creates and returns the analyze cobra command
func NewAnalyzeCmd() *cobra.Command {
	analyzeCommand := NewAnalyzeCommand()
	return analyzeCommand.CreateCobraCommand()
}
