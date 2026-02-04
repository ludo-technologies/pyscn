package service

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// MockDataConfigurationLoaderImpl implements the MockDataConfigurationLoader interface
type MockDataConfigurationLoaderImpl struct{}

// NewMockDataConfigurationLoader creates a new mock data configuration loader service
func NewMockDataConfigurationLoader() *MockDataConfigurationLoaderImpl {
	return &MockDataConfigurationLoaderImpl{}
}

// LoadConfig loads mock data configuration from the specified path using TOML-only strategy
func (cl *MockDataConfigurationLoaderImpl) LoadConfig(path string) (*domain.MockDataRequest, error) {
	// Use TOML-only loader
	tomlLoader := config.NewTomlConfigLoader()
	pyscnCfg, err := tomlLoader.LoadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	// Convert pyscn config to MockData request
	return cl.configToRequest(pyscnCfg), nil
}

// LoadDefaultConfig loads the default mock data configuration, first checking for .pyscn.toml
func (cl *MockDataConfigurationLoaderImpl) LoadDefaultConfig() *domain.MockDataRequest {
	// First, try to find and load a config file in the current directory
	configFile := cl.FindDefaultConfigFile()
	if configFile != "" {
		if configReq, err := cl.LoadConfig(configFile); err == nil {
			return configReq
		}
		// If loading failed, fall back to hardcoded defaults
	}

	// Fall back to hardcoded default configuration
	return domain.DefaultMockDataRequest()
}

// MergeConfig merges CLI flags with configuration file
func (cl *MockDataConfigurationLoaderImpl) MergeConfig(base *domain.MockDataRequest, override *domain.MockDataRequest) *domain.MockDataRequest {
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

	// Boolean flags - use pointer type to distinguish "not set" (nil) from "set to false"
	if override.IgnoreTests != nil {
		merged.IgnoreTests = override.IgnoreTests
	}

	// Analysis options
	merged.Recursive = override.Recursive

	// Array values - override if provided
	if len(override.IncludePatterns) > 0 {
		merged.IncludePatterns = override.IncludePatterns
	}
	if len(override.ExcludePatterns) > 0 {
		merged.ExcludePatterns = override.ExcludePatterns
	}
	if len(override.Keywords) > 0 {
		merged.Keywords = override.Keywords
	}
	if len(override.Domains) > 0 {
		merged.Domains = override.Domains
	}
	if len(override.IgnorePatterns) > 0 {
		merged.IgnorePatterns = override.IgnorePatterns
	}
	if len(override.EnabledTypes) > 0 {
		merged.EnabledTypes = override.EnabledTypes
	}

	return &merged
}

// configToRequest converts a PyscnConfig to domain.MockDataRequest
func (cl *MockDataConfigurationLoaderImpl) configToRequest(pyscnCfg *config.PyscnConfig) *domain.MockDataRequest {
	if pyscnCfg == nil {
		return domain.DefaultMockDataRequest()
	}

	// Convert config values, falling back to defaults
	minSeverity := domain.MockDataSeverity(pyscnCfg.MockDataMinSeverity)
	if minSeverity == "" {
		minSeverity = domain.MockDataSeverityWarning
	}

	sortBy := domain.MockDataSortCriteria(pyscnCfg.MockDataSortBy)
	if sortBy == "" {
		sortBy = domain.MockDataSortBySeverity
	}

	keywords := pyscnCfg.MockDataKeywords
	if len(keywords) == 0 {
		keywords = domain.DefaultMockDataKeywords()
	}

	domains := pyscnCfg.MockDataDomains
	if len(domains) == 0 {
		domains = domain.DefaultMockDataDomains()
	}

	return &domain.MockDataRequest{
		OutputFormat:    domain.OutputFormat(pyscnCfg.Output.Format),
		MinSeverity:     minSeverity,
		SortBy:          sortBy,
		IgnoreTests:     pyscnCfg.MockDataIgnoreTests,
		Keywords:        keywords,
		Domains:         domains,
		IgnorePatterns:  pyscnCfg.MockDataIgnorePatterns,
		Recursive:       domain.BoolValue(pyscnCfg.AnalysisRecursive, true),
		IncludePatterns: pyscnCfg.AnalysisIncludePatterns,
		ExcludePatterns: pyscnCfg.AnalysisExcludePatterns,
	}
}

// FindDefaultConfigFile looks for TOML config files starting from the current directory
// and walking up the directory tree
func (cl *MockDataConfigurationLoaderImpl) FindDefaultConfigFile() string {
	tomlLoader := config.NewTomlConfigLoader()
	return tomlLoader.FindConfigFileFromPath("")
}
