package reporter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/pyqol/pyqol/internal/config"
	"gopkg.in/yaml.v3"
)

// ComplexityResult represents complexity metrics for a function (interface to avoid import cycle)
type ComplexityResult interface {
	GetComplexity() int
	GetFunctionName() string
	GetRiskLevel() string
	GetNodes() int
	GetEdges() int
	GetIfStatements() int
	GetLoopStatements() int
	GetExceptionHandlers() int
	GetSwitchCases() int
}

// SerializableComplexityResult is a concrete type for JSON/YAML serialization
type SerializableComplexityResult struct {
	Complexity        int    `json:"complexity" yaml:"complexity"`
	FunctionName      string `json:"function_name" yaml:"function_name"`
	RiskLevel         string `json:"risk_level" yaml:"risk_level"`
	Nodes             int    `json:"nodes" yaml:"nodes"`
	Edges             int    `json:"edges" yaml:"edges"`
	IfStatements      int    `json:"if_statements" yaml:"if_statements"`
	LoopStatements    int    `json:"loop_statements" yaml:"loop_statements"`
	ExceptionHandlers int    `json:"exception_handlers" yaml:"exception_handlers"`
	SwitchCases       int    `json:"switch_cases" yaml:"switch_cases"`
}

// ComplexityReport represents a complete complexity analysis report
type ComplexityReport struct {
	// Summary contains aggregate statistics
	Summary ReportSummary `json:"summary" yaml:"summary"`
	
	// Results contains individual function complexity results
	Results []SerializableComplexityResult `json:"results" yaml:"results"`
	
	// Metadata contains report generation information
	Metadata ReportMetadata `json:"metadata" yaml:"metadata"`
	
	// Warnings contains threshold violations and issues
	Warnings []ReportWarning `json:"warnings,omitempty" yaml:"warnings,omitempty"`
}

// ReportSummary contains aggregate statistics for the complexity report
type ReportSummary struct {
	// TotalFunctions is the total number of functions analyzed
	TotalFunctions int `json:"total_functions" yaml:"total_functions"`
	
	// AverageComplexity is the mean complexity across all functions
	AverageComplexity float64 `json:"average_complexity" yaml:"average_complexity"`
	
	// MaxComplexity is the highest complexity found
	MaxComplexity int `json:"max_complexity" yaml:"max_complexity"`
	
	// MinComplexity is the lowest complexity found  
	MinComplexity int `json:"min_complexity" yaml:"min_complexity"`
	
	// RiskDistribution shows count by risk level
	RiskDistribution RiskDistribution `json:"risk_distribution" yaml:"risk_distribution"`
	
	// ComplexityDistribution shows count by complexity ranges
	ComplexityDistribution map[string]int `json:"complexity_distribution" yaml:"complexity_distribution"`
}

// RiskDistribution shows the count of functions by risk level
type RiskDistribution struct {
	Low    int `json:"low" yaml:"low"`
	Medium int `json:"medium" yaml:"medium"`
	High   int `json:"high" yaml:"high"`
}

// ReportMetadata contains information about report generation
type ReportMetadata struct {
	// GeneratedAt is when the report was generated
	GeneratedAt time.Time `json:"generated_at" yaml:"generated_at"`
	
	// Version is the pyqol version used
	Version string `json:"version" yaml:"version"`
	
	// Configuration is the analysis configuration used
	Configuration *config.Config `json:"configuration,omitempty" yaml:"configuration,omitempty"`
	
	// FilesAnalyzed is the number of files processed
	FilesAnalyzed int `json:"files_analyzed" yaml:"files_analyzed"`
}

// ReportWarning represents a warning or threshold violation
type ReportWarning struct {
	// Type of warning (threshold_exceeded, max_complexity_exceeded, etc.)
	Type string `json:"type" yaml:"type"`
	
	// Message describes the warning
	Message string `json:"message" yaml:"message"`
	
	// FunctionName is the function that triggered the warning (if applicable)
	FunctionName string `json:"function_name,omitempty" yaml:"function_name,omitempty"`
	
	// Complexity is the complexity value that triggered the warning
	Complexity int `json:"complexity,omitempty" yaml:"complexity,omitempty"`
}

// ComplexityReporter handles formatting and outputting complexity reports
type ComplexityReporter struct {
	config *config.Config
	writer io.Writer
}

// NewComplexityReporter creates a new complexity reporter
func NewComplexityReporter(cfg *config.Config, writer io.Writer) *ComplexityReporter {
	return &ComplexityReporter{
		config: cfg,
		writer: writer,
	}
}

// GetWriter returns the writer used by this reporter
func (r *ComplexityReporter) GetWriter() io.Writer {
	return r.writer
}

