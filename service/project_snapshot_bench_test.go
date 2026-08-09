package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ludo-technologies/pyscn/internal/analyzer"
)

func BenchmarkAggregateModuleGraphReuse(b *testing.B) {
	projectRoot, paths := writeModuleGraphBenchmarkProject(b, 40)
	ctx := context.Background()

	b.Run("reparse_for_each_consumer", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			for range 2 {
				moduleAnalyzer, err := analyzer.NewModuleAnalyzer(&analyzer.ModuleAnalysisOptions{ProjectRoot: projectRoot})
				if err != nil {
					b.Fatal(err)
				}
				if _, err := moduleAnalyzer.AnalyzeFiles(paths); err != nil {
					b.Fatal(err)
				}
			}
		}
	})

	b.Run("capture_once_and_clone", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			snapshot := BuildProjectSnapshot(ctx, paths)
			graph, err := snapshot.BuildDependencyGraph(ctx, &ModuleGraphOptions{ProjectRoot: projectRoot})
			if err != nil {
				b.Fatal(err)
			}
			firstConsumer := graph.Clone()
			secondConsumer := graph.Clone()
			if firstConsumer == nil || secondConsumer == nil {
				b.Fatal("expected independent graph projections")
			}
		}
	})
}

func writeModuleGraphBenchmarkProject(b *testing.B, moduleCount int) (string, []string) {
	b.Helper()
	projectRoot := b.TempDir()
	paths := make([]string, moduleCount)
	for index := range moduleCount {
		path := filepath.Join(projectRoot, fmt.Sprintf("module_%03d.py", index))
		source := "VALUE = 1\n"
		if index+1 < moduleCount {
			source = fmt.Sprintf("import module_%03d\nVALUE = 1\n", index+1)
		}
		if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
			b.Fatalf("write benchmark module: %v", err)
		}
		paths[index] = path
	}
	return projectRoot, paths
}
