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

	// Summary counts describe the complete population used by aggregate metrics.
	stats := map[string]interface{}{
		"Total Functions": formatFunctionCoverage(response.Summary.TotalFunctions, response.Summary.FunctionsParsed),
		"Class Scopes":    response.Summary.TotalClassScopes,
		"Files Analyzed":  response.Summary.FilesAnalyzed,
	}
	if response.Summary.TotalFunctions > 0 {
		stats["Average Complexity"] = fmt.Sprintf("%.1f", response.Summary.AverageComplexity)
		stats["Max Complexity"] = response.Summary.MaxComplexity
		stats["Min Complexity"] = response.Summary.MinComplexity
	}
	if response.Summary.TotalClassScopes > 0 {
		stats["Max Class Complexity"] = response.Summary.MaxClassComplexity
	}
	builder.WriteString(utils.FormatSummaryStats(stats))

	// Risk Distribution
	builder.WriteString(utils.FormatRiskDistribution(
		response.Summary.HighRiskFunctions,
		response.Summary.MediumRiskFunctions,
		response.Summary.LowRiskFunctions))
	builder.WriteString(formatDirectoryComplexityText(response.ByDirectory, 0))

	if response.RawMetricsSummary != nil {
		builder.WriteString(utils.FormatSectionHeader("RAW CODE METRICS"))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Files Analyzed", response.RawMetricsSummary.FilesAnalyzed))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "SLOC", response.RawMetricsSummary.SLOC))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "LLOC", response.RawMetricsSummary.LLOC))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Comment Lines", response.RawMetricsSummary.CommentLines))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Docstring Lines", response.RawMetricsSummary.DocstringLines))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Blank Lines", response.RawMetricsSummary.BlankLines))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Total Lines", response.RawMetricsSummary.TotalLines))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Comment Ratio", utils.FormatPercentage(response.RawMetricsSummary.CommentRatio*100)))
		builder.WriteString(utils.FormatSectionSeparator())

		for _, metrics := range response.RawMetrics {
			builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "File", metrics.FilePath))
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "SLOC", metrics.SLOC))
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "LLOC", metrics.LLOC))
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Comment Lines", metrics.CommentLines))
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Docstring Lines", metrics.DocstringLines))
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Blank Lines", metrics.BlankLines))
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Total Lines", metrics.TotalLines))
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Comment Ratio", utils.FormatPercentage(metrics.CommentRatio*100)))
		}

		if len(response.RawMetrics) > 0 {
			builder.WriteString(utils.FormatSectionSeparator())
		}
	}

	writeComplexityScopeDetails(&builder, utils, "FUNCTION DETAILS", response.Functions)
	writeComplexityScopeDetails(&builder, utils, "CLASS SCOPE DETAILS", response.ClassScopes)

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

