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
