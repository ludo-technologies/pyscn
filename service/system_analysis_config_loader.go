package service

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/ludo-technologies/pyscn/domain"
)

// SystemAnalysisConfigurationLoaderImpl implements the SystemAnalysisConfigurationLoader interface
type SystemAnalysisConfigurationLoaderImpl struct{}

// NewSystemAnalysisConfigurationLoader creates a new system analysis configuration loader
func NewSystemAnalysisConfigurationLoader() *SystemAnalysisConfigurationLoaderImpl {
	return &SystemAnalysisConfigurationLoaderImpl{}
}

// LoadConfig loads configuration from the specified path
func (cl *SystemAnalysisConfigurationLoaderImpl) LoadConfig(path string) (*domain.SystemAnalysisRequest, error) {
	// Create a new viper instance
	v := viper.New()

	// Set configuration file
	if path != "" {
		// Use the specified file
		v.SetConfigFile(path)
	} else {
		// Look for default configuration files
		v.SetConfigName(".pyscn")
		v.SetConfigType("toml")
		v.AddConfigPath(".")

		// Also check for pyproject.toml
		pyprojectPath := "pyproject.toml"
		if _, err := os.Stat(pyprojectPath); err == nil {
			v.SetConfigFile(pyprojectPath)
		}
	}

	// Read the config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; use default configuration
			return cl.LoadDefaultConfig(), nil
		} else {
			// Config file was found but another error was produced
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Load system analysis specific configuration
	request := cl.LoadDefaultConfig()

	// Load from [tool.pyscn.system_analysis] section if it exists (for pyproject.toml)
	if v.IsSet("tool.pyscn.system_analysis") {
		if err := cl.loadFromViperSection(v, "tool.pyscn.system_analysis", request); err != nil {
			return nil, err
		}
	}

	// Load from [system_analysis] section if it exists (for .pyscn.toml)
	if v.IsSet("system_analysis") {
		if err := cl.loadFromViperSection(v, "system_analysis", request); err != nil {
			return nil, err
		}
	}

	// Load dependency analysis configuration
	if v.IsSet("tool.pyscn.dependencies") {
		if err := cl.loadDependencyConfig(v, "tool.pyscn.dependencies", request); err != nil {
			return nil, err
		}
	}
	if v.IsSet("dependencies") {
		if err := cl.loadDependencyConfig(v, "dependencies", request); err != nil {
			return nil, err
		}
	}

	// Load architecture analysis configuration
	if v.IsSet("tool.pyscn.architecture") {
		if err := cl.loadArchitectureConfig(v, "tool.pyscn.architecture", request); err != nil {
			return nil, err
		}
	}
	if v.IsSet("architecture") {
		if err := cl.loadArchitectureConfig(v, "architecture", request); err != nil {
			return nil, err
		}
	}

	return request, nil
}

// LoadDefaultConfig loads the default configuration
func (cl *SystemAnalysisConfigurationLoaderImpl) LoadDefaultConfig() *domain.SystemAnalysisRequest {
	return &domain.SystemAnalysisRequest{
		OutputFormat:        domain.OutputFormatText,
		AnalyzeDependencies: true,
		AnalyzeArchitecture: true,
		AnalyzeQuality:      true,
		IncludeStdLib:       false,
		IncludeThirdParty:   true,
		FollowRelative:      true,
		DetectCycles:        true,
		Recursive:           true,
		IncludePatterns:     []string{"*.py"},
		ExcludePatterns:     []string{},
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

	// Analysis type overrides - CLI takes precedence
	merged.AnalyzeDependencies = override.AnalyzeDependencies
	merged.AnalyzeArchitecture = override.AnalyzeArchitecture
	merged.AnalyzeQuality = override.AnalyzeQuality

	// (Filtering options removed - fields no longer exist)

	// Analysis options - CLI takes precedence
	merged.IncludeStdLib = override.IncludeStdLib
	merged.IncludeThirdParty = override.IncludeThirdParty
	merged.FollowRelative = override.FollowRelative
	merged.DetectCycles = override.DetectCycles

	// File selection - override if provided
	if len(override.IncludePatterns) > 0 {
		merged.IncludePatterns = override.IncludePatterns
	}
	if len(override.ExcludePatterns) > 0 {
		merged.ExcludePatterns = override.ExcludePatterns
	}
	merged.Recursive = override.Recursive

	return &merged
}

// loadFromViperSection loads configuration from a specific viper section
func (cl *SystemAnalysisConfigurationLoaderImpl) loadFromViperSection(v *viper.Viper, section string, request *domain.SystemAnalysisRequest) error {
	// Analysis types
	if v.IsSet(section + ".analyze_dependencies") {
		request.AnalyzeDependencies = v.GetBool(section + ".analyze_dependencies")
	}
	if v.IsSet(section + ".analyze_architecture") {
		request.AnalyzeArchitecture = v.GetBool(section + ".analyze_architecture")
	}
	if v.IsSet(section + ".analyze_quality") {
		request.AnalyzeQuality = v.GetBool(section + ".analyze_quality")
	}

	// (Output options removed - ShowDetails field no longer exists)

	// File patterns
	if v.IsSet(section + ".include_patterns") {
		request.IncludePatterns = v.GetStringSlice(section + ".include_patterns")
	}
	if v.IsSet(section + ".exclude_patterns") {
		request.ExcludePatterns = v.GetStringSlice(section + ".exclude_patterns")
	}
	if v.IsSet(section + ".recursive") {
		request.Recursive = v.GetBool(section + ".recursive")
	}

	return nil
}

// loadDependencyConfig loads dependency-specific configuration
func (cl *SystemAnalysisConfigurationLoaderImpl) loadDependencyConfig(v *viper.Viper, section string, request *domain.SystemAnalysisRequest) error {
	// (Filtering options removed - fields no longer exist)

	// Analysis options
	if v.IsSet(section + ".include_stdlib") {
		request.IncludeStdLib = v.GetBool(section + ".include_stdlib")
	}
	if v.IsSet(section + ".include_third_party") {
		request.IncludeThirdParty = v.GetBool(section + ".include_third_party")
	}
	if v.IsSet(section + ".follow_relative") {
		request.FollowRelative = v.GetBool(section + ".follow_relative")
	}
	if v.IsSet(section + ".detect_cycles") {
		request.DetectCycles = v.GetBool(section + ".detect_cycles")
	}

	return nil
}

// loadArchitectureConfig loads architecture-specific configuration
func (cl *SystemAnalysisConfigurationLoaderImpl) loadArchitectureConfig(v *viper.Viper, section string, request *domain.SystemAnalysisRequest) error {
	// Architecture analysis doesn't have specific config options yet
	// This method is here for future extensibility
	
	// Future options could include:
	// - Layer definitions
	// - Architecture pattern detection settings
	// - Violation severity thresholds
	
	return nil
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
analyze_quality = true
show_details = false
recursive = true
include_patterns = ["*.py"]
exclude_patterns = ["test_*.py", "*_test.py"]

[dependencies]
min_coupling = 0
max_coupling = 0  # 0 means no limit
min_instability = 0.0
max_distance = 1.0
include_stdlib = false
include_third_party = true
follow_relative = true
detect_cycles = true

[architecture]
# Architecture-specific settings
# (to be extended in future versions)
`