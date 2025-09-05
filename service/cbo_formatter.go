package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"github.com/pyqol/pyqol/domain"
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

	// Header
	builder.WriteString("=== CBO (Coupling Between Objects) Analysis Results ===\n\n")

	// Summary
	builder.WriteString("📊 SUMMARY\n")
	builder.WriteString(fmt.Sprintf("Total Classes:     %d\n", response.Summary.TotalClasses))
	builder.WriteString(fmt.Sprintf("Files Analyzed:    %d\n", response.Summary.FilesAnalyzed))
	builder.WriteString(fmt.Sprintf("Average CBO:       %.2f\n", response.Summary.AverageCBO))
	builder.WriteString(fmt.Sprintf("Max CBO:           %d\n", response.Summary.MaxCBO))
	builder.WriteString(fmt.Sprintf("Min CBO:           %d\n", response.Summary.MinCBO))
	builder.WriteString("\n")

	// Risk distribution
	builder.WriteString("🚦 RISK DISTRIBUTION\n")
	builder.WriteString(fmt.Sprintf("Low Risk:          %d classes\n", response.Summary.LowRiskClasses))
	builder.WriteString(fmt.Sprintf("Medium Risk:       %d classes\n", response.Summary.MediumRiskClasses))
	builder.WriteString(fmt.Sprintf("High Risk:         %d classes\n", response.Summary.HighRiskClasses))
	builder.WriteString("\n")

	// CBO distribution
	if len(response.Summary.CBODistribution) > 0 {
		builder.WriteString("📈 CBO DISTRIBUTION\n")
		
		// Sort ranges for consistent output
		ranges := make([]string, 0, len(response.Summary.CBODistribution))
		for rang := range response.Summary.CBODistribution {
			ranges = append(ranges, rang)
		}
		sort.Strings(ranges)

		for _, rang := range ranges {
			count := response.Summary.CBODistribution[rang]
			builder.WriteString(fmt.Sprintf("CBO %s:           %d classes\n", rang, count))
		}
		builder.WriteString("\n")
	}

	// Most coupled classes
	if len(response.Summary.MostCoupledClasses) > 0 {
		builder.WriteString("🔗 MOST COUPLED CLASSES\n")
		for i, class := range response.Summary.MostCoupledClasses {
			if i >= 10 { // Limit to top 10
				break
			}
			riskIcon := f.getRiskIcon(class.RiskLevel)
			builder.WriteString(fmt.Sprintf("%d. %s %s (CBO: %d) - %s:%d\n", 
				i+1, riskIcon, class.Name, class.Metrics.CouplingCount, class.FilePath, class.StartLine))
		}
		builder.WriteString("\n")
	}

	// Detailed class information
	if len(response.Classes) > 0 {
		builder.WriteString("📋 CLASS DETAILS\n")
		for _, class := range response.Classes {
			f.writeClassDetails(&builder, class)
			builder.WriteString("\n")
		}
	}

	// Warnings
	if len(response.Warnings) > 0 {
		builder.WriteString("⚠️  WARNINGS\n")
		for _, warning := range response.Warnings {
			builder.WriteString(fmt.Sprintf("• %s\n", warning))
		}
		builder.WriteString("\n")
	}

	// Errors
	if len(response.Errors) > 0 {
		builder.WriteString("❌ ERRORS\n")
		for _, err := range response.Errors {
			builder.WriteString(fmt.Sprintf("• %s\n", err))
		}
		builder.WriteString("\n")
	}

	// Footer
	builder.WriteString(fmt.Sprintf("Generated at: %s\n", response.GeneratedAt))
	builder.WriteString(fmt.Sprintf("Version: %s\n", response.Version))

	return builder.String(), nil
}

// writeClassDetails writes detailed information about a class
func (f *CBOFormatterImpl) writeClassDetails(builder *strings.Builder, class domain.ClassCoupling) {
	riskIcon := f.getRiskIcon(class.RiskLevel)
	
	builder.WriteString(fmt.Sprintf("%s %s (CBO: %d, Risk: %s)\n", 
		riskIcon, class.Name, class.Metrics.CouplingCount, class.RiskLevel))
	builder.WriteString(fmt.Sprintf("  Location: %s:%d-%d\n", class.FilePath, class.StartLine, class.EndLine))
	
	if class.IsAbstract {
		builder.WriteString("  Type: Abstract Class\n")
	}

	// Base classes
	if len(class.BaseClasses) > 0 {
		builder.WriteString(fmt.Sprintf("  Inherits from: %s\n", strings.Join(class.BaseClasses, ", ")))
	}

	// Dependency breakdown
	if class.Metrics.CouplingCount > 0 {
		builder.WriteString("  Dependencies:\n")
		if class.Metrics.InheritanceDependencies > 0 {
			builder.WriteString(fmt.Sprintf("    Inheritance: %d\n", class.Metrics.InheritanceDependencies))
		}
		if class.Metrics.TypeHintDependencies > 0 {
			builder.WriteString(fmt.Sprintf("    Type Hints: %d\n", class.Metrics.TypeHintDependencies))
		}
		if class.Metrics.InstantiationDependencies > 0 {
			builder.WriteString(fmt.Sprintf("    Instantiation: %d\n", class.Metrics.InstantiationDependencies))
		}
		if class.Metrics.AttributeAccessDependencies > 0 {
			builder.WriteString(fmt.Sprintf("    Attribute Access: %d\n", class.Metrics.AttributeAccessDependencies))
		}
		if class.Metrics.ImportDependencies > 0 {
			builder.WriteString(fmt.Sprintf("    Imports: %d\n", class.Metrics.ImportDependencies))
		}

		// List dependent classes
		if len(class.Metrics.DependentClasses) > 0 {
			builder.WriteString(fmt.Sprintf("    Coupled to: %s\n", strings.Join(class.Metrics.DependentClasses, ", ")))
		}
	}
}

