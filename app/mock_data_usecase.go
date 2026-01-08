package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ludo-technologies/pyscn/domain"
	svc "github.com/ludo-technologies/pyscn/service"
)

// MockDataUseCase orchestrates the mock data analysis workflow
type MockDataUseCase struct {
	service      domain.MockDataService
	fileReader   domain.FileReader
	formatter    domain.MockDataFormatter
	configLoader domain.MockDataConfigurationLoader
	output       domain.ReportWriter
}

// NewMockDataUseCase creates a new mock data use case
func NewMockDataUseCase(
	service domain.MockDataService,
	fileReader domain.FileReader,
	formatter domain.MockDataFormatter,
	configLoader domain.MockDataConfigurationLoader,
) *MockDataUseCase {
	return &MockDataUseCase{
		service:      service,
		fileReader:   fileReader,
		formatter:    formatter,
		configLoader: configLoader,
		output:       svc.NewFileOutputWriter(nil),
	}
}

// Execute performs the complete mock data analysis workflow
func (uc *MockDataUseCase) Execute(ctx context.Context, req domain.MockDataRequest) error {
	// Validate input
	if err := uc.validateRequest(req); err != nil {
		return domain.NewInvalidInputError("invalid request", err)
	}

	// Load configuration if specified
	finalReq, err := uc.loadAndMergeConfig(req)
	if err != nil {
		return domain.NewConfigError("failed to load configuration", err)
	}

	// Collect Python files
	files, err := uc.fileReader.CollectPythonFiles(
		finalReq.Paths,
		finalReq.Recursive,
		finalReq.IncludePatterns,
		finalReq.ExcludePatterns,
	)
	if err != nil {
		return domain.NewFileNotFoundError("failed to collect files", err)
	}

	if len(files) == 0 {
		return domain.NewInvalidInputError("no Python files found in the specified paths", nil)
	}

	// Update request with collected files
	finalReq.Paths = files

	// Perform analysis
	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return domain.NewAnalysisError("mock data analysis failed", err)
	}

	// Delegate output handling to ReportWriter
	var out io.Writer
	if finalReq.OutputPath == "" {
		out = finalReq.OutputWriter
	}
	if err := uc.output.Write(out, finalReq.OutputPath, finalReq.OutputFormat, finalReq.NoOpen, func(w io.Writer) error {
		return uc.formatter.Write(response, finalReq.OutputFormat, w)
	}); err != nil {
		return domain.NewOutputError("failed to write output", err)
	}

	return nil
}

// AnalyzeAndReturn performs mock data analysis and returns the response without formatting
func (uc *MockDataUseCase) AnalyzeAndReturn(ctx context.Context, req domain.MockDataRequest) (*domain.MockDataResponse, error) {
	// Validate input
	if err := uc.validateRequest(req); err != nil {
		return nil, domain.NewInvalidInputError("invalid request", err)
	}

	// Load configuration if specified
	finalReq, err := uc.loadAndMergeConfig(req)
	if err != nil {
		return nil, domain.NewConfigError("failed to load configuration", err)
	}

	// Resolve file paths
	files, err := ResolveFilePaths(
		uc.fileReader,
		finalReq.Paths,
		finalReq.Recursive,
		finalReq.IncludePatterns,
		finalReq.ExcludePatterns,
		false, // validatePythonFile: mock data doesn't need strict Python validation
	)
	if err != nil {
		return nil, domain.NewFileNotFoundError("failed to collect files", err)
	}

	if len(files) == 0 {
		return nil, domain.NewInvalidInputError("no Python files found in the specified paths", nil)
	}

	// Update request with collected files
	finalReq.Paths = files

	// Perform analysis and return the response
	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return nil, domain.NewAnalysisError("mock data analysis failed", err)
	}

	return response, nil
}

// AnalyzeFile analyzes a single file for mock data
func (uc *MockDataUseCase) AnalyzeFile(ctx context.Context, filePath string, req domain.MockDataRequest) error {
	// Validate file
	if !uc.fileReader.IsValidPythonFile(filePath) {
		return domain.NewInvalidInputError(fmt.Sprintf("not a valid Python file: %s", filePath), nil)
	}

	// Check if file exists through abstraction
	exists, err := uc.fileReader.FileExists(filePath)
	if err != nil {
		return domain.NewFileNotFoundError(filePath, err)
	}
	if !exists {
		return domain.NewFileNotFoundError(filePath, fmt.Errorf("file does not exist"))
	}

	// Load configuration if specified
	finalReq, err := uc.loadAndMergeConfig(req)
	if err != nil {
		return domain.NewConfigError("failed to load configuration", err)
	}

	// Analyze the file
	fileResult, err := uc.service.AnalyzeFile(ctx, filePath, finalReq)
	if err != nil {
		return domain.NewAnalysisError("mock data analysis failed", err)
	}

	// Create response
	response := &domain.MockDataResponse{
		Files: []domain.FileMockData{*fileResult},
		Summary: domain.MockDataSummary{
			TotalFiles:         1,
			TotalFindings:      fileResult.TotalFindings,
			FilesWithMockData:  0,
			ErrorFindings:      fileResult.ErrorCount,
			WarningFindings:    fileResult.WarningCount,
			InfoFindings:       fileResult.InfoCount,
		},
	}
	if fileResult.HasFindings() {
		response.Summary.FilesWithMockData = 1
	}
	response.Summary.CalculateTypeCounts(response.Files)

	// Delegate output handling to ReportWriter
	var out io.Writer
	if finalReq.OutputPath == "" {
		out = finalReq.OutputWriter
	}
	if err := uc.output.Write(out, finalReq.OutputPath, finalReq.OutputFormat, finalReq.NoOpen, func(w io.Writer) error {
		return uc.formatter.Write(response, finalReq.OutputFormat, w)
	}); err != nil {
		return domain.NewOutputError("failed to write output", err)
	}

	return nil
}

