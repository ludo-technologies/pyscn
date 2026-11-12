package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindProjectRoot_FromSubpackageFiles(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, "src", "myapp")
	require.NoError(t, os.MkdirAll(srcDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte("[project]\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "a.py"), []byte("pass\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "b.py"), []byte("pass\n"), 0o644))

	got := FindProjectRoot([]string{
		filepath.Join(srcDir, "a.py"),
		filepath.Join(srcDir, "b.py"),
	})
	assert.Equal(t, root, got)
}

func TestFindProjectRoot_FromSiblingPackages(t *testing.T) {
	root := t.TempDir()
	pkgA := filepath.Join(root, "src", "pkg_a")
	pkgB := filepath.Join(root, "src", "pkg_b")
	require.NoError(t, os.MkdirAll(pkgA, 0o755))
	require.NoError(t, os.MkdirAll(pkgB, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte("[project]\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(pkgA, "a.py"), []byte("pass\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(pkgB, "b.py"), []byte("pass\n"), 0o644))

	got := FindProjectRoot([]string{
		filepath.Join(pkgA, "a.py"),
		filepath.Join(pkgB, "b.py"),
	})
	assert.Equal(t, root, got)
}

func TestFindProjectRoot_UsesPathComponentBoundaries(t *testing.T) {
	root := t.TempDir()
	pkg := filepath.Join(root, "pkg")
	pkgExtra := filepath.Join(root, "pkg_extra")
	require.NoError(t, os.MkdirAll(pkg, 0o755))
	require.NoError(t, os.MkdirAll(pkgExtra, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte("[project]\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(pkg, "pyproject.toml"), []byte("[project]\n"), 0o644))

	got := FindProjectRoot([]string{pkg, pkgExtra})
	assert.Equal(t, root, got)
}

func TestFindProjectRoot_UsesConfiguredProjectRoot(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, "src")
	require.NoError(t, os.MkdirAll(srcDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".pyscn.toml"), []byte("project_root = \"src\"\n"), 0o644))
	file := filepath.Join(root, "package", "module.py")
	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
	require.NoError(t, os.WriteFile(file, []byte("pass\n"), 0o644))

	got := FindProjectRoot([]string{file})
	want, err := filepath.Abs(srcDir)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestFindProjectRoot_DiscoveredConfigWithoutProjectRootUsesConfigDirectory(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, ".pyscn.toml"), []byte("[output]\nmin_complexity = 15\n"), 0o644))
	file := filepath.Join(root, "package", "module.py")
	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
	require.NoError(t, os.WriteFile(file, []byte("pass\n"), 0o644))

	got := FindProjectRoot([]string{file})
	want, err := filepath.Abs(root)
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestFindProjectRoot_MarkerBelowAncestorConfigWins(t *testing.T) {
	outer := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outer, ".pyscn.toml"), []byte("[complexity]\nenabled = true\n"), 0o644))
	target := filepath.Join(outer, "vendor", "lp")
	require.NoError(t, os.MkdirAll(filepath.Join(target, "lp"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "setup.py"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(target, "lp", "__init__.py"), []byte(""), 0o644))

	assert.Equal(t, target, FindProjectRoot([]string{target}))
}

func TestFindProjectRoot_TargetHoldingTopLevelPackageWinsOverAncestorConfig(t *testing.T) {
	outer := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outer, ".pyscn.toml"), []byte("[complexity]\nenabled = true\n"), 0o644))
	target := filepath.Join(outer, "vendor", "lp")
	require.NoError(t, os.MkdirAll(filepath.Join(target, "lp"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "lp", "__init__.py"), []byte(""), 0o644))

	assert.Equal(t, target, FindProjectRoot([]string{target}))
}

func TestFindProjectRoot_PackageInsideAncestorConfigProjectUsesConfigDirectory(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, ".pyscn.toml"), []byte("[complexity]\nenabled = true\n"), 0o644))
	pkg := filepath.Join(root, "pkg")
	require.NoError(t, os.MkdirAll(filepath.Join(pkg, "sub"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkg, "__init__.py"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(pkg, "sub", "__init__.py"), []byte(""), 0o644))

	assert.Equal(t, root, FindProjectRoot([]string{pkg}))
}

