package service

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// CBOConfigurationLoaderImpl implements the CBOConfigurationLoader interface
type CBOConfigurationLoaderImpl struct{}

// NewCBOConfigurationLoader creates a new CBO configuration loader service
func NewCBOConfigurationLoader() *CBOConfigurationLoaderImpl {
	return &CBOConfigurationLoaderImpl{}
}

// LoadConfig loads CBO configuration from the specified path using TOML-only strategy
func (cl *CBOConfigurationLoaderImpl) LoadConfig(path string) (*domain.CBORequest, error) {
	// Use TOML-only loader
	tomlLoader := config.NewTomlConfigLoader()
	pyscnCfg, err := tomlLoader.LoadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	// Convert pyscn config to CBO request
	return cl.configToRequest(pyscnCfg), nil
}

// LoadDefaultConfig loads the default CBO configuration, first checking for .pyscn.toml
func (cl *CBOConfigurationLoaderImpl) LoadDefaultConfig() *domain.CBORequest {
	// First, try to find and load a config file in the current directory
	configFile := cl.FindDefaultConfigFile()
	if configFile != "" {
		if configReq, err := cl.LoadConfig(configFile); err == nil {
			return configReq
		}
		// If loading failed, fall back to hardcoded defaults
	}

	// Fall back to hardcoded default configuration
	pyscnCfg := config.DefaultPyscnConfig()
	return cl.configToRequest(pyscnCfg)
}

// MergeConfig merges CLI flags with configuration file
func (cl *CBOConfigurationLoaderImpl) MergeConfig(base *domain.CBORequest, override *domain.CBORequest) *domain.CBORequest {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	// Start with base config
	merged := *base

	// Always override paths as they come from command arguments
	if len(override.Paths) > 0 {
		merged.Paths = override.Paths
	}

	// Output configuration
	if override.OutputFormat != "" {
		merged.OutputFormat = override.OutputFormat
	}
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	if override.OutputPath != "" {
		merged.OutputPath = override.OutputPath
	}

	// NoOpen flag
	merged.NoOpen = override.NoOpen

	// Filtering - only override if explicitly set (non-zero values)
	if override.MinCBO > 0 {
		merged.MinCBO = override.MinCBO
	}
	if override.MaxCBO > 0 {
		merged.MaxCBO = override.MaxCBO
	}

	// SortBy - override only if non-default
	if override.SortBy != "" && override.SortBy != domain.SortByComplexity {
		merged.SortBy = override.SortBy
	}

	// Thresholds - only override if explicitly set (non-zero values)
	if override.LowThreshold > 0 {
		merged.LowThreshold = override.LowThreshold
	}
	if override.MediumThreshold > 0 {
		merged.MediumThreshold = override.MediumThreshold
	}

	// ConfigPath - always override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// Boolean flags - use pointer type to distinguish "not set" (nil) from "set to false"
	// Only override if explicitly set (non-nil)
	if override.ShowZeros != nil {
		merged.ShowZeros = override.ShowZeros
	}
	// ShowDetails: default is false, so if override is true, use it
	if override.ShowDetails {
		merged.ShowDetails = true
	}
	if override.IncludeBuiltins != nil {
		merged.IncludeBuiltins = override.IncludeBuiltins
	}
	if override.IncludeImports != nil {
		merged.IncludeImports = override.IncludeImports
	}

	// Analysis options
	if override.Recursive != nil {
		merged.Recursive = override.Recursive
	}

	// Array values - override if provided
	if len(override.IncludePatterns) > 0 {
		merged.IncludePatterns = override.IncludePatterns
	}
	if len(override.ExcludePatterns) > 0 {
		merged.ExcludePatterns = override.ExcludePatterns
	}

	return &merged
}

// configToRequest converts a PyscnConfig to domain.CBORequest
func (cl *CBOConfigurationLoaderImpl) configToRequest(pyscnCfg *config.PyscnConfig) *domain.CBORequest {
	if pyscnCfg == nil {
		return &domain.CBORequest{
			LowThreshold:    domain.DefaultCBOLowThreshold,
			MediumThreshold: domain.DefaultCBOMediumThreshold,
			MinCBO:          0,
			MaxCBO:          0,
			ShowZeros:       domain.BoolPtr(false),
			ShowDetails:     false,
			IncludeBuiltins: domain.BoolPtr(false),
			IncludeImports:  domain.BoolPtr(true),
			SortBy:          domain.SortByComplexity,
			OutputFormat:    domain.OutputFormatText,
			Recursive:       domain.BoolPtr(true),
			IncludePatterns: []string{"**/*.py"},
			ExcludePatterns: []string{},
		}
	}

	return &domain.CBORequest{
		OutputFormat:    domain.OutputFormat(pyscnCfg.Output.Format),
		ShowDetails:     domain.BoolValue(pyscnCfg.Output.ShowDetails, false),
		LowThreshold:    pyscnCfg.CboLowThreshold,
		MediumThreshold: pyscnCfg.CboMediumThreshold,
		MinCBO:          pyscnCfg.CboMinCbo,
		MaxCBO:          pyscnCfg.CboMaxCbo,
		ShowZeros:       pyscnCfg.CboShowZeros,
		IncludeBuiltins: pyscnCfg.CboIncludeBuiltins,
		IncludeImports:  pyscnCfg.CboIncludeImports,
		SortBy:          domain.SortByComplexity, // Default, can be overridden
		Recursive:       pyscnCfg.AnalysisRecursive,
		IncludePatterns: pyscnCfg.AnalysisIncludePatterns,
		ExcludePatterns: pyscnCfg.AnalysisExcludePatterns,
	}
}

// FindDefaultConfigFile looks for TOML config files from the current directory upward.
func (cl *CBOConfigurationLoaderImpl) FindDefaultConfigFile() string {
	tomlLoader := config.NewTomlConfigLoader()
	return tomlLoader.FindConfigFileFromPath("")
}
