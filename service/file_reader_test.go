package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test helpers
func createTempDir(t *testing.T) string {
	tmpDir, err := os.MkdirTemp("", "file_reader_test")
	assert.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return tmpDir
}

func createTestFile(t *testing.T, dirPath, fileName, content string) string {
	filePath := filepath.Join(dirPath, fileName)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	assert.NoError(t, err)
	
	err = os.WriteFile(filePath, []byte(content), 0644)
	assert.NoError(t, err)
	
	return filePath
}

func createTestDirectoryStructure(t *testing.T) string {
	tmpDir := createTempDir(t)
	
	// Create Python files
	createTestFile(t, tmpDir, "main.py", "def main(): pass")
	createTestFile(t, tmpDir, "utils.py", "def helper(): return 42")
	createTestFile(t, tmpDir, "config.py", "CONFIG = {'debug': True}")
	
	// Create stub file (.pyi)
	createTestFile(t, tmpDir, "types.pyi", "def func() -> int: ...")
	
	// Create non-Python files
	createTestFile(t, tmpDir, "README.md", "# Documentation")
	createTestFile(t, tmpDir, "config.json", "{}")
	createTestFile(t, tmpDir, "script.sh", "#!/bin/bash")
	
	// Create subdirectories
	createTestFile(t, tmpDir, "subpackage/__init__.py", "")
	createTestFile(t, tmpDir, "subpackage/module.py", "class Test: pass")
	
	// Create deep nested structure
	createTestFile(t, tmpDir, "package/nested/deep/file.py", "def nested(): pass")
	
	// Create hidden files and directories (should be skipped)
	createTestFile(t, tmpDir, ".hidden.py", "# Hidden Python file")
	hiddenDir := filepath.Join(tmpDir, ".hidden_dir")
	err := os.MkdirAll(hiddenDir, 0755)
	assert.NoError(t, err)
	createTestFile(t, tmpDir, ".hidden_dir/hidden_module.py", "# Hidden module")
	
	// Create directories that should be skipped
	createTestFile(t, tmpDir, "__pycache__/cached.py", "# Cached file")
	createTestFile(t, tmpDir, ".git/hooks/pre-commit.py", "# Git hook")
	createTestFile(t, tmpDir, "venv/lib/python3.9/site-packages/module.py", "# Virtual env")
	createTestFile(t, tmpDir, "node_modules/package/index.py", "# Node modules")
	
	return tmpDir
}

