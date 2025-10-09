package service

import (
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
)

func TestHTMLFormatter_NewHTMLFormatter(t *testing.T) {
	formatter := NewHTMLFormatter()
	assert.NotNil(t, formatter)
	assert.IsType(t, &HTMLFormatterImpl{}, formatter)
}

func TestHTMLFormatter_CalculateComplexityScore(t *testing.T) {
	formatter := NewHTMLFormatter()

	tests := []struct {
		name     string
		response *domain.ComplexityResponse
		expected struct {
			minScore int
			maxScore int
			status   string
		}
	}{
		{
			name: "No functions",
			response: &domain.ComplexityResponse{
				Summary: domain.ComplexitySummary{
					TotalFunctions: 0,
				},
			},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{100, 100, "pass"},
		},
		{
			name: "Low complexity",
			response: &domain.ComplexityResponse{
				Summary: domain.ComplexitySummary{
					TotalFunctions:    5,
					AverageComplexity: 2.0,
				},
			},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{80, 90, "average"},
		},
		{
			name: "High complexity",
			response: &domain.ComplexityResponse{
				Summary: domain.ComplexitySummary{
					TotalFunctions:    10,
					AverageComplexity: 15.0,
				},
			},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{0, 50, "fail"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := formatter.CalculateComplexityScore(tt.response)

			assert.GreaterOrEqual(t, score.Score, tt.expected.minScore)
			assert.LessOrEqual(t, score.Score, tt.expected.maxScore)
			assert.Equal(t, tt.expected.status, score.Status)
			assert.Equal(t, "complexity", score.Category)
		})
	}
}

func TestHTMLFormatter_CalculateDeadCodeScore(t *testing.T) {
	formatter := NewHTMLFormatter()

	tests := []struct {
		name     string
		response *domain.DeadCodeResponse
		expected struct {
			minScore int
			maxScore int
			status   string
		}
	}{
		{
			name: "No code blocks",
			response: &domain.DeadCodeResponse{
				Summary: domain.DeadCodeSummary{
					TotalBlocks: 0,
				},
			},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{100, 100, "pass"},
		},
		{
			name: "No dead code",
			response: &domain.DeadCodeResponse{
				Summary: domain.DeadCodeSummary{
					TotalBlocks:      100,
					DeadBlocks:       0,
					OverallDeadRatio: 0.0,
				},
			},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{100, 100, "pass"},
		},
		{
			name: "Some dead code",
			response: &domain.DeadCodeResponse{
				Summary: domain.DeadCodeSummary{
					TotalBlocks:      100,
					DeadBlocks:       50,
					OverallDeadRatio: 0.5,
				},
			},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{40, 60, "average"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := formatter.CalculateDeadCodeScore(tt.response)

			assert.GreaterOrEqual(t, score.Score, tt.expected.minScore)
			assert.LessOrEqual(t, score.Score, tt.expected.maxScore)
			assert.Equal(t, tt.expected.status, score.Status)
			assert.Equal(t, "dead_code", score.Category)
		})
	}
}

func TestHTMLFormatter_CalculateCloneScore(t *testing.T) {
	formatter := NewHTMLFormatter()

	tests := []struct {
		name     string
		response *domain.CloneResponse
		expected struct {
			minScore int
			maxScore int
			status   string
		}
	}{
		{
			name: "No analysis data",
			response: &domain.CloneResponse{
				Statistics: nil,
			},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{100, 100, "pass"},
		},
		{
			name: "No clones found",
			response: &domain.CloneResponse{
				Statistics: &domain.CloneStatistics{
					LinesAnalyzed:   1000,
					TotalClonePairs: 0,
				},
			},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{100, 100, "pass"},
		},
		{
			name: "Some clones found",
			response: &domain.CloneResponse{
				Statistics: &domain.CloneStatistics{
					LinesAnalyzed:   1000,
					TotalClonePairs: 10,
				},
			},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{60, 80, "average"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := formatter.CalculateCloneScore(tt.response)

			assert.GreaterOrEqual(t, score.Score, tt.expected.minScore)
			assert.LessOrEqual(t, score.Score, tt.expected.maxScore)
			assert.Equal(t, tt.expected.status, score.Status)
			assert.Equal(t, "clone", score.Category)
		})
	}
}

