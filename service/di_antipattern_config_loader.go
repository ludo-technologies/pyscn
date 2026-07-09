package service

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// DIAntipatternConfigurationLoaderImpl implements the DIAntipatternConfigurationLoader interface
type DIAntipatternConfigurationLoaderImpl struct{}

// NewDIAntipatternConfigurationLoader creates a new DI anti-pattern configuration loader service
func NewDIAntipatternConfigurationLoader() *DIAntipatternConfigurationLoaderImpl {
	return &DIAntipatternConfigurationLoaderImpl{}
}

// LoadConfig loads DI anti-pattern configuration from the specified path using TOML-only strategy
func (cl *DIAntipatternConfigurationLoaderImpl) LoadConfig(path string) (*domain.DIAntipatternRequest, error) {
	// Use TOML-only loader
	tomlLoader := config.NewTomlConfigLoader()
	pyscnCfg, err := tomlLoader.LoadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	// Convert pyscn config to DIAntipattern request
	return cl.configToRequest(pyscnCfg), nil
}

// LoadDefaultConfig loads the default DI anti-pattern configuration, first checking for .pyscn.toml
func (cl *DIAntipatternConfigurationLoaderImpl) LoadDefaultConfig() *domain.DIAntipatternRequest {
	// First, try to find and load a config file in the current directory
	configFile := cl.FindDefaultConfigFile()
	if configFile != "" {
		if configReq, err := cl.LoadConfig(configFile); err == nil {
			return configReq
		}
		// If loading failed, fall back to hardcoded defaults
	}

	// Fall back to hardcoded default configuration
	return domain.DefaultDIAntipatternRequest()
}

// MergeConfig merges CLI flags with configuration file
func (cl *DIAntipatternConfigurationLoaderImpl) MergeConfig(base *domain.DIAntipatternRequest, override *domain.DIAntipatternRequest) *domain.DIAntipatternRequest {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	// Start with base config
	merged := *base

	// Always override paths as they come from command arguments
	merged.Paths = config.MergeSlice(merged.Paths, override.Paths)

	// Output configuration
	merged.OutputFormat = config.Merge(merged.OutputFormat, override.OutputFormat)
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	merged.OutputPath = config.Merge(merged.OutputPath, override.OutputPath)

	// NoOpen flag
	merged.NoOpen = override.NoOpen

	// Filtering
	merged.MinSeverity = config.Merge(merged.MinSeverity, override.MinSeverity)
	merged.SortBy = config.Merge(merged.SortBy, override.SortBy)

	// ConfigPath
	merged.ConfigPath = config.Merge(merged.ConfigPath, override.ConfigPath)

	// DI-specific options
	merged.ConstructorParamThreshold = config.Merge(merged.ConstructorParamThreshold, override.ConstructorParamThreshold)

	// Analysis options
	merged.Recursive = config.MergePtr(merged.Recursive, override.Recursive)

	// Array values
	merged.IncludePatterns = config.MergeSlice(merged.IncludePatterns, override.IncludePatterns)
	merged.ExcludePatterns = config.MergeSlice(merged.ExcludePatterns, override.ExcludePatterns)

	return &merged
}

// configToRequest converts a PyscnConfig to domain.DIAntipatternRequest
func (cl *DIAntipatternConfigurationLoaderImpl) configToRequest(pyscnCfg *config.PyscnConfig) *domain.DIAntipatternRequest {
	if pyscnCfg == nil {
		return domain.DefaultDIAntipatternRequest()
	}

	// Convert config values, falling back to defaults
	minSeverity := domain.DIAntipatternSeverity(pyscnCfg.DIMinSeverity)
	if minSeverity == "" {
		minSeverity = domain.DIAntipatternSeverityWarning
	}

	threshold := pyscnCfg.DIConstructorParamThreshold
	if threshold == 0 {
		threshold = domain.DefaultDIConstructorParamThreshold
	}

	return &domain.DIAntipatternRequest{
		OutputFormat:              domain.OutputFormat(pyscnCfg.Output.Format),
		MinSeverity:               minSeverity,
		ConstructorParamThreshold: threshold,
		Recursive:                 domain.BoolPtr(domain.BoolValue(pyscnCfg.AnalysisRecursive, true)),
		IncludePatterns:           pyscnCfg.AnalysisIncludePatterns,
		ExcludePatterns:           pyscnCfg.AnalysisExcludePatterns,
	}
}

// FindDefaultConfigFile looks for TOML config files from the current directory upward.
func (cl *DIAntipatternConfigurationLoaderImpl) FindDefaultConfigFile() string {
	tomlLoader := config.NewTomlConfigLoader()
	return tomlLoader.FindConfigFileFromPath("")
}
