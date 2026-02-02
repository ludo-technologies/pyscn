package analyzer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReExportResolverBasic(t *testing.T) {
	dir := t.TempDir()

	// Create package structure:
	// pkg_a/__init__.py: from .module_x import SomeClass
	// pkg_a/module_x.py: class SomeClass: pass
	pkgDir := filepath.Join(dir, "pkg_a")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_a directory: %v", err)
	}

	initContent := `from .module_x import SomeClass
`
	if err := os.WriteFile(filepath.Join(pkgDir, "__init__.py"), []byte(initContent), 0o644); err != nil {
		t.Fatalf("failed to write __init__.py: %v", err)
	}

	moduleContent := `class SomeClass:
    pass
`
	if err := os.WriteFile(filepath.Join(pkgDir, "module_x.py"), []byte(moduleContent), 0o644); err != nil {
		t.Fatalf("failed to write module_x.py: %v", err)
	}

	resolver := NewReExportResolver(dir)

	// Test resolution
	sourceModule, found := resolver.ResolveReExport("pkg_a", "SomeClass")
	if !found {
		t.Fatal("expected to find SomeClass in pkg_a re-exports")
	}
	if sourceModule != "pkg_a.module_x" {
		t.Fatalf("expected source module 'pkg_a.module_x', got '%s'", sourceModule)
	}
}

func TestReExportResolverMultipleNames(t *testing.T) {
	dir := t.TempDir()

	pkgDir := filepath.Join(dir, "pkg_a")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_a directory: %v", err)
	}

	initContent := `from .module_x import ClassA, ClassB
from .module_y import ClassC
`
	if err := os.WriteFile(filepath.Join(pkgDir, "__init__.py"), []byte(initContent), 0o644); err != nil {
		t.Fatalf("failed to write __init__.py: %v", err)
	}

	if err := os.WriteFile(filepath.Join(pkgDir, "module_x.py"), []byte("class ClassA: pass\nclass ClassB: pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write module_x.py: %v", err)
	}

	if err := os.WriteFile(filepath.Join(pkgDir, "module_y.py"), []byte("class ClassC: pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write module_y.py: %v", err)
	}

	resolver := NewReExportResolver(dir)

	tests := []struct {
		name           string
		expectedModule string
	}{
		{"ClassA", "pkg_a.module_x"},
		{"ClassB", "pkg_a.module_x"},
		{"ClassC", "pkg_a.module_y"},
	}

	for _, tc := range tests {
		sourceModule, found := resolver.ResolveReExport("pkg_a", tc.name)
		if !found {
			t.Errorf("expected to find %s in pkg_a re-exports", tc.name)
			continue
		}
		if sourceModule != tc.expectedModule {
			t.Errorf("for %s: expected source module '%s', got '%s'", tc.name, tc.expectedModule, sourceModule)
		}
	}
}

func TestReExportResolverWithAll(t *testing.T) {
	dir := t.TempDir()

	pkgDir := filepath.Join(dir, "pkg_a")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_a directory: %v", err)
	}

	// __all__ limits what's considered public
	initContent := `__all__ = ['PublicClass']

from .module_x import PublicClass, PrivateClass
`
	if err := os.WriteFile(filepath.Join(pkgDir, "__init__.py"), []byte(initContent), 0o644); err != nil {
		t.Fatalf("failed to write __init__.py: %v", err)
	}

	if err := os.WriteFile(filepath.Join(pkgDir, "module_x.py"), []byte("class PublicClass: pass\nclass PrivateClass: pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write module_x.py: %v", err)
	}

	resolver := NewReExportResolver(dir)

	// PublicClass should be found
	sourceModule, found := resolver.ResolveReExport("pkg_a", "PublicClass")
	if !found {
		t.Fatal("expected to find PublicClass in pkg_a re-exports")
	}
	if sourceModule != "pkg_a.module_x" {
		t.Fatalf("expected source module 'pkg_a.module_x', got '%s'", sourceModule)
	}

	// PrivateClass should NOT be found (not in __all__)
	_, found = resolver.ResolveReExport("pkg_a", "PrivateClass")
	if found {
		t.Fatal("expected PrivateClass to NOT be found (not in __all__)")
	}
}

func TestReExportResolverCaching(t *testing.T) {
	dir := t.TempDir()

	pkgDir := filepath.Join(dir, "pkg_a")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_a directory: %v", err)
	}

	initContent := `from .module_x import SomeClass
`
	if err := os.WriteFile(filepath.Join(pkgDir, "__init__.py"), []byte(initContent), 0o644); err != nil {
		t.Fatalf("failed to write __init__.py: %v", err)
	}

	if err := os.WriteFile(filepath.Join(pkgDir, "module_x.py"), []byte("class SomeClass: pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write module_x.py: %v", err)
	}

	resolver := NewReExportResolver(dir)

	// First call
	map1, err := resolver.GetReExportMap("pkg_a")
	if err != nil {
		t.Fatalf("first GetReExportMap failed: %v", err)
	}

	// Second call should return cached result
	map2, err := resolver.GetReExportMap("pkg_a")
	if err != nil {
		t.Fatalf("second GetReExportMap failed: %v", err)
	}

	// Should be the same pointer (cached)
	if map1 != map2 {
		t.Fatal("expected cached result, got different pointers")
	}
}

func TestReExportResolverNoInitFile(t *testing.T) {
	dir := t.TempDir()

	resolver := NewReExportResolver(dir)

	// Should return empty for non-existent package
	_, found := resolver.ResolveReExport("nonexistent_pkg", "SomeClass")
	if found {
		t.Fatal("expected not to find re-export for non-existent package")
	}
}

func TestReExportResolverNestedPackage(t *testing.T) {
	dir := t.TempDir()

	// Create nested package: pkg_a/sub/__init__.py
	subDir := filepath.Join(dir, "pkg_a", "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_a/sub directory: %v", err)
	}

	// Parent package
	if err := os.WriteFile(filepath.Join(dir, "pkg_a", "__init__.py"), []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write pkg_a/__init__.py: %v", err)
	}

	// Sub-package re-exports
	initContent := `from .impl import Thing
`
	if err := os.WriteFile(filepath.Join(subDir, "__init__.py"), []byte(initContent), 0o644); err != nil {
		t.Fatalf("failed to write sub/__init__.py: %v", err)
	}

	if err := os.WriteFile(filepath.Join(subDir, "impl.py"), []byte("class Thing: pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write impl.py: %v", err)
	}

	resolver := NewReExportResolver(dir)

	sourceModule, found := resolver.ResolveReExport("pkg_a.sub", "Thing")
	if !found {
		t.Fatal("expected to find Thing in pkg_a.sub re-exports")
	}
	if sourceModule != "pkg_a.sub.impl" {
		t.Fatalf("expected source module 'pkg_a.sub.impl', got '%s'", sourceModule)
	}
}
