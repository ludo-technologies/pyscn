package main

import (
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/app"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeCommand_BuildAnalyzeUseCase_WiresCommunityAnalysis(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "..", "testdata", "python", "mvc_app"))
	require.NoError(t, err)

	cmd := &AnalyzeCommand{}
	cobraCmd := &cobra.Command{}
	builder := app.NewAnalyzeUseCaseBuilder()
	require.NoError(t, cmd.buildIndividualUseCases(builder, cobraCmd))

	progressManager := service.NewProgressManager()
	progressManager.SetWriter(io.Discard)

	useCase, err := builder.
		WithFileReader(service.NewFileReader()).
		WithConfigLoader(service.NewAnalyzeConfigurationLoader()).
		WithFormatter(service.NewAnalyzeFormatter()).
		WithProgressManager(progressManager).
		WithParallelExecutor(service.NewParallelExecutor()).
		WithErrorCategorizer(service.NewErrorCategorizer()).
		Build()
	require.NoError(t, err)

	response, err := useCase.Execute(context.Background(), app.AnalyzeUseCaseConfig{
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
