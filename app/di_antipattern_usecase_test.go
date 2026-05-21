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

type mockDIAntipatternService struct {
	mock.Mock
}

func (m *mockDIAntipatternService) Analyze(ctx context.Context, req domain.DIAntipatternRequest) (*domain.DIAntipatternResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DIAntipatternResponse), args.Error(1)
}

func (m *mockDIAntipatternService) AnalyzeFile(ctx context.Context, filePath string, req domain.DIAntipatternRequest) (*domain.DIAntipatternResponse, error) {
	args := m.Called(ctx, filePath, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DIAntipatternResponse), args.Error(1)
}

type mockDIAntipatternOutputFormatter struct {
	mock.Mock
}

func (m *mockDIAntipatternOutputFormatter) Format(response *domain.DIAntipatternResponse, format domain.OutputFormat) (string, error) {
	args := m.Called(response, format)
	return args.String(0), args.Error(1)
}

func (m *mockDIAntipatternOutputFormatter) Write(response *domain.DIAntipatternResponse, format domain.OutputFormat, writer io.Writer) error {
	args := m.Called(response, format, writer)
	return args.Error(0)
}

type mockDIAntipatternConfigurationLoader struct {
	mock.Mock
}

func (m *mockDIAntipatternConfigurationLoader) LoadConfig(path string) (*domain.DIAntipatternRequest, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DIAntipatternRequest), args.Error(1)
}

func (m *mockDIAntipatternConfigurationLoader) LoadDefaultConfig() *domain.DIAntipatternRequest {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.DIAntipatternRequest)
}

func (m *mockDIAntipatternConfigurationLoader) MergeConfig(base *domain.DIAntipatternRequest, override *domain.DIAntipatternRequest) *domain.DIAntipatternRequest {
	args := m.Called(base, override)
	return args.Get(0).(*domain.DIAntipatternRequest)
}

type mockReportWriter struct {
	mock.Mock
}

func (m *mockReportWriter) Write(writer io.Writer, outputPath string, format domain.OutputFormat, noOpen bool, writeFunc func(io.Writer) error) error {
	args := m.Called(writer, outputPath, format, noOpen, writeFunc)
	return args.Error(0)
}

// Helper functions

func setupDIAntipatternUseCaseMocks() (*DIAntipatternUseCase, *mockDIAntipatternService, *mockFileReader, *mockDIAntipatternOutputFormatter, *mockDIAntipatternConfigurationLoader) {
	service := &mockDIAntipatternService{}
	fileReader := &mockFileReader{}
	formatter := &mockDIAntipatternOutputFormatter{}
	configLoader := &mockDIAntipatternConfigurationLoader{}

	useCase := NewDIAntipatternUseCase(service, fileReader, formatter, configLoader)
	return useCase, service, fileReader, formatter, configLoader
}

func createValidDIAntipatternRequest() domain.DIAntipatternRequest {
	return domain.DIAntipatternRequest{
		Paths:                     []string{"/test/file.py"},
		OutputWriter:              os.Stdout,
		OutputFormat:              domain.OutputFormatText,
		ConstructorParamThreshold: 5,
		MinSeverity:               domain.DIAntipatternSeverityWarning,
		Recursive:                 domain.BoolPtr(true),
		IncludePatterns:           []string{"**/*.py"},
		ExcludePatterns:           []string{},
	}
}

func createMockDIAntipatternResponse() *domain.DIAntipatternResponse {
	return &domain.DIAntipatternResponse{
		Findings: []domain.DIAntipatternFinding{
			{
				Type:        domain.DIAntipatternConstructorOverInjection,
				Severity:    domain.DIAntipatternSeverityWarning,
				ClassName:   "TestService",
				MethodName:  "__init__",
				Description: "Constructor has too many parameters",
				Suggestion:  "Consider splitting responsibilities or using a factory",
			},
		},
		Summary: domain.DIAntipatternSummary{
			TotalFindings: 1,
			ByType: map[domain.DIAntipatternType]int{
				domain.DIAntipatternConstructorOverInjection: 1,
			},
			BySeverity: map[domain.DIAntipatternSeverity]int{
				domain.DIAntipatternSeverityWarning: 1,
			},
			FilesAnalyzed:   1,
			AffectedClasses: 1,
		},
		GeneratedAt: "2025-01-01T00:00:00Z",
	}
}

