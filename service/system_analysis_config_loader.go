package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/pelletier/go-toml/v2"
)

// SystemAnalysisConfigurationLoaderImpl implements the SystemAnalysisConfigurationLoader interface
type SystemAnalysisConfigurationLoaderImpl struct{}

// NewSystemAnalysisConfigurationLoader creates a new system analysis configuration loader
func NewSystemAnalysisConfigurationLoader() *SystemAnalysisConfigurationLoaderImpl {
	return &SystemAnalysisConfigurationLoaderImpl{}
}

// LoadConfig loads configuration from the specified path using TOML-only loader
func (cl *SystemAnalysisConfigurationLoaderImpl) LoadConfig(path string) (*domain.SystemAnalysisRequest, error) {
	// Start with default configuration
	request := cl.LoadDefaultConfig()

	// Determine the directory to search for config
	startDir := "."
	if path != "" {
		info, err := os.Stat(path)
		if err == nil {
			if info.IsDir() {
				startDir = path
			} else {
				startDir = filepath.Dir(path)
			}
		}
	}

	// Try to load .pyscn.toml first
	pyscnTomlPath := cl.findPyscnToml(startDir)
	if pyscnTomlPath != "" {
		if err := cl.loadFromPyscnToml(pyscnTomlPath, request); err != nil {
			return nil, fmt.Errorf("error reading .pyscn.toml: %w", err)
		}
		return request, nil
	}

	// Try to load pyproject.toml
	pyprojectPath := cl.findPyprojectToml(startDir)
	if pyprojectPath != "" {
		if err := cl.loadFromPyprojectToml(pyprojectPath, request); err != nil {
			return nil, fmt.Errorf("error reading pyproject.toml: %w", err)
		}
		return request, nil
	}

	// No config file found, return defaults
	return request, nil
}

// findPyscnToml walks up the directory tree to find .pyscn.toml
func (cl *SystemAnalysisConfigurationLoaderImpl) findPyscnToml(startDir string) string {
	dir := startDir
	for {
		configPath := filepath.Join(dir, ".pyscn.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// findPyprojectToml walks up the directory tree to find pyproject.toml
func (cl *SystemAnalysisConfigurationLoaderImpl) findPyprojectToml(startDir string) string {
	dir := startDir
	for {
		configPath := filepath.Join(dir, "pyproject.toml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// loadFromPyscnToml loads configuration from .pyscn.toml
func (cl *SystemAnalysisConfigurationLoaderImpl) loadFromPyscnToml(configPath string, request *domain.SystemAnalysisRequest) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cfg config.PyscnTomlConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	// Merge system_analysis section
	cl.mergeSystemAnalysisSection(&cfg.SystemAnalysis, request)

	// Merge dependencies section
	cl.mergeDependenciesSection(&cfg.Dependencies, request)

	// Merge architecture section
	cl.mergeArchitectureSection(&cfg.Architecture, request)

	// Merge analysis section (for include/exclude patterns)
	cl.mergeAnalysisSection(&cfg.Analysis, request)

	return nil
}

// loadFromPyprojectToml loads configuration from pyproject.toml
func (cl *SystemAnalysisConfigurationLoaderImpl) loadFromPyprojectToml(configPath string, request *domain.SystemAnalysisRequest) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var pyproject config.PyprojectToml
	if err := toml.Unmarshal(data, &pyproject); err != nil {
		return err
	}

	pyscn := pyproject.Tool.Pyscn

	// Merge system_analysis section
	cl.mergeSystemAnalysisSection(&pyscn.SystemAnalysis, request)

	// Merge dependencies section
	cl.mergeDependenciesSection(&pyscn.Dependencies, request)

	// Merge architecture section
	cl.mergeArchitectureSection(&pyscn.Architecture, request)

	// Merge analysis section (for include/exclude patterns)
	cl.mergeAnalysisSection(&pyscn.Analysis, request)

	return nil
}

// mergeSystemAnalysisSection merges system_analysis settings into the request
func (cl *SystemAnalysisConfigurationLoaderImpl) mergeSystemAnalysisSection(cfg *config.SystemAnalysisTomlConfig, request *domain.SystemAnalysisRequest) {
	if cfg.EnableDependencies != nil {
		request.AnalyzeDependencies = cfg.EnableDependencies
	}
	if cfg.EnableArchitecture != nil {
		request.AnalyzeArchitecture = cfg.EnableArchitecture
	}
}

// mergeDependenciesSection merges dependencies settings into the request
func (cl *SystemAnalysisConfigurationLoaderImpl) mergeDependenciesSection(cfg *config.DependenciesTomlConfig, request *domain.SystemAnalysisRequest) {
	if cfg.IncludeStdLib != nil {
		request.IncludeStdLib = cfg.IncludeStdLib
	}
	if cfg.IncludeThirdParty != nil {
		request.IncludeThirdParty = cfg.IncludeThirdParty
	}
	if cfg.FollowRelative != nil {
		request.FollowRelative = cfg.FollowRelative
	}
	if cfg.DetectCycles != nil {
		request.DetectCycles = cfg.DetectCycles
	}
}

// mergeArchitectureSection merges architecture settings into the request
func (cl *SystemAnalysisConfigurationLoaderImpl) mergeArchitectureSection(cfg *config.ArchitectureTomlConfig, request *domain.SystemAnalysisRequest) {
	// Initialize architecture rules if needed
	if request.ArchitectureRules == nil {
		request.ArchitectureRules = &domain.ArchitectureRules{}
	}

	if cfg.StrictMode != nil {
		request.ArchitectureRules.StrictMode = *cfg.StrictMode
	}
	if len(cfg.AllowedPatterns) > 0 {
		request.ArchitectureRules.AllowedPatterns = cfg.AllowedPatterns
	}
	if len(cfg.ForbiddenPatterns) > 0 {
		request.ArchitectureRules.ForbiddenPatterns = cfg.ForbiddenPatterns
	}
}

// mergeAnalysisSection merges analysis settings into the request
func (cl *SystemAnalysisConfigurationLoaderImpl) mergeAnalysisSection(cfg *config.AnalysisTomlConfig, request *domain.SystemAnalysisRequest) {
	if len(cfg.IncludePatterns) > 0 {
		request.IncludePatterns = cfg.IncludePatterns
	}
	if len(cfg.ExcludePatterns) > 0 {
		request.ExcludePatterns = cfg.ExcludePatterns
	}
	if cfg.Recursive != nil {
		request.Recursive = cfg.Recursive
	}
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
