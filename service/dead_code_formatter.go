package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pyqol/pyqol/domain"
	"gopkg.in/yaml.v3"
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
		return f.formatFindingText(finding), nil
	case domain.OutputFormatJSON:
		data, err := json.Marshal(finding)
		return string(data), err
	case domain.OutputFormatYAML:
		data, err := yaml.Marshal(finding)
		return string(data), err
	default:
		return "", domain.NewUnsupportedFormatError(string(format))
	}
}

// formatText formats the response as human-readable text
func (f *DeadCodeFormatterImpl) formatText(response *domain.DeadCodeResponse) (string, error) {
	var output strings.Builder

	// Header
	output.WriteString("Dead Code Detection Results\n")
	output.WriteString("============================\n\n")

	// Summary
	output.WriteString(fmt.Sprintf("Files analyzed: %d\n", response.Summary.TotalFiles))
	output.WriteString(fmt.Sprintf("Files with dead code: %d\n", response.Summary.FilesWithDeadCode))
	output.WriteString(fmt.Sprintf("Total findings: %d\n", response.Summary.TotalFindings))
	output.WriteString(fmt.Sprintf("Functions analyzed: %d\n", response.Summary.TotalFunctions))
	output.WriteString(fmt.Sprintf("Functions with dead code: %d\n\n", response.Summary.FunctionsWithDeadCode))

	// Severity breakdown
	output.WriteString("Severity Breakdown:\n")
	output.WriteString(fmt.Sprintf("  Critical: %d\n", response.Summary.CriticalFindings))
	output.WriteString(fmt.Sprintf("  Warning:  %d\n", response.Summary.WarningFindings))
	output.WriteString(fmt.Sprintf("  Info:     %d\n\n", response.Summary.InfoFindings))

	// Files with findings
	for _, file := range response.Files {
		output.WriteString(fmt.Sprintf("File: %s\n", file.FilePath))
		output.WriteString(strings.Repeat("=", len(file.FilePath)+6) + "\n")

		for _, function := range file.Functions {
			output.WriteString(fmt.Sprintf("\nFunction: %s\n", function.Name))
			for _, finding := range function.Findings {
				output.WriteString(f.formatFindingText(finding) + "\n")
			}
		}
		output.WriteString("\n")
	}

	// Warnings and errors
	if len(response.Warnings) > 0 {
		output.WriteString("Warnings:\n")
		for _, warning := range response.Warnings {
			output.WriteString(fmt.Sprintf("  - %s\n", warning))
		}
		output.WriteString("\n")
	}

	if len(response.Errors) > 0 {
		output.WriteString("Errors:\n")
		for _, error := range response.Errors {
			output.WriteString(fmt.Sprintf("  - %s\n", error))
		}
		output.WriteString("\n")
	}

	return output.String(), nil
}

// formatFindingText formats a single finding as text
func (f *DeadCodeFormatterImpl) formatFindingText(finding domain.DeadCodeFinding) string {
	return fmt.Sprintf("  [%s] Line %d-%d: %s (%s)",
		strings.ToUpper(string(finding.Severity)),
		finding.Location.StartLine,
		finding.Location.EndLine,
		finding.Description,
		finding.Reason)
}

// formatJSON formats the response as JSON
func (f *DeadCodeFormatterImpl) formatJSON(response *domain.DeadCodeResponse) (string, error) {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", domain.NewOutputError("failed to marshal JSON", err)
	}
	return string(data), nil
}

// formatYAML formats the response as YAML
func (f *DeadCodeFormatterImpl) formatYAML(response *domain.DeadCodeResponse) (string, error) {
	data, err := yaml.Marshal(response)
	if err != nil {
		return "", domain.NewOutputError("failed to marshal YAML", err)
	}
	return string(data), nil
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