// GenerateReport creates a comprehensive complexity report from results
func (r *ComplexityReporter) GenerateReport(results []ComplexityResult) *ComplexityReport {
	filtered := r.filterAndSortResults(results)
	
	// Convert interface results to serializable results
	serializableResults := make([]SerializableComplexityResult, len(filtered))
	for i, result := range filtered {
		serializableResults[i] = SerializableComplexityResult{
			Complexity:        result.GetComplexity(),
			FunctionName:      result.GetFunctionName(),
			RiskLevel:         result.GetRiskLevel(),
			Nodes:             result.GetNodes(),
			Edges:             result.GetEdges(),
			IfStatements:      result.GetIfStatements(),
			LoopStatements:    result.GetLoopStatements(),
			ExceptionHandlers: result.GetExceptionHandlers(),
			SwitchCases:       result.GetSwitchCases(),
		}
	}
	
	report := &ComplexityReport{
		Results: serializableResults,
		Metadata: ReportMetadata{
			GeneratedAt:   time.Now(),
			Version:       "dev", // TODO: Get from version package
			Configuration: r.config,
			FilesAnalyzed: 1, // TODO: Track actual file count
		},
	}
	
	report.Summary = r.generateSummaryFromSerializable(report.Results)
	report.Warnings = r.generateWarningsFromSerializable(report.Results)
	
	return report
}

// ReportComplexity formats and outputs the complexity results
func (r *ComplexityReporter) ReportComplexity(results []ComplexityResult) error {
	report := r.GenerateReport(results)
	
	switch strings.ToLower(r.config.Output.Format) {
	case "json":
		return r.outputJSON(report)
	case "yaml":
		return r.outputYAML(report)
	case "csv":
		return r.outputCSV(report)
	case "text":
		fallthrough
	default:
		return r.outputText(report)
	}
}

// filterAndSortResults applies configuration-based filtering and sorting
func (r *ComplexityReporter) filterAndSortResults(results []ComplexityResult) []ComplexityResult {
	// Filter by minimum complexity
	filtered := make([]ComplexityResult, 0, len(results))
	for _, result := range results {
		if result.GetComplexity() >= r.config.Output.MinComplexity {
			filtered = append(filtered, result)
		}
	}
	
	// Sort results based on configuration
	sort.Slice(filtered, func(i, j int) bool {
		switch r.config.Output.SortBy {
		case "complexity":
			return filtered[i].GetComplexity() > filtered[j].GetComplexity() // Descending
		case "risk":
			return r.compareRiskLevel(filtered[i].GetRiskLevel(), filtered[j].GetRiskLevel())
		case "name":
			fallthrough
		default:
			return filtered[i].GetFunctionName() < filtered[j].GetFunctionName() // Ascending
		}
	})
	
	return filtered
}

// compareRiskLevel compares risk levels for sorting (high > medium > low)
func (r *ComplexityReporter) compareRiskLevel(risk1, risk2 string) bool {
	riskValue := map[string]int{"high": 3, "medium": 2, "low": 1}
	return riskValue[risk1] > riskValue[risk2]
}

// generateSummary creates summary statistics from results
func (r *ComplexityReporter) generateSummary(results []ComplexityResult) ReportSummary {
	if len(results) == 0 {
		return ReportSummary{}
	}
	
	summary := ReportSummary{
		TotalFunctions: len(results),
		MinComplexity:  results[0].GetComplexity(),
		MaxComplexity:  results[0].GetComplexity(),
		ComplexityDistribution: make(map[string]int),
	}
	
	totalComplexity := 0
	for _, result := range results {
		complexity := result.GetComplexity()
		totalComplexity += complexity
		
		if complexity > summary.MaxComplexity {
			summary.MaxComplexity = complexity
		}
		if complexity < summary.MinComplexity {
			summary.MinComplexity = complexity
		}
		
		// Count by risk level
		switch result.GetRiskLevel() {
		case "high":
			summary.RiskDistribution.High++
		case "medium":
			summary.RiskDistribution.Medium++
		case "low":
			summary.RiskDistribution.Low++
		}
		
		// Count by complexity ranges
		switch {
		case complexity == 1:
			summary.ComplexityDistribution["1"]++
		case complexity <= 5:
			summary.ComplexityDistribution["2-5"]++
		case complexity <= 10:
			summary.ComplexityDistribution["6-10"]++
		case complexity <= 20:
			summary.ComplexityDistribution["11-20"]++
		default:
			summary.ComplexityDistribution["21+"]++
		}
	}
	
	summary.AverageComplexity = float64(totalComplexity) / float64(len(results))
	
	return summary
}