func TestHTMLFormatter_CalculateOverallScore(t *testing.T) {
	formatter := NewHTMLFormatter()

	tests := []struct {
		name     string
		scores   []ScoreData
		expected struct {
			minScore int
			maxScore int
			status   string
		}
	}{
		{
			name:   "No scores",
			scores: []ScoreData{},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{100, 100, "pass"},
		},
		{
			name: "High scores",
			scores: []ScoreData{
				{Score: 95, Category: "complexity"},
				{Score: 90, Category: "dead_code"},
				{Score: 85, Category: "clone"},
			},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{90, 95, "pass"},
		},
		{
			name: "Mixed scores",
			scores: []ScoreData{
				{Score: 70, Category: "complexity"},
				{Score: 40, Category: "dead_code"},
				{Score: 80, Category: "clone"},
			},
			expected: struct {
				minScore int
				maxScore int
				status   string
			}{60, 70, "average"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overall := formatter.CalculateOverallScore(tt.scores, "Test Project")

			assert.GreaterOrEqual(t, overall.Score, tt.expected.minScore)
			assert.LessOrEqual(t, overall.Score, tt.expected.maxScore)
			assert.Equal(t, tt.expected.status, overall.Status)
			assert.Equal(t, "Test Project", overall.ProjectName)
			assert.NotEmpty(t, overall.Timestamp)
			assert.Equal(t, len(tt.scores), len(overall.Breakdown))
		})
	}
}

func TestHTMLFormatter_FormatComplexityAsHTML(t *testing.T) {
	formatter := NewHTMLFormatter()

	response := &domain.ComplexityResponse{
		Summary: domain.ComplexitySummary{
			TotalFunctions:    5,
			AverageComplexity: 3.2,
			FilesAnalyzed:     3,
		},
		Functions: []domain.FunctionComplexity{
			{
				Name:     "test_function",
				FilePath: "test.py",
				Metrics: domain.ComplexityMetrics{
					Complexity:   5,
					NestingDepth: 3,
				},
				RiskLevel: domain.RiskLevelMedium,
			},
		},
	}

	html, err := formatter.FormatComplexityAsHTML(response, "Test Project")

	assert.NoError(t, err)
	assert.NotEmpty(t, html)

	// Check HTML structure
	assert.Contains(t, html, "<!DOCTYPE html>")
	assert.Contains(t, html, "<title>pyscn Code Quality Report - Test Project</title>")
	assert.Contains(t, html, "Test Project")
	assert.Contains(t, html, "Overall Score")
	assert.Contains(t, html, "Complexity Score")

	// Check CSS is included
	assert.Contains(t, html, "<style>")
	assert.Contains(t, html, "font-family:")

	// Check responsive design
	assert.Contains(t, html, "@media")

	// Check function table is present with nesting depth column
	assert.Contains(t, html, "Function Details")
	assert.Contains(t, html, "Nesting Depth")
	assert.Contains(t, html, "test_function")
}

