package app

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFileReader is a mock implementation of domain.FileReader
type MockFileReader struct {
	mock.Mock
}

func (m *MockFileReader) FileExists(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

func (m *MockFileReader) IsValidPythonFile(path string) bool {
	args := m.Called(path)
	return args.Bool(0)
}

func (m *MockFileReader) CollectPythonFiles(paths []string, recursive bool, includePatterns []string, excludePatterns []string) ([]string, error) {
	args := m.Called(paths, recursive, includePatterns, excludePatterns)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockFileReader) ReadFile(path string) ([]byte, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func TestResolveFilePaths_AllPathsAreFiles(t *testing.T) {
	// Setup
	mockReader := new(MockFileReader)
	paths := []string{"file1.py", "file2.py", "file3.py"}

	// Mock: All paths exist as files
	for _, path := range paths {
		mockReader.On("FileExists", path).Return(true, nil)
	}

	// Execute
	result, err := ResolveFilePaths(
		mockReader,
		paths,
		false,
		[]string{"*.py"},
		[]string{},
		false,
	)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, paths, result, "Should return paths directly when all are files")
	mockReader.AssertExpectations(t)
	mockReader.AssertNotCalled(t, "CollectPythonFiles") // Should not call CollectPythonFiles
}

func TestResolveFilePaths_AllPathsAreFilesWithValidation(t *testing.T) {
	// Setup
	mockReader := new(MockFileReader)
	paths := []string{"file1.py", "file2.py"}

	// Mock: All paths are valid Python files and exist
	for _, path := range paths {
		mockReader.On("IsValidPythonFile", path).Return(true)
		mockReader.On("FileExists", path).Return(true, nil)
	}

	// Execute
	result, err := ResolveFilePaths(
		mockReader,
		paths,
		false,
		[]string{"*.py"},
		[]string{},
		true, // validatePythonFile enabled
	)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, paths, result, "Should return paths directly when all are valid Python files")
	mockReader.AssertExpectations(t)
	mockReader.AssertNotCalled(t, "CollectPythonFiles")
}

func TestResolveFilePaths_InvalidPythonFileWithValidation(t *testing.T) {
	// Setup
	mockReader := new(MockFileReader)
	paths := []string{"file1.py", "file2.txt"} // file2.txt is not a Python file

	// Mock: First file is valid Python and exists, second is not valid Python
	mockReader.On("IsValidPythonFile", "file1.py").Return(true)
	mockReader.On("FileExists", "file1.py").Return(true, nil) // After IsValidPythonFile check, FileExists is called
	mockReader.On("IsValidPythonFile", "file2.txt").Return(false)

	// Mock: Should fall back to CollectPythonFiles
	collectedFiles := []string{"file1.py"}
	mockReader.On("CollectPythonFiles", paths, false, []string{"*.py"}, []string{}).Return(collectedFiles, nil)

	// Execute
	result, err := ResolveFilePaths(
		mockReader,
		paths,
		false,
		[]string{"*.py"},
		[]string{},
		true, // validatePythonFile enabled
	)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, collectedFiles, result, "Should collect files when validation fails")
	mockReader.AssertExpectations(t)
}

func TestResolveFilePaths_MixedFilesAndDirectories(t *testing.T) {
	// Setup
	mockReader := new(MockFileReader)
	paths := []string{"file1.py", "directory"}

	// Mock: First path is a file, second doesn't exist as a file (is a directory)
	mockReader.On("FileExists", "file1.py").Return(true, nil)
	mockReader.On("FileExists", "directory").Return(false, nil)

	// Mock: Should call CollectPythonFiles
	collectedFiles := []string{"file1.py", "directory/file2.py", "directory/file3.py"}
	mockReader.On("CollectPythonFiles", paths, true, []string{"*.py"}, []string{"*_test.py"}).Return(collectedFiles, nil)

	// Execute
	result, err := ResolveFilePaths(
		mockReader,
		paths,
		true,
		[]string{"*.py"},
		[]string{"*_test.py"},
		false,
	)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, collectedFiles, result, "Should collect files when paths include directories")
	mockReader.AssertExpectations(t)
}

