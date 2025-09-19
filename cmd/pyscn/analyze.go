package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/version"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// AnalysisStatus represents the status of an individual analysis
type AnalysisStatus struct {
	Name              string
	Enabled           bool
	Started           bool
	Completed         bool
	Success           bool
	Error             error
	Duration          time.Duration
	StartTime         time.Time
	EstimatedDuration time.Duration
	ProcessedFiles    int
	TotalFiles        int
	ProgressBar       *progressbar.ProgressBar
}

// AnalysisResult aggregates all analysis results
type AnalysisResult struct {
	Complexity *AnalysisStatus
	DeadCode   *AnalysisStatus
	Clones     *AnalysisStatus
	CBO        *AnalysisStatus
	System     *AnalysisStatus // System-level analysis
	Overall    struct {
		StartTime    time.Time
		EndTime      time.Time
		TotalTime    time.Duration
		SuccessCount int
		FailureCount int
		SkippedCount int
	}

	// Store actual analysis responses for unified report
	ComplexityResponse *domain.ComplexityResponse
	DeadCodeResponse   *domain.DeadCodeResponse
	CloneResponse      *domain.CloneResponse
	CBOResponse        *domain.CBOResponse
	SystemResponse     *domain.SystemAnalysisResponse // System analysis response
}

// AnalyzeCommand represents the comprehensive analysis command
type AnalyzeCommand struct {
	// Output format flags (only one should be true)
	html   bool
	json   bool
	csv    bool
	yaml   bool
	noOpen bool

	// Configuration
	configFile string
	verbose    bool

	// Analysis selection
	skipComplexity bool
	skipDeadCode   bool
	skipClones     bool
	skipCBO        bool
	skipSystem     bool     // Skip system-level analysis (deps + architecture)
	selectAnalyses []string // Only run specified analyses

	// Quick filters
	minComplexity   int
	minSeverity     string
	cloneSimilarity float64
	minCBO          int

	// System analysis options
	enableSystemAnalysis bool // Deprecated: not used; use skipSystem instead
	detectCycles         bool // Detect circular dependencies
	validateArch         bool // Validate architecture rules
}

// NewAnalyzeCommand creates a new analyze command
func NewAnalyzeCommand() *AnalyzeCommand {
	return &AnalyzeCommand{
		html:                 false,
		json:                 false,
		csv:                  false,
		yaml:                 false,
		noOpen:               false,
		configFile:           "",
		verbose:              false,
		skipComplexity:       false,
		skipDeadCode:         false,
		skipClones:           false,
		skipCBO:              false,
		skipSystem:           false,
		minComplexity:        5,
		minSeverity:          "warning",
		cloneSimilarity:      0.8,
		minCBO:               0,
		enableSystemAnalysis: true, // Deprecated: keep true; actual control is skipSystem
		detectCycles:         true,
		validateArch:         true,
	}
}

// NewAnalysisResult creates a new analysis result tracker
func NewAnalysisResult() *AnalysisResult {
	return &AnalysisResult{
		Complexity: &AnalysisStatus{Name: "Complexity Analysis", Enabled: false},
		DeadCode:   &AnalysisStatus{Name: "Dead Code Detection", Enabled: false},
		Clones:     &AnalysisStatus{Name: "Clone Detection", Enabled: false},
		CBO:        &AnalysisStatus{Name: "Class Coupling (CBO)", Enabled: false},
		System:     &AnalysisStatus{Name: "System Analysis", Enabled: false},
	}
}

