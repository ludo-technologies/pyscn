package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	internalconfig "github.com/ludo-technologies/pyscn/internal/config"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/spf13/cobra"
)

// CheckCommand represents a quick check command with sensible defaults
type CheckCommand struct {
	// Configuration
	configFile string
	quiet      bool

	// Quick override flags
	maxComplexity     int
	allowDeadCode     bool
	allowParseErrors  bool
	skipClones        bool
	allowCircularDeps bool
	maxCycles         int

	// Select specific analyses to run
	selectAnalyses []string
}

// NewCheckCommand creates a new check command
func NewCheckCommand() *CheckCommand {
	return &CheckCommand{
		configFile:        "",
		quiet:             false,
		maxComplexity:     10,    // Fail if complexity > 10
		allowDeadCode:     false, // Fail on any dead code
		allowParseErrors:  false, // Fail on any file that cannot be analyzed
		skipClones:        false,
		allowCircularDeps: false, // Fail on any circular dependencies
		maxCycles:         0,     // Fail if more than 0 cycles found
		selectAnalyses:    []string{},
	}
}

// CreateCobraCommand creates the cobra command for quick checking
func (c *CheckCommand) CreateCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check [files...]",
		Short: "Quick code quality check with sensible defaults",
		Long: `Quick code quality check optimized for CI/CD pipelines.

This command performs a fast analysis with predefined thresholds:
• Complexity: Fails if any function has complexity > 10
• Dead Code: Fails if any critical dead code is found
• Clones: Reports clones with similarity > 0.8 (warning only)
• Circular Dependencies: Fails if any cycles are detected
• DI Anti-patterns: Detects dependency injection anti-patterns
• Parse Errors: Fails if any target file cannot be read or parsed
By default, complexity, dead code, and clones analyses are run. Use --select to choose specific analyses.

Exit codes:
• 0: No issues found
• 1: Quality issues found (see output for details)
• 2: Analysis failed (invalid input, missing files, unparseable files, etc.)

The check command is designed to be fast and CI-friendly with minimal output
unless issues are found.

Examples:
  # Check current directory (typical CI usage)
  pyscn check .

  # Check only complexity (like ruff --select C901)
  pyscn check --select complexity --max-complexity 10 src/

  # Check only dead code
  pyscn check --select deadcode src/

  # Check complexity and dead code, skip clones
  pyscn check --select complexity,deadcode src/

  # Check only for circular dependencies
  pyscn check --select deps src/

  # Check for DI anti-patterns
  pyscn check --select di src/

	# Check with higher complexity threshold
  pyscn check --max-complexity 15 src/

  # Allow dead code, only check complexity
  pyscn check --allow-dead-code src/

  # Report unparseable files without failing the gate
  pyscn check --allow-parse-errors src/

  # Allow circular dependencies (warning only)
  pyscn check --allow-circular-deps src/
  
	# Allow up to 3 circular dependency cycles
  pyscn check --max-cycles 3 src/
  
	# Skip clone detection for faster analysis
  pyscn check --skip-clones src/`,
		Args: cobra.ArbitraryArgs,
		RunE: c.runCheck,
		// A failing gate is a verdict, not a usage mistake: dumping the flag
		// list after every failed CI run buries the findings above it.
		SilenceUsage: true,
	}

	// An unusable invocation is "invalid input", which the exit-code contract
	// above assigns to 2. Without this, an unknown flag would exit 1 and read
	// as a quality failure.
	cmd.SetFlagErrorFunc(func(_ *cobra.Command, err error) error {
		return newAnalysisError(err)
	})

	// Configuration flags
	cmd.Flags().StringVarP(&c.configFile, "config", "c", "", "Configuration file path")
	cmd.Flags().BoolVarP(&c.quiet, "quiet", "q", false, "Suppress output unless issues found")

	// Override flags for quick adjustments
	cmd.Flags().IntVar(&c.maxComplexity, "max-complexity", 10, "Maximum allowed complexity")
	cmd.Flags().BoolVar(&c.allowDeadCode, "allow-dead-code", false, "Allow dead code (don't fail)")
	cmd.Flags().BoolVar(&c.allowParseErrors, "allow-parse-errors", false, "Allow files that cannot be parsed (report them but don't fail)")
	cmd.Flags().BoolVar(&c.skipClones, "skip-clones", false, "Skip clone detection")
	cmd.Flags().BoolVar(&c.allowCircularDeps, "allow-circular-deps", false, "Allow circular dependencies (warnings only)")
	cmd.Flags().IntVar(&c.maxCycles, "max-cycles", 0, "Maximum allowed circular dependency cycles before failing")

	// Select specific analyses to run
	cmd.Flags().StringSliceVarP(&c.selectAnalyses, "select", "s", []string{},
		"Comma-separated list of analyses to run: complexity, deadcode, clones, deps, mockdata, di")

	return cmd
}

