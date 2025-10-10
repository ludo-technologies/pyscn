package analyzer

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/parser"
)

func TestTypeCheckingImports(t *testing.T) {
	// Create a temporary Python file with TYPE_CHECKING imports
	content := `# Test file for TYPE_CHECKING imports
import os
import sys
from typing import TYPE_CHECKING

# Regular imports that should be included
from collections import defaultdict
import json

# TYPE_CHECKING imports that should be ignored for circular dependency detection
if TYPE_CHECKING:
    from typing import List, Dict, Optional
    from some.circular.dependency import CircularClass
    import circular_module

# Another TYPE_CHECKING block
if TYPE_CHECKING:
    from another.circular import AnotherCircular

# Regular code after TYPE_CHECKING
def some_function():
    pass

# Nested TYPE_CHECKING (should also be detected)
def some_other_function():
    if TYPE_CHECKING:
        from nested.circular import NestedCircular
    pass

# Not TYPE_CHECKING - should be included
if sys.version_info >= (3, 8):
    from new_feature import something

# Complex TYPE_CHECKING condition - should still be detected
if TYPE_CHECKING and sys.version_info >= (3, 9):
    from complex.circular import ComplexCircular
`

	// Create temporary directory and file
	tmpDir, err := os.MkdirTemp("", "pyscn_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test_module.py")
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create module analyzer
	options := &ModuleAnalysisOptions{
		ProjectRoot:       tmpDir,
		IncludePatterns:   []string{"**/*.py"},
		ExcludePatterns:   []string{},
		IncludeStdLib:     false,
		IncludeThirdParty: true,
		FollowRelative:    true,
	}

	analyzer, err := NewModuleAnalyzer(options)
	if err != nil {
		t.Fatalf("Failed to create module analyzer: %v", err)
	}

	// Parse the file and collect imports
	fileContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	p := parser.New()
	ctx := context.Background()
	result, err := p.Parse(ctx, fileContent)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	imports := analyzer.collectModuleImports(result.AST, testFile)

	// Verify the imports
	typeCheckingImports := 0
	regularImports := 0

	expectedTypeCheckingImports := map[string]bool{
		"typing":                   true, // from typing import List, Dict, Optional
		"some.circular.dependency": true, // from some.circular.dependency import CircularClass
		"import circular_module":   true, // import circular_module (note: includes "import " prefix)
		"another.circular":         true, // from another.circular import AnotherCircular
		"nested.circular":          true, // from nested.circular import NestedCircular (nested)
		"complex.circular":         true, // from complex.circular import ComplexCircular (complex condition)
	}

	expectedRegularImports := map[string]bool{
		"import os":   true, // import os (note: includes "import " prefix)
		"import sys":  true, // import sys (note: includes "import " prefix)
		"typing":      true, // from typing import TYPE_CHECKING (this should not be TYPE_CHECKING)
		"collections": true, // from collections import defaultdict
		"import json": true, // import json (note: includes "import " prefix)
		"new_feature": true, // from new_feature import something (not in TYPE_CHECKING)
	}

	for _, imp := range imports {
		t.Logf("Import: '%s' | IsTypeChecking: %t | Line: %d", imp.Statement, imp.IsTypeChecking, imp.Line)
		if imp.IsTypeChecking {
			typeCheckingImports++
			if !expectedTypeCheckingImports[imp.Statement] {
				t.Errorf("Unexpected TYPE_CHECKING import: %s", imp.Statement)
			}
		} else {
			regularImports++
			if !expectedRegularImports[imp.Statement] {
				t.Logf("Regular import: %s (line %d)", imp.Statement, imp.Line)
			}
		}
	}

	t.Logf("Found %d TYPE_CHECKING imports and %d regular imports", typeCheckingImports, regularImports)

	// Check that we found the expected number of TYPE_CHECKING imports
	expectedTypeCheckingCount := len(expectedTypeCheckingImports)
	if typeCheckingImports != expectedTypeCheckingCount {
		t.Errorf("Expected %d TYPE_CHECKING imports, got %d", expectedTypeCheckingCount, typeCheckingImports)
	}

	// Verify that regular imports are not marked as TYPE_CHECKING
	if regularImports < 4 { // At least os, sys, collections, json
		t.Errorf("Expected at least 4 regular imports, got %d", regularImports)
	}
}

func TestModuleAnalyzerIgnoresTypeCheckingForCircularDeps(t *testing.T) {
	// Create temporary directory structure with circular dependencies
	tmpDir, err := os.MkdirTemp("", "pyscn_circular_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create module A that imports B in TYPE_CHECKING
	moduleA := `from typing import TYPE_CHECKING
import sys

if TYPE_CHECKING:
    from module_b import ClassB  # This should be ignored for circular dependency detection

def function_a():
    pass
`

	// Create module B that imports A normally (this would create a cycle if TYPE_CHECKING wasn't ignored)
	moduleB := `from module_a import function_a  # This should be counted

def function_b():
    function_a()

class ClassB:
    pass
`

	// Write the modules
	moduleAPath := filepath.Join(tmpDir, "module_a.py")
	moduleBPath := filepath.Join(tmpDir, "module_b.py")

	err = os.WriteFile(moduleAPath, []byte(moduleA), 0644)
	if err != nil {
		t.Fatalf("Failed to write module A: %v", err)
	}

	err = os.WriteFile(moduleBPath, []byte(moduleB), 0644)
	if err != nil {
		t.Fatalf("Failed to write module B: %v", err)
	}

	// Create module analyzer
	options := &ModuleAnalysisOptions{
		ProjectRoot:       tmpDir,
		IncludePatterns:   []string{"**/*.py"},
		ExcludePatterns:   []string{},
		IncludeStdLib:     false,
		IncludeThirdParty: true,
		FollowRelative:    true,
	}

	analyzer, err := NewModuleAnalyzer(options)
	if err != nil {
		t.Fatalf("Failed to create module analyzer: %v", err)
	}

	// Analyze the project
	graph, err := analyzer.AnalyzeProject()
	if err != nil {
		t.Fatalf("Failed to analyze project: %v", err)
	}

	// Check that we have both modules
	moduleANode := graph.GetModule("module_a")
	moduleBNode := graph.GetModule("module_b")

	if moduleANode == nil {
		t.Fatalf("Module A not found in graph")
	}
	if moduleBNode == nil {
		t.Fatalf("Module B not found in graph")
	}

	// Check dependencies - there should be no circular dependency
	// Module B should depend on Module A, but Module A should not depend on Module B
	// (because the import is in TYPE_CHECKING block)

	aDependencies := graph.GetDependencies("module_a")
	bDependencies := graph.GetDependencies("module_b")

	t.Logf("Module A dependencies: %d", len(aDependencies))
	t.Logf("Module B dependencies: %d", len(bDependencies))

	// Module A should have no dependencies to module_b (TYPE_CHECKING import ignored)
	hasAToB := false
	for _, dep := range aDependencies {
		if dep == "module_b" {
			hasAToB = true
			break
		}
	}

	if hasAToB {
		t.Errorf("Module A should not depend on Module B (TYPE_CHECKING import should be ignored)")
	}

	// Module B should depend on Module A (normal import)
	hasBToA := false
	for _, dep := range bDependencies {
		if dep == "module_a" {
			hasBToA = true
			break
		}
	}

	if !hasBToA {
		t.Errorf("Module B should depend on Module A (normal import)")
	}

	// Check for cycles - there should be none
	if graph.HasCycle() {
		cyclicModules := graph.GetModulesInCycles()
		t.Errorf("Found unexpected cycles with modules: %v", cyclicModules)
	}
}
