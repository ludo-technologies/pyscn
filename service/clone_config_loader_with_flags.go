package service

import (
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"
)

// CloneConfigurationLoaderWithFlags wraps clone configuration loading with explicit flag tracking
type CloneConfigurationLoaderWithFlags struct {
	loader      *CloneConfigurationLoader
	flagTracker *config.FlagTracker
}

// NewCloneConfigurationLoaderWithFlags creates a new clone configuration loader that tracks explicit flags
func NewCloneConfigurationLoaderWithFlags(explicitFlags map[string]bool) *CloneConfigurationLoaderWithFlags {
	return &CloneConfigurationLoaderWithFlags{
		loader:      NewCloneConfigurationLoader(),
		flagTracker: config.NewFlagTrackerWithFlags(explicitFlags),
	}
}

// LoadCloneConfig loads clone configuration from the specified path
func (cl *CloneConfigurationLoaderWithFlags) LoadCloneConfig(path string) (*domain.CloneRequest, error) {
	return cl.loader.LoadCloneConfig(path)
}

// GetDefaultCloneConfig loads the default clone configuration
func (cl *CloneConfigurationLoaderWithFlags) GetDefaultCloneConfig() *domain.CloneRequest {
	return cl.loader.GetDefaultCloneConfig()
}

// MergeConfig merges CLI flags with configuration file, respecting explicit flags
func (cl *CloneConfigurationLoaderWithFlags) MergeConfig(base *domain.CloneRequest, override *domain.CloneRequest) *domain.CloneRequest {
	merged := cl.loader.MergeConfig(base, override)
	if merged == nil || override == nil {
		return merged
	}

	// Boolean flags
	if cl.flagTracker.WasSet("recursive") {
		merged.Recursive = override.Recursive
	}
	if cl.flagTracker.WasSet("show-details") {
		merged.ShowDetails = override.ShowDetails
	}
	if cl.flagTracker.WasSet("show-content") {
		merged.ShowContent = override.ShowContent
	}
	if cl.flagTracker.WasSet("group") {
		merged.GroupClones = override.GroupClones
	}
	if cl.flagTracker.WasSet("ignore-literals") {
		merged.IgnoreLiterals = override.IgnoreLiterals
	}
	if cl.flagTracker.WasSet("ignore-identifiers") {
		merged.IgnoreIdentifiers = override.IgnoreIdentifiers
	}

	// Numeric values
	if cl.flagTracker.WasSet("min-lines") {
		merged.MinLines = override.MinLines
	}
	if cl.flagTracker.WasSet("min-nodes") {
		merged.MinNodes = override.MinNodes
	}
	if cl.flagTracker.WasSet("similarity") {
		merged.SimilarityThreshold = override.SimilarityThreshold
	}
	if cl.flagTracker.WasSet("max-edit-distance") {
		merged.MaxEditDistance = override.MaxEditDistance
	}

	// Threshold values
	if cl.flagTracker.WasSet("type1-threshold") {
		merged.Type1Threshold = override.Type1Threshold
	}
	if cl.flagTracker.WasSet("type2-threshold") {
		merged.Type2Threshold = override.Type2Threshold
	}
	if cl.flagTracker.WasSet("type3-threshold") {
		merged.Type3Threshold = override.Type3Threshold
	}
	if cl.flagTracker.WasSet("type4-threshold") {
		merged.Type4Threshold = override.Type4Threshold
	}

	// Output settings - always use override format when explicitly set
	// Since we removed --format flag, check for individual format flags or non-text format
	if override.OutputFormat != "" {
		// If a specific format was set (not text), use it
		if override.OutputFormat != domain.OutputFormatText {
			merged.OutputFormat = override.OutputFormat
		} else if cl.flagTracker.WasSet("html") || cl.flagTracker.WasSet("json") ||
			cl.flagTracker.WasSet("csv") || cl.flagTracker.WasSet("yaml") {
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

	// Always preserve NoOpen from override
	merged.NoOpen = override.NoOpen

	if cl.flagTracker.WasSet("sort") {
		merged.SortBy = override.SortBy
	}

	// Similarity filters
	if cl.flagTracker.WasSet("min-similarity") {
		merged.MinSimilarity = override.MinSimilarity
	}
	if cl.flagTracker.WasSet("max-similarity") {
		merged.MaxSimilarity = override.MaxSimilarity
	}

	// Patterns
	if cl.flagTracker.WasSet("include") && len(override.IncludePatterns) > 0 {
		merged.IncludePatterns = append([]string(nil), override.IncludePatterns...)
	}
	if cl.flagTracker.WasSet("exclude") && len(override.ExcludePatterns) > 0 {
		merged.ExcludePatterns = append([]string(nil), override.ExcludePatterns...)
	}

	// Clone types
	if cl.flagTracker.WasSet("types") && len(override.CloneTypes) > 0 {
		merged.CloneTypes = append([]domain.CloneType(nil), override.CloneTypes...)
	}

	// Config path is always from override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	return merged
}

// FindDefaultConfigFile looks for .pyscn.toml in the current directory
func (cl *CloneConfigurationLoaderWithFlags) FindDefaultConfigFile() string {
	return cl.loader.FindDefaultConfigFile()
}

// SaveCloneConfig saves clone configuration to the specified path
func (cl *CloneConfigurationLoaderWithFlags) SaveCloneConfig(config *domain.CloneRequest, path string) error {
	return cl.loader.SaveCloneConfig(config, path)
}
