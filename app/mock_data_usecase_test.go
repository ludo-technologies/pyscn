package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	return args.Get(0).(*domain.MockDataRequest)
}

type passthroughReportWriter struct {
	err error
}

func (w passthroughReportWriter) Write(writer io.Writer, _ string, _ domain.OutputFormat, _ bool, writeFunc func(io.Writer) error) error {
	if w.err != nil {
		return w.err
	}
	return writeFunc(writer)
}

func newMockDataUseCaseForTest() (*MockDataUseCase, *mockMockDataService, *mockFileReader, *mockMockDataFormatter, *mockMockDataConfigLoader) {
	service := &mockMockDataService{}
	fileReader := &mockFileReader{}
	formatter := &mockMockDataFormatter{}
	configLoader := &mockMockDataConfigLoader{}
	useCase := NewMockDataUseCase(service, fileReader, formatter, configLoader)
	useCase.output = passthroughReportWriter{}
	return useCase, service, fileReader, formatter, configLoader
}

func validMockDataRequest() domain.MockDataRequest {
	req := *domain.DefaultMockDataRequest()
	req.Paths = []string{"src"}
	req.OutputWriter = &bytes.Buffer{}
	return req
}

func mockDataResponse() *domain.MockDataResponse {
	return &domain.MockDataResponse{
		Files:   []domain.FileMockData{{FilePath: "src/example.py"}},
		Summary: domain.MockDataSummary{TotalFiles: 1},
	}
}

