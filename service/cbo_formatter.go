package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"gopkg.in/yaml.v3"
)

// CBOFormatterImpl implements CBOOutputFormatter interface
type CBOFormatterImpl struct{}

// NewCBOFormatter creates a new CBO output formatter
func NewCBOFormatter() *CBOFormatterImpl {
	return &CBOFormatterImpl{}
}

// Format formats the CBO analysis response according to the specified format
func (f *CBOFormatterImpl) Format(response *domain.CBOResponse, format domain.OutputFormat) (string, error) {
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
func (f *CBOFormatterImpl) Write(response *domain.CBOResponse, format domain.OutputFormat, writer io.Writer) error {
	formatted, err := f.Format(response, format)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(formatted))
	return err
}

// formatText formats the response as human-readable text
func (f *CBOFormatterImpl) formatText(response *domain.CBOResponse) (string, error) {
	var builder strings.Builder
	utils := NewFormatUtils()

	// Header
	builder.WriteString(utils.FormatMainHeader("CBO (Coupling Between Objects) Analysis Report"))

	// Summary
	stats := map[string]interface{}{
		"Total Classes":  response.Summary.TotalClasses,
		"Files Analyzed": response.Summary.FilesAnalyzed,
		"Average CBO":    fmt.Sprintf("%.1f", response.Summary.AverageCBO),
		"Max CBO":        response.Summary.MaxCBO,
		"Min CBO":        response.Summary.MinCBO,
	}
	builder.WriteString(utils.FormatSummaryStats(stats))

	// Risk distribution
	builder.WriteString(utils.FormatRiskDistribution(
		response.Summary.HighRiskClasses,
		response.Summary.MediumRiskClasses,
		response.Summary.LowRiskClasses))

	// CBO distribution
	if len(response.Summary.CBODistribution) > 0 {
		builder.WriteString(utils.FormatSectionHeader("CBO DISTRIBUTION"))

		// Sort ranges for consistent output
		ranges := make([]string, 0, len(response.Summary.CBODistribution))
		for rang := range response.Summary.CBODistribution {
			ranges = append(ranges, rang)
		}
		sort.Strings(ranges)

		for _, rang := range ranges {
			count := response.Summary.CBODistribution[rang]
			builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, fmt.Sprintf("CBO %s", rang), fmt.Sprintf("%d classes", count)))
		}
		builder.WriteString(utils.FormatSectionSeparator())
	}

	// Most coupled classes
	if len(response.Summary.MostCoupledClasses) > 0 {
		builder.WriteString(utils.FormatSectionHeader("MOST COUPLED CLASSES"))
		for i, class := range response.Summary.MostCoupledClasses {
			if i >= 10 { // Limit to top 10
				break
			}
			// Convert domain risk level to standard risk level
			var standardRisk RiskLevel
			switch class.RiskLevel {
			case "High":
				standardRisk = RiskHigh
			case "Medium":
				standardRisk = RiskMedium
			case "Low":
				standardRisk = RiskLow
			default:
				standardRisk = RiskLow
			}
			coloredRisk := utils.FormatRiskWithColor(standardRisk)

			builder.WriteString(fmt.Sprintf("%s%d. %s %s (CBO: %d) - %s:%d\n",
				strings.Repeat(" ", SectionPadding), i+1, coloredRisk, class.Name, class.Metrics.CouplingCount, class.FilePath, class.StartLine))
		}
		builder.WriteString(utils.FormatSectionSeparator())
	}

	// Detailed class information
	if len(response.Classes) > 0 {
		builder.WriteString(utils.FormatSectionHeader("CLASS DETAILS"))
		for _, class := range response.Classes {
			f.writeClassDetails(&builder, class, utils)
			builder.WriteString("\n")
		}
		builder.WriteString(utils.FormatSectionSeparator())
	}

	// Warnings
	if len(response.Warnings) > 0 {
		builder.WriteString(utils.FormatWarningsSection(response.Warnings))
	}

	// Errors
	if len(response.Errors) > 0 {
		builder.WriteString(utils.FormatSectionHeader("ERRORS"))
		for _, err := range response.Errors {
			builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "❌", err))
		}
		builder.WriteString(utils.FormatSectionSeparator())
	}

	// Footer
	builder.WriteString(utils.FormatSectionHeader("METADATA"))
	builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Generated at", response.GeneratedAt))
	builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Version", response.Version))

	return builder.String(), nil
}

