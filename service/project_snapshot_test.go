package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/parser"
)

func TestProjectSnapshotCachesParsedFileState(t *testing.T) {
	ctx := context.Background()
	sourcePath := writeSnapshotFixture(t)

	snapshot := BuildProjectSnapshot(ctx, []string{sourcePath})
	files := snapshot.FileProjections()
	if len(files) != 1 {
		t.Fatalf("expected 1 snapshot file, got %d", len(files))
	}

	file := files[0]
	if !file.Parsed() {
		t.Fatalf("expected parsed file, read err: %v, parse err: %v", file.ReadErr, file.ParseErr)
	}
	if file.RawMetrics == nil {
		t.Fatal("expected raw metrics")
	}

	firstCFGs, err := file.CFGs()
	if err != nil {
		t.Fatalf("first CFG build failed: %v", err)
	}
	secondCFGs, err := file.CFGs()
	if err != nil {
		t.Fatalf("second CFG build failed: %v", err)
	}
	if len(firstCFGs) == 0 {
		t.Fatal("expected CFGs")
	}
	if firstCFGs[domain.ModuleFunctionName] != secondCFGs[domain.ModuleFunctionName] {
		t.Fatal("expected cached CFG objects to be reused")
	}
}

func TestProjectSnapshotOptionsSkipRawMetrics(t *testing.T) {
	ctx := context.Background()
	sourcePath := writeSnapshotFixture(t)

	snapshot := BuildProjectSnapshotWithOptions(ctx, []string{sourcePath}, ProjectSnapshotOptions{})
	files := snapshot.FileProjections()
	if len(files) != 1 {
		t.Fatalf("expected 1 snapshot file, got %d", len(files))
	}

	file := files[0]
	if !file.Parsed() {
		t.Fatalf("expected parsed file, read err: %v, parse err: %v", file.ReadErr, file.ParseErr)
	}
	if file.RawMetrics != nil {
		t.Fatal("expected raw metrics to be skipped")
	}
	if _, err := file.CFGs(); err != nil {
		t.Fatalf("expected CFGs without raw metrics: %v", err)
	}
}

