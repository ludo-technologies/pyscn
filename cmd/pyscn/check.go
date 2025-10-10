package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/spf13/cobra"
)

// CheckCommand represents a quick check command with sensible defaults
type CheckCommand struct {
	// Configuration
	configFile string
	quiet      bool

	// Quick override flags
	maxComplexity int
	allowDeadCode bool
	skipClones    bool
}

// NewCheckCommand creates a new check command
func NewCheckCommand() *CheckCommand {
	return &CheckCommand{
		configFile:    "",
		quiet:         false,
		maxComplexity: 10,    // Fail if complexity > 10
		allowDeadCode: false, // Fail on any dead code
		skipClones:    false,
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

Exit codes:
• 0: No issues found
• 1: Quality issues found (see output for details)
• 2: Analysis failed (invalid input, missing files, etc.)

The check command is designed to be fast and CI-friendly with minimal output
unless issues are found.

Examples:
  # Check current directory (typical CI usage)
  pyscn check .

  # Check with higher complexity threshold
  pyscn check --max-complexity 15 src/

  # Allow dead code, only check complexity
  pyscn check --allow-dead-code src/

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

	return cmd
}

// runCheck executes the quick check analysis
func (c *CheckCommand) runCheck(cmd *cobra.Command, args []string) error {
	// Default to current directory if no args
	if len(args) == 0 {
		args = []string{"."}
	}

	// Count issues found
	var issueCount int
	var hasErrors bool

	if !c.quiet {
		fmt.Fprintf(cmd.ErrOrStderr(), "🔍 Running quality check...\n")
	}

	// Run complexity check
	complexityIssues, err := c.checkComplexity(cmd, args)
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "❌ Complexity analysis failed: %v\n", err)
		hasErrors = true
	} else {
		issueCount += complexityIssues
	}

	// Run dead code check (if not explicitly allowed)
	if !c.allowDeadCode {
		deadCodeIssues, err := c.checkDeadCode(cmd, args)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "❌ Dead code analysis failed: %v\n", err)
			hasErrors = true
		} else {
			issueCount += deadCodeIssues
		}
	}

	// Run clone check (if not skipped)
	if !c.skipClones {
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

	// Handle results
	if hasErrors {
		return fmt.Errorf("analysis failed with errors")
	}

	if issueCount > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "❌ Found %d quality issue(s)\n", issueCount)
		os.Exit(1) // Exit with code 1 to indicate issues found
	}

	if !c.quiet {
		fmt.Fprintf(cmd.ErrOrStderr(), "✅ Code quality check passed\n")
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
				fmt.Fprintf(cmd.ErrOrStderr(), "❌ High complexity in %s:%s (complexity: %d > %d)\n",
					function.FilePath, function.Name, function.Metrics.Complexity, c.maxComplexity)
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

	// Count critical dead code findings
	issueCount := response.Summary.CriticalFindings
	if issueCount > 0 && !c.quiet {
		fmt.Fprintf(cmd.ErrOrStderr(), "❌ Found %d critical dead code issue(s)\n", issueCount)
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

	// Count clone pairs above the similarity threshold
	issueCount := len(response.ClonePairs)
	if issueCount > 0 && !c.quiet {
		fmt.Fprintf(cmd.ErrOrStderr(), "⚠️  Found %d code clone pair(s) (similarity > %.1f)\n",
			issueCount, request.SimilarityThreshold)
	}

	return issueCount, nil
}

// NewCheckCmd creates and returns the check cobra command
func NewCheckCmd() *cobra.Command {
	checkCommand := NewCheckCommand()
	return checkCommand.CreateCobraCommand()
}
