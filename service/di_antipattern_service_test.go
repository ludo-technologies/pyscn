package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestDIAntipatternServiceAnalyzeSnapshotDoesNotReadSourceFiles(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "service.py")
	source := "class Service:\n    def __init__(self, a, b, c, d, e, f):\n        pass\n"
	if err := os.WriteFile(sourcePath, []byte(source), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	snapshot := BuildProjectSnapshot(context.Background(), []string{sourcePath})
	if err := os.Remove(sourcePath); err != nil {
		t.Fatalf("remove captured source: %v", err)
	}
	req := *domain.DefaultDIAntipatternRequest()

	response, err := NewDIAntipatternService().AnalyzeSnapshot(context.Background(), snapshot, req)
	if err != nil {
		t.Fatalf("analyze snapshot: %v", err)
	}
	if response.Summary.FilesAnalyzed != 1 {
		t.Fatalf("expected one captured file analyzed, got %+v", response.Summary)
	}
}
