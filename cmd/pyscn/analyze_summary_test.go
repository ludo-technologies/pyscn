package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/ludo-technologies/pyscn/domain"
)

// TestPrintSummaryLabelsModulesInCycles checks that the dependencies line
// reports modules in cycles as modules, not as cycles (issue #869).
func TestPrintSummaryLabelsModulesInCycles(t *testing.T) {
	tests := []struct {
		modulesInCycles int
		want            string
	}{
		{0, "(no cycles, depth: 2)"},
		{1, "(1 module in cycles, depth: 2)"},
		{3, "(3 modules in cycles, depth: 2)"},
	}
	for _, tt := range tests {
		var stderr bytes.Buffer
		cmd := &cobra.Command{}
		cmd.SetErr(&stderr)
		response := &domain.AnalyzeResponse{Summary: domain.AnalyzeSummary{
			DepsEnabled:         true,
			DepsModulesInCycles: tt.modulesInCycles,
			DepsMaxDepth:        2,
		}}

		(&AnalyzeCommand{}).printSummary(cmd, response)

		if !strings.Contains(stderr.String(), tt.want) {
			t.Errorf("modules in cycles = %d: want %q in output:\n%s", tt.modulesInCycles, tt.want, stderr.String())
		}
	}
}
