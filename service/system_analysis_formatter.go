package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
	"gopkg.in/yaml.v3"
)

// SystemAnalysisFormatterImpl implements the SystemAnalysisOutputFormatter interface
type SystemAnalysisFormatterImpl struct{}

// NewSystemAnalysisFormatter creates a new system analysis output formatter
func NewSystemAnalysisFormatter() *SystemAnalysisFormatterImpl {
	return &SystemAnalysisFormatterImpl{}
}

// Format formats the system analysis response according to the specified format
func (f *SystemAnalysisFormatterImpl) Format(response *domain.SystemAnalysisResponse, format domain.OutputFormat) (string, error) {
	switch format {
	case domain.OutputFormatText:
		return f.formatText(response)
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
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

// Write writes the formatted output to the writer
func (f *SystemAnalysisFormatterImpl) Write(response *domain.SystemAnalysisResponse, format domain.OutputFormat, writer io.Writer) error {
	formatted, err := f.Format(response, format)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(formatted))
	return err
}

// formatText formats the response as human-readable text
func (f *SystemAnalysisFormatterImpl) formatText(response *domain.SystemAnalysisResponse) (string, error) {
	var builder strings.Builder
	utils := NewFormatUtils()

	// Main header
	builder.WriteString(utils.FormatMainHeader("System-Level Structural Quality Analysis"))

	// Dependencies section
	if response.DependencyAnalysis != nil {
		f.writeDependenciesSection(&builder, response.DependencyAnalysis, utils)
	}

	// Architecture section
	if response.ArchitectureAnalysis != nil {
		f.writeArchitectureSection(&builder, response.ArchitectureAnalysis, utils)
	}

	// Quality section
	if response.QualityMetrics != nil {
		f.writeQualitySection(&builder, response.QualityMetrics, utils)
	}

	// Warnings section
	if len(response.Warnings) > 0 {
		builder.WriteString(utils.FormatWarningsSection(response.Warnings))
	}

	// Errors section
	if len(response.Errors) > 0 {
		builder.WriteString(utils.FormatSectionHeader("ERRORS"))
		for _, err := range response.Errors {
			builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "❌", err))
		}
		builder.WriteString(utils.FormatSectionSeparator())
	}

	// Metadata section
	builder.WriteString(utils.FormatSectionHeader("METADATA"))
	builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Generated at", response.GeneratedAt))
	builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Duration", fmt.Sprintf("%dms", response.Duration)))
	builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Version", response.Version))

	return builder.String(), nil
}

