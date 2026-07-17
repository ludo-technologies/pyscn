package domain

import (
	"context"
	"io"
)

// OutputFormat represents the supported output formats
type OutputFormat string

const (
	OutputFormatText OutputFormat = "text"
	OutputFormatJSON OutputFormat = "json"
	OutputFormatYAML OutputFormat = "yaml"
	OutputFormatCSV  OutputFormat = "csv"
	OutputFormatHTML OutputFormat = "html"
	OutputFormatDOT  OutputFormat = "dot"
)

// SortCriteria represents the criteria for sorting results
type SortCriteria string

const (
	SortByComplexity SortCriteria = "complexity"
	SortByName       SortCriteria = "name"
	SortByRisk       SortCriteria = "risk"
	SortBySimilarity SortCriteria = "similarity"
	SortBySize       SortCriteria = "size"
	SortByLocation   SortCriteria = "location"
	SortByCoupling   SortCriteria = "coupling" // For CBO metrics
	SortBySeverity   SortCriteria = "severity" // For anti-pattern findings
	SortByCohesion   SortCriteria = "cohesion" // For LCOM metrics
)

// RiskLevel represents the complexity risk level
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
)

// ModuleFunctionName is the user-facing label used for module-scope (top-level) code
// in places that key/display per-function results. The angle brackets follow Python's
// own convention (e.g. tracebacks and `dis` output) and signal that this is not a real
// function defined in the source.
const ModuleFunctionName = "<module>"

// ComplexityRequest represents a request for complexity analysis
type ComplexityRequest struct {
	// Input files or directories to analyze
	Paths []string

	// Output configuration
	OutputFormat OutputFormat
	OutputWriter io.Writer
	OutputPath   string // Path to save output file (for HTML format)
	NoOpen       bool   // Don't auto-open HTML in browser
	ShowDetails  *bool  // nil = unset, non-nil = explicitly set

	// Filtering and sorting
	MinComplexity int
	MaxComplexity int // 0 means no limit
	SortBy        SortCriteria

	// Complexity thresholds
	LowThreshold                 int
	MediumThreshold              int
	CognitiveComplexityThreshold int
	NestingDepthThreshold        int

	// Function SLOC thresholds
	FunctionSLOCWarnThreshold     int
	FunctionSLOCCriticalThreshold int

	// Analysis toggles loaded from configuration when present.
	// Nil means "use the default enabled behavior".
	Enabled         *bool
	ReportUnchanged *bool

	// Configuration
	ConfigPath string

	// Analysis options
	Recursive       *bool // nil = unset, non-nil = explicitly set
	IncludePatterns []string
	ExcludePatterns []string
}

// ComplexityMetrics represents detailed complexity metrics for a function
type ComplexityMetrics struct {
	// McCabe cyclomatic complexity
	Complexity int

	// Cognitive complexity (SonarQube-style)
	CognitiveComplexity int

	// CFG metrics
	Nodes int
	Edges int

	// Nesting depth
	NestingDepth int

	// Statement counts
	IfStatements      int
	LoopStatements    int
	ExceptionHandlers int
	SwitchCases       int

	// SLOC is the source lines of code within this function's line range.
	// Computed using the same line-classification logic as raw_metrics.
	SLOC int
}

// FunctionComplexity represents complexity analysis result for a single function
type FunctionComplexity struct {
	// Function identification
	Name        string
	FilePath    string
	StartLine   int
	StartColumn int
	EndLine     int

	// Complexity metrics
	Metrics ComplexityMetrics

	// Risk assessment
	RiskLevel RiskLevel
}

// RawMetrics represents file-level raw code metrics.
type RawMetrics struct {
	FilePath       string  `json:"file_path" yaml:"file_path"`
	SLOC           int     `json:"sloc" yaml:"sloc"`
	LLOC           int     `json:"lloc" yaml:"lloc"`
	CommentLines   int     `json:"comment_lines" yaml:"comment_lines"`
	DocstringLines int     `json:"docstring_lines" yaml:"docstring_lines"`
	BlankLines     int     `json:"blank_lines" yaml:"blank_lines"`
	TotalLines     int     `json:"total_lines" yaml:"total_lines"`
	CommentRatio   float64 `json:"comment_ratio" yaml:"comment_ratio"`
}

