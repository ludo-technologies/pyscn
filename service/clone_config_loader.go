package service

import (
	"fmt"
	"os"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// CloneConfigurationLoader implements the domain.CloneConfigurationLoader interface
type CloneConfigurationLoader struct{}

// NewCloneConfigurationLoader creates a new clone configuration loader
func NewCloneConfigurationLoader() *CloneConfigurationLoader {
	return &CloneConfigurationLoader{}
}

// LoadCloneConfig loads clone detection configuration from file using TOML-only strategy
func (c *CloneConfigurationLoader) LoadCloneConfig(configPath string) (*domain.CloneRequest, error) {
	// Use TOML-only loader
	tomlLoader := config.NewTomlConfigLoader()
	cloneCfg, err := tomlLoader.LoadConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Convert clone config directly to clone request
	request := c.cloneConfigToCloneRequest(cloneCfg)
	return request, nil
}

// SaveCloneConfig saves clone detection configuration to a TOML file
func (c *CloneConfigurationLoader) SaveCloneConfig(cloneConfig *domain.CloneRequest, configPath string) error {
	// Load existing config or create new one using TOML loader
	var cfg *config.Config
	if _, err := os.Stat(configPath); err == nil {
		// File exists, load it using TOML loader
		loadedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			return fmt.Errorf("failed to load existing config: %w", err)
		}
		cfg = loadedCfg
	} else {
		// File doesn't exist, create default config
		cfg = config.DefaultConfig()
	}

	// Update clone detection section
	c.updateConfigFromCloneRequest(cfg, cloneConfig)

	// Save the updated configuration as TOML
	return config.SaveConfig(cfg, configPath)
}

// GetDefaultCloneConfig returns default clone detection configuration
func (c *CloneConfigurationLoader) GetDefaultCloneConfig() *domain.CloneRequest {
	// First, try to find and load a config file in the current directory
	configFile := c.FindDefaultConfigFile()
	if configFile != "" {
		if configReq, err := c.LoadCloneConfig(configFile); err == nil {
			return configReq
		}
		// If loading failed, fall back to hardcoded defaults
	}

	// Fall back to hardcoded default configuration
	defaultCloneConfig := config.DefaultPyscnConfig()
	return c.cloneConfigToCloneRequest(defaultCloneConfig)
}

// MergeConfig merges request values over loaded clone configuration.
// Zero values (0, "", nil) in override mean "not set" and preserve the
// config-backed base value; any other value is an explicit override.
func (c *CloneConfigurationLoader) MergeConfig(base *domain.CloneRequest, override *domain.CloneRequest) *domain.CloneRequest {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	merged := *base

	merged.Paths = config.MergeSlice(merged.Paths, override.Paths)

	merged.OutputFormat = config.Merge(merged.OutputFormat, override.OutputFormat)
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	merged.OutputPath = config.Merge(merged.OutputPath, override.OutputPath)
	merged.NoOpen = override.NoOpen

	merged.Recursive = config.MergePtr(merged.Recursive, override.Recursive)
	merged.IgnoreLiterals = config.MergePtr(merged.IgnoreLiterals, override.IgnoreLiterals)
	merged.IgnoreIdentifiers = config.MergePtr(merged.IgnoreIdentifiers, override.IgnoreIdentifiers)
	merged.SkipDocstrings = config.MergePtr(merged.SkipDocstrings, override.SkipDocstrings)
	merged.ShowDetails = config.MergePtr(merged.ShowDetails, override.ShowDetails)
	merged.ShowContent = config.MergePtr(merged.ShowContent, override.ShowContent)
	merged.GroupClones = config.MergePtr(merged.GroupClones, override.GroupClones)

	merged.MinLines = config.Merge(merged.MinLines, override.MinLines)
	merged.MinNodes = config.Merge(merged.MinNodes, override.MinNodes)
	merged.SimilarityThreshold = config.Merge(merged.SimilarityThreshold, override.SimilarityThreshold)
	merged.MaxEditDistance = config.Merge(merged.MaxEditDistance, override.MaxEditDistance)
	merged.Type1Threshold = config.Merge(merged.Type1Threshold, override.Type1Threshold)
	merged.Type2Threshold = config.Merge(merged.Type2Threshold, override.Type2Threshold)
	merged.Type3Threshold = config.Merge(merged.Type3Threshold, override.Type3Threshold)
	merged.Type4Threshold = config.Merge(merged.Type4Threshold, override.Type4Threshold)
	merged.MinSimilarity = config.Merge(merged.MinSimilarity, override.MinSimilarity)
	merged.MaxSimilarity = config.Merge(merged.MaxSimilarity, override.MaxSimilarity)
	merged.GroupThreshold = config.Merge(merged.GroupThreshold, override.GroupThreshold)
	merged.KCoreK = config.Merge(merged.KCoreK, override.KCoreK)
	merged.Timeout = config.Merge(merged.Timeout, override.Timeout)
	merged.LSHAutoThreshold = config.Merge(merged.LSHAutoThreshold, override.LSHAutoThreshold)
	merged.LSHSimilarityThreshold = config.Merge(merged.LSHSimilarityThreshold, override.LSHSimilarityThreshold)
	merged.LSHBands = config.Merge(merged.LSHBands, override.LSHBands)
	merged.LSHRows = config.Merge(merged.LSHRows, override.LSHRows)
	merged.LSHHashes = config.Merge(merged.LSHHashes, override.LSHHashes)

	merged.SortBy = config.Merge(merged.SortBy, override.SortBy)
	merged.GroupMode = config.Merge(merged.GroupMode, override.GroupMode)
	merged.LSHEnabled = config.Merge(merged.LSHEnabled, override.LSHEnabled)
	merged.ConfigPath = config.Merge(merged.ConfigPath, override.ConfigPath)

	merged.IncludePatterns = config.MergeSlice(merged.IncludePatterns, override.IncludePatterns)
	merged.ExcludePatterns = config.MergeSlice(merged.ExcludePatterns, override.ExcludePatterns)
	merged.CloneTypes = config.MergeSlice(merged.CloneTypes, override.CloneTypes)

	return &merged
}