// writeDependenciesSection writes the dependencies analysis section
func (f *SystemAnalysisFormatterImpl) writeDependenciesSection(builder *strings.Builder, deps *domain.DependencyAnalysisResult, utils *FormatUtils) {
	builder.WriteString(utils.FormatSectionHeader("DEPENDENCY ANALYSIS"))

	// Summary statistics
	stats := map[string]interface{}{
		"Total Modules":      deps.TotalModules,
		"Total Dependencies": deps.TotalDependencies,
		"Root Modules":       len(deps.RootModules),
		"Leaf Modules":       len(deps.LeafModules),
		"Max Depth":          deps.MaxDepth,
	}
	builder.WriteString(utils.FormatSummaryStats(stats))

	// Circular dependencies
	if deps.CircularDependencies != nil {
		builder.WriteString(utils.FormatSectionHeader("CIRCULAR DEPENDENCIES"))
		if deps.CircularDependencies.HasCircularDependencies {
			builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Status", fmt.Sprintf("⚠️  %d cycles found involving %d modules",
				deps.CircularDependencies.TotalCycles, deps.CircularDependencies.TotalModulesInCycles)))

			// List cycles
			for i, cycle := range deps.CircularDependencies.CircularDependencies {
				if i >= 5 { // Limit to top 5
					builder.WriteString(utils.FormatLabelWithIndent(SectionPadding*2, "...", fmt.Sprintf("and %d more cycles", len(deps.CircularDependencies.CircularDependencies)-i)))
					break
				}
				builder.WriteString(utils.FormatLabelWithIndent(SectionPadding*2, fmt.Sprintf("Cycle %d", i+1),
					fmt.Sprintf("%s (%d modules)", cycle.Description, len(cycle.Modules))))
			}

			// Cycle breaking suggestions
			if len(deps.CircularDependencies.CycleBreakingSuggestions) > 0 {
				builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Suggestions", ""))
				for _, suggestion := range deps.CircularDependencies.CycleBreakingSuggestions {
					builder.WriteString(utils.FormatLabelWithIndent(SectionPadding*2, "•", suggestion))
				}
			}
		} else {
			builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Status", "✅ No circular dependencies"))
		}
		builder.WriteString("\n")
	}

	// Coupling analysis
	if deps.CouplingAnalysis != nil {
		builder.WriteString(utils.FormatSectionHeader("COUPLING ANALYSIS"))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Average Coupling", fmt.Sprintf("%.2f", deps.CouplingAnalysis.AverageCoupling)))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Average Instability", fmt.Sprintf("%.3f", deps.CouplingAnalysis.AverageInstability)))
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Main Sequence Deviation", fmt.Sprintf("%.3f", deps.CouplingAnalysis.MainSequenceDeviation)))

		if len(deps.CouplingAnalysis.HighlyCoupledModules) > 0 {
			builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Highly Coupled", strings.Join(deps.CouplingAnalysis.HighlyCoupledModules[:min(3, len(deps.CouplingAnalysis.HighlyCoupledModules))], ", ")))
		}
		builder.WriteString("\n")
	}

	// Longest chains
	if len(deps.LongestChains) > 0 {
		builder.WriteString(utils.FormatSectionHeader("LONGEST DEPENDENCY CHAINS"))
		for i, chain := range deps.LongestChains {
			if i >= 5 { // Limit to top 5
				break
			}
			pathStr := f.formatDependencyPath(chain.Path)
			builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, fmt.Sprintf("Chain %d (depth %d)", i+1, chain.Length), pathStr))
		}
		builder.WriteString("\n")
	}

	builder.WriteString(utils.FormatSectionSeparator())
}

// writeArchitectureSection writes the architecture analysis section
func (f *SystemAnalysisFormatterImpl) writeArchitectureSection(builder *strings.Builder, arch *domain.ArchitectureAnalysisResult, utils *FormatUtils) {
	builder.WriteString(utils.FormatSectionHeader("ARCHITECTURE ANALYSIS"))

	// Layer analysis
	if arch.LayerAnalysis != nil {
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Layers Analyzed", strconv.Itoa(arch.LayerAnalysis.LayersAnalyzed)))

		// Layer violations
		if len(arch.LayerAnalysis.LayerViolations) > 0 {
			builder.WriteString(utils.FormatSectionHeader("LAYER VIOLATIONS"))
			for i, violation := range arch.LayerAnalysis.LayerViolations {
				if i >= 10 { // Limit violations
					builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "...", fmt.Sprintf("and %d more violations", len(arch.LayerAnalysis.LayerViolations)-i)))
					break
				}
				builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Rule",
					fmt.Sprintf("%s: %s -> %s (%s)", violation.Rule, violation.FromModule, violation.ToModule, violation.Severity)))
			}
			builder.WriteString("\n")
		}
	}

	// Summary stats
	stats := map[string]interface{}{
		"Total Violations": arch.TotalViolations,
		"Total Rules":      arch.TotalRules,
		"Compliance Score": fmt.Sprintf("%.1f%%", arch.ComplianceScore*100),
	}
	builder.WriteString(utils.FormatSummaryStats(stats))

	builder.WriteString(utils.FormatSectionSeparator())
}

