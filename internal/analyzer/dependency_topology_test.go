package analyzer

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyzeDependencyTopologyReportsDepthAndChains(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "left", "right", "leaf")
	graph.AddDependency("entry", "left", DependencyEdgeImport, nil)
	graph.AddDependency("entry", "right", DependencyEdgeImport, nil)
	graph.AddDependency("left", "leaf", DependencyEdgeImport, nil)
	graph.AddDependency("right", "leaf", DependencyEdgeImport, nil)

	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 2)

	require.NoError(t, err)
	assert.Equal(t, 2, topology.MaxDepth())
	chains := topology.LongestChains()
	require.Len(t, chains, 2)
	assert.Equal(t, []string{"entry", "left", "leaf"}, chains[0].Path)
	assert.Equal(t, []string{"entry", "right", "leaf"}, chains[1].Path)

	chains[0].Path[0] = "changed"
	assert.Equal(t, "entry", topology.LongestChains()[0].Path[0])
}

func TestAnalyzeDependencyTopologyCollapsesCycles(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "cycle_a", "cycle_b", "leaf")
	graph.AddDependency("entry", "cycle_a", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_a", "cycle_b", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_b", "cycle_a", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_b", "leaf", DependencyEdgeImport, nil)

	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 0)

	require.NoError(t, err)
	assert.Equal(t, 2, topology.MaxDepth())
}

func TestAnalyzeDependencyTopologyExcludesLazyImports(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "dependency")
	graph.AddDependency("entry", "dependency", DependencyEdgeImport, nil)
	graph.AddDependency(
		"dependency",
		"entry",
		DependencyEdgeFromImport,
		&ImportInfo{IsLazy: true},
	)

	cycles := NewCircularDependencyDetector(graph).DetectCircularDependencies()
	require.False(t, cycles.HasCircularDependencies)

	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 1)

	require.NoError(t, err)
	assert.Equal(t, 1, topology.MaxDepth())
	chains := topology.LongestChains()
	require.Len(t, chains, 1)
	assert.Equal(t, []string{"entry", "dependency"}, chains[0].Path)

	calculator := NewCouplingMetricsCalculator(graph, DefaultCouplingMetricsOptions())
	require.NoError(t, calculator.CalculateMetrics(context.Background(), topology))
	assert.Equal(t, 2, graph.SystemMetrics.TotalDependencies)
	assert.Equal(t, 1, graph.SystemMetrics.MaxDependencyDepth)
}

func TestAnalyzeDependencyTopologyHonorsCancellation(t *testing.T) {
	graph := dependencyGraphWithModules("module")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := AnalyzeDependencyTopology(ctx, graph, 0)

	require.ErrorIs(t, err, context.Canceled)
}

func TestAnalyzeDependencyTopologyRejectsUnknownDependencies(t *testing.T) {
	graph := dependencyGraphWithModules("entry")
	graph.Nodes["entry"].Dependencies["missing"] = true

	_, err := AnalyzeDependencyTopology(context.Background(), graph, 0)

	require.ErrorContains(t, err, `module "entry" depends on unknown module "missing"`)
}

func TestCouplingMetricsUsesDependencyTopologyDepth(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "middle", "leaf")
	graph.AddDependency("entry", "middle", DependencyEdgeImport, nil)
	graph.AddDependency("middle", "leaf", DependencyEdgeImport, nil)

	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 0)
	require.NoError(t, err)
	calculator := NewCouplingMetricsCalculator(graph, DefaultCouplingMetricsOptions())
	require.NoError(t, calculator.CalculateMetrics(context.Background(), topology))

	assert.Equal(t, 2, graph.SystemMetrics.MaxDependencyDepth)
}

func TestCouplingMetricsHonorsCancellation(t *testing.T) {
	graph := dependencyGraphWithModules("module")
	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 0)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calculator := NewCouplingMetricsCalculator(graph, DefaultCouplingMetricsOptions())
	err = calculator.CalculateMetrics(ctx, topology)

	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, graph.ModuleMetrics)
	assert.Zero(t, graph.SystemMetrics.TotalModules)
}

