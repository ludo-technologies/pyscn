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
