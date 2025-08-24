package service

import (
	"github.com/pyqol/pyqol/domain"
	"github.com/pyqol/pyqol/internal/config"
)

// DeadCodeConfigurationLoaderWithFlags wraps dead code configuration loading with explicit flag tracking
type DeadCodeConfigurationLoaderWithFlags struct {
	loader        *DeadCodeConfigurationLoaderImpl
	explicitFlags map[string]bool
}

// NewDeadCodeConfigurationLoaderWithFlags creates a new dead code configuration loader that tracks explicit flags
func NewDeadCodeConfigurationLoaderWithFlags(explicitFlags map[string]bool) *DeadCodeConfigurationLoaderWithFlags {
	return &DeadCodeConfigurationLoaderWithFlags{
		loader:        NewDeadCodeConfigurationLoader(),
		explicitFlags: explicitFlags,
	}
}

// LoadConfig loads dead code configuration from the specified path
func (cl *DeadCodeConfigurationLoaderWithFlags) LoadConfig(path string) (*domain.DeadCodeRequest, error) {
	return cl.loader.LoadConfig(path)
}

// LoadDefaultConfig loads the default dead code configuration
func (cl *DeadCodeConfigurationLoaderWithFlags) LoadDefaultConfig() *domain.DeadCodeRequest {
	return cl.loader.LoadDefaultConfig()
}

// MergeConfig merges CLI flags with configuration file, respecting explicit flags
func (cl *DeadCodeConfigurationLoaderWithFlags) MergeConfig(base *domain.DeadCodeRequest, override *domain.DeadCodeRequest) *domain.DeadCodeRequest {
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
	if config.WasExplicitlySet(cl.explicitFlags, "format") {
		merged.OutputFormat = override.OutputFormat
	}
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	merged.ShowContext = config.MergeBool(merged.ShowContext, override.ShowContext, "show-context", cl.explicitFlags)
	merged.ContextLines = config.MergeInt(merged.ContextLines, override.ContextLines, "context-lines", cl.explicitFlags)

	// Filtering and sorting
	if config.WasExplicitlySet(cl.explicitFlags, "min-severity") {
		merged.MinSeverity = override.MinSeverity
	}
	if config.WasExplicitlySet(cl.explicitFlags, "sort") {
		merged.SortBy = override.SortBy
	}

	// Config path is always from override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// Dead code detection options
	merged.DetectAfterReturn = config.MergeBool(merged.DetectAfterReturn, override.DetectAfterReturn, "detect-after-return", cl.explicitFlags)
	merged.DetectAfterBreak = config.MergeBool(merged.DetectAfterBreak, override.DetectAfterBreak, "detect-after-break", cl.explicitFlags)
	merged.DetectAfterContinue = config.MergeBool(merged.DetectAfterContinue, override.DetectAfterContinue, "detect-after-continue", cl.explicitFlags)
	merged.DetectAfterRaise = config.MergeBool(merged.DetectAfterRaise, override.DetectAfterRaise, "detect-after-raise", cl.explicitFlags)
	merged.DetectUnreachableBranches = config.MergeBool(merged.DetectUnreachableBranches, override.DetectUnreachableBranches, "detect-unreachable-branches", cl.explicitFlags)

	// For recursive, only override if explicitly set
	merged.Recursive = config.MergeBool(merged.Recursive, override.Recursive, "recursive", cl.explicitFlags)

	// Patterns
	merged.IncludePatterns = config.MergeStringSlice(merged.IncludePatterns, override.IncludePatterns, "include", cl.explicitFlags)
	merged.ExcludePatterns = config.MergeStringSlice(merged.ExcludePatterns, override.ExcludePatterns, "exclude", cl.explicitFlags)
	merged.IgnorePatterns = config.MergeStringSlice(merged.IgnorePatterns, override.IgnorePatterns, "ignore", cl.explicitFlags)

	return &merged
}

// FindDefaultConfigFile looks for .pyqol.yaml in the current directory
func (cl *DeadCodeConfigurationLoaderWithFlags) FindDefaultConfigFile() string {
	return cl.loader.FindDefaultConfigFile()
}