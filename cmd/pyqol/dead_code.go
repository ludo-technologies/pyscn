package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pyqol/pyqol/app"
	"github.com/pyqol/pyqol/domain"
	"github.com/pyqol/pyqol/service"
	"github.com/spf13/cobra"
)

// DeadCodeCommand represents the dead code command
type DeadCodeCommand struct {
	// Output format flags (only one should be true)
	html      bool
	json      bool
	csv       bool
	yaml      bool
	noOpen    bool
	
	// Analysis flags
	minSeverity  string
	sortBy       string
	showContext  bool
	contextLines int
	configFile   string
	verbose      bool

	// Dead code detection options
	detectAfterReturn         bool
	detectAfterBreak          bool
	detectAfterContinue       bool
	detectAfterRaise          bool
	detectUnreachableBranches bool
}

// NewDeadCodeCommand creates a new dead code command
func NewDeadCodeCommand() *DeadCodeCommand {
	return &DeadCodeCommand{
		html:                      false,
		json:                      false,
		csv:                       false,
		yaml:                      false,
		noOpen:                    false,
		minSeverity:               "warning",
		sortBy:                    "severity",
		showContext:               false,
		contextLines:              3,
		configFile:                "",
		verbose:                   false,
		detectAfterReturn:         true,
		detectAfterBreak:          true,
		detectAfterContinue:       true,
		detectAfterRaise:          true,
		detectUnreachableBranches: true,
	}
}

// CreateCobraCommand creates the cobra command for dead code analysis
func (c *DeadCodeCommand) CreateCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deadcode [files...]",
		Short: "Detect dead code in Python files",
		Long: `Detect dead code in Python files using Control Flow Graph (CFG) analysis.

Dead code refers to parts of a program that are executed but whose result is never used, 
or that can never be executed. This tool identifies various types of dead code:

Severity levels:
  • Critical:  Code that is definitely unreachable (after return, break, etc.)
  • Warning:   Code that is likely unreachable (unreachable branches)
  • Info:      Potential optimization opportunities

Detection types:
  • Code after return statements
  • Code after break/continue statements  
  • Code after raise statements
  • Unreachable conditional branches
  • Code after infinite loops

Examples:
  pyqol deadcode myfile.py
  pyqol deadcode src/
  pyqol deadcode --format json --min-severity critical src/
  pyqol deadcode --show-context --context-lines 5 myfile.py`,
		Args: cobra.MinimumNArgs(1),
		RunE: c.runDeadCodeAnalysis,
	}

	// Add output format flags
	cmd.Flags().BoolVar(&c.html, "html", false, "Generate HTML report file")
	cmd.Flags().BoolVar(&c.json, "json", false, "Generate JSON report file")
	cmd.Flags().BoolVar(&c.csv, "csv", false, "Generate CSV report file")
	cmd.Flags().BoolVar(&c.yaml, "yaml", false, "Generate YAML report file")
	cmd.Flags().BoolVar(&c.noOpen, "no-open", false, "Don't auto-open HTML in browser")
	
	// Add analysis flags
	cmd.Flags().StringVar(&c.minSeverity, "min-severity", "warning", "Minimum severity to report (critical, warning, info)")
	cmd.Flags().StringVar(&c.sortBy, "sort", "severity", "Sort results by (severity, line, file, function)")
	cmd.Flags().BoolVar(&c.showContext, "show-context", false, "Show surrounding code context")
	cmd.Flags().IntVar(&c.contextLines, "context-lines", 3, "Number of context lines to show")
	cmd.Flags().StringVarP(&c.configFile, "config", "c", "", "Configuration file path")

	// Dead code detection options
	cmd.Flags().BoolVar(&c.detectAfterReturn, "detect-after-return", true, "Detect code after return statements")
	cmd.Flags().BoolVar(&c.detectAfterBreak, "detect-after-break", true, "Detect code after break statements")
	cmd.Flags().BoolVar(&c.detectAfterContinue, "detect-after-continue", true, "Detect code after continue statements")
	cmd.Flags().BoolVar(&c.detectAfterRaise, "detect-after-raise", true, "Detect code after raise statements")
	cmd.Flags().BoolVar(&c.detectUnreachableBranches, "detect-unreachable-branches", true, "Detect unreachable conditional branches")

	return cmd
}

