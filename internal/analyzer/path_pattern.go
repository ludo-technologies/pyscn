package analyzer

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// ProjectRelativePath returns filePath relative to root with forward slashes.
// File selection patterns match against this form, so directories above the
// project root (a checkout under ~/docs, say) never match a pattern such as
// "**/docs/**".
func ProjectRelativePath(root, filePath string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve project root %s: %w", root, err)
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return "", fmt.Errorf("resolve path %s: %w", filePath, err)
	}
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return "", fmt.Errorf("relate %s to project root %s: %w", filePath, root, err)
	}
	// ToSlash only replaces the platform separator; also fold backslashes so
	// directory globs work for Windows-style paths on every platform.
	return strings.ReplaceAll(filepath.ToSlash(rel), "\\", "/"), nil
}

// MatchPathPattern reports whether a glob pattern matches a project-relative
// path from ProjectRelativePath. A leading "./" anchors the pattern at the
// project root. Bare filename patterns match the basename at any depth.
func MatchPathPattern(pattern, relPath string) bool {
	pattern = strings.TrimPrefix(pattern, "./")
	if matched, _ := doublestar.Match(pattern, relPath); matched {
		return true
	}
	if !strings.ContainsAny(pattern, "/\\") {
		matched, _ := doublestar.Match(pattern, path.Base(relPath))
		return matched
	}
	return false
}
