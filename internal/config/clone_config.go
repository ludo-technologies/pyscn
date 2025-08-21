package config

import (
	"fmt"
	"io"

	"github.com/pyqol/pyqol/internal/constants"
)

// CloneConfig represents the unified clone detection configuration
// This replaces the duplicated configurations across the codebase
type CloneConfig struct {
	// Analysis Configuration
	Analysis CloneAnalysisConfig `mapstructure:"analysis" yaml:"analysis" json:"analysis"`

	// Thresholds Configuration
	Thresholds ThresholdConfig `mapstructure:"thresholds" yaml:"thresholds" json:"thresholds"`

	// Filtering Configuration
	Filtering FilteringConfig `mapstructure:"filtering" yaml:"filtering" json:"filtering"`

	// Input Configuration
	Input InputConfig `mapstructure:"input" yaml:"input" json:"input"`

	// Output Configuration
	Output CloneOutputConfig `mapstructure:"output" yaml:"output" json:"output"`

	// Performance Configuration
	Performance PerformanceConfig `mapstructure:"performance" yaml:"performance" json:"performance"`
}

// CloneAnalysisConfig holds core analysis parameters
type CloneAnalysisConfig struct {
	// Minimum requirements for clone candidates
	MinLines int `mapstructure:"min_lines" yaml:"min_lines" json:"min_lines"`
	MinNodes int `mapstructure:"min_nodes" yaml:"min_nodes" json:"min_nodes"`

	// Edit distance configuration
	MaxEditDistance float64 `mapstructure:"max_edit_distance" yaml:"max_edit_distance" json:"max_edit_distance"`

	// Normalization options
	IgnoreLiterals    bool `mapstructure:"ignore_literals" yaml:"ignore_literals" json:"ignore_literals"`
	IgnoreIdentifiers bool `mapstructure:"ignore_identifiers" yaml:"ignore_identifiers" json:"ignore_identifiers"`

	// Cost model configuration
	CostModelType string `mapstructure:"cost_model_type" yaml:"cost_model_type" json:"cost_model_type"`
}

// ThresholdConfig holds similarity thresholds for different clone types
type ThresholdConfig struct {
	// Type-specific thresholds (these determine clone classification)
	Type1Threshold float64 `mapstructure:"type1_threshold" yaml:"type1_threshold" json:"type1_threshold"`
	Type2Threshold float64 `mapstructure:"type2_threshold" yaml:"type2_threshold" json:"type2_threshold"`
	Type3Threshold float64 `mapstructure:"type3_threshold" yaml:"type3_threshold" json:"type3_threshold"`
	Type4Threshold float64 `mapstructure:"type4_threshold" yaml:"type4_threshold" json:"type4_threshold"`

	// General similarity threshold (minimum for any clone to be reported)
	SimilarityThreshold float64 `mapstructure:"similarity_threshold" yaml:"similarity_threshold" json:"similarity_threshold"`
}

// FilteringConfig holds filtering and selection criteria
type FilteringConfig struct {
	// Similarity range filtering
	MinSimilarity float64 `mapstructure:"min_similarity" yaml:"min_similarity" json:"min_similarity"`
	MaxSimilarity float64 `mapstructure:"max_similarity" yaml:"max_similarity" json:"max_similarity"`

	// Clone type filtering
	EnabledCloneTypes []string `mapstructure:"enabled_clone_types" yaml:"enabled_clone_types" json:"enabled_clone_types"`

	// Result limiting
	MaxResults int `mapstructure:"max_results" yaml:"max_results" json:"max_results"`
}

// InputConfig holds input processing configuration
type InputConfig struct {
	// File selection
	Paths           []string `mapstructure:"paths" yaml:"paths" json:"paths"`
	Recursive       bool     `mapstructure:"recursive" yaml:"recursive" json:"recursive"`
	IncludePatterns []string `mapstructure:"include_patterns" yaml:"include_patterns" json:"include_patterns"`
	ExcludePatterns []string `mapstructure:"exclude_patterns" yaml:"exclude_patterns" json:"exclude_patterns"`
}