// runDeadCodeAnalysis executes the dead code analysis
func (c *DeadCodeCommand) runDeadCodeAnalysis(cmd *cobra.Command, args []string) error {
	// Get verbose flag from parent command
	if cmd.Parent() != nil {
		c.verbose, _ = cmd.Parent().Flags().GetBool("verbose")
	}

	// Build the domain request from CLI flags
	request, err := c.buildDeadCodeRequest(cmd, args)
	if err != nil {
		return fmt.Errorf("invalid command arguments: %w", err)
	}

	// Create the use case with dependencies
	useCase, err := c.createDeadCodeUseCase(cmd)
	if err != nil {
		return fmt.Errorf("failed to initialize dead code analyzer: %w", err)
	}

	// Execute the analysis
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	if err := useCase.Execute(ctx, request); err != nil {
		return c.handleAnalysisError(err)
	}

	return nil
}

// determineOutputFormat determines the output format based on flags
func (c *DeadCodeCommand) determineOutputFormat() (domain.OutputFormat, string, error) {
    resolver := service.NewOutputFormatResolver()
    return resolver.Determine(c.html, c.json, c.csv, c.yaml)
}

// buildDeadCodeRequest creates a domain request from CLI flags
func (c *DeadCodeCommand) buildDeadCodeRequest(cmd *cobra.Command, args []string) (domain.DeadCodeRequest, error) {
	// Determine output format from flags
	outputFormat, extension, err := c.determineOutputFormat()
	if err != nil {
		return domain.DeadCodeRequest{}, err
	}

	// Convert severity level
	minSeverity, err := c.parseSeverityLevel(c.minSeverity)
	if err != nil {
		return domain.DeadCodeRequest{}, err
	}

	// Convert sort criteria
	sortBy, err := c.parseSortCriteria(c.sortBy)
	if err != nil {
		return domain.DeadCodeRequest{}, err
	}

	// Validate context lines
	if err := c.validateContextLines(); err != nil {
		return domain.DeadCodeRequest{}, err
	}

	// Expand any directory paths and validate files
	paths, err := c.expandAndValidatePaths(args)
	if err != nil {
		return domain.DeadCodeRequest{}, err
	}

	// Determine output destination
	var outputWriter io.Writer
	var outputPath string
	
	if outputFormat == domain.OutputFormatText {
		// Text format goes to stdout
		outputWriter = cmd.OutOrStdout()
	} else {
		// Other formats generate a file
		// Use first path as target for config discovery
		targetPath := getTargetPathFromArgs(args)
		var err error
		outputPath, err = generateOutputFilePath("deadcode", extension, targetPath)
		if err != nil {
			return domain.DeadCodeRequest{}, fmt.Errorf("failed to generate output path: %w", err)
		}
	}

	return domain.DeadCodeRequest{
		Paths:           paths,
		OutputFormat:    outputFormat,
		OutputWriter:    outputWriter,
		OutputPath:      outputPath,
		NoOpen:          c.noOpen,
		ShowContext:     c.showContext,
		ContextLines:    c.contextLines,
		MinSeverity:     minSeverity,
		SortBy:          sortBy,
		ConfigPath:      c.configFile,
		Recursive:       true, // Always recursive for directories
		IncludePatterns: []string{"*.py", "*.pyi"},
		ExcludePatterns: []string{"test_*.py", "*_test.py"},
		IgnorePatterns:  []string{},

		// Dead code detection options
		DetectAfterReturn:         c.detectAfterReturn,
		DetectAfterBreak:          c.detectAfterBreak,
		DetectAfterContinue:       c.detectAfterContinue,
		DetectAfterRaise:          c.detectAfterRaise,
		DetectUnreachableBranches: c.detectUnreachableBranches,
	}, nil
}

// createDeadCodeUseCase creates the use case with all dependencies
func (c *DeadCodeCommand) createDeadCodeUseCase(cmd *cobra.Command) (*app.DeadCodeUseCase, error) {
	// Track which flags were explicitly set by the user
	explicitFlags := GetExplicitFlags(cmd)

	// Create services
	fileReader := service.NewFileReader()
	formatter := service.NewDeadCodeFormatter()
	configLoader := service.NewDeadCodeConfigurationLoaderWithFlags(explicitFlags)

	deadCodeService := service.NewDeadCodeService()

    // Build use case
    useCase, err := app.NewDeadCodeUseCaseBuilder().
        WithService(deadCodeService).
        WithFileReader(fileReader).
        WithFormatter(formatter).
        WithConfigLoader(configLoader).
        WithOutputWriter(service.NewFileOutputWriter(cmd.ErrOrStderr())).
        Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build use case: %w", err)
	}

	return useCase, nil
}

