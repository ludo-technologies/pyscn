package service

import (
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
)

// AnalyzeFormatter handles formatting of unified analysis reports
type AnalyzeFormatter struct {
	complexityFormatter *OutputFormatterImpl
	deadCodeFormatter   *DeadCodeFormatterImpl
	cloneFormatter      *CloneOutputFormatter
}

// NewAnalyzeFormatter creates a new analyze formatter
func NewAnalyzeFormatter() *AnalyzeFormatter {
	return &AnalyzeFormatter{
		complexityFormatter: NewOutputFormatter(),
		deadCodeFormatter:   NewDeadCodeFormatter(),
		cloneFormatter:      NewCloneOutputFormatter(),
	}
}

// Write formats and writes the unified analysis response
func (f *AnalyzeFormatter) Write(response *domain.AnalyzeResponse, format domain.OutputFormat, writer io.Writer) error {
	switch format {
	case domain.OutputFormatText:
		return f.writeText(response, writer)
	case domain.OutputFormatJSON:
		return WriteJSON(writer, response)
	case domain.OutputFormatYAML:
		return WriteYAML(writer, response)
	case domain.OutputFormatCSV:
		return f.writeCSV(response, writer)
	case domain.OutputFormatHTML:
		return f.writeHTML(response, writer)
	default:
		return domain.NewUnsupportedFormatError(string(format))
	}
}

// writeText formats the response as plain text
func (f *AnalyzeFormatter) writeText(response *domain.AnalyzeResponse, writer io.Writer) error {
	utils := NewFormatUtils()

	// Header
	fmt.Fprint(writer, utils.FormatMainHeader("Comprehensive Analysis Report"))

	// Overall health and duration
	healthStats := map[string]interface{}{
		"Health Score":      fmt.Sprintf("%d/100 (%s)", response.Summary.HealthScore, response.Summary.Grade),
		"Analysis Duration": fmt.Sprintf("%.2fs", float64(response.Duration)/1000.0),
		"Generated":         response.GeneratedAt.Format(time.RFC3339),
	}
	fmt.Fprint(writer, utils.FormatSummaryStats(healthStats))

	// File statistics
	fmt.Fprint(writer, utils.FormatFileStats(
		response.Summary.AnalyzedFiles,
		response.Summary.TotalFiles,
		response.Summary.TotalFiles-response.Summary.AnalyzedFiles))

	// Analysis modules results
	if response.Summary.ComplexityEnabled {
		fmt.Fprint(writer, utils.FormatSectionHeader("COMPLEXITY ANALYSIS"))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Total Functions", response.Summary.TotalFunctions))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Average Complexity", fmt.Sprintf("%.1f", response.Summary.AverageComplexity)))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "High Complexity Count", response.Summary.HighComplexityCount))
		fmt.Fprint(writer, utils.FormatSectionSeparator())
	}

	if response.Summary.DeadCodeEnabled {
		fmt.Fprint(writer, utils.FormatSectionHeader("DEAD CODE DETECTION"))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Total Issues", response.Summary.DeadCodeCount))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Critical Issues", response.Summary.CriticalDeadCode))
		fmt.Fprint(writer, utils.FormatSectionSeparator())
	}

	if response.Summary.CloneEnabled {
		fmt.Fprint(writer, utils.FormatSectionHeader("CLONE DETECTION"))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Clone Pairs", response.Summary.ClonePairs))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Clone Groups", response.Summary.CloneGroups))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Code Duplication", utils.FormatPercentage(response.Summary.CodeDuplication)))
		fmt.Fprint(writer, utils.FormatSectionSeparator())
	}

	if response.Summary.CBOEnabled {
		fmt.Fprint(writer, utils.FormatSectionHeader("DEPENDENCY ANALYSIS"))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Classes Analyzed", response.Summary.CBOClasses))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "High Coupling Classes", response.Summary.HighCouplingClasses))
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Average Coupling", fmt.Sprintf("%.1f", response.Summary.AverageCoupling)))
		fmt.Fprint(writer, utils.FormatSectionSeparator())
	}

	// Recommendations
	fmt.Fprint(writer, utils.FormatSectionHeader("RECOMMENDATIONS"))
	recommendationCount := 0

	if response.Summary.HighComplexityCount > 0 {
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "•",
			fmt.Sprintf("Refactor %d high-complexity functions", response.Summary.HighComplexityCount)))
		recommendationCount++
	}
	if response.Summary.DeadCodeCount > 0 {
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "•",
			fmt.Sprintf("Remove %d dead code segments", response.Summary.DeadCodeCount)))
		recommendationCount++
	}
	if response.Summary.CodeDuplication > 10 {
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "•",
			fmt.Sprintf("Reduce code duplication (currently %.1f%%)", response.Summary.CodeDuplication)))
		recommendationCount++
	}
	if response.Summary.HighCouplingClasses > 0 {
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "•",
			fmt.Sprintf("Reduce coupling in %d high-dependency classes", response.Summary.HighCouplingClasses)))
		recommendationCount++
	}

	if recommendationCount == 0 {
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Status", "No major issues detected"))
	}

	return nil
}