// CreateCobraCommand creates the cobra command for comprehensive analysis
func (c *AnalyzeCommand) CreateCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze [files...]",
		Short: "Run comprehensive analysis on Python files",
		Long: `Run comprehensive analysis including complexity, dead code detection, clone detection, and CBO analysis.

This command performs all available static analyses on Python code:
• Cyclomatic complexity analysis
• Dead code detection using CFG analysis
• Code clone detection using APTED algorithm
• Dependency analysis (class coupling)

The analyses run concurrently for optimal performance. Results are combined
and presented in a unified format.

Examples:
  # Analyze current directory
  pyscn analyze .

  # Analyze specific files with JSON output
  pyscn analyze --json src/myfile.py

  # Skip clone detection, focus on complexity, dead code, and dependencies
  pyscn analyze --skip-clones src/

  # Quick analysis with higher thresholds  
  pyscn analyze --min-complexity 10 --min-severity critical --min-cbo 5 src/

  # Skip dependency analysis
  pyscn analyze --skip-cbo src/`,
		Args: cobra.MinimumNArgs(1),
		RunE: c.runAnalyze,
	}

	// Output format flags
	cmd.Flags().BoolVar(&c.html, "html", false, "Generate HTML report file")
	cmd.Flags().BoolVar(&c.json, "json", false, "Generate JSON report file")
	cmd.Flags().BoolVar(&c.csv, "csv", false, "Generate CSV report file")
	cmd.Flags().BoolVar(&c.yaml, "yaml", false, "Generate YAML report file")
	cmd.Flags().BoolVar(&c.noOpen, "no-open", false, "Don't auto-open HTML in browser")
	cmd.Flags().StringVarP(&c.configFile, "config", "c", "", "Configuration file path")

	// Analysis selection flags
	cmd.Flags().BoolVar(&c.skipComplexity, "skip-complexity", false, "Skip complexity analysis")
	cmd.Flags().BoolVar(&c.skipDeadCode, "skip-deadcode", false, "Skip dead code detection")
	cmd.Flags().BoolVar(&c.skipClones, "skip-clones", false, "Skip clone detection")
	cmd.Flags().BoolVar(&c.skipCBO, "skip-cbo", false, "Skip class coupling (CBO) analysis")
	cmd.Flags().BoolVar(&c.skipSystem, "skip-deps", false, "Skip module dependencies and architecture analysis")
	cmd.Flags().StringSliceVar(&c.selectAnalyses, "select", []string{}, "Only run specified analyses (complexity,deadcode,clones,cbo,deps)")

	// Quick filter flags
	cmd.Flags().IntVar(&c.minComplexity, "min-complexity", 5, "Minimum complexity to report")
	cmd.Flags().StringVar(&c.minSeverity, "min-severity", "warning", "Minimum dead code severity (critical, warning, info)")
	cmd.Flags().Float64Var(&c.cloneSimilarity, "clone-threshold", 0.8, "Minimum similarity for clone detection (0.0-1.0)")
	cmd.Flags().IntVar(&c.minCBO, "min-cbo", 0, "Minimum CBO to report")

	return cmd
}

// determineOutputFormat determines the output format based on flags
func (c *AnalyzeCommand) determineOutputFormat() (string, string, error) {
	// Count how many format flags are set
	formatCount := 0
	var format string
	var extension string

	if c.html {
		formatCount++
		format = "html"
		extension = "html"
	}
	if c.json {
		formatCount++
		format = "json"
		extension = "json"
	}
	if c.csv {
		formatCount++
		format = "csv"
		extension = "csv"
	}
	if c.yaml {
		formatCount++
		format = "yaml"
		extension = "yaml"
	}

	// Check for conflicting flags
	if formatCount > 1 {
		return "", "", fmt.Errorf("only one output format flag can be specified")
	}

	// Default to HTML if no format specified (hybrid default: file + stderr summary)
	if formatCount == 0 {
		return "html", "html", nil
	}

	return format, extension, nil
}

// shouldGenerateUnifiedReport returns true if a unified report should be generated
func (c *AnalyzeCommand) shouldGenerateUnifiedReport() bool {
	// Always generate unified report in analyze mode.
	// Format is determined by flags or defaults to HTML.
	return true
}

// shouldUseProgressBars returns true when the session appears to be interactive
// and progress bars won't pollute machine-readable stdout consumers.
func (c *AnalyzeCommand) shouldUseProgressBars(cmd *cobra.Command) bool {
	if !isInteractiveEnvironment() {
		return false
	}

	if errWriter, ok := cmd.ErrOrStderr().(*os.File); ok {
		return term.IsTerminal(int(errWriter.Fd()))
	}

	return false
}

// initializeProgressTracking sets up progress tracking for each analysis
func (c *AnalyzeCommand) initializeProgressTracking(result *AnalysisResult, totalFiles int, writer io.Writer) {
	if writer == nil {
		writer = io.Discard
	}
	// Estimation coefficients (ms per file)
	complexityCoeff := 50.0 // Fast parsing and CFG analysis
	deadCodeCoeff := 50.0   // Similar to complexity
	cboCoeff := 30.0        // Simpler analysis
	systemCoeff := 40.0     // Module dependency analysis

	// Clone detection is O(n²) for file comparisons, so use quadratic estimation
	cloneCoeff := 100.0 + float64(totalFiles)*2.0 // Base cost + scaling factor

	if result.Complexity.Enabled {
		result.Complexity.TotalFiles = totalFiles
		result.Complexity.EstimatedDuration = time.Duration(float64(totalFiles)*complexityCoeff) * time.Millisecond
		result.Complexity.ProgressBar = c.createProgressBar("Complexity Analysis", totalFiles, writer)
	}

	if result.DeadCode.Enabled {
		result.DeadCode.TotalFiles = totalFiles
		result.DeadCode.EstimatedDuration = time.Duration(float64(totalFiles)*deadCodeCoeff) * time.Millisecond
		result.DeadCode.ProgressBar = c.createProgressBar("Dead Code Detection", totalFiles, writer)
	}

	if result.Clones.Enabled {
		result.Clones.TotalFiles = totalFiles
		result.Clones.EstimatedDuration = time.Duration(float64(totalFiles)*cloneCoeff) * time.Millisecond
		result.Clones.ProgressBar = c.createProgressBar("Clone Detection", totalFiles, writer)
	}

	if result.CBO.Enabled {
		result.CBO.TotalFiles = totalFiles
		result.CBO.EstimatedDuration = time.Duration(float64(totalFiles)*cboCoeff) * time.Millisecond
		result.CBO.ProgressBar = c.createProgressBar("Class Coupling (CBO)", totalFiles, writer)
	}

	if result.System.Enabled {
		result.System.TotalFiles = totalFiles
		result.System.EstimatedDuration = time.Duration(float64(totalFiles)*systemCoeff) * time.Millisecond
		result.System.ProgressBar = c.createProgressBar("System Analysis", totalFiles, writer)
	}
}

