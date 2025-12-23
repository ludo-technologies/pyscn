package app

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type mockCBOService struct {
	mock.Mock
}

func (m *mockCBOService) Analyze(ctx context.Context, req domain.CBORequest) (*domain.CBOResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CBOResponse), args.Error(1)
}

func (m *mockCBOService) AnalyzeFile(ctx context.Context, filePath string, req domain.CBORequest) (*domain.CBOResponse, error) {
	args := m.Called(ctx, filePath, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CBOResponse), args.Error(1)
}

type mockCBOOutputFormatter struct {
	mock.Mock
}

func (m *mockCBOOutputFormatter) Format(response *domain.CBOResponse, format domain.OutputFormat) (string, error) {
	args := m.Called(response, format)
	return args.String(0), args.Error(1)
}

func (m *mockCBOOutputFormatter) Write(response *domain.CBOResponse, format domain.OutputFormat, writer io.Writer) error {
	args := m.Called(response, format, writer)
	return args.Error(0)
}

type mockCBOConfigurationLoader struct {
	mock.Mock
}

func (m *mockCBOConfigurationLoader) LoadConfig(path string) (*domain.CBORequest, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CBORequest), args.Error(1)
}

func (m *mockCBOConfigurationLoader) LoadDefaultConfig() *domain.CBORequest {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.CBORequest)
}

func (m *mockCBOConfigurationLoader) MergeConfig(base *domain.CBORequest, override *domain.CBORequest) *domain.CBORequest {
	args := m.Called(base, override)
	return args.Get(0).(*domain.CBORequest)
}

// Helper functions
func setupCBOUseCaseMocks() (*CBOUseCase, *mockCBOService, *mockFileReader, *mockCBOOutputFormatter, *mockCBOConfigurationLoader) {
	service := &mockCBOService{}
	fileReader := &mockFileReader{}
	formatter := &mockCBOOutputFormatter{}
	configLoader := &mockCBOConfigurationLoader{}

	useCase := NewCBOUseCase(service, fileReader, formatter, configLoader)
	return useCase, service, fileReader, formatter, configLoader
}

func createValidCBORequest() domain.CBORequest {
	return domain.CBORequest{
		Paths:           []string{"/test/file.py"},
		OutputWriter:    os.Stdout,
		OutputFormat:    domain.OutputFormatText,
		SortBy:          domain.SortByCoupling,
		MinCBO:          0,
		MaxCBO:          10,
		LowThreshold:    3,
		MediumThreshold: 7,
		Recursive:       domain.BoolPtr(true),
		IncludePatterns: []string{"**/*.py"},
		ExcludePatterns: []string{},
	}
}

func createMockCBOResponse() *domain.CBOResponse {
	return &domain.CBOResponse{
		Classes: []domain.ClassCoupling{
			{
				Name:      "TestClass",
				FilePath:  "/test/file.py",
				StartLine: 1,
				EndLine:   20,
				RiskLevel: domain.RiskLevelMedium,
				Metrics: domain.CBOMetrics{
					CouplingCount:               5,
					InheritanceDependencies:     1,
					TypeHintDependencies:        2,
					InstantiationDependencies:   1,
					AttributeAccessDependencies: 1,
					ImportDependencies:          0,
					DependentClasses:            []string{"BaseClass", "Helper", "Util", "Config", "Logger"},
				},
			},
		},
		Summary: domain.CBOSummary{
			TotalClasses:      1,
			AverageCBO:        5.0,
			MaxCBO:            5,
			MinCBO:            5,
			ClassesAnalyzed:   1,
			FilesAnalyzed:     1,
			LowRiskClasses:    0,
			MediumRiskClasses: 1,
			HighRiskClasses:   0,
			CBODistribution:   map[string]int{"5": 1},
		},
		GeneratedAt: "2025-01-01T00:00:00Z",
	}
}

func TestCBOUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockCBOService, *mockFileReader, *mockCBOOutputFormatter, *mockCBOConfigurationLoader)
		request     domain.CBORequest
		expectError bool
		errorType   string
		errorMsg    string
	}{
		{
			name: "successful execution with valid request",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.CBORequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.CBORequest")).
					Return(createMockCBOResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			request:     createValidCBORequest(),
			expectError: false,
		},
		{
			name: "validation error - empty paths",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.CBORequest{
				Paths:        []string{},
				OutputWriter: os.Stdout,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "no input paths specified",
		},
		{
			name: "validation error - nil output writer",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.CBORequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: nil,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "output writer or output path is required",
		},
		{
			name: "validation error - negative min CBO",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.CBORequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				MinCBO:          -1,
				LowThreshold:    3,
				MediumThreshold: 7,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByCoupling,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "minimum CBO cannot be negative",
		},
		{
			name: "validation error - invalid output format",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.CBORequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormat("invalid"),
				SortBy:          domain.SortByCoupling,
				LowThreshold:    3,
				MediumThreshold: 7,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "unsupported output format: invalid",
		},
		{
			name: "configuration loading error",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				configLoader.On("LoadConfig", "/invalid/config.yaml").
					Return((*domain.CBORequest)(nil), errors.New("config file not found"))
			},
			request: domain.CBORequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByCoupling,
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
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.CBORequest)(nil))
				fileReader.On("FileExists", "/invalid/path").Return(false, nil)
				fileReader.On("CollectPythonFiles", []string{"/invalid/path"}, true, []string{"**/*.py"}, []string{}).
					Return([]string{}, errors.New("path not found"))
			},
			request: domain.CBORequest{
				Paths:           []string{"/invalid/path"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByCoupling,
				LowThreshold:    3,
				MediumThreshold: 7,
				Recursive:       domain.BoolPtr(true),
				IncludePatterns: []string{"**/*.py"},
				ExcludePatterns: []string{},
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "failed to collect files",
		},
		{
			name: "no files found error",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.CBORequest)(nil))
				fileReader.On("FileExists", "/empty/path").Return(false, nil)
				fileReader.On("CollectPythonFiles", []string{"/empty/path"}, true, []string{"**/*.py"}, []string{}).
					Return([]string{}, nil)
			},
			request: domain.CBORequest{
				Paths:           []string{"/empty/path"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByCoupling,
				LowThreshold:    3,
				MediumThreshold: 7,
				Recursive:       domain.BoolPtr(true),
				IncludePatterns: []string{"**/*.py"},
				ExcludePatterns: []string{},
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "no Python files found in the specified paths",
		},
		{
			name: "analysis service error",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.CBORequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.CBORequest")).
					Return((*domain.CBOResponse)(nil), errors.New("parsing failed"))
			},
			request:     createValidCBORequest(),
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "CBO analysis failed",
		},
		{
			name: "output formatting error",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.CBORequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.CBORequest")).
					Return(createMockCBOResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, os.Stdout).
					Return(errors.New("write failed"))
			},
			request:     createValidCBORequest(),
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "failed to write output",
		},
		{
			name: "successful execution with config loading",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				configReq := &domain.CBORequest{
					MinCBO:          2,
					MaxCBO:          15,
					LowThreshold:    4,
					MediumThreshold: 8,
					OutputFormat:    domain.OutputFormatJSON,
					SortBy:          domain.SortByName,
				}
				mergedReq := createValidCBORequest()
				mergedReq.MinCBO = 2

				configLoader.On("LoadConfig", "/config.yaml").Return(configReq, nil)
				configLoader.On("MergeConfig", configReq, mock.AnythingOfType("*domain.CBORequest")).
					Return(&mergedReq)
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.CBORequest")).
					Return(createMockCBOResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			request: func() domain.CBORequest {
				req := createValidCBORequest()
				req.ConfigPath = "/config.yaml"
				return req
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader := setupCBOUseCaseMocks()

			tt.setupMocks(service, fileReader, formatter, configLoader)

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
		})
	}
}

