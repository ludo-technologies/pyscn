package app

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for DeadCodeUseCase
type mockDeadCodeService struct {
	mock.Mock
}

func (m *mockDeadCodeService) Analyze(ctx context.Context, req domain.DeadCodeRequest) (*domain.DeadCodeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DeadCodeResponse), args.Error(1)
}

func (m *mockDeadCodeService) AnalyzeFile(ctx context.Context, filePath string, req domain.DeadCodeRequest) (*domain.FileDeadCode, error) {
	args := m.Called(ctx, filePath, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FileDeadCode), args.Error(1)
}

func (m *mockDeadCodeService) AnalyzeFunction(ctx context.Context, functionCFG interface{}, req domain.DeadCodeRequest) (*domain.FunctionDeadCode, error) {
	args := m.Called(ctx, functionCFG, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FunctionDeadCode), args.Error(1)
}

type mockDeadCodeFormatter struct {
	mock.Mock
}

func (m *mockDeadCodeFormatter) Format(response *domain.DeadCodeResponse, format domain.OutputFormat) (string, error) {
	args := m.Called(response, format)
	return args.String(0), args.Error(1)
}

func (m *mockDeadCodeFormatter) Write(response *domain.DeadCodeResponse, format domain.OutputFormat, writer io.Writer) error {
	args := m.Called(response, format, writer)
	return args.Error(0)
}

func (m *mockDeadCodeFormatter) FormatFinding(finding domain.DeadCodeFinding, format domain.OutputFormat) (string, error) {
	args := m.Called(finding, format)
	return args.String(0), args.Error(1)
}

type mockDeadCodeConfigurationLoader struct {
	mock.Mock
}

func (m *mockDeadCodeConfigurationLoader) LoadConfig(path string) (*domain.DeadCodeRequest, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DeadCodeRequest), args.Error(1)
}

func (m *mockDeadCodeConfigurationLoader) LoadDefaultConfig() *domain.DeadCodeRequest {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.DeadCodeRequest)
}

func (m *mockDeadCodeConfigurationLoader) MergeConfig(base *domain.DeadCodeRequest, override *domain.DeadCodeRequest) *domain.DeadCodeRequest {
	args := m.Called(base, override)
	return args.Get(0).(*domain.DeadCodeRequest)
}

// mockFileReader is defined in clone_usecase_test.go

// Helper functions for DeadCodeUseCase tests
func setupDeadCodeUseCaseMocks() (*DeadCodeUseCase, *mockDeadCodeService, *mockFileReader, *mockDeadCodeFormatter, *mockDeadCodeConfigurationLoader) {
	service := &mockDeadCodeService{}
	fileReader := &mockFileReader{}
	formatter := &mockDeadCodeFormatter{}
	configLoader := &mockDeadCodeConfigurationLoader{}

	useCase := NewDeadCodeUseCase(service, fileReader, formatter, configLoader)
	return useCase, service, fileReader, formatter, configLoader
}

func createValidDeadCodeRequest() domain.DeadCodeRequest {
	return domain.DeadCodeRequest{
		Paths:                     []string{"/test/file.py"},
		OutputWriter:              os.Stdout,
		OutputFormat:              domain.OutputFormatText,
		SortBy:                    domain.DeadCodeSortBySeverity,
		MinSeverity:               domain.DeadCodeSeverityWarning,
		ContextLines:              3,
		ShowContext:               true,
		Recursive:                 true,
		IncludePatterns:           []string{"**/*.py"},
		ExcludePatterns:           []string{},
		IgnorePatterns:            []string{},
		DetectAfterReturn:         true,
		DetectAfterBreak:          true,
		DetectAfterContinue:       true,
		DetectAfterRaise:          true,
		DetectUnreachableBranches: true,
	}
}

