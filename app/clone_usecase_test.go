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

// Mock implementations for CloneUseCase
type mockCloneService struct {
	mock.Mock
}

func (m *mockCloneService) DetectClones(ctx context.Context, req *domain.CloneRequest) (*domain.CloneResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CloneResponse), args.Error(1)
}

func (m *mockCloneService) DetectClonesInFiles(ctx context.Context, filePaths []string, req *domain.CloneRequest) (*domain.CloneResponse, error) {
	args := m.Called(ctx, filePaths, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CloneResponse), args.Error(1)
}

func (m *mockCloneService) ComputeSimilarity(ctx context.Context, fragment1, fragment2 string) (float64, error) {
	args := m.Called(ctx, fragment1, fragment2)
	return args.Get(0).(float64), args.Error(1)
}

type mockFileReader struct {
	mock.Mock
}

func (m *mockFileReader) CollectPythonFiles(paths []string, recursive bool, includePatterns, excludePatterns []string) ([]string, error) {
	args := m.Called(paths, recursive, includePatterns, excludePatterns)
	return args.Get(0).([]string), args.Error(1)
}

func (m *mockFileReader) ReadFile(path string) ([]byte, error) {
	args := m.Called(path)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockFileReader) IsValidPythonFile(filePath string) bool {
	args := m.Called(filePath)
	return args.Bool(0)
}

func (m *mockFileReader) FileExists(path string) (bool, error) {
	args := m.Called(path)
	return args.Bool(0), args.Error(1)
}

type mockProgressReporter struct {
	mock.Mock
}

func (m *mockProgressReporter) StartProgress(totalFiles int) {
	m.Called(totalFiles)
}

func (m *mockProgressReporter) UpdateProgress(currentFile string, processed, total int) {
	m.Called(currentFile, processed, total)
}

func (m *mockProgressReporter) FinishProgress() {
	m.Called()
}

type mockCloneOutputFormatter struct {
	mock.Mock
}

func (m *mockCloneOutputFormatter) FormatCloneResponse(response *domain.CloneResponse, format domain.OutputFormat, writer io.Writer) error {
	args := m.Called(response, format, writer)
	return args.Error(0)
}

func (m *mockCloneOutputFormatter) FormatCloneStatistics(stats *domain.CloneStatistics, format domain.OutputFormat, writer io.Writer) error {
	args := m.Called(stats, format, writer)
	return args.Error(0)
}

type mockCloneConfigurationLoader struct {
	mock.Mock
}

func (m *mockCloneConfigurationLoader) LoadCloneConfig(path string) (*domain.CloneRequest, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CloneRequest), args.Error(1)
}

func (m *mockCloneConfigurationLoader) GetDefaultCloneConfig() *domain.CloneRequest {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*domain.CloneRequest)
}

func (m *mockCloneConfigurationLoader) SaveCloneConfig(req *domain.CloneRequest, path string) error {
	args := m.Called(req, path)
	return args.Error(0)
}

// Helper functions for CloneUseCase tests
func setupCloneUseCaseMocks() (*CloneUseCase, *mockCloneService, *mockFileReader, *mockCloneOutputFormatter, *mockCloneConfigurationLoader, *mockProgressReporter) {
	service := &mockCloneService{}
	fileReader := &mockFileReader{}
	formatter := &mockCloneOutputFormatter{}
	configLoader := &mockCloneConfigurationLoader{}
	progress := &mockProgressReporter{}

	useCase := NewCloneUseCase(service, fileReader, formatter, configLoader, progress)
	return useCase, service, fileReader, formatter, configLoader, progress
}

func createValidCloneRequest() domain.CloneRequest {
	return domain.CloneRequest{
		Paths:               []string{"/test/file1.py", "/test/file2.py"},
		OutputWriter:        os.Stdout,
		OutputFormat:        domain.OutputFormatText,
		SortBy:              domain.SortBySimilarity,
		MinLines:            5,
		MinNodes:            10,
		SimilarityThreshold: 0.8,
		MaxEditDistance:     3,
		Type1Threshold:      0.95,
		Type2Threshold:      0.85,
		Type3Threshold:      0.75,
		Type4Threshold:      0.65,
		CloneTypes:          []domain.CloneType{domain.Type1Clone, domain.Type2Clone, domain.Type3Clone},
		Recursive:           true,
		IncludePatterns:     []string{"*.py"},
		ExcludePatterns:     []string{"*test*"},
		ShowDetails:         true,
		ShowContent:         false,
		GroupClones:         true,
	}
}

