package domain

import (
	"context"
	"io"
	"math"
)

// CommunityAnalysisRequest represents a request for module community detection.
type CommunityAnalysisRequest struct {
	// Input files or directories to analyze
	Paths []string

	// SourcePaths preserves the original user-provided paths before file expansion.
	// Used for project-root detection when Paths contains only resolved files.
	SourcePaths []string

	// Output configuration
	OutputFormat OutputFormat
	OutputWriter io.Writer
	OutputPath   string
	NoOpen       bool

	// Configuration
	ConfigPath      string
	Recursive       *bool
	IncludePatterns []string
	ExcludePatterns []string

	// Community detection options
	Algorithm           string
	Scope               string
	MinCommunitySize    int
	IncludeLazyEdges    *bool
	ReportBridgeModules *bool
	Resolution          float64

	// Module graph options
	IncludeStdLib     *bool
	IncludeThirdParty *bool
	FollowRelative    *bool

	// ArchitectureRules supplies configured layers for layer mismatch scoring.
	// Loaded from config when not set explicitly.
	ArchitectureRules *ArchitectureRules
}

// CommunityMetrics describes one detected module community.
type CommunityMetrics struct {
	ID                          string   `json:"id" yaml:"id"`
	Modules                     []string `json:"modules" yaml:"modules"`
	Packages                    []string `json:"packages" yaml:"packages"`
	InternalEdges               int      `json:"internal_edges" yaml:"internal_edges"`
	ExternalEdges               int      `json:"external_edges" yaml:"external_edges"`
	ExternalDependencyRatio     float64  `json:"external_dependency_ratio" yaml:"external_dependency_ratio"`
	IncomingCrossCommunityEdges int      `json:"incoming_cross_community_edges" yaml:"incoming_cross_community_edges"`
	OutgoingCrossCommunityEdges int      `json:"outgoing_cross_community_edges" yaml:"outgoing_cross_community_edges"`
	Size                        int      `json:"size" yaml:"size"`

	// Package mismatch metrics (omitted when package metadata is unavailable).
	DominantPackage  string  `json:"dominant_package,omitempty" yaml:"dominant_package,omitempty"`
	PackageCount     int     `json:"package_count,omitempty" yaml:"package_count,omitempty"`
	PackageAlignment float64 `json:"package_alignment,omitempty" yaml:"package_alignment,omitempty"`

	// Layer mismatch metrics (omitted when architecture layers are not configured).
	DominantLayer  string   `json:"dominant_layer,omitempty" yaml:"dominant_layer,omitempty"`
	LayerCount     int      `json:"layer_count,omitempty" yaml:"layer_count,omitempty"`
	Layers         []string `json:"layers,omitempty" yaml:"layers,omitempty"`
	LayerAlignment *float64 `json:"layer_alignment,omitempty" yaml:"layer_alignment,omitempty"`

	// RiskLevel classifies the community as low/medium/high using documented
	// thresholds (see docs/ANALYZE_SCORING.md). Populated by ScoreCommunityResult.
	RiskLevel string `json:"risk_level,omitempty" yaml:"risk_level,omitempty"`
}

// CommunityModuleDependency is a directed module dependency edge used for graph export.
type CommunityModuleDependency struct {
	From string
	To   string
}

// BridgeModule describes a module that couples multiple communities.
type BridgeModule struct {
	Module              string   `json:"module" yaml:"module"`
	Community           string   `json:"community" yaml:"community"`
	CrossCommunityEdges int      `json:"cross_community_edges" yaml:"cross_community_edges"`
	TargetCommunities   []string `json:"target_communities" yaml:"target_communities"`
}

// CommunityAnalysisResult represents the complete community analysis output.
type CommunityAnalysisResult struct {
	Algorithm        string             `json:"algorithm" yaml:"algorithm"`
	Scope            string             `json:"scope" yaml:"scope"`
	TotalCommunities int                `json:"total_communities" yaml:"total_communities"`
	Modularity       float64            `json:"modularity" yaml:"modularity"`
	Communities      []CommunityMetrics `json:"communities" yaml:"communities"`
	BridgeModules    []BridgeModule     `json:"bridge_modules" yaml:"bridge_modules"`

	// Package mismatch metrics compare inferred communities to declared package boundaries.
	PackageAlignmentScore *float64 `json:"package_alignment_score,omitempty" yaml:"package_alignment_score,omitempty"`
	SplitPackages         []string `json:"split_packages,omitempty" yaml:"split_packages,omitempty"`
	MixedCommunities      []string `json:"mixed_communities,omitempty" yaml:"mixed_communities,omitempty"`

	// Layer mismatch metrics compare inferred communities to configured architecture layers.
	LayerAlignmentScore   *float64 `json:"layer_alignment_score,omitempty" yaml:"layer_alignment_score,omitempty"`
	CrossLayerCommunities []string `json:"cross_layer_communities,omitempty" yaml:"cross_layer_communities,omitempty"`
	LayerBridgeModules    []string `json:"layer_bridge_modules,omitempty" yaml:"layer_bridge_modules,omitempty"`

	// RiskScore is a system-level community risk score (0-100, higher = worse),
	// populated by ScoreCommunityResult. Nil when fewer than two communities were
	// detected (no meaningful modular structure to score).
	RiskScore *int `json:"community_risk_score,omitempty" yaml:"community_risk_score,omitempty"`

	// ModuleDependencies holds directed edges for DOT export and is omitted from JSON/YAML.
	ModuleDependencies []CommunityModuleDependency `json:"-" yaml:"-"`

	// BridgeModuleCount is the number of detected bridge modules from the
	// underlying analysis. It is tracked independently of BridgeModules (which is
	// only populated when bridge reporting is enabled) so risk scoring does not
	// depend on a presentation option. Omitted from JSON/YAML.
	BridgeModuleCount int `json:"-" yaml:"-"`

	Warnings []string `json:"warnings,omitempty" yaml:"warnings,omitempty"`
	Errors   []string `json:"errors,omitempty" yaml:"errors,omitempty"`

	GeneratedAt string      `json:"generated_at" yaml:"generated_at"`
	Version     string      `json:"version" yaml:"version"`
	Config      interface{} `json:"config,omitempty" yaml:"config,omitempty"`
}

