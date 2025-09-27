package service

import (
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

// ErrorCategorizerImpl implements the ErrorCategorizer interface
type ErrorCategorizerImpl struct {
	patterns map[domain.ErrorCategory][]string
}

// NewErrorCategorizer creates a new error categorizer
func NewErrorCategorizer() domain.ErrorCategorizer {
	return &ErrorCategorizerImpl{
		patterns: initializeErrorPatterns(),
	}
}

// initializeErrorPatterns initializes error pattern mappings
func initializeErrorPatterns() map[domain.ErrorCategory][]string {
	return map[domain.ErrorCategory][]string{
		domain.ErrorCategoryInput: {
			"invalid input",
			"no files found",
			"path",
			"directory",
			"file not found",
			"cannot access",
			"permission denied",
		},
		domain.ErrorCategoryConfig: {
			"config",
			"configuration",
			"invalid format",
			"invalid settings",
			"missing configuration",
			"toml",
			"yaml",
			"json",
		},
		domain.ErrorCategoryTimeout: {
			"timeout",
			"deadline",
			"context canceled",
			"operation timed out",
			"exceeded",
		},
		domain.ErrorCategoryOutput: {
			"write",
			"output",
			"format",
			"cannot create",
			"failed to generate",
			"report generation",
		},
		domain.ErrorCategoryProcessing: {
			"parse",
			"syntax",
			"analysis",
			"process",
			"failed to analyze",
			"invalid python",
			"compilation",
			"ast",
		},
	}
}

// Categorize determines the category of an error
func (ec *ErrorCategorizerImpl) Categorize(err error) *domain.CategorizedError {
	if err == nil {
		return nil
	}

	errMsg := strings.ToLower(err.Error())

	// Check each category's patterns
	for category, patterns := range ec.patterns {
		if containsAnyPattern(errMsg, patterns) {
			message := ec.getCategoryMessage(category)
			return &domain.CategorizedError{
				Category: category,
				Message:  message,
				Original: err,
			}
		}
	}

	// Default to unknown category
	return &domain.CategorizedError{
		Category: domain.ErrorCategoryUnknown,
		Message:  err.Error(),
		Original: err,
	}
}

// GetRecoverySuggestions returns recovery suggestions for an error category
func (ec *ErrorCategorizerImpl) GetRecoverySuggestions(category domain.ErrorCategory) []string {
	suggestions := map[domain.ErrorCategory][]string{
		domain.ErrorCategoryInput: {
			"Check that files/directories exist and contain Python files",
			"Try: pyscn analyze . --verbose to see detailed file discovery",
			"Ensure you have read permissions for the target files",
			"Use absolute paths if relative paths are causing issues",
		},
		domain.ErrorCategoryConfig: {
			"Verify configuration file format and values",
			"Try: pyscn init to generate a valid config file",
			"Check for syntax errors in .pyscn.toml or pyproject.toml",
			"Ensure all required configuration fields are present",
		},
		domain.ErrorCategoryTimeout: {
			"Consider analyzing smaller file sets or increasing timeout",
			"Try: Analyze specific files instead of entire directories",
			"Check if any files are unusually large or complex",
			"Consider using --skip-clones to speed up analysis",
		},
		domain.ErrorCategoryOutput: {
			"Check write permissions and output format validity",
			"Use --format text or check file system permissions",
			"Ensure output directory exists and is writable",
			"Try writing to a different location",
		},
		domain.ErrorCategoryProcessing: {
			"Some files may have syntax errors or be corrupted",
			"Run individual analysis types to isolate the problem",
			"Check Python version compatibility",
			"Try: python -m py_compile on files to check for syntax errors",
		},
		domain.ErrorCategoryUnknown: {
			"Run with --verbose for detailed error information",
			"Try: pyscn analyze . --verbose or check GitHub issues",
			"Check the documentation for known issues",
			"Report the issue if it persists",
		},
	}

	if sug, ok := suggestions[category]; ok {
		return sug
	}
	return []string{"Check the error message for more details"}
}

// getCategoryMessage returns a user-friendly message for an error category
func (ec *ErrorCategorizerImpl) getCategoryMessage(category domain.ErrorCategory) string {
	messages := map[domain.ErrorCategory]string{
		domain.ErrorCategoryInput:      "Failed to process input files or directories",
		domain.ErrorCategoryConfig:     "Configuration file or settings error",
		domain.ErrorCategoryTimeout:    "Analysis timed out",
		domain.ErrorCategoryOutput:     "Failed to generate or write output",
		domain.ErrorCategoryProcessing: "Error during code analysis processing",
		domain.ErrorCategoryUnknown:    "An unexpected error occurred",
	}

	if msg, ok := messages[category]; ok {
		return msg
	}
	return "An error occurred"
}

// containsAnyPattern checks if a string contains any of the given patterns
func containsAnyPattern(str string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(str, pattern) {
			return true
		}
	}
	return false
}