func TestAnalysisProjectSnapshotKeepsSourceAndModuleScopesDistinct(t *testing.T) {
	projectRoot := t.TempDir()
	sourcePath := filepath.Join(projectRoot, "runtime.py")
	stubPath := filepath.Join(projectRoot, "contract.pyi")
	if err := os.WriteFile(sourcePath, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(stubPath, []byte("VALUE: int\n"), 0o644); err != nil {
		t.Fatalf("write stub: %v", err)
	}

	snapshot := BuildAnalysisProjectSnapshot(
		context.Background(),
		[]string{sourcePath},
		[]string{sourcePath, stubPath},
		ProjectSnapshotOptions{},
	)
	if projections := snapshot.FileProjections(); len(projections) != 2 {
		t.Fatalf("expected two on-demand file projections, got %d", len(projections))
	}
	if got := snapshot.Paths(); len(got) != 1 || got[0] != sourcePath {
		t.Fatalf("expected only the implementation path, got %v", got)
	}
	if got := snapshot.analysisProjectFiles(); len(got) != 1 || got[0].Path != sourcePath {
		t.Fatalf("expected only the implementation file, got %v", got)
	}
	coverage := snapshot.Coverage()
	if coverage.TotalFiles != 1 || coverage.AnalyzedFiles != 1 || coverage.SkippedFiles != 0 {
		t.Fatalf("expected coverage to use the implementation scope, got %+v", coverage)
	}
	modules := snapshot.selectedModuleFiles()
	if len(modules) != 2 {
		t.Fatalf("expected source and stub module files, got %d", len(modules))
	}
}

func TestAnalysisProjectSnapshotRetainsCancelledFilesInCoverage(t *testing.T) {
	projectRoot := t.TempDir()
	paths := make([]string, 100)
	for index := range paths {
		paths[index] = filepath.Join(projectRoot, fmt.Sprintf("module_%d.py", index))
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	snapshot := BuildAnalysisProjectSnapshot(ctx, paths, paths, ProjectSnapshotOptions{})

	if got := snapshot.Paths(); len(got) != len(paths) {
		t.Fatalf("expected %d cancelled paths, got %d", len(paths), len(got))
	}
	coverage := snapshot.Coverage()
	if coverage.TotalFiles != len(paths) || coverage.AnalyzedFiles != 0 || coverage.SkippedFiles != len(paths) {
		t.Fatalf("expected all cancelled files in coverage, got %+v", coverage)
	}
	if len(coverage.Diagnostics) != len(paths) {
		t.Fatalf("expected one cancellation diagnostic per file, got %d", len(coverage.Diagnostics))
	}
}

func TestAnalysisProjectSnapshotMatchesBareIncludePatternsByFilename(t *testing.T) {
	projectRoot := t.TempDir()
	nestedDir := filepath.Join(projectRoot, "pkg")
	if err := os.MkdirAll(nestedDir, 0o755); err != nil {
		t.Fatalf("create nested directory: %v", err)
	}
	sourcePath := filepath.Join(nestedDir, "source.py")
	if err := os.WriteFile(sourcePath, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}

	snapshot := BuildAnalysisProjectSnapshot(
		context.Background(),
		[]string{sourcePath},
		[]string{sourcePath},
		ProjectSnapshotOptions{},
	)
	files := snapshot.analysisProjectFilesMatching([]string{"*.py"}, nil)
	if len(files) != 1 || files[0].Path != sourcePath {
		t.Fatalf("expected bare include pattern to retain nested source, got %+v", files)
	}
}

func TestProjectSnapshotProjectionDoesNotShareValueNodes(t *testing.T) {
	sourcePath := filepath.Join(t.TempDir(), "source.py")
	if err := os.WriteFile(sourcePath, []byte("def call():\n    return target()\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	snapshot := BuildProjectSnapshot(context.Background(), []string{sourcePath})

	projectedCalls := snapshot.FileProjections()[0].AST.FindByType(parser.NodeCall)
	internalCalls := snapshot.files[0].AST.FindByType(parser.NodeCall)
	if len(projectedCalls) != 1 || len(internalCalls) != 1 {
		t.Fatalf("expected one call in each AST, got projected=%d internal=%d", len(projectedCalls), len(internalCalls))
	}
	projectedCallee, projectedOK := projectedCalls[0].Value.(*parser.Node)
	internalCallee, internalOK := internalCalls[0].Value.(*parser.Node)
	if !projectedOK || !internalOK {
		t.Fatal("expected node-valued call targets")
	}
	projectedCallee.Name = "mutated"
	if internalCallee.Name != "target" {
		t.Fatalf("public projection mutated sealed syntax to %q", internalCallee.Name)
	}
}

func TestProjectSnapshotBuildDependencyGraphUsesOnlyCapturedParsedFiles(t *testing.T) {
	ctx := context.Background()
	projectRoot := t.TempDir()
	sourcePath := filepath.Join(projectRoot, "source.py")
	targetPath := filepath.Join(projectRoot, "target.py")
	brokenPath := filepath.Join(projectRoot, "broken.py")

	if err := os.WriteFile(sourcePath, []byte("import target\n"), 0o644); err != nil {
		t.Fatalf("write source module: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("write target module: %v", err)
	}
	if err := os.WriteFile(brokenPath, []byte("def broken(:\n"), 0o644); err != nil {
		t.Fatalf("write broken source: %v", err)
	}

	snapshot := BuildProjectSnapshotWithOptions(ctx, []string{sourcePath, targetPath, brokenPath}, ProjectSnapshotOptions{})
	projection := snapshot.FileProjections()[0]
	projection.Path = filepath.Join(projectRoot, "mutated.py")
	projection.AST = nil
	if err := os.WriteFile(sourcePath, []byte("VALUE = 2\n"), 0o644); err != nil {
		t.Fatalf("replace captured source: %v", err)
	}

	graph, err := snapshot.BuildDependencyGraph(ctx, &ModuleGraphOptions{ProjectRoot: projectRoot})
	if err != nil {
		t.Fatalf("build dependency graph: %v", err)
	}
	if graph.graph.TotalModules != 2 {
		t.Fatalf("expected only the two parsed modules, got %d", graph.graph.TotalModules)
	}
	if graph.graph.Nodes["source"] == nil || !graph.graph.Nodes["source"].Dependencies["target"] {
		t.Fatalf("expected dependency from captured syntax, got %+v", graph.graph.Nodes)
	}
	if graph.graph.Nodes["broken"] != nil {
		t.Fatal("broken module must not be added to the graph")
	}
}

func TestComplexitySnapshotRequiresRawMetrics(t *testing.T) {
	ctx := context.Background()
	sourcePath := writeSnapshotFixture(t)
	paths := []string{sourcePath}
	snapshot := BuildProjectSnapshotWithOptions(ctx, paths, ProjectSnapshotOptions{})

	_, err := NewComplexityService().AnalyzeSnapshot(ctx, snapshot, domain.ComplexityRequest{
		Paths:           paths,
		OutputFormat:    domain.OutputFormatJSON,
		MinComplexity:   1,
		SortBy:          domain.SortByName,
		LowThreshold:    domain.DefaultComplexityLowThreshold,
		MediumThreshold: domain.DefaultComplexityMediumThreshold,
	})
	if err == nil {
		t.Fatal("expected complexity snapshot without raw metrics to fail")
	}
}

func TestProjectSnapshotCapturesSrcModuleRoot(t *testing.T) {
	projectRoot := t.TempDir()
	packageDir := filepath.Join(projectRoot, "src", "pkg")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("create src package: %v", err)
	}
	sourcePath := filepath.Join(packageDir, "source.py")
	targetPath := filepath.Join(packageDir, "target.py")
	if err := os.WriteFile(sourcePath, []byte("import pkg.target\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(targetPath, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	snapshot := BuildProjectSnapshot(context.Background(), []string{sourcePath, targetPath})
	for _, path := range []string{sourcePath, targetPath, packageDir, filepath.Join(projectRoot, "src")} {
		if err := os.Remove(path); err != nil {
			t.Fatalf("remove captured path %s: %v", path, err)
		}
	}

	graph, err := snapshot.BuildDependencyGraph(context.Background(), &ModuleGraphOptions{ProjectRoot: projectRoot})
	if err != nil {
		t.Fatalf("build captured src graph: %v", err)
	}
	if graph.graph.Nodes["pkg.source"] == nil || !graph.graph.Nodes["pkg.source"].Dependencies["pkg.target"] {
		t.Fatalf("expected src-rooted captured modules, got %+v", graph.graph.Nodes)
	}
}

func TestProjectSnapshotPreservesNamespacePackageFromSelectedSubtree(t *testing.T) {
	projectRoot := t.TempDir()
	packageDir := filepath.Join(projectRoot, "src", "acme")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("create namespace package: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectRoot, "pyproject.toml"), []byte("[project]\nname = 'sample'\n"), 0o644); err != nil {
		t.Fatalf("write project marker: %v", err)
	}
	alphaPath := filepath.Join(packageDir, "alpha.py")
	betaPath := filepath.Join(packageDir, "beta.py")
	if err := os.WriteFile(alphaPath, []byte("from acme import beta\n"), 0o644); err != nil {
		t.Fatalf("write alpha: %v", err)
	}
	if err := os.WriteFile(betaPath, []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("write beta: %v", err)
	}

	snapshot := BuildAnalysisProjectSnapshot(
		context.Background(),
		[]string{alphaPath, betaPath},
		[]string{alphaPath, betaPath},
		ProjectSnapshotOptions{},
	)
	graph, err := snapshot.BuildDependencyGraph(context.Background(), &ModuleGraphOptions{ProjectRoot: projectRoot})
	if err != nil {
		t.Fatalf("build namespace graph: %v", err)
	}
	if graph.graph.Nodes["acme.alpha"] == nil || graph.graph.Nodes["acme.beta"] == nil {
		t.Fatalf("expected namespace-qualified modules, got %+v", graph.graph.Nodes)
	}
	if !graph.graph.Nodes["acme.alpha"].Dependencies["acme.beta"] {
		t.Fatalf("expected namespace import edge, got %+v", graph.graph.Nodes["acme.alpha"].Dependencies)
	}
}

func TestProjectSnapshotResolvesRelativeModulePathsAtCaptureTime(t *testing.T) {
	originalWorkingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalWorkingDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	projectRoot := t.TempDir()
	packageDir := filepath.Join(projectRoot, "src", "pkg")
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		t.Fatalf("create src package: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "source.py"), []byte("import pkg.target\n"), 0o644); err != nil {
		t.Fatalf("write source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "target.py"), []byte("VALUE = 1\n"), 0o644); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.Chdir(projectRoot); err != nil {
		t.Fatalf("enter project root: %v", err)
	}
	relativePaths := []string{filepath.Join("src", "pkg", "source.py"), filepath.Join("src", "pkg", "target.py")}
	snapshot := BuildProjectSnapshot(context.Background(), relativePaths)

	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("leave captured project: %v", err)
	}
	graph, err := snapshot.BuildDependencyGraph(context.Background(), &ModuleGraphOptions{})
	if err != nil {
		t.Fatalf("build graph after working directory changed: %v", err)
	}
	if graph.graph.Nodes["pkg.source"] == nil || !graph.graph.Nodes["pkg.source"].Dependencies["pkg.target"] {
		t.Fatalf("expected capture-rooted modules, got %+v", graph.graph.Nodes)
	}
	if got := snapshot.Paths(); len(got) != len(relativePaths) || got[0] != relativePaths[0] || got[1] != relativePaths[1] {
		t.Fatalf("expected public paths to preserve caller spelling, got %v", got)
	}
}

func TestSnapshotServicesMatchFileServices(t *testing.T) {
	ctx := context.Background()
	sourcePath := writeSnapshotFixture(t)
	paths := []string{sourcePath}
	snapshot := BuildProjectSnapshot(ctx, paths)

	complexityReq := domain.ComplexityRequest{
		Paths:           paths,
		OutputFormat:    domain.OutputFormatJSON,
		MinComplexity:   1,
		SortBy:          domain.SortByName,
		LowThreshold:    domain.DefaultComplexityLowThreshold,
		MediumThreshold: domain.DefaultComplexityMediumThreshold,
	}
	complexitySvc := NewComplexityService()
	regularComplexity, err := complexitySvc.Analyze(ctx, complexityReq)
	if err != nil {
		t.Fatalf("regular complexity failed: %v", err)
	}
	snapshotComplexity, err := complexitySvc.AnalyzeSnapshot(ctx, snapshot, complexityReq)
	if err != nil {
		t.Fatalf("snapshot complexity failed: %v", err)
	}
	if len(regularComplexity.Functions) != len(snapshotComplexity.Functions) {
		t.Fatalf("complexity function count mismatch: regular=%d snapshot=%d", len(regularComplexity.Functions), len(snapshotComplexity.Functions))
	}
	if regularComplexity.Summary.TotalFunctions != snapshotComplexity.Summary.TotalFunctions {
		t.Fatalf("complexity summary mismatch: regular=%d snapshot=%d", regularComplexity.Summary.TotalFunctions, snapshotComplexity.Summary.TotalFunctions)
	}

	cboReq := *domain.DefaultCBORequest()
	cboReq.Paths = paths
	cboReq.ShowZeros = domain.BoolPtr(true)
	cboSvc := NewCBOService()
	regularCBO, err := cboSvc.Analyze(ctx, cboReq)
	if err != nil {
		t.Fatalf("regular CBO failed: %v", err)
	}
	snapshotCBO, err := cboSvc.AnalyzeSnapshot(ctx, snapshot, cboReq)
	if err != nil {
		t.Fatalf("snapshot CBO failed: %v", err)
	}
	if len(regularCBO.Classes) != len(snapshotCBO.Classes) {
		t.Fatalf("CBO class count mismatch: regular=%d snapshot=%d", len(regularCBO.Classes), len(snapshotCBO.Classes))
	}

	lcomReq := *domain.DefaultLCOMRequest()
	lcomReq.Paths = paths
	lcomSvc := NewLCOMService()
	regularLCOM, err := lcomSvc.Analyze(ctx, lcomReq)
	if err != nil {
		t.Fatalf("regular LCOM failed: %v", err)
	}
	snapshotLCOM, err := lcomSvc.AnalyzeSnapshot(ctx, snapshot, lcomReq)
	if err != nil {
		t.Fatalf("snapshot LCOM failed: %v", err)
	}
	if len(regularLCOM.Classes) != len(snapshotLCOM.Classes) {
		t.Fatalf("LCOM class count mismatch: regular=%d snapshot=%d", len(regularLCOM.Classes), len(snapshotLCOM.Classes))
	}
}

func writeSnapshotFixture(t *testing.T) string {
	t.Helper()

	source := `import json

class Example:
    def __init__(self, value):
        self.value = value

    def duplicate_one(self, items):
        total = 0
        for item in items:
            if item:
                total += item
        return total

    def duplicate_two(self, items):
        total = 0
        for item in items:
            if item:
                total += item
        return total

def top_level(flag):
    if flag:
        return json.dumps({"ok": True})
    return "{}"
`

	sourcePath := filepath.Join(t.TempDir(), "sample.py")
	if err := os.WriteFile(sourcePath, []byte(source), 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	return sourcePath
}
