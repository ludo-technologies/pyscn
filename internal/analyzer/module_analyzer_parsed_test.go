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
	parsedModules := parseModuleSources(t, sources)

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

func TestModuleAnalyzer_AnalyzeParsedModulesUsesCapturedReExports(t *testing.T) {
	projectRoot := t.TempDir()
	packageDir := filepath.Join(projectRoot, "pkg")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("create package directory: %v", err)
	}

	sources := map[string]string{
		filepath.Join(projectRoot, "consumer.py"): "from pkg import Target\n",
		filepath.Join(packageDir, "__init__.py"):  "from .target import Target\n",
		filepath.Join(packageDir, "target.py"):    "class Target:\n    pass\n",
	}
	parsedModules := parseModuleSources(t, sources)
	if err := os.Remove(filepath.Join(packageDir, "__init__.py")); err != nil {
		t.Fatalf("remove captured package source: %v", err)
	}
	if err := os.Remove(filepath.Join(packageDir, "target.py")); err != nil {
		t.Fatalf("remove captured target source: %v", err)
	}

	moduleAnalyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: projectRoot})
	if err != nil {
		t.Fatalf("create module analyzer: %v", err)
	}
	graph, err := moduleAnalyzer.AnalyzeParsedModules(context.Background(), parsedModules)
	if err != nil {
		t.Fatalf("analyze parsed modules: %v", err)
	}
	if !graph.Nodes["consumer"].Dependencies["pkg.target"] {
		t.Fatalf("expected captured re-export edge, got %+v", graph.Nodes["consumer"].Dependencies)
	}
}

func TestModuleAnalyzer_AbsoluteThirdPartyImportDoesNotBindByLocalSuffix(t *testing.T) {
	projectRoot := t.TempDir()
	localModule := filepath.Join(projectRoot, "piccolo_api", "fastapi.py")
	consumer := filepath.Join(projectRoot, "piccolo_api", "routers", "admin", "api.py")
	if err := os.MkdirAll(filepath.Dir(consumer), 0o755); err != nil {
		t.Fatalf("create package directories: %v", err)
	}
	sources := map[string]string{
		localModule: "class LocalFastAPI:\n    pass\n",
		consumer:    "from fastapi import FastAPI\n",
	}
	parsedModules := parseModuleSources(t, sources)

	tests := map[string]func(*ModuleAnalyzer) (*DependencyGraph, error){
		"captured": func(moduleAnalyzer *ModuleAnalyzer) (*DependencyGraph, error) {
			return moduleAnalyzer.AnalyzeParsedModules(context.Background(), parsedModules)
		},
		"standalone": func(moduleAnalyzer *ModuleAnalyzer) (*DependencyGraph, error) {
			return moduleAnalyzer.AnalyzeFiles([]string{localModule, consumer})
		},
	}
	for name, analyze := range tests {
		t.Run(name, func(t *testing.T) {
			moduleAnalyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: projectRoot})
			if err != nil {
				t.Fatalf("create module analyzer: %v", err)
			}
			graph, err := analyze(moduleAnalyzer)
			if err != nil {
				t.Fatalf("analyze modules: %v", err)
			}
			consumerName := moduleAnalyzer.filePathToModuleName(consumer)
			if dependencies := graph.Nodes[consumerName].Dependencies; len(dependencies) != 0 {
				t.Fatalf("expected third-party import to remain external, got %v", dependencies)
			}
		})
	}
}

func TestModuleAnalyzer_AnalyzeParsedModulesResolvesSiblingScriptImports(t *testing.T) {
	projectRoot := t.TempDir()
	moduleDir := filepath.Join(projectRoot, "sub")
	if err := os.MkdirAll(moduleDir, 0o755); err != nil {
		t.Fatalf("create module directory: %v", err)
	}

	parsedModules := parseModuleSources(t, map[string]string{
		filepath.Join(moduleDir, "ma.py"): "import mb\n",
		filepath.Join(moduleDir, "mb.py"): "import ma\n",
	})
	moduleAnalyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: projectRoot})
	if err != nil {
		t.Fatalf("create module analyzer: %v", err)
	}

	graph, err := moduleAnalyzer.AnalyzeParsedModules(context.Background(), parsedModules)
	if err != nil {
		t.Fatalf("analyze parsed modules: %v", err)
	}
	if !graph.Nodes["sub.ma"].Dependencies["sub.mb"] {
		t.Fatalf("expected sub.ma to depend on sub.mb, got %v", graph.Nodes["sub.ma"].Dependencies)
	}
	if !graph.Nodes["sub.mb"].Dependencies["sub.ma"] {
		t.Fatalf("expected sub.mb to depend on sub.ma, got %v", graph.Nodes["sub.mb"].Dependencies)
	}
}

func parseModuleSources(t *testing.T, sources map[string]string) []ParsedModule {
	t.Helper()

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
	return parsedModules
}
