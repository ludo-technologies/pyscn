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
	}
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
		SourcePaths: []string{srcDir},
	})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Greater(t, result.TotalCommunities, 0)
}