// RawMetricsSummary represents aggregated raw code metrics across files.
type RawMetricsSummary struct {
	FilesAnalyzed  int     `json:"files_analyzed" yaml:"files_analyzed"`
	SLOC           int     `json:"sloc" yaml:"sloc"`
	LLOC           int     `json:"lloc" yaml:"lloc"`
	CommentLines   int     `json:"comment_lines" yaml:"comment_lines"`
	DocstringLines int     `json:"docstring_lines" yaml:"docstring_lines"`
	BlankLines     int     `json:"blank_lines" yaml:"blank_lines"`
	TotalLines     int     `json:"total_lines" yaml:"total_lines"`
	CommentRatio   float64 `json:"comment_ratio" yaml:"comment_ratio"`
}

// ComplexitySummary represents aggregate statistics
type ComplexitySummary struct {
	// TotalFunctions is the post-filter count (functions included in results after min_complexity filtering).
	TotalFunctions int
	// FunctionsParsed is the pre-filter count of all functions parsed before min_complexity filtering.
	// When min_complexity drops trivial functions, FunctionsParsed > TotalFunctions.
	FunctionsParsed            int
	AverageComplexity          float64
	AverageCognitiveComplexity float64
	AverageNestingDepth        float64
	MaxComplexity              int
	MinComplexity              int
	FilesAnalyzed              int

	// Risk distribution
	LowRiskFunctions    int
	MediumRiskFunctions int
	HighRiskFunctions   int

	// Complexity distribution
	ComplexityDistribution map[string]int
}

// ComplexityResponse represents the complete analysis result
type ComplexityResponse struct {
	// Analysis results
	Functions []FunctionComplexity
	Summary   ComplexitySummary

	// File-level raw code metrics
	RawMetrics        []RawMetrics       `json:"raw_metrics,omitempty" yaml:"raw_metrics,omitempty"`
	RawMetricsSummary *RawMetricsSummary `json:"raw_metrics_summary,omitempty" yaml:"raw_metrics_summary,omitempty"`

	// Warnings and issues
	Warnings []string
	Errors   []string

	// Metadata
	GeneratedAt string
	Version     string
	Config      interface{}        // Configuration used for analysis
	Request     *ComplexityRequest `json:"request,omitempty"` // Merged configuration request
}

// ComplexityService defines the core business logic for complexity analysis
type ComplexityService interface {
	// Analyze performs complexity analysis on the given request
	Analyze(ctx context.Context, req ComplexityRequest) (*ComplexityResponse, error)

	// AnalyzeFile analyzes a single Python file
	AnalyzeFile(ctx context.Context, filePath string, req ComplexityRequest) (*ComplexityResponse, error)
}

// FileReader defines the interface for reading and collecting Python files
type FileReader interface {
	// CollectPythonFiles recursively finds all Python files in the given paths
	CollectPythonFiles(paths []string, recursive bool, includePatterns, excludePatterns []string) ([]string, error)

	// ReadFile reads the content of a file
	ReadFile(path string) ([]byte, error)

	// IsValidPythonFile checks if a file is a valid Python file
	IsValidPythonFile(path string) bool

	// FileExists checks if a file exists and returns an error if not
	FileExists(path string) (bool, error)
}

// OutputFormatter defines the interface for formatting analysis results
type OutputFormatter interface {
	// Format formats the analysis response according to the specified format
	Format(response *ComplexityResponse, format OutputFormat) (string, error)

	// Write writes the formatted output to the writer
	Write(response *ComplexityResponse, format OutputFormat, writer io.Writer) error
}

// ConfigurationLoader defines the interface for loading configuration
type ConfigurationLoader interface {
	// LoadConfig loads configuration from the specified path
	LoadConfig(path string) (*ComplexityRequest, error)

	// LoadDefaultConfig loads the default configuration
	LoadDefaultConfig() *ComplexityRequest

	// MergeConfig merges CLI flags with configuration file
	MergeConfig(base *ComplexityRequest, override *ComplexityRequest) *ComplexityRequest
}
