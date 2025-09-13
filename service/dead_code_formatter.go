package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

// DeadCodeFormatterImpl implements the DeadCodeFormatter interface
type DeadCodeFormatterImpl struct{}

// NewDeadCodeFormatter creates a new dead code formatter service
func NewDeadCodeFormatter() *DeadCodeFormatterImpl {
	return &DeadCodeFormatterImpl{}
}

// Format formats the dead code analysis response according to the specified format
func (f *DeadCodeFormatterImpl) Format(response *domain.DeadCodeResponse, format domain.OutputFormat) (string, error) {
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

// Write writes the formatted dead code output to the writer
func (f *DeadCodeFormatterImpl) Write(response *domain.DeadCodeResponse, format domain.OutputFormat, writer io.Writer) error {
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

// FormatFinding formats a single dead code finding
func (f *DeadCodeFormatterImpl) FormatFinding(finding domain.DeadCodeFinding, format domain.OutputFormat) (string, error) {
	switch format {
	case domain.OutputFormatText:
		return f.formatFindingTextLegacy(finding), nil
	case domain.OutputFormatJSON:
		return EncodeJSON(finding)
	case domain.OutputFormatYAML:
		return EncodeYAML(finding)
	case domain.OutputFormatHTML:
		// HTML formatting for individual findings is not typically needed
		// Fall back to text format for individual findings
		return f.formatFindingTextLegacy(finding), nil
	default:
		return "", domain.NewUnsupportedFormatError(string(format))
	}
}

// formatText formats the response as human-readable text
func (f *DeadCodeFormatterImpl) formatText(response *domain.DeadCodeResponse) (string, error) {
	var output strings.Builder
	utils := NewFormatUtils()

	// Header
	output.WriteString(utils.FormatMainHeader("Dead Code Analysis Report"))

	// Summary
	stats := map[string]interface{}{
		"Total Files":           response.Summary.TotalFiles,
		"Files with Dead Code":  response.Summary.FilesWithDeadCode,
		"Total Findings":        response.Summary.TotalFindings,
		"Functions Analyzed":    response.Summary.TotalFunctions,
		"Functions with Issues": response.Summary.FunctionsWithDeadCode,
	}
	output.WriteString(utils.FormatSummaryStats(stats))

	// Severity distribution (using standard risk levels)
	output.WriteString(utils.FormatRiskDistribution(
		response.Summary.CriticalFindings, // Map Critical to High
		response.Summary.WarningFindings,  // Map Warning to Medium
		response.Summary.InfoFindings))    // Map Info to Low

	// File Details
	if len(response.Files) > 0 && response.Summary.TotalFindings > 0 {
		output.WriteString(utils.FormatSectionHeader("DETAILED FINDINGS"))

		for _, file := range response.Files {
			if len(file.Functions) > 0 {
				output.WriteString(utils.FormatLabelWithIndent(0, "File", file.FilePath))
				output.WriteString(strings.Repeat("-", HeaderWidth) + "\n")

				for _, function := range file.Functions {
					if len(function.Findings) > 0 {
						output.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Function", function.Name))
						for _, finding := range function.Findings {
							output.WriteString(f.formatFindingText(finding, utils) + "\n")
						}
						output.WriteString("\n")
					}
				}
			}
		}
		output.WriteString(utils.FormatSectionSeparator())
	}

	// Warnings
	if len(response.Warnings) > 0 {
		output.WriteString(utils.FormatWarningsSection(response.Warnings))
	}

	// Errors
	if len(response.Errors) > 0 {
		output.WriteString(utils.FormatSectionHeader("ERRORS"))
		for _, error := range response.Errors {
			output.WriteString(utils.FormatLabelWithIndent(SectionPadding, "❌", error))
		}
		output.WriteString(utils.FormatSectionSeparator())
	}

	return output.String(), nil
}

// formatFindingText formats a single finding as text
func (f *DeadCodeFormatterImpl) formatFindingText(finding domain.DeadCodeFinding, utils *FormatUtils) string {
	// Convert severity to standard risk level
	var standardRisk RiskLevel
	switch finding.Severity {
	case "critical":
		standardRisk = RiskHigh
	case "warning":
		standardRisk = RiskMedium
	case "info":
		standardRisk = RiskLow
	default:
		standardRisk = RiskLow
	}

	coloredSeverity := utils.FormatRiskWithColor(standardRisk)
	return fmt.Sprintf("    [%s] Line %d-%d: %s (%s)",
		coloredSeverity,
		finding.Location.StartLine,
		finding.Location.EndLine,
		finding.Description,
		finding.Reason)
}

// Keep the old method for backward compatibility with FormatFinding
func (f *DeadCodeFormatterImpl) formatFindingTextLegacy(finding domain.DeadCodeFinding) string {
	return fmt.Sprintf("  [%s] Line %d-%d: %s (%s)",
		strings.ToUpper(string(finding.Severity)),
		finding.Location.StartLine,
		finding.Location.EndLine,
		finding.Description,
		finding.Reason)
}

// formatJSON formats the response as JSON
func (f *DeadCodeFormatterImpl) formatJSON(response *domain.DeadCodeResponse) (string, error) {
	return EncodeJSON(response)
}

// formatYAML formats the response as YAML
func (f *DeadCodeFormatterImpl) formatYAML(response *domain.DeadCodeResponse) (string, error) {
	return EncodeYAML(response)
}

// formatCSV formats the response as CSV
func (f *DeadCodeFormatterImpl) formatCSV(response *domain.DeadCodeResponse) (string, error) {
	var output strings.Builder
	writer := csv.NewWriter(&output)

	// Write header
	header := []string{"File", "Function", "Severity", "StartLine", "EndLine", "Reason", "Description"}
	if err := writer.Write(header); err != nil {
		return "", domain.NewOutputError("failed to write CSV header", err)
	}

	// Write findings
	for _, file := range response.Files {
		for _, function := range file.Functions {
			for _, finding := range function.Findings {
				record := []string{
					finding.Location.FilePath,
					finding.FunctionName,
					string(finding.Severity),
					fmt.Sprintf("%d", finding.Location.StartLine),
					fmt.Sprintf("%d", finding.Location.EndLine),
					finding.Reason,
					finding.Description,
				}
				if err := writer.Write(record); err != nil {
					return "", domain.NewOutputError("failed to write CSV record", err)
				}
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", domain.NewOutputError("CSV writer error", err)
	}

	return output.String(), nil
}

// formatHTML formats the response as Lighthouse-style HTML
func (f *DeadCodeFormatterImpl) formatHTML(response *domain.DeadCodeResponse) (string, error) {
	htmlFormatter := NewHTMLFormatter()
	projectName := "Python Project" // Default project name, could be configurable
	return htmlFormatter.FormatDeadCodeAsHTML(response, projectName)
}