// ScoreCommunityResult computes the system-level community risk score and the
// per-community risk_level, mutating the result in place. It is the single entry
// point used by both the standalone community command and the analyze health
// score, so the numbers stay consistent. Safe to call with a nil result.
func ScoreCommunityResult(result *CommunityAnalysisResult) {
	if result == nil {
		return
	}

	// Per-community risk levels are always classifiable from local metrics.
	for i := range result.Communities {
		result.Communities[i].RiskLevel = communityRiskLevel(&result.Communities[i])
	}

	// The system risk score needs at least two communities to be meaningful.
	if result.TotalCommunities < 2 {
		result.RiskScore = nil
		return
	}

	internalEdges, crossEdges := 0, 0
	for i := range result.Communities {
		internalEdges += result.Communities[i].InternalEdges
		crossEdges += result.Communities[i].OutgoingCrossCommunityEdges
	}

	ratio := computeCommunityRiskRatio(communityRiskInputs{
		communityCount:   result.TotalCommunities,
		modularity:       result.Modularity,
		bridgeModules:    result.BridgeModuleCount,
		internalEdges:    internalEdges,
		crossEdges:       crossEdges,
		packageAlignment: result.PackageAlignmentScore,
		layerAlignment:   result.LayerAlignmentScore,
	})
	score := int(math.Round(ratio * 100))
	result.RiskScore = &score
}

// communityRiskLevel classifies a single community as low/medium/high. It blends
// the community's external dependency ratio with its package and layer alignment
// (each included only when the underlying metadata is available).
func communityRiskLevel(c *CommunityMetrics) string {
	var weightedSum, totalWeight float64

	weightedSum += 0.5 * clamp01(c.ExternalDependencyRatio)
	totalWeight += 0.5

	if c.PackageCount > 0 {
		weightedSum += 0.25 * clamp01(1-c.PackageAlignment)
		totalWeight += 0.25
	}
	if c.LayerAlignment != nil {
		weightedSum += 0.25 * clamp01(1-*c.LayerAlignment)
		totalWeight += 0.25
	}

	ratio := 0.0
	if totalWeight > 0 {
		ratio = weightedSum / totalWeight
	}

	switch {
	case ratio >= CommunityRiskHighRatio:
		return "high"
	case ratio >= CommunityRiskMediumRatio:
		return "medium"
	default:
		return "low"
	}
}

// CommunityAnalysisService defines the core business logic for community analysis.
type CommunityAnalysisService interface {
	Analyze(ctx context.Context, req CommunityAnalysisRequest) (*CommunityAnalysisResult, error)
}

// CommunityConfigurationLoader defines the interface for loading community configuration.
type CommunityConfigurationLoader interface {
	LoadConfig(path string) (*CommunityAnalysisRequest, error)
	LoadDefaultConfig() *CommunityAnalysisRequest
	MergeConfig(base *CommunityAnalysisRequest, override *CommunityAnalysisRequest) *CommunityAnalysisRequest
}

// CommunityAnalysisOutputFormatter defines the interface for formatting community results.
type CommunityAnalysisOutputFormatter interface {
	Format(response *CommunityAnalysisResult, format OutputFormat) (string, error)
	Write(response *CommunityAnalysisResult, format OutputFormat, writer io.Writer) error
}

// DefaultCommunityAnalysisRequest returns a CommunityAnalysisRequest with default values.
func DefaultCommunityAnalysisRequest() *CommunityAnalysisRequest {
	return &CommunityAnalysisRequest{
		OutputFormat:        OutputFormatText,
		Algorithm:           DefaultCommunityAlgorithm,
		Scope:               DefaultCommunityScope,
		MinCommunitySize:    DefaultCommunityMinSize,
		IncludeLazyEdges:    BoolPtr(true),
		ReportBridgeModules: BoolPtr(true),
		Resolution:          DefaultCommunityResolution,
		Recursive:           BoolPtr(true),
		IncludePatterns:     DefaultPythonModuleIncludePatterns(),
		ExcludePatterns:     DefaultAnalysisExcludePatterns(),
		IncludeStdLib:       BoolPtr(false),
		IncludeThirdParty:   BoolPtr(true),
		FollowRelative:      BoolPtr(true),
	}
}
