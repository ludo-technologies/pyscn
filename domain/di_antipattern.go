package domain

import (
	"context"
	"io"
)

// DIAntipatternType represents the type of DI anti-pattern detected
type DIAntipatternType string

const (
	// DIAntipatternConstructorOverInjection indicates too many constructor parameters
	DIAntipatternConstructorOverInjection DIAntipatternType = "constructor_over_injection"
	// DIAntipatternHiddenDependency indicates a hidden dependency pattern
	DIAntipatternHiddenDependency DIAntipatternType = "hidden_dependency"
	// DIAntipatternConcreteDependency indicates dependency on concrete class
	DIAntipatternConcreteDependency DIAntipatternType = "concrete_dependency"
	// DIAntipatternServiceLocator indicates service locator anti-pattern
	DIAntipatternServiceLocator DIAntipatternType = "service_locator"
)

// HiddenDependencySubtype represents the subtype of hidden dependency
type HiddenDependencySubtype string

const (
	// HiddenDepGlobal indicates use of global statement
	HiddenDepGlobal HiddenDependencySubtype = "global_statement"
	// HiddenDepModuleVariable indicates direct access to module-level variable
	HiddenDepModuleVariable HiddenDependencySubtype = "module_variable"
	// HiddenDepSingleton indicates singleton pattern via _instance
	HiddenDepSingleton HiddenDependencySubtype = "singleton"
)

// ConcreteDependencySubtype represents the subtype of concrete dependency
type ConcreteDependencySubtype string

const (
	// ConcreteDepTypeHint indicates type hint with concrete class
	ConcreteDepTypeHint ConcreteDependencySubtype = "type_hint"
	// ConcreteDepInstantiation indicates direct instantiation in constructor
	ConcreteDepInstantiation ConcreteDependencySubtype = "instantiation"
)

// DIAntipatternSeverity represents the severity level of a DI anti-pattern
type DIAntipatternSeverity string

const (
	// DIAntipatternSeverityInfo indicates informational severity
	DIAntipatternSeverityInfo DIAntipatternSeverity = "info"
	// DIAntipatternSeverityWarning indicates warning severity
	DIAntipatternSeverityWarning DIAntipatternSeverity = "warning"
	// DIAntipatternSeverityError indicates error severity
	DIAntipatternSeverityError DIAntipatternSeverity = "error"
)

// DIAntipatternFinding represents a single DI anti-pattern detection result
type DIAntipatternFinding struct {
	// Type of the anti-pattern
	Type DIAntipatternType `json:"type"`

	// Subtype for hidden dependency and concrete dependency patterns
	Subtype string `json:"subtype,omitempty"`

	// Severity of the finding
	Severity DIAntipatternSeverity `json:"severity"`

	// Class name where the anti-pattern was found
	ClassName string `json:"class_name,omitempty"`

	// Method name where the anti-pattern was found
	MethodName string `json:"method_name,omitempty"`

	// Location in source code
	Location SourceLocation `json:"location"`

	// Human-readable description of the issue
	Description string `json:"description"`

	// Suggestion for fixing the issue
	Suggestion string `json:"suggestion"`

	// Additional details specific to the anti-pattern type
	Details map[string]interface{} `json:"details,omitempty"`
}

// DIAntipatternRequest represents a request for DI anti-pattern analysis
type DIAntipatternRequest struct {
	// Input files or directories to analyze
	Paths []string

	// Output configuration
	OutputFormat OutputFormat
	OutputWriter io.Writer
	OutputPath   string
	NoOpen       bool

	// Analysis options
	Recursive       *bool
	IncludePatterns []string
	ExcludePatterns []string

	// Configuration
	ConfigPath string

	// DI-specific options
	// ConstructorParamThreshold is the maximum allowed constructor parameters (default: 5)
	ConstructorParamThreshold int

	// MinSeverity filters findings by minimum severity level
	MinSeverity DIAntipatternSeverity

	// SortBy specifies the sort order
	SortBy SortCriteria
}