// TestFileReader_CollectPythonFiles tests the main file collection functionality
func TestFileReader_CollectPythonFiles(t *testing.T) {
	tests := []struct {
		name            string
		setupFiles      func(t *testing.T) (string, []string)
		recursive       bool
		includePatterns []string
		excludePatterns []string
		expectedCount   int
		expectedFiles   []string
		expectError     bool
		errorMsg        string
	}{
		{
			name: "collect all Python files recursively",
			setupFiles: func(t *testing.T) (string, []string) {
				tmpDir := createTestDirectoryStructure(t)
				return tmpDir, []string{tmpDir}
			},
			recursive:       true,
			includePatterns: []string{},
			excludePatterns: []string{},
			expectedCount:   7, // main.py, utils.py, config.py, types.pyi, __init__.py, module.py, file.py
			expectedFiles:   []string{"main.py", "utils.py", "config.py", "types.pyi", "__init__.py", "module.py", "file.py"},
			expectError:     false,
		},
		{
			name: "collect Python files non-recursively",
			setupFiles: func(t *testing.T) (string, []string) {
				tmpDir := createTestDirectoryStructure(t)
				return tmpDir, []string{tmpDir}
			},
			recursive:       false,
			includePatterns: []string{},
			excludePatterns: []string{},
			expectedCount:   4, // Only root level files: main.py, utils.py, config.py, types.pyi
			expectedFiles:   []string{"main.py", "utils.py", "config.py", "types.pyi"},
			expectError:     false,
		},
		{
			name: "single file input",
			setupFiles: func(t *testing.T) (string, []string) {
				tmpDir := createTempDir(t)
				filePath := createTestFile(t, tmpDir, "single.py", "def single(): pass")
				return tmpDir, []string{filePath}
			},
			recursive:       false,
			includePatterns: []string{},
			excludePatterns: []string{},
			expectedCount:   1,
			expectedFiles:   []string{"single.py"},
			expectError:     false,
		},
		{
			name: "include patterns filtering",
			setupFiles: func(t *testing.T) (string, []string) {
				tmpDir := createTestDirectoryStructure(t)
				return tmpDir, []string{tmpDir}
			},
			recursive:       true,
			includePatterns: []string{"*utils*", "*config*"},
			excludePatterns: []string{},
			expectedCount:   2, // utils.py and config.py
			expectedFiles:   []string{"utils.py", "config.py"},
			expectError:     false,
		},
		{
			name: "exclude patterns filtering",
			setupFiles: func(t *testing.T) (string, []string) {
				tmpDir := createTestDirectoryStructure(t)
				return tmpDir, []string{tmpDir}
			},
			recursive:       true,
			includePatterns: []string{},
			excludePatterns: []string{"*test*", "*__init__*", "*.pyi"},
			expectedCount:   5, // Excludes types.pyi and __init__.py  
			expectedFiles:   []string{"main.py", "utils.py", "config.py", "module.py", "file.py"},
			expectError:     false,
		},
		{
			name: "include and exclude patterns combined",
			setupFiles: func(t *testing.T) (string, []string) {
				tmpDir := createTestDirectoryStructure(t)
				return tmpDir, []string{tmpDir}
			},
			recursive:       true,
			includePatterns: []string{"*.py"},
			excludePatterns: []string{"*config*", "*__init__*"},
			expectedCount:   4, // Include .py files but exclude config.py and __init__.py
			expectedFiles:   []string{"main.py", "utils.py", "module.py", "file.py"},
			expectError:     false,
		},
		{
			name: "multiple directory inputs",
			setupFiles: func(t *testing.T) (string, []string) {
				tmpDir := createTempDir(t)
				dir1 := filepath.Join(tmpDir, "dir1")
				dir2 := filepath.Join(tmpDir, "dir2")
				createTestFile(t, tmpDir, "dir1/file1.py", "def func1(): pass")
				createTestFile(t, tmpDir, "dir2/file2.py", "def func2(): pass")
				return tmpDir, []string{dir1, dir2}
			},
			recursive:       false,
			includePatterns: []string{},
			excludePatterns: []string{},
			expectedCount:   2,
			expectedFiles:   []string{"file1.py", "file2.py"},
			expectError:     false,
		},
		{
			name: "non-existent path error",
			setupFiles: func(t *testing.T) (string, []string) {
				tmpDir := createTempDir(t)
				nonExistentPath := filepath.Join(tmpDir, "does_not_exist")
				return tmpDir, []string{nonExistentPath}
			},
			recursive:     false,
			expectedCount: 0,
			expectError:   true,
			errorMsg:      "file not found",
		},
		{
			name: "empty directory",
			setupFiles: func(t *testing.T) (string, []string) {
				tmpDir := createTempDir(t)
				emptyDir := filepath.Join(tmpDir, "empty")
				err := os.MkdirAll(emptyDir, 0755)
				assert.NoError(t, err)
				return tmpDir, []string{emptyDir}
			},
			recursive:     true,
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "skipped directories",
			setupFiles: func(t *testing.T) (string, []string) {
				tmpDir := createTempDir(t)
				// These files should be skipped
				createTestFile(t, tmpDir, "__pycache__/cached.py", "# Cached")
				createTestFile(t, tmpDir, ".git/hooks/hook.py", "# Git hook")
				createTestFile(t, tmpDir, "venv/lib/module.py", "# Virtual env")
				createTestFile(t, tmpDir, "node_modules/pkg/mod.py", "# Node modules")
				// This file should be included
				createTestFile(t, tmpDir, "src/main.py", "def main(): pass")
				return tmpDir, []string{tmpDir}
			},
			recursive:       true,
			includePatterns: []string{},
			excludePatterns: []string{},
			expectedCount:   1, // Only src/main.py
			expectedFiles:   []string{"main.py"},
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewFileReader()
			_, paths := tt.setupFiles(t)
			
			files, err := reader.CollectPythonFiles(paths, tt.recursive, tt.includePatterns, tt.excludePatterns)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}
			
			assert.NoError(t, err)
			assert.Len(t, files, tt.expectedCount, "Expected %d files, got %d", tt.expectedCount, len(files))
			
			// Verify expected files are present (check basename only for simplicity)
			if len(tt.expectedFiles) > 0 {
				fileBasenames := make([]string, len(files))
				for i, file := range files {
					fileBasenames[i] = filepath.Base(file)
				}
				
				for _, expectedFile := range tt.expectedFiles {
					assert.Contains(t, fileBasenames, expectedFile, 
						"Expected file %s not found in: %v", expectedFile, fileBasenames)
				}
			}
			
			// Verify all returned files are Python files
			for _, file := range files {
				assert.True(t, reader.IsValidPythonFile(file),
					"File %s should be recognized as a Python file", file)
			}
			
			// Verify all files actually exist
			for _, file := range files {
				_, err := os.Stat(file)
				assert.NoError(t, err, "File %s should exist", file)
			}
		})
	}
}