// writeJSON formats the response as JSON
// writeCSV formats the response as CSV (summary only)
func (f *AnalyzeFormatter) writeCSV(response *domain.AnalyzeResponse, writer io.Writer) error {
	// Write header
	fmt.Fprintf(writer, "Metric,Value\n")

	// Write summary metrics
	fmt.Fprintf(writer, "Health Score,%d\n", response.Summary.HealthScore)
	fmt.Fprintf(writer, "Grade,%s\n", response.Summary.Grade)
	fmt.Fprintf(writer, "Total Files,%d\n", response.Summary.TotalFiles)
	fmt.Fprintf(writer, "Analyzed Files,%d\n", response.Summary.AnalyzedFiles)
	fmt.Fprintf(writer, "Average Complexity,%.2f\n", response.Summary.AverageComplexity)
	fmt.Fprintf(writer, "High Complexity Count,%d\n", response.Summary.HighComplexityCount)
	fmt.Fprintf(writer, "Dead Code Count,%d\n", response.Summary.DeadCodeCount)
	fmt.Fprintf(writer, "Critical Dead Code,%d\n", response.Summary.CriticalDeadCode)
	fmt.Fprintf(writer, "Clone Pairs,%d\n", response.Summary.ClonePairs)
	fmt.Fprintf(writer, "Clone Groups,%d\n", response.Summary.CloneGroups)
	fmt.Fprintf(writer, "Code Duplication,%.2f\n", response.Summary.CodeDuplication)
	fmt.Fprintf(writer, "Total Classes Analyzed,%d\n", response.Summary.CBOClasses)
    fmt.Fprintf(writer, "High Coupling (CBO) Classes,%d\n", response.Summary.HighCouplingClasses)
    fmt.Fprintf(writer, "Average CBO,%.2f\n", response.Summary.AverageCoupling)

	return nil
}

// writeHTML formats the response as HTML
func (f *AnalyzeFormatter) writeHTML(response *domain.AnalyzeResponse, writer io.Writer) error {
    funcMap := template.FuncMap{
        "join": func(elems []string, sep string) string {
            return strings.Join(elems, sep)
        },
        "add": func(a, b int) int {
            return a + b
        },
        "mul100": func(v float64) float64 { return v * 100.0 },
    }
    tmpl := template.Must(template.New("analyze").Funcs(funcMap).Parse(analyzeHTMLTemplate))
    return tmpl.Execute(writer, response)
}

