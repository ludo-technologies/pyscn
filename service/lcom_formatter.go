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

// LCOMFormatterImpl implements LCOMOutputFormatter interface
type LCOMFormatterImpl struct{}

// NewLCOMFormatter creates a new LCOM output formatter
func NewLCOMFormatter() *LCOMFormatterImpl {
	return &LCOMFormatterImpl{}
}

// Format formats the LCOM analysis response according to the specified format
func (f *LCOMFormatterImpl) Format(response *domain.LCOMResponse, format domain.OutputFormat) (string, error) {
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
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

// Write writes the formatted output to the writer
func (f *LCOMFormatterImpl) Write(response *domain.LCOMResponse, format domain.OutputFormat, writer io.Writer) error {
	formatted, err := f.Format(response, format)
	if err != nil {
		return err
	}
	_, err = writer.Write([]byte(formatted))
	return err
}

// formatText formats the response as human-readable text
func (f *LCOMFormatterImpl) formatText(response *domain.LCOMResponse) (string, error) {
	var builder strings.Builder
	utils := NewFormatUtils()

	builder.WriteString(utils.FormatMainHeader("LCOM4 (Lack of Cohesion of Methods) Analysis Report"))

	stats := map[string]any{
		"Total Classes":  response.Summary.TotalClasses,
		"Files Analyzed": response.Summary.FilesAnalyzed,
		"Average LCOM4":  fmt.Sprintf("%.1f", response.Summary.AverageLCOM),
		"Max LCOM4":      response.Summary.MaxLCOM,
		"Min LCOM4":      response.Summary.MinLCOM,
	}
	builder.WriteString(utils.FormatSummaryStats(stats))

	builder.WriteString(utils.FormatRiskDistribution(
		response.Summary.HighRiskClasses,
		response.Summary.MediumRiskClasses,
		response.Summary.LowRiskClasses))

	// LCOM distribution
	if len(response.Summary.LCOMDistribution) > 0 {
		builder.WriteString(utils.FormatSectionHeader("LCOM4 DISTRIBUTION"))
		ranges := make([]string, 0, len(response.Summary.LCOMDistribution))
		for rang := range response.Summary.LCOMDistribution {
			ranges = append(ranges, rang)
		}
		sort.Strings(ranges)

		for _, rang := range ranges {
			count := response.Summary.LCOMDistribution[rang]
			builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, fmt.Sprintf("LCOM4 %s", rang), fmt.Sprintf("%d classes", count)))
		}
		builder.WriteString(utils.FormatSectionSeparator())
	}

	// Least cohesive classes
	if len(response.Summary.LeastCohesiveClasses) > 0 {
		builder.WriteString(utils.FormatSectionHeader("LEAST COHESIVE CLASSES"))
		for i, class := range response.Summary.LeastCohesiveClasses {
			if i >= 10 {
				break
			}
			standardRisk := domainRiskToStandard(class.RiskLevel)
			coloredRisk := utils.FormatRiskWithColor(standardRisk)
			builder.WriteString(fmt.Sprintf("%s%d. %s %s (LCOM4: %d) - %s:%d\n",
				strings.Repeat(" ", SectionPadding), i+1, coloredRisk, class.Name, class.Metrics.LCOM4, class.FilePath, class.StartLine))
		}
		builder.WriteString(utils.FormatSectionSeparator())
	}

	// Class details
	if len(response.Classes) > 0 {
		builder.WriteString(utils.FormatSectionHeader("CLASS DETAILS"))
		for _, class := range response.Classes {
			f.writeClassDetails(&builder, class, utils)
			builder.WriteString("\n")
		}
		builder.WriteString(utils.FormatSectionSeparator())
	}

	if len(response.Warnings) > 0 {
		builder.WriteString(utils.FormatWarningsSection(response.Warnings))
	}

	if len(response.Errors) > 0 {
		builder.WriteString(utils.FormatSectionHeader("ERRORS"))
		for _, err := range response.Errors {
			builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Error", err))
		}
		builder.WriteString(utils.FormatSectionSeparator())
	}

	builder.WriteString(utils.FormatSectionHeader("METADATA"))
	builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Generated at", response.GeneratedAt))
	builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Version", response.Version))

	return builder.String(), nil
}

// writeClassDetails writes detailed information about a class
func (f *LCOMFormatterImpl) writeClassDetails(builder *strings.Builder, class domain.ClassCohesion, utils *FormatUtils) {
	standardRisk := domainRiskToStandard(class.RiskLevel)
	coloredRisk := utils.FormatRiskWithColor(standardRisk)

	builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Class", fmt.Sprintf("%s %s (LCOM4: %d)",
		coloredRisk, class.Name, class.Metrics.LCOM4)))
	builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Location", fmt.Sprintf("%s:%d-%d", class.FilePath, class.StartLine, class.EndLine)))
	builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Methods", fmt.Sprintf("%d total, %d excluded", class.Metrics.TotalMethods, class.Metrics.ExcludedMethods)))
	builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Instance Variables", strconv.Itoa(class.Metrics.InstanceVariables)))

	if len(class.Metrics.MethodGroups) > 0 {
		builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Method Groups", fmt.Sprintf("%d connected components", len(class.Metrics.MethodGroups))))
		for i, group := range class.Metrics.MethodGroups {
			builder.WriteString(fmt.Sprintf("%sGroup %d: %s\n",
				strings.Repeat(" ", ItemPadding+2), i+1, strings.Join(group, ", ")))
		}
	}
}

