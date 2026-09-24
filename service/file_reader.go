package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
)

// FileReaderImpl implements the FileReader interface
type FileReaderImpl struct{}

// NewFileReader creates a new file reader service
func NewFileReader() *FileReaderImpl {
	return &FileReaderImpl{}
}

// CollectPythonFiles recursively finds all Python files in the given paths
func (f *FileReaderImpl) CollectPythonFiles(paths []string, recursive bool, includePatterns, excludePatterns []string) ([]string, error) {

	var files []string
	selection := domain.PythonFileSelection{IncludePatterns: includePatterns, ExcludePatterns: excludePatterns}
	projectRoot := FindProjectRoot(paths)

	for _, path := range paths {
		// Check if path exists
		info, err := os.Stat(path)
		if err != nil {
			return nil, domain.NewFileNotFoundError(path, err)
		}

		if info.IsDir() {
			// Process directory
			dirFiles, err := f.collectFromDirectory(path, recursive, projectRoot, selection)
			if err != nil {
				return nil, err
			}
			files = append(files, dirFiles...)
		} else {
			// Process single file
			if !f.IsValidPythonFile(path) {
				continue
			}
			included, err := MatchesPythonFileSelection(projectRoot, path, selection)
			if err != nil {
				return nil, err
			}
			if included {
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
func (f *FileReaderImpl) collectFromDirectory(dirPath string, recursive bool, projectRoot string, selection domain.PythonFileSelection) ([]string, error) {
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

		// Skip hidden directories and files (but not the root directory being walked)
		if strings.HasPrefix(info.Name(), ".") && path != dirPath {
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
			included, err := MatchesPythonFileSelection(projectRoot, path, selection)
			if err != nil {
				return err
			}
			if included {
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

// MatchesPythonFileSelection reports whether path passes the selection's
// include and exclude patterns. Patterns match the path relative to
// projectRoot, never the directories above it.
func MatchesPythonFileSelection(projectRoot, path string, selection domain.PythonFileSelection) (bool, error) {
	relPath, err := analyzer.ProjectRelativePath(projectRoot, path)
	if err != nil {
		return false, err
	}
	for _, pattern := range selection.ExcludePatterns {
		if analyzer.MatchPathPattern(pattern, relPath) {
			return false, nil
		}
	}
	if len(selection.IncludePatterns) == 0 {
		return true, nil
	}
	for _, pattern := range selection.IncludePatterns {
		if analyzer.MatchPathPattern(pattern, relPath) {
			return true, nil
		}
	}
	return false, nil
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
