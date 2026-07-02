package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

// communityGraphNodeThreshold is the default module-count limit above which the
// macro-architecture visualization collapses module-level nodes into one node
// per community. Keeping the rendered node count bounded keeps the client-side
// force layout responsive on large codebases. Documented in
// website/docs/guides/module-community-detection.md.
const communityGraphNodeThreshold = 100

// communityGraphNode is a single rendered node in the macro-architecture graph.
// In module mode it is a module; in collapsed mode it is a whole community.
type communityGraphNode struct {
	ID              string   `json:"id"`
	Label           string   `json:"label"`
	Community       string   `json:"community"`
	Color           string   `json:"color"`
	Bridge          bool     `json:"bridge,omitempty"`
	Group           bool     `json:"group,omitempty"`
	Count           int      `json:"count,omitempty"`
	CrossEdges      int      `json:"crossEdges,omitempty"`
	Targets         []string `json:"targets,omitempty"`
	DominantPackage string   `json:"dominantPackage,omitempty"`
	DominantLayer   string   `json:"dominantLayer,omitempty"`
}

// communityGraphEdge is a directed edge between two rendered nodes.
type communityGraphEdge struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Cross  bool   `json:"cross,omitempty"`
	Weight int    `json:"weight,omitempty"`
}

// communityGraphCommunity describes a community for the legend / filter list.
type communityGraphCommunity struct {
	ID              string `json:"id"`
	Color           string `json:"color"`
	Size            int    `json:"size"`
	DominantPackage string `json:"dominantPackage,omitempty"`
	DominantLayer   string `json:"dominantLayer,omitempty"`
}

// communityGraphData is the compact JSON blob embedded in the HTML report and
// consumed by the client-side renderer.
type communityGraphData struct {
	Nodes        []communityGraphNode      `json:"nodes"`
	Edges        []communityGraphEdge      `json:"edges"`
	Communities  []communityGraphCommunity `json:"communities"`
	Collapsed    bool                      `json:"collapsed"`
	Threshold    int                       `json:"threshold"`
	TotalModules int                       `json:"totalModules"`
}

// BuildCommunityGraphData converts a community analysis result into the compact
// graph payload used by the HTML visualization. When the module count exceeds
// threshold, the graph collapses to one node per community so the layout stays
// responsive. A threshold <= 0 falls back to communityGraphNodeThreshold.
func BuildCommunityGraphData(response *domain.CommunityAnalysisResult, threshold int) communityGraphData {
	if threshold <= 0 {
		threshold = communityGraphNodeThreshold
	}
	data := communityGraphData{Threshold: threshold}
	if response == nil {
		return data
	}

	moduleCommunity := make(map[string]string)
	communityColor := make(map[string]string)
	sorted := sortedCommunitiesByID(response.Communities)
	for i, community := range sorted {
		color := communityDOTColors[i%len(communityDOTColors)]
		communityColor[community.ID] = color
		for _, module := range community.Modules {
			moduleCommunity[module] = community.ID
		}
		data.Communities = append(data.Communities, communityGraphCommunity{
			ID:              community.ID,
			Color:           color,
			Size:            community.Size,
			DominantPackage: community.DominantPackage,
			DominantLayer:   community.DominantLayer,
		})
	}
	data.TotalModules = len(moduleCommunity)

	bridges := make(map[string]domain.BridgeModule)
	for _, bridge := range response.BridgeModules {
		bridges[bridge.Module] = bridge
	}

	if data.TotalModules > threshold {
		data.Collapsed = true
		buildCollapsedGraph(&data, response, moduleCommunity, communityColor)
		return data
	}

	buildModuleGraph(&data, response, sorted, moduleCommunity, communityColor, bridges)
	return data
}