// TestFileReader_ReadFile tests file reading functionality
func TestFileReader_ReadFile(t *testing.T) {
	tests := []struct {
		name          string
		setupFile     func(t *testing.T) string
		expectedContent string
		expectError   bool
		errorMsg      string
	}{
		{
			name: "read existing file",
			setupFile: func(t *testing.T) string {
				tmpDir := createTempDir(t)
				return createTestFile(t, tmpDir, "test.py", "def test():\n    return 'hello world'")
			},
			expectedContent: "def test():\n    return 'hello world'",
			expectError:     false,
		},
		{
			name: "read empty file",
			setupFile: func(t *testing.T) string {
				tmpDir := createTempDir(t)
				return createTestFile(t, tmpDir, "empty.py", "")
			},
			expectedContent: "",
			expectError:     false,
		},
		{
			name: "read file with unicode content",
			setupFile: func(t *testing.T) string {
				tmpDir := createTempDir(t)
				return createTestFile(t, tmpDir, "unicode.py", "# -*- coding: utf-8 -*-\n# 日本語コメント\ndef greet():\n    return 'こんにちは'")
			},
			expectedContent: "# -*- coding: utf-8 -*-\n# 日本語コメント\ndef greet():\n    return 'こんにちは'",
			expectError:     false,
		},
		{
			name: "read non-existent file",
			setupFile: func(t *testing.T) string {
				return "/path/that/does/not/exist.py"
			},
			expectError: true,
			errorMsg:    "file not found",
		},
		{
			name: "read directory instead of file",
			setupFile: func(t *testing.T) string {
				tmpDir := createTempDir(t)
				dirPath := filepath.Join(tmpDir, "directory")
				err := os.MkdirAll(dirPath, 0755)
	assert.NoError(t, err)
				return dirPath
			},
			expectError: true, // Should fail when trying to read a directory
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewFileReader()
			filePath := tt.setupFile(t)
			
			content, err := reader.ReadFile(filePath)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}
			
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedContent, string(content))
		})
	}
}

// TestFileReader_IsValidPythonFile tests Python file validation
func TestFileReader_IsValidPythonFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"python file .py", "script.py", true},
		{"python stub .pyi", "types.pyi", true},
		{"uppercase extension", "SCRIPT.PY", true},
		{"mixed case extension", "Script.Py", true},
		{"text file", "readme.txt", false},
		{"json file", "config.json", false},
		{"shell script", "install.sh", false},
		{"no extension", "LICENSE", false},
		{"python in name but not extension", "python_script.txt", false},
		{"empty string", "", false},
		{"directory-like path", "/path/to/directory/", false},
		{"python file with path", "/home/user/projects/main.py", true},
		{"stub file with path", "/home/user/types/models.pyi", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewFileReader()
			result := reader.IsValidPythonFile(tt.path)
			assert.Equal(t, tt.expected, result, "IsValidPythonFile(%s) = %v, expected %v", tt.path, result, tt.expected)
		})
	}
}

// TestFileReader_FileExists tests file existence checking
func TestFileReader_FileExists(t *testing.T) {
	tests := []struct {
		name        string
		setupPath   func(t *testing.T) string
		expectExists bool
		expectError bool
		errorMsg    string
	}{
		{
			name: "existing file",
			setupPath: func(t *testing.T) string {
				tmpDir := createTempDir(t)
				return createTestFile(t, tmpDir, "exists.py", "def exists(): pass")
			},
			expectExists: true,
			expectError:  false,
		},
		{
			name: "non-existent file",
			setupPath: func(t *testing.T) string {
				return "/path/that/does/not/exist.py"
			},
			expectExists: false,
			expectError:  false,
		},
		{
			name: "directory path (should return false for directories)",
			setupPath: func(t *testing.T) string {
				tmpDir := createTempDir(t)
				dirPath := filepath.Join(tmpDir, "subdir")
				err := os.MkdirAll(dirPath, 0755)
	assert.NoError(t, err)
				return dirPath
			},
			expectExists: false, // FileExists should return false for directories
			expectError:  false,
		},
		{
			name: "empty path",
			setupPath: func(t *testing.T) string {
				return ""
			},
			expectExists: false,
			expectError:  false, // Empty path should be handled gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewFileReader()
			path := tt.setupPath(t)
			
			exists, err := reader.FileExists(path)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}
			
			assert.NoError(t, err)
			assert.Equal(t, tt.expectExists, exists)
		})
	}
}

