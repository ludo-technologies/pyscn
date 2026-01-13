package service

import (
	"fmt"
	"os"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
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

	// Convert pyscn config to unified config format, then to dead code request
	cfg := cl.pyscnConfigToUnifiedConfig(cloneCfg)
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

	// Always override paths as they come from command arguments
	if len(override.Paths) > 0 {
		merged.Paths = override.Paths
	}

	// Output configuration
	if override.OutputFormat != "" {
		merged.OutputFormat = override.OutputFormat
	}
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}

	// MinSeverity - override only if non-default
	if override.MinSeverity != "" && override.MinSeverity != domain.DeadCodeSeverityWarning {
		merged.MinSeverity = override.MinSeverity
	}

	// SortBy - override only if non-default
	if override.SortBy != "" && override.SortBy != domain.DeadCodeSortBySeverity {
		merged.SortBy = override.SortBy
	}

	// ConfigPath - always override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// Boolean pointer values - nil means not set, non-nil means explicitly set
	// With pointer types, we can now distinguish between "not set" and "set to default value"
	// This allows proper precedence: CLI override > config file > defaults

	// ShowContext: use override if explicitly set (non-nil), otherwise use base
	if override.ShowContext != nil {
		merged.ShowContext = override.ShowContext
	} else {
		merged.ShowContext = base.ShowContext
	}

	// Detection flags: use override if explicitly set (non-nil), otherwise use base
	if override.DetectAfterReturn != nil {
		merged.DetectAfterReturn = override.DetectAfterReturn
	} else {
		merged.DetectAfterReturn = base.DetectAfterReturn
	}

	if override.DetectAfterBreak != nil {
		merged.DetectAfterBreak = override.DetectAfterBreak
	} else {
		merged.DetectAfterBreak = base.DetectAfterBreak
	}

	if override.DetectAfterContinue != nil {
		merged.DetectAfterContinue = override.DetectAfterContinue
	} else {
		merged.DetectAfterContinue = base.DetectAfterContinue
	}

	if override.DetectAfterRaise != nil {
		merged.DetectAfterRaise = override.DetectAfterRaise
	} else {
		merged.DetectAfterRaise = base.DetectAfterRaise
	}

	if override.DetectUnreachableBranches != nil {
		merged.DetectUnreachableBranches = override.DetectUnreachableBranches
	} else {
		merged.DetectUnreachableBranches = base.DetectUnreachableBranches
	}

	// ContextLines - override only if explicitly set (non-zero)
	// 0 means "use config file or default value"
	if override.ContextLines > 0 {
		merged.ContextLines = override.ContextLines
	}

	// Recursive - preserve override value
	merged.Recursive = override.Recursive

	// Array values - override if provided
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
		ShowContext:               domain.BoolPtr(cfg.DeadCode.ShowContext),
		ContextLines:              cfg.DeadCode.ContextLines,
		MinSeverity:               minSeverity,
		SortBy:                    sortBy,
		Recursive:                 cfg.Analysis.Recursive,
		IncludePatterns:           cfg.Analysis.IncludePatterns,
		ExcludePatterns:           cfg.Analysis.ExcludePatterns,
		IgnorePatterns:            cfg.DeadCode.IgnorePatterns,
		DetectAfterReturn:         domain.BoolPtr(cfg.DeadCode.DetectAfterReturn),
		DetectAfterBreak:          domain.BoolPtr(cfg.DeadCode.DetectAfterBreak),
		DetectAfterContinue:       domain.BoolPtr(cfg.DeadCode.DetectAfterContinue),
		DetectAfterRaise:          domain.BoolPtr(cfg.DeadCode.DetectAfterRaise),
		DetectUnreachableBranches: domain.BoolPtr(cfg.DeadCode.DetectUnreachableBranches),
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

// SaveConfig saves dead code configuration to a TOML file
func (cl *DeadCodeConfigurationLoaderImpl) SaveConfig(req *domain.DeadCodeRequest, path string) error {
	if req == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Convert request back to config format
	cfg := cl.requestToConfig(req)

	// Use the TOML-based SaveConfig
	return config.SaveConfig(cfg, path)
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
	cfg.DeadCode.ShowContext = domain.BoolValue(req.ShowContext, false)
	cfg.DeadCode.ContextLines = req.ContextLines
	cfg.DeadCode.DetectAfterReturn = domain.BoolValue(req.DetectAfterReturn, true)
	cfg.DeadCode.DetectAfterBreak = domain.BoolValue(req.DetectAfterBreak, true)
	cfg.DeadCode.DetectAfterContinue = domain.BoolValue(req.DetectAfterContinue, true)
	cfg.DeadCode.DetectAfterRaise = domain.BoolValue(req.DetectAfterRaise, true)
	cfg.DeadCode.DetectUnreachableBranches = domain.BoolValue(req.DetectUnreachableBranches, true)
	cfg.DeadCode.IgnorePatterns = req.IgnorePatterns

	// Set analysis config
	cfg.Analysis.Recursive = req.Recursive
	cfg.Analysis.IncludePatterns = req.IncludePatterns
	cfg.Analysis.ExcludePatterns = req.ExcludePatterns

	return cfg
}

