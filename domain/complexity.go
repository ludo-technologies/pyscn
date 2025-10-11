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
)

// RiskLevel represents the complexity risk level
type RiskLevel string

const (
	RiskLevelLow    RiskLevel = "low"
	RiskLevelMedium RiskLevel = "medium"
	RiskLevelHigh   RiskLevel = "high"
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
	ShowDetails  bool

	// Filtering and sorting
	MinComplexity int
	MaxComplexity int // 0 means no limit
	SortBy        SortCriteria

	// Complexity thresholds
	LowThreshold    int
	MediumThreshold int

	// Configuration
	ConfigPath string

	// Analysis options
	Recursive       bool
	IncludePatterns []string
	ExcludePatterns []string
}

// ComplexityMetrics represents detailed complexity metrics for a function
type ComplexityMetrics struct {
	// McCabe cyclomatic complexity
	Complexity int

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

// ComplexitySummary represents aggregate statistics
type ComplexitySummary struct {
	TotalFunctions    int
	AverageComplexity float64
	MaxComplexity     int
	MinComplexity     int
	FilesAnalyzed     int

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

	// Warnings and issues
	Warnings []string
	Errors   []string

	// Metadata
	GeneratedAt string
	Version     string
	Config      interface{} // Configuration used for analysis
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
