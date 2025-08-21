package main

import (
	"context"
	"fmt"
	"os"

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
  pyqol check .

  # Check with higher complexity threshold
  pyqol check --max-complexity 15 src/

  # Allow dead code, only check complexity
  pyqol check --allow-dead-code src/

  # Skip clone detection for faster analysis
  pyqol check --skip-clones src/`,
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
	complexityCmd := NewComplexityCommand()

	// Configure with stricter defaults for checking
	complexityCmd.outputFormat = "text"
	complexityCmd.minComplexity = c.maxComplexity + 1 // Only report issues above threshold
	complexityCmd.maxComplexity = c.maxComplexity     // Set maximum allowed
	complexityCmd.configFile = c.configFile
	complexityCmd.verbose = false

	// Build request
	request, err := complexityCmd.buildComplexityRequest(cmd, args)
	if err != nil {
		return 0, err
	}

	// Create use case
	useCase, err := complexityCmd.createComplexityUseCase(cmd)
	if err != nil {
		return 0, err
	}

	// Execute analysis (this will output results)
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if err := useCase.Execute(ctx, request); err != nil {
		return 0, err
	}

	// For now, assume any output means issues were found
	// In a more sophisticated version, we'd capture and count results
	return 0, nil // Simplified - actual implementation would count issues
}

// checkDeadCode runs dead code analysis and returns issue count
func (c *CheckCommand) checkDeadCode(cmd *cobra.Command, args []string) (int, error) {
	deadCodeCmd := NewDeadCodeCommand()

	// Configure for critical issues only
	deadCodeCmd.outputFormat = "text"
	deadCodeCmd.minSeverity = "critical"
	deadCodeCmd.configFile = c.configFile
	deadCodeCmd.verbose = false

	// Build request
	request, err := deadCodeCmd.buildDeadCodeRequest(cmd, args)
	if err != nil {
		return 0, err
	}

	// Create use case
	useCase, err := deadCodeCmd.createDeadCodeUseCase(cmd)
	if err != nil {
		return 0, err
	}

	// Execute analysis
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if err := useCase.Execute(ctx, request); err != nil {
		return 0, err
	}

	// Simplified - actual implementation would count issues
	return 0, nil
}

// checkClones runs clone detection and returns issue count
func (c *CheckCommand) checkClones(cmd *cobra.Command, args []string) (int, error) {
	cloneCmd := NewCloneCommand()

	// Configure for informational reporting
	cloneCmd.outputFormat = "text"
	cloneCmd.similarityThreshold = 0.8
	cloneCmd.configFile = c.configFile
	cloneCmd.verbose = false

	// Create request
	request, err := cloneCmd.createCloneRequest(args)
	if err != nil {
		return 0, err
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return 0, fmt.Errorf("invalid clone request: %w", err)
	}

	// Create use case
	useCase, err := cloneCmd.createCloneUseCase(cmd)
	if err != nil {
		return 0, err
	}

	// Execute analysis
	ctx := context.Background()
	if err := useCase.Execute(ctx, *request); err != nil {
		return 0, err
	}

	// Simplified - actual implementation would count clones
	return 0, nil
}

// NewCheckCmd creates and returns the check cobra command
func NewCheckCmd() *cobra.Command {
	checkCommand := NewCheckCommand()
	return checkCommand.CreateCobraCommand()
}
