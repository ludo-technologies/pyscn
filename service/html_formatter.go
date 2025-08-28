package service

import (
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"

	"github.com/pyqol/pyqol/domain"
)

// HTMLFormatterImpl provides common HTML formatting functionality with Lighthouse-style scoring
type HTMLFormatterImpl struct{}

// NewHTMLFormatter creates a new HTML formatter service
func NewHTMLFormatter() *HTMLFormatterImpl {
	return &HTMLFormatterImpl{}
}

// ScoreData represents scoring information for HTML output
type ScoreData struct {
	Score    int    `json:"score"`
	Label    string `json:"label"`
	Color    string `json:"color"`
	Status   string `json:"status"`
	Category string `json:"category"`
}

// OverallScoreData represents the combined score information
type OverallScoreData struct {
	Score       int         `json:"score"`
	Color       string      `json:"color"`
	Status      string      `json:"status"`
	Breakdown   []ScoreData `json:"breakdown"`
	ProjectName string      `json:"project_name"`
	Timestamp   string      `json:"timestamp"`
}

// ComplexityHTMLData represents complexity analysis data for HTML template
type ComplexityHTMLData struct {
	OverallScore OverallScoreData          `json:"overall_score"`
	Response     *domain.ComplexityResponse `json:"response"`
	ScoreDetails ScoreData                 `json:"score_details"`
}

// DeadCodeHTMLData represents dead code analysis data for HTML template
type DeadCodeHTMLData struct {
	OverallScore OverallScoreData        `json:"overall_score"`
	Response     *domain.DeadCodeResponse `json:"response"`
	ScoreDetails ScoreData               `json:"score_details"`
}

// CloneHTMLData represents clone detection data for HTML template
type CloneHTMLData struct {
	OverallScore OverallScoreData      `json:"overall_score"`
	Response     *domain.CloneResponse `json:"response"`
	ScoreDetails ScoreData             `json:"score_details"`
}

// CalculateComplexityScore calculates a Lighthouse-style score (0-100) for complexity
func (f *HTMLFormatterImpl) CalculateComplexityScore(response *domain.ComplexityResponse) ScoreData {
	if response == nil || response.Summary.TotalFunctions == 0 {
		return ScoreData{
			Score:    100,
			Label:    "No Functions",
			Color:    "#0CCE6B",
			Status:   "pass",
			Category: "complexity",
		}
	}

	// Score based on average complexity
	// Lower complexity = higher score
	avgComplexity := response.Summary.AverageComplexity
	
	// Use logarithmic scale to avoid extreme scores
	rawScore := 100 - (avgComplexity * 8) // Adjust multiplier as needed
	score := int(math.Max(0, math.Min(100, rawScore)))

	var color, status string
	switch {
	case score >= 90:
		color = "#0CCE6B" // Green
		status = "pass"
	case score >= 50:
		color = "#FFA500" // Orange
		status = "average"
	default:
		color = "#FF5722" // Red
		status = "fail"
	}

	return ScoreData{
		Score:    score,
		Label:    fmt.Sprintf("Avg Complexity: %.1f", avgComplexity),
		Color:    color,
		Status:   status,
		Category: "complexity",
	}
}

// CalculateDeadCodeScore calculates a Lighthouse-style score for dead code detection
func (f *HTMLFormatterImpl) CalculateDeadCodeScore(response *domain.DeadCodeResponse) ScoreData {
	if response.Summary.TotalBlocks == 0 {
		return ScoreData{
			Score:    100,
			Label:    "No Code Blocks",
			Color:    "#0CCE6B",
			Status:   "pass",
			Category: "dead_code",
		}
	}

	// Score based on reachable code ratio
	reachableRatio := 1.0 - response.Summary.OverallDeadRatio
	score := int(reachableRatio * 100)

	var color, status string
	switch {
	case score >= 90:
		color = "#0CCE6B" // Green
		status = "pass"
	case score >= 50:
		color = "#FFA500" // Orange
		status = "average"
	default:
		color = "#FF5722" // Red
		status = "fail"
	}

	return ScoreData{
		Score:    score,
		Label:    fmt.Sprintf("%.1f%% Reachable", reachableRatio*100),
		Color:    color,
		Status:   status,
		Category: "dead_code",
	}
}

