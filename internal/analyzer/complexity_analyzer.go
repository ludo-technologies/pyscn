package analyzer

import (
	"io"
	"os"

	"github.com/pyqol/pyqol/internal/config"
	"github.com/pyqol/pyqol/internal/reporter"
)

// ComplexityAnalyzer provides high-level complexity analysis functionality
type ComplexityAnalyzer struct {
	config   *config.Config
	reporter *reporter.ComplexityReporter
}

// NewComplexityAnalyzer creates a new complexity analyzer with configuration
func NewComplexityAnalyzer(cfg *config.Config, output io.Writer) *ComplexityAnalyzer {
	if output == nil {
		output = os.Stdout
	}
	
	return &ComplexityAnalyzer{
		config:   cfg,
		reporter: reporter.NewComplexityReporter(cfg, output),
	}
}

// NewComplexityAnalyzerWithDefaults creates a new analyzer with default configuration
func NewComplexityAnalyzerWithDefaults(output io.Writer) *ComplexityAnalyzer {
	cfg := config.DefaultConfig()
	return NewComplexityAnalyzer(cfg, output)
}

// AnalyzeAndReport performs complexity analysis and generates a formatted report
func (ca *ComplexityAnalyzer) AnalyzeAndReport(cfgs []*CFG) error {
	// Calculate complexity with configuration
	results := CalculateFileComplexityWithConfig(cfgs, &ca.config.Complexity)
	
	// Convert to interface slice for reporter
	interfaceResults := make([]reporter.ComplexityResult, len(results))
	for i, result := range results {
		interfaceResults[i] = result
	}
	
	// Generate and output report
	return ca.reporter.ReportComplexity(interfaceResults)
}

// AnalyzeFunction analyzes a single function and returns the result
func (ca *ComplexityAnalyzer) AnalyzeFunction(cfg *CFG) *ComplexityResult {
	return CalculateComplexityWithConfig(cfg, &ca.config.Complexity)
}

// AnalyzeFunctions analyzes multiple functions and returns filtered results
func (ca *ComplexityAnalyzer) AnalyzeFunctions(cfgs []*CFG) []*ComplexityResult {
	return CalculateFileComplexityWithConfig(cfgs, &ca.config.Complexity)
}

// CheckComplexityLimits checks if any functions exceed the configured maximum complexity
// Returns true if all functions are within limits, false otherwise
func (ca *ComplexityAnalyzer) CheckComplexityLimits(cfgs []*CFG) (bool, []*ComplexityResult) {
	results := CalculateFileComplexityWithConfig(cfgs, &ca.config.Complexity)
	
	var violations []*ComplexityResult
	for _, result := range results {
		if ca.config.Complexity.ExceedsMaxComplexity(result.Complexity) {
			violations = append(violations, result)
		}
	}
	
	return len(violations) == 0, violations
}

// GetConfiguration returns the current configuration
func (ca *ComplexityAnalyzer) GetConfiguration() *config.Config {
	return ca.config
}

// UpdateConfiguration updates the analyzer configuration
func (ca *ComplexityAnalyzer) UpdateConfiguration(cfg *config.Config) {
	ca.config = cfg
	// Create new reporter with updated config
	ca.reporter = reporter.NewComplexityReporter(cfg, ca.reporter.GetWriter())
}

// SetOutput changes the output destination for reports
func (ca *ComplexityAnalyzer) SetOutput(output io.Writer) {
	ca.reporter = reporter.NewComplexityReporter(ca.config, output)
}

// GenerateReport creates a comprehensive report without outputting it
func (ca *ComplexityAnalyzer) GenerateReport(cfgs []*CFG) *reporter.ComplexityReport {
	results := CalculateFileComplexityWithConfig(cfgs, &ca.config.Complexity)
	
	// Convert to interface slice for reporter
	interfaceResults := make([]reporter.ComplexityResult, len(results))
	for i, result := range results {
		interfaceResults[i] = result
	}
	
	return ca.reporter.GenerateReport(interfaceResults)
}