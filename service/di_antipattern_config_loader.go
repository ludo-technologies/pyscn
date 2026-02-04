package service

import (
	"fmt"
	"os"

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

	// Filtering - only override if explicitly set
	if override.MinSeverity != "" {
		merged.MinSeverity = override.MinSeverity
	}
	if override.SortBy != "" {
		merged.SortBy = override.SortBy
	}

	// ConfigPath - always override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// DI-specific options
	if override.ConstructorParamThreshold > 0 {
		merged.ConstructorParamThreshold = override.ConstructorParamThreshold
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

// FindDefaultConfigFile looks for TOML config files in the current directory
func (cl *DIAntipatternConfigurationLoaderImpl) FindDefaultConfigFile() string {
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
