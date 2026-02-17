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
		fmt.Fprint(writer, utils.FormatLabelWithIndent(SectionPadding, "Unique Fragments", response.Summary.TotalClones))
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
	fmt.Fprintf(writer, "Unique Fragments,%d\n", response.Summary.TotalClones)
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
		"sub": func(a, b int) int {
			return a - b
		},
		"mul100": func(v float64) float64 { return v * 100.0 },
		"slice": func(s []string, start, end int) []string {
			if start < 0 || start >= len(s) {
				return []string{}
			}
			if end > len(s) {
				end = len(s)
			}
			return s[start:end]
		},
		"scoreQuality": func(score int) string {
			switch {
			case score >= domain.ScoreThresholdExcellent:
				return "excellent"
			case score >= domain.ScoreThresholdGood:
				return "good"
			case score >= domain.ScoreThresholdFair:
				return "fair"
			default:
				return "poor"
			}
		},
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

        /* Score bars */
        .score-bars {
            margin: 20px 0;
        }
        .score-bar-item {
            margin-bottom: 24px;
        }
        .score-bar-header {
            display: flex;
            justify-content: space-between;
            margin-bottom: 6px;
            font-size: 14px;
        }
        .score-label {
            font-weight: 600;
            color: #333;
        }
        .score-value {
            font-weight: 700;
            color: #667eea;
        }
        .score-bar-container {
            width: 100%;
            height: 12px;
            background: #e0e0e0;
            border-radius: 6px;
            overflow: hidden;
            box-shadow: inset 0 1px 3px rgba(0,0,0,0.1);
        }
        .score-bar-fill {
            height: 100%;
            transition: width 0.3s ease;
            border-radius: 6px;
        }
        .score-excellent { background: linear-gradient(90deg, #4caf50, #66bb6a); }
        .score-good { background: linear-gradient(90deg, #8bc34a, #9ccc65); }
        .score-fair { background: linear-gradient(90deg, #ff9800, #ffa726); }
        .score-poor { background: linear-gradient(90deg, #f44336, #ef5350); }
        .score-detail {
            margin-top: 4px;
            font-size: 12px;
            color: #666;
        }

        /* Tab header with score badge */
        .tab-header-with-score {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 20px;
            padding-bottom: 12px;
            border-bottom: 2px solid #e0e0e0;
        }

        .score-badge-compact {
            display: inline-block;
            padding: 6px 14px;
            border-radius: 16px;
            font-size: 13px;
            font-weight: 700;
            color: white;
            white-space: nowrap;
        }
        .score-badge-compact.score-excellent {
            background: linear-gradient(135deg, #4caf50, #66bb6a);
            box-shadow: 0 2px 6px rgba(76, 175, 80, 0.4);
        }
        .score-badge-compact.score-good {
            background: linear-gradient(135deg, #8bc34a, #9ccc65);
            box-shadow: 0 2px 6px rgba(139, 195, 74, 0.4);
        }
        .score-badge-compact.score-fair {
            background: linear-gradient(135deg, #ff9800, #ffa726);
            box-shadow: 0 2px 6px rgba(255, 152, 0, 0.4);
        }
        .score-badge-compact.score-poor {
            background: linear-gradient(135deg, #f44336, #ef5350);
            box-shadow: 0 2px 6px rgba(244, 67, 54, 0.4);
        }
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
                <button class="tab-button" onclick="showTab('clone', this)">Clone</button>
                {{end}}
                {{if .Summary.CBOEnabled}}
                <button class="tab-button" onclick="showTab('cbo', this)">Coupling</button>
                {{end}}
                {{if .Summary.LCOMEnabled}}
                <button class="tab-button" onclick="showTab('lcom', this)">Cohesion</button>
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

                <h3 style="margin-top: 20px; margin-bottom: 16px; color: #2c3e50;">Quality Scores</h3>
                <div class="score-bars">
                    {{if .Summary.ComplexityEnabled}}
                    <div class="score-bar-item">
                        <div class="score-bar-header">
                            <span class="score-label">Complexity</span>
                            <span class="score-value">{{.Summary.ComplexityScore}}/100</span>
                        </div>
                        <div class="score-bar-container">
                            <div class="score-bar-fill score-{{scoreQuality .Summary.ComplexityScore}}" style="width: {{.Summary.ComplexityScore}}%"></div>
                        </div>
                        <div class="score-detail">Avg: {{printf "%.1f" .Summary.AverageComplexity}}, High-risk: {{.Summary.HighComplexityCount}}</div>
                    </div>
                    {{end}}

                    {{if .Summary.DeadCodeEnabled}}
                    <div class="score-bar-item">
                        <div class="score-bar-header">
                            <span class="score-label">Dead Code</span>
                            <span class="score-value">{{.Summary.DeadCodeScore}}/100</span>
                        </div>
                        <div class="score-bar-container">
                            <div class="score-bar-fill score-{{scoreQuality .Summary.DeadCodeScore}}" style="width: {{.Summary.DeadCodeScore}}%"></div>
                        </div>
                        <div class="score-detail">{{.Summary.DeadCodeCount}} issues, {{.Summary.CriticalDeadCode}} critical</div>
                    </div>
                    {{end}}

                    {{if .Summary.CloneEnabled}}
                    <div class="score-bar-item">
                        <div class="score-bar-header">
                            <span class="score-label">Duplication</span>
                            <span class="score-value">{{.Summary.DuplicationScore}}/100</span>
                        </div>
                        <div class="score-bar-container">
                            <div class="score-bar-fill score-{{scoreQuality .Summary.DuplicationScore}}" style="width: {{.Summary.DuplicationScore}}%"></div>
                        </div>
                        <div class="score-detail">{{printf "%.1f%%" .Summary.CodeDuplication}} duplication, {{.Summary.CloneGroups}} groups</div>
                    </div>
                    {{end}}

                    {{if .Summary.CBOEnabled}}
                    <div class="score-bar-item">
                        <div class="score-bar-header">
                            <span class="score-label">Coupling (CBO)</span>
                            <span class="score-value">{{.Summary.CouplingScore}}/100</span>
                        </div>
                        <div class="score-bar-container">
                            <div class="score-bar-fill score-{{scoreQuality .Summary.CouplingScore}}" style="width: {{.Summary.CouplingScore}}%"></div>
                        </div>
                        <div class="score-detail">Avg: {{printf "%.1f" .Summary.AverageCoupling}}, High-coupling: {{.Summary.HighCouplingClasses}}/{{.Summary.CBOClasses}}</div>
                    </div>
                    {{end}}

                    {{if .Summary.LCOMEnabled}}
                    <div class="score-bar-item">
                        <div class="score-bar-header">
                            <span class="score-label">Cohesion (LCOM)</span>
                            <span class="score-value">{{.Summary.CohesionScore}}/100</span>
                        </div>
                        <div class="score-bar-container">
                            <div class="score-bar-fill score-{{scoreQuality .Summary.CohesionScore}}" style="width: {{.Summary.CohesionScore}}%"></div>
                        </div>
                        <div class="score-detail">Avg: {{printf "%.1f" .Summary.AverageLCOM}}, Low-cohesion: {{.Summary.HighLCOMClasses}}/{{.Summary.LCOMClasses}}</div>
                    </div>
                    {{end}}

                    {{if .Summary.DepsEnabled}}
                    <div class="score-bar-item">
                        <div class="score-bar-header">
                            <span class="score-label">Dependencies</span>
                            <span class="score-value">{{.Summary.DependencyScore}}/100</span>
                        </div>
                        <div class="score-bar-container">
                            <div class="score-bar-fill score-{{scoreQuality .Summary.DependencyScore}}" style="width: {{.Summary.DependencyScore}}%"></div>
                        </div>
                        <div class="score-detail">{{if eq .Summary.DepsModulesInCycles 0}}No cycles{{else}}{{.Summary.DepsModulesInCycles}} cycles{{end}}, Depth: {{.Summary.DepsMaxDepth}}</div>
                    </div>
                    {{end}}

                    {{if .Summary.ArchEnabled}}
                    <div class="score-bar-item">
                        <div class="score-bar-header">
                            <span class="score-label">Architecture</span>
                            <span class="score-value">{{.Summary.ArchitectureScore}}/100</span>
                        </div>
                        <div class="score-bar-container">
                            <div class="score-bar-fill score-{{scoreQuality .Summary.ArchitectureScore}}" style="width: {{.Summary.ArchitectureScore}}%"></div>
                        </div>
                        <div class="score-detail">{{printf "%.0f%%" (mul100 .Summary.ArchCompliance)}} compliant</div>
                    </div>
                    {{end}}
                </div>

                <h3 style="margin-top: 24px; margin-bottom: 16px; color: #2c3e50;">File Statistics</h3>
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
                        <div class="metric-value">{{.Summary.TotalClones}}</div>
                        <div class="metric-label">Unique Fragments</div>
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
                    {{if .Summary.LCOMEnabled}}
                    <div class="metric-card">
                        <div class="metric-value">{{.Summary.LCOMClasses}}</div>
                        <div class="metric-label">Classes (LCOM)</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.Summary.HighLCOMClasses}}</div>
                        <div class="metric-label">Low Cohesion</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.2f" .Summary.AverageLCOM}}</div>
                        <div class="metric-label">Avg LCOM4</div>
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
                <div class="tab-header-with-score">
                    <h2 style="margin: 0;">Complexity Analysis</h2>
                    <div class="score-badge-compact score-{{scoreQuality .Summary.ComplexityScore}}">
                        {{.Summary.ComplexityScore}}/100
                    </div>
                </div>
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
                            <th>Nesting Depth</th>
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
                            <td>{{$f.Metrics.NestingDepth}}</td>
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
                <div class="tab-header-with-score">
                    <h2 style="margin: 0;">Dead Code Detection</h2>
                    <div class="score-badge-compact score-{{scoreQuality .Summary.DeadCodeScore}}">
                        {{.Summary.DeadCodeScore}}/100
                    </div>
                </div>
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
                <div class="tab-header-with-score">
                    <h2 style="margin: 0;">Clone</h2>
                    <div class="score-badge-compact score-{{scoreQuality .Summary.DuplicationScore}}">
                        {{.Summary.DuplicationScore}}/100
                    </div>
                </div>
                {{if .Clone}}
                <div class="metric-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{.Clone.Statistics.TotalClones}}</div>
                        <div class="metric-label">Unique Fragments</div>
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
                
                {{if gt .Clone.Statistics.TotalCloneGroups 0}}
                <h3>Clone Groups</h3>
                <p style="color: #666; margin-bottom: 15px;">Code fragments grouped by similarity</p>
                {{range $i, $group := .Clone.CloneGroups}}
                {{if lt $i 10}}
                <div style="background: #f8f9fa; padding: 15px; margin-bottom: 15px; border-radius: 8px; border-left: 4px solid #667eea;">
                    <h4 style="margin-top: 0; color: #333;">Group {{$group.ID}} - {{len $group.Clones}} clones (Type {{$group.Type}}, similarity: {{printf "%.2f" $group.Similarity}})</h4>
                    <table class="table" style="margin-bottom: 0;">
                        <thead>
                            <tr>
                                <th>File</th>
                                <th>Lines</th>
                                <th>Size</th>
                            </tr>
                        </thead>
                        <tbody>
                            {{range $j, $clone := $group.Clones}}
                            {{if lt $j 10}}
                            <tr>
                                <td>{{$clone.Location.FilePath}}</td>
                                <td>{{$clone.Location.StartLine}}-{{$clone.Location.EndLine}}</td>
                                <td>{{$clone.LineCount}} lines</td>
                            </tr>
                            {{end}}
                            {{end}}
                            {{if gt (len $group.Clones) 10}}
                            <tr>
                                <td colspan="3" style="color: #666; font-style: italic;">... and {{sub (len $group.Clones) 10}} more clones</td>
                            </tr>
                            {{end}}
                        </tbody>
                    </table>
                </div>
                {{end}}
                {{end}}
                {{if gt .Clone.Statistics.TotalCloneGroups 10}}
                <p style="color: #666; margin-top: 10px;">Showing top 10 of {{.Clone.Statistics.TotalCloneGroups}} clone groups</p>
                {{end}}
                {{else if gt .Clone.Statistics.TotalClonePairs 0}}
                <h3>Clone Pairs</h3>
                <p style="color: #666; margin-bottom: 15px;">No groups formed, showing individual pairs</p>
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
                <div class="tab-header-with-score">
                    <h2 style="margin: 0;">Coupling</h2>
                    <div class="score-badge-compact score-{{scoreQuality .Summary.CouplingScore}}">
                        {{.Summary.CouplingScore}}/100
                    </div>
                </div>
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

            {{if .Summary.LCOMEnabled}}
            <div id="lcom" class="tab-content">
                <div class="tab-header-with-score">
                    <h2 style="margin: 0;">Class Cohesion</h2>
                    <div class="score-badge-compact score-{{scoreQuality .Summary.CohesionScore}}">
                        {{.Summary.CohesionScore}}/100
                    </div>
                </div>
                <p style="margin-bottom: 20px; color: #666;">Lack of Cohesion of Methods (LCOM4) metrics</p>
                {{if .LCOM}}
                <div class="metric-grid">
                    <div class="metric-card">
                        <div class="metric-value">{{.LCOM.Summary.TotalClasses}}</div>
                        <div class="metric-label">Total Classes</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.LCOM.Summary.HighRiskClasses}}</div>
                        <div class="metric-label">Low Cohesion</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.2f" .LCOM.Summary.AverageLCOM}}</div>
                        <div class="metric-label">Average LCOM4</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.LCOM.Summary.MaxLCOM}}</div>
                        <div class="metric-label">Max LCOM4</div>
                    </div>
                </div>

                <h3>Least Cohesive Classes</h3>
                <table class="table">
                    <thead>
                        <tr>
                            <th>Class</th>
                            <th>File</th>
                            <th>LCOM4</th>
                            <th>Risk</th>
                            <th>Methods</th>
                            <th>Instance Vars</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range $i, $c := .LCOM.Classes}}
                        {{if lt $i 10}}
                        <tr>
                            <td>{{$c.Name}}</td>
                            <td>{{$c.FilePath}}:{{$c.StartLine}}</td>
                            <td>{{$c.Metrics.LCOM4}}</td>
                            <td class="risk-{{$c.RiskLevel}}">{{$c.RiskLevel}}</td>
                            <td>{{sub $c.Metrics.TotalMethods $c.Metrics.ExcludedMethods}}</td>
                            <td>{{$c.Metrics.InstanceVariables}}</td>
                        </tr>
                        {{end}}
                        {{end}}
                    </tbody>
                </table>
                {{if gt (len .LCOM.Classes) 10}}
                <p style="color: #666; margin-top: 10px;">Showing top 10 of {{len .LCOM.Classes}} classes</p>
                {{end}}
                {{end}}
            </div>
            {{end}}

            {{if .System}}
            {{if .System.DependencyAnalysis}}
            <div id="sys-deps" class="tab-content">
                <div class="tab-header-with-score">
                    <h2 style="margin: 0;">Module Dependencies</h2>
                    <div class="score-badge-compact score-{{scoreQuality .Summary.DependencyScore}}">
                        {{.Summary.DependencyScore}}/100
                    </div>
                </div>
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

                {{/* Circular Dependencies Details Section */}}
                {{if .System.DependencyAnalysis.CircularDependencies}}
                <h3 style="margin-top: 30px;">Circular Dependencies</h3>
                {{if not .System.DependencyAnalysis.CircularDependencies.HasCircularDependencies}}
                <div style="padding: 20px; background: #d4edda; border-left: 4px solid #28a745; border-radius: 4px; margin: 20px 0;">
                    <strong style="color: #155724;">✅ No circular dependencies detected</strong>
                    <p style="color: #155724; margin: 10px 0 0 0;">All modules have acyclic dependency relationships.</p>
                </div>
                {{else}}
                <table class="table">
                    <thead>
                        <tr>
                            <th style="width: 10%;">Severity</th>
                            <th style="width: 8%;">Size</th>
                            <th>Dependency Paths</th>
                        </tr>
                    </thead>
                    <tbody>
                        {{range $i, $cycle := .System.DependencyAnalysis.CircularDependencies.CircularDependencies}}
                        {{if lt $i 20}}
                        <tr>
                            <td>
                                {{if eq $cycle.Severity "critical"}}<span style="background: #f8d7da; color: #721c24; padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: bold;">CRITICAL</span>
                                {{else if eq $cycle.Severity "high"}}<span style="background: #fff3cd; color: #856404; padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: bold;">HIGH</span>
                                {{else if eq $cycle.Severity "medium"}}<span style="background: #fff3cd; color: #856404; padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: bold;">MEDIUM</span>
                                {{else}}<span style="background: #d1ecf1; color: #0c5460; padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: bold;">LOW</span>{{end}}
                            </td>
                            <td>{{$cycle.Size}}</td>
                            <td>
                                {{range $j, $path := $cycle.Dependencies}}
                                    {{if lt $j 5}}
                                        {{if gt $j 0}}<br>{{end}}
                                        <code style="font-size: 11px;">{{join $path.Path " → "}}</code>
                                    {{end}}
                                {{end}}
                                {{if gt (len $cycle.Dependencies) 5}}
                                    <br><em style="font-size: 11px; color: #666;">... and {{sub (len $cycle.Dependencies) 5}} more paths</em>
                                {{end}}
                            </td>
                        </tr>
                        {{end}}
                        {{end}}
                        {{if gt (len .System.DependencyAnalysis.CircularDependencies.CircularDependencies) 20}}
                        <tr>
                            <td colspan="3"><em>... and {{sub (len .System.DependencyAnalysis.CircularDependencies.CircularDependencies) 20}} more circular dependencies</em></td>
                        </tr>
                        {{end}}
                    </tbody>
                </table>

                {{/* Core Infrastructure Modules */}}
                {{if gt (len .System.DependencyAnalysis.CircularDependencies.CoreInfrastructure) 0}}
                <div style="padding: 15px; background: #fff3cd; border-left: 4px solid #ffc107; border-radius: 4px; margin: 20px 0;">
                    <strong style="color: #856404;">⚠️ Core Infrastructure Modules (appear in multiple cycles):</strong>
                    <p style="color: #856404; margin: 10px 0 0 0;">{{join .System.DependencyAnalysis.CircularDependencies.CoreInfrastructure ", "}}</p>
                </div>
                {{end}}

                {{/* Cycle Breaking Suggestions */}}
                {{if gt (len .System.DependencyAnalysis.CircularDependencies.CycleBreakingSuggestions) 0}}
                <div style="padding: 15px; background: #d1ecf1; border-left: 4px solid #17a2b8; border-radius: 4px; margin: 20px 0;">
                    <strong style="color: #0c5460;">💡 Suggestions for Breaking Cycles:</strong>
                    <ul style="margin: 10px 0 0 20px; color: #0c5460;">
                        {{range .System.DependencyAnalysis.CircularDependencies.CycleBreakingSuggestions}}
                        <li>{{.}}</li>
                        {{end}}
                    </ul>
                </div>
                {{end}}
                {{end}}
                {{end}}

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
                <div class="tab-header-with-score">
                    <h2 style="margin: 0;">Architecture Validation</h2>
                    <div class="score-badge-compact score-{{scoreQuality .Summary.ArchitectureScore}}">
                        {{.Summary.ArchitectureScore}}/100
                    </div>
                </div>
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
