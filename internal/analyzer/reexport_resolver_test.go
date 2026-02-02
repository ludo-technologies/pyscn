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

func TestReExportResolverLevel2RelativeImport(t *testing.T) {
	dir := t.TempDir()

	// Create nested package structure:
	// pkg_a/__init__.py: (empty)
	// pkg_a/sub/__init__.py: from ..sibling import Thing (Level 2 relative import)
	// pkg_a/sibling.py: class Thing: pass
	pkgADir := filepath.Join(dir, "pkg_a")
	subDir := filepath.Join(pkgADir, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	// Parent package
	if err := os.WriteFile(filepath.Join(pkgADir, "__init__.py"), []byte(""), 0o644); err != nil {
		t.Fatalf("failed to write pkg_a/__init__.py: %v", err)
	}

	// Sibling module
	if err := os.WriteFile(filepath.Join(pkgADir, "sibling.py"), []byte("class Thing: pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write sibling.py: %v", err)
	}

	// Sub-package with level 2 relative import (from .. import sibling)
	initContent := `from ..sibling import Thing
`
	if err := os.WriteFile(filepath.Join(subDir, "__init__.py"), []byte(initContent), 0o644); err != nil {
		t.Fatalf("failed to write sub/__init__.py: %v", err)
	}

	resolver := NewReExportResolver(dir)

	sourceModule, found := resolver.ResolveReExport("pkg_a.sub", "Thing")
	if !found {
		t.Fatal("expected to find Thing in pkg_a.sub re-exports (level 2 relative import)")
	}
	// from ..sibling means: go up 2 levels from pkg_a.sub -> pkg_a, then .sibling -> pkg_a.sibling
	if sourceModule != "pkg_a.sibling" {
		t.Fatalf("expected source module 'pkg_a.sibling', got '%s'", sourceModule)
	}
}

func TestReExportResolverLevel3RelativeImport(t *testing.T) {
	dir := t.TempDir()

	// Create deeply nested package structure:
	// root/__init__.py: (empty)
	// root/pkg_a/__init__.py: (empty)
	// root/pkg_a/sub/__init__.py: (empty)
	// root/pkg_a/sub/deep/__init__.py: from ...sibling import Thing (Level 3)
	// root/pkg_a/sibling.py: class Thing: pass
	rootDir := filepath.Join(dir, "root")
	pkgADir := filepath.Join(rootDir, "pkg_a")
	subDir := filepath.Join(pkgADir, "sub")
	deepDir := filepath.Join(subDir, "deep")
	if err := os.MkdirAll(deepDir, 0o755); err != nil {
		t.Fatalf("failed to create directories: %v", err)
	}

	// Create __init__.py files
	for _, d := range []string{rootDir, pkgADir, subDir} {
		if err := os.WriteFile(filepath.Join(d, "__init__.py"), []byte(""), 0o644); err != nil {
			t.Fatalf("failed to write __init__.py in %s: %v", d, err)
		}
	}

	// Sibling module at pkg_a level
	if err := os.WriteFile(filepath.Join(pkgADir, "sibling.py"), []byte("class Thing: pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write sibling.py: %v", err)
	}

	// Deep package with level 3 relative import
	initContent := `from ...sibling import Thing
`
	if err := os.WriteFile(filepath.Join(deepDir, "__init__.py"), []byte(initContent), 0o644); err != nil {
		t.Fatalf("failed to write deep/__init__.py: %v", err)
	}

	resolver := NewReExportResolver(dir)

	// from ...sibling in root.pkg_a.sub.deep means:
	// go up 3 levels -> root.pkg_a, then .sibling -> root.pkg_a.sibling
	sourceModule, found := resolver.ResolveReExport("root.pkg_a.sub.deep", "Thing")
	if !found {
		t.Fatal("expected to find Thing in root.pkg_a.sub.deep re-exports (level 3 relative import)")
	}
	if sourceModule != "root.pkg_a.sibling" {
		t.Fatalf("expected source module 'root.pkg_a.sibling', got '%s'", sourceModule)
	}
}

func TestReExportResolverLevelExceedsDepth(t *testing.T) {
	dir := t.TempDir()

	// Create shallow package with too-deep relative import:
	// pkg_a/__init__.py: from ...too_deep import Thing (Level 3, but only 1 level deep)
	pkgDir := filepath.Join(dir, "pkg_a")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_a directory: %v", err)
	}

	// This import is invalid - goes beyond project root
	initContent := `from ...too_deep import Thing
`
	if err := os.WriteFile(filepath.Join(pkgDir, "__init__.py"), []byte(initContent), 0o644); err != nil {
		t.Fatalf("failed to write __init__.py: %v", err)
	}

	resolver := NewReExportResolver(dir)

	// Should NOT find this re-export since level (3) > package depth (1)
	_, found := resolver.ResolveReExport("pkg_a", "Thing")
	if found {
		t.Fatal("expected re-export to NOT be found when level exceeds package depth")
	}
}

func TestReExportResolverWithAlias(t *testing.T) {
	dir := t.TempDir()

	// Create package with aliased re-export:
	// pkg_a/__init__.py: from .module_x import OriginalClass as AliasedClass
	pkgDir := filepath.Join(dir, "pkg_a")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_a directory: %v", err)
	}

	initContent := `from .module_x import OriginalClass as AliasedClass
`
	if err := os.WriteFile(filepath.Join(pkgDir, "__init__.py"), []byte(initContent), 0o644); err != nil {
		t.Fatalf("failed to write __init__.py: %v", err)
	}

	if err := os.WriteFile(filepath.Join(pkgDir, "module_x.py"), []byte("class OriginalClass: pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write module_x.py: %v", err)
	}

	resolver := NewReExportResolver(dir)

	// Should find AliasedClass (the alias name)
	sourceModule, found := resolver.ResolveReExport("pkg_a", "AliasedClass")
	if !found {
		t.Fatal("expected to find AliasedClass in pkg_a re-exports")
	}
	if sourceModule != "pkg_a.module_x" {
		t.Fatalf("expected source module 'pkg_a.module_x', got '%s'", sourceModule)
	}

	// OriginalClass should NOT be found (it's aliased)
	_, found = resolver.ResolveReExport("pkg_a", "OriginalClass")
	if found {
		t.Fatal("expected OriginalClass to NOT be found (it's exported as AliasedClass)")
	}
}

func TestReExportResolverWildcardNotSupported(t *testing.T) {
	dir := t.TempDir()

	// Create package with wildcard re-export:
	// pkg_a/__init__.py: from .module_x import *
	pkgDir := filepath.Join(dir, "pkg_a")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("failed to create pkg_a directory: %v", err)
	}

	initContent := `from .module_x import *
`
	if err := os.WriteFile(filepath.Join(pkgDir, "__init__.py"), []byte(initContent), 0o644); err != nil {
		t.Fatalf("failed to write __init__.py: %v", err)
	}

	if err := os.WriteFile(filepath.Join(pkgDir, "module_x.py"), []byte("class SomeClass: pass\n"), 0o644); err != nil {
		t.Fatalf("failed to write module_x.py: %v", err)
	}

	resolver := NewReExportResolver(dir)

	// Wildcard imports are not tracked - should NOT find any re-exports
	_, found := resolver.ResolveReExport("pkg_a", "SomeClass")
	if found {
		t.Fatal("expected wildcard re-exports to NOT be tracked")
	}
}