func createMockDeadCodeResponse() *domain.DeadCodeResponse {
	return &domain.DeadCodeResponse{
		Files: []domain.FileDeadCode{
			{
				FilePath:          "/test/file.py",
				TotalFindings:     2,
				TotalFunctions:    1,
				AffectedFunctions: 1,
				Functions: []domain.FunctionDeadCode{
					{
						Name:           "test_function",
						FilePath:       "/test/file.py",
						TotalBlocks:    5,
						DeadBlocks:     2,
						ReachableRatio: 0.6,
						CriticalCount:  0,
						WarningCount:   2,
						Findings: []domain.DeadCodeFinding{
							{
								Location: domain.DeadCodeLocation{
									FilePath:    "/test/file.py",
									StartLine:   15,
									EndLine:     15,
									StartColumn: 4,
									EndColumn:   30,
								},
								FunctionName: "test_function",
								Code:         "print('unreachable')  # Dead code",
								Reason:       "Code after return statement is unreachable",
								Severity:     domain.DeadCodeSeverityWarning,
								Description:  "Remove unreachable code after return statement",
								Context:      []string{"def test_function():", "    return True", "    print('unreachable')  # Dead code"},
								BlockID:      "block_1",
							},
							{
								Location: domain.DeadCodeLocation{
									FilePath:    "/test/file.py",
									StartLine:   18,
									EndLine:     18,
									StartColumn: 8,
									EndColumn:   35,
								},
								FunctionName: "test_function",
								Code:         "print('never executed')  # Dead code",
								Reason:       "This branch is never reached",
								Severity:     domain.DeadCodeSeverityInfo,
								Description:  "Remove or fix condition that makes this branch unreachable",
								Context:      []string{"if False:", "    print('never executed')  # Dead code"},
								BlockID:      "block_2",
							},
						},
					},
				},
			},
		},
		Summary: domain.DeadCodeSummary{
			TotalFiles:            1,
			FilesWithDeadCode:     1,
			TotalFindings:         2,
			TotalFunctions:        1,
			FunctionsWithDeadCode: 1,
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}
}

func TestDeadCodeUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockDeadCodeService, *mockFileReader, *mockDeadCodeFormatter, *mockDeadCodeConfigurationLoader)
		request     domain.DeadCodeRequest
		expectError bool
		errorType   string
		errorMsg    string
	}{
		{
			name: "successful execution with valid request",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DeadCodeRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file.py"}, true, []string{"**/*.py"}, []string{}).
					Return([]string{"/test/file.py"}, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DeadCodeRequest")).
					Return(createMockDeadCodeResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			request:     createValidDeadCodeRequest(),
			expectError: false,
		},
		{
			name: "validation error - empty paths",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.DeadCodeRequest{
				Paths:        []string{},
				OutputWriter: os.Stdout,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "no input paths specified",
		},
		{
			name: "validation error - nil output writer",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.DeadCodeRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: nil,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "output writer or output path is required",
		},
		{
			name: "validation error - negative context lines",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.DeadCodeRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: os.Stdout,
				ContextLines: -1,
				OutputFormat: domain.OutputFormatText,
				SortBy:       domain.DeadCodeSortBySeverity,
				MinSeverity:  domain.DeadCodeSeverityWarning,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "context lines cannot be negative",
		},
		{
			name: "validation error - invalid severity level",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.DeadCodeRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: os.Stdout,
				ContextLines: 3,
				OutputFormat: domain.OutputFormatText,
				SortBy:       domain.DeadCodeSortBySeverity,
				MinSeverity:  domain.DeadCodeSeverity("invalid"),
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "unsupported minimum severity: invalid",
		},
		{
			name: "validation error - invalid output format",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.DeadCodeRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: os.Stdout,
				ContextLines: 3,
				OutputFormat: domain.OutputFormat("invalid"),
				SortBy:       domain.DeadCodeSortBySeverity,
				MinSeverity:  domain.DeadCodeSeverityWarning,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "unsupported output format: invalid",
		},
		{
			name: "validation error - invalid sort criteria",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.DeadCodeRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: os.Stdout,
				ContextLines: 3,
				OutputFormat: domain.OutputFormatText,
				SortBy:       domain.DeadCodeSortCriteria("invalid"),
				MinSeverity:  domain.DeadCodeSeverityWarning,
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "unsupported sort criteria: invalid",
		},
		{
			name: "configuration loading error",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				configLoader.On("LoadConfig", "/invalid/config.yaml").
					Return((*domain.DeadCodeRequest)(nil), errors.New("config file not found"))
			},
			request: domain.DeadCodeRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: os.Stdout,
				ContextLines: 3,
				OutputFormat: domain.OutputFormatText,
				SortBy:       domain.DeadCodeSortBySeverity,
				MinSeverity:  domain.DeadCodeSeverityWarning,
				ConfigPath:   "/invalid/config.yaml",
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "failed to load configuration",
		},
		{
			name: "file collection error",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DeadCodeRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/invalid/path"}, true, []string{"**/*.py"}, []string{}).
					Return([]string{}, errors.New("path not found"))
			},
			request: domain.DeadCodeRequest{
				Paths:           []string{"/invalid/path"},
				OutputWriter:    os.Stdout,
				ContextLines:    3,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.DeadCodeSortBySeverity,
				MinSeverity:     domain.DeadCodeSeverityWarning,
				Recursive:       true,
				IncludePatterns: []string{"**/*.py"},
				ExcludePatterns: []string{},
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "failed to collect files",
		},
		{
			name: "no files found error",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DeadCodeRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/empty/path"}, true, []string{"**/*.py"}, []string{}).
					Return([]string{}, nil)
			},
			request: domain.DeadCodeRequest{
				Paths:           []string{"/empty/path"},
				OutputWriter:    os.Stdout,
				ContextLines:    3,
				OutputFormat:    domain.OutputFormatText,
				SortBy:          domain.DeadCodeSortBySeverity,
				MinSeverity:     domain.DeadCodeSeverityWarning,
				Recursive:       true,
				IncludePatterns: []string{"**/*.py"},
				ExcludePatterns: []string{},
			},
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "no Python files found in the specified paths",
		},
		{
			name: "analysis service error",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DeadCodeRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file.py"}, true, []string{"**/*.py"}, []string{}).
					Return([]string{"/test/file.py"}, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DeadCodeRequest")).
					Return((*domain.DeadCodeResponse)(nil), errors.New("CFG analysis failed"))
			},
			request:     createValidDeadCodeRequest(),
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "dead code analysis failed",
		},
		{
			name: "output formatting error",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DeadCodeRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file.py"}, true, []string{"**/*.py"}, []string{}).
					Return([]string{"/test/file.py"}, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DeadCodeRequest")).
					Return(createMockDeadCodeResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, os.Stdout).
					Return(errors.New("write failed"))
			},
			request:     createValidDeadCodeRequest(),
			expectError: true,
			errorType:   "*domain.DomainError",
			errorMsg:    "failed to write output",
		},
		{
			name: "successful execution with config loading",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				configReq := &domain.DeadCodeRequest{
					MinSeverity:  domain.DeadCodeSeverityCritical,
					ContextLines: 5,
					ShowContext:  false,
				}
				mergedReq := createValidDeadCodeRequest()
				mergedReq.MinSeverity = domain.DeadCodeSeverityCritical

				configLoader.On("LoadConfig", "/config.yaml").Return(configReq, nil)
				configLoader.On("MergeConfig", configReq, mock.AnythingOfType("*domain.DeadCodeRequest")).
					Return(&mergedReq)
				fileReader.On("CollectPythonFiles", []string{"/test/file.py"}, true, []string{"**/*.py"}, []string{}).
					Return([]string{"/test/file.py"}, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DeadCodeRequest")).
					Return(createMockDeadCodeResponse(), nil)
				formatter.On("Write", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			request: func() domain.DeadCodeRequest {
				req := createValidDeadCodeRequest()
				req.ConfigPath = "/config.yaml"
				return req
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader := setupDeadCodeUseCaseMocks()

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

func TestDeadCodeUseCase_AnalyzeAndReturn(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockDeadCodeService, *mockFileReader, *mockDeadCodeFormatter, *mockDeadCodeConfigurationLoader)
		request     domain.DeadCodeRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful analysis without formatting",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DeadCodeRequest)(nil))
				// Mock file detection logic
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DeadCodeRequest")).
					Return(createMockDeadCodeResponse(), nil)
			},
			request:     createValidDeadCodeRequest(),
			expectError: false,
		},
		{
			name: "validation error in analyze and return",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.DeadCodeRequest{
				Paths:        []string{},
				OutputWriter: os.Stdout,
			},
			expectError: true,
			errorMsg:    "no input paths specified",
		},
		{
			name: "analysis error in analyze and return",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DeadCodeRequest)(nil))
				// Mock file detection logic
				fileReader.On("FileExists", "/test/file.py").Return(true, nil)
				service.On("Analyze", mock.Anything, mock.AnythingOfType("domain.DeadCodeRequest")).
					Return((*domain.DeadCodeResponse)(nil), errors.New("CFG construction failed"))
			},
			request:     createValidDeadCodeRequest(),
			expectError: true,
			errorMsg:    "dead code analysis failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader := setupDeadCodeUseCaseMocks()

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
				assert.Equal(t, 1, len(response.Files))
				assert.Equal(t, "/test/file.py", response.Files[0].FilePath)
				assert.Equal(t, 2, response.Files[0].TotalFindings)
			}

			// Verify all mock expectations
			service.AssertExpectations(t)
			fileReader.AssertExpectations(t)
			formatter.AssertExpectations(t)
			configLoader.AssertExpectations(t)
		})
	}
}

