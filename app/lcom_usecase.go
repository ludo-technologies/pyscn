package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ludo-technologies/pyscn/domain"
	svc "github.com/ludo-technologies/pyscn/service"
)

// LCOMUseCase orchestrates the LCOM analysis workflow
type LCOMUseCase struct {
	service      domain.LCOMService
	fileReader   domain.FileReader
	formatter    domain.LCOMOutputFormatter
	configLoader domain.LCOMConfigurationLoader
	output       domain.ReportWriter
}

// NewLCOMUseCase creates a new LCOM use case
func NewLCOMUseCase(
	service domain.LCOMService,
	fileReader domain.FileReader,
	formatter domain.LCOMOutputFormatter,
	configLoader domain.LCOMConfigurationLoader,
) *LCOMUseCase {
	return &LCOMUseCase{
		service:      service,
		fileReader:   fileReader,
		formatter:    formatter,
		configLoader: configLoader,
		output:       svc.NewFileOutputWriter(nil),
	}
}

// prepareAnalysis handles common preparation steps for analysis.
// Config is merged before validation so that callers may leave threshold
// fields at zero and rely on the config file (or built-in defaults) to
// supply valid values.
func (uc *LCOMUseCase) prepareAnalysis(ctx context.Context, req domain.LCOMRequest) (domain.LCOMRequest, error) {
	finalReq, err := uc.loadAndMergeConfig(req)
	if err != nil {
		return req, domain.NewConfigError("failed to load configuration", err)
	}

	if err := uc.validateRequest(finalReq); err != nil {
		return req, domain.NewInvalidInputError("invalid request", err)
	}

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

// Execute performs the complete LCOM analysis workflow
func (uc *LCOMUseCase) Execute(ctx context.Context, req domain.LCOMRequest) error {
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return err
	}

	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return domain.NewAnalysisError("LCOM analysis failed", err)
	}

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

// AnalyzeAndReturn performs LCOM analysis and returns the response without formatting
func (uc *LCOMUseCase) AnalyzeAndReturn(ctx context.Context, req domain.LCOMRequest) (*domain.LCOMResponse, error) {
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return nil, err
	}

	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return nil, domain.NewAnalysisError("LCOM analysis failed", err)
	}

	return response, nil
}

// validateRequest validates the LCOM request
func (uc *LCOMUseCase) validateRequest(req domain.LCOMRequest) error {
	if len(req.Paths) == 0 {
		return fmt.Errorf("no input paths specified")
	}
	if req.OutputWriter == nil && req.OutputPath == "" {
		return fmt.Errorf("output writer or output path is required")
	}
	if req.MinLCOM < 0 {
		return fmt.Errorf("minimum LCOM cannot be negative")
	}
	if req.MaxLCOM < 0 {
		return fmt.Errorf("maximum LCOM cannot be negative")
	}
	if req.MaxLCOM > 0 && req.MinLCOM > req.MaxLCOM {
		return fmt.Errorf("minimum LCOM cannot be greater than maximum LCOM")
	}
	if req.LowThreshold <= 0 {
		return fmt.Errorf("low threshold must be positive")
	}
	if req.MediumThreshold <= req.LowThreshold {
		return fmt.Errorf("medium threshold must be greater than low threshold")
	}
	switch req.OutputFormat {
	case domain.OutputFormatText, domain.OutputFormatJSON, domain.OutputFormatYAML, domain.OutputFormatCSV, domain.OutputFormatHTML:
		// Valid formats
	default:
		return fmt.Errorf("unsupported output format: %s", req.OutputFormat)
	}
	return nil
}

// loadAndMergeConfig loads configuration from file and merges with request
func (uc *LCOMUseCase) loadAndMergeConfig(req domain.LCOMRequest) (domain.LCOMRequest, error) {
	if uc.configLoader == nil {
		return req, nil
	}

	var configReq *domain.LCOMRequest
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

// LCOMUseCaseBuilder provides a builder pattern for creating LCOMUseCase
type LCOMUseCaseBuilder struct {
	service      domain.LCOMService
	fileReader   domain.FileReader
	formatter    domain.LCOMOutputFormatter
	configLoader domain.LCOMConfigurationLoader
	output       domain.ReportWriter
}

// NewLCOMUseCaseBuilder creates a new builder
func NewLCOMUseCaseBuilder() *LCOMUseCaseBuilder {
	return &LCOMUseCaseBuilder{}
}

// WithService sets the LCOM service
func (b *LCOMUseCaseBuilder) WithService(service domain.LCOMService) *LCOMUseCaseBuilder {
	b.service = service
	return b
}

// WithFileReader sets the file reader
func (b *LCOMUseCaseBuilder) WithFileReader(fileReader domain.FileReader) *LCOMUseCaseBuilder {
	b.fileReader = fileReader
	return b
}

// WithFormatter sets the output formatter
func (b *LCOMUseCaseBuilder) WithFormatter(formatter domain.LCOMOutputFormatter) *LCOMUseCaseBuilder {
	b.formatter = formatter
	return b
}

// WithConfigLoader sets the configuration loader
func (b *LCOMUseCaseBuilder) WithConfigLoader(configLoader domain.LCOMConfigurationLoader) *LCOMUseCaseBuilder {
	b.configLoader = configLoader
	return b
}

// WithOutputWriter sets the report writer
func (b *LCOMUseCaseBuilder) WithOutputWriter(output domain.ReportWriter) *LCOMUseCaseBuilder {
	b.output = output
	return b
}

// Build creates the LCOMUseCase with the configured dependencies
func (b *LCOMUseCaseBuilder) Build() (*LCOMUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("LCOM service is required")
	}
	if b.fileReader == nil {
		return nil, fmt.Errorf("file reader is required")
	}
	if b.formatter == nil {
		return nil, fmt.Errorf("output formatter is required")
	}

	uc := NewLCOMUseCase(
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
