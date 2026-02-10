package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/version"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

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
	skipSystem     bool
	selectAnalyses []string // Only run specified analyses

	// Quick filters
	minComplexity   int
	minSeverity     string
	cloneSimilarity float64
	minCBO          int

	// Clone detection options
	enableDFA bool // Enable Data Flow Analysis for enhanced Type-4 detection

	// System analysis options
	detectCycles bool // Detect circular dependencies
	validateArch bool // Validate architecture rules
}

// NewAnalyzeCommand creates a new analyze command
func NewAnalyzeCommand() *AnalyzeCommand {
	return &AnalyzeCommand{
		html:            false,
		json:            false,
		csv:             false,
		yaml:            false,
		noOpen:          false,
		configFile:      "",
		verbose:         false,
		skipComplexity:  false,
		skipDeadCode:    false,
		skipClones:      false,
		skipCBO:         false,
		skipSystem:      false,
		minComplexity:   5,
		minSeverity:     "warning",
		cloneSimilarity: 0.65,
		minCBO:          0,
		enableDFA:       true,
		detectCycles:    true,
		validateArch:    true,
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
• System-level analysis (module dependencies and architecture)

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
	cmd.Flags().Float64Var(&c.cloneSimilarity, "clone-threshold", 0.65, "Minimum similarity for clone detection (0.0-1.0)")
	cmd.Flags().IntVar(&c.minCBO, "min-cbo", 0, "Minimum CBO to report")

	return cmd
}

// runAnalyze executes the comprehensive analysis using the clean architecture
func (c *AnalyzeCommand) runAnalyze(cmd *cobra.Command, args []string) error {
	// Get verbose flag from parent command
	if cmd.Parent() != nil {
		c.verbose, _ = cmd.Parent().Flags().GetBool("verbose")
	}

	// Create use case configuration
	config := c.createUseCaseConfig()

	// Build the analyze use case
	useCase, err := c.buildAnalyzeUseCase(cmd)
	if err != nil {
		return fmt.Errorf("failed to build analyze use case: %w", err)
	}

	// Execute analysis with timeout and cancellation support
	ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Minute)
	defer cancel()
	response, analysisErr := useCase.Execute(ctx, config, args)

	// Generate output even if there were partial failures
	if response != nil {
		// Generate output
		if err := c.generateOutput(cmd, response, args); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Failed to generate output: %v\n", err)
		}

		// Print summary
		c.printSummary(cmd, response)
	}

	// Return the analysis error so CLI exits with non-zero status
	if analysisErr != nil {
		return analysisErr
	}

	return nil
}

// createUseCaseConfig creates the use case configuration from command flags
func (c *AnalyzeCommand) createUseCaseConfig() app.AnalyzeUseCaseConfig {
	config := app.AnalyzeUseCaseConfig{
		ConfigFile:      c.configFile,
		Verbose:         c.verbose,
		MinComplexity:   c.minComplexity,
		CloneSimilarity: c.cloneSimilarity,
		MinCBO:          c.minCBO,
		EnableDFA:       c.enableDFA,
	}

	// Handle analysis selection
	if len(c.selectAnalyses) > 0 {
		// If --select is used, only run selected analyses
		config.SkipComplexity = !c.containsAnalysis("complexity")
		config.SkipDeadCode = !c.containsAnalysis("deadcode")
		config.SkipClones = !c.containsAnalysis("clones")
		config.SkipCBO = !c.containsAnalysis("cbo")
		config.SkipSystem = !c.containsAnalysis("deps")
	} else {
		// Otherwise use skip flags
		config.SkipComplexity = c.skipComplexity
		config.SkipDeadCode = c.skipDeadCode
		config.SkipClones = c.skipClones
		config.SkipCBO = c.skipCBO
		config.SkipSystem = c.skipSystem
	}

	// Parse severity
	switch c.minSeverity {
	case "critical":
		config.MinSeverity = domain.DeadCodeSeverityCritical
	case "warning":
		config.MinSeverity = domain.DeadCodeSeverityWarning
	case "info":
		config.MinSeverity = domain.DeadCodeSeverityInfo
	default:
		config.MinSeverity = domain.DeadCodeSeverityWarning
	}

	return config
}