func TestDIAntipatternUseCase_validateRequest(t *testing.T) {
	useCase := &DIAntipatternUseCase{}

	tests := []struct {
		name    string
		request domain.DIAntipatternRequest
		wantErr string
	}{
		{
			name: "valid request with output writer",
			request: domain.DIAntipatternRequest{
				Paths:                     []string{"/test/file.py"},
				OutputWriter:              os.Stdout,
				ConstructorParamThreshold: 5,
			},
			wantErr: "",
		},
		{
			name: "valid request with output path",
			request: domain.DIAntipatternRequest{
				Paths:                     []string{"/test/file.py"},
				OutputPath:                "/output/report.txt",
				ConstructorParamThreshold: 0,
			},
			wantErr: "",
		},
		{
			name: "empty paths",
			request: domain.DIAntipatternRequest{
				OutputWriter: os.Stdout,
			},
			wantErr: "no input paths specified",
		},
		{
			name: "nil output writer and no output path",
			request: domain.DIAntipatternRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: nil,
				OutputPath:   "",
			},
			wantErr: "output writer or output path is required",
		},
		{
			name: "negative constructor param threshold",
			request: domain.DIAntipatternRequest{
				Paths:                     []string{"/test/file.py"},
				OutputWriter:              os.Stdout,
				ConstructorParamThreshold: -1,
			},
			wantErr: "constructor parameter threshold cannot be negative",
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

func TestDIAntipatternUseCase_loadAndMergeConfig(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*mockDIAntipatternConfigurationLoader)
		request   domain.DIAntipatternRequest
		expectErr bool
		errorMsg  string
	}{
		{
			name:      "no config loader",
			setupMock: func(configLoader *mockDIAntipatternConfigurationLoader) {},
			request:   createValidDIAntipatternRequest(),
			expectErr: false,
		},
		{
			name: "load default config - nil result",
			setupMock: func(configLoader *mockDIAntipatternConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DIAntipatternRequest)(nil))
			},
			request:   createValidDIAntipatternRequest(),
			expectErr: false,
		},
		{
			name: "load default config - non-nil result",
			setupMock: func(configLoader *mockDIAntipatternConfigurationLoader) {
				defaultConfig := &domain.DIAntipatternRequest{
					ConstructorParamThreshold: 3,
					MinSeverity:               domain.DIAntipatternSeverityError,
				}
				merged := createValidDIAntipatternRequest()
				merged.ConstructorParamThreshold = 3

				configLoader.On("LoadDefaultConfig").Return(defaultConfig)
				configLoader.On("MergeConfig", defaultConfig, mock.AnythingOfType("*domain.DIAntipatternRequest")).
					Return(&merged)
			},
			request:   createValidDIAntipatternRequest(),
			expectErr: false,
		},
		{
			name: "load specific config file successfully",
			setupMock: func(configLoader *mockDIAntipatternConfigurationLoader) {
				configReq := &domain.DIAntipatternRequest{
					ConstructorParamThreshold: 7,
				}
				merged := createValidDIAntipatternRequest()
				merged.ConstructorParamThreshold = 7

				configLoader.On("LoadConfig", "/config.yaml").Return(configReq, nil)
				configLoader.On("MergeConfig", configReq, mock.AnythingOfType("*domain.DIAntipatternRequest")).
					Return(&merged)
			},
			request: func() domain.DIAntipatternRequest {
				req := createValidDIAntipatternRequest()
				req.ConfigPath = "/config.yaml"
				return req
			}(),
			expectErr: false,
		},
		{
			name: "config loading error",
			setupMock: func(configLoader *mockDIAntipatternConfigurationLoader) {
				configLoader.On("LoadConfig", "/bad.yaml").
					Return((*domain.DIAntipatternRequest)(nil), errors.New("config not found"))
			},
			request: func() domain.DIAntipatternRequest {
				req := createValidDIAntipatternRequest()
				req.ConfigPath = "/bad.yaml"
				return req
			}(),
			expectErr: true,
			errorMsg:  "failed to load config from /bad.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var useCase *DIAntipatternUseCase
			var configLoader *mockDIAntipatternConfigurationLoader

			if strings.Contains(tt.name, "no config loader") {
				useCase = &DIAntipatternUseCase{configLoader: nil}
			} else {
				configLoader = &mockDIAntipatternConfigurationLoader{}
				useCase = &DIAntipatternUseCase{configLoader: configLoader}
				tt.setupMock(configLoader)
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

func TestDIAntipatternUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockDIAntipatternService, *mockFileReader, *mockDIAntipatternOutputFormatter, *mockDIAntipatternConfigurationLoader)
		request     domain.DIAntipatternRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful execution",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DIAntipatternRequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DIAntipatternRequest")).
					Return(createMockDIAntipatternResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			request:     createValidDIAntipatternRequest(),
			expectError: false,
		},
		{
			name: "validation error - empty paths",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
			},
			request: domain.DIAntipatternRequest{
				Paths:        []string{},
				OutputWriter: os.Stdout,
			},
			expectError: true,
			errorMsg:    "no input paths specified",
		},
		{
			name: "validation error - nil output writer and no output path",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
			},
			request: domain.DIAntipatternRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: nil,
				OutputPath:   "",
			},
			expectError: true,
			errorMsg:    "output writer or output path is required",
		},
		{
			name: "validation error - negative constructor threshold",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
			},
			request: domain.DIAntipatternRequest{
				Paths:                     []string{"/test/file.py"},
				OutputWriter:              os.Stdout,
				ConstructorParamThreshold: -1,
			},
			expectError: true,
			errorMsg:    "constructor parameter threshold cannot be negative",
		},
		{
			name: "configuration loading error",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
				configLoader.On("LoadConfig", "/invalid/config.yaml").
					Return((*domain.DIAntipatternRequest)(nil), errors.New("config file not found"))
			},
			request: func() domain.DIAntipatternRequest {
				req := createValidDIAntipatternRequest()
				req.ConfigPath = "/invalid/config.yaml"
				return req
			}(),
			expectError: true,
			errorMsg:    "failed to load configuration",
		},
		{
			name: "file collection error",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DIAntipatternRequest)(nil))
				fileReader.On("FileExists", "/invalid/path").Return(false, nil)
				fileReader.On("CollectPythonFiles", []string{"/invalid/path"}, true, []string{"**/*.py"}, []string{}).
					Return([]string{}, errors.New("path not found"))
			},
			request: domain.DIAntipatternRequest{
				Paths:           []string{"/invalid/path"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormatText,
				Recursive:       domain.BoolPtr(true),
				IncludePatterns: []string{"**/*.py"},
				ExcludePatterns: []string{},
			},
			expectError: true,
			errorMsg:    "failed to collect files",
		},
		{
			name: "no files found",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DIAntipatternRequest)(nil))
				fileReader.On("FileExists", "/empty/path").Return(false, nil)
				fileReader.On("CollectPythonFiles", []string{"/empty/path"}, true, []string{"**/*.py"}, []string{}).
					Return([]string{}, nil)
			},
			request: domain.DIAntipatternRequest{
				Paths:           []string{"/empty/path"},
				OutputWriter:    os.Stdout,
				OutputFormat:    domain.OutputFormatText,
				Recursive:       domain.BoolPtr(true),
				IncludePatterns: []string{"**/*.py"},
				ExcludePatterns: []string{},
			},
			expectError: true,
			errorMsg:    "no Python files found in the specified paths",
		},
		{
			name: "analysis service error",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DIAntipatternRequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DIAntipatternRequest")).
					Return((*domain.DIAntipatternResponse)(nil), errors.New("parsing failed"))
			},
			request:     createValidDIAntipatternRequest(),
			expectError: true,
			errorMsg:    "DI anti-pattern analysis failed",
		},
		{
			name: "output formatting error",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DIAntipatternRequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DIAntipatternRequest")).
					Return(createMockDIAntipatternResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, os.Stdout).
					Return(errors.New("write failed"))
			},
			request:     createValidDIAntipatternRequest(),
			expectError: true,
			errorMsg:    "failed to write output",
		},
		{
			name: "successful execution with config loading",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
				configReq := &domain.DIAntipatternRequest{
					ConstructorParamThreshold: 3,
					MinSeverity:               domain.DIAntipatternSeverityError,
				}
				mergedReq := createValidDIAntipatternRequest()
				mergedReq.ConstructorParamThreshold = 3

				configLoader.On("LoadConfig", "/config.yaml").Return(configReq, nil)
				configLoader.On("MergeConfig", configReq, mock.AnythingOfType("*domain.DIAntipatternRequest")).
					Return(&mergedReq)
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DIAntipatternRequest")).
					Return(createMockDIAntipatternResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			request: func() domain.DIAntipatternRequest {
				req := createValidDIAntipatternRequest()
				req.ConfigPath = "/config.yaml"
				return req
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader := setupDIAntipatternUseCaseMocks()

			tt.setupMocks(service, fileReader, formatter, configLoader)

			err := useCase.Execute(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.IsType(t, domain.DomainError{}, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			service.AssertExpectations(t)
			fileReader.AssertExpectations(t)
			formatter.AssertExpectations(t)
			configLoader.AssertExpectations(t)
		})
	}
}

func TestDIAntipatternUseCase_AnalyzeAndReturn(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockDIAntipatternService, *mockFileReader, *mockDIAntipatternOutputFormatter, *mockDIAntipatternConfigurationLoader)
		request     domain.DIAntipatternRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful analysis without formatting",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DIAntipatternRequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DIAntipatternRequest")).
					Return(createMockDIAntipatternResponse(), nil)
			},
			request:     createValidDIAntipatternRequest(),
			expectError: false,
		},
		{
			name: "validation error in analyze and return",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
			},
			request: domain.DIAntipatternRequest{
				Paths:        []string{},
				OutputWriter: os.Stdout,
			},
			expectError: true,
			errorMsg:    "no input paths specified",
		},
		{
			name: "service error in analyze and return",
			setupMocks: func(service *mockDIAntipatternService, fileReader *mockFileReader, formatter *mockDIAntipatternOutputFormatter, configLoader *mockDIAntipatternConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DIAntipatternRequest)(nil))
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DIAntipatternRequest")).
					Return((*domain.DIAntipatternResponse)(nil), errors.New("analysis failed"))
			},
			request:     createValidDIAntipatternRequest(),
			expectError: true,
			errorMsg:    "DI anti-pattern analysis failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader := setupDIAntipatternUseCaseMocks()

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
				assert.Equal(t, 1, len(response.Findings))
				assert.Equal(t, "TestService", response.Findings[0].ClassName)
			}

			service.AssertExpectations(t)
			fileReader.AssertExpectations(t)
			formatter.AssertExpectations(t)
			configLoader.AssertExpectations(t)
		})
	}
}

