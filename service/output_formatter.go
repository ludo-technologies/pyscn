package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
)

// OutputFormatterImpl implements the OutputFormatter interface
type OutputFormatterImpl struct{}

// NewOutputFormatter creates a new output formatter service
func NewOutputFormatter() *OutputFormatterImpl {
	return &OutputFormatterImpl{}
}

// Format formats the analysis response according to the specified format
func (f *OutputFormatterImpl) Format(response *domain.ComplexityResponse, format domain.OutputFormat) (string, error) {
	switch format {
	case domain.OutputFormatText:
		return f.formatText(response)
	case domain.OutputFormatJSON:
		return f.formatJSON(response)
	case domain.OutputFormatYAML:
		return f.formatYAML(response)
	case domain.OutputFormatCSV:
		return f.formatCSV(response)
	case domain.OutputFormatHTML:
		return f.formatHTML(response)
	default:
		return "", domain.NewUnsupportedFormatError(string(format))
	}
}

// Write writes the formatted output to the writer
func (f *OutputFormatterImpl) Write(response *domain.ComplexityResponse, format domain.OutputFormat, writer io.Writer) error {
	output, err := f.Format(response, format)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(output))
	if err != nil {
		return domain.NewOutputError("failed to write output", err)
	}

	return nil
}

// formatText formats the response as human-readable text
func (f *OutputFormatterImpl) formatText(response *domain.ComplexityResponse) (string, error) {
	var builder strings.Builder
	utils := NewFormatUtils()

	// Header
	builder.WriteString(utils.FormatMainHeader("Complexity Analysis Report"))

	// Summary
	stats := map[string]interface{}{
		"Total Functions": response.Summary.TotalFunctions,
		"Files Analyzed":  response.Summary.FilesAnalyzed,
	}
	if response.Summary.TotalFunctions > 0 {
		stats["Average Complexity"] = fmt.Sprintf("%.1f", response.Summary.AverageComplexity)
		stats["Max Complexity"] = response.Summary.MaxComplexity
		stats["Min Complexity"] = response.Summary.MinComplexity
	}
	builder.WriteString(utils.FormatSummaryStats(stats))

	// Risk Distribution
	builder.WriteString(utils.FormatRiskDistribution(
		response.Summary.HighRiskFunctions,
		response.Summary.MediumRiskFunctions,
		response.Summary.LowRiskFunctions))

	// Function Details
	if len(response.Functions) > 0 {
		builder.WriteString(utils.FormatSectionHeader("FUNCTION DETAILS"))
		builder.WriteString(utils.FormatTableHeader("Function", "Complexity", "Risk"))

		for _, function := range response.Functions {
			// Convert domain risk level to standard risk level
			var standardRisk RiskLevel
			switch function.RiskLevel {
			case "High":
				standardRisk = RiskHigh
			case "Medium":
				standardRisk = RiskMedium
			case "Low":
				standardRisk = RiskLow
			default:
				standardRisk = RiskLow
			}

			coloredRisk := utils.FormatRiskWithColor(standardRisk)
			builder.WriteString(fmt.Sprintf("%-30s %10d  %s\n",
				function.Name,
				function.Metrics.Complexity,
				coloredRisk))
		}
		builder.WriteString(utils.FormatSectionSeparator())
	}

	// Warnings
	if len(response.Warnings) > 0 {
		builder.WriteString(utils.FormatWarningsSection(response.Warnings))
	}

	// Errors
	if len(response.Errors) > 0 {
		builder.WriteString(utils.FormatSectionHeader("ERRORS"))
		for _, err := range response.Errors {
			builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "❌", err))
		}
		builder.WriteString(utils.FormatSectionSeparator())
	}

	// Footer
	if parsedTime, err := time.Parse(time.RFC3339, response.GeneratedAt); err == nil {
		builder.WriteString(utils.FormatSectionHeader("METADATA"))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Generated at", parsedTime.Format("2006-01-02T15:04:05-07:00")))
	}

	return builder.String(), nil
}

// formatJSON formats the response as JSON
func (f *OutputFormatterImpl) formatJSON(response *domain.ComplexityResponse) (string, error) {
	// Create a JSON-friendly structure
	jsonResponse := f.createJSONResponse(response)
	return EncodeJSON(jsonResponse)
}

// formatYAML formats the response as YAML
func (f *OutputFormatterImpl) formatYAML(response *domain.ComplexityResponse) (string, error) {
	// Create a YAML-friendly structure
	yamlResponse := f.createJSONResponse(response) // Same structure works for YAML
	return EncodeYAML(yamlResponse)
}

