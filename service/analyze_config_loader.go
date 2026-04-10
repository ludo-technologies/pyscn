package service

import (
	"fmt"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// AnalyzeConfigurationLoaderImpl resolves and loads config for AnalyzeUseCase.
type AnalyzeConfigurationLoaderImpl struct{}

// NewAnalyzeConfigurationLoader creates a new analyze configuration loader.
func NewAnalyzeConfigurationLoader() *AnalyzeConfigurationLoaderImpl {
	return &AnalyzeConfigurationLoaderImpl{}
}

// analyze includes stub files by default because they participate in the same
// module surface as runtime Python files.
var defaultAnalyzeIncludePatterns = []string{"**/*.py", "*.pyi"}

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

	executionCfg := analyzeExecutionConfigFromConfig(cfg)
	executionCfg.ConfigPath = resolvedConfigPath

	return executionCfg, nil
}

func defaultAnalyzeExecutionConfig() domain.AnalyzeExecutionConfig {
	defaultCfg := config.DefaultConfig()
	defaultCloneReq := domain.DefaultCloneRequest()

	return domain.AnalyzeExecutionConfig{
		ConfigPath:                "",
		IncludePatterns:           append([]string(nil), defaultAnalyzeIncludePatterns...),
		ExcludePatterns:           append([]string(nil), defaultCfg.Analysis.ExcludePatterns...),
		Recursive:                 defaultCfg.Analysis.Recursive,
		ComplexityEnabled:         defaultCfg.Complexity.Enabled,
		ComplexityReportUnchanged: defaultCfg.Complexity.ReportUnchanged,
		ComplexityMinComplexity:   defaultCfg.Output.MinComplexity,
		ComplexityLowThreshold:    defaultCfg.Complexity.LowThreshold,
		ComplexityMediumThreshold: defaultCfg.Complexity.MediumThreshold,
		ComplexityMaxComplexity:   defaultCfg.Complexity.MaxComplexity,
		CloneLSHEnabled:           defaultCloneReq.LSHEnabled,
		CloneLSHAutoThreshold:     defaultCloneReq.LSHAutoThreshold,
	}
}

func analyzeExecutionConfigFromConfig(cfg *config.Config) domain.AnalyzeExecutionConfig {
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
	executionCfg.ComplexityEnabled = cfg.Complexity.Enabled
	executionCfg.ComplexityReportUnchanged = cfg.Complexity.ReportUnchanged
	executionCfg.ComplexityMinComplexity = cfg.Output.MinComplexity
	executionCfg.ComplexityLowThreshold = cfg.Complexity.LowThreshold
	executionCfg.ComplexityMediumThreshold = cfg.Complexity.MediumThreshold
	executionCfg.ComplexityMaxComplexity = cfg.Complexity.MaxComplexity

	if cfg.Clones != nil {
		if cfg.Clones.LSH.Enabled != "" {
			executionCfg.CloneLSHEnabled = cfg.Clones.LSH.Enabled
		}
		if cfg.Clones.LSH.AutoThreshold > 0 {
			executionCfg.CloneLSHAutoThreshold = cfg.Clones.LSH.AutoThreshold
		}
	}

	return executionCfg
}