func TestDeadCodeUseCase_validateRequest(t *testing.T) {
	useCase := &DeadCodeUseCase{}

	tests := []struct {
		name    string
		request domain.DeadCodeRequest
		wantErr string
	}{
		{
			name: "valid request",
			request: domain.DeadCodeRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: os.Stdout,
				ContextLines: 3,
				MinSeverity:  domain.DeadCodeSeverityWarning,
				OutputFormat: domain.OutputFormatText,
				SortBy:       domain.DeadCodeSortBySeverity,
			},
			wantErr: "",
		},
		{
			name: "empty paths",
			request: domain.DeadCodeRequest{
				OutputWriter: os.Stdout,
			},
			wantErr: "no input paths specified",
		},
		{
			name: "nil output writer",
			request: domain.DeadCodeRequest{
				Paths: []string{"/test/file.py"},
			},
			wantErr: "output writer or output path is required",
		},
		{
			name: "negative context lines",
			request: domain.DeadCodeRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: os.Stdout,
				ContextLines: -1,
			},
			wantErr: "context lines cannot be negative",
		},
		{
			name: "invalid severity level",
			request: domain.DeadCodeRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: os.Stdout,
				ContextLines: 3,
				MinSeverity:  domain.DeadCodeSeverity("invalid"),
				OutputFormat: domain.OutputFormatText,
				SortBy:       domain.DeadCodeSortBySeverity,
			},
			wantErr: "unsupported minimum severity: invalid",
		},
		{
			name: "invalid output format",
			request: domain.DeadCodeRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: os.Stdout,
				ContextLines: 3,
				MinSeverity:  domain.DeadCodeSeverityWarning,
				OutputFormat: domain.OutputFormat("invalid"),
				SortBy:       domain.DeadCodeSortBySeverity,
			},
			wantErr: "unsupported output format: invalid",
		},
		{
			name: "invalid sort criteria",
			request: domain.DeadCodeRequest{
				Paths:        []string{"/test/file.py"},
				OutputWriter: os.Stdout,
				ContextLines: 3,
				MinSeverity:  domain.DeadCodeSeverityWarning,
				OutputFormat: domain.OutputFormatText,
				SortBy:       domain.DeadCodeSortCriteria("invalid"),
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

func TestDeadCodeUseCase_AnalyzeFile(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		setupMocks  func(*mockDeadCodeService, *mockFileReader, *mockDeadCodeFormatter, *mockDeadCodeConfigurationLoader)
		expectError bool
		errorMsg    string
	}{
		{
			name:     "file not found error",
			filePath: "/test/file.py",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				fileReader.On("IsValidPythonFile", "/test/file.py").Return(true)
				fileReader.On("FileExists", "/test/file.py").Return(false, nil)
			},
			expectError: true,
			errorMsg:    "file not found: /test/file.py",
		},
		{
			name:     "invalid python file",
			filePath: "/test/file.txt",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				fileReader.On("IsValidPythonFile", "/test/file.txt").Return(false)
			},
			expectError: true,
			errorMsg:    "not a valid Python file",
		},
		{
			name:     "analysis error - file not found",
			filePath: "/test/nonexistent.py",
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				fileReader.On("IsValidPythonFile", "/test/nonexistent.py").Return(true)
				fileReader.On("FileExists", "/test/nonexistent.py").Return(false, nil)
			},
			expectError: true,
			errorMsg:    "file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader := setupDeadCodeUseCaseMocks()

			tt.setupMocks(service, fileReader, formatter, configLoader)

			req := createValidDeadCodeRequest()
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

func TestDeadCodeUseCase_AnalyzeFunction(t *testing.T) {
	tests := []struct {
		name        string
		functionCFG interface{}
		setupMocks  func(*mockDeadCodeService, *mockFileReader, *mockDeadCodeFormatter, *mockDeadCodeConfigurationLoader)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful function analysis",
			functionCFG: map[string]interface{}{"nodes": []int{1, 2, 3}, "edges": [][]int{{1, 2}, {2, 3}}},
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DeadCodeRequest)(nil))

				mockFunctionResult := &domain.FunctionDeadCode{
					Name:           "test_function",
					FilePath:       "/test/file.py",
					TotalBlocks:    3,
					DeadBlocks:     1,
					ReachableRatio: 0.67,
					CriticalCount:  0,
					WarningCount:   1,
					Findings:       []domain.DeadCodeFinding{},
				}

				service.On("AnalyzeFunction", mock.Anything, mock.Anything, mock.AnythingOfType("domain.DeadCodeRequest")).
					Return(mockFunctionResult, nil)
			},
			expectError: false,
		},
		{
			name:        "nil function CFG",
			functionCFG: nil,
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				// No mocks needed - validation fails before any service calls
			},
			expectError: true,
			errorMsg:    "function CFG cannot be nil",
		},
		{
			name:        "analysis error",
			functionCFG: map[string]interface{}{"nodes": []int{1, 2, 3}},
			setupMocks: func(service *mockDeadCodeService, fileReader *mockFileReader, formatter *mockDeadCodeFormatter, configLoader *mockDeadCodeConfigurationLoader) {
				configLoader.On("LoadDefaultConfig").Return((*domain.DeadCodeRequest)(nil))
				service.On("AnalyzeFunction", mock.Anything, mock.Anything, mock.AnythingOfType("domain.DeadCodeRequest")).
					Return((*domain.FunctionDeadCode)(nil), errors.New("CFG analysis failed"))
			},
			expectError: true,
			errorMsg:    "function analysis failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader := setupDeadCodeUseCaseMocks()

			tt.setupMocks(service, fileReader, formatter, configLoader)

			req := createValidDeadCodeRequest()
			result, err := useCase.AnalyzeFunction(context.Background(), tt.functionCFG, req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "test_function", result.Name)
				assert.Equal(t, "/test/file.py", result.FilePath)
			}

			// Verify all mock expectations
			service.AssertExpectations(t)
			fileReader.AssertExpectations(t)
			formatter.AssertExpectations(t)
			configLoader.AssertExpectations(t)
		})
	}
}

