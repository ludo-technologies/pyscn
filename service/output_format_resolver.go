package service

import (
    "fmt"

    "github.com/ludo-technologies/pyscn/domain"
)

// OutputFormatResolver resolves output format and file extension from flags.
type OutputFormatResolver struct{}

func NewOutputFormatResolver() *OutputFormatResolver { return &OutputFormatResolver{} }

// Determine evaluates format flags and returns the selected format and extension.
// Exactly one of html/json/csv/yaml may be true; if none are true, defaults to text.
func (r *OutputFormatResolver) Determine(html, json, csv, yaml bool) (domain.OutputFormat, string, error) {
    formatCount := 0
    var format domain.OutputFormat
    var ext string

    if html {
        formatCount++
        format = domain.OutputFormatHTML
        ext = "html"
    }
    if json {
        formatCount++
        format = domain.OutputFormatJSON
        ext = "json"
    }
    if csv {
        formatCount++
        format = domain.OutputFormatCSV
        ext = "csv"
    }
    if yaml {
        formatCount++
        format = domain.OutputFormatYAML
        ext = "yaml"
    }

    if formatCount > 1 {
        return "", "", fmt.Errorf("only one output format flag can be specified")
    }
    if formatCount == 0 {
        return domain.OutputFormatText, "", nil
    }
    return format, ext, nil
}

