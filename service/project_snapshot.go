package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/analyzer"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

// ProjectSnapshot stores the parsed source needed by multiple analyzers.
type ProjectSnapshot struct {
	// Files is a compatibility projection. Mutating it does not affect the
	// sealed snapshot state consumed by analyzers.
	// Deprecated: use FileProjections to materialize a copy on demand.
	Files              []*ProjectFile
	files              []*ProjectFile
	analysisFiles      map[string]struct{}
	moduleFiles        map[string]struct{}
	defaultProjectRoot string
}

// ProjectSnapshotOptions controls which optional per-file analysis caches are built.
type ProjectSnapshotOptions struct {
	IncludeRawMetrics bool
}

// ModuleGraphOptions controls how a project snapshot is projected into a
// module graph.
type ModuleGraphOptions struct {
	ProjectRoot     string
	Graph           domain.ModuleGraphOptions
	IncludePatterns []string
	ExcludePatterns []string
}

// ProjectModuleGraph is an owned graph projection of a project snapshot. Its
// analyzer representation remains private to the service layer.
type ProjectModuleGraph struct {
	graph  *analyzer.DependencyGraph
	policy domain.ModuleGraphOptions
}

// Clone returns an independent graph for one analyzer to consume.
func (g *ProjectModuleGraph) Clone() *ProjectModuleGraph {
	if g == nil || g.graph == nil {
		return nil
	}
	return &ProjectModuleGraph{graph: g.graph.Clone(), policy: g.policy}
}

// ProjectFile stores one Python file after read and parse.
type ProjectFile struct {
	Path         string
	AST          *parser.Node
	RawMetrics   *analyzer.RawMetricsResult
	ReadErr      error
	ParseErr     error
	identityPath string
	source       []byte
	parseResult  *parser.ParseResult

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
	return buildProjectSnapshot(ctx, paths, paths, paths, options, true)
}

// BuildAnalysisProjectSnapshot captures the implementation and importable
// module surfaces once while retaining their distinct analyzer scopes.
func BuildAnalysisProjectSnapshot(ctx context.Context, analysisPaths, modulePaths []string, options ProjectSnapshotOptions) *ProjectSnapshot {
	paths := mergeSnapshotPaths(analysisPaths, modulePaths)
	return buildProjectSnapshot(ctx, paths, analysisPaths, modulePaths, options, false)
}

func buildProjectSnapshot(ctx context.Context, paths, analysisPaths, modulePaths []string, options ProjectSnapshotOptions, exposeCompatibilityFiles bool) *ProjectSnapshot {
	if ctx == nil {
		ctx = context.Background()
	}

	defaultProjectRoot, _ := os.Getwd()
	snapshot := &ProjectSnapshot{
		files:              make([]*ProjectFile, len(paths)),
		analysisFiles:      snapshotPathSet(defaultProjectRoot, analysisPaths),
		moduleFiles:        snapshotPathSet(defaultProjectRoot, modulePaths),
		defaultProjectRoot: defaultProjectRoot,
	}
	if len(paths) == 0 {
		if exposeCompatibilityFiles {
			snapshot.Files = []*ProjectFile{}
		}
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
				identityPath := snapshotPath(defaultProjectRoot, path)
				snapshot.files[idx] = buildProjectFile(ctx, pyParser, path, identityPath, options)
			}
		}()
	}

	cancelled := false
	for idx := range paths {
		if cancelled {
			snapshot.files[idx] = cancelledProjectFile(paths[idx], ctx.Err())
			continue
		}

		select {
		case <-ctx.Done():
			snapshot.files[idx] = cancelledProjectFile(paths[idx], ctx.Err())
			cancelled = true
		case jobs <- idx:
		}
	}

	close(jobs)
	wg.Wait()

	for idx, path := range paths {
		if snapshot.files[idx] == nil {
			snapshot.files[idx] = cancelledProjectFile(path, ctx.Err())
		}
	}
	if exposeCompatibilityFiles {
		snapshot.Files = snapshot.FileProjections()
	}

	return snapshot
}

// FileProjections returns independent copies of the captured project files.
func (s *ProjectSnapshot) FileProjections() []*ProjectFile {
	if s == nil {
		return nil
	}
	return projectFileProjections(s.files)
}

