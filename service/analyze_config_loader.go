package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/pelletier/go-toml/v2"
)

// AnalyzeConfigurationLoaderImpl resolves and loads config for AnalyzeUseCase.
type AnalyzeConfigurationLoaderImpl struct{}

// NewAnalyzeConfigurationLoader creates a new analyze configuration loader.
func NewAnalyzeConfigurationLoader() *AnalyzeConfigurationLoaderImpl {
	return &AnalyzeConfigurationLoaderImpl{}
}

// LoadAnalyzeExecutionConfig resolves the effective config path and loads the
// AnalyzeUseCase-specific settings it needs.
func (l *AnalyzeConfigurationLoaderImpl) LoadAnalyzeExecutionConfig(configPath string, targetPath string) (domain.AnalyzeExecutionConfig, error) {
	tomlLoader := config.NewTomlConfigLoader()
	resolvedConfigPath, err := tomlLoader.ResolveConfigPath(configPath, targetPath)
	if err != nil {
		return domain.AnalyzeExecutionConfig{}, fmt.Errorf("failed to resolve configuration: %w", err)
	}

	if resolvedConfigPath == "" {
		return defaultAnalyzeExecutionConfig(), nil
	}

	cfg, err := config.LoadConfig(resolvedConfigPath)
	if err != nil {
		return domain.AnalyzeExecutionConfig{}, fmt.Errorf("failed to load configuration: %w", err)
	}

	overrides, err := loadAnalyzeEnabledOverrides(resolvedConfigPath)
	if err != nil {
		return domain.AnalyzeExecutionConfig{}, fmt.Errorf("failed to load analyze enabled settings: %w", err)
	}

	executionCfg := analyzeExecutionConfigFromConfig(cfg, overrides)
	executionCfg.ConfigPath = resolvedConfigPath

	return executionCfg, nil
}

type analyzeEnabledOverrides struct {
	SystemEnabled             *bool
	SystemAnalyzeDependencies *bool
	SystemAnalyzeArchitecture *bool
	DependenciesEnabled       *bool
	ArchitectureEnabled       *bool
	CommunitiesEnabled        *bool
}

func defaultAnalyzeExecutionConfig() domain.AnalyzeExecutionConfig {
	defaultCfg := config.DefaultConfig()
	defaultCloneReq := domain.DefaultCloneRequest()

	return domain.AnalyzeExecutionConfig{
		ConfigPath:                   "",
		IncludePatterns:              domain.DefaultAnalysisIncludePatterns(),
		ExcludePatterns:              append([]string(nil), defaultCfg.Analysis.ExcludePatterns...),
		Recursive:                    defaultCfg.Analysis.Recursive,
		ShowDetails:                  defaultCfg.Output.ShowDetails,
		ComplexityEnabled:            defaultCfg.Complexity.Enabled,
		ComplexityReportUnchanged:    defaultCfg.Complexity.ReportUnchanged,
		ComplexityMinComplexity:      defaultCfg.Output.MinComplexity,
		ComplexityLowThreshold:       defaultCfg.Complexity.LowThreshold,
		ComplexityMediumThreshold:    defaultCfg.Complexity.MediumThreshold,
		ComplexityMaxComplexity:      defaultCfg.Complexity.MaxComplexity,
		CognitiveComplexityThreshold: defaultCfg.Complexity.CognitiveComplexityThreshold,
		NestingDepthThreshold:        defaultCfg.Complexity.NestingDepthThreshold,
		DeadCodeEnabled:              defaultCfg.DeadCode.Enabled,
		CloneLSHEnabled:              defaultCloneReq.LSHEnabled,
		CloneLSHAutoThreshold:        defaultCloneReq.LSHAutoThreshold,
		SystemEnabled:                true,
		SystemAnalyzeDependencies:    true,
		SystemAnalyzeArchitecture:    true,
		CommunitiesEnabled:           false,
		CommunitiesEnabledExplicit:   false,
	}
}

