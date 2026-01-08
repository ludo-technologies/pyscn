package domain

import (
	"context"
	"io"
)

// MockDataSeverity represents the severity level of mock data findings
type MockDataSeverity string

const (
	MockDataSeverityError   MockDataSeverity = "error"
	MockDataSeverityWarning MockDataSeverity = "warning"
	MockDataSeverityInfo    MockDataSeverity = "info"
)

// MockDataSortCriteria represents the criteria for sorting mock data results
type MockDataSortCriteria string

const (
	MockDataSortBySeverity MockDataSortCriteria = "severity"
	MockDataSortByLine     MockDataSortCriteria = "line"
	MockDataSortByFile     MockDataSortCriteria = "file"
	MockDataSortByType     MockDataSortCriteria = "type"
)

// MockDataType represents the type of mock data detected
type MockDataType string

const (
	MockDataTypeKeyword       MockDataType = "keyword"        // mock, fake, dummy, etc.
	MockDataTypeDomain        MockDataType = "domain"         // example.com, test.com, etc.
	MockDataTypeEmail         MockDataType = "email"          // test@example.com, etc.
	MockDataTypePhone         MockDataType = "phone"          // 000-0000-0000, etc.
	MockDataTypeUUID          MockDataType = "uuid"           // low-entropy UUIDs
	MockDataTypePlaceholder   MockDataType = "placeholder"    // TODO, FIXME, XXX, etc.
	MockDataTypeRepetitive    MockDataType = "repetitive"     // aaaa, 1111, etc.
	MockDataTypeTestCredential MockDataType = "test_credential" // password123, secret, etc.
)

// MockDataRequest represents a request for mock data analysis
type MockDataRequest struct {
	// Input files or directories to analyze
	Paths []string

	// Output configuration
	OutputFormat OutputFormat
	OutputWriter io.Writer
	OutputPath   string // Path to save output file (for HTML format)
	NoOpen       bool   // Don't auto-open HTML in browser

	// Filtering and sorting
	MinSeverity MockDataSeverity
	SortBy      MockDataSortCriteria

	// Analysis options
	Recursive       bool
	IncludePatterns []string
	ExcludePatterns []string
	IgnoreTests     *bool // nil = use default (true), non-nil = explicitly set

	// Configuration
	ConfigPath string

	// Mock data specific options
	Keywords         []string // Keywords to detect (mock, fake, dummy, etc.)
	Domains          []string // Domains to detect (example.com, etc.)
	IgnorePatterns   []string // Patterns in code to ignore
	EnabledTypes     []MockDataType // Types of mock data to detect (empty = all)
}

// MockDataLocation represents the location of detected mock data
type MockDataLocation struct {
	FilePath    string `json:"file_path"`
	StartLine   int    `json:"start_line"`
	EndLine     int    `json:"end_line"`
	StartColumn int    `json:"start_column"`
	EndColumn   int    `json:"end_column"`
}

// MockDataFinding represents a single mock data detection result
type MockDataFinding struct {
	// Location information
	Location MockDataLocation `json:"location"`

	// Mock data details
	Value       string           `json:"value"`       // The detected mock value
	Type        MockDataType     `json:"type"`        // Type of mock data
	Severity    MockDataSeverity `json:"severity"`
	Description string           `json:"description"` // Why this was flagged
	Rationale   string           `json:"rationale"`   // Detection rationale

	// Context information
	Context     string `json:"context,omitempty"`      // Surrounding code
	VariableName string `json:"variable_name,omitempty"` // Variable name if applicable
}

// FileMockData represents mock data analysis result for a single file
type FileMockData struct {
	// File identification
	FilePath string `json:"file_path"`

	// Findings
	Findings []MockDataFinding `json:"findings"`

	// File-level summary
	TotalFindings int `json:"total_findings"`
	ErrorCount    int `json:"error_count"`
	WarningCount  int `json:"warning_count"`
	InfoCount     int `json:"info_count"`
}

// MockDataSummary represents aggregate statistics for mock data analysis
type MockDataSummary struct {
	// Overall metrics
	TotalFiles         int `json:"total_files"`
	TotalFindings      int `json:"total_findings"`
	FilesWithMockData  int `json:"files_with_mock_data"`

	// Severity distribution
	ErrorFindings   int `json:"error_findings"`
	WarningFindings int `json:"warning_findings"`
	InfoFindings    int `json:"info_findings"`

	// Type distribution
	FindingsByType map[MockDataType]int `json:"findings_by_type"`
}