// generateSummaryFromSerializable creates summary statistics from serializable results
func (r *ComplexityReporter) generateSummaryFromSerializable(results []SerializableComplexityResult) ReportSummary {
	if len(results) == 0 {
		return ReportSummary{}
	}
	
	summary := ReportSummary{
		TotalFunctions: len(results),
		MinComplexity:  results[0].Complexity,
		MaxComplexity:  results[0].Complexity,
		ComplexityDistribution: make(map[string]int),
	}
	
	totalComplexity := 0
	for _, result := range results {
		complexity := result.Complexity
		totalComplexity += complexity
		
		if complexity > summary.MaxComplexity {
			summary.MaxComplexity = complexity
		}
		if complexity < summary.MinComplexity {
			summary.MinComplexity = complexity
		}
		
		// Count by risk level
		switch result.RiskLevel {
		case "high":
			summary.RiskDistribution.High++
		case "medium":
			summary.RiskDistribution.Medium++
		case "low":
			summary.RiskDistribution.Low++
		}
		
		// Count by complexity ranges
		switch {
		case complexity == 1:
			summary.ComplexityDistribution["1"]++
		case complexity <= 5:
			summary.ComplexityDistribution["2-5"]++
		case complexity <= 10:
			summary.ComplexityDistribution["6-10"]++
		case complexity <= 20:
			summary.ComplexityDistribution["11-20"]++
		default:
			summary.ComplexityDistribution["21+"]++
		}
	}
	
	summary.AverageComplexity = float64(totalComplexity) / float64(len(results))
	
	return summary
}

// generateWarnings creates warnings based on thresholds and configuration
func (r *ComplexityReporter) generateWarnings(results []ComplexityResult) []ReportWarning {
	var warnings []ReportWarning
	
	for _, result := range results {
		complexity := result.GetComplexity()
		functionName := result.GetFunctionName()
		riskLevel := result.GetRiskLevel()
		
		// Check if function exceeds maximum allowed complexity
		if r.config.Complexity.ExceedsMaxComplexity(complexity) {
			warnings = append(warnings, ReportWarning{
				Type:         "max_complexity_exceeded",
				Message:      fmt.Sprintf("Function complexity %d exceeds maximum allowed %d", complexity, r.config.Complexity.MaxComplexity),
				FunctionName: functionName,
				Complexity:   complexity,
			})
		}
		
		// Check high complexity threshold
		if riskLevel == "high" {
			warnings = append(warnings, ReportWarning{
				Type:         "high_complexity",
				Message:      fmt.Sprintf("Function has high complexity (%d > %d)", complexity, r.config.Complexity.MediumThreshold),
				FunctionName: functionName,
				Complexity:   complexity,
			})
		}
	}
	
	return warnings
}

// generateWarningsFromSerializable creates warnings based on thresholds and configuration
func (r *ComplexityReporter) generateWarningsFromSerializable(results []SerializableComplexityResult) []ReportWarning {
	var warnings []ReportWarning
	
	for _, result := range results {
		complexity := result.Complexity
		functionName := result.FunctionName
		riskLevel := result.RiskLevel
		
		// Check if function exceeds maximum allowed complexity
		if r.config.Complexity.ExceedsMaxComplexity(complexity) {
			warnings = append(warnings, ReportWarning{
				Type:         "max_complexity_exceeded",
				Message:      fmt.Sprintf("Function complexity %d exceeds maximum allowed %d", complexity, r.config.Complexity.MaxComplexity),
				FunctionName: functionName,
				Complexity:   complexity,
			})
		}
		
		// Check high complexity threshold
		if riskLevel == "high" {
			warnings = append(warnings, ReportWarning{
				Type:         "high_complexity",
				Message:      fmt.Sprintf("Function has high complexity (%d > %d)", complexity, r.config.Complexity.MediumThreshold),
				FunctionName: functionName,
				Complexity:   complexity,
			})
		}
	}
	
	return warnings
}

