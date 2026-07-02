package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeCommunityMetrics_SeparatedCommunities(t *testing.T) {
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
	leiden := DetectCommunitiesLeiden(cg, nil)
	metrics := ComputeCommunityMetrics(graph, cg, leiden, nil)

	require.Equal(t, 2, metrics.TotalCommunities)
	require.Len(t, metrics.Communities, 2)

	for _, comm := range metrics.Communities {
		assert.Equal(t, 0, comm.ExternalEdges)
		assert.Equal(t, 0.0, comm.ExternalDependencyRatio)
		assert.Equal(t, 0, comm.IncomingCrossCommunityEdges)
		assert.Equal(t, 0, comm.OutgoingCrossCommunityEdges)
		assert.Greater(t, comm.InternalEdges, 0)
	}
	assert.Empty(t, metrics.BridgeModules)
}

func TestComputeCommunityMetrics_BridgeModule(t *testing.T) {
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
	leiden := DetectCommunitiesLeiden(cg, nil)
	metrics := ComputeCommunityMetrics(graph, cg, leiden, nil)

	require.NotEmpty(t, metrics.BridgeModules)

	var bridgeEntry *BridgeModuleMetrics
	for i := range metrics.BridgeModules {
		if metrics.BridgeModules[i].Module == "bridge" {
			bridgeEntry = &metrics.BridgeModules[i]
			break
		}
	}
	require.NotNil(t, bridgeEntry)
	assert.GreaterOrEqual(t, bridgeEntry.CrossCommunityEdges, 1)
	assert.NotEmpty(t, bridgeEntry.TargetCommunities)
	assert.NotEmpty(t, bridgeEntry.CommunityID)
}

func TestComputeCommunityMetrics_CrossCommunityEdgeCounts(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("left.a", "/project/left/a.py")
	graph.AddModule("left.b", "/project/left/b.py")
	graph.AddModule("hub", "/project/hub.py")
	graph.AddModule("right.a", "/project/right/a.py")
	graph.AddModule("right.b", "/project/right/b.py")

	graph.AddDependency("left.a", "left.b", DependencyEdgeImport, nil)
	graph.AddDependency("left.b", "hub", DependencyEdgeImport, nil)
	graph.AddDependency("hub", "right.a", DependencyEdgeImport, nil)
	graph.AddDependency("right.a", "right.b", DependencyEdgeImport, nil)
	graph.AddDependency("hub", "right.b", DependencyEdgeImport, nil)

	cg := BuildCommunityGraph(graph, nil)

	membership := make([]int, cg.NodeCount)
	for i, name := range cg.NodeNames {
		switch name {
		case "left.a", "left.b":
			membership[i] = 0
		case "hub":
			membership[i] = 1
		default:
			membership[i] = 2
		}
	}

	leiden := &LeidenResult{
		Membership:     membership,
		NumCommunities: 3,
		Modularity:     0.42,
	}
	metrics := ComputeCommunityMetrics(graph, cg, leiden, nil)

	byID := make(map[string]CommunityPartition)
	for _, comm := range metrics.Communities {
		byID[comm.ID] = comm
	}

	hubComm := byID["community_2"]
	assert.Equal(t, 2, hubComm.OutgoingCrossCommunityEdges)
	assert.Equal(t, 1, hubComm.IncomingCrossCommunityEdges)
	assert.Equal(t, 3, hubComm.ExternalEdges)

	leftComm := byID["community_1"]
	assert.Equal(t, 1, leftComm.OutgoingCrossCommunityEdges)
	assert.Equal(t, 0, leftComm.IncomingCrossCommunityEdges)

	rightComm := byID["community_3"]
	assert.Equal(t, 0, rightComm.OutgoingCrossCommunityEdges)
	assert.Equal(t, 2, rightComm.IncomingCrossCommunityEdges)
}

func TestComputeCommunityMetrics_EmptyInput(t *testing.T) {
	metrics := ComputeCommunityMetrics(nil, nil, nil, nil)
	assert.Empty(t, metrics.Communities)
	assert.Empty(t, metrics.BridgeModules)
	assert.Equal(t, 0, metrics.TotalCommunities)
}
