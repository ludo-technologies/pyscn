package app

import (
    "bytes"
    "context"
    "errors"
    "io"
    "testing"

    "github.com/ludo-technologies/pyscn/domain"
)

// Mocks
type mockDepService struct{ resp *domain.DependencyResponse; err error }
func (m *mockDepService) Analyze(ctx context.Context, req domain.DependencyRequest) (*domain.DependencyResponse, error) {
    return m.resp, m.err
}

type mockDepsFileReader struct{ files []string; err error }
func (m *mockDepsFileReader) CollectPythonFiles(paths []string, recursive bool, includePatterns, excludePatterns []string) ([]string, error) {
    if m.err != nil { return nil, m.err }
    if m.files != nil { return m.files, nil }
    return []string{"x.py"}, nil
}
func (m *mockDepsFileReader) ReadFile(path string) ([]byte, error) { return nil, nil }
func (m *mockDepsFileReader) IsValidPythonFile(path string) bool { return true }
func (m *mockDepsFileReader) FileExists(path string) (bool, error) { return true, nil }

type mockFormatter struct{ called bool; lastFormat domain.OutputFormat }
func (m *mockFormatter) Write(resp *domain.DependencyResponse, format domain.OutputFormat, w io.Writer) error {
    m.called = true
    m.lastFormat = format
    if w != nil { _, _ = w.Write([]byte("ok")) }
    return nil
}

type mockReportWriter struct{ called bool; lastPath string; lastFormat domain.OutputFormat; err error }
func (mw *mockReportWriter) Write(writer io.Writer, outputPath string, format domain.OutputFormat, noOpen bool, writeFunc func(io.Writer) error) error {
    mw.called = true
    mw.lastPath = outputPath
    mw.lastFormat = format
    // Simulate writing to buffer instead of filesystem
    var buf bytes.Buffer
    if err := writeFunc(&buf); err != nil { return err }
    if mw.err != nil { return mw.err }
    return nil
}

func TestDepsUseCase_Execute_Success(t *testing.T) {
    svc := &mockDepService{resp: &domain.DependencyResponse{Summary: domain.DependencySummary{Modules:1}}}
    fr := &mockDepsFileReader{files: []string{"a.py"}}
    fmt := &mockFormatter{}
    out := &mockReportWriter{}

    uc, err := NewDepsUseCaseBuilder().
        WithService(svc).
        WithFileReader(fr).
        WithFormatter(fmt).
        WithOutputWriter(out).
        Build()
    if err != nil { t.Fatalf("build usecase: %v", err) }

    req := domain.DependencyRequest{Paths: []string{"."}, OutputWriter: &bytes.Buffer{}, OutputFormat: domain.OutputFormatText}
    if err := uc.Execute(context.Background(), req); err != nil {
        t.Fatalf("execute: %v", err)
    }
    if !out.called || !fmt.called {
        t.Fatalf("expected formatter and report writer to be called")
    }
}

func TestDepsUseCase_Execute_InvalidRequest_NoPaths(t *testing.T) {
    uc := NewDepsUseCase(&mockDepService{}, &mockDepsFileReader{}, &mockFormatter{})
    err := uc.Execute(context.Background(), domain.DependencyRequest{Paths: []string{}, OutputWriter: &bytes.Buffer{}, OutputFormat: domain.OutputFormatText})
    if err == nil { t.Fatalf("expected error for empty paths") }
}

func TestDepsUseCase_Execute_FileReaderError(t *testing.T) {
    fr := &mockDepsFileReader{err: errors.New("collect failed")}
    uc := NewDepsUseCase(&mockDepService{}, fr, &mockFormatter{})
    err := uc.Execute(context.Background(), domain.DependencyRequest{Paths: []string{"."}, OutputWriter: &bytes.Buffer{}, OutputFormat: domain.OutputFormatText})
    if err == nil { t.Fatalf("expected error from file reader") }
}

func TestDepsUseCase_Execute_AnalysisError(t *testing.T) {
    svc := &mockDepService{err: errors.New("analyze failed")}
    fr := &mockDepsFileReader{files: []string{"a.py"}}
    uc := NewDepsUseCase(svc, fr, &mockFormatter{})
    err := uc.Execute(context.Background(), domain.DependencyRequest{Paths: []string{"."}, OutputWriter: &bytes.Buffer{}, OutputFormat: domain.OutputFormatText})
    if err == nil { t.Fatalf("expected analysis error") }
}

func TestDepsUseCase_Execute_ReportWriterError(t *testing.T) {
    svc := &mockDepService{resp: &domain.DependencyResponse{Summary: domain.DependencySummary{}}}
    fr := &mockDepsFileReader{files: []string{"a.py"}}
    fmt := &mockFormatter{}
    out := &mockReportWriter{err: errors.New("write failed")}
    uc, err := NewDepsUseCaseBuilder().WithService(svc).WithFileReader(fr).WithFormatter(fmt).WithOutputWriter(out).Build()
    if err != nil { t.Fatalf("build usecase: %v", err) }
    if err := uc.Execute(context.Background(), domain.DependencyRequest{Paths: []string{"."}, OutputWriter: &bytes.Buffer{}, OutputFormat: domain.OutputFormatText}); err == nil {
        t.Fatalf("expected write error")
    }
}
