package domain

import (
	"context"
	"io"
)

// LCOMRequest represents a request for LCOM (Lack of Cohesion of Methods) analysis
type LCOMRequest struct {
	// Input files or directories to analyze
	Paths []string

	// Output configuration
	OutputFormat OutputFormat
	OutputWriter io.Writer
	OutputPath   string // Path to save output file (for HTML format)
	NoOpen       bool   // Don't auto-open HTML in browser
	ShowDetails  bool

	// Filtering and sorting
	MinLCOM int
	MaxLCOM int // 0 means no limit
	SortBy  SortCriteria

	// LCOM thresholds for risk assessment
	LowThreshold    int // Default: 2 (LCOM4 <= 2 is low risk)
	MediumThreshold int // Default: 5 (LCOM4 3-5 is medium risk)

	// Configuration
	ConfigPath string

	// Analysis options
	Recursive       *bool
	IncludePatterns []string
	ExcludePatterns []string
}

// LCOMMetrics represents detailed LCOM metrics for a class
type LCOMMetrics struct {
	// Core LCOM4 metric - number of connected components in method-variable graph
	LCOM4 int

	// Method statistics
	TotalMethods    int // All methods in the class
	ExcludedMethods int // Methods excluded (@classmethod, @staticmethod)

	// Instance variable statistics
	InstanceVariables int // Distinct self.xxx variables accessed

	// Connected component details
	MethodGroups [][]string // Method names grouped by connected component
}

// ClassCohesion represents LCOM analysis result for a single class
type ClassCohesion struct {
	// Class identification
	Name      string
	FilePath  string
	StartLine int
	EndLine   int

	// LCOM metrics
	Metrics LCOMMetrics

	// Risk assessment
	RiskLevel RiskLevel
}

// LCOMSummary represents aggregate LCOM statistics
type LCOMSummary struct {
	TotalClasses    int
	AverageLCOM     float64
	MaxLCOM         int
	MinLCOM         int
	ClassesAnalyzed int
	FilesAnalyzed   int

	// Risk distribution
	LowRiskClasses    int
	MediumRiskClasses int
	HighRiskClasses   int

	// LCOM distribution
	LCOMDistribution map[string]int

	// Least cohesive classes (top 10)
	LeastCohesiveClasses []ClassCohesion
}

// LCOMResponse represents the complete LCOM analysis result
type LCOMResponse struct {
	// Analysis results
	Classes []ClassCohesion
	Summary LCOMSummary

	// Warnings and issues
	Warnings []string
	Errors   []string

	// Metadata
	GeneratedAt string
	Version     string
	Config      interface{} // Configuration used for analysis
}

// LCOMService defines the core business logic for LCOM analysis
type LCOMService interface {
	// Analyze performs LCOM analysis on the given request
	Analyze(ctx context.Context, req LCOMRequest) (*LCOMResponse, error)

	// AnalyzeFile analyzes a single Python file
	AnalyzeFile(ctx context.Context, filePath string, req LCOMRequest) (*LCOMResponse, error)
}

// LCOMConfigurationLoader defines the interface for loading LCOM configuration
type LCOMConfigurationLoader interface {
	// LoadConfig loads configuration from the specified path
	LoadConfig(path string) (*LCOMRequest, error)

	// LoadDefaultConfig loads the default configuration
	LoadDefaultConfig() *LCOMRequest

	// MergeConfig merges CLI flags with configuration file
	MergeConfig(base *LCOMRequest, override *LCOMRequest) *LCOMRequest
}

// LCOMOutputFormatter defines the interface for formatting LCOM analysis results
type LCOMOutputFormatter interface {
	// Format formats the analysis response according to the specified format
	Format(response *LCOMResponse, format OutputFormat) (string, error)

	// Write writes the formatted output to the writer
	Write(response *LCOMResponse, format OutputFormat, writer io.Writer) error
}

// DefaultLCOMRequest returns a LCOMRequest with default values
// Threshold values are sourced from domain/defaults.go
func DefaultLCOMRequest() *LCOMRequest {
	return &LCOMRequest{
		OutputFormat:    OutputFormatText,
		ShowDetails:     false,
		MinLCOM:         0,
		MaxLCOM:         0,                          // No limit
		SortBy:          SortByCohesion,             // Sort by LCOM value
		LowThreshold:    DefaultLCOMLowThreshold,    // LCOM4 <= 2 is low risk
		MediumThreshold: DefaultLCOMMediumThreshold, // LCOM4 3-5 is medium risk
		Recursive:       BoolPtr(true),
		IncludePatterns: []string{"**/*.py"},
		ExcludePatterns: []string{},
	}
}
