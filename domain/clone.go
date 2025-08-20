package domain

import (
	"context"
	"fmt"
	"io"

	"github.com/pyqol/pyqol/internal/constants"
)

// CloneType represents different types of code clones
type CloneType int

const (
	// Type1Clone - Identical code fragments (except whitespace and comments)
	Type1Clone CloneType = iota + 1
	// Type2Clone - Syntactically identical but with different identifiers/literals  
	Type2Clone
	// Type3Clone - Syntactically similar with small modifications
	Type3Clone
	// Type4Clone - Functionally similar but syntactically different
	Type4Clone
)

// String returns string representation of CloneType
func (ct CloneType) String() string {
	switch ct {
	case Type1Clone:
		return "Type-1"
	case Type2Clone:
		return "Type-2"
	case Type3Clone:
		return "Type-3"
	case Type4Clone:
		return "Type-4"
	default:
		return "Unknown"
	}
}

// CloneLocation represents a location of a clone in source code
type CloneLocation struct {
	FilePath  string `json:"file_path" yaml:"file_path" csv:"file_path"`
	StartLine int    `json:"start_line" yaml:"start_line" csv:"start_line"`
	EndLine   int    `json:"end_line" yaml:"end_line" csv:"end_line"`
	StartCol  int    `json:"start_col" yaml:"start_col" csv:"start_col"`
	EndCol    int    `json:"end_col" yaml:"end_col" csv:"end_col"`
}

// String returns string representation of CloneLocation
func (cl *CloneLocation) String() string {
	return fmt.Sprintf("%s:%d:%d-%d:%d", cl.FilePath, cl.StartLine, cl.StartCol, cl.EndLine, cl.EndCol)
}

// LineCount returns the number of lines in this location
func (cl *CloneLocation) LineCount() int {
	return cl.EndLine - cl.StartLine + 1
}

// Clone represents a detected code clone
type Clone struct {
	ID         int            `json:"id" yaml:"id" csv:"id"`
	Type       CloneType      `json:"type" yaml:"type" csv:"type"`
	Location   *CloneLocation `json:"location" yaml:"location" csv:"location"`
	Content    string         `json:"content,omitempty" yaml:"content,omitempty" csv:"content"`
	Hash       string         `json:"hash" yaml:"hash" csv:"hash"`
	Size       int            `json:"size" yaml:"size" csv:"size"`           // Number of AST nodes
	LineCount  int            `json:"line_count" yaml:"line_count" csv:"line_count"`
	Complexity int            `json:"complexity" yaml:"complexity" csv:"complexity"`
}

// String returns string representation of Clone
func (c *Clone) String() string {
	return fmt.Sprintf("Clone{ID: %d, Type: %s, Location: %s, Size: %d}",
		c.ID, c.Type.String(), c.Location.String(), c.Size)
}

// ClonePair represents a pair of similar code clones
type ClonePair struct {
	ID         int     `json:"id" yaml:"id" csv:"id"`
	Clone1     *Clone  `json:"clone1" yaml:"clone1" csv:"clone1"`
	Clone2     *Clone  `json:"clone2" yaml:"clone2" csv:"clone2"`
	Similarity float64 `json:"similarity" yaml:"similarity" csv:"similarity"`
	Distance   float64 `json:"distance" yaml:"distance" csv:"distance"`
	Type       CloneType `json:"type" yaml:"type" csv:"type"`
	Confidence float64 `json:"confidence" yaml:"confidence" csv:"confidence"`
}

// String returns string representation of ClonePair
func (cp *ClonePair) String() string {
	return fmt.Sprintf("%s clone: %s <-> %s (similarity: %.3f)",
		cp.Type.String(),
		cp.Clone1.Location.String(),
		cp.Clone2.Location.String(),
		cp.Similarity)
}

// CloneGroup represents a group of related clones
type CloneGroup struct {
	ID         int      `json:"id" yaml:"id" csv:"id"`
	Clones     []*Clone `json:"clones" yaml:"clones" csv:"clones"`
	Type       CloneType `json:"type" yaml:"type" csv:"type"`
	Similarity float64  `json:"similarity" yaml:"similarity" csv:"similarity"`
	Size       int      `json:"size" yaml:"size" csv:"size"`
}

// String returns string representation of CloneGroup
func (cg *CloneGroup) String() string {
	return fmt.Sprintf("CloneGroup{ID: %d, Type: %s, Size: %d, Similarity: %.3f}",
		cg.ID, cg.Type.String(), cg.Size, cg.Similarity)
}