func createMockCloneResponse() *domain.CloneResponse {
	return &domain.CloneResponse{
		Clones: []*domain.Clone{
			{
				ID:   1,
				Type: domain.Type2Clone,
				Location: &domain.CloneLocation{
					FilePath:  "/test/file1.py",
					StartLine: 10,
					EndLine:   20,
					StartCol:  0,
					EndCol:    10,
				},
				Content:    "def calculate_sum(a, b):\n    return a + b",
				Hash:       "hash1",
				Size:       15,
				LineCount:  11,
				Complexity: 2,
			},
		},
		ClonePairs: []*domain.ClonePair{
			{
				ID: 1,
				Clone1: &domain.Clone{
					ID:   1,
					Type: domain.Type2Clone,
					Location: &domain.CloneLocation{
						FilePath:  "/test/file1.py",
						StartLine: 10,
						EndLine:   20,
						StartCol:  0,
						EndCol:    10,
					},
				},
				Clone2: &domain.Clone{
					ID:   2,
					Type: domain.Type2Clone,
					Location: &domain.CloneLocation{
						FilePath:  "/test/file2.py",
						StartLine: 15,
						EndLine:   25,
						StartCol:  0,
						EndCol:    10,
					},
				},
				Similarity: 0.85,
				Distance:   2.0,
				Type:       domain.Type2Clone,
				Confidence: 0.9,
			},
		},
		CloneGroups: []*domain.CloneGroup{
			{
				ID: 1,
				Clones: []*domain.Clone{
					{
						ID:   1,
						Type: domain.Type2Clone,
					},
				},
				Type:       domain.Type2Clone,
				Similarity: 0.85,
				Size:       2,
			},
		},
		Statistics: &domain.CloneStatistics{
			TotalClones:       1,
			TotalClonePairs:   1,
			TotalCloneGroups:  1,
			ClonesByType:      map[string]int{"Type-2": 1},
			AverageSimilarity: 0.85,
			LinesAnalyzed:     1000,
			FilesAnalyzed:     2,
		},
		Request:  nil, // Will be set by the use case
		Duration: 1500,
		Success:  true,
	}
}

