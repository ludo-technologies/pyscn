package service

import (
	"errors"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewErrorCategorizer tests the constructor
func TestNewErrorCategorizer(t *testing.T) {
	categorizer := NewErrorCategorizer()
	assert.NotNil(t, categorizer)
	assert.IsType(t, &ErrorCategorizerImpl{}, categorizer)
}

// TestCategorize_InputErrors tests categorization of input errors
func TestCategorize_InputErrors(t *testing.T) {
	categorizer := NewErrorCategorizer()

	tests := []struct {
		name         string
		errMsg       string
		wantCategory domain.ErrorCategory
		wantMessage  string
	}{
		{
			name:         "invalid input",
			errMsg:       "invalid input provided",
			wantCategory: domain.ErrorCategoryInput,
			wantMessage:  "Failed to process input files or directories",
		},
		{
			name:         "no files found",
			errMsg:       "no files found in directory",
			wantCategory: domain.ErrorCategoryInput,
			wantMessage:  "Failed to process input files or directories",
		},
		{
			name:         "path error",
			errMsg:       "path does not exist",
			wantCategory: domain.ErrorCategoryInput,
			wantMessage:  "Failed to process input files or directories",
		},
		{
			name:         "directory error",
			errMsg:       "directory is not accessible",
			wantCategory: domain.ErrorCategoryInput,
			wantMessage:  "Failed to process input files or directories",
		},
		{
			name:         "file not found",
			errMsg:       "file not found: /some/path.py",
			wantCategory: domain.ErrorCategoryInput,
			wantMessage:  "Failed to process input files or directories",
		},
		{
			name:         "cannot access",
			errMsg:       "cannot access the specified file",
			wantCategory: domain.ErrorCategoryInput,
			wantMessage:  "Failed to process input files or directories",
		},
		{
			name:         "permission denied",
			errMsg:       "permission denied when reading file",
			wantCategory: domain.ErrorCategoryInput,
			wantMessage:  "Failed to process input files or directories",
		},
		{
			name:         "case insensitive - uppercase",
			errMsg:       "PERMISSION DENIED",
			wantCategory: domain.ErrorCategoryInput,
			wantMessage:  "Failed to process input files or directories",
		},
		{
			name:         "case insensitive - mixed case",
			errMsg:       "No Files Found",
			wantCategory: domain.ErrorCategoryInput,
			wantMessage:  "Failed to process input files or directories",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := categorizer.Categorize(err)

			require.NotNil(t, result)
			assert.Equal(t, tt.wantCategory, result.Category)
			assert.Equal(t, tt.wantMessage, result.Message)
			assert.Equal(t, err, result.Original)
		})
	}
}

// TestCategorize_ConfigErrors tests categorization of configuration errors
func TestCategorize_ConfigErrors(t *testing.T) {
	categorizer := NewErrorCategorizer()

	tests := []struct {
		name         string
		errMsg       string
		wantCategory domain.ErrorCategory
		wantMessage  string
	}{
		{
			name:         "config error",
			errMsg:       "config file error",
			wantCategory: domain.ErrorCategoryConfig,
			wantMessage:  "Configuration file or settings error",
		},
		{
			name:         "configuration error",
			errMsg:       "configuration is invalid",
			wantCategory: domain.ErrorCategoryConfig,
			wantMessage:  "Configuration file or settings error",
		},
		{
			name:         "invalid settings",
			errMsg:       "invalid settings detected",
			wantCategory: domain.ErrorCategoryConfig,
			wantMessage:  "Configuration file or settings error",
		},
		{
			name:         "missing configuration",
			errMsg:       "missing configuration file",
			wantCategory: domain.ErrorCategoryConfig,
			wantMessage:  "Configuration file or settings error",
		},
		{
			name:         "toml error",
			errMsg:       "toml file is invalid",
			wantCategory: domain.ErrorCategoryConfig,
			wantMessage:  "Configuration file or settings error",
		},
		{
			name:         "yaml error",
			errMsg:       "yaml parsing failed",
			wantCategory: domain.ErrorCategoryConfig,
			wantMessage:  "Configuration file or settings error",
		},
		{
			name:         "json error",
			errMsg:       "json is malformed",
			wantCategory: domain.ErrorCategoryConfig,
			wantMessage:  "Configuration file or settings error",
		},
		{
			name:         "case insensitive - uppercase",
			errMsg:       "CONFIG ERROR",
			wantCategory: domain.ErrorCategoryConfig,
			wantMessage:  "Configuration file or settings error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := categorizer.Categorize(err)

			require.NotNil(t, result)
			assert.Equal(t, tt.wantCategory, result.Category)
			assert.Equal(t, tt.wantMessage, result.Message)
			assert.Equal(t, err, result.Original)
		})
	}
}

