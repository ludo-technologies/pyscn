package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// TestLazyImportsAreFlagged verifies that imports nested in function/method
// bodies are marked IsLazy, while module-level and class-body imports are not.
func TestLazyImportsAreFlagged(t *testing.T) {
	content := `from collections import defaultdict   # module-level, not lazy

class B:
    from typing import List           # class body, executes at load time -> not lazy

    def expand(self):
        from foo.a import use         # lazy: inside a method body
        return use()

def helper():
    import json                       # lazy: inside a function body
    return json

async def async_helper():
    from foo.c import thing           # lazy: inside an async function body
    return thing
`

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_module.py")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	options := &ModuleAnalysisOptions{
		ProjectRoot:       tmpDir,
		IncludeStdLib:     domain.BoolPtr(false),
		IncludeThirdParty: domain.BoolPtr(true),
		FollowRelative:    domain.BoolPtr(true),
	}
	analyzer, err := NewModuleAnalyzer(options)
	if err != nil {
		t.Fatalf("Failed to create module analyzer: %v", err)
	}

	fileContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	p := parser.New()
	result, err := p.Parse(context.Background(), fileContent)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	imports := analyzer.collectModuleFacts(result.AST).imports

	// Map a recognizable substring of each import to its expected IsLazy value.
	expectedLazy := map[string]bool{
		"collections": false, // module-level
		"typing":      false, // class body executes at load time
		"foo.a":       true,  // inside method
		"import json": true,  // inside function
		"foo.c":       true,  // inside async function
	}

	seen := map[string]bool{}
	for _, imp := range imports {
		want, ok := expectedLazy[imp.Statement]
		if !ok {
			continue
		}
		seen[imp.Statement] = true
		if imp.IsLazy != want {
			t.Errorf("import %q: IsLazy = %t, want %t", imp.Statement, imp.IsLazy, want)
		}
	}

	for stmt := range expectedLazy {
		if !seen[stmt] {
			t.Errorf("expected to find import %q but it was not collected", stmt)
		}
	}
}

// TestModuleAnalyzerIgnoresLazyImportsForCircularDeps reproduces issue #460:
// a module pair where one side imports the other only via a function-body lazy
// import must not be reported as a load-time circular dependency. The lazy edge
// is still retained in the graph (it is a real runtime dependency that matters
// for coupling/architecture analyses) but flagged so cycle detection skips it.
func TestModuleAnalyzerIgnoresLazyImportsForCircularDeps(t *testing.T) {
	tmpDir := t.TempDir()

	pkgDir := filepath.Join(tmpDir, "foo")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("Failed to create package dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "__init__.py"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write __init__.py: %v", err)
	}

	// a.py imports b at the top level.
	moduleA := `from .b import B   # top-level

def use():
    return B()
`
	// b.py imports a back, but only lazily inside a method body.
	moduleB := `class B:
    def expand(self):
        from .a import use   # lazy, inside method
        return use()
`

	if err := os.WriteFile(filepath.Join(pkgDir, "a.py"), []byte(moduleA), 0644); err != nil {
		t.Fatalf("Failed to write a.py: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "b.py"), []byte(moduleB), 0644); err != nil {
		t.Fatalf("Failed to write b.py: %v", err)
	}

	options := &ModuleAnalysisOptions{
		ProjectRoot:       tmpDir,
		IncludeStdLib:     domain.BoolPtr(false),
		IncludeThirdParty: domain.BoolPtr(true),
		FollowRelative:    domain.BoolPtr(true),
	}
	analyzer, err := NewModuleAnalyzer(options)
	if err != nil {
		t.Fatalf("Failed to create module analyzer: %v", err)
	}

	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("Failed to analyze project: %v", err)
	}

	// Both directions must remain as real runtime edges in the graph, so that
	// coupling metrics, dependency matrices, and architecture layer checks
	// still see the dependency.
	if !graph.Nodes["foo.b"].Dependencies["foo.a"] {
		t.Errorf("foo.b -> foo.a edge should be retained in the graph (real runtime dependency)")
	}
	if !graph.Nodes["foo.a"].Dependencies["foo.b"] {
		t.Errorf("foo.a -> foo.b edge should be present (top-level import)")
	}

	// foo.b -> foo.a must be flagged lazy; foo.a -> foo.b must not be.
	if !graph.Nodes["foo.b"].LazyDependencies["foo.a"] {
		t.Errorf("foo.b -> foo.a should be flagged as a lazy dependency")
	}
	if graph.Nodes["foo.a"].LazyDependencies["foo.b"] {
		t.Errorf("foo.a -> foo.b (top-level import) should NOT be flagged lazy")
	}

	loadTimeGraph := loadTimeDependencyGraph{graph}
	if successors := loadTimeGraph.Successors("foo.b"); len(successors) != 0 {
		t.Errorf("load-time successors of foo.b = %v, want none", successors)
	}
	if predecessors := loadTimeGraph.Predecessors("foo.a"); len(predecessors) != 0 {
		t.Errorf("load-time predecessors of foo.a = %v, want none", predecessors)
	}
	predecessors := loadTimeGraph.Predecessors("foo.b")
	if len(predecessors) != 1 || predecessors[0] != "foo.a" {
		t.Errorf("load-time predecessors of foo.b = %v, want [foo.a]", predecessors)
	}

	// Cycle detection must NOT report a cycle, because the only edge closing the
	// loop is lazy (function-body).
	result := NewCircularDependencyDetector(graph).DetectCircularDependencies()
	if result.HasCircularDependencies {
		t.Errorf("expected no load-time cycle, got %d: %+v", result.TotalCycles, result.CircularDependencies)
	}
}

// TestLazyImportPromotedToLoadTimeWhenAlsoImportedAtModuleLevel verifies that a
// pair which has BOTH a lazy and a module-level import to the same target is
// treated as a real load-time dependency (and so can form a cycle).
func TestLazyImportPromotedToLoadTimeWhenAlsoImportedAtModuleLevel(t *testing.T) {
	tmpDir := t.TempDir()
	pkgDir := filepath.Join(tmpDir, "foo")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatalf("Failed to create package dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "__init__.py"), []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write __init__.py: %v", err)
	}

	moduleA := `from .b import B   # top-level

def use():
    return B()
`
	// b.py imports a both lazily AND at the top level -> real load-time cycle.
	moduleB := `from .a import use   # top-level

class B:
    def expand(self):
        from .a import use   # also lazy
        return use()
`
	if err := os.WriteFile(filepath.Join(pkgDir, "a.py"), []byte(moduleA), 0644); err != nil {
		t.Fatalf("Failed to write a.py: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "b.py"), []byte(moduleB), 0644); err != nil {
		t.Fatalf("Failed to write b.py: %v", err)
	}

	analyzer, err := NewModuleAnalyzer(&ModuleAnalysisOptions{
		ProjectRoot:       tmpDir,
		IncludeStdLib:     domain.BoolPtr(false),
		IncludeThirdParty: domain.BoolPtr(true),
		FollowRelative:    domain.BoolPtr(true),
	})
	if err != nil {
		t.Fatalf("Failed to create module analyzer: %v", err)
	}
	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("Failed to analyze project: %v", err)
	}

	// The module-level import must promote the pair out of lazy-only status.
	if graph.Nodes["foo.b"].LazyDependencies["foo.a"] {
		t.Errorf("foo.b -> foo.a should NOT be lazy-only (it also has a top-level import)")
	}

	result := NewCircularDependencyDetector(graph).DetectCircularDependencies()
	if !result.HasCircularDependencies {
		t.Errorf("expected a load-time cycle because both modules import each other at module level")
	}
}
