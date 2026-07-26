package app

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestAggregateDirectoryQuality_UsesRecursiveWeightedRollups(t *testing.T) {
	root := t.TempDir()
	modules := []domain.ModuleQualityMetrics{
		{
			FilePath:      filepath.Join(root, "pkg", "hot.py"),
			LinesOfCode:   100,
			FunctionCount: 4,
			ModuleComplexityMetrics: domain.ModuleComplexityMetrics{
				AnalyzedFunctionCount:      2,
				AverageComplexity:          8,
				AverageCognitiveComplexity: 6,
				MaxComplexity:              12,
				HighRiskFunctionCount:      1,
				ExceptionHandlerCount:      3,
			},
			ModuleDeadCodeMetrics: domain.ModuleDeadCodeMetrics{
				DeadCodeFindingCount: 2,
				DeadCodeBlockCount:   1,
			},
		},
		{
			FilePath:      filepath.Join(root, "pkg", "sub", "helper.py"),
			LinesOfCode:   30,
			FunctionCount: 2,
			ModuleComplexityMetrics: domain.ModuleComplexityMetrics{
				AnalyzedFunctionCount:      1,
				AverageComplexity:          2,
				AverageCognitiveComplexity: 3,
				MaxComplexity:              2,
				ExceptionHandlerCount:      1,
			},
			ModuleDeadCodeMetrics: domain.ModuleDeadCodeMetrics{
				DeadCodeFindingCount: 1,
				DeadCodeBlockCount:   2,
			},
		},
	}

	rollups, err := aggregateDirectoryQuality(modules, root)
	if err != nil {
		t.Fatalf("aggregate directory quality: %v", err)
	}
	byPath := make(map[string]domain.DirectoryQualityMetrics, len(rollups))
	for _, rollup := range rollups {
		byPath[rollup.DirectoryPath] = rollup
	}

	pkg := byPath["pkg"]
	if pkg.ModuleCount != 2 || pkg.DirectModuleCount != 1 {
		t.Fatalf("expected recursive and direct module counts, got %+v", pkg)
	}
	if pkg.LinesOfCode != 130 || pkg.FunctionCount != 6 || pkg.AnalyzedFunctionCount != 3 {
		t.Fatalf("expected additive directory metrics, got %+v", pkg)
	}
	if math.Abs(pkg.AverageComplexity-6) > 0.0001 {
		t.Fatalf("expected function-weighted average complexity 6, got %f", pkg.AverageComplexity)
	}
	if math.Abs(pkg.AverageCognitiveComplexity-5) > 0.0001 {
		t.Fatalf("expected function-weighted average cognitive complexity 5, got %f", pkg.AverageCognitiveComplexity)
	}
	if pkg.MaxComplexity != 12 || pkg.HighRiskFunctionCount != 1 || pkg.ExceptionHandlerCount != 4 {
		t.Fatalf("expected recursive complexity totals, got %+v", pkg)
	}
	if pkg.DeadCodeFindingCount != 3 || pkg.DeadCodeBlockCount != 3 {
		t.Fatalf("expected recursive dead-code totals, got %+v", pkg)
	}
}

func TestDirectoryQualityRoot_UsesExplicitAnalysisScope(t *testing.T) {
	root := t.TempDir()
	left := filepath.Join(root, "left")
	right := filepath.Join(root, "right")
	if err := os.MkdirAll(left, 0o755); err != nil {
		t.Fatalf("create left directory: %v", err)
	}
	if err := os.MkdirAll(right, 0o755); err != nil {
		t.Fatalf("create right directory: %v", err)
	}
	file := filepath.Join(left, "one.py")
	if err := os.WriteFile(file, []byte("pass\n"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}

	got, err := directoryQualityRoot([]string{file, right})
	if err != nil {
		t.Fatalf("resolve directory quality root: %v", err)
	}
	if got != root {
		t.Fatalf("expected common explicit root %q, got %q", root, got)
	}
}

func TestAggregateDirectoryQuality_RejectsModulesOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside.py")
	_, err := aggregateDirectoryQuality([]domain.ModuleQualityMetrics{{FilePath: outside}}, root)
	if err == nil {
		t.Fatal("expected module outside the analysis root to be rejected")
	}
}
