package service

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ludo-technologies/pyscn/internal/config"
)

// FindProjectRoot locates the project root from the given paths by finding their
// discovered config file first, then falling back to the common parent and
// standard Python project markers.
func FindProjectRoot(paths []string) string {
	if len(paths) == 0 {
		cwd, _ := os.Getwd()
		return cwd
	}

	if configuredRoot := configuredProjectRoot(paths); configuredRoot != "" {
		return configuredRoot
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
		for !pathWithinDirectory(path, commonParent) {
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

		if !pathWithinDirectory(absPaths[0], parent) {
			break
		}

		commonParent = parent
	}

	return commonParent
}

func configuredProjectRoot(paths []string) string {
	searchPath := commonAnalysisParent(paths)
	if searchPath == "" {
		searchPath = "."
	}

	loader := config.NewTomlConfigLoader()
	configPath := loader.FindConfigFileFromPath(searchPath)
	if configPath == "" {
		return ""
	}

	cfg, err := loader.LoadConfig(configPath)
	if err != nil {
		return ""
	}
	return cfg.ProjectRoot
}

func commonAnalysisParent(paths []string) string {
	var absPaths []string
	for _, path := range paths {
		if path == "" {
			continue
		}
		absolute, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		if info, err := os.Stat(absolute); err == nil && !info.IsDir() {
			absolute = filepath.Dir(absolute)
		}
		absPaths = append(absPaths, absolute)
	}
	if len(absPaths) == 0 {
		return ""
	}

	common := absPaths[0]
	for _, path := range absPaths[1:] {
		for !pathWithinDirectory(path, common) {
			parent := filepath.Dir(common)
			if parent == common {
				break
			}
			common = parent
		}
	}
	return common
}

func pathWithinDirectory(path, root string) bool {
	relative, err := filepath.Rel(root, path)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
