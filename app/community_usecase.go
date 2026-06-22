package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ludo-technologies/pyscn/domain"
	svc "github.com/ludo-technologies/pyscn/service"
)

// CommunityUseCase orchestrates the community analysis workflow.
type CommunityUseCase struct {
	service      domain.CommunityAnalysisService
	fileReader   domain.FileReader
	formatter    domain.CommunityAnalysisOutputFormatter
	configLoader domain.CommunityConfigurationLoader
	output       domain.ReportWriter
}

// NewCommunityUseCase creates a new community use case.
func NewCommunityUseCase(
	service domain.CommunityAnalysisService,
	fileReader domain.FileReader,
	formatter domain.CommunityAnalysisOutputFormatter,
	configLoader domain.CommunityConfigurationLoader,
) *CommunityUseCase {
	return &CommunityUseCase{
		service:      service,
		fileReader:   fileReader,
		formatter:    formatter,
		configLoader: configLoader,
		output:       svc.NewFileOutputWriter(nil),
	}
}

func (uc *CommunityUseCase) prepareAnalysis(ctx context.Context, req domain.CommunityAnalysisRequest) (domain.CommunityAnalysisRequest, error) {
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

	if len(finalReq.SourcePaths) == 0 {
		finalReq.SourcePaths = append([]string(nil), finalReq.Paths...)
	}
	finalReq.Paths = files
	return finalReq, nil
}

// Execute performs the complete community analysis workflow.
func (uc *CommunityUseCase) Execute(ctx context.Context, req domain.CommunityAnalysisRequest) error {
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return err
	}

	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return domain.NewAnalysisError("community analysis failed", err)
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

// AnalyzeAndReturn performs community analysis and returns the response without formatting.
func (uc *CommunityUseCase) AnalyzeAndReturn(ctx context.Context, req domain.CommunityAnalysisRequest) (*domain.CommunityAnalysisResult, error) {
	finalReq, err := uc.prepareAnalysis(ctx, req)
	if err != nil {
		return nil, err
	}

	response, err := uc.service.Analyze(ctx, finalReq)
	if err != nil {
		return nil, domain.NewAnalysisError("community analysis failed", err)
	}

	return response, nil
}

func (uc *CommunityUseCase) validateRequest(req domain.CommunityAnalysisRequest) error {
	if len(req.Paths) == 0 {
		return fmt.Errorf("no input paths specified")
	}
	if req.OutputWriter == nil && req.OutputPath == "" {
		return fmt.Errorf("output writer or output path is required")
	}
	if req.MinCommunitySize < 0 {
		return fmt.Errorf("minimum community size cannot be negative")
	}
	if req.Resolution < 0 {
		return fmt.Errorf("resolution cannot be negative")
	}
	switch req.OutputFormat {
	case domain.OutputFormatText, domain.OutputFormatJSON, domain.OutputFormatYAML, domain.OutputFormatCSV, domain.OutputFormatHTML:
	default:
		return fmt.Errorf("unsupported output format: %s", req.OutputFormat)
	}
	return nil
}

func (uc *CommunityUseCase) loadAndMergeConfig(req domain.CommunityAnalysisRequest) (domain.CommunityAnalysisRequest, error) {
	if uc.configLoader == nil {
		return req, nil
	}

	var configReq *domain.CommunityAnalysisRequest
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

// CommunityUseCaseBuilder provides a builder pattern for creating CommunityUseCase.
type CommunityUseCaseBuilder struct {
	service      domain.CommunityAnalysisService
	fileReader   domain.FileReader
	formatter    domain.CommunityAnalysisOutputFormatter
	configLoader domain.CommunityConfigurationLoader
	output       domain.ReportWriter
}

// NewCommunityUseCaseBuilder creates a new builder.
func NewCommunityUseCaseBuilder() *CommunityUseCaseBuilder {
	return &CommunityUseCaseBuilder{}
}

func (b *CommunityUseCaseBuilder) WithService(service domain.CommunityAnalysisService) *CommunityUseCaseBuilder {
	b.service = service
	return b
}

func (b *CommunityUseCaseBuilder) WithFileReader(fileReader domain.FileReader) *CommunityUseCaseBuilder {
	b.fileReader = fileReader
	return b
}

func (b *CommunityUseCaseBuilder) WithFormatter(formatter domain.CommunityAnalysisOutputFormatter) *CommunityUseCaseBuilder {
	b.formatter = formatter
	return b
}

func (b *CommunityUseCaseBuilder) WithConfigLoader(configLoader domain.CommunityConfigurationLoader) *CommunityUseCaseBuilder {
	b.configLoader = configLoader
	return b
}

func (b *CommunityUseCaseBuilder) WithOutputWriter(output domain.ReportWriter) *CommunityUseCaseBuilder {
	b.output = output
	return b
}

func (b *CommunityUseCaseBuilder) Build() (*CommunityUseCase, error) {
	if b.service == nil {
		return nil, fmt.Errorf("community analysis service is required")
	}
	if b.fileReader == nil {
		return nil, fmt.Errorf("file reader is required")
	}
	if b.formatter == nil {
		return nil, fmt.Errorf("output formatter is required")
	}

	uc := NewCommunityUseCase(
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