// CalculateCloneScore calculates a Lighthouse-style score for clone detection
func (f *HTMLFormatterImpl) CalculateCloneScore(response *domain.CloneResponse) ScoreData {
	if response.Statistics == nil || response.Statistics.LinesAnalyzed == 0 {
		return ScoreData{
			Score:    100,
			Label:    "No Lines Analyzed",
			Color:    "#0CCE6B",
			Status:   "pass",
			Category: "clone",
		}
	}

	// Calculate score based on clone pairs density
	totalPairs := response.Statistics.TotalClonePairs
	linesAnalyzed := response.Statistics.LinesAnalyzed
	
	// Calculate clone density: pairs per 1000 lines of code
	cloneDensity := float64(totalPairs) / (float64(linesAnalyzed) / 1000.0)
	
	// Convert to score using logarithmic scale
	// 0 clones = 100 score, higher density = lower score
	var score int
	if totalPairs == 0 {
		score = 100
	} else {
		// Use log scale to prevent extreme scores, but less aggressive
		rawScore := 100 - (math.Log(cloneDensity + 1) * 10)
		score = int(math.Max(5, math.Min(100, rawScore))) // Minimum score of 5
	}

	var color, status string
	switch {
	case score >= 90:
		color = "#0CCE6B" // Green
		status = "pass"
	case score >= 50:
		color = "#FFA500" // Orange
		status = "average"
	default:
		color = "#FF5722" // Red
		status = "fail"
	}

	return ScoreData{
		Score:    score,
		Label:    fmt.Sprintf("%d Clone Pairs", totalPairs),
		Color:    color,
		Status:   status,
		Category: "clone",
	}
}