func TestMockDataUseCaseValidateRequest(t *testing.T) {
	useCase := &MockDataUseCase{}

	assert.NoError(t, useCase.validateRequest(domain.MockDataRequest{Paths: []string{"src"}}))
	err := useCase.validateRequest(domain.MockDataRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one path must be specified")
}

func TestMockDataUseCaseLoadAndMergeConfig(t *testing.T) {
	t.Run("passes through without loader", func(t *testing.T) {
		req := validMockDataRequest()
		useCase := &MockDataUseCase{}

		got, err := useCase.loadAndMergeConfig(req)

		require.NoError(t, err)
		assert.Equal(t, req, got)
	})

	t.Run("loads explicit config and merges request", func(t *testing.T) {
		loader := &mockMockDataConfigLoader{}
		useCase := &MockDataUseCase{configLoader: loader}
		req := validMockDataRequest()
		req.ConfigPath = "pyscn.yaml"
		configReq := &domain.MockDataRequest{Recursive: domain.BoolPtr(false)}
		mergedReq := req
		mergedReq.Recursive = domain.BoolPtr(false)
		loader.On("LoadConfig", "pyscn.yaml").Return(configReq, nil)
		loader.On("MergeConfig", configReq, mock.AnythingOfType("*domain.MockDataRequest")).Return(&mergedReq)

		got, err := useCase.loadAndMergeConfig(req)

		require.NoError(t, err)
		assert.Equal(t, mergedReq, got)
		loader.AssertExpectations(t)
	})

	t.Run("loads default config", func(t *testing.T) {
		loader := &mockMockDataConfigLoader{}
		useCase := &MockDataUseCase{configLoader: loader}
		req := validMockDataRequest()
		configReq := &domain.MockDataRequest{MinSeverity: domain.MockDataSeverityError}
		mergedReq := req
		mergedReq.MinSeverity = domain.MockDataSeverityError
		loader.On("LoadDefaultConfig").Return(configReq)
		loader.On("MergeConfig", configReq, mock.AnythingOfType("*domain.MockDataRequest")).Return(&mergedReq)

		got, err := useCase.loadAndMergeConfig(req)

		require.NoError(t, err)
		assert.Equal(t, mergedReq, got)
		loader.AssertExpectations(t)
	})

	t.Run("returns explicit config error", func(t *testing.T) {
		loader := &mockMockDataConfigLoader{}
		useCase := &MockDataUseCase{configLoader: loader}
		req := validMockDataRequest()
		req.ConfigPath = "missing.yaml"
		loader.On("LoadConfig", "missing.yaml").Return((*domain.MockDataRequest)(nil), errors.New("missing"))

		_, err := useCase.loadAndMergeConfig(req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load config from missing.yaml")
	})
}

func TestMockDataUseCaseExecute(t *testing.T) {
	t.Run("writes successful analysis", func(t *testing.T) {
		useCase, service, fileReader, formatter, configLoader := newMockDataUseCaseForTest()
		req := validMockDataRequest()
		response := mockDataResponse()
		configLoader.On("LoadDefaultConfig").Return((*domain.MockDataRequest)(nil))
		fileReader.On("CollectPythonFiles", req.Paths, true, req.IncludePatterns, req.ExcludePatterns).
			Return([]string{"src/example.py"}, nil)
		service.On("Analyze", mock.Anything, mock.MatchedBy(func(got domain.MockDataRequest) bool {
			return assert.ObjectsAreEqual([]string{"src/example.py"}, got.Paths)
		})).Return(response, nil)
		formatter.On("Write", response, domain.OutputFormatText, req.OutputWriter).Return(nil)

		err := useCase.Execute(context.Background(), req)

		require.NoError(t, err)
		service.AssertExpectations(t)
		fileReader.AssertExpectations(t)
		formatter.AssertExpectations(t)
	})

	tests := []struct {
		name     string
		setup    func(*MockDataUseCase, *mockMockDataService, *mockFileReader, *mockMockDataConfigLoader, domain.MockDataRequest)
		request  domain.MockDataRequest
		contains string
	}{
		{
			name: "rejects empty paths",
			setup: func(*MockDataUseCase, *mockMockDataService, *mockFileReader, *mockMockDataConfigLoader, domain.MockDataRequest) {
			},
			request:  domain.MockDataRequest{},
			contains: "invalid request",
		},
		{
			name: "rejects empty collection",
			setup: func(_ *MockDataUseCase, _ *mockMockDataService, reader *mockFileReader, loader *mockMockDataConfigLoader, req domain.MockDataRequest) {
				loader.On("LoadDefaultConfig").Return((*domain.MockDataRequest)(nil))
				reader.On("CollectPythonFiles", req.Paths, true, req.IncludePatterns, req.ExcludePatterns).Return([]string{}, nil)
			},
			request:  validMockDataRequest(),
			contains: "no Python files found",
		},
		{
			name: "wraps service failure",
			setup: func(_ *MockDataUseCase, service *mockMockDataService, reader *mockFileReader, loader *mockMockDataConfigLoader, req domain.MockDataRequest) {
				loader.On("LoadDefaultConfig").Return((*domain.MockDataRequest)(nil))
				reader.On("CollectPythonFiles", req.Paths, true, req.IncludePatterns, req.ExcludePatterns).Return([]string{"src/example.py"}, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.MockDataRequest")).Return((*domain.MockDataResponse)(nil), errors.New("parse failed"))
			},
			request:  validMockDataRequest(),
			contains: "mock data analysis failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, _, configLoader := newMockDataUseCaseForTest()
			tt.setup(useCase, service, fileReader, configLoader, tt.request)

			err := useCase.Execute(context.Background(), tt.request)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.contains)
		})
	}

	t.Run("wraps report writer failure", func(t *testing.T) {
		useCase, service, fileReader, _, configLoader := newMockDataUseCaseForTest()
		useCase.output = passthroughReportWriter{err: errors.New("disk full")}
		req := validMockDataRequest()
		configLoader.On("LoadDefaultConfig").Return((*domain.MockDataRequest)(nil))
		fileReader.On("CollectPythonFiles", req.Paths, true, req.IncludePatterns, req.ExcludePatterns).Return([]string{"src/example.py"}, nil)
		service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.MockDataRequest")).Return(mockDataResponse(), nil)

		err := useCase.Execute(context.Background(), req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to write output")
	})
}

func TestMockDataUseCaseAnalyzeAndReturn(t *testing.T) {
	t.Run("returns successful analysis", func(t *testing.T) {
		useCase, service, fileReader, _, configLoader := newMockDataUseCaseForTest()
		req := validMockDataRequest()
		response := mockDataResponse()
		configLoader.On("LoadDefaultConfig").Return((*domain.MockDataRequest)(nil))
		fileReader.On("FileExists", "src").Return(false, nil)
		fileReader.On("CollectPythonFiles", req.Paths, true, req.IncludePatterns, req.ExcludePatterns).
			Return([]string{"src/example.py"}, nil)
		service.On("Analyze", mock.Anything, mock.MatchedBy(func(got domain.MockDataRequest) bool {
			return assert.ObjectsAreEqual([]string{"src/example.py"}, got.Paths)
		})).Return(response, nil)

		got, err := useCase.AnalyzeAndReturn(context.Background(), req)

		require.NoError(t, err)
		assert.Same(t, response, got)
	})

	tests := []struct {
		name     string
		request  domain.MockDataRequest
		setup    func(*mockMockDataService, *mockFileReader, *mockMockDataConfigLoader, domain.MockDataRequest)
		contains string
	}{
		{
			name:    "rejects empty paths",
			request: domain.MockDataRequest{},
			setup: func(*mockMockDataService, *mockFileReader, *mockMockDataConfigLoader, domain.MockDataRequest) {
			},
			contains: "invalid request",
		},
		{
			name:    "rejects empty collection",
			request: validMockDataRequest(),
			setup: func(_ *mockMockDataService, reader *mockFileReader, loader *mockMockDataConfigLoader, req domain.MockDataRequest) {
				loader.On("LoadDefaultConfig").Return((*domain.MockDataRequest)(nil))
				reader.On("FileExists", "src").Return(false, nil)
				reader.On("CollectPythonFiles", req.Paths, true, req.IncludePatterns, req.ExcludePatterns).Return([]string{}, nil)
			},
			contains: "no Python files found",
		},
		{
			name:    "wraps service failure",
			request: validMockDataRequest(),
			setup: func(service *mockMockDataService, reader *mockFileReader, loader *mockMockDataConfigLoader, req domain.MockDataRequest) {
				loader.On("LoadDefaultConfig").Return((*domain.MockDataRequest)(nil))
				reader.On("FileExists", "src").Return(false, nil)
				reader.On("CollectPythonFiles", req.Paths, true, req.IncludePatterns, req.ExcludePatterns).Return([]string{"src/example.py"}, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.MockDataRequest")).Return((*domain.MockDataResponse)(nil), errors.New("parse failed"))
			},
			contains: "mock data analysis failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, reader, _, loader := newMockDataUseCaseForTest()
			tt.setup(service, reader, loader, tt.request)

			response, err := useCase.AnalyzeAndReturn(context.Background(), tt.request)

			require.Error(t, err)
			assert.Nil(t, response)
			assert.Contains(t, err.Error(), tt.contains)
		})
	}
}

func TestMockDataUseCaseAnalyzeFile(t *testing.T) {
	t.Run("builds and writes single-file response", func(t *testing.T) {
		useCase, service, fileReader, formatter, configLoader := newMockDataUseCaseForTest()
		req := validMockDataRequest()
		result := &domain.FileMockData{
			FilePath:      "src/example.py",
			Findings:      []domain.MockDataFinding{{Type: domain.MockDataTypeEmail}},
			TotalFindings: 1,
			WarningCount:  1,
		}
		fileReader.On("IsValidPythonFile", "src/example.py").Return(true)
		fileReader.On("FileExists", "src/example.py").Return(true, nil)
		configLoader.On("LoadDefaultConfig").Return((*domain.MockDataRequest)(nil))
		service.On("AnalyzeFile", mock.Anything, "src/example.py", req).Return(result, nil)
		formatter.On("Write", mock.MatchedBy(func(response *domain.MockDataResponse) bool {
			return response.Summary.TotalFiles == 1 &&
				response.Summary.TotalFindings == 1 &&
				response.Summary.FilesWithMockData == 1 &&
				response.Summary.FindingsByType[domain.MockDataTypeEmail] == 1
		}), domain.OutputFormatText, req.OutputWriter).Return(nil)

		err := useCase.AnalyzeFile(context.Background(), "src/example.py", req)

		require.NoError(t, err)
		formatter.AssertExpectations(t)
	})

	tests := []struct {
		name     string
		path     string
		setup    func(*mockMockDataService, *mockFileReader, *mockMockDataConfigLoader, domain.MockDataRequest)
		contains string
	}{
		{
			name: "rejects non-Python file",
			path: "notes.txt",
			setup: func(_ *mockMockDataService, reader *mockFileReader, _ *mockMockDataConfigLoader, _ domain.MockDataRequest) {
				reader.On("IsValidPythonFile", "notes.txt").Return(false)
			},
			contains: "not a valid Python file",
		},
		{
			name: "rejects missing file",
			path: "missing.py",
			setup: func(_ *mockMockDataService, reader *mockFileReader, _ *mockMockDataConfigLoader, _ domain.MockDataRequest) {
				reader.On("IsValidPythonFile", "missing.py").Return(true)
				reader.On("FileExists", "missing.py").Return(false, nil)
			},
			contains: "file does not exist",
		},
		{
			name: "wraps analysis error",
			path: "src/example.py",
			setup: func(service *mockMockDataService, reader *mockFileReader, loader *mockMockDataConfigLoader, req domain.MockDataRequest) {
				reader.On("IsValidPythonFile", "src/example.py").Return(true)
				reader.On("FileExists", "src/example.py").Return(true, nil)
				loader.On("LoadDefaultConfig").Return((*domain.MockDataRequest)(nil))
				service.On("AnalyzeFile", mock.Anything, "src/example.py", req).Return((*domain.FileMockData)(nil), errors.New("parse failed"))
			},
			contains: "mock data analysis failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, reader, _, loader := newMockDataUseCaseForTest()
			req := validMockDataRequest()
			tt.setup(service, reader, loader, req)

			err := useCase.AnalyzeFile(context.Background(), tt.path, req)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.contains)
		})
	}
}