// Paths returns the file paths represented by the snapshot.
func (s *ProjectSnapshot) Paths() []string {
	if s == nil {
		return nil
	}

	paths := make([]string, 0, len(s.files))
	for _, file := range s.files {
		if file != nil && s.hasAnalysisFile(file) {
			paths = append(paths, file.Path)
		}
	}
	return paths
}

// Coverage reports project-wide read and parse coverage from the snapshot.
func (s *ProjectSnapshot) Coverage() domain.AnalysisCoverage {
	if s == nil {
		return domain.AnalysisCoverage{}
	}

	coverage := domain.AnalysisCoverage{
		TotalFiles:  len(s.files),
		Diagnostics: make([]domain.AnalysisDiagnostic, 0),
	}
	for _, file := range s.files {
		if file == nil {
			coverage.SkippedFiles++
			continue
		}
		if file.ReadErr != nil {
			coverage.SkippedFiles++
			coverage.Diagnostics = append(coverage.Diagnostics, domain.AnalysisDiagnostic{
				FilePath: file.Path,
				Code:     domain.DiagnosticCodeRead,
				Message:  file.ReadErr.Error(),
			})
			continue
		}
		if file.ParseErr != nil || file.AST == nil {
			coverage.SkippedFiles++
			message := "invalid parse result"
			if file.ParseErr != nil {
				message = file.ParseErr.Error()
			}
			coverage.Diagnostics = append(coverage.Diagnostics, domain.AnalysisDiagnostic{
				FilePath: file.Path,
				Code:     domain.DiagnosticCodeParse,
				Message:  message,
			})
			continue
		}
		coverage.AnalyzedFiles++
	}

	return coverage
}

// BuildDependencyGraph projects the snapshot's successfully parsed files into
// a new dependency graph. The returned graph is owned by the caller.
func (s *ProjectSnapshot) BuildDependencyGraph(ctx context.Context, options *ModuleGraphOptions) (*ProjectModuleGraph, error) {
	if s == nil {
		return nil, fmt.Errorf("project snapshot is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("build dependency graph: %w", err)
	}

	moduleFiles := s.selectedModuleFiles(options)
	parsedModules := make([]analyzer.ParsedModule, 0, len(moduleFiles))
	for _, file := range moduleFiles {
		if !file.Parsed() {
			continue
		}
		parsedModule, err := analyzer.NewParsedModule(file.identityPath, file.source, file.AST)
		if err != nil {
			return nil, fmt.Errorf("project file %s: %w", file.Path, err)
		}
		parsedModules = append(parsedModules, parsedModule)
	}

	projectRoot := s.defaultProjectRoot
	graphOptions := domain.ModuleGraphOptions{}
	includePatterns := []string(nil)
	excludePatterns := []string(nil)
	if options != nil {
		if options.ProjectRoot != "" {
			projectRoot = snapshotPath(s.defaultProjectRoot, options.ProjectRoot)
		}
		graphOptions = options.Graph
		includePatterns = options.IncludePatterns
		excludePatterns = options.ExcludePatterns
	}
	analyzerOptions := &analyzer.ModuleAnalysisOptions{
		ProjectRoot:       projectRoot,
		ModuleRoots:       capturedModuleRoots(projectRoot, moduleFiles),
		IncludeStdLib:     domain.BoolPtr(graphOptions.IncludeStdLib),
		IncludeThirdParty: domain.BoolPtr(graphOptions.IncludeThirdParty),
		FollowRelative:    domain.BoolPtr(graphOptions.FollowRelative),
		IncludePatterns:   includePatterns,
		ExcludePatterns:   excludePatterns,
	}
	moduleAnalyzer, err := analyzer.NewModuleAnalyzer(analyzerOptions)
	if err != nil {
		return nil, fmt.Errorf("create module analyzer: %w", err)
	}
	graph, err := moduleAnalyzer.AnalyzeParsedModules(ctx, parsedModules)
	if err != nil {
		return nil, fmt.Errorf("analyze parsed modules: %w", err)
	}
	policy := domain.ModuleGraphOptions{}
	if options != nil {
		policy = options.Graph
	}
	return &ProjectModuleGraph{graph: graph, policy: policy}, nil
}

func (s *ProjectSnapshot) analysisProjectFiles() []*ProjectFile {
	return s.analysisProjectFilesMatching(nil, nil)
}

func (s *ProjectSnapshot) analysisProjectFilesMatching(includePatterns, excludePatterns []string) []*ProjectFile {
	if s == nil {
		return nil
	}
	files := make([]*ProjectFile, 0, len(s.analysisFiles))
	for _, file := range s.files {
		if file != nil && s.hasAnalysisFile(file) && matchesFileSelection(file.identityPath, includePatterns, excludePatterns) {
			files = append(files, file)
		}
	}
	return files
}

func (s *ProjectSnapshot) hasAnalysisFile(file *ProjectFile) bool {
	if s == nil || file == nil {
		return false
	}
	_, ok := s.analysisFiles[file.identityPath]
	return ok
}

func (s *ProjectSnapshot) selectedModuleFiles(options *ModuleGraphOptions) []*ProjectFile {
	if s == nil {
		return nil
	}
	var includePatterns, excludePatterns []string
	if options != nil {
		includePatterns = options.IncludePatterns
		excludePatterns = options.ExcludePatterns
	}
	files := make([]*ProjectFile, 0, len(s.moduleFiles))
	for _, file := range s.files {
		if file == nil {
			continue
		}
		if _, ok := s.moduleFiles[file.identityPath]; !ok {
			continue
		}
		if !matchesFileSelection(file.identityPath, includePatterns, excludePatterns) {
			continue
		}
		files = append(files, file)
	}
	return files
}

func matchesFileSelection(path string, includePatterns, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		if patternMatches(pattern, path, true) {
			return false
		}
	}
	if len(includePatterns) == 0 {
		return true
	}
	for _, pattern := range includePatterns {
		if patternMatches(pattern, path, false) {
			return true
		}
	}
	return false
}