// domainRiskToStandard converts domain.RiskLevel to formatter's RiskLevel
func domainRiskToStandard(risk domain.RiskLevel) RiskLevel {
	switch risk {
	case domain.RiskLevelHigh:
		return RiskHigh
	case domain.RiskLevelMedium:
		return RiskMedium
	case domain.RiskLevelLow:
		return RiskLow
	default:
		return RiskLow
	}
}

// formatJSON formats the response as JSON
func (f *LCOMFormatterImpl) formatJSON(response *domain.LCOMResponse) (string, error) {
	data, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal LCOM response to JSON: %w", err)
	}
	return string(data), nil
}

// formatYAML formats the response as YAML
func (f *LCOMFormatterImpl) formatYAML(response *domain.LCOMResponse) (string, error) {
	data, err := yaml.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal LCOM response to YAML: %w", err)
	}
	return string(data), nil
}

// formatCSV formats the response as CSV
func (f *LCOMFormatterImpl) formatCSV(response *domain.LCOMResponse) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	// Header
	header := []string{"ClassName", "FilePath", "StartLine", "EndLine", "LCOM4", "RiskLevel",
		"TotalMethods", "ExcludedMethods", "InstanceVariables"}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	for _, class := range response.Classes {
		record := []string{
			class.Name,
			class.FilePath,
			strconv.Itoa(class.StartLine),
			strconv.Itoa(class.EndLine),
			strconv.Itoa(class.Metrics.LCOM4),
			string(class.RiskLevel),
			strconv.Itoa(class.Metrics.TotalMethods),
			strconv.Itoa(class.Metrics.ExcludedMethods),
			strconv.Itoa(class.Metrics.InstanceVariables),
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	return builder.String(), nil
}

// formatHTML formats the response as an HTML report
func (f *LCOMFormatterImpl) formatHTML(response *domain.LCOMResponse) (string, error) {
	var builder strings.Builder

	builder.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>LCOM4 Analysis Report</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 20px; background: #f5f5f5; }
.container { max-width: 1200px; margin: 0 auto; }
h1 { color: #333; }
.card { background: white; border-radius: 8px; padding: 20px; margin: 10px 0; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.metrics { display: flex; gap: 20px; flex-wrap: wrap; }
.metric { text-align: center; padding: 15px; min-width: 120px; }
.metric .value { font-size: 2em; font-weight: bold; color: #333; }
.metric .label { color: #666; font-size: 0.9em; }
table { width: 100%; border-collapse: collapse; }
th, td { padding: 10px; text-align: left; border-bottom: 1px solid #eee; }
th { background: #f8f9fa; font-weight: 600; }
.risk-high { color: #dc3545; font-weight: bold; }
.risk-medium { color: #ffc107; font-weight: bold; }
.risk-low { color: #28a745; font-weight: bold; }
</style>
</head>
<body>
<div class="container">
<h1>LCOM4 Analysis Report</h1>
`)

	// Summary metrics
	builder.WriteString(`<div class="card"><div class="metrics">`)
	builder.WriteString(fmt.Sprintf(`<div class="metric"><div class="value">%d</div><div class="label">Total Classes</div></div>`, response.Summary.TotalClasses))
	builder.WriteString(fmt.Sprintf(`<div class="metric"><div class="value">%.1f</div><div class="label">Average LCOM4</div></div>`, response.Summary.AverageLCOM))
	builder.WriteString(fmt.Sprintf(`<div class="metric"><div class="value">%d</div><div class="label">Max LCOM4</div></div>`, response.Summary.MaxLCOM))
	builder.WriteString(fmt.Sprintf(`<div class="metric"><div class="value risk-high">%d</div><div class="label">High Risk</div></div>`, response.Summary.HighRiskClasses))
	builder.WriteString(fmt.Sprintf(`<div class="metric"><div class="value risk-medium">%d</div><div class="label">Medium Risk</div></div>`, response.Summary.MediumRiskClasses))
	builder.WriteString(fmt.Sprintf(`<div class="metric"><div class="value risk-low">%d</div><div class="label">Low Risk</div></div>`, response.Summary.LowRiskClasses))
	builder.WriteString(`</div></div>`)

	// Class table
	if len(response.Classes) > 0 {
		builder.WriteString(`<div class="card"><h2>Classes</h2><table>`)
		builder.WriteString(`<tr><th>Class</th><th>File</th><th>LCOM4</th><th>Risk</th><th>Methods</th><th>Instance Vars</th></tr>`)
		for _, class := range response.Classes {
			riskClass := "risk-low"
			switch class.RiskLevel {
			case domain.RiskLevelHigh:
				riskClass = "risk-high"
			case domain.RiskLevelMedium:
				riskClass = "risk-medium"
			}
			builder.WriteString(fmt.Sprintf(`<tr><td>%s</td><td>%s:%d</td><td>%d</td><td class="%s">%s</td><td>%d</td><td>%d</td></tr>`,
				class.Name, class.FilePath, class.StartLine,
				class.Metrics.LCOM4, riskClass, class.RiskLevel,
				class.Metrics.TotalMethods-class.Metrics.ExcludedMethods,
				class.Metrics.InstanceVariables))
		}
		builder.WriteString(`</table></div>`)
	}

	builder.WriteString(`</div></body></html>`)
	return builder.String(), nil
}
