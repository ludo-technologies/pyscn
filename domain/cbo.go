package domain

import (
	"context"
	"io"
)

// CBORequest represents a request for CBO (Coupling Between Objects) analysis
type CBORequest struct {
	// Input files or directories to analyze
	Paths []string

	// Output configuration
	OutputFormat OutputFormat
	OutputWriter io.Writer
	OutputPath   string // Path to save output file (for HTML format)
	NoOpen       bool   // Don't auto-open HTML in browser
	ShowDetails  bool

	// Filtering and sorting
	MinCBO    int
	MaxCBO    int // 0 means no limit
	SortBy    SortCriteria
	ShowZeros *bool // Include classes with CBO = 0

	// CBO thresholds for risk assessment
	LowThreshold    int // Default: 3 (industry standard)
	MediumThreshold int // Default: 7 (industry standard)

	// Configuration
	ConfigPath string

	// Analysis options
	Recursive       *bool
	IncludePatterns []string
	ExcludePatterns []string

	// Analysis scope
	IncludeBuiltins *bool // Include dependencies on built-in types
	IncludeImports  *bool // Include imported modules in dependency count
}

// CBOMetrics represents detailed CBO metrics for a class
type CBOMetrics struct {
	// Core CBO metric - number of classes this class depends on
	CouplingCount int

	// Breakdown by dependency type
	InheritanceDependencies     int // Base classes
	TypeHintDependencies        int // Type annotations
	InstantiationDependencies   int // Object creation
	AttributeAccessDependencies int // Method calls and attribute access
	ImportDependencies          int // Explicitly imported classes

	// Dependency details
	DependentClasses []string // List of class names this class depends on
}

// ClassCoupling represents CBO analysis result for a single class
type ClassCoupling struct {
	// Class identification
	Name      string
	FilePath  string
	StartLine int
	EndLine   int

	// CBO metrics
	Metrics CBOMetrics

	// Risk assessment
	RiskLevel RiskLevel

	// Additional context
	IsAbstract  bool
	BaseClasses []string
}

// CBOSummary represents aggregate CBO statistics
type CBOSummary struct {
	TotalClasses    int
	AverageCBO      float64
	MaxCBO          int
	MinCBO          int
	ClassesAnalyzed int
	FilesAnalyzed   int

	// Risk distribution
	LowRiskClasses    int
	MediumRiskClasses int
	HighRiskClasses   int

	// CBO distribution
	CBODistribution map[string]int

	// Most coupled classes (top 10)
	MostCoupledClasses []ClassCoupling

	// Classes with highest impact (most depended upon)
	MostDependedUponClasses []string
}

// CBOResponse represents the complete CBO analysis result
type CBOResponse struct {
	// Analysis results
	Classes []ClassCoupling
	Summary CBOSummary

	// Warnings and issues
	Warnings []string
	Errors   []string

	// Metadata
	GeneratedAt string
	Version     string
	Config      interface{} // Configuration used for analysis
}

// CBOService defines the core business logic for CBO analysis
type CBOService interface {
	// Analyze performs CBO analysis on the given request
	Analyze(ctx context.Context, req CBORequest) (*CBOResponse, error)

	// AnalyzeFile analyzes a single Python file
	AnalyzeFile(ctx context.Context, filePath string, req CBORequest) (*CBOResponse, error)
}

// CBOConfigurationLoader defines the interface for loading CBO configuration
type CBOConfigurationLoader interface {
	// LoadConfig loads configuration from the specified path
	LoadConfig(path string) (*CBORequest, error)

	// LoadDefaultConfig loads the default configuration
	LoadDefaultConfig() *CBORequest

	// MergeConfig merges CLI flags with configuration file
	MergeConfig(base *CBORequest, override *CBORequest) *CBORequest
}

// CBOOutputFormatter defines the interface for formatting CBO analysis results
type CBOOutputFormatter interface {
	// Format formats the analysis response according to the specified format
	Format(response *CBOResponse, format OutputFormat) (string, error)

	// Write writes the formatted output to the writer
	Write(response *CBOResponse, format OutputFormat, writer io.Writer) error
}

// CBOAnalysisOptions provides configuration for CBO analysis behavior
type CBOAnalysisOptions struct {
	// Include system and built-in dependencies
	IncludeBuiltins bool

	// Maximum depth for dependency resolution
	MaxDependencyDepth int

	// Exclude patterns for class names
	ExcludeClassPatterns []string

	// Only analyze public classes (exclude private classes starting with _)
	PublicClassesOnly bool
}

// DefaultCBORequest returns a CBORequest with default values
func DefaultCBORequest() *CBORequest {
	return &CBORequest{
		OutputFormat:    OutputFormatText,
		ShowDetails:     false,
		MinCBO:          0,
		MaxCBO:          0,              // No limit
		SortBy:          SortByCoupling, // Sort by CBO value
		ShowZeros:       BoolPtr(false),
		LowThreshold:    3, // Industry standard: CBO <= 3 is low risk
		MediumThreshold: 7, // Industry standard: 3 < CBO <= 7 is medium risk
		Recursive:       BoolPtr(true),
		IncludeBuiltins: BoolPtr(false),
		IncludeImports:  BoolPtr(true),
		IncludePatterns: []string{"**/*.py"},
		ExcludePatterns: []string{},
	}
}
