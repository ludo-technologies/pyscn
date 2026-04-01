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

	if len(override.Paths) > 0 {
		merged.Paths = override.Paths
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

	if override.MinLCOM > 0 {
		merged.MinLCOM = override.MinLCOM
	}
	if override.MaxLCOM > 0 {
		merged.MaxLCOM = override.MaxLCOM
	}
	if override.SortBy != "" {
		merged.SortBy = override.SortBy
	}
	if override.LowThreshold > 0 {
		merged.LowThreshold = override.LowThreshold
	}
	if override.MediumThreshold > 0 {
		merged.MediumThreshold = override.MediumThreshold
	}
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}
	if override.ShowDetails {
		merged.ShowDetails = true
	}
	if override.Recursive != nil {
		merged.Recursive = override.Recursive
	}
	if len(override.IncludePatterns) > 0 {
		merged.IncludePatterns = override.IncludePatterns
	}
	if len(override.ExcludePatterns) > 0 {
		merged.ExcludePatterns = override.ExcludePatterns
	}

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
		ShowDetails:     domain.BoolValue(pyscnCfg.Output.ShowDetails, false),
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
