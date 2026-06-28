package service

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// graphFixture builds a small 2-community + bridge result mirroring the
// testdata/python/community_bridge fixture.
func graphFixtureResult() *domain.CommunityAnalysisResult {
	return &domain.CommunityAnalysisResult{
		Algorithm:        "leiden",
		Scope:            "module",
		TotalCommunities: 2,
		Modularity:       0.42,
		Communities: []domain.CommunityMetrics{
			{ID: "community_1", Size: 3, Modules: []string{"bridge", "mod.a", "mod.b"}},
			{ID: "community_2", Size: 2, Modules: []string{"mod.c", "mod.d"}},
		},
		BridgeModules: []domain.BridgeModule{
			{Module: "bridge", Community: "community_1", CrossCommunityEdges: 2, TargetCommunities: []string{"community_2"}},
			{Module: "mod.c", Community: "community_2", CrossCommunityEdges: 1, TargetCommunities: []string{"community_1"}},
		},
		ModuleDependencies: []domain.CommunityModuleDependency{
			{From: "mod.a", To: "mod.b"},
			{From: "bridge", To: "mod.c"},
			{From: "mod.c", To: "mod.d"},
			{From: "mod.b", To: "bridge"},
		},
	}
}

func extractGraphData(t *testing.T, html string) communityGraphData {
	t.Helper()
	re := regexp.MustCompile(`(?s)<script type="application/json" id="community-graph-data">(.*?)</script>`)
	m := re.FindStringSubmatch(html)
	require.Len(t, m, 2, "embedded graph data blob not found")
	var data communityGraphData
	require.NoError(t, json.Unmarshal([]byte(m[1]), &data))
	return data
}

func TestBuildCommunityGraphData_ModuleLevel(t *testing.T) {
	data := BuildCommunityGraphData(graphFixtureResult(), 0)

	assert.False(t, data.Collapsed)
	assert.Equal(t, 100, data.Threshold)
	assert.Equal(t, 5, data.TotalModules)
	assert.Len(t, data.Nodes, 5)
	assert.Len(t, data.Edges, 4)
	assert.Len(t, data.Communities, 2)

	bridges := map[string]bool{}
	colors := map[string]string{}
	for _, n := range data.Nodes {
		if n.Bridge {
			bridges[n.ID] = true
		}
		colors[n.ID] = n.Color
	}
	assert.True(t, bridges["bridge"], "bridge module should be flagged")
	assert.True(t, bridges["mod.c"], "mod.c should be flagged as bridge")
	assert.False(t, bridges["mod.a"])
	// Nodes in the same community share a color; different communities differ.
	assert.Equal(t, colors["mod.a"], colors["mod.b"])
	assert.NotEqual(t, colors["mod.a"], colors["mod.c"])

	// The bridge -> mod.c edge crosses communities.
	var crossFound bool
	for _, e := range data.Edges {
		if e.From == "bridge" && e.To == "mod.c" {
			crossFound = e.Cross
		}
	}
	assert.True(t, crossFound, "bridge->mod.c should be a cross-community edge")
}

func TestBuildCommunityGraphData_CollapsedAboveThreshold(t *testing.T) {
	// Threshold of 3 forces the 5-module fixture to collapse to communities.
	data := BuildCommunityGraphData(graphFixtureResult(), 3)

	assert.True(t, data.Collapsed)
	assert.Len(t, data.Nodes, 2, "collapsed graph has one node per community")
	for _, n := range data.Nodes {
		assert.True(t, n.Group)
		assert.Greater(t, n.Count, 0)
	}
	// Only cross-community edges survive collapse, aggregated with weights.
	require.NotEmpty(t, data.Edges)
	for _, e := range data.Edges {
		assert.True(t, e.Cross)
		assert.Greater(t, e.Weight, 0)
	}
}

func TestCommunityFormatter_HTML_ContainsGraph(t *testing.T) {
	formatter := NewCommunityFormatter()
	html, err := formatter.Format(graphFixtureResult(), domain.OutputFormatHTML)
	require.NoError(t, err)

	assert.Contains(t, html, `id="community-graph"`)
	assert.Contains(t, html, `id="community-graph-canvas"`)
	assert.Contains(t, html, `id="community-graph-data"`)
	assert.Contains(t, html, "Macro Architecture")
	// Existing accessible tables remain.
	assert.Contains(t, html, "Bridge Modules")

	data := extractGraphData(t, html)
	assert.Equal(t, 5, len(data.Nodes), "node count should match the fixture")
	assert.False(t, data.Collapsed)

	var bridgeMarked bool
	for _, n := range data.Nodes {
		if n.ID == "bridge" {
			bridgeMarked = n.Bridge
		}
	}
	assert.True(t, bridgeMarked, "bridge module must be marked in the graph data")
}

func TestBuildCommunityGraphData_Nil(t *testing.T) {
	data := BuildCommunityGraphData(nil, 0)
	assert.Empty(t, data.Nodes)
	assert.Equal(t, 100, data.Threshold)
}

func TestWriteCommunityGraphHTML_EmptyCommunitiesNoOutput(t *testing.T) {
	var b strings.Builder
	writeCommunityGraphHTML(&b, &domain.CommunityAnalysisResult{})
	assert.Empty(t, b.String(), "no graph section when there are no communities")
}
