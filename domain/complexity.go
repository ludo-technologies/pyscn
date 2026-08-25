package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sort"
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
	AnalysisScopeUnknown  AnalysisScopeKind = ""
	AnalysisScopeModule   AnalysisScopeKind = "module"
	AnalysisScopeFunction AnalysisScopeKind = "function"
	AnalysisScopeClass    AnalysisScopeKind = "class"
)

// Validate reports whether the kind names a supported Python execution scope.
// Analyzer-owned records must never use AnalysisScopeUnknown.
func (k AnalysisScopeKind) Validate() error {
	switch k {
	case AnalysisScopeModule, AnalysisScopeFunction, AnalysisScopeClass:
		return nil
	default:
		return fmt.Errorf("invalid analysis scope kind %q", k)
	}
}

// ComplexityRequest represents a request for complexity analysis
type ComplexityRequest struct {
	// Input files or directories to analyze
	Paths []string `json:"paths" yaml:"paths"`

	// Output configuration
	OutputFormat OutputFormat `json:"output_format" yaml:"output_format"`
	OutputWriter io.Writer    `json:"-" yaml:"-"`
	OutputPath   string       `json:"output_path" yaml:"output_path"`   // Path to save output file (for HTML format)
	NoOpen       bool         `json:"no_open" yaml:"no_open"`           // Don't auto-open HTML in browser
	ShowDetails  *bool        `json:"show_details" yaml:"show_details"` // nil = unset, non-nil = explicitly set

	// Filtering and sorting
	MinComplexity int          `json:"min_complexity" yaml:"min_complexity"`
	MaxComplexity int          `json:"max_complexity" yaml:"max_complexity"` // 0 means no limit
	SortBy        SortCriteria `json:"sort_by" yaml:"sort_by"`

	// Complexity thresholds
	LowThreshold                 int `json:"low_threshold" yaml:"low_threshold"`
	MediumThreshold              int `json:"medium_threshold" yaml:"medium_threshold"`
	CognitiveComplexityThreshold int `json:"cognitive_complexity_threshold" yaml:"cognitive_complexity_threshold"`
	NestingDepthThreshold        int `json:"nesting_depth_threshold" yaml:"nesting_depth_threshold"`

	// Function SLOC thresholds
	FunctionSLOCWarnThreshold     int `json:"function_sloc_warn_threshold" yaml:"function_sloc_warn_threshold"`
	FunctionSLOCCriticalThreshold int `json:"function_sloc_critical_threshold" yaml:"function_sloc_critical_threshold"`

	// Analysis toggles loaded from configuration when present.
	// Nil means "use the default enabled behavior".
	Enabled         *bool `json:"enabled" yaml:"enabled"`
	ReportUnchanged *bool `json:"report_unchanged" yaml:"report_unchanged"`

	// Configuration
	ConfigPath string `json:"config_path" yaml:"config_path"`

	// Analysis options
	Recursive       *bool    `json:"recursive" yaml:"recursive"` // nil = unset, non-nil = explicitly set
	IncludePatterns []string `json:"include_patterns" yaml:"include_patterns"`
	ExcludePatterns []string `json:"exclude_patterns" yaml:"exclude_patterns"`
}

// ComplexityMetrics represents detailed complexity metrics for a function
type ComplexityMetrics struct {
	// McCabe cyclomatic complexity
	Complexity int `json:"complexity" yaml:"complexity"`

	// Cognitive complexity (SonarQube-style)
	CognitiveComplexity int `json:"cognitive_complexity" yaml:"cognitive_complexity"`

	// CFG metrics
	Nodes int `json:"nodes" yaml:"nodes"`
	Edges int `json:"edges" yaml:"edges"`

	// Nesting depth
	NestingDepth int `json:"nesting_depth" yaml:"nesting_depth"`

	// Statement counts
	IfStatements      int `json:"if_statements" yaml:"if_statements"`
	LoopStatements    int `json:"loop_statements" yaml:"loop_statements"`
	ExceptionHandlers int `json:"exception_handlers" yaml:"exception_handlers"`
	SwitchCases       int `json:"switch_cases" yaml:"switch_cases"`

	// SLOC is the source lines of code within this function's line range.
	// Computed using the same line-classification logic as raw_metrics.
	SLOC int `json:"sloc" yaml:"sloc"`
}

