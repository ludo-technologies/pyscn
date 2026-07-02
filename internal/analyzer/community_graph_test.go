package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCommunityGraph_EmptyGraph(t *testing.T) {
	t.Run("nil graph", func(t *testing.T) {
		cg := BuildCommunityGraph(nil, nil)
		assert.Equal(t, 0, cg.NodeCount)
		assert.Empty(t, cg.NodeNames)
		assert.Empty(t, cg.UndirectedAdj)
		assert.Empty(t, cg.DirectedEdges)
	})

	t.Run("no modules", func(t *testing.T) {
		cg := BuildCommunityGraph(NewDependencyGraph("/project"), nil)
		assert.Equal(t, 0, cg.NodeCount)
		assert.Empty(t, cg.NodeNames)
	})
}

func TestBuildCommunityGraph_TwoCommunities(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("mod.a", "/project/mod/a.py")
	graph.AddModule("mod.b", "/project/mod/b.py")
	graph.AddModule("mod.c", "/project/mod/c.py")
	graph.AddModule("mod.d", "/project/mod/d.py")

	graph.AddDependency("mod.a", "mod.b", DependencyEdgeImport, nil)
	graph.AddDependency("mod.b", "mod.a", DependencyEdgeImport, nil)
	graph.AddDependency("mod.c", "mod.d", DependencyEdgeImport, nil)
	graph.AddDependency("mod.d", "mod.c", DependencyEdgeImport, nil)

	cg := BuildCommunityGraph(graph, nil)

	require.Equal(t, 4, cg.NodeCount)
	assert.Equal(t, []string{"mod.a", "mod.b", "mod.c", "mod.d"}, cg.NodeNames)
	assert.Equal(t, 4, len(cg.DirectedEdges))
	assert.Equal(t, 2.0, undirectedWeight(cg, "mod.a", "mod.b"))
	assert.Equal(t, 2.0, undirectedWeight(cg, "mod.c", "mod.d"))
	assert.Equal(t, 0.0, undirectedWeight(cg, "mod.a", "mod.c"))
	assert.Equal(t, 4.0, cg.TotalUndirectedWeight)
}

func TestBuildCommunityGraph_BridgeNode(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("mod.a", "/project/mod/a.py")
	graph.AddModule("mod.b", "/project/mod/b.py")
	graph.AddModule("bridge", "/project/bridge.py")
	graph.AddModule("mod.c", "/project/mod/c.py")
	graph.AddModule("mod.d", "/project/mod/d.py")

	graph.AddDependency("mod.a", "mod.b", DependencyEdgeImport, nil)
	graph.AddDependency("mod.b", "bridge", DependencyEdgeImport, nil)
	graph.AddDependency("bridge", "mod.c", DependencyEdgeImport, nil)
	graph.AddDependency("mod.c", "mod.d", DependencyEdgeImport, nil)

	cg := BuildCommunityGraph(graph, nil)

	require.Equal(t, 5, cg.NodeCount)
	assert.Equal(t, 1.0, undirectedWeight(cg, "mod.a", "mod.b"))
	assert.Equal(t, 1.0, undirectedWeight(cg, "mod.b", "bridge"))
	assert.Equal(t, 1.0, undirectedWeight(cg, "bridge", "mod.c"))
	assert.Equal(t, 1.0, undirectedWeight(cg, "mod.c", "mod.d"))
	assert.Equal(t, 0.0, undirectedWeight(cg, "mod.a", "mod.d"))
}

func TestBuildCommunityGraph_FullyDisconnected(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("alpha", "/project/alpha.py")
	graph.AddModule("beta", "/project/beta.py")
	graph.AddModule("gamma", "/project/gamma.py")

	cg := BuildCommunityGraph(graph, nil)

	require.Equal(t, 3, cg.NodeCount)
	assert.Empty(t, cg.DirectedEdges)
	assert.Equal(t, 0.0, cg.TotalUndirectedWeight)
	for i := range cg.UndirectedAdj {
		assert.Empty(t, cg.UndirectedAdj[i])
	}
}

func TestBuildCommunityGraph_Deterministic(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("zeta", "/project/zeta.py")
	graph.AddModule("alpha", "/project/alpha.py")
	graph.AddModule("middle", "/project/middle.py")

	graph.AddDependency("alpha", "middle", DependencyEdgeImport, nil)
	graph.AddDependency("middle", "zeta", DependencyEdgeImport, nil)
	graph.AddDependency("zeta", "alpha", DependencyEdgeImport, nil)

	first := BuildCommunityGraph(graph, nil)
	second := BuildCommunityGraph(graph, nil)

	assert.Equal(t, first.NodeNames, second.NodeNames)
	assert.Equal(t, first.DirectedEdges, second.DirectedEdges)
	assert.Equal(t, first.UndirectedAdj, second.UndirectedAdj)
	assert.Equal(t, first.TotalUndirectedWeight, second.TotalUndirectedWeight)
}

