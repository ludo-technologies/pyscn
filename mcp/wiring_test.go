package mcp

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAnalyzeUseCase_WiresCommunityAnalysis(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "mvc_app"))
	require.NoError(t, err)

	uc, err := buildAnalyzeUseCase(service.NewFileReader())
	require.NoError(t, err)

	response, err := uc.Execute(context.Background(), app.AnalyzeUseCaseConfig{
		SkipComplexity:  true,
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      true,
		SkipCommunities: false,
	}, []string{fixtureRoot})
	require.NoError(t, err)
	require.NotNil(t, response)
	require.NotNil(t, response.Communities)
	assert.Greater(t, response.Communities.TotalCommunities, 0)
}