// TestCategorize_TimeoutErrors tests categorization of timeout errors
func TestCategorize_TimeoutErrors(t *testing.T) {
	categorizer := NewErrorCategorizer()

	tests := []struct {
		name         string
		errMsg       string
		wantCategory domain.ErrorCategory
		wantMessage  string
	}{
		{
			name:         "timeout",
			errMsg:       "timeout waiting for response",
			wantCategory: domain.ErrorCategoryTimeout,
			wantMessage:  "Analysis timed out",
		},
		{
			name:         "deadline",
			errMsg:       "deadline exceeded",
			wantCategory: domain.ErrorCategoryTimeout,
			wantMessage:  "Analysis timed out",
		},
		{
			name:         "context canceled",
			errMsg:       "context canceled",
			wantCategory: domain.ErrorCategoryTimeout,
			wantMessage:  "Analysis timed out",
		},
		{
			name:         "operation timed out",
			errMsg:       "operation timed out",
			wantCategory: domain.ErrorCategoryTimeout,
			wantMessage:  "Analysis timed out",
		},
		{
			name:         "exceeded",
			errMsg:       "time limit exceeded",
			wantCategory: domain.ErrorCategoryTimeout,
			wantMessage:  "Analysis timed out",
		},
		{
			name:         "case insensitive - uppercase",
			errMsg:       "TIMEOUT ERROR",
			wantCategory: domain.ErrorCategoryTimeout,
			wantMessage:  "Analysis timed out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := categorizer.Categorize(err)

			require.NotNil(t, result)
			assert.Equal(t, tt.wantCategory, result.Category)
			assert.Equal(t, tt.wantMessage, result.Message)
			assert.Equal(t, err, result.Original)
		})
	}
}

// TestCategorize_OutputErrors tests categorization of output errors
func TestCategorize_OutputErrors(t *testing.T) {
	categorizer := NewErrorCategorizer()

	tests := []struct {
		name         string
		errMsg       string
		wantCategory domain.ErrorCategory
		wantMessage  string
	}{
		{
			name:         "write error",
			errMsg:       "write failed",
			wantCategory: domain.ErrorCategoryOutput,
			wantMessage:  "Failed to generate or write output",
		},
		{
			name:         "output error",
			errMsg:       "output generation failed",
			wantCategory: domain.ErrorCategoryOutput,
			wantMessage:  "Failed to generate or write output",
		},
		{
			name:         "format error",
			errMsg:       "format is not supported",
			wantCategory: domain.ErrorCategoryOutput,
			wantMessage:  "Failed to generate or write output",
		},
		{
			name:         "cannot create",
			errMsg:       "cannot create output file",
			wantCategory: domain.ErrorCategoryOutput,
			wantMessage:  "Failed to generate or write output",
		},
		{
			name:         "failed to generate",
			errMsg:       "failed to generate report",
			wantCategory: domain.ErrorCategoryOutput,
			wantMessage:  "Failed to generate or write output",
		},
		{
			name:         "report generation",
			errMsg:       "report generation error occurred",
			wantCategory: domain.ErrorCategoryOutput,
			wantMessage:  "Failed to generate or write output",
		},
		{
			name:         "case insensitive - uppercase",
			errMsg:       "WRITE ERROR",
			wantCategory: domain.ErrorCategoryOutput,
			wantMessage:  "Failed to generate or write output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := categorizer.Categorize(err)

			require.NotNil(t, result)
			assert.Equal(t, tt.wantCategory, result.Category)
			assert.Equal(t, tt.wantMessage, result.Message)
			assert.Equal(t, err, result.Original)
		})
	}
}