// createProgressBar creates a new progress bar with consistent styling
func (c *AnalyzeCommand) createProgressBar(description string, max int, writer io.Writer) *progressbar.ProgressBar {
	if writer == nil {
		writer = io.Discard
	}
	return progressbar.NewOptions(max,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetWriter(writer),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprintln(writer)
		}),
	)
}

// outputComplexityReport outputs an individual complexity report (for backward compatibility)
func (c *AnalyzeCommand) outputComplexityReport(cmd *cobra.Command, response *domain.ComplexityResponse) error {
	// This is a placeholder for backward compatibility
	// In unified mode, we don't generate individual reports
	return nil
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

	// Configure which analyses to run based on --select or skip flags
	if len(c.selectAnalyses) > 0 {
		// --select flag specified: only run the specified analyses
		result.Complexity.Enabled = c.containsAnalysis("complexity")
		result.DeadCode.Enabled = c.containsAnalysis("deadcode")
		result.Clones.Enabled = c.containsAnalysis("clones")
		result.CBO.Enabled = c.containsAnalysis("cbo")
		result.System.Enabled = c.containsAnalysis("deps")
	} else {
		// No --select flag: run all except those explicitly skipped
		result.Complexity.Enabled = !c.skipComplexity
		result.DeadCode.Enabled = !c.skipDeadCode
		result.Clones.Enabled = !c.skipClones
		result.CBO.Enabled = !c.skipCBO
		result.System.Enabled = !c.skipSystem
	}

	// Early validation: Check if there are any Python files to analyze
	fileReader := service.NewFileReader()
	pythonFiles, err := fileReader.CollectPythonFiles(
		args,
		true, // recursive
		[]string{"*.py", "*.pyi"},
		[]string{"test_*.py", "*_test.py"},
	)
	if err != nil {
		return fmt.Errorf("failed to collect Python files: %w", err)
	}

	if len(pythonFiles) == 0 {
		// Provide helpful error message with suggestions
		fmt.Fprintf(cmd.ErrOrStderr(), "❌ No Python files found in the specified paths\n\n")
		fmt.Fprintf(cmd.ErrOrStderr(), "Searched in:\n")
		for _, arg := range args {
			fmt.Fprintf(cmd.ErrOrStderr(), "  • %s\n", arg)
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "\n💡 Suggestions:\n")
		fmt.Fprintf(cmd.ErrOrStderr(), "  • Check that the path exists and contains Python files (*.py, *.pyi)\n")
		fmt.Fprintf(cmd.ErrOrStderr(), "  • Try running from a directory containing Python code\n")
		fmt.Fprintf(cmd.ErrOrStderr(), "  • Use 'pyscn analyze .' to analyze the current directory\n")
		fmt.Fprintf(cmd.ErrOrStderr(), "  • Specify a valid Python file or directory path\n")
		return fmt.Errorf("no Python files found to analyze")
	}

	// Log file discovery if verbose
	if c.verbose {
		fmt.Fprintf(cmd.ErrOrStderr(), "📁 Found %d Python file(s) to analyze\n", len(pythonFiles))
	}

	// Decide whether interactive progress bars should be rendered
	enableProgressBars := c.shouldUseProgressBars(cmd)

	// Initialize progress tracking for enabled analyses when interactive
	if enableProgressBars {
		c.initializeProgressTracking(result, len(pythonFiles), cmd.ErrOrStderr())
	} else if c.verbose {
		c.printAnalysisPlan(cmd, result)
	}

	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Run analyses concurrently with status tracking
	var wg sync.WaitGroup
	statusMutex := &sync.Mutex{}

	// Start real-time status monitoring for interactive or verbose sessions
	var statusDone chan bool
	if enableProgressBars || c.verbose {
		statusDone = make(chan bool)
		go c.monitorAnalysisProgress(cmd, result, statusMutex, statusDone, enableProgressBars)
	}

	// Ensure status monitoring is stopped on exit
	defer func() {
		if statusDone != nil {
			select {
			case <-statusDone:
				// Already closed
			default:
				close(statusDone)
			}
		}
	}()

	// Run complexity analysis
	if result.Complexity.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				result.Complexity.Error = fmt.Errorf("analysis timed out")
				return
			default:
				c.runAnalysisWithStatus(result.Complexity, statusMutex, func() error {
					if c.shouldGenerateUnifiedReport() {
						response, err := c.runComplexityAnalysisWithResult(cmd, args)
						if err == nil {
							result.ComplexityResponse = response
						}
						return err
					}
					return c.runComplexityAnalysis(cmd, args)
				})
			}
		}()
	}

	// Run dead code analysis
	if result.DeadCode.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				result.DeadCode.Error = fmt.Errorf("analysis timed out")
				return
			default:
				c.runAnalysisWithStatus(result.DeadCode, statusMutex, func() error {
					if c.shouldGenerateUnifiedReport() {
						response, err := c.runDeadCodeAnalysisWithResult(cmd, args)
						if err == nil {
							result.DeadCodeResponse = response
						}
						return err
					}
					return c.runDeadCodeAnalysis(cmd, args)
				})
			}
		}()
	}

	// Run clone analysis
	if result.Clones.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				result.Clones.Error = fmt.Errorf("analysis timed out")
				return
			default:
				c.runAnalysisWithStatus(result.Clones, statusMutex, func() error {
					if c.shouldGenerateUnifiedReport() {
						response, err := c.runCloneAnalysisWithResult(cmd, args)
						if err == nil {
							result.CloneResponse = response
						}
						return err
					}
					return c.runCloneAnalysis(cmd, args)
				})
			}
		}()
	}

	// Run CBO analysis
	if result.CBO.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				result.CBO.Error = fmt.Errorf("analysis timed out")
				return
			default:
				c.runAnalysisWithStatus(result.CBO, statusMutex, func() error {
					if c.shouldGenerateUnifiedReport() {
						response, err := c.runCBOAnalysisWithResult(cmd, args)
						if err == nil {
							result.CBOResponse = response
						}
						return err
					}
					return c.runCBOAnalysis(cmd, args)
				})
			}
		}()
	}

	// Run System (deps + architecture) analysis
	if result.System.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ctx.Done():
				result.System.Error = fmt.Errorf("analysis timed out")
				return
			default:
				c.runAnalysisWithStatus(result.System, statusMutex, func() error {
					response, err := c.runSystemAnalysisWithResult(cmd, args)
					if err == nil {
						result.SystemResponse = response
					}
					return err
				})
			}
		}()
	}

	// Wait for all analyses to complete or timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All analyses completed
	case <-ctx.Done():
		// Timeout occurred
		fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Analysis timed out after 10 minutes\n")
	}

	// Stop status monitoring
	if statusDone != nil {
		select {
		case <-statusDone:
			// Already closed
		default:
			close(statusDone)
		}
	}
	result.Overall.EndTime = time.Now()
	result.Overall.TotalTime = result.Overall.EndTime.Sub(result.Overall.StartTime)

	// Calculate overall statistics
	c.calculateOverallStats(result)

	// Generate unified report if requested
	if c.shouldGenerateUnifiedReport() {
		if err := c.generateUnifiedReport(cmd, result, args); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Failed to generate unified report: %v\n", err)
		}
	}

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
		// Mark progress bar as complete
		if status.ProgressBar != nil {
			_ = status.ProgressBar.Finish()
		}
	}
	mutex.Unlock()
}