func TestMockDataUseCaseBuilder(t *testing.T) {
	service := &mockMockDataService{}
	reader := &mockFileReader{}
	formatter := &mockMockDataFormatter{}
	output := passthroughReportWriter{}
	builder := NewMockDataUseCaseBuilder()

	assert.Same(t, builder, builder.WithService(service))
	assert.Same(t, builder, builder.WithFileReader(reader))
	assert.Same(t, builder, builder.WithFormatter(formatter))
	assert.Same(t, builder, builder.WithConfigLoader(nil))
	assert.Same(t, builder, builder.WithOutput(output))

	useCase, err := builder.Build()
	require.NoError(t, err)
	assert.Same(t, service, useCase.service)
	assert.Same(t, reader, useCase.fileReader)
	assert.Same(t, formatter, useCase.formatter)
	assert.Equal(t, output, useCase.output)

	withDefaultOutput, err := NewMockDataUseCaseBuilder().
		WithService(service).
		WithFileReader(reader).
		WithFormatter(formatter).
		Build()
	require.NoError(t, err)
	assert.NotNil(t, withDefaultOutput.output)

	for _, tt := range []struct {
		name    string
		builder *MockDataUseCaseBuilder
		wantErr string
	}{
		{"missing service", NewMockDataUseCaseBuilder(), "service is required"},
		{"missing file reader", NewMockDataUseCaseBuilder().WithService(service), "file reader is required"},
		{"missing formatter", NewMockDataUseCaseBuilder().WithService(service).WithFileReader(reader), "formatter is required"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.builder.Build()
			require.Error(t, err)
			assert.EqualError(t, err, tt.wantErr)
		})
	}

	defaults, err := NewMockDataUseCaseBuilder().BuildWithDefaults()
	require.NoError(t, err)
	assert.NotNil(t, defaults.service)
	assert.NotNil(t, defaults.fileReader)
	assert.NotNil(t, defaults.formatter)
	assert.NotNil(t, defaults.configLoader)
	assert.NotNil(t, defaults.output)
}