func TestDeadCodeUseCase_loadAndMergeConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*mockDeadCodeConfigurationLoader)
		request    domain.DeadCodeRequest
		expectErr  bool
		errorMsg   string
	}{
		{
			name: "no config loader",
			setupMocks: func(configLoader *mockDeadCodeConfigurationLoader) {
				// No setup needed - configLoader will be nil
			},
			request:   createValidDeadCodeRequest(),
			expectErr: false,
		},
		{
			name: "load default config successfully",
			setupMocks: func(configLoader *mockDeadCodeConfigurationLoader) {
				defaultConfig := &domain.DeadCodeRequest{
					MinSeverity:  domain.DeadCodeSeverityCritical,
					ContextLines: 5,
				}
				configLoader.On("LoadDefaultConfig").Return(defaultConfig)
				configLoader.On("MergeConfig", defaultConfig, mock.AnythingOfType("*domain.DeadCodeRequest")).
					Return(&domain.DeadCodeRequest{MinSeverity: domain.DeadCodeSeverityCritical})
			},
			request:   createValidDeadCodeRequest(),
			expectErr: false,
		},
		{
			name: "load specific config file successfully",
			setupMocks: func(configLoader *mockDeadCodeConfigurationLoader) {
				configReq := &domain.DeadCodeRequest{
					MinSeverity:  domain.DeadCodeSeverityInfo,
					ContextLines: 7,
					ShowContext:  false,
				}
				configLoader.On("LoadConfig", "/config.yaml").Return(configReq, nil)
				configLoader.On("MergeConfig", configReq, mock.AnythingOfType("*domain.DeadCodeRequest")).
					Return(&domain.DeadCodeRequest{MinSeverity: domain.DeadCodeSeverityInfo})
			},
			request: func() domain.DeadCodeRequest {
				req := createValidDeadCodeRequest()
				req.ConfigPath = "/config.yaml"
				return req
			}(),
			expectErr: false,
		},
		{
			name: "config loading error",
			setupMocks: func(configLoader *mockDeadCodeConfigurationLoader) {
				configLoader.On("LoadConfig", "/invalid.yaml").Return((*domain.DeadCodeRequest)(nil), errors.New("config not found"))
			},
			request: func() domain.DeadCodeRequest {
				req := createValidDeadCodeRequest()
				req.ConfigPath = "/invalid.yaml"
				return req
			}(),
			expectErr: true,
			errorMsg:  "failed to load config from /invalid.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var useCase *DeadCodeUseCase
			var configLoader *mockDeadCodeConfigurationLoader

			if strings.Contains(tt.name, "no config loader") {
				useCase = &DeadCodeUseCase{configLoader: nil}
			} else {
				configLoader = &mockDeadCodeConfigurationLoader{}
				useCase = &DeadCodeUseCase{configLoader: configLoader}
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
