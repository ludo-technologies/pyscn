package app

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	svc "github.com/ludo-technologies/pyscn/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockLCOMService struct {
	mock.Mock
}

func (m *mockLCOMService) Analyze(ctx context.Context, req domain.LCOMRequest) (*domain.LCOMResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LCOMResponse), args.Error(1)
}

func (m *mockLCOMService) AnalyzeFile(ctx context.Context, filePath string, req domain.LCOMRequest) (*domain.LCOMResponse, error) {
	args := m.Called(ctx, filePath, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LCOMResponse), args.Error(1)
}

type mockSnapshotLCOMService struct {
	mockLCOMService
}

func (m *mockSnapshotLCOMService) AnalyzeSnapshot(ctx context.Context, snapshot *svc.ProjectSnapshot, req domain.LCOMRequest) (*domain.LCOMResponse, error) {
	args := m.Called(ctx, snapshot, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LCOMResponse), args.Error(1)
}

type mockLCOMFileReader struct {
	mock.Mock
}

func (m *mockLCOMFileReader) CollectPythonFiles(paths []string, recursive bool, includePatterns, excludePatterns []string) ([]string, error) {
	args := m.Called(paths, recursive, includePatterns, excludePatterns)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockLCOMFileReader) ReadFile(path string) ([]byte, error) {
	args := m.Called(path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockLCOMFileReader) IsValidPythonFile(path string) bool {
	args := m.Called(path)
	return args.Bool(0)
}

func (m *mockLCOMFileReader) FileExists(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

type mockLCOMOutputFormatter struct {
	mock.Mock
}

func (m *mockLCOMOutputFormatter) Format(response *domain.LCOMResponse, format domain.OutputFormat) (string, error) {
	args := m.Called(response, format)
	return args.String(0), args.Error(1)
}

func (m *mockLCOMOutputFormatter) Write(response *domain.LCOMResponse, format domain.OutputFormat, writer io.Writer) error {
	args := m.Called(response, format, writer)
	return args.Error(0)
}

type mockLCOMConfigurationLoader struct {
	mock.Mock
}

func (m *mockLCOMConfigurationLoader) LoadConfig(path string) (*domain.LCOMRequest, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LCOMRequest), args.Error(1)
}

func (m *mockLCOMConfigurationLoader) LoadDefaultConfig() *domain.LCOMRequest {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.LCOMRequest)
}

func (m *mockLCOMConfigurationLoader) MergeConfig(base *domain.LCOMRequest, override *domain.LCOMRequest) *domain.LCOMRequest {
	args := m.Called(base, override)
	return args.Get(0).(*domain.LCOMRequest)
}

type stubLCOMReportWriter struct {
	calls      int
	writer     io.Writer
	outputPath string
	format     domain.OutputFormat
	noOpen     bool
	err        error
}

func (w *stubLCOMReportWriter) Write(writer io.Writer, outputPath string, format domain.OutputFormat, noOpen bool, writeFunc func(io.Writer) error) error {
	w.calls++
	w.writer = writer
	w.outputPath = outputPath
	w.format = format
	w.noOpen = noOpen
	if w.err != nil {
		return w.err
	}
	return writeFunc(writer)
}

func setupLCOMUseCaseMocks() (*LCOMUseCase, *mockLCOMService, *mockLCOMFileReader, *mockLCOMOutputFormatter, *mockLCOMConfigurationLoader, *stubLCOMReportWriter) {
	service := &mockLCOMService{}
	fileReader := &mockLCOMFileReader{}
	formatter := &mockLCOMOutputFormatter{}
	configLoader := &mockLCOMConfigurationLoader{}
	output := &stubLCOMReportWriter{}

	useCase := NewLCOMUseCase(service, fileReader, formatter, configLoader)
	useCase.output = output

	return useCase, service, fileReader, formatter, configLoader, output
}

func createValidLCOMRequest() domain.LCOMRequest {
	return domain.LCOMRequest{
		Paths:           []string{"/test/file.py"},
		OutputWriter:    os.Stdout,
		OutputFormat:    domain.OutputFormatText,
		SortBy:          domain.SortByCohesion,
		MinLCOM:         0,
		MaxLCOM:         10,
		LowThreshold:    2,
		MediumThreshold: 5,
		Recursive:       domain.BoolPtr(true),
		IncludePatterns: []string{"**/*.py"},
		ExcludePatterns: []string{},
	}
}

