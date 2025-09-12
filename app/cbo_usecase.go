package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ludo-technologies/pyscn/domain"
	svc "github.com/ludo-technologies/pyscn/service"
)

// CBOUseCase orchestrates the CBO analysis workflow
type CBOUseCase struct {
	service      domain.CBOService
	fileReader   domain.FileReader
	formatter    domain.CBOOutputFormatter
	configLoader domain.CBOConfigurationLoader
	output       domain.ReportWriter
}

// NewCBOUseCase creates a new CBO use case
func NewCBOUseCase(
	service domain.CBOService,
	fileReader domain.FileReader,
	formatter domain.CBOOutputFormatter,
	configLoader domain.CBOConfigurationLoader,
) *CBOUseCase {
	return &CBOUseCase{
		service:      service,
		fileReader:   fileReader,
		formatter:    formatter,
		configLoader: configLoader,
		output:       svc.NewFileOutputWriter(nil),
	}
}

// prepareAnalysis handles common preparation steps for analysis
func (uc *CBOUseCase) prepareAnalysis(ctx context.Context, req domain.CBORequest) (domain.CBORequest, error) {
	// Validate input
	if err := uc.validateRequest(req); err != nil {
		return req, domain.NewInvalidInputError("invalid request", err)
	}

	// Load configuration if specified
	finalReq, err := uc.loadAndMergeConfig(req)
	if err != nil {
		return req, domain.NewConfigError("failed to load configuration", err)
	}

	// Collect Python files
	files, err := uc.fileReader.CollectPythonFiles(
		finalReq.Paths,
		finalReq.Recursive,
		finalReq.IncludePatterns,
		finalReq.ExcludePatterns,
	)
	if err != nil {
		return req, domain.NewFileNotFoundError("failed to collect files", err)
	}

	if len(files) == 0 {
		return req, domain.NewInvalidInputError("no Python files found in the specified paths", nil)
	}

	// Progress reporting removed - not meaningful for file parsing

	// Update request with collected files
	finalReq.Paths = files
	return finalReq, nil
}

// Execute performs the complete CBO analysis workflow
func (uc *CBOUseCase) Execute(ctx context.Context, req domain.CBORequest) error {
	// Prepare for analysis
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return err
	}
	// Progress reporting removed

	// Perform analysis
	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return domain.NewAnalysisError("CBO analysis failed", err)
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

// AnalyzeAndReturn performs CBO analysis and returns the response without formatting
func (uc *CBOUseCase) AnalyzeAndReturn(ctx context.Context, req domain.CBORequest) (*domain.CBOResponse, error) {
	// Prepare for analysis
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return nil, err
	}
	// Progress reporting removed

	// Perform analysis and return the response
	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return nil, domain.NewAnalysisError("CBO analysis failed", err)
	}

	return response, nil
}