// writeQualitySection writes the quality metrics section
func (f *SystemAnalysisFormatterImpl) writeQualitySection(builder *strings.Builder, quality *domain.QualityMetricsResult, utils *FormatUtils) {
	builder.WriteString(utils.FormatSectionHeader("QUALITY METRICS"))

	// Summary stats
	stats := map[string]interface{}{
		"Overall Quality":       fmt.Sprintf("%.1f", quality.OverallQuality),
		"Maintainability Index": fmt.Sprintf("%.1f", quality.MaintainabilityIndex),
		"Technical Debt":        fmt.Sprintf("%.1f hours", quality.TechnicalDebtTotal),
		"System Instability":    fmt.Sprintf("%.3f", quality.SystemInstability),
		"System Complexity":     fmt.Sprintf("%.1f", quality.SystemComplexity),
	}
	builder.WriteString(utils.FormatSummaryStats(stats))

	// High quality modules
	if len(quality.HighQualityModules) > 0 {
		builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "High Quality Modules",
			strings.Join(quality.HighQualityModules[:min(5, len(quality.HighQualityModules))], ", ")))
	}

	builder.WriteString(utils.FormatSectionSeparator())
}

// formatJSON formats the response as JSON
func (f *SystemAnalysisFormatterImpl) formatJSON(response *domain.SystemAnalysisResponse) (string, error) {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatYAML formats the response as YAML
func (f *SystemAnalysisFormatterImpl) formatYAML(response *domain.SystemAnalysisResponse) (string, error) {
	data, err := yaml.Marshal(response)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// formatCSV formats the response as CSV
func (f *SystemAnalysisFormatterImpl) formatCSV(response *domain.SystemAnalysisResponse) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	// CSV headers for dependency analysis
	if response.DependencyAnalysis != nil {
		_ = writer.Write([]string{"Analysis Type", "Metric", "Value"})

		// Dependency metrics
		_ = writer.Write([]string{"Dependencies", "Total Modules", strconv.Itoa(response.DependencyAnalysis.TotalModules)})
		_ = writer.Write([]string{"Dependencies", "Total Dependencies", strconv.Itoa(response.DependencyAnalysis.TotalDependencies)})
		_ = writer.Write([]string{"Dependencies", "Root Modules", strconv.Itoa(len(response.DependencyAnalysis.RootModules))})
		_ = writer.Write([]string{"Dependencies", "Leaf Modules", strconv.Itoa(len(response.DependencyAnalysis.LeafModules))})
		_ = writer.Write([]string{"Dependencies", "Max Depth", strconv.Itoa(response.DependencyAnalysis.MaxDepth)})

		if response.DependencyAnalysis.CircularDependencies != nil {
			_ = writer.Write([]string{"Dependencies", "Circular Dependencies", strconv.FormatBool(response.DependencyAnalysis.CircularDependencies.HasCircularDependencies)})
			_ = writer.Write([]string{"Dependencies", "Total Cycles", strconv.Itoa(response.DependencyAnalysis.CircularDependencies.TotalCycles)})
		}

		if response.DependencyAnalysis.CouplingAnalysis != nil {
			_ = writer.Write([]string{"Dependencies", "Average Coupling", fmt.Sprintf("%.2f", response.DependencyAnalysis.CouplingAnalysis.AverageCoupling)})
			_ = writer.Write([]string{"Dependencies", "Average Instability", fmt.Sprintf("%.3f", response.DependencyAnalysis.CouplingAnalysis.AverageInstability)})
		}
	}

	// Architecture metrics
	if response.ArchitectureAnalysis != nil {
		_ = writer.Write([]string{"Architecture", "Layer Count", strconv.Itoa(response.ArchitectureAnalysis.LayerAnalysis.LayersAnalyzed)})
		_ = writer.Write([]string{"Architecture", "Violations", strconv.Itoa(response.ArchitectureAnalysis.TotalViolations)})
		_ = writer.Write([]string{"Architecture", "Compliance Score", fmt.Sprintf("%.3f", response.ArchitectureAnalysis.ComplianceScore)})
		_ = writer.Write([]string{"Architecture", "Detected Rules", strconv.Itoa(response.ArchitectureAnalysis.TotalRules)})
	}

	// Quality metrics
	if response.QualityMetrics != nil {
		_ = writer.Write([]string{"Quality", "Overall Quality", fmt.Sprintf("%.1f", response.QualityMetrics.OverallQuality)})
		_ = writer.Write([]string{"Quality", "Maintainability Index", fmt.Sprintf("%.1f", response.QualityMetrics.MaintainabilityIndex)})
		_ = writer.Write([]string{"Quality", "Technical Debt Hours", fmt.Sprintf("%.1f", response.QualityMetrics.TechnicalDebtTotal)})
		_ = writer.Write([]string{"Quality", "System Instability", fmt.Sprintf("%.3f", response.QualityMetrics.SystemInstability)})
		_ = writer.Write([]string{"Quality", "System Complexity", fmt.Sprintf("%.1f", response.QualityMetrics.SystemComplexity)})
	}

	writer.Flush()
	return builder.String(), nil
}

// formatHTML formats the response as HTML
func (f *SystemAnalysisFormatterImpl) formatHTML(response *domain.SystemAnalysisResponse) (string, error) {
	var builder strings.Builder

	// Create HTML template
	template := HTMLTemplate{
		Title:       "Dependency Analysis Report",
		Subtitle:    "System Dependencies and Structural Quality",
		GeneratedAt: response.GeneratedAt,
		Version:     response.Version,
		Duration:    response.Duration,
		ShowScore:   false, // System analysis doesn't have a single score
	}

	// Generate header
	builder.WriteString(template.GenerateHTMLHeader())

	// Count sections to determine if we need tabs
	sectionCount := 0
	if response.DependencyAnalysis != nil {
		sectionCount++
	}
	if response.ArchitectureAnalysis != nil {
		sectionCount++
	}
	if response.QualityMetrics != nil {
		sectionCount++
	}

	// Use tabs if multiple sections, otherwise single page
	if sectionCount > 1 {
		builder.WriteString(GenerateTabsStart())

		// Generate tab buttons
		activeTab := true
		if response.DependencyAnalysis != nil {
			builder.WriteString(GenerateTabButton("dependencies", "Dependencies", activeTab))
			activeTab = false
		}
		if response.ArchitectureAnalysis != nil {
			builder.WriteString(GenerateTabButton("architecture", "Architecture", activeTab))
			activeTab = false
		}
		if response.QualityMetrics != nil {
			builder.WriteString(GenerateTabButton("quality", "Quality", activeTab))
		}

		builder.WriteString(GenerateTabsMiddle())

		// Generate tab content
		activeTab = true
		if response.DependencyAnalysis != nil {
			var content strings.Builder
			f.writeHTMLDependenciesContent(&content, response.DependencyAnalysis)
			builder.WriteString(GenerateTabContent("dependencies", activeTab, content.String()))
			activeTab = false
		}
		if response.ArchitectureAnalysis != nil {
			var content strings.Builder
			f.writeHTMLArchitectureContent(&content, response.ArchitectureAnalysis)
			builder.WriteString(GenerateTabContent("architecture", activeTab, content.String()))
			activeTab = false
		}
		if response.QualityMetrics != nil {
			var content strings.Builder
			f.writeHTMLQualityContent(&content, response.QualityMetrics)
			builder.WriteString(GenerateTabContent("quality", activeTab, content.String()))
		}

		builder.WriteString(GenerateTabsEnd())
		builder.WriteString(GenerateTabScript())
	} else {
		// Single section - no tabs needed
		var content strings.Builder
		if response.DependencyAnalysis != nil {
			f.writeHTMLDependenciesContent(&content, response.DependencyAnalysis)
		}
		if response.ArchitectureAnalysis != nil {
			f.writeHTMLArchitectureContent(&content, response.ArchitectureAnalysis)
		}
		if response.QualityMetrics != nil {
			f.writeHTMLQualityContent(&content, response.QualityMetrics)
		}
		builder.WriteString(GenerateSinglePageContent(content.String()))
	}

	// Close HTML
	builder.WriteString(GenerateHTMLFooter())

	return builder.String(), nil
}

// formatDOT formats the response as DOT graph for visualization
func (f *SystemAnalysisFormatterImpl) formatDOT(response *domain.SystemAnalysisResponse) (string, error) {
	if response.DependencyAnalysis == nil {
		return "", fmt.Errorf("no dependency data available for DOT format")
	}

	var builder strings.Builder

	builder.WriteString("digraph SystemDependencies {\n")
	builder.WriteString("  rankdir=LR;\n")
	builder.WriteString("  node [shape=box, style=filled, fillcolor=lightblue];\n")
	builder.WriteString("  edge [color=gray];\n\n")

	// Add nodes for all modules
	modules := make(map[string]bool)

	// Collect all unique modules from dependency matrix
	if response.DependencyAnalysis.DependencyMatrix != nil {
		for modName, deps := range response.DependencyAnalysis.DependencyMatrix {
			modules[modName] = true
			for depName := range deps {
				if deps[depName] {
					modules[depName] = true
				}
			}
		}
	}

	// Add nodes
	for modName := range modules {
		// Clean module name for DOT format
		cleanName := strings.ReplaceAll(modName, ".", "_")
		cleanName = strings.ReplaceAll(cleanName, "-", "_")

		// Color nodes based on type
		color := "lightblue"
		if f.isInSlice(response.DependencyAnalysis.RootModules, modName) {
			color = "lightgreen" // Root modules
		} else if f.isInSlice(response.DependencyAnalysis.LeafModules, modName) {
			color = "lightyellow" // Leaf modules
		}

		// Check if module is in a cycle
		if response.DependencyAnalysis.CircularDependencies != nil {
			for _, cycle := range response.DependencyAnalysis.CircularDependencies.CircularDependencies {
				if f.isInSlice(cycle.Modules, modName) {
					color = "lightcoral" // Cyclic modules
					break
				}
			}
		}

		builder.WriteString(fmt.Sprintf("  %s [label=\"%s\", fillcolor=%s];\n", cleanName, modName, color))
	}

	builder.WriteString("\n")

	// Add edges
	if response.DependencyAnalysis.DependencyMatrix != nil {
		for modName, deps := range response.DependencyAnalysis.DependencyMatrix {
			cleanModule := strings.ReplaceAll(modName, ".", "_")
			cleanModule = strings.ReplaceAll(cleanModule, "-", "_")

			for depName := range deps {
				if deps[depName] {
					cleanDep := strings.ReplaceAll(depName, ".", "_")
					cleanDep = strings.ReplaceAll(cleanDep, "-", "_")

					// Color edges in cycles differently
					edgeColor := "gray"
					if response.DependencyAnalysis.CircularDependencies != nil {
						for _, cycle := range response.DependencyAnalysis.CircularDependencies.CircularDependencies {
							if f.isInSlice(cycle.Modules, modName) && f.isInSlice(cycle.Modules, depName) {
								edgeColor = "red"
								break
							}
						}
					}

					builder.WriteString(fmt.Sprintf("  %s -> %s [color=%s];\n", cleanModule, cleanDep, edgeColor))
				}
			}
		}
	}

	// Add legend
	builder.WriteString("\n  // Legend\n")
	builder.WriteString("  subgraph cluster_legend {\n")
	builder.WriteString("    label=\"Legend\";\n")
	builder.WriteString("    style=filled;\n")
	builder.WriteString("    fillcolor=white;\n")
	builder.WriteString("    legend_root [label=\"Root Module\", fillcolor=lightgreen, shape=box];\n")
	builder.WriteString("    legend_leaf [label=\"Leaf Module\", fillcolor=lightyellow, shape=box];\n")
	builder.WriteString("    legend_cycle [label=\"In Cycle\", fillcolor=lightcoral, shape=box];\n")
	builder.WriteString("    legend_normal [label=\"Normal Module\", fillcolor=lightblue, shape=box];\n")
	builder.WriteString("  }\n")

	builder.WriteString("}\n")

	return builder.String(), nil
}

// HTML section writers

func (f *SystemAnalysisFormatterImpl) writeHTMLDependenciesContent(builder *strings.Builder, deps *domain.DependencyAnalysisResult) {
	builder.WriteString(GenerateSectionHeader("📊 Dependency Analysis"))
	builder.WriteString(`<div class="metric-grid">`)
	builder.WriteString(GenerateMetricCard(strconv.Itoa(deps.TotalModules), "Total Modules"))
	builder.WriteString(GenerateMetricCard(strconv.Itoa(deps.TotalDependencies), "Total Dependencies"))
	builder.WriteString(GenerateMetricCard(strconv.Itoa(deps.MaxDepth), "Max Dependency Depth"))

	if deps.CircularDependencies != nil {
		statusText := "✅ None"
		severity := "success"
		if deps.CircularDependencies.HasCircularDependencies {
			statusText = fmt.Sprintf("❌ %d Cycles", deps.CircularDependencies.TotalCycles)
			severity = "danger"
		}
		builder.WriteString(GenerateMetricCard(GenerateStatusBadge(statusText, severity), "Circular Dependencies"))
	}

	builder.WriteString(`
            </div>`)

	// Coupling analysis
	if deps.CouplingAnalysis != nil {
		builder.WriteString(GenerateSectionHeader("Coupling Metrics"))
		builder.WriteString(`<div class="metric-grid">`)
		builder.WriteString(GenerateMetricCard(fmt.Sprintf("%.2f", deps.CouplingAnalysis.AverageCoupling), "Average Coupling"))
		builder.WriteString(GenerateMetricCard(fmt.Sprintf("%.3f", deps.CouplingAnalysis.AverageInstability), "Average Instability"))
		builder.WriteString(GenerateMetricCard(fmt.Sprintf("%.3f", deps.CouplingAnalysis.MainSequenceDeviation), "Main Sequence Deviation"))
		builder.WriteString(`</div>`)
	}

	// Add detailed dependency list if available
	if len(deps.DependencyMatrix) > 0 {
		builder.WriteString(GenerateSectionHeader("Module Dependencies"))
		builder.WriteString(`
            <table class="table">
                <thead>
                    <tr>
                        <th>Module</th>
                        <th>Depends On</th>
                    </tr>
                </thead>
                <tbody>`)

		// Sort modules for consistent display
		var modules []string
		for module := range deps.DependencyMatrix {
			modules = append(modules, module)
		}
		sort.Strings(modules)

		for _, module := range modules {
			depMap := deps.DependencyMatrix[module]
			var dependencies []string
			for dep, hasDep := range depMap {
				if hasDep {
					dependencies = append(dependencies, dep)
				}
			}

			if len(dependencies) > 0 {
				sort.Strings(dependencies)
				builder.WriteString(`
                        <tr>
                            <td><strong>` + module + `</strong></td>
                            <td>`)
				for i, dep := range dependencies {
					if i > 0 {
						builder.WriteString(`<br>`)
					}
					builder.WriteString(dep)
				}
				builder.WriteString(`</td>
                        </tr>`)
			}
		}

		builder.WriteString(`
                    </tbody>
                </table>`)
	}

	// Add longest dependency chains
	if len(deps.LongestChains) > 0 {
		builder.WriteString(GenerateSectionHeader("Longest Dependency Chains"))
		builder.WriteString(`
            <table class="table">
                <thead>
                    <tr>
                        <th>Chain #</th>
                        <th>Depth</th>
                        <th>Path</th>
                    </tr>
                </thead>
                <tbody>`)
		for i, chain := range deps.LongestChains {
			if i >= 10 { // Limit to 10 chains
				break
			}
			builder.WriteString(`
                    <tr>
                        <td><strong>` + strconv.Itoa(i+1) + `</strong></td>
                        <td>` + strconv.Itoa(chain.Length) + `</td>
                        <td>`)
			for j, module := range chain.Path {
				if j > 0 {
					builder.WriteString(` → `)
				}
				// Show only first 3 and last module for long chains
				if len(chain.Path) > 5 && j == 3 {
					builder.WriteString(`... → ` + chain.Path[len(chain.Path)-1])
					break
				}
				builder.WriteString(module)
			}
			builder.WriteString(`</td>
                    </tr>`)
		}
		builder.WriteString(`
                </tbody>
            </table>`)
	}
}

func (f *SystemAnalysisFormatterImpl) writeHTMLArchitectureContent(builder *strings.Builder, arch *domain.ArchitectureAnalysisResult) {
	layersAnalyzed := 0
	if arch.LayerAnalysis != nil {
		layersAnalyzed = arch.LayerAnalysis.LayersAnalyzed
	}

	// Use the same section structure as Dependencies
	builder.WriteString(GenerateSectionHeader("🏛️ Architecture Analysis"))
	builder.WriteString(`<div class="metric-grid">`)
	builder.WriteString(GenerateMetricCard(strconv.Itoa(layersAnalyzed), "Layers Analyzed"))
	builder.WriteString(GenerateMetricCard(strconv.Itoa(arch.TotalRules), "Total Rules"))

    // Violations: show as large metric number (not inside small badge)
    violationValue := strconv.Itoa(arch.TotalViolations)
    if arch.TotalViolations > 0 {
        violationValue = "❌ " + violationValue
    } else {
        violationValue = "✅ " + violationValue
    }
    builder.WriteString(GenerateMetricCard(violationValue, "Violations"))

	// Compliance Score with color coding
	complianceScore := arch.ComplianceScore * 100
	scoreColor := "success"
	if complianceScore < 80 {
		scoreColor = "warning"
	}
	if complianceScore < 60 {
		scoreColor = "danger"
	}
	builder.WriteString(GenerateMetricCard(
		`<span class="badge bg-` + scoreColor + `">` + fmt.Sprintf("%.1f%%", complianceScore) + `</span>`,
		"Compliance Score"))
	builder.WriteString(`</div>`)

	// Layer Analysis Details
	if arch.LayerAnalysis != nil {
		// Layer Coupling if available
		if len(arch.LayerAnalysis.LayerCoupling) > 0 {
			builder.WriteString(GenerateSectionHeader("Layer Dependencies"))
			builder.WriteString(`
            <table class="table">
                <thead>
                    <tr>
                        <th>From Layer</th>
                        <th>To Layer</th>
                        <th>Dependencies</th>
                    </tr>
                </thead>
                <tbody>`)

			for fromLayer, toMap := range arch.LayerAnalysis.LayerCoupling {
				for toLayer, count := range toMap {
					builder.WriteString(`
                    <tr>
                        <td><strong>` + fromLayer + `</strong></td>
                        <td>` + toLayer + `</td>
                        <td>` + strconv.Itoa(count) + `</td>
                    </tr>`)
				}
			}
			builder.WriteString(`
                </tbody>
            </table>`)
		}

		// Layer Violations
		if len(arch.LayerAnalysis.LayerViolations) > 0 {
			builder.WriteString(GenerateSectionHeader("Architecture Violations"))
			builder.WriteString(`
            <table class="table">
                <thead>
                    <tr>
                        <th>Severity</th>
                        <th>Rule</th>
                        <th>From Module</th>
                        <th>To Module</th>
                    </tr>
                </thead>
                <tbody>`)

			for i, violation := range arch.LayerAnalysis.LayerViolations {
				if i >= 20 { // Limit to 20 violations
					builder.WriteString(`
                    <tr>
                        <td colspan="4"><em>... and ` + strconv.Itoa(len(arch.LayerAnalysis.LayerViolations)-20) + ` more violations</em></td>
                    </tr>`)
					break
				}

				severityClass := "warning"
				severityIcon := "⚠️"
				if string(violation.Severity) == "error" {
					severityClass = "danger"
					severityIcon = "❌"
				} else if string(violation.Severity) == "info" {
					severityClass = "info"
					severityIcon = "ℹ️"
				}

				builder.WriteString(`
                    <tr>
                        <td><span class="badge bg-` + severityClass + `">` + severityIcon + ` ` + string(violation.Severity) + `</span></td>
                        <td>` + violation.Rule + `</td>
                        <td>` + violation.FromModule + `</td>
                        <td>` + violation.ToModule + `</td>
                    </tr>`)
			}
			builder.WriteString(`
                </tbody>
            </table>`)
		}

		// Layer Cohesion if available
		if len(arch.LayerAnalysis.LayerCohesion) > 0 {
			builder.WriteString(GenerateSectionHeader("Layer Cohesion"))
			builder.WriteString(`<div class="metric-grid">`)
			for layer, cohesion := range arch.LayerAnalysis.LayerCohesion {
				cohesionText := fmt.Sprintf("%.2f", cohesion)
				builder.WriteString(GenerateMetricCard(cohesionText, layer))
			}
			builder.WriteString(`</div>`)
		}
	}

	// Recommendations if available
	if len(arch.Recommendations) > 0 {
		builder.WriteString(GenerateSectionHeader("Recommendations"))
		builder.WriteString(`<ul class="list-group">`)
		for _, rec := range arch.Recommendations {
			builder.WriteString(`<li class="list-group-item">` + rec.Description + `</li>`)
		}
		builder.WriteString(`</ul>`)
	}
}

func (f *SystemAnalysisFormatterImpl) writeHTMLQualityContent(builder *strings.Builder, quality *domain.QualityMetricsResult) {
    builder.WriteString(`
        <div class="section">
            <h2>⚡ Quality Metrics</h2>
            <div class="metric-grid">
                <div class="metric-card">
                    <div class="metric-value">` + fmt.Sprintf("%.1f", quality.OverallQuality) + `</div>
                    <div class="metric-label">Overall Quality</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">` + fmt.Sprintf("%.1f", quality.MaintainabilityIndex) + `</div>
                    <div class="metric-label">Maintainability Index</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">` + fmt.Sprintf("%.1f", quality.TechnicalDebtTotal) + `h</div>
                    <div class="metric-label">Technical Debt</div>
                </div>
                <div class="metric-card">
                    <div class="metric-value">` + fmt.Sprintf("%.3f", quality.SystemInstability) + `</div>
                    <div class="metric-label">System Instability</div>
                </div>
            </div>`)

	if len(quality.HighQualityModules) > 0 {
		builder.WriteString(`
            <h3>High Quality Modules</h3>
            <div class="chart-container">`)
		for i, module := range quality.HighQualityModules {
			if i >= 10 { // Limit to 10 modules
				break
			}
			builder.WriteString(`<div style="padding: 8px 0; border-bottom: 1px solid #e9ecef;">` + module + `</div>`)
		}
		builder.WriteString(`</div>`)
	}
}

// Helper methods

func (f *SystemAnalysisFormatterImpl) formatDependencyPath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	if len(path) <= 4 {
		return strings.Join(path, " → ")
	}
	return fmt.Sprintf("%s → ... → %s", path[0], path[len(path)-1])
}

func (f *SystemAnalysisFormatterImpl) isInSlice(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