func TestCBOUseCase_AnalyzeAndReturn(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockCBOService, *mockFileReader, *mockCBOOutputFormatter, *mockCBOConfigurationLoader)
		request     domain.CBORequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful analysis without formatting",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.CBORequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.CBORequest")).
					Return(createMockCBOResponse(), nil)
			},
			request:     createValidCBORequest(),
			expectError: false,
		},
		{
			name: "validation error in analyze and return",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.CBORequest{
				Paths:        []string{},
				OutputWriter: os.Stdout,
			},
			expectError: true,
			errorMsg:    "no input paths specified",
		},
		{
			name: "analysis error in analyze and return",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.CBORequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.CBORequest")).
					Return((*domain.CBOResponse)(nil), errors.New("analysis failed"))
			},
			request:     createValidCBORequest(),
			expectError: true,
			errorMsg:    "CBO analysis failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader := setupCBOUseCaseMocks()

			tt.setupMocks(service, fileReader, formatter, configLoader)

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
				assert.Equal(t, 1, len(response.Classes))
				assert.Equal(t, "TestClass", response.Classes[0].Name)
			}

			// Verify all mock expectations
			service.AssertExpectations(t)
			fileReader.AssertExpectations(t)
			formatter.AssertExpectations(t)
			configLoader.AssertExpectations(t)
		})
	}
}

func TestCBOUseCase_AnalyzeFile(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		setupMocks  func(*mockCBOService, *mockFileReader, *mockCBOOutputFormatter, *mockCBOConfigurationLoader)
		expectError bool
		errorMsg    string
	}{
		{
			name:     "successful file analysis",
			filePath: "/test/file.py",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				fileReader.On("IsValidPythonFile", "/test/file.py").Return(true)
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				configLoader.On("LoadDefaultConfig").Return((*domain.CBORequest)(nil))
				service.On("AnalyzeFile", mock.Anything, "/test/file.py", mock.AnythingOfType("domain.CBORequest")).
					Return(createMockCBOResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			expectError: false,
		},
		{
			name:     "invalid python file",
			filePath: "/test/file.txt",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				fileReader.On("IsValidPythonFile", "/test/file.txt").Return(false)
			},
			expectError: true,
			errorMsg:    "not a valid Python file",
		},
		{
			name:     "file does not exist",
			filePath: "/test/nonexistent.py",
			setupMocks: func(service *mockCBOService, fileReader *mockFileReader, formatter *mockCBOOutputFormatter, configLoader *mockCBOConfigurationLoader) {
				fileReader.On("IsValidPythonFile", "/test/nonexistent.py").Return(true)
				fileReader.On("FileExists", "/test/nonexistent.py").Return(false, nil)
			},
			expectError: true,
			errorMsg:    "file does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader := setupCBOUseCaseMocks()

			tt.setupMocks(service, fileReader, formatter, configLoader)

			req := createValidCBORequest()
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
		})
	}
}

