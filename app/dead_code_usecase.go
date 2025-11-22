package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	svc "github.com/ludo-technologies/pyscn/service"
)

// DeadCodeUseCase orchestrates the dead code analysis workflow
type DeadCodeUseCase struct {
	service      domain.DeadCodeService
	fileReader   domain.FileReader
	formatter    domain.DeadCodeFormatter
	configLoader domain.DeadCodeConfigurationLoader
	output       domain.ReportWriter
}

// NewDeadCodeUseCase creates a new dead code use case
func NewDeadCodeUseCase(
	service domain.DeadCodeService,
	fileReader domain.FileReader,
	formatter domain.DeadCodeFormatter,
	configLoader domain.DeadCodeConfigurationLoader,
) *DeadCodeUseCase {
	return &DeadCodeUseCase{
		service:      service,
		fileReader:   fileReader,
		formatter:    formatter,
		configLoader: configLoader,
		output:       svc.NewFileOutputWriter(nil),
	}
}

// Execute performs the complete dead code analysis workflow
func (uc *DeadCodeUseCase) Execute(ctx context.Context, req domain.DeadCodeRequest) error {
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
		return domain.NewAnalysisError("dead code analysis failed", err)
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

// AnalyzeAndReturn performs dead code analysis and returns the response without formatting
func (uc *DeadCodeUseCase) AnalyzeAndReturn(ctx context.Context, req domain.DeadCodeRequest) (*domain.DeadCodeResponse, error) {
	// Validate input
	if err := uc.validateRequest(req); err != nil {
		return nil, domain.NewInvalidInputError("invalid request", err)
	}

	// Load configuration if specified
	finalReq, err := uc.loadAndMergeConfig(req)
	if err != nil {
		return nil, domain.NewConfigError("failed to load configuration", err)
	}

	// Resolve file paths (use helper to avoid duplication)
	files, err := ResolveFilePaths(
		uc.fileReader,
		finalReq.Paths,
		finalReq.Recursive,
		finalReq.IncludePatterns,
		finalReq.ExcludePatterns,
		false, // validatePythonFile: dead code doesn't need strict Python validation
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
		return nil, domain.NewAnalysisError("dead code analysis failed", err)
	}

	return response, nil
}

// AnalyzeFile analyzes a single file for dead code
func (uc *DeadCodeUseCase) AnalyzeFile(ctx context.Context, filePath string, req domain.DeadCodeRequest) error {
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

	// Perform analysis
	fileResult, err := uc.service.AnalyzeFile(ctx, filePath, finalReq)
	if err != nil {
		return domain.NewAnalysisError("file analysis failed", err)
	}

	// Create a response with the single file
	response := &domain.DeadCodeResponse{
		Files: []domain.FileDeadCode{*fileResult},
		Summary: domain.DeadCodeSummary{
			TotalFiles:            1,
			FilesWithDeadCode:     1,
			TotalFindings:         fileResult.TotalFindings,
			TotalFunctions:        fileResult.TotalFunctions,
			FunctionsWithDeadCode: fileResult.AffectedFunctions,
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	// Delegate output handling to ReportWriter
	var out2 io.Writer
	if finalReq.OutputPath == "" {
		out2 = finalReq.OutputWriter
	}
	if err := uc.output.Write(out2, finalReq.OutputPath, finalReq.OutputFormat, finalReq.NoOpen, func(w io.Writer) error {
		return uc.formatter.Write(response, finalReq.OutputFormat, w)
	}); err != nil {
		return domain.NewOutputError("failed to write output", err)
	}

	return nil
}

// AnalyzeFunction analyzes a single function for dead code
func (uc *DeadCodeUseCase) AnalyzeFunction(ctx context.Context, functionCFG interface{}, req domain.DeadCodeRequest) (*domain.FunctionDeadCode, error) {
	// Validate input
	if functionCFG == nil {
		return nil, domain.NewInvalidInputError("function CFG cannot be nil", nil)
	}

	// Load configuration if specified
	finalReq, err := uc.loadAndMergeConfig(req)
	if err != nil {
		return nil, domain.NewConfigError("failed to load configuration", err)
	}

	// Perform analysis
	result, err := uc.service.AnalyzeFunction(ctx, functionCFG, finalReq)
	if err != nil {
		return nil, domain.NewAnalysisError("function analysis failed", err)
	}

	return result, nil
}

// validateRequest validates the dead code request
func (uc *DeadCodeUseCase) validateRequest(req domain.DeadCodeRequest) error {
	if len(req.Paths) == 0 {
		return fmt.Errorf("no input paths specified")
	}

	if req.OutputWriter == nil && req.OutputPath == "" {
		return fmt.Errorf("output writer or output path is required")
	}

	if req.ContextLines < 0 {
		return fmt.Errorf("context lines cannot be negative")
	}

	// Validate severity level
	switch req.MinSeverity {
	case domain.DeadCodeSeverityCritical, domain.DeadCodeSeverityWarning, domain.DeadCodeSeverityInfo:
		// Valid severities
	default:
		return fmt.Errorf("unsupported minimum severity: %s", req.MinSeverity)
	}

	// Validate output format
	switch req.OutputFormat {
	case domain.OutputFormatText, domain.OutputFormatJSON, domain.OutputFormatYAML, domain.OutputFormatCSV, domain.OutputFormatHTML:
		// Valid formats
	default:
		return fmt.Errorf("unsupported output format: %s", req.OutputFormat)
	}

	// Validate sort criteria
	switch req.SortBy {
	case domain.DeadCodeSortBySeverity, domain.DeadCodeSortByLine, domain.DeadCodeSortByFile, domain.DeadCodeSortByFunction:
		// Valid criteria
	default:
		return fmt.Errorf("unsupported sort criteria: %s", req.SortBy)
	}

	return nil
}

// loadAndMergeConfig loads configuration from file and merges with request
func (uc *DeadCodeUseCase) loadAndMergeConfig(req domain.DeadCodeRequest) (domain.DeadCodeRequest, error) {
	if uc.configLoader == nil {
		return req, nil
	}

	var configReq *domain.DeadCodeRequest
	var err error

	if req.ConfigPath != "" {
		// Load from specified config file
		configReq, err = uc.configLoader.LoadConfig(req.ConfigPath)
		if err != nil {
			return req, fmt.Errorf("failed to load config from %s: %w", req.ConfigPath, err)
		}
	} else {
		// Try to load default config
		configReq = uc.configLoader.LoadDefaultConfig()
	}

	if configReq != nil {
		// Merge config with request (request takes precedence)
		merged := uc.configLoader.MergeConfig(configReq, &req)
		return *merged, nil
	}

	return req, nil
}

// DeadCodeUseCaseBuilder provides a builder pattern for creating DeadCodeUseCase
type DeadCodeUseCaseBuilder struct {
	service      domain.DeadCodeService
	fileReader   domain.FileReader
	formatter    domain.DeadCodeFormatter
	configLoader domain.DeadCodeConfigurationLoader
	output       domain.ReportWriter
}

// NewDeadCodeUseCaseBuilder creates a new builder
func NewDeadCodeUseCaseBuilder() *DeadCodeUseCaseBuilder {
	return &DeadCodeUseCaseBuilder{}
}

// WithService sets the dead code service
func (b *DeadCodeUseCaseBuilder) WithService(service domain.DeadCodeService) *DeadCodeUseCaseBuilder {
	b.service = service
	return b
}

// WithFileReader sets the file reader
func (b *DeadCodeUseCaseBuilder) WithFileReader(fileReader domain.FileReader) *DeadCodeUseCaseBuilder {
	b.fileReader = fileReader
	return b
}

// WithFormatter sets the output formatter
func (b *DeadCodeUseCaseBuilder) WithFormatter(formatter domain.DeadCodeFormatter) *DeadCodeUseCaseBuilder {
	b.formatter = formatter
	return b
}

// WithConfigLoader sets the configuration loader
func (b *DeadCodeUseCaseBuilder) WithConfigLoader(configLoader domain.DeadCodeConfigurationLoader) *DeadCodeUseCaseBuilder {
	b.configLoader = configLoader
	return b
}

// WithOutputWriter sets the report writer
func (b *DeadCodeUseCaseBuilder) WithOutputWriter(output domain.ReportWriter) *DeadCodeUseCaseBuilder {
	b.output = output
	return b
}

// Build creates the DeadCodeUseCase with the configured dependencies
func (b *DeadCodeUseCaseBuilder) Build() (*DeadCodeUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("dead code service is required")
	}
	if b.fileReader == nil {
		return nil, fmt.Errorf("file reader is required")
	}
	if b.formatter == nil {
		return nil, fmt.Errorf("output formatter is required")
	}

	// Provide sensible defaults for optional dependencies
	if b.configLoader == nil {
		// ConfigLoader is optional - will skip config loading if nil
		b.configLoader = nil
	}

	uc := NewDeadCodeUseCase(
		b.service,
		b.fileReader,
		b.formatter,
		b.configLoader,
	)
	if b.output != nil {
		uc.output = b.output
	}
	return uc, nil
}

// BuildWithDefaults creates the DeadCodeUseCase with default implementations for optional dependencies
func (b *DeadCodeUseCaseBuilder) BuildWithDefaults() (*DeadCodeUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("dead code service is required")
	}
	if b.fileReader == nil {
		return nil, fmt.Errorf("file reader is required")
	}
	if b.formatter == nil {
		return nil, fmt.Errorf("output formatter is required")
	}

	// Provide default implementations for optional dependencies
	if b.configLoader == nil {
		// Create a no-op config loader that returns nil
		b.configLoader = &noOpDeadCodeConfigLoader{}
	}

	uc := NewDeadCodeUseCase(
		b.service,
		b.fileReader,
		b.formatter,
		b.configLoader,
	)
	if b.output != nil {
		uc.output = b.output
	}
	return uc, nil
}

// noOpDeadCodeConfigLoader is a no-op implementation of DeadCodeConfigurationLoader
type noOpDeadCodeConfigLoader struct{}

func (n *noOpDeadCodeConfigLoader) LoadConfig(path string) (*domain.DeadCodeRequest, error) {
	return nil, nil
}

func (n *noOpDeadCodeConfigLoader) LoadDefaultConfig() *domain.DeadCodeRequest {
	return nil
}

func (n *noOpDeadCodeConfigLoader) MergeConfig(base *domain.DeadCodeRequest, override *domain.DeadCodeRequest) *domain.DeadCodeRequest {
	return override
}

// DeadCodeUseCaseOptions provides configuration options for the dead code use case
type DeadCodeUseCaseOptions struct {
	EnableProgress   bool
	ProgressInterval time.Duration
	MaxConcurrency   int
	TimeoutPerFile   time.Duration
	ShowContext      bool
	ContextLines     int
}

// DefaultDeadCodeUseCaseOptions returns default options
func DefaultDeadCodeUseCaseOptions() DeadCodeUseCaseOptions {
	return DeadCodeUseCaseOptions{
		EnableProgress:   true,
		ProgressInterval: 100 * time.Millisecond,
		MaxConcurrency:   4,
		TimeoutPerFile:   30 * time.Second,
		ShowContext:      false,
		ContextLines:     3,
	}
}

// ExecuteWithOptions performs dead code analysis with custom options
func (uc *DeadCodeUseCase) ExecuteWithOptions(ctx context.Context, req domain.DeadCodeRequest, options DeadCodeUseCaseOptions) error {
	// Apply options to request
	req.ShowContext = domain.BoolPtr(options.ShowContext)
	req.ContextLines = options.ContextLines

	// Create context with timeout
	if options.TimeoutPerFile > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.TimeoutPerFile*time.Duration(len(req.Paths)))
		defer cancel()
	}

	return uc.Execute(ctx, req)
}

