package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputePackageMismatchMetrics_SplitPackage(t *testing.T) {
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
	mismatch := ComputePackageMismatchMetrics(metrics.Communities)

	require.NotNil(t, mismatch)
	assert.Equal(t, []string{"mod"}, mismatch.SplitPackages)
	assert.Empty(t, mismatch.MixedCommunities)
	assert.InDelta(t, 0.0, mismatch.PackageAlignmentScore, 1e-9)

	byID := communityPartitionsByID(metrics.Communities)
	for _, id := range []string{"community_1", "community_2"} {
		comm, ok := byID[id]
		require.True(t, ok)
		assert.Equal(t, 1, comm.PackageCount)
		assert.Equal(t, "mod", comm.DominantPackage)
		assert.InDelta(t, 1.0, comm.PackageAlignment, 1e-9)
	}
}

func TestComputePackageMismatchMetrics_AlignedPackages(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("billing.invoice", "/project/billing/invoice.py")
	graph.AddModule("billing.tax", "/project/billing/tax.py")
	graph.AddModule("inventory.stock", "/project/inventory/stock.py")
	graph.AddModule("inventory.warehouse", "/project/inventory/warehouse.py")

	graph.AddDependency("billing.invoice", "billing.tax", DependencyEdgeImport, nil)
	graph.AddDependency("inventory.stock", "inventory.warehouse", DependencyEdgeImport, nil)

	cg := BuildCommunityGraph(graph, nil)
	leiden := DetectCommunitiesLeiden(cg, nil)
	metrics := ComputeCommunityMetrics(graph, cg, leiden, nil)
	mismatch := ComputePackageMismatchMetrics(metrics.Communities)

	require.NotNil(t, mismatch)
	assert.Empty(t, mismatch.SplitPackages)
	assert.Empty(t, mismatch.MixedCommunities)
	assert.InDelta(t, 1.0, mismatch.PackageAlignmentScore, 1e-9)
}

func TestComputePackageMismatchMetrics_MixedCommunity(t *testing.T) {
	graph := NewDependencyGraph("/project")
	graph.AddModule("pkg_alpha.one", "/project/pkg_alpha/one.py")
	graph.AddModule("pkg_alpha.two", "/project/pkg_alpha/two.py")
	graph.AddModule("pkg_beta.one", "/project/pkg_beta/one.py")
	graph.AddModule("pkg_beta.two", "/project/pkg_beta/two.py")

	graph.AddDependency("pkg_alpha.one", "pkg_alpha.two", DependencyEdgeImport, nil)
	graph.AddDependency("pkg_alpha.one", "pkg_beta.one", DependencyEdgeImport, nil)
	graph.AddDependency("pkg_alpha.two", "pkg_beta.two", DependencyEdgeImport, nil)
	graph.AddDependency("pkg_beta.one", "pkg_alpha.one", DependencyEdgeImport, nil)
	graph.AddDependency("pkg_beta.two", "pkg_beta.one", DependencyEdgeImport, nil)

	cg := BuildCommunityGraph(graph, nil)
	membership := make([]int, cg.NodeCount)
	for i, name := range cg.NodeNames {
		membership[i] = 0
		_ = name
	}
	leiden := &LeidenResult{
		Membership:     membership,
		NumCommunities: 1,
		Modularity:     0.1,
	}
	metrics := ComputeCommunityMetrics(graph, cg, leiden, nil)
	mismatch := ComputePackageMismatchMetrics(metrics.Communities)

	require.NotNil(t, mismatch)
	require.Len(t, metrics.Communities, 1)
	assert.Equal(t, []string{"pkg_alpha", "pkg_beta"}, metrics.Communities[0].Packages)
	assert.Equal(t, 2, metrics.Communities[0].PackageCount)
	assert.Equal(t, []string{"community_1"}, mismatch.MixedCommunities)
	assert.Empty(t, mismatch.SplitPackages)
	assert.InDelta(t, 1.0, mismatch.PackageAlignmentScore, 1e-9)
}

func TestComputePackageMismatchMetrics_NoPackageMetadata(t *testing.T) {
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
	mismatch := ComputePackageMismatchMetrics(metrics.Communities)

	require.NotNil(t, mismatch)
	assert.Equal(t, 0.0, mismatch.PackageAlignmentScore)
	assert.Empty(t, mismatch.SplitPackages)
	assert.Empty(t, mismatch.MixedCommunities)
	assert.Equal(t, 0, metrics.Communities[0].PackageCount)
}

func communityPartitionsByID(partitions []CommunityPartition) map[string]CommunityPartition {
	out := make(map[string]CommunityPartition, len(partitions))
	for _, partition := range partitions {
		out[partition.ID] = partition
	}
	return out
}