// MockDataResponse represents the complete mock data analysis result
type MockDataResponse struct {
	// Analysis results
	Files   []FileMockData  `json:"files"`
	Summary MockDataSummary `json:"summary"`

	// Warnings and issues
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`

	// Metadata
	GeneratedAt string      `json:"generated_at"`
	Version     string      `json:"version"`
	Config      interface{} `json:"config"` // Configuration used for analysis
}

// MockDataService defines the core business logic for mock data analysis
type MockDataService interface {
	// Analyze performs mock data analysis on the given request
	Analyze(ctx context.Context, req MockDataRequest) (*MockDataResponse, error)

	// AnalyzeFile analyzes a single Python file for mock data
	AnalyzeFile(ctx context.Context, filePath string, req MockDataRequest) (*FileMockData, error)
}

// MockDataConfigurationLoader defines the interface for loading mock data configuration
type MockDataConfigurationLoader interface {
	// LoadConfig loads mock data configuration from the specified path
	LoadConfig(path string) (*MockDataRequest, error)

	// LoadDefaultConfig loads the default mock data configuration
	LoadDefaultConfig() *MockDataRequest

	// MergeConfig merges CLI flags with configuration file
	MergeConfig(base *MockDataRequest, override *MockDataRequest) *MockDataRequest
}

// MockDataFormatter defines the interface for formatting mock data analysis results
type MockDataFormatter interface {
	// Format formats the mock data analysis response according to the specified format
	Format(response *MockDataResponse, format OutputFormat) (string, error)

	// Write writes the formatted mock data output to the writer
	Write(response *MockDataResponse, format OutputFormat, writer io.Writer) error
}

// Default configuration values for mock data analysis
func DefaultMockDataRequest() *MockDataRequest {
	return &MockDataRequest{
		OutputFormat:    OutputFormatText,
		MinSeverity:     MockDataSeverityWarning,
		SortBy:          MockDataSortBySeverity,
		Recursive:       true,
		IncludePatterns: []string{"**/*.py"},
		ExcludePatterns: []string{},
		IgnoreTests:     BoolPtr(true),
		Keywords:        DefaultMockDataKeywords(),
		Domains:         DefaultMockDataDomains(),
		IgnorePatterns:  []string{},
		EnabledTypes:    []MockDataType{}, // Empty means all types enabled
	}
}

// Validation methods

// Validate validates the mock data request
func (req *MockDataRequest) Validate() error {
	if len(req.Paths) == 0 {
		return NewInvalidInputError("at least one path must be specified", nil)
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
	validSeverities := map[MockDataSeverity]bool{
		MockDataSeverityError:   true,
		MockDataSeverityWarning: true,
		MockDataSeverityInfo:    true,
	}

	if !validSeverities[req.MinSeverity] {
		return NewInvalidInputError("invalid minimum severity level", nil)
	}

	// Validate sort criteria
	validSortBy := map[MockDataSortCriteria]bool{
		MockDataSortBySeverity: true,
		MockDataSortByLine:     true,
		MockDataSortByFile:     true,
		MockDataSortByType:     true,
	}

	if !validSortBy[req.SortBy] {
		return NewInvalidInputError("invalid sort criteria", nil)
	}

	return nil
}

// Helper methods for severity comparison

// Level returns the numeric level for comparison
func (s MockDataSeverity) Level() int {
	switch s {
	case MockDataSeverityInfo:
		return 1
	case MockDataSeverityWarning:
		return 2
	case MockDataSeverityError:
		return 3
	default:
		return 0
	}
}

// IsAtLeast checks if the severity is at least the specified level
func (s MockDataSeverity) IsAtLeast(minSeverity MockDataSeverity) bool {
	return s.Level() >= minSeverity.Level()
}

// Summary calculation helpers

// CalculateSeverityCounts calculates the count of findings by severity
func (fmd *FileMockData) CalculateSeverityCounts() {
	fmd.ErrorCount = 0
	fmd.WarningCount = 0
	fmd.InfoCount = 0

	for _, finding := range fmd.Findings {
		switch finding.Severity {
		case MockDataSeverityError:
			fmd.ErrorCount++
		case MockDataSeverityWarning:
			fmd.WarningCount++
		case MockDataSeverityInfo:
			fmd.InfoCount++
		}
	}
	fmd.TotalFindings = len(fmd.Findings)
}

// HasFindings returns true if the file has any mock data findings
func (fmd *FileMockData) HasFindings() bool {
	return len(fmd.Findings) > 0
}

// HasFindingsAtSeverity returns true if the file has findings at or above the specified severity
func (fmd *FileMockData) HasFindingsAtSeverity(minSeverity MockDataSeverity) bool {
	for _, finding := range fmd.Findings {
		if finding.Severity.IsAtLeast(minSeverity) {
			return true
		}
	}
	return false
}

// GetFindingsAtSeverity returns findings at or above the specified severity level
func (fmd *FileMockData) GetFindingsAtSeverity(minSeverity MockDataSeverity) []MockDataFinding {
	var filtered []MockDataFinding
	for _, finding := range fmd.Findings {
		if finding.Severity.IsAtLeast(minSeverity) {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

// CalculateTypeCounts calculates the count of findings by type
func (s *MockDataSummary) CalculateTypeCounts(files []FileMockData) {
	s.FindingsByType = make(map[MockDataType]int)
	for _, file := range files {
		for _, finding := range file.Findings {
			s.FindingsByType[finding.Type]++
		}
	}
}
