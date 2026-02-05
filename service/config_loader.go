package service

import (
	"os"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// ConfigurationLoaderImpl implements the ConfigurationLoader interface
type ConfigurationLoaderImpl struct{}

// NewConfigurationLoader creates a new configuration loader service
func NewConfigurationLoader() *ConfigurationLoaderImpl {
	return &ConfigurationLoaderImpl{}
}

// LoadConfig loads configuration from the specified path
func (c *ConfigurationLoaderImpl) LoadConfig(path string) (*domain.ComplexityRequest, error) {
	// Use TOML-only loader
	tomlLoader := config.NewTomlConfigLoader()
	pyscnCfg, err := tomlLoader.LoadConfig(path)
	if err != nil {
		return nil, domain.NewConfigError("failed to load configuration file", err)
	}

	// Convert pyscn config to unified config format, then to complexity request
	cfg := c.pyscnConfigToUnifiedConfig(pyscnCfg)
	return c.convertToComplexityRequest(cfg), nil
}

// LoadDefaultConfig loads the default configuration, first checking for .pyscn.toml
func (c *ConfigurationLoaderImpl) LoadDefaultConfig() *domain.ComplexityRequest {
	// First, try to find and load a config file in the current directory
	configFile := c.FindDefaultConfigFile()
	if configFile != "" {
		if configReq, err := c.LoadConfig(configFile); err == nil {
			return configReq
		}
		// If loading failed, fall back to hardcoded defaults
	}

	// Fall back to hardcoded default configuration
	cfg := config.DefaultConfig()
	return c.convertToComplexityRequest(cfg)
}

// MergeConfig merges CLI flags with configuration file
func (c *ConfigurationLoaderImpl) MergeConfig(base *domain.ComplexityRequest, override *domain.ComplexityRequest) *domain.ComplexityRequest {
	// Start with base configuration
	merged := *base

	// Override with non-zero values from override
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

	// Only override if values differ from defaults
	if override.ShowDetails {
		merged.ShowDetails = override.ShowDetails
	}

	// Filtering and sorting - override if non-default
	if override.MinComplexity != 1 {
		merged.MinComplexity = override.MinComplexity
	}

	if override.MaxComplexity != 0 {
		merged.MaxComplexity = override.MaxComplexity
	}

	if override.SortBy != "" && override.SortBy != "complexity" {
		merged.SortBy = override.SortBy
	}

	// Complexity thresholds - override if non-default
	if override.LowThreshold != domain.DefaultComplexityLowThreshold && override.LowThreshold > 0 {
		merged.LowThreshold = override.LowThreshold
	}

	if override.MediumThreshold != domain.DefaultComplexityMediumThreshold && override.MediumThreshold > 0 {
		merged.MediumThreshold = override.MediumThreshold
	}

	// Config path is always from override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// For recursive, preserve the override value
	merged.Recursive = override.Recursive

	// Patterns - override if provided and different from defaults
	if len(override.IncludePatterns) > 0 {
		merged.IncludePatterns = override.IncludePatterns
	}

	if len(override.ExcludePatterns) > 0 {
		merged.ExcludePatterns = override.ExcludePatterns
	}

	return &merged
}

// convertToComplexityRequest converts internal config to domain request
func (c *ConfigurationLoaderImpl) convertToComplexityRequest(cfg *config.Config) *domain.ComplexityRequest {
	// Convert output format
	var outputFormat domain.OutputFormat
	switch cfg.Output.Format {
	case "json":
		outputFormat = domain.OutputFormatJSON
	case "yaml":
		outputFormat = domain.OutputFormatYAML
	case "csv":
		outputFormat = domain.OutputFormatCSV
	case "html":
		outputFormat = domain.OutputFormatHTML
	default:
		outputFormat = domain.OutputFormatText
	}

	// Convert sort criteria
	var sortBy domain.SortCriteria
	switch cfg.Output.SortBy {
	case "name":
		sortBy = domain.SortByName
	case "risk":
		sortBy = domain.SortByRisk
	default:
		sortBy = domain.SortByComplexity
	}

	return &domain.ComplexityRequest{
		OutputFormat:    outputFormat,
		OutputWriter:    os.Stdout, // Default to stdout
		ShowDetails:     cfg.Output.ShowDetails,
		MinComplexity:   cfg.Output.MinComplexity,
		MaxComplexity:   cfg.Complexity.MaxComplexity,
		SortBy:          sortBy,
		LowThreshold:    cfg.Complexity.LowThreshold,
		MediumThreshold: cfg.Complexity.MediumThreshold,
		Recursive:       cfg.Analysis.Recursive,
		IncludePatterns: cfg.Analysis.IncludePatterns,
		ExcludePatterns: cfg.Analysis.ExcludePatterns,
	}
}

// ValidateConfig validates a configuration request
func (c *ConfigurationLoaderImpl) ValidateConfig(req *domain.ComplexityRequest) error {
	if req.LowThreshold <= 0 {
		return domain.NewConfigError("low threshold must be positive", nil)
	}

	if req.MediumThreshold <= req.LowThreshold {
		return domain.NewConfigError("medium threshold must be greater than low threshold", nil)
	}

	if req.MaxComplexity > 0 && req.MaxComplexity <= req.MediumThreshold {
		return domain.NewConfigError("max complexity must be greater than medium threshold or 0 for no limit", nil)
	}

	if req.MinComplexity < 0 {
		return domain.NewConfigError("minimum complexity cannot be negative", nil)
	}

	if req.MaxComplexity > 0 && req.MinComplexity > req.MaxComplexity {
		return domain.NewConfigError("minimum complexity cannot be greater than maximum complexity", nil)
	}

	return nil
}

// GetDefaultThresholds returns the default complexity thresholds
func (c *ConfigurationLoaderImpl) GetDefaultThresholds() (low, medium int) {
	return config.DefaultLowComplexityThreshold, config.DefaultMediumComplexityThreshold
}

// CreateConfigTemplate creates a template configuration file
func (c *ConfigurationLoaderImpl) CreateConfigTemplate(path string) error {
	cfg := config.DefaultConfig()
	return config.SaveConfig(cfg, path)
}

// FindDefaultConfigFile looks for TOML config files from the current directory upward.
func (c *ConfigurationLoaderImpl) FindDefaultConfigFile() string {
	tomlLoader := config.NewTomlConfigLoader()
	return tomlLoader.FindConfigFileFromPath("")
}

// pyscnConfigToUnifiedConfig converts PyscnConfig to unified Config format
func (c *ConfigurationLoaderImpl) pyscnConfigToUnifiedConfig(pyscnCfg *config.PyscnConfig) *config.Config {
	cfg := config.DefaultConfig()

	// Map clone detection settings (backward compatibility)
	cfg.Analysis.IncludePatterns = pyscnCfg.Input.IncludePatterns
	cfg.Analysis.ExcludePatterns = pyscnCfg.Input.ExcludePatterns
	cfg.Analysis.Recursive = domain.BoolValue(pyscnCfg.Input.Recursive, true)

	// Map clone output settings (backward compatibility)
	cfg.Output.Format = pyscnCfg.Output.Format
	cfg.Output.ShowDetails = domain.BoolValue(pyscnCfg.Output.ShowDetails, false)

	// Map complexity settings from [complexity] section
	cfg.Complexity.LowThreshold = pyscnCfg.ComplexityLowThreshold
	cfg.Complexity.MediumThreshold = pyscnCfg.ComplexityMediumThreshold
	cfg.Complexity.MaxComplexity = pyscnCfg.ComplexityMaxComplexity
	cfg.Output.MinComplexity = pyscnCfg.ComplexityMinComplexity

	// Map dead code settings from [dead_code] section
	cfg.DeadCode.Enabled = domain.BoolValue(pyscnCfg.DeadCodeEnabled, true)
	cfg.DeadCode.MinSeverity = pyscnCfg.DeadCodeMinSeverity
	cfg.DeadCode.ShowContext = domain.BoolValue(pyscnCfg.DeadCodeShowContext, false)
	cfg.DeadCode.ContextLines = pyscnCfg.DeadCodeContextLines
	cfg.DeadCode.SortBy = pyscnCfg.DeadCodeSortBy
	cfg.DeadCode.DetectAfterReturn = domain.BoolValue(pyscnCfg.DeadCodeDetectAfterReturn, true)
	cfg.DeadCode.DetectAfterBreak = domain.BoolValue(pyscnCfg.DeadCodeDetectAfterBreak, true)
	cfg.DeadCode.DetectAfterContinue = domain.BoolValue(pyscnCfg.DeadCodeDetectAfterContinue, true)
	cfg.DeadCode.DetectAfterRaise = domain.BoolValue(pyscnCfg.DeadCodeDetectAfterRaise, true)
	cfg.DeadCode.DetectUnreachableBranches = domain.BoolValue(pyscnCfg.DeadCodeDetectUnreachableBranches, true)
	cfg.DeadCode.IgnorePatterns = pyscnCfg.DeadCodeIgnorePatterns

	// Map general output settings from [output] section (override clone-specific if set)
	if pyscnCfg.OutputFormat != "" {
		cfg.Output.Format = pyscnCfg.OutputFormat
	}
	if pyscnCfg.OutputSortBy != "" {
		cfg.Output.SortBy = pyscnCfg.OutputSortBy
	}
	if pyscnCfg.OutputDirectory != "" {
		cfg.Output.Directory = pyscnCfg.OutputDirectory
	}
	// Override if explicitly set
	if pyscnCfg.OutputShowDetails != nil {
		cfg.Output.ShowDetails = *pyscnCfg.OutputShowDetails
	}
	if pyscnCfg.OutputMinComplexity > 0 {
		cfg.Output.MinComplexity = pyscnCfg.OutputMinComplexity
	}

	// Map general analysis settings from [analysis] section (override clone-specific if set)
	if len(pyscnCfg.AnalysisIncludePatterns) > 0 {
		cfg.Analysis.IncludePatterns = pyscnCfg.AnalysisIncludePatterns
	}
	if len(pyscnCfg.AnalysisExcludePatterns) > 0 {
		cfg.Analysis.ExcludePatterns = pyscnCfg.AnalysisExcludePatterns
	}
	// Override if explicitly set
	if pyscnCfg.AnalysisRecursive != nil {
		cfg.Analysis.Recursive = *pyscnCfg.AnalysisRecursive
	}

	if pyscnCfg.AnalysisFollowSymlinks != nil {
		cfg.Analysis.FollowSymlinks = *pyscnCfg.AnalysisFollowSymlinks
	}

	// Map architecture settings from [architecture] section
	cfg.Architecture.Enabled = domain.BoolValue(pyscnCfg.ArchitectureEnabled, false)
	cfg.Architecture.ValidateLayers = domain.BoolValue(pyscnCfg.ArchitectureValidateLayers, true)
	cfg.Architecture.ValidateCohesion = domain.BoolValue(pyscnCfg.ArchitectureValidateCohesion, true)
	cfg.Architecture.ValidateResponsibility = domain.BoolValue(pyscnCfg.ArchitectureValidateResponsibility, true)
	cfg.Architecture.MinCohesion = pyscnCfg.ArchitectureMinCohesion
	cfg.Architecture.MaxCoupling = pyscnCfg.ArchitectureMaxCoupling
	cfg.Architecture.MaxResponsibilities = pyscnCfg.ArchitectureMaxResponsibilities
	cfg.Architecture.LayerViolationSeverity = pyscnCfg.ArchitectureLayerViolationSeverity
	cfg.Architecture.CohesionViolationSeverity = pyscnCfg.ArchitectureCohesionViolationSeverity
	cfg.Architecture.ResponsibilityViolationSeverity = pyscnCfg.ArchitectureResponsibilityViolationSeverity
	cfg.Architecture.ShowAllViolations = domain.BoolValue(pyscnCfg.ArchitectureShowAllViolations, true)
	cfg.Architecture.GroupByType = domain.BoolValue(pyscnCfg.ArchitectureGroupByType, true)
	cfg.Architecture.IncludeSuggestions = domain.BoolValue(pyscnCfg.ArchitectureIncludeSuggestions, true)
	cfg.Architecture.MaxViolationsToShow = pyscnCfg.ArchitectureMaxViolationsToShow
	cfg.Architecture.CustomPatterns = pyscnCfg.ArchitectureCustomPatterns
	cfg.Architecture.AllowedPatterns = pyscnCfg.ArchitectureAllowedPatterns
	cfg.Architecture.ForbiddenPatterns = pyscnCfg.ArchitectureForbiddenPatterns
	cfg.Architecture.StrictMode = domain.BoolValue(pyscnCfg.ArchitectureStrictMode, true)
	cfg.Architecture.FailOnViolations = domain.BoolValue(pyscnCfg.ArchitectureFailOnViolations, false)

	// Map system analysis settings from [system_analysis] section
	cfg.SystemAnalysis.Enabled = domain.BoolValue(pyscnCfg.SystemAnalysisEnabled, false)
	cfg.SystemAnalysis.EnableDependencies = domain.BoolValue(pyscnCfg.SystemAnalysisEnableDependencies, true)
	cfg.SystemAnalysis.EnableArchitecture = domain.BoolValue(pyscnCfg.SystemAnalysisEnableArchitecture, true)
	cfg.SystemAnalysis.UseComplexityData = domain.BoolValue(pyscnCfg.SystemAnalysisUseComplexityData, false)
	cfg.SystemAnalysis.UseClonesData = domain.BoolValue(pyscnCfg.SystemAnalysisUseClonesData, false)
	cfg.SystemAnalysis.UseDeadCodeData = domain.BoolValue(pyscnCfg.SystemAnalysisUseDeadCodeData, false)
	cfg.SystemAnalysis.GenerateUnifiedReport = domain.BoolValue(pyscnCfg.SystemAnalysisGenerateUnifiedReport, true)

	// Map dependencies settings from [dependencies] section
	cfg.Dependencies.Enabled = domain.BoolValue(pyscnCfg.DependenciesEnabled, false)
	cfg.Dependencies.IncludeStdLib = domain.BoolValue(pyscnCfg.DependenciesIncludeStdLib, false)
	cfg.Dependencies.IncludeThirdParty = domain.BoolValue(pyscnCfg.DependenciesIncludeThirdParty, true)
	cfg.Dependencies.FollowRelative = domain.BoolValue(pyscnCfg.DependenciesFollowRelative, true)
	cfg.Dependencies.DetectCycles = domain.BoolValue(pyscnCfg.DependenciesDetectCycles, true)
	cfg.Dependencies.CalculateMetrics = domain.BoolValue(pyscnCfg.DependenciesCalculateMetrics, true)
	cfg.Dependencies.FindLongChains = domain.BoolValue(pyscnCfg.DependenciesFindLongChains, true)
	cfg.Dependencies.MinCoupling = pyscnCfg.DependenciesMinCoupling
	cfg.Dependencies.MaxCoupling = pyscnCfg.DependenciesMaxCoupling
	cfg.Dependencies.MinInstability = pyscnCfg.DependenciesMinInstability
	cfg.Dependencies.MaxDistance = pyscnCfg.DependenciesMaxDistance
	cfg.Dependencies.SortBy = pyscnCfg.DependenciesSortBy
	cfg.Dependencies.ShowMatrix = domain.BoolValue(pyscnCfg.DependenciesShowMatrix, true)
	cfg.Dependencies.ShowMetrics = domain.BoolValue(pyscnCfg.DependenciesShowMetrics, true)
	cfg.Dependencies.ShowChains = domain.BoolValue(pyscnCfg.DependenciesShowChains, true)
	cfg.Dependencies.GenerateDotGraph = domain.BoolValue(pyscnCfg.DependenciesGenerateDotGraph, false)
	cfg.Dependencies.CycleReporting = pyscnCfg.DependenciesCycleReporting
	cfg.Dependencies.MaxCyclesToShow = pyscnCfg.DependenciesMaxCyclesToShow
	cfg.Dependencies.ShowCyclePaths = domain.BoolValue(pyscnCfg.DependenciesShowCyclePaths, true)

	// Keep the clone config reference for backward compatibility
	cfg.Clones = pyscnCfg

	return cfg
}