// CloneOutputConfig holds output formatting configuration
// (This extends the existing OutputConfig with clone-specific fields)
type CloneOutputConfig struct {
	// Format and display
	Format      string `mapstructure:"format" yaml:"format" json:"format"`
	ShowDetails bool   `mapstructure:"show_details" yaml:"show_details" json:"show_details"`
	ShowContent bool   `mapstructure:"show_content" yaml:"show_content" json:"show_content"`

	// Sorting and grouping
	SortBy      string `mapstructure:"sort_by" yaml:"sort_by" json:"sort_by"`
	GroupClones bool   `mapstructure:"group_clones" yaml:"group_clones" json:"group_clones"`

	// Output destination (not serialized)
	Writer io.Writer `json:"-" yaml:"-" mapstructure:"-"`
}

// PerformanceConfig holds performance-related settings
type PerformanceConfig struct {
	// Memory management
	MaxMemoryMB    int  `mapstructure:"max_memory_mb" yaml:"max_memory_mb" json:"max_memory_mb"`
	BatchSize      int  `mapstructure:"batch_size" yaml:"batch_size" json:"batch_size"`
	EnableBatching bool `mapstructure:"enable_batching" yaml:"enable_batching" json:"enable_batching"`

	// Parallelization
	MaxGoroutines int `mapstructure:"max_goroutines" yaml:"max_goroutines" json:"max_goroutines"`

	// Early termination
	TimeoutSeconds int `mapstructure:"timeout_seconds" yaml:"timeout_seconds" json:"timeout_seconds"`
}

// DefaultCloneConfig returns a configuration with sensible defaults
func DefaultCloneConfig() *CloneConfig {
	return &CloneConfig{
		Analysis: CloneAnalysisConfig{
			MinLines:          5,
			MinNodes:          10,
			MaxEditDistance:   50.0,
			IgnoreLiterals:    false,
			IgnoreIdentifiers: false,
			CostModelType:     "python",
		},
		Thresholds: ThresholdConfig{
			Type1Threshold:      constants.DefaultType1CloneThreshold,
			Type2Threshold:      constants.DefaultType2CloneThreshold,
			Type3Threshold:      constants.DefaultType3CloneThreshold,
			Type4Threshold:      constants.DefaultType4CloneThreshold,
			SimilarityThreshold: 0.8, // General threshold for clone reporting
		},
		Filtering: FilteringConfig{
			MinSimilarity:     0.0,
			MaxSimilarity:     1.0,
			EnabledCloneTypes: []string{"type1", "type2", "type3", "type4"},
			MaxResults:        10000,
		},
		Input: InputConfig{
			Paths:           []string{"."},
			Recursive:       true,
			IncludePatterns: []string{"*.py"},
			ExcludePatterns: []string{"*test*.py", "*_test.py", "test_*.py"},
		},
		Output: CloneOutputConfig{
			Format:      "text",
			ShowDetails: false,
			ShowContent: false,
			SortBy:      "similarity",
			GroupClones: true,
		},
		Performance: PerformanceConfig{
			MaxMemoryMB:    100,
			BatchSize:      1000,
			EnableBatching: true,
			MaxGoroutines:  4,
			TimeoutSeconds: 300, // 5 minutes
		},
	}
}

// Validate checks if the configuration is valid
func (c *CloneConfig) Validate() error {
	// Validate analysis config
	if err := c.Analysis.Validate(); err != nil {
		return fmt.Errorf("analysis config invalid: %w", err)
	}

	// Validate thresholds
	if err := c.Thresholds.Validate(); err != nil {
		return fmt.Errorf("thresholds config invalid: %w", err)
	}

	// Validate filtering config
	if err := c.Filtering.Validate(); err != nil {
		return fmt.Errorf("filtering config invalid: %w", err)
	}

	// Validate input config
	if err := c.Input.Validate(); err != nil {
		return fmt.Errorf("input config invalid: %w", err)
	}

	// Validate output config
	if err := c.Output.Validate(); err != nil {
		return fmt.Errorf("output config invalid: %w", err)
	}

	// Validate performance config
	if err := c.Performance.Validate(); err != nil {
		return fmt.Errorf("performance config invalid: %w", err)
	}

	return nil
}

