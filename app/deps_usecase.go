package app

import (
	"context"
	"fmt"
	"io"

	"github.com/ludo-technologies/pyscn/domain"
	svc "github.com/ludo-technologies/pyscn/service"
)

// DepsUseCase orchestrates the dependency analysis workflow
type DepsUseCase struct {
	service    domain.DependencyService
	fileReader domain.FileReader
	formatter  domain.DepsOutputFormatter
	output     domain.ReportWriter
}

// NewDepsUseCase creates a new dependency analysis use case
func NewDepsUseCase(service domain.DependencyService, fileReader domain.FileReader, formatter domain.DepsOutputFormatter) *DepsUseCase {
	return &DepsUseCase{
		service:    service,
		fileReader: fileReader,
		formatter:  formatter,
		output:     svc.NewFileOutputWriter(nil),
	}
}

// Execute performs dependency analysis and writes formatted output
func (uc *DepsUseCase) Execute(ctx context.Context, req domain.DependencyRequest) error {
	if err := uc.validateRequest(req); err != nil {
		return domain.NewInvalidInputError("invalid request", err)
	}

	// Collect files
	files, err := uc.fileReader.CollectPythonFiles(req.Paths, req.Recursive, req.IncludePatterns, req.ExcludePatterns)
	if err != nil {
		return domain.NewFileNotFoundError("failed to collect files", err)
	}
	if len(files) == 0 {
		return domain.NewInvalidInputError("no Python files found in the specified paths", nil)
	}
	req.Paths = files

	// Analyze
	response, err := uc.service.Analyze(ctx, req)
	if err != nil {
		return domain.NewAnalysisError("dependency analysis failed", err)
	}

	// Output via ReportWriter
	var out io.Writer
	if req.OutputPath == "" {
		out = req.OutputWriter
	}
	if err := uc.output.Write(out, req.OutputPath, req.OutputFormat, req.NoOpen, func(w io.Writer) error {
		return uc.formatter.Write(response, req.OutputFormat, w)
	}); err != nil {
		return domain.NewOutputError("failed to write output", err)
	}
	return nil
}

func (uc *DepsUseCase) validateRequest(req domain.DependencyRequest) error {
	if len(req.Paths) == 0 {
		return fmt.Errorf("no input paths specified")
	}
	if req.OutputWriter == nil && req.OutputPath == "" {
		return fmt.Errorf("output writer or output path is required")
	}
	return nil
}

// DepsUseCaseBuilder provides a fluent builder for DepsUseCase
type DepsUseCaseBuilder struct {
	service    domain.DependencyService
	fileReader domain.FileReader
	formatter  domain.DepsOutputFormatter
	output     domain.ReportWriter
}

func NewDepsUseCaseBuilder() *DepsUseCaseBuilder { return &DepsUseCaseBuilder{} }

func (b *DepsUseCaseBuilder) WithService(s domain.DependencyService) *DepsUseCaseBuilder {
	b.service = s
	return b
}
func (b *DepsUseCaseBuilder) WithFileReader(fr domain.FileReader) *DepsUseCaseBuilder {
	b.fileReader = fr
	return b
}
func (b *DepsUseCaseBuilder) WithFormatter(f domain.DepsOutputFormatter) *DepsUseCaseBuilder {
	b.formatter = f
	return b
}
func (b *DepsUseCaseBuilder) WithOutputWriter(w domain.ReportWriter) *DepsUseCaseBuilder {
	b.output = w
	return b
}

func (b *DepsUseCaseBuilder) Build() (*DepsUseCase, error) {
	if b.service == nil || b.fileReader == nil || b.formatter == nil {
		return nil, fmt.Errorf("missing required dependencies")
	}
	uc := &DepsUseCase{
		service:    b.service,
		fileReader: b.fileReader,
		formatter:  b.formatter,
		output:     b.output,
	}
	if uc.output == nil {
		uc.output = svc.NewFileOutputWriter(nil)
	}
	return uc, nil
}
