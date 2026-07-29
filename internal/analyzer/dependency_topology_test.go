package analyzer

import (
	"context"
	"fmt"
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

func TestCalculateMaxDependencyDepthRejectsUnknownDependencies(t *testing.T) {
	graph := dependencyGraphWithModules("entry")
	graph.Nodes["entry"].Dependencies["missing"] = true

	_, err := CalculateMaxDependencyDepth(context.Background(), graph)

	require.ErrorContains(t, err, `module "entry" depends on unknown module "missing"`)
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
	assert.Empty(t, graph.ModuleMetrics)
	assert.Zero(t, graph.SystemMetrics.TotalModules)
}

func TestFindLongestDependencyChainsReturnsActualLongestPath(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "left", "right", "leaf")
	graph.AddDependency("entry", "left", DependencyEdgeImport, nil)
	graph.AddDependency("entry", "right", DependencyEdgeImport, nil)
	graph.AddDependency("left", "leaf", DependencyEdgeImport, nil)
	graph.AddDependency("right", "leaf", DependencyEdgeImport, nil)

	chains, err := FindLongestDependencyChains(context.Background(), graph, 2)

	require.NoError(t, err)
	require.Len(t, chains, 2)
	assert.Equal(t, []string{"entry", "left", "leaf"}, chains[0].Path)
	assert.Equal(t, []string{"entry", "right", "leaf"}, chains[1].Path)
}

func TestFindLongestDependencyChainsRanksDeepBranchAfterShallowBranches(t *testing.T) {
	moduleNames := []string{"root", "z_deep", "z_middle", "z_leaf"}
	for index := 0; index < 10; index++ {
		moduleNames = append(moduleNames, fmt.Sprintf("a_shallow_%02d", index))
	}
	graph := dependencyGraphWithModules(moduleNames...)
	for index := 0; index < 10; index++ {
		graph.AddDependency(
			"root",
			fmt.Sprintf("a_shallow_%02d", index),
			DependencyEdgeImport,
			nil,
		)
	}
	graph.AddDependency("root", "z_deep", DependencyEdgeImport, nil)
	graph.AddDependency("z_deep", "z_middle", DependencyEdgeImport, nil)
	graph.AddDependency("z_middle", "z_leaf", DependencyEdgeImport, nil)

	chains, err := FindLongestDependencyChains(context.Background(), graph, 10)

	require.NoError(t, err)
	require.NotEmpty(t, chains)
	assert.Equal(t, []string{"root", "z_deep", "z_middle", "z_leaf"}, chains[0].Path)
}

func TestFindLongestDependencyChainsExpandsCycleComponents(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "cycle_a", "cycle_b", "leaf")
	graph.AddDependency("entry", "cycle_a", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_a", "cycle_b", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_b", "cycle_a", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_b", "leaf", DependencyEdgeImport, nil)

	chains, err := FindLongestDependencyChains(context.Background(), graph, 1)

	require.NoError(t, err)
	require.Len(t, chains, 1)
	assert.Equal(t, []string{"entry", "cycle_a", "cycle_b", "leaf"}, chains[0].Path)
}

func TestFindLongestDependencyChainsExpandsMultipleCycleComponents(t *testing.T) {
	graph := dependencyGraphWithModules(
		"entry",
		"first_a",
		"first_b",
		"second_a",
		"second_b",
		"leaf",
	)
	graph.AddDependency("entry", "first_a", DependencyEdgeImport, nil)
	graph.AddDependency("first_a", "first_b", DependencyEdgeImport, nil)
	graph.AddDependency("first_b", "first_a", DependencyEdgeImport, nil)
	graph.AddDependency("first_b", "second_a", DependencyEdgeImport, nil)
	graph.AddDependency("second_a", "second_b", DependencyEdgeImport, nil)
	graph.AddDependency("second_b", "second_a", DependencyEdgeImport, nil)
	graph.AddDependency("second_b", "leaf", DependencyEdgeImport, nil)

	chains, err := FindLongestDependencyChains(context.Background(), graph, 1)

	require.NoError(t, err)
	require.Len(t, chains, 1)
	assert.Equal(
		t,
		[]string{"entry", "first_a", "first_b", "second_a", "second_b", "leaf"},
		chains[0].Path,
	)
}

func TestFindLongestDependencyChainsHonorsCancellation(t *testing.T) {
	graph := dependencyGraphWithModules("module")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := FindLongestDependencyChains(ctx, graph, 1)

	require.ErrorIs(t, err, context.Canceled)
}

func dependencyGraphWithModules(moduleNames ...string) *DependencyGraph {
	graph := NewDependencyGraph("/project")
	for _, moduleName := range moduleNames {
		graph.AddModule(moduleName, "/project/"+moduleName+".py")
	}
	return graph
}
