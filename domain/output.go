package domain

import (
	"context"
	"io"
	"time"
)

// ReportWriter abstracts writing reports to a destination (file or writer)
// and handling side-effects like opening HTML reports in a browser.
//
// Implementations live in the service layer.
type ReportWriter interface {
	// Write writes formatted content using the provided writeFunc.
	// - If outputPath is non-empty, implementations should create/truncate the file
	//   at that path and pass the file as the writer to writeFunc.
	// - If outputPath is empty, implementations should pass the provided writer to writeFunc.
	// Implementations may emit user-facing status messages (e.g., file paths) and
	// optionally open HTML outputs in a browser when format is OutputFormatHTML and noOpen is false.
	Write(writer io.Writer, outputPath string, format OutputFormat, noOpen bool, writeFunc func(io.Writer) error) error
}

// ProgressManager manages progress tracking for analysis
type ProgressManager interface {
	// Initialize sets up progress tracking with the maximum value
	Initialize(maxValue int)

	// Start starts the progress bar
	Start()

	// Complete marks the progress as completed
	Complete(success bool)

	// Update updates the progress
	Update(processed, total int)

	// SetWriter sets the output writer for progress bars
	SetWriter(writer io.Writer)

	// IsInteractive returns true if progress bars should be shown
	IsInteractive() bool

	// Close cleans up any resources
	Close()
}

// ParallelExecutor manages parallel execution of tasks
type ParallelExecutor interface {
	// Execute runs tasks in parallel with the given configuration
	Execute(ctx context.Context, tasks []ExecutableTask) error

	// SetMaxConcurrency sets the maximum number of concurrent tasks
	SetMaxConcurrency(max int)

	// SetTimeout sets the timeout for all tasks
	SetTimeout(timeout time.Duration)
}

// ExecutableTask represents a task that can be executed in parallel
type ExecutableTask interface {
	// Name returns the name of the task
	Name() string

	// Execute runs the task and returns the result
	Execute(ctx context.Context) (interface{}, error)

	// IsEnabled returns whether the task should be executed
	IsEnabled() bool
}

// ErrorCategory represents the category of an error
type ErrorCategory string

const (
	ErrorCategoryInput      ErrorCategory = "Input Error"
	ErrorCategoryConfig     ErrorCategory = "Configuration Error"
	ErrorCategoryProcessing ErrorCategory = "Processing Error"
	ErrorCategoryOutput     ErrorCategory = "Output Error"
	ErrorCategoryTimeout    ErrorCategory = "Timeout Error"
	ErrorCategoryUnknown    ErrorCategory = "Unknown Error"
)

// CategorizedError represents an error with category information
type CategorizedError struct {
	Category ErrorCategory
	Message  string
	Original error
}

// Error implements the error interface
func (e *CategorizedError) Error() string {
	if e.Original != nil {
		return e.Original.Error()
	}
	return e.Message
}

// ErrorCategorizer categorizes errors for better reporting
type ErrorCategorizer interface {
	// Categorize determines the category of an error
	Categorize(err error) *CategorizedError

	// GetRecoverySuggestions returns recovery suggestions for an error category
	GetRecoverySuggestions(category ErrorCategory) []string
}
