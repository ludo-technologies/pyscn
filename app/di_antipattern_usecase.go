package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ludo-technologies/pyscn/domain"
	svc "github.com/ludo-technologies/pyscn/service"
)

// DIAntipatternUseCase orchestrates the DI anti-pattern analysis workflow
type DIAntipatternUseCase struct {
	service      domain.DIAntipatternService
	fileReader   domain.FileReader
	formatter    domain.DIAntipatternOutputFormatter
	configLoader domain.DIAntipatternConfigurationLoader
	output       domain.ReportWriter
}

// NewDIAntipatternUseCase creates a new DI anti-pattern use case
func NewDIAntipatternUseCase(
	service domain.DIAntipatternService,
	fileReader domain.FileReader,
	formatter domain.DIAntipatternOutputFormatter,
	configLoader domain.DIAntipatternConfigurationLoader,
) *DIAntipatternUseCase {
	return &DIAntipatternUseCase{
		service:      service,
		fileReader:   fileReader,
		formatter:    formatter,
		configLoader: configLoader,
		output:       svc.NewFileOutputWriter(nil),
	}
}

// prepareAnalysis handles common preparation steps for analysis
func (uc *DIAntipatternUseCase) prepareAnalysis(ctx context.Context, req domain.DIAntipatternRequest) (domain.DIAntipatternRequest, error) {
	// Validate input
	if err := uc.validateRequest(req); err != nil {
		return req, domain.NewInvalidInputError("invalid request", err)
	}

	// Load configuration if specified
	finalReq, err := uc.loadAndMergeConfig(req)
	if err != nil {
		return req, domain.NewConfigError("failed to load configuration", err)
	}

	// Resolve file paths
	files, err := ResolveFilePaths(
		uc.fileReader,
		finalReq.Paths,
		domain.BoolValue(finalReq.Recursive, true),
		finalReq.IncludePatterns,
		finalReq.ExcludePatterns,
		false,
	)
	if err != nil {
		return req, domain.NewFileNotFoundError("failed to collect files", err)
	}

	if len(files) == 0 {
		return req, domain.NewInvalidInputError("no Python files found in the specified paths", nil)
	}

	finalReq.Paths = files
	return finalReq, nil
}

// Execute performs the complete DI anti-pattern analysis workflow
func (uc *DIAntipatternUseCase) Execute(ctx context.Context, req domain.DIAntipatternRequest) error {
	// Prepare for analysis
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return err
	}

	// Perform analysis
	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return domain.NewAnalysisError("DI anti-pattern analysis failed", err)
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

// AnalyzeAndReturn performs DI anti-pattern analysis and returns the response without formatting
func (uc *DIAntipatternUseCase) AnalyzeAndReturn(ctx context.Context, req domain.DIAntipatternRequest) (*domain.DIAntipatternResponse, error) {
	// Prepare for analysis
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return nil, err
	}

	// Perform analysis and return the response
	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return nil, domain.NewAnalysisError("DI anti-pattern analysis failed", err)
	}

	return response, nil
}

// validateRequest validates the DI anti-pattern request
func (uc *DIAntipatternUseCase) validateRequest(req domain.DIAntipatternRequest) error {
	if len(req.Paths) == 0 {
		return fmt.Errorf("no input paths specified")
	}

	if req.OutputWriter == nil && req.OutputPath == "" {
		return fmt.Errorf("output writer or output path is required")
	}

	if req.ConstructorParamThreshold < 0 {
		return fmt.Errorf("constructor parameter threshold cannot be negative")
	}

	return nil
}

