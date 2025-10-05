package config

import (
	"io"

	"github.com/ludo-technologies/pyscn/domain"
)

// ToCloneRequest converts unified CloneConfig to domain's CloneRequest
// This maintains backward compatibility with the domain package
func (c *CloneConfig) ToCloneRequest(outputWriter io.Writer) *domain.CloneRequest {
	// Convert clone types from strings to domain.CloneType
	var cloneTypes []domain.CloneType
	for _, typeStr := range c.Filtering.EnabledCloneTypes {
		switch typeStr {
		case "type1":
			cloneTypes = append(cloneTypes, domain.Type1Clone)
		case "type2":
			cloneTypes = append(cloneTypes, domain.Type2Clone)
		case "type3":
			cloneTypes = append(cloneTypes, domain.Type3Clone)
		case "type4":
			cloneTypes = append(cloneTypes, domain.Type4Clone)
		}
	}

	// Convert output format
	var outputFormat domain.OutputFormat
	switch c.Output.Format {
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
	switch c.Output.SortBy {
	case "size":
		sortBy = domain.SortBySize
	case "location":
		sortBy = domain.SortByLocation
	case "type":
		sortBy = domain.SortByComplexity
	default:
		sortBy = domain.SortBySimilarity
	}

	return &domain.CloneRequest{
		// Input parameters
		Paths:           c.Input.Paths,
		Recursive:       c.Input.Recursive,
		IncludePatterns: c.Input.IncludePatterns,
		ExcludePatterns: c.Input.ExcludePatterns,

		// Analysis configuration
		MinLines:            c.Analysis.MinLines,
		MinNodes:            c.Analysis.MinNodes,
		SimilarityThreshold: c.Thresholds.SimilarityThreshold,
		MaxEditDistance:     c.Analysis.MaxEditDistance,
		IgnoreLiterals:      c.Analysis.IgnoreLiterals,
		IgnoreIdentifiers:   c.Analysis.IgnoreIdentifiers,

		// Type-specific thresholds
		Type1Threshold: c.Thresholds.Type1Threshold,
		Type2Threshold: c.Thresholds.Type2Threshold,
		Type3Threshold: c.Thresholds.Type3Threshold,
		Type4Threshold: c.Thresholds.Type4Threshold,

		// Output configuration
		OutputFormat: outputFormat,
		OutputWriter: outputWriter,
		ShowDetails:  c.Output.ShowDetails,
		ShowContent:  c.Output.ShowContent,
		SortBy:       sortBy,
		GroupClones:  c.Output.GroupClones,

		// Filtering
		MinSimilarity: c.Filtering.MinSimilarity,
		MaxSimilarity: c.Filtering.MaxSimilarity,
		CloneTypes:    cloneTypes,
	}
}

// FromCloneRequest creates unified CloneConfig from domain's CloneRequest
func FromCloneRequest(request *domain.CloneRequest) *CloneConfig {
	config := DefaultCloneConfig()

	// Input parameters
	config.Input.Paths = request.Paths
	config.Input.Recursive = request.Recursive
	config.Input.IncludePatterns = request.IncludePatterns
	config.Input.ExcludePatterns = request.ExcludePatterns

	// Analysis configuration
	config.Analysis.MinLines = request.MinLines
	config.Analysis.MinNodes = request.MinNodes
	config.Analysis.MaxEditDistance = request.MaxEditDistance
	config.Analysis.IgnoreLiterals = request.IgnoreLiterals
	config.Analysis.IgnoreIdentifiers = request.IgnoreIdentifiers

	// Thresholds
	config.Thresholds.Type1Threshold = request.Type1Threshold
	config.Thresholds.Type2Threshold = request.Type2Threshold
	config.Thresholds.Type3Threshold = request.Type3Threshold
	config.Thresholds.Type4Threshold = request.Type4Threshold
	config.Thresholds.SimilarityThreshold = request.SimilarityThreshold

	// Output configuration
	switch request.OutputFormat {
	case domain.OutputFormatJSON:
		config.Output.Format = "json"
	case domain.OutputFormatYAML:
		config.Output.Format = "yaml"
	case domain.OutputFormatCSV:
		config.Output.Format = "csv"
	default:
		config.Output.Format = "text"
	}

	config.Output.ShowDetails = request.ShowDetails
	config.Output.ShowContent = request.ShowContent
	config.Output.GroupClones = request.GroupClones
	config.Output.Writer = request.OutputWriter

	switch request.SortBy {
	case domain.SortBySize:
		config.Output.SortBy = "size"
	case domain.SortByLocation:
		config.Output.SortBy = "location"
	case domain.SortByComplexity:
		config.Output.SortBy = "type"
	default:
		config.Output.SortBy = "similarity"
	}

	// Filtering
	config.Filtering.MinSimilarity = request.MinSimilarity
	config.Filtering.MaxSimilarity = request.MaxSimilarity

	// Convert clone types to strings
	config.Filtering.EnabledCloneTypes = make([]string, 0, len(request.CloneTypes))
	for _, cloneType := range request.CloneTypes {
		switch cloneType {
		case domain.Type1Clone:
			config.Filtering.EnabledCloneTypes = append(config.Filtering.EnabledCloneTypes, "type1")
		case domain.Type2Clone:
			config.Filtering.EnabledCloneTypes = append(config.Filtering.EnabledCloneTypes, "type2")
		case domain.Type3Clone:
			config.Filtering.EnabledCloneTypes = append(config.Filtering.EnabledCloneTypes, "type3")
		case domain.Type4Clone:
			config.Filtering.EnabledCloneTypes = append(config.Filtering.EnabledCloneTypes, "type4")
		}
	}

	return config
}
