package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

// FileReaderImpl implements the FileReader interface
type FileReaderImpl struct{}

// NewFileReader creates a new file reader service
func NewFileReader() *FileReaderImpl {
	return &FileReaderImpl{}
}

// CollectPythonFiles recursively finds all Python files in the given paths
func (f *FileReaderImpl) CollectPythonFiles(paths []string, recursive bool, includePatterns, excludePatterns []string) ([]string, error) {
	// Validate patterns early to catch common issues
	if err := f.validatePatterns(includePatterns, "include"); err != nil {
		return nil, err
	}
	if err := f.validatePatterns(excludePatterns, "exclude"); err != nil {
		return nil, err
	}

	var files []string

	for _, path := range paths {
		// Check if path exists
		info, err := os.Stat(path)
		if err != nil {
			return nil, domain.NewFileNotFoundError(path, err)
		}

		if info.IsDir() {
			// Process directory
			dirFiles, err := f.collectFromDirectory(path, recursive, includePatterns, excludePatterns)
			if err != nil {
				return nil, err
			}
			files = append(files, dirFiles...)
		} else {
			// Process single file
			if f.IsValidPythonFile(path) && f.shouldIncludeFile(path, includePatterns, excludePatterns) {
				files = append(files, path)
			}
		}
	}

	return files, nil
}

// ReadFile reads the content of a file
func (f *FileReaderImpl) ReadFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, domain.NewFileNotFoundError(path, err)
	}
	return content, nil
}

// IsValidPythonFile checks if a file is a valid Python file
func (f *FileReaderImpl) IsValidPythonFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".py" || ext == ".pyi"
}

// FileExists checks if a file exists
func (f *FileReaderImpl) FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

// collectFromDirectory collects Python files from a directory
func (f *FileReaderImpl) collectFromDirectory(dirPath string, recursive bool, includePatterns, excludePatterns []string) ([]string, error) {
	var files []string

	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log warning but continue processing other files
			return nil
		}

		// Skip directories if not recursive
		if info.IsDir() && !recursive && path != dirPath {
			return filepath.SkipDir
		}

		// Skip hidden directories and files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip common directories that shouldn't contain Python source files
		if info.IsDir() && f.shouldSkipDirectory(info.Name()) {
			return filepath.SkipDir
		}

		// Check if it's a Python file
		if !info.IsDir() && f.IsValidPythonFile(path) {
			if f.shouldIncludeFile(path, includePatterns, excludePatterns) {
				files = append(files, path)
			}
		}

		return nil
	}

	if err := filepath.Walk(dirPath, walkFunc); err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dirPath, err)
	}

	return files, nil
}

// shouldIncludeFile checks if a file should be included based on patterns
func (f *FileReaderImpl) shouldIncludeFile(path string, includePatterns, excludePatterns []string) bool {
	// Check exclude patterns first
	for _, pattern := range excludePatterns {
		if f.matchesPattern(pattern, path) {
			return false
		}
	}

	// If no include patterns specified, include by default
	if len(includePatterns) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range includePatterns {
		if f.matchesPattern(pattern, path) {
			return true
		}
	}

	return false
}

// matchesPattern checks if a path matches a pattern
func (f *FileReaderImpl) matchesPattern(pattern, path string) bool {
	// First try matching against just the filename
	if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
		return true
	}

	// Handle globstar (**) patterns
	if strings.Contains(pattern, "**") {
		return f.matchesGlobstarPattern(pattern, path)
	}

	// For patterns with path separators, try full path matching
	if strings.Contains(pattern, "/") {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
	}

	return false
}

