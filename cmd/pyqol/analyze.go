package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/spf13/cobra"
)

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
	minComplexity       int
	minSeverity         string
	cloneSimilarity     float64
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

	// Create context
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Run analyses concurrently
	var wg sync.WaitGroup
	
	// Channel to collect all errors
	errChan := make(chan error, 3)
	
	// Run complexity analysis
	if !c.skipComplexity {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.runComplexityAnalysis(cmd, args); err != nil {
				errChan <- fmt.Errorf("complexity analysis failed: %w", err)
			}
		}()
	}
	
	// Run dead code analysis
	if !c.skipDeadCode {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.runDeadCodeAnalysis(cmd, args); err != nil {
				errChan <- fmt.Errorf("dead code analysis failed: %w", err)
			}
		}()
	}
	
	// Run clone analysis
	if !c.skipClones {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.runCloneAnalysis(cmd, args); err != nil {
				errChan <- fmt.Errorf("clone detection failed: %w", err)
			}
		}()
	}
	
	// Wait for all analyses to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()
	
	// Collect errors
	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}
	
	// Report any errors
	if len(errors) > 0 {
		fmt.Fprintf(cmd.ErrOrStderr(), "Analysis completed with errors:\n")
		for _, err := range errors {
			fmt.Fprintf(cmd.ErrOrStderr(), "  • %v\n", err)
		}
		return fmt.Errorf("%d analysis(es) failed", len(errors))
	}
	
	// Print summary if verbose
	if c.verbose {
		analysisCount := 0
		if !c.skipComplexity { analysisCount++ }
		if !c.skipDeadCode { analysisCount++ }
		if !c.skipClones { analysisCount++ }
		fmt.Fprintf(cmd.ErrOrStderr(), "✅ Completed %d analysis type(s) successfully\n", analysisCount)
	}
	
	return nil
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