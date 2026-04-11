package service

import (
	"fmt"
	"os"
	"slices"

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
// Without explicit flag tracking, values that still match domain defaults are
// treated as "not overridden" so config-backed values are preserved.
func (c *CloneConfigurationLoader) MergeConfig(base *domain.CloneRequest, override *domain.CloneRequest) *domain.CloneRequest {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	merged := *base
	defaultReq := domain.DefaultCloneRequest()

	if len(override.Paths) > 0 {
		merged.Paths = append([]string(nil), override.Paths...)
	}

	if override.OutputFormat != "" {
		merged.OutputFormat = override.OutputFormat
	}
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	if override.OutputPath != "" {
		merged.OutputPath = override.OutputPath
	}
	merged.NoOpen = override.NoOpen

	if override.Recursive != defaultReq.Recursive {
		merged.Recursive = override.Recursive
	}
	if override.IgnoreLiterals != defaultReq.IgnoreLiterals {
		merged.IgnoreLiterals = override.IgnoreLiterals
	}
	if override.IgnoreIdentifiers != defaultReq.IgnoreIdentifiers {
		merged.IgnoreIdentifiers = override.IgnoreIdentifiers
	}
	if override.SkipDocstrings != defaultReq.SkipDocstrings {
		merged.SkipDocstrings = override.SkipDocstrings
	}
	if override.EnableDFA != defaultReq.EnableDFA {
		merged.EnableDFA = override.EnableDFA
	}
	if override.ShowDetails != defaultReq.ShowDetails {
		merged.ShowDetails = override.ShowDetails
	}
	if override.ShowContent != defaultReq.ShowContent {
		merged.ShowContent = override.ShowContent
	}
	if override.GroupClones != defaultReq.GroupClones {
		merged.GroupClones = override.GroupClones
	}

	if override.MinLines != defaultReq.MinLines {
		merged.MinLines = override.MinLines
	}
	if override.MinNodes != defaultReq.MinNodes {
		merged.MinNodes = override.MinNodes
	}
	if override.SimilarityThreshold != defaultReq.SimilarityThreshold {
		merged.SimilarityThreshold = override.SimilarityThreshold
	}
	if override.MaxEditDistance != defaultReq.MaxEditDistance {
		merged.MaxEditDistance = override.MaxEditDistance
	}
	if override.Type1Threshold != defaultReq.Type1Threshold {
		merged.Type1Threshold = override.Type1Threshold
	}
	if override.Type2Threshold != defaultReq.Type2Threshold {
		merged.Type2Threshold = override.Type2Threshold
	}
	if override.Type3Threshold != defaultReq.Type3Threshold {
		merged.Type3Threshold = override.Type3Threshold
	}
	if override.Type4Threshold != defaultReq.Type4Threshold {
		merged.Type4Threshold = override.Type4Threshold
	}
	if override.MinSimilarity != defaultReq.MinSimilarity {
		merged.MinSimilarity = override.MinSimilarity
	}
	if override.MaxSimilarity != defaultReq.MaxSimilarity {
		merged.MaxSimilarity = override.MaxSimilarity
	}
	if override.GroupThreshold != 0 && override.GroupThreshold != defaultReq.GroupThreshold {
		merged.GroupThreshold = override.GroupThreshold
	}
	if override.KCoreK != 0 && override.KCoreK != defaultReq.KCoreK {
		merged.KCoreK = override.KCoreK
	}
	if override.Timeout != 0 {
		merged.Timeout = override.Timeout
	}
	if override.LSHAutoThreshold != 0 && override.LSHAutoThreshold != defaultReq.LSHAutoThreshold {
		merged.LSHAutoThreshold = override.LSHAutoThreshold
	}
	if override.LSHSimilarityThreshold != 0 && override.LSHSimilarityThreshold != defaultReq.LSHSimilarityThreshold {
		merged.LSHSimilarityThreshold = override.LSHSimilarityThreshold
	}
	if override.LSHBands != 0 && override.LSHBands != defaultReq.LSHBands {
		merged.LSHBands = override.LSHBands
	}
	if override.LSHRows != 0 && override.LSHRows != defaultReq.LSHRows {
		merged.LSHRows = override.LSHRows
	}
	if override.LSHHashes != 0 && override.LSHHashes != defaultReq.LSHHashes {
		merged.LSHHashes = override.LSHHashes
	}

	if override.SortBy != "" && override.SortBy != defaultReq.SortBy {
		merged.SortBy = override.SortBy
	}
	if override.GroupMode != "" && override.GroupMode != defaultReq.GroupMode {
		merged.GroupMode = override.GroupMode
	}
	if override.LSHEnabled != "" && override.LSHEnabled != defaultReq.LSHEnabled {
		merged.LSHEnabled = override.LSHEnabled
	}
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	if len(override.IncludePatterns) > 0 && !slices.Equal(override.IncludePatterns, defaultReq.IncludePatterns) {
		merged.IncludePatterns = append([]string(nil), override.IncludePatterns...)
	}
	if len(override.ExcludePatterns) > 0 && !slices.Equal(override.ExcludePatterns, defaultReq.ExcludePatterns) {
		merged.ExcludePatterns = append([]string(nil), override.ExcludePatterns...)
	}
	if len(override.CloneTypes) > 0 && !slices.Equal(override.CloneTypes, defaultReq.CloneTypes) {
		merged.CloneTypes = append([]domain.CloneType(nil), override.CloneTypes...)
	}

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
		IgnoreLiterals:      domain.BoolValue(cloneCfg.Analysis.IgnoreLiterals, false),
		IgnoreIdentifiers:   domain.BoolValue(cloneCfg.Analysis.IgnoreIdentifiers, false),
		Type1Threshold:      cloneCfg.Thresholds.Type1Threshold,
		Type2Threshold:      cloneCfg.Thresholds.Type2Threshold,
		Type3Threshold:      cloneCfg.Thresholds.Type3Threshold,
		Type4Threshold:      cloneCfg.Thresholds.Type4Threshold,
		ShowDetails:         domain.BoolValue(cloneCfg.Output.ShowDetails, false),
		ShowContent:         domain.BoolValue(cloneCfg.Output.ShowContent, false),
		SortBy:              sortBy,
		GroupClones:         domain.BoolValue(cloneCfg.Output.GroupClones, true),
		MinSimilarity:       cloneCfg.Filtering.MinSimilarity,
		MaxSimilarity:       cloneCfg.Filtering.MaxSimilarity,
		CloneTypes:          cloneTypes,
		OutputFormat:        domain.OutputFormatText, // Default, overridden by CLI
		Recursive:           domain.BoolValue(cloneCfg.Input.Recursive, true),
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
	cfg.Clones.Analysis.IgnoreLiterals = domain.BoolPtr(req.IgnoreLiterals)
	cfg.Clones.Analysis.IgnoreIdentifiers = domain.BoolPtr(req.IgnoreIdentifiers)

	cfg.Clones.Thresholds.Type1Threshold = req.Type1Threshold
	cfg.Clones.Thresholds.Type2Threshold = req.Type2Threshold
	cfg.Clones.Thresholds.Type3Threshold = req.Type3Threshold
	cfg.Clones.Thresholds.Type4Threshold = req.Type4Threshold
	cfg.Clones.Thresholds.SimilarityThreshold = req.SimilarityThreshold

	cfg.Clones.Output.ShowContent = domain.BoolPtr(req.ShowContent)
	cfg.Clones.Output.GroupClones = domain.BoolPtr(req.GroupClones)
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
	cfg.Analysis.Recursive = req.Recursive
}

// FindDefaultConfigFile looks for TOML config files from the current directory upward.
func (c *CloneConfigurationLoader) FindDefaultConfigFile() string {
	tomlLoader := config.NewTomlConfigLoader()
	return tomlLoader.FindConfigFileFromPath("")
}
