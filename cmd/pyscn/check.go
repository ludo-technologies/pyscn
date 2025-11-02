package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
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

By default, complexity, dead code, and clones analyses are run. Use --select to choose specific analyses.

Exit codes:
• 0: No issues found
• 1: Quality issues found (see output for details)
• 2: Analysis failed (invalid input, missing files, etc.)

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

  # Check with higher complexity threshold
  pyscn check --max-complexity 15 src/

  # Allow dead code, only check complexity
  pyscn check --allow-dead-code src/

  # Allow circular dependencies (warning only)
  pyscn check --allow-circular-deps src/

  # Allow up to 3 circular dependency cycles
  pyscn check --max-cycles 3 src/

  # Skip clone detection for faster analysis
  pyscn check --skip-clones src/`,
		Args: cobra.ArbitraryArgs,
		RunE: c.runCheck,
	}

	// Configuration flags
	cmd.Flags().StringVarP(&c.configFile, "config", "c", "", "Configuration file path")
	cmd.Flags().BoolVarP(&c.quiet, "quiet", "q", false, "Suppress output unless issues found")

	// Override flags for quick adjustments
	cmd.Flags().IntVar(&c.maxComplexity, "max-complexity", 10, "Maximum allowed complexity")
	cmd.Flags().BoolVar(&c.allowDeadCode, "allow-dead-code", false, "Allow dead code (don't fail)")
	cmd.Flags().BoolVar(&c.skipClones, "skip-clones", false, "Skip clone detection")
	cmd.Flags().BoolVar(&c.allowCircularDeps, "allow-circular-deps", false, "Allow circular dependencies (warnings only)")
	cmd.Flags().IntVar(&c.maxCycles, "max-cycles", 0, "Maximum allowed circular dependency cycles before failing")

	// Select specific analyses to run
	cmd.Flags().StringSliceVarP(&c.selectAnalyses, "select", "s", []string{},
		"Comma-separated list of analyses to run: complexity, deadcode, clones, deps")

	return cmd
}

// runCheck executes the quick check analysis
func (c *CheckCommand) runCheck(cmd *cobra.Command, args []string) error {
	// Default to current directory if no args
	if len(args) == 0 {
		args = []string{"."}
	}

	// Validate selected analyses before creating config
	if len(c.selectAnalyses) > 0 {
		if err := c.validateSelectedAnalyses(); err != nil {
			return fmt.Errorf("invalid --select flag: %w", err)
		}
	}

	// Create use case configuration
	skipComplexity, skipDeadCode, skipClones, skipDeps := c.determineEnabledAnalyses()

	// Count issues found
	var issueCount int
	var hasErrors bool

	if !c.quiet {
		fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Running quality check (%s)...\n", strings.Join(c.getEnabledAnalyses(skipComplexity, skipDeadCode, skipClones, skipDeps), ", "))
	}

	// Run complexity check if enabled
	if !skipComplexity {
		complexityIssues, err := c.checkComplexity(cmd, args)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Complexity analysis failed: %v\n", err)
			hasErrors = true
		} else {
			issueCount += complexityIssues
		}
	}

	// Run dead code check if enabled
	if !skipDeadCode {
		deadCodeIssues, err := c.checkDeadCode(cmd, args)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Dead code analysis failed: %v\n", err)
			hasErrors = true
		} else {
			// Only count dead code issues if not explicitly allowed
			if !c.allowDeadCode {
				issueCount += deadCodeIssues
			} else if deadCodeIssues > 0 && !c.quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "Found %d dead code issue(s) (ignored due to --allow-dead-code)\n", deadCodeIssues)
			}
		}
	}

	// Run clone check if enabled
	if !skipClones {
		cloneIssues, err := c.checkClones(cmd, args)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Clone detection failed: %v\n", err)
			// Don't treat clone detection failures as hard errors
		} else if cloneIssues > 0 {
			if !c.quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Found %d code clone(s) (informational)\n", cloneIssues)
			}
		}
	}

	// Run circular dependency check if enabled
	if !skipDeps {
		depsIssues, err := c.checkCircularDependencies(cmd, args)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Circular dependency check failed: %v\n", err)
			hasErrors = true
		} else {
			// Handle max-cycles threshold
			if depsIssues > c.maxCycles {
				if !c.allowCircularDeps {
					issueCount += depsIssues
				} else if depsIssues > 0 && !c.quiet {
					fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Found %d circular dependency cycle(s) (allowed by --allow-circular-deps)\n", depsIssues)
				}
			} else if depsIssues > 0 && !c.quiet {
				// Within max-cycles threshold
				fmt.Fprintf(cmd.ErrOrStderr(), "✓ Found %d circular dependency cycle(s) (within allowed limit of %d)\n", depsIssues, c.maxCycles)
			}
		}
	}

	// Handle results
	if hasErrors {
		return fmt.Errorf("analysis failed with errors")
	}

	if issueCount > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "❌ Found %d quality issue(s)\n", issueCount)
		return fmt.Errorf("found %d quality issue(s)", issueCount)
	}

	if !c.quiet {
		fmt.Fprintf(cmd.ErrOrStderr(), "✅ Code quality check passed\n")
	}

	return nil
}

// determineEnabledAnalyses determines which analyses should run based on flags
func (c *CheckCommand) determineEnabledAnalyses() (skipComplexity bool, skipDeadCode bool, skipClones bool, skipDeps bool) {
	if len(c.selectAnalyses) > 0 {
		// If --select is used, only run selected analyses
		skipComplexity = !c.containsAnalysis("complexity")
		skipDeadCode = !c.containsAnalysis("deadcode")
		skipClones = !c.containsAnalysis("clones")
		skipDeps = !c.containsAnalysis("deps") && !c.containsAnalysis("circular")
	} else {
		// Otherwise use original behavior (backward compatible)
		skipComplexity = false    // Always run complexity
		skipDeadCode = false      // Always run dead code analysis
		skipClones = c.skipClones // Only skip clones if explicitly requested
		skipDeps = true           // Skip deps by default (opt-in via --select)
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
func (c *CheckCommand) getEnabledAnalyses(skipComplexity bool, skipDeadCode bool, skipClones bool, skipDeps bool) []string {
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
	}
	for _, analysis := range c.selectAnalyses {
		if !validAnalyses[strings.ToLower(analysis)] {
			return fmt.Errorf("invalid analysis type: %s. Valid options: complexity, deadcode, clones, deps", analysis)
		}
	}
	if len(c.selectAnalyses) == 0 {
		return fmt.Errorf("--select flag requires at least one analysis type")
	}

	return nil
}

// checkComplexity runs complexity analysis and returns issue count
func (c *CheckCommand) checkComplexity(cmd *cobra.Command, args []string) (int, error) {
	// Create request with check-specific settings
	request := &domain.ComplexityRequest{
		Paths:           args,
		OutputFormat:    domain.OutputFormatText,
		OutputWriter:    io.Discard,
		MinComplexity:   1,
		MaxComplexity:   0, // No filter
		LowThreshold:    5,
		MediumThreshold: 9,
		ShowDetails:     false,
		SortBy:          domain.SortByComplexity,
		Recursive:       true,
		IncludePatterns: []string{"**/*.py"},
		ExcludePatterns: []string{"__pycache__/*", "*.pyc"},
		ConfigPath:      c.configFile,
	}

	// Create use case with services
	configLoader := service.NewConfigurationLoader()
	fileReader := service.NewFileReader()
	complexityService := service.NewComplexityService()
	outputFormatter := service.NewOutputFormatter()

	useCase := app.NewComplexityUseCase(
		complexityService,
		fileReader,
		outputFormatter,
		configLoader,
	)

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Run analysis
	response, err := useCase.AnalyzeAndReturn(ctx, *request)
	if err != nil {
		return 0, err
	}

	// Count functions that exceed the maximum complexity threshold
	issueCount := 0
	for _, function := range response.Functions {
		if function.Metrics.Complexity > c.maxComplexity {
			issueCount++
			if !c.quiet {
				fmt.Fprintf(cmd.ErrOrStderr(), "%s:%d:%d: %s is too complex (%d > %d)\n",
					function.FilePath, function.StartLine, function.StartColumn+1, function.Name, function.Metrics.Complexity, c.maxComplexity)
			}
		}
	}

	return issueCount, nil
}

// checkDeadCode runs dead code analysis and returns issue count
func (c *CheckCommand) checkDeadCode(cmd *cobra.Command, args []string) (int, error) {
	// Create request with check-specific settings
	request := &domain.DeadCodeRequest{
		Paths:                     args,
		OutputFormat:              domain.OutputFormatText,
		OutputWriter:              io.Discard,
		ShowContext:               false,
		ContextLines:              0,
		MinSeverity:               domain.DeadCodeSeverityCritical,
		SortBy:                    domain.DeadCodeSortBySeverity,
		Recursive:                 true,
		IncludePatterns:           []string{"**/*.py"},
		ExcludePatterns:           []string{"__pycache__/*", "*.pyc"},
		IgnorePatterns:            []string{},
		DetectAfterReturn:         true,
		DetectAfterBreak:          true,
		DetectAfterContinue:       true,
		DetectAfterRaise:          true,
		DetectUnreachableBranches: true,
		ConfigPath:                c.configFile,
	}

	// Create use case with services
	configLoader := service.NewDeadCodeConfigurationLoader()
	fileReader := service.NewFileReader()
	deadCodeService := service.NewDeadCodeService()
	deadCodeFormatter := service.NewDeadCodeFormatter()

	useCase := app.NewDeadCodeUseCase(
		deadCodeService,
		fileReader,
		deadCodeFormatter,
		configLoader,
	)

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Run analysis
	response, err := useCase.AnalyzeAndReturn(ctx, *request)
	if err != nil {
		return 0, err
	}

	// Count and output critical dead code findings
	issueCount := 0
	for _, file := range response.Files {
		for _, function := range file.Functions {
			for _, finding := range function.Findings {
				if finding.Severity == domain.DeadCodeSeverityCritical {
					issueCount++
					if !c.quiet {
						fmt.Fprintf(cmd.ErrOrStderr(), "%s:%d:%d: %s (%s)\n",
							finding.Location.FilePath,
							finding.Location.StartLine,
							finding.Location.StartColumn+1,
							finding.Reason,
							finding.Severity)
					}
				}
			}
		}
	}

	return issueCount, nil
}

// checkClones runs clone detection and returns issue count
func (c *CheckCommand) checkClones(cmd *cobra.Command, args []string) (int, error) {
	// Create request with check-specific settings
	request := &domain.CloneRequest{
		Paths:               args,
		OutputFormat:        domain.OutputFormatText,
		OutputWriter:        io.Discard,
		MinLines:            5,
		MinNodes:            10,
		SimilarityThreshold: 0.8,
		MaxEditDistance:     50.0,
		IgnoreLiterals:      false,
		IgnoreIdentifiers:   false,
		Type1Threshold:      0.98,
		Type2Threshold:      0.95,
		Type3Threshold:      0.85,
		Type4Threshold:      0.70,
		ShowDetails:         false,
		ShowContent:         false,
		SortBy:              domain.SortBySimilarity,
		GroupClones:         true,
		MinSimilarity:       0.0,
		MaxSimilarity:       1.0,
		CloneTypes:          []domain.CloneType{domain.Type1Clone, domain.Type2Clone, domain.Type3Clone, domain.Type4Clone},
		Recursive:           true,
		IncludePatterns:     []string{"**/*.py"},
		ExcludePatterns:     []string{"__pycache__/*", "*.pyc"},
		ConfigPath:          c.configFile,
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return 0, fmt.Errorf("invalid clone request: %w", err)
	}

	// Create use case with services
	configLoader := service.NewCloneConfigurationLoader()
	fileReader := service.NewFileReader()
	cloneService := service.NewCloneService()
	cloneFormatter := service.NewCloneOutputFormatter()

	useCase := app.NewCloneUseCase(
		cloneService,
		fileReader,
		cloneFormatter,
		configLoader,
	)

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Run analysis
	response, err := useCase.ExecuteAndReturn(ctx, *request)
	if err != nil {
		return 0, err
	}

	// Output clone pairs above the similarity threshold
	issueCount := 0
	for _, pair := range response.ClonePairs {
		issueCount++
		if !c.quiet {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s:%d:%d: clone of %s:%d:%d (similarity: %.1f%%)\n",
				pair.Clone1.Location.FilePath,
				pair.Clone1.Location.StartLine,
				pair.Clone1.Location.StartCol+1,
				pair.Clone2.Location.FilePath,
				pair.Clone2.Location.StartLine,
				pair.Clone2.Location.StartCol+1,
				pair.Similarity*100)
		}
	}

	return issueCount, nil
}

// checkCircularDependencies runs circular dependency detection and returns issue count
func (c *CheckCommand) checkCircularDependencies(cmd *cobra.Command, args []string) (int, error) {
	// Determine project root (default to current directory if no args)
	projectRoot := "."
	if len(args) > 0 {
		projectRoot = args[0]
	}

	// Create module analyzer with check-optimized options
	opts := &analyzer.ModuleAnalysisOptions{
		ProjectRoot:       projectRoot,
		IncludePatterns:   []string{"**/*.py"},
		ExcludePatterns:   []string{"__pycache__", "*.pyc", ".venv", "venv"},
		IncludeStdLib:     false, // Exclude standard library for faster analysis
		IncludeThirdParty: false, // Exclude third-party for faster analysis
		FollowRelative:    true,  // Follow relative imports
	}

	moduleAnalyzer, err := analyzer.NewModuleAnalyzer(opts)
	if err != nil {
		return 0, fmt.Errorf("failed to create module analyzer: %w", err)
	}

	// Build dependency graph
	graph, err := moduleAnalyzer.AnalyzeProject()
	if err != nil {
		return 0, fmt.Errorf("failed to analyze dependencies: %w", err)
	}

	// Detect circular dependencies
	result := analyzer.DetectCircularDependencies(graph)

	if !result.HasCircularDependencies {
		return 0, nil
	}

	// Output circular dependencies in linter format
	for _, cycle := range result.CircularDependencies {
		if len(cycle.Modules) == 0 {
			continue
		}

		// Get the first module's file path
		firstModule := cycle.Modules[0]
		node := graph.Nodes[firstModule]
		if node == nil {
			continue
		}

		// Format: file:line:col: message
		cyclePath := strings.Join(cycle.Modules, " -> ")
		if !c.quiet {
			fmt.Fprintf(cmd.ErrOrStderr(), "%s:1:1: circular dependency detected: %s\n",
				node.FilePath, cyclePath)
		}
	}

	return result.TotalCycles, nil
}

// NewCheckCmd creates and returns the check cobra command
func NewCheckCmd() *cobra.Command {
	checkCommand := NewCheckCommand()
	return checkCommand.CreateCobraCommand()
}