// matchesGlobstarPattern handles patterns with ** (recursive directory matching)
func (f *FileReaderImpl) matchesGlobstarPattern(pattern, path string) bool {
	// Handle simple globstar patterns using stdlib approach

	// Pattern: "**/suffix" - matches anything ending with suffix
	if strings.HasPrefix(pattern, "**/") {
		suffix := strings.TrimPrefix(pattern, "**/")
		// Try matching the suffix directly
		if matched, _ := filepath.Match(suffix, filepath.Base(path)); matched {
			return true
		}
		// Try matching the full suffix as a path segment
		if strings.HasSuffix(path, "/"+suffix) || path == suffix {
			return true
		}
		return false
	}

	// Pattern: "prefix/**" - matches anything starting with prefix
	if strings.HasSuffix(pattern, "/**") {
		prefix := strings.TrimSuffix(pattern, "/**")
		// Check if path starts with prefix or contains it as a path segment
		return strings.HasPrefix(path, prefix+"/") || strings.Contains(path, "/"+prefix+"/") || path == prefix
	}

	// Pattern: "prefix/**/suffix" - more complex, for now just check contains
	parts := strings.Split(pattern, "**")
	if len(parts) == 2 {
		prefix := strings.TrimSuffix(parts[0], "/")
		suffix := strings.TrimPrefix(parts[1], "/")

		prefixMatch := prefix == "" || strings.Contains(path, prefix)
		suffixMatch := suffix == "" || strings.Contains(path, suffix)

		return prefixMatch && suffixMatch
	}

	return false
}

// validatePatterns checks for common pattern syntax issues and provides helpful error messages
func (f *FileReaderImpl) validatePatterns(patterns []string, patternType string) error {
	for _, pattern := range patterns {
		if err := f.validatePattern(pattern); err != nil {
			return fmt.Errorf("invalid %s pattern '%s': %w", patternType, pattern, err)
		}
	}
	return nil
}

// validatePattern validates a single pattern for common issues
func (f *FileReaderImpl) validatePattern(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("empty pattern not allowed")
	}

	// Check for escaped characters first (before other bracket checks)
	if strings.Contains(pattern, "\\") {
		return fmt.Errorf("escaped characters not fully supported, avoid backslashes in patterns")
	}

	// Check for multiple ** (not supported)
	if strings.Count(pattern, "**") > 1 {
		return fmt.Errorf("multiple ** globstars not supported, use single ** instead")
	}

	// Check for regex-like patterns (common confusion)
	if strings.Contains(pattern, ".*") {
		return fmt.Errorf("looks like regex syntax, use glob syntax instead (e.g., '*.py' not '.*\\.py')")
	}
	if strings.HasSuffix(pattern, "$") || strings.HasPrefix(pattern, "^") {
		return fmt.Errorf("regex anchors (^ $) not supported, use glob syntax instead")
	}

	// Check for character classes (not supported by filepath.Match)
	if strings.Contains(pattern, "[") || strings.Contains(pattern, "]") {
		return fmt.Errorf("character classes [abc] not supported, use separate patterns instead")
	}

	// Check for brace expansion (not supported)
	if strings.Contains(pattern, "{") || strings.Contains(pattern, "}") {
		return fmt.Errorf("brace expansion {a,b} not supported, use separate patterns instead")
	}

	// Validate the pattern with filepath.Match to catch syntax errors early
	_, err := filepath.Match(pattern, "test")
	if err != nil {
		return fmt.Errorf("invalid glob syntax: %w", err)
	}

	// Note: We could add more specific validations here for complex ** patterns,
	// but most patterns that pass filepath.Match will work reasonably well

	return nil
}

// shouldSkipDirectory checks if a directory should be skipped entirely
func (f *FileReaderImpl) shouldSkipDirectory(dirName string) bool {
	skipDirs := []string{
		"__pycache__",
		".git",
		".svn",
		".hg",
		".bzr",
		"node_modules",
		".tox",
		".pytest_cache",
		".mypy_cache",
		"venv",
		"env",
		".venv",
		".env",
		"build",
		"dist",
		"*.egg-info",
	}

	dirLower := strings.ToLower(dirName)
	for _, skipDir := range skipDirs {
		if matched, _ := filepath.Match(strings.ToLower(skipDir), dirLower); matched {
			return true
		}
	}

	return false
}

// GetFileInfo provides additional information about a file
func (f *FileReaderImpl) GetFileInfo(path string) (os.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, domain.NewFileNotFoundError(path, err)
	}
	return info, nil
}

// ValidatePaths validates that all provided paths exist and are accessible
func (f *FileReaderImpl) ValidatePaths(paths []string) error {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return domain.NewFileNotFoundError(path, err)
			}
			return domain.NewInvalidInputError(fmt.Sprintf("cannot access path: %s", path), err)
		}
	}
	return nil
}
