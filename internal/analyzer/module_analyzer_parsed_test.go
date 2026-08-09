package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

func TestModuleAnalyzer_AnalyzeParsedModulesUsesCapturedSyntax(t *testing.T) {
	projectRoot := t.TempDir()
	packageDir := filepath.Join(projectRoot, "pkg")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("create package directory: %v", err)
	}

	sources := map[string]string{
		filepath.Join(packageDir, "a.py"): "from pkg import b\n",
		filepath.Join(packageDir, "b.py"): "VALUE = 1\n",
	}
	parsedModules := make([]ParsedModule, 0, len(sources))
	pyParser := parser.New()
	for path, source := range sources {
		content := []byte(source)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("write Python source: %v", err)
		}
		result, err := pyParser.Parse(context.Background(), content)
		if err != nil {
			t.Fatalf("parse Python source: %v", err)
		}
		parsedModule, err := NewParsedModule(path, content, result.AST)
		if err != nil {
			t.Fatalf("create parsed module: %v", err)
		}
		parsedModules = append(parsedModules, parsedModule)
	}

	if err := os.WriteFile(filepath.Join(packageDir, "a.py"), []byte("VALUE = 2\n"), 0o644); err != nil {
		t.Fatalf("replace captured source: %v", err)
	}

	moduleAnalyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: projectRoot})
	if err != nil {
		t.Fatalf("create module analyzer: %v", err)
	}
	graph, err := moduleAnalyzer.AnalyzeParsedModules(context.Background(), parsedModules)
	if err != nil {
		t.Fatalf("analyze parsed modules: %v", err)
	}
	if !graph.Nodes["pkg.a"].Dependencies["pkg.b"] {
		t.Fatalf("expected captured import edge, got %+v", graph.Nodes["pkg.a"].Dependencies)
	}
}
