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
	text   bool
	noOpen bool

	// Configuration
	configFile string
	verbose    bool

	// Analysis selection
	skipComplexity  bool
	skipDeadCode    bool
	skipClones      bool
	skipCBO         bool
	skipLCOM        bool
	skipSystem      bool
	skipCommunities bool
	selectAnalyses  []string // Only run specified analyses

	// Quick filters
	minComplexity   int
	minSeverity     string
	cloneSimilarity float64
	minCBO          int

	// Complexity thresholds (0 = unset, use config/default)
	lowThreshold                 int
	mediumThreshold              int
	cognitiveComplexityThreshold int
	nestingDepthThreshold        int

	functionSLOCWarnThreshold     int
	functionSLOCCriticalThreshold int

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
		text:            false,
		noOpen:          false,
		configFile:      "",
		verbose:         false,
		skipComplexity:  false,
		skipDeadCode:    false,
		skipClones:      false,
		skipCBO:         false,
		skipLCOM:        false,
		skipSystem:      false,
		minComplexity:   0,
		minSeverity:     "",
		cloneSimilarity: 0,
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
• Class cohesion analysis (LCOM4)
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
	cmd.Flags().BoolVar(&c.text, "text", false, "Generate plain-text report file")
	cmd.Flags().BoolVar(&c.noOpen, "no-open", false, "Don't auto-open HTML in browser")
	cmd.Flags().StringVarP(&c.configFile, "config", "c", "", "Configuration file path")

	// Analysis selection flags
	cmd.Flags().BoolVar(&c.skipComplexity, "skip-complexity", false, "Skip complexity analysis")
	cmd.Flags().BoolVar(&c.skipDeadCode, "skip-deadcode", false, "Skip dead code detection")
	cmd.Flags().BoolVar(&c.skipClones, "skip-clones", false, "Skip clone detection")
	cmd.Flags().BoolVar(&c.skipCBO, "skip-cbo", false, "Skip class coupling (CBO) analysis")
	cmd.Flags().BoolVar(&c.skipLCOM, "skip-lcom", false, "Skip class cohesion (LCOM4) analysis")
	cmd.Flags().BoolVar(&c.skipSystem, "skip-deps", false, "Skip module dependencies and architecture analysis")
	cmd.Flags().BoolVar(&c.skipCommunities, "skip-communities", false, "Skip module community detection")
	cmd.Flags().StringSliceVar(&c.selectAnalyses, "select", []string{}, "Only run specified analyses (complexity,deadcode,clones,cbo,lcom,deps,communities)")

	// Quick filter flags
	cmd.Flags().IntVar(&c.minComplexity, "min-complexity", 0, "Minimum complexity to report (default: 1)")
	cmd.Flags().StringVar(&c.minSeverity, "min-severity", "", "Minimum dead code severity: critical, warning, info (default: warning)")
	cmd.Flags().Float64Var(&c.cloneSimilarity, "clone-threshold", 0, "Minimum similarity for clone detection, 0.0-1.0 (default: 0.65)")
	cmd.Flags().IntVar(&c.minCBO, "min-cbo", 0, "Minimum CBO to report")

	// Complexity threshold flags (0 = unset, use config file or default)
	cmd.Flags().IntVar(&c.lowThreshold, "low-threshold", 0, "Upper bound for low-risk complexity (default: 9)")
	cmd.Flags().IntVar(&c.mediumThreshold, "medium-threshold", 0, "Upper bound for medium-risk complexity (default: 19)")
	cmd.Flags().IntVar(&c.cognitiveComplexityThreshold, "cognitive-complexity-threshold", 0, "High-risk threshold for cognitive complexity (default: 25)")
	cmd.Flags().IntVar(&c.nestingDepthThreshold, "nesting-depth-threshold", 0, "High-risk threshold for maximum nesting depth (default: 7)")
	cmd.Flags().IntVar(&c.functionSLOCWarnThreshold, "function-sloc-warn-threshold", 0, "Function length in source lines reported as long (default: 50)")
	cmd.Flags().IntVar(&c.functionSLOCCriticalThreshold, "function-sloc-critical-threshold", 0, "Function length in source lines reported as too long (default: 100)")

	return cmd
}

