package main

import (
	"github.com/pyqol/pyqol/domain"
	"github.com/pyqol/pyqol/service"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	explicitFlags := make(map[string]bool)
	if cmd != nil {
		cmd.Flags().Visit(func(f *pflag.Flag) {
			explicitFlags[f.Name] = true
		})
	}

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

	// Helper function to check if a flag was explicitly set
	wasExplicitlySet := func(flagName string) bool {
		if w.explicitFlags == nil {
			return false
		}
		return w.explicitFlags[flagName]
	}

	// Always override paths as they come from command arguments
	if len(requestReq.Paths) > 0 {
		merged.Paths = requestReq.Paths
	}

	// Override boolean flags only if explicitly set
	if wasExplicitlySet("recursive") {
		merged.Recursive = requestReq.Recursive
	}
	if wasExplicitlySet("show-details") {
		merged.ShowDetails = requestReq.ShowDetails
	}
	if wasExplicitlySet("show-content") {
		merged.ShowContent = requestReq.ShowContent
	}
	if wasExplicitlySet("group") {
		merged.GroupClones = requestReq.GroupClones
	}
	if wasExplicitlySet("ignore-literals") {
		merged.IgnoreLiterals = requestReq.IgnoreLiterals
	}
	if wasExplicitlySet("ignore-identifiers") {
		merged.IgnoreIdentifiers = requestReq.IgnoreIdentifiers
	}

	// Override numeric values only if explicitly set
	if wasExplicitlySet("min-lines") {
		merged.MinLines = requestReq.MinLines
	}
	if wasExplicitlySet("min-nodes") {
		merged.MinNodes = requestReq.MinNodes
	}
	if wasExplicitlySet("similarity") {
		merged.SimilarityThreshold = requestReq.SimilarityThreshold
	}
	if wasExplicitlySet("max-edit-distance") {
		merged.MaxEditDistance = requestReq.MaxEditDistance
	}

	// Override threshold values only if explicitly set
	if wasExplicitlySet("type1-threshold") {
		merged.Type1Threshold = requestReq.Type1Threshold
	}
	if wasExplicitlySet("type2-threshold") {
		merged.Type2Threshold = requestReq.Type2Threshold
	}
	if wasExplicitlySet("type3-threshold") {
		merged.Type3Threshold = requestReq.Type3Threshold
	}
	if wasExplicitlySet("type4-threshold") {
		merged.Type4Threshold = requestReq.Type4Threshold
	}

	// Override output settings only if explicitly set
	if wasExplicitlySet("format") {
		merged.OutputFormat = requestReq.OutputFormat
	}
	merged.OutputWriter = requestReq.OutputWriter // Always use from request
	if wasExplicitlySet("sort") {
		merged.SortBy = requestReq.SortBy
	}

	// Override similarity filters only if explicitly set
	if wasExplicitlySet("min-similarity") {
		merged.MinSimilarity = requestReq.MinSimilarity
	}
	if wasExplicitlySet("max-similarity") {
		merged.MaxSimilarity = requestReq.MaxSimilarity
	}

	// Override patterns only if explicitly set
	if wasExplicitlySet("include") && len(requestReq.IncludePatterns) > 0 {
		merged.IncludePatterns = requestReq.IncludePatterns
	}
	if wasExplicitlySet("exclude") && len(requestReq.ExcludePatterns) > 0 {
		merged.ExcludePatterns = requestReq.ExcludePatterns
	}

	// Override clone types only if explicitly set
	if wasExplicitlySet("types") && len(requestReq.CloneTypes) > 0 {
		merged.CloneTypes = requestReq.CloneTypes
	}

	return &merged
}