// buildAnalyzeUseCase builds the analyze use case with all dependencies
func (c *AnalyzeCommand) buildAnalyzeUseCase(cmd *cobra.Command) (*app.AnalyzeUseCase, error) {
	builder := app.NewAnalyzeUseCaseBuilder()

	// Set up file reader
	fileReader := service.NewFileReader()
	builder.WithFileReader(fileReader)

	// Set up formatter
	formatter := service.NewAnalyzeFormatter()
	builder.WithFormatter(formatter)

	// Set up progress manager
	progressManager := service.NewProgressManager()
	if c.shouldUseProgressBars(cmd) {
		progressManager.SetWriter(cmd.ErrOrStderr())
	} else {
		progressManager.SetWriter(io.Discard)
	}
	builder.WithProgressManager(progressManager)

	// Set up parallel executor
	parallelExecutor := service.NewParallelExecutor()
	builder.WithParallelExecutor(parallelExecutor)

	// Set up error categorizer
	errorCategorizer := service.NewErrorCategorizer()
	builder.WithErrorCategorizer(errorCategorizer)

	// Build individual use cases
	if err := c.buildIndividualUseCases(builder); err != nil {
		return nil, err
	}

	return builder.Build()
}

// buildIndividualUseCases builds and sets individual analysis use cases
func (c *AnalyzeCommand) buildIndividualUseCases(builder *app.AnalyzeUseCaseBuilder) error {
	// Complexity use case
	complexityService := service.NewComplexityService()
	complexityFormatter := service.NewOutputFormatter()
	complexityConfigLoader := service.NewConfigurationLoader()
	complexityUseCase := app.NewComplexityUseCase(
		complexityService,
		service.NewFileReader(),
		complexityFormatter,
		complexityConfigLoader,
	)
	builder.WithComplexityUseCase(complexityUseCase)

	// Dead code use case
	deadCodeService := service.NewDeadCodeService()
	deadCodeFormatter := service.NewDeadCodeFormatter()
	deadCodeConfigLoader := service.NewDeadCodeConfigurationLoader()
	deadCodeUseCase := app.NewDeadCodeUseCase(
		deadCodeService,
		service.NewFileReader(),
		deadCodeFormatter,
		deadCodeConfigLoader,
	)
	builder.WithDeadCodeUseCase(deadCodeUseCase)

	// Clone use case
	cloneService := service.NewCloneService()
	cloneFormatter := service.NewCloneOutputFormatter()
	cloneConfigLoader := service.NewCloneConfigurationLoader()
	cloneUseCase, err := app.NewCloneUseCaseBuilder().
		WithService(cloneService).
		WithFileReader(service.NewFileReader()).
		WithFormatter(cloneFormatter).
		WithConfigLoader(cloneConfigLoader).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build clone use case: %w", err)
	}
	builder.WithCloneUseCase(cloneUseCase)

	// CBO use case
	cboService := service.NewCBOService()
	cboFormatter := service.NewCBOFormatter()
	cboConfigLoader := service.NewCBOConfigurationLoader()
	cboUseCase, err := app.NewCBOUseCaseBuilder().
		WithService(cboService).
		WithFileReader(service.NewFileReader()).
		WithFormatter(cboFormatter).
		WithConfigLoader(cboConfigLoader).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build CBO use case: %w", err)
	}
	builder.WithCBOUseCase(cboUseCase)

	// System analysis use case
	systemService := service.NewSystemAnalysisService()
	systemFormatter := service.NewSystemAnalysisFormatter()
	systemConfigLoader := service.NewSystemAnalysisConfigurationLoader()
	systemUseCase, err := app.NewSystemAnalysisUseCaseBuilder().
		WithService(systemService).
		WithFileReader(service.NewFileReader()).
		WithFormatter(systemFormatter).
		WithConfigLoader(systemConfigLoader).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build system analysis use case: %w", err)
	}
	builder.WithSystemUseCase(systemUseCase)

	return nil
}

