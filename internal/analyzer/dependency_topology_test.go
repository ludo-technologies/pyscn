package analyzer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculateMaxDependencyDepthCountsDAGEdges(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "left", "right", "leaf")
	graph.AddDependency("entry", "left", DependencyEdgeImport, nil)
	graph.AddDependency("entry", "right", DependencyEdgeImport, nil)
	graph.AddDependency("left", "leaf", DependencyEdgeImport, nil)
	graph.AddDependency("right", "leaf", DependencyEdgeImport, nil)

	depth, err := CalculateMaxDependencyDepth(context.Background(), graph)

	require.NoError(t, err)
	assert.Equal(t, 2, depth)
}

func TestCalculateMaxDependencyDepthCollapsesCycles(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "cycle_a", "cycle_b", "leaf")
	graph.AddDependency("entry", "cycle_a", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_a", "cycle_b", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_b", "cycle_a", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_b", "leaf", DependencyEdgeImport, nil)

	depth, err := CalculateMaxDependencyDepth(context.Background(), graph)

	require.NoError(t, err)
	assert.Equal(t, 2, depth)
}

func TestCalculateMaxDependencyDepthHonorsCancellation(t *testing.T) {
	graph := dependencyGraphWithModules("module")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := CalculateMaxDependencyDepth(ctx, graph)

	require.ErrorIs(t, err, context.Canceled)
}

func TestCouplingMetricsUsesDependencyTopologyDepth(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "middle", "leaf")
	graph.AddDependency("entry", "middle", DependencyEdgeImport, nil)
	graph.AddDependency("middle", "leaf", DependencyEdgeImport, nil)

	calculator := NewCouplingMetricsCalculator(graph, DefaultCouplingMetricsOptions())
	require.NoError(t, calculator.CalculateMetrics(context.Background()))

	assert.Equal(t, 2, graph.SystemMetrics.MaxDependencyDepth)
}

func TestCouplingMetricsHonorsCancellation(t *testing.T) {
	graph := dependencyGraphWithModules("module")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calculator := NewCouplingMetricsCalculator(graph, DefaultCouplingMetricsOptions())
	err := calculator.CalculateMetrics(ctx)

	require.ErrorIs(t, err, context.Canceled)
}

func dependencyGraphWithModules(moduleNames ...string) *DependencyGraph {
	graph := NewDependencyGraph("/project")
	for _, moduleName := range moduleNames {
		graph.AddModule(moduleName, "/project/"+moduleName+".py")
	}
	return graph
}
