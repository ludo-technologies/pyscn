package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModuleAnalyzerResolvesPlainImportWithinProject(t *testing.T) {
	dir := t.TempDir()

	moduleA := filepath.Join(dir, "module_a.py")
	moduleB := filepath.Join(dir, "module_b.py")

	if err := os.WriteFile(moduleA, []byte("import module_b\n\n"), 0o644); err != nil {
		t.Fatalf("failed to write module_a: %v", err)
	}
	if err := os.WriteFile(moduleB, []byte("value = 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write module_b: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeFiles([]string{moduleA, moduleB})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}

	fromModule := analyzer.filePathToModuleName(moduleA)
	toModule := analyzer.filePathToModuleName(moduleB)

	node := graph.Nodes[fromModule]
	if node == nil {
		t.Fatalf("expected module %s in graph", fromModule)
	}
	if !node.Dependencies[toModule] {
		t.Fatalf("expected dependency from %s to %s, got %v", fromModule, toModule, node.Dependencies)
	}
}

func TestModuleAnalyzerResolvesImportWithAlias(t *testing.T) {
	dir := t.TempDir()

	moduleA := filepath.Join(dir, "module_alias.py")
	moduleB := filepath.Join(dir, "module_target.py")

	if err := os.WriteFile(moduleA, []byte("import module_target as target\n"), 0o644); err != nil {
		t.Fatalf("failed to write module_alias: %v", err)
	}
	if err := os.WriteFile(moduleB, []byte("value = 2\n"), 0o644); err != nil {
		t.Fatalf("failed to write module_target: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeFiles([]string{moduleA, moduleB})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}

	fromModule := analyzer.filePathToModuleName(moduleA)
	toModule := analyzer.filePathToModuleName(moduleB)

	node := graph.Nodes[fromModule]
	if node == nil {
		t.Fatalf("expected module %s in graph", fromModule)
	}
	if !node.Dependencies[toModule] {
		t.Fatalf("expected dependency from %s to %s, got %v", fromModule, toModule, node.Dependencies)
	}
}
