package domain

import (
	"context"
	"encoding/json"
	"fmt"
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

// AnalysisScopeKind identifies the Python execution scope that owns a
// complexity result. Methods and nested functions are function scopes; class
// suites are class scopes because Python executes their statements separately
// while constructing the class object.
type AnalysisScopeKind string

const (
	AnalysisScopeModule   AnalysisScopeKind = "module"
	AnalysisScopeFunction AnalysisScopeKind = "function"
	AnalysisScopeClass    AnalysisScopeKind = "class"
)

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

// FunctionComplexity represents one executable Python scope. The historical
// type and field names remain part of the public API; ScopeKind distinguishes
// modules, functions, and class suites without duplicating the result model.
type FunctionComplexity struct {
	// Function identification
	Name        string
	ScopeKind   AnalysisScopeKind `json:"scope_kind" yaml:"scope_kind"`
	FilePath    string
	StartLine   int
	StartColumn int
	EndLine     int

	// Complexity metrics
	Metrics ComplexityMetrics

	// Risk assessment
	RiskLevel RiskLevel
}

// ValidateFunctionSLOCThresholds checks the long-function tiers against each
// other. A non-positive value means "not configured" and is left to the layer's
// own defaulting; only a fully specified, inverted pair is an error. Messages
// name the configuration keys, which match the CLI flags.
func ValidateFunctionSLOCThresholds(warn, critical int) error {
	if warn < 0 {
		return fmt.Errorf("function_sloc_warn_threshold must be >= 0, got %d", warn)
	}
	if critical < 0 {
		return fmt.Errorf("function_sloc_critical_threshold must be >= 0, got %d", critical)
	}
	if warn > 0 && critical > 0 && critical <= warn {
		return fmt.Errorf("function_sloc_critical_threshold (%d) must be > function_sloc_warn_threshold (%d)", critical, warn)
	}

	return nil
}

// ExceedsSLOC reports whether this function is longer than the given SLOC
// threshold. Module and class scopes never qualify because this is explicitly
// a long-function rule. A non-positive threshold disables the check.
func (f FunctionComplexity) ExceedsSLOC(threshold int) bool {
	if threshold <= 0 || f.ScopeKind != AnalysisScopeFunction {
		return false
	}
	return f.Metrics.SLOC > threshold
}

// DirectoryComplexityMetrics aggregates the complete analyzed scope population
// for one project-root-relative directory. FunctionCount and
// HighRiskFunctionCount retain their historical serialized names. Presentation
// filters do not change these metrics, matching the project-wide summary contract.
type DirectoryComplexityMetrics struct {
	DirectoryPath         string  `json:"directory_path" yaml:"directory_path"`
	FunctionCount         int     `json:"function_count" yaml:"function_count"`
	AverageComplexity     float64 `json:"average_complexity" yaml:"average_complexity"`
	MaxComplexity         int     `json:"max_complexity" yaml:"max_complexity"`
	HighRiskFunctionCount int     `json:"high_risk_function_count" yaml:"high_risk_function_count"`
	AverageNestingDepth   float64 `json:"average_nesting_depth" yaml:"average_nesting_depth"`
	MaxNestingDepth       int     `json:"max_nesting_depth" yaml:"max_nesting_depth"`
}

// DirectoryComplexityMetricsList is the stable serialized collection contract.
// A zero value is encoded as an empty array so callers never need to distinguish
// an uninitialized collection from a completed analysis with no reported rows.
type DirectoryComplexityMetricsList []DirectoryComplexityMetrics

// MarshalJSON encodes an uninitialized collection as an empty JSON array.
func (metrics DirectoryComplexityMetricsList) MarshalJSON() ([]byte, error) {
	if metrics == nil {
		return []byte("[]"), nil
	}
	type plainDirectoryComplexityMetricsList DirectoryComplexityMetricsList
	return json.Marshal(plainDirectoryComplexityMetricsList(metrics))
}

// MarshalYAML encodes an uninitialized collection as an empty YAML array.
func (metrics DirectoryComplexityMetricsList) MarshalYAML() (interface{}, error) {
	if metrics == nil {
		return []DirectoryComplexityMetrics{}, nil
	}
	return []DirectoryComplexityMetrics(metrics), nil
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

// ComplexitySummary represents aggregate statistics. Historical public field
// names refer to functions, but the population contains every executable scope
// and each record exposes its ScopeKind.
// Averages, min/max, risk counts, and the distribution are computed over every
// analyzed scope; min_complexity only limits which scopes are reported.
type ComplexitySummary struct {
	// TotalFunctions is the complete analyzed scope population used by all
	// aggregate metrics. Its historical name is retained for output compatibility.
	// Presentation filters only limit ComplexityResponse.Functions.
	TotalFunctions int
	// FunctionsParsed is retained for output compatibility and describes the same
	// complete analyzed scope population as TotalFunctions.
	FunctionsParsed            int
	AverageComplexity          float64
	AverageCognitiveComplexity float64
	AverageNestingDepth        float64
	MaxComplexity              int
	MinComplexity              int
	// FilesAnalyzed is the number of files that were successfully parsed and
	// contributed to the metrics above.
	FilesAnalyzed int
	// TotalFiles is the number of files the request covered, parsed or not.
	TotalFiles int
	// SkippedFiles is the number of files dropped because they could not be
	// read or parsed. Their contents are absent from every metric, so a
	// consumer must read this before trusting the aggregates.
	SkippedFiles int

	// Risk distribution
	LowRiskFunctions    int
	MediumRiskFunctions int
	HighRiskFunctions   int

	// Complexity distribution
	ComplexityDistribution map[string]int
}

// ComplexityResponse represents the complete analysis result
type ComplexityResponse struct {
	// Analysis results. Functions is the historical public field name for typed
	// execution scopes; inspect FunctionComplexity.ScopeKind before applying
	// function-only rules.
	Functions   []FunctionComplexity
	ByDirectory DirectoryComplexityMetricsList `json:"by_directory" yaml:"by_directory"`
	Summary     ComplexitySummary
	// AnalyzedFunctions is the complete population before presentation filters.
	// It is consumed by app-level aggregations and is not part of public output.
	AnalyzedFunctions []FunctionComplexity `json:"-" yaml:"-"`
	// ModuleRollups are derived before report filters are applied. They are consumed
	// by the unified analyze command and are not part of standalone complexity output.
	ModuleRollups map[string]ModuleComplexityMetrics `json:"-" yaml:"-"`

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

	// LoadDefaultConfig discovers configuration from targetPath (the analyzed
	// path) and falls back to built-in defaults when none is found
	LoadDefaultConfig(targetPath string) *ComplexityRequest

	// MergeConfig merges CLI flags with configuration file
	MergeConfig(base *ComplexityRequest, override *ComplexityRequest) *ComplexityRequest
}