// QuickAnalysis performs a quick dead code analysis with minimal configuration
func (uc *DeadCodeUseCase) QuickAnalysis(ctx context.Context, filePaths []string, outputWriter *os.File) error {
	req := domain.DeadCodeRequest{
		Paths:           filePaths,
		OutputFormat:    domain.OutputFormatText,
		OutputWriter:    outputWriter,
		MinSeverity:     domain.DeadCodeSeverityWarning,
		SortBy:          domain.DeadCodeSortBySeverity,
		ShowContext:     domain.BoolPtr(false),
		ContextLines:    0,
		Recursive:       false,
		IncludePatterns: []string{"**/*.py"},
		ExcludePatterns: []string{},
		IgnorePatterns:  []string{},

		// Enable all detection types
		DetectAfterReturn:         domain.BoolPtr(true),
		DetectAfterBreak:          domain.BoolPtr(true),
		DetectAfterContinue:       domain.BoolPtr(true),
		DetectAfterRaise:          domain.BoolPtr(true),
		DetectUnreachableBranches: domain.BoolPtr(true),
	}

	return uc.Execute(ctx, req)
}

// ValidateConfiguration validates a dead code configuration
func (uc *DeadCodeUseCase) ValidateConfiguration(req domain.DeadCodeRequest) error {
	return uc.validateRequest(req)
}

// GetSupportedFormats returns the list of supported output formats
func (uc *DeadCodeUseCase) GetSupportedFormats() []domain.OutputFormat {
	return []domain.OutputFormat{
		domain.OutputFormatText,
		domain.OutputFormatJSON,
		domain.OutputFormatYAML,
		domain.OutputFormatCSV,
	}
}

// GetSupportedSortCriteria returns the list of supported sort criteria
func (uc *DeadCodeUseCase) GetSupportedSortCriteria() []domain.DeadCodeSortCriteria {
	return []domain.DeadCodeSortCriteria{
		domain.DeadCodeSortBySeverity,
		domain.DeadCodeSortByLine,
		domain.DeadCodeSortByFile,
		domain.DeadCodeSortByFunction,
	}
}

// GetSupportedSeverityLevels returns the list of supported severity levels
func (uc *DeadCodeUseCase) GetSupportedSeverityLevels() []domain.DeadCodeSeverity {
	return []domain.DeadCodeSeverity{
		domain.DeadCodeSeverityInfo,
		domain.DeadCodeSeverityWarning,
		domain.DeadCodeSeverityCritical,
	}
}