// formatCSV formats the response as CSV
func (f *OutputFormatterImpl) formatCSV(response *domain.ComplexityResponse) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	// Write header
	header := []string{"Function", "Complexity", "Risk", "Nodes", "Edges", "If Statements", "Loop Statements", "Exception Handlers"}
	if err := writer.Write(header); err != nil {
		return "", domain.NewOutputError("failed to write CSV header", err)
	}

	// Write data rows
	for _, function := range response.Functions {
		row := []string{
			function.Name,
			fmt.Sprintf("%d", function.Metrics.Complexity),
			string(function.RiskLevel),
			fmt.Sprintf("%d", function.Metrics.Nodes),
			fmt.Sprintf("%d", function.Metrics.Edges),
			fmt.Sprintf("%d", function.Metrics.IfStatements),
			fmt.Sprintf("%d", function.Metrics.LoopStatements),
			fmt.Sprintf("%d", function.Metrics.ExceptionHandlers),
		}
		if err := writer.Write(row); err != nil {
			return "", domain.NewOutputError("failed to write CSV row", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", domain.NewOutputError("CSV writer error", err)
	}

	return builder.String(), nil
}

// createJSONResponse creates a JSON/YAML-friendly response structure
func (f *OutputFormatterImpl) createJSONResponse(response *domain.ComplexityResponse) map[string]interface{} {
	// Convert domain types to serializable types
	functions := make([]map[string]interface{}, len(response.Functions))
	for i, function := range response.Functions {
		functions[i] = map[string]interface{}{
			"complexity":         function.Metrics.Complexity,
			"function_name":      function.Name,
			"file_path":          function.FilePath,
			"risk_level":         string(function.RiskLevel),
			"nodes":              function.Metrics.Nodes,
			"edges":              function.Metrics.Edges,
			"if_statements":      function.Metrics.IfStatements,
			"loop_statements":    function.Metrics.LoopStatements,
			"exception_handlers": function.Metrics.ExceptionHandlers,
			"switch_cases":       function.Metrics.SwitchCases,
		}
	}

	// Create risk distribution map
	riskDistribution := map[string]int{
		"low":    response.Summary.LowRiskFunctions,
		"medium": response.Summary.MediumRiskFunctions,
		"high":   response.Summary.HighRiskFunctions,
	}

	// Create summary
	summary := map[string]interface{}{
		"total_functions":         response.Summary.TotalFunctions,
		"files_analyzed":          response.Summary.FilesAnalyzed,
		"risk_distribution":       riskDistribution,
		"complexity_distribution": response.Summary.ComplexityDistribution,
	}

	if response.Summary.TotalFunctions > 0 {
		summary["average_complexity"] = response.Summary.AverageComplexity
		summary["max_complexity"] = response.Summary.MaxComplexity
		summary["min_complexity"] = response.Summary.MinComplexity
	}

	// Create metadata
	metadata := map[string]interface{}{
		"generated_at":   response.GeneratedAt,
		"version":        response.Version,
		"files_analyzed": response.Summary.FilesAnalyzed,
	}

	if response.Config != nil {
		metadata["configuration"] = response.Config
	}

	result := map[string]interface{}{
		"summary":  summary,
		"results":  functions,
		"metadata": metadata,
	}

	// Add warnings and errors if present
	if len(response.Warnings) > 0 {
		result["warnings"] = response.Warnings
	}
	if len(response.Errors) > 0 {
		result["errors"] = response.Errors
	}

	return result
}

// FormatSummaryOnly formats only the summary information
func (f *OutputFormatterImpl) FormatSummaryOnly(response *domain.ComplexityResponse, format domain.OutputFormat) (string, error) {
	switch format {
	case domain.OutputFormatText:
		return f.formatSummaryText(response), nil
	case domain.OutputFormatJSON:
		summary := map[string]interface{}{
			"summary": f.createJSONResponse(response)["summary"],
		}
		return EncodeJSON(summary)
	default:
		return f.Format(response, format)
	}
}

// formatSummaryText formats only the summary as text
func (f *OutputFormatterImpl) formatSummaryText(response *domain.ComplexityResponse) string {
	var builder strings.Builder

	builder.WriteString("Summary:\n")
	builder.WriteString(fmt.Sprintf("  Total Functions: %d\n", response.Summary.TotalFunctions))
	if response.Summary.TotalFunctions > 0 {
		builder.WriteString(fmt.Sprintf("  Average Complexity: %.2f\n", response.Summary.AverageComplexity))
		builder.WriteString(fmt.Sprintf("  Max Complexity: %d\n", response.Summary.MaxComplexity))
		builder.WriteString(fmt.Sprintf("  Min Complexity: %d\n", response.Summary.MinComplexity))
	}

	builder.WriteString("\nRisk Distribution:\n")
	builder.WriteString(fmt.Sprintf("  High: %d\n", response.Summary.HighRiskFunctions))
	builder.WriteString(fmt.Sprintf("  Medium: %d\n", response.Summary.MediumRiskFunctions))
	builder.WriteString(fmt.Sprintf("  Low: %d\n", response.Summary.LowRiskFunctions))

	if len(response.Summary.ComplexityDistribution) > 0 {
		builder.WriteString("\nComplexity Distribution:\n")

		// Sort the keys for consistent output
		var keys []string
		for k := range response.Summary.ComplexityDistribution {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			builder.WriteString(fmt.Sprintf("  %s: %d\n", k, response.Summary.ComplexityDistribution[k]))
		}
	}

	return builder.String()
}

// formatHTML formats the response as Lighthouse-style HTML
func (f *OutputFormatterImpl) formatHTML(response *domain.ComplexityResponse) (string, error) {
	htmlFormatter := NewHTMLFormatter()
	projectName := "Python Project" // Default project name, could be configurable
	return htmlFormatter.FormatComplexityAsHTML(response, projectName)
}
