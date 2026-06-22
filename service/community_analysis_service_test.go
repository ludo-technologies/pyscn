package service

import (
	"context"
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
