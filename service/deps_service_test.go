package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
)

func writePy(t *testing.T, dir, rel, content string) string {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return path
}

func TestDependencyService_Edges(t *testing.T) {
	tmp := t.TempDir()
	// pkg/a.py imports pkg.b
	writePy(t, tmp, "pkg/a.py", "import pkg.b\n")
	writePy(t, tmp, "pkg/b.py", "# noop\n")

	svc := NewDependencyService()
	req := domain.DependencyRequest{
		Paths:           []string{tmp},
		Recursive:       true,
		IncludePatterns: []string{"*.py"},
		ExcludePatterns: []string{},
	}
	resp, err := svc.Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	if len(resp.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d: %#v", len(resp.Edges), resp.Edges)
	}
	if resp.Edges[0].From != "pkg.a" || resp.Edges[0].To != "pkg.b" {
		t.Fatalf("unexpected edge: %#v", resp.Edges[0])
	}
}

func TestDependencyService_Cycles(t *testing.T) {
	tmp := t.TempDir()
	// Simple cycle: x <-> y
	writePy(t, tmp, "pkg/x.py", "import pkg.y\n")
	writePy(t, tmp, "pkg/y.py", "import pkg.x\n")

	svc := NewDependencyService()
	req := domain.DependencyRequest{
		Paths:           []string{tmp},
		Recursive:       true,
		IncludePatterns: []string{"*.py"},
	}
	resp, err := svc.Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	if len(resp.Cycles) < 1 {
		t.Fatalf("expected at least 1 cycle, got %d", len(resp.Cycles))
	}
	// Check that a cycle contains pkg.x and pkg.y
	found := false
	for _, c := range resp.Cycles {
		if len(c.Modules) == 2 {
			a, b := c.Modules[0], c.Modules[1]
			if (a == "pkg.x" && b == "pkg.y") || (a == "pkg.y" && b == "pkg.x") {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("expected cycle involving pkg.x and pkg.y, got %#v", resp.Cycles)
	}
}

func TestDependencyService_LayerViolations(t *testing.T) {
	tmp := t.TempDir()
	// presentation.controller -> domain.model (violation: presentation should depend only on application)
	writePy(t, tmp, "pkg/presentation/controller.py", "import pkg.domain.model\n")
	writePy(t, tmp, "pkg/domain/model.py", "# domain model\n")
	writePy(t, tmp, "pkg/application/usecase.py", "# application usecase\n")

	svc := NewDependencyService()
	arch := &domain.ArchitectureConfigSpec{
		Layers: []domain.ArchitectureLayer{
			{Name: "presentation", Packages: []string{"pkg.presentation.**"}},
			{Name: "application", Packages: []string{"pkg.application.**"}},
			{Name: "domain", Packages: []string{"pkg.domain.**"}},
		},
		Rules: []domain.ArchitectureRule{
			{From: "presentation", Allow: []string{"application"}},
			{From: "application", Allow: []string{"domain"}},
			{From: "domain", Allow: []string{}},
		},
	}
	req := domain.DependencyRequest{
		Paths:           []string{tmp},
		Recursive:       true,
		IncludePatterns: []string{"*.py"},
		Architecture:    arch,
	}
	resp, err := svc.Analyze(context.Background(), req)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	if resp.Summary.LayerViolations < 1 {
		t.Fatalf("expected at least 1 layer violation, got %d", resp.Summary.LayerViolations)
	}
}
