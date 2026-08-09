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

// FindAnalysisRoot returns the narrowest common directory explicitly selected
// by the caller, without widening the scope through project markers.
func FindAnalysisRoot(paths []string) string {
	if len(paths) == 0 {
		cwd, _ := os.Getwd()
		return cwd
	}

	directories := make([]string, 0, len(paths))
	for _, path := range paths {
		absolute, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absolute); err == nil && !info.IsDir() {
			absolute = filepath.Dir(absolute)
		}
		directories = append(directories, filepath.Clean(absolute))
	}
	if len(directories) == 0 {
		cwd, _ := os.Getwd()
		return cwd
	}

	root := directories[0]
	for _, directory := range directories[1:] {
		for !pathWithinDirectory(directory, root) {
			parent := filepath.Dir(root)
			if parent == root {
				break
			}
			root = parent
		}
	}
	return root
}

func pathWithinDirectory(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