// writeClassDetails writes detailed information about a class
func (f *CBOFormatterImpl) writeClassDetails(builder *strings.Builder, class domain.ClassCoupling, utils *FormatUtils) {
	// Convert domain risk level to standard risk level
	var standardRisk RiskLevel
	switch class.RiskLevel {
	case "High":
		standardRisk = RiskHigh
	case "Medium":
		standardRisk = RiskMedium
	case "Low":
		standardRisk = RiskLow
	default:
		standardRisk = RiskLow
	}
	coloredRisk := utils.FormatRiskWithColor(standardRisk)

	builder.WriteString(utils.FormatLabelWithIndent(SectionPadding, "Class", fmt.Sprintf("%s %s (CBO: %d)",
		coloredRisk, class.Name, class.Metrics.CouplingCount)))
	builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Location", fmt.Sprintf("%s:%d-%d", class.FilePath, class.StartLine, class.EndLine)))

	if class.IsAbstract {
		builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Type", "Abstract Class"))
	}

	// Base classes
	if len(class.BaseClasses) > 0 {
		builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Inherits from", strings.Join(class.BaseClasses, ", ")))
	}

	// Dependency breakdown
	if class.Metrics.CouplingCount > 0 {
		builder.WriteString(utils.FormatLabelWithIndent(ItemPadding, "Dependencies", ""))
		if class.Metrics.InheritanceDependencies > 0 {
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding+2, "Inheritance", class.Metrics.InheritanceDependencies))
		}
		if class.Metrics.TypeHintDependencies > 0 {
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding+2, "Type Hints", class.Metrics.TypeHintDependencies))
		}
		if class.Metrics.InstantiationDependencies > 0 {
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding+2, "Instantiation", class.Metrics.InstantiationDependencies))
		}
		if class.Metrics.AttributeAccessDependencies > 0 {
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding+2, "Attribute Access", class.Metrics.AttributeAccessDependencies))
		}
		if class.Metrics.ImportDependencies > 0 {
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding+2, "Imports", class.Metrics.ImportDependencies))
		}

		// List dependent classes
		if len(class.Metrics.DependentClasses) > 0 {
			builder.WriteString(utils.FormatLabelWithIndent(ItemPadding+2, "Coupled to", strings.Join(class.Metrics.DependentClasses, ", ")))
		}
	}
}

// formatJSON formats the response as JSON
func (f *CBOFormatterImpl) formatJSON(response *domain.CBOResponse) (string, error) {
	jsonBytes, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(jsonBytes), nil
}

// formatYAML formats the response as YAML
func (f *CBOFormatterImpl) formatYAML(response *domain.CBOResponse) (string, error) {
	yamlBytes, err := yaml.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}
	return string(yamlBytes), nil
}

