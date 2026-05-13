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

func TestModuleAnalyzerExtractsAbstractClassCount(t *testing.T) {
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "contracts.py")
	source := []byte(`
from abc import ABC, ABCMeta, abstractmethod
import abc

class Repository(ABC):
    pass

class Service(abc.ABC):
    pass

class Controller(metaclass=ABCMeta):
    pass

class Worker:
    @abstractmethod
    def run(self):
        pass

class AsyncWorker:
    @abc.abstractmethod
    async def run(self):
        pass

class Concrete:
    def run(self):
        pass
`)

	if err := os.WriteFile(modulePath, source, 0o644); err != nil {
		t.Fatalf("failed to write contracts module: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeFiles([]string{modulePath})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}

	moduleName := analyzer.filePathToModuleName(modulePath)
	node := graph.Nodes[moduleName]
	if node == nil {
		t.Fatalf("expected module %s in graph", moduleName)
	}

	if node.ClassCount != 6 {
		t.Fatalf("expected 6 classes, got %d", node.ClassCount)
	}
	if node.AbstractClassCount != 5 {
		t.Fatalf("expected 5 abstract classes, got %d", node.AbstractClassCount)
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

func TestModuleAnalyzerDoesNotResolveStdlibImportToShadowingProjectModule(t *testing.T) {
	dir := t.TempDir()

	shadowModule := filepath.Join(dir, "src", "mypkg", "time.py")
	samePackageUser := filepath.Join(dir, "src", "mypkg", "widget.py")
	otherPackageUser := filepath.Join(dir, "utils", "serve.py")

	for _, path := range []string{shadowModule, samePackageUser, otherPackageUser} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
	}
	if err := os.WriteFile(shadowModule, []byte(`"""Project-local time picker module."""`), 0o644); err != nil {
		t.Fatalf("failed to write shadow module: %v", err)
	}
	// Python 3 absolute imports should treat this as stdlib time, not
	// src.mypkg.time, even though a same-basename project module exists.
	if err := os.WriteFile(samePackageUser, []byte("import time\n"), 0o644); err != nil {
		t.Fatalf("failed to write same-package user: %v", err)
	}
	if err := os.WriteFile(otherPackageUser, []byte("import time\n"), 0o644); err != nil {
		t.Fatalf("failed to write other-package user: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeFiles([]string{samePackageUser, shadowModule, otherPackageUser})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}

	shadowName := analyzer.filePathToModuleName(shadowModule)
	for _, path := range []string{samePackageUser, otherPackageUser} {
		moduleName := analyzer.filePathToModuleName(path)
		node := graph.Nodes[moduleName]
		if node == nil {
			t.Fatalf("expected module %s in graph", moduleName)
		}
		if node.Dependencies[shadowName] {
			t.Fatalf("did not expect %s to depend on stdlib-shadowing module %s; dependencies: %v", moduleName, shadowName, node.Dependencies)
		}
	}
}