// FunctionComplexity represents one executable Python scope. The historical
// type and field names remain part of the public API; ScopeKind distinguishes
// modules, functions, and class suites without duplicating the result model.
type FunctionComplexity struct {
	// Function identification
	Name        string            `json:"name" yaml:"name"`
	ScopeKind   AnalysisScopeKind `json:"scope_kind" yaml:"scope_kind"`
	FilePath    string            `json:"file_path" yaml:"file_path"`
	StartLine   int               `json:"start_line" yaml:"start_line"`
	StartColumn int               `json:"start_column" yaml:"start_column"`
	EndLine     int               `json:"end_line" yaml:"end_line"`

	// Complexity metrics
	Metrics ComplexityMetrics `json:"metrics" yaml:"metrics"`

	// Risk assessment
	RiskLevel RiskLevel `json:"risk_level" yaml:"risk_level"`
}

// ScopeLabel returns the canonical user-facing name for an executable scope.
func (f FunctionComplexity) ScopeLabel() string {
	return executionScopeLabel(f.ScopeKind, f.Name)
}

func executionScopeLabel(kind AnalysisScopeKind, name string) string {
	if kind == AnalysisScopeClass {
		return "class scope " + name
	}
	return name
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

// DirectoryComplexityMetrics aggregates the complete analyzed function
// population for one project-root-relative directory. Presentation filters do
// not change these metrics, matching the project-wide summary contract.
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

// ComplexitySummary represents aggregate function statistics. The established
// module pseudo-record remains in this population; executable class suites are
// reported separately and do not alter function counts or averages.
type ComplexitySummary struct {
	// TotalFunctions is the complete analyzed function population used by all
	// aggregate metrics, including the established module pseudo-record.
	// Presentation filters only limit ComplexityResponse.Functions.
	TotalFunctions int `json:"total_functions" yaml:"total_functions"`
	// TotalClassScopes is the complete executable class-suite population.
	// Class maxima are published separately and do not alter function aggregates
	// or health-score semantics.
	TotalClassScopes            int `json:"total_class_scopes" yaml:"total_class_scopes"`
	MaxClassComplexity          int `json:"max_class_complexity" yaml:"max_class_complexity"`
	MaxClassCognitiveComplexity int `json:"max_class_cognitive_complexity" yaml:"max_class_cognitive_complexity"`
	MaxClassNestingDepth        int `json:"max_class_nesting_depth" yaml:"max_class_nesting_depth"`
	HighRiskClassScopes         int `json:"high_risk_class_scopes" yaml:"high_risk_class_scopes"`
	// FunctionsParsed is retained for output compatibility and describes the same
	// complete analyzed function population as TotalFunctions.
	FunctionsParsed            int     `json:"functions_parsed" yaml:"functions_parsed"`
	AverageComplexity          float64 `json:"average_complexity" yaml:"average_complexity"`
	AverageCognitiveComplexity float64 `json:"average_cognitive_complexity" yaml:"average_cognitive_complexity"`
	AverageNestingDepth        float64 `json:"average_nesting_depth" yaml:"average_nesting_depth"`
	MaxComplexity              int     `json:"max_complexity" yaml:"max_complexity"`
	MinComplexity              int     `json:"min_complexity" yaml:"min_complexity"`
	// FilesAnalyzed is the number of files that were successfully parsed and
	// contributed to the metrics above.
	FilesAnalyzed int `json:"files_analyzed" yaml:"files_analyzed"`
	// TotalFiles is the number of files the request covered, parsed or not.
	TotalFiles int `json:"total_files" yaml:"total_files"`
	// SkippedFiles is the number of files dropped because they could not be
	// read or parsed. Their contents are absent from every metric, so a
	// consumer must read this before trusting the aggregates.
	SkippedFiles int `json:"skipped_files" yaml:"skipped_files"`

	// Risk distribution
	LowRiskFunctions    int `json:"low_risk_functions" yaml:"low_risk_functions"`
	MediumRiskFunctions int `json:"medium_risk_functions" yaml:"medium_risk_functions"`
	HighRiskFunctions   int `json:"high_risk_functions" yaml:"high_risk_functions"`

	// Complexity distribution
	ComplexityDistribution map[string]int `json:"complexity_distribution" yaml:"complexity_distribution"`
}

// ComplexityResponse represents the complete analysis result
type ComplexityResponse struct {
	// Functions retains the established module and function population.
	Functions []FunctionComplexity `json:"functions" yaml:"functions"`
	// ClassScopes is an additive collection of executable class-suite results.
	// It uses the same typed metric record without changing function summaries.
	ClassScopes []FunctionComplexity           `json:"class_scopes,omitempty" yaml:"class_scopes,omitempty"`
	ByDirectory DirectoryComplexityMetricsList `json:"by_directory" yaml:"by_directory"`
	Summary     ComplexitySummary              `json:"summary" yaml:"summary"`
	// AnalyzedFunctions is the complete population before presentation filters.
	// It is consumed by app-level aggregations and is not part of public output.
	AnalyzedFunctions []FunctionComplexity `json:"-" yaml:"-"`
	// AnalyzedClassScopes is the complete class-suite population before
	// presentation filters and is not part of public output.
	AnalyzedClassScopes []FunctionComplexity `json:"-" yaml:"-"`
	// ModuleRollups are derived before report filters are applied. They are consumed
	// by the unified analyze command and are not part of standalone complexity output.
	ModuleRollups map[string]ModuleComplexityMetrics `json:"-" yaml:"-"`

	// File-level raw code metrics
	RawMetrics        []RawMetrics       `json:"raw_metrics,omitempty" yaml:"raw_metrics,omitempty"`
	RawMetricsSummary *RawMetricsSummary `json:"raw_metrics_summary,omitempty" yaml:"raw_metrics_summary,omitempty"`

	// Warnings and issues
	Warnings []string          `json:"warnings" yaml:"warnings"`
	Errors   []string          `json:"errors" yaml:"errors"`
	Failures []AnalysisFailure `json:"failures,omitempty" yaml:"failures,omitempty"`

	// Metadata
	GeneratedAt string             `json:"generated_at" yaml:"generated_at"`
	Version     string             `json:"version" yaml:"version"`
	Config      interface{}        `json:"config" yaml:"config"` // Configuration used for analysis
	Request     *ComplexityRequest `json:"request,omitempty"`    // Merged configuration request
}

// ReportedScopes returns the complete visible execution-scope population in
// the requested order. The returned slice never aliases response storage.
func (r *ComplexityResponse) ReportedScopes(sortBy SortCriteria) ([]FunctionComplexity, error) {
	if r == nil {
		return nil, fmt.Errorf("complexity response is nil")
	}
	return SortComplexityScopesBy(r.reportedScopes(), sortBy)
}

func (r *ComplexityResponse) reportedScopes() []FunctionComplexity {
	scopes := make([]FunctionComplexity, 0, len(r.Functions)+len(r.ClassScopes))
	scopes = append(scopes, r.Functions...)
	scopes = append(scopes, r.ClassScopes...)
	return scopes
}

// ReportedScopesByComplexity returns all visible scopes in a stable severity
// order for presentation. It never mutates response storage.
func (r *ComplexityResponse) ReportedScopesByComplexity() []FunctionComplexity {
	if r == nil {
		return nil
	}
	return SortComplexityScopes(r.reportedScopes())
}

// AnalyzedScopes returns an independently owned copy of the complete,
// pre-filter population. Both collections must be initialized by the analysis
// producer, including when either population is empty.
func (r *ComplexityResponse) AnalyzedScopes() ([]FunctionComplexity, error) {
	if err := r.ValidateAnalyzedScopes(); err != nil {
		return nil, err
	}

	scopes := make([]FunctionComplexity, 0, len(r.AnalyzedFunctions)+len(r.AnalyzedClassScopes))
	scopes = append(scopes, r.AnalyzedFunctions...)
	scopes = append(scopes, r.AnalyzedClassScopes...)
	return scopes, nil
}

// ValidateAnalyzedScopes enforces the producer-owned population contract
// without allocating a combined result slice.
func (r *ComplexityResponse) ValidateAnalyzedScopes() error {
	if r == nil {
		return fmt.Errorf("complexity response is nil")
	}
	if r.AnalyzedFunctions == nil {
		return fmt.Errorf("analyzed function population is not initialized")
	}
	if r.AnalyzedClassScopes == nil {
		return fmt.Errorf("analyzed class-scope population is not initialized")
	}

	for i, scope := range r.AnalyzedFunctions {
		if err := scope.ScopeKind.Validate(); err != nil {
			return fmt.Errorf("analyzed function %d: %w", i, err)
		}
		if scope.ScopeKind == AnalysisScopeClass {
			return fmt.Errorf("analyzed function %d has class scope kind", i)
		}
	}
	for i, scope := range r.AnalyzedClassScopes {
		if err := scope.ScopeKind.Validate(); err != nil {
			return fmt.Errorf("analyzed class scope %d: %w", i, err)
		}
		if scope.ScopeKind != AnalysisScopeClass {
			return fmt.Errorf("analyzed class scope %d has %q scope kind", i, scope.ScopeKind)
		}
	}
	return nil
}

// SortComplexityScopes returns an independently owned severity-ranked copy.
// Ties use source identity so output is deterministic.
func SortComplexityScopes(scopes []FunctionComplexity) []FunctionComplexity {
	return sortComplexityScopes(scopes, SortByComplexity)
}

// SortComplexityScopesBy returns an independently owned copy ordered by one
// of the complexity report's supported criteria. Ties use source identity so
// independently collected scope kinds still produce deterministic output.
func SortComplexityScopesBy(scopes []FunctionComplexity, sortBy SortCriteria) ([]FunctionComplexity, error) {
	switch sortBy {
	case SortByComplexity, SortByName, SortByRisk:
		return sortComplexityScopes(scopes, sortBy), nil
	default:
		return nil, fmt.Errorf("unsupported complexity sort criteria: %s", sortBy)
	}
}

func sortComplexityScopes(scopes []FunctionComplexity, sortBy SortCriteria) []FunctionComplexity {
	scopes = append([]FunctionComplexity(nil), scopes...)
	sort.SliceStable(scopes, func(i, j int) bool {
		left, right := scopes[i], scopes[j]
		switch sortBy {
		case SortByName:
			if left.Name != right.Name {
				return left.Name < right.Name
			}
		case SortByRisk:
			leftRisk, rightRisk := complexityRiskRank(left.RiskLevel), complexityRiskRank(right.RiskLevel)
			if leftRisk != rightRisk {
				return leftRisk > rightRisk
			}
			if left.Metrics.Complexity != right.Metrics.Complexity {
				return left.Metrics.Complexity > right.Metrics.Complexity
			}
		default:
			if left.Metrics.Complexity != right.Metrics.Complexity {
				return left.Metrics.Complexity > right.Metrics.Complexity
			}
		}
		if left.FilePath != right.FilePath {
			return left.FilePath < right.FilePath
		}
		if left.StartLine != right.StartLine {
			return left.StartLine < right.StartLine
		}
		if left.StartColumn != right.StartColumn {
			return left.StartColumn < right.StartColumn
		}
		if left.ScopeKind != right.ScopeKind {
			return left.ScopeKind < right.ScopeKind
		}
		return left.Name < right.Name
	})
	return scopes
}

func complexityRiskRank(risk RiskLevel) int {
	switch risk {
	case RiskLevelHigh:
		return 3
	case RiskLevelMedium:
		return 2
	case RiskLevelLow:
		return 1
	default:
		return 0
	}
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
