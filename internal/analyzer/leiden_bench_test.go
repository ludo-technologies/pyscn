package analyzer

import (
	"fmt"
	"testing"
)

func buildSyntheticLeidenGraph(nodeCount, avgDegree int) *CommunityGraph {
	graph := NewDependencyGraph("/project")
	for i := 0; i < nodeCount; i++ {
		name := fmt.Sprintf("mod.%d", i)
		graph.AddModule(name, "/project/"+name+".py")
	}

	names := graph.GetModuleNames()
	for i, from := range names {
		for d := 1; d <= avgDegree; d++ {
			to := names[(i+d*7)%len(names)]
			if from == to {
				continue
			}
			graph.AddDependency(from, to, DependencyEdgeImport, nil)
		}
	}

	return BuildCommunityGraph(graph, nil)
}

func BenchmarkDetectCommunitiesLeiden_MediumGraph(b *testing.B) {
	cg := buildSyntheticLeidenGraph(200, 4)
	opts := DefaultLeidenOptions()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetectCommunitiesLeiden(cg, opts)
	}
}