// generateOutput generates the output report
func (c *AnalyzeCommand) generateOutput(cmd *cobra.Command, response *domain.AnalyzeResponse, args []string) error {
	// Determine output format
	format, extension, err := c.determineOutputFormat()
	if err != nil {
		return err
	}

	// Generate filename with timestamp
	targetPath := getTargetPathFromArgs(args)
	filename, err := generateOutputFilePath("analyze", extension, targetPath)
	if err != nil {
		return fmt.Errorf("failed to generate output path: %w", err)
	}

	// Add version to response
	response.Version = version.Version

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
		// Auto-open only when explicitly allowed, environment is interactive, and not over SSH
		if !c.noOpen && service.IsInteractiveEnvironment() && !service.IsSSH() {
			fileURL := "file://" + absPath
			if err := service.OpenBrowser(fileURL); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: Could not open browser: %v\n", err)
			} else {
				fmt.Fprintf(cmd.ErrOrStderr(), "📊 Unified HTML report generated and opened: %s\n", absPath)
				return nil
			}
		}
	}

	// Display success message
	formatName := strings.ToUpper(format)
	fmt.Fprintf(cmd.ErrOrStderr(), "📊 Unified %s report generated: %s\n", formatName, absPath)

	return nil
}

// getScoreIcon returns an icon representing the score quality
func getScoreIcon(score int) string {
	switch {
	case score >= domain.ScoreThresholdExcellent:
		return "✅" // Excellent
	case score >= domain.ScoreThresholdGood:
		return "👍" // Good
	case score >= domain.ScoreThresholdFair:
		return "⚠️" // Fair
	default:
		return "❌" // Poor
	}
}

// printSummary prints a summary of the analysis results
func (c *AnalyzeCommand) printSummary(cmd *cobra.Command, response *domain.AnalyzeResponse) {
	fmt.Fprintf(cmd.ErrOrStderr(), "\n📊 Analysis Summary:\n")
	fmt.Fprintf(cmd.ErrOrStderr(), "Health Score: %d/100 (Grade: %s)\n", response.Summary.HealthScore, response.Summary.Grade)
	fmt.Fprintf(cmd.ErrOrStderr(), "Total time: %dms\n\n", response.Duration)

	// Print detailed scores section
	fmt.Fprintf(cmd.ErrOrStderr(), "📈 Detailed Scores:\n")

	if response.Summary.ComplexityEnabled {
		icon := getScoreIcon(response.Summary.ComplexityScore)
		fmt.Fprintf(cmd.ErrOrStderr(), "  Complexity:     %3d/100 %s  (avg: %.1f, high-risk: %d functions)\n",
			response.Summary.ComplexityScore, icon,
			response.Summary.AverageComplexity, response.Summary.HighComplexityCount)
	}

	if response.Summary.DeadCodeEnabled {
		icon := getScoreIcon(response.Summary.DeadCodeScore)
		fmt.Fprintf(cmd.ErrOrStderr(), "  Dead Code:      %3d/100 %s  (%d issues, %d critical)\n",
			response.Summary.DeadCodeScore, icon,
			response.Summary.DeadCodeCount, response.Summary.CriticalDeadCode)
	}

	if response.Summary.CloneEnabled {
		icon := getScoreIcon(response.Summary.DuplicationScore)
		fmt.Fprintf(cmd.ErrOrStderr(), "  Duplication:    %3d/100 %s  (%.1f%% duplication, %d groups)\n",
			response.Summary.DuplicationScore, icon,
			response.Summary.CodeDuplication, response.Summary.CloneGroups)
	}

	if response.Summary.CBOEnabled {
		icon := getScoreIcon(response.Summary.CouplingScore)
		fmt.Fprintf(cmd.ErrOrStderr(), "  Coupling (CBO): %3d/100 %s  (avg: %.1f, %d/%d high-coupling)\n",
			response.Summary.CouplingScore, icon,
			response.Summary.AverageCoupling, response.Summary.HighCouplingClasses, response.Summary.CBOClasses)
	}

	if response.Summary.DepsEnabled {
		icon := getScoreIcon(response.Summary.DependencyScore)
		cyclesMsg := fmt.Sprintf("%d cycles", response.Summary.DepsModulesInCycles)
		if response.Summary.DepsModulesInCycles == 0 {
			cyclesMsg = "no cycles"
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "  Dependencies:   %3d/100 %s  (%s, depth: %d)\n",
			response.Summary.DependencyScore, icon,
			cyclesMsg, response.Summary.DepsMaxDepth)
	}

	if response.Summary.ArchEnabled {
		icon := getScoreIcon(response.Summary.ArchitectureScore)
		fmt.Fprintf(cmd.ErrOrStderr(), "  Architecture:   %3d/100 %s  (%.0f%% compliant)\n",
			response.Summary.ArchitectureScore, icon,
			response.Summary.ArchCompliance*100)
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "\n")

	// Print analysis status summary
	if response.Summary.ComplexityEnabled {
		if response.Complexity != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "✅ Complexity Analysis: %d functions analyzed\n",
				response.Summary.TotalFunctions)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Complexity Analysis: Failed\n")
		}
	}

	if response.Summary.DeadCodeEnabled {
		if response.DeadCode != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "✅ Dead Code Detection: Completed\n")
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Dead Code Detection: Failed\n")
		}
	}

	if response.Summary.CloneEnabled {
		if response.Clone != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "✅ Clone Detection: Completed\n")
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Clone Detection: Failed\n")
		}
	}

	if response.Summary.CBOEnabled {
		if response.CBO != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "✅ Class Coupling: %d classes analyzed\n",
				response.Summary.CBOClasses)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Class Coupling: Failed\n")
		}
	}

	if response.Summary.DepsEnabled {
		if response.System != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "✅ System Analysis: %d modules analyzed\n",
				response.Summary.DepsTotalModules)
		} else {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ System Analysis: Failed\n")
		}
	}

	// Print README badge snippet
	c.printBadge(cmd, response.Summary.Grade)
}

