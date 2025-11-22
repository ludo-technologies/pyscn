package domain

import (
	"context"
	"io"
)

// DeadCodeSeverity represents the severity level of dead code findings
type DeadCodeSeverity string

const (
	DeadCodeSeverityCritical DeadCodeSeverity = "critical"
	DeadCodeSeverityWarning  DeadCodeSeverity = "warning"
	DeadCodeSeverityInfo     DeadCodeSeverity = "info"
)

// DeadCodeSortCriteria represents the criteria for sorting dead code results
type DeadCodeSortCriteria string

const (
	DeadCodeSortBySeverity DeadCodeSortCriteria = "severity"
	DeadCodeSortByLine     DeadCodeSortCriteria = "line"
	DeadCodeSortByFile     DeadCodeSortCriteria = "file"
	DeadCodeSortByFunction DeadCodeSortCriteria = "function"
)

// DeadCodeRequest represents a request for dead code analysis
type DeadCodeRequest struct {
	// Input files or directories to analyze
	Paths []string

	// Output configuration
	OutputFormat OutputFormat
	OutputWriter io.Writer
	OutputPath   string // Path to save output file (for HTML format)
	NoOpen       bool   // Don't auto-open HTML in browser
	ShowContext  *bool  // nil = use default (false), non-nil = explicitly set
	ContextLines int    // Number of lines to show around dead code

	// Filtering and sorting
	MinSeverity DeadCodeSeverity
	SortBy      DeadCodeSortCriteria

	// Analysis options
	Recursive       bool
	IncludePatterns []string
	ExcludePatterns []string
	IgnorePatterns  []string // Patterns for code to ignore (e.g., comments, debug code)

	// Configuration
	ConfigPath string

	// Dead code specific options
	DetectAfterReturn         *bool // nil = use default (true), non-nil = explicitly set
	DetectAfterBreak          *bool // nil = use default (true), non-nil = explicitly set
	DetectAfterContinue       *bool // nil = use default (true), non-nil = explicitly set
	DetectAfterRaise          *bool // nil = use default (true), non-nil = explicitly set
	DetectUnreachableBranches *bool // nil = use default (true), non-nil = explicitly set
}

// DeadCodeLocation represents the location of dead code
type DeadCodeLocation struct {
	FilePath    string `json:"file_path"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	StartColumn int    `json:"start_column"`
	EndColumn   int    `json:"end_column"`
}

// DeadCodeFinding represents a single dead code detection result
type DeadCodeFinding struct {
	// Location information
	Location DeadCodeLocation `json:"location"`

	// Function context
	FunctionName string `json:"function_name"`

	// Dead code details
	Code        string           `json:"code"`
	Reason      string           `json:"reason"`
	Severity    DeadCodeSeverity `json:"severity"`
	Description string           `json:"description"`

	// Context information (surrounding code)
	Context []string `json:"context,omitempty"`

	// Metadata
	BlockID string `json:"block_id,omitempty"`
}

// FunctionDeadCode represents dead code analysis result for a single function
type FunctionDeadCode struct {
	// Function identification
	Name     string `json:"name"`
	FilePath string `json:"file_path"`

	// Dead code findings
	Findings []DeadCodeFinding `json:"findings"`

	// Function metrics
	TotalBlocks    int     `json:"total_blocks"`
	DeadBlocks     int     `json:"dead_blocks"`
	ReachableRatio float64 `json:"reachable_ratio"`

	// Summary by severity
	CriticalCount int `json:"critical_count"`
	WarningCount  int `json:"warning_count"`
	InfoCount     int `json:"info_count"`
}

// FileDeadCode represents dead code analysis result for a single file
type FileDeadCode struct {
	// File identification
	FilePath string `json:"file_path"`

	// Functions analyzed
	Functions []FunctionDeadCode `json:"functions"`

	// File-level summary
	TotalFindings     int     `json:"total_findings"`
	TotalFunctions    int     `json:"total_functions"`
	AffectedFunctions int     `json:"affected_functions"`
	DeadCodeRatio     float64 `json:"dead_code_ratio"`
}

// DeadCodeSummary represents aggregate statistics for dead code analysis
type DeadCodeSummary struct {
	// Overall metrics
	TotalFiles            int `json:"total_files"`
	TotalFunctions        int `json:"total_functions"`
	TotalFindings         int `json:"total_findings"`
	FilesWithDeadCode     int `json:"files_with_dead_code"`
	FunctionsWithDeadCode int `json:"functions_with_dead_code"`

	// Severity distribution
	CriticalFindings int `json:"critical_findings"`
	WarningFindings  int `json:"warning_findings"`
	InfoFindings     int `json:"info_findings"`

	// Reason distribution
	FindingsByReason map[string]int `json:"findings_by_reason"`

	// Coverage metrics
	TotalBlocks      int     `json:"total_blocks"`
	DeadBlocks       int     `json:"dead_blocks"`
	OverallDeadRatio float64 `json:"overall_dead_ratio"`
}

// DeadCodeResponse represents the complete dead code analysis result
type DeadCodeResponse struct {
	// Analysis results
	Files   []FileDeadCode  `json:"files"`
	Summary DeadCodeSummary `json:"summary"`

	// Warnings and issues
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`

	// Metadata
	GeneratedAt string      `json:"generated_at"`
	Version     string      `json:"version"`
	Config      interface{} `json:"config"` // Configuration used for analysis
}