func TestResolveFilePaths_FileExistsError(t *testing.T) {
	// Setup
	mockReader := new(MockFileReader)
	paths := []string{"file1.py", "file2.py"}

	// Mock: First file exists, second returns an error
	mockReader.On("FileExists", "file1.py").Return(true, nil)
	mockReader.On("FileExists", "file2.py").Return(false, errors.New("permission denied"))

	// Mock: Should fall back to CollectPythonFiles
	collectedFiles := []string{"file1.py"}
	mockReader.On("CollectPythonFiles", paths, false, []string{"*.py"}, []string{}).Return(collectedFiles, nil)

	// Execute
	result, err := ResolveFilePaths(
		mockReader,
		paths,
		false,
		[]string{"*.py"},
		[]string{},
		false,
	)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, collectedFiles, result, "Should collect files when FileExists returns error")
	mockReader.AssertExpectations(t)
}

func TestResolveFilePaths_CollectFilesError(t *testing.T) {
	// Setup
	mockReader := new(MockFileReader)
	paths := []string{"directory"}

	// Mock: Path doesn't exist as a file
	mockReader.On("FileExists", "directory").Return(false, nil)

	// Mock: CollectPythonFiles returns an error
	collectError := errors.New("failed to collect files")
	mockReader.On("CollectPythonFiles", paths, true, []string{"*.py"}, []string{}).Return([]string(nil), collectError)

	// Execute
	result, err := ResolveFilePaths(
		mockReader,
		paths,
		true,
		[]string{"*.py"},
		[]string{},
		false,
	)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, collectError, err, "Should return the CollectPythonFiles error")
	assert.Nil(t, result)
	mockReader.AssertExpectations(t)
}

func TestResolveFilePaths_EmptyPaths(t *testing.T) {
	// Setup
	mockReader := new(MockFileReader)
	paths := []string{}

	// Execute
	result, err := ResolveFilePaths(
		mockReader,
		paths,
		false,
		[]string{"*.py"},
		[]string{},
		false,
	)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, []string{}, result, "Should return empty slice for empty paths")
}

func TestResolveFilePaths_RecursiveWithPatterns(t *testing.T) {
	// Setup
	mockReader := new(MockFileReader)
	paths := []string{"src"}

	// Mock: Path is not a file (is a directory)
	mockReader.On("FileExists", "src").Return(false, nil)

	// Mock: Should call CollectPythonFiles with correct parameters
	includePatterns := []string{"**/*.py", "!test_*.py"}
	excludePatterns := []string{"**/migrations/*.py"}
	collectedFiles := []string{"src/main.py", "src/utils/helper.py"}
	mockReader.On("CollectPythonFiles", paths, true, includePatterns, excludePatterns).Return(collectedFiles, nil)

	// Execute
	result, err := ResolveFilePaths(
		mockReader,
		paths,
		true,
		includePatterns,
		excludePatterns,
		false,
	)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, collectedFiles, result)
	mockReader.AssertExpectations(t)
	mockReader.AssertCalled(t, "CollectPythonFiles", paths, true, includePatterns, excludePatterns)
}

func TestResolveFilePaths_NoFilesCollected(t *testing.T) {
	// Setup
	mockReader := new(MockFileReader)
	paths := []string{"empty_directory"}

	// Mock: Path is not a file
	mockReader.On("FileExists", "empty_directory").Return(false, nil)

	// Mock: CollectPythonFiles returns empty slice
	mockReader.On("CollectPythonFiles", paths, false, []string{"*.py"}, []string{}).Return([]string{}, nil)

	// Execute
	result, err := ResolveFilePaths(
		mockReader,
		paths,
		false,
		[]string{"*.py"},
		[]string{},
		false,
	)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, result, "Should return empty slice when no files are collected")
	mockReader.AssertExpectations(t)
}