func TestCouplingMetricsRejectsTopologyFromAnotherGraph(t *testing.T) {
	topologyGraph := dependencyGraphWithModules("topology")
	topology, err := AnalyzeDependencyTopology(context.Background(), topologyGraph, 0)
	require.NoError(t, err)

	metricsGraph := dependencyGraphWithModules("metrics")
	calculator := NewCouplingMetricsCalculator(metricsGraph, DefaultCouplingMetricsOptions())
	err = calculator.CalculateMetrics(context.Background(), topology)

	require.ErrorContains(t, err, "dependency topology belongs to another graph")
	assert.Empty(t, metricsGraph.ModuleMetrics)
}

func TestCouplingMetricsRejectsStaleTopology(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "leaf")
	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 0)
	require.NoError(t, err)
	graph.AddDependency("entry", "leaf", DependencyEdgeImport, nil)

	calculator := NewCouplingMetricsCalculator(graph, DefaultCouplingMetricsOptions())
	err = calculator.CalculateMetrics(context.Background(), topology)

	require.ErrorContains(t, err, "dependency graph changed after topology analysis")
	assert.Empty(t, graph.ModuleMetrics)
}

func TestCouplingMetricsRejectsTopologyAfterLazyImportPromotion(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "dependency")
	graph.AddDependency(
		"entry",
		"dependency",
		DependencyEdgeImport,
		&ImportInfo{IsLazy: true},
	)
	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 0)
	require.NoError(t, err)

	graph.AddDependency("entry", "dependency", DependencyEdgeImport, nil)

	calculator := NewCouplingMetricsCalculator(graph, DefaultCouplingMetricsOptions())
	err = calculator.CalculateMetrics(context.Background(), topology)

	require.ErrorContains(t, err, "dependency graph changed after topology analysis")
	assert.Empty(t, graph.ModuleMetrics)
}

func TestCouplingMetricsCanDisableAbstractness(t *testing.T) {
	graph := dependencyGraphWithModules("abstract")
	graph.Nodes["abstract"].ClassCount = 1
	graph.Nodes["abstract"].AbstractClassCount = 1
	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 0)
	require.NoError(t, err)

	options := DefaultCouplingMetricsOptions()
	options.IncludeAbstractness = false
	calculator := NewCouplingMetricsCalculator(graph, options)
	require.NoError(t, calculator.CalculateMetrics(context.Background(), topology))

	assert.Zero(t, graph.ModuleMetrics["abstract"].Abstractness)
}

func TestAnalyzeDependencyTopologyRanksDeepBranchAfterShallowBranches(t *testing.T) {
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

	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 10)

	require.NoError(t, err)
	chains := topology.LongestChains()
	require.NotEmpty(t, chains)
	assert.Equal(
		t,
		[]string{"root", "z_deep", "z_middle", "z_leaf"},
		chains[0].Path,
	)
}

func TestAnalyzeDependencyTopologyExpandsCycleComponents(t *testing.T) {
	graph := dependencyGraphWithModules("entry", "cycle_a", "cycle_b", "leaf")
	graph.AddDependency("entry", "cycle_a", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_a", "cycle_b", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_b", "cycle_a", DependencyEdgeImport, nil)
	graph.AddDependency("cycle_b", "leaf", DependencyEdgeImport, nil)

	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 1)

	require.NoError(t, err)
	chains := topology.LongestChains()
	require.Len(t, chains, 1)
	assert.Equal(
		t,
		[]string{"entry", "cycle_a", "cycle_b", "leaf"},
		chains[0].Path,
	)
}

func TestAnalyzeDependencyTopologyExpandsMultipleCycleComponents(t *testing.T) {
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

	topology, err := AnalyzeDependencyTopology(context.Background(), graph, 1)

	require.NoError(t, err)
	chains := topology.LongestChains()
	require.Len(t, chains, 1)
	assert.Equal(
		t,
		[]string{"entry", "first_a", "first_b", "second_a", "second_b", "leaf"},
		chains[0].Path,
	)
}

func TestAnalyzeDependencyTopologyChainSearchHonorsCancellation(t *testing.T) {
	graph := dependencyGraphWithModules("module")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := AnalyzeDependencyTopology(ctx, graph, 1)

	require.ErrorIs(t, err, context.Canceled)
}

func dependencyGraphWithModules(moduleNames ...string) *DependencyGraph {
	graph := NewDependencyGraph("/project")
	for _, moduleName := range moduleNames {
		graph.AddModule(moduleName, "/project/"+moduleName+".py")
	}
	return graph
}
