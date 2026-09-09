package service

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/domain"
)

// OutputFormatResolver resolves output format and file extension selections.
type OutputFormatResolver struct{}

// NewOutputFormatResolver creates an output format resolver.
func NewOutputFormatResolver() *OutputFormatResolver {
	return &OutputFormatResolver{}
}

// ReportFormat pairs an output format with the file extension its report uses.
type ReportFormat struct {
	Format    domain.OutputFormat
	Extension string
}

// Determine preserves the original four-flag resolver contract. With no
// selection it returns text output without a file extension.
func (r *OutputFormatResolver) Determine(html, json, csv, yaml bool) (domain.OutputFormat, string, error) {
	selected := selectedReportFormats(html, json, csv, yaml, false)
	switch len(selected) {
	case 0:
		return domain.OutputFormatText, "", nil
	case 1:
		return selected[0].Format, selected[0].Extension, nil
	default:
		return "", "", fmt.Errorf("only one output format flag can be specified")
	}
}

// DetermineAnalyzeReports resolves the five analyze report flags into every
// requested report, in flag order. With no selection it returns the command's
// default HTML report.
func (r *OutputFormatResolver) DetermineAnalyzeReports(html, json, csv, yaml, text bool) []ReportFormat {
	selected := selectedReportFormats(html, json, csv, yaml, text)
	if len(selected) == 0 {
		return []ReportFormat{{Format: domain.OutputFormatHTML, Extension: "html"}}
	}
	return selected
}

func selectedReportFormats(html, json, csv, yaml, text bool) []ReportFormat {
	var selected []ReportFormat
	add := func(on bool, format domain.OutputFormat, extension string) {
		if on {
			selected = append(selected, ReportFormat{Format: format, Extension: extension})
		}
	}
	add(html, domain.OutputFormatHTML, "html")
	add(json, domain.OutputFormatJSON, "json")
	add(csv, domain.OutputFormatCSV, "csv")
	add(yaml, domain.OutputFormatYAML, "yaml")
	add(text, domain.OutputFormatText, "txt")
	return selected
}