// pyscnConfigToUnifiedConfig converts PyscnConfig to unified Config format
//
// Configuration priority (lower priority to higher priority):
//  1. DefaultConfig() - base defaults
//  2. Clone-specific legacy fields (Input.*, Output.*) - backward compatibility
//  3. General sections ([analysis], [output]) - override legacy if set
//  4. Feature-specific sections ([dead_code], [cbo], etc.) - highest priority
func (cl *DeadCodeConfigurationLoaderImpl) pyscnConfigToUnifiedConfig(pyscnCfg *config.PyscnConfig) *config.Config {
	cfg := config.DefaultConfig()

	// Step 1: Map clone-specific legacy fields (backward compatibility)
	// These are from [clone.input] and [clone.output] sections
	cfg.Analysis.IncludePatterns = pyscnCfg.Input.IncludePatterns
	cfg.Analysis.ExcludePatterns = pyscnCfg.Input.ExcludePatterns
	cfg.Analysis.Recursive = domain.BoolValue(pyscnCfg.Input.Recursive, true)
	cfg.Output.Format = pyscnCfg.Output.Format
	cfg.Output.ShowDetails = domain.BoolValue(pyscnCfg.Output.ShowDetails, false)

	// Step 2: Map feature-specific settings from [dead_code] section
	cfg.DeadCode.Enabled = domain.BoolValue(pyscnCfg.DeadCodeEnabled, true)
	cfg.DeadCode.MinSeverity = pyscnCfg.DeadCodeMinSeverity
	cfg.DeadCode.ShowContext = domain.BoolValue(pyscnCfg.DeadCodeShowContext, false)
	cfg.DeadCode.ContextLines = pyscnCfg.DeadCodeContextLines
	cfg.DeadCode.SortBy = pyscnCfg.DeadCodeSortBy
	cfg.DeadCode.DetectAfterReturn = domain.BoolValue(pyscnCfg.DeadCodeDetectAfterReturn, true)
	cfg.DeadCode.DetectAfterBreak = domain.BoolValue(pyscnCfg.DeadCodeDetectAfterBreak, true)
	cfg.DeadCode.DetectAfterContinue = domain.BoolValue(pyscnCfg.DeadCodeDetectAfterContinue, true)
	cfg.DeadCode.DetectAfterRaise = domain.BoolValue(pyscnCfg.DeadCodeDetectAfterRaise, true)
	cfg.DeadCode.DetectUnreachableBranches = domain.BoolValue(pyscnCfg.DeadCodeDetectUnreachableBranches, true)
	cfg.DeadCode.IgnorePatterns = pyscnCfg.DeadCodeIgnorePatterns

	// Step 3: Apply general [analysis] section overrides (highest priority for analysis settings)
	// Only override if explicitly set (non-empty/non-zero values)
	if len(pyscnCfg.AnalysisIncludePatterns) > 0 {
		cfg.Analysis.IncludePatterns = pyscnCfg.AnalysisIncludePatterns
	}
	if len(pyscnCfg.AnalysisExcludePatterns) > 0 {
		cfg.Analysis.ExcludePatterns = pyscnCfg.AnalysisExcludePatterns
	}
	// Only override if explicitly set (non-nil)
	if pyscnCfg.AnalysisRecursive != nil {
		cfg.Analysis.Recursive = *pyscnCfg.AnalysisRecursive
	}
	cfg.Analysis.FollowSymlinks = domain.BoolValue(pyscnCfg.AnalysisFollowSymlinks, false)

	// Step 4: Apply general [output] section overrides (highest priority for output settings)
	// Only override if explicitly set (non-empty values)
	if pyscnCfg.OutputFormat != "" {
		cfg.Output.Format = pyscnCfg.OutputFormat
	}
	if pyscnCfg.OutputSortBy != "" {
		cfg.Output.SortBy = pyscnCfg.OutputSortBy
	}
	if pyscnCfg.OutputDirectory != "" {
		cfg.Output.Directory = pyscnCfg.OutputDirectory
	}
	// Only override if explicitly set (non-nil)
	if pyscnCfg.OutputShowDetails != nil {
		cfg.Output.ShowDetails = *pyscnCfg.OutputShowDetails
	}

	return cfg
}