// printAnalysisPlan prints the planned analyses when verbose mode is enabled.
func (c *AnalyzeCommand) printAnalysisPlan(cmd *cobra.Command, result *AnalysisResult) {
	if !c.verbose {
		return
	}

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
	if result.CBO.Enabled {
		fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ Class Coupling (CBO) (min CBO: %d)\n", c.minCBO)
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "  ⏭ Class Coupling (CBO) (skipped)\n")
	}
	if result.System.Enabled {
		fmt.Fprintf(cmd.ErrOrStderr(), "  ✓ System Analysis (deps + architecture)\n")
	} else {
		fmt.Fprintf(cmd.ErrOrStderr(), "  ⏭ System Analysis (skipped)\n")
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "\n")
}

// calculateOverallStats calculates overall statistics from individual analysis statuses
func (c *AnalyzeCommand) calculateOverallStats(result *AnalysisResult) {
	analyses := []*AnalysisStatus{result.Complexity, result.DeadCode, result.Clones, result.CBO, result.System}

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
	analyses := []*AnalysisStatus{result.Complexity, result.DeadCode, result.Clones, result.CBO, result.System}
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
			fmt.Fprintf(cmd.ErrOrStderr(), "     Try: pyscn analyze . --verbose to see detailed file discovery\n")
		case ErrorCategoryConfig:
			fmt.Fprintf(cmd.ErrOrStderr(), "  ⚙️  Config Issues: Verify configuration file format and values\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "     Try: pyscn init to generate a valid config file\n")
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
			fmt.Fprintf(cmd.ErrOrStderr(), "     Try: pyscn analyze . --verbose or check GitHub issues\n")
		}
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "\n  📖 For more help: pyscn --help or visit the documentation\n")
}