// getRiskIcon returns an emoji icon for the risk level
func (f *CBOFormatterImpl) getRiskIcon(risk domain.RiskLevel) string {
	switch risk {
	case domain.RiskLevelLow:
		return "🟢"
	case domain.RiskLevelMedium:
		return "🟡"
	case domain.RiskLevelHigh:
		return "🔴"
	default:
		return "⚪"
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

	// HTML header
	builder.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CBO Analysis Report</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background-color: #f9f9f9;
        }
        .header {
            text-align: center;
            margin-bottom: 30px;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border-radius: 10px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
        }
        .summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .summary-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .summary-card h3 {
            margin-top: 0;
            color: #555;
            border-bottom: 2px solid #eee;
            padding-bottom: 10px;
        }
        .risk-low { color: #28a745; }
        .risk-medium { color: #ffc107; }
        .risk-high { color: #dc3545; }
        .classes-table {
            background: white;
            border-radius: 8px;
            overflow: hidden;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        th {
            background-color: #f8f9fa;
            font-weight: 600;
            color: #555;
        }
        tr:hover {
            background-color: #f5f5f5;
        }
        .cbo-badge {
            display: inline-block;
            padding: 4px 8px;
            border-radius: 12px;
            font-weight: bold;
            color: white;
            font-size: 12px;
        }
        .cbo-low { background-color: #28a745; }
        .cbo-medium { background-color: #ffc107; color: #000; }
        .cbo-high { background-color: #dc3545; }
        .dependencies {
            font-size: 12px;
            color: #666;
            margin-top: 4px;
        }
        .footer {
            text-align: center;
            margin-top: 40px;
            padding: 20px;
            background: white;
            border-radius: 8px;
            color: #666;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🔗 CBO Analysis Report</h1>
        <p>Coupling Between Objects Analysis</p>
    </div>
`)

	// Summary section
	builder.WriteString(`    <div class="summary">
        <div class="summary-card">
            <h3>📊 Overview</h3>`)
	builder.WriteString(fmt.Sprintf(`            <p><strong>Total Classes:</strong> %d</p>`, response.Summary.TotalClasses))
	builder.WriteString(fmt.Sprintf(`            <p><strong>Files Analyzed:</strong> %d</p>`, response.Summary.FilesAnalyzed))
	builder.WriteString(fmt.Sprintf(`            <p><strong>Average CBO:</strong> %.2f</p>`, response.Summary.AverageCBO))
	builder.WriteString(fmt.Sprintf(`            <p><strong>Max CBO:</strong> %d</p>`, response.Summary.MaxCBO))
	builder.WriteString(`        </div>
        <div class="summary-card">
            <h3>🚦 Risk Distribution</h3>`)
	builder.WriteString(fmt.Sprintf(`            <p class="risk-low"><strong>Low Risk:</strong> %d classes</p>`, response.Summary.LowRiskClasses))
	builder.WriteString(fmt.Sprintf(`            <p class="risk-medium"><strong>Medium Risk:</strong> %d classes</p>`, response.Summary.MediumRiskClasses))
	builder.WriteString(fmt.Sprintf(`            <p class="risk-high"><strong>High Risk:</strong> %d classes</p>`, response.Summary.HighRiskClasses))
	builder.WriteString(`        </div>
    </div>
`)

	// Classes table
	if len(response.Classes) > 0 {
		builder.WriteString(`    <div class="classes-table">
        <table>
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
			cboBadgeClass := f.getCBOBadgeClass(class.RiskLevel)
			builder.WriteString(fmt.Sprintf(`                <tr>
                    <td><strong>%s</strong></td>
                    <td><span class="cbo-badge %s">%d</span></td>
                    <td class="%s">%s</td>
                    <td>%s:%d</td>
                    <td>
`, class.Name, cboBadgeClass, class.Metrics.CouplingCount, 
   f.getRiskClass(class.RiskLevel), class.RiskLevel, class.FilePath, class.StartLine))

			if len(class.Metrics.DependentClasses) > 0 {
				builder.WriteString(strings.Join(class.Metrics.DependentClasses, ", "))
			} else {
				builder.WriteString("No dependencies")
			}

			// Add dependency breakdown
			if class.Metrics.CouplingCount > 0 {
				builder.WriteString(`<div class="dependencies">`)
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
				builder.WriteString(strings.Join(deps, ", "))
				builder.WriteString(`</div>`)
			}

			builder.WriteString(`                    </td>
                </tr>
`)
		}

		builder.WriteString(`            </tbody>
        </table>
    </div>
`)
	}

	// Footer
	builder.WriteString(fmt.Sprintf(`    <div class="footer">
        <p>Generated at: %s | Version: %s</p>
    </div>
</body>
</html>`, response.GeneratedAt, response.Version))

	return builder.String(), nil
}

// getCBOBadgeClass returns CSS class for CBO badge
func (f *CBOFormatterImpl) getCBOBadgeClass(risk domain.RiskLevel) string {
	switch risk {
	case domain.RiskLevelLow:
		return "cbo-low"
	case domain.RiskLevelMedium:
		return "cbo-medium"
	case domain.RiskLevelHigh:
		return "cbo-high"
	default:
		return "cbo-low"
	}
}

// getRiskClass returns CSS class for risk level
func (f *CBOFormatterImpl) getRiskClass(risk domain.RiskLevel) string {
	switch risk {
	case domain.RiskLevelLow:
		return "risk-low"
	case domain.RiskLevelMedium:
		return "risk-medium"
	case domain.RiskLevelHigh:
		return "risk-high"
	default:
		return "risk-low"
	}
}