func TestDIAntipatternUseCaseBuilder(t *testing.T) {
	t.Run("missing service", func(t *testing.T) {
		_, err := NewDIAntipatternUseCaseBuilder().
			WithFileReader(&mockFileReader{}).
			WithFormatter(&mockDIAntipatternOutputFormatter{}).
			Build()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "DI anti-pattern service is required")
	})

	t.Run("missing file reader", func(t *testing.T) {
		_, err := NewDIAntipatternUseCaseBuilder().
			WithService(&mockDIAntipatternService{}).
			WithFormatter(&mockDIAntipatternOutputFormatter{}).
			Build()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "file reader is required")
	})

	t.Run("missing formatter", func(t *testing.T) {
		_, err := NewDIAntipatternUseCaseBuilder().
			WithService(&mockDIAntipatternService{}).
			WithFileReader(&mockFileReader{}).
			Build()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "output formatter is required")
	})

	t.Run("all required deps, no configLoader - injects noOp", func(t *testing.T) {
		uc, err := NewDIAntipatternUseCaseBuilder().
			WithService(&mockDIAntipatternService{}).
			WithFileReader(&mockFileReader{}).
			WithFormatter(&mockDIAntipatternOutputFormatter{}).
			Build()

		assert.NoError(t, err)
		assert.NotNil(t, uc)
		assert.IsType(t, &noOpDIAntipatternConfigLoader{}, uc.configLoader)
	})

	t.Run("with configLoader - uses provided loader", func(t *testing.T) {
		configLoader := &mockDIAntipatternConfigurationLoader{}

		uc, err := NewDIAntipatternUseCaseBuilder().
			WithService(&mockDIAntipatternService{}).
			WithFileReader(&mockFileReader{}).
			WithFormatter(&mockDIAntipatternOutputFormatter{}).
			WithConfigLoader(configLoader).
			Build()

		assert.NoError(t, err)
		assert.NotNil(t, uc)
		assert.Equal(t, configLoader, uc.configLoader)
	})

	t.Run("with output writer - overrides default", func(t *testing.T) {
		reportWriter := &mockReportWriter{}

		uc, err := NewDIAntipatternUseCaseBuilder().
			WithService(&mockDIAntipatternService{}).
			WithFileReader(&mockFileReader{}).
			WithFormatter(&mockDIAntipatternOutputFormatter{}).
			WithOutputWriter(reportWriter).
			Build()

		assert.NoError(t, err)
		assert.NotNil(t, uc)
		assert.Equal(t, reportWriter, uc.output)
	})

	t.Run("builder methods are chainable", func(t *testing.T) {
		builder := NewDIAntipatternUseCaseBuilder()
		result := builder.
			WithService(&mockDIAntipatternService{}).
			WithFileReader(&mockFileReader{}).
			WithFormatter(&mockDIAntipatternOutputFormatter{}).
			WithConfigLoader(&mockDIAntipatternConfigurationLoader{}).
			WithOutputWriter(&mockReportWriter{})

		assert.Equal(t, builder, result)
	})
}

func TestNoOpDIAntipatternConfigLoader(t *testing.T) {
	loader := &noOpDIAntipatternConfigLoader{}

	t.Run("LoadConfig returns nil", func(t *testing.T) {
		result, err := loader.LoadConfig("/any/path.yaml")

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("LoadDefaultConfig returns nil", func(t *testing.T) {
		result := loader.LoadDefaultConfig()

		assert.Nil(t, result)
	})

	t.Run("MergeConfig returns override", func(t *testing.T) {
		base := &domain.DIAntipatternRequest{ConstructorParamThreshold: 3}
		override := &domain.DIAntipatternRequest{ConstructorParamThreshold: 7}

		result := loader.MergeConfig(base, override)

		assert.Equal(t, override, result)
	})
}
