package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func TestMockDataServiceUsesRequestDetectorConfig(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "data.py")
	if err := os.WriteFile(sourcePath, []byte("productionfixture = 42\n"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	req := *domain.DefaultMockDataRequest()
	req.Paths = []string{sourcePath}
	req.Keywords = []string{"productionfixture"}

	response, err := NewMockDataService().Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("Analyze returned an error: %v", err)
	}
	if len(response.Files) != 1 || len(response.Files[0].Findings) == 0 {
		t.Fatalf("expected a finding from the request keyword, got: %#v", response.Files)
	}
}

func TestMockDataServiceUsesRequestDomains(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "data.py")
	if err := os.WriteFile(sourcePath, []byte("endpoint = \"corp.internal\"\n"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	req := *domain.DefaultMockDataRequest()
	req.Paths = []string{sourcePath}
	req.Keywords = []string{"not-present"}
	req.Domains = []string{"corp.internal"}

	response, err := NewMockDataService().Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("Analyze returned an error: %v", err)
	}
	if len(response.Files) != 1 || len(response.Files[0].Findings) == 0 {
		t.Fatalf("expected a finding from the request domain, got: %#v", response.Files)
	}
}

func TestMockDataServiceWithConfigKeepsConstructorDetector(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "data.py")
	if err := os.WriteFile(sourcePath, []byte("constructiononly = 42\n"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	req := *domain.DefaultMockDataRequest()
	req.Paths = []string{sourcePath}

	response, err := NewMockDataServiceWithConfig([]string{"constructiononly"}, nil).Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("Analyze returned an error: %v", err)
	}
	if len(response.Files) != 1 || len(response.Files[0].Findings) == 0 {
		t.Fatalf("expected constructor-provided detector config to remain authoritative, got: %#v", response.Files)
	}
}

func TestMockDataServiceUsesIgnorePatterns(t *testing.T) {
	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "ignored.py")
	if err := os.WriteFile(sourcePath, []byte("email = \"test@example.com\"\n"), 0o644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	req := *domain.DefaultMockDataRequest()
	req.Paths = []string{sourcePath}
	req.IgnorePatterns = []string{`ignored\.py$`}

	response, err := NewMockDataService().Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("Analyze returned an error: %v", err)
	}
	if response.Summary.TotalFiles != 0 || len(response.Files) != 0 {
		t.Fatalf("expected ignored file to be skipped, got summary=%#v files=%#v", response.Summary, response.Files)
	}
}

func TestMockDataServiceRejectsInvalidIgnorePattern(t *testing.T) {
	req := *domain.DefaultMockDataRequest()
	req.Paths = []string{"data.py"}
	req.IgnorePatterns = []string{"["}

	if _, err := NewMockDataService().Analyze(context.Background(), req); err == nil {
		t.Fatal("expected invalid ignore pattern to return an error")
	}
}

func TestMockDataServiceReportsParseDiagnostics(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "broken.py")
	if err := os.WriteFile(sourcePath, []byte("def broken(:\n"), 0o644); err != nil {
		t.Fatalf("write broken source: %v", err)
	}
	req := *domain.DefaultMockDataRequest()
	req.Paths = []string{sourcePath}
	req.IgnoreTests = domain.BoolPtr(false)

	response, err := NewMockDataService().Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("analyze broken source: %v", err)
	}
	if len(response.Diagnostics) != 1 || response.Diagnostics[0].Code != domain.DiagnosticCodeParse {
		t.Fatalf("expected typed parse diagnostic, got %+v", response.Diagnostics)
	}
	if len(response.Failures) != 0 {
		t.Fatalf("parse diagnostics must not be execution failures: %+v", response.Failures)
	}
}

func TestMockDataServiceAnalyzeSnapshotDoesNotReadSourceFiles(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "data.py")
	if err := os.WriteFile(sourcePath, []byte("email = \"test@example.com\"\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	snapshot := BuildProjectSnapshot(context.Background(), []string{sourcePath})
	if err := os.Remove(sourcePath); err != nil {
		t.Fatalf("remove captured source: %v", err)
	}
	req := *domain.DefaultMockDataRequest()
	req.IgnoreTests = domain.BoolPtr(false)

	response, err := NewMockDataService().AnalyzeSnapshot(context.Background(), snapshot, req)
	if err != nil {
		t.Fatalf("analyze snapshot: %v", err)
	}
	if response.Summary.TotalFiles != 1 || response.Summary.TotalFindings == 0 {
		t.Fatalf("expected captured source findings, got %+v", response.Summary)
	}
}
