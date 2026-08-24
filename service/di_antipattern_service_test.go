package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestDIAntipatternServiceAnalyzeSnapshotHonorsCancellationWithoutParsedFiles(t *testing.T) {
	brokenPath := filepath.Join(t.TempDir(), "broken.py")
	if err := os.WriteFile(brokenPath, []byte("def broken(:\n"), 0o644); err != nil {
		t.Fatalf("write broken source: %v", err)
	}
	snapshot := BuildProjectSnapshot(context.Background(), []string{brokenPath})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := NewDIAntipatternService().AnalyzeSnapshot(ctx, snapshot, *domain.DefaultDIAntipatternRequest())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
}

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
