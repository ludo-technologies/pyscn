package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
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

func TestModuleAnalyzerCollectsImportsFromCompoundBranches(t *testing.T) {
	dir := t.TempDir()

	consumer := filepath.Join(dir, "consumer.py")

	source := []byte(`
if ready:
    pass
else:
    import else_target

try:
    pass
except Exception:
    import except_target
finally:
    import finally_target
`)
	if err := os.WriteFile(consumer, source, 0o644); err != nil {
		t.Fatalf("failed to write consumer: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	imports := collectImportsForTest(t, analyzer, consumer)
	statements := make(map[string]bool, len(imports))
	for _, imp := range imports {
		statements[imp.Statement] = true
	}

	for _, statement := range []string{"import else_target", "import except_target", "import finally_target"} {
		if !statements[statement] {
			t.Fatalf("expected import %s, got %v", statement, statements)
		}
	}
}

func TestModuleAnalyzerDetectsCycleThroughCompoundBranchImport(t *testing.T) {
	dir := t.TempDir()

	moduleA := filepath.Join(dir, "a.py")
	moduleB := filepath.Join(dir, "b.py")

	sourceA := []byte(`
if ready:
    pass
else:
    import b
`)
	if err := os.WriteFile(moduleA, sourceA, 0o644); err != nil {
		t.Fatalf("failed to write module a: %v", err)
	}
	if err := os.WriteFile(moduleB, []byte("import a\n"), 0o644); err != nil {
		t.Fatalf("failed to write module b: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeFiles([]string{moduleA, moduleB})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}

	result := DetectCircularDependencies(graph)
	if !result.HasCircularDependencies {
		aNode := graph.GetModule("a")
		bNode := graph.GetModule("b")
		if aNode == nil || bNode == nil {
			t.Fatalf("expected modules a and b in graph")
		}
		t.Fatalf("expected cycle through else import, got dependencies a=%v b=%v",
			aNode.Dependencies,
			bNode.Dependencies)
	}
}

func TestModuleAnalyzerResolvesSingleDotRelativeImportInCurrentPackage(t *testing.T) {
	dir := t.TempDir()

	initFile := filepath.Join(dir, "pkg", "__init__.py")
	consumer := filepath.Join(dir, "pkg", "consumer.py")
	target := filepath.Join(dir, "pkg", "target.py")

	for _, path := range []string{initFile, consumer, target} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
	}
	if err := os.WriteFile(initFile, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write package init: %v", err)
	}
	if err := os.WriteFile(consumer, []byte("from .target import Thing\n"), 0o644); err != nil {
		t.Fatalf("failed to write consumer: %v", err)
	}
	if err := os.WriteFile(target, []byte("class Thing:\n    pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write target: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	imports := collectImportsForTest(t, analyzer, consumer)
	if len(imports) != 1 || !imports[0].IsRelative || imports[0].Level != 1 || imports[0].Statement != "target" {
		t.Fatalf("unexpected relative import metadata: %#v", imports)
	}
	if resolved := analyzer.resolveImport(imports[0], consumer); resolved != "pkg.target" {
		t.Fatalf("relative import resolved to %q, want pkg.target", resolved)
	}

	graph, err := analyzer.AnalyzeFiles([]string{initFile, consumer, target})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}

	fromModule := analyzer.filePathToModuleName(consumer)
	toModule := analyzer.filePathToModuleName(target)
	node := graph.Nodes[fromModule]
	if node == nil {
		t.Fatalf("expected module %s in graph", fromModule)
	}
	if !node.Dependencies[toModule] {
		t.Fatalf("expected dependency from %s to %s, got %v", fromModule, toModule, node.Dependencies)
	}

	var edge *DependencyEdge
	for _, candidate := range graph.Edges {
		if candidate.From == fromModule && candidate.To == toModule {
			edge = candidate
			break
		}
	}
	if edge == nil {
		t.Fatalf("expected edge from %s to %s", fromModule, toModule)
	}
	if edge.EdgeType != DependencyEdgeRelative {
		t.Fatalf("expected relative edge, got %s", edge.EdgeType)
	}
	if edge.ImportInfo == nil || !edge.ImportInfo.IsRelative || edge.ImportInfo.Level != 1 {
		t.Fatalf("expected level-1 relative import info, got %#v", edge.ImportInfo)
	}
}

func TestModuleAnalyzerResolvesSameImportNameFromEachPackage(t *testing.T) {
	dir := t.TempDir()

	userA := filepath.Join(dir, "pkg_a", "user.py")
	utilA := filepath.Join(dir, "pkg_a", "util.py")
	userB := filepath.Join(dir, "pkg_b", "user.py")
	utilB := filepath.Join(dir, "pkg_b", "util.py")

	for _, path := range []string{userA, utilA, userB, utilB} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
	}
	if err := os.WriteFile(userA, []byte("import util\n"), 0o644); err != nil {
		t.Fatalf("failed to write pkg_a user: %v", err)
	}
	if err := os.WriteFile(utilA, []byte("A = 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write pkg_a util: %v", err)
	}
	if err := os.WriteFile(userB, []byte("import util\n"), 0o644); err != nil {
		t.Fatalf("failed to write pkg_b user: %v", err)
	}
	if err := os.WriteFile(utilB, []byte("B = 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write pkg_b util: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeFiles([]string{userA, utilA, userB, utilB})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}

	assertDependency := func(fromFile, toFile string) {
		t.Helper()
		fromModule := analyzer.filePathToModuleName(fromFile)
		toModule := analyzer.filePathToModuleName(toFile)
		node := graph.Nodes[fromModule]
		if node == nil {
			t.Fatalf("expected module %s in graph", fromModule)
		}
		if !node.Dependencies[toModule] {
			t.Fatalf("expected dependency from %s to %s, got %v", fromModule, toModule, node.Dependencies)
		}
	}

	assertDependency(userA, utilA)
	assertDependency(userB, utilB)
}

func TestModuleAnalyzerResolvesStubOnlyModule(t *testing.T) {
	dir := t.TempDir()

	consumer := filepath.Join(dir, "pkg", "consumer.py")
	stub := filepath.Join(dir, "pkg", "contracts.pyi")

	for _, path := range []string{consumer, stub} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
	}
	if err := os.WriteFile(consumer, []byte("from .contracts import Service\n"), 0o644); err != nil {
		t.Fatalf("failed to write consumer: %v", err)
	}
	if err := os.WriteFile(stub, []byte("class Service: ...\n"), 0o644); err != nil {
		t.Fatalf("failed to write contracts stub: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeFiles([]string{consumer, stub})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}

	fromModule := analyzer.filePathToModuleName(consumer)
	toModule := analyzer.filePathToModuleName(stub)
	if toModule != "pkg.contracts" {
		t.Fatalf("stub module name = %q, want pkg.contracts", toModule)
	}

	node := graph.Nodes[fromModule]
	if node == nil {
		t.Fatalf("expected module %s in graph", fromModule)
	}
	if !node.Dependencies[toModule] {
		t.Fatalf("expected dependency from %s to %s, got %v", fromModule, toModule, node.Dependencies)
	}
}

func TestModuleAnalyzerResolvesAbsoluteFromImportToSubmodule(t *testing.T) {
	dir := t.TempDir()

	initFile := filepath.Join(dir, "pkg", "__init__.py")
	consumer := filepath.Join(dir, "consumer.py")
	submodule := filepath.Join(dir, "pkg", "submodule.py")

	for _, path := range []string{initFile, consumer, submodule} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
	}
	if err := os.WriteFile(initFile, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write package init: %v", err)
	}
	if err := os.WriteFile(consumer, []byte("from pkg import submodule\n"), 0o644); err != nil {
		t.Fatalf("failed to write consumer: %v", err)
	}
	if err := os.WriteFile(submodule, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write submodule: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	consumerNode := graph.GetModule("consumer")
	if consumerNode == nil {
		t.Fatalf("expected consumer module in graph, got %v", graph.GetModuleNames())
	}
	if !consumerNode.Dependencies["pkg.submodule"] {
		t.Fatalf("expected consumer to depend on pkg.submodule, got %v", consumerNode.Dependencies)
	}
	if consumerNode.Dependencies["pkg"] {
		t.Fatalf("did not expect consumer to depend on package fallback when submodule exists: %v", consumerNode.Dependencies)
	}
}

func TestModuleAnalyzerResolvesRelativeFromImportToSubmodule(t *testing.T) {
	dir := t.TempDir()

	initFile := filepath.Join(dir, "pkg", "__init__.py")
	consumer := filepath.Join(dir, "pkg", "consumer.py")
	submodule := filepath.Join(dir, "pkg", "submodule.py")

	for _, path := range []string{initFile, consumer, submodule} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
	}
	if err := os.WriteFile(initFile, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write package init: %v", err)
	}
	if err := os.WriteFile(consumer, []byte("from . import submodule\n"), 0o644); err != nil {
		t.Fatalf("failed to write consumer: %v", err)
	}
	if err := os.WriteFile(submodule, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write submodule: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	consumerNode := graph.GetModule("pkg.consumer")
	if consumerNode == nil {
		t.Fatalf("expected pkg.consumer module in graph, got %v", graph.GetModuleNames())
	}
	if !consumerNode.Dependencies["pkg.submodule"] {
		t.Fatalf("expected pkg.consumer to depend on pkg.submodule, got %v", consumerNode.Dependencies)
	}
	if consumerNode.Dependencies["pkg"] {
		t.Fatalf("did not expect pkg.consumer to depend on package fallback when submodule exists: %v", consumerNode.Dependencies)
	}
}

func TestModuleAnalyzerSkipsPackageInitSubmoduleImportDependency(t *testing.T) {
	dir := t.TempDir()

	initFile := filepath.Join(dir, "pkg", "__init__.py")
	submodule := filepath.Join(dir, "pkg", "submodule.py")

	for _, path := range []string{initFile, submodule} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
	}
	if err := os.WriteFile(initFile, []byte("from . import submodule\n"), 0o644); err != nil {
		t.Fatalf("failed to write package init: %v", err)
	}
	if err := os.WriteFile(submodule, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write submodule: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	packageNode := graph.GetModule("pkg")
	if packageNode == nil {
		t.Fatalf("expected pkg module in graph, got %v", graph.GetModuleNames())
	}
	if packageNode.Dependencies["pkg.submodule"] {
		t.Fatalf("did not expect package init to depend on own submodule, got %v", packageNode.Dependencies)
	}
}

func TestModuleAnalyzerAnalyzeProjectCollectsNestedStubModules(t *testing.T) {
	dir := t.TempDir()

	stub := filepath.Join(dir, "pkg", "contracts.pyi")
	if err := os.MkdirAll(filepath.Dir(stub), 0o755); err != nil {
		t.Fatalf("failed to create package directory: %v", err)
	}
	if err := os.WriteFile(stub, []byte("class Service: ...\n"), 0o644); err != nil {
		t.Fatalf("failed to write contracts stub: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	if graph.Nodes["pkg.contracts"] == nil {
		t.Fatalf("expected pkg.contracts stub module in graph, got %v", graph.GetModuleNames())
	}
}

func TestModuleAnalyzerPrefersRuntimeModuleWhenStubAlsoExists(t *testing.T) {
	dir := t.TempDir()

	consumer := filepath.Join(dir, "pkg", "consumer.py")
	runtimeModule := filepath.Join(dir, "pkg", "contracts.py")
	stub := filepath.Join(dir, "pkg", "contracts.pyi")

	for _, path := range []string{consumer, runtimeModule, stub} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
	}
	if err := os.WriteFile(consumer, []byte("from .contracts import RuntimeContract\n"), 0o644); err != nil {
		t.Fatalf("failed to write consumer: %v", err)
	}
	if err := os.WriteFile(runtimeModule, []byte("class RuntimeContract:\n    pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write runtime module: %v", err)
	}
	if err := os.WriteFile(stub, []byte("class StubContract: ...\n"), 0o644); err != nil {
		t.Fatalf("failed to write stub: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	module := graph.GetModule("pkg.contracts")
	if module == nil {
		t.Fatalf("expected pkg.contracts module in graph, got %v", graph.GetModuleNames())
	}
	if module.FilePath != runtimeModule {
		t.Fatalf("module file path = %q, want runtime file %q", module.FilePath, runtimeModule)
	}
	if module.ClassCount != 1 {
		t.Fatalf("class count = %d, want runtime metadata only", module.ClassCount)
	}
	if len(module.PublicNames) != 1 || module.PublicNames[0] != "RuntimeContract" {
		t.Fatalf("public names = %v, want runtime metadata only", module.PublicNames)
	}

	consumerNode := graph.GetModule("pkg.consumer")
	if consumerNode == nil || !consumerNode.Dependencies["pkg.contracts"] {
		t.Fatalf("expected consumer to depend on runtime module, got %#v", consumerNode)
	}
}

func TestModuleAnalyzerHonorsExplicitFalseOptions(t *testing.T) {
	dir := t.TempDir()

	module := filepath.Join(dir, "module.py")
	if err := os.WriteFile(module, []byte("import requests\n"), 0o644); err != nil {
		t.Fatalf("failed to write module: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{
		ProjectRoot:       dir,
		IncludeThirdParty: domain.BoolPtr(false),
	})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeFiles([]string{module})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}

	node := graph.Nodes[analyzer.filePathToModuleName(module)]
	if node == nil {
		t.Fatalf("expected module in graph")
	}
	if len(node.Dependencies) != 0 {
		t.Fatalf("expected third-party import to be excluded, got %v", node.Dependencies)
	}
}

func TestModuleAnalyzerHonorsExplicitEmptyExcludePatterns(t *testing.T) {
	dir := t.TempDir()

	module := filepath.Join(dir, "test_contract.py")
	if err := os.WriteFile(module, []byte("class Contract:\n    pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write module: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{
		ProjectRoot:     dir,
		ExcludePatterns: []string{},
	})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	if graph.Nodes["test_contract"] == nil {
		t.Fatalf("expected explicit empty excludes to include test_contract, got %v", graph.GetModuleNames())
	}
}

func TestModuleAnalyzerEmptyIncludesStillCollectOnlyPythonModules(t *testing.T) {
	dir := t.TempDir()

	pythonFile := filepath.Join(dir, "module.py")
	textFile := filepath.Join(dir, "README.md")
	if err := os.WriteFile(pythonFile, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write module: %v", err)
	}
	if err := os.WriteFile(textFile, []byte("# not python\n"), 0o644); err != nil {
		t.Fatalf("failed to write text file: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{
		ProjectRoot:       dir,
		IncludePatterns:   []string{},
		ExcludePatterns:   []string{},
		IncludeStdLib:     domain.BoolPtr(false),
		FollowRelative:    domain.BoolPtr(true),
		IncludeThirdParty: domain.BoolPtr(true),
	})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	if graph.TotalModules != 1 || graph.GetModule("module") == nil {
		t.Fatalf("expected only Python module, got %v", graph.GetModuleNames())
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

func TestModuleAnalyzerResolvesSrcLayoutAbsoluteImport(t *testing.T) {
	dir := t.TempDir()

	serviceModule := filepath.Join(dir, "src", "myapp", "service.py")
	repoModule := filepath.Join(dir, "src", "myapp", "repo.py")
	for _, path := range []string{serviceModule, repoModule} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("failed to create directory for %s: %v", path, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "myapp", "__init__.py"), []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write package init: %v", err)
	}
	if err := os.WriteFile(serviceModule, []byte("from myapp import repo\n"), 0o644); err != nil {
		t.Fatalf("failed to write service module: %v", err)
	}
	if err := os.WriteFile(repoModule, []byte("value = 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write repo module: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeFiles([]string{serviceModule, repoModule})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}

	serviceNode := graph.GetModule("myapp.service")
	if serviceNode == nil {
		t.Fatalf("expected myapp.service module, got %v", graph.GetModuleNames())
	}
	if !serviceNode.Dependencies["myapp.repo"] {
		t.Fatalf("expected myapp.service to depend on myapp.repo; dependencies: %v", serviceNode.Dependencies)
	}
	if graph.GetModule("src.myapp.service") != nil {
		t.Fatalf("did not expect src-prefixed module name in src-layout graph")
	}
}

func TestModuleAnalyzerKeepsSrcPackagePrefixWhenSrcIsPackage(t *testing.T) {
	dir := t.TempDir()

	srcInit := filepath.Join(dir, "src", "__init__.py")
	module := filepath.Join(dir, "src", "module.py")
	if err := os.MkdirAll(filepath.Dir(module), 0o755); err != nil {
		t.Fatalf("failed to create src package: %v", err)
	}
	if err := os.WriteFile(srcInit, []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write src init: %v", err)
	}
	if err := os.WriteFile(module, []byte("value = 1\n"), 0o644); err != nil {
		t.Fatalf("failed to write module: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeFiles([]string{srcInit, module})
	if err != nil {
		t.Fatalf("AnalyzeFiles failed: %v", err)
	}
	if graph.GetModule("src.module") == nil {
		t.Fatalf("expected src.module in graph, got %v", graph.GetModuleNames())
	}
}

func collectImportsForTest(t *testing.T, analyzer *ModuleAnalyzer, path string) []*ImportInfo {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", path, err)
	}
	result, err := parser.New().Parse(context.Background(), content)
	if err != nil {
		t.Fatalf("failed to parse %s: %v", path, err)
	}
	return analyzer.collectModuleFacts(result.AST).imports
}