// TestCategorize_ProcessingErrors tests categorization of processing errors
func TestCategorize_ProcessingErrors(t *testing.T) {
	categorizer := NewErrorCategorizer()

	tests := []struct {
		name         string
		errMsg       string
		wantCategory domain.ErrorCategory
		wantMessage  string
	}{
		{
			name:         "parse error",
			errMsg:       "parse error in file",
			wantCategory: domain.ErrorCategoryProcessing,
			wantMessage:  "Error during code analysis processing",
		},
		{
			name:         "syntax error",
			errMsg:       "syntax error detected",
			wantCategory: domain.ErrorCategoryProcessing,
			wantMessage:  "Error during code analysis processing",
		},
		{
			name:         "analysis error",
			errMsg:       "analysis failed",
			wantCategory: domain.ErrorCategoryProcessing,
			wantMessage:  "Error during code analysis processing",
		},
		{
			name:         "process error",
			errMsg:       "process terminated unexpectedly",
			wantCategory: domain.ErrorCategoryProcessing,
			wantMessage:  "Error during code analysis processing",
		},
		{
			name:         "failed to analyze",
			errMsg:       "failed to analyze code",
			wantCategory: domain.ErrorCategoryProcessing,
			wantMessage:  "Error during code analysis processing",
		},
		{
			name:         "invalid python",
			errMsg:       "invalid python syntax",
			wantCategory: domain.ErrorCategoryProcessing,
			wantMessage:  "Error during code analysis processing",
		},
		{
			name:         "compilation error",
			errMsg:       "compilation failed",
			wantCategory: domain.ErrorCategoryProcessing,
			wantMessage:  "Error during code analysis processing",
		},
		{
			name:         "ast error",
			errMsg:       "ast parsing failed",
			wantCategory: domain.ErrorCategoryProcessing,
			wantMessage:  "Error during code analysis processing",
		},
		{
			name:         "case insensitive - uppercase",
			errMsg:       "PARSE ERROR",
			wantCategory: domain.ErrorCategoryProcessing,
			wantMessage:  "Error during code analysis processing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := categorizer.Categorize(err)

			require.NotNil(t, result)
			assert.Equal(t, tt.wantCategory, result.Category)
			assert.Equal(t, tt.wantMessage, result.Message)
			assert.Equal(t, err, result.Original)
		})
	}
}

// TestCategorize_UnknownErrors tests categorization of unknown errors
func TestCategorize_UnknownErrors(t *testing.T) {
	categorizer := NewErrorCategorizer()

	tests := []struct {
		name         string
		errMsg       string
		wantCategory domain.ErrorCategory
		wantMessage  string
	}{
		{
			name:         "random error",
			errMsg:       "something went wrong",
			wantCategory: domain.ErrorCategoryUnknown,
			wantMessage:  "something went wrong",
		},
		{
			name:         "unexpected error",
			errMsg:       "unexpected error occurred",
			wantCategory: domain.ErrorCategoryUnknown,
			wantMessage:  "unexpected error occurred",
		},
		{
			name:         "generic error",
			errMsg:       "error",
			wantCategory: domain.ErrorCategoryUnknown,
			wantMessage:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			result := categorizer.Categorize(err)

			require.NotNil(t, result)
			assert.Equal(t, tt.wantCategory, result.Category)
			assert.Equal(t, tt.wantMessage, result.Message)
			assert.Equal(t, err, result.Original)
		})
	}
}

// TestCategorize_NilError tests handling of nil errors
func TestCategorize_NilError(t *testing.T) {
	categorizer := NewErrorCategorizer()
	result := categorizer.Categorize(nil)
	assert.Nil(t, result)
}

