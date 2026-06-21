package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestCommunityGraph(nodes []string, edges [][2]string) *CommunityGraph {
	graph := NewDependencyGraph("/project")
	for _, node := range nodes {
		graph.AddModule(node, "/project/"+node+".py")
	}
	for _, edge := range edges {
		graph.AddDependency(edge[0], edge[1], DependencyEdgeImport, nil)
	}
	return BuildCommunityGraph(graph, nil)
}

func groupNodesByCommunity(cg *CommunityGraph, result *LeidenResult) map[int][]string {
	groups := make(map[int][]string)
	for i, comm := range result.Membership {
		groups[comm] = append(groups[comm], cg.NodeNames[i])
	}
	return groups
}

func assertSamePartition(t *testing.T, cg *CommunityGraph, result *LeidenResult, expectedGroups [][]string) {
	t.Helper()
	groups := groupNodesByCommunity(cg, result)
	require.Equal(t, len(expectedGroups), len(groups), "community count mismatch: %v", groups)

	matched := make([]bool, len(expectedGroups))
	for _, members := range groups {
		found := false
		for i, expected := range expectedGroups {
			if matched[i] {
				continue
			}
			if sameStringSet(members, expected) {
				matched[i] = true
				found = true
				break
			}
		}
		require.True(t, found, "unexpected community %v", members)
	}
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	setA := make(map[string]struct{}, len(a))
	for _, v := range a {
		setA[v] = struct{}{}
	}
	for _, v := range b {
		if _, ok := setA[v]; !ok {
			return false
		}
	}
	return true
}

func TestDetectCommunitiesLeiden_EmptyGraph(t *testing.T) {
	t.Run("nil graph", func(t *testing.T) {
		result := DetectCommunitiesLeiden(nil, nil)
		assert.Empty(t, result.Membership)
		assert.Equal(t, 0, result.NumCommunities)
		assert.Equal(t, 0.0, result.Modularity)
	})

	t.Run("no nodes", func(t *testing.T) {
		cg := BuildCommunityGraph(NewDependencyGraph("/project"), nil)
		result := DetectCommunitiesLeiden(cg, nil)
		assert.Empty(t, result.Membership)
		assert.Equal(t, 0, result.NumCommunities)
		assert.Equal(t, 0.0, result.Modularity)
	})
}

func TestDetectCommunitiesLeiden_TwoCliques(t *testing.T) {
	nodes := []string{"a1", "a2", "a3", "b1", "b2", "b3"}
	edges := [][2]string{
		{"a1", "a2"}, {"a2", "a1"},
		{"a1", "a3"}, {"a3", "a1"},
		{"a2", "a3"}, {"a3", "a2"},
		{"b1", "b2"}, {"b2", "b1"},
		{"b1", "b3"}, {"b3", "b1"},
		{"b2", "b3"}, {"b3", "b2"},
	}
	cg := buildTestCommunityGraph(nodes, edges)

	result := DetectCommunitiesLeiden(cg, nil)
	require.Len(t, result.Membership, 6)
	assert.Equal(t, 2, result.NumCommunities)
	assert.Greater(t, result.Modularity, 0.3)
	assertSamePartition(t, cg, result, [][]string{
		{"a1", "a2", "a3"},
		{"b1", "b2", "b3"},
	})
}

func TestDetectCommunitiesLeiden_Barbell(t *testing.T) {
	nodes := []string{"a1", "a2", "bridge", "b1", "b2"}
	edges := [][2]string{
		{"a1", "a2"}, {"a2", "a1"},
		{"a2", "bridge"},
		{"bridge", "b1"},
		{"b1", "b2"}, {"b2", "b1"},
	}
	cg := buildTestCommunityGraph(nodes, edges)

	result := DetectCommunitiesLeiden(cg, nil)
	require.Len(t, result.Membership, 5)
	assert.Equal(t, 2, result.NumCommunities)
	assert.Greater(t, result.Modularity, 0.0)
	assertSamePartition(t, cg, result, [][]string{
		{"a1", "a2", "bridge"},
		{"b1", "b2"},
	})
}

