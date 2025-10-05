package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// PyscnTomlConfig represents the structure of .pyscn.toml
type PyscnTomlConfig struct {
	Analysis    PyscnTomlAnalysisConfig  `toml:"analysis"`
	Thresholds  ThresholdConfig          `toml:"thresholds"`
	Filtering   PyscnTomlFilteringConfig `toml:"filtering"`
	Input       PyscnTomlInputConfig     `toml:"input"`
	Output      PyscnTomlOutputConfig    `toml:"output"`
	Performance PerformanceConfig        `toml:"performance"`
	Grouping    GroupingConfig           `toml:"grouping"`
	LSH         LSHConfig                `toml:"lsh"`
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
func (l *TomlConfigLoader) mergePyscnTomlConfigs(defaults *CloneConfig, pyscnToml *PyscnTomlConfig) {
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

// GetSupportedConfigFiles returns the list of supported TOML config files
// in order of precedence
func (l *TomlConfigLoader) GetSupportedConfigFiles() []string {
	return []string{
		".pyscn.toml",    // dedicated config file (highest priority)
		"pyproject.toml", // with [tool.pyscn] section
	}
}