// TestCategorize_MultiplePatternMatches tests that first match wins
func TestCategorize_MultiplePatternMatches(t *testing.T) {
	categorizer := NewErrorCategorizer()

	/*
	   Error message that could match multiple categories "parse" is in processing, but if "config" appears first in the message
	   and we check config patterns first, it should match config However,
	   the order depends on map iteration which is random in Go So we test with a pattern that's unique to one category
	*/

	err := errors.New("failed to parse file: timeout exceeded")
	result := categorizer.Categorize(err)

	require.NotNil(t, result)
	// Should match the first category found (parse or timeout)
	// Both are valid since the message contains both patterns
	assert.NotEqual(t, domain.ErrorCategoryUnknown, result.Category)
}

// TestGetRecoverySuggestions tests recovery suggestions for each category
func TestGetRecoverySuggestions(t *testing.T) {
	categorizer := NewErrorCategorizer()

	tests := []struct {
		name               string
		category           domain.ErrorCategory
		wantMinSuggestions int
	}{
		{
			name:               "input category",
			category:           domain.ErrorCategoryInput,
			wantMinSuggestions: 4,
		},
		{
			name:               "config category",
			category:           domain.ErrorCategoryConfig,
			wantMinSuggestions: 4,
		},
		{
			name:               "timeout category",
			category:           domain.ErrorCategoryTimeout,
			wantMinSuggestions: 4,
		},
		{
			name:               "output category",
			category:           domain.ErrorCategoryOutput,
			wantMinSuggestions: 4,
		},
		{
			name:               "processing category",
			category:           domain.ErrorCategoryProcessing,
			wantMinSuggestions: 4,
		},
		{
			name:               "unknown category",
			category:           domain.ErrorCategoryUnknown,
			wantMinSuggestions: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := categorizer.GetRecoverySuggestions(tt.category)
			assert.NotNil(t, suggestions)
			assert.GreaterOrEqual(t, len(suggestions), tt.wantMinSuggestions,
				"Expected at least %d suggestions for %s", tt.wantMinSuggestions, tt.category)

			// Verify suggestions are not empty strings
			for i, suggestion := range suggestions {
				assert.NotEmpty(t, suggestion, "Suggestion %d should not be empty", i)
			}
		})
	}
}

// TestGetRecoverySuggestions_SpecificContent tests specific suggestion content
func TestGetRecoverySuggestions_SpecificContent(t *testing.T) {
	categorizer := NewErrorCategorizer()

	t.Run("input suggestions contain relevant advice", func(t *testing.T) {
		suggestions := categorizer.GetRecoverySuggestions(domain.ErrorCategoryInput)

		// Check that at least one suggestion mentions common input-related solutions
		hasRelevant := false
		for _, s := range suggestions {
			if strings.Contains(s, "files") || strings.Contains(s, "permissions") || strings.Contains(s, "paths") {
				hasRelevant = true
				break
			}
		}
		assert.True(t, hasRelevant, "Input suggestions should contain relevant advice")
	})

	t.Run("config suggestions contain relevant advice", func(t *testing.T) {
		suggestions := categorizer.GetRecoverySuggestions(domain.ErrorCategoryConfig)

		hasRelevant := false
		for _, s := range suggestions {
			if strings.Contains(s, "config") || strings.Contains(s, "toml") || strings.Contains(s, "init") {
				hasRelevant = true
				break
			}
		}
		assert.True(t, hasRelevant, "Config suggestions should contain relevant advice")
	})

	t.Run("timeout suggestions contain relevant advice", func(t *testing.T) {
		suggestions := categorizer.GetRecoverySuggestions(domain.ErrorCategoryTimeout)

		hasRelevant := false
		for _, s := range suggestions {
			if strings.Contains(s, "timeout") || strings.Contains(s, "smaller") || strings.Contains(s, "files") {
				hasRelevant = true
				break
			}
		}
		assert.True(t, hasRelevant, "Timeout suggestions should contain relevant advice")
	})
}