// buildModuleGraph populates module-level nodes and edges.
func buildModuleGraph(
	data *communityGraphData,
	response *domain.CommunityAnalysisResult,
	sorted []domain.CommunityMetrics,
	moduleCommunity map[string]string,
	communityColor map[string]string,
	bridges map[string]domain.BridgeModule,
) {
	for _, community := range sorted {
		for _, module := range sortedStringCopy(community.Modules) {
			node := communityGraphNode{
				ID:              module,
				Label:           module,
				Community:       community.ID,
				Color:           communityColor[community.ID],
				DominantPackage: community.DominantPackage,
				DominantLayer:   community.DominantLayer,
			}
			if bridge, ok := bridges[module]; ok {
				node.Bridge = true
				node.CrossEdges = bridge.CrossCommunityEdges
				node.Targets = bridge.TargetCommunities
			}
			data.Nodes = append(data.Nodes, node)
		}
	}

	for _, dep := range response.ModuleDependencies {
		fromCommunity, fromOK := moduleCommunity[dep.From]
		toCommunity, toOK := moduleCommunity[dep.To]
		if !fromOK || !toOK {
			continue
		}
		data.Edges = append(data.Edges, communityGraphEdge{
			From:  dep.From,
			To:    dep.To,
			Cross: fromCommunity != toCommunity,
		})
	}
}

// buildCollapsedGraph populates one node per community with aggregated
// cross-community edges, used when the module count exceeds the threshold.
func buildCollapsedGraph(
	data *communityGraphData,
	response *domain.CommunityAnalysisResult,
	moduleCommunity map[string]string,
	communityColor map[string]string,
) {
	for _, community := range sortedCommunitiesByID(response.Communities) {
		data.Nodes = append(data.Nodes, communityGraphNode{
			ID:              community.ID,
			Label:           community.ID,
			Community:       community.ID,
			Color:           communityColor[community.ID],
			Group:           true,
			Count:           community.Size,
			DominantPackage: community.DominantPackage,
			DominantLayer:   community.DominantLayer,
		})
	}

	weights := make(map[string]int)
	order := make([]string, 0)
	for _, dep := range response.ModuleDependencies {
		from, fromOK := moduleCommunity[dep.From]
		to, toOK := moduleCommunity[dep.To]
		if !fromOK || !toOK || from == to {
			continue
		}
		key := from + "\x00" + to
		if _, seen := weights[key]; !seen {
			order = append(order, key)
		}
		weights[key]++
	}
	for _, key := range order {
		parts := strings.SplitN(key, "\x00", 2)
		data.Edges = append(data.Edges, communityGraphEdge{
			From:   parts[0],
			To:     parts[1],
			Cross:  true,
			Weight: weights[key],
		})
	}
}

// writeCommunityGraphHTML renders the macro-architecture graph section: a data
// blob, the interactive container, and a self-contained renderer script. It is
// safe to embed in both the standalone community report and the unified analyze
// report (the script is wrapped in an IIFE and scoped to a fixed container id).
func writeCommunityGraphHTML(builder *strings.Builder, response *domain.CommunityAnalysisResult) {
	if response == nil || len(response.Communities) == 0 {
		return
	}

	data := BuildCommunityGraphData(response, communityGraphNodeThreshold)
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}

	builder.WriteString(GenerateSectionHeader("Macro Architecture"))
	builder.WriteString(`<p class="community-graph-help">Module dependency graph colored by community. ` +
		`Bridge modules are outlined in <span style="color:#d9534f;font-weight:600;">red</span>; ` +
		`red edges cross community boundaries. Hover a node for details; click to highlight its neighbors.</p>`)

	if data.Collapsed {
		fmt.Fprintf(builder,
			`<p class="community-graph-help"><em>%d modules exceed the %d-module display threshold; `+
				`showing one node per community. Each node is sized by module count.</em></p>`,
			data.TotalModules, data.Threshold,
		)
	}

	builder.WriteString(`
        <div id="community-graph" class="community-graph">
            <div class="community-graph-controls">
                <label>Community
                    <select id="community-graph-filter"><option value="">All</option></select>
                </label>
                <label><input type="checkbox" id="community-graph-bridges"> Bridges only</label>
                <label><input type="checkbox" id="community-graph-isolated"> Hide isolated</label>
                <span id="community-graph-count" class="community-graph-count"></span>
            </div>
            <div class="community-graph-canvas-wrap">
                <canvas id="community-graph-canvas" width="800" height="480" role="img"
                    aria-label="Module community dependency graph"></canvas>
                <div id="community-graph-tooltip" class="community-graph-tooltip" hidden></div>
            </div>
            <div id="community-graph-legend" class="community-graph-legend"></div>
        </div>`)

	builder.WriteString(`<script type="application/json" id="community-graph-data">`)
	builder.Write(payload)
	builder.WriteString(`</script>`)
	builder.WriteString(communityGraphStyle)
	builder.WriteString(communityGraphScript)
}
