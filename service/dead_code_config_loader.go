package service

import (
	"fmt"
	"os"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
	"github.com/spf13/viper"
)

// DeadCodeConfigurationLoaderImpl implements the DeadCodeConfigurationLoader interface
type DeadCodeConfigurationLoaderImpl struct{}

// NewDeadCodeConfigurationLoader creates a new dead code configuration loader service
func NewDeadCodeConfigurationLoader() *DeadCodeConfigurationLoaderImpl {
	return &DeadCodeConfigurationLoaderImpl{}
}

// LoadConfig loads dead code configuration from the specified path using TOML-only strategy
func (cl *DeadCodeConfigurationLoaderImpl) LoadConfig(path string) (*domain.DeadCodeRequest, error) {
	// Use TOML-only loader
	tomlLoader := config.NewTomlConfigLoader()
	cloneCfg, err := tomlLoader.LoadConfig(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
	}

	// Convert clone config to unified config format, then to dead code request
	cfg := cl.cloneConfigToUnifiedConfig(cloneCfg)
	return cl.configToRequest(cfg), nil
}

// LoadDefaultConfig loads the default dead code configuration, first checking for .pyscn.toml
func (cl *DeadCodeConfigurationLoaderImpl) LoadDefaultConfig() *domain.DeadCodeRequest {
	// First, try to find and load a config file in the current directory
	configFile := cl.FindDefaultConfigFile()
	if configFile != "" {
		if configReq, err := cl.LoadConfig(configFile); err == nil {
			return configReq
		}
		// If loading failed, fall back to hardcoded defaults
	}

	// Fall back to hardcoded default configuration
	cfg := config.DefaultConfig()
	return cl.configToRequest(cfg)
}

// MergeConfig merges CLI flags with configuration file
func (cl *DeadCodeConfigurationLoaderImpl) MergeConfig(base *domain.DeadCodeRequest, override *domain.DeadCodeRequest) *domain.DeadCodeRequest {
	if base == nil {
		return override
	}
	if override == nil {
		return base
	}

	// Start with base config
	merged := *base

	// Override with CLI values (only if they're not default/empty)
	if len(override.Paths) > 0 {
		merged.Paths = override.Paths
	}
	if override.OutputFormat != "" {
		merged.OutputFormat = override.OutputFormat
	}
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	if override.MinSeverity != "" {
		merged.MinSeverity = override.MinSeverity
	}
	if override.SortBy != "" {
		merged.SortBy = override.SortBy
	}
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// Boolean values - use override values
	merged.ShowContext = override.ShowContext
	merged.Recursive = override.Recursive
	merged.DetectAfterReturn = override.DetectAfterReturn
	merged.DetectAfterBreak = override.DetectAfterBreak
	merged.DetectAfterContinue = override.DetectAfterContinue
	merged.DetectAfterRaise = override.DetectAfterRaise
	merged.DetectUnreachableBranches = override.DetectUnreachableBranches

	// Integer values - use override if positive
	if override.ContextLines >= 0 {
		merged.ContextLines = override.ContextLines
	}

	// Array values - use override if not empty
	if len(override.IncludePatterns) > 0 {
		merged.IncludePatterns = override.IncludePatterns
	}
	if len(override.ExcludePatterns) > 0 {
		merged.ExcludePatterns = override.ExcludePatterns
	}
	if len(override.IgnorePatterns) > 0 {
		merged.IgnorePatterns = override.IgnorePatterns
	}

	return &merged
}

// configToRequest converts a config.Config to domain.DeadCodeRequest
func (cl *DeadCodeConfigurationLoaderImpl) configToRequest(cfg *config.Config) *domain.DeadCodeRequest {
	if cfg == nil {
		return domain.DefaultDeadCodeRequest()
	}

	// Convert output format
	var outputFormat domain.OutputFormat
	switch cfg.Output.Format {
	case "json":
		outputFormat = domain.OutputFormatJSON
	case "yaml", "yml":
		outputFormat = domain.OutputFormatYAML
	case "csv":
		outputFormat = domain.OutputFormatCSV
	case "html":
		outputFormat = domain.OutputFormatHTML
	default:
		outputFormat = domain.OutputFormatText
	}

	// Convert severity level
	var minSeverity domain.DeadCodeSeverity
	switch cfg.DeadCode.MinSeverity {
	case "critical":
		minSeverity = domain.DeadCodeSeverityCritical
	case "info":
		minSeverity = domain.DeadCodeSeverityInfo
	default:
		minSeverity = domain.DeadCodeSeverityWarning
	}

	// Convert sort criteria
	var sortBy domain.DeadCodeSortCriteria
	switch cfg.DeadCode.SortBy {
	case "line":
		sortBy = domain.DeadCodeSortByLine
	case "file":
		sortBy = domain.DeadCodeSortByFile
	case "function":
		sortBy = domain.DeadCodeSortByFunction
	default:
		sortBy = domain.DeadCodeSortBySeverity
	}

	return &domain.DeadCodeRequest{
		OutputFormat:              outputFormat,
		ShowContext:               cfg.DeadCode.ShowContext,
		ContextLines:              cfg.DeadCode.ContextLines,
		MinSeverity:               minSeverity,
		SortBy:                    sortBy,
		Recursive:                 cfg.Analysis.Recursive,
		IncludePatterns:           cfg.Analysis.IncludePatterns,
		ExcludePatterns:           cfg.Analysis.ExcludePatterns,
		IgnorePatterns:            cfg.DeadCode.IgnorePatterns,
		DetectAfterReturn:         cfg.DeadCode.DetectAfterReturn,
		DetectAfterBreak:          cfg.DeadCode.DetectAfterBreak,
		DetectAfterContinue:       cfg.DeadCode.DetectAfterContinue,
		DetectAfterRaise:          cfg.DeadCode.DetectAfterRaise,
		DetectUnreachableBranches: cfg.DeadCode.DetectUnreachableBranches,
	}
}

