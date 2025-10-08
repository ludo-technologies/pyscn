package service

import (
	"testing"
)

func TestFileReader_GlobstarPatterns(t *testing.T) {
	fr := NewFileReader()

	tests := []struct {
		name     string
		pattern  string
		path     string
		expected bool
	}{
		// Basic globstar patterns
		{
			name:     "directory with globstar matches files in subdirs",
			pattern:  "postrp/cli/**",
			path:     "postrp/cli/main.py",
			expected: true,
		},
		{
			name:     "directory with globstar matches files in nested subdirs",
			pattern:  "postrp/cli/**",
			path:     "postrp/cli/subdir/file.py",
			expected: true,
		},
		{
			name:     "directory with globstar doesn't match outside directory",
			pattern:  "postrp/cli/**",
			path:     "other/dir/file.py",
			expected: false,
		},
		{
			name:     "globstar with suffix matches anywhere",
			pattern:  "**/test.py",
			path:     "deep/nested/test.py",
			expected: true,
		},
		{
			name:     "globstar with suffix matches at root",
			pattern:  "**/test.py",
			path:     "test.py",
			expected: true,
		},
		// Test the actual default config patterns
		{
			name:     "venv directory exclusion",
			pattern:  "venv/**",
			path:     "venv/lib/python3.9/site-packages/module.py",
			expected: true,
		},
		{
			name:     "__pycache__ directory exclusion",
			pattern:  "__pycache__/**",
			path:     "src/__pycache__/module.cpython-39.pyc",
			expected: true,
		},
		{
			name:     ".pytest_cache directory exclusion",
			pattern:  ".pytest_cache/**",
			path:     ".pytest_cache/v/cache/nodeids",
			expected: true,
		},
		{
			name:     ".tox directory exclusion",
			pattern:  ".tox/**",
			path:     ".tox/py39/lib/python3.9/site-packages/pytest.py",
			expected: true,
		},
		{
			name:     "virtual env variants",
			pattern:  ".venv/**",
			path:     ".venv/bin/python",
			expected: true,
		},
		// Regular patterns (should still work)
		{
			name:     "simple wildcard pattern",
			pattern:  "test_*.py",
			path:     "test_example.py",
			expected: true,
		},
		{
			name:     "simple wildcard pattern no match",
			pattern:  "test_*.py",
			path:     "example_test.py",
			expected: false,
		},
		{
			name:     "directory pattern without globstar",
			pattern:  "postrp/cli/*.py",
			path:     "postrp/cli/main.py",
			expected: true,
		},
		{
			name:     "directory pattern without globstar doesn't match subdirs",
			pattern:  "postrp/cli/*.py",
			path:     "postrp/cli/subdir/file.py",
			expected: false,
		},
		// Edge cases
		{
			name:     "globstar at end matches directory itself",
			pattern:  "build/**",
			path:     "build",
			expected: true,
		},
		{
			name:     "nested globstar pattern (realistic use case)",
			pattern:  "__pycache__/**",
			path:     "/home/user/project/src/__pycache__/module.pyc",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fr.matchesPattern(tt.pattern, tt.path)
			if result != tt.expected {
				t.Errorf("matchesPattern(%q, %q) = %v, expected %v", tt.pattern, tt.path, result, tt.expected)
			}
		})
	}
}

func TestFileReader_ShouldIncludeFile_ExcludePatterns(t *testing.T) {
	fr := NewFileReader()

	excludePatterns := []string{
		"test_*.py",
		"*_test.py",
		"postrp/cli/**",
		"venv/**",
	}

	tests := []struct {
		name     string
		path     string
		expected bool // true = should include, false = should exclude
	}{
		{
			name:     "normal file should be included",
			path:     "src/main.py",
			expected: true,
		},
		{
			name:     "test file should be excluded",
			path:     "test_example.py",
			expected: false,
		},
		{
			name:     "another test file should be excluded",
			path:     "example_test.py",
			expected: false,
		},
		{
			name:     "file in postrp/cli should be excluded",
			path:     "postrp/cli/main.py",
			expected: false,
		},
		{
			name:     "file in postrp/cli subdir should be excluded",
			path:     "postrp/cli/commands/run.py",
			expected: false,
		},
		{
			name:     "file in venv should be excluded",
			path:     "venv/lib/python3.9/site-packages/module.py",
			expected: false,
		},
		{
			name:     "file outside excluded paths should be included",
			path:     "postrp/core/main.py",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fr.shouldIncludeFile(tt.path, []string{"*.py"}, excludePatterns)
			if result != tt.expected {
				if tt.expected {
					t.Errorf("shouldIncludeFile(%q) = false, expected true (file should be included)", tt.path)
				} else {
					t.Errorf("shouldIncludeFile(%q) = true, expected false (file should be excluded)", tt.path)
				}
			}
		})
	}
}
