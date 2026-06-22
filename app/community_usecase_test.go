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

type noopCommunityFormatter struct{}

func (noopCommunityFormatter) Format(*domain.CommunityAnalysisResult, domain.OutputFormat) (string, error) {
	return "", nil
}

func (noopCommunityFormatter) Write(*domain.CommunityAnalysisResult, domain.OutputFormat, io.Writer) error {
	return nil
}

func TestCommunityUseCase_AnalyzeAndReturn_FixtureProject(t *testing.T) {
	fixtureRoot := filepath.Join("..", "testdata", "python", "mvc_app")
	absRoot, err := filepath.Abs(fixtureRoot)
	require.NoError(t, err)

	uc, err := NewCommunityUseCaseBuilder().
		WithService(service.NewCommunityAnalysisService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(noopCommunityFormatter{}).
		Build()
	require.NoError(t, err)

	req := domain.CommunityAnalysisRequest{
		Paths:        []string{absRoot},
		OutputFormat: domain.OutputFormatJSON,
		OutputWriter: io.Discard,
		Recursive:    domain.BoolPtr(true),
	}

	result, err := uc.AnalyzeAndReturn(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "leiden", result.Algorithm)
	assert.Equal(t, "module", result.Scope)
	assert.Greater(t, result.TotalCommunities, 0)
	assert.NotEmpty(t, result.Communities)
	assert.NotEmpty(t, result.GeneratedAt)
	assert.NotEmpty(t, result.Version)

	totalModules := 0
	for _, community := range result.Communities {
		assert.NotEmpty(t, community.ID)
		assert.Greater(t, community.Size, 0)
		assert.NotEmpty(t, community.Modules)
		totalModules += community.Size
	}
	assert.Greater(t, totalModules, 0)
}

func TestCommunityUseCaseBuilder_RequiresDependencies(t *testing.T) {
	_, err := NewCommunityUseCaseBuilder().Build()
	require.Error(t, err)

	_, err = NewCommunityUseCaseBuilder().
		WithService(service.NewCommunityAnalysisService()).
		Build()
	require.Error(t, err)

	_, err = NewCommunityUseCaseBuilder().
		WithService(service.NewCommunityAnalysisService()).
		WithFileReader(service.NewFileReader()).
		Build()
	require.Error(t, err)
}

func TestCommunityUseCase_ValidateRequest(t *testing.T) {
	uc := NewCommunityUseCase(
		service.NewCommunityAnalysisService(),
		service.NewFileReader(),
		noopCommunityFormatter{},
		nil,
	)

	err := uc.validateRequest(domain.CommunityAnalysisRequest{})
	require.Error(t, err)

	err = uc.validateRequest(domain.CommunityAnalysisRequest{
		Paths:            []string{"."},
		OutputWriter:     os.Stdout,
		MinCommunitySize: -1,
	})
	require.Error(t, err)
}
