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
	Complexity ComplexityTomlConfig `toml:"complexity"`
	Clones     ClonesConfig         `toml:"clones"`
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

	// Merge with defaults using shared merge logic
	config := DefaultCloneConfig()
	mergeComplexitySection(config, &pyproject.Tool.Pyscn.Complexity)
	mergeClonesSection(config, &pyproject.Tool.Pyscn.Clones)

	return config, nil
}

// mergeComplexitySection merges settings from the [complexity] section
// This function is shared between .pyscn.toml and pyproject.toml loaders
func mergeComplexitySection(defaults *CloneConfig, complexity *ComplexityTomlConfig) {
	if complexity.LowThreshold != nil {
		defaults.ComplexityLowThreshold = *complexity.LowThreshold
	}
	if complexity.MediumThreshold != nil {
		defaults.ComplexityMediumThreshold = *complexity.MediumThreshold
	}
	if complexity.MaxComplexity != nil {
		defaults.ComplexityMaxComplexity = *complexity.MaxComplexity
	}
	if complexity.MinComplexity != nil {
		defaults.ComplexityMinComplexity = *complexity.MinComplexity
	}
}

// mergeClonesSection merges settings from the [clones] section
// This function is shared between .pyscn.toml and pyproject.toml loaders
func mergeClonesSection(defaults *CloneConfig, clones *ClonesConfig) {
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

	// LSH settings
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

	// Input
	if len(clones.Paths) > 0 {
		defaults.Input.Paths = clones.Paths
	}
	if clones.Recursive != nil {
		defaults.Input.Recursive = *clones.Recursive
	}
	if len(clones.IncludePatterns) > 0 {
		defaults.Input.IncludePatterns = clones.IncludePatterns
	}
	if len(clones.ExcludePatterns) > 0 {
		defaults.Input.ExcludePatterns = clones.ExcludePatterns
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
	if clones.Format != "" {
		defaults.Output.Format = clones.Format
	}
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
