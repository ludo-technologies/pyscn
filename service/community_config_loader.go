package service

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// CommunityConfigurationLoaderImpl implements domain.CommunityConfigurationLoader.
type CommunityConfigurationLoaderImpl struct{}

// NewCommunityConfigurationLoader creates a new community configuration loader.
func NewCommunityConfigurationLoader() *CommunityConfigurationLoaderImpl {
	return &CommunityConfigurationLoaderImpl{}
}

// LoadConfig loads community configuration from the specified path.
func (cl *CommunityConfigurationLoaderImpl) LoadConfig(path string) (*domain.CommunityAnalysisRequest, error) {
	tomlLoader := config.NewTomlConfigLoader()
	pyscnCfg, err := tomlLoader.LoadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	return cl.configToRequest(pyscnCfg), nil
}

// LoadDefaultConfig loads the default community configuration, checking for project config first.
func (cl *CommunityConfigurationLoaderImpl) LoadDefaultConfig() *domain.CommunityAnalysisRequest {
	configFile := cl.FindDefaultConfigFile()
	if configFile != "" {
		if configReq, err := cl.LoadConfig(configFile); err == nil {
			return configReq
		}
	}

	return cl.configToRequest(config.DefaultPyscnConfig())
}

// MergeConfig merges configuration file values with request overrides.
func (cl *CommunityConfigurationLoaderImpl) MergeConfig(base *domain.CommunityAnalysisRequest, override *domain.CommunityAnalysisRequest) *domain.CommunityAnalysisRequest {
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
	if len(override.SourcePaths) > 0 {
		merged.SourcePaths = override.SourcePaths
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

	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
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

	if override.Algorithm != "" {
		merged.Algorithm = override.Algorithm
	}
	if override.Scope != "" {
		merged.Scope = override.Scope
	}
	if override.MinCommunitySize > 0 {
		merged.MinCommunitySize = override.MinCommunitySize
	}
	if override.IncludeLazyEdges != nil {
		merged.IncludeLazyEdges = override.IncludeLazyEdges
	}
	if override.ReportBridgeModules != nil {
		merged.ReportBridgeModules = override.ReportBridgeModules
	}
	if override.Resolution > 0 {
		merged.Resolution = override.Resolution
	}

	if override.IncludeStdLib != nil {
		merged.IncludeStdLib = override.IncludeStdLib
	}
	if override.IncludeThirdParty != nil {
		merged.IncludeThirdParty = override.IncludeThirdParty
	}
	if override.FollowRelative != nil {
		merged.FollowRelative = override.FollowRelative
	}

	if override.ArchitectureRules != nil {
		merged.ArchitectureRules = override.ArchitectureRules
	}

	return &merged
}

func (cl *CommunityConfigurationLoaderImpl) configToRequest(pyscnCfg *config.PyscnConfig) *domain.CommunityAnalysisRequest {
	if pyscnCfg == nil {
		return domain.DefaultCommunityAnalysisRequest()
	}

	req := domain.DefaultCommunityAnalysisRequest()
	req.Algorithm = pyscnCfg.CommunitiesAlgorithm
	req.Scope = pyscnCfg.CommunitiesScope
	req.MinCommunitySize = pyscnCfg.CommunitiesMinCommunitySize
	req.IncludeLazyEdges = pyscnCfg.CommunitiesIncludeLazyEdges
	req.ReportBridgeModules = pyscnCfg.CommunitiesReportBridgeModules
	req.Resolution = pyscnCfg.CommunitiesResolution
	req.Recursive = pyscnCfg.AnalysisRecursive
	req.IncludePatterns = pyscnCfg.AnalysisIncludePatterns
	req.ExcludePatterns = pyscnCfg.AnalysisExcludePatterns

	if req.Algorithm == "" {
		req.Algorithm = domain.DefaultCommunityAlgorithm
	}
	if req.Scope == "" {
		req.Scope = domain.DefaultCommunityScope
	}
	if req.MinCommunitySize <= 0 {
		req.MinCommunitySize = 2
	}
	if req.Resolution <= 0 {
		req.Resolution = domain.DefaultCommunityResolution
	}

	req.ArchitectureRules = ArchitectureRulesFromPyscnConfig(pyscnCfg)

	return req
}

// FindDefaultConfigFile looks for TOML config files from the current directory upward.
func (cl *CommunityConfigurationLoaderImpl) FindDefaultConfigFile() string {
	tomlLoader := config.NewTomlConfigLoader()
	return tomlLoader.FindConfigFileFromPath("")
}
