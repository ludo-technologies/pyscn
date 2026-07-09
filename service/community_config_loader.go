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

	merged.Paths = config.MergeSlice(merged.Paths, override.Paths)
	merged.SourcePaths = config.MergeSlice(merged.SourcePaths, override.SourcePaths)
	merged.OutputFormat = config.Merge(merged.OutputFormat, override.OutputFormat)
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	merged.OutputPath = config.Merge(merged.OutputPath, override.OutputPath)
	merged.NoOpen = override.NoOpen

	merged.ConfigPath = config.Merge(merged.ConfigPath, override.ConfigPath)
	merged.Recursive = config.MergePtr(merged.Recursive, override.Recursive)
	merged.IncludePatterns = config.MergeSlice(merged.IncludePatterns, override.IncludePatterns)
	merged.ExcludePatterns = config.MergeSlice(merged.ExcludePatterns, override.ExcludePatterns)

	merged.Algorithm = config.Merge(merged.Algorithm, override.Algorithm)
	merged.Scope = config.Merge(merged.Scope, override.Scope)
	merged.MinCommunitySize = config.Merge(merged.MinCommunitySize, override.MinCommunitySize)
	merged.IncludeLazyEdges = config.MergePtr(merged.IncludeLazyEdges, override.IncludeLazyEdges)
	merged.ReportBridgeModules = config.MergePtr(merged.ReportBridgeModules, override.ReportBridgeModules)
	merged.Resolution = config.Merge(merged.Resolution, override.Resolution)

	merged.IncludeStdLib = config.MergePtr(merged.IncludeStdLib, override.IncludeStdLib)
	merged.IncludeThirdParty = config.MergePtr(merged.IncludeThirdParty, override.IncludeThirdParty)
	merged.FollowRelative = config.MergePtr(merged.FollowRelative, override.FollowRelative)

	merged.ArchitectureRules = config.MergePtr(merged.ArchitectureRules, override.ArchitectureRules)

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
