package service

import (
	"github.com/pyqol/pyqol/domain"
	"github.com/pyqol/pyqol/internal/config"
)

// DeadCodeConfigurationLoaderWithFlags wraps dead code configuration loading with explicit flag tracking
type DeadCodeConfigurationLoaderWithFlags struct {
	loader      *DeadCodeConfigurationLoaderImpl
	flagTracker *config.FlagTracker
}

// NewDeadCodeConfigurationLoaderWithFlags creates a new dead code configuration loader that tracks explicit flags
func NewDeadCodeConfigurationLoaderWithFlags(explicitFlags map[string]bool) *DeadCodeConfigurationLoaderWithFlags {
	return &DeadCodeConfigurationLoaderWithFlags{
		loader:      NewDeadCodeConfigurationLoader(),
		flagTracker: config.NewFlagTrackerWithFlags(explicitFlags),
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
	if cl.flagTracker.WasSet("format") {
		merged.OutputFormat = override.OutputFormat
	}
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	merged.ShowContext = cl.flagTracker.MergeBool(merged.ShowContext, override.ShowContext, "show-context")
	merged.ContextLines = cl.flagTracker.MergeInt(merged.ContextLines, override.ContextLines, "context-lines")

	// Filtering and sorting
	if cl.flagTracker.WasSet("min-severity") {
		merged.MinSeverity = override.MinSeverity
	}
	if cl.flagTracker.WasSet("sort") {
		merged.SortBy = override.SortBy
	}

	// Config path is always from override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// Dead code detection options
	merged.DetectAfterReturn = cl.flagTracker.MergeBool(merged.DetectAfterReturn, override.DetectAfterReturn, "detect-after-return")
	merged.DetectAfterBreak = cl.flagTracker.MergeBool(merged.DetectAfterBreak, override.DetectAfterBreak, "detect-after-break")
	merged.DetectAfterContinue = cl.flagTracker.MergeBool(merged.DetectAfterContinue, override.DetectAfterContinue, "detect-after-continue")
	merged.DetectAfterRaise = cl.flagTracker.MergeBool(merged.DetectAfterRaise, override.DetectAfterRaise, "detect-after-raise")
	merged.DetectUnreachableBranches = cl.flagTracker.MergeBool(merged.DetectUnreachableBranches, override.DetectUnreachableBranches, "detect-unreachable-branches")

	// For recursive, only override if explicitly set
	merged.Recursive = cl.flagTracker.MergeBool(merged.Recursive, override.Recursive, "recursive")

	// Patterns
	merged.IncludePatterns = cl.flagTracker.MergeStringSlice(merged.IncludePatterns, override.IncludePatterns, "include")
	merged.ExcludePatterns = cl.flagTracker.MergeStringSlice(merged.ExcludePatterns, override.ExcludePatterns, "exclude")
	merged.IgnorePatterns = cl.flagTracker.MergeStringSlice(merged.IgnorePatterns, override.IgnorePatterns, "ignore")

	return &merged
}

// FindDefaultConfigFile looks for .pyqol.yaml in the current directory
func (cl *DeadCodeConfigurationLoaderWithFlags) FindDefaultConfigFile() string {
	return cl.loader.FindDefaultConfigFile()
}