// monitorAnalysisProgress provides real-time status updates during analysis
func (c *AnalyzeCommand) monitorAnalysisProgress(cmd *cobra.Command, result *AnalysisResult, mutex *sync.Mutex, done chan bool, useProgressBars bool) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Add a timeout to prevent infinite blocking
	timeout := time.NewTimer(5 * time.Minute)
	defer timeout.Stop()

	for {
		select {
		case <-done:
			return
		case <-timeout.C:
			// Timeout after 5 minutes to prevent infinite blocking
			fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Progress monitoring timed out after 5 minutes\n")
			return
		case <-ticker.C:
			if useProgressBars {
				mutex.Lock()
				c.updateProgressBars(result)
				mutex.Unlock()
			} else if c.verbose {
				mutex.Lock()
				running := c.collectRunningAnalyses(result)
				mutex.Unlock()
				if len(running) > 0 {
					fmt.Fprintf(cmd.ErrOrStderr(), "🔄 Running: %s\n", joinStrings(running, ", "))
				}
			}
		}
	}
}

// updateProgressBars updates the progress bars based on elapsed time
func (c *AnalyzeCommand) updateProgressBars(result *AnalysisResult) {
	analyses := []*AnalysisStatus{result.Complexity, result.DeadCode, result.Clones, result.CBO, result.System}

	for _, analysis := range analyses {
		if analysis.Enabled && analysis.Started && !analysis.Completed && analysis.ProgressBar != nil {
			elapsed := time.Since(analysis.StartTime)
			if analysis.EstimatedDuration > 0 {
				// Calculate progress based on elapsed time vs estimated duration
				progress := float64(elapsed) / float64(analysis.EstimatedDuration)
				if progress > 1.0 {
					progress = 1.0
				}

				// Set the progress bar to the calculated progress
				targetValue := int(progress * float64(analysis.TotalFiles))

				// Only update if we haven't exceeded the max
				if targetValue <= analysis.TotalFiles {
					_ = analysis.ProgressBar.Set(targetValue)
				}
			}
		}
	}
}

