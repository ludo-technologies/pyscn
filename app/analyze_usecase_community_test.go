package app

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeUseCase_CommunityTask(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "mvc_app"))
	require.NoError(t, err)

	communityUC, err := NewCommunityUseCaseBuilder().
		WithService(service.NewCommunityAnalysisService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(noopCommunityFormatter{}).
		Build()
	require.NoError(t, err)

	useCase, err := NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewAnalyzeFormatter()).
		WithCommunityUseCase(communityUC).
		Build()
	require.NoError(t, err)

	config := AnalyzeUseCaseConfig{
		SkipComplexity:  true,
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      true,
		SkipCommunities: false,
	}

	response, err := useCase.Execute(context.Background(), config, []string{fixtureRoot})
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.True(t, response.Summary.CommunitiesEnabled)
	require.NotNil(t, response.Communities)
	assert.Greater(t, response.Communities.TotalCommunities, 0)
	assert.NotEmpty(t, response.Communities.Communities)
}

func TestAnalyzeUseCase_CommunityEnabledFromConfig(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "mvc_app"))
	require.NoError(t, err)

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, ".pyscn.toml")
	configContent := `[communities]
enabled = true
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	communityUC, err := NewCommunityUseCaseBuilder().
		WithService(service.NewCommunityAnalysisService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(noopCommunityFormatter{}).
		WithConfigLoader(service.NewCommunityConfigurationLoader()).
		Build()
	require.NoError(t, err)

	useCase, err := NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewAnalyzeFormatter()).
		WithConfigLoader(service.NewAnalyzeConfigurationLoader()).
		WithCommunityUseCase(communityUC).
		Build()
	require.NoError(t, err)

	config := AnalyzeUseCaseConfig{
		SkipComplexity:  true,
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      true,
		SkipCommunities: true,
		ConfigFile:      configPath,
	}

	response, err := useCase.Execute(context.Background(), config, []string{fixtureRoot})
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.True(t, response.Summary.CommunitiesEnabled)
	require.NotNil(t, response.Communities)
}

func TestAnalyzeUseCase_CommunityDisabledFromConfig(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "mvc_app"))
	require.NoError(t, err)

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, ".pyscn.toml")
	configContent := `[communities]
enabled = false
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	communityUC, err := NewCommunityUseCaseBuilder().
		WithService(service.NewCommunityAnalysisService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(noopCommunityFormatter{}).
		WithConfigLoader(service.NewCommunityConfigurationLoader()).
		Build()
	require.NoError(t, err)

	useCase, err := NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewAnalyzeFormatter()).
		WithConfigLoader(service.NewAnalyzeConfigurationLoader()).
		WithCommunityUseCase(communityUC).
		Build()
	require.NoError(t, err)

	config := AnalyzeUseCaseConfig{
		SkipComplexity:  true,
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      true,
		SkipCommunities: false,
		ConfigFile:      configPath,
	}

	response, err := useCase.Execute(context.Background(), config, []string{fixtureRoot})
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.False(t, response.Summary.CommunitiesEnabled)
	assert.Nil(t, response.Communities)
}

func TestAnalyzeUseCase_CommunityConfigSkippedWhenSelectOmitsCommunities(t *testing.T) {
	fixtureRoot, err := filepath.Abs(filepath.Join("..", "testdata", "python", "mvc_app"))
	require.NoError(t, err)

	configDir := t.TempDir()
	configPath := filepath.Join(configDir, ".pyscn.toml")
	configContent := `[communities]
enabled = true
`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

	communityUC, err := NewCommunityUseCaseBuilder().
		WithService(service.NewCommunityAnalysisService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(noopCommunityFormatter{}).
		Build()
	require.NoError(t, err)

	useCase, err := NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewAnalyzeFormatter()).
		WithConfigLoader(service.NewAnalyzeConfigurationLoader()).
		WithCommunityUseCase(communityUC).
		Build()
	require.NoError(t, err)

	config := AnalyzeUseCaseConfig{
		SkipComplexity:     true,
		SkipDeadCode:       true,
		SkipClones:         true,
		SkipCBO:            true,
		SkipLCOM:           true,
		SkipSystem:         true,
		SkipCommunities:    true,
		SelectAnalysesUsed: true,
		ConfigFile:         configPath,
	}

	response, err := useCase.Execute(context.Background(), config, []string{fixtureRoot})
	require.NoError(t, err)
	require.NotNil(t, response)

	assert.False(t, response.Summary.CommunitiesEnabled)
	assert.Nil(t, response.Communities)
}

func TestAnalyzeUseCase_CommunityTaskSkippedByDefault(t *testing.T) {
	communityUC, err := NewCommunityUseCaseBuilder().
		WithService(service.NewCommunityAnalysisService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(noopCommunityFormatter{}).
		Build()
	require.NoError(t, err)

	useCase, err := NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewAnalyzeFormatter()).
		WithCommunityUseCase(communityUC).
		Build()
	require.NoError(t, err)

	tasks := useCase.createAnalysisTasks(AnalyzeUseCaseConfig{
		SkipComplexity:  true,
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      true,
		SkipCommunities: true,
	}, []string{"."}, []string{"."}, nil, domain.AnalyzeExecutionConfig{})

	var communityTask *AnalysisTask
	for _, task := range tasks {
		if task.Name == taskNameCommunities {
			communityTask = task
			break
		}
	}
	require.NotNil(t, communityTask)
	assert.False(t, communityTask.Enabled)
}

func TestAnalyzeUseCase_CommunityTaskRequestUsesDiscardWriter(t *testing.T) {
	communityUC, err := NewCommunityUseCaseBuilder().
		WithService(service.NewCommunityAnalysisService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(noopCommunityFormatter{}).
		Build()
	require.NoError(t, err)

	useCase, err := NewAnalyzeUseCaseBuilder().
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewAnalyzeFormatter()).
		WithCommunityUseCase(communityUC).
		Build()
	require.NoError(t, err)

	tasks := useCase.createAnalysisTasks(AnalyzeUseCaseConfig{
		SkipComplexity:  true,
		SkipDeadCode:    true,
		SkipClones:      true,
		SkipCBO:         true,
		SkipLCOM:        true,
		SkipSystem:      true,
		SkipCommunities: false,
	}, []string{filepath.Join("..", "testdata", "python", "mvc_app")}, []string{filepath.Join("..", "testdata", "python", "mvc_app")}, nil, domain.AnalyzeExecutionConfig{Recursive: true})

	var communityTask *AnalysisTask
	for _, task := range tasks {
		if task.Name == taskNameCommunities {
			communityTask = task
			break
		}
	}
	require.NotNil(t, communityTask)
	require.True(t, communityTask.Enabled)

	_, err = communityTask.Execute(context.Background())
	require.NoError(t, err)
	_ = io.Discard
}
