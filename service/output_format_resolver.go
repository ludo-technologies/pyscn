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

// Determine preserves the original four-flag resolver contract. With no
// selection it returns text output without a file extension.
func (r *OutputFormatResolver) Determine(html, json, csv, yaml bool) (domain.OutputFormat, string, error) {
	return r.determine(outputFormatSelection{
		html: html,
		json: json,
		csv:  csv,
		yaml: yaml,
	}, domain.OutputFormatText, "")
}

// DetermineAnalyzeReport resolves the five analyze report flags. With no
// selection it returns the command's default HTML report.
func (r *OutputFormatResolver) DetermineAnalyzeReport(html, json, csv, yaml, text bool) (domain.OutputFormat, string, error) {
	return r.determine(outputFormatSelection{
		html: html,
		json: json,
		csv:  csv,
		yaml: yaml,
		text: text,
	}, domain.OutputFormatHTML, "html")
}

type outputFormatSelection struct {
	html bool
	json bool
	csv  bool
	yaml bool
	text bool
}

func (r *OutputFormatResolver) determine(selection outputFormatSelection, defaultFormat domain.OutputFormat, defaultExtension string) (domain.OutputFormat, string, error) {
	formatCount := 0
	format := defaultFormat
	extension := defaultExtension

	if selection.html {
		formatCount++
		format = domain.OutputFormatHTML
		extension = "html"
	}
	if selection.json {
		formatCount++
		format = domain.OutputFormatJSON
		extension = "json"
	}
	if selection.csv {
		formatCount++
		format = domain.OutputFormatCSV
		extension = "csv"
	}
	if selection.yaml {
		formatCount++
		format = domain.OutputFormatYAML
		extension = "yaml"
	}
	if selection.text {
		formatCount++
		format = domain.OutputFormatText
		extension = "txt"
	}

	if formatCount > 1 {
		return "", "", fmt.Errorf("only one output format flag can be specified")
	}
	return format, extension, nil
}
