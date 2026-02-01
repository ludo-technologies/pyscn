package service

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
	"gopkg.in/yaml.v3"
)

// DIAntipatternFormatter implements the DIAntipatternOutputFormatter interface
type DIAntipatternFormatter struct{}

// NewDIAntipatternFormatter creates a new DI anti-pattern formatter
func NewDIAntipatternFormatter() *DIAntipatternFormatter {
	return &DIAntipatternFormatter{}
}

// Format formats the analysis response according to the specified format
func (f *DIAntipatternFormatter) Format(response *domain.DIAntipatternResponse, format domain.OutputFormat) (string, error) {
	var sb strings.Builder
	if err := f.Write(response, format, &sb); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// Write writes the formatted output to the writer
func (f *DIAntipatternFormatter) Write(response *domain.DIAntipatternResponse, format domain.OutputFormat, writer io.Writer) error {
	switch format {
	case domain.OutputFormatJSON:
		return f.writeJSON(response, writer)
	case domain.OutputFormatYAML:
		return f.writeYAML(response, writer)
	case domain.OutputFormatText:
		return f.writeText(response, writer)
	case domain.OutputFormatCSV:
		return f.writeCSV(response, writer)
	default:
		return f.writeJSON(response, writer)
	}
}

// writeJSON writes output in JSON format
func (f *DIAntipatternFormatter) writeJSON(response *domain.DIAntipatternResponse, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

// writeYAML writes output in YAML format
func (f *DIAntipatternFormatter) writeYAML(response *domain.DIAntipatternResponse, writer io.Writer) error {
	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(2)
	return encoder.Encode(response)
}

// writeText writes output in human-readable text format
func (f *DIAntipatternFormatter) writeText(response *domain.DIAntipatternResponse, writer io.Writer) error {
	// Header
	fmt.Fprintf(writer, "DI Anti-pattern Analysis Results\n")
	fmt.Fprintf(writer, "================================\n\n")

	// Summary
	fmt.Fprintf(writer, "Summary:\n")
	fmt.Fprintf(writer, "  Total Findings: %d\n", response.Summary.TotalFindings)
	fmt.Fprintf(writer, "  Files Analyzed: %d\n", response.Summary.FilesAnalyzed)
	fmt.Fprintf(writer, "  Classes Analyzed: %d\n", response.Summary.ClassesAnalyzed)
	fmt.Fprintf(writer, "  Affected Classes: %d\n", response.Summary.AffectedClasses)
	fmt.Fprintf(writer, "\n")

	// By type
	if len(response.Summary.ByType) > 0 {
		fmt.Fprintf(writer, "By Type:\n")
		for t, count := range response.Summary.ByType {
			fmt.Fprintf(writer, "  %s: %d\n", t, count)
		}
		fmt.Fprintf(writer, "\n")
	}

	// By severity
	if len(response.Summary.BySeverity) > 0 {
		fmt.Fprintf(writer, "By Severity:\n")
		for s, count := range response.Summary.BySeverity {
			fmt.Fprintf(writer, "  %s: %d\n", s, count)
		}
		fmt.Fprintf(writer, "\n")
	}

	// Findings
	if len(response.Findings) > 0 {
		fmt.Fprintf(writer, "Findings:\n")
		fmt.Fprintf(writer, "---------\n")

		for i, finding := range response.Findings {
			fmt.Fprintf(writer, "\n%d. [%s] %s\n", i+1, strings.ToUpper(string(finding.Severity)), finding.Type)
			fmt.Fprintf(writer, "   Location: %s:%d\n", finding.Location.FilePath, finding.Location.StartLine)
			if finding.ClassName != "" {
				fmt.Fprintf(writer, "   Class: %s\n", finding.ClassName)
			}
			if finding.MethodName != "" {
				fmt.Fprintf(writer, "   Method: %s\n", finding.MethodName)
			}
			fmt.Fprintf(writer, "   Description: %s\n", finding.Description)
			fmt.Fprintf(writer, "   Suggestion: %s\n", finding.Suggestion)
		}
	}

	// Warnings
	if len(response.Warnings) > 0 {
		fmt.Fprintf(writer, "\nWarnings:\n")
		for _, warning := range response.Warnings {
			fmt.Fprintf(writer, "  - %s\n", warning)
		}
	}

	// Errors
	if len(response.Errors) > 0 {
		fmt.Fprintf(writer, "\nErrors:\n")
		for _, err := range response.Errors {
			fmt.Fprintf(writer, "  - %s\n", err)
		}
	}

	fmt.Fprintf(writer, "\nGenerated at: %s\n", response.GeneratedAt)
	fmt.Fprintf(writer, "Version: %s\n", response.Version)

	return nil
}

// writeCSV writes output in CSV format
func (f *DIAntipatternFormatter) writeCSV(response *domain.DIAntipatternResponse, writer io.Writer) error {
	// Header
	fmt.Fprintf(writer, "type,subtype,severity,class_name,method_name,file_path,start_line,description,suggestion\n")

	// Findings
	for _, finding := range response.Findings {
		fmt.Fprintf(writer, "%s,%s,%s,%s,%s,%s,%d,%q,%q\n",
			finding.Type,
			finding.Subtype,
			finding.Severity,
			finding.ClassName,
			finding.MethodName,
			finding.Location.FilePath,
			finding.Location.StartLine,
			finding.Description,
			finding.Suggestion,
		)
	}

	return nil
}
