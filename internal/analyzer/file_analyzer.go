package analyzer

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/pyqol/pyqol/internal/config"
	"github.com/pyqol/pyqol/internal/parser"
	"github.com/pyqol/pyqol/internal/reporter"
)

// FileComplexityAnalyzer provides high-level file analysis capabilities
type FileComplexityAnalyzer struct {
	config   *config.Config
	reporter *reporter.ComplexityReporter
}

// NewFileComplexityAnalyzer creates a new file analyzer with configuration
func NewFileComplexityAnalyzer(cfg *config.Config, output io.Writer) (*FileComplexityAnalyzer, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	if output == nil {
		output = os.Stdout
	}

	reporter, err := reporter.NewComplexityReporter(cfg, output)
	if err != nil {
		return nil, fmt.Errorf("failed to create reporter: %w", err)
	}

	return &FileComplexityAnalyzer{
		config:   cfg,
		reporter: reporter,
	}, nil
}

// AnalyzeFile analyzes a single Python file and outputs complexity results
func (fca *FileComplexityAnalyzer) AnalyzeFile(filename string) error {
	// Read file content
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	// Parse the Python code
	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to parse Python code in %s: %w", filename, err)
	}

	// Build CFGs for all functions
	builder := NewCFGBuilder()
	cfgs, err := builder.BuildAll(result.AST)
	if err != nil {
		return fmt.Errorf("failed to build control flow graphs for %s: %w", filename, err)
	}

	if len(cfgs) == 0 {
		return fmt.Errorf("no functions found in %s", filename)
	}

	// Convert map to slice for the complexity analyzer
	cfgList := make([]*CFG, 0, len(cfgs))
	for _, cfg := range cfgs {
		cfgList = append(cfgList, cfg)
	}

	// Calculate complexity results
	results := CalculateFileComplexityWithConfig(cfgList, &fca.config.Complexity)

	// Convert to reporter interface
	interfaceResults := make([]reporter.ComplexityResult, len(results))
	for i, result := range results {
		interfaceResults[i] = result
	}

	// Generate and output report
	return fca.reporter.ReportComplexityWithFileCount(interfaceResults, 1)
}

// AnalyzeFiles analyzes multiple Python files and outputs combined complexity results
func (fca *FileComplexityAnalyzer) AnalyzeFiles(filenames []string) error {
	var allResults []*ComplexityResult

	for _, filename := range filenames {
		// Read and parse file
		content, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", filename, err)
		}

		p := parser.New()
		ctx := context.Background()
		result, err := p.Parse(ctx, content)
		if err != nil {
			return fmt.Errorf("failed to parse Python code in %s: %w", filename, err)
		}

		// Build CFGs
		builder := NewCFGBuilder()
		cfgs, err := builder.BuildAll(result.AST)
		if err != nil {
			return fmt.Errorf("failed to build control flow graphs for %s: %w", filename, err)
		}

		// Convert map to slice
		cfgList := make([]*CFG, 0, len(cfgs))
		for _, cfg := range cfgs {
			cfgList = append(cfgList, cfg)
		}

		// Calculate complexity for this file
		fileResults := CalculateFileComplexityWithConfig(cfgList, &fca.config.Complexity)
		allResults = append(allResults, fileResults...)
	}

	if len(allResults) == 0 {
		return fmt.Errorf("no functions found in any of the files")
	}

	// Convert to reporter interface
	interfaceResults := make([]reporter.ComplexityResult, len(allResults))
	for i, result := range allResults {
		interfaceResults[i] = result
	}

	// Generate and output report
	return fca.reporter.ReportComplexityWithFileCount(interfaceResults, len(filenames))
}
