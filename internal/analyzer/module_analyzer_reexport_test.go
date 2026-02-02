package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCircularDependencyThroughReExport(t *testing.T) {
	dir := t.TempDir()

	// Create package structure:
	// pkg_a/__init__.py: from .module_x import SomeClass
	// pkg_a/module_x.py: from pkg_b.file import OtherClass
	// pkg_b/__init__.py: (empty)
	// pkg_b/file.py: from pkg_a import SomeClass
	//
	// Expected cycle: pkg_a.module_x -> pkg_b.file -> pkg_a.module_x

	// Create directories
	pkgADir := filepath.Join(dir, "pkg_a")
	pkgBDir := filepath.Join(dir, "pkg_b")
	if err := os.MkdirAll(pkgADir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_a directory: %v", err)
	}
	if err := os.MkdirAll(pkgBDir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_b directory: %v", err)
	}

	// pkg_a/__init__.py - re-exports SomeClass from module_x
	pkgAInit := `from .module_x import SomeClass
`
	if err := os.WriteFile(filepath.Join(pkgADir, "__init__.py"), []byte(pkgAInit), 0o644); err != nil {
		t.Fatalf("failed to write pkg_a/__init__.py: %v", err)
	}

	// pkg_a/module_x.py - imports from pkg_b.file
	moduleX := `from pkg_b.file import OtherClass

class SomeClass:
    pass
`
	if err := os.WriteFile(filepath.Join(pkgADir, "module_x.py"), []byte(moduleX), 0o644); err != nil {
		t.Fatalf("failed to write pkg_a/module_x.py: %v", err)
	}

	// pkg_b/__init__.py - empty
	if err := os.WriteFile(filepath.Join(pkgBDir, "__init__.py"), []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write pkg_b/__init__.py: %v", err)
	}

	// pkg_b/file.py - imports SomeClass from pkg_a (should resolve to pkg_a.module_x)
	pkgBFile := `from pkg_a import SomeClass

class OtherClass:
    pass
`
	if err := os.WriteFile(filepath.Join(pkgBDir, "file.py"), []byte(pkgBFile), 0o644); err != nil {
		t.Fatalf("failed to write pkg_b/file.py: %v", err)
	}

	// Analyze the project
	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	// Run circular dependency detection
	detector := NewCircularDependencyDetector(graph)
	result := detector.DetectCircularDependencies()

	// We expect to find a circular dependency
	if !result.HasCircularDependencies {
		t.Errorf("expected to detect circular dependency through re-export")
		t.Logf("Graph nodes: %v", getNodeNames(graph))
		t.Logf("Graph edges:")
		for _, edge := range graph.Edges {
			t.Logf("  %s -> %s", edge.From, edge.To)
		}
	}

	// The cycle should involve pkg_a.module_x and pkg_b.file
	foundExpectedCycle := false
	for _, cycle := range result.CircularDependencies {
		moduleXInCycle := false
		pkgBFileInCycle := false
		for _, module := range cycle.Modules {
			if module == "pkg_a.module_x" {
				moduleXInCycle = true
			}
			if module == "pkg_b.file" {
				pkgBFileInCycle = true
			}
		}
		if moduleXInCycle && pkgBFileInCycle {
			foundExpectedCycle = true
			break
		}
	}

	if !foundExpectedCycle {
		t.Errorf("expected cycle to contain pkg_a.module_x and pkg_b.file")
		for _, cycle := range result.CircularDependencies {
			t.Logf("  Cycle: %v", cycle.Modules)
		}
	}
}

func TestNoFalsePositiveForInternalReExports(t *testing.T) {
	dir := t.TempDir()

	// Create package structure with internal re-exports only (no external cycle):
	// pkg_a/__init__.py: from .module_x import SomeClass
	// pkg_a/module_x.py: class SomeClass: pass
	//
	// This should NOT create a cycle (pkg_a -> pkg_a.module_x is internal structure)

	pkgADir := filepath.Join(dir, "pkg_a")
	if err := os.MkdirAll(pkgADir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_a directory: %v", err)
	}

	pkgAInit := `from .module_x import SomeClass
`
	if err := os.WriteFile(filepath.Join(pkgADir, "__init__.py"), []byte(pkgAInit), 0o644); err != nil {
		t.Fatalf("failed to write pkg_a/__init__.py: %v", err)
	}

	moduleX := `class SomeClass:
    pass
`
	if err := os.WriteFile(filepath.Join(pkgADir, "module_x.py"), []byte(moduleX), 0o644); err != nil {
		t.Fatalf("failed to write pkg_a/module_x.py: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	detector := NewCircularDependencyDetector(graph)
	result := detector.DetectCircularDependencies()

	// Should NOT find any circular dependencies
	if result.HasCircularDependencies {
		t.Errorf("unexpected circular dependency detected in internal re-export")
		for _, cycle := range result.CircularDependencies {
			t.Logf("  Cycle: %v", cycle.Modules)
		}
	}
}

func TestReExportResolutionInGraph(t *testing.T) {
	dir := t.TempDir()

	// Create a simple re-export scenario and verify the graph edges
	// pkg_a/__init__.py: from .module_x import SomeClass
	// pkg_a/module_x.py: class SomeClass: pass
	// consumer.py: from pkg_a import SomeClass

	pkgADir := filepath.Join(dir, "pkg_a")
	if err := os.MkdirAll(pkgADir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_a directory: %v", err)
	}

	pkgAInit := `from .module_x import SomeClass
`
	if err := os.WriteFile(filepath.Join(pkgADir, "__init__.py"), []byte(pkgAInit), 0o644); err != nil {
		t.Fatalf("failed to write pkg_a/__init__.py: %v", err)
	}

	moduleX := `class SomeClass:
    pass
`
	if err := os.WriteFile(filepath.Join(pkgADir, "module_x.py"), []byte(moduleX), 0o644); err != nil {
		t.Fatalf("failed to write pkg_a/module_x.py: %v", err)
	}

	consumer := `from pkg_a import SomeClass
`
	if err := os.WriteFile(filepath.Join(dir, "consumer.py"), []byte(consumer), 0o644); err != nil {
		t.Fatalf("failed to write consumer.py: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{ProjectRoot: dir})
	if err != nil {
		t.Fatalf("failed to create analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("AnalyzeProject failed: %v", err)
	}

	// Verify that consumer.py has a dependency on pkg_a.module_x (resolved through re-export)
	consumerNode := graph.GetModule("consumer")
	if consumerNode == nil {
		t.Fatalf("consumer module not found in graph")
	}

	// The dependency should be resolved to pkg_a.module_x, not just pkg_a
	if !consumerNode.Dependencies["pkg_a.module_x"] {
		t.Errorf("expected consumer to depend on pkg_a.module_x (resolved through re-export)")
		t.Logf("Consumer dependencies: %v", consumerNode.Dependencies)
	}
}

func getNodeNames(graph *DependencyGraph) []string {
	var names []string
	for name := range graph.Nodes {
		names = append(names, name)
	}
	return names
}