// validateRequest validates the mock data request
func (uc *MockDataUseCase) validateRequest(req domain.MockDataRequest) error {
	if len(req.Paths) == 0 {
		return fmt.Errorf("at least one path must be specified")
	}
	return nil
}

// loadAndMergeConfig loads and merges configuration with CLI flags
func (uc *MockDataUseCase) loadAndMergeConfig(req domain.MockDataRequest) (domain.MockDataRequest, error) {
	if uc.configLoader == nil {
		return req, nil
	}

	var configReq *domain.MockDataRequest

	if req.ConfigPath != "" {
		// Load from specified file
		var err error
		configReq, err = uc.configLoader.LoadConfig(req.ConfigPath)
		if err != nil {
			return req, fmt.Errorf("failed to load config from %s: %w", req.ConfigPath, err)
		}
	} else {
		// Try to load default config
		configReq = uc.configLoader.LoadDefaultConfig()
	}

	if configReq != nil {
		// Request takes precedence (CLI overrides config file)
		merged := uc.configLoader.MergeConfig(configReq, &req)
		return *merged, nil
	}

	return req, nil
}

// MockDataUseCaseBuilder provides a builder pattern for constructing MockDataUseCase
type MockDataUseCaseBuilder struct {
	service      domain.MockDataService
	fileReader   domain.FileReader
	formatter    domain.MockDataFormatter
	configLoader domain.MockDataConfigurationLoader
	output       domain.ReportWriter
}

// NewMockDataUseCaseBuilder creates a new builder
func NewMockDataUseCaseBuilder() *MockDataUseCaseBuilder {
	return &MockDataUseCaseBuilder{}
}

// WithService sets the mock data service
func (b *MockDataUseCaseBuilder) WithService(service domain.MockDataService) *MockDataUseCaseBuilder {
	b.service = service
	return b
}

// WithFileReader sets the file reader
func (b *MockDataUseCaseBuilder) WithFileReader(fileReader domain.FileReader) *MockDataUseCaseBuilder {
	b.fileReader = fileReader
	return b
}

// WithFormatter sets the output formatter
func (b *MockDataUseCaseBuilder) WithFormatter(formatter domain.MockDataFormatter) *MockDataUseCaseBuilder {
	b.formatter = formatter
	return b
}

// WithConfigLoader sets the configuration loader
func (b *MockDataUseCaseBuilder) WithConfigLoader(configLoader domain.MockDataConfigurationLoader) *MockDataUseCaseBuilder {
	b.configLoader = configLoader
	return b
}

// WithOutput sets the report writer
func (b *MockDataUseCaseBuilder) WithOutput(output domain.ReportWriter) *MockDataUseCaseBuilder {
	b.output = output
	return b
}

// Build creates the MockDataUseCase with validation
func (b *MockDataUseCaseBuilder) Build() (*MockDataUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("service is required")
	}
	if b.fileReader == nil {
		return nil, fmt.Errorf("file reader is required")
	}
	if b.formatter == nil {
		return nil, fmt.Errorf("formatter is required")
	}

	uc := &MockDataUseCase{
		service:      b.service,
		fileReader:   b.fileReader,
		formatter:    b.formatter,
		configLoader: b.configLoader,
	}

	if b.output != nil {
		uc.output = b.output
	} else {
		uc.output = svc.NewFileOutputWriter(nil)
	}

	return uc, nil
}

// BuildWithDefaults creates the MockDataUseCase with default implementations
func (b *MockDataUseCaseBuilder) BuildWithDefaults() (*MockDataUseCase, error) {
	if b.service == nil {
		b.service = svc.NewMockDataService()
	}
	if b.fileReader == nil {
		b.fileReader = svc.NewFileReader()
	}
	if b.formatter == nil {
		b.formatter = svc.NewMockDataFormatter()
	}
	if b.configLoader == nil {
		b.configLoader = svc.NewMockDataConfigurationLoader()
	}
	if b.output == nil {
		b.output = svc.NewFileOutputWriter(nil)
	}

	return &MockDataUseCase{
		service:      b.service,
		fileReader:   b.fileReader,
		formatter:    b.formatter,
		configLoader: b.configLoader,
		output:       b.output,
	}, nil
}
