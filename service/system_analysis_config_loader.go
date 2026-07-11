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
	if cfg.ArchitectureValidateCohesion != nil {
		request.ValidateCohesion = cfg.ArchitectureValidateCohesion
	}
	if cfg.ArchitectureValidateResponsibility != nil {
		request.ValidateResponsibility = cfg.ArchitectureValidateResponsibility
	}
	if cfg.ArchitectureMinCohesion > 0 {
		request.MinCohesion = cfg.ArchitectureMinCohesion
	}
	if cfg.ArchitectureMaxResponsibilities > 0 {
		request.MaxResponsibilities = cfg.ArchitectureMaxResponsibilities
	}
	if cfg.ArchitectureCohesionViolationSeverity != "" {
		request.CohesionViolationSeverity = parseViolationSeverity(cfg.ArchitectureCohesionViolationSeverity)
	}
	if cfg.ArchitectureResponsibilityViolationSeverity != "" {
		request.ResponsibilityViolationSeverity = parseViolationSeverity(cfg.ArchitectureResponsibilityViolationSeverity)
	}

	// Architecture settings
	if rules := ArchitectureRulesFromPyscnConfig(cfg); rules != nil {
		request.ArchitectureRules = rules
	}

	// Analysis settings (include/exclude patterns)
	if cfg.HasExplicitAnalysisIncludePatterns() {
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
		OutputFormat:                    domain.OutputFormatText,
		AnalyzeDependencies:             domain.BoolPtr(true),
		AnalyzeArchitecture:             domain.BoolPtr(true),
		IncludeStdLib:                   domain.BoolPtr(false),
		IncludeThirdParty:               domain.BoolPtr(true),
		FollowRelative:                  domain.BoolPtr(true),
		DetectCycles:                    domain.BoolPtr(true),
		ValidateArchitecture:            domain.BoolPtr(true),
		ValidateCohesion:                domain.BoolPtr(true),
		ValidateResponsibility:          domain.BoolPtr(true),
		MinCohesion:                     domain.DefaultArchitectureMinCohesion,
		MaxResponsibilities:             domain.DefaultArchitectureMaxResponsibilities,
		CohesionViolationSeverity:       domain.ViolationSeverityWarning,
		ResponsibilityViolationSeverity: domain.ViolationSeverityWarning,
		Recursive:                       domain.BoolPtr(true),
		IncludePatterns:                 domain.DefaultPythonModuleIncludePatterns(),
		ExcludePatterns:                 domain.DefaultAnalysisExcludePatterns(),
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

	// Override with CLI flags (zero values mean "not set")
	merged.Paths = config.MergeSlice(merged.Paths, override.Paths)
	merged.OutputFormat = config.Merge(merged.OutputFormat, override.OutputFormat)
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	merged.OutputPath = config.Merge(merged.OutputPath, override.OutputPath)
	merged.ConfigPath = config.Merge(merged.ConfigPath, override.ConfigPath)

	// Boolean flags - CLI always takes precedence for explicit settings
	merged.NoOpen = override.NoOpen

	// Analysis type overrides - only override if explicitly set (non-nil)
	merged.AnalyzeDependencies = config.MergePtr(merged.AnalyzeDependencies, override.AnalyzeDependencies)
	merged.AnalyzeArchitecture = config.MergePtr(merged.AnalyzeArchitecture, override.AnalyzeArchitecture)

	// Analysis options - only override if explicitly set (non-nil)
	merged.IncludeStdLib = config.MergePtr(merged.IncludeStdLib, override.IncludeStdLib)
	merged.IncludeThirdParty = config.MergePtr(merged.IncludeThirdParty, override.IncludeThirdParty)
	merged.FollowRelative = config.MergePtr(merged.FollowRelative, override.FollowRelative)
	merged.DetectCycles = config.MergePtr(merged.DetectCycles, override.DetectCycles)
	merged.ValidateArchitecture = config.MergePtr(merged.ValidateArchitecture, override.ValidateArchitecture)
	merged.ValidateCohesion = config.MergePtr(merged.ValidateCohesion, override.ValidateCohesion)
	merged.ValidateResponsibility = config.MergePtr(merged.ValidateResponsibility, override.ValidateResponsibility)
	merged.MinCohesion = config.Merge(merged.MinCohesion, override.MinCohesion)
	merged.MaxResponsibilities = config.Merge(merged.MaxResponsibilities, override.MaxResponsibilities)
	merged.CohesionViolationSeverity = config.Merge(merged.CohesionViolationSeverity, override.CohesionViolationSeverity)
	merged.ResponsibilityViolationSeverity = config.Merge(merged.ResponsibilityViolationSeverity, override.ResponsibilityViolationSeverity)

	// File selection
	merged.IncludePatterns = config.MergeSlice(merged.IncludePatterns, override.IncludePatterns)
	merged.ExcludePatterns = config.MergeSlice(merged.ExcludePatterns, override.ExcludePatterns)
	merged.Recursive = config.MergePtr(merged.Recursive, override.Recursive)

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
			if override.ArchitectureRules.Style != "" {
				merged.ArchitectureRules.Style = override.ArchitectureRules.Style
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
			if len(override.ArchitectureRules.NeutralPrefixes) > 0 {
				merged.ArchitectureRules.NeutralPrefixes = override.ArchitectureRules.NeutralPrefixes
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

// convertLayerDefinitions converts config.LayerDefinition slice to domain.Layer slice.
func convertLayerDefinitions(layers []config.LayerDefinition) []domain.Layer {
	out := make([]domain.Layer, len(layers))
	for i, l := range layers {
		out[i] = domain.Layer{
			Name:        l.Name,
			Packages:    l.Packages,
			Description: l.Description,
		}
	}
	return out
}

// convertLayerRules converts config.LayerRule slice to domain.LayerRule slice.
func convertLayerRules(rules []config.LayerRule) []domain.LayerRule {
	out := make([]domain.LayerRule, len(rules))
	for i, r := range rules {
		out[i] = domain.LayerRule{
			From:  r.From,
			Allow: r.Allow,
			Deny:  r.Deny,
			Warn:  r.Warn,
		}
	}
	return out
}

// Example configuration file content for documentation
var ExampleSystemAnalysisConfig = `
# System Analysis Configuration

[system_analysis]
analyze_dependencies = true
analyze_architecture = true
recursive = true
include_patterns = ["**/*.py"]
exclude_patterns = ["test_*.py", "*_test.py", "**/migrations/**"]

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
