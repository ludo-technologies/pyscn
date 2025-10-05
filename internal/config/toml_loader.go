package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// PyscnTomlConfig represents the structure of .pyscn.toml
type PyscnTomlConfig struct {
	Clones      ClonesConfig             `toml:"clones"`      // Primary: [clones] section
	Analysis    PyscnTomlAnalysisConfig  `toml:"analysis"`    // Fallback for compatibility
	Thresholds  ThresholdConfig          `toml:"thresholds"`  // Fallback for compatibility
	Filtering   PyscnTomlFilteringConfig `toml:"filtering"`   // Fallback for compatibility
	Input       PyscnTomlInputConfig     `toml:"input"`       // Fallback for compatibility
	Output      PyscnTomlOutputConfig    `toml:"output"`      // Fallback for compatibility
	Performance PerformanceConfig        `toml:"performance"` // Fallback for compatibility
	Grouping    GroupingConfig           `toml:"grouping"`    // Fallback for compatibility
	LSH         LSHConfig                `toml:"lsh"`         // Fallback for compatibility
}

// ClonesConfig represents the [clones] section (flat structure)
type ClonesConfig struct {
	// Analysis settings
	MinLines          int     `toml:"min_lines"`
	MinNodes          int     `toml:"min_nodes"`
	MaxEditDistance   float64 `toml:"max_edit_distance"`
	IgnoreLiterals    *bool   `toml:"ignore_literals"`    // pointer to detect unset
	IgnoreIdentifiers *bool   `toml:"ignore_identifiers"` // pointer to detect unset
	CostModelType     string  `toml:"cost_model_type"`

	// Thresholds
	Type1Threshold      float64 `toml:"type1_threshold"`
	Type2Threshold      float64 `toml:"type2_threshold"`
	Type3Threshold      float64 `toml:"type3_threshold"`
	Type4Threshold      float64 `toml:"type4_threshold"`
	SimilarityThreshold float64 `toml:"similarity_threshold"`

	// Filtering
	MinSimilarity     float64  `toml:"min_similarity"`
	MaxSimilarity     float64  `toml:"max_similarity"`
	EnabledCloneTypes []string `toml:"enabled_clone_types"`
	MaxResults        int      `toml:"max_results"`

	// Grouping
	GroupingMode      string  `toml:"grouping_mode"`
	GroupingThreshold float64 `toml:"grouping_threshold"`
	KCoreK            int     `toml:"k_core_k"`

	// LSH (flat structure with lsh_ prefix)
	LSHEnabled             string  `toml:"lsh_enabled"`
	LSHAutoThreshold       int     `toml:"lsh_auto_threshold"`
	LSHSimilarityThreshold float64 `toml:"lsh_similarity_threshold"`
	LSHBands               int     `toml:"lsh_bands"`
	LSHRows                int     `toml:"lsh_rows"`
	LSHHashes              int     `toml:"lsh_hashes"`

	// Performance
	MaxMemoryMB    int  `toml:"max_memory_mb"`
	BatchSize      int  `toml:"batch_size"`
	EnableBatching *bool `toml:"enable_batching"` // pointer to detect unset
	MaxGoroutines  int  `toml:"max_goroutines"`
	TimeoutSeconds int  `toml:"timeout_seconds"`

	// Output
	ShowDetails *bool  `toml:"show_details"` // pointer to detect unset
	ShowContent *bool  `toml:"show_content"` // pointer to detect unset
	SortBy      string `toml:"sort_by"`
	GroupClones *bool  `toml:"group_clones"` // pointer to detect unset
}

type PyscnTomlAnalysisConfig struct {
	MinLines          int     `toml:"min_lines"`
	MinNodes          int     `toml:"min_nodes"`
	MaxEditDistance   float64 `toml:"max_edit_distance"`
	IgnoreLiterals    *bool   `toml:"ignore_literals"`    // pointer to detect unset
	IgnoreIdentifiers *bool   `toml:"ignore_identifiers"` // pointer to detect unset
	CostModelType     string  `toml:"cost_model_type"`
}

type PyscnTomlFilteringConfig struct {
	MinSimilarity     float64  `toml:"min_similarity"`
	MaxSimilarity     float64  `toml:"max_similarity"`
	EnabledCloneTypes []string `toml:"enabled_clone_types"`
	MaxResults        int      `toml:"max_results"`
}

type PyscnTomlInputConfig struct {
	Paths           []string `toml:"paths"`
	Recursive       *bool    `toml:"recursive"` // pointer to detect unset
	IncludePatterns []string `toml:"include_patterns"`
	ExcludePatterns []string `toml:"exclude_patterns"`
}