// AnalyzeFile analyzes a single file
func (uc *CBOUseCase) AnalyzeFile(ctx context.Context, filePath string, req domain.CBORequest) error {
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

// validatePaths validates input paths
func (uc *CBOUseCase) validatePaths(req domain.CBORequest) error {
	if len(req.Paths) == 0 {
		return fmt.Errorf("no input paths specified")
	}
	return nil
}

// validateOutput validates output configuration
func (uc *CBOUseCase) validateOutput(req domain.CBORequest) error {
	if req.OutputWriter == nil && req.OutputPath == "" {
		return fmt.Errorf("output writer or output path is required")
	}
	return nil
}

// validateCBORange validates CBO range parameters
func (uc *CBOUseCase) validateCBORange(req domain.CBORequest) error {
	if req.MinCBO < 0 {
		return fmt.Errorf("minimum CBO cannot be negative")
	}

	if req.MaxCBO < 0 {
		return fmt.Errorf("maximum CBO cannot be negative")
	}

	if req.MaxCBO > 0 && req.MinCBO > req.MaxCBO {
		return fmt.Errorf("minimum CBO cannot be greater than maximum CBO")
	}

	return nil
}

// validateThresholds validates threshold parameters
func (uc *CBOUseCase) validateThresholds(req domain.CBORequest) error {
	if req.LowThreshold <= 0 {
		return fmt.Errorf("low threshold must be positive")
	}

	if req.MediumThreshold <= req.LowThreshold {
		return fmt.Errorf("medium threshold must be greater than low threshold")
	}

	return nil
}

// validateFormats validates output format and sort criteria
func (uc *CBOUseCase) validateFormats(req domain.CBORequest) error {
	// Validate output format
	switch req.OutputFormat {
	case domain.OutputFormatText, domain.OutputFormatJSON, domain.OutputFormatYAML, domain.OutputFormatCSV, domain.OutputFormatHTML:
		// Valid formats
	default:
		return fmt.Errorf("unsupported output format: %s", req.OutputFormat)
	}

	// Validate sort criteria
	switch req.SortBy {
	case domain.SortByCoupling, domain.SortByName, domain.SortByRisk, domain.SortByLocation:
		// Valid criteria for CBO
	default:
		return fmt.Errorf("unsupported sort criteria: %s", req.SortBy)
	}

	return nil
}

// validateRequest validates the CBO request
func (uc *CBOUseCase) validateRequest(req domain.CBORequest) error {
	validators := []func(domain.CBORequest) error{
		uc.validatePaths,
		uc.validateOutput,
		uc.validateCBORange,
		uc.validateThresholds,
		uc.validateFormats,
	}

	for _, validator := range validators {
		if err := validator(req); err != nil {
			return err
		}
	}

	return nil
}

// loadAndMergeConfig loads configuration from file and merges with request
func (uc *CBOUseCase) loadAndMergeConfig(req domain.CBORequest) (domain.CBORequest, error) {
	if uc.configLoader == nil {
		return req, nil
	}

	var configReq *domain.CBORequest
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

// CBOUseCaseBuilder provides a builder pattern for creating CBOUseCase
type CBOUseCaseBuilder struct {
	service      domain.CBOService
	fileReader   domain.FileReader
	formatter    domain.CBOOutputFormatter
	configLoader domain.CBOConfigurationLoader
	output       domain.ReportWriter
}

// NewCBOUseCaseBuilder creates a new builder
func NewCBOUseCaseBuilder() *CBOUseCaseBuilder {
	return &CBOUseCaseBuilder{}
}

// WithService sets the CBO service
func (b *CBOUseCaseBuilder) WithService(service domain.CBOService) *CBOUseCaseBuilder {
	b.service = service
	return b
}

// WithFileReader sets the file reader
func (b *CBOUseCaseBuilder) WithFileReader(fileReader domain.FileReader) *CBOUseCaseBuilder {
	b.fileReader = fileReader
	return b
}

// WithFormatter sets the output formatter
func (b *CBOUseCaseBuilder) WithFormatter(formatter domain.CBOOutputFormatter) *CBOUseCaseBuilder {
	b.formatter = formatter
	return b
}

// WithConfigLoader sets the configuration loader
func (b *CBOUseCaseBuilder) WithConfigLoader(configLoader domain.CBOConfigurationLoader) *CBOUseCaseBuilder {
	b.configLoader = configLoader
	return b
}

// WithOutputWriter sets the report writer
func (b *CBOUseCaseBuilder) WithOutputWriter(output domain.ReportWriter) *CBOUseCaseBuilder {
	b.output = output
	return b
}

// Build creates the CBOUseCase with the configured dependencies
func (b *CBOUseCaseBuilder) Build() (*CBOUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("CBO service is required")
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

	uc := NewCBOUseCase(
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

// BuildWithDefaults creates the CBOUseCase with default implementations for optional dependencies
func (b *CBOUseCaseBuilder) BuildWithDefaults() (*CBOUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("CBO service is required")
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
		b.configLoader = &noOpCBOConfigLoader{}
	}

	uc := NewCBOUseCase(
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

// noOpCBOConfigLoader is a no-op implementation of CBOConfigurationLoader
type noOpCBOConfigLoader struct{}

func (n *noOpCBOConfigLoader) LoadConfig(path string) (*domain.CBORequest, error) {
	return nil, nil
}

func (n *noOpCBOConfigLoader) LoadDefaultConfig() *domain.CBORequest {
	return nil
}

func (n *noOpCBOConfigLoader) MergeConfig(base *domain.CBORequest, override *domain.CBORequest) *domain.CBORequest {
	return override
}