func createMockLCOMResponse() *domain.LCOMResponse {
	return &domain.LCOMResponse{
		Classes: []domain.ClassCohesion{
			{
				Name:      "TestClass",
				FilePath:  "/test/file.py",
				StartLine: 1,
				EndLine:   20,
				Metrics: domain.LCOMMetrics{
					LCOM4:             3,
					TotalMethods:      4,
					ExcludedMethods:   1,
					InstanceVariables: 2,
					MethodGroups:      [][]string{{"method_one", "method_two"}, {"method_three"}},
				},
				RiskLevel: domain.RiskLevelMedium,
			},
		},
		Summary: domain.LCOMSummary{
			TotalClasses:      1,
			AverageLCOM:       3.0,
			MaxLCOM:           3,
			MinLCOM:           3,
			ClassesAnalyzed:   1,
			FilesAnalyzed:     1,
			MediumRiskClasses: 1,
			LCOMDistribution:  map[string]int{"3": 1},
		},
		GeneratedAt: "2025-01-01T00:00:00Z",
		Version:     "test",
	}
}

func createLCOMSnapshot() *svc.ProjectSnapshot {
	return &svc.ProjectSnapshot{
		Files: []*svc.ProjectFile{
			{Path: "/snapshot/one.py"},
			{Path: "/snapshot/two.py"},
		},
	}
}

func assertDomainError(t *testing.T, err error, code string, contains string) {
	t.Helper()

	var domainErr domain.DomainError
	require.ErrorAs(t, err, &domainErr)
	assert.Equal(t, code, domainErr.Code)
	if contains != "" {
		assert.Contains(t, err.Error(), contains)
	}
}

func TestLCOMUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockLCOMService, *mockLCOMFileReader, *mockLCOMOutputFormatter, *mockLCOMConfigurationLoader, *stubLCOMReportWriter)
		request     domain.LCOMRequest
		expectError bool
		errorCode   string
		errorMsg    string
		errorIs     error
	}{
		{
			name: "successful execution with valid request",
			setupMocks: func(service *mockLCOMService, fileReader *mockLCOMFileReader, formatter *mockLCOMOutputFormatter, configLoader *mockLCOMConfigurationLoader, output *stubLCOMReportWriter) {
				response := createMockLCOMResponse()
				configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.LCOMRequest")).Return(response, nil)
				formatter.On("Write", response, domain.OutputFormatText, os.Stdout).Return(nil)
			},
			request: createValidLCOMRequest(),
		},
		{
			name: "validation error is wrapped",
			setupMocks: func(service *mockLCOMService, fileReader *mockLCOMFileReader, formatter *mockLCOMOutputFormatter, configLoader *mockLCOMConfigurationLoader, output *stubLCOMReportWriter) {
				configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
			},
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.Paths = nil
				return req
			}(),
			expectError: true,
			errorCode:   domain.ErrCodeInvalidInput,
			errorMsg:    "no input paths specified",
		},
		{
			name: "configuration loading error is wrapped",
			setupMocks: func(service *mockLCOMService, fileReader *mockLCOMFileReader, formatter *mockLCOMOutputFormatter, configLoader *mockLCOMConfigurationLoader, output *stubLCOMReportWriter) {
				configLoader.On("LoadConfig", "/invalid/config.yaml").Return((*domain.LCOMRequest)(nil), errors.New("config file not found"))
			},
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.ConfigPath = "/invalid/config.yaml"
				return req
			}(),
			expectError: true,
			errorCode:   domain.ErrCodeConfigError,
			errorMsg:    "failed to load configuration",
		},
		{
			name: "file collection error is wrapped",
			setupMocks: func(service *mockLCOMService, fileReader *mockLCOMFileReader, formatter *mockLCOMOutputFormatter, configLoader *mockLCOMConfigurationLoader, output *stubLCOMReportWriter) {
				collectErr := errors.New("path not found")
				configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
				fileReader.On("FileExists", "/invalid/path").Return(false, nil)
				fileReader.On("CollectPythonFiles", []string{"/invalid/path"}, true, []string{"**/*.py"}, []string{}).Return([]string{}, collectErr)
			},
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.Paths = []string{"/invalid/path"}
				return req
			}(),
			expectError: true,
			errorCode:   domain.ErrCodeFileNotFound,
			errorMsg:    "failed to collect files",
		},
		{
			name: "no files found is invalid input",
			setupMocks: func(service *mockLCOMService, fileReader *mockLCOMFileReader, formatter *mockLCOMOutputFormatter, configLoader *mockLCOMConfigurationLoader, output *stubLCOMReportWriter) {
				configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
				fileReader.On("FileExists", "/empty/path").Return(false, nil)
				fileReader.On("CollectPythonFiles", []string{"/empty/path"}, true, []string{"**/*.py"}, []string{}).Return([]string{}, nil)
			},
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.Paths = []string{"/empty/path"}
				return req
			}(),
			expectError: true,
			errorCode:   domain.ErrCodeInvalidInput,
			errorMsg:    "no Python files found",
		},
		{
			name: "analysis service error is wrapped",
			setupMocks: func(service *mockLCOMService, fileReader *mockLCOMFileReader, formatter *mockLCOMOutputFormatter, configLoader *mockLCOMConfigurationLoader, output *stubLCOMReportWriter) {
				analysisErr := errors.New("parsing failed")
				configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.LCOMRequest")).Return((*domain.LCOMResponse)(nil), analysisErr)
			},
			request:     createValidLCOMRequest(),
			expectError: true,
			errorCode:   domain.ErrCodeAnalysisError,
			errorMsg:    "LCOM analysis failed",
		},
		{
			name: "output formatting error is wrapped",
			setupMocks: func(service *mockLCOMService, fileReader *mockLCOMFileReader, formatter *mockLCOMOutputFormatter, configLoader *mockLCOMConfigurationLoader, output *stubLCOMReportWriter) {
				response := createMockLCOMResponse()
				writeErr := errors.New("write failed")
				configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.LCOMRequest")).Return(response, nil)
				formatter.On("Write", response, domain.OutputFormatText, os.Stdout).Return(writeErr)
			},
			request:     createValidLCOMRequest(),
			expectError: true,
			errorCode:   domain.ErrCodeOutputError,
			errorMsg:    "failed to write output",
		},
		{
			name: "report writer error is wrapped",
			setupMocks: func(service *mockLCOMService, fileReader *mockLCOMFileReader, formatter *mockLCOMOutputFormatter, configLoader *mockLCOMConfigurationLoader, output *stubLCOMReportWriter) {
				output.err = errors.New("output failed")
				configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.LCOMRequest")).Return(createMockLCOMResponse(), nil)
			},
			request:     createValidLCOMRequest(),
			expectError: true,
			errorCode:   domain.ErrCodeOutputError,
			errorMsg:    "failed to write output",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader, output := setupLCOMUseCaseMocks()
			tt.setupMocks(service, fileReader, formatter, configLoader, output)

			err := useCase.Execute(context.Background(), tt.request)

			if tt.expectError {
				require.Error(t, err)
				assertDomainError(t, err, tt.errorCode, tt.errorMsg)
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, 1, output.calls)
				assert.Equal(t, os.Stdout, output.writer)
				assert.Equal(t, domain.OutputFormatText, output.format)
			}

			service.AssertExpectations(t)
			fileReader.AssertExpectations(t)
			formatter.AssertExpectations(t)
			configLoader.AssertExpectations(t)
		})
	}
}

func TestLCOMUseCase_AnalyzeAndReturn(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockLCOMService, *mockLCOMFileReader, *mockLCOMOutputFormatter, *mockLCOMConfigurationLoader)
		request     domain.LCOMRequest
		expectError bool
		errorCode   string
		errorMsg    string
	}{
		{
			name: "successful analysis without formatting",
			setupMocks: func(service *mockLCOMService, fileReader *mockLCOMFileReader, formatter *mockLCOMOutputFormatter, configLoader *mockLCOMConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.LCOMRequest")).Return(createMockLCOMResponse(), nil)
			},
			request: createValidLCOMRequest(),
		},
		{
			name: "validation error in analyze and return",
			setupMocks: func(service *mockLCOMService, fileReader *mockLCOMFileReader, formatter *mockLCOMOutputFormatter, configLoader *mockLCOMConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
			},
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.OutputWriter = nil
				req.OutputPath = ""
				return req
			}(),
			expectError: true,
			errorCode:   domain.ErrCodeInvalidInput,
			errorMsg:    "output writer or output path is required",
		},
		{
			name: "analysis error in analyze and return",
			setupMocks: func(service *mockLCOMService, fileReader *mockLCOMFileReader, formatter *mockLCOMOutputFormatter, configLoader *mockLCOMConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.LCOMRequest")).Return((*domain.LCOMResponse)(nil), errors.New("analysis failed"))
			},
			request:     createValidLCOMRequest(),
			expectError: true,
			errorCode:   domain.ErrCodeAnalysisError,
			errorMsg:    "LCOM analysis failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader, _ := setupLCOMUseCaseMocks()
			tt.setupMocks(service, fileReader, formatter, configLoader)

			response, err := useCase.AnalyzeAndReturn(context.Background(), tt.request)

			if tt.expectError {
				require.Error(t, err)
				assert.Nil(t, response)
				assertDomainError(t, err, tt.errorCode, tt.errorMsg)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				require.Len(t, response.Classes, 1)
				assert.Equal(t, "TestClass", response.Classes[0].Name)
			}

			service.AssertExpectations(t)
			fileReader.AssertExpectations(t)
			formatter.AssertExpectations(t)
			configLoader.AssertExpectations(t)
		})
	}
}