// DeadCodeService defines the core business logic for dead code analysis
type DeadCodeService interface {
	// Analyze performs dead code analysis on the given request
	Analyze(ctx context.Context, req DeadCodeRequest) (*DeadCodeResponse, error)

	// AnalyzeFile analyzes a single Python file for dead code
	AnalyzeFile(ctx context.Context, filePath string, req DeadCodeRequest) (*FileDeadCode, error)

	// AnalyzeFunction analyzes a single function for dead code
	AnalyzeFunction(ctx context.Context, functionCFG interface{}, req DeadCodeRequest) (*FunctionDeadCode, error)
}

// DeadCodeConfigurationLoader defines the interface for loading dead code configuration
type DeadCodeConfigurationLoader interface {
	// LoadConfig loads dead code configuration from the specified path
	LoadConfig(path string) (*DeadCodeRequest, error)

	// LoadDefaultConfig loads the default dead code configuration
	LoadDefaultConfig() *DeadCodeRequest

	// MergeConfig merges CLI flags with configuration file
	MergeConfig(base *DeadCodeRequest, override *DeadCodeRequest) *DeadCodeRequest
}

// DeadCodeFormatter defines the interface for formatting dead code analysis results
type DeadCodeFormatter interface {
	// Format formats the dead code analysis response according to the specified format
	Format(response *DeadCodeResponse, format OutputFormat) (string, error)

	// Write writes the formatted dead code output to the writer
	Write(response *DeadCodeResponse, format OutputFormat, writer io.Writer) error

	// FormatFinding formats a single dead code finding
	FormatFinding(finding DeadCodeFinding, format OutputFormat) (string, error)
}

// Helper functions for pointer boolean handling

// BoolPtr creates a pointer to a boolean value
// This is useful for creating pointer boolean values inline
func BoolPtr(b bool) *bool {
	return &b
}

// BoolValue safely dereferences a boolean pointer, returning defaultVal if nil
// This allows safe access to pointer booleans with explicit defaults
func BoolValue(b *bool, defaultVal bool) bool {
	if b == nil {
		return defaultVal
	}
	return *b
}

