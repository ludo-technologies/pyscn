package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

// MockDataFormatterImpl implements the MockDataFormatter interface
type MockDataFormatterImpl struct{}

// NewMockDataFormatter creates a new mock data formatter service
func NewMockDataFormatter() *MockDataFormatterImpl {
	return &MockDataFormatterImpl{}
}

// Format formats the mock data analysis response according to the specified format
func (f *MockDataFormatterImpl) Format(response *domain.MockDataResponse, format domain.OutputFormat) (string, error) {
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
		return "", domain.NewUnsupportedFormatError(string(format))
	}
}

// Write writes the formatted mock data output to the writer
func (f *MockDataFormatterImpl) Write(response *domain.MockDataResponse, format domain.OutputFormat, writer io.Writer) error {
	output, err := f.Format(response, format)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(output))
	if err != nil {
		return domain.NewOutputError("failed to write output", err)
	}

	return nil
}

// formatText formats the response as human-readable text
func (f *MockDataFormatterImpl) formatText(response *domain.MockDataResponse) (string, error) {
	var output strings.Builder
	utils := NewFormatUtils()

	// Header
	output.WriteString(utils.FormatMainHeader("Mock Data Analysis Report"))

	// Summary
	stats := map[string]interface{}{
		"Total Files":          response.Summary.TotalFiles,
		"Files with Mock Data": response.Summary.FilesWithMockData,
		"Total Findings":       response.Summary.TotalFindings,
	}
	output.WriteString(utils.FormatSummaryStats(stats))

	// Severity distribution
	output.WriteString(utils.FormatRiskDistribution(
		response.Summary.ErrorFindings,
		response.Summary.WarningFindings,
		response.Summary.InfoFindings))

	// Type distribution
	if len(response.Summary.FindingsByType) > 0 {
		output.WriteString(utils.FormatSectionHeader("FINDINGS BY TYPE"))
		for typeKey, count := range response.Summary.FindingsByType {
			output.WriteString(utils.FormatLabelWithIndent(SectionPadding, string(typeKey), fmt.Sprintf("%d", count)))
		}
		output.WriteString(utils.FormatSectionSeparator())
	}

	// File Details
	if len(response.Files) > 0 && response.Summary.TotalFindings > 0 {
		output.WriteString(utils.FormatSectionHeader("DETAILED FINDINGS"))

		for _, file := range response.Files {
			if len(file.Findings) > 0 {
				output.WriteString(utils.FormatLabelWithIndent(0, "File", file.FilePath))
				output.WriteString(strings.Repeat("-", HeaderWidth) + "\n")

				for _, finding := range file.Findings {
					output.WriteString(f.formatFindingText(finding, utils) + "\n")
				}
				output.WriteString("\n")
			}
		}
		output.WriteString(utils.FormatSectionSeparator())
	}

	// Warnings
	if len(response.Warnings) > 0 {
		output.WriteString(utils.FormatWarningsSection(response.Warnings))
	}

	// Errors
	if len(response.Errors) > 0 {
		output.WriteString(utils.FormatSectionHeader("ERRORS"))
		for _, error := range response.Errors {
			output.WriteString(utils.FormatLabelWithIndent(SectionPadding, "X", error))
		}
		output.WriteString(utils.FormatSectionSeparator())
	}

	return output.String(), nil
}

// formatFindingText formats a single finding as text
func (f *MockDataFormatterImpl) formatFindingText(finding domain.MockDataFinding, utils *FormatUtils) string {
	var standardRisk RiskLevel
	switch finding.Severity {
	case domain.MockDataSeverityError:
		standardRisk = RiskHigh
	case domain.MockDataSeverityWarning:
		standardRisk = RiskMedium
	case domain.MockDataSeverityInfo:
		standardRisk = RiskLow
	default:
		standardRisk = RiskLow
	}

	coloredSeverity := utils.FormatRiskWithColor(standardRisk)

	// Truncate long values
	value := finding.Value
	if len(value) > 50 {
		value = value[:47] + "..."
	}

	return fmt.Sprintf("    [%s] Line %d: %s - %s (%s)",
		coloredSeverity,
		finding.Location.StartLine,
		finding.Type,
		value,
		finding.Rationale)
}

// formatJSON formats the response as JSON
func (f *MockDataFormatterImpl) formatJSON(response *domain.MockDataResponse) (string, error) {
	return EncodeJSON(response)
}

// formatYAML formats the response as YAML
func (f *MockDataFormatterImpl) formatYAML(response *domain.MockDataResponse) (string, error) {
	return EncodeYAML(response)
}

