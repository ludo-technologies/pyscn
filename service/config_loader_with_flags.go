package service

import (
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// ConfigurationLoaderWithFlags wraps configuration loading with explicit flag tracking
type ConfigurationLoaderWithFlags struct {
	loader      *ConfigurationLoaderImpl
	flagTracker *config.FlagTracker
}

// NewConfigurationLoaderWithFlags creates a new configuration loader that tracks explicit flags
func NewConfigurationLoaderWithFlags(explicitFlags map[string]bool) *ConfigurationLoaderWithFlags {
	return &ConfigurationLoaderWithFlags{
		loader:      NewConfigurationLoader(),
		flagTracker: config.NewFlagTrackerWithFlags(explicitFlags),
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
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	// Start with base configuration
	merged := *base

	// Always override paths as they come from command arguments
	if len(override.Paths) > 0 {
		merged.Paths = override.Paths
	}

	// Output configuration - always use override format when explicitly set
	// Since we removed --format flag, check for individual format flags or non-text format
	if override.OutputFormat != "" {
		// If a specific format was set (not text), use it
		if override.OutputFormat != domain.OutputFormatText {
			merged.OutputFormat = override.OutputFormat
		} else if c.flagTracker.WasSet("html") || c.flagTracker.WasSet("json") ||
			c.flagTracker.WasSet("csv") || c.flagTracker.WasSet("yaml") {
			// If any format flag was set, use the override format
			merged.OutputFormat = override.OutputFormat
		}
	}

	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}

	// Always preserve output path and no-open flag from override (command line)
	// These are generated based on format flags, not set directly
	if override.OutputPath != "" {
		merged.OutputPath = override.OutputPath
	}

	// Always preserve NoOpen from override if it's been set
	merged.NoOpen = override.NoOpen

	merged.ShowDetails = c.flagTracker.MergeBool(merged.ShowDetails, override.ShowDetails, "details")

	// Filtering and sorting
	merged.MinComplexity = c.flagTracker.MergeInt(merged.MinComplexity, override.MinComplexity, "min")
	merged.MaxComplexity = c.flagTracker.MergeInt(merged.MaxComplexity, override.MaxComplexity, "max")

	// Only override sort if explicitly set
	if c.flagTracker.WasSet("sort") {
		merged.SortBy = override.SortBy
	}

	// Complexity thresholds
	merged.LowThreshold = c.flagTracker.MergeInt(merged.LowThreshold, override.LowThreshold, "low-threshold")
	merged.MediumThreshold = c.flagTracker.MergeInt(merged.MediumThreshold, override.MediumThreshold, "medium-threshold")

	// Config path is always from override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// For recursive, only override if explicitly set
	merged.Recursive = c.flagTracker.MergeBool(merged.Recursive, override.Recursive, "recursive")

	// Patterns
	merged.IncludePatterns = c.flagTracker.MergeStringSlice(merged.IncludePatterns, override.IncludePatterns, "include")
	merged.ExcludePatterns = c.flagTracker.MergeStringSlice(merged.ExcludePatterns, override.ExcludePatterns, "exclude")

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

// FindDefaultConfigFile looks for .pyscn.toml in the current directory
func (c *ConfigurationLoaderWithFlags) FindDefaultConfigFile() string {
	return c.loader.FindDefaultConfigFile()
}
