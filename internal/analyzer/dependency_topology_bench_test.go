package analyzer

import (
	"context"
	"fmt"
	"testing"
)

func BenchmarkAnalyzeDependencyTopologyBranchingGraph(b *testing.B) {
	for _, moduleCount := range []int{44, 1000} {
		b.Run(fmt.Sprintf("modules_%d", moduleCount), func(b *testing.B) {
			moduleNames := make([]string, moduleCount)
			for index := range moduleNames {
				moduleNames[index] = fmt.Sprintf("module_%04d", index)
			}
			graph := dependencyGraphWithModules(moduleNames...)
			for index, moduleName := range moduleNames {
				if index+1 < moduleCount {
					graph.AddDependency(
						moduleName,
						moduleNames[index+1],
						DependencyEdgeImport,
						nil,
					)
				}
				if index+2 < moduleCount {
					graph.AddDependency(
						moduleName,
						moduleNames[index+2],
						DependencyEdgeImport,
						nil,
					)
				}
			}

			b.ReportAllocs()
			for b.Loop() {
				topology, err := AnalyzeDependencyTopology(context.Background(), graph, 10)
				if err != nil {
					b.Fatal(err)
				}
				if topology.MaxDepth != moduleCount-1 {
					b.Fatalf(
						"expected depth %d, got %d",
						moduleCount-1,
						topology.MaxDepth,
					)
				}
			}
		})
	}
}