func analyzeExecutionConfigFromConfig(cfg *config.Config, overrides analyzeEnabledOverrides) domain.AnalyzeExecutionConfig {
	executionCfg := defaultAnalyzeExecutionConfig()
	if cfg == nil {
		return executionCfg
	}

	if len(cfg.Analysis.IncludePatterns) > 0 {
		executionCfg.IncludePatterns = append([]string(nil), cfg.Analysis.IncludePatterns...)
	}
	if len(cfg.Analysis.ExcludePatterns) > 0 {
		executionCfg.ExcludePatterns = append([]string(nil), cfg.Analysis.ExcludePatterns...)
	}

	executionCfg.Recursive = cfg.Analysis.Recursive
	executionCfg.ShowDetails = cfg.Output.ShowDetails
	executionCfg.ComplexityEnabled = cfg.Complexity.Enabled
	executionCfg.ComplexityReportUnchanged = cfg.Complexity.ReportUnchanged
	executionCfg.ComplexityMinComplexity = cfg.Output.MinComplexity
	executionCfg.ComplexityLowThreshold = cfg.Complexity.LowThreshold
	executionCfg.ComplexityMediumThreshold = cfg.Complexity.MediumThreshold
	executionCfg.ComplexityMaxComplexity = cfg.Complexity.MaxComplexity
	executionCfg.CognitiveComplexityThreshold = cfg.Complexity.CognitiveComplexityThreshold
	executionCfg.NestingDepthThreshold = cfg.Complexity.NestingDepthThreshold
	executionCfg.DeadCodeEnabled = cfg.DeadCode.Enabled

	if cfg.Clones != nil {
		if cfg.Clones.LSH.Enabled != "" {
			executionCfg.CloneLSHEnabled = cfg.Clones.LSH.Enabled
		}
		if cfg.Clones.LSH.AutoThreshold > 0 {
			executionCfg.CloneLSHAutoThreshold = cfg.Clones.LSH.AutoThreshold
		}
	}

	applySystemEnabledOverrides(&executionCfg, overrides)
	applyCommunitiesEnabledOverrides(&executionCfg, overrides, cfg)

	return executionCfg
}

func applyCommunitiesEnabledOverrides(executionCfg *domain.AnalyzeExecutionConfig, overrides analyzeEnabledOverrides, cfg *config.Config) {
	if overrides.CommunitiesEnabled != nil {
		executionCfg.CommunitiesEnabled = *overrides.CommunitiesEnabled
		executionCfg.CommunitiesEnabledExplicit = true
		return
	}
	if cfg != nil {
		executionCfg.CommunitiesEnabled = cfg.Communities.Enabled
	}
}

func applySystemEnabledOverrides(executionCfg *domain.AnalyzeExecutionConfig, overrides analyzeEnabledOverrides) {
	hasSystemOverride := overrides.SystemEnabled != nil ||
		overrides.SystemAnalyzeDependencies != nil ||
		overrides.SystemAnalyzeArchitecture != nil ||
		overrides.DependenciesEnabled != nil ||
		overrides.ArchitectureEnabled != nil
	if !hasSystemOverride {
		return
	}

	systemEnabled := true
	if overrides.SystemEnabled != nil {
		systemEnabled = *overrides.SystemEnabled
	}

	analyzeDependencies := systemEnabled
	if overrides.SystemAnalyzeDependencies != nil {
		analyzeDependencies = systemEnabled && *overrides.SystemAnalyzeDependencies
	}
	analyzeArchitecture := systemEnabled
	if overrides.SystemAnalyzeArchitecture != nil {
		analyzeArchitecture = systemEnabled && *overrides.SystemAnalyzeArchitecture
	}

	if overrides.DependenciesEnabled != nil {
		analyzeDependencies = *overrides.DependenciesEnabled
	}
	if overrides.ArchitectureEnabled != nil {
		analyzeArchitecture = *overrides.ArchitectureEnabled
	}

	executionCfg.SystemAnalyzeDependencies = analyzeDependencies
	executionCfg.SystemAnalyzeArchitecture = analyzeArchitecture
	executionCfg.SystemEnabled = analyzeDependencies || analyzeArchitecture
}

func loadAnalyzeEnabledOverrides(configPath string) (analyzeEnabledOverrides, error) {
	if configPath == "" {
		return analyzeEnabledOverrides{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return analyzeEnabledOverrides{}, err
	}

	if filepath.Base(configPath) == "pyproject.toml" {
		var parsed config.PyprojectToml
		if err := toml.Unmarshal(data, &parsed); err != nil {
			return analyzeEnabledOverrides{}, err
		}
		return enabledOverridesFromSections(
			parsed.Tool.Pyscn.SystemAnalysis,
			parsed.Tool.Pyscn.Dependencies,
			parsed.Tool.Pyscn.Architecture,
			parsed.Tool.Pyscn.Communities,
		), nil
	}

	var parsed config.PyscnTomlConfig
	if err := toml.Unmarshal(data, &parsed); err != nil {
		return analyzeEnabledOverrides{}, err
	}
	return enabledOverridesFromSections(parsed.SystemAnalysis, parsed.Dependencies, parsed.Architecture, parsed.Communities), nil
}

func enabledOverridesFromSections(system config.SystemAnalysisTomlConfig, dependencies config.DependenciesTomlConfig, architecture config.ArchitectureTomlConfig, communities config.CommunitiesTomlConfig) analyzeEnabledOverrides {
	return analyzeEnabledOverrides{
		SystemEnabled:             system.Enabled,
		SystemAnalyzeDependencies: system.EnableDependencies,
		SystemAnalyzeArchitecture: system.EnableArchitecture,
		DependenciesEnabled:       dependencies.Enabled,
		ArchitectureEnabled:       architecture.Enabled,
		CommunitiesEnabled:        communities.Enabled,
	}
}
