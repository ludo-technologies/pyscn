package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommunityAnalysisService_Analyze_FixtureProject(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "hexagonal_ports"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)
	require.NotEmpty(t, files)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:        files,
		OutputFormat: domain.OutputFormatJSON,
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "leiden", result.Algorithm)
	assert.Equal(t, "module", result.Scope)
	assert.Greater(t, result.TotalCommunities, 0)
	assert.NotEmpty(t, result.Communities)
	assert.GreaterOrEqual(t, result.Modularity, 0.0)
}

func TestCommunityAnalysisService_Analyze_MinimalSingleModule(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_minimal"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:       files,
		SourcePaths: []string{fixtureRoot},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 1, result.TotalCommunities)
	assert.Len(t, result.Communities, 1)
	assert.Equal(t, []string{"solo"}, result.Communities[0].Modules)
	assert.Empty(t, result.BridgeModules)
}

func TestCommunityAnalysisService_Analyze_IsolatedModulesNoEdges(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_isolated"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:       files,
		SourcePaths: []string{fixtureRoot},
	})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 3, result.TotalCommunities)
	assert.Equal(t, 0.0, result.Modularity)
	assert.Empty(t, result.BridgeModules)
}

func TestCommunityAnalysisService_Analyze_PackageMismatch_BridgeFixture(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_bridge"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:            files,
		SourcePaths:      []string{fixtureRoot},
		MinCommunitySize: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.PackageAlignmentScore)
	assert.InDelta(t, 0.0, *result.PackageAlignmentScore, 1e-9)
	assert.Equal(t, []string{"mod"}, result.SplitPackages)
	assert.Empty(t, result.MixedCommunities)

	for _, community := range result.Communities {
		assert.Equal(t, 1, community.PackageCount)
		assert.Equal(t, "mod", community.DominantPackage)
		assert.InDelta(t, 1.0, community.PackageAlignment, 1e-9)
	}
}

func TestCommunityAnalysisService_Analyze_PackageMismatch_SeparatedFixture(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_separated"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:            files,
		SourcePaths:      []string{fixtureRoot},
		MinCommunitySize: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.PackageAlignmentScore)
	assert.InDelta(t, 1.0, *result.PackageAlignmentScore, 1e-9)
	assert.Empty(t, result.SplitPackages)
	assert.Empty(t, result.MixedCommunities)
}

func TestCommunityAnalysisService_Analyze_PackageMismatch_MixedFixture(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_package_mismatch"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:            files,
		SourcePaths:      []string{fixtureRoot},
		MinCommunitySize: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.PackageAlignmentScore)
	assert.InDelta(t, 0.0, *result.PackageAlignmentScore, 1e-9)
	assert.Equal(t, []string{"pkg_alpha", "pkg_beta"}, result.SplitPackages)
	assert.Equal(t, []string{"community_1", "community_2"}, result.MixedCommunities)

	for _, community := range result.Communities {
		assert.Equal(t, 2, community.PackageCount)
		assert.Contains(t, []string{"pkg_alpha", "pkg_beta"}, community.DominantPackage)
	}
}

// analyzeCommunityFixtureForRisk runs community analysis against a fixture and
// returns the scored result (RiskScore and per-community risk levels populated).
func analyzeCommunityFixtureForRisk(t *testing.T, fixture string) *domain.CommunityAnalysisResult {
	t.Helper()
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", fixture))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:            files,
		SourcePaths:      []string{fixtureRoot},
		MinCommunitySize: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	return result
}

// A high-bridge / low-Q project must carry more community risk (a lower
// community score) than a cleanly separated one.
func TestCommunityAnalysisService_RiskScore_BridgeWorseThanSeparated(t *testing.T) {
	bridge := analyzeCommunityFixtureForRisk(t, "community_bridge")
	separated := analyzeCommunityFixtureForRisk(t, "community_separated")

	require.NotNil(t, bridge.RiskScore, "bridge fixture should be scored (>= 2 communities)")
	require.NotNil(t, separated.RiskScore, "separated fixture should be scored (>= 2 communities)")

	assert.Greater(t, *bridge.RiskScore, *separated.RiskScore,
		"high-bridge/low-Q fixture (Q=%.3f) should score riskier than clean separation (Q=%.3f)",
		bridge.Modularity, separated.Modularity)

	// The cleanly separated fixture should be essentially risk-free.
	assert.LessOrEqual(t, *separated.RiskScore, 10)
}

// Risk scoring must not depend on the bridge-module reporting option: disabling
// report_bridge_modules omits the emitted list but must leave the score intact.
func TestCommunityAnalysisService_RiskScore_IndependentOfBridgeReporting(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_bridge"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	service := NewCommunityAnalysisService()
	base := domain.CommunityAnalysisRequest{
		Paths:            files,
		SourcePaths:      []string{fixtureRoot},
		MinCommunitySize: 2,
	}

	withReport := base
	withReport.ReportBridgeModules = domain.BoolPtr(true)
	reported, err := service.Analyze(context.Background(), withReport)
	require.NoError(t, err)

	withoutReport := base
	withoutReport.ReportBridgeModules = domain.BoolPtr(false)
	suppressed, err := service.Analyze(context.Background(), withoutReport)
	require.NoError(t, err)

	// The emitted list differs, but the analysis bridge count and risk score must not.
	assert.NotEmpty(t, reported.BridgeModules)
	assert.Empty(t, suppressed.BridgeModules)
	assert.Equal(t, reported.BridgeModuleCount, suppressed.BridgeModuleCount)
	assert.Positive(t, suppressed.BridgeModuleCount)
	require.NotNil(t, reported.RiskScore)
	require.NotNil(t, suppressed.RiskScore)
	assert.Equal(t, *reported.RiskScore, *suppressed.RiskScore)
}