// runCheck executes the quick check analysis
func (c *CheckCommand) runCheck(cmd *cobra.Command, args []string) error {
	// Default to current directory if no args
	if len(args) == 0 {
		args = []string{"."}
	}

	// Resolve the configuration discovery result once and load it explicitly.
	// Discovery starts at the analyzed target - not the working directory - so
	// checking a repository nested in another working tree cannot pick up the
	// outer project's config (issue #666). Loading here also ensures a
	// discovered but malformed config fails the quality gate instead of being
	// silently replaced with defaults by individual loaders.
	originalConfigFile := c.configFile
	resolvedConfigFile, err := resolveCheckConfig(c.configFile, getTargetPathFromArgs(args))
	if err != nil {
		return newAnalysisError(err)
	}
	c.configFile = resolvedConfigFile
	defer func() {
		c.configFile = originalConfigFile
	}()

	// Validate selected analyses before creating config
	if len(c.selectAnalyses) > 0 {
		if err := c.validateSelectedAnalyses(); err != nil {
			return newAnalysisError(fmt.Errorf("invalid --select flag: %w", err))
		}
	}

	// Create use case configuration
	skipComplexity, skipDeadCode, skipClones, skipDeps, skipMockdata, skipDI := c.determineEnabledAnalyses()

	// Count issues found
	var issueCount int
	var hasErrors bool

	if !c.quiet {
		fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Running quality check (%s)...\n", strings.Join(c.getEnabledAnalyses(skipComplexity, skipDeadCode, skipClones, skipDeps, skipMockdata, skipDI), ", "))
	}

	response, snapshot, analysisErr := c.runCoreAnalysis(cmd, args, skipComplexity, skipDeadCode, skipClones, skipDeps)
	if response != nil {
		if err := c.reportProjectDiagnostics(cmd.ErrOrStderr(), response.Diagnostics); err != nil {
			hasErrors = true
		}

		if !skipComplexity {
			issueCount += c.countComplexityIssues(cmd, response.Complexity)
		}
		if !skipDeadCode {
			deadCodeIssues := c.countDeadCodeIssues(cmd, response.DeadCode)
			if !c.allowDeadCode {
				issueCount += deadCodeIssues
			} else if deadCodeIssues > 0 && !c.quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "Found %d dead code issue(s) (ignored due to --allow-dead-code)\n", deadCodeIssues)
			}
		}
		if !skipClones {
			cloneIssues := c.countCloneIssues(cmd, response.Clone)
			if cloneIssues > 0 && !c.quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Found %d code clone(s) (informational)\n", cloneIssues)
			}
		}
		if !skipDeps {
			depsIssues := c.countCircularDependencyIssues(cmd, response.System)
			if depsIssues > c.maxCycles {
				if !c.allowCircularDeps {
					issueCount += depsIssues
				} else if depsIssues > 0 && !c.quiet {
					fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Found %d circular dependency cycle(s) (allowed by --allow-circular-deps)\n", depsIssues)
				}
			} else if depsIssues > 0 && !c.quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "✓ Found %d circular dependency cycle(s) (within allowed limit of %d)\n", depsIssues, c.maxCycles)
			}
		}
	}
	if analysisErr != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "❌ Analysis failed: %v\n", analysisErr)
		hasErrors = true
	}

	// Run mock data check if enabled
	if !skipMockdata {
		mockdataIssues, err := c.checkMockdata(cmd, args, snapshot)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Mock data check failed: %v\n", err)
			hasErrors = true
		} else {
			issueCount += mockdataIssues
		}
	}

	// Run DI anti-pattern check if enabled
	if !skipDI {
		diIssues, err := c.checkDIAntipatterns(cmd, args, snapshot)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ DI anti-pattern check failed: %v\n", err)
			hasErrors = true
		} else {
			issueCount += diIssues
		}
	}

	// Handle results
	if hasErrors {
		return newAnalysisError(fmt.Errorf("analysis failed with errors"))
	}

	// Generic issue handling
	if issueCount > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(),
			"❌ Found %d quality issue(s)\n", issueCount)
		return fmt.Errorf("found %d quality issue(s)", issueCount)
	}

	if !c.quiet {
		fmt.Fprintf(cmd.ErrOrStderr(), "✅ Code quality check passed\n")
	}

	return nil
}