// TestGetRecoverySuggestions_UnknownCategory tests fallback for unknown categories
func TestGetRecoverySuggestions_UnknownCategory(t *testing.T) {
	categorizer := NewErrorCategorizer()

	// Use a custom category that doesn't exist in the map
	suggestions := categorizer.GetRecoverySuggestions(domain.ErrorCategory("NonExistent"))

	assert.NotNil(t, suggestions)
	assert.Len(t, suggestions, 1)
	assert.Equal(t, "Check the error message for more details", suggestions[0])
}

// TestGetCategoryMessage tests category message generation
func TestGetCategoryMessage(t *testing.T) {
	categorizer := &ErrorCategorizerImpl{
		patterns: initializeErrorPatterns(),
	}

	tests := []struct {
		name        string
		category    domain.ErrorCategory
		wantMessage string
	}{
		{
			name:        "input category",
			category:    domain.ErrorCategoryInput,
			wantMessage: "Failed to process input files or directories",
		},
		{
			name:        "config category",
			category:    domain.ErrorCategoryConfig,
			wantMessage: "Configuration file or settings error",
		},
		{
			name:        "timeout category",
			category:    domain.ErrorCategoryTimeout,
			wantMessage: "Analysis timed out",
		},
		{
			name:        "output category",
			category:    domain.ErrorCategoryOutput,
			wantMessage: "Failed to generate or write output",
		},
		{
			name:        "processing category",
			category:    domain.ErrorCategoryProcessing,
			wantMessage: "Error during code analysis processing",
		},
		{
			name:        "unknown category",
			category:    domain.ErrorCategoryUnknown,
			wantMessage: "An unexpected error occurred",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := categorizer.getCategoryMessage(tt.category)
			assert.Equal(t, tt.wantMessage, message)
		})
	}
}

// TestGetCategoryMessage_UnknownCategory tests fallback for unknown category
func TestGetCategoryMessage_UnknownCategory(t *testing.T) {
	categorizer := &ErrorCategorizerImpl{
		patterns: initializeErrorPatterns(),
	}

	message := categorizer.getCategoryMessage(domain.ErrorCategory("NonExistent"))
	assert.Equal(t, "An error occurred", message)
}

// TestContainsAnyPattern tests the pattern matching helper function
func TestContainsAnyPattern(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		patterns []string
		want     bool
	}{
		{
			name:     "single pattern match",
			str:      "file not found",
			patterns: []string{"not found", "missing"},
			want:     true,
		},
		{
			name:     "multiple patterns - first match",
			str:      "invalid configuration",
			patterns: []string{"invalid", "missing", "error"},
			want:     true,
		},
		{
			name:     "multiple patterns - last match",
			str:      "an error occurred",
			patterns: []string{"invalid", "missing", "error"},
			want:     true,
		},
		{
			name:     "no match",
			str:      "everything is fine",
			patterns: []string{"error", "failed", "invalid"},
			want:     false,
		},
		{
			name:     "empty patterns",
			str:      "some error",
			patterns: []string{},
			want:     false,
		},
		{
			name:     "empty string",
			str:      "",
			patterns: []string{"error"},
			want:     false,
		},
		{
			name:     "partial match",
			str:      "configuration error",
			patterns: []string{"config"},
			want:     true,
		},
		{
			name:     "case sensitive - lowercase pattern in string",
			str:      "timeout occurred",
			patterns: []string{"timeout"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsAnyPattern(tt.str, tt.patterns)
			assert.Equal(t, tt.want, result)
		})
	}
}

