package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	communityBridgeDir          = "../testdata/python/community_bridge"
	communitySeparatedDir       = "../testdata/python/community_separated"
	communityPackageMismatchDir = "../testdata/python/community_package_mismatch"
	communityLayerBridgeDir     = "../testdata/python/community_layer_bridge"
	communityLayerAlignedDir    = "../testdata/python/community_layer_aligned"
	communityLayerMixedDir      = "../testdata/python/community_layer_mixed"
	communityCycleDir           = "../testdata/python/community_cycle"
	communityMinimalDir         = "../testdata/python/community_minimal"
	communityIsolatedDir        = "../testdata/python/community_isolated"
)

func newCommunityUseCase() *app.CommunityUseCase {
	uc, err := app.NewCommunityUseCaseBuilder().
		WithService(service.NewCommunityAnalysisService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewCommunityFormatter()).
		Build()
	if err != nil {
		panic(err)
	}
	return uc
}

func analyzeCommunityFixture(t *testing.T, fixtureDir string) *domain.CommunityAnalysisResult {
	t.Helper()

	absDir, err := filepath.Abs(fixtureDir)
	require.NoError(t, err)

	uc := newCommunityUseCase()
	result, err := uc.AnalyzeAndReturn(context.Background(), domain.CommunityAnalysisRequest{
		Paths:            []string{absDir},
		SourcePaths:      []string{absDir},
		OutputWriter:     ioDiscard{},
		OutputFormat:     domain.OutputFormatJSON,
		Recursive:        domain.BoolPtr(true),
		MinCommunitySize: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	return result
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }

func TestCommunity_SeparatedCommunities(t *testing.T) {
	result := analyzeCommunityFixture(t, communitySeparatedDir)

	assert.Equal(t, 2, result.TotalCommunities)
	assert.Empty(t, result.BridgeModules)
	assert.Greater(t, result.Modularity, 0.0)

	for _, community := range result.Communities {
		assert.Equal(t, 0, community.ExternalEdges)
		assert.Equal(t, 0.0, community.ExternalDependencyRatio)
		assert.GreaterOrEqual(t, community.Size, 2)
		assert.Greater(t, community.InternalEdges, 0)
	}
}

func TestCommunity_BridgeModules(t *testing.T) {
	result := analyzeCommunityFixture(t, communityBridgeDir)

	require.Equal(t, 2, result.TotalCommunities)
	require.Len(t, result.BridgeModules, 2)

	bridgeByModule := make(map[string]domain.BridgeModule, len(result.BridgeModules))
	for _, bridge := range result.BridgeModules {
		bridgeByModule[bridge.Module] = bridge
	}

	bridge, ok := bridgeByModule["bridge"]
	require.True(t, ok)
	assert.Equal(t, 1, bridge.CrossCommunityEdges)
	assert.Contains(t, bridge.TargetCommunities, "community_2")

	modC, ok := bridgeByModule["mod.c"]
	require.True(t, ok)
	assert.Equal(t, 1, modC.CrossCommunityEdges)
	assert.Contains(t, modC.TargetCommunities, "community_1")
}

func TestCommunity_PackageMismatchSplitPackage(t *testing.T) {
	result := analyzeCommunityFixture(t, communityBridgeDir)

	require.NotNil(t, result.PackageAlignmentScore)
	assert.InDelta(t, 0.0, *result.PackageAlignmentScore, 1e-9)
	assert.Equal(t, []string{"mod"}, result.SplitPackages)
	assert.Empty(t, result.MixedCommunities)
}

func TestCommunity_PackageMismatchAlignedPackages(t *testing.T) {
	result := analyzeCommunityFixture(t, communitySeparatedDir)

	require.NotNil(t, result.PackageAlignmentScore)
	assert.InDelta(t, 1.0, *result.PackageAlignmentScore, 1e-9)
	assert.Empty(t, result.SplitPackages)
	assert.Empty(t, result.MixedCommunities)
}

func TestCommunity_LayerMismatchBridgeFixture(t *testing.T) {
	result := analyzeCommunityFixture(t, communityLayerBridgeDir)

	require.NotNil(t, result.LayerAlignmentScore)
	assert.InDelta(t, 1.0, *result.LayerAlignmentScore, 1e-9)
	assert.Empty(t, result.CrossLayerCommunities)
	assert.Equal(t, []string{"bridge", "infra.c"}, result.LayerBridgeModules)
}

func TestCommunity_LayerMismatchAlignedLayers(t *testing.T) {
	result := analyzeCommunityFixture(t, communityLayerAlignedDir)

	require.NotNil(t, result.LayerAlignmentScore)
	assert.InDelta(t, 1.0, *result.LayerAlignmentScore, 1e-9)
	assert.Empty(t, result.CrossLayerCommunities)
	assert.Empty(t, result.LayerBridgeModules)
}

func TestCommunity_LayerMismatchCrossLayerCommunity(t *testing.T) {
	result := analyzeCommunityFixture(t, communityLayerMixedDir)

	require.Equal(t, 2, result.TotalCommunities)
	require.NotNil(t, result.LayerAlignmentScore)
	assert.InDelta(t, 0.0, *result.LayerAlignmentScore, 1e-9)
	assert.Equal(t, []string{"community_1", "community_2"}, result.CrossLayerCommunities)
	for _, community := range result.Communities {
		assert.Equal(t, 2, community.LayerCount)
	}
}

func TestCommunity_LayerMismatchOmittedWithoutArchitecture(t *testing.T) {
	result := analyzeCommunityFixture(t, communityBridgeDir)

	assert.Nil(t, result.LayerAlignmentScore)
	assert.Empty(t, result.CrossLayerCommunities)
	assert.Empty(t, result.LayerBridgeModules)
}

func TestCommunity_PackageMismatchMixedCommunities(t *testing.T) {
	result := analyzeCommunityFixture(t, communityPackageMismatchDir)

	require.Equal(t, 2, result.TotalCommunities)
	require.NotNil(t, result.PackageAlignmentScore)
	assert.InDelta(t, 0.0, *result.PackageAlignmentScore, 1e-9)
	assert.Equal(t, []string{"pkg_alpha", "pkg_beta"}, result.SplitPackages)
	assert.Equal(t, []string{"community_1", "community_2"}, result.MixedCommunities)

	for _, community := range result.Communities {
		assert.Equal(t, 2, community.PackageCount)
	}
}

func TestCommunity_CycleGraph(t *testing.T) {
	result := analyzeCommunityFixture(t, communityCycleDir)

	require.Equal(t, 1, result.TotalCommunities)
	require.Len(t, result.Communities, 1)
	assert.Equal(t, 3, result.Communities[0].Size)
	assert.Equal(t, 3, result.Communities[0].InternalEdges)
	assert.Equal(t, 0, result.Communities[0].ExternalEdges)
	assert.Empty(t, result.BridgeModules)
}

func TestCommunity_MinimalSingleModule(t *testing.T) {
	result := analyzeCommunityFixture(t, communityMinimalDir)

	require.Equal(t, 1, result.TotalCommunities)
	require.Len(t, result.Communities, 1)
	assert.Equal(t, []string{"solo"}, result.Communities[0].Modules)
	assert.Equal(t, 0, result.Communities[0].InternalEdges)
	assert.Equal(t, 0.0, result.Modularity)
	assert.Empty(t, result.BridgeModules)
}

func TestCommunity_IsolatedModulesNoEdges(t *testing.T) {
	result := analyzeCommunityFixture(t, communityIsolatedDir)

	require.Equal(t, 3, result.TotalCommunities)
	assert.Equal(t, 0.0, result.Modularity)
	assert.Empty(t, result.BridgeModules)

	for _, community := range result.Communities {
		assert.Equal(t, 1, community.Size)
		assert.Equal(t, 0, community.InternalEdges)
		assert.Equal(t, 0, community.ExternalEdges)
	}
}

func TestCommunity_DeterministicRepeatedRuns(t *testing.T) {
	absDir, err := filepath.Abs(communityBridgeDir)
	require.NoError(t, err)

	uc := newCommunityUseCase()
	req := domain.CommunityAnalysisRequest{
		Paths:            []string{absDir},
		SourcePaths:      []string{absDir},
		OutputWriter:     ioDiscard{},
		OutputFormat:     domain.OutputFormatJSON,
		Recursive:        domain.BoolPtr(true),
		MinCommunitySize: 2,
	}

	first, err := uc.AnalyzeAndReturn(context.Background(), req)
	require.NoError(t, err)
	second, err := uc.AnalyzeAndReturn(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, first.TotalCommunities, second.TotalCommunities)
	assert.InDelta(t, first.Modularity, second.Modularity, 1e-12)
	assert.Equal(t, communityIDs(first), communityIDs(second))
	assert.Equal(t, communityModuleSets(first), communityModuleSets(second))
	assert.Equal(t, bridgeModuleNames(first), bridgeModuleNames(second))
}

func TestCommunity_AnalyzeUseCase_SelectCommunities(t *testing.T) {
	absDir, err := filepath.Abs(communityBridgeDir)
	require.NoError(t, err)

	communityUC, err := app.NewCommunityUseCaseBuilder().
		WithService(service.NewCommunityAnalysisService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewCommunityFormatter()).
		Build()
	require.NoError(t, err)

	analyzeUC, err := app.NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewAnalyzeFormatter()).
		WithCommunityUseCase(communityUC).
		Build()
	require.NoError(t, err)

	response, err := analyzeUC.Execute(context.Background(), app.AnalyzeUseCaseConfig{
		SkipComplexity:     true,
		SkipDeadCode:       true,
		SkipClones:         true,
		SkipCBO:            true,
		SkipLCOM:           true,
		SkipSystem:         true,
		SkipCommunities:    false,
		SelectAnalysesUsed: true,
	}, []string{absDir})
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.True(t, response.Summary.CommunitiesEnabled)
	require.NotNil(t, response.Communities)
	assert.Equal(t, 2, response.Communities.TotalCommunities)
	assert.NotEmpty(t, response.Communities.BridgeModules)

	formatter := service.NewCommunityFormatter()
	response.Communities.GeneratedAt = "2026-01-15T12:00:00Z"
	response.Communities.Version = "0.0.0-test"
	response.Communities.Config = map[string]any{
		"algorithm":           "leiden",
		"scope":               "module",
		"minCommunitySize":    2,
		"includeLazyEdges":    true,
		"reportBridgeModules": true,
		"resolution":          1.0,
		"includeStdLib":       false,
		"includeThirdParty":   true,
		"followRelative":      true,
	}

	var buf bytes.Buffer
	require.NoError(t, formatter.Write(response.Communities, domain.OutputFormatJSON, &buf))

	goldenPath := filepath.Join("..", "service", "testdata", "golden", "community_analysis.json")
	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err)

	assertCommunityJSONEqual(t, expected, buf.Bytes())
}

func communityIDs(result *domain.CommunityAnalysisResult) []string {
	ids := make([]string, len(result.Communities))
	for i, community := range result.Communities {
		ids[i] = community.ID
	}
	return ids
}

func communityModuleSets(result *domain.CommunityAnalysisResult) [][]string {
	sets := make([][]string, len(result.Communities))
	for i, community := range result.Communities {
		sets[i] = append([]string(nil), community.Modules...)
	}
	return sets
}

func bridgeModuleNames(result *domain.CommunityAnalysisResult) []string {
	names := make([]string, len(result.BridgeModules))
	for i, bridge := range result.BridgeModules {
		names[i] = bridge.Module
	}
	return names
}

func assertCommunityJSONEqual(t *testing.T, expected, actual []byte) {
	t.Helper()

	var exp map[string]any
	var act map[string]any
	require.NoError(t, json.Unmarshal(expected, &exp))
	require.NoError(t, json.Unmarshal(actual, &act))
	assert.Equal(t, exp, act)
}
