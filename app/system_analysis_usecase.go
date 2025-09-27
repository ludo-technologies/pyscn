package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ludo-technologies/pyscn/domain"
	svc "github.com/ludo-technologies/pyscn/service"
)

// SystemAnalysisUseCase orchestrates the system analysis workflow
type SystemAnalysisUseCase struct {
	service      domain.SystemAnalysisService
	fileReader   domain.FileReader
	formatter    domain.SystemAnalysisOutputFormatter
	configLoader domain.SystemAnalysisConfigurationLoader
	output       domain.ReportWriter
}

// NewSystemAnalysisUseCase creates a new system analysis use case
func NewSystemAnalysisUseCase(
	service domain.SystemAnalysisService,
	fileReader domain.FileReader,
	formatter domain.SystemAnalysisOutputFormatter,
	configLoader domain.SystemAnalysisConfigurationLoader,
) *SystemAnalysisUseCase {
	return &SystemAnalysisUseCase{
		service:      service,
		fileReader:   fileReader,
		formatter:    formatter,
		configLoader: configLoader,
		output:       svc.NewFileOutputWriter(nil),
	}
}

// prepareAnalysis handles common preparation steps for analysis
func (uc *SystemAnalysisUseCase) prepareAnalysis(ctx context.Context, req domain.SystemAnalysisRequest) (domain.SystemAnalysisRequest, error) {
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

	// Update request with collected files
	finalReq.Paths = files
	return finalReq, nil
}

// Execute performs the complete system analysis workflow
func (uc *SystemAnalysisUseCase) Execute(ctx context.Context, req domain.SystemAnalysisRequest) error {
	// Prepare for analysis
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return err
	}

	// Perform analysis
	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return domain.NewAnalysisError("system analysis failed", err)
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

// AnalyzeAndReturn performs system analysis and returns the response without formatting
func (uc *SystemAnalysisUseCase) AnalyzeAndReturn(ctx context.Context, req domain.SystemAnalysisRequest) (*domain.SystemAnalysisResponse, error) {
	// Prepare for analysis
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return nil, err
	}

	// Perform analysis and return the response
	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return nil, domain.NewAnalysisError("system analysis failed", err)
	}

	return response, nil
}

// AnalyzeDependenciesOnly performs dependency analysis only
func (uc *SystemAnalysisUseCase) AnalyzeDependenciesOnly(ctx context.Context, req domain.SystemAnalysisRequest) (*domain.DependencyAnalysisResult, error) {
	// Prepare for analysis
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return nil, err
	}

	// Perform dependency analysis only
	result, err := uc.service.AnalyzeDependencies(ctx, finalReq)
	if err != nil {
		return nil, domain.NewAnalysisError("dependency analysis failed", err)
	}

	return result, nil
}

// AnalyzeArchitectureOnly performs architecture analysis only
func (uc *SystemAnalysisUseCase) AnalyzeArchitectureOnly(ctx context.Context, req domain.SystemAnalysisRequest) (*domain.ArchitectureAnalysisResult, error) {
	// Prepare for analysis
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return nil, err
	}

	// Perform architecture analysis only
	result, err := uc.service.AnalyzeArchitecture(ctx, finalReq)
	if err != nil {
		return nil, domain.NewAnalysisError("architecture analysis failed", err)
	}

	return result, nil
}

// validateRequest validates the analysis request
func (uc *SystemAnalysisUseCase) validateRequest(req domain.SystemAnalysisRequest) error {
	// Validate paths
	if err := uc.validatePaths(req); err != nil {
		return err
	}

	// Validate output
	if err := uc.validateOutput(req); err != nil {
		return err
	}

	// Validate thresholds
	if err := uc.validateThresholds(req); err != nil {
		return err
	}

	// Validate analysis options
	if err := uc.validateAnalysisOptions(req); err != nil {
		return err
	}

	return nil
}

// validatePaths validates input paths
func (uc *SystemAnalysisUseCase) validatePaths(req domain.SystemAnalysisRequest) error {
	if len(req.Paths) == 0 {
		return fmt.Errorf("no input paths specified")
	}
	return nil
}

// validateOutput validates output configuration
func (uc *SystemAnalysisUseCase) validateOutput(req domain.SystemAnalysisRequest) error {
	if req.OutputWriter == nil && req.OutputPath == "" {
		return fmt.Errorf("output writer or output path is required")
	}
	return nil
}

// validateThresholds validates threshold parameters
func (uc *SystemAnalysisUseCase) validateThresholds(req domain.SystemAnalysisRequest) error {
	// All threshold fields have been removed
	return nil
}

// validateAnalysisOptions validates analysis type options
func (uc *SystemAnalysisUseCase) validateAnalysisOptions(req domain.SystemAnalysisRequest) error {
	// At least one analysis type must be enabled
	if !req.AnalyzeDependencies && !req.AnalyzeArchitecture {
		return fmt.Errorf("at least one analysis type must be enabled")
	}

	// Validate output format
	validFormats := map[domain.OutputFormat]bool{
		domain.OutputFormatText: true,
		domain.OutputFormatJSON: true,
		domain.OutputFormatYAML: true,
		domain.OutputFormatCSV:  true,
		domain.OutputFormatHTML: true,
		"dot":                   true, // Special DOT format
	}

	if !validFormats[req.OutputFormat] {
		return fmt.Errorf("unsupported output format: %s", req.OutputFormat)
	}

	return nil
}