type PyscnTomlOutputConfig struct {
	Format      string `toml:"format"`
	ShowDetails *bool  `toml:"show_details"` // pointer to detect unset
	ShowContent *bool  `toml:"show_content"` // pointer to detect unset
	SortBy      string `toml:"sort_by"`
	GroupClones *bool  `toml:"group_clones"` // pointer to detect unset
}

// TomlConfigLoader handles TOML-only configuration loading
type TomlConfigLoader struct{}

// NewTomlConfigLoader creates a new TOML configuration loader
func NewTomlConfigLoader() *TomlConfigLoader {
	return &TomlConfigLoader{}
}

// LoadConfig loads configuration from TOML files with ruff-like priority:
// 1. .pyscn.toml (dedicated config file)
// 2. pyproject.toml (with [tool.pyscn] section)
// 3. defaults
func (l *TomlConfigLoader) LoadConfig(startDir string) (*CloneConfig, error) {
	// Try .pyscn.toml first (highest priority)
	if config, err := l.loadFromPyscnToml(startDir); err == nil {
		return config, nil
	}

	// Try pyproject.toml as fallback
	if config, err := l.loadFromPyprojectToml(startDir); err == nil {
		return config, nil
	}

	// Return defaults if no config found
	return DefaultCloneConfig(), nil
}

// loadFromPyprojectToml loads config from pyproject.toml
func (l *TomlConfigLoader) loadFromPyprojectToml(startDir string) (*CloneConfig, error) {
	_, err := l.findPyprojectToml(startDir)
	if err != nil {
		return nil, err
	}

	return LoadPyprojectConfig(startDir)
}

// loadFromPyscnToml loads config from .pyscn.toml (dedicated config file)
func (l *TomlConfigLoader) loadFromPyscnToml(startDir string) (*CloneConfig, error) {
	configPath, err := l.findPyscnToml(startDir)
	if err != nil {
		return nil, err
	}

	// Read and parse .pyscn.toml
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config PyscnTomlConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	// Merge with defaults
	defaults := DefaultCloneConfig()
	l.mergePyscnTomlConfigs(defaults, &config)

	return defaults, nil
}

// findPyprojectToml walks up the directory tree to find pyproject.toml
func (l *TomlConfigLoader) findPyprojectToml(startDir string) (string, error) {
	return findPyprojectToml(startDir) // Reuse existing function
}

