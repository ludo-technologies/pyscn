package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
)

const (
	communitySummaryLimit = 5
)

var communityDOTColors = []string{
	"lightblue",
	"lightgreen",
	"lightyellow",
	"lightpink",
	"lavender",
	"peachpuff",
	"lightgray",
	"honeydew",
}

// CommunityFormatter formats community analysis results for output.
type CommunityFormatter struct{}

// NewCommunityFormatter creates a community analysis formatter.
func NewCommunityFormatter() *CommunityFormatter {
	return &CommunityFormatter{}
}

// Format formats community analysis results according to the requested output format.
func (f *CommunityFormatter) Format(response *domain.CommunityAnalysisResult, format domain.OutputFormat) (string, error) {
	if response == nil {
		return "", fmt.Errorf("community analysis result is nil")
	}

	switch format {
	case domain.OutputFormatText:
		return f.formatText(response), nil
	case domain.OutputFormatJSON:
		return f.formatJSON(response)
	case domain.OutputFormatYAML:
		return f.formatYAML(response)
	case domain.OutputFormatCSV:
		return f.formatCSV(response)
	case domain.OutputFormatHTML:
		return f.formatHTML(response)
	case domain.OutputFormatDOT:
		return f.formatDOT(response)
	default:
		return "", fmt.Errorf("community formatter does not support format %s", format)
	}
}

// Write formats and writes community analysis output.
func (f *CommunityFormatter) Write(response *domain.CommunityAnalysisResult, format domain.OutputFormat, writer io.Writer) error {
	if response == nil {
		return fmt.Errorf("community analysis result is nil")
	}

	switch format {
	case domain.OutputFormatJSON:
		return WriteJSON(writer, normalizeCommunityResult(response))
	case domain.OutputFormatYAML:
		return WriteYAML(writer, normalizeCommunityResult(response))
	default:
		content, err := f.Format(response, format)
		if err != nil {
			return err
		}
		_, err = io.WriteString(writer, content)
		return err
	}
}

func (f *CommunityFormatter) formatText(response *domain.CommunityAnalysisResult) string {
	var builder strings.Builder
	utils := NewFormatUtils()
	builder.WriteString(utils.FormatMainHeader("Module Community Analysis"))
	f.writeTextSummary(&builder, response, utils)
	return builder.String()
}

// WriteCommunityTextSummary writes a concise community section for unified analyze text output.
func WriteCommunityTextSummary(writer io.Writer, response *domain.CommunityAnalysisResult) {
	if response == nil {
		return
	}

	var builder strings.Builder
	utils := NewFormatUtils()
	builder.WriteString(utils.FormatSectionHeader("COMMUNITY DETECTION"))
	f := &CommunityFormatter{}
	f.writeTextSummary(&builder, response, utils)
	_, _ = io.WriteString(writer, builder.String())
}

