package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

// DepsFormatterImpl implements domain.DepsOutputFormatter
type DepsFormatterImpl struct{}

func NewDepsFormatter() *DepsFormatterImpl { return &DepsFormatterImpl{} }

func (f *DepsFormatterImpl) Write(resp *domain.DependencyResponse, format domain.OutputFormat, w io.Writer) error {
	switch format {
	case domain.OutputFormatText:
		_, err := w.Write([]byte(f.formatText(resp)))
		return err
	case domain.OutputFormatJSON:
		return WriteJSON(w, resp)
	case domain.OutputFormatYAML:
		return WriteYAML(w, resp)
	case domain.OutputFormatCSV:
		cw := csv.NewWriter(w)
		if err := cw.Write([]string{"from", "to"}); err != nil {
			return err
		}
		for _, e := range resp.Edges {
			if err := cw.Write([]string{e.From, e.To}); err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case domain.OutputFormatHTML:
		html := NewHTMLFormatter()
		content, err := html.FormatDepsAsHTML(resp, "Python Project")
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(content))
		return err
	default:
		return domain.NewUnsupportedFormatError(string(format))
	}
}

func (f *DepsFormatterImpl) formatText(resp *domain.DependencyResponse) string {
	var b strings.Builder
	b.WriteString("Dependency Analysis\n=====================\n\n")
	fmt.Fprintf(&b, "Modules: %d\nEdges:   %d\nCycles:  %d\n", resp.Summary.Modules, resp.Summary.Edges, resp.Summary.Cycles)
	if resp.Summary.LayerViolations > 0 {
		fmt.Fprintf(&b, "Layer Violations: %d\n", resp.Summary.LayerViolations)
	}
	b.WriteString("\n")
	if len(resp.Cycles) > 0 {
		b.WriteString("Cycles:\n")
		for i, cyc := range resp.Cycles {
			fmt.Fprintf(&b, "  %d) %s\n", i+1, strings.Join(cyc.Modules, " -> "))
		}
		b.WriteString("\n")
	}
	if len(resp.Errors) > 0 {
		b.WriteString("Errors:\n")
		for _, e := range resp.Errors {
			fmt.Fprintf(&b, "  - %s\n", e)
		}
		b.WriteString("\n")
	}
	if len(resp.Warnings) > 0 {
		b.WriteString("Warnings:\n")
		for _, w := range resp.Warnings {
			fmt.Fprintf(&b, "  - %s\n", w)
		}
		b.WriteString("\n")
	}
	if len(resp.LayerViolations) > 0 {
		b.WriteString("Layer Rule Violations:\n")
		for _, v := range resp.LayerViolations {
			fmt.Fprintf(&b, "  - %s (%s) -> %s (%s)\n", v.FromModule, v.FromLayer, v.ToModule, v.ToLayer)
		}
		b.WriteString("\n")
	}
	return b.String()
}
