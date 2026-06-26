package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeLayerMismatchMetrics_AlignedLayers(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("billing.invoice", "/project/billing/invoice.py")
	graph.AddModule("billing.tax", "/project/billing/tax.py")
	graph.AddModule("inventory.stock", "/project/inventory/stock.py")
	graph.AddModule("inventory.warehouse", "/project/inventory/warehouse.py")

	graph.AddDependency("billing.invoice", "billing.tax", DependencyEdgeImport, nil)
	graph.AddDependency("inventory.stock", "inventory.warehouse", DependencyEdgeImport, nil)

	moduleToLayer := map[string]string{
		"billing.invoice":     "billing",
		"billing.tax":         "billing",
		"inventory.stock":     "inventory",
		"inventory.warehouse": "inventory",
	}

	cg := BuildCommunityGraph(graph, nil)
	leiden := DetectCommunitiesLeiden(cg, nil)
	metrics := ComputeCommunityMetrics(graph, cg, leiden, moduleToLayer)
	mismatch := ComputeLayerMismatchMetrics(metrics.Communities, metrics.BridgeModules)

	require.NotNil(t, mismatch)
	assert.Empty(t, mismatch.CrossLayerCommunities)
	assert.Empty(t, mismatch.LayerBridgeModules)
	assert.InDelta(t, 1.0, mismatch.LayerAlignmentScore, 1e-9)
}

func TestComputeLayerMismatchMetrics_BridgeFixture(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("api.a", "/project/api/a.py")
	graph.AddModule("api.b", "/project/api/b.py")
	graph.AddModule("bridge", "/project/bridge.py")
	graph.AddModule("infra.c", "/project/infra/c.py")
	graph.AddModule("infra.d", "/project/infra/d.py")

	graph.AddDependency("api.a", "api.b", DependencyEdgeImport, nil)
	graph.AddDependency("api.b", "bridge", DependencyEdgeImport, nil)
	graph.AddDependency("bridge", "infra.c", DependencyEdgeImport, nil)
	graph.AddDependency("infra.c", "infra.d", DependencyEdgeImport, nil)

	moduleToLayer := map[string]string{
		"api.a":   "api",
		"api.b":   "api",
		"bridge":  "unknown",
		"infra.c": "infra",
		"infra.d": "infra",
	}

	cg := BuildCommunityGraph(graph, nil)
	leiden := DetectCommunitiesLeiden(cg, nil)
	metrics := ComputeCommunityMetrics(graph, cg, leiden, moduleToLayer)
	mismatch := ComputeLayerMismatchMetrics(metrics.Communities, metrics.BridgeModules)

	require.NotNil(t, mismatch)
	assert.InDelta(t, 1.0, mismatch.LayerAlignmentScore, 1e-9)
	assert.Empty(t, mismatch.CrossLayerCommunities)
	assert.Equal(t, []string{"bridge", "infra.c"}, mismatch.LayerBridgeModules)
}

func TestComputeLayerMismatchMetrics_CrossLayerCommunity(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("api.one", "/project/api/one.py")
	graph.AddModule("api.two", "/project/api/two.py")
	graph.AddModule("infra.one", "/project/infra/one.py")
	graph.AddModule("infra.two", "/project/infra/two.py")

	graph.AddDependency("api.one", "api.two", DependencyEdgeImport, nil)
	graph.AddDependency("api.one", "infra.one", DependencyEdgeImport, nil)
	graph.AddDependency("api.two", "infra.two", DependencyEdgeImport, nil)
	graph.AddDependency("infra.one", "api.one", DependencyEdgeImport, nil)
	graph.AddDependency("infra.two", "infra.one", DependencyEdgeImport, nil)

	moduleToLayer := map[string]string{
		"api.one":   "api",
		"api.two":   "api",
		"infra.one": "infra",
		"infra.two": "infra",
	}

	cg := BuildCommunityGraph(graph, nil)
	membership := make([]int, cg.NodeCount)
	leiden := &LeidenResult{
		Membership:     membership,
		NumCommunities: 1,
		Modularity:     0.1,
	}
	metrics := ComputeCommunityMetrics(graph, cg, leiden, moduleToLayer)
	mismatch := ComputeLayerMismatchMetrics(metrics.Communities, metrics.BridgeModules)

	require.NotNil(t, mismatch)
	require.Len(t, metrics.Communities, 1)
	assert.Equal(t, 2, metrics.Communities[0].LayerCount)
	assert.Equal(t, []string{"community_1"}, mismatch.CrossLayerCommunities)
	assert.InDelta(t, 1.0, mismatch.LayerAlignmentScore, 1e-9)
}

func TestComputeLayerMismatchMetrics_NoLayerMapping(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.Nodes["solo"] = &ModuleNode{Name: "solo", FilePath: "/project/solo.py"}

	cg := &CommunityGraph{
		NodeNames: []string{"solo"},
		NodeCount: 1,
	}
	leiden := &LeidenResult{
		Membership:     []int{0},
		NumCommunities: 1,
	}
	metrics := ComputeCommunityMetrics(graph, cg, leiden, nil)
	mismatch := ComputeLayerMismatchMetrics(metrics.Communities, metrics.BridgeModules)

	require.NotNil(t, mismatch)
	assert.Equal(t, 0.0, mismatch.LayerAlignmentScore)
	assert.Empty(t, mismatch.CrossLayerCommunities)
	assert.Empty(t, mismatch.LayerBridgeModules)
	assert.Equal(t, 0, metrics.Communities[0].LayerCount)
}