func TestCommunityAnalysisService_Analyze_Deterministic(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_bridge"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	req := domain.CommunityAnalysisRequest{
		Paths:            files,
		SourcePaths:      []string{fixtureRoot},
		MinCommunitySize: 2,
	}
	service := NewCommunityAnalysisService()

	first, err := service.Analyze(context.Background(), req)
	require.NoError(t, err)
	second, err := service.Analyze(context.Background(), req)
	require.NoError(t, err)

	assert.Equal(t, first.TotalCommunities, second.TotalCommunities)
	assert.InDelta(t, first.Modularity, second.Modularity, 1e-12)
	assert.Equal(t, first.Communities, second.Communities)
	assert.Equal(t, first.BridgeModules, second.BridgeModules)
	assert.Equal(t, first.PackageAlignmentScore, second.PackageAlignmentScore)
	assert.Equal(t, first.SplitPackages, second.SplitPackages)
	assert.Equal(t, first.MixedCommunities, second.MixedCommunities)
	assert.Equal(t, first.LayerAlignmentScore, second.LayerAlignmentScore)
	assert.Equal(t, first.CrossLayerCommunities, second.CrossLayerCommunities)
	assert.Equal(t, first.LayerBridgeModules, second.LayerBridgeModules)
}

func TestCommunityAnalysisService_Analyze_LayerMismatch_BridgeFixture(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_layer_bridge"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:            files,
		SourcePaths:      []string{fixtureRoot},
		MinCommunitySize: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.LayerAlignmentScore)
	assert.InDelta(t, 1.0, *result.LayerAlignmentScore, 1e-9)
	assert.Empty(t, result.CrossLayerCommunities)
	assert.Equal(t, []string{"bridge", "infra.c"}, result.LayerBridgeModules)
}

func TestCommunityAnalysisService_Analyze_LayerMismatch_AlignedFixture(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_layer_aligned"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:            files,
		SourcePaths:      []string{fixtureRoot},
		MinCommunitySize: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.LayerAlignmentScore)
	assert.InDelta(t, 1.0, *result.LayerAlignmentScore, 1e-9)
	assert.Empty(t, result.CrossLayerCommunities)
	assert.Empty(t, result.LayerBridgeModules)
}

func TestCommunityAnalysisService_Analyze_LayerMismatch_MixedFixture(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_layer_mixed"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:            files,
		SourcePaths:      []string{fixtureRoot},
		MinCommunitySize: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 2, result.TotalCommunities)
	require.NotNil(t, result.LayerAlignmentScore)
	assert.InDelta(t, 0.0, *result.LayerAlignmentScore, 1e-9)
	assert.Equal(t, []string{"community_1", "community_2"}, result.CrossLayerCommunities)

	for _, community := range result.Communities {
		assert.Equal(t, 2, community.LayerCount)
		require.NotNil(t, community.LayerAlignment, "layer_alignment must be present when layer_count > 0")
	}

	formatter := NewCommunityFormatter()
	output, err := formatter.Format(result, domain.OutputFormatJSON)
	require.NoError(t, err)
	assert.Contains(t, output, `"layer_alignment": 0`)
}

func TestCommunityAnalysisService_Analyze_LayerMismatch_OmittedWithoutArchitecture(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "community_bridge"))
	require.NoError(t, err)

	fileReader := NewFileReader()
	files, err := fileReader.CollectPythonFiles([]string{fixtureRoot}, true, nil, nil)
	require.NoError(t, err)

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:            files,
		SourcePaths:      []string{fixtureRoot},
		MinCommunitySize: 2,
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Nil(t, result.LayerAlignmentScore)
	assert.Empty(t, result.CrossLayerCommunities)
	assert.Empty(t, result.LayerBridgeModules)
}

func TestCommunityAnalysisService_Analyze_UsesSourcePathsForProjectRoot(t *testing.T) {
	root := t.TempDir()
	srcDir := filepath.Join(root, "src", "myapp")
	require.NoError(t, os.MkdirAll(srcDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(root, "pyproject.toml"), []byte("[project]\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "__init__.py"), []byte(""), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "service.py"), []byte("from myapp import repo\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(srcDir, "repo.py"), []byte("pass\n"), 0o644))

	fileA := filepath.Join(srcDir, "service.py")
	fileB := filepath.Join(srcDir, "repo.py")

	service := NewCommunityAnalysisService()
	result, err := service.Analyze(context.Background(), domain.CommunityAnalysisRequest{
		Paths:       []string{fileA, fileB},
		SourcePaths: []string{root},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, result.TotalCommunities, 0)
	assert.Contains(t, communityModuleNames(result), "myapp.service")
	assert.Contains(t, communityModuleNames(result), "myapp.repo")
	assert.Contains(t, result.ModuleDependencies, domain.CommunityModuleDependency{From: "myapp.service", To: "myapp.repo"})
}

func communityModuleNames(result *domain.CommunityAnalysisResult) []string {
	if result == nil {
		return nil
	}
	names := make([]string, 0)
	for _, community := range result.Communities {
		names = append(names, community.Modules...)
	}
	return names
}
