package service

import (
	"os"

	"github.com/pyqol/pyqol/domain"
	"github.com/pyqol/pyqol/internal/config"
)

// ConfigurationLoaderImpl implements the ConfigurationLoader interface
type ConfigurationLoaderImpl struct{}

// NewConfigurationLoader creates a new configuration loader service
func NewConfigurationLoader() *ConfigurationLoaderImpl {
	return &ConfigurationLoaderImpl{}
}

// LoadConfig loads configuration from the specified path
func (c *ConfigurationLoaderImpl) LoadConfig(path string) (*domain.ComplexityRequest, error) {
	cfg, err := config.LoadConfig(path)
	if err != nil {
		return nil, domain.NewConfigError("failed to load configuration file", err)
	}

	return c.convertToComplexityRequest(cfg), nil
}

// LoadDefaultConfig loads the default configuration, first checking for .pyqol.yaml
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

	// Helper function to check if a flag was explicitly set
	wasExplicitlySet := func(flagName string) bool {
		if override.ExplicitFlags == nil {
			return false
		}
		return override.ExplicitFlags[flagName]
	}

	// Override with values from override only if they were explicitly set
	// Always override paths as they come from command arguments
	if len(override.Paths) > 0 {
		merged.Paths = override.Paths
	}

	// Output configuration
	if wasExplicitlySet("format") || override.OutputFormat != "" {
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

// FindDefaultConfigFile looks for .pyqol.yaml in the current directory
func (c *ConfigurationLoaderImpl) FindDefaultConfigFile() string {
	configFiles := []string{".pyqol.yaml", ".pyqol.yml", "pyqol.yaml"}

	for _, filename := range configFiles {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}

	return "" // No config file found
}
