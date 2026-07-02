package service

import (
	"os"
	"path/filepath"
	"strings"
)

// FindProjectRoot locates the project root from the given paths by finding their
// common parent and walking upward for standard Python project markers.
func FindProjectRoot(paths []string) string {
	if len(paths) == 0 {
		cwd, _ := os.Getwd()
		return cwd
	}

	absPaths := make([]string, 0, len(paths))
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			continue
		}

		info, err := os.Stat(absPath)
		if err == nil && !info.IsDir() {
			absPath = filepath.Dir(absPath)
		}

		absPaths = append(absPaths, absPath)
	}

	if len(absPaths) == 0 {
		cwd, _ := os.Getwd()
		return cwd
	}

	commonParent := absPaths[0]
	for _, path := range absPaths[1:] {
		for !strings.HasPrefix(path, commonParent) {
			commonParent = filepath.Dir(commonParent)
			if commonParent == "/" || commonParent == "." {
				break
			}
		}
	}

	for {
		markers := []string{"setup.py", "pyproject.toml", "setup.cfg", ".git", "requirements.txt"}
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(commonParent, marker)); err == nil {
				return commonParent
			}
		}

		parent := filepath.Dir(commonParent)
		if parent == commonParent || parent == "/" || parent == "." {
			break
		}

		if !strings.HasPrefix(absPaths[0], parent) {
			break
		}

		commonParent = parent
	}

	return commonParent
}