// findPyscnToml walks up the directory tree to find .pyscn.toml
func (l *TomlConfigLoader) findPyscnToml(startDir string) (string, error) {
	dir := startDir
	for {
		configPath := filepath.Join(dir, ".pyscn.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}

	return "", os.ErrNotExist
}

// mergePyscnTomlConfigs merges .pyscn.toml config into defaults
// using pointer booleans to detect unset values
// Priority: [clones] section > individual sections (for backward compatibility)
func (l *TomlConfigLoader) mergePyscnTomlConfigs(defaults *CloneConfig, pyscnToml *PyscnTomlConfig) {
	// First, merge from [clones] section if it exists (highest priority)
	l.mergeClonesSection(defaults, &pyscnToml.Clones)

	// Then, merge from individual sections as fallback (for backward compatibility)
	// Only apply if not already set by [clones] section

	// Analysis config
	if pyscnToml.Analysis.MinLines > 0 {
		defaults.Analysis.MinLines = pyscnToml.Analysis.MinLines
	}
	if pyscnToml.Analysis.MinNodes > 0 {
		defaults.Analysis.MinNodes = pyscnToml.Analysis.MinNodes
	}
	if pyscnToml.Analysis.MaxEditDistance > 0 {
		defaults.Analysis.MaxEditDistance = pyscnToml.Analysis.MaxEditDistance
	}
	if pyscnToml.Analysis.CostModelType != "" {
		defaults.Analysis.CostModelType = pyscnToml.Analysis.CostModelType
	}
	// Boolean fields: only override if explicitly set (pointer not nil)
	if pyscnToml.Analysis.IgnoreLiterals != nil {
		defaults.Analysis.IgnoreLiterals = *pyscnToml.Analysis.IgnoreLiterals
	}
	if pyscnToml.Analysis.IgnoreIdentifiers != nil {
		defaults.Analysis.IgnoreIdentifiers = *pyscnToml.Analysis.IgnoreIdentifiers
	}

	// Threshold config
	if pyscnToml.Thresholds.Type1Threshold > 0 {
		defaults.Thresholds.Type1Threshold = pyscnToml.Thresholds.Type1Threshold
	}
	if pyscnToml.Thresholds.Type2Threshold > 0 {
		defaults.Thresholds.Type2Threshold = pyscnToml.Thresholds.Type2Threshold
	}
	if pyscnToml.Thresholds.Type3Threshold > 0 {
		defaults.Thresholds.Type3Threshold = pyscnToml.Thresholds.Type3Threshold
	}
	if pyscnToml.Thresholds.Type4Threshold > 0 {
		defaults.Thresholds.Type4Threshold = pyscnToml.Thresholds.Type4Threshold
	}
	if pyscnToml.Thresholds.SimilarityThreshold > 0 {
		defaults.Thresholds.SimilarityThreshold = pyscnToml.Thresholds.SimilarityThreshold
	}

	// Grouping config
	if pyscnToml.Grouping.Mode != "" {
		defaults.Grouping.Mode = pyscnToml.Grouping.Mode
	}
	if pyscnToml.Grouping.Threshold > 0 {
		defaults.Grouping.Threshold = pyscnToml.Grouping.Threshold
	}
	if pyscnToml.Grouping.KCoreK > 0 {
		defaults.Grouping.KCoreK = pyscnToml.Grouping.KCoreK
	}

	// LSH config
	if pyscnToml.LSH.Enabled != "" {
		defaults.LSH.Enabled = pyscnToml.LSH.Enabled
	}
	if pyscnToml.LSH.AutoThreshold > 0 {
		defaults.LSH.AutoThreshold = pyscnToml.LSH.AutoThreshold
	}
	if pyscnToml.LSH.SimilarityThreshold > 0 {
		defaults.LSH.SimilarityThreshold = pyscnToml.LSH.SimilarityThreshold
	}
	if pyscnToml.LSH.Bands > 0 {
		defaults.LSH.Bands = pyscnToml.LSH.Bands
	}
	if pyscnToml.LSH.Rows > 0 {
		defaults.LSH.Rows = pyscnToml.LSH.Rows
	}
	if pyscnToml.LSH.Hashes > 0 {
		defaults.LSH.Hashes = pyscnToml.LSH.Hashes
	}

	// Input config
	if len(pyscnToml.Input.Paths) > 0 {
		defaults.Input.Paths = pyscnToml.Input.Paths
	}
	if len(pyscnToml.Input.IncludePatterns) > 0 {
		defaults.Input.IncludePatterns = pyscnToml.Input.IncludePatterns
	}
	if len(pyscnToml.Input.ExcludePatterns) > 0 {
		defaults.Input.ExcludePatterns = pyscnToml.Input.ExcludePatterns
	}
	// Boolean field: only override if explicitly set
	if pyscnToml.Input.Recursive != nil {
		defaults.Input.Recursive = *pyscnToml.Input.Recursive
	}

	// Output config
	if pyscnToml.Output.Format != "" {
		defaults.Output.Format = pyscnToml.Output.Format
	}
	if pyscnToml.Output.SortBy != "" {
		defaults.Output.SortBy = pyscnToml.Output.SortBy
	}
	// Boolean fields: only override if explicitly set
	if pyscnToml.Output.ShowDetails != nil {
		defaults.Output.ShowDetails = *pyscnToml.Output.ShowDetails
	}
	if pyscnToml.Output.ShowContent != nil {
		defaults.Output.ShowContent = *pyscnToml.Output.ShowContent
	}
	if pyscnToml.Output.GroupClones != nil {
		defaults.Output.GroupClones = *pyscnToml.Output.GroupClones
	}

	// Filtering config
	if pyscnToml.Filtering.MinSimilarity >= 0 {
		defaults.Filtering.MinSimilarity = pyscnToml.Filtering.MinSimilarity
	}
	if pyscnToml.Filtering.MaxSimilarity > 0 {
		defaults.Filtering.MaxSimilarity = pyscnToml.Filtering.MaxSimilarity
	}
	if len(pyscnToml.Filtering.EnabledCloneTypes) > 0 {
		defaults.Filtering.EnabledCloneTypes = pyscnToml.Filtering.EnabledCloneTypes
	}
	if pyscnToml.Filtering.MaxResults > 0 {
		defaults.Filtering.MaxResults = pyscnToml.Filtering.MaxResults
	}
}

// mergeClonesSection merges settings from the [clones] section
func (l *TomlConfigLoader) mergeClonesSection(defaults *CloneConfig, clones *ClonesConfig) {
	// Analysis settings
	if clones.MinLines > 0 {
		defaults.Analysis.MinLines = clones.MinLines
	}
	if clones.MinNodes > 0 {
		defaults.Analysis.MinNodes = clones.MinNodes
	}
	if clones.MaxEditDistance > 0 {
		defaults.Analysis.MaxEditDistance = clones.MaxEditDistance
	}
	if clones.CostModelType != "" {
		defaults.Analysis.CostModelType = clones.CostModelType
	}
	if clones.IgnoreLiterals != nil {
		defaults.Analysis.IgnoreLiterals = *clones.IgnoreLiterals
	}
	if clones.IgnoreIdentifiers != nil {
		defaults.Analysis.IgnoreIdentifiers = *clones.IgnoreIdentifiers
	}

	// Thresholds
	if clones.Type1Threshold > 0 {
		defaults.Thresholds.Type1Threshold = clones.Type1Threshold
	}
	if clones.Type2Threshold > 0 {
		defaults.Thresholds.Type2Threshold = clones.Type2Threshold
	}
	if clones.Type3Threshold > 0 {
		defaults.Thresholds.Type3Threshold = clones.Type3Threshold
	}
	if clones.Type4Threshold > 0 {
		defaults.Thresholds.Type4Threshold = clones.Type4Threshold
	}
	if clones.SimilarityThreshold > 0 {
		defaults.Thresholds.SimilarityThreshold = clones.SimilarityThreshold
	}

	// Filtering
	if clones.MinSimilarity >= 0 {
		defaults.Filtering.MinSimilarity = clones.MinSimilarity
	}
	if clones.MaxSimilarity > 0 {
		defaults.Filtering.MaxSimilarity = clones.MaxSimilarity
	}
	if len(clones.EnabledCloneTypes) > 0 {
		defaults.Filtering.EnabledCloneTypes = clones.EnabledCloneTypes
	}
	if clones.MaxResults > 0 {
		defaults.Filtering.MaxResults = clones.MaxResults
	}

	// Grouping
	if clones.GroupingMode != "" {
		defaults.Grouping.Mode = clones.GroupingMode
	}
	if clones.GroupingThreshold > 0 {
		defaults.Grouping.Threshold = clones.GroupingThreshold
	}
	if clones.KCoreK > 0 {
		defaults.Grouping.KCoreK = clones.KCoreK
	}

	// LSH settings (this is the critical part!)
	if clones.LSHEnabled != "" {
		defaults.LSH.Enabled = clones.LSHEnabled
	}
	if clones.LSHAutoThreshold > 0 {
		defaults.LSH.AutoThreshold = clones.LSHAutoThreshold
	}
	if clones.LSHSimilarityThreshold > 0 {
		defaults.LSH.SimilarityThreshold = clones.LSHSimilarityThreshold
	}
	if clones.LSHBands > 0 {
		defaults.LSH.Bands = clones.LSHBands
	}
	if clones.LSHRows > 0 {
		defaults.LSH.Rows = clones.LSHRows
	}
	if clones.LSHHashes > 0 {
		defaults.LSH.Hashes = clones.LSHHashes
	}

	// Performance
	if clones.MaxMemoryMB > 0 {
		defaults.Performance.MaxMemoryMB = clones.MaxMemoryMB
	}
	if clones.BatchSize > 0 {
		defaults.Performance.BatchSize = clones.BatchSize
	}
	if clones.EnableBatching != nil {
		defaults.Performance.EnableBatching = *clones.EnableBatching
	}
	if clones.MaxGoroutines > 0 {
		defaults.Performance.MaxGoroutines = clones.MaxGoroutines
	}
	if clones.TimeoutSeconds > 0 {
		defaults.Performance.TimeoutSeconds = clones.TimeoutSeconds
	}

	// Output
	if clones.ShowDetails != nil {
		defaults.Output.ShowDetails = *clones.ShowDetails
	}
	if clones.ShowContent != nil {
		defaults.Output.ShowContent = *clones.ShowContent
	}
	if clones.SortBy != "" {
		defaults.Output.SortBy = clones.SortBy
	}
	if clones.GroupClones != nil {
		defaults.Output.GroupClones = *clones.GroupClones
	}
}

// GetSupportedConfigFiles returns the list of supported TOML config files
// in order of precedence
func (l *TomlConfigLoader) GetSupportedConfigFiles() []string {
	return []string{
		".pyscn.toml",    // dedicated config file (highest priority)
		"pyproject.toml", // with [tool.pyscn] section
	}
}
