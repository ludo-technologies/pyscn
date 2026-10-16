package analyzer

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func calculateTestCouplingMetrics(t *testing.T, graph *DependencyGraph) {
	t.Helper()
	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 0)
	require.NoError(t, err)
	calculator := NewCouplingMetricsCalculator(graph, DefaultCouplingMetricsOptions())
	require.NoError(t, calculator.CalculateMetrics(context.Background(), topology))
}

func TestCouplingMetricsAllIsolatedModulesHaveNoDeviation(t *testing.T) {
	graph := dependencyGraphWithModules("pkg", "pkg.a", "pkg.b", "pkg.c", "pkg.d")

	calculateTestCouplingMetrics(t, graph)

	for name, metrics := range graph.ModuleMetrics {
		assert.True(t, IsIsolatedModule(metrics), name)
	}
	assert.Zero(t, graph.SystemMetrics.MainSequenceDeviation)
	assert.Zero(t, graph.SystemMetrics.AverageInstability)
	assert.Empty(t, graph.SystemMetrics.RefactoringPriority)
}

func TestCouplingMetricsExcludeIsolatedModulesFromAverages(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "leaf", "isolated")
	graph.AddDependency("entry", "leaf", DependencyEdgeImport, nil)

	calculateTestCouplingMetrics(t, graph)

	entry := graph.ModuleMetrics["entry"]
	leaf := graph.ModuleMetrics["leaf"]
	require.True(t, IsIsolatedModule(graph.ModuleMetrics["isolated"]))
	require.False(t, IsIsolatedModule(entry))
	require.False(t, IsIsolatedModule(leaf))

	assert.InDelta(t, (entry.Distance+leaf.Distance)/2, graph.SystemMetrics.MainSequenceDeviation, 1e-9)
	assert.InDelta(t, (entry.Instability+leaf.Instability)/2, graph.SystemMetrics.AverageInstability, 1e-9)
	assert.NotContains(t, graph.SystemMetrics.RefactoringPriority, "isolated")
}