// Helper methods for parsing and validation

func (c *DeadCodeCommand) parseSeverityLevel(severity string) (domain.DeadCodeSeverity, error) {
	switch strings.ToLower(severity) {
	case "critical":
		return domain.DeadCodeSeverityCritical, nil
	case "warning":
		return domain.DeadCodeSeverityWarning, nil
	case "info":
		return domain.DeadCodeSeverityInfo, nil
	default:
		return "", fmt.Errorf("unsupported severity level: %s (supported: critical, warning, info)", severity)
	}
}

func (c *DeadCodeCommand) parseSortCriteria(sort string) (domain.DeadCodeSortCriteria, error) {
	switch strings.ToLower(sort) {
	case "severity":
		return domain.DeadCodeSortBySeverity, nil
	case "line":
		return domain.DeadCodeSortByLine, nil
	case "file":
		return domain.DeadCodeSortByFile, nil
	case "function":
		return domain.DeadCodeSortByFunction, nil
	default:
		return "", fmt.Errorf("unsupported sort criteria: %s (supported: severity, line, file, function)", sort)
	}
}

func (c *DeadCodeCommand) validateContextLines() error {
	if c.contextLines < 0 {
		return fmt.Errorf("context lines cannot be negative")
	}
	if c.contextLines > 20 {
		return fmt.Errorf("context lines cannot exceed 20 (got %d)", c.contextLines)
	}
	return nil
}

func (c *DeadCodeCommand) expandAndValidatePaths(args []string) ([]string, error) {
	var paths []string

	for _, arg := range args {
		// Expand the path
		expanded, err := filepath.Abs(arg)
		if err != nil {
			return nil, fmt.Errorf("invalid path %s: %w", arg, err)
		}

		// Check if path exists
		if _, err := os.Stat(expanded); err != nil {
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("path does not exist: %s", arg)
			}
			return nil, fmt.Errorf("cannot access path %s: %w", arg, err)
		}

		paths = append(paths, expanded)
	}

	return paths, nil
}

func (c *DeadCodeCommand) handleAnalysisError(err error) error {
	// Convert domain errors to user-friendly messages
	if domainErr, ok := err.(domain.DomainError); ok {
		switch domainErr.Code {
		case domain.ErrCodeFileNotFound:
			return fmt.Errorf("file not found: %s", domainErr.Message)
		case domain.ErrCodeInvalidInput:
			return fmt.Errorf("invalid input: %s", domainErr.Message)
		case domain.ErrCodeParseError:
			return fmt.Errorf("parsing failed: %s", domainErr.Message)
		case domain.ErrCodeAnalysisError:
			return fmt.Errorf("analysis failed: %s", domainErr.Message)
		case domain.ErrCodeConfigError:
			return fmt.Errorf("configuration error: %s", domainErr.Message)
		case domain.ErrCodeOutputError:
			return fmt.Errorf("output error: %s", domainErr.Message)
		case domain.ErrCodeUnsupportedFormat:
			return fmt.Errorf("unsupported format: %s", domainErr.Message)
		default:
			return fmt.Errorf("analysis error: %s", domainErr.Message)
		}
	}

	// Return original error if not a domain error
	return err
}

// GetUsageExamples returns example usage commands
func (c *DeadCodeCommand) GetUsageExamples() []string {
	return []string{
		"pyqol deadcode myfile.py",
		"pyqol deadcode src/",
		"pyqol deadcode --format json src/",
		"pyqol deadcode --min-severity critical --show-context src/",
		"pyqol deadcode --sort line --context-lines 5 myfile.py",
		"pyqol deadcode --config .pyqol.yaml src/",
	}
}

// GetSupportedFormats returns supported output formats
func (c *DeadCodeCommand) GetSupportedFormats() []string {
	return []string{"text", "json", "yaml", "csv"}
}

// GetSupportedSeverities returns supported severity levels
func (c *DeadCodeCommand) GetSupportedSeverities() []string {
	return []string{"critical", "warning", "info"}
}

// GetSupportedSortCriteria returns supported sort criteria
func (c *DeadCodeCommand) GetSupportedSortCriteria() []string {
	return []string{"severity", "line", "file", "function"}
}

// NewDeadCodeCmd creates and returns the dead code cobra command
func NewDeadCodeCmd() *cobra.Command {
	deadCodeCommand := NewDeadCodeCommand()
	return deadCodeCommand.CreateCobraCommand()
}
