package service

import (
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// SystemAnalysisConfigurationLoaderImpl implements the SystemAnalysisConfigurationLoader interface
type SystemAnalysisConfigurationLoaderImpl struct{}

// NewSystemAnalysisConfigurationLoader creates a new system analysis configuration loader
func NewSystemAnalysisConfigurationLoader() *SystemAnalysisConfigurationLoaderImpl {
	return &SystemAnalysisConfigurationLoaderImpl{}
}

// LoadConfig loads configuration from the specified path using the shared TomlConfigLoader
func (cl *SystemAnalysisConfigurationLoaderImpl) LoadConfig(path string) (*domain.SystemAnalysisRequest, error) {
	// Use the shared TomlConfigLoader to load configuration
	tomlLoader := config.NewTomlConfigLoader()
	pyscnCfg, err := tomlLoader.LoadConfig(path)
	if err != nil {
		return nil, err
	}

	// Convert PyscnConfig to SystemAnalysisRequest
	return cl.pyscnConfigToSystemAnalysisRequest(pyscnCfg), nil
}

// pyscnConfigToSystemAnalysisRequest converts PyscnConfig to SystemAnalysisRequest
func (cl *SystemAnalysisConfigurationLoaderImpl) pyscnConfigToSystemAnalysisRequest(cfg *config.PyscnConfig) *domain.SystemAnalysisRequest {
	request := cl.LoadDefaultConfig()

	// System analysis settings
	if cfg.SystemAnalysisEnableDependencies != nil {
		request.AnalyzeDependencies = cfg.SystemAnalysisEnableDependencies
	}
	if cfg.SystemAnalysisEnableArchitecture != nil {
		request.AnalyzeArchitecture = cfg.SystemAnalysisEnableArchitecture
	}

	// Dependencies settings
	if cfg.DependenciesIncludeStdLib != nil {
		request.IncludeStdLib = cfg.DependenciesIncludeStdLib
	}
	if cfg.DependenciesIncludeThirdParty != nil {
		request.IncludeThirdParty = cfg.DependenciesIncludeThirdParty
	}
	if cfg.DependenciesFollowRelative != nil {
		request.FollowRelative = cfg.DependenciesFollowRelative
	}
	if cfg.DependenciesDetectCycles != nil {
		request.DetectCycles = cfg.DependenciesDetectCycles
	}

	// Architecture settings
	if cfg.ArchitectureStrictMode != nil || len(cfg.ArchitectureAllowedPatterns) > 0 || len(cfg.ArchitectureForbiddenPatterns) > 0 {
		if request.ArchitectureRules == nil {
			request.ArchitectureRules = &domain.ArchitectureRules{}
		}
		if cfg.ArchitectureStrictMode != nil {
			request.ArchitectureRules.StrictMode = *cfg.ArchitectureStrictMode
		}
		if len(cfg.ArchitectureAllowedPatterns) > 0 {
			request.ArchitectureRules.AllowedPatterns = cfg.ArchitectureAllowedPatterns
		}
		if len(cfg.ArchitectureForbiddenPatterns) > 0 {
			request.ArchitectureRules.ForbiddenPatterns = cfg.ArchitectureForbiddenPatterns
		}
	}

	// Analysis settings (include/exclude patterns)
	if len(cfg.AnalysisIncludePatterns) > 0 {
		request.IncludePatterns = cfg.AnalysisIncludePatterns
	}
	if len(cfg.AnalysisExcludePatterns) > 0 {
		request.ExcludePatterns = cfg.AnalysisExcludePatterns
	}
	if cfg.AnalysisRecursive != nil {
		request.Recursive = cfg.AnalysisRecursive
	}

	return request
}

// LoadDefaultConfig loads the default configuration
func (cl *SystemAnalysisConfigurationLoaderImpl) LoadDefaultConfig() *domain.SystemAnalysisRequest {
	return &domain.SystemAnalysisRequest{
		OutputFormat:         domain.OutputFormatText,
		AnalyzeDependencies:  domain.BoolPtr(true),
		AnalyzeArchitecture:  domain.BoolPtr(true),
		IncludeStdLib:        domain.BoolPtr(false),
		IncludeThirdParty:    domain.BoolPtr(true),
		FollowRelative:       domain.BoolPtr(true),
		DetectCycles:         domain.BoolPtr(true),
		ValidateArchitecture: domain.BoolPtr(true),
		Recursive:            domain.BoolPtr(true),
		IncludePatterns:      []string{"**/*.py"},
		ExcludePatterns:      []string{},
	}
}

// MergeConfig merges CLI flags with configuration file
func (cl *SystemAnalysisConfigurationLoaderImpl) MergeConfig(base *domain.SystemAnalysisRequest, override *domain.SystemAnalysisRequest) *domain.SystemAnalysisRequest {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}

	// Start with base configuration
	merged := *base

	// Override with CLI flags (non-zero values take precedence)
	if len(override.Paths) > 0 {
		merged.Paths = override.Paths
	}
	if override.OutputFormat != "" && override.OutputFormat != domain.OutputFormatText {
		merged.OutputFormat = override.OutputFormat
	}
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	if override.OutputPath != "" {
		merged.OutputPath = override.OutputPath
	}
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// Boolean flags - CLI always takes precedence for explicit settings
	merged.NoOpen = override.NoOpen

	// Analysis type overrides - only override if explicitly set (non-nil)
	if override.AnalyzeDependencies != nil {
		merged.AnalyzeDependencies = override.AnalyzeDependencies
	}
	if override.AnalyzeArchitecture != nil {
		merged.AnalyzeArchitecture = override.AnalyzeArchitecture
	}

	// Analysis options - only override if explicitly set (non-nil)
	if override.IncludeStdLib != nil {
		merged.IncludeStdLib = override.IncludeStdLib
	}
	if override.IncludeThirdParty != nil {
		merged.IncludeThirdParty = override.IncludeThirdParty
	}
	if override.FollowRelative != nil {
		merged.FollowRelative = override.FollowRelative
	}
	if override.DetectCycles != nil {
		merged.DetectCycles = override.DetectCycles
	}
	if override.ValidateArchitecture != nil {
		merged.ValidateArchitecture = override.ValidateArchitecture
	}

	// File selection - override if provided
	if len(override.IncludePatterns) > 0 {
		merged.IncludePatterns = override.IncludePatterns
	}
	if len(override.ExcludePatterns) > 0 {
		merged.ExcludePatterns = override.ExcludePatterns
	}
	if override.Recursive != nil {
		merged.Recursive = override.Recursive
	}

	// Architecture rules - merge carefully to preserve config while applying CLI overrides
	if override.ArchitectureRules != nil {
		if merged.ArchitectureRules == nil {
			// No config rules, use override as-is
			merged.ArchitectureRules = override.ArchitectureRules
		} else {
			// Merge: apply StrictMode from CLI while preserving config rules
			if override.ArchitectureRules.StrictMode {
				merged.ArchitectureRules.StrictMode = true
			}
			// If CLI provides layers/rules, they override config (unlikely in deps command)
			if len(override.ArchitectureRules.Layers) > 0 {
				merged.ArchitectureRules.Layers = override.ArchitectureRules.Layers
			}
			if len(override.ArchitectureRules.Rules) > 0 {
				merged.ArchitectureRules.Rules = override.ArchitectureRules.Rules
			}
			if len(override.ArchitectureRules.AllowedPatterns) > 0 {
				merged.ArchitectureRules.AllowedPatterns = override.ArchitectureRules.AllowedPatterns
			}
			if len(override.ArchitectureRules.ForbiddenPatterns) > 0 {
				merged.ArchitectureRules.ForbiddenPatterns = override.ArchitectureRules.ForbiddenPatterns
			}
		}
	}

	return &merged
}

