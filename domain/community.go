package domain

import (
	"context"
	"fmt"
	"io"
	"math"
	"sort"
	"strings"
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

// CommunityContextMapVersion is the schema version of CommunityContextMap.
// Bump it when the shape of the context map changes in a breaking way.
const CommunityContextMapVersion = 1

// CommunityContextMapModuleLimit caps how many modules are listed per bundle
// before the remainder is collapsed into a "... +N more" marker, keeping the
// map token-efficient for AI agents on large repositories.
const CommunityContextMapModuleLimit = 10

// CommunityContextMap is a compact, agent-optimized view of the community
// analysis. It tells AI coding/review agents which modules to inspect together
// and which modules bridge otherwise-separate clusters. It is derived entirely
// from CommunityAnalysisResult and carries no LLM-generated content.
type CommunityContextMap struct {
	Version       int                   `json:"version" yaml:"version"`
	Bundles       []CommunityBundle     `json:"bundles" yaml:"bundles"`
	BridgeModules []ContextBridgeModule `json:"bridge_modules" yaml:"bridge_modules"`
}

// CommunityBundle is a single cluster of modules an agent should review together.
type CommunityBundle struct {
	CommunityID          string   `json:"community_id" yaml:"community_id"`
	Modules              []string `json:"modules" yaml:"modules"`
	ModuleCount          int      `json:"module_count" yaml:"module_count"`
	Packages             []string `json:"packages" yaml:"packages"`
	RiskLevel            string   `json:"risk_level" yaml:"risk_level"`
	BridgeModules        []string `json:"bridge_modules" yaml:"bridge_modules"`
	SuggestedReviewScope string   `json:"suggested_review_scope,omitempty" yaml:"suggested_review_scope,omitempty"`
	Summary              string   `json:"summary" yaml:"summary"`
}

// ContextBridgeModule is a module that couples two or more communities, surfaced
// at the top level so agents widen review scope across cluster boundaries.
type ContextBridgeModule struct {
	Module   string   `json:"module" yaml:"module"`
	Connects []string `json:"connects" yaml:"connects"`
	Reason   string   `json:"reason" yaml:"reason"`
}

// BuildCommunityContextMap derives a compact, deterministic context map from a
// scored community analysis result. It returns nil when there is nothing to map
// (no result or no communities). Call ScoreCommunityResult first so per-community
// risk levels are populated.
func BuildCommunityContextMap(result *CommunityAnalysisResult) *CommunityContextMap {
	if result == nil || len(result.Communities) == 0 {
		return nil
	}

	// Group bridge modules by their owning community for per-bundle listing.
	bridgesByCommunity := make(map[string][]string, len(result.BridgeModules))
	for _, b := range result.BridgeModules {
		bridgesByCommunity[b.Community] = append(bridgesByCommunity[b.Community], b.Module)
	}

	bundles := make([]CommunityBundle, 0, len(result.Communities))
	for i := range result.Communities {
		c := &result.Communities[i]

		modules := append([]string(nil), c.Modules...)
		sort.Strings(modules)
		packages := append([]string(nil), c.Packages...)
		sort.Strings(packages)

		bridges := append([]string(nil), bridgesByCommunity[c.ID]...)
		sort.Strings(bridges)

		bundles = append(bundles, CommunityBundle{
			CommunityID:          c.ID,
			Modules:              capModuleList(modules, CommunityContextMapModuleLimit),
			ModuleCount:          len(modules),
			Packages:             packages,
			RiskLevel:            c.RiskLevel,
			BridgeModules:        bridges,
			SuggestedReviewScope: suggestedReviewScope(modules),
			Summary:              bundleSummary(c, len(bridges)),
		})
	}
	sort.Slice(bundles, func(i, j int) bool {
		return bundles[i].CommunityID < bundles[j].CommunityID
	})

	bridgeModules := make([]ContextBridgeModule, 0, len(result.BridgeModules))
	for _, b := range result.BridgeModules {
		connects := append([]string{b.Community}, b.TargetCommunities...)
		connects = sortedUnique(connects)
		bridgeModules = append(bridgeModules, ContextBridgeModule{
			Module:   b.Module,
			Connects: connects,
			Reason:   pluralizeEdges(b.CrossCommunityEdges, "cross-community import edge"),
		})
	}
	sort.Slice(bridgeModules, func(i, j int) bool {
		return bridgeModules[i].Module < bridgeModules[j].Module
	})

	return &CommunityContextMap{
		Version:       CommunityContextMapVersion,
		Bundles:       bundles,
		BridgeModules: bridgeModules,
	}
}

// capModuleList truncates a sorted module list to limit entries, appending a
// "... +N more" marker when modules are omitted. A non-positive limit disables
// truncation.
func capModuleList(modules []string, limit int) []string {
	if limit <= 0 || len(modules) <= limit {
		return modules
	}
	out := append([]string(nil), modules[:limit]...)
	return append(out, fmt.Sprintf("... +%d more", len(modules)-limit))
}

// suggestedReviewScope derives a filesystem-style review scope from the longest
// common dotted prefix of the member modules (e.g. ["app.orders.service",
// "app.orders.repository"] -> "app/orders/"). For a single module it drops the
// leaf so the scope points at the containing package. Returns "" when modules
// share no common package prefix.
func suggestedReviewScope(modules []string) string {
	if len(modules) == 0 {
		return ""
	}

	split := func(m string) []string { return strings.Split(m, ".") }
	common := split(modules[0])
	for _, m := range modules[1:] {
		common = commonPrefix(common, split(m))
		if len(common) == 0 {
			return ""
		}
	}

	// A single module's "common prefix" is the whole module; drop the leaf to
	// land on its package directory.
	if len(modules) == 1 && len(common) > 0 {
		common = common[:len(common)-1]
	}
	if len(common) == 0 {
		return ""
	}
	return strings.Join(common, "/") + "/"
}

func commonPrefix(a, b []string) []string {
	n := min(len(a), len(b))
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return a[:i]
}

// bundleSummary produces a short, deterministic, fact-based description of a
// community for agent consumption (no natural-language inference).
func bundleSummary(c *CommunityMetrics, bridgeCount int) string {
	cross := c.IncomingCrossCommunityEdges + c.OutgoingCrossCommunityEdges
	parts := []string{
		pluralize(c.Size, "module"),
		pluralize(len(c.Packages), "package"),
		fmt.Sprintf("risk %s", riskLevelOrUnknown(c.RiskLevel)),
		pluralizeEdges(cross, "cross-community edge"),
	}
	if bridgeCount > 0 {
		parts = append(parts, pluralize(bridgeCount, "bridge module"))
	}
	return strings.Join(parts, "; ") + "."
}

func riskLevelOrUnknown(level string) string {
	if level == "" {
		return "unknown"
	}
	return level
}

func pluralize(n int, noun string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, noun)
	}
	return fmt.Sprintf("%d %ss", n, noun)
}

func pluralizeEdges(n int, noun string) string {
	return pluralize(n, noun)
}

func sortedUnique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
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

	// ContextMap is a compact, agent-optimized view of the communities (which
	// modules to inspect together, which modules bridge clusters). Populated by
	// ScoreCommunityResult whenever at least one community was detected.
	ContextMap *CommunityContextMap `json:"community_context_map,omitempty" yaml:"community_context_map,omitempty"`

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

	// Build the agent-facing context map once risk levels are populated.
	result.ContextMap = BuildCommunityContextMap(result)

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
