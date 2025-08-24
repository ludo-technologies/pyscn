package main

import (
	"github.com/pyqol/pyqol/domain"
	"github.com/pyqol/pyqol/internal/config"
	"github.com/pyqol/pyqol/service"
	"github.com/spf13/cobra"
)

// CloneConfigWrapper wraps clone configuration loading with explicit flag tracking
type CloneConfigWrapper struct {
	loader        *service.CloneConfigurationLoader
	explicitFlags map[string]bool
	request       *domain.CloneRequest
}

// NewCloneConfigWrapper creates a new clone configuration wrapper
func NewCloneConfigWrapper(cmd *cobra.Command, request *domain.CloneRequest) *CloneConfigWrapper {
	// Track which flags were explicitly set by the user
	explicitFlags := GetExplicitFlags(cmd)

	return &CloneConfigWrapper{
		loader:        service.NewCloneConfigurationLoader(),
		explicitFlags: explicitFlags,
		request:       request,
	}
}

// MergeWithConfig merges the request with configuration from file
func (w *CloneConfigWrapper) MergeWithConfig() *domain.CloneRequest {
	if w.request.ConfigPath == "" {
		// Try to load default config
		defaultConfig := w.loader.GetDefaultCloneConfig()
		if defaultConfig != nil {
			return w.mergeConfiguration(*defaultConfig, *w.request)
		}
		return w.request
	}

	// Load specified config
	configReq, err := w.loader.LoadCloneConfig(w.request.ConfigPath)
	if err != nil {
		// If config loading fails, return original request
		return w.request
	}

	return w.mergeConfiguration(*configReq, *w.request)
}

// mergeConfiguration merges configuration from file with request parameters
func (w *CloneConfigWrapper) mergeConfiguration(configReq, requestReq domain.CloneRequest) *domain.CloneRequest {
	// Start with configuration from file
	merged := configReq


	// Always override paths as they come from command arguments
	if len(requestReq.Paths) > 0 {
		merged.Paths = requestReq.Paths
	}

	// Override boolean flags only if explicitly set
	merged.Recursive = config.MergeBool(merged.Recursive, requestReq.Recursive, "recursive", w.explicitFlags)
	merged.ShowDetails = config.MergeBool(merged.ShowDetails, requestReq.ShowDetails, "show-details", w.explicitFlags)
	merged.ShowContent = config.MergeBool(merged.ShowContent, requestReq.ShowContent, "show-content", w.explicitFlags)
	merged.GroupClones = config.MergeBool(merged.GroupClones, requestReq.GroupClones, "group", w.explicitFlags)
	merged.IgnoreLiterals = config.MergeBool(merged.IgnoreLiterals, requestReq.IgnoreLiterals, "ignore-literals", w.explicitFlags)
	merged.IgnoreIdentifiers = config.MergeBool(merged.IgnoreIdentifiers, requestReq.IgnoreIdentifiers, "ignore-identifiers", w.explicitFlags)

	// Override numeric values only if explicitly set
	merged.MinLines = config.MergeInt(merged.MinLines, requestReq.MinLines, "min-lines", w.explicitFlags)
	merged.MinNodes = config.MergeInt(merged.MinNodes, requestReq.MinNodes, "min-nodes", w.explicitFlags)
	merged.SimilarityThreshold = config.MergeFloat64(merged.SimilarityThreshold, requestReq.SimilarityThreshold, "similarity", w.explicitFlags)
	merged.MaxEditDistance = config.MergeFloat64(merged.MaxEditDistance, requestReq.MaxEditDistance, "max-edit-distance", w.explicitFlags)

	// Override threshold values only if explicitly set
	merged.Type1Threshold = config.MergeFloat64(merged.Type1Threshold, requestReq.Type1Threshold, "type1-threshold", w.explicitFlags)
	merged.Type2Threshold = config.MergeFloat64(merged.Type2Threshold, requestReq.Type2Threshold, "type2-threshold", w.explicitFlags)
	merged.Type3Threshold = config.MergeFloat64(merged.Type3Threshold, requestReq.Type3Threshold, "type3-threshold", w.explicitFlags)
	merged.Type4Threshold = config.MergeFloat64(merged.Type4Threshold, requestReq.Type4Threshold, "type4-threshold", w.explicitFlags)

	// Override output settings only if explicitly set
	if config.WasExplicitlySet(w.explicitFlags, "format") {
		merged.OutputFormat = requestReq.OutputFormat
	}
	merged.OutputWriter = requestReq.OutputWriter // Always use from request
	if config.WasExplicitlySet(w.explicitFlags, "sort") {
		merged.SortBy = requestReq.SortBy
	}

	// Override similarity filters only if explicitly set
	merged.MinSimilarity = config.MergeFloat64(merged.MinSimilarity, requestReq.MinSimilarity, "min-similarity", w.explicitFlags)
	merged.MaxSimilarity = config.MergeFloat64(merged.MaxSimilarity, requestReq.MaxSimilarity, "max-similarity", w.explicitFlags)

	// Override patterns only if explicitly set
	merged.IncludePatterns = config.MergeStringSlice(merged.IncludePatterns, requestReq.IncludePatterns, "include", w.explicitFlags)
	merged.ExcludePatterns = config.MergeStringSlice(merged.ExcludePatterns, requestReq.ExcludePatterns, "exclude", w.explicitFlags)

	// Override clone types only if explicitly set
	if config.WasExplicitlySet(w.explicitFlags, "types") && len(requestReq.CloneTypes) > 0 {
		merged.CloneTypes = requestReq.CloneTypes
	}

	return &merged
}