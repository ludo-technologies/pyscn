package app

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMockDataService struct {
	mock.Mock
}

func (m *mockMockDataService) Analyze(ctx context.Context, req domain.MockDataRequest) (*domain.MockDataResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MockDataResponse), args.Error(1)
}

func (m *mockMockDataService) AnalyzeFile(ctx context.Context, filePath string, req domain.MockDataRequest) (*domain.FileMockData, error) {
	args := m.Called(ctx, filePath, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FileMockData), args.Error(1)
}

type mockMockDataFormatter struct {
	mock.Mock
}

func (m *mockMockDataFormatter) Format(response *domain.MockDataResponse, format domain.OutputFormat) (string, error) {
	args := m.Called(response, format)
	return args.String(0), args.Error(1)
}

func (m *mockMockDataFormatter) Write(response *domain.MockDataResponse, format domain.OutputFormat, writer io.Writer) error {
	args := m.Called(response, format, writer)
	return args.Error(0)
}

type mockMockDataConfigLoader struct {
	mock.Mock
}

func (m *mockMockDataConfigLoader) LoadConfig(path string) (*domain.MockDataRequest, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MockDataRequest), args.Error(1)
}

func (m *mockMockDataConfigLoader) LoadDefaultConfig() *domain.MockDataRequest {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.MockDataRequest)
}

func (m *mockMockDataConfigLoader) MergeConfig(base *domain.MockDataRequest, override *domain.MockDataRequest) *domain.MockDataRequest {
	args := m.Called(base, override)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.MockDataRequest)
}

type mockReportWriter struct {
	mock.Mock
}

func (m *mockReportWriter) Write(writer io.Writer, outputPath string, format domain.OutputFormat, noOpen bool, writeFunc func(io.Writer) error) error {
	args := m.Called(writer, outputPath, format, noOpen)
	if invoke, ok := args.Get(0).(bool); ok && invoke {
		if err := writeFunc(writer); err != nil {
			return err
		}
	}
	return args.Error(1)
}

func createValidMockDataRequest() domain.MockDataRequest {
	return domain.MockDataRequest{
		Paths:           []string{"/tmp/sample.py"},
		OutputWriter:    os.Stdout,
		OutputFormat:    domain.OutputFormatText,
		Recursive:       true,
		IncludePatterns: []string{"**/*.py"},
		ExcludePatterns: []string{},
	}
}

func createMockDataResponse() *domain.MockDataResponse {
	return &domain.MockDataResponse{
		Files: []domain.FileMockData{
			{FilePath: "/tmp/sample.py", TotalFindings: 1, WarningCount: 1},
		},
		Summary: domain.MockDataSummary{
			TotalFiles:        1,
			TotalFindings:     1,
			FilesWithMockData: 1,
		},
	}
}

func TestMockDataUseCase_Execute_Success(t *testing.T) {
	serviceMock := &mockMockDataService{}
	fileReader := &mockFileReader{}
	formatter := &mockMockDataFormatter{}
	configLoader := &mockMockDataConfigLoader{}
	reportWriter := &mockReportWriter{}

	uc := NewMockDataUseCase(serviceMock, fileReader, formatter, configLoader)
	uc.output = reportWriter

	req := createValidMockDataRequest()
	files := []string{"/tmp/sample.py"}
	response := createMockDataResponse()

	configLoader.On("LoadDefaultConfig").Return((*domain.MockDataRequest)(nil))
	fileReader.On("CollectPythonFiles", req.Paths, true, req.IncludePatterns, req.ExcludePatterns).Return(files, nil)
	serviceMock.On("Analyze", mock.Anything, mock.AnythingOfType("domain.MockDataRequest")).Return(response, nil)
	reportWriter.On("Write", req.OutputWriter, "", domain.OutputFormatText, false).Return(true, nil)
	formatter.On("Write", response, domain.OutputFormatText, req.OutputWriter).Return(nil)

	err := uc.Execute(context.Background(), req)

	assert.NoError(t, err)
	serviceMock.AssertExpectations(t)
	fileReader.AssertExpectations(t)
	formatter.AssertExpectations(t)
	reportWriter.AssertExpectations(t)
}