// CalculateOverallScore calculates weighted average of all scores
func (f *HTMLFormatterImpl) CalculateOverallScore(scores []ScoreData, projectName string) OverallScoreData {
	if len(scores) == 0 {
		return OverallScoreData{
			Score:       100,
			Color:       "#0CCE6B",
			Status:      "pass",
			Breakdown:   []ScoreData{},
			ProjectName: projectName,
			Timestamp:   time.Now().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	// Weighted average: Complexity 40%, Dead Code 30%, Clone 30%
	var weightedSum float64
	var totalWeight float64

	for _, score := range scores {
		var weight float64
		switch score.Category {
		case "complexity":
			weight = 0.4
		case "dead_code":
			weight = 0.3
		case "clone":
			weight = 0.3
		default:
			weight = 1.0 / float64(len(scores))
		}
		
		weightedSum += float64(score.Score) * weight
		totalWeight += weight
	}

	overallScore := int(weightedSum / totalWeight)

	var color, status string
	switch {
	case overallScore >= 90:
		color = "#0CCE6B" // Green
		status = "pass"
	case overallScore >= 50:
		color = "#FFA500" // Orange
		status = "average"
	default:
		color = "#FF5722" // Red
		status = "fail"
	}

	return OverallScoreData{
		Score:       overallScore,
		Color:       color,
		Status:      status,
		Breakdown:   scores,
		ProjectName: projectName,
		Timestamp:   time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}
}

// getHTMLTemplate returns the Lighthouse-style HTML template
func (f *HTMLFormatterImpl) getHTMLTemplate() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PyQol Code Quality Report - {{.OverallScore.ProjectName}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            background-color: #f5f5f5;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        
        .header {
            text-align: center;
            background: white;
            padding: 40px 20px;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            margin-bottom: 30px;
        }
        
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            color: #1a1a1a;
        }
        
        .header .timestamp {
            color: #666;
            font-size: 0.9em;
        }
        
        .score-section {
            display: flex;
            gap: 20px;
            margin-bottom: 30px;
            flex-wrap: wrap;
        }
        
        .score-card {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            flex: 1;
            min-width: 250px;
            text-align: center;
        }
        
        .score-gauge {
            position: relative;
            width: 120px;
            height: 120px;
            margin: 0 auto 20px;
        }
        
        .score-circle {
            width: 120px;
            height: 120px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 2em;
            font-weight: bold;
            color: white;
            position: relative;
        }
        
        .score-label {
            font-size: 1.1em;
            font-weight: 600;
            margin-bottom: 10px;
        }
        
        .score-description {
            color: #666;
            font-size: 0.9em;
        }
        
        .details-section {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            margin-bottom: 20px;
        }
        
        .details-section h2 {
            margin-bottom: 20px;
            color: #1a1a1a;
            border-bottom: 2px solid #eee;
            padding-bottom: 10px;
        }
        
        .metric-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        
        .metric-item {
            padding: 15px;
            background: #f8f9fa;
            border-radius: 4px;
        }
        
        .metric-item .value {
            font-size: 1.5em;
            font-weight: bold;
            color: #1a1a1a;
        }
        
        .metric-item .label {
            color: #666;
            font-size: 0.9em;
        }
        
        .risk-bar {
            height: 8px;
            background: #eee;
            border-radius: 4px;
            overflow: hidden;
            margin: 10px 0;
        }
        
        .risk-fill {
            height: 100%;
            transition: width 0.3s ease;
        }
        
        .risk-high { background-color: #FF5722; }
        .risk-medium { background-color: #FFA500; }
        .risk-low { background-color: #0CCE6B; }
        
        @media (max-width: 768px) {
            .score-section {
                flex-direction: column;
            }
            
            .metric-grid {
                grid-template-columns: 1fr;
            }
        }
        
        .footer {
            text-align: center;
            padding: 20px;
            color: #666;
            font-size: 0.9em;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Code Quality Report</h1>
            <div class="project-name">{{.OverallScore.ProjectName}}</div>
            <div class="timestamp">Generated on {{.OverallScore.Timestamp}}</div>
        </div>
        
        <div class="score-section">
            <div class="score-card">
                <div class="score-gauge">
                    <div class="score-circle" style="background-color: {{.OverallScore.Color}};">
                        {{.OverallScore.Score}}
                    </div>
                </div>
                <div class="score-label">Overall Score</div>
                <div class="score-description">Weighted average of all quality metrics</div>
            </div>
            
            {{range .OverallScore.Breakdown}}
            <div class="score-card">
                <div class="score-gauge">
                    <div class="score-circle" style="background-color: {{.Color}};">
                        {{.Score}}
                    </div>
                </div>
                <div class="score-label">{{.Category | title}} Score</div>
                <div class="score-description">{{.Label}}</div>
            </div>
            {{end}}
        </div>
        
        <div class="details-section">
            <h2>Analysis Details</h2>
            <div class="metric-grid">
                <div class="metric-item">
                    <div class="value">{{.ScoreDetails.Score}}</div>
                    <div class="label">{{.ScoreDetails.Category | title}} Score</div>
                </div>
                <div class="metric-item">
                    <div class="value">{{.ScoreDetails.Label}}</div>
                    <div class="label">Details</div>
                </div>
            </div>
        </div>
        
        <div class="footer">
            Generated by <strong>PyQol</strong> - Python Code Quality Analysis Tool
        </div>
    </div>
</body>
</html>`
}

// FormatComplexityAsHTML formats complexity analysis as HTML
func (f *HTMLFormatterImpl) FormatComplexityAsHTML(response *domain.ComplexityResponse, projectName string) (string, error) {
	if response == nil {
		return "", fmt.Errorf("response cannot be nil")
	}
	
	scoreDetails := f.CalculateComplexityScore(response)
	overallScore := f.CalculateOverallScore([]ScoreData{scoreDetails}, projectName)
	
	data := ComplexityHTMLData{
		OverallScore: overallScore,
		Response:     response,
		ScoreDetails: scoreDetails,
	}
	
	return f.renderTemplate(data)
}

// FormatDeadCodeAsHTML formats dead code analysis as HTML
func (f *HTMLFormatterImpl) FormatDeadCodeAsHTML(response *domain.DeadCodeResponse, projectName string) (string, error) {
	if response == nil {
		return "", fmt.Errorf("response cannot be nil")
	}
	
	scoreDetails := f.CalculateDeadCodeScore(response)
	overallScore := f.CalculateOverallScore([]ScoreData{scoreDetails}, projectName)
	
	data := DeadCodeHTMLData{
		OverallScore: overallScore,
		Response:     response,
		ScoreDetails: scoreDetails,
	}
	
	return f.renderTemplate(data)
}

// FormatCloneAsHTML formats clone detection as HTML
func (f *HTMLFormatterImpl) FormatCloneAsHTML(response *domain.CloneResponse, projectName string) (string, error) {
	if response == nil {
		return "", fmt.Errorf("response cannot be nil")
	}
	
	scoreDetails := f.CalculateCloneScore(response)
	overallScore := f.CalculateOverallScore([]ScoreData{scoreDetails}, projectName)
	
	data := CloneHTMLData{
		OverallScore: overallScore,
		Response:     response,
		ScoreDetails: scoreDetails,
	}
	
	return f.renderTemplate(data)
}

// renderTemplate renders the HTML template with the provided data
func (f *HTMLFormatterImpl) renderTemplate(data interface{}) (string, error) {
	tmplStr := f.getHTMLTemplate()
	
	// Add custom template functions
	funcMap := template.FuncMap{
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
		},
	}
	
	tmpl, err := template.New("html_report").Funcs(funcMap).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML template: %w", err)
	}
	
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute HTML template: %w", err)
	}
	
	return buf.String(), nil
}