func TestBuildCommunityGraph_ExcludeLazyEdges(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("mod.a", "/project/mod/a.py")
	graph.AddModule("mod.b", "/project/mod/b.py")
	graph.AddModule("mod.c", "/project/mod/c.py")

	lazyInfo := &ImportInfo{IsLazy: true}
	graph.AddDependency("mod.a", "mod.b", DependencyEdgeImport, lazyInfo)
	graph.AddDependency("mod.b", "mod.c", DependencyEdgeImport, nil)

	cg := BuildCommunityGraph(graph, &CommunityGraphBuildOptions{ExcludeLazyEdges: true})

	assert.Equal(t, 1, len(cg.DirectedEdges))
	assert.Equal(t, cg.NameToIndex["mod.b"], cg.DirectedEdges[0].FromIndex)
	assert.Equal(t, cg.NameToIndex["mod.c"], cg.DirectedEdges[0].ToIndex)
	assert.Equal(t, 1.0, undirectedWeight(cg, "mod.b", "mod.c"))
	assert.Equal(t, 0.0, undirectedWeight(cg, "mod.a", "mod.b"))
}

func TestBuildCommunityGraph_BidirectionalAggregation(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("mod.a", "/project/mod/a.py")
	graph.AddModule("mod.b", "/project/mod/b.py")

	graph.AddDependency("mod.a", "mod.b", DependencyEdgeImport, nil)
	graph.AddDependency("mod.b", "mod.a", DependencyEdgeImport, nil)

	cg := BuildCommunityGraph(graph, nil)

	assert.Equal(t, 2, len(cg.DirectedEdges))
	assert.Equal(t, 2.0, undirectedWeight(cg, "mod.a", "mod.b"))
	assert.Equal(t, 2.0, cg.TotalUndirectedWeight)
}

func TestBuildCommunityGraph_DefaultIncludesLazyEdges(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("mod.a", "/project/mod/a.py")
	graph.AddModule("mod.b", "/project/mod/b.py")

	lazyInfo := &ImportInfo{IsLazy: true}
	graph.AddDependency("mod.a", "mod.b", DependencyEdgeImport, lazyInfo)

	t.Run("empty options struct", func(t *testing.T) {
		cg := BuildCommunityGraph(graph, &CommunityGraphBuildOptions{})
		require.Len(t, cg.DirectedEdges, 1)
		assert.True(t, cg.DirectedEdges[0].IsLazy)
	})

	t.Run("only EdgeWeightFunc set", func(t *testing.T) {
		cg := BuildCommunityGraph(graph, &CommunityGraphBuildOptions{
			EdgeWeightFunc: func(*DependencyEdge) float64 { return 1.0 },
		})
		require.Len(t, cg.DirectedEdges, 1)
		assert.True(t, cg.DirectedEdges[0].IsLazy)
	})
}

func TestBuildCommunityGraph_CustomEdgeWeightFunc(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("mod.a", "/project/mod/a.py")
	graph.AddModule("mod.b", "/project/mod/b.py")
	graph.AddDependency("mod.a", "mod.b", DependencyEdgeFromImport, nil)

	cg := BuildCommunityGraph(graph, &CommunityGraphBuildOptions{
		EdgeWeightFunc: func(edge *DependencyEdge) float64 {
			if edge.EdgeType == DependencyEdgeFromImport {
				return 2.5
			}
			return 1.0
		},
	})

	assert.Equal(t, 2.5, cg.DirectedEdges[0].Weight)
	assert.Equal(t, 2.5, undirectedWeight(cg, "mod.a", "mod.b"))
	assert.Equal(t, 2.5, cg.TotalUndirectedWeight)
}

func undirectedWeight(cg *CommunityGraph, from, to string) float64 {
	fromIdx, fromOK := cg.NameToIndex[from]
	toIdx, toOK := cg.NameToIndex[to]
	if !fromOK || !toOK {
		return 0
	}

	for _, neighbor := range cg.UndirectedAdj[fromIdx] {
		if neighbor.Index == toIdx {
			return neighbor.Weight
		}
	}
	return 0
}
