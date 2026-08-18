package domain

import (
	"context"
	"io"
)

// LCOMRequest represents a request for LCOM (Lack of Cohesion of Methods) analysis
type LCOMRequest struct {
	// Input files or directories to analyze
	Paths []string `json:"paths" yaml:"paths"`

	// Output configuration
	OutputFormat OutputFormat `json:"output_format" yaml:"output_format"`
	OutputWriter io.Writer    `json:"-" yaml:"-"`
	OutputPath   string       `json:"output_path" yaml:"output_path"`   // Path to save output file (for HTML format)
	NoOpen       bool         `json:"no_open" yaml:"no_open"`           // Don't auto-open HTML in browser
	ShowDetails  *bool        `json:"show_details" yaml:"show_details"` // nil = unset, non-nil = explicitly set

	// Filtering and sorting
	MinLCOM int          `json:"min_lcom" yaml:"min_lcom"`
	MaxLCOM int          `json:"max_lcom" yaml:"max_lcom"` // 0 means no limit
	SortBy  SortCriteria `json:"sort_by" yaml:"sort_by"`

	// LCOM thresholds for risk assessment
	LowThreshold    int `json:"low_threshold" yaml:"low_threshold"`       // Default: 2 (LCOM4 <= 2 is low risk)
	MediumThreshold int `json:"medium_threshold" yaml:"medium_threshold"` // Default: 5 (LCOM4 3-5 is medium risk)

	// Configuration
	ConfigPath string `json:"config_path" yaml:"config_path"`

	// Analysis options
	Recursive       *bool    `json:"recursive" yaml:"recursive"`
	IncludePatterns []string `json:"include_patterns" yaml:"include_patterns"`
	ExcludePatterns []string `json:"exclude_patterns" yaml:"exclude_patterns"`
}

// LCOMMetrics represents detailed LCOM metrics for a class
type LCOMMetrics struct {
	// Core LCOM4 metric - number of connected components in method-variable graph
	LCOM4 int `json:"lcom4" yaml:"lcom4"`

	// Method statistics
	TotalMethods    int `json:"total_methods" yaml:"total_methods"`       // All methods in the class
	ExcludedMethods int `json:"excluded_methods" yaml:"excluded_methods"` // Methods kept out of the graph (@classmethod, @staticmethod, @abstractmethod, constructors)

	// Instance variable statistics
	InstanceVariables int `json:"instance_variables" yaml:"instance_variables"` // Distinct self.xxx variables accessed

	// Connected component details
	MethodGroups [][]string `json:"method_groups" yaml:"method_groups"` // Method names grouped by connected component
}

// ClassCohesion represents LCOM analysis result for a single class
type ClassCohesion struct {
	// Class identification
	Name      string `json:"name" yaml:"name"`
	FilePath  string `json:"file_path" yaml:"file_path"`
	StartLine int    `json:"start_line" yaml:"start_line"`
	EndLine   int    `json:"end_line" yaml:"end_line"`

	// LCOM metrics
	Metrics LCOMMetrics `json:"metrics" yaml:"metrics"`

	// Risk assessment
	RiskLevel RiskLevel `json:"risk_level" yaml:"risk_level"`
}

// LCOMSummary represents aggregate LCOM statistics
type LCOMSummary struct {
	TotalClasses    int     `json:"total_classes" yaml:"total_classes"`
	AverageLCOM     float64 `json:"average_lcom" yaml:"average_lcom"`
	MaxLCOM         int     `json:"max_lcom" yaml:"max_lcom"`
	MinLCOM         int     `json:"min_lcom" yaml:"min_lcom"`
	ClassesAnalyzed int     `json:"classes_analyzed" yaml:"classes_analyzed"`
	FilesAnalyzed   int     `json:"files_analyzed" yaml:"files_analyzed"`

	// Risk distribution
	LowRiskClasses    int `json:"low_risk_classes" yaml:"low_risk_classes"`
	MediumRiskClasses int `json:"medium_risk_classes" yaml:"medium_risk_classes"`
	HighRiskClasses   int `json:"high_risk_classes" yaml:"high_risk_classes"`

	// LCOM distribution
	LCOMDistribution map[string]int `json:"lcom_distribution" yaml:"lcom_distribution"`

	// Least cohesive classes (top 10)
	LeastCohesiveClasses []ClassCohesion `json:"least_cohesive_classes" yaml:"least_cohesive_classes"`
}

// LCOMResponse represents the complete LCOM analysis result
type LCOMResponse struct {
	// Analysis results
	Classes []ClassCohesion `json:"classes" yaml:"classes"`
	Summary LCOMSummary     `json:"summary" yaml:"summary"`

	// Warnings and issues
	Warnings []string `json:"warnings" yaml:"warnings"`
	Errors   []string `json:"errors" yaml:"errors"`

	// Metadata
	GeneratedAt string      `json:"generated_at" yaml:"generated_at"`
	Version     string      `json:"version" yaml:"version"`
	Config      interface{} `json:"config" yaml:"config"` // Configuration used for analysis
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

	// LoadDefaultConfig discovers configuration from targetPath (the analyzed
	// path) and falls back to built-in defaults when none is found
	LoadDefaultConfig(targetPath string) *LCOMRequest

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
		ShowDetails:     BoolPtr(false),
		MinLCOM:         0,
		MaxLCOM:         0,                          // No limit
		SortBy:          SortByCohesion,             // Sort by LCOM value
		LowThreshold:    DefaultLCOMLowThreshold,    // LCOM4 <= 2 is low risk
		MediumThreshold: DefaultLCOMMediumThreshold, // LCOM4 3-5 is medium risk
		Recursive:       BoolPtr(true),
		IncludePatterns: DefaultAnalysisIncludePatterns(),
		ExcludePatterns: []string{},
	}
}