const badgeLandingURL = "https://pyscn.ludo-tech.org"

// printBadge prints a Markdown badge snippet for the user's README
func (c *AnalyzeCommand) printBadge(cmd *cobra.Command, grade string) {
	color := gradeBadgeColor(grade)
	badge := fmt.Sprintf("[![pyscn quality](https://img.shields.io/badge/pyscn-%s-%s)](%s)",
		grade, color, badgeLandingURL)

	fmt.Fprintf(cmd.ErrOrStderr(), "\n--------------------------------------------------\n")
	fmt.Fprintf(cmd.ErrOrStderr(), "[Badge] Add this to your README to show off your score:\n")
	fmt.Fprintf(cmd.ErrOrStderr(), "%s\n", badge)
	fmt.Fprintf(cmd.ErrOrStderr(), "--------------------------------------------------\n")
}

// gradeBadgeColor returns a shields.io color name for the given grade
func gradeBadgeColor(grade string) string {
	switch grade {
	case "A":
		return "brightgreen"
	case "B":
		return "yellow"
	case "C":
		return "orange"
	case "D":
		return "red"
	case "F":
		return "critical"
	default:
		return "lightgrey"
	}
}

// Helper methods

// determineOutputFormat determines the output format based on flags
func (c *AnalyzeCommand) determineOutputFormat() (string, string, error) {
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

	// Default to HTML if no format specified
	if formatCount == 0 {
		return "html", "html", nil
	}

	return format, extension, nil
}

// shouldUseProgressBars returns true when the session appears to be interactive
func (c *AnalyzeCommand) shouldUseProgressBars(cmd *cobra.Command) bool {
	if !service.IsInteractiveEnvironment() {
		return false
	}

	if errWriter, ok := cmd.ErrOrStderr().(*os.File); ok {
		return term.IsTerminal(int(errWriter.Fd()))
	}

	return false
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
