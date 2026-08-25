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

func TestDIAntipatternServiceAnalyzeSnapshotHonorsFileSelection(t *testing.T) {
	projectRoot := t.TempDir()
	includedPath := filepath.Join(projectRoot, "included.py")
	excludedPath := filepath.Join(projectRoot, "migrations", "excluded.py")
	if err := os.MkdirAll(filepath.Dir(excludedPath), 0o755); err != nil {
		t.Fatalf("create migrations directory: %v", err)
	}
	source := []byte("class Service:\n    def __init__(self, a, b, c, d, e, f):\n        pass\n")
	for _, path := range []string{includedPath, excludedPath} {
		if err := os.WriteFile(path, source, 0o644); err != nil {
			t.Fatalf("write source %s: %v", path, err)
		}
	}
	snapshot := BuildProjectSnapshotWithOptions(context.Background(), []string{includedPath, excludedPath}, ProjectSnapshotOptions{ProjectRoot: projectRoot})
	req := *domain.DefaultDIAntipatternRequest()
	req.IncludePatterns = []string{"included.py", "**/migrations/**"}
	req.ExcludePatterns = []string{"**/migrations/**"}

	response, err := NewDIAntipatternService().AnalyzeSnapshot(context.Background(), snapshot, req)
	if err != nil {
		t.Fatalf("analyze snapshot: %v", err)
	}
	if response.Summary.FilesAnalyzed != 1 {
		t.Fatalf("expected one selected implementation file, got %+v", response.Summary)
	}
	for _, finding := range response.Findings {
		if finding.Location.FilePath != includedPath {
			t.Fatalf("unexpected finding outside the selected file: %+v", finding)
		}
	}
}