// Default configuration values for dead code analysis
func DefaultDeadCodeRequest() *DeadCodeRequest {
	return &DeadCodeRequest{
		OutputFormat:    OutputFormatText,
		ShowContext:     BoolPtr(false),
		ContextLines:    3,
		MinSeverity:     DeadCodeSeverityWarning,
		SortBy:          DeadCodeSortBySeverity,
		Recursive:       true,
		IncludePatterns: []string{"**/*.py"},
		ExcludePatterns: []string{"test_*.py", "*_test.py"},
		IgnorePatterns:  []string{},

		// Dead code detection options (all enabled by default)
		DetectAfterReturn:         BoolPtr(true),
		DetectAfterBreak:          BoolPtr(true),
		DetectAfterContinue:       BoolPtr(true),
		DetectAfterRaise:          BoolPtr(true),
		DetectUnreachableBranches: BoolPtr(true),
	}
}

// Validation methods

// Validate validates the dead code request
func (req *DeadCodeRequest) Validate() error {
	if len(req.Paths) == 0 {
		return NewInvalidInputError("at least one path must be specified", nil)
	}

	if req.ContextLines < 0 {
		return NewInvalidInputError("context lines must be >= 0", nil)
	}

	// Validate output format
	validFormats := map[OutputFormat]bool{
		OutputFormatText: true,
		OutputFormatJSON: true,
		OutputFormatYAML: true,
		OutputFormatCSV:  true,
		OutputFormatHTML: true,
	}

	if !validFormats[req.OutputFormat] {
		return NewInvalidInputError("invalid output format", nil)
	}

	// Validate severity level
	validSeverities := map[DeadCodeSeverity]bool{
		DeadCodeSeverityCritical: true,
		DeadCodeSeverityWarning:  true,
		DeadCodeSeverityInfo:     true,
	}

	if !validSeverities[req.MinSeverity] {
		return NewInvalidInputError("invalid minimum severity level", nil)
	}

	// Validate sort criteria
	validSortBy := map[DeadCodeSortCriteria]bool{
		DeadCodeSortBySeverity: true,
		DeadCodeSortByLine:     true,
		DeadCodeSortByFile:     true,
		DeadCodeSortByFunction: true,
	}

	if !validSortBy[req.SortBy] {
		return NewInvalidInputError("invalid sort criteria", nil)
	}

	return nil
}

// Helper methods for severity comparison

// SeverityLevel returns the numeric level for comparison
func (s DeadCodeSeverity) Level() int {
	switch s {
	case DeadCodeSeverityInfo:
		return 1
	case DeadCodeSeverityWarning:
		return 2
	case DeadCodeSeverityCritical:
		return 3
	default:
		return 0
	}
}

// IsAtLeast checks if the severity is at least the specified level
func (s DeadCodeSeverity) IsAtLeast(minSeverity DeadCodeSeverity) bool {
	return s.Level() >= minSeverity.Level()
}

// Summary calculation helpers

// CalculateSeverityCounts calculates the count of findings by severity
func (fdc *FunctionDeadCode) CalculateSeverityCounts() {
	fdc.CriticalCount = 0
	fdc.WarningCount = 0
	fdc.InfoCount = 0

	for _, finding := range fdc.Findings {
		switch finding.Severity {
		case DeadCodeSeverityCritical:
			fdc.CriticalCount++
		case DeadCodeSeverityWarning:
			fdc.WarningCount++
		case DeadCodeSeverityInfo:
			fdc.InfoCount++
		}
	}
}

// HasFindings returns true if the function has any dead code findings
func (fdc *FunctionDeadCode) HasFindings() bool {
	return len(fdc.Findings) > 0
}

// HasFindingsAtSeverity returns true if the function has findings at or above the specified severity
func (fdc *FunctionDeadCode) HasFindingsAtSeverity(minSeverity DeadCodeSeverity) bool {
	for _, finding := range fdc.Findings {
		if finding.Severity.IsAtLeast(minSeverity) {
			return true
		}
	}
	return false
}

// GetFindingsAtSeverity returns findings at or above the specified severity level
func (fdc *FunctionDeadCode) GetFindingsAtSeverity(minSeverity DeadCodeSeverity) []DeadCodeFinding {
	var filtered []DeadCodeFinding
	for _, finding := range fdc.Findings {
		if finding.Severity.IsAtLeast(minSeverity) {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}