func TestLCOMUseCase_analyzeSnapshotRequest(t *testing.T) {
	t.Run("nil snapshot returns analysis error", func(t *testing.T) {
		useCase := &LCOMUseCase{service: &mockSnapshotLCOMService{}}

		response, err := useCase.analyzeSnapshotRequest(context.Background(), nil, createValidLCOMRequest())

		require.Error(t, err)
		assert.Nil(t, response)
		assertDomainError(t, err, domain.ErrCodeAnalysisError, "project snapshot is required")
	})

	t.Run("config loading error is wrapped", func(t *testing.T) {
		configLoader := &mockLCOMConfigurationLoader{}
		configErr := errors.New("config not found")
		req := createValidLCOMRequest()
		req.ConfigPath = "/invalid.yaml"
		configLoader.On("LoadConfig", "/invalid.yaml").Return((*domain.LCOMRequest)(nil), configErr)
		useCase := &LCOMUseCase{service: &mockSnapshotLCOMService{}, configLoader: configLoader}

		response, err := useCase.analyzeSnapshotRequest(context.Background(), createLCOMSnapshot(), req)

		require.Error(t, err)
		assert.Nil(t, response)
		assertDomainError(t, err, domain.ErrCodeConfigError, "failed to load configuration")
		assert.ErrorIs(t, err, configErr)
		configLoader.AssertExpectations(t)
	})

	t.Run("validation error is wrapped", func(t *testing.T) {
		configLoader := &mockLCOMConfigurationLoader{}
		configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
		useCase := &LCOMUseCase{service: &mockSnapshotLCOMService{}, configLoader: configLoader}
		req := createValidLCOMRequest()
		req.Paths = nil

		response, err := useCase.analyzeSnapshotRequest(context.Background(), createLCOMSnapshot(), req)

		require.Error(t, err)
		assert.Nil(t, response)
		assertDomainError(t, err, domain.ErrCodeInvalidInput, "no input paths specified")
		configLoader.AssertExpectations(t)
	})

	t.Run("service without snapshot support returns analysis error", func(t *testing.T) {
		service := &mockLCOMService{}
		configLoader := &mockLCOMConfigurationLoader{}
		configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
		useCase := &LCOMUseCase{service: service, configLoader: configLoader}

		response, err := useCase.analyzeSnapshotRequest(context.Background(), createLCOMSnapshot(), createValidLCOMRequest())

		require.Error(t, err)
		assert.Nil(t, response)
		assertDomainError(t, err, domain.ErrCodeAnalysisError, "LCOM service does not support project snapshots")
		service.AssertExpectations(t)
		configLoader.AssertExpectations(t)
	})

	t.Run("snapshot service error is wrapped", func(t *testing.T) {
		snapshot := createLCOMSnapshot()
		service := &mockSnapshotLCOMService{}
		configLoader := &mockLCOMConfigurationLoader{}
		analysisErr := errors.New("snapshot analysis failed")
		configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
		service.On("AnalyzeSnapshot", mock.Anything, snapshot, mock.AnythingOfType("domain.LCOMRequest")).Return((*domain.LCOMResponse)(nil), analysisErr)
		useCase := &LCOMUseCase{service: service, configLoader: configLoader}

		response, err := useCase.analyzeSnapshotRequest(context.Background(), snapshot, createValidLCOMRequest())

		require.Error(t, err)
		assert.Nil(t, response)
		assertDomainError(t, err, domain.ErrCodeAnalysisError, "LCOM analysis failed")
		assert.ErrorIs(t, err, analysisErr)
		service.AssertExpectations(t)
		configLoader.AssertExpectations(t)
	})

	t.Run("happy path delegates to snapshot service", func(t *testing.T) {
		snapshot := createLCOMSnapshot()
		service := &mockSnapshotLCOMService{}
		configLoader := &mockLCOMConfigurationLoader{}
		response := createMockLCOMResponse()
		configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
		service.On("AnalyzeSnapshot", mock.Anything, snapshot, mock.MatchedBy(func(req domain.LCOMRequest) bool {
			return reflect.DeepEqual(req.Paths, snapshot.Paths())
		})).Return(response, nil)
		useCase := &LCOMUseCase{service: service, configLoader: configLoader}

		result, err := useCase.analyzeSnapshotRequest(context.Background(), snapshot, createValidLCOMRequest())

		require.NoError(t, err)
		assert.Same(t, response, result)
		service.AssertExpectations(t)
		configLoader.AssertExpectations(t)
	})
}