func resolveCheckConfig(configPath string, targetPath string) (string, error) {
	loader := internalconfig.NewTomlConfigLoader()
	resolvedPath, err := loader.ResolveConfigPath(configPath, targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve configuration: %w", err)
	}
	if resolvedPath == "" {
		return "", nil
	}

	if _, err := loader.LoadConfig(resolvedPath); err != nil {
		return "", fmt.Errorf("failed to load configuration from %s: %w", resolvedPath, err)
	}

	return resolvedPath, nil
}

// determineEnabledAnalyses determines which analyses should run based on flags
func (c *CheckCommand) determineEnabledAnalyses() (skipComplexity bool, skipDeadCode bool, skipClones bool, skipDeps bool, skipMockdata bool, skipDI bool) {
	if len(c.selectAnalyses) > 0 {
		// If --select is used, only run selected analyses
		skipComplexity = !c.containsAnalysis("complexity")
		skipDeadCode = !c.containsAnalysis("deadcode")
		skipClones = !c.containsAnalysis("clones")
		skipDeps = !c.containsAnalysis("deps") && !c.containsAnalysis("circular")
		skipMockdata = !c.containsAnalysis("mockdata")
		skipDI = !c.containsAnalysis("di")
	} else {
		// Otherwise use original behavior (backward compatible)
		skipComplexity = false    // Always run complexity
		skipDeadCode = false      // Always run dead code analysis
		skipClones = c.skipClones // Only skip clones if explicitly requested
		skipDeps = true           // Skip deps by default (opt-in via --select)
		skipMockdata = true       // Skip mockdata by default (opt-in via --select)
		skipDI = true             // Skip DI by default (opt-in via --select)
	}
	return
}

// containsAnalysis checks if the specified analysis is in the select list
func (c *CheckCommand) containsAnalysis(analysis string) bool {
	for _, a := range c.selectAnalyses {
		lowered := strings.ToLower(a)
		if lowered == analysis {
			return true
		}
		// Support both 'deps' and 'circular' for circular dependency analysis
		if (analysis == "deps" && lowered == "circular") || (analysis == "circular" && lowered == "deps") {
			return true
		}
	}
	return false
}

// getEnabledAnalyses returns a list of enabled analyses for display
func (c *CheckCommand) getEnabledAnalyses(skipComplexity bool, skipDeadCode bool, skipClones bool, skipDeps bool, skipMockdata bool, skipDI bool) []string {
	var enabled []string
	if !skipComplexity {
		enabled = append(enabled, "complexity")
	}
	if !skipDeadCode {
		enabled = append(enabled, "deadcode")
	}
	if !skipClones {
		enabled = append(enabled, "clones")
	}
	if !skipDeps {
		enabled = append(enabled, "deps")
	}
	if !skipMockdata {
		enabled = append(enabled, "mockdata")
	}
	if !skipDI {
		enabled = append(enabled, "di")
	}
	return enabled
}

// validateSelectedAnalyses validates the --select flag values
func (c *CheckCommand) validateSelectedAnalyses() error {
	validAnalyses := map[string]bool{
		"complexity": true,
		"deadcode":   true,
		"clones":     true,
		"deps":       true,
		"circular":   true,
		"mockdata":   true,
		"di":         true,
	}
	for _, analysis := range c.selectAnalyses {
		if !validAnalyses[strings.ToLower(analysis)] {
			return fmt.Errorf("invalid analysis type: %s. Valid options: complexity, deadcode, clones, deps, mockdata, di", analysis)
		}
	}
	if len(c.selectAnalyses) == 0 {
		return fmt.Errorf("--select flag requires at least one analysis type")
	}

	return nil
}

