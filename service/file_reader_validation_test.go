package service

import (
	"strings"
	"testing"
)

func TestFileReader_ValidatePattern(t *testing.T) {
	fr := NewFileReader()

	tests := []struct {
		name        string
		pattern     string
		expectError bool
		errorSubstr string
	}{
		// Valid patterns
		{
			name:        "simple wildcard",
			pattern:     "*.py",
			expectError: false,
		},
		{
			name:        "test file pattern",
			pattern:     "test_*.py",
			expectError: false,
		},
		{
			name:        "directory with globstar",
			pattern:     "venv/**",
			expectError: false,
		},
		{
			name:        "globstar with suffix",
			pattern:     "**/test.py",
			expectError: false,
		},
		{
			name:        "complex but valid path",
			pattern:     "src/*/tests/*.py",
			expectError: false,
		},

		// Invalid patterns - Multiple globstars
		{
			name:        "multiple globstars",
			pattern:     "**/dir/**/file.py",
			expectError: true,
			errorSubstr: "multiple ** globstars",
		},

		// Invalid patterns - Regex syntax
		{
			name:        "regex dot-star",
			pattern:     ".*py", // Changed to avoid backslash which gets caught first
			expectError: true,
			errorSubstr: "looks like regex syntax",
		},
		{
			name:        "regex with dollar",
			pattern:     "test.py$",
			expectError: true,
			errorSubstr: "regex anchors",
		},
		{
			name:        "regex with caret",
			pattern:     "^test.py",
			expectError: true,
			errorSubstr: "regex anchors",
		},

		// Invalid patterns - Character classes
		{
			name:        "character class",
			pattern:     "[abc]*.py",
			expectError: true,
			errorSubstr: "character classes",
		},
		{
			name:        "character range",
			pattern:     "[a-z]*.py",
			expectError: true,
			errorSubstr: "character classes",
		},
		{
			name:        "negated character class",
			pattern:     "[!test]*.py",
			expectError: true,
			errorSubstr: "character classes",
		},

		// Invalid patterns - Brace expansion
		{
			name:        "brace expansion alternatives",
			pattern:     "{test,spec}_*.py",
			expectError: true,
			errorSubstr: "brace expansion",
		},
		{
			name:        "brace expansion extensions",
			pattern:     "*.{py,pyx}",
			expectError: true,
			errorSubstr: "brace expansion",
		},

		// Invalid patterns - Escaped characters
		{
			name:        "escaped asterisk",
			pattern:     "\\*.py",
			expectError: true,
			errorSubstr: "escaped characters",
		},
		{
			name:        "escaped bracket",
			pattern:     "\\[test\\].py",
			expectError: true,
			errorSubstr: "escaped characters",
		},

		// Invalid patterns - Empty
		{
			name:        "empty pattern",
			pattern:     "",
			expectError: true,
			errorSubstr: "empty pattern",
		},

		// Problematic but technically valid patterns - Let's remove this test for now
		// The pattern "*/dir/**" actually works fine, so let's not warn about it

		// Invalid glob syntax
		{
			name:        "malformed glob",
			pattern:     "test[.py",
			expectError: true,
			errorSubstr: "character classes", // This gets caught by character class check first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fr.validatePattern(tt.pattern)

			if tt.expectError {
				if err == nil {
					t.Errorf("validatePattern(%q) should have returned an error", tt.pattern)
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("validatePattern(%q) error %q should contain %q", tt.pattern, err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("validatePattern(%q) should not have returned an error, got: %v", tt.pattern, err)
				}
			}
		})
	}
}

func TestFileReader_ValidatePatterns(t *testing.T) {
	fr := NewFileReader()

	tests := []struct {
		name        string
		patterns    []string
		patternType string
		expectError bool
		errorSubstr string
	}{
		{
			name:        "all valid patterns",
			patterns:    []string{"*.py", "test_*.py", "venv/**"},
			patternType: "exclude",
			expectError: false,
		},
		{
			name:        "mixed valid and invalid",
			patterns:    []string{"*.py", "[abc]*.py", "venv/**"},
			patternType: "include",
			expectError: true,
			errorSubstr: "invalid include pattern '[abc]*.py'",
		},
		{
			name:        "multiple invalid patterns - reports first",
			patterns:    []string{"[abc]*.py", "{test,spec}*.py"},
			patternType: "exclude",
			expectError: true,
			errorSubstr: "invalid exclude pattern '[abc]*.py'",
		},
		{
			name:        "empty patterns list",
			patterns:    []string{},
			patternType: "exclude",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fr.validatePatterns(tt.patterns, tt.patternType)

			if tt.expectError {
				if err == nil {
					t.Errorf("validatePatterns(%v, %q) should have returned an error", tt.patterns, tt.patternType)
					return
				}
				if tt.errorSubstr != "" && !strings.Contains(err.Error(), tt.errorSubstr) {
					t.Errorf("validatePatterns(%v, %q) error %q should contain %q", tt.patterns, tt.patternType, err.Error(), tt.errorSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("validatePatterns(%v, %q) should not have returned an error, got: %v", tt.patterns, tt.patternType, err)
				}
			}
		})
	}
}

func TestFileReader_CollectPythonFiles_ValidationIntegration(t *testing.T) {
	fr := NewFileReader()

	// Test that pattern validation happens during CollectPythonFiles
	_, err := fr.CollectPythonFiles(
		[]string{"."},
		true,
		[]string{"*.py"},      // valid include pattern
		[]string{"[abc]*.py"}, // invalid exclude pattern
	)

	if err == nil {
		t.Error("CollectPythonFiles should have failed due to invalid exclude pattern")
		return
	}

	if !strings.Contains(err.Error(), "invalid exclude pattern") {
		t.Errorf("Error should mention invalid exclude pattern, got: %v", err)
	}
}
