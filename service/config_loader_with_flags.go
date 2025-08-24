package service

import (
	"github.com/pyqol/pyqol/domain"
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

	// Helper function to check if a flag was explicitly set
	wasExplicitlySet := func(flagName string) bool {
		if c.explicitFlags == nil {
			return false
		}
		return c.explicitFlags[flagName]
	}

	// Always override paths as they come from command arguments
	if len(override.Paths) > 0 {
		merged.Paths = override.Paths
	}

	// Output configuration
	if wasExplicitlySet("format") {
		merged.OutputFormat = override.OutputFormat
	}

	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}

	if wasExplicitlySet("details") {
		merged.ShowDetails = override.ShowDetails
	}

	// Filtering and sorting
	if wasExplicitlySet("min") {
		merged.MinComplexity = override.MinComplexity
	}

	if wasExplicitlySet("max") {
		merged.MaxComplexity = override.MaxComplexity
	}

	if wasExplicitlySet("sort") {
		merged.SortBy = override.SortBy
	}

	// Complexity thresholds
	if wasExplicitlySet("low-threshold") {
		merged.LowThreshold = override.LowThreshold
	}

	if wasExplicitlySet("medium-threshold") {
		merged.MediumThreshold = override.MediumThreshold
	}

	// Config path is always from override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// For recursive, only override if explicitly set
	if wasExplicitlySet("recursive") {
		merged.Recursive = override.Recursive
	}

	// Patterns
	if wasExplicitlySet("include") && len(override.IncludePatterns) > 0 {
		merged.IncludePatterns = override.IncludePatterns
	}

	if wasExplicitlySet("exclude") && len(override.ExcludePatterns) > 0 {
		merged.ExcludePatterns = override.ExcludePatterns
	}

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