// DIAntipatternSummary represents aggregate statistics for DI anti-pattern analysis
type DIAntipatternSummary struct {
	// TotalFindings is the total number of findings
	TotalFindings int `json:"total_findings"`

	// ByType breaks down findings by anti-pattern type
	ByType map[DIAntipatternType]int `json:"by_type"`

	// BySeverity breaks down findings by severity
	BySeverity map[DIAntipatternSeverity]int `json:"by_severity"`

	// FilesAnalyzed is the number of files analyzed
	FilesAnalyzed int `json:"files_analyzed"`

	// ClassesAnalyzed is the number of classes analyzed
	ClassesAnalyzed int `json:"classes_analyzed"`

	// AffectedClasses is the number of classes with at least one finding
	AffectedClasses int `json:"affected_classes"`
}

// DIAntipatternResponse represents the complete DI anti-pattern analysis result
type DIAntipatternResponse struct {
	// Findings contains all detected anti-patterns
	Findings []DIAntipatternFinding `json:"findings"`

	// Summary contains aggregate statistics
	Summary DIAntipatternSummary `json:"summary"`

	// Warnings contains non-fatal issues encountered during analysis
	Warnings []string `json:"warnings,omitempty"`

	// Errors contains errors encountered during analysis
	Errors []string `json:"errors,omitempty"`

	// Metadata
	GeneratedAt string      `json:"generated_at"`
	Version     string      `json:"version"`
	Config      interface{} `json:"config,omitempty"`
}

// DIAntipatternService defines the interface for DI anti-pattern analysis
type DIAntipatternService interface {
	// Analyze performs DI anti-pattern analysis on the given request
	Analyze(ctx context.Context, req DIAntipatternRequest) (*DIAntipatternResponse, error)

	// AnalyzeFile analyzes a single Python file
	AnalyzeFile(ctx context.Context, filePath string, req DIAntipatternRequest) (*DIAntipatternResponse, error)
}

// DIAntipatternConfigurationLoader defines the interface for loading DI anti-pattern configuration
type DIAntipatternConfigurationLoader interface {
	// LoadConfig loads configuration from the specified path
	LoadConfig(path string) (*DIAntipatternRequest, error)

	// LoadDefaultConfig loads the default configuration
	LoadDefaultConfig() *DIAntipatternRequest

	// MergeConfig merges CLI flags with configuration file
	MergeConfig(base *DIAntipatternRequest, override *DIAntipatternRequest) *DIAntipatternRequest
}

// DIAntipatternOutputFormatter defines the interface for formatting DI anti-pattern analysis results
type DIAntipatternOutputFormatter interface {
	// Format formats the analysis response according to the specified format
	Format(response *DIAntipatternResponse, format OutputFormat) (string, error)

	// Write writes the formatted output to the writer
	Write(response *DIAntipatternResponse, format OutputFormat, writer io.Writer) error
}

// DefaultDIAntipatternRequest returns a DIAntipatternRequest with default values
func DefaultDIAntipatternRequest() *DIAntipatternRequest {
	return &DIAntipatternRequest{
		OutputFormat:              OutputFormatJSON,
		Recursive:                 BoolPtr(true),
		IncludePatterns:           []string{"**/*.py"},
		ExcludePatterns:           []string{},
		ConstructorParamThreshold: DefaultDIConstructorParamThreshold,
		MinSeverity:               DIAntipatternSeverityWarning,
		SortBy:                    SortBySeverity,
	}
}

// SeverityOrder returns numeric order for severity (higher = more severe)
func (s DIAntipatternSeverity) SeverityOrder() int {
	switch s {
	case DIAntipatternSeverityError:
		return 3
	case DIAntipatternSeverityWarning:
		return 2
	case DIAntipatternSeverityInfo:
		return 1
	default:
		return 0
	}
}