func TestDetectCommunitiesLeiden_Cycle(t *testing.T) {
	nodes := []string{"n1", "n2", "n3", "n4", "n5"}
	edges := [][2]string{
		{"n1", "n2"}, {"n2", "n3"}, {"n3", "n4"}, {"n4", "n5"}, {"n5", "n1"},
	}
	cg := buildTestCommunityGraph(nodes, edges)

	result := DetectCommunitiesLeiden(cg, nil)
	require.Len(t, result.Membership, 5)
	assert.GreaterOrEqual(t, result.NumCommunities, 1)
	assert.LessOrEqual(t, result.NumCommunities, 2)
}

func TestDetectCommunitiesLeiden_Star(t *testing.T) {
	nodes := []string{"center", "s1", "s2", "s3", "s4"}
	edges := [][2]string{
		{"center", "s1"}, {"center", "s2"}, {"center", "s3"}, {"center", "s4"},
	}
	cg := buildTestCommunityGraph(nodes, edges)

	result := DetectCommunitiesLeiden(cg, nil)
	require.Len(t, result.Membership, 5)
	assert.Equal(t, 1, result.NumCommunities)
}

func TestDetectCommunitiesLeiden_DisconnectedComponents(t *testing.T) {
	nodes := []string{"a", "b", "c", "d", "e"}
	edges := [][2]string{
		{"a", "b"}, {"b", "a"},
		{"d", "e"}, {"e", "d"},
	}
	cg := buildTestCommunityGraph(nodes, edges)

	result := DetectCommunitiesLeiden(cg, nil)
	require.Len(t, result.Membership, 5)
	assert.Equal(t, 3, result.NumCommunities)
	assertSamePartition(t, cg, result, [][]string{
		{"a", "b"},
		{"c"},
		{"d", "e"},
	})
}

func TestDetectCommunitiesLeiden_Deterministic(t *testing.T) {
	nodes := []string{"z", "a", "m", "b", "y"}
	edges := [][2]string{
		{"a", "b"}, {"b", "m"}, {"m", "y"}, {"y", "z"}, {"z", "a"},
		{"a", "m"},
	}
	cg := buildTestCommunityGraph(nodes, edges)

	first := DetectCommunitiesLeiden(cg, nil)
	second := DetectCommunitiesLeiden(cg, nil)

	assert.Equal(t, first.Membership, second.Membership)
	assert.Equal(t, first.NumCommunities, second.NumCommunities)
	assert.InDelta(t, first.Modularity, second.Modularity, 1e-12)
}

func TestDetectCommunitiesLeiden_MinCommunitySize(t *testing.T) {
	nodes := []string{"a1", "a2", "b1", "b2", "iso"}
	edges := [][2]string{
		{"a1", "a2"}, {"a2", "a1"},
		{"b1", "b2"}, {"b2", "b1"},
	}
	cg := buildTestCommunityGraph(nodes, edges)

	result := DetectCommunitiesLeiden(cg, &LeidenOptions{
		Resolution:       1.0,
		MinCommunitySize: 2,
		MaxIterations:    64,
		MaxPasses:        16,
	})

	require.Len(t, result.Membership, 5)
	assert.Equal(t, 3, result.NumCommunities)
	assertSamePartition(t, cg, result, [][]string{
		{"a1", "a2"},
		{"b1", "b2"},
		{"iso"},
	})
}

func TestDetectCommunitiesLeiden_IsolatedNodesNoEdges(t *testing.T) {
	cg := buildTestCommunityGraph([]string{"solo1", "solo2", "solo3"}, nil)

	result := DetectCommunitiesLeiden(cg, nil)
	require.Len(t, result.Membership, 3)
	assert.Equal(t, 3, result.NumCommunities)
	assert.Equal(t, 0.0, result.Modularity)
	assert.Equal(t, []int{0, 1, 2}, result.Membership)
}
