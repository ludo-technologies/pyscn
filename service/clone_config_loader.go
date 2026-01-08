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

// FindDefaultConfigFile looks for TOML config files in the current directory
func (c *CloneConfigurationLoader) FindDefaultConfigFile() string {
	// Use TOML-only strategy
	tomlLoader := config.NewTomlConfigLoader()
	configFiles := tomlLoader.GetSupportedConfigFiles()

	for _, filename := range configFiles {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}

	return "" // No config file found
}