func (c *AnalyzeCommand) collectRunningAnalyses(result *AnalysisResult) []string {
	analyses := []*AnalysisStatus{result.Complexity, result.DeadCode, result.Clones, result.CBO, result.System}
	var running []string
	for _, analysis := range analyses {
		if analysis.Enabled && analysis.Started && !analysis.Completed {
			elapsed := time.Since(analysis.StartTime)
			running = append(running, fmt.Sprintf("%s (%v)", analysis.Name, elapsed.Round(time.Second)))
		}
	}
	return running
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

// isInteractiveEnvironment returns true if the environment appears to be
// an interactive TTY session (and not CI), used to decide auto-open behavior.
func isInteractiveEnvironment() bool {
	if os.Getenv("CI") != "" {
		return false
	}
	// Best-effort TTY detection without external deps
	if fi, err := os.Stderr.Stat(); err == nil {
		return (fi.Mode() & os.ModeCharDevice) != 0
	}
	return false
}

// runComplexityAnalysis runs complexity analysis with configured parameters
func (c *AnalyzeCommand) runComplexityAnalysis(cmd *cobra.Command, args []string) error {
	response, err := c.runComplexityAnalysisWithResult(cmd, args)
	if err != nil {
		return err
	}

	// For backward compatibility, still generate individual file if format is specified
	// and we're not generating a unified report
	if !c.shouldGenerateUnifiedReport() {
		return c.outputComplexityReport(cmd, response)
	}

	return nil
}

// runComplexityAnalysisWithResult runs complexity analysis and returns the result
func (c *AnalyzeCommand) runComplexityAnalysisWithResult(cmd *cobra.Command, args []string) (*domain.ComplexityResponse, error) {
	complexityCmd := NewComplexityCommand()

	// Configure complexity command with analyze parameters
	complexityCmd.minComplexity = c.minComplexity
	complexityCmd.configFile = c.configFile
	complexityCmd.verbose = false // Disable verbose output for unified analysis

	// Build complexity request - bypass validation since paths will be validated by use case
	request := domain.ComplexityRequest{
		Paths:           args,
		OutputFormat:    domain.OutputFormatText,
		OutputWriter:    os.Stdout, // Provide a dummy writer
		MinComplexity:   complexityCmd.minComplexity,
		MaxComplexity:   complexityCmd.maxComplexity,
		LowThreshold:    complexityCmd.lowThreshold,
		MediumThreshold: complexityCmd.mediumThreshold,
		SortBy:          domain.SortByComplexity, // Default sort
		ConfigPath:      complexityCmd.configFile,
		Recursive:       true, // Always recursive for directories
		IncludePatterns: []string{"*.py", "*.pyi"},
		ExcludePatterns: []string{"test_*.py", "*_test.py"},
	}

	// Create use case
	useCase, err := complexityCmd.createComplexityUseCase(cmd)
	if err != nil {
		return nil, err
	}

	// Execute analysis and return result
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	return useCase.AnalyzeAndReturn(ctx, request)
}

// runDeadCodeAnalysis runs dead code analysis with configured parameters
func (c *AnalyzeCommand) runDeadCodeAnalysis(cmd *cobra.Command, args []string) error {
	response, err := c.runDeadCodeAnalysisWithResult(cmd, args)
	if err != nil {
		return err
	}

	// For backward compatibility
	if !c.shouldGenerateUnifiedReport() {
		return c.outputDeadCodeReport(cmd, response)
	}

	return nil
}

// runDeadCodeAnalysisWithResult runs dead code analysis and returns the result
func (c *AnalyzeCommand) runDeadCodeAnalysisWithResult(cmd *cobra.Command, args []string) (*domain.DeadCodeResponse, error) {
	deadCodeCmd := NewDeadCodeCommand()

	// Configure dead code command with analyze parameters
	deadCodeCmd.minSeverity = c.minSeverity
	deadCodeCmd.configFile = c.configFile
	deadCodeCmd.verbose = false // Disable verbose output for unified analysis

	// Parse severity
	minSeverity := domain.DeadCodeSeverityWarning
	switch c.minSeverity {
	case "critical":
		minSeverity = domain.DeadCodeSeverityCritical
	case "warning":
		minSeverity = domain.DeadCodeSeverityWarning
	case "info":
		minSeverity = domain.DeadCodeSeverityInfo
	}

	// Build dead code request - bypass validation since paths will be validated by use case
	request := domain.DeadCodeRequest{
		Paths:           args,
		OutputFormat:    domain.OutputFormatText,
		OutputWriter:    os.Stdout, // Provide a dummy writer
		MinSeverity:     minSeverity,
		ShowContext:     deadCodeCmd.showContext,
		ContextLines:    3,
		SortBy:          domain.DeadCodeSortBySeverity, // Default sort
		ConfigPath:      deadCodeCmd.configFile,
		Recursive:       true, // Always recursive for directories
		IncludePatterns: []string{"*.py", "*.pyi"},
		ExcludePatterns: []string{"test_*.py", "*_test.py"},
	}

	// Create use case
	useCase, err := deadCodeCmd.createDeadCodeUseCase(cmd)
	if err != nil {
		return nil, err
	}

	// Execute analysis and return result
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	return useCase.AnalyzeAndReturn(ctx, request)
}

// outputDeadCodeReport outputs an individual dead code report (for backward compatibility)
func (c *AnalyzeCommand) outputDeadCodeReport(cmd *cobra.Command, response *domain.DeadCodeResponse) error {
	// Placeholder for backward compatibility
	return nil
}

// runCloneAnalysis runs clone detection with configured parameters
func (c *AnalyzeCommand) runCloneAnalysis(cmd *cobra.Command, args []string) error {
	response, err := c.runCloneAnalysisWithResult(cmd, args)
	if err != nil {
		return err
	}

	// For backward compatibility
	if !c.shouldGenerateUnifiedReport() {
		return c.outputCloneReport(cmd, response)
	}

	return nil
}

// runCloneAnalysisWithResult runs clone detection and returns the result
func (c *AnalyzeCommand) runCloneAnalysisWithResult(cmd *cobra.Command, args []string) (*domain.CloneResponse, error) {
	cloneCmd := NewCloneCommand()

	// Configure clone command with analyze parameters
	cloneCmd.similarityThreshold = c.cloneSimilarity
	cloneCmd.configFile = c.configFile
	cloneCmd.verbose = false // Disable verbose output for unified analysis

	// Pre-calculate file count for progress reporting (not used in concurrent mode)
	// In analyze mode, we disable progress to avoid conflicts with concurrent analysis

	// Create clone request without output settings
	request, err := cloneCmd.createCloneRequest(cmd, args)
	if err != nil {
		return nil, err
	}

	// Clear output settings to prevent individual file generation
	request.OutputPath = ""
	request.OutputWriter = nil

	// Validate request
	if err := request.Validate(); err != nil {
		return nil, fmt.Errorf("invalid clone request: %w", err)
	}

	// Create use case (progress reporting disabled for concurrent analysis)
	useCase, err := cloneCmd.createCloneUseCase(cmd)
	if err != nil {
		return nil, err
	}

	// Execute analysis and return result
	ctx := context.Background()
	return useCase.ExecuteAndReturn(ctx, *request)
}

// outputCloneReport outputs an individual clone report (for backward compatibility)
func (c *AnalyzeCommand) outputCloneReport(cmd *cobra.Command, response *domain.CloneResponse) error {
	// Placeholder for backward compatibility
	return nil
}

// runSystemAnalysisWithResult runs module dependency + architecture analysis and returns the result
func (c *AnalyzeCommand) runSystemAnalysisWithResult(cmd *cobra.Command, args []string) (*domain.SystemAnalysisResponse, error) {
	// Build request
	request := domain.SystemAnalysisRequest{
		Paths:               args,
		OutputFormat:        domain.OutputFormatText,
		OutputWriter:        os.Stdout, // required by usecase validation
		AnalyzeDependencies: true,
		AnalyzeArchitecture: true,
		AnalyzeQuality:      true,

		// Options (lean defaults; config can override)
		IncludeStdLib:     false,
		IncludeThirdParty: true,
		FollowRelative:    true,
		DetectCycles:      true,

		// File selection
		ConfigPath:      c.configFile,
		Recursive:       true,
		IncludePatterns: []string{"*.py", "*.pyi"},
		ExcludePatterns: []string{"test_*.py", "*_test.py"},
	}

	// Build use case
	systemService := service.NewSystemAnalysisService()
	fileReader := service.NewFileReader()
	formatter := service.NewSystemAnalysisFormatter()
	configLoader := service.NewSystemAnalysisConfigurationLoader()

	uc, err := app.NewSystemAnalysisUseCaseBuilder().
		WithService(systemService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		WithConfigLoader(configLoader).
		Build()
	if err != nil {
		return nil, err
	}

	// Execute
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	return uc.AnalyzeAndReturn(ctx, request)
}

// generateUnifiedReport generates a unified report from all analysis results
func (c *AnalyzeCommand) generateUnifiedReport(cmd *cobra.Command, result *AnalysisResult, args []string) error {
	// Determine output format
	format, extension, err := c.determineOutputFormat()
	if err != nil {
		return err
	}

	// Generate filename with timestamp
	// Use first path as target for config discovery
	targetPath := getTargetPathFromArgs(args)
	filename, err := generateOutputFilePath("analyze", extension, targetPath)
	if err != nil {
		return fmt.Errorf("failed to generate output path: %w", err)
	}

	// Create the unified response
	response := &domain.AnalyzeResponse{
		Complexity:  result.ComplexityResponse,
		DeadCode:    result.DeadCodeResponse,
		Clone:       result.CloneResponse,
		CBO:         result.CBOResponse,
		System:      result.SystemResponse,
		GeneratedAt: time.Now(),
		Duration:    result.Overall.TotalTime.Milliseconds(),
		Version:     version.Version,
	}

	// Calculate summary statistics
	summary := &response.Summary

	// File statistics
	if result.ComplexityResponse != nil {
		summary.TotalFiles = result.ComplexityResponse.Summary.FilesAnalyzed
		summary.AnalyzedFiles = result.ComplexityResponse.Summary.FilesAnalyzed
		summary.ComplexityEnabled = true
		summary.TotalFunctions = len(result.ComplexityResponse.Functions)
		summary.AverageComplexity = result.ComplexityResponse.Summary.AverageComplexity

		// Count high complexity functions
		summary.HighComplexityCount = result.ComplexityResponse.Summary.HighRiskFunctions
	}

	// Dead code statistics
	if result.DeadCodeResponse != nil {
		summary.DeadCodeEnabled = true
		summary.DeadCodeCount = result.DeadCodeResponse.Summary.TotalFindings
		summary.CriticalDeadCode = result.DeadCodeResponse.Summary.CriticalFindings
	}

	// Clone statistics
	if result.CloneResponse != nil {
		summary.CloneEnabled = true
		summary.TotalClones = result.CloneResponse.Statistics.TotalClones
		summary.ClonePairs = result.CloneResponse.Statistics.TotalClonePairs
		summary.CloneGroups = result.CloneResponse.Statistics.TotalCloneGroups

		// Calculate code duplication percentage
		if result.CloneResponse.Statistics.LinesAnalyzed > 0 {
			// Track unique lines involved in clones to avoid double-counting
			duplicatedLineSet := make(map[string]map[int]bool)

			for _, pair := range result.CloneResponse.ClonePairs {
				// Add lines from Clone1
				if pair.Clone1 != nil && pair.Clone1.Location != nil {
					filePath := pair.Clone1.Location.FilePath
					if duplicatedLineSet[filePath] == nil {
						duplicatedLineSet[filePath] = make(map[int]bool)
					}
					for line := pair.Clone1.Location.StartLine; line <= pair.Clone1.Location.EndLine; line++ {
						duplicatedLineSet[filePath][line] = true
					}
				}
				// Add lines from Clone2
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

			summary.CodeDuplication = (float64(duplicatedLines) / float64(result.CloneResponse.Statistics.LinesAnalyzed)) * 100
		}
	}

	// CBO statistics
	if result.CBOResponse != nil {
		summary.CBOEnabled = true
		summary.CBOClasses = result.CBOResponse.Summary.TotalClasses
		summary.HighCouplingClasses = result.CBOResponse.Summary.HighRiskClasses
		summary.AverageCoupling = result.CBOResponse.Summary.AverageCBO
	}

	// System (deps + architecture) statistics for scoring
	if result.SystemResponse != nil {
		if result.SystemResponse.DependencyAnalysis != nil {
			da := result.SystemResponse.DependencyAnalysis
			summary.DepsEnabled = true
			summary.DepsTotalModules = da.TotalModules
			summary.DepsMaxDepth = da.MaxDepth
			if da.CircularDependencies != nil {
				summary.DepsModulesInCycles = da.CircularDependencies.TotalModulesInCycles
			}
			if da.CouplingAnalysis != nil {
				summary.DepsMainSequenceDeviation = da.CouplingAnalysis.MainSequenceDeviation
			}
		}
		if result.SystemResponse.ArchitectureAnalysis != nil {
			aa := result.SystemResponse.ArchitectureAnalysis
			summary.ArchEnabled = true
			summary.ArchCompliance = aa.ComplianceScore
		}
	}

	// Calculate health score
	summary.CalculateHealthScore()

	// Create formatter
	formatter := service.NewAnalyzeFormatter()

	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", filename, err)
	}
	defer file.Close()

	// Write the unified report
	formatType := domain.OutputFormat(format)
	if err := formatter.Write(response, formatType, file); err != nil {
		return fmt.Errorf("failed to write unified report: %w", err)
	}

	// Get absolute path for display
	absPath, err := filepath.Abs(filename)
	if err != nil {
		absPath = filename
	}

	// Handle browser opening for HTML
	if format == "html" {
		// Auto-open only when explicitly allowed and environment appears interactive
		if !c.noOpen && isInteractiveEnvironment() {
			fileURL := "file://" + absPath
			if err := service.OpenBrowser(fileURL); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Could not open browser: %v\n", err)
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "📊 Unified HTML report generated and opened: %s\n", absPath)
				return nil
			}
		}
		// If not opened, fall through to standard success message
	}

	// Display success message
	formatName := strings.ToUpper(format)
	fmt.Fprintf(cmd.ErrOrStderr(), "📊 Unified %s report generated: %s\n", formatName, absPath)

	return nil
}