// cloneConfigToCloneRequest converts a config.CloneConfig (TOML-based) to domain.CloneRequest
func (c *CloneConfigurationLoader) cloneConfigToCloneRequest(cloneCfg *config.PyscnConfig) *domain.CloneRequest {
	// Convert enabled clone types from string slice to domain clone types
	cloneTypes := make([]domain.CloneType, 0, len(cloneCfg.Filtering.EnabledCloneTypes))
	for _, typeStr := range cloneCfg.Filtering.EnabledCloneTypes {
		switch typeStr {
		case "type1":
			cloneTypes = append(cloneTypes, domain.Type1Clone)
		case "type2":
			cloneTypes = append(cloneTypes, domain.Type2Clone)
		case "type3":
			cloneTypes = append(cloneTypes, domain.Type3Clone)
		case "type4":
			cloneTypes = append(cloneTypes, domain.Type4Clone)
		}
	}

	// Convert sort criteria
	var sortBy domain.SortCriteria
	switch cloneCfg.Output.SortBy {
	case "similarity":
		sortBy = domain.SortBySimilarity
	case "size":
		sortBy = domain.SortBySize
	case "location":
		sortBy = domain.SortByLocation
	case "type":
		sortBy = domain.SortByComplexity // Reuse existing sort criteria
	default:
		sortBy = domain.SortBySimilarity
	}

	return &domain.CloneRequest{
		Paths:               cloneCfg.Input.Paths,
		MinLines:            cloneCfg.Analysis.MinLines,
		MinNodes:            cloneCfg.Analysis.MinNodes,
		SimilarityThreshold: cloneCfg.Thresholds.SimilarityThreshold,
		MaxEditDistance:     cloneCfg.Analysis.MaxEditDistance,
		IgnoreLiterals:      domain.BoolPtr(domain.BoolValue(cloneCfg.Analysis.IgnoreLiterals, false)),
		IgnoreIdentifiers:   domain.BoolPtr(domain.BoolValue(cloneCfg.Analysis.IgnoreIdentifiers, false)),
		SkipDocstrings:      domain.BoolPtr(domain.BoolValue(cloneCfg.Analysis.SkipDocstrings, true)),
		Type1Threshold:      cloneCfg.Thresholds.Type1Threshold,
		Type2Threshold:      cloneCfg.Thresholds.Type2Threshold,
		Type3Threshold:      cloneCfg.Thresholds.Type3Threshold,
		Type4Threshold:      cloneCfg.Thresholds.Type4Threshold,
		ShowDetails:         domain.BoolPtr(domain.BoolValue(cloneCfg.Output.ShowDetails, false)),
		ShowContent:         domain.BoolPtr(domain.BoolValue(cloneCfg.Output.ShowContent, false)),
		SortBy:              sortBy,
		GroupClones:         domain.BoolPtr(domain.BoolValue(cloneCfg.Output.GroupClones, true)),
		GroupMode:           cloneCfg.Grouping.Mode,
		GroupThreshold:      cloneCfg.Grouping.Threshold,
		KCoreK:              cloneCfg.Grouping.KCoreK,
		MinSimilarity:       cloneCfg.Filtering.MinSimilarity,
		MaxSimilarity:       cloneCfg.Filtering.MaxSimilarity,
		CloneTypes:          cloneTypes,
		OutputFormat:        domain.OutputFormatText, // Default, overridden by CLI
		Recursive:           domain.BoolPtr(domain.BoolValue(cloneCfg.Input.Recursive, true)),
		IncludePatterns:     cloneCfg.Input.IncludePatterns,
		ExcludePatterns:     cloneCfg.Input.ExcludePatterns,
		// DFA (Data Flow Analysis) - default enabled for multi-dimensional classification
		EnableDFA: domain.BoolValue(cloneCfg.Analysis.EnableDFA, true),
		// LSH settings
		LSHEnabled:             cloneCfg.LSH.Enabled,
		LSHAutoThreshold:       cloneCfg.LSH.AutoThreshold,
		LSHSimilarityThreshold: cloneCfg.LSH.SimilarityThreshold,
		LSHBands:               cloneCfg.LSH.Bands,
		LSHRows:                cloneCfg.LSH.Rows,
		LSHHashes:              cloneCfg.LSH.Hashes,
	}
}

