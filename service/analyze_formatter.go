package service

import (
    "fmt"
    "html/template"
    "io"
    "strings"
    "time"

    "github.com/pyqol/pyqol/domain"
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
	fmt.Fprintf(writer, "PyQOL Comprehensive Analysis Report\n")
	fmt.Fprintf(writer, "====================================\n\n")
	fmt.Fprintf(writer, "Generated: %s\n\n", response.GeneratedAt.Format(time.RFC3339))

	// Summary section
	fmt.Fprintf(writer, "Overall Health Score: %d/100 (Grade: %s)\n", 
		response.Summary.HealthScore, response.Summary.Grade)
	fmt.Fprintf(writer, "Analysis Duration: %.2fs\n\n", float64(response.Duration)/1000.0)

	// File statistics
	fmt.Fprintf(writer, "File Statistics:\n")
	fmt.Fprintf(writer, "  Total Files: %d\n", response.Summary.TotalFiles)
	fmt.Fprintf(writer, "  Analyzed: %d\n", response.Summary.AnalyzedFiles)
	fmt.Fprintf(writer, "  Skipped: %d\n\n", response.Summary.SkippedFiles)

	// Complexity analysis results
	if response.Complexity != nil && response.Summary.ComplexityEnabled {
		fmt.Fprintf(writer, "Complexity Analysis:\n")
		fmt.Fprintf(writer, "--------------------\n")
		fmt.Fprintf(writer, "  Total Functions: %d\n", response.Summary.TotalFunctions)
		fmt.Fprintf(writer, "  Average Complexity: %.2f\n", response.Summary.AverageComplexity)
		fmt.Fprintf(writer, "  High Complexity Count: %d\n\n", response.Summary.HighComplexityCount)
	}

	// Dead code analysis results
	if response.DeadCode != nil && response.Summary.DeadCodeEnabled {
		fmt.Fprintf(writer, "Dead Code Detection:\n")
		fmt.Fprintf(writer, "-------------------\n")
		fmt.Fprintf(writer, "  Total Issues: %d\n", response.Summary.DeadCodeCount)
		fmt.Fprintf(writer, "  Critical Issues: %d\n\n", response.Summary.CriticalDeadCode)
	}

	// Clone detection results
	if response.Clone != nil && response.Summary.CloneEnabled {
		fmt.Fprintf(writer, "Clone Detection:\n")
		fmt.Fprintf(writer, "---------------\n")
		fmt.Fprintf(writer, "  Clone Pairs: %d\n", response.Summary.ClonePairs)
		fmt.Fprintf(writer, "  Clone Groups: %d\n", response.Summary.CloneGroups)
		fmt.Fprintf(writer, "  Code Duplication: %.2f%%\n\n", response.Summary.CodeDuplication)
	}

	// Recommendations
	fmt.Fprintf(writer, "Recommendations:\n")
	fmt.Fprintf(writer, "---------------\n")
	if response.Summary.HighComplexityCount > 0 {
		fmt.Fprintf(writer, "  • Refactor %d high-complexity functions\n", response.Summary.HighComplexityCount)
	}
	if response.Summary.DeadCodeCount > 0 {
		fmt.Fprintf(writer, "  • Remove %d dead code segments\n", response.Summary.DeadCodeCount)
	}
	if response.Summary.CodeDuplication > 10 {
		fmt.Fprintf(writer, "  • Reduce code duplication (currently %.1f%%)\n", response.Summary.CodeDuplication)
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
	fmt.Fprintf(writer, "CBO Classes,%d\n", response.Summary.CBOClasses)
	fmt.Fprintf(writer, "High Coupling Classes,%d\n", response.Summary.HighCouplingClasses)
	fmt.Fprintf(writer, "Average Coupling,%.2f\n", response.Summary.AverageCoupling)
	
	return nil
}

// writeHTML formats the response as HTML
func (f *AnalyzeFormatter) writeHTML(response *domain.AnalyzeResponse, writer io.Writer) error {
	funcMap := template.FuncMap{
		"join": func(elems []string, sep string) string {
			return strings.Join(elems, sep)
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
    <title>PyQOL Analysis Report</title>
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
            <h1>PyQOL Analysis Report</h1>
            <p>Generated: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
            <div class="score-badge grade-{{if eq .Summary.Grade "A"}}a{{else if eq .Summary.Grade "B"}}b{{else if eq .Summary.Grade "C"}}c{{else if eq .Summary.Grade "D"}}d{{else}}f{{end}}">
                Health Score: {{.Summary.HealthScore}}/100 (Grade: {{.Summary.Grade}})
            </div>
        </div>

        <div class="tabs">
            <div class="tab-buttons">
                <button class="tab-button active" onclick="showTab('summary')">Summary</button>
                {{if .Summary.ComplexityEnabled}}
                <button class="tab-button" onclick="showTab('complexity')">Complexity</button>
                {{end}}
                {{if .Summary.DeadCodeEnabled}}
                <button class="tab-button" onclick="showTab('deadcode')">Dead Code</button>
                {{end}}
                {{if .Summary.CloneEnabled}}
                <button class="tab-button" onclick="showTab('clone')">Clone Detection</button>
                {{end}}
                {{if .Summary.CBOEnabled}}
                <button class="tab-button" onclick="showTab('cbo')">CBO Analysis</button>
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
                        <div class="metric-label">CBO Classes</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{.Summary.HighCouplingClasses}}</div>
                        <div class="metric-label">High Coupling</div>
                    </div>
                    <div class="metric-card">
                        <div class="metric-value">{{printf "%.2f" .Summary.AverageCoupling}}</div>
                        <div class="metric-label">Avg Coupling</div>
                    </div>
                    {{end}}
                </div>
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
                {{end}}
            </div>
            {{end}}

            {{if .Summary.CBOEnabled}}
            <div id="cbo" class="tab-content">
                <h2>CBO Analysis</h2>
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
                
                <h3>Top Coupled Classes</h3>
                <table class="table">
                    <thead>
                        <tr>
                            <th>Class</th>
                            <th>File</th>
                            <th>CBO Count</th>
                            <th>Risk Level</th>
                            <th>Dependencies</th>
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
                {{end}}
            </div>
            {{end}}
        </div>
    </div>

    <script>
        function showTab(tabName) {
            // Hide all tabs
            const tabs = document.querySelectorAll('.tab-content');
            tabs.forEach(tab => tab.classList.remove('active'));
            
            // Remove active from all buttons
            const buttons = document.querySelectorAll('.tab-button');
            buttons.forEach(btn => btn.classList.remove('active'));
            
            // Show selected tab
            document.getElementById(tabName).classList.add('active');
            
            // Mark button as active
            event.target.classList.add('active');
        }
    </script>
</body>
</html>`
