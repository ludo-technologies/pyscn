package service

import (
	"github.com/pyqol/pyqol/domain"
	"github.com/pyqol/pyqol/internal/config"
)

// ConfigurationLoaderWithFlags wraps configuration loading with explicit flag tracking
type ConfigurationLoaderWithFlags struct {
	loader        *ConfigurationLoaderImpl
	explicitFlags map[string]bool
}

// NewConfigurationLoaderWithFlags creates a new configuration loader that tracks explicit flags
func NewConfigurationLoaderWithFlags(explicitFlags map[string]bool) *ConfigurationLoaderWithFlags {
	return &ConfigurationLoaderWithFlags{
		loader:        NewConfigurationLoader(),
		explicitFlags: explicitFlags,
	}
}

// LoadConfig loads configuration from the specified path
func (c *ConfigurationLoaderWithFlags) LoadConfig(path string) (*domain.ComplexityRequest, error) {
	return c.loader.LoadConfig(path)
}

// LoadDefaultConfig loads the default configuration
func (c *ConfigurationLoaderWithFlags) LoadDefaultConfig() *domain.ComplexityRequest {
	return c.loader.LoadDefaultConfig()
}

// MergeConfig merges CLI flags with configuration file, respecting explicit flags
func (c *ConfigurationLoaderWithFlags) MergeConfig(base *domain.ComplexityRequest, override *domain.ComplexityRequest) *domain.ComplexityRequest {
	// Start with base configuration
	merged := *base


	// Always override paths as they come from command arguments
	if len(override.Paths) > 0 {
		merged.Paths = override.Paths
	}

	// Output configuration
	if config.WasExplicitlySet(c.explicitFlags, "format") {
		merged.OutputFormat = override.OutputFormat
	}

	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}

	merged.ShowDetails = config.MergeBool(merged.ShowDetails, override.ShowDetails, "details", c.explicitFlags)

	// Filtering and sorting
	merged.MinComplexity = config.MergeInt(merged.MinComplexity, override.MinComplexity, "min", c.explicitFlags)
	merged.MaxComplexity = config.MergeInt(merged.MaxComplexity, override.MaxComplexity, "max", c.explicitFlags)

	if config.WasExplicitlySet(c.explicitFlags, "sort") {
		merged.SortBy = override.SortBy
	}

	// Complexity thresholds
	merged.LowThreshold = config.MergeInt(merged.LowThreshold, override.LowThreshold, "low-threshold", c.explicitFlags)
	merged.MediumThreshold = config.MergeInt(merged.MediumThreshold, override.MediumThreshold, "medium-threshold", c.explicitFlags)

	// Config path is always from override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// For recursive, only override if explicitly set
	merged.Recursive = config.MergeBool(merged.Recursive, override.Recursive, "recursive", c.explicitFlags)

	// Patterns
	merged.IncludePatterns = config.MergeStringSlice(merged.IncludePatterns, override.IncludePatterns, "include", c.explicitFlags)
	merged.ExcludePatterns = config.MergeStringSlice(merged.ExcludePatterns, override.ExcludePatterns, "exclude", c.explicitFlags)

	return &merged
}

// ValidateConfig validates a configuration request
func (c *ConfigurationLoaderWithFlags) ValidateConfig(req *domain.ComplexityRequest) error {
	return c.loader.ValidateConfig(req)
}

// GetDefaultThresholds returns the default complexity thresholds
func (c *ConfigurationLoaderWithFlags) GetDefaultThresholds() (low, medium int) {
	return c.loader.GetDefaultThresholds()
}

// CreateConfigTemplate creates a template configuration file
func (c *ConfigurationLoaderWithFlags) CreateConfigTemplate(path string) error {
	return c.loader.CreateConfigTemplate(path)
}

// FindDefaultConfigFile looks for .pyqol.yaml in the current directory
func (c *ConfigurationLoaderWithFlags) FindDefaultConfigFile() string {
	return c.loader.FindDefaultConfigFile()
}