func (f *CommunityFormatter) writeTextSummary(builder *strings.Builder, response *domain.CommunityAnalysisResult, utils *FormatUtils) {
	stats := map[string]interface{}{
		"Total Communities": response.TotalCommunities,
		"Modularity (Q)":    fmt.Sprintf("%.3f", response.Modularity),
		"Algorithm":         response.Algorithm,
		"Scope":             response.Scope,
	}
	if response.PackageAlignmentScore != nil {
		stats["Package Alignment"] = fmt.Sprintf("%.3f", *response.PackageAlignmentScore)
	}
	builder.WriteString(utils.FormatSummaryStats(stats))

	if len(response.SplitPackages) > 0 || len(response.MixedCommunities) > 0 {
		builder.WriteString(utils.FormatSectionHeader("PACKAGE MISMATCH"))
		if len(response.SplitPackages) > 0 {
			builder.WriteString(utils.FormatLabelWithIndent(
				SectionPadding,
				"Split Packages",
				strings.Join(response.SplitPackages, ", "),
			))
		}
		if len(response.MixedCommunities) > 0 {
			builder.WriteString(utils.FormatLabelWithIndent(
				SectionPadding,
				"Mixed Communities",
				strings.Join(response.MixedCommunities, ", "),
			))
		}
		builder.WriteString("\n")
	}

	communities := communitiesBySize(response.Communities)
	if len(communities) > 0 {
		builder.WriteString(utils.FormatSectionHeader("LARGEST COMMUNITIES"))
		limit := min(len(communities), communitySummaryLimit)
		for i := 0; i < limit; i++ {
			community := communities[i]
			detail := fmt.Sprintf(
				"%d modules (internal: %d, external: %d, cross-in: %d, cross-out: %d)",
				community.Size,
				community.InternalEdges,
				community.ExternalEdges,
				community.IncomingCrossCommunityEdges,
				community.OutgoingCrossCommunityEdges,
			)
			if community.PackageCount > 0 {
				detail += fmt.Sprintf(
					", pkg-align: %.3f (%s, %d pkgs)",
					community.PackageAlignment,
					community.DominantPackage,
					community.PackageCount,
				)
			}
			builder.WriteString(utils.FormatLabelWithIndent(
				SectionPadding,
				community.ID,
				detail,
			))
		}
		if len(communities) > limit {
			builder.WriteString(utils.FormatLabelWithIndent(
				SectionPadding,
				"...",
				fmt.Sprintf("and %d more communities", len(communities)-limit),
			))
		}
		builder.WriteString("\n")
	}

	bridges := bridgesByCoupling(response.BridgeModules)
	if len(bridges) > 0 {
		builder.WriteString(utils.FormatSectionHeader("BRIDGE MODULES"))
		limit := min(len(bridges), communitySummaryLimit)
		for i := 0; i < limit; i++ {
			bridge := bridges[i]
			builder.WriteString(utils.FormatLabelWithIndent(
				SectionPadding,
				bridge.Module,
				fmt.Sprintf(
					"%d cross-community edges in %s -> %s",
					bridge.CrossCommunityEdges,
					bridge.Community,
					strings.Join(bridge.TargetCommunities, ", "),
				),
			))
		}
		if len(bridges) > limit {
			builder.WriteString(utils.FormatLabelWithIndent(
				SectionPadding,
				"...",
				fmt.Sprintf("and %d more bridge modules", len(bridges)-limit),
			))
		}
		builder.WriteString("\n")
	}

	if len(response.Warnings) > 0 {
		builder.WriteString(utils.FormatWarningsSection(response.Warnings))
	}

	builder.WriteString(utils.FormatSectionSeparator())
}

func (f *CommunityFormatter) formatJSON(response *domain.CommunityAnalysisResult) (string, error) {
	normalized := normalizeCommunityResult(response)
	data, err := EncodeJSON(normalized)
	if err != nil {
		return "", err
	}
	return data + "\n", nil
}

func (f *CommunityFormatter) formatYAML(response *domain.CommunityAnalysisResult) (string, error) {
	return EncodeYAML(normalizeCommunityResult(response))
}

