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

	// Boolean flags - use pointer type to distinguish "not set" (nil) from "set to false"
	merged.IgnoreTests = config.MergePtr(merged.IgnoreTests, override.IgnoreTests)

	// Analysis options
	merged.Recursive = override.Recursive

	// Array values
	merged.IncludePatterns = config.MergeSlice(merged.IncludePatterns, override.IncludePatterns)
	merged.ExcludePatterns = config.MergeSlice(merged.ExcludePatterns, override.ExcludePatterns)
	merged.Keywords = config.MergeSlice(merged.Keywords, override.Keywords)
	merged.Domains = config.MergeSlice(merged.Domains, override.Domains)
	merged.IgnorePatterns = config.MergeSlice(merged.IgnorePatterns, override.IgnorePatterns)
	merged.EnabledTypes = config.MergeSlice(merged.EnabledTypes, override.EnabledTypes)

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

// FindDefaultConfigFile looks for TOML config files from the current directory upward.
func (cl *MockDataConfigurationLoaderImpl) FindDefaultConfigFile() string {
	tomlLoader := config.NewTomlConfigLoader()
	return tomlLoader.FindConfigFileFromPath("")
}