// HTML template for unified report
const analyzeHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>pyscn Analysis Report</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            background: white;
            border-radius: 10px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
        }
        .header h1 {
            color: #667eea;
            margin-bottom: 10px;
        }
        .score-badge {
            display: inline-block;
            padding: 10px 20px;
            border-radius: 50px;
            font-size: 24px;
            font-weight: bold;
            margin: 10px 0;
        }
        .grade-a { background: #4caf50; color: white; }
        .grade-b { background: #8bc34a; color: white; }
        .grade-c { background: #ff9800; color: white; }
        .grade-d { background: #ff5722; color: white; }
        .grade-f { background: #f44336; color: white; }
        
        .tabs {
            background: white;
            border-radius: 10px;
            overflow: hidden;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
        }
        .tab-buttons {
            display: flex;
            background: #f5f5f5;
        }
        .tab-button {
            flex: 1;
            padding: 15px;
            border: none;
            background: transparent;
            cursor: pointer;
            font-size: 16px;
            transition: all 0.3s;
        }
        .tab-button.active {
            background: white;
            color: #667eea;
            font-weight: bold;
        }
        .tab-content {
            display: none;
            padding: 30px;
        }
        .tab-content.active {
            display: block;
        }
        
        .metric-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .metric-card {
            background: #f8f9fa;
            padding: 20px;
            border-radius: 8px;
            text-align: center;
        }
        .metric-value {
            font-size: 32px;
            font-weight: bold;
            color: #667eea;
        }
        .metric-label {
            color: #666;
            margin-top: 5px;
        }
        
        .table {
            width: 100%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        .table th, .table td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        .table th {
            background: #f8f9fa;
            font-weight: 600;
        }
        
        .risk-low { color: #4caf50; }
        .risk-medium { color: #ff9800; }
        .risk-high { color: #f44336; }
        
        .severity-critical { color: #f44336; }
        .severity-warning { color: #ff9800; }
        .severity-info { color: #2196f3; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>pyscn Analysis Report</h1>
            <p>Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
            <div class="score-badge grade-{{if eq .Summary.Grade "A"}}a{{else if eq .Summary.Grade "B"}}b{{else if eq .Summary.Grade "C"}}c{{else if eq .Summary.Grade "D"}}d{{else}}f{{end}}">
                Health Score: {{.Summary.HealthScore}}/100 (Grade: {{.Summary.Grade}})
            </div>
        </div>

        <div class="tabs">
            <div class="tab-buttons">
                <button class="tab-button active" onclick="showTab('summary', this)">Summary</button>
                {{if .Summary.ComplexityEnabled}}
                <button class="tab-button" onclick="showTab('complexity', this)">Complexity</button>
                {{end}}
                {{if .Summary.DeadCodeEnabled}}
                <button class="tab-button" onclick="showTab('deadcode', this)">Dead Code</button>
                {{end}}
                {{if .Summary.CloneEnabled}}
                <button class="tab-button" onclick="showTab('clone', this)">Clone Detection</button>
                {{end}}
                {{if .Summary.CBOEnabled}}
                <button class="tab-button" onclick="showTab('cbo', this)">Class Coupling</button>
                {{end}}
                {{if .System}}
                {{if .System.DependencyAnalysis}}
                <button class="tab-button" onclick="showTab('sys-deps', this)">Dependencies</button>
                {{end}}
                {{if .System.ArchitectureAnalysis}}
                <button class="tab-button" onclick="showTab('sys-arch', this)">Architecture</button>
                {{end}}
                {{end}}
            </div>

            <div id="summary" class="tab-content active">
                <h2>Analysis Summary</h2>
                <div class="metric-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{.Summary.TotalFiles}}</div>
                        <div class="metric-label">Total Files</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.Summary.AnalyzedFiles}}</div>
                        <div class="metric-label">Analyzed Files</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.2f" .Summary.AverageComplexity}}</div>
                        <div class="metric-label">Avg Complexity</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.Summary.DeadCodeCount}}</div>
                        <div class="metric-label">Dead Code Issues</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.Summary.ClonePairs}}</div>
                        <div class="metric-label">Clone Pairs</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.1f%%" .Summary.CodeDuplication}}</div>
                        <div class="metric-label">Code Duplication</div>
                    </div>
                    {{if .Summary.CBOEnabled}}
                    <div class="metric-card">
                        <div class="metric-value">{{.Summary.CBOClasses}}</div>
                        <div class="metric-label">Total Classes</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.Summary.HighCouplingClasses}}</div>
                        <div class="metric-label">High Coupling (CBO)</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.2f" .Summary.AverageCoupling}}</div>
                        <div class="metric-label">Avg CBO</div>
                    </div>
                    {{end}}
                </div>

                {{/* System-level quick glance */}}
                {{if .System}}
                {{if .System.DependencyAnalysis}}
                <h3 style="margin-top: 16px; color: #2c3e50;">Dependencies</h3>
                <div class="metric-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{.System.DependencyAnalysis.TotalModules}}</div>
                        <div class="metric-label">Total Modules</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.System.DependencyAnalysis.TotalDependencies}}</div>
                        <div class="metric-label">Total Dependencies</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.System.DependencyAnalysis.MaxDepth}}</div>
                        <div class="metric-label">Max Depth</div>
                    </div>
                    {{if .System.DependencyAnalysis.CircularDependencies}}
                    <div class="metric-card">
                        <div class="metric-value">{{if .System.DependencyAnalysis.CircularDependencies.HasCircularDependencies}}❌ {{.System.DependencyAnalysis.CircularDependencies.TotalCycles}}{{else}}✅ 0{{end}}</div>
                        <div class="metric-label">Circular Dependencies</div>
                    </div>
                    {{end}}
                </div>
                {{end}}

                {{if .System.ArchitectureAnalysis}}
                <h3 style="margin-top: 8px; color: #2c3e50;">Architecture</h3>
                <div class="metric-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{.System.ArchitectureAnalysis.TotalViolations}}</div>
                        <div class="metric-label">Violations</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.1f%%" (mul100 .System.ArchitectureAnalysis.ComplianceScore)}}</div>
                        <div class="metric-label">Compliance</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{if .System.ArchitectureAnalysis.LayerAnalysis}}{{.System.ArchitectureAnalysis.LayerAnalysis.LayersAnalyzed}}{{else}}0{{end}}</div>
                        <div class="metric-label">Layers Analyzed</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.System.ArchitectureAnalysis.TotalRules}}</div>
                        <div class="metric-label">Total Rules</div>
                    </div>
                </div>
                {{end}}
                {{end}}
            </div>

            {{if .Summary.ComplexityEnabled}}
            <div id="complexity" class="tab-content">
                <h2>Complexity Analysis</h2>
                {{if .Complexity}}
                <div class="metric-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{len .Complexity.Functions}}</div>
                        <div class="metric-label">Total Functions</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.2f" .Complexity.Summary.AverageComplexity}}</div>
                        <div class="metric-label">Average</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.Complexity.Summary.MaxComplexity}}</div>
                        <div class="metric-label">Maximum</div>
                    </div>
                </div>
                
                <h3>Top Complex Functions</h3>
                <table class="table">
                    <thead>
                        <tr>
                            <th>Function</th>
                            <th>File</th>
                            <th>Complexity</th>
                            <th>Risk</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range $i, $f := .Complexity.Functions}}
                        {{if lt $i 10}}
                        <tr>
                            <td>{{$f.Name}}</td>
                            <td>{{$f.FilePath}}</td>
                            <td>{{$f.Metrics.Complexity}}</td>
                            <td class="risk-{{$f.RiskLevel}}">{{$f.RiskLevel}}</td>
                        </tr>
                        {{end}}
                        {{end}}
                    </tbody>
                </table>
                {{if gt (len .Complexity.Functions) 10}}
                <p style="color: #666; margin-top: 10px;">Showing top 10 of {{len .Complexity.Functions}} functions</p>
                {{end}}
                {{end}}
            </div>
            {{end}}

            {{if .Summary.DeadCodeEnabled}}
            <div id="deadcode" class="tab-content">
                <h2>Dead Code Detection</h2>
                {{if .DeadCode}}
                <div class="metric-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{.DeadCode.Summary.TotalFindings}}</div>
                        <div class="metric-label">Total Issues</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.DeadCode.Summary.CriticalFindings}}</div>
                        <div class="metric-label">Critical</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.DeadCode.Summary.WarningFindings}}</div>
                        <div class="metric-label">Warnings</div>
                    </div>
                </div>
                
                {{if gt .DeadCode.Summary.TotalFindings 0}}
                <h3>Top Dead Code Issues</h3>
                <table class="table">
                    <thead>
                        <tr>
                            <th>File</th>
                            <th>Function</th>
                            <th>Lines</th>
                            <th>Severity</th>
                            <th>Reason</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range $file := .DeadCode.Files}}
                        {{range $func := $file.Functions}}
                        {{range $i, $finding := $func.Findings}}
                        {{if lt $i 10}}
                        <tr>
                            <td>{{$finding.Location.FilePath}}</td>
                            <td>{{$finding.FunctionName}}</td>
                            <td>{{$finding.Location.StartLine}}-{{$finding.Location.EndLine}}</td>
                            <td class="severity-{{$finding.Severity}}">{{$finding.Severity}}</td>
                            <td>{{$finding.Reason}}</td>
                        </tr>
                        {{end}}
                        {{end}}
                        {{end}}
                        {{end}}
                    </tbody>
                </table>
                {{if gt .DeadCode.Summary.TotalFindings 10}}
                <p style="color: #666; margin-top: 10px;">Showing top 10 of {{.DeadCode.Summary.TotalFindings}} dead code issues</p>
                {{end}}
                {{else}}
                <p style="color: #4caf50; font-weight: bold; margin-top: 20px;">✓ No dead code detected</p>
                {{end}}
                {{end}}
            </div>
            {{end}}

            {{if .Summary.CloneEnabled}}
            <div id="clone" class="tab-content">
                <h2>Clone Detection</h2>
                {{if .Clone}}
                <div class="metric-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{.Clone.Statistics.TotalClonePairs}}</div>
                        <div class="metric-label">Clone Pairs</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.Clone.Statistics.TotalCloneGroups}}</div>
                        <div class="metric-label">Clone Groups</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.2f" .Clone.Statistics.AverageSimilarity}}</div>
                        <div class="metric-label">Avg Similarity</div>
                    </div>
                </div>
                
                {{if gt .Clone.Statistics.TotalClonePairs 0}}
                <h3>Major Clone Pairs</h3>
                <table class="table">
                    <thead>
                        <tr>
                            <th>File 1</th>
                            <th>File 2</th>
                            <th>Lines 1</th>
                            <th>Lines 2</th>
                            <th>Similarity</th>
                            <th>Type</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range $i, $pair := .Clone.ClonePairs}}
                        {{if lt $i 15}}
                        <tr>
                            <td>{{$pair.Clone1.Location.FilePath}}</td>
                            <td>{{$pair.Clone2.Location.FilePath}}</td>
                            <td>{{$pair.Clone1.Location.StartLine}}-{{$pair.Clone1.Location.EndLine}}</td>
                            <td>{{$pair.Clone2.Location.StartLine}}-{{$pair.Clone2.Location.EndLine}}</td>
                            <td>{{printf "%.3f" $pair.Similarity}}</td>
                            <td>{{$pair.Type}}</td>
                        </tr>
                        {{end}}
                        {{end}}
                    </tbody>
                </table>
                {{if gt .Clone.Statistics.TotalClonePairs 15}}
                <p style="color: #666; margin-top: 10px;">Showing top 15 of {{.Clone.Statistics.TotalClonePairs}} clone pairs</p>
                {{end}}
                {{else}}
                <p style="color: #4caf50; font-weight: bold; margin-top: 20px;">✓ No clones detected</p>
                {{end}}
                {{end}}
            </div>
            {{end}}

            {{if .Summary.CBOEnabled}}
            <div id="cbo" class="tab-content">
                <h2>Class Coupling</h2>
                <p style="margin-bottom: 20px; color: #666;">Coupling Between Objects (CBO) metrics</p>
                {{if .CBO}}
                <div class="metric-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{.CBO.Summary.TotalClasses}}</div>
                        <div class="metric-label">Total Classes</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.CBO.Summary.HighRiskClasses}}</div>
                        <div class="metric-label">High Risk Classes</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.2f" .CBO.Summary.AverageCBO}}</div>
                        <div class="metric-label">Average CBO</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.CBO.Summary.MaxCBO}}</div>
                        <div class="metric-label">Max CBO</div>
                    </div>
                </div>
                
                <h3>Most Dependent Classes</h3>
                <table class="table">
                    <thead>
                        <tr>
                            <th>Class</th>
                            <th>File</th>
                            <th>CBO</th>
                            <th>Risk Level</th>
                            <th>Dependent Classes</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range $i, $c := .CBO.Classes}}
                        {{if lt $i 10}}
                        <tr>
                            <td>{{$c.Name}}</td>
                            <td>{{$c.FilePath}}</td>
                            <td>{{$c.Metrics.CouplingCount}}</td>
                            <td class="risk-{{$c.RiskLevel}}">{{$c.RiskLevel}}</td>
                            <td>{{join $c.Metrics.DependentClasses ", "}}</td>
                        </tr>
                        {{end}}
                        {{end}}
                    </tbody>
                </table>
                {{if gt (len .CBO.Classes) 10}}
                <p style="color: #666; margin-top: 10px;">Showing top 10 of {{len .CBO.Classes}} classes</p>
                {{end}}
                {{end}}
            </div>
            {{end}}

            {{if .System}}
            {{if .System.DependencyAnalysis}}
            <div id="sys-deps" class="tab-content">
                <h2>Module Dependencies</h2>
                <p style="margin-bottom: 20px; color: #666;">Project-wide module dependency graph metrics</p>
                <div class="metric-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{.System.DependencyAnalysis.TotalModules}}</div>
                        <div class="metric-label">Total Modules</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.System.DependencyAnalysis.TotalDependencies}}</div>
                        <div class="metric-label">Total Dependencies</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.System.DependencyAnalysis.MaxDepth}}</div>
                        <div class="metric-label">Max Depth</div>
                    </div>
                    {{if .System.DependencyAnalysis.CircularDependencies}}
                    <div class="metric-card">
                        <div class="metric-value">{{if .System.DependencyAnalysis.CircularDependencies.HasCircularDependencies}}❌ {{.System.DependencyAnalysis.CircularDependencies.TotalCycles}}{{else}}✅ 0{{end}}</div>
                        <div class="metric-label">Circular Dependencies</div>
                    </div>
                    {{end}}
                </div>

                {{if gt (len .System.DependencyAnalysis.LongestChains) 0}}
                <h3>Longest Dependency Chains</h3>
                <table class="table">
                    <thead>
                        <tr>
                            <th>#</th>
                            <th>Depth</th>
                            <th>Path</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range $i, $chain := .System.DependencyAnalysis.LongestChains}}
                        {{if lt $i 10}}
                        <tr>
                            <td>{{add $i 1}}</td>
                            <td>{{$chain.Length}}</td>
                            <td>{{join $chain.Path " → "}}</td>
                        </tr>
                        {{end}}
                        {{end}}
                    </tbody>
                </table>
                {{end}}
            </div>
            {{end}}

            {{if .System.ArchitectureAnalysis}}
            <div id="sys-arch" class="tab-content">
                <h2>Architecture Validation</h2>
                <div class="metric-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{if .System.ArchitectureAnalysis.LayerAnalysis}}{{.System.ArchitectureAnalysis.LayerAnalysis.LayersAnalyzed}}{{else}}0{{end}}</div>
                        <div class="metric-label">Layers Analyzed</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.System.ArchitectureAnalysis.TotalRules}}</div>
                        <div class="metric-label">Total Rules</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.System.ArchitectureAnalysis.TotalViolations}}</div>
                        <div class="metric-label">Violations</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.1f%%" (mul100 .System.ArchitectureAnalysis.ComplianceScore)}}</div>
                        <div class="metric-label">Compliance</div>
                    </div>
                </div>

                {{if and .System.ArchitectureAnalysis.LayerAnalysis (gt (len .System.ArchitectureAnalysis.LayerAnalysis.LayerViolations) 0)}}
                <h3>Top Rule Violations</h3>
                <table class="table">
                    <thead>
                        <tr>
                            <th>Severity</th>
                            <th>Rule</th>
                            <th>From</th>
                            <th>To</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range $i, $v := .System.ArchitectureAnalysis.LayerAnalysis.LayerViolations}}
                        {{if lt $i 20}}
                        <tr>
                            <td>{{$v.Severity}}</td>
                            <td>{{$v.Rule}}</td>
                            <td>{{$v.FromModule}}</td>
                            <td>{{$v.ToModule}}</td>
                        </tr>
                        {{end}}
                        {{end}}
                    </tbody>
                </table>
                {{else}}
                <p style="color: #4caf50; font-weight: bold; margin-top: 20px;">✓ No architecture violations</p>
                {{end}}
            </div>
            {{end}}
            {{end}}
        </div>
    </div>

    <script>
        function showTab(tabName, el) {
            // Hide all tabs
            const tabs = document.querySelectorAll('.tab-content');
            tabs.forEach(tab => tab.classList.remove('active'));
            
            // Remove active from all buttons
            const buttons = document.querySelectorAll('.tab-button');
            buttons.forEach(btn => btn.classList.remove('active'));
            
            // Show selected tab
            document.getElementById(tabName).classList.add('active');
            
            // Mark button as active
            if (el) { el.classList.add('active'); }
        }
    </script>
</body>
</html>`