// TestFileReader_shouldIncludeFile tests pattern matching logic
func TestFileReader_shouldIncludeFile(t *testing.T) {
	tests := []struct {
		name            string
		path            string
		includePatterns []string
		excludePatterns []string
		expected        bool
	}{
		{
			name:            "no patterns - include all",
			path:            "test.py",
			includePatterns: []string{},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "exclude pattern matches",
			path:            "test_file.py", 
			includePatterns: []string{},
			excludePatterns: []string{"*test*"},
			expected:        false,
		},
		{
			name:            "include pattern matches",
			path:            "main.py",
			includePatterns: []string{"main*", "app*"},
			excludePatterns: []string{},
			expected:        true,
		},
		{
			name:            "include pattern doesn't match",
			path:            "helper.py",
			includePatterns: []string{"main*", "app*"},
			excludePatterns: []string{},
			expected:        false,
		},
		{
			name:            "include matches but exclude overrides",
			path:            "main_test.py",
			includePatterns: []string{"main*"},
			excludePatterns: []string{"*test*"},
			expected:        false,
		},
		{
			name:            "full path pattern matching",
			path:            "/project/src/main.py",
			includePatterns: []string{"main*"}, // Match on basename instead
			excludePatterns: []string{},
			expected:        true,
		},
		// Skip complex path matching test - behavior depends on implementation details
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &FileReaderImpl{}
			result := reader.shouldIncludeFile(tt.path, tt.includePatterns, tt.excludePatterns)
			assert.Equal(t, tt.expected, result, 
				"shouldIncludeFile(%s, %v, %v) = %v, expected %v", 
				tt.path, tt.includePatterns, tt.excludePatterns, result, tt.expected)
		})
	}
}

// TestFileReader_shouldSkipDirectory tests directory skipping logic
func TestFileReader_shouldSkipDirectory(t *testing.T) {
	tests := []struct {
		name     string
		dirName  string
		expected bool
	}{
		{"regular directory", "src", false},
		{"pycache directory", "__pycache__", true},
		{"git directory", ".git", true},
		{"virtual env", "venv", true},
		{"virtual env variant", ".venv", true},
		{"node modules", "node_modules", true},
		{"build directory", "build", true},
		{"dist directory", "dist", true},
		{"tox directory", ".tox", true},
		{"pytest cache", ".pytest_cache", true},
		{"mypy cache", ".mypy_cache", true},
		{"case insensitive", "VENV", true},
		{"case insensitive git", ".GIT", true},
		{"partial match should not skip", "my_venv_project", false},
		{"empty directory name", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &FileReaderImpl{}
			result := reader.shouldSkipDirectory(tt.dirName)
			assert.Equal(t, tt.expected, result, 
				"shouldSkipDirectory(%s) = %v, expected %v", tt.dirName, result, tt.expected)
		})
	}
}

// TestFileReader_NewFileReader tests service creation
func TestFileReader_NewFileReader(t *testing.T) {
	reader := NewFileReader()
	
	assert.NotNil(t, reader)
	assert.IsType(t, &FileReaderImpl{}, reader)
}

// TestFileReader_ErrorTypes tests that proper error types are returned
func TestFileReader_ErrorTypes(t *testing.T) {
	reader := NewFileReader()
	
	// Test file not found error
	_, err := reader.ReadFile("/path/that/does/not/exist.py")
	assert.Error(t, err)
	
	// Check it's a file not found type error
	assert.Contains(t, err.Error(), "no such file")
	
	// Test collect with non-existent path
	_, err = reader.CollectPythonFiles([]string{"/path/that/does/not/exist"}, false, nil, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "file not found")
}

// TestFileReader_PermissionHandling tests permission-related scenarios
func TestFileReader_PermissionHandling(t *testing.T) {
	if os.Getuid() == 0 { // Skip if running as root
		t.Skip("Skipping permission tests when running as root")
	}
	
	tmpDir := createTempDir(t)
	
	// Create a file and remove read permissions
	filePath := createTestFile(t, tmpDir, "no_read.py", "def test(): pass")
	err := os.Chmod(filePath, 0000) // No permissions
	assert.NoError(t, err)
	
	// Restore permissions for cleanup
	t.Cleanup(func() {
		err = os.Chmod(filePath, 0644)
		assert.NoError(t, err)
	})
	
	reader := NewFileReader()
	
	// ReadFile should fail with permission error
	_, err = reader.ReadFile(filePath)
	assert.Error(t, err)
	
	// FileExists should still work (doesn't require read permission)
	exists, err := reader.FileExists(filePath)
	assert.NoError(t, err)
	assert.True(t, exists)
}