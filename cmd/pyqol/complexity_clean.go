package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pyqol/pyqol/app"
	"github.com/pyqol/pyqol/domain"
	"github.com/pyqol/pyqol/service"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ComplexityCommand represents the complexity command
type ComplexityCommand struct {
	// Command line flags
	outputFormat    string
	minComplexity   int
	maxComplexity   int
	sortBy          string
	showDetails     bool
	configFile      string
	lowThreshold    int
	mediumThreshold int
	verbose         bool
}

// NewComplexityCommand creates a new complexity command
func NewComplexityCommand() *ComplexityCommand {
	return &ComplexityCommand{
		outputFormat:    "text",
		minComplexity:   1,
		maxComplexity:   0,
		sortBy:          "complexity",
		showDetails:     false,
		configFile:      "",
		lowThreshold:    9,
		mediumThreshold: 19,
		verbose:         false,
	}
}

// CreateCobraCommand creates the cobra command for complexity analysis
func (c *ComplexityCommand) CreateCobraCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "complexity [files...]",
		Short: "Analyze cyclomatic complexity of Python files",
		Long: `Analyze the cyclomatic complexity of Python files using Control Flow Graph (CFG) analysis.

McCabe cyclomatic complexity measures the number of linearly independent paths through a program's source code.
Lower complexity indicates easier to understand and maintain code.

Risk levels:
  • Low (1-9):     Easy to understand and maintain
  • Medium (10-19): Moderate complexity, consider refactoring
  • High (20+):     Complex, should be refactored

Examples:
  pyqol complexity myfile.py
  pyqol complexity src/
  pyqol complexity --format json src/`,
		Args: cobra.MinimumNArgs(1),
		RunE: c.runComplexityAnalysis,
	}

	// Add flags
	cmd.Flags().StringVarP(&c.outputFormat, "format", "f", "text", "Output format (text, json, yaml, csv)")
	cmd.Flags().IntVar(&c.minComplexity, "min", 1, "Minimum complexity to report")
	cmd.Flags().IntVar(&c.maxComplexity, "max", 0, "Maximum complexity limit (0 = no limit)")
	cmd.Flags().StringVar(&c.sortBy, "sort", "complexity", "Sort results by (name, complexity, risk)")
	cmd.Flags().BoolVar(&c.showDetails, "details", false, "Show detailed complexity breakdown")
	cmd.Flags().StringVarP(&c.configFile, "config", "c", "", "Configuration file path")
	cmd.Flags().IntVar(&c.lowThreshold, "low-threshold", 9, "Low complexity threshold")
	cmd.Flags().IntVar(&c.mediumThreshold, "medium-threshold", 19, "Medium complexity threshold")

	return cmd
}