func (f *CommunityFormatter) formatCSV(response *domain.CommunityAnalysisResult) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	if err := writer.Write([]string{"Section", "Metric", "Value"}); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	summaryRows := [][]string{
		{"Summary", "Algorithm", response.Algorithm},
		{"Summary", "Scope", response.Scope},
		{"Summary", "Total Communities", strconv.Itoa(response.TotalCommunities)},
		{"Summary", "Modularity", fmt.Sprintf("%.4f", response.Modularity)},
		{"Summary", "Bridge Modules", strconv.Itoa(len(response.BridgeModules))},
	}
	for _, row := range summaryRows {
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write CSV summary row: %w", err)
		}
	}

	for _, community := range sortedCommunitiesByID(response.Communities) {
		rows := [][]string{
			{"Community", community.ID + " Size", strconv.Itoa(community.Size)},
			{"Community", community.ID + " Internal Edges", strconv.Itoa(community.InternalEdges)},
			{"Community", community.ID + " External Edges", strconv.Itoa(community.ExternalEdges)},
			{"Community", community.ID + " External Ratio", fmt.Sprintf("%.4f", community.ExternalDependencyRatio)},
			{"Community", community.ID + " Cross-In", strconv.Itoa(community.IncomingCrossCommunityEdges)},
			{"Community", community.ID + " Cross-Out", strconv.Itoa(community.OutgoingCrossCommunityEdges)},
			{"Community", community.ID + " Module Count", strconv.Itoa(len(community.Modules))},
		}
		for _, row := range rows {
			if err := writer.Write(row); err != nil {
				return "", fmt.Errorf("failed to write CSV community row: %w", err)
			}
		}
	}

	for _, bridge := range sortedBridgeModules(response.BridgeModules) {
		rows := [][]string{
			{"Bridge", bridge.Module + " Community", bridge.Community},
			{"Bridge", bridge.Module + " Cross Edges", strconv.Itoa(bridge.CrossCommunityEdges)},
			{"Bridge", bridge.Module + " Targets", strings.Join(bridge.TargetCommunities, ";")},
		}
		for _, row := range rows {
			if err := writer.Write(row); err != nil {
				return "", fmt.Errorf("failed to write CSV bridge row: %w", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %w", err)
	}
	return builder.String(), nil
}

func (f *CommunityFormatter) formatHTML(response *domain.CommunityAnalysisResult) (string, error) {
	var builder strings.Builder

	generatedAt, _ := time.Parse(time.RFC3339, response.GeneratedAt)
	if generatedAt.IsZero() {
		generatedAt = time.Now()
	}

	template := HTMLTemplate{
		Title:       "Community Analysis Report",
		Subtitle:    "Module Community Detection Summary",
		GeneratedAt: generatedAt,
		Version:     response.Version,
		ShowScore:   false,
	}

	builder.WriteString(template.GenerateHTMLHeader())

	var content strings.Builder
	f.writeHTMLSummary(&content, response)
	builder.WriteString(GenerateSinglePageContent(content.String()))
	builder.WriteString(GenerateHTMLFooter())

	return builder.String(), nil
}

// WriteCommunityHTMLSummary writes a concise community section for unified analyze HTML output.
func WriteCommunityHTMLSummary(builder *strings.Builder, response *domain.CommunityAnalysisResult) {
	if response == nil {
		return
	}
	formatter := &CommunityFormatter{}
	formatter.writeHTMLSummary(builder, response)
}

func (f *CommunityFormatter) writeHTMLSummary(builder *strings.Builder, response *domain.CommunityAnalysisResult) {
	builder.WriteString(GenerateSectionHeader("Module Communities"))
	builder.WriteString(`<div class="metric-grid">`)
	builder.WriteString(GenerateMetricCard(strconv.Itoa(response.TotalCommunities), "Communities"))
	builder.WriteString(GenerateMetricCard(fmt.Sprintf("%.3f", response.Modularity), "Modularity (Q)"))
	builder.WriteString(GenerateMetricCard(response.Algorithm, "Algorithm"))
	builder.WriteString(GenerateMetricCard(strconv.Itoa(len(response.BridgeModules)), "Bridge Modules"))
	if response.PackageAlignmentScore != nil {
		builder.WriteString(GenerateMetricCard(fmt.Sprintf("%.3f", *response.PackageAlignmentScore), "Package Alignment"))
	}
	builder.WriteString(`</div>`)

	if len(response.SplitPackages) > 0 || len(response.MixedCommunities) > 0 {
		builder.WriteString(GenerateSectionHeader("Package Mismatch"))
		builder.WriteString(`<ul>`)
		if len(response.SplitPackages) > 0 {
			builder.WriteString(fmt.Sprintf(
				`<li><strong>Split packages:</strong> %s</li>`,
				JoinEscapedHTML(response.SplitPackages, ", "),
			))
		}
		if len(response.MixedCommunities) > 0 {
			builder.WriteString(fmt.Sprintf(
				`<li><strong>Mixed communities:</strong> %s</li>`,
				JoinEscapedHTML(response.MixedCommunities, ", "),
			))
		}
		builder.WriteString(`</ul>`)
	}

	communities := communitiesBySize(response.Communities)
	if len(communities) > 0 {
		builder.WriteString(GenerateSectionHeader("Largest Communities"))
		builder.WriteString(`
            <table class="table">
                <thead>
                    <tr>
                        <th>Community</th>
                        <th>Modules</th>
                        <th>Internal</th>
                        <th>External</th>
                        <th>Cross-In</th>
                        <th>Cross-Out</th>
                    </tr>
                </thead>
                <tbody>`)
		limit := min(len(communities), communitySummaryLimit)
		for i := 0; i < limit; i++ {
			community := communities[i]
			builder.WriteString(fmt.Sprintf(`
                    <tr>
                        <td><strong>%s</strong></td>
                        <td>%d</td>
                        <td>%d</td>
                        <td>%d</td>
                        <td>%d</td>
                        <td>%d</td>
                    </tr>`,
				EscapeHTML(community.ID),
				community.Size,
				community.InternalEdges,
				community.ExternalEdges,
				community.IncomingCrossCommunityEdges,
				community.OutgoingCrossCommunityEdges,
			))
		}
		if len(communities) > limit {
			builder.WriteString(fmt.Sprintf(`
                    <tr>
                        <td colspan="6"><em>... and %d more communities</em></td>
                    </tr>`, len(communities)-limit))
		}
		builder.WriteString(`
                </tbody>
            </table>`)
	}

	bridges := bridgesByCoupling(response.BridgeModules)
	if len(bridges) > 0 {
		builder.WriteString(GenerateSectionHeader("Bridge Modules"))
		builder.WriteString(`
            <table class="table">
                <thead>
                    <tr>
                        <th>Module</th>
                        <th>Community</th>
                        <th>Cross Edges</th>
                        <th>Target Communities</th>
                    </tr>
                </thead>
                <tbody>`)
		limit := min(len(bridges), communitySummaryLimit)
		for i := 0; i < limit; i++ {
			bridge := bridges[i]
			builder.WriteString(fmt.Sprintf(`
                    <tr>
                        <td><code>%s</code></td>
                        <td>%s</td>
                        <td>%d</td>
                        <td>%s</td>
                    </tr>`,
				EscapeHTML(bridge.Module),
				EscapeHTML(bridge.Community),
				bridge.CrossCommunityEdges,
				JoinEscapedHTML(bridge.TargetCommunities, ", "),
			))
		}
		if len(bridges) > limit {
			builder.WriteString(fmt.Sprintf(`
                    <tr>
                        <td colspan="4"><em>... and %d more bridge modules</em></td>
                    </tr>`, len(bridges)-limit))
		}
		builder.WriteString(`
                </tbody>
            </table>`)
	}
}

func (f *CommunityFormatter) formatDOT(response *domain.CommunityAnalysisResult) (string, error) {
	if response == nil {
		return "", fmt.Errorf("community analysis result is nil")
	}

	moduleCommunity := make(map[string]string)
	bridgeModules := make(map[string]bool)
	for _, bridge := range response.BridgeModules {
		bridgeModules[bridge.Module] = true
	}

	communityColors := make(map[string]string)
	for i, community := range sortedCommunitiesByID(response.Communities) {
		color := communityDOTColors[i%len(communityDOTColors)]
		communityColors[community.ID] = color
		for _, module := range community.Modules {
			moduleCommunity[module] = community.ID
		}
	}

	var builder strings.Builder
	builder.WriteString("digraph ModuleCommunities {\n")
	builder.WriteString("  rankdir=LR;\n")
	builder.WriteString("  node [shape=box, style=filled];\n")
	builder.WriteString("  edge [color=gray];\n\n")

	for _, community := range sortedCommunitiesByID(response.Communities) {
		clusterID := communityDOTClusterID(community.ID)
		color := communityColors[community.ID]
		builder.WriteString(fmt.Sprintf("  subgraph %s {\n", clusterID))
		builder.WriteString(fmt.Sprintf("    label=\"%s\";\n", community.ID))
		builder.WriteString("    style=filled;\n")
		builder.WriteString(fmt.Sprintf("    color=%s;\n", color))
		for _, module := range sortedStringCopy(community.Modules) {
			nodeColor := color
			if bridgeModules[module] {
				nodeColor = "lightcoral"
			}
			builder.WriteString(fmt.Sprintf(
				"    %s [label=\"%s\", fillcolor=%s];\n",
				communityDOTNodeID(module),
				module,
				nodeColor,
			))
		}
		builder.WriteString("  }\n\n")
	}

	for _, dep := range response.ModuleDependencies {
		fromCommunity := moduleCommunity[dep.From]
		toCommunity := moduleCommunity[dep.To]
		edgeColor := "gray"
		if fromCommunity != "" && toCommunity != "" && fromCommunity != toCommunity {
			edgeColor = "red"
		}
		builder.WriteString(fmt.Sprintf(
			"  %s -> %s [color=%s];\n",
			communityDOTNodeID(dep.From),
			communityDOTNodeID(dep.To),
			edgeColor,
		))
	}

	builder.WriteString("\n  subgraph cluster_legend {\n")
	builder.WriteString("    label=\"Legend\";\n")
	builder.WriteString("    style=filled;\n")
	builder.WriteString("    fillcolor=white;\n")
	builder.WriteString("    legend_bridge [label=\"Bridge Module\", fillcolor=lightcoral, shape=box];\n")
	builder.WriteString("    legend_cross [label=\"Cross-Community Edge\", color=red, shape=plaintext];\n")
	builder.WriteString("  }\n")
	builder.WriteString("}\n")

	return builder.String(), nil
}

func communitiesBySize(communities []domain.CommunityMetrics) []domain.CommunityMetrics {
	out := append([]domain.CommunityMetrics(nil), communities...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Size == out[j].Size {
			return out[i].ID < out[j].ID
		}
		return out[i].Size > out[j].Size
	})
	return out
}

func sortedCommunitiesByID(communities []domain.CommunityMetrics) []domain.CommunityMetrics {
	out := append([]domain.CommunityMetrics(nil), communities...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func bridgesByCoupling(bridges []domain.BridgeModule) []domain.BridgeModule {
	out := append([]domain.BridgeModule(nil), bridges...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].CrossCommunityEdges == out[j].CrossCommunityEdges {
			return out[i].Module < out[j].Module
		}
		return out[i].CrossCommunityEdges > out[j].CrossCommunityEdges
	})
	return out
}

func sortedBridgeModules(bridges []domain.BridgeModule) []domain.BridgeModule {
	out := append([]domain.BridgeModule(nil), bridges...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Module == out[j].Module {
			return out[i].Community < out[j].Community
		}
		return out[i].Module < out[j].Module
	})
	return out
}