func TestHTMLFormatter_FormatComplexityAsHTML_WithNestingDepth(t *testing.T) {
	formatter := NewHTMLFormatter()

	response := &domain.ComplexityResponse{
		Summary: domain.ComplexitySummary{
			TotalFunctions:    2,
			AverageComplexity: 5.5,
			FilesAnalyzed:     1,
		},
		Functions: []domain.FunctionComplexity{
			{
				Name:     "low_nesting_func",
				FilePath: "example.py",
				Metrics: domain.ComplexityMetrics{
					Complexity:   3,
					NestingDepth: 1,
					Nodes:        10,
					Edges:        12,
				},
				RiskLevel: domain.RiskLevelLow,
			},
			{
				Name:     "high_nesting_func",
				FilePath: "example.py",
				Metrics: domain.ComplexityMetrics{
					Complexity:   8,
					NestingDepth: 5,
					Nodes:        25,
					Edges:        35,
				},
				RiskLevel: domain.RiskLevelHigh,
			},
		},
	}

	html, err := formatter.FormatComplexityAsHTML(response, "Nesting Test")

	assert.NoError(t, err)
	assert.NotEmpty(t, html)

	// Verify function table is present
	assert.Contains(t, html, "Function Details")
	assert.Contains(t, html, "Nesting Depth")

	// Verify both functions are listed
	assert.Contains(t, html, "low_nesting_func")
	assert.Contains(t, html, "high_nesting_func")

	// Verify nesting depth values are displayed (as they appear in table cells)
	assert.Contains(t, html, ">1<")
	assert.Contains(t, html, ">5<")
}

func TestHTMLFormatter_FormatComplexityAsHTML_EmptyFunctions(t *testing.T) {
	formatter := NewHTMLFormatter()

	response := &domain.ComplexityResponse{
		Summary: domain.ComplexitySummary{
			TotalFunctions:    0,
			AverageComplexity: 0,
			FilesAnalyzed:     0,
		},
		Functions: []domain.FunctionComplexity{},
	}

	html, err := formatter.FormatComplexityAsHTML(response, "Empty Test")

	assert.NoError(t, err)
	assert.NotEmpty(t, html)

	// Function table should NOT be present when there are no functions
	assert.NotContains(t, html, "Function Details")
	assert.NotContains(t, html, "<table")

	// But the report structure should still be present
	assert.Contains(t, html, "Empty Test")
	assert.Contains(t, html, "Overall Score")
}

func TestHTMLFormatter_FormatDeadCodeAsHTML(t *testing.T) {
	formatter := NewHTMLFormatter()

	response := &domain.DeadCodeResponse{
		Summary: domain.DeadCodeSummary{
			TotalFiles:       3,
			TotalFindings:    2,
			TotalBlocks:      50,
			DeadBlocks:       5,
			OverallDeadRatio: 0.1,
		},
	}

	html, err := formatter.FormatDeadCodeAsHTML(response, "Test Project")

	assert.NoError(t, err)
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "Test Project")
	assert.Contains(t, html, "Dead_code Score")
}

func TestHTMLFormatter_FormatCloneAsHTML(t *testing.T) {
	formatter := NewHTMLFormatter()

	response := &domain.CloneResponse{
		Success: true,
		Statistics: &domain.CloneStatistics{
			LinesAnalyzed:   1000,
			TotalClonePairs: 5,
			FilesAnalyzed:   10,
		},
	}

	html, err := formatter.FormatCloneAsHTML(response, "Test Project")

	assert.NoError(t, err)
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "Test Project")
	assert.Contains(t, html, "Clone Score")
}

func TestHTMLFormatter_renderTemplate(t *testing.T) {
	formatter := NewHTMLFormatter()

	data := ComplexityHTMLData{
		OverallScore: OverallScoreData{
			Score:       85,
			Color:       "#0CCE6B",
			Status:      "pass",
			ProjectName: "Test Project",
			Timestamp:   "2024-01-01T00:00:00Z",
			Breakdown:   []ScoreData{},
		},
		Response: &domain.ComplexityResponse{
			Summary: domain.ComplexitySummary{
				TotalFunctions: 5,
				FilesAnalyzed:  3,
			},
		},
	}

	html, err := formatter.renderTemplate(data)

	assert.NoError(t, err)
	assert.NotEmpty(t, html)
	assert.Contains(t, html, "Test Project")
	assert.Contains(t, html, "85")
}

func TestHTMLFormatter_ErrorHandling(t *testing.T) {
	formatter := NewHTMLFormatter()

	// Test with nil response
	_, err := formatter.FormatComplexityAsHTML(nil, "Test")
	assert.Error(t, err)
}