// runComplexityAnalysis executes the complexity analysis
func (c *ComplexityCommand) runComplexityAnalysis(cmd *cobra.Command, args []string) error {
	// Get verbose flag from parent command
	if cmd.Parent() != nil {
		c.verbose, _ = cmd.Parent().Flags().GetBool("verbose")
	}

	// Build the domain request from CLI flags
	request, err := c.buildComplexityRequest(cmd, args)
	if err != nil {
		return fmt.Errorf("invalid command arguments: %w", err)
	}

	// Create the use case with dependencies
	useCase, err := c.createComplexityUseCase(cmd)
	if err != nil {
		return fmt.Errorf("failed to initialize complexity analyzer: %w", err)
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

// buildComplexityRequest creates a domain request from CLI flags
func (c *ComplexityCommand) buildComplexityRequest(cmd *cobra.Command, args []string) (domain.ComplexityRequest, error) {
	// Convert output format
	outputFormat, err := c.parseOutputFormat(c.outputFormat)
	if err != nil {
		return domain.ComplexityRequest{}, err
	}

	// Convert sort criteria
	sortBy, err := c.parseSortCriteria(c.sortBy)
	if err != nil {
		return domain.ComplexityRequest{}, err
	}

	// Validate thresholds
	if err := c.validateThresholds(); err != nil {
		return domain.ComplexityRequest{}, err
	}

	// Expand any directory paths and validate files
	paths, err := c.expandAndValidatePaths(args)
	if err != nil {
		return domain.ComplexityRequest{}, err
	}

	return domain.ComplexityRequest{
		Paths:           paths,
		OutputFormat:    outputFormat,
		OutputWriter:    cmd.OutOrStdout(),
		ShowDetails:     c.showDetails,
		MinComplexity:   c.minComplexity,
		MaxComplexity:   c.maxComplexity,
		SortBy:          sortBy,
		LowThreshold:    c.lowThreshold,
		MediumThreshold: c.mediumThreshold,
		ConfigPath:      c.configFile,
		Recursive:       true, // Always recursive for directories
		IncludePatterns: []string{"*.py", "*.pyi"},
		ExcludePatterns: []string{"test_*.py", "*_test.py"},
	}, nil
}

// createComplexityUseCase creates the use case with all dependencies
func (c *ComplexityCommand) createComplexityUseCase(cmd *cobra.Command) (*app.ComplexityUseCase, error) {
	// Track which flags were explicitly set by the user
	explicitFlags := make(map[string]bool)
	cmd.Flags().Visit(func(f *pflag.Flag) {
		explicitFlags[f.Name] = true
	})

	// Create services
	fileReader := service.NewFileReader()
	formatter := service.NewOutputFormatter()
	configLoader := service.NewConfigurationLoaderWithFlags(explicitFlags)

	// Create progress reporter
	progress := service.CreateProgressReporter(cmd.ErrOrStderr(), 0, c.verbose)
	complexityService := service.NewComplexityService(progress)

	// Build use case
	useCase, err := app.NewComplexityUseCaseBuilder().
		WithService(complexityService).
		WithFileReader(fileReader).
		WithFormatter(formatter).
		WithConfigLoader(configLoader).
		WithProgress(progress).
		Build()

	if err != nil {
		return nil, fmt.Errorf("failed to build use case: %w", err)
	}

	return useCase, nil
}

// Helper methods for parsing and validation

func (c *ComplexityCommand) parseOutputFormat(format string) (domain.OutputFormat, error) {
	switch strings.ToLower(format) {
	case "text":
		return domain.OutputFormatText, nil
	case "json":
		return domain.OutputFormatJSON, nil
	case "yaml", "yml":
		return domain.OutputFormatYAML, nil
	case "csv":
		return domain.OutputFormatCSV, nil
	default:
		return "", fmt.Errorf("unsupported output format: %s (supported: text, json, yaml, csv)", format)
	}
}

func (c *ComplexityCommand) parseSortCriteria(sort string) (domain.SortCriteria, error) {
	switch strings.ToLower(sort) {
	case "complexity":
		return domain.SortByComplexity, nil
	case "name":
		return domain.SortByName, nil
	case "risk":
		return domain.SortByRisk, nil
	default:
		return "", fmt.Errorf("unsupported sort criteria: %s (supported: complexity, name, risk)", sort)
	}
}

func (c *ComplexityCommand) validateThresholds() error {
	if c.lowThreshold <= 0 {
		return fmt.Errorf("low threshold must be positive")
	}
	if c.mediumThreshold <= c.lowThreshold {
		return fmt.Errorf("medium threshold (%d) must be greater than low threshold (%d)", c.mediumThreshold, c.lowThreshold)
	}
	if c.maxComplexity > 0 && c.maxComplexity <= c.mediumThreshold {
		return fmt.Errorf("max complexity (%d) must be greater than medium threshold (%d) or 0 for no limit", c.maxComplexity, c.mediumThreshold)
	}
	if c.minComplexity < 0 {
		return fmt.Errorf("minimum complexity cannot be negative")
	}
	if c.maxComplexity > 0 && c.minComplexity > c.maxComplexity {
		return fmt.Errorf("minimum complexity (%d) cannot be greater than maximum complexity (%d)", c.minComplexity, c.maxComplexity)
	}
	return nil
}

func (c *ComplexityCommand) expandAndValidatePaths(args []string) ([]string, error) {
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

func (c *ComplexityCommand) handleAnalysisError(err error) error {
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

// Global complexity command instance for the cobra command
var complexityCommand = NewComplexityCommand()

// complexityCmd is the cobra command that will be added to the root command
var complexityCmd = complexityCommand.CreateCobraCommand()
