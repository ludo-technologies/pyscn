package service

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// LCOMConfigurationLoaderImpl implements the LCOMConfigurationLoader interface
type LCOMConfigurationLoaderImpl struct{}

// NewLCOMConfigurationLoader creates a new LCOM configuration loader service
func NewLCOMConfigurationLoader() *LCOMConfigurationLoaderImpl {
	return &LCOMConfigurationLoaderImpl{}
}

// LoadConfig loads LCOM configuration from the specified path using TOML-only strategy
func (cl *LCOMConfigurationLoaderImpl) LoadConfig(path string) (*domain.LCOMRequest, error) {
	tomlLoader := config.NewTomlConfigLoader()
	pyscnCfg, err := tomlLoader.LoadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}
	return cl.configToRequest(pyscnCfg), nil
}

// LoadDefaultConfig loads the default LCOM configuration, first checking for .pyscn.toml
func (cl *LCOMConfigurationLoaderImpl) LoadDefaultConfig() *domain.LCOMRequest {
	configFile := cl.FindDefaultConfigFile()
	if configFile != "" {
		if configReq, err := cl.LoadConfig(configFile); err == nil {
			return configReq
		}
	}

	pyscnCfg := config.DefaultPyscnConfig()
	return cl.configToRequest(pyscnCfg)
}

// MergeConfig merges CLI flags with configuration file
func (cl *LCOMConfigurationLoaderImpl) MergeConfig(base *domain.LCOMRequest, override *domain.LCOMRequest) *domain.LCOMRequest {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	merged := *base

	// A zero-value override field means "not set", so the base wins. Never
	// compare against domain defaults: an explicit override that happens to
	// equal a default must still take precedence (issue #553).
	merged.Paths = config.MergeSlice(merged.Paths, override.Paths)
	merged.OutputFormat = config.Merge(merged.OutputFormat, override.OutputFormat)
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	merged.OutputPath = config.Merge(merged.OutputPath, override.OutputPath)
	// NoOpen flag - plain bool, always take override
	merged.NoOpen = override.NoOpen

	merged.MinLCOM = config.Merge(merged.MinLCOM, override.MinLCOM)
	merged.MaxLCOM = config.Merge(merged.MaxLCOM, override.MaxLCOM)
	merged.SortBy = config.Merge(merged.SortBy, override.SortBy)
	merged.LowThreshold = config.Merge(merged.LowThreshold, override.LowThreshold)
	merged.MediumThreshold = config.Merge(merged.MediumThreshold, override.MediumThreshold)
	merged.ConfigPath = config.Merge(merged.ConfigPath, override.ConfigPath)
	merged.ShowDetails = config.MergePtr(merged.ShowDetails, override.ShowDetails)
	merged.Recursive = config.MergePtr(merged.Recursive, override.Recursive)
	merged.IncludePatterns = config.MergeSlice(merged.IncludePatterns, override.IncludePatterns)
	merged.ExcludePatterns = config.MergeSlice(merged.ExcludePatterns, override.ExcludePatterns)

	return &merged
}

// FindDefaultConfigFile searches for a config file in the current directory
func (cl *LCOMConfigurationLoaderImpl) FindDefaultConfigFile() string {
	tomlLoader := config.NewTomlConfigLoader()
	return tomlLoader.FindConfigFileFromPath("")
}

// configToRequest converts a PyscnConfig to domain.LCOMRequest
func (cl *LCOMConfigurationLoaderImpl) configToRequest(pyscnCfg *config.PyscnConfig) *domain.LCOMRequest {
	if pyscnCfg == nil {
		return domain.DefaultLCOMRequest()
	}

	return &domain.LCOMRequest{
		OutputFormat:    domain.OutputFormat(pyscnCfg.Output.Format),
		ShowDetails:     domain.BoolPtr(domain.BoolValue(pyscnCfg.Output.ShowDetails, false)),
		LowThreshold:    pyscnCfg.LcomLowThreshold,
		MediumThreshold: pyscnCfg.LcomMediumThreshold,
		MinLCOM:         0,
		MaxLCOM:         0,
		SortBy:          domain.SortByCohesion,
		Recursive:       pyscnCfg.AnalysisRecursive,
		IncludePatterns: pyscnCfg.AnalysisIncludePatterns,
		ExcludePatterns: pyscnCfg.AnalysisExcludePatterns,
	}
}