// outputJSON formats the report as JSON
func (r *ComplexityReporter) outputJSON(report *ComplexityReport) error {
	encoder := json.NewEncoder(r.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

// outputYAML formats the report as YAML
func (r *ComplexityReporter) outputYAML(report *ComplexityReport) error {
	encoder := yaml.NewEncoder(r.writer)
	encoder.SetIndent(2)
	defer encoder.Close()
	return encoder.Encode(report)
}

// outputCSV formats the report as CSV
func (r *ComplexityReporter) outputCSV(report *ComplexityReport) error {
	writer := csv.NewWriter(r.writer)
	defer writer.Flush()
	
	// Write header
	header := []string{
		"Function", "Complexity", "Risk", "Nodes", "Edges", 
		"If Statements", "Loop Statements", "Exception Handlers",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}
	
	// Write data rows
	for _, result := range report.Results {
		row := []string{
			result.FunctionName,
			fmt.Sprintf("%d", result.Complexity),
			result.RiskLevel,
			fmt.Sprintf("%d", result.Nodes),
			fmt.Sprintf("%d", result.Edges),
			fmt.Sprintf("%d", result.IfStatements),
			fmt.Sprintf("%d", result.LoopStatements),
			fmt.Sprintf("%d", result.ExceptionHandlers),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}
	
	return nil
}

// outputText formats the report as human-readable text
func (r *ComplexityReporter) outputText(report *ComplexityReport) error {
	// Write summary
	fmt.Fprintf(r.writer, "Complexity Analysis Report\n")
	fmt.Fprintf(r.writer, "==========================\n\n")
	
	fmt.Fprintf(r.writer, "Summary:\n")
	fmt.Fprintf(r.writer, "  Total Functions: %d\n", report.Summary.TotalFunctions)
	fmt.Fprintf(r.writer, "  Average Complexity: %.2f\n", report.Summary.AverageComplexity)
	fmt.Fprintf(r.writer, "  Max Complexity: %d\n", report.Summary.MaxComplexity)
	fmt.Fprintf(r.writer, "  Min Complexity: %d\n", report.Summary.MinComplexity)
	
	fmt.Fprintf(r.writer, "\nRisk Distribution:\n")
	fmt.Fprintf(r.writer, "  High: %d\n", report.Summary.RiskDistribution.High)
	fmt.Fprintf(r.writer, "  Medium: %d\n", report.Summary.RiskDistribution.Medium)
	fmt.Fprintf(r.writer, "  Low: %d\n", report.Summary.RiskDistribution.Low)
	
	// Write warnings if any
	if len(report.Warnings) > 0 {
		fmt.Fprintf(r.writer, "\nWarnings:\n")
		for _, warning := range report.Warnings {
			fmt.Fprintf(r.writer, "  [%s] %s\n", strings.ToUpper(warning.Type), warning.Message)
		}
	}
	
	// Write individual function results
	if len(report.Results) > 0 {
		fmt.Fprintf(r.writer, "\nFunction Details:\n")
		fmt.Fprintf(r.writer, "%-30s %10s %8s", "Function", "Complexity", "Risk")
		
		if r.config.Output.ShowDetails {
			fmt.Fprintf(r.writer, " %6s %6s %4s %4s %4s", "Nodes", "Edges", "Ifs", "Loops", "Excps")
		}
		fmt.Fprintf(r.writer, "\n")
		
		fmt.Fprint(r.writer, strings.Repeat("-", 30+10+8))
		if r.config.Output.ShowDetails {
			fmt.Fprint(r.writer, strings.Repeat("-", 6+6+4+4+4+5))
		}
		fmt.Fprintf(r.writer, "\n")
		
		for _, result := range report.Results {
			riskColor := r.getRiskColor(result.RiskLevel)
			fmt.Fprintf(r.writer, "%-30s %10d %s%8s%s", 
				result.FunctionName, result.Complexity, riskColor, result.RiskLevel, "\033[0m")
			
			if r.config.Output.ShowDetails {
				fmt.Fprintf(r.writer, " %6d %6d %4d %4d %4d", 
					result.Nodes, result.Edges, result.IfStatements, 
					result.LoopStatements, result.ExceptionHandlers)
			}
			fmt.Fprintf(r.writer, "\n")
		}
	}
	
	fmt.Fprintf(r.writer, "\nGenerated at: %s\n", report.Metadata.GeneratedAt.Format(time.RFC3339))
	
	return nil
}

// getRiskColor returns ANSI color codes for risk levels
func (r *ComplexityReporter) getRiskColor(riskLevel string) string {
	switch riskLevel {
	case "high":
		return "\033[31m" // Red
	case "medium":
		return "\033[33m" // Yellow
	case "low":
		return "\033[32m" // Green
	default:
		return "\033[0m"  // Reset
	}
}

// FormatComplexityBrief returns a brief one-line summary
func FormatComplexityBrief(results []ComplexityResult) string {
	if len(results) == 0 {
		return "No functions analyzed"
	}
	
	// Calculate aggregate stats inline to avoid circular dependency
	totalComplexity := 0
	maxComplexity := 0
	highRiskCount := 0
	
	for _, result := range results {
		complexity := result.GetComplexity()
		totalComplexity += complexity
		
		if complexity > maxComplexity {
			maxComplexity = complexity
		}
		
		if result.GetRiskLevel() == "high" {
			highRiskCount++
		}
	}
	
	avgComplexity := float64(totalComplexity) / float64(len(results))
	
	return fmt.Sprintf("%d functions analyzed - Avg: %.1f, Max: %d, High Risk: %d", 
		len(results), avgComplexity, maxComplexity, highRiskCount)
}