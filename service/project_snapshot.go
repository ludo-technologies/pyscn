package service

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// ProjectSnapshot stores the parsed source needed by multiple analyzers.
type ProjectSnapshot struct {
	Files []*ProjectFile
}

// ProjectSnapshotOptions controls which optional per-file analysis caches are built.
type ProjectSnapshotOptions struct {
	IncludeRawMetrics bool
}

// ProjectFile stores one Python file after read and parse.
type ProjectFile struct {
	Path       string
	AST        *parser.Node
	RawMetrics *analyzer.RawMetricsResult
	ReadErr    error
	ParseErr   error

	cfgOnce sync.Once
	cfgs    map[string]*analyzer.CFG
	cfgErr  error
}

// BuildProjectSnapshot reads and parses each file once for the full analyze command.
func BuildProjectSnapshot(ctx context.Context, paths []string) *ProjectSnapshot {
	return BuildProjectSnapshotWithOptions(ctx, paths, ProjectSnapshotOptions{IncludeRawMetrics: true})
}

// BuildProjectSnapshotWithOptions reads and parses each file once with analyzer-scoped caches.
func BuildProjectSnapshotWithOptions(ctx context.Context, paths []string, options ProjectSnapshotOptions) *ProjectSnapshot {
	if ctx == nil {
		ctx = context.Background()
	}

	snapshot := &ProjectSnapshot{Files: make([]*ProjectFile, len(paths))}
	if len(paths) == 0 {
		return snapshot
	}

	workerCount := min(len(paths), runtime.GOMAXPROCS(0))
	if workerCount < 1 {
		workerCount = 1
	}

	jobs := make(chan int)
	var wg sync.WaitGroup

	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()

			pyParser := parser.New()
			for idx := range jobs {
				path := paths[idx]
				snapshot.Files[idx] = buildProjectFile(ctx, pyParser, path, options)
			}
		}()
	}

	cancelled := false
	for idx := range paths {
		if cancelled {
			snapshot.Files[idx] = cancelledProjectFile(paths[idx], ctx.Err())
			continue
		}

		select {
		case <-ctx.Done():
			snapshot.Files[idx] = cancelledProjectFile(paths[idx], ctx.Err())
			cancelled = true
		case jobs <- idx:
		}
	}

	close(jobs)
	wg.Wait()

	for idx, path := range paths {
		if snapshot.Files[idx] == nil {
			snapshot.Files[idx] = cancelledProjectFile(path, ctx.Err())
		}
	}

	return snapshot
}

// Paths returns the file paths represented by the snapshot.
func (s *ProjectSnapshot) Paths() []string {
	if s == nil {
		return nil
	}

	paths := make([]string, 0, len(s.Files))
	for _, file := range s.Files {
		if file != nil {
			paths = append(paths, file.Path)
		}
	}
	return paths
}

// Parsed reports whether the file has a valid parsed AST.
func (f *ProjectFile) Parsed() bool {
	return f != nil && f.ReadErr == nil && f.ParseErr == nil && f.AST != nil
}

// CFGs builds CFGs once and shares them across CFG-backed analyzers.
func (f *ProjectFile) CFGs() (map[string]*analyzer.CFG, error) {
	if f == nil {
		return nil, fmt.Errorf("project file cannot be nil")
	}
	if f.ReadErr != nil {
		return nil, f.ReadErr
	}
	if f.ParseErr != nil {
		return nil, f.ParseErr
	}
	if f.AST == nil {
		return nil, fmt.Errorf("invalid parse result")
	}

	f.cfgOnce.Do(func() {
		builder := analyzer.NewCFGBuilder()
		f.cfgs, f.cfgErr = builder.BuildAll(f.AST)
	})

	return f.cfgs, f.cfgErr
}

func buildProjectFile(ctx context.Context, pyParser *parser.Parser, path string, options ProjectSnapshotOptions) *ProjectFile {
	file := &ProjectFile{Path: path}

	select {
	case <-ctx.Done():
		file.ReadErr = fmt.Errorf("analysis cancelled: %w", ctx.Err())
		return file
	default:
	}

	content, err := os.ReadFile(path)
	if err != nil {
		file.ReadErr = err
		return file
	}

	if options.IncludeRawMetrics {
		file.RawMetrics = analyzer.CalculateRawMetrics(content, path)
	}

	result, err := pyParser.Parse(ctx, content)
	if err != nil {
		file.ParseErr = err
		return file
	}
	if result == nil || result.AST == nil {
		file.ParseErr = fmt.Errorf("invalid parse result")
		return file
	}

	file.AST = result.AST
	if file.RawMetrics != nil {
		analyzer.PopulateLogicalLines(file.RawMetrics, file.AST)
	}

	return file
}

func cancelledProjectFile(path string, err error) *ProjectFile {
	if err == nil {
		err = context.Canceled
	}
	return &ProjectFile{
		Path:    path,
		ReadErr: fmt.Errorf("analysis cancelled: %w", err),
	}
}

func countSourceLines(content []byte) int {
	return len(strings.Split(string(content), "\n"))
}