func writeComplexityScopeDetails(builder *strings.Builder, utils *FormatUtils, title string, scopes []domain.FunctionComplexity) {
	if len(scopes) == 0 {
		return
	}

	builder.WriteString(utils.FormatSectionHeader(title))
	builder.WriteString(utils.FormatTableHeader("Scope", "Complexity", "Cognitive", "SLOC", "Risk"))

	for _, scope := range scopes {
		var standardRisk RiskLevel
		switch scope.RiskLevel {
		case domain.RiskLevelHigh:
			standardRisk = RiskHigh
		case domain.RiskLevelMedium:
			standardRisk = RiskMedium
		case domain.RiskLevelLow:
			standardRisk = RiskLow
		default:
			standardRisk = RiskLow
		}

		coloredRisk := utils.FormatRiskWithColor(standardRisk)
		builder.WriteString(fmt.Sprintf("%-30s %10d %10d %10d  %s\n",
			fmt.Sprintf("%s (%s)", scope.Name, scope.ScopeKind),
			scope.Metrics.Complexity,
			scope.Metrics.CognitiveComplexity,
			scope.Metrics.SLOC,
			coloredRisk))
	}
	builder.WriteString(utils.FormatSectionSeparator())
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

	// Append new columns so existing positions remain stable for CSV consumers.
	header := []string{"Function", "Complexity", "Cognitive Complexity", "Risk", "Nodes", "Edges", "Nesting Depth", "If Statements", "Loop Statements", "Exception Handlers", "SLOC", "Scope Kind"}
	if err := writer.Write(header); err != nil {
		return "", domain.NewOutputError("failed to write CSV header", err)
	}

	// Write all reported executable scopes. The appended kind column keeps the
	// established function columns stable while making class rows unambiguous.
	for _, function := range response.ReportedScopes() {
		row := []string{
			function.Name,
			fmt.Sprintf("%d", function.Metrics.Complexity),
			fmt.Sprintf("%d", function.Metrics.CognitiveComplexity),
			string(function.RiskLevel),
			fmt.Sprintf("%d", function.Metrics.Nodes),
			fmt.Sprintf("%d", function.Metrics.Edges),
			fmt.Sprintf("%d", function.Metrics.NestingDepth),
			fmt.Sprintf("%d", function.Metrics.IfStatements),
			fmt.Sprintf("%d", function.Metrics.LoopStatements),
			fmt.Sprintf("%d", function.Metrics.ExceptionHandlers),
			fmt.Sprintf("%d", function.Metrics.SLOC),
			string(function.ScopeKind),
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
	functions := serializeComplexityScopes(response.Functions)

	// Create risk distribution map
	riskDistribution := map[string]int{
		"low":    response.Summary.LowRiskFunctions,
		"medium": response.Summary.MediumRiskFunctions,
		"high":   response.Summary.HighRiskFunctions,
	}

	// Create summary
	// Both function counts describe the complete population used by aggregate metrics;
	// presentation filters only limit the top-level functions list.
	summary := map[string]interface{}{
		"total_functions":                response.Summary.TotalFunctions,
		"total_class_scopes":             response.Summary.TotalClassScopes,
		"max_class_complexity":           response.Summary.MaxClassComplexity,
		"max_class_cognitive_complexity": response.Summary.MaxClassCognitiveComplexity,
		"max_class_nesting_depth":        response.Summary.MaxClassNestingDepth,
		"high_risk_class_scopes":         response.Summary.HighRiskClassScopes,
		"functions_parsed":               response.Summary.FunctionsParsed,
		"files_analyzed":                 response.Summary.FilesAnalyzed,
		"risk_distribution":              riskDistribution,
		"complexity_distribution":        response.Summary.ComplexityDistribution,
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
		"summary":      summary,
		"results":      functions,
		"by_directory": response.ByDirectory,
		"metadata":     metadata,
	}
	if len(response.ClassScopes) > 0 {
		result["class_scopes"] = serializeComplexityScopes(response.ClassScopes)
	}

	if response.RawMetricsSummary != nil {
		result["raw_metrics_summary"] = map[string]interface{}{
			"files_analyzed":  response.RawMetricsSummary.FilesAnalyzed,
			"sloc":            response.RawMetricsSummary.SLOC,
			"lloc":            response.RawMetricsSummary.LLOC,
			"comment_lines":   response.RawMetricsSummary.CommentLines,
			"docstring_lines": response.RawMetricsSummary.DocstringLines,
			"blank_lines":     response.RawMetricsSummary.BlankLines,
			"total_lines":     response.RawMetricsSummary.TotalLines,
			"comment_ratio":   response.RawMetricsSummary.CommentRatio,
		}
	}

	if len(response.RawMetrics) > 0 {
		rawMetrics := make([]map[string]interface{}, len(response.RawMetrics))
		for i, metrics := range response.RawMetrics {
			rawMetrics[i] = map[string]interface{}{
				"file_path":       metrics.FilePath,
				"sloc":            metrics.SLOC,
				"lloc":            metrics.LLOC,
				"comment_lines":   metrics.CommentLines,
				"docstring_lines": metrics.DocstringLines,
				"blank_lines":     metrics.BlankLines,
				"total_lines":     metrics.TotalLines,
				"comment_ratio":   metrics.CommentRatio,
			}
		}
		result["raw_metrics"] = rawMetrics
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

func serializeComplexityScopes(scopes []domain.FunctionComplexity) []map[string]interface{} {
	rows := make([]map[string]interface{}, len(scopes))
	for i, scope := range scopes {
		rows[i] = map[string]interface{}{
			"complexity":           scope.Metrics.Complexity,
			"cognitive_complexity": scope.Metrics.CognitiveComplexity,
			"function_name":        scope.Name,
			"scope_kind":           string(scope.ScopeKind),
			"file_path":            scope.FilePath,
			"risk_level":           string(scope.RiskLevel),
			"sloc":                 scope.Metrics.SLOC,
			"nodes":                scope.Metrics.Nodes,
			"edges":                scope.Metrics.Edges,
			"nesting_depth":        scope.Metrics.NestingDepth,
			"if_statements":        scope.Metrics.IfStatements,
			"loop_statements":      scope.Metrics.LoopStatements,
			"exception_handlers":   scope.Metrics.ExceptionHandlers,
			"switch_cases":         scope.Metrics.SwitchCases,
		}
	}
	return rows
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
	builder.WriteString(fmt.Sprintf("  Total Functions: %s\n", formatFunctionCoverage(response.Summary.TotalFunctions, response.Summary.FunctionsParsed)))
	builder.WriteString(fmt.Sprintf("  Class Scopes: %d\n", response.Summary.TotalClassScopes))
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

// formatFunctionCoverage preserves the legacy "reported / parsed" display for
// externally constructed responses whose counts differ. Service-produced
// responses use one complete-population count for both values.
func formatFunctionCoverage(reported, parsed int) string {
	if parsed > 0 && parsed != reported {
		return fmt.Sprintf("%d reported / %d parsed", reported, parsed)
	}
	return fmt.Sprintf("%d", reported)
}