func TestMockDataUseCase_Execute_InvalidRequest(t *testing.T) {
	uc := NewMockDataUseCase(
		&mockMockDataService{},
		&mockFileReader{},
		&mockMockDataFormatter{},
		&mockMockDataConfigLoader{},
	)

	err := uc.Execute(context.Background(), domain.MockDataRequest{OutputWriter: os.Stdout})

	assert.Error(t, err)
	assert.ErrorContains(t, err, "invalid request")
	assert.ErrorContains(t, err, domain.ErrCodeInvalidInput)
}

func TestMockDataUseCase_AnalyzeAndReturn_Success(t *testing.T) {
	serviceMock := &mockMockDataService{}
	fileReader := &mockFileReader{}
	formatter := &mockMockDataFormatter{}
	configLoader := &mockMockDataConfigLoader{}

	uc := NewMockDataUseCase(serviceMock, fileReader, formatter, configLoader)

	req := createValidMockDataRequest()
	response := createMockDataResponse()

	configLoader.On("LoadDefaultConfig").Return((*domain.MockDataRequest)(nil))
	fileReader.On("FileExists", "/tmp/sample.py").Return(true, nil)
	serviceMock.On("Analyze", mock.Anything, mock.AnythingOfType("domain.MockDataRequest")).Return(response, nil)

	result, err := uc.AnalyzeAndReturn(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, response, result)
	serviceMock.AssertExpectations(t)
	fileReader.AssertExpectations(t)
}

func TestMockDataUseCase_AnalyzeFile_InvalidPythonFile(t *testing.T) {
	serviceMock := &mockMockDataService{}
	fileReader := &mockFileReader{}
	formatter := &mockMockDataFormatter{}

	uc := NewMockDataUseCase(serviceMock, fileReader, formatter, nil)

	fileReader.On("IsValidPythonFile", "not_python.txt").Return(false)

	err := uc.AnalyzeFile(context.Background(), "not_python.txt", createValidMockDataRequest())

	assert.Error(t, err)
	assert.ErrorContains(t, err, "not a valid Python file")
	assert.ErrorContains(t, err, domain.ErrCodeInvalidInput)
	fileReader.AssertExpectations(t)
}

func TestMockDataUseCase_LoadAndMergeConfig_LoadError(t *testing.T) {
	configLoader := &mockMockDataConfigLoader{}
	uc := NewMockDataUseCase(
		&mockMockDataService{},
		&mockFileReader{},
		&mockMockDataFormatter{},
		configLoader,
	)

	req := createValidMockDataRequest()
	req.ConfigPath = "/tmp/missing.toml"
	configLoader.On("LoadConfig", req.ConfigPath).Return((*domain.MockDataRequest)(nil), errors.New("missing file"))

	_, err := uc.loadAndMergeConfig(req)

	assert.Error(t, err)
	assert.ErrorContains(t, err, "failed to load config")
}

func TestMockDataUseCaseBuilder_BuildAndDefaults(t *testing.T) {
	builder := NewMockDataUseCaseBuilder()

	_, err := builder.Build()
	assert.Error(t, err)
	assert.ErrorContains(t, err, "service is required")

	useCase, err := builder.
		WithService(service.NewMockDataService()).
		WithFileReader(service.NewFileReader()).
		WithFormatter(service.NewMockDataFormatter()).
		Build()
	assert.NoError(t, err)
	assert.NotNil(t, useCase)

	defaultUseCase, err := NewMockDataUseCaseBuilder().BuildWithDefaults()
	assert.NoError(t, err)
	assert.NotNil(t, defaultUseCase)
}