// formatCSV formats the response as CSV
func (f *CBOFormatterImpl) formatCSV(response *domain.CBOResponse) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)

	// Write header
	header := []string{
		"ClassName", "FilePath", "StartLine", "EndLine", "CBO", "RiskLevel", "IsAbstract",
		"InheritanceDeps", "TypeHintDeps", "InstantiationDeps", "AttributeAccessDeps", "ImportDeps",
		"BaseClasses", "DependentClasses",
	}
	if err := writer.Write(header); err != nil {
		return "", fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, class := range response.Classes {
		row := []string{
			class.Name,
			class.FilePath,
			strconv.Itoa(class.StartLine),
			strconv.Itoa(class.EndLine),
			strconv.Itoa(class.Metrics.CouplingCount),
			string(class.RiskLevel),
			strconv.FormatBool(class.IsAbstract),
			strconv.Itoa(class.Metrics.InheritanceDependencies),
			strconv.Itoa(class.Metrics.TypeHintDependencies),
			strconv.Itoa(class.Metrics.InstantiationDependencies),
			strconv.Itoa(class.Metrics.AttributeAccessDependencies),
			strconv.Itoa(class.Metrics.ImportDependencies),
			strings.Join(class.BaseClasses, ";"),
			strings.Join(class.Metrics.DependentClasses, ";"),
		}
		if err := writer.Write(row); err != nil {
			return "", fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("CSV writer error: %w", err)
	}

	return builder.String(), nil
}

// formatHTML formats the response as HTML
func (f *CBOFormatterImpl) formatHTML(response *domain.CBOResponse) (string, error) {
	var builder strings.Builder

	// Create HTML template
	generatedAt, _ := time.Parse("2006-01-02 15:04:05", response.GeneratedAt)
	if generatedAt.IsZero() {
		generatedAt = time.Now()
	}

	template := HTMLTemplate{
		Title:       "CBO Analysis Report",
		Subtitle:    "Coupling Between Objects Analysis",
		GeneratedAt: generatedAt,
		Version:     response.Version,
		Duration:    0,     // CBO doesn't track duration
		ShowScore:   false, // CBO doesn't have a single score
	}

	// Generate header
	builder.WriteString(template.GenerateHTMLHeader())

	// Generate content
	var content strings.Builder

	// Summary section with metrics
	content.WriteString(GenerateSectionHeader("📊 Overview"))
	content.WriteString(`<div class="metric-grid">`)
	content.WriteString(GenerateMetricCard(strconv.Itoa(response.Summary.TotalClasses), "Total Classes"))
	content.WriteString(GenerateMetricCard(strconv.Itoa(response.Summary.FilesAnalyzed), "Files Analyzed"))
	content.WriteString(GenerateMetricCard(fmt.Sprintf("%.2f", response.Summary.AverageCBO), "Average CBO"))
	content.WriteString(GenerateMetricCard(strconv.Itoa(response.Summary.MaxCBO), "Max CBO"))
	content.WriteString(`</div>`)

	// Risk distribution
	content.WriteString(GenerateSectionHeader("🚦 Risk Distribution"))
	content.WriteString(`<div class="metric-grid">`)
	content.WriteString(GenerateMetricCard(
		GenerateStatusBadge(strconv.Itoa(response.Summary.LowRiskClasses), "success"),
		"Low Risk Classes"))
	content.WriteString(GenerateMetricCard(
		GenerateStatusBadge(strconv.Itoa(response.Summary.MediumRiskClasses), "warning"),
		"Medium Risk Classes"))
	content.WriteString(GenerateMetricCard(
		GenerateStatusBadge(strconv.Itoa(response.Summary.HighRiskClasses), "danger"),
		"High Risk Classes"))
	content.WriteString(`</div>`)

	// Classes table
	if len(response.Classes) > 0 {
		content.WriteString(GenerateSectionHeader("📋 Class Details"))
		content.WriteString(`
        <table class="table">
            <thead>
                <tr>
                    <th>Class Name</th>
                    <th>CBO</th>
                    <th>Risk</th>
                    <th>Location</th>
                    <th>Dependencies</th>
                </tr>
            </thead>
            <tbody>
`)

		for _, class := range response.Classes {
			riskSeverity := "info"
			switch class.RiskLevel {
			case domain.RiskLevelLow:
				riskSeverity = "success"
			case domain.RiskLevelMedium:
				riskSeverity = "warning"
			case domain.RiskLevelHigh:
				riskSeverity = "danger"
			}

			content.WriteString(fmt.Sprintf(`                <tr>
                    <td><strong>%s</strong></td>
                    <td>%s</td>
                    <td>%s</td>
                    <td>%s:%d</td>
                    <td>
`, class.Name,
				GenerateStatusBadge(strconv.Itoa(class.Metrics.CouplingCount), riskSeverity),
				GenerateStatusBadge(string(class.RiskLevel), riskSeverity),
				class.FilePath, class.StartLine))

			if len(class.Metrics.DependentClasses) > 0 {
				content.WriteString(strings.Join(class.Metrics.DependentClasses, ", "))
			} else {
				content.WriteString("No dependencies")
			}

			// Add dependency breakdown
			if class.Metrics.CouplingCount > 0 {
				content.WriteString(`<br><small style="color: #666;">`)
				deps := []string{}
				if class.Metrics.InheritanceDependencies > 0 {
					deps = append(deps, fmt.Sprintf("Inheritance: %d", class.Metrics.InheritanceDependencies))
				}
				if class.Metrics.TypeHintDependencies > 0 {
					deps = append(deps, fmt.Sprintf("Type Hints: %d", class.Metrics.TypeHintDependencies))
				}
				if class.Metrics.InstantiationDependencies > 0 {
					deps = append(deps, fmt.Sprintf("Instantiation: %d", class.Metrics.InstantiationDependencies))
				}
				if class.Metrics.AttributeAccessDependencies > 0 {
					deps = append(deps, fmt.Sprintf("Attribute Access: %d", class.Metrics.AttributeAccessDependencies))
				}
				content.WriteString(strings.Join(deps, ", "))
				content.WriteString(`</small>`)
			}

			content.WriteString(`                    </td>
                </tr>
`)
		}

		content.WriteString(`            </tbody>
        </table>`)
	}

	// Wrap content in single page container
	builder.WriteString(GenerateSinglePageContent(content.String()))

	// Close HTML
	builder.WriteString(GenerateHTMLFooter())

	return builder.String(), nil
}