// updateConfigFromCloneRequest updates a config.Config from a domain.CloneRequest
func (c *CloneConfigurationLoader) updateConfigFromCloneRequest(cfg *config.Config, req *domain.CloneRequest) {
	// Convert domain clone types to string slice
	cloneTypes := make([]string, 0, len(req.CloneTypes))
	for _, cloneType := range req.CloneTypes {
		switch cloneType {
		case domain.Type1Clone:
			cloneTypes = append(cloneTypes, "type1")
		case domain.Type2Clone:
			cloneTypes = append(cloneTypes, "type2")
		case domain.Type3Clone:
			cloneTypes = append(cloneTypes, "type3")
		case domain.Type4Clone:
			cloneTypes = append(cloneTypes, "type4")
		}
	}

	// Convert sort criteria to string
	var sortBy string
	switch req.SortBy {
	case domain.SortBySimilarity:
		sortBy = "similarity"
	case domain.SortBySize:
		sortBy = "size"
	case domain.SortByLocation:
		sortBy = "location"
	default:
		sortBy = "similarity"
	}

	// Update clone detection configuration using unified config
	if cfg.Clones == nil {
		cfg.Clones = config.DefaultPyscnConfig()
	}

	cfg.Clones.Analysis.MinLines = req.MinLines
	cfg.Clones.Analysis.MinNodes = req.MinNodes
	cfg.Clones.Analysis.MaxEditDistance = req.MaxEditDistance
	cfg.Clones.Analysis.CostModelType = "python" // Default cost model
	cfg.Clones.Analysis.IgnoreLiterals = domain.BoolPtr(domain.BoolValue(req.IgnoreLiterals, false))
	cfg.Clones.Analysis.IgnoreIdentifiers = domain.BoolPtr(domain.BoolValue(req.IgnoreIdentifiers, false))
	cfg.Clones.Analysis.SkipDocstrings = domain.BoolPtr(domain.BoolValue(req.SkipDocstrings, true))

	cfg.Clones.Thresholds.Type1Threshold = req.Type1Threshold
	cfg.Clones.Thresholds.Type2Threshold = req.Type2Threshold
	cfg.Clones.Thresholds.Type3Threshold = req.Type3Threshold
	cfg.Clones.Thresholds.Type4Threshold = req.Type4Threshold
	cfg.Clones.Thresholds.SimilarityThreshold = req.SimilarityThreshold

	cfg.Clones.Output.ShowContent = domain.BoolPtr(req.ShouldShowContent())
	cfg.Clones.Output.GroupClones = domain.BoolPtr(req.ShouldGroupClones())
	cfg.Clones.Output.SortBy = sortBy

	cfg.Clones.Filtering.MinSimilarity = req.MinSimilarity
	cfg.Clones.Filtering.MaxSimilarity = req.MaxSimilarity
	cfg.Clones.Filtering.EnabledCloneTypes = cloneTypes

	// Update analysis patterns if provided
	if len(req.IncludePatterns) > 0 {
		cfg.Analysis.IncludePatterns = req.IncludePatterns
	}
	if len(req.ExcludePatterns) > 0 {
		cfg.Analysis.ExcludePatterns = req.ExcludePatterns
	}
	cfg.Analysis.Recursive = domain.BoolValue(req.Recursive, true)
}

// FindDefaultConfigFile looks for TOML config files from the current directory upward.
func (c *CloneConfigurationLoader) FindDefaultConfigFile() string {
	tomlLoader := config.NewTomlConfigLoader()
	return tomlLoader.FindConfigFileFromPath("")
}
