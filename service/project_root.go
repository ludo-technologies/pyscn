package service

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// projectRootMarkers name files whose presence marks a Python project root.
var projectRootMarkers = []string{"setup.py", "pyproject.toml", "setup.cfg", ".git", "requirements.txt"}

// FindProjectRoot locates the project root for the analyzed paths.
//
// An explicit project_root in the discovered config always wins. Otherwise the
// nearest ancestor of the analyzed paths holding a project marker is the root.
// The walk stops at the discovered config file's directory: an ancestor config
// found above a marker (a vendored package, a monorepo subpackage) must not
// relocate the root, or absolute imports stop resolving (issue #753). Without
// any marker the root is the directory holding the analyzed top-level package
// when there is one, and the config directory otherwise.
func FindProjectRoot(paths []string) string {
	target := commonAnalysisParent(paths)
	if target == "" {
		cwd, _ := os.Getwd()
		return cwd
	}

	configDir := ""
	if configPath := discoverConfigFile(target); configPath != "" {
		if cfg, err := config.NewTomlConfigLoader().LoadConfig(configPath); err == nil && cfg.ProjectRoot != "" {
			return cfg.ProjectRoot
		}
		if absConfigPath, err := filepath.Abs(configPath); err == nil {
			configDir = filepath.Dir(absConfigPath)
		}
	}

	for dir := target; ; dir = filepath.Dir(dir) {
		// A marker inside a package (a requirements.txt shipped with the
		// package, say) does not make that package the root: module names
		// would lose the package prefix. A repository checkout with a stray
		// __init__.py is still the root unless its sources import it as a
		// package (issue #885).
		if hasProjectMarker(dir) && (!isPythonPackage(dir) || isRepositoryImportRoot(dir)) {
			return dir
		}
		if dir == configDir || filepath.Dir(dir) == dir {
			break
		}
	}

	importRoot := packageContainer(target, configDir)
	if configDir == "" || holdsTopLevelPackage(importRoot) {
		return importRoot
	}
	return configDir
}

// packageContainer climbs out of the package tree holding dir: a project root
// inside a package would strip the package prefix from every module name.
// The climb stops at bound when set.
func packageContainer(dir, bound string) string {
	for dir != bound && isPythonPackage(dir) {
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return dir
}

func hasProjectMarker(dir string) bool {
	for _, marker := range projectRootMarkers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}

var (
	fromImportLine = regexp.MustCompile(`^\s*from\s+(\.*)\s*([\w.]*)\s+import\b`)
	importLine     = regexp.MustCompile(`^\s*import\s+(.+)`)
)

// isRepositoryImportRoot reports whether dir, a package directory, is a
// repository checkout whose sources never import it as a package. Such a
// checkout is run from inside (python main.py) and its __init__.py takes part
// in no import, so the root must not climb above the repository.
func isRepositoryImportRoot(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		return false
	}
	files, err := NewFileReader().collectFromDirectory(dir, true, dir, domain.PythonFileSelection{})
	if err != nil {
		return false
	}
	for _, file := range files {
		if importsCheckoutAsPackage(file, dir) {
			return false
		}
	}
	return true
}

// importsCheckoutAsPackage reports whether file imports checkout by its own
// name or through a relative import that climbs to the checkout directory.
func importsCheckoutAsPackage(file, checkout string) bool {
	content, err := os.ReadFile(file)
	if err != nil {
		return false
	}
	relDir, err := filepath.Rel(checkout, filepath.Dir(file))
	if err != nil {
		return false
	}
	depth := 0
	if relDir != "." {
		depth = strings.Count(relDir, string(filepath.Separator)) + 1
	}
	name := filepath.Base(checkout)
	for line := range strings.Lines(string(content)) {
		if match := fromImportLine.FindStringSubmatch(line); match != nil {
			level := len(match[1])
			if level > depth || (level == 0 && topLevelModule(match[2]) == name) {
				return true
			}
			continue
		}
		if match := importLine.FindStringSubmatch(line); match != nil {
			for _, imported := range strings.Split(match[1], ",") {
				if fields := strings.Fields(imported); len(fields) > 0 && topLevelModule(fields[0]) == name {
					return true
				}
			}
		}
	}
	return false
}

func topLevelModule(module string) string {
	top, _, _ := strings.Cut(module, ".")
	return top
}

// holdsTopLevelPackage reports whether dir is not a package itself but directly
// contains one, the shape of a checkout whose import root is the checkout.
func holdsTopLevelPackage(dir string) bool {
	if isPythonPackage(dir) {
		return false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() && isPythonPackage(filepath.Join(dir, entry.Name())) {
			return true
		}
	}
	return false
}

func isPythonPackage(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "__init__.py"))
	return err == nil
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