func TestCloneUseCase_Execute(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockCloneService, *mockFileReader, *mockCloneOutputFormatter, *mockCloneConfigurationLoader, *mockProgressReporter)
		request     domain.CloneRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful execution with valid request",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("GetDefaultCloneConfig").Return((*domain.CloneRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file1.py", "/test/file2.py"}, true, []string{"*.py"}, []string{"*test*"}).
					Return([]string{"/test/file1.py", "/test/file2.py"}, nil)
				service.On("DetectClones", mock.Anything, mock.AnythingOfType("*domain.CloneRequest")).
					Return(createMockCloneResponse(), nil)
				formatter.On("FormatCloneResponse", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			request:     createValidCloneRequest(),
			expectError: false,
		},
		{
			name: "validation error - request validation fails",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.CloneRequest{
				Paths:               []string{},
				MinLines:            -1,
				SimilarityThreshold: 1.5, // Invalid threshold > 1.0
			},
			expectError: true,
			errorMsg:    "validation failed",
		},
		{
			name: "configuration loading error",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("LoadCloneConfig", "/invalid/config.yaml").
					Return((*domain.CloneRequest)(nil), errors.New("config file not found"))
			},
			request: func() domain.CloneRequest {
				req := createValidCloneRequest()
				req.ConfigPath = "/invalid/config.yaml"
				return req
			}(),
			expectError: true,
			errorMsg:    "failed to load configuration",
		},
		{
			name: "file collection error",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("GetDefaultCloneConfig").Return((*domain.CloneRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/invalid/path"}, true, []string{"*.py"}, []string{"*test*"}).
					Return([]string{}, errors.New("path not found"))
			},
			request: func() domain.CloneRequest {
				req := createValidCloneRequest()
				req.Paths = []string{"/invalid/path"}
				return req
			}(),
			expectError: true,
			errorMsg:    "failed to collect files",
		},
		{
			name: "no files found - outputs empty results",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("GetDefaultCloneConfig").Return((*domain.CloneRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/empty/path"}, true, []string{"*.py"}, []string{"*test*"}).
					Return([]string{}, nil)
				formatter.On("FormatCloneResponse", mock.MatchedBy(func(resp *domain.CloneResponse) bool {
					return resp.Statistics.TotalClones == 0
				}), domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			request: func() domain.CloneRequest {
				req := createValidCloneRequest()
				req.Paths = []string{"/empty/path"}
				return req
			}(),
			expectError: false,
		},
		{
			name: "clone detection service error",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("GetDefaultCloneConfig").Return((*domain.CloneRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file1.py", "/test/file2.py"}, true, []string{"*.py"}, []string{"*test*"}).
					Return([]string{"/test/file1.py", "/test/file2.py"}, nil)
				service.On("DetectClones", mock.Anything, mock.AnythingOfType("*domain.CloneRequest")).
					Return((*domain.CloneResponse)(nil), errors.New("APTED analysis failed"))
			},
			request:     createValidCloneRequest(),
			expectError: true,
			errorMsg:    "clone detection failed",
		},
		{
			name: "invalid output writer error",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("GetDefaultCloneConfig").Return((*domain.CloneRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file1.py", "/test/file2.py"}, true, []string{"*.py"}, []string{"*test*"}).
					Return([]string{"/test/file1.py", "/test/file2.py"}, nil)
				service.On("DetectClones", mock.Anything, mock.AnythingOfType("*domain.CloneRequest")).
					Return(createMockCloneResponse(), nil)
			},
			request: func() domain.CloneRequest {
				req := createValidCloneRequest()
				req.OutputWriter = nil
				return req
			}(),
			expectError: true,
			errorMsg:    "no valid output writer specified",
		},
		{
			name: "output formatting error",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("GetDefaultCloneConfig").Return((*domain.CloneRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file1.py", "/test/file2.py"}, true, []string{"*.py"}, []string{"*test*"}).
					Return([]string{"/test/file1.py", "/test/file2.py"}, nil)
				service.On("DetectClones", mock.Anything, mock.AnythingOfType("*domain.CloneRequest")).
					Return(createMockCloneResponse(), nil)
				formatter.On("FormatCloneResponse", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).
					Return(errors.New("write failed"))
			},
			request:     createValidCloneRequest(),
			expectError: true,
			errorMsg:    "failed to format output",
		},
		{
			name: "successful execution with config loading",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configReq := &domain.CloneRequest{
					MinLines:            10,
					SimilarityThreshold: 0.9,
					ShowDetails:         false,
				}
				
				configLoader.On("LoadCloneConfig", "/config.yaml").Return(configReq, nil)
				fileReader.On("CollectPythonFiles", []string{"/test/file1.py", "/test/file2.py"}, true, []string{"*.py"}, []string{"*test*"}).
					Return([]string{"/test/file1.py", "/test/file2.py"}, nil)
				service.On("DetectClones", mock.Anything, mock.AnythingOfType("*domain.CloneRequest")).
					Return(createMockCloneResponse(), nil)
				formatter.On("FormatCloneResponse", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			request: func() domain.CloneRequest {
				req := createValidCloneRequest()
				req.ConfigPath = "/config.yaml"
				return req
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader, progress := setupCloneUseCaseMocks()

			tt.setupMocks(service, fileReader, formatter, configLoader, progress)

			err := useCase.Execute(context.Background(), tt.request)

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

func TestCloneUseCase_ExecuteAndReturn(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(*mockCloneService, *mockFileReader, *mockCloneOutputFormatter, *mockCloneConfigurationLoader, *mockProgressReporter)
		request     domain.CloneRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful analysis without formatting",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("GetDefaultCloneConfig").Return((*domain.CloneRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file1.py", "/test/file2.py"}, true, []string{"*.py"}, []string{"*test*"}).
					Return([]string{"/test/file1.py", "/test/file2.py"}, nil)
				service.On("DetectClones", mock.Anything, mock.AnythingOfType("*domain.CloneRequest")).
					Return(createMockCloneResponse(), nil)
			},
			request:     createValidCloneRequest(),
			expectError: false,
		},
		{
			name: "validation error in execute and return",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.CloneRequest{
				Paths:               []string{},
				MinLines:            -1,
				SimilarityThreshold: 1.5,
			},
			expectError: true,
			errorMsg:    "invalid request",
		},
		{
			name: "empty paths error",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				// No mocks needed - validation fails before any service calls
			},
			request: domain.CloneRequest{
				Paths:               []string{},
				MinLines:            5,
				SimilarityThreshold: 0.8,
			},
			expectError: true,
			errorMsg:    "paths cannot be empty",
		},
		{
			name: "file reader not initialized",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("GetDefaultCloneConfig").Return((*domain.CloneRequest)(nil))
			},
			request: domain.CloneRequest{
				Paths:               []string{"/test/file.py"},
				MinLines:            5,
				MinNodes:            10,
				SimilarityThreshold: 0.8,
				Type1Threshold:      0.95,
				Type2Threshold:      0.85,
				Type3Threshold:      0.75,
				Type4Threshold:      0.65,
			},
			expectError: true,
			errorMsg:    "file reader not initialized",
		},
		{
			name: "no files found error",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("GetDefaultCloneConfig").Return((*domain.CloneRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/empty/path"}, true, []string{"*.py"}, []string{"*test*"}).
					Return([]string{}, nil)
			},
			request: func() domain.CloneRequest {
				req := createValidCloneRequest()
				req.Paths = []string{"/empty/path"}
				return req
			}(),
			expectError: true,
			errorMsg:    "no Python files found in the specified paths",
		},
		{
			name: "analysis error in execute and return",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("GetDefaultCloneConfig").Return((*domain.CloneRequest)(nil))
				fileReader.On("CollectPythonFiles", []string{"/test/file1.py", "/test/file2.py"}, true, []string{"*.py"}, []string{"*test*"}).
					Return([]string{"/test/file1.py", "/test/file2.py"}, nil)
				service.On("DetectClones", mock.Anything, mock.AnythingOfType("*domain.CloneRequest")).
					Return((*domain.CloneResponse)(nil), errors.New("AST comparison failed"))
			},
			request:     createValidCloneRequest(),
			expectError: true,
			errorMsg:    "clone detection failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader, progress := setupCloneUseCaseMocks()
			
			// Set fileReader to nil for the specific test case
			if strings.Contains(tt.name, "file reader not initialized") {
				useCase.fileReader = nil
			}

			tt.setupMocks(service, fileReader, formatter, configLoader, progress)

			response, err := useCase.ExecuteAndReturn(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, 1, len(response.Clones))
				assert.Equal(t, domain.Type2Clone, response.Clones[0].Type)
				assert.Equal(t, 1, response.Clones[0].ID)
			}

			// Verify all mock expectations
			service.AssertExpectations(t)
			if useCase.fileReader != nil {
				fileReader.AssertExpectations(t)
			}
			formatter.AssertExpectations(t)
			configLoader.AssertExpectations(t)
			progress.AssertExpectations(t)
		})
	}
}

func TestCloneUseCase_ExecuteWithFiles(t *testing.T) {
	tests := []struct {
		name        string
		filePaths   []string
		setupMocks  func(*mockCloneService, *mockFileReader, *mockCloneOutputFormatter, *mockCloneConfigurationLoader, *mockProgressReporter)
		expectError bool
		errorMsg    string
	}{
		{
			name:      "successful analysis with specific files",
			filePaths: []string{"/test/file1.py", "/test/file2.py"},
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				fileReader.On("IsValidPythonFile", "/test/file1.py").Return(true)
				fileReader.On("IsValidPythonFile", "/test/file2.py").Return(true)
				service.On("DetectClonesInFiles", mock.Anything, []string{"/test/file1.py", "/test/file2.py"}, mock.AnythingOfType("*domain.CloneRequest")).
					Return(createMockCloneResponse(), nil)
				formatter.On("FormatCloneResponse", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			expectError: false,
		},
		{
			name:      "validation error in execute with files",
			filePaths: []string{"/test/file.py"},
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				// No mocks needed - validation fails before any service calls
			},
			expectError: true,
			errorMsg:    "validation failed",
		},
		{
			name:      "no valid python files",
			filePaths: []string{"/test/file.txt", "/test/file.java"},
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				fileReader.On("IsValidPythonFile", "/test/file.txt").Return(false)
				fileReader.On("IsValidPythonFile", "/test/file.java").Return(false)
				formatter.On("FormatCloneResponse", mock.MatchedBy(func(resp *domain.CloneResponse) bool {
					return resp.Statistics.TotalClones == 0
				}), domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			expectError: false, // Should output empty results, not error
		},
		{
			name:      "mixed valid and invalid files",
			filePaths: []string{"/test/file.py", "/test/file.txt"},
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				fileReader.On("IsValidPythonFile", "/test/file.py").Return(true)
				fileReader.On("IsValidPythonFile", "/test/file.txt").Return(false)
				service.On("DetectClonesInFiles", mock.Anything, []string{"/test/file.py"}, mock.AnythingOfType("*domain.CloneRequest")).
					Return(createMockCloneResponse(), nil)
				formatter.On("FormatCloneResponse", mock.Anything, domain.OutputFormatText, mock.AnythingOfType("*os.File")).Return(nil)
			},
			expectError: false,
		},
		{
			name:      "clone detection service error",
			filePaths: []string{"/test/file.py"},
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				fileReader.On("IsValidPythonFile", "/test/file.py").Return(true)
				service.On("DetectClonesInFiles", mock.Anything, []string{"/test/file.py"}, mock.AnythingOfType("*domain.CloneRequest")).
					Return((*domain.CloneResponse)(nil), errors.New("tree comparison failed"))
			},
			expectError: true,
			errorMsg:    "clone detection failed",
		},
		{
			name:      "invalid output writer error",
			filePaths: []string{"/test/file.py"},
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				fileReader.On("IsValidPythonFile", "/test/file.py").Return(true)
				service.On("DetectClonesInFiles", mock.Anything, []string{"/test/file.py"}, mock.AnythingOfType("*domain.CloneRequest")).
					Return(createMockCloneResponse(), nil)
			},
			expectError: true,
			errorMsg:    "no valid output writer specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader, progress := setupCloneUseCaseMocks()

			tt.setupMocks(service, fileReader, formatter, configLoader, progress)

			req := createValidCloneRequest()
			if strings.Contains(tt.name, "validation error") {
				req.MinLines = -1 // Invalid value
			}
			if strings.Contains(tt.name, "invalid output writer") {
				req.OutputWriter = nil
			}

			err := useCase.ExecuteWithFiles(context.Background(), tt.filePaths, req)

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

func TestCloneUseCase_ComputeFragmentSimilarity(t *testing.T) {
	tests := []struct {
		name        string
		fragment1   string
		fragment2   string
		setupMocks  func(*mockCloneService, *mockFileReader, *mockCloneOutputFormatter, *mockCloneConfigurationLoader, *mockProgressReporter)
		expectError bool
		expectedSim float64
		errorMsg    string
	}{
		{
			name:      "successful similarity computation",
			fragment1: "def func1():\n    return x + y",
			fragment2: "def func2():\n    return a + b",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				service.On("ComputeSimilarity", mock.Anything, "def func1():\n    return x + y", "def func2():\n    return a + b").
					Return(0.75, nil)
			},
			expectError: false,
			expectedSim: 0.75,
		},
		{
			name:      "similarity computation error",
			fragment1: "def func1():\n    return x + y",
			fragment2: "invalid syntax >>>",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				service.On("ComputeSimilarity", mock.Anything, "def func1():\n    return x + y", "invalid syntax >>>").
					Return(0.0, errors.New("failed to parse fragment"))
			},
			expectError: true,
			expectedSim: 0.0,
			errorMsg:    "failed to compute similarity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader, progress := setupCloneUseCaseMocks()

			tt.setupMocks(service, fileReader, formatter, configLoader, progress)

			similarity, err := useCase.ComputeFragmentSimilarity(context.Background(), tt.fragment1, tt.fragment2)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedSim, similarity)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedSim, similarity)
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

func TestCloneUseCase_SaveConfiguration(t *testing.T) {
	tests := []struct {
		name        string
		configPath  string
		setupMocks  func(*mockCloneService, *mockFileReader, *mockCloneOutputFormatter, *mockCloneConfigurationLoader, *mockProgressReporter)
		expectError bool
		errorMsg    string
	}{
		{
			name:       "successful config save",
			configPath: "/config/clone.yaml",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("SaveCloneConfig", mock.AnythingOfType("*domain.CloneRequest"), "/config/clone.yaml").Return(nil)
			},
			expectError: false,
		},
		{
			name:       "config save error",
			configPath: "/invalid/path/config.yaml",
			setupMocks: func(service *mockCloneService, fileReader *mockFileReader, formatter *mockCloneOutputFormatter, configLoader *mockCloneConfigurationLoader, progress *mockProgressReporter) {
				configLoader.On("SaveCloneConfig", mock.AnythingOfType("*domain.CloneRequest"), "/invalid/path/config.yaml").
					Return(errors.New("permission denied"))
			},
			expectError: true,
			errorMsg:    "failed to save configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase, service, fileReader, formatter, configLoader, progress := setupCloneUseCaseMocks()

			tt.setupMocks(service, fileReader, formatter, configLoader, progress)

			req := createValidCloneRequest()
			err := useCase.SaveConfiguration(req, tt.configPath)

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

func TestCloneUseCase_mergeConfiguration(t *testing.T) {
	useCase := &CloneUseCase{}

	tests := []struct {
		name      string
		configReq domain.CloneRequest
		requestReq domain.CloneRequest
		expected  domain.CloneRequest
	}{
		{
			name: "merge with request taking precedence",
			configReq: domain.CloneRequest{
				MinLines:            10,
				SimilarityThreshold: 0.9,
				ShowDetails:         true,
				CloneTypes:          []domain.CloneType{domain.Type1Clone},
			},
			requestReq: domain.CloneRequest{
				Paths:               []string{"/test/file.py"},
				MinLines:            5,  // Should override config
				ShowDetails:         false, // Should override config
				SimilarityThreshold: 0.8, // Should override config
				OutputFormat:        domain.OutputFormatJSON,
				OutputWriter:        os.Stdout,
			},
			expected: domain.CloneRequest{
				Paths:               []string{"/test/file.py"}, // From request
				MinLines:            10,                        // From config (request matched default, didn't override)
				SimilarityThreshold: 0.9,                      // From config (request matched default, didn't override)
				ShowDetails:         false,                    // From request (overrides config)
				CloneTypes:          []domain.CloneType{domain.Type1Clone}, // From config (not overridden)
				OutputFormat:        domain.OutputFormatJSON,  // From request
				OutputWriter:        os.Stdout,                // From request
			},
		},
		{
			name: "merge with default values not overriding config",
			configReq: domain.CloneRequest{
				MinLines:            15,
				SimilarityThreshold: 0.95,
				Type1Threshold:      0.98,
				CloneTypes:          []domain.CloneType{domain.Type1Clone, domain.Type2Clone},
			},
			requestReq: *domain.DefaultCloneRequest(), // All default values
			expected: domain.CloneRequest{
				MinLines:            15,   // From config (not overridden by default)
				SimilarityThreshold: 0.95, // From config (not overridden by default)
				Type1Threshold:      0.98, // From config (not overridden by default)
				CloneTypes:          []domain.CloneType{domain.Type1Clone, domain.Type2Clone}, // From config
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := useCase.mergeConfiguration(tt.configReq, tt.requestReq)

			// Check key fields that should be merged correctly
			if len(tt.expected.Paths) > 0 {
				assert.Equal(t, tt.expected.Paths, result.Paths)
			}
			if tt.expected.MinLines != 0 {
				assert.Equal(t, tt.expected.MinLines, result.MinLines)
			}
			if tt.expected.SimilarityThreshold != 0 {
				assert.Equal(t, tt.expected.SimilarityThreshold, result.SimilarityThreshold)
			}
			if tt.expected.Type1Threshold != 0 {
				assert.Equal(t, tt.expected.Type1Threshold, result.Type1Threshold)
			}
			if len(tt.expected.CloneTypes) > 0 {
				assert.Equal(t, tt.expected.CloneTypes, result.CloneTypes)
			}

			// Output settings should always come from request
			assert.Equal(t, tt.requestReq.OutputFormat, result.OutputFormat)
			assert.Equal(t, tt.requestReq.OutputWriter, result.OutputWriter)
			assert.Equal(t, tt.requestReq.SortBy, result.SortBy)
		})
	}
}