// AddClone adds a clone to the group
func (cg *CloneGroup) AddClone(clone *Clone) {
	cg.Clones = append(cg.Clones, clone)
	cg.Size = len(cg.Clones)
}

// CloneStatistics provides statistics about clone detection results
type CloneStatistics struct {
	TotalClones      int            `json:"total_clones" yaml:"total_clones" csv:"total_clones"`
	TotalClonePairs  int            `json:"total_clone_pairs" yaml:"total_clone_pairs" csv:"total_clone_pairs"`
	TotalCloneGroups int            `json:"total_clone_groups" yaml:"total_clone_groups" csv:"total_clone_groups"`
	ClonesByType     map[string]int `json:"clones_by_type" yaml:"clones_by_type" csv:"clones_by_type"`
	AverageSimilarity float64       `json:"average_similarity" yaml:"average_similarity" csv:"average_similarity"`
	LinesAnalyzed    int            `json:"lines_analyzed" yaml:"lines_analyzed" csv:"lines_analyzed"`
	FilesAnalyzed    int            `json:"files_analyzed" yaml:"files_analyzed" csv:"files_analyzed"`
}

// CloneRequest represents a request for clone detection
type CloneRequest struct {
	// Input parameters
	Paths               []string `json:"paths"`
	Recursive           bool     `json:"recursive"`
	IncludePatterns     []string `json:"include_patterns"`
	ExcludePatterns     []string `json:"exclude_patterns"`

	// Analysis configuration
	MinLines            int     `json:"min_lines"`
	MinNodes            int     `json:"min_nodes"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
	MaxEditDistance     float64 `json:"max_edit_distance"`
	IgnoreLiterals      bool    `json:"ignore_literals"`
	IgnoreIdentifiers   bool    `json:"ignore_identifiers"`

	// Type-specific thresholds
	Type1Threshold      float64 `json:"type1_threshold"`
	Type2Threshold      float64 `json:"type2_threshold"`
	Type3Threshold      float64 `json:"type3_threshold"`
	Type4Threshold      float64 `json:"type4_threshold"`

	// Output configuration
	OutputFormat        OutputFormat `json:"output_format"`
	OutputWriter        io.Writer    `json:"-"`
	ShowDetails         bool         `json:"show_details"`
	ShowContent         bool         `json:"show_content"`
	SortBy              SortCriteria `json:"sort_by"`
	GroupClones         bool         `json:"group_clones"`
	
	// Filtering
	MinSimilarity       float64   `json:"min_similarity"`
	MaxSimilarity       float64   `json:"max_similarity"`
	CloneTypes          []CloneType `json:"clone_types"`

	// Configuration file
	ConfigPath          string `json:"config_path"`
}

// CloneResponse represents the response from clone detection
type CloneResponse struct {
	// Results
	Clones      []*Clone       `json:"clones" yaml:"clones" csv:"clones"`
	ClonePairs  []*ClonePair   `json:"clone_pairs" yaml:"clone_pairs" csv:"clone_pairs"`
	CloneGroups []*CloneGroup  `json:"clone_groups" yaml:"clone_groups" csv:"clone_groups"`
	Statistics  *CloneStatistics `json:"statistics" yaml:"statistics" csv:"statistics"`
	
	// Metadata
	Request     *CloneRequest `json:"request,omitempty" yaml:"request,omitempty" csv:"-"`
	Duration    int64         `json:"duration_ms" yaml:"duration_ms" csv:"duration_ms"`
	Success     bool          `json:"success" yaml:"success" csv:"success"`
	Error       string        `json:"error,omitempty" yaml:"error,omitempty" csv:"error"`
}

// CloneSortCriteria defines how to sort clone results
type CloneSortCriteria string

const (
	SortClonesByLocation    CloneSortCriteria = "location"
	SortClonesBySimilarity  CloneSortCriteria = "similarity"
	SortClonesBySize        CloneSortCriteria = "size"
	SortClonesByType        CloneSortCriteria = "type"
	SortClonesByConfidence  CloneSortCriteria = "confidence"
)

// CloneService defines the interface for clone detection services
type CloneService interface {
	// DetectClones performs clone detection on the given request
	DetectClones(ctx context.Context, req *CloneRequest) (*CloneResponse, error)
	
	// DetectClonesInFiles performs clone detection on specific files
	DetectClonesInFiles(ctx context.Context, filePaths []string, req *CloneRequest) (*CloneResponse, error)
	
	// ComputeSimilarity computes similarity between two code fragments
	ComputeSimilarity(ctx context.Context, fragment1, fragment2 string) (float64, error)
}

// CloneOutputFormatter defines the interface for formatting clone detection results
type CloneOutputFormatter interface {
	// FormatCloneResponse formats a clone response according to the specified format
	FormatCloneResponse(response *CloneResponse, format OutputFormat, writer io.Writer) error
	
	// FormatCloneStatistics formats clone statistics
	FormatCloneStatistics(stats *CloneStatistics, format OutputFormat, writer io.Writer) error
}

// CloneConfigurationLoader defines the interface for loading clone detection configuration
type CloneConfigurationLoader interface {
	// LoadCloneConfig loads clone detection configuration from file
	LoadCloneConfig(configPath string) (*CloneRequest, error)
	
	// SaveCloneConfig saves clone detection configuration to file
	SaveCloneConfig(config *CloneRequest, configPath string) error
	
	// GetDefaultCloneConfig returns default clone detection configuration
	GetDefaultCloneConfig() *CloneRequest
}

// Validation methods

// Validate validates a clone request
func (req *CloneRequest) Validate() error {
	if len(req.Paths) == 0 {
		return NewValidationError("paths cannot be empty")
	}
	
	if req.MinLines < 1 {
		return NewValidationError("min_lines must be >= 1")
	}
	
	if req.MinNodes < 1 {
		return NewValidationError("min_nodes must be >= 1")
	}
	
	if req.SimilarityThreshold < 0.0 || req.SimilarityThreshold > 1.0 {
		return NewValidationError("similarity_threshold must be between 0.0 and 1.0")
	}
	
	if req.MaxEditDistance < 0.0 {
		return NewValidationError("max_edit_distance must be >= 0.0")
	}
	
	// Validate type-specific thresholds
	if req.Type1Threshold < 0.0 || req.Type1Threshold > 1.0 {
		return NewValidationError("type1_threshold must be between 0.0 and 1.0")
	}
	
	if req.Type2Threshold < 0.0 || req.Type2Threshold > 1.0 {
		return NewValidationError("type2_threshold must be between 0.0 and 1.0")
	}
	
	if req.Type3Threshold < 0.0 || req.Type3Threshold > 1.0 {
		return NewValidationError("type3_threshold must be between 0.0 and 1.0")
	}
	
	if req.Type4Threshold < 0.0 || req.Type4Threshold > 1.0 {
		return NewValidationError("type4_threshold must be between 0.0 and 1.0")
	}
	
	// Validate threshold ordering (Type1 > Type2 > Type3 > Type4)
	if req.Type1Threshold <= req.Type2Threshold {
		return NewValidationError("type1_threshold should be > type2_threshold")
	}
	
	if req.Type2Threshold <= req.Type3Threshold {
		return NewValidationError("type2_threshold should be > type3_threshold")
	}
	
	if req.Type3Threshold <= req.Type4Threshold {
		return NewValidationError("type3_threshold should be > type4_threshold")
	}
	
	return nil
}

// HasValidOutputWriter checks if the request has a valid output writer
func (req *CloneRequest) HasValidOutputWriter() bool {
	return req.OutputWriter != nil
}

// ShouldShowContent determines if content should be included in output
func (req *CloneRequest) ShouldShowContent() bool {
	return req.ShowContent
}

// ShouldGroupClones determines if clones should be grouped
func (req *CloneRequest) ShouldGroupClones() bool {
	return req.GroupClones
}

// DefaultCloneRequest returns a default clone request
func DefaultCloneRequest() *CloneRequest {
	return &CloneRequest{
		Paths:               []string{"."},
		Recursive:           true,
		IncludePatterns:     []string{"*.py"},
		ExcludePatterns:     []string{"*test*.py", "*_test.py", "test_*.py"},
		MinLines:            5,
		MinNodes:            10,
		SimilarityThreshold: 0.8,
		MaxEditDistance:     50.0,
		IgnoreLiterals:      false,
		IgnoreIdentifiers:   false,
		Type1Threshold:      constants.DefaultType1CloneThreshold,
		Type2Threshold:      constants.DefaultType2CloneThreshold,
		Type3Threshold:      constants.DefaultType3CloneThreshold,
		Type4Threshold:      constants.DefaultType4CloneThreshold,
		OutputFormat:        OutputFormatText,
		ShowDetails:         false,
		ShowContent:         false,
		SortBy:              SortBySimilarity,
		GroupClones:         true,
		MinSimilarity:       0.0,
		MaxSimilarity:       1.0,
		CloneTypes:          []CloneType{Type1Clone, Type2Clone, Type3Clone, Type4Clone},
	}
}

// NewCloneStatistics creates a new clone statistics instance
func NewCloneStatistics() *CloneStatistics {
	return &CloneStatistics{
		ClonesByType: make(map[string]int),
	}
}