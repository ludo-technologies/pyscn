package service

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// createTestAnalyzeResponse creates a test AnalyzeResponse for testing
func createTestAnalyzeResponse() *domain.AnalyzeResponse {
	return &domain.AnalyzeResponse{
		GeneratedAt: time.Now(),
		Duration:    1500,
		Summary: domain.AnalyzeSummary{
			HealthScore:           85,
			Grade:                 "B",
			TotalFiles:            10,
			AnalyzedFiles:         10,
			TotalFunctions:        25,
			AverageComplexity:     5.5,
			HighComplexityCount:   2,
			DeadCodeCount:         3,
			CriticalDeadCode:      1,
			WarningDeadCode:       2,
			InfoDeadCode:          0,
			TotalClones:           5,
			ClonePairs:            3,
			CloneGroups:           2,
			CodeDuplication:       8.5,
			CBOClasses:            8,
			HighCouplingClasses:   1,
			MediumCouplingClasses: 2,
			AverageCoupling:       3.2,
			ComplexityEnabled:     true,
			DeadCodeEnabled:       true,
			CloneEnabled:          true,
			CBOEnabled:            true,
		},
		Complexity: &domain.ComplexityResponse{
			Functions: []domain.FunctionComplexity{
				{
					Name:      "complex_func",
					FilePath:  "test.py",
					Metrics:   domain.ComplexityMetrics{Complexity: 15},
					RiskLevel: domain.RiskLevelHigh,
				},
			},
			Summary: domain.ComplexitySummary{
				TotalFunctions:    25,
				AverageComplexity: 5.5,
				MaxComplexity:     15,
			},
		},
		DeadCode: &domain.DeadCodeResponse{
			Summary: domain.DeadCodeSummary{
				TotalFindings:    3,
				CriticalFindings: 1,
				WarningFindings:  2,
				InfoFindings:     0,
			},
		},
		Clone: &domain.CloneResponse{
			Statistics: &domain.CloneStatistics{
				TotalClones:      5,
				TotalClonePairs:  3,
				TotalCloneGroups: 2,
			},
		},
		CBO: &domain.CBOResponse{
			Summary: domain.CBOSummary{
				TotalClasses:      8,
				HighRiskClasses:   1,
				MediumRiskClasses: 2,
				AverageCBO:        3.2,
			},
		},
	}
}

func createMinimalAnalyzeResponse() *domain.AnalyzeResponse {
	return &domain.AnalyzeResponse{
		GeneratedAt: time.Now(),
		Duration:    500,
		Summary: domain.AnalyzeSummary{
			HealthScore:   100,
			Grade:         "A",
			TotalFiles:    5,
			AnalyzedFiles: 5,
		},
	}
}

func TestNewAnalyzeFormatter(t *testing.T) {
	formatter := NewAnalyzeFormatter()

	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.complexityFormatter)
	assert.NotNil(t, formatter.deadCodeFormatter)
	assert.NotNil(t, formatter.cloneFormatter)
}

func TestAnalyzeFormatter_Write_Text(t *testing.T) {
	tests := []struct {
		name          string
		response      *domain.AnalyzeResponse
		expectedParts []string
		notExpected   []string
	}{
		{
			name:     "full response with all analyses",
			response: createTestAnalyzeResponse(),
			expectedParts: []string{
				"Comprehensive Analysis Report",
				"Health Score",
				"85/100",
				"COMPLEXITY ANALYSIS",
				"DEAD CODE DETECTION",
				"CLONE DETECTION",
				"DEPENDENCY ANALYSIS",
				"RECOMMENDATIONS",
			},
		},
		{
			name:     "minimal response no issues",
			response: createMinimalAnalyzeResponse(),
			expectedParts: []string{
				"Comprehensive Analysis Report",
				"Health Score",
				"100/100",
				"No major issues detected",
			},
			notExpected: []string{
				"COMPLEXITY ANALYSIS",
				"DEAD CODE DETECTION",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewAnalyzeFormatter()
			var buf bytes.Buffer

			err := formatter.Write(tt.response, domain.OutputFormatText, &buf)
			require.NoError(t, err)

			output := buf.String()
			for _, part := range tt.expectedParts {
				assert.Contains(t, output, part, "expected output to contain: %s", part)
			}
			for _, part := range tt.notExpected {
				assert.NotContains(t, output, part, "expected output NOT to contain: %s", part)
			}
		})
	}
}

func TestAnalyzeFormatter_Write_JSON(t *testing.T) {
	formatter := NewAnalyzeFormatter()
	response := createTestAnalyzeResponse()
	var buf bytes.Buffer

	err := formatter.Write(response, domain.OutputFormatJSON, &buf)
	require.NoError(t, err)

	// Verify valid JSON
	var decoded domain.AnalyzeResponse
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.Summary.HealthScore, decoded.Summary.HealthScore)
	assert.Equal(t, response.Summary.Grade, decoded.Summary.Grade)
	assert.Equal(t, response.Summary.TotalFiles, decoded.Summary.TotalFiles)
}

