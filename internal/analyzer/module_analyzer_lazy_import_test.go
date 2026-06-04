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
// import must not be reported as a load-time circular dependency.
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

	// b.a -> a must NOT be recorded (lazy import).
	bDeps := graph.GetDependencies("foo.b")
	for _, dep := range bDeps {
		if dep == "foo.a" {
			t.Errorf("foo.b should not depend on foo.a (the import is lazy / function-body)")
		}
	}

	// a -> b must still be recorded (top-level import).
	aDeps := graph.GetDependencies("foo.a")
	hasAToB := false
	for _, dep := range aDeps {
		if dep == "foo.b" {
			hasAToB = true
		}
	}
	if !hasAToB {
		t.Errorf("foo.a should depend on foo.b (top-level import)")
	}

	// No load-time cycle should exist.
	if graph.HasCycle() {
		t.Errorf("Found unexpected cycle: %v", graph.GetModulesInCycles())
	}
}
