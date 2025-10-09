package service

import (
	"fmt"
	"os"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/spf13/viper"
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
		IncludeStdLib:       false,
		IncludeThirdParty:   true,
		FollowRelative:      true,
		DetectCycles:        true,
		Recursive:           true,
		IncludePatterns:     []string{"**/*.py"},
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

// loadFromViperSection loads configuration from a specific viper section
func (cl *SystemAnalysisConfigurationLoaderImpl) loadFromViperSection(v *viper.Viper, section string, request *domain.SystemAnalysisRequest) error {
	// Analysis types
	if v.IsSet(section + ".analyze_dependencies") {
		request.AnalyzeDependencies = v.GetBool(section + ".analyze_dependencies")
	}
	if v.IsSet(section + ".analyze_architecture") {
		request.AnalyzeArchitecture = v.GetBool(section + ".analyze_architecture")
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
	// Initialize rules if missing
	if request.ArchitectureRules == nil {
		request.ArchitectureRules = &domain.ArchitectureRules{}
	}

	// strict_mode
	if v.IsSet(section + ".strict_mode") {
		request.ArchitectureRules.StrictMode = v.GetBool(section + ".strict_mode")
	}

	// allowed_patterns / forbidden_patterns
	if v.IsSet(section + ".allowed_patterns") {
		request.ArchitectureRules.AllowedPatterns = v.GetStringSlice(section + ".allowed_patterns")
	}
	if v.IsSet(section + ".forbidden_patterns") {
		request.ArchitectureRules.ForbiddenPatterns = v.GetStringSlice(section + ".forbidden_patterns")
	}

	// Load layers
	if v.IsSet(section + ".layers") {
		layers, err := cl.unmarshalLayers(v, section+".layers")
		if err != nil {
			return err
		}
		request.ArchitectureRules.Layers = layers
	}

	// Load rules
	if v.IsSet(section + ".rules") {
		rules, err := cl.unmarshalRules(v, section+".rules")
		if err != nil {
			return err
		}
		request.ArchitectureRules.Rules = rules
	}

	return nil
}

// unmarshalLayers extracts layer configuration from viper
func (cl *SystemAnalysisConfigurationLoaderImpl) unmarshalLayers(v *viper.Viper, key string) ([]domain.Layer, error) {
	var rawLayers []map[string]interface{}
	if err := v.UnmarshalKey(key, &rawLayers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal architecture layers config: %w", err)
	}

	layers := make([]domain.Layer, 0, len(rawLayers))
	for _, l := range rawLayers {
		layer := cl.parseLayer(l)
		if layer.Name != "" {
			layers = append(layers, layer)
		}
	}
	return layers, nil
}

// parseLayer converts a raw map to a Layer struct
func (cl *SystemAnalysisConfigurationLoaderImpl) parseLayer(raw map[string]interface{}) domain.Layer {
	layer := domain.Layer{}

	if name, ok := raw["name"].(string); ok {
		layer.Name = name
	}
	if desc, ok := raw["description"].(string); ok {
		layer.Description = desc
	}

	layer.Packages = cl.extractStringSlice(raw["packages"])
	return layer
}

// unmarshalRules extracts rule configuration from viper
func (cl *SystemAnalysisConfigurationLoaderImpl) unmarshalRules(v *viper.Viper, key string) ([]domain.LayerRule, error) {
	var rawRules []map[string]interface{}
	if err := v.UnmarshalKey(key, &rawRules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal architecture rules config: %w", err)
	}

	rules := make([]domain.LayerRule, 0, len(rawRules))
	for _, r := range rawRules {
		rule := cl.parseRule(r)
		if rule.From != "" {
			rules = append(rules, rule)
		}
	}
	return rules, nil
}

// parseRule converts a raw map to a LayerRule struct
func (cl *SystemAnalysisConfigurationLoaderImpl) parseRule(raw map[string]interface{}) domain.LayerRule {
	rule := domain.LayerRule{}

	if from, ok := raw["from"].(string); ok {
		rule.From = from
	}

	rule.Allow = cl.extractStringSlice(raw["allow"])
	rule.Deny = cl.extractStringSlice(raw["deny"])
	return rule
}

// extractStringSlice handles type conversion for string slices
func (cl *SystemAnalysisConfigurationLoaderImpl) extractStringSlice(value interface{}) []string {
	var result []string

	switch v := value.(type) {
	case []string:
		result = append(result, v...)
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
	}

	return result
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
show_details = false
recursive = true
include_patterns = ["**/*.py"]
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
