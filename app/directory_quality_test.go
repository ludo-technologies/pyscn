package app

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestAggregateComplexityByDirectory_UsesReportedFunctions(t *testing.T) {
	root := t.TempDir()
	functions := []domain.FunctionComplexity{
		{
			Name:      "hotspot",
			FilePath:  filepath.Join(root, "pkg", "hot.py"),
			Metrics:   domain.ComplexityMetrics{Complexity: 10, NestingDepth: 4},
			RiskLevel: domain.RiskLevelHigh,
		},
		{
			Name:      "helper",
			FilePath:  filepath.Join(root, "pkg", "helper.py"),
			Metrics:   domain.ComplexityMetrics{Complexity: 2, NestingDepth: 2},
			RiskLevel: domain.RiskLevelLow,
		},
		{
			Name:      "nested",
			FilePath:  filepath.Join(root, "pkg", "sub", "nested.py"),
			Metrics:   domain.ComplexityMetrics{Complexity: 3, NestingDepth: 1},
			RiskLevel: domain.RiskLevelLow,
		},
	}

	rollups, err := aggregateComplexityByDirectory(functions, root)
	if err != nil {
		t.Fatalf("aggregate complexity by directory: %v", err)
	}
	byPath := make(map[string]domain.DirectoryComplexityMetrics, len(rollups))
	for _, rollup := range rollups {
		byPath[rollup.DirectoryPath] = rollup
	}

	pkg := byPath["pkg"]
	if pkg.FunctionCount != 2 || pkg.HighRiskFunctionCount != 1 || pkg.MaxComplexity != 10 {
		t.Fatalf("expected direct directory metrics, got %+v", pkg)
	}
	if math.Abs(pkg.AverageComplexity-6) > 0.0001 {
		t.Fatalf("expected average complexity 6, got %f", pkg.AverageComplexity)
	}
	if math.Abs(pkg.AverageNestingDepth-3) > 0.0001 || pkg.MaxNestingDepth != 4 {
		t.Fatalf("expected average/max nesting 3/4, got %+v", pkg)
	}
	if byPath[filepath.Join("pkg", "sub")].FunctionCount != 1 {
		t.Fatalf("expected nested directory to remain distinct, got %+v", byPath)
	}
}

func TestComplexityDirectoryRoot_UsesExplicitAnalysisScope(t *testing.T) {
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

	got, err := complexityDirectoryRoot([]string{file, right})
	if err != nil {
		t.Fatalf("resolve complexity directory root: %v", err)
	}
	if got != root {
		t.Fatalf("expected common explicit root %q, got %q", root, got)
	}
}

func TestAggregateComplexityByDirectory_RejectsFunctionsOutsideRoot(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(filepath.Dir(root), "outside.py")
	_, err := aggregateComplexityByDirectory([]domain.FunctionComplexity{{FilePath: outside}}, root)
	if err == nil {
		t.Fatal("expected function outside the analysis root to be rejected")
	}
}
