package app

import (
    "context"
    "fmt"
    "io"
    "time"

    "github.com/ludo-technologies/pyscn/domain"
    svc "github.com/ludo-technologies/pyscn/service"
)

// ComplexityUseCase orchestrates the complexity analysis workflow
type ComplexityUseCase struct {
    service      domain.ComplexityService
    fileReader   domain.FileReader
    formatter    domain.OutputFormatter
    configLoader domain.ConfigurationLoader
    output       domain.ReportWriter
}

// NewComplexityUseCase creates a new complexity use case
func NewComplexityUseCase(
    service domain.ComplexityService,
    fileReader domain.FileReader,
    formatter domain.OutputFormatter,
    configLoader domain.ConfigurationLoader,
) *ComplexityUseCase {
    return &ComplexityUseCase{
        service:      service,
        fileReader:   fileReader,
        formatter:    formatter,
        configLoader: configLoader,
        output:       svc.NewFileOutputWriter(nil),
    }
}

// Execute performs the complete complexity analysis workflow
func (uc *ComplexityUseCase) Execute(ctx context.Context, req domain.ComplexityRequest) error {
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

	// Progress reporting removed - not meaningful for file parsing

	// Update request with collected files
	finalReq.Paths = files

	// Perform analysis
	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return domain.NewAnalysisError("complexity analysis failed", err)
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

// AnalyzeAndReturn performs complexity analysis and returns the response without formatting
func (uc *ComplexityUseCase) AnalyzeAndReturn(ctx context.Context, req domain.ComplexityRequest) (*domain.ComplexityResponse, error) {
	// Validate input
	if err := uc.validateRequest(req); err != nil {
		return nil, domain.NewInvalidInputError("invalid request", err)
	}

	// Load configuration if specified
	finalReq, err := uc.loadAndMergeConfig(req)
	if err != nil {
		return nil, domain.NewConfigError("failed to load configuration", err)
	}

	// Collect Python files
	files, err := uc.fileReader.CollectPythonFiles(
		finalReq.Paths,
		finalReq.Recursive,
		finalReq.IncludePatterns,
		finalReq.ExcludePatterns,
	)
	if err != nil {
		return nil, domain.NewFileNotFoundError("failed to collect files", err)
	}

	if len(files) == 0 {
		return nil, domain.NewInvalidInputError("no Python files found in the specified paths", nil)
	}

	// Progress reporting removed - not meaningful for file parsing

	// Update request with collected files
	finalReq.Paths = files

	// Perform analysis and return the response
	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return nil, domain.NewAnalysisError("complexity analysis failed", err)
	}

	return response, nil
}

// AnalyzeFile analyzes a single file
func (uc *ComplexityUseCase) AnalyzeFile(ctx context.Context, filePath string, req domain.ComplexityRequest) error {
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
	response, err := uc.service.AnalyzeFile(ctx, filePath, finalReq)
	if err != nil {
		return domain.NewAnalysisError("file analysis failed", err)
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

// validateRequest validates the complexity request
func (uc *ComplexityUseCase) validateRequest(req domain.ComplexityRequest) error {
	if len(req.Paths) == 0 {
		return fmt.Errorf("no input paths specified")
	}

	if req.OutputWriter == nil && req.OutputPath == "" {
		return fmt.Errorf("output writer or output path is required")
	}

	if req.MinComplexity < 0 {
		return fmt.Errorf("minimum complexity cannot be negative")
	}

	if req.MaxComplexity < 0 {
		return fmt.Errorf("maximum complexity cannot be negative")
	}

	if req.MaxComplexity > 0 && req.MinComplexity > req.MaxComplexity {
		return fmt.Errorf("minimum complexity cannot be greater than maximum complexity")
	}

	if req.LowThreshold <= 0 {
		return fmt.Errorf("low threshold must be positive")
	}

	if req.MediumThreshold <= req.LowThreshold {
		return fmt.Errorf("medium threshold must be greater than low threshold")
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
	case domain.SortByComplexity, domain.SortByName, domain.SortByRisk:
		// Valid criteria
	default:
		return fmt.Errorf("unsupported sort criteria: %s", req.SortBy)
	}

	return nil
}

// loadAndMergeConfig loads configuration from file and merges with request
func (uc *ComplexityUseCase) loadAndMergeConfig(req domain.ComplexityRequest) (domain.ComplexityRequest, error) {
	if uc.configLoader == nil {
		return req, nil
	}

	var configReq *domain.ComplexityRequest
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

// ComplexityUseCaseBuilder provides a builder pattern for creating ComplexityUseCase
type ComplexityUseCaseBuilder struct {
    service      domain.ComplexityService
    fileReader   domain.FileReader
    formatter    domain.OutputFormatter
    configLoader domain.ConfigurationLoader
    output       domain.ReportWriter
}

// NewComplexityUseCaseBuilder creates a new builder
func NewComplexityUseCaseBuilder() *ComplexityUseCaseBuilder {
	return &ComplexityUseCaseBuilder{}
}

// WithService sets the complexity service
func (b *ComplexityUseCaseBuilder) WithService(service domain.ComplexityService) *ComplexityUseCaseBuilder {
	b.service = service
	return b
}

// WithFileReader sets the file reader
func (b *ComplexityUseCaseBuilder) WithFileReader(fileReader domain.FileReader) *ComplexityUseCaseBuilder {
	b.fileReader = fileReader
	return b
}

// WithFormatter sets the output formatter
func (b *ComplexityUseCaseBuilder) WithFormatter(formatter domain.OutputFormatter) *ComplexityUseCaseBuilder {
	b.formatter = formatter
	return b
}

// WithConfigLoader sets the configuration loader
func (b *ComplexityUseCaseBuilder) WithConfigLoader(configLoader domain.ConfigurationLoader) *ComplexityUseCaseBuilder {
	b.configLoader = configLoader
	return b
}


// WithOutputWriter sets the report writer
func (b *ComplexityUseCaseBuilder) WithOutputWriter(output domain.ReportWriter) *ComplexityUseCaseBuilder {
    b.output = output
    return b
}

// Build creates the ComplexityUseCase with the configured dependencies
func (b *ComplexityUseCaseBuilder) Build() (*ComplexityUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("complexity service is required")
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

    uc := NewComplexityUseCase(
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

// BuildWithDefaults creates the ComplexityUseCase with default implementations for optional dependencies
func (b *ComplexityUseCaseBuilder) BuildWithDefaults() (*ComplexityUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("complexity service is required")
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
		b.configLoader = &noOpConfigLoader{}
	}

    uc := NewComplexityUseCase(
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

// noOpConfigLoader is a no-op implementation of ConfigurationLoader
type noOpConfigLoader struct{}

func (n *noOpConfigLoader) LoadConfig(path string) (*domain.ComplexityRequest, error) {
	return nil, nil
}

func (n *noOpConfigLoader) LoadDefaultConfig() *domain.ComplexityRequest {
	return nil
}

func (n *noOpConfigLoader) MergeConfig(base *domain.ComplexityRequest, override *domain.ComplexityRequest) *domain.ComplexityRequest {
	return override
}


// UseCaseOptions provides configuration options for the use case
type UseCaseOptions struct {
	EnableProgress   bool
	ProgressInterval time.Duration
	MaxConcurrency   int
	TimeoutPerFile   time.Duration
}

// DefaultUseCaseOptions returns default options
func DefaultUseCaseOptions() UseCaseOptions {
	return UseCaseOptions{
		EnableProgress:   true,
		ProgressInterval: 100 * time.Millisecond,
		MaxConcurrency:   4,
		TimeoutPerFile:   30 * time.Second,
	}
}
