package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// PyscnTomlConfig represents the structure of .pyscn.toml
type PyscnTomlConfig struct {
	Clones ClonesConfig `toml:"clones"` // [clones] section - unified flat structure
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
	MaxMemoryMB    int   `toml:"max_memory_mb"`
	BatchSize      int   `toml:"batch_size"`
	EnableBatching *bool `toml:"enable_batching"` // pointer to detect unset
	MaxGoroutines  int   `toml:"max_goroutines"`
	TimeoutSeconds int   `toml:"timeout_seconds"`

	// Input
	Paths           []string `toml:"paths"`
	Recursive       *bool    `toml:"recursive"`        // pointer to detect unset
	IncludePatterns []string `toml:"include_patterns"`
	ExcludePatterns []string `toml:"exclude_patterns"`

	// Output
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
	// Merge from [clones] section (unified flat structure)
	mergeClonesSection(defaults, &pyscnToml.Clones)
}

// mergeClonesSection is moved to pyproject_loader.go and is now shared
// between .pyscn.toml and pyproject.toml loaders

// GetSupportedConfigFiles returns the list of supported TOML config files
// in order of precedence
func (l *TomlConfigLoader) GetSupportedConfigFiles() []string {
	return []string{
		".pyscn.toml",    // dedicated config file (highest priority)
		"pyproject.toml", // with [tool.pyscn] section
	}
}
