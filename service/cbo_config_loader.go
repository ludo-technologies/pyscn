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

	// A zero-value override field means "not set", so the base wins. Never
	// compare against domain defaults: an explicit override that happens to
	// equal a default must still take precedence (issue #553).
	merged.Paths = config.MergeSlice(merged.Paths, override.Paths)

	// Output configuration
	merged.OutputFormat = config.Merge(merged.OutputFormat, override.OutputFormat)
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	merged.OutputPath = config.Merge(merged.OutputPath, override.OutputPath)

	// NoOpen flag - plain bool, always take override
	merged.NoOpen = override.NoOpen

	// Filtering
	merged.MinCBO = config.Merge(merged.MinCBO, override.MinCBO)
	merged.MaxCBO = config.Merge(merged.MaxCBO, override.MaxCBO)

	merged.SortBy = config.Merge(merged.SortBy, override.SortBy)

	// Thresholds
	merged.LowThreshold = config.Merge(merged.LowThreshold, override.LowThreshold)
	merged.MediumThreshold = config.Merge(merged.MediumThreshold, override.MediumThreshold)

	// ConfigPath - always override if provided
	merged.ConfigPath = config.Merge(merged.ConfigPath, override.ConfigPath)

	// Boolean flags - pointer type distinguishes "not set" (nil) from an
	// explicit value (including false).
	merged.ShowZeros = config.MergePtr(merged.ShowZeros, override.ShowZeros)
	merged.ShowDetails = config.MergePtr(merged.ShowDetails, override.ShowDetails)
	merged.IncludeBuiltins = config.MergePtr(merged.IncludeBuiltins, override.IncludeBuiltins)
	merged.IncludeImports = config.MergePtr(merged.IncludeImports, override.IncludeImports)
	merged.GroupNamespaceImports = config.MergePtr(merged.GroupNamespaceImports, override.GroupNamespaceImports)

	// Analysis options
	merged.Recursive = config.MergePtr(merged.Recursive, override.Recursive)

	// Array values
	merged.IncludePatterns = config.MergeSlice(merged.IncludePatterns, override.IncludePatterns)
	merged.ExcludePatterns = config.MergeSlice(merged.ExcludePatterns, override.ExcludePatterns)

	return &merged
}

// configToRequest converts a PyscnConfig to domain.CBORequest
func (cl *CBOConfigurationLoaderImpl) configToRequest(pyscnCfg *config.PyscnConfig) *domain.CBORequest {
	if pyscnCfg == nil {
		return &domain.CBORequest{
			LowThreshold:          domain.DefaultCBOLowThreshold,
			MediumThreshold:       domain.DefaultCBOMediumThreshold,
			MinCBO:                0,
			MaxCBO:                0,
			ShowZeros:             domain.BoolPtr(false),
			ShowDetails:           domain.BoolPtr(false),
			IncludeBuiltins:       domain.BoolPtr(false),
			IncludeImports:        domain.BoolPtr(true),
			GroupNamespaceImports: domain.BoolPtr(true),
			SortBy:                domain.SortByComplexity,
			OutputFormat:          domain.OutputFormatText,
			Recursive:             domain.BoolPtr(true),
			IncludePatterns:       domain.DefaultAnalysisIncludePatterns(),
			ExcludePatterns:       []string{},
		}
	}

	return &domain.CBORequest{
		OutputFormat:          domain.OutputFormat(pyscnCfg.Output.Format),
		ShowDetails:           domain.BoolPtr(domain.BoolValue(pyscnCfg.Output.ShowDetails, false)),
		LowThreshold:          pyscnCfg.CboLowThreshold,
		MediumThreshold:       pyscnCfg.CboMediumThreshold,
		MinCBO:                pyscnCfg.CboMinCbo,
		MaxCBO:                pyscnCfg.CboMaxCbo,
		ShowZeros:             pyscnCfg.CboShowZeros,
		IncludeBuiltins:       pyscnCfg.CboIncludeBuiltins,
		IncludeImports:        pyscnCfg.CboIncludeImports,
		GroupNamespaceImports: pyscnCfg.CboGroupNamespaceImports,
		SortBy:                domain.SortByComplexity, // Default, can be overridden
		Recursive:             pyscnCfg.AnalysisRecursive,
		IncludePatterns:       pyscnCfg.AnalysisIncludePatterns,
		ExcludePatterns:       pyscnCfg.AnalysisExcludePatterns,
	}
}

// FindDefaultConfigFile looks for TOML config files from the current directory upward.
func (cl *CBOConfigurationLoaderImpl) FindDefaultConfigFile() string {
	tomlLoader := config.NewTomlConfigLoader()
	return tomlLoader.FindConfigFileFromPath("")
}