// loadAndMergeConfig loads configuration from file and merges with request
func (uc *DIAntipatternUseCase) loadAndMergeConfig(req domain.DIAntipatternRequest) (domain.DIAntipatternRequest, error) {
	if uc.configLoader == nil {
		return req, nil
	}

	var configReq *domain.DIAntipatternRequest
	var err error

	if req.ConfigPath != "" {
		configReq, err = uc.configLoader.LoadConfig(req.ConfigPath)
		if err != nil {
			return req, fmt.Errorf("failed to load config from %s: %w", req.ConfigPath, err)
		}
	} else {
		configReq = uc.configLoader.LoadDefaultConfig()
	}

	if configReq != nil {
		merged := uc.configLoader.MergeConfig(configReq, &req)
		return *merged, nil
	}

	return req, nil
}

// DIAntipatternUseCaseBuilder provides a builder pattern for creating DIAntipatternUseCase
type DIAntipatternUseCaseBuilder struct {
	service      domain.DIAntipatternService
	fileReader   domain.FileReader
	formatter    domain.DIAntipatternOutputFormatter
	configLoader domain.DIAntipatternConfigurationLoader
	output       domain.ReportWriter
}

// NewDIAntipatternUseCaseBuilder creates a new builder
func NewDIAntipatternUseCaseBuilder() *DIAntipatternUseCaseBuilder {
	return &DIAntipatternUseCaseBuilder{}
}

// WithService sets the DI anti-pattern service
func (b *DIAntipatternUseCaseBuilder) WithService(service domain.DIAntipatternService) *DIAntipatternUseCaseBuilder {
	b.service = service
	return b
}

// WithFileReader sets the file reader
func (b *DIAntipatternUseCaseBuilder) WithFileReader(fileReader domain.FileReader) *DIAntipatternUseCaseBuilder {
	b.fileReader = fileReader
	return b
}

// WithFormatter sets the output formatter
func (b *DIAntipatternUseCaseBuilder) WithFormatter(formatter domain.DIAntipatternOutputFormatter) *DIAntipatternUseCaseBuilder {
	b.formatter = formatter
	return b
}

// WithConfigLoader sets the configuration loader
func (b *DIAntipatternUseCaseBuilder) WithConfigLoader(configLoader domain.DIAntipatternConfigurationLoader) *DIAntipatternUseCaseBuilder {
	b.configLoader = configLoader
	return b
}

// WithOutputWriter sets the report writer
func (b *DIAntipatternUseCaseBuilder) WithOutputWriter(output domain.ReportWriter) *DIAntipatternUseCaseBuilder {
	b.output = output
	return b
}

// Build creates the DIAntipatternUseCase with the configured dependencies
func (b *DIAntipatternUseCaseBuilder) Build() (*DIAntipatternUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("DI anti-pattern service is required")
	}
	if b.fileReader == nil {
		return nil, fmt.Errorf("file reader is required")
	}
	if b.formatter == nil {
		return nil, fmt.Errorf("output formatter is required")
	}

	uc := NewDIAntipatternUseCase(
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

// BuildWithDefaults creates the DIAntipatternUseCase with default implementations for optional dependencies
func (b *DIAntipatternUseCaseBuilder) BuildWithDefaults() (*DIAntipatternUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("DI anti-pattern service is required")
	}
	if b.fileReader == nil {
		return nil, fmt.Errorf("file reader is required")
	}
	if b.formatter == nil {
		return nil, fmt.Errorf("output formatter is required")
	}

	if b.configLoader == nil {
		b.configLoader = &noOpDIAntipatternConfigLoader{}
	}

	uc := NewDIAntipatternUseCase(
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

// noOpDIAntipatternConfigLoader is a no-op implementation
type noOpDIAntipatternConfigLoader struct{}

func (n *noOpDIAntipatternConfigLoader) LoadConfig(path string) (*domain.DIAntipatternRequest, error) {
	return nil, nil
}

func (n *noOpDIAntipatternConfigLoader) LoadDefaultConfig() *domain.DIAntipatternRequest {
	return nil
}

func (n *noOpDIAntipatternConfigLoader) MergeConfig(base *domain.DIAntipatternRequest, override *domain.DIAntipatternRequest) *domain.DIAntipatternRequest {
	return override
}
