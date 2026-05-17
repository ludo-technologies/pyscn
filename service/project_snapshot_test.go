package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestProjectSnapshotCachesParsedFileState(t *testing.T) {
	ctx := context.Background()
	sourcePath := writeSnapshotFixture(t)

	snapshot := BuildProjectSnapshot(ctx, []string{sourcePath})
	if len(snapshot.Files) != 1 {
		t.Fatalf("expected 1 snapshot file, got %d", len(snapshot.Files))
	}

	file := snapshot.Files[0]
	if !file.Parsed() {
		t.Fatalf("expected parsed file, read err: %v, parse err: %v", file.ReadErr, file.ParseErr)
	}
	if file.RawMetrics == nil {
		t.Fatal("expected raw metrics")
	}

	firstCFGs, err := file.CFGs()
	if err != nil {
		t.Fatalf("first CFG build failed: %v", err)
	}
	secondCFGs, err := file.CFGs()
	if err != nil {
		t.Fatalf("second CFG build failed: %v", err)
	}
	if len(firstCFGs) == 0 {
		t.Fatal("expected CFGs")
	}
	if firstCFGs[domain.ModuleFunctionName] != secondCFGs[domain.ModuleFunctionName] {
		t.Fatal("expected cached CFG objects to be reused")
	}
}

func TestProjectSnapshotOptionsSkipRawMetrics(t *testing.T) {
	ctx := context.Background()
	sourcePath := writeSnapshotFixture(t)

	snapshot := BuildProjectSnapshotWithOptions(ctx, []string{sourcePath}, ProjectSnapshotOptions{})
	if len(snapshot.Files) != 1 {
		t.Fatalf("expected 1 snapshot file, got %d", len(snapshot.Files))
	}

	file := snapshot.Files[0]
	if !file.Parsed() {
		t.Fatalf("expected parsed file, read err: %v, parse err: %v", file.ReadErr, file.ParseErr)
	}
	if file.RawMetrics != nil {
		t.Fatal("expected raw metrics to be skipped")
	}
	if _, err := file.CFGs(); err != nil {
		t.Fatalf("expected CFGs without raw metrics: %v", err)
	}
}

func TestComplexitySnapshotRequiresRawMetrics(t *testing.T) {
	ctx := context.Background()
	sourcePath := writeSnapshotFixture(t)
	paths := []string{sourcePath}
	snapshot := BuildProjectSnapshotWithOptions(ctx, paths, ProjectSnapshotOptions{})

	_, err := NewComplexityService().AnalyzeSnapshot(ctx, snapshot, domain.ComplexityRequest{
		Paths:           paths,
		OutputFormat:    domain.OutputFormatJSON,
		MinComplexity:   1,
		SortBy:          domain.SortByName,
		LowThreshold:    domain.DefaultComplexityLowThreshold,
		MediumThreshold: domain.DefaultComplexityMediumThreshold,
	})
	if err == nil {
		t.Fatal("expected complexity snapshot without raw metrics to fail")
	}
}

func TestSnapshotServicesMatchFileServices(t *testing.T) {
	ctx := context.Background()
	sourcePath := writeSnapshotFixture(t)
	paths := []string{sourcePath}
	snapshot := BuildProjectSnapshot(ctx, paths)

	complexityReq := domain.ComplexityRequest{
		Paths:           paths,
		OutputFormat:    domain.OutputFormatJSON,
		MinComplexity:   1,
		SortBy:          domain.SortByName,
		LowThreshold:    domain.DefaultComplexityLowThreshold,
		MediumThreshold: domain.DefaultComplexityMediumThreshold,
	}
	complexitySvc := NewComplexityService()
	regularComplexity, err := complexitySvc.Analyze(ctx, complexityReq)
	if err != nil {
		t.Fatalf("regular complexity failed: %v", err)
	}
	snapshotComplexity, err := complexitySvc.AnalyzeSnapshot(ctx, snapshot, complexityReq)
	if err != nil {
		t.Fatalf("snapshot complexity failed: %v", err)
	}
	if len(regularComplexity.Functions) != len(snapshotComplexity.Functions) {
		t.Fatalf("complexity function count mismatch: regular=%d snapshot=%d", len(regularComplexity.Functions), len(snapshotComplexity.Functions))
	}
	if regularComplexity.Summary.TotalFunctions != snapshotComplexity.Summary.TotalFunctions {
		t.Fatalf("complexity summary mismatch: regular=%d snapshot=%d", regularComplexity.Summary.TotalFunctions, snapshotComplexity.Summary.TotalFunctions)
	}

	cboReq := *domain.DefaultCBORequest()
	cboReq.Paths = paths
	cboReq.ShowZeros = domain.BoolPtr(true)
	cboSvc := NewCBOService()
	regularCBO, err := cboSvc.Analyze(ctx, cboReq)
	if err != nil {
		t.Fatalf("regular CBO failed: %v", err)
	}
	snapshotCBO, err := cboSvc.AnalyzeSnapshot(ctx, snapshot, cboReq)
	if err != nil {
		t.Fatalf("snapshot CBO failed: %v", err)
	}
	if len(regularCBO.Classes) != len(snapshotCBO.Classes) {
		t.Fatalf("CBO class count mismatch: regular=%d snapshot=%d", len(regularCBO.Classes), len(snapshotCBO.Classes))
	}

	lcomReq := *domain.DefaultLCOMRequest()
	lcomReq.Paths = paths
	lcomSvc := NewLCOMService()
	regularLCOM, err := lcomSvc.Analyze(ctx, lcomReq)
	if err != nil {
		t.Fatalf("regular LCOM failed: %v", err)
	}
	snapshotLCOM, err := lcomSvc.AnalyzeSnapshot(ctx, snapshot, lcomReq)
	if err != nil {
		t.Fatalf("snapshot LCOM failed: %v", err)
	}
	if len(regularLCOM.Classes) != len(snapshotLCOM.Classes) {
		t.Fatalf("LCOM class count mismatch: regular=%d snapshot=%d", len(regularLCOM.Classes), len(snapshotLCOM.Classes))
	}
}

func writeSnapshotFixture(t *testing.T) string {
	t.Helper()

	source := `import json

class Example:
    def __init__(self, value):
        self.value = value

    def duplicate_one(self, items):
        total = 0
        for item in items:
            if item:
                total += item
        return total

    def duplicate_two(self, items):
        total = 0
        for item in items:
            if item:
                total += item
        return total

def top_level(flag):
    if flag:
        return json.dumps({"ok": True})
    return "{}"
`

	sourcePath := filepath.Join(t.TempDir(), "sample.py")
	if err := os.WriteFile(sourcePath, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	return sourcePath
}