// checkMockdata runs mock data analysis and returns issue count
func (c *CheckCommand) checkMockdata(cmd *cobra.Command, args []string, snapshot *service.ProjectSnapshot) (int, error) {
	// Create request with check-specific settings
	request := &domain.MockDataRequest{
		Paths:        args,
		OutputFormat: domain.OutputFormatText,
		OutputWriter: io.Discard,
		ConfigPath:   c.configFile,
	}

	// Create service components
	mockDataService := service.NewMockDataService()
	mockDataFormatter := service.NewMockDataFormatter()
	mockDataConfigLoader := service.NewMockDataConfigurationLoader()

	useCase := app.NewSnapshotMockDataUseCase(mockDataService, service.NewFileReader(), mockDataFormatter, mockDataConfigLoader)

	// Run analysis
	response, err := useCase.AnalyzeSnapshotAndReturn(cmd.Context(), snapshot, *request)
	if err != nil {
		return 0, err
	}
	if err := c.reportAnalysisFailures(cmd.ErrOrStderr(), response.Failures); err != nil {
		return 0, err
	}

	// Count and output issues
	issueCount := 0
	for _, file := range response.Files {
		for _, finding := range file.Findings {
			// Only count warning and error level findings
			if finding.Severity.IsAtLeast(domain.MockDataSeverityWarning) {
				issueCount++
				if !c.quiet {
					fmt.Fprintf(cmd.ErrOrStderr(), "%s:%d:%d: mock data detected: %s (%s)\n",
						finding.Location.FilePath,
						finding.Location.StartLine,
						finding.Location.StartColumn,
						finding.Description,
						finding.Rationale)
				}
			}
		}
	}

	return issueCount, nil
}

// checkDIAntipatterns runs DI anti-pattern detection and returns issue count
func (c *CheckCommand) checkDIAntipatterns(cmd *cobra.Command, args []string, snapshot *service.ProjectSnapshot) (int, error) {
	// Create request with check-specific settings
	request := &domain.DIAntipatternRequest{
		Paths:        args,
		OutputFormat: domain.OutputFormatText,
		OutputWriter: io.Discard,
		ConfigPath:   c.configFile,
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return 0, fmt.Errorf("invalid DI anti-pattern request: %w", err)
	}

	// Create service components
	diService := service.NewDIAntipatternService()
	diFormatter := service.NewDIAntipatternFormatter()
	diConfigLoader := service.NewDIAntipatternConfigurationLoader()

	useCase := app.NewSnapshotDIAntipatternUseCase(diService, service.NewFileReader(), diFormatter, diConfigLoader)

	// Run analysis
	response, err := useCase.AnalyzeSnapshotAndReturn(cmd.Context(), snapshot, *request)
	if err != nil {
		return 0, err
	}
	if err := c.reportAnalysisFailures(cmd.ErrOrStderr(), response.Failures); err != nil {
		return 0, err
	}

	return c.countDIAntipatternIssues(cmd.ErrOrStderr(), response)
}

func (c *CheckCommand) reportAnalysisFailures(writer io.Writer, failures []domain.AnalysisFailure) error {
	if len(failures) == 0 {
		return nil
	}
	if !c.quiet {
		for _, failure := range failures {
			fmt.Fprintf(writer, "%s: %s: %s\n", failure.FilePath, failure.Code, failure.Message)
		}
	}
	return fmt.Errorf("%d analyzer execution failure(s)", len(failures))
}

func (c *CheckCommand) countDIAntipatternIssues(writer io.Writer, response *domain.DIAntipatternResponse) (int, error) {
	issueCount := 0
	for _, finding := range response.Findings {
		if finding.Severity.IsAtLeast(domain.DIAntipatternSeverityWarning) {
			issueCount++
			if !c.quiet {
				fmt.Fprintf(writer, "%s:%d:%d: %s: %s\n",
					finding.Location.FilePath,
					finding.Location.StartLine,
					finding.Location.StartCol+1,
					finding.Type,
					finding.Description)
			}
		}
	}

	return issueCount, nil
}

// NewCheckCmd creates and returns the check cobra command
func NewCheckCmd() *cobra.Command {
	checkCommand := NewCheckCommand()
	return checkCommand.CreateCobraCommand()
}
