package app

import "github.com/ludo-technologies/pyscn/domain"

// ResolveFilePaths resolves file paths for analysis.
// If all paths are already files (not directories), returns them directly.
// Otherwise, collects Python files from the provided paths using the specified filters.
//
// Parameters:
//   - fileReader: The file reader abstraction for file operations
//   - paths: The input paths to resolve (can be files or directories)
//   - recursive: Whether to recursively collect files from subdirectories
//   - includePatterns: Glob patterns for files to include
//   - excludePatterns: Glob patterns for files to exclude
//   - validatePythonFile: If true, also validates paths are Python files (stricter check)
//
// Returns:
//   - []string: List of resolved Python file paths
//   - error: Any error encountered during resolution
//
// This function optimizes the case where AnalyzeUseCase pre-collects files
// and passes them to individual analysis use cases, avoiding redundant file collection.
func ResolveFilePaths(
	fileReader domain.FileReader,
	paths []string,
	recursive bool,
	includePatterns []string,
	excludePatterns []string,
	validatePythonFile bool,
) ([]string, error) {
	// Check if all paths are already files (not directories)
	// This happens when called from AnalyzeUseCase which pre-collects files
	allFiles := true
	for _, path := range paths {
		// Optional: Validate that path is a Python file (used by clone detection)
		if validatePythonFile && !fileReader.IsValidPythonFile(path) {
			allFiles = false
			break
		}

		// Check if file exists (FileExists returns true only for files, not directories)
		exists, err := fileReader.FileExists(path)
		if err != nil || !exists {
			allFiles = false
			break
		}
	}

	// If all paths are already files, no need to collect again
	if allFiles {
		return paths, nil
	}

	// Collect Python files from directories
	files, err := fileReader.CollectPythonFiles(
		paths,
		recursive,
		includePatterns,
		excludePatterns,
	)
	if err != nil {
		return nil, err
	}

	return files, nil
}