// formatCSV formats the response as CSV
func (f *MockDataFormatterImpl) formatCSV(response *domain.MockDataResponse) (string, error) {
	var output strings.Builder
	writer := csv.NewWriter(&output)

	// Write header
	header := []string{"File", "Line", "Type", "Severity", "Value", "Description", "Rationale"}
	if err := writer.Write(header); err != nil {
		return "", domain.NewOutputError("failed to write CSV header", err)
	}

	// Write findings
	for _, file := range response.Files {
		for _, finding := range file.Findings {
			record := []string{
				finding.Location.FilePath,
				fmt.Sprintf("%d", finding.Location.StartLine),
				string(finding.Type),
				string(finding.Severity),
				finding.Value,
				finding.Description,
				finding.Rationale,
			}
			if err := writer.Write(record); err != nil {
				return "", domain.NewOutputError("failed to write CSV record", err)
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", domain.NewOutputError("CSV write error", err)
	}

	return output.String(), nil
}

// formatHTML formats the response as HTML
func (f *MockDataFormatterImpl) formatHTML(response *domain.MockDataResponse) (string, error) {
	var output strings.Builder

	output.WriteString(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Mock Data Analysis Report</title>
    <style>
        :root {
            --bg-color: #1a1a2e;
            --card-bg: #16213e;
            --text-color: #eee;
            --border-color: #0f3460;
            --error-color: #e94560;
            --warning-color: #f39c12;
            --info-color: #3498db;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-color);
            margin: 0;
            padding: 20px;
        }
        .container { max-width: 1200px; margin: 0 auto; }
        h1 { color: #fff; border-bottom: 2px solid var(--border-color); padding-bottom: 10px; }
        .summary-cards { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; margin-bottom: 30px; }
        .card {
            background: var(--card-bg);
            padding: 20px;
            border-radius: 8px;
            border: 1px solid var(--border-color);
        }
        .card h3 { margin: 0 0 10px 0; color: #888; font-size: 14px; }
        .card .value { font-size: 32px; font-weight: bold; }
        .findings-table { width: 100%; border-collapse: collapse; }
        .findings-table th, .findings-table td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid var(--border-color);
        }
        .findings-table th { background: var(--card-bg); }
        .severity-error { color: var(--error-color); }
        .severity-warning { color: var(--warning-color); }
        .severity-info { color: var(--info-color); }
        .type-badge {
            display: inline-block;
            padding: 2px 8px;
            border-radius: 4px;
            font-size: 12px;
            background: var(--border-color);
        }
        .value-cell { max-width: 300px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Mock Data Analysis Report</h1>

        <div class="summary-cards">
            <div class="card">
                <h3>Total Files</h3>
                <div class="value">` + fmt.Sprintf("%d", response.Summary.TotalFiles) + `</div>
            </div>
            <div class="card">
                <h3>Files with Mock Data</h3>
                <div class="value">` + fmt.Sprintf("%d", response.Summary.FilesWithMockData) + `</div>
            </div>
            <div class="card">
                <h3>Total Findings</h3>
                <div class="value">` + fmt.Sprintf("%d", response.Summary.TotalFindings) + `</div>
            </div>
            <div class="card">
                <h3>Error / Warning / Info</h3>
                <div class="value">
                    <span class="severity-error">` + fmt.Sprintf("%d", response.Summary.ErrorFindings) + `</span> /
                    <span class="severity-warning">` + fmt.Sprintf("%d", response.Summary.WarningFindings) + `</span> /
                    <span class="severity-info">` + fmt.Sprintf("%d", response.Summary.InfoFindings) + `</span>
                </div>
            </div>
        </div>
`)

	if len(response.Files) > 0 {
		output.WriteString(`
        <h2>Findings</h2>
        <table class="findings-table">
            <thead>
                <tr>
                    <th>File</th>
                    <th>Line</th>
                    <th>Type</th>
                    <th>Severity</th>
                    <th>Value</th>
                    <th>Rationale</th>
                </tr>
            </thead>
            <tbody>
`)
		for _, file := range response.Files {
			for _, finding := range file.Findings {
				severityClass := "severity-info"
				switch finding.Severity {
				case domain.MockDataSeverityError:
					severityClass = "severity-error"
				case domain.MockDataSeverityWarning:
					severityClass = "severity-warning"
				}

				value := finding.Value
				if len(value) > 50 {
					value = value[:47] + "..."
				}

				output.WriteString(fmt.Sprintf(`
                <tr>
                    <td>%s</td>
                    <td>%d</td>
                    <td><span class="type-badge">%s</span></td>
                    <td class="%s">%s</td>
                    <td class="value-cell" title="%s">%s</td>
                    <td>%s</td>
                </tr>
`, file.FilePath, finding.Location.StartLine, finding.Type, severityClass, finding.Severity, finding.Value, value, finding.Rationale))
			}
		}
		output.WriteString(`
            </tbody>
        </table>
`)
	}

	output.WriteString(`
        <footer style="margin-top: 40px; color: #666; font-size: 12px;">
            Generated at: ` + response.GeneratedAt + ` | Version: ` + response.Version + `
        </footer>
    </div>
</body>
</html>
`)

	return output.String(), nil
}