func TestLCOMUseCase_validateRequest(t *testing.T) {
	useCase := &LCOMUseCase{}

	tests := []struct {
		name    string
		request domain.LCOMRequest
		wantErr string
	}{
		{
			name:    "valid request",
			request: createValidLCOMRequest(),
		},
		{
			name: "valid request with output path",
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.OutputWriter = nil
				req.OutputPath = "/tmp/lcom-report.txt"
				return req
			}(),
		},
		{
			name: "empty paths",
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.Paths = nil
				return req
			}(),
			wantErr: "no input paths specified",
		},
		{
			name: "missing output target",
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.OutputWriter = nil
				req.OutputPath = ""
				return req
			}(),
			wantErr: "output writer or output path is required",
		},
		{
			name: "negative min LCOM",
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.MinLCOM = -1
				return req
			}(),
			wantErr: "minimum LCOM cannot be negative",
		},
		{
			name: "negative max LCOM",
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.MaxLCOM = -1
				return req
			}(),
			wantErr: "maximum LCOM cannot be negative",
		},
		{
			name: "min greater than max LCOM",
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.MinLCOM = 11
				req.MaxLCOM = 10
				return req
			}(),
			wantErr: "minimum LCOM cannot be greater than maximum LCOM",
		},
		{
			name: "invalid low threshold",
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.LowThreshold = 0
				return req
			}(),
			wantErr: "low threshold must be positive",
		},
		{
			name: "medium threshold not greater than low",
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.LowThreshold = 5
				req.MediumThreshold = 5
				return req
			}(),
			wantErr: "medium threshold must be greater than low threshold",
		},
		{
			name: "invalid output format",
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.OutputFormat = domain.OutputFormat("invalid")
				return req
			}(),
			wantErr: "unsupported output format: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.validateRequest(tt.request)

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestLCOMUseCase_loadAndMergeConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*mockLCOMConfigurationLoader)
		request    domain.LCOMRequest
		expectErr  bool
		errorMsg   string
		errorIs    error
		want       domain.LCOMRequest
	}{
		{
			name:    "no config loader",
			request: createValidLCOMRequest(),
			want:    createValidLCOMRequest(),
		},
		{
			name: "load default config successfully",
			setupMocks: func(configLoader *mockLCOMConfigurationLoader) {
				defaultConfig := &domain.LCOMRequest{
					MinLCOM:         2,
					MaxLCOM:         15,
					LowThreshold:    3,
					MediumThreshold: 6,
				}
				merged := createValidLCOMRequest()
				merged.MinLCOM = 2
				merged.MaxLCOM = 15
				configLoader.On("LoadDefaultConfig").Return(defaultConfig)
				configLoader.On("MergeConfig", defaultConfig, mock.AnythingOfType("*domain.LCOMRequest")).Return(&merged)
			},
			request: createValidLCOMRequest(),
			want: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.MinLCOM = 2
				req.MaxLCOM = 15
				return req
			}(),
		},
		{
			name: "load specific config file successfully",
			setupMocks: func(configLoader *mockLCOMConfigurationLoader) {
				configReq := &domain.LCOMRequest{
					MinLCOM:         3,
					MaxLCOM:         20,
					LowThreshold:    4,
					MediumThreshold: 8,
				}
				merged := createValidLCOMRequest()
				merged.ConfigPath = "/config.yaml"
				merged.MinLCOM = 3
				merged.MaxLCOM = 20
				configLoader.On("LoadConfig", "/config.yaml").Return(configReq, nil)
				configLoader.On("MergeConfig", configReq, mock.AnythingOfType("*domain.LCOMRequest")).Return(&merged)
			},
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.ConfigPath = "/config.yaml"
				return req
			}(),
			want: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.ConfigPath = "/config.yaml"
				req.MinLCOM = 3
				req.MaxLCOM = 20
				return req
			}(),
		},
		{
			name: "nil default config passes request through",
			setupMocks: func(configLoader *mockLCOMConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.LCOMRequest)(nil))
			},
			request: createValidLCOMRequest(),
			want:    createValidLCOMRequest(),
		},
		{
			name: "config loading error",
			setupMocks: func(configLoader *mockLCOMConfigurationLoader) {
				configLoader.On("LoadConfig", "/invalid.yaml").Return((*domain.LCOMRequest)(nil), errors.New("config not found"))
			},
			request: func() domain.LCOMRequest {
				req := createValidLCOMRequest()
				req.ConfigPath = "/invalid.yaml"
				return req
			}(),
			expectErr: true,
			errorMsg:  "failed to load config from /invalid.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var useCase *LCOMUseCase
			var configLoader *mockLCOMConfigurationLoader

			if strings.Contains(tt.name, "no config loader") {
				useCase = &LCOMUseCase{configLoader: nil}
			} else {
				configLoader = &mockLCOMConfigurationLoader{}
				useCase = &LCOMUseCase{configLoader: configLoader}
				tt.setupMocks(configLoader)
			}

			result, err := useCase.loadAndMergeConfig(tt.request)

			if tt.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				if tt.errorIs != nil {
					assert.ErrorIs(t, err, tt.errorIs)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)
			}

			if configLoader != nil {
				configLoader.AssertExpectations(t)
			}
		})
	}
}

