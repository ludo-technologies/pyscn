package service

import (
	"fmt"
	"strings"
	"time"
)

// HTMLTemplate provides a standard HTML template for all reports
type HTMLTemplate struct {
	Title       string
	Subtitle    string
	GeneratedAt time.Time
	Version     string
	Duration    int64
	ScoreValue  int
	ScoreGrade  string
	ShowScore   bool
}

// GenerateHTMLHeader generates the standard HTML header with consistent styling
func (t *HTMLTemplate) GenerateHTMLHeader() string {
	var scoreHTML string
	if t.ShowScore {
		scoreHTML = fmt.Sprintf(`
        <div class="score-badge grade-%s">
            Health Score: %d/100 (Grade: %s)
        </div>`, strings.ToLower(t.ScoreGrade), t.ScoreValue, t.ScoreGrade)
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
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
        .header p {
            color: #666;
            font-size: 14px;
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
        .tab-button:hover {
            background: #e0e0e0;
        }
        .tab-button.active {
            background: white;
            color: #667eea;
            font-weight: bold;
        }
        .tab-content {
            display: none;
            padding: 30px;
            background: white;
        }
        .tab-content.active {
            display: block;
        }
        
        .content {
            background: white;
            border-radius: 10px;
            padding: 30px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.1);
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
            border-left: 4px solid #667eea;
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
        
        .section {
            margin: 30px 0;
        }
        .section-header {
            font-size: 24px;
            color: #2c3e50;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 2px solid #e9ecef;
        }
        
        .table {
            width: 100%%;
            border-collapse: collapse;
            margin: 20px 0;
        }
        .table th {
            background: #667eea;
            color: white;
            padding: 12px;
            text-align: left;
            font-weight: 600;
        }
        .table td {
            padding: 12px;
            border-bottom: 1px solid #e9ecef;
        }
        .table tr:hover {
            background: #f8f9fa;
        }
        
        .status-badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 12px;
            font-size: 12px;
            font-weight: bold;
        }
        .status-success { background: #d4edda; color: #155724; }
        .status-warning { background: #fff3cd; color: #856404; }
        .status-danger { background: #f8d7da; color: #721c24; }
        .status-info { background: #d1ecf1; color: #0c5460; }
        
        .footer {
            margin-top: 40px;
            padding: 20px;
            text-align: center;
            color: white;
            font-size: 14px;
        }
        
        /* Responsive */
        @media (max-width: 768px) {
            .container { padding: 10px; }
            .header { padding: 20px; }
            .tab-button { font-size: 14px; padding: 10px; }
            .metric-grid { grid-template-columns: 1fr; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s</h1>
            <p>%s</p>
            <p>Generated: %s | Duration: %dms | Version: %s</p>%s
        </div>`,
		t.Title,
		t.Title,
		t.Subtitle,
		t.GeneratedAt.Format("2006-01-02 15:04:05"),
		t.Duration,
		t.Version,
		scoreHTML)
}

// GenerateTabsStart generates the start of a tabbed interface
func GenerateTabsStart() string {
	return `
        <div class="tabs">
            <div class="tab-buttons">`
}

// GenerateTabButton generates a tab button
func GenerateTabButton(id, label string, active bool) string {
	activeClass := ""
	if active {
		activeClass = " active"
	}
	return fmt.Sprintf(`
                <button class="tab-button%s" onclick="showTab('%s')">%s</button>`,
		activeClass, id, label)
}

// GenerateTabsMiddle generates the middle section between tab buttons and content
func GenerateTabsMiddle() string {
	return `
            </div>`
}

// GenerateTabContent generates a tab content wrapper
func GenerateTabContent(id string, active bool, content string) string {
	activeClass := ""
	if active {
		activeClass = " active"
	}
	return fmt.Sprintf(`
            <div id="%s" class="tab-content%s">
                %s
            </div>`, id, activeClass, content)
}

// GenerateTabsEnd generates the end of a tabbed interface
func GenerateTabsEnd() string {
	return `
        </div>`
}

// GenerateSinglePageContent generates a single page content wrapper (no tabs)
func GenerateSinglePageContent(content string) string {
	return fmt.Sprintf(`
        <div class="content">
            %s
        </div>`, content)
}

// GenerateTabScript generates the JavaScript for tab switching
func GenerateTabScript() string {
	return `
    <script>
        function showTab(tabId) {
            // Hide all tabs
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.classList.remove('active');
            });
            document.querySelectorAll('.tab-button').forEach(button => {
                button.classList.remove('active');
            });
            
            // Show selected tab
            document.getElementById(tabId).classList.add('active');
            event.target.classList.add('active');
        }
    </script>`
}

// GenerateHTMLFooter generates the standard HTML footer
func GenerateHTMLFooter() string {
	return `
        <div class="footer">
            Generated by pyscn - Python Static Code Analyzer
        </div>
    </div>
    </body>
</html>`
}

// GenerateMetricCard generates a metric card HTML
func GenerateMetricCard(value, label string) string {
	return fmt.Sprintf(`
        <div class="metric-card">
            <div class="metric-value">%s</div>
            <div class="metric-label">%s</div>
        </div>`, value, label)
}

// GenerateSectionHeader generates a section header
func GenerateSectionHeader(title string) string {
	return fmt.Sprintf(`
        <h2 class="section-header">%s</h2>`, title)
}

// GenerateStatusBadge generates a status badge based on severity
func GenerateStatusBadge(text, severity string) string {
	class := "status-info"
	switch strings.ToLower(severity) {
	case "success", "good", "low":
		class = "status-success"
	case "warning", "medium":
		class = "status-warning"
	case "danger", "high", "critical":
		class = "status-danger"
	}
	return fmt.Sprintf(`<span class="status-badge %s">%s</span>`, class, text)
}
