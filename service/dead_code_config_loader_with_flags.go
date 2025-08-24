package service

import (
	"github.com/pyqol/pyqol/domain"
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

	// Helper function to check if a flag was explicitly set
	wasExplicitlySet := func(flagName string) bool {
		if cl.explicitFlags == nil {
			return false
		}
		return cl.explicitFlags[flagName]
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
	if wasExplicitlySet("show-context") {
		merged.ShowContext = override.ShowContext
	}
	if wasExplicitlySet("context-lines") {
		merged.ContextLines = override.ContextLines
	}

	// Filtering and sorting
	if wasExplicitlySet("min-severity") {
		merged.MinSeverity = override.MinSeverity
	}
	if wasExplicitlySet("sort") {
		merged.SortBy = override.SortBy
	}

	// Config path is always from override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// Dead code detection options
	if wasExplicitlySet("detect-after-return") {
		merged.DetectAfterReturn = override.DetectAfterReturn
	}
	if wasExplicitlySet("detect-after-break") {
		merged.DetectAfterBreak = override.DetectAfterBreak
	}
	if wasExplicitlySet("detect-after-continue") {
		merged.DetectAfterContinue = override.DetectAfterContinue
	}
	if wasExplicitlySet("detect-after-raise") {
		merged.DetectAfterRaise = override.DetectAfterRaise
	}
	if wasExplicitlySet("detect-unreachable-branches") {
		merged.DetectUnreachableBranches = override.DetectUnreachableBranches
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
	if wasExplicitlySet("ignore") && len(override.IgnorePatterns) > 0 {
		merged.IgnorePatterns = override.IgnorePatterns
	}

	return &merged
}

// FindDefaultConfigFile looks for .pyqol.yaml in the current directory
func (cl *DeadCodeConfigurationLoaderWithFlags) FindDefaultConfigFile() string {
	return cl.loader.FindDefaultConfigFile()
}