func TestLCOMUseCaseBuilder(t *testing.T) {
	t.Run("chainable setters and build success", func(t *testing.T) {
		service := &mockLCOMService{}
		fileReader := &mockLCOMFileReader{}
		formatter := &mockLCOMOutputFormatter{}
		configLoader := &mockLCOMConfigurationLoader{}
		output := &stubLCOMReportWriter{}
		builder := NewLCOMUseCaseBuilder()

		assert.Same(t, builder, builder.WithService(service))
		assert.Same(t, builder, builder.WithFileReader(fileReader))
		assert.Same(t, builder, builder.WithFormatter(formatter))
		assert.Same(t, builder, builder.WithConfigLoader(configLoader))
		assert.Same(t, builder, builder.WithOutputWriter(output))

		useCase, err := builder.Build()

		require.NoError(t, err)
		require.NotNil(t, useCase)
		assert.Equal(t, service, useCase.service)
		assert.Equal(t, fileReader, useCase.fileReader)
		assert.Equal(t, formatter, useCase.formatter)
		assert.Equal(t, configLoader, useCase.configLoader)
		assert.Equal(t, output, useCase.output)
	})

	t.Run("build succeeds with required dependencies only", func(t *testing.T) {
		useCase, err := NewLCOMUseCaseBuilder().
			WithService(&mockLCOMService{}).
			WithFileReader(&mockLCOMFileReader{}).
			WithFormatter(&mockLCOMOutputFormatter{}).
			Build()

		require.NoError(t, err)
		require.NotNil(t, useCase)
		assert.Nil(t, useCase.configLoader)
		assert.NotNil(t, useCase.output)
	})

	tests := []struct {
		name     string
		builder  *LCOMUseCaseBuilder
		errorMsg string
	}{
		{
			name:     "missing service",
			builder:  NewLCOMUseCaseBuilder(),
			errorMsg: "LCOM service is required",
		},
		{
			name: "missing file reader",
			builder: NewLCOMUseCaseBuilder().
				WithService(&mockLCOMService{}),
			errorMsg: "file reader is required",
		},
		{
			name: "missing formatter",
			builder: NewLCOMUseCaseBuilder().
				WithService(&mockLCOMService{}).
				WithFileReader(&mockLCOMFileReader{}),
			errorMsg: "output formatter is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, err := tt.builder.Build()

			require.Error(t, err)
			assert.Nil(t, useCase)
			assert.Contains(t, err.Error(), tt.errorMsg)
		})
	}
}