func communityDOTNodeID(module string) string {
	clean := strings.ReplaceAll(module, ".", "_")
	clean = strings.ReplaceAll(clean, "-", "_")
	return clean
}

func communityDOTClusterID(communityID string) string {
	return "cluster_" + communityDOTNodeID(communityID)
}

// normalizeCommunityResult returns a schema-stable copy with sorted lists and rounded ratios.
func normalizeCommunityResult(response *domain.CommunityAnalysisResult) *domain.CommunityAnalysisResult {
	if response == nil {
		return nil
	}

	out := *response
	out.Modularity = roundCommunityFloat(response.Modularity)
	if response.PackageAlignmentScore != nil {
		score := roundCommunityFloat(*response.PackageAlignmentScore)
		out.PackageAlignmentScore = &score
	}
	out.SplitPackages = sortedStringCopy(response.SplitPackages)
	out.MixedCommunities = sortedStringCopy(response.MixedCommunities)

	communities := make([]domain.CommunityMetrics, len(response.Communities))
	copy(communities, response.Communities)
	sort.Slice(communities, func(i, j int) bool {
		return communities[i].ID < communities[j].ID
	})

	for i := range communities {
		communities[i].Modules = sortedStringCopy(communities[i].Modules)
		communities[i].Packages = sortedStringCopy(communities[i].Packages)
		communities[i].ExternalDependencyRatio = roundCommunityFloat(communities[i].ExternalDependencyRatio)
		if communities[i].PackageCount > 0 {
			communities[i].PackageAlignment = roundCommunityFloat(communities[i].PackageAlignment)
		}
	}
	out.Communities = communities

	bridges := make([]domain.BridgeModule, len(response.BridgeModules))
	copy(bridges, response.BridgeModules)
	sort.Slice(bridges, func(i, j int) bool {
		if bridges[i].Module == bridges[j].Module {
			return bridges[i].Community < bridges[j].Community
		}
		return bridges[i].Module < bridges[j].Module
	})
	for i := range bridges {
		bridges[i].TargetCommunities = sortedStringCopy(bridges[i].TargetCommunities)
	}
	out.BridgeModules = bridges
	out.ModuleDependencies = nil

	return &out
}

func sortedStringCopy(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	out := append([]string(nil), values...)
	sort.Strings(out)
	return out
}

func roundCommunityFloat(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return math.Round(value*10000) / 10000
}

func normalizeAnalyzeResponseForJSON(response *domain.AnalyzeResponse) *domain.AnalyzeResponse {
	if response == nil {
		return nil
	}
	if response.Communities == nil {
		return response
	}

	out := *response
	out.Communities = normalizeCommunityResult(response.Communities)
	return &out
}