func TestCBOUseCase_validateRequest(t *testing.T) {
	useCase := &CBOUseCase{}

	tests := []struct {
		name    string
		request domain.CBORequest
		wantErr string
	}{
		{
			name: "valid request",
			request: domain.CBORequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByCoupling,
				MinCBO:          0,
				MaxCBO:          10,
				LowThreshold:    3,
				MediumThreshold: 7,
			},
			wantErr: "",
		},
		{
			name: "empty paths",
			request: domain.CBORequest{
				OutputWriter: os.Stdout,
			},
			wantErr: "no input paths specified",
		},
		{
			name: "nil output writer",
			request: domain.CBORequest{
				Paths: []string{"/test/file.py"},
			},
			wantErr: "output writer or output path is required",
		},
		{
			name: "negative min CBO",
			request: domain.CBORequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				MinCBO:          -1,
				LowThreshold:    3,
				MediumThreshold: 7,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByCoupling,
			},
			wantErr: "minimum CBO cannot be negative",
		},
		{
			name: "negative max CBO",
			request: domain.CBORequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				MaxCBO:          -1,
				LowThreshold:    3,
				MediumThreshold: 7,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByCoupling,
			},
			wantErr: "maximum CBO cannot be negative",
		},
		{
			name: "min greater than max CBO",
			request: domain.CBORequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				MinCBO:          10,
				MaxCBO:          5,
				LowThreshold:    3,
				MediumThreshold: 7,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByCoupling,
			},
			wantErr: "minimum CBO cannot be greater than maximum CBO",
		},
		{
			name: "invalid low threshold",
			request: domain.CBORequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				LowThreshold:    0,
				MediumThreshold: 7,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByCoupling,
			},
			wantErr: "low threshold must be positive",
		},
		{
			name: "medium threshold not greater than low",
			request: domain.CBORequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				LowThreshold:    5,
				MediumThreshold: 3,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.SortByCoupling,
			},
			wantErr: "medium threshold must be greater than low threshold",
		},
		{
			name: "invalid output format",
			request: domain.CBORequest{
				Paths:           []string{"/test/file.py"},
				OutputWriter:    os.Stdout,
				LowThreshold:    3,
				MediumThreshold: 7,
				OutputFormat:    domain.OutputFormat("invalid"),
				SortBy:          domain.SortByCoupling,
			},
			wantErr: "unsupported output format: invalid",
		},
		{
			name: "invalid sort criteria",
			request: domain.CBORequest{
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

func TestCBOUseCase_loadAndMergeConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*mockCBOConfigurationLoader)
		request    domain.CBORequest
		expectErr  bool
		errorMsg   string
	}{
		{
			name: "no config loader",
			setupMocks: func(configLoader *mockCBOConfigurationLoader) {
				// No setup needed - configLoader will be nil
			},
			request:   createValidCBORequest(),
			expectErr: false,
		},
		{
			name: "load default config successfully",
			setupMocks: func(configLoader *mockCBOConfigurationLoader) {
				defaultConfig := &domain.CBORequest{
					MinCBO: 2,
					MaxCBO: 15,
				}
				configLoader.On("LoadDefaultConfig").Return(defaultConfig)
				configLoader.On("MergeConfig", defaultConfig, mock.AnythingOfType("*domain.CBORequest")).
					Return(&domain.CBORequest{MinCBO: 2})
			},
			request:   createValidCBORequest(),
			expectErr: false,
		},
		{
			name: "load specific config file successfully",
			setupMocks: func(configLoader *mockCBOConfigurationLoader) {
				configReq := &domain.CBORequest{
					MinCBO: 3,
					MaxCBO: 20,
				}
				configLoader.On("LoadConfig", "/config.yaml").Return(configReq, nil)
				configLoader.On("MergeConfig", configReq, mock.AnythingOfType("*domain.CBORequest")).
					Return(&domain.CBORequest{MinCBO: 3})
			},
			request: func() domain.CBORequest {
				req := createValidCBORequest()
				req.ConfigPath = "/config.yaml"
				return req
			}(),
			expectErr: false,
		},
		{
			name: "config loading error",
			setupMocks: func(configLoader *mockCBOConfigurationLoader) {
				configLoader.On("LoadConfig", "/invalid.yaml").Return((*domain.CBORequest)(nil), errors.New("config not found"))
			},
			request: func() domain.CBORequest {
				req := createValidCBORequest()
				req.ConfigPath = "/invalid.yaml"
				return req
			}(),
			expectErr: true,
			errorMsg:  "failed to load config from /invalid.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var useCase *CBOUseCase
			var configLoader *mockCBOConfigurationLoader

			if strings.Contains(tt.name, "no config loader") {
				useCase = &CBOUseCase{configLoader: nil}
			} else {
				configLoader = &mockCBOConfigurationLoader{}
				useCase = &CBOUseCase{configLoader: configLoader}
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
