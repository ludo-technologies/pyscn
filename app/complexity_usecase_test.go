package app

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/pyqol/pyqol/domain"
)

// Mock implementations
type mockComplexityService struct {
	mock.Mock
}

func (m *mockComplexityService) Analyze(ctx context.Context, req domain.ComplexityRequest) (*domain.ComplexityResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ComplexityResponse), args.Error(1)
}

func (m *mockComplexityService) AnalyzeFile(ctx context.Context, filePath string, req domain.ComplexityRequest) (*domain.ComplexityResponse, error) {
	args := m.Called(ctx, filePath, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ComplexityResponse), args.Error(1)
}

type mockComplexityOutputFormatter struct {
	mock.Mock
}

func (m *mockComplexityOutputFormatter) Format(response *domain.ComplexityResponse, format domain.OutputFormat) (string, error) {
	args := m.Called(response, format)
	return args.String(0), args.Error(1)
}

func (m *mockComplexityOutputFormatter) Write(response *domain.ComplexityResponse, format domain.OutputFormat, writer io.Writer) error {
	args := m.Called(response, format, writer)
	return args.Error(0)
}

type mockComplexityConfigurationLoader struct {
	mock.Mock
}

func (m *mockComplexityConfigurationLoader) LoadConfig(path string) (*domain.ComplexityRequest, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ComplexityRequest), args.Error(1)
}

func (m *mockComplexityConfigurationLoader) LoadDefaultConfig() *domain.ComplexityRequest {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.ComplexityRequest)
}

func (m *mockComplexityConfigurationLoader) MergeConfig(base *domain.ComplexityRequest, override *domain.ComplexityRequest) *domain.ComplexityRequest {
	args := m.Called(base, override)
	return args.Get(0).(*domain.ComplexityRequest)
}

// Helper functions
func setupComplexityUseCaseMocks() (*ComplexityUseCase, *mockComplexityService, *mockFileReader, *mockComplexityOutputFormatter, *mockComplexityConfigurationLoader, *mockProgressReporter) {
	service := &mockComplexityService{}
	fileReader := &mockFileReader{}
	formatter := &mockComplexityOutputFormatter{}
	configLoader := &mockComplexityConfigurationLoader{}
	progress := &mockProgressReporter{}

	useCase := NewComplexityUseCase(service, fileReader, formatter, configLoader, progress)
	return useCase, service, fileReader, formatter, configLoader, progress
}

func createValidComplexityRequest() domain.ComplexityRequest {
	return domain.ComplexityRequest{
		Paths:           []string{"/test/file.py"},
		OutputWriter:    os.Stdout,
		OutputFormat:    domain.OutputFormatText,
		SortBy:          domain.SortByComplexity,
		MinComplexity:   1,
		MaxComplexity:   10,
		LowThreshold:    3,
		MediumThreshold: 7,
		Recursive:       true,
		IncludePatterns: []string{"*.py"},
		ExcludePatterns: []string{},
	}
}

func createMockComplexityResponse() *domain.ComplexityResponse {
	return &domain.ComplexityResponse{
		Functions: []domain.FunctionComplexity{
			{
				Name:      "test_function",
				FilePath:  "/test/file.py",
				RiskLevel: domain.RiskLevelMedium,
				Metrics: domain.ComplexityMetrics{
					Complexity:        5,
					Nodes:             10,
					Edges:             12,
					IfStatements:      2,
					LoopStatements:    1,
					ExceptionHandlers: 1,
				},
			},
		},
		Summary: domain.ComplexitySummary{
			FilesAnalyzed:           1,
			TotalFunctions:          1,
			AverageComplexity:       5.0,
			MaxComplexity:           5,
			MinComplexity:           5,
			HighRiskFunctions:       0,
			MediumRiskFunctions:     1,
			LowRiskFunctions:        0,
			ComplexityDistribution:  map[string]int{"5": 1},
		},
		GeneratedAt: "2025-01-01T00:00:00Z",
	}
}

func TestComplexityUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockComplexityService, *mockFileReader, *mockComplexityOutputFormatter, *mockComplexityConfigurationLoader, *mockProgressReporter)
		request     domain.ComplexityRequest
		expectError bool
		errorType   string
		errorMsg    string
	}{
		{
			name: "successful execution with valid request",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("LoadDefaultConfig").Return((*domain.ComplexityRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file.py"}, true, []string{"*.py"}, []string{}).
					Return([]string{"/test/file.py"}, nil)
				progress.On("StartProgress", 1)
				progress.On("FinishProgress")
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.ComplexityRequest")).
					Return(createMockComplexityResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			request:     createValidComplexityRequest(),
			expectError: false,
		},
		{
			name: "validation error - empty paths",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.ComplexityRequest{
				Paths:        []string{},
				OutputWriter: os.Stdout,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "no input paths specified",
		},
		{
			name: "validation error - nil output writer",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.ComplexityRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: nil,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "output writer is required",
		},
		{
			name: "validation error - negative min complexity",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.ComplexityRequest{
				Paths:         []string{"/test/file.py"},
				OutputWriter:  os.Stdout,
				MinComplexity: -1,
				LowThreshold:  3,
				MediumThreshold: 7,
				OutputFormat:  domain.OutputFormatText,
				SortBy:        domain.SortByComplexity,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "minimum complexity cannot be negative",
		},
		{
			name: "validation error - invalid output format",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.ComplexityRequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormat("invalid"),
				SortBy:          domain.SortByComplexity,
				LowThreshold:    3,
				MediumThreshold: 7,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "unsupported output format: invalid",
		},
		{
			name: "configuration loading error",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("LoadConfig", "/invalid/config.yaml").
					Return((*domain.ComplexityRequest)(nil), errors.New("config file not found"))
			},
			request: domain.ComplexityRequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByComplexity,
				LowThreshold:    3,
				MediumThreshold: 7,
				ConfigPath:      "/invalid/config.yaml",
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "failed to load configuration",
		},
		{
			name: "file collection error",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("LoadDefaultConfig").Return((*domain.ComplexityRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/invalid/path"}, true, []string{"*.py"}, []string{}).
					Return([]string{}, errors.New("path not found"))
			},
			request: domain.ComplexityRequest{
				Paths:           []string{"/invalid/path"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByComplexity,
				LowThreshold:    3,
				MediumThreshold: 7,
				Recursive:       true,
				IncludePatterns: []string{"*.py"},
				ExcludePatterns: []string{},
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "failed to collect files",
		},
		{
			name: "no files found error",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("LoadDefaultConfig").Return((*domain.ComplexityRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/empty/path"}, true, []string{"*.py"}, []string{}).
					Return([]string{}, nil)
			},
			request: domain.ComplexityRequest{
				Paths:           []string{"/empty/path"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByComplexity,
				LowThreshold:    3,
				MediumThreshold: 7,
				Recursive:       true,
				IncludePatterns: []string{"*.py"},
				ExcludePatterns: []string{},
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "no Python files found in the specified paths",
		},
		{
			name: "analysis service error",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("LoadDefaultConfig").Return((*domain.ComplexityRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file.py"}, true, []string{"*.py"}, []string{}).
					Return([]string{"/test/file.py"}, nil)
				progress.On("StartProgress", 1)
				progress.On("FinishProgress")
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.ComplexityRequest")).
					Return((*domain.ComplexityResponse)(nil), errors.New("parsing failed"))
			},
			request:     createValidComplexityRequest(),
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "complexity analysis failed",
		},
		{
			name: "output formatting error",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("LoadDefaultConfig").Return((*domain.ComplexityRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file.py"}, true, []string{"*.py"}, []string{}).
					Return([]string{"/test/file.py"}, nil)
				progress.On("StartProgress", 1)
				progress.On("FinishProgress")
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.ComplexityRequest")).
					Return(createMockComplexityResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, os.Stdout).
					Return(errors.New("write failed"))
			},
			request:     createValidComplexityRequest(),
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "failed to write output",
		},
		{
			name: "successful execution with config loading",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				configReq := &domain.ComplexityRequest{
					MinComplexity:   2,
					MaxComplexity:   15,
					LowThreshold:    4,
					MediumThreshold: 8,
					OutputFormat:    domain.OutputFormatJSON,
					SortBy:          domain.SortByName,
				}
				mergedReq := createValidComplexityRequest()
				mergedReq.MinComplexity = 2
				
				configLoader.On("LoadConfig", "/config.yaml").Return(configReq, nil)
				configLoader.On("MergeConfig", configReq, mock.AnythingOfType("*domain.ComplexityRequest")).
					Return(&mergedReq)
				fileReader.On("CollectPythonFiles", []string{"/test/file.py"}, true, []string{"*.py"}, []string{}).
					Return([]string{"/test/file.py"}, nil)
				progress.On("StartProgress", 1)
				progress.On("FinishProgress")
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.ComplexityRequest")).
					Return(createMockComplexityResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			request: func() domain.ComplexityRequest {
				req := createValidComplexityRequest()
				req.ConfigPath = "/config.yaml"
				return req
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader, progress := setupComplexityUseCaseMocks()
			
			tt.setupMocks(service, fileReader, formatter, configLoader, progress)

			err := useCase.Execute(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorType != "" {
					assert.IsType(t, domain.DomainError{}, err)
				}
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all mock expectations
			service.AssertExpectations(t)
			fileReader.AssertExpectations(t)
			formatter.AssertExpectations(t)
			configLoader.AssertExpectations(t)
			progress.AssertExpectations(t)
		})
	}
}

func TestComplexityUseCase_AnalyzeAndReturn(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockComplexityService, *mockFileReader, *mockComplexityOutputFormatter, *mockComplexityConfigurationLoader, *mockProgressReporter)
		request     domain.ComplexityRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful analysis without formatting",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("LoadDefaultConfig").Return((*domain.ComplexityRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file.py"}, true, []string{"*.py"}, []string{}).
					Return([]string{"/test/file.py"}, nil)
				progress.On("StartProgress", 1)
				progress.On("FinishProgress")
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.ComplexityRequest")).
					Return(createMockComplexityResponse(), nil)
			},
			request:     createValidComplexityRequest(),
			expectError: false,
		},
		{
			name: "validation error in analyze and return",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.ComplexityRequest{
				Paths:        []string{},
				OutputWriter: os.Stdout,
			},
			expectError: true,
			errorMsg:    "no input paths specified",
		},
		{
			name: "analysis error in analyze and return",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("LoadDefaultConfig").Return((*domain.ComplexityRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file.py"}, true, []string{"*.py"}, []string{}).
					Return([]string{"/test/file.py"}, nil)
				progress.On("StartProgress", 1)
				progress.On("FinishProgress")
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.ComplexityRequest")).
					Return((*domain.ComplexityResponse)(nil), errors.New("analysis failed"))
			},
			request:     createValidComplexityRequest(),
			expectError: true,
			errorMsg:    "complexity analysis failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader, progress := setupComplexityUseCaseMocks()
			
			tt.setupMocks(service, fileReader, formatter, configLoader, progress)

			response, err := useCase.AnalyzeAndReturn(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, 1, len(response.Functions))
				assert.Equal(t, "test_function", response.Functions[0].Name)
			}

			// Verify all mock expectations
			service.AssertExpectations(t)
			fileReader.AssertExpectations(t)
			formatter.AssertExpectations(t)
			configLoader.AssertExpectations(t)
			progress.AssertExpectations(t)
		})
	}
}

func TestComplexityUseCase_validateRequest(t *testing.T) {
	useCase := &ComplexityUseCase{}

	tests := []struct {
		name    string
		request domain.ComplexityRequest
		wantErr string
	}{
		{
			name: "valid request",
			request: domain.ComplexityRequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByComplexity,
				MinComplexity:   0,
				MaxComplexity:   10,
				LowThreshold:    3,
				MediumThreshold: 7,
			},
			wantErr: "",
		},
		{
			name: "empty paths",
			request: domain.ComplexityRequest{
				OutputWriter: os.Stdout,
			},
			wantErr: "no input paths specified",
		},
		{
			name: "nil output writer",
			request: domain.ComplexityRequest{
				Paths: []string{"/test/file.py"},
			},
			wantErr: "output writer is required",
		},
		{
			name: "negative min complexity",
			request: domain.ComplexityRequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				MinComplexity:   -1,
				LowThreshold:    3,
				MediumThreshold: 7,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByComplexity,
			},
			wantErr: "minimum complexity cannot be negative",
		},
		{
			name: "min greater than max complexity",
			request: domain.ComplexityRequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				MinComplexity:   10,
				MaxComplexity:   5,
				LowThreshold:    3,
				MediumThreshold: 7,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByComplexity,
			},
			wantErr: "minimum complexity cannot be greater than maximum complexity",
		},
		{
			name: "invalid low threshold",
			request: domain.ComplexityRequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				LowThreshold:    0,
				MediumThreshold: 7,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByComplexity,
			},
			wantErr: "low threshold must be positive",
		},
		{
			name: "medium threshold not greater than low",
			request: domain.ComplexityRequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				LowThreshold:    5,
				MediumThreshold: 3,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByComplexity,
			},
			wantErr: "medium threshold must be greater than low threshold",
		},
		{
			name: "invalid output format",
			request: domain.ComplexityRequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				LowThreshold:    3,
				MediumThreshold: 7,
				OutputFormat:    domain.OutputFormat("invalid"),
				SortBy:          domain.SortByComplexity,
			},
			wantErr: "unsupported output format: invalid",
		},
		{
			name: "invalid sort criteria",
			request: domain.ComplexityRequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				LowThreshold:    3,
				MediumThreshold: 7,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortCriteria("invalid"),
			},
			wantErr: "unsupported sort criteria: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := useCase.validateRequest(tt.request)
			
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestComplexityUseCase_AnalyzeFile(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		setupMocks  func(*mockComplexityService, *mockFileReader, *mockComplexityOutputFormatter, *mockComplexityConfigurationLoader, *mockProgressReporter)
		expectError bool
		errorMsg    string
	}{
		{
			name:     "successful file analysis",
			filePath: "/test/file.py",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				fileReader.On("IsValidPythonFile", "/test/file.py").Return(true)
				configLoader.On("LoadDefaultConfig").Return((*domain.ComplexityRequest)(nil))
				service.On("AnalyzeFile", mock.Anything, "/test/file.py", mock.AnythingOfType("domain.ComplexityRequest")).
					Return(createMockComplexityResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			expectError: false,
		},
		{
			name:     "invalid python file",
			filePath: "/test/file.txt",
			setupMocks: func(service *mockComplexityService, fileReader *mockFileReader, formatter *mockComplexityOutputFormatter, configLoader *mockComplexityConfigurationLoader, progress *mockProgressReporter) {
				fileReader.On("IsValidPythonFile", "/test/file.txt").Return(false)
			},
			expectError: true,
			errorMsg:    "not a valid Python file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader, progress := setupComplexityUseCaseMocks()
			
			tt.setupMocks(service, fileReader, formatter, configLoader, progress)

			req := createValidComplexityRequest()
			err := useCase.AnalyzeFile(context.Background(), tt.filePath, req)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify all mock expectations
			service.AssertExpectations(t)
			fileReader.AssertExpectations(t)
			formatter.AssertExpectations(t)
			configLoader.AssertExpectations(t)
			progress.AssertExpectations(t)
		})
	}
}

func TestComplexityUseCase_loadAndMergeConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*mockComplexityConfigurationLoader)
		request    domain.ComplexityRequest
		expectErr  bool
		errorMsg   string
	}{
		{
			name: "no config loader",
			setupMocks: func(configLoader *mockComplexityConfigurationLoader) {
				// No setup needed - configLoader will be nil
			},
			request:   createValidComplexityRequest(),
			expectErr: false,
		},
		{
			name: "load default config successfully",
			setupMocks: func(configLoader *mockComplexityConfigurationLoader) {
				defaultConfig := &domain.ComplexityRequest{
					MinComplexity: 2,
					MaxComplexity: 15,
				}
				configLoader.On("LoadDefaultConfig").Return(defaultConfig)
				configLoader.On("MergeConfig", defaultConfig, mock.AnythingOfType("*domain.ComplexityRequest")).
					Return(&domain.ComplexityRequest{MinComplexity: 2})
			},
			request:   createValidComplexityRequest(),
			expectErr: false,
		},
		{
			name: "load specific config file successfully",
			setupMocks: func(configLoader *mockComplexityConfigurationLoader) {
				configReq := &domain.ComplexityRequest{
					MinComplexity: 3,
					MaxComplexity: 20,
				}
				configLoader.On("LoadConfig", "/config.yaml").Return(configReq, nil)
				configLoader.On("MergeConfig", configReq, mock.AnythingOfType("*domain.ComplexityRequest")).
					Return(&domain.ComplexityRequest{MinComplexity: 3})
			},
			request: func() domain.ComplexityRequest {
				req := createValidComplexityRequest()
				req.ConfigPath = "/config.yaml"
				return req
			}(),
			expectErr: false,
		},
		{
			name: "config loading error",
			setupMocks: func(configLoader *mockComplexityConfigurationLoader) {
				configLoader.On("LoadConfig", "/invalid.yaml").Return((*domain.ComplexityRequest)(nil), errors.New("config not found"))
			},
			request: func() domain.ComplexityRequest {
				req := createValidComplexityRequest()
				req.ConfigPath = "/invalid.yaml"
				return req
			}(),
			expectErr: true,
			errorMsg:  "failed to load config from /invalid.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var useCase *ComplexityUseCase
			var configLoader *mockComplexityConfigurationLoader

			if strings.Contains(tt.name, "no config loader") {
				useCase = &ComplexityUseCase{configLoader: nil}
			} else {
				configLoader = &mockComplexityConfigurationLoader{}
				useCase = &ComplexityUseCase{configLoader: configLoader}
				tt.setupMocks(configLoader)
			}

			result, err := useCase.loadAndMergeConfig(tt.request)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			if configLoader != nil {
				configLoader.AssertExpectations(t)
			}
		})
	}
}