// loadAndMergeConfig loads and merges configuration
func (uc *SystemAnalysisUseCase) loadAndMergeConfig(req domain.SystemAnalysisRequest) (domain.SystemAnalysisRequest, error) {
	if uc.configLoader == nil {
		// No config loader available, return request as-is
		return req, nil
	}

	var baseConfig *domain.SystemAnalysisRequest
	var err error

	// Load configuration from file if specified
	if req.ConfigPath != "" {
		baseConfig, err = uc.configLoader.LoadConfig(req.ConfigPath)
		if err != nil {
			return req, fmt.Errorf("failed to load config from %s: %w", req.ConfigPath, err)
		}
	} else {
		// Load default configuration
		baseConfig = uc.configLoader.LoadDefaultConfig()
	}

	// Merge with request (request takes precedence)
	mergedConfig := uc.configLoader.MergeConfig(baseConfig, &req)
	return *mergedConfig, nil
}

// SystemAnalysisUseCaseBuilder provides a builder pattern for creating SystemAnalysisUseCase
type SystemAnalysisUseCaseBuilder struct {
	service      domain.SystemAnalysisService
	fileReader   domain.FileReader
	formatter    domain.SystemAnalysisOutputFormatter
	configLoader domain.SystemAnalysisConfigurationLoader
	output       domain.ReportWriter
}

// NewSystemAnalysisUseCaseBuilder creates a new builder
func NewSystemAnalysisUseCaseBuilder() *SystemAnalysisUseCaseBuilder {
	return &SystemAnalysisUseCaseBuilder{}
}

// WithService sets the system analysis service
func (b *SystemAnalysisUseCaseBuilder) WithService(service domain.SystemAnalysisService) *SystemAnalysisUseCaseBuilder {
	b.service = service
	return b
}

// WithFileReader sets the file reader
func (b *SystemAnalysisUseCaseBuilder) WithFileReader(fileReader domain.FileReader) *SystemAnalysisUseCaseBuilder {
	b.fileReader = fileReader
	return b
}

// WithFormatter sets the output formatter
func (b *SystemAnalysisUseCaseBuilder) WithFormatter(formatter domain.SystemAnalysisOutputFormatter) *SystemAnalysisUseCaseBuilder {
	b.formatter = formatter
	return b
}

// WithConfigLoader sets the configuration loader
func (b *SystemAnalysisUseCaseBuilder) WithConfigLoader(configLoader domain.SystemAnalysisConfigurationLoader) *SystemAnalysisUseCaseBuilder {
	b.configLoader = configLoader
	return b
}

// WithOutputWriter sets the report writer
func (b *SystemAnalysisUseCaseBuilder) WithOutputWriter(output domain.ReportWriter) *SystemAnalysisUseCaseBuilder {
	b.output = output
	return b
}

// Build creates the SystemAnalysisUseCase with the configured dependencies
func (b *SystemAnalysisUseCaseBuilder) Build() (*SystemAnalysisUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("system analysis service is required")
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

	uc := NewSystemAnalysisUseCase(
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

// BuildWithDefaults creates the SystemAnalysisUseCase with default implementations for optional dependencies
func (b *SystemAnalysisUseCaseBuilder) BuildWithDefaults() (*SystemAnalysisUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("system analysis service is required")
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
		b.configLoader = &noOpSystemAnalysisConfigLoader{}
	}

	uc := NewSystemAnalysisUseCase(
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

// noOpSystemAnalysisConfigLoader is a no-op implementation of SystemAnalysisConfigurationLoader
type noOpSystemAnalysisConfigLoader struct{}

func (n *noOpSystemAnalysisConfigLoader) LoadConfig(path string) (*domain.SystemAnalysisRequest, error) {
	return domain.DefaultSystemAnalysisRequest(), nil
}

func (n *noOpSystemAnalysisConfigLoader) LoadDefaultConfig() *domain.SystemAnalysisRequest {
	return domain.DefaultSystemAnalysisRequest()
}

func (n *noOpSystemAnalysisConfigLoader) MergeConfig(base *domain.SystemAnalysisRequest, override *domain.SystemAnalysisRequest) *domain.SystemAnalysisRequest {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}

	// Simple merge - override takes precedence
	merged := *base

	// Override non-zero values
	if len(override.Paths) > 0 {
		merged.Paths = override.Paths
	}
	if override.OutputFormat != "" {
		merged.OutputFormat = override.OutputFormat
	}
	if override.OutputWriter != nil {
		merged.OutputWriter = override.OutputWriter
	}
	if override.OutputPath != "" {
		merged.OutputPath = override.OutputPath
	}
	if override.ConfigPath != "" {
		merged.ConfigPath = override.ConfigPath
	}

	// Boolean overrides
	merged.NoOpen = override.NoOpen
	merged.AnalyzeDependencies = override.AnalyzeDependencies
	merged.AnalyzeArchitecture = override.AnalyzeArchitecture
	merged.IncludeStdLib = override.IncludeStdLib
	merged.IncludeThirdParty = override.IncludeThirdParty
	merged.FollowRelative = override.FollowRelative
	merged.DetectCycles = override.DetectCycles
	merged.Recursive = override.Recursive

	// (Numeric overrides removed - fields no longer exist)

	// String slices
	if len(override.IncludePatterns) > 0 {
		merged.IncludePatterns = override.IncludePatterns
	}
	if len(override.ExcludePatterns) > 0 {
		merged.ExcludePatterns = override.ExcludePatterns
	}

	return &merged
}