func TestAnalyzeFormatter_Write_YAML(t *testing.T) {
	formatter := NewAnalyzeFormatter()
	response := createTestAnalyzeResponse()
	var buf bytes.Buffer

	err := formatter.Write(response, domain.OutputFormatYAML, &buf)
	require.NoError(t, err)

	// Verify valid YAML
	var decoded map[string]interface{}
	err = yaml.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Contains(t, decoded, "summary")
	assert.Contains(t, decoded, "generated_at")
}

func TestAnalyzeFormatter_Write_CSV(t *testing.T) {
	formatter := NewAnalyzeFormatter()
	response := createTestAnalyzeResponse()
	var buf bytes.Buffer

	err := formatter.Write(response, domain.OutputFormatCSV, &buf)
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check header
	assert.Equal(t, "Metric,Value", lines[0])

	// Check some expected rows
	assert.Contains(t, output, "Health Score,85")
	assert.Contains(t, output, "Grade,B")
	assert.Contains(t, output, "Total Files,10")
	assert.Contains(t, output, "Analyzed Files,10")
}

func TestAnalyzeFormatter_Write_HTML(t *testing.T) {
	formatter := NewAnalyzeFormatter()
	response := createTestAnalyzeResponse()
	var buf bytes.Buffer

	err := formatter.Write(response, domain.OutputFormatHTML, &buf)
	require.NoError(t, err)

	output := buf.String()

	// Verify HTML structure
	assert.Contains(t, output, "<!DOCTYPE html>")
	assert.Contains(t, output, "<html")
	assert.Contains(t, output, "</html>")
	assert.Contains(t, output, "pyscn Analysis Report")

	// Verify tabs are present for enabled analyses
	assert.Contains(t, output, "Complexity")
	assert.Contains(t, output, "Dead Code")
	assert.Contains(t, output, "Clone Detection")
	assert.Contains(t, output, "Class Coupling")
}

func TestAnalyzeFormatter_Write_UnsupportedFormat(t *testing.T) {
	formatter := NewAnalyzeFormatter()
	response := createTestAnalyzeResponse()
	var buf bytes.Buffer

	err := formatter.Write(response, domain.OutputFormat("invalid"), &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestAnalyzeFormatter_WriteText_Recommendations(t *testing.T) {
	tests := []struct {
		name             string
		modifyResponse   func(*domain.AnalyzeResponse)
		expectedContains []string
	}{
		{
			name: "high complexity recommendation",
			modifyResponse: func(r *domain.AnalyzeResponse) {
				r.Summary.HighComplexityCount = 5
				r.Summary.ComplexityEnabled = true
			},
			expectedContains: []string{"Refactor 5 high-complexity functions"},
		},
		{
			name: "dead code recommendation",
			modifyResponse: func(r *domain.AnalyzeResponse) {
				r.Summary.DeadCodeCount = 10
				r.Summary.DeadCodeEnabled = true
			},
			expectedContains: []string{"Remove 10 dead code segments"},
		},
		{
			name: "high duplication recommendation",
			modifyResponse: func(r *domain.AnalyzeResponse) {
				r.Summary.CodeDuplication = 15.5
				r.Summary.CloneEnabled = true
			},
			expectedContains: []string{"Reduce code duplication", "15.5%"},
		},
		{
			name: "high coupling recommendation",
			modifyResponse: func(r *domain.AnalyzeResponse) {
				r.Summary.HighCouplingClasses = 3
				r.Summary.CBOEnabled = true
			},
			expectedContains: []string{"Reduce coupling in 3 high-dependency classes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := createMinimalAnalyzeResponse()
			tt.modifyResponse(response)

			formatter := NewAnalyzeFormatter()
			var buf bytes.Buffer

			err := formatter.Write(response, domain.OutputFormatText, &buf)
			require.NoError(t, err)

			output := buf.String()
			for _, expected := range tt.expectedContains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestAnalyzeFormatter_WriteHTML_ScoreQuality(t *testing.T) {
	tests := []struct {
		name          string
		healthScore   int
		expectedClass string
	}{
		{"excellent score", 95, "grade-a"},
		{"good score", 80, "grade-b"},
		{"fair score", 65, "grade-c"},
		{"poor score", 50, "grade-d"},
		{"failing score", 30, "grade-f"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := createMinimalAnalyzeResponse()
			response.Summary.HealthScore = tt.healthScore
			switch {
			case tt.healthScore >= 90:
				response.Summary.Grade = "A"
			case tt.healthScore >= 75:
				response.Summary.Grade = "B"
			case tt.healthScore >= 60:
				response.Summary.Grade = "C"
			case tt.healthScore >= 45:
				response.Summary.Grade = "D"
			default:
				response.Summary.Grade = "F"
			}

			formatter := NewAnalyzeFormatter()
			var buf bytes.Buffer

			err := formatter.Write(response, domain.OutputFormatHTML, &buf)
			require.NoError(t, err)

			assert.Contains(t, buf.String(), tt.expectedClass)
		})
	}
}
