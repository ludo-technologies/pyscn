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