// FindDefaultConfigFile looks for TOML config files in the current directory
func (cl *DeadCodeConfigurationLoaderImpl) FindDefaultConfigFile() string {
	// Use TOML-only strategy
	tomlLoader := config.NewTomlConfigLoader()
	configFiles := tomlLoader.GetSupportedConfigFiles()
	
	for _, filename := range configFiles {
		if _, err := os.Stat(filename); err == nil {
			return filename
		}
	}

	return "" // No config file found
}

// ValidateConfig validates a dead code configuration
func (cl *DeadCodeConfigurationLoaderImpl) ValidateConfig(req *domain.DeadCodeRequest) error {
	if req == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	return req.Validate()
}

// SaveConfig saves dead code configuration to a file
func (cl *DeadCodeConfigurationLoaderImpl) SaveConfig(req *domain.DeadCodeRequest, path string) error {
	if req == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Convert request back to config format
	cfg := cl.requestToConfig(req)

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// Set dead code config values
	viper.Set("dead_code", cfg.DeadCode)
	viper.Set("output", cfg.Output)
	viper.Set("analysis", cfg.Analysis)

	return viper.WriteConfig()
}

// requestToConfig converts a domain.DeadCodeRequest back to config.Config
func (cl *DeadCodeConfigurationLoaderImpl) requestToConfig(req *domain.DeadCodeRequest) *config.Config {
	cfg := config.DefaultConfig()

	// Convert output format
	switch req.OutputFormat {
	case domain.OutputFormatJSON:
		cfg.Output.Format = "json"
	case domain.OutputFormatYAML:
		cfg.Output.Format = "yaml"
	case domain.OutputFormatCSV:
		cfg.Output.Format = "csv"
	default:
		cfg.Output.Format = "text"
	}

	// Convert severity level
	switch req.MinSeverity {
	case domain.DeadCodeSeverityCritical:
		cfg.DeadCode.MinSeverity = "critical"
	case domain.DeadCodeSeverityInfo:
		cfg.DeadCode.MinSeverity = "info"
	default:
		cfg.DeadCode.MinSeverity = "warning"
	}

	// Convert sort criteria
	switch req.SortBy {
	case domain.DeadCodeSortByLine:
		cfg.DeadCode.SortBy = "line"
	case domain.DeadCodeSortByFile:
		cfg.DeadCode.SortBy = "file"
	case domain.DeadCodeSortByFunction:
		cfg.DeadCode.SortBy = "function"
	default:
		cfg.DeadCode.SortBy = "severity"
	}

	// Set dead code specific config
	cfg.DeadCode.ShowContext = req.ShowContext
	cfg.DeadCode.ContextLines = req.ContextLines
	cfg.DeadCode.DetectAfterReturn = req.DetectAfterReturn
	cfg.DeadCode.DetectAfterBreak = req.DetectAfterBreak
	cfg.DeadCode.DetectAfterContinue = req.DetectAfterContinue
	cfg.DeadCode.DetectAfterRaise = req.DetectAfterRaise
	cfg.DeadCode.DetectUnreachableBranches = req.DetectUnreachableBranches
	cfg.DeadCode.IgnorePatterns = req.IgnorePatterns

	// Set analysis config
	cfg.Analysis.Recursive = req.Recursive
	cfg.Analysis.IncludePatterns = req.IncludePatterns
	cfg.Analysis.ExcludePatterns = req.ExcludePatterns

	return cfg
}

// cloneConfigToUnifiedConfig converts CloneConfig to unified Config format
func (cl *DeadCodeConfigurationLoaderImpl) cloneConfigToUnifiedConfig(cloneCfg *config.CloneConfig) *config.Config {
	cfg := config.DefaultConfig()
	
	// Map analysis settings
	cfg.Analysis.IncludePatterns = cloneCfg.Input.IncludePatterns
	cfg.Analysis.ExcludePatterns = cloneCfg.Input.ExcludePatterns
	cfg.Analysis.Recursive = cloneCfg.Input.Recursive
	
	// Map output settings
	cfg.Output.Format = cloneCfg.Output.Format
	cfg.Output.ShowDetails = cloneCfg.Output.ShowDetails
	
	// Dead code settings use defaults from DefaultConfig()
	// since TOML-only config focuses on clone detection
	
	return cfg
}