// runCBOAnalysis runs CBO analysis for the given arguments (but doesn't output in analyze mode)
func (c *AnalyzeCommand) runCBOAnalysis(cmd *cobra.Command, args []string) error {
	// In analyze mode, we don't output individual CBO reports to stdout
	// The results will be included in the unified status report at the end
	// This prevents CBO output from interfering with other analyses

	// Run analysis silently and just return success/failure for status tracking
	_, err := c.runCBOAnalysisWithResult(cmd, args)
	if err != nil {
		// Add more context for debugging
		return fmt.Errorf("CBO analysis failed: %w", err)
	}
	return nil
}

// runCBOAnalysisWithResult runs CBO analysis and returns the result for unified reporting
func (c *AnalyzeCommand) runCBOAnalysisWithResult(cmd *cobra.Command, args []string) (*domain.CBOResponse, error) {
	// Import CBO-related packages
	cboService := service.NewCBOService()
	fileReader := service.NewFileReader()
	formatter := service.NewCBOFormatter()

	// Build CBO request
	request := domain.CBORequest{
		Paths:           args,
		OutputFormat:    domain.OutputFormatJSON, // For unified report
		OutputWriter:    cmd.OutOrStdout(),       // Required for validation
		MinCBO:          c.minCBO,
		MaxCBO:          0, // No limit
		SortBy:          domain.SortByCoupling,
		ShowZeros:       c.minCBO == 0,
		LowThreshold:    5,
		MediumThreshold: 10,
		Recursive:       true,
		IncludePatterns: []string{"*.py"},
		ExcludePatterns: []string{},
		IncludeBuiltins: false,
		IncludeImports:  true,
	}

	// Create use case
	cboUseCase, err := app.NewCBOUseCaseBuilder().
		WithService(cboService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create CBO use case: %w", err)
	}

	// Execute analysis and return result
	return cboUseCase.AnalyzeAndReturn(cmd.Context(), request)
}

// containsAnalysis checks if the given analysis is in the selectAnalyses list
func (c *AnalyzeCommand) containsAnalysis(analysis string) bool {
	for _, a := range c.selectAnalyses {
		if strings.ToLower(a) == analysis {
			return true
		}
	}
	return false
}

// NewAnalyzeCmd creates and returns the analyze cobra command
func NewAnalyzeCmd() *cobra.Command {
	analyzeCommand := NewAnalyzeCommand()
	return analyzeCommand.CreateCobraCommand()
}
