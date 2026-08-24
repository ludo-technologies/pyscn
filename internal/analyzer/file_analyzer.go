package analyzer

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/ludo-technologies/pyscn/internal/parser"
	"github.com/ludo-technologies/pyscn/internal/reporter"
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

	// Build a typed CFG for each Python execution scope.
	builder := NewCFGBuilder()
	cfgs, err := builder.BuildAll(result.AST)
	if err != nil {
		return fmt.Errorf("failed to build control flow graphs for %s: %w", filename, err)
	}

	if len(cfgs) == 0 {
		return fmt.Errorf("no execution scopes found in %s", filename)
	}

	results := calculateScopedReporterResults(cfgs, &fca.config.Complexity, filename)

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
	var allResults []scopedReporterResult

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

		// Build typed CFGs for all execution scopes.
		builder := NewCFGBuilder()
		cfgs, err := builder.BuildAll(result.AST)
		if err != nil {
			return fmt.Errorf("failed to build control flow graphs for %s: %w", filename, err)
		}

		fileResults := calculateScopedReporterResults(cfgs, &fca.config.Complexity, filename)
		allResults = append(allResults, fileResults...)
	}

	if len(allResults) == 0 {
		return fmt.Errorf("no execution scopes found in any of the files")
	}

	// Convert to reporter interface
	interfaceResults := make([]reporter.ComplexityResult, len(allResults))
	for i, result := range allResults {
		interfaceResults[i] = result
	}

	// Generate and output report
	return fca.reporter.ReportComplexityWithFileCount(interfaceResults, len(filenames))
}

type scopedReporterResult struct {
	*ComplexityResult
	scopeKind domain.AnalysisScopeKind
	filePath  string
}

func (r scopedReporterResult) GetScopeKind() domain.AnalysisScopeKind {
	return r.scopeKind
}

func (r scopedReporterResult) GetSourceLocation() reporter.ComplexitySourceLocation {
	return reporter.ComplexitySourceLocation{
		FilePath:    r.filePath,
		StartLine:   r.StartLine,
		StartColumn: r.StartCol,
		EndLine:     r.EndLine,
	}
}

func calculateScopedReporterResults(cfgs ControlFlowGraphs, complexityConfig *config.ComplexityConfig, filePath string) []scopedReporterResult {
	results := make([]scopedReporterResult, 0, len(cfgs))
	for _, scopedCFG := range cfgs {
		result := CalculateComplexityWithConfig(scopedCFG.Graph, complexityConfig)
		if complexityConfig.ShouldReport(result.Complexity) {
			results = append(results, scopedReporterResult{
				ComplexityResult: result,
				scopeKind:        scopedCFG.Scope.Kind,
				filePath:         filePath,
			})
		}
	}
	return results
}
