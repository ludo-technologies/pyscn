package config

import (
	"os"
	"path/filepath"
	
	"github.com/pelletier/go-toml/v2"
)

// PyprojectToml represents the structure of pyproject.toml
type PyprojectToml struct {
	Tool ToolConfig `toml:"tool"`
}

// ToolConfig represents the [tool] section
type ToolConfig struct {
	Pyscn PyscnConfig `toml:"pyscn"`
}

// PyscnConfig represents the [tool.pyscn] section  
type PyscnConfig struct {
	Clone CloneConfig `toml:"clone"`
}

// LoadPyprojectConfig loads clone configuration from pyproject.toml
func LoadPyprojectConfig(startDir string) (*CloneConfig, error) {
	// Find pyproject.toml file (walk up directory tree)
	configPath, err := findPyprojectToml(startDir)
	if err != nil {
		// Return default config if no pyproject.toml found
		return DefaultCloneConfig(), nil
	}
	
	// Read and parse pyproject.toml
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	
	var pyproject PyprojectToml
	if err := toml.Unmarshal(data, &pyproject); err != nil {
		return nil, err
	}
	
	// Merge with defaults
	config := DefaultCloneConfig()
	mergeConfigs(config, &pyproject.Tool.Pyscn.Clone)
	
	return config, nil
}

// findPyprojectToml walks up the directory tree to find pyproject.toml
func findPyprojectToml(startDir string) (string, error) {
	dir := startDir
	for {
		configPath := filepath.Join(dir, "pyproject.toml")
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

// mergeConfigs merges pyproject.toml config into default config
func mergeConfigs(defaults *CloneConfig, pyproject *CloneConfig) {
	// Only override non-zero values from pyproject.toml
	
	// Analysis config
	if pyproject.Analysis.MinLines > 0 {
		defaults.Analysis.MinLines = pyproject.Analysis.MinLines
	}
	if pyproject.Analysis.MinNodes > 0 {
		defaults.Analysis.MinNodes = pyproject.Analysis.MinNodes
	}
	if pyproject.Analysis.MaxEditDistance > 0 {
		defaults.Analysis.MaxEditDistance = pyproject.Analysis.MaxEditDistance
	}
	if pyproject.Analysis.CostModelType != "" {
		defaults.Analysis.CostModelType = pyproject.Analysis.CostModelType
	}
	// Boolean fields need special handling (false is a valid value)
	defaults.Analysis.IgnoreLiterals = pyproject.Analysis.IgnoreLiterals
	defaults.Analysis.IgnoreIdentifiers = pyproject.Analysis.IgnoreIdentifiers
	
	// Threshold config
	if pyproject.Thresholds.Type1Threshold > 0 {
		defaults.Thresholds.Type1Threshold = pyproject.Thresholds.Type1Threshold
	}
	if pyproject.Thresholds.Type2Threshold > 0 {
		defaults.Thresholds.Type2Threshold = pyproject.Thresholds.Type2Threshold
	}
	if pyproject.Thresholds.Type3Threshold > 0 {
		defaults.Thresholds.Type3Threshold = pyproject.Thresholds.Type3Threshold
	}
	if pyproject.Thresholds.Type4Threshold > 0 {
		defaults.Thresholds.Type4Threshold = pyproject.Thresholds.Type4Threshold
	}
	if pyproject.Thresholds.SimilarityThreshold > 0 {
		defaults.Thresholds.SimilarityThreshold = pyproject.Thresholds.SimilarityThreshold
	}
	
	// Grouping config
	if pyproject.Grouping.Mode != "" {
		defaults.Grouping.Mode = pyproject.Grouping.Mode
	}
	if pyproject.Grouping.Threshold > 0 {
		defaults.Grouping.Threshold = pyproject.Grouping.Threshold
	}
	if pyproject.Grouping.KCoreK > 0 {
		defaults.Grouping.KCoreK = pyproject.Grouping.KCoreK
	}
	
	// LSH config
	if pyproject.LSH.Enabled != "" {
		defaults.LSH.Enabled = pyproject.LSH.Enabled
	}
	if pyproject.LSH.AutoThreshold > 0 {
		defaults.LSH.AutoThreshold = pyproject.LSH.AutoThreshold
	}
	if pyproject.LSH.SimilarityThreshold > 0 {
		defaults.LSH.SimilarityThreshold = pyproject.LSH.SimilarityThreshold
	}
	if pyproject.LSH.Bands > 0 {
		defaults.LSH.Bands = pyproject.LSH.Bands
	}
	if pyproject.LSH.Rows > 0 {
		defaults.LSH.Rows = pyproject.LSH.Rows
	}
	if pyproject.LSH.Hashes > 0 {
		defaults.LSH.Hashes = pyproject.LSH.Hashes
	}
	
	// Input config  
	if len(pyproject.Input.Paths) > 0 {
		defaults.Input.Paths = pyproject.Input.Paths
	}
	if len(pyproject.Input.IncludePatterns) > 0 {
		defaults.Input.IncludePatterns = pyproject.Input.IncludePatterns
	}
	if len(pyproject.Input.ExcludePatterns) > 0 {
		defaults.Input.ExcludePatterns = pyproject.Input.ExcludePatterns
	}
	defaults.Input.Recursive = pyproject.Input.Recursive // Boolean field
	
	// Output config
	if pyproject.Output.Format != "" {
		defaults.Output.Format = pyproject.Output.Format
	}
	if pyproject.Output.SortBy != "" {
		defaults.Output.SortBy = pyproject.Output.SortBy
	}
	defaults.Output.ShowDetails = pyproject.Output.ShowDetails
	defaults.Output.ShowContent = pyproject.Output.ShowContent
	defaults.Output.GroupClones = pyproject.Output.GroupClones
	
	// Filtering config
	if pyproject.Filtering.MinSimilarity >= 0 {
		defaults.Filtering.MinSimilarity = pyproject.Filtering.MinSimilarity
	}
	if pyproject.Filtering.MaxSimilarity > 0 {
		defaults.Filtering.MaxSimilarity = pyproject.Filtering.MaxSimilarity
	}
	if len(pyproject.Filtering.EnabledCloneTypes) > 0 {
		defaults.Filtering.EnabledCloneTypes = pyproject.Filtering.EnabledCloneTypes
	}
	if pyproject.Filtering.MaxResults > 0 {
		defaults.Filtering.MaxResults = pyproject.Filtering.MaxResults
	}
}