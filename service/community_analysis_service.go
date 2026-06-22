package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/version"
)

// CommunityAnalysisServiceImpl implements domain.CommunityAnalysisService.
type CommunityAnalysisServiceImpl struct{}

// NewCommunityAnalysisService creates a new community analysis service.
func NewCommunityAnalysisService() *CommunityAnalysisServiceImpl {
	return &CommunityAnalysisServiceImpl{}
}

// Analyze performs community detection over the module dependency graph.
func (s *CommunityAnalysisServiceImpl) Analyze(ctx context.Context, req domain.CommunityAnalysisRequest) (*domain.CommunityAnalysisResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("community analysis cancelled: %w", err)
	}

	graph, err := s.buildDependencyGraph(ctx, req)
	if err != nil {
		return nil, err
	}

	graphOpts := &analyzer.CommunityGraphBuildOptions{
		ExcludeLazyEdges: !domain.BoolValue(req.IncludeLazyEdges, true),
	}
	cg := analyzer.BuildCommunityGraph(graph, graphOpts)

	leidenOpts := &analyzer.LeidenOptions{
		Resolution:       req.Resolution,
		MinCommunitySize: req.MinCommunitySize,
	}
	if leidenOpts.Resolution <= 0 {
		leidenOpts.Resolution = domain.DefaultCommunityResolution
	}
	if leidenOpts.MinCommunitySize <= 0 {
		leidenOpts.MinCommunitySize = domain.DefaultCommunityMinSize
	}

	leiden := analyzer.DetectCommunitiesLeiden(cg, leidenOpts)
	metrics := analyzer.ComputeCommunityMetrics(graph, cg, leiden)

	result := &domain.CommunityAnalysisResult{
		Algorithm:        s.resolveAlgorithm(req.Algorithm),
		Scope:            s.resolveScope(req.Scope),
		TotalCommunities: metrics.TotalCommunities,
		Modularity:       metrics.Modularity,
		Communities:      s.convertCommunities(metrics.Communities),
		GeneratedAt:      time.Now().Format(time.RFC3339),
		Version:          version.Version,
		Config:           s.buildConfigForResponse(req),
	}

	if domain.BoolValue(req.ReportBridgeModules, true) {
		result.BridgeModules = s.convertBridgeModules(metrics.BridgeModules)
	}

	if graph.TotalModules == 0 {
		result.Warnings = append(result.Warnings, "No modules found to analyze")
	}

	return result, nil
}

func (s *CommunityAnalysisServiceImpl) buildDependencyGraph(ctx context.Context, req domain.CommunityAnalysisRequest) (*analyzer.DependencyGraph, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("module graph cancelled: %w", err)
	}

	rootPaths := req.SourcePaths
	if len(rootPaths) == 0 {
		rootPaths = req.Paths
	}
	projectRoot := FindProjectRoot(rootPaths)
	options := &analyzer.ModuleAnalysisOptions{
		ProjectRoot:       projectRoot,
		IncludeStdLib:     req.IncludeStdLib,
		IncludeThirdParty: req.IncludeThirdParty,
		FollowRelative:    req.FollowRelative,
		IncludePatterns:   req.IncludePatterns,
		ExcludePatterns:   req.ExcludePatterns,
	}

	ma, err := analyzer.NewModuleAnalyzer(options)
	if err != nil {
		return nil, fmt.Errorf("failed to create module analyzer: %w", err)
	}

	graph, err := ma.AnalyzeFiles(req.Paths)
	if err != nil {
		return nil, fmt.Errorf("failed to build module graph: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("module graph cancelled: %w", err)
	}
	return graph, nil
}

func (s *CommunityAnalysisServiceImpl) resolveAlgorithm(algorithm string) string {
	if algorithm == "" {
		return domain.DefaultCommunityAlgorithm
	}
	return algorithm
}

func (s *CommunityAnalysisServiceImpl) resolveScope(scope string) string {
	if scope == "" {
		return domain.DefaultCommunityScope
	}
	return scope
}

func (s *CommunityAnalysisServiceImpl) convertCommunities(partitions []analyzer.CommunityPartition) []domain.CommunityMetrics {
	if len(partitions) == 0 {
		return []domain.CommunityMetrics{}
	}

	out := make([]domain.CommunityMetrics, 0, len(partitions))
	for _, partition := range partitions {
		out = append(out, domain.CommunityMetrics{
			ID:                          partition.ID,
			Modules:                     append([]string(nil), partition.Modules...),
			Packages:                    append([]string(nil), partition.Packages...),
			InternalEdges:               partition.InternalEdges,
			ExternalEdges:               partition.ExternalEdges,
			ExternalDependencyRatio:     partition.ExternalDependencyRatio,
			IncomingCrossCommunityEdges: partition.IncomingCrossCommunityEdges,
			OutgoingCrossCommunityEdges: partition.OutgoingCrossCommunityEdges,
			Size:                        partition.Size,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func (s *CommunityAnalysisServiceImpl) convertBridgeModules(bridges []analyzer.BridgeModuleMetrics) []domain.BridgeModule {
	if len(bridges) == 0 {
		return []domain.BridgeModule{}
	}

	out := make([]domain.BridgeModule, 0, len(bridges))
	for _, bridge := range bridges {
		out = append(out, domain.BridgeModule{
			Module:              bridge.Module,
			Community:           bridge.CommunityID,
			CrossCommunityEdges: bridge.CrossCommunityEdges,
			TargetCommunities:   append([]string(nil), bridge.TargetCommunities...),
		})
	}
	return out
}

func (s *CommunityAnalysisServiceImpl) buildConfigForResponse(req domain.CommunityAnalysisRequest) any {
	return map[string]any{
		"algorithm":           s.resolveAlgorithm(req.Algorithm),
		"scope":               s.resolveScope(req.Scope),
		"minCommunitySize":    req.MinCommunitySize,
		"includeLazyEdges":    domain.BoolValue(req.IncludeLazyEdges, true),
		"reportBridgeModules": domain.BoolValue(req.ReportBridgeModules, true),
		"resolution":          req.Resolution,
		"includeStdLib":       domain.BoolValue(req.IncludeStdLib, false),
		"includeThirdParty":   domain.BoolValue(req.IncludeThirdParty, true),
		"followRelative":      domain.BoolValue(req.FollowRelative, true),
	}
}