// SystemAnalysisConfigurationLoaderWithFlags extends the base loader with CLI flag integration
type SystemAnalysisConfigurationLoaderWithFlags struct {
	*SystemAnalysisConfigurationLoaderImpl
}

// NewSystemAnalysisConfigurationLoaderWithFlags creates a configuration loader that integrates CLI flags
func NewSystemAnalysisConfigurationLoaderWithFlags() *SystemAnalysisConfigurationLoaderWithFlags {
	return &SystemAnalysisConfigurationLoaderWithFlags{
		SystemAnalysisConfigurationLoaderImpl: NewSystemAnalysisConfigurationLoader(),
	}
}

// LoadConfigWithFlags loads configuration and merges with CLI flags
func (cl *SystemAnalysisConfigurationLoaderWithFlags) LoadConfigWithFlags(
	configPath string,
	cliRequest *domain.SystemAnalysisRequest,
) (*domain.SystemAnalysisRequest, error) {
	// Load base configuration
	baseConfig, err := cl.LoadConfig(configPath)
	if err != nil {
		return nil, err
	}

	// Merge with CLI flags
	mergedConfig := cl.MergeConfig(baseConfig, cliRequest)
	return mergedConfig, nil
}

// Example configuration file content for documentation
var ExampleSystemAnalysisConfig = `
# System Analysis Configuration

[system_analysis]
analyze_dependencies = true
analyze_architecture = true
recursive = true
include_patterns = ["**/*.py"]
exclude_patterns = ["test_*.py", "*_test.py"]

[dependencies]
include_stdlib = false
include_third_party = true
follow_relative = true
detect_cycles = true

[architecture]
strict_mode = false
allowed_patterns = []
forbidden_patterns = []
`