// TestInitializeErrorPatterns tests pattern initialization
func TestInitializeErrorPatterns(t *testing.T) {
	patterns := initializeErrorPatterns()

	t.Run("has all categories", func(t *testing.T) {
		assert.Contains(t, patterns, domain.ErrorCategoryInput)
		assert.Contains(t, patterns, domain.ErrorCategoryConfig)
		assert.Contains(t, patterns, domain.ErrorCategoryTimeout)
		assert.Contains(t, patterns, domain.ErrorCategoryOutput)
		assert.Contains(t, patterns, domain.ErrorCategoryProcessing)
	})

	t.Run("input patterns not empty", func(t *testing.T) {
		inputPatterns := patterns[domain.ErrorCategoryInput]
		assert.NotEmpty(t, inputPatterns)
		assert.Contains(t, inputPatterns, "file not found")
		assert.Contains(t, inputPatterns, "permission denied")
	})

	t.Run("config patterns not empty", func(t *testing.T) {
		configPatterns := patterns[domain.ErrorCategoryConfig]
		assert.NotEmpty(t, configPatterns)
		assert.Contains(t, configPatterns, "config")
		assert.Contains(t, configPatterns, "toml")
	})

	t.Run("timeout patterns not empty", func(t *testing.T) {
		timeoutPatterns := patterns[domain.ErrorCategoryTimeout]
		assert.NotEmpty(t, timeoutPatterns)
		assert.Contains(t, timeoutPatterns, "timeout")
		assert.Contains(t, timeoutPatterns, "deadline")
	})

	t.Run("output patterns not empty", func(t *testing.T) {
		outputPatterns := patterns[domain.ErrorCategoryOutput]
		assert.NotEmpty(t, outputPatterns)
		assert.Contains(t, outputPatterns, "write")
		assert.Contains(t, outputPatterns, "output")
	})

	t.Run("processing patterns not empty", func(t *testing.T) {
		processingPatterns := patterns[domain.ErrorCategoryProcessing]
		assert.NotEmpty(t, processingPatterns)
		assert.Contains(t, processingPatterns, "parse")
		assert.Contains(t, processingPatterns, "syntax")
	})
}

// TestCategorizedError_Error tests the Error() method of CategorizedError
func TestCategorizedError_Error(t *testing.T) {
	t.Run("with original error", func(t *testing.T) {
		originalErr := errors.New("original error message")
		catErr := &domain.CategorizedError{
			Category: domain.ErrorCategoryInput,
			Message:  "Failed to process input",
			Original: originalErr,
		}

		assert.Equal(t, "original error message", catErr.Error())
	})

	t.Run("without original error", func(t *testing.T) {
		catErr := &domain.CategorizedError{
			Category: domain.ErrorCategoryInput,
			Message:  "Failed to process input",
			Original: nil,
		}

		assert.Equal(t, "Failed to process input", catErr.Error())
	})
}

// TestIntegration_FullErrorFlow tests the full error categorization flow
func TestIntegration_FullErrorFlow(t *testing.T) {
	categorizer := NewErrorCategorizer()

	t.Run("categorize and get suggestions", func(t *testing.T) {
		err := errors.New("file not found: test.py")

		// Categorize the error
		catErr := categorizer.Categorize(err)
		require.NotNil(t, catErr)
		assert.Equal(t, domain.ErrorCategoryInput, catErr.Category)

		// Get recovery suggestions
		suggestions := categorizer.GetRecoverySuggestions(catErr.Category)
		assert.NotEmpty(t, suggestions)
		assert.GreaterOrEqual(t, len(suggestions), 4)
	})

	t.Run("multiple errors with different categories", func(t *testing.T) {
		testCases := []struct {
			errMsg       string
			wantCategory domain.ErrorCategory
		}{
			{"file not found", domain.ErrorCategoryInput},
			{"config error", domain.ErrorCategoryConfig},
			{"timeout exceeded", domain.ErrorCategoryTimeout},
			{"write failed", domain.ErrorCategoryOutput},
			{"parse error", domain.ErrorCategoryProcessing},
			{"unknown problem", domain.ErrorCategoryUnknown},
		}

		for _, tc := range testCases {
			err := errors.New(tc.errMsg)
			catErr := categorizer.Categorize(err)

			require.NotNil(t, catErr)
			assert.Equal(t, tc.wantCategory, catErr.Category)

			suggestions := categorizer.GetRecoverySuggestions(catErr.Category)
			assert.NotEmpty(t, suggestions)
		}
	})
}
