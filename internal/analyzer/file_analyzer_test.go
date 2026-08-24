package analyzer

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

func TestFileComplexityAnalyzerReportsClassScopeKind(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.py")
	if err := os.WriteFile(path, []byte("class Config:\n    if enabled:\n        value = 1\n"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Output.Format = "json"
	cfg.Output.MinComplexity = 1
	var output bytes.Buffer
	analyzer, err := NewFileComplexityAnalyzer(cfg, &output)
	if err != nil {
		t.Fatalf("create analyzer: %v", err)
	}
	if err := analyzer.AnalyzeFile(path); err != nil {
		t.Fatalf("analyze file: %v", err)
	}

	var report struct {
		Results []struct {
			FunctionName string                   `json:"function_name"`
			ScopeKind    domain.AnalysisScopeKind `json:"scope_kind"`
		} `json:"results"`
		ClassScopes []struct {
			FunctionName string                   `json:"function_name"`
			ScopeKind    domain.AnalysisScopeKind `json:"scope_kind"`
			FilePath     string                   `json:"file_path"`
		} `json:"class_scopes"`
	}
	if err := json.Unmarshal(output.Bytes(), &report); err != nil {
		t.Fatalf("decode report: %v", err)
	}
	if len(report.ClassScopes) != 1 {
		t.Fatalf("class_scopes = %+v", report.ClassScopes)
	}
	classScope := report.ClassScopes[0]
	if classScope.FunctionName != "Config" || classScope.ScopeKind != domain.AnalysisScopeClass {
		t.Fatalf("class scope = %+v", classScope)
	}
	if classScope.FilePath != path {
		t.Fatalf("class source path = %q, want %q", classScope.FilePath, path)
	}
}