// runAnalyze executes the comprehensive analysis using the clean architecture
func (c *AnalyzeCommand) runAnalyze(cmd *cobra.Command, args []string) error {
	// Get verbose flag from parent command
	if cmd.Parent() != nil {
		c.verbose, _ = cmd.Parent().Flags().GetBool("verbose")
	}

	if len(c.selectAnalyses) > 0 {
		if err := c.validateSelectedAnalyses(); err != nil {
			return fmt.Errorf("invalid --select flag: %w", err)
		}
	}

	switch c.minSeverity {
	case "", "critical", "warning", "info":
	default:
		return fmt.Errorf("invalid --min-severity value %q (expected: critical, warning, info)", c.minSeverity)
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
	var outputErr error
	if response != nil {
		// Generate output
		if err := c.generateOutput(cmd, response, args); err != nil {
			outputErr = err
		}

		// Print summary
		c.printSummary(cmd, response)
	}

	// Return the analysis error so CLI exits with non-zero status
	if analysisErr != nil {
		return analysisErr
	}

	// Return output error if analysis succeeded but output generation failed
	if outputErr != nil {
		return outputErr
	}

	return nil
}

// createUseCaseConfig creates the use case configuration from command flags
func (c *AnalyzeCommand) createUseCaseConfig() app.AnalyzeUseCaseConfig {
	config := app.AnalyzeUseCaseConfig{
		ConfigFile:              c.configFile,
		Verbose:                 c.verbose,
		MinComplexity:           c.minComplexity,
		CloneSimilarity:         c.cloneSimilarity,
		MinCBO:                  c.minCBO,
		EnableDFA:               c.enableDFA,
		SkipCommunities:         false,
		SkipCommunitiesExplicit: c.skipCommunities,

		LowThreshold:                 c.lowThreshold,
		MediumThreshold:              c.mediumThreshold,
		CognitiveComplexityThreshold: c.cognitiveComplexityThreshold,
		NestingDepthThreshold:        c.nestingDepthThreshold,

		FunctionSLOCWarnThreshold:     c.functionSLOCWarnThreshold,
		FunctionSLOCCriticalThreshold: c.functionSLOCCriticalThreshold,
	}
	config = app.ApplyAnalyzeSelection(config, c.selectAnalyses)

	// Handle analysis selection
	if len(c.selectAnalyses) == 0 {
		// Otherwise use skip flags
		config.SkipComplexity = c.skipComplexity
		config.SkipDeadCode = c.skipDeadCode
		config.SkipClones = c.skipClones
		config.SkipCBO = c.skipCBO
		config.SkipLCOM = c.skipLCOM
		config.SkipSystem = c.skipSystem
	}
	config.SkipComplexity = config.SkipComplexity || c.skipComplexity
	config.SkipDeadCode = config.SkipDeadCode || c.skipDeadCode
	config.SkipClones = config.SkipClones || c.skipClones
	config.SkipCBO = config.SkipCBO || c.skipCBO
	config.SkipLCOM = config.SkipLCOM || c.skipLCOM
	config.SkipSystem = config.SkipSystem || c.skipSystem
	if c.skipCommunities {
		config.SkipCommunities = true
	}

	// Parse severity; empty means "not set" and is resolved from the
	// config file (or defaults) during merge. Invalid values are rejected
	// in runAnalyze before this is called.
	switch c.minSeverity {
	case "critical":
		config.MinSeverity = domain.DeadCodeSeverityCritical
	case "warning":
		config.MinSeverity = domain.DeadCodeSeverityWarning
	case "info":
		config.MinSeverity = domain.DeadCodeSeverityInfo
	}

	return config
}

// buildAnalyzeUseCase builds the analyze use case with all dependencies
func (c *AnalyzeCommand) buildAnalyzeUseCase(cmd *cobra.Command) (*app.AnalyzeUseCase, error) {
	return buildAnalyzeUseCase(cmd, c.shouldUseProgressBars(cmd))
}

func buildAnalyzeUseCase(cmd *cobra.Command, showProgress bool) (*app.AnalyzeUseCase, error) {
	builder := app.NewAnalyzeUseCaseBuilder()

	// Set up file reader
	fileReader := service.NewFileReader()
	builder.WithFileReader(fileReader)
	builder.WithConfigLoader(service.NewAnalyzeConfigurationLoader())

	// Set up formatter
	formatter := service.NewAnalyzeFormatter()
	builder.WithFormatter(formatter)

	// Set up progress manager
	progressManager := service.NewProgressManager()
	if showProgress {
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
	if err := buildIndividualUseCases(builder); err != nil {
		return nil, err
	}

	return builder.Build()
}

// buildIndividualUseCases builds and sets individual analysis use cases
func buildIndividualUseCases(builder *app.AnalyzeUseCaseBuilder) error {
	// Complexity use case
	complexityService := service.NewComplexityService()
	complexityFormatter := service.NewOutputFormatter()
	complexityConfigLoader := service.NewConfigurationLoader()
	complexityUseCase := app.NewSnapshotComplexityUseCase(
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
	deadCodeUseCase := app.NewSnapshotDeadCodeUseCase(
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
		WithSnapshotService(cloneService).
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
		WithSnapshotService(cboService).
		WithFileReader(service.NewFileReader()).
		WithFormatter(cboFormatter).
		WithConfigLoader(cboConfigLoader).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build CBO use case: %w", err)
	}
	builder.WithCBOUseCase(cboUseCase)

	// LCOM use case
	lcomService := service.NewLCOMService()
	lcomFormatter := service.NewLCOMFormatter()
	lcomConfigLoader := service.NewLCOMConfigurationLoader()
	lcomUseCase, err := app.NewLCOMUseCaseBuilder().
		WithSnapshotService(lcomService).
		WithFileReader(service.NewFileReader()).
		WithFormatter(lcomFormatter).
		WithConfigLoader(lcomConfigLoader).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build LCOM use case: %w", err)
	}
	builder.WithLCOMUseCase(lcomUseCase)

	// System analysis use case
	systemService := service.NewSystemAnalysisService()
	systemFormatter := service.NewSystemAnalysisFormatter()
	systemConfigLoader := service.NewSystemAnalysisConfigurationLoader()
	systemUseCase, err := app.NewSystemAnalysisUseCaseBuilder().
		WithGraphService(systemService).
		WithFileReader(service.NewFileReader()).
		WithFormatter(systemFormatter).
		WithConfigLoader(systemConfigLoader).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build system analysis use case: %w", err)
	}
	builder.WithSystemUseCase(systemUseCase)

	// Community analysis use case
	communityService := service.NewCommunityAnalysisService()
	communityFormatter := service.NewCommunityFormatter()
	communityConfigLoader := service.NewCommunityConfigurationLoader()
	communityUseCase, err := app.NewCommunityUseCaseBuilder().
		WithGraphService(communityService).
		WithFileReader(service.NewFileReader()).
		WithFormatter(communityFormatter).
		WithConfigLoader(communityConfigLoader).
		Build()
	if err != nil {
		return fmt.Errorf("failed to build community analysis use case: %w", err)
	}
	builder.WithCommunityUseCase(communityUseCase)

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

	// Write standalone community JSON when only communities were selected.
	formatType := domain.OutputFormat(format)
	if c.shouldWriteStandaloneCommunityJSON(response) {
		communityFormatter := service.NewCommunityFormatter()
		if err := communityFormatter.Write(response.Communities, formatType, file); err != nil {
			return fmt.Errorf("failed to write community analysis report: %w", err)
		}
	} else if err := formatter.Write(response, formatType, file); err != nil {
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
	// Restate the shortfall next to the score. The per-file warnings are
	// printed before the analysis runs and scroll away, which is how an
	// unparseable file used to reach a Grade A unnoticed (issue #690).
	if response.Summary.SkippedFiles > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  %d of %d files skipped (parse errors) - excluded from every score below\n",
			response.Summary.SkippedFiles, response.Summary.TotalFiles)
		// Name them. The count alone tells you the report is incomplete but
		// not what to fix, and Errors was previously collected and never read.
		if response.Complexity != nil {
			for _, e := range response.Complexity.Errors {
				fmt.Fprintf(cmd.ErrOrStderr(), "    %s\n", e)
			}
		}
	}
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
		fmt.Fprintf(cmd.ErrOrStderr(), "  Duplication:    %3d/100 %s  (%.1f%% fragments cloned, %d groups)\n",
			response.Summary.DuplicationScore, icon,
			response.Summary.CodeDuplication, response.Summary.CloneGroups)
	}

	if response.Summary.CBOEnabled {
		icon := getScoreIcon(response.Summary.CouplingScore)
		fmt.Fprintf(cmd.ErrOrStderr(), "  Coupling (CBO): %3d/100 %s  (avg: %.1f, %d/%d high-coupling)\n",
			response.Summary.CouplingScore, icon,
			response.Summary.AverageCoupling, response.Summary.HighCouplingClasses, response.Summary.CBOClasses)
	}

	if response.Summary.LCOMEnabled {
		icon := getScoreIcon(response.Summary.CohesionScore)
		fmt.Fprintf(cmd.ErrOrStderr(), "  Cohesion (LCOM):%3d/100 %s  (avg: %.1f, %d/%d high-lcom)\n",
			response.Summary.CohesionScore, icon,
			response.Summary.AverageLCOM, response.Summary.HighLCOMClasses, response.Summary.LCOMClasses)
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

	// Print README badge snippet
	c.printBadge(cmd, response.Summary.Grade)
}