func TestFindProjectRoot_NoMarkersFallsBackToAnalyzedDirectory(t *testing.T) {
	target := filepath.Join(t.TempDir(), "a", "b")
	require.NoError(t, os.MkdirAll(target, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(target, "m.py"), []byte("pass\n"), 0o644))

	assert.Equal(t, target, FindProjectRoot([]string{filepath.Join(target, "m.py")}))
}

func TestFindProjectRoot_PackageTargetWithoutMarkersUsesPackageParent(t *testing.T) {
	parent := t.TempDir()
	pkg := filepath.Join(parent, "pkg")
	require.NoError(t, os.MkdirAll(filepath.Join(pkg, "sub"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkg, "__init__.py"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(pkg, "sub", "__init__.py"), []byte(""), 0o644))

	assert.Equal(t, parent, FindProjectRoot([]string{filepath.Join(pkg, "sub")}))
}

func TestFindProjectRoot_MarkerInsidePackageDoesNotMoveRootIntoPackage(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(root, ".pyscn.toml"), []byte("[complexity]\nenabled = true\n"), 0o644))
	pkg := filepath.Join(root, "pkg")
	require.NoError(t, os.MkdirAll(pkg, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(pkg, "__init__.py"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(pkg, "requirements.txt"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(pkg, "a.py"), []byte("pass\n"), 0o644))

	assert.Equal(t, root, FindProjectRoot([]string{filepath.Join(pkg, "a.py")}))
}

func writeProjectTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(root, ".git"), 0o755))
	for name, content := range files {
		path := filepath.Join(root, name)
		require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	}
}

func TestFindProjectRoot_RepositoryWithRootInitStaysRoot(t *testing.T) {
	repo := filepath.Join(t.TempDir(), "myrepo")
	writeProjectTree(t, repo, map[string]string{
		"__init__.py":         "",
		"requirements.txt":    "pygame\n",
		"engine/core/a.py":    "def f():\n    return 1\n",
		"engine/game/b.py":    "from engine.core.a import f\n",
		"main.py":             "from engine.game.b import g\n",
		"engine/game/data.py": "import os, engine.core.a as core\n",
	})

	assert.Equal(t, repo, FindProjectRoot([]string{repo}))
	assert.Equal(t, repo, FindProjectRoot([]string{filepath.Join(repo, "engine", "game", "b.py")}))
}

func TestFindProjectRoot_RepositoryImportedByOwnNameUsesParent(t *testing.T) {
	parent := t.TempDir()
	repo := filepath.Join(parent, "myrepo")
	writeProjectTree(t, repo, map[string]string{
		"__init__.py":     "",
		"sub/__init__.py": "",
		"sub/a.py":        "def f():\n    return 1\n",
		"sub/b.py":        "from myrepo.sub.a import f\n",
	})

	assert.Equal(t, parent, FindProjectRoot([]string{repo}))
}

func TestFindProjectRoot_RepositoryImportedRelativelyUsesParent(t *testing.T) {
	parent := t.TempDir()
	repo := filepath.Join(parent, "addon")
	writeProjectTree(t, repo, map[string]string{
		"__init__.py":     "from . import ui\n",
		"ui.py":           "",
		"sub/__init__.py": "",
		"sub/a.py":        "from .b import g\n",
		"sub/b.py":        "def g():\n    return 1\n",
	})

	assert.Equal(t, parent, FindProjectRoot([]string{repo}))
}

func TestFindProjectRoot_RelativeImportClimbingToRepositoryUsesParent(t *testing.T) {
	parent := t.TempDir()
	repo := filepath.Join(parent, "addon")
	writeProjectTree(t, repo, map[string]string{
		"__init__.py":     "",
		"ui.py":           "",
		"sub/__init__.py": "",
		"sub/a.py":        "from ..ui import draw\n",
	})

	assert.Equal(t, parent, FindProjectRoot([]string{repo}))
}