// Validate validates the analysis configuration
func (a *CloneAnalysisConfig) Validate() error {
	if a.MinLines < 1 {
		return fmt.Errorf("min_lines must be >= 1, got %d", a.MinLines)
	}
	if a.MinNodes < 1 {
		return fmt.Errorf("min_nodes must be >= 1, got %d", a.MinNodes)
	}
	if a.MaxEditDistance < 0 {
		return fmt.Errorf("max_edit_distance must be >= 0, got %f", a.MaxEditDistance)
	}

	validCostModels := []string{"default", "python", "weighted"}
	valid := false
	for _, model := range validCostModels {
		if a.CostModelType == model {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("cost_model_type must be one of %v, got %s", validCostModels, a.CostModelType)
	}

	return nil
}

// Validate validates the threshold configuration
func (t *ThresholdConfig) Validate() error {
	// Use the threshold validation from constants
	thresholdConfig := constants.CloneThresholdConfig{
		Type1Threshold: t.Type1Threshold,
		Type2Threshold: t.Type2Threshold,
		Type3Threshold: t.Type3Threshold,
		Type4Threshold: t.Type4Threshold,
	}

	if err := thresholdConfig.ValidateThresholds(); err != nil {
		return err
	}

	// Validate general similarity threshold
	if t.SimilarityThreshold < 0.0 || t.SimilarityThreshold > 1.0 {
		return fmt.Errorf("similarity_threshold must be between 0.0 and 1.0, got %f", t.SimilarityThreshold)
	}

	return nil
}

// Validate validates the filtering configuration
func (f *FilteringConfig) Validate() error {
	if f.MinSimilarity < 0.0 || f.MinSimilarity > 1.0 {
		return fmt.Errorf("min_similarity must be between 0.0 and 1.0, got %f", f.MinSimilarity)
	}
	if f.MaxSimilarity < 0.0 || f.MaxSimilarity > 1.0 {
		return fmt.Errorf("max_similarity must be between 0.0 and 1.0, got %f", f.MaxSimilarity)
	}
	if f.MinSimilarity > f.MaxSimilarity {
		return fmt.Errorf("min_similarity (%f) must be <= max_similarity (%f)", f.MinSimilarity, f.MaxSimilarity)
	}
	if f.MaxResults < 0 {
		return fmt.Errorf("max_results must be >= 0, got %d", f.MaxResults)
	}

	// Validate clone types
	validTypes := []string{"type1", "type2", "type3", "type4"}
	for _, cloneType := range f.EnabledCloneTypes {
		valid := false
		for _, validType := range validTypes {
			if cloneType == validType {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid clone type %s, must be one of %v", cloneType, validTypes)
		}
	}

	return nil
}

// Validate validates the input configuration
func (i *InputConfig) Validate() error {
	if len(i.Paths) == 0 {
		return fmt.Errorf("paths cannot be empty")
	}
	return nil
}

// Validate validates the output configuration
func (o *CloneOutputConfig) Validate() error {
	validFormats := []string{"text", "json", "yaml", "csv"}
	valid := false
	for _, format := range validFormats {
		if o.Format == format {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("format must be one of %v, got %s", validFormats, o.Format)
	}

	validSortBy := []string{"similarity", "size", "location", "type"}
	valid = false
	for _, sort := range validSortBy {
		if o.SortBy == sort {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("sort_by must be one of %v, got %s", validSortBy, o.SortBy)
	}

	return nil
}

// Validate validates the performance configuration
func (p *PerformanceConfig) Validate() error {
	if p.MaxMemoryMB <= 0 {
		return fmt.Errorf("max_memory_mb must be > 0, got %d", p.MaxMemoryMB)
	}
	if p.BatchSize <= 0 {
		return fmt.Errorf("batch_size must be > 0, got %d", p.BatchSize)
	}
	if p.MaxGoroutines <= 0 {
		return fmt.Errorf("max_goroutines must be > 0, got %d", p.MaxGoroutines)
	}
	if p.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout_seconds must be > 0, got %d", p.TimeoutSeconds)
	}

	return nil
}