func mergeSnapshotPaths(pathSets ...[]string) []string {
	paths := make([]string, 0)
	seen := make(map[string]struct{})
	for _, pathSet := range pathSets {
		for _, path := range pathSet {
			identity, err := filepath.Abs(path)
			if err != nil {
				continue
			}
			identity = filepath.Clean(identity)
			if _, duplicate := seen[identity]; duplicate {
				continue
			}
			seen[identity] = struct{}{}
			paths = append(paths, path)
		}
	}
	return paths
}

func snapshotPathSet(captureRoot string, paths []string) map[string]struct{} {
	set := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		set[snapshotPath(captureRoot, path)] = struct{}{}
	}
	return set
}

func capturedModuleRoots(projectRoot string, files []*ProjectFile) []string {
	absoluteRoot := filepath.Clean(projectRoot)
	srcRoot := filepath.Join(absoluteRoot, "src")
	hasSrcModule := false
	hasSrcPackageInit := false
	for _, file := range files {
		if file == nil {
			continue
		}
		absolutePath := file.identityPath
		relative, err := filepath.Rel(srcRoot, absolutePath)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			continue
		}
		hasSrcModule = true
		if relative == "__init__.py" || relative == "__init__.pyi" {
			hasSrcPackageInit = true
		}
	}
	if hasSrcModule && !hasSrcPackageInit {
		return []string{srcRoot, absoluteRoot}
	}
	return []string{absoluteRoot}
}

func projectFileProjections(files []*ProjectFile) []*ProjectFile {
	projections := make([]*ProjectFile, len(files))
	for index, file := range files {
		if file == nil {
			continue
		}
		projection := &ProjectFile{
			Path:       file.Path,
			AST:        parser.CloneNode(file.AST),
			RawMetrics: file.RawMetrics.Clone(),
			ReadErr:    file.ReadErr,
			ParseErr:   file.ParseErr,
			source:     append([]byte(nil), file.source...),
		}
		projections[index] = projection
	}
	return projections
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

func buildProjectFile(ctx context.Context, pyParser *parser.Parser, path, identityPath string, options ProjectSnapshotOptions) *ProjectFile {
	file := &ProjectFile{Path: path, identityPath: identityPath}

	select {
	case <-ctx.Done():
		file.ReadErr = fmt.Errorf("analysis cancelled: %w", ctx.Err())
		return file
	default:
	}

	content, err := os.ReadFile(identityPath)
	if err != nil {
		file.ReadErr = err
		return file
	}

	if options.IncludeRawMetrics {
		file.RawMetrics = analyzer.CalculateRawMetrics(content, path)
	}
	file.source = content

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
	file.parseResult = result
	if file.RawMetrics != nil {
		analyzer.PopulateLogicalLines(file.RawMetrics, file.AST)
	}

	return file
}

func snapshotPath(captureRoot, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(captureRoot, path))
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
