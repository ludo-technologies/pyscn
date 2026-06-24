package service

import (
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/ludo-technologies/pyscn/domain"
)

// CommunityFormatter formats community analysis results for output.
type CommunityFormatter struct{}

// NewCommunityFormatter creates a community analysis formatter.
func NewCommunityFormatter() *CommunityFormatter {
	return &CommunityFormatter{}
}

// Format formats community analysis results according to the requested output format.
func (f *CommunityFormatter) Format(response *domain.CommunityAnalysisResult, format domain.OutputFormat) (string, error) {
	if response == nil {
		return "", fmt.Errorf("community analysis result is nil")
	}

	switch format {
	case domain.OutputFormatText:
		return f.formatText(response), nil
	case domain.OutputFormatJSON:
		return f.formatJSON(response)
	default:
		return "", fmt.Errorf("community formatter does not yet support format %s", format)
	}
}

// Write formats and writes community analysis output.
func (f *CommunityFormatter) Write(response *domain.CommunityAnalysisResult, format domain.OutputFormat, writer io.Writer) error {
	content, err := f.Format(response, format)
	if err != nil {
		return err
	}
	_, err = io.WriteString(writer, content)
	return err
}

func (f *CommunityFormatter) formatText(response *domain.CommunityAnalysisResult) string {
	return fmt.Sprintf(
		"Community analysis: %d communities (modularity %.3f, algorithm %s)\n",
		response.TotalCommunities,
		response.Modularity,
		response.Algorithm,
	)
}

func (f *CommunityFormatter) formatJSON(response *domain.CommunityAnalysisResult) (string, error) {
	normalized := normalizeCommunityResult(response)
	return EncodeJSON(normalized)
}

// normalizeCommunityResult returns a schema-stable copy with sorted lists and rounded ratios.
func normalizeCommunityResult(response *domain.CommunityAnalysisResult) *domain.CommunityAnalysisResult {
	if response == nil {
		return nil
	}

	out := *response
	out.Modularity = roundCommunityFloat(response.Modularity)

	communities := make([]domain.CommunityMetrics, len(response.Communities))
	copy(communities, response.Communities)
	sort.Slice(communities, func(i, j int) bool {
		return communities[i].ID < communities[j].ID
	})

	for i := range communities {
		communities[i].Modules = sortedStringCopy(communities[i].Modules)
		communities[i].Packages = sortedStringCopy(communities[i].Packages)
		communities[i].ExternalDependencyRatio = roundCommunityFloat(communities[i].ExternalDependencyRatio)
	}
	out.Communities = communities

	bridges := make([]domain.BridgeModule, len(response.BridgeModules))
	copy(bridges, response.BridgeModules)
	sort.Slice(bridges, func(i, j int) bool {
		if bridges[i].Module == bridges[j].Module {
			return bridges[i].Community < bridges[j].Community
		}
		return bridges[i].Module < bridges[j].Module
	})
	for i := range bridges {
		bridges[i].TargetCommunities = sortedStringCopy(bridges[i].TargetCommunities)
	}
	out.BridgeModules = bridges

	return &out
}

func sortedStringCopy(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func roundCommunityFloat(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return math.Round(value*10000) / 10000
}
