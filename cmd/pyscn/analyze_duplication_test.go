package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

// A few eligible clones in a mostly small-function project must not make the
// whole project appear duplicated. A clone-heavy project still reports its
// full density even after the penalty saturates.
func TestAnalyzeDuplicationScoreUsesDistinctSourceLines(t *testing.T) {
	for _, smallFunctions := range []int{0, 40} {
		t.Run(fmt.Sprintf("%d small functions", smallFunctions), func(t *testing.T) {
			project := t.TempDir()
			var source strings.Builder
			for _, name := range []string{"big_a", "big_b"} {
				fmt.Fprintf(&source, "def %s(x):\n", name)
				for i := 1; i < 14; i++ {
					fmt.Fprintf(&source, "    x = x + %d\n", i)
				}
				source.WriteString("    return x\n\n")
			}
			for i := 0; i < smallFunctions; i++ {
				fmt.Fprintf(&source, "def small_%d(x):\n    y = x + %d\n    return y * 2\n\n", i, i)
			}
			if err := os.WriteFile(filepath.Join(project, "sample.py"), []byte(source.String()), 0o644); err != nil {
				t.Fatal(err)
			}

			analyze := func() domain.AnalyzeResponse {
				t.Helper()
				var output bytes.Buffer
				cmd := NewAnalyzeCmd()
				cmd.SetOut(&output)
				cmd.SetErr(io.Discard)
				cmd.SetArgs([]string{"--json", "--output", "-", "--no-open", "--select", "clones", project})
				if err := cmd.Execute(); err != nil {
					t.Fatalf("analyze clones: %v", err)
				}
				var report domain.AnalyzeResponse
				if err := json.Unmarshal(output.Bytes(), &report); err != nil {
					t.Fatalf("decode JSON report: %v", err)
				}
				return report
			}
			report := analyze()
			if report.Clone == nil || report.Clone.Statistics == nil {
				t.Fatal("missing clone statistics")
			}
			stats := report.Clone.Statistics
			if stats.TotalFragments != 2 || stats.DuplicatedFragments != 2 || stats.DuplicatedLines != 30 {
				t.Fatalf("expected two 15-line clone fragments, got %+v", stats)
			}
			want := float64(30) / float64(stats.LinesAnalyzed) * 100
			if math.Abs(report.Summary.CodeDuplication-want) > 1e-9 {
				t.Fatalf("duplication = %.2f%%, want %.2f%%", report.Summary.CodeDuplication, want)
			}
			if smallFunctions == 0 {
				if report.Summary.CodeDuplication <= 60 || report.Summary.DuplicationScore != 0 {
					t.Fatalf("high density must remain unclamped and saturate scoring: %+v", report.Summary)
				}
			} else {
				if report.Summary.CodeDuplication >= 60 || report.Summary.DuplicationScore == 0 {
					t.Fatalf("small functions must lower source-line density and penalty: %+v", report.Summary)
				}
				if err := os.WriteFile(filepath.Join(project, ".pyscn.toml"), []byte("[clones]\nmax_similarity = 0.5\n"), 0o644); err != nil {
					t.Fatal(err)
				}
				filtered := analyze()
				if filtered.Clone == nil || filtered.Clone.Statistics == nil || filtered.Clone.Statistics.TotalClonePairs != 0 {
					t.Fatalf("expected the presentation filter to hide every pair: %+v", filtered.Clone)
				}
				if filtered.Summary.CodeDuplication != report.Summary.CodeDuplication ||
					filtered.Summary.DuplicationScore != report.Summary.DuplicationScore {
					t.Fatalf("presentation filter changed the score: before=%+v after=%+v", report.Summary, filtered.Summary)
				}
			}
		})
	}
}