// badgeLandingURL is where the README badge points. It targets the repository
// so that a click from someone else's README lands on the project itself.
const badgeLandingURL = service.RepositoryURL

// printBadge prints a Markdown badge snippet for the user's README
func (c *AnalyzeCommand) printBadge(cmd *cobra.Command, grade string) {
	color := gradeBadgeColor(grade)
	badge := fmt.Sprintf("[![pyscn quality](https://img.shields.io/badge/pyscn-%s-%s)](%s)",
		grade, color, badgeLandingURL)

	fmt.Fprintf(cmd.ErrOrStderr(), "\n--------------------------------------------------\n")
	fmt.Fprintf(cmd.ErrOrStderr(), "[Badge] Add this to your README to show off your score:\n")
	fmt.Fprintf(cmd.ErrOrStderr(), "%s\n", badge)
	fmt.Fprintf(cmd.ErrOrStderr(), "\n[Star] Finding pyscn useful? %s\n", service.RepositoryURL)
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
	format, extension, err := service.NewOutputFormatResolver().DetermineAnalyzeReport(
		c.html,
		c.json,
		c.csv,
		c.yaml,
		c.text,
	)
	return string(format), extension, err
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

// shouldWriteStandaloneCommunityJSON returns true when community analysis is the
// only selected analyzer and JSON output is requested.
func (c *AnalyzeCommand) shouldWriteStandaloneCommunityJSON(response *domain.AnalyzeResponse) bool {
	return c.json &&
		len(c.selectAnalyses) == 1 &&
		c.containsAnalysis("communities") &&
		response != nil &&
		response.Communities != nil
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

func (c *AnalyzeCommand) validateSelectedAnalyses() error {
	validAnalyses := map[string]bool{
		"complexity":  true,
		"deadcode":    true,
		"clones":      true,
		"cbo":         true,
		"lcom":        true,
		"deps":        true,
		"communities": true,
	}
	for _, analysis := range c.selectAnalyses {
		if !validAnalyses[strings.ToLower(analysis)] {
			return fmt.Errorf("invalid analysis type: %s. Valid options: complexity, deadcode, clones, cbo, lcom, deps, communities", analysis)
		}
	}
	return nil
}

// NewAnalyzeCmd creates and returns the analyze cobra command
func NewAnalyzeCmd() *cobra.Command {
	analyzeCommand := NewAnalyzeCommand()
	return analyzeCommand.CreateCobraCommand()
}
