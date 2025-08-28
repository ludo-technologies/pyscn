package service

import (
	"github.com/pyqol/pyqol/domain"
	"github.com/pyqol/pyqol/internal/config"
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

	// Boolean flags
	merged.Recursive = cl.flagTracker.MergeBool(merged.Recursive, override.Recursive, "recursive")
	merged.ShowDetails = cl.flagTracker.MergeBool(merged.ShowDetails, override.ShowDetails, "show-details")
	merged.ShowContent = cl.flagTracker.MergeBool(merged.ShowContent, override.ShowContent, "show-content")
	merged.GroupClones = cl.flagTracker.MergeBool(merged.GroupClones, override.GroupClones, "group")
	merged.IgnoreLiterals = cl.flagTracker.MergeBool(merged.IgnoreLiterals, override.IgnoreLiterals, "ignore-literals")
	merged.IgnoreIdentifiers = cl.flagTracker.MergeBool(merged.IgnoreIdentifiers, override.IgnoreIdentifiers, "ignore-identifiers")

	// Numeric values
	merged.MinLines = cl.flagTracker.MergeInt(merged.MinLines, override.MinLines, "min-lines")
	merged.MinNodes = cl.flagTracker.MergeInt(merged.MinNodes, override.MinNodes, "min-nodes")
	merged.SimilarityThreshold = cl.flagTracker.MergeFloat64(merged.SimilarityThreshold, override.SimilarityThreshold, "similarity")
	merged.MaxEditDistance = cl.flagTracker.MergeFloat64(merged.MaxEditDistance, override.MaxEditDistance, "max-edit-distance")

	// Threshold values
	merged.Type1Threshold = cl.flagTracker.MergeFloat64(merged.Type1Threshold, override.Type1Threshold, "type1-threshold")
	merged.Type2Threshold = cl.flagTracker.MergeFloat64(merged.Type2Threshold, override.Type2Threshold, "type2-threshold")
	merged.Type3Threshold = cl.flagTracker.MergeFloat64(merged.Type3Threshold, override.Type3Threshold, "type3-threshold")
	merged.Type4Threshold = cl.flagTracker.MergeFloat64(merged.Type4Threshold, override.Type4Threshold, "type4-threshold")

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
	merged.MinSimilarity = cl.flagTracker.MergeFloat64(merged.MinSimilarity, override.MinSimilarity, "min-similarity")
	merged.MaxSimilarity = cl.flagTracker.MergeFloat64(merged.MaxSimilarity, override.MaxSimilarity, "max-similarity")

	// Patterns
	merged.IncludePatterns = cl.flagTracker.MergeStringSlice(merged.IncludePatterns, override.IncludePatterns, "include")
	merged.ExcludePatterns = cl.flagTracker.MergeStringSlice(merged.ExcludePatterns, override.ExcludePatterns, "exclude")

	// Clone types
	if cl.flagTracker.WasSet("types") && len(override.CloneTypes) > 0 {
		merged.CloneTypes = override.CloneTypes
	}

	// Config path is always from override if provided
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	return &merged
}

// FindDefaultConfigFile looks for .pyqol.yaml in the current directory
func (cl *CloneConfigurationLoaderWithFlags) FindDefaultConfigFile() string {
	return cl.loader.FindDefaultConfigFile()
}

// SaveCloneConfig saves clone configuration to the specified path
func (cl *CloneConfigurationLoaderWithFlags) SaveCloneConfig(config *domain.CloneRequest, path string) error {
	return cl.loader.SaveCloneConfig(config, path)
}