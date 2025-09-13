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
	Clone PyprojectCloneConfig `toml:"clone"`
}

// PyprojectCloneConfig is a version of CloneConfig with pointer booleans
// to distinguish between unset and explicitly set false values
type PyprojectCloneConfig struct {
	Analysis    PyprojectCloneAnalysisConfig `toml:"analysis"`
	Thresholds  ThresholdConfig              `toml:"thresholds"`
	Filtering   PyprojectFilteringConfig     `toml:"filtering"`
	Input       PyprojectInputConfig         `toml:"input"`
	Output      PyprojectCloneOutputConfig   `toml:"output"`
	Performance PerformanceConfig            `toml:"performance"`
	Grouping    GroupingConfig               `toml:"grouping"`
	LSH         LSHConfig                    `toml:"lsh"`
}

type PyprojectCloneAnalysisConfig struct {
	MinLines          int     `toml:"min_lines"`
	MinNodes          int     `toml:"min_nodes"`
	MaxEditDistance   float64 `toml:"max_edit_distance"`
	IgnoreLiterals    *bool   `toml:"ignore_literals"`    // pointer to detect unset
	IgnoreIdentifiers *bool   `toml:"ignore_identifiers"` // pointer to detect unset
	CostModelType     string  `toml:"cost_model_type"`
}

type PyprojectFilteringConfig struct {
	MinSimilarity     float64  `toml:"min_similarity"`
	MaxSimilarity     float64  `toml:"max_similarity"`
	EnabledCloneTypes []string `toml:"enabled_clone_types"`
	MaxResults        int      `toml:"max_results"`
}

type PyprojectInputConfig struct {
	Paths           []string `toml:"paths"`
	Recursive       *bool    `toml:"recursive"` // pointer to detect unset
	IncludePatterns []string `toml:"include_patterns"`
	ExcludePatterns []string `toml:"exclude_patterns"`
}

type PyprojectCloneOutputConfig struct {
	Format      string `toml:"format"`
	ShowDetails *bool  `toml:"show_details"` // pointer to detect unset
	ShowContent *bool  `toml:"show_content"` // pointer to detect unset
	SortBy      string `toml:"sort_by"`
	GroupClones *bool  `toml:"group_clones"` // pointer to detect unset
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
	mergePyprojectConfigs(config, &pyproject.Tool.Pyscn.Clone)

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

// mergePyprojectConfigs merges pyproject.toml config into default config
// using pointer booleans to detect unset values
func mergePyprojectConfigs(defaults *CloneConfig, pyproject *PyprojectCloneConfig) {
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
	// Boolean fields: only override if explicitly set (pointer not nil)
	if pyproject.Analysis.IgnoreLiterals != nil {
		defaults.Analysis.IgnoreLiterals = *pyproject.Analysis.IgnoreLiterals
	}
	if pyproject.Analysis.IgnoreIdentifiers != nil {
		defaults.Analysis.IgnoreIdentifiers = *pyproject.Analysis.IgnoreIdentifiers
	}

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
	// Boolean field: only override if explicitly set
	if pyproject.Input.Recursive != nil {
		defaults.Input.Recursive = *pyproject.Input.Recursive
	}

	// Output config
	if pyproject.Output.Format != "" {
		defaults.Output.Format = pyproject.Output.Format
	}
	if pyproject.Output.SortBy != "" {
		defaults.Output.SortBy = pyproject.Output.SortBy
	}
	// Boolean fields: only override if explicitly set
	if pyproject.Output.ShowDetails != nil {
		defaults.Output.ShowDetails = *pyproject.Output.ShowDetails
	}
	if pyproject.Output.ShowContent != nil {
		defaults.Output.ShowContent = *pyproject.Output.ShowContent
	}
	if pyproject.Output.GroupClones != nil {
		defaults.Output.GroupClones = *pyproject.Output.GroupClones
	}

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
