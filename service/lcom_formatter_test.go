package service

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLCOMResponse() *domain.LCOMResponse {
	return &domain.LCOMResponse{
		Classes: []domain.ClassCohesion{
			{
				Name:      "HighLCOM",
				FilePath:  "test.py",
				StartLine: 1,
				EndLine:   20,
				Metrics: domain.LCOMMetrics{
					LCOM4:             6,
					TotalMethods:      8,
					ExcludedMethods:   2,
					InstanceVariables: 6,
					MethodGroups:      [][]string{{"a"}, {"b"}, {"c"}, {"d"}, {"e"}, {"f"}},
				},
				RiskLevel: domain.RiskLevelHigh,
			},
			{
				Name:      "LowLCOM",
				FilePath:  "test.py",
				StartLine: 25,
				EndLine:   40,
				Metrics: domain.LCOMMetrics{
					LCOM4:             1,
					TotalMethods:      3,
					ExcludedMethods:   0,
					InstanceVariables: 1,
					MethodGroups:      [][]string{{"get", "set", "__init__"}},
				},
				RiskLevel: domain.RiskLevelLow,
			},
		},
		Summary: domain.LCOMSummary{
			TotalClasses:    2,
			ClassesAnalyzed: 2,
			FilesAnalyzed:   1,
			AverageLCOM:     3.5,
			MaxLCOM:         6,
			MinLCOM:         1,
			HighRiskClasses: 1,
			LowRiskClasses:  1,
			LCOMDistribution: map[string]int{
				"1":  1,
				"6+": 1,
			},
			LeastCohesiveClasses: []domain.ClassCohesion{
				{
					Name:      "HighLCOM",
					FilePath:  "test.py",
					StartLine: 1,
					Metrics:   domain.LCOMMetrics{LCOM4: 6},
					RiskLevel: domain.RiskLevelHigh,
				},
			},
		},
		GeneratedAt: "2025-01-01T00:00:00Z",
		Version:     "1.0.0",
	}
}

func TestLCOMFormatter_Format(t *testing.T) {
	formatter := NewLCOMFormatter()
	response := newTestLCOMResponse()

	tests := []struct {
		name   string
		format domain.OutputFormat
		check  func(t *testing.T, output string)
	}{
		{
			name:   "text format",
			format: domain.OutputFormatText,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "LCOM4")
				assert.Contains(t, output, "HighLCOM")
				assert.Contains(t, output, "LowLCOM")
				assert.Contains(t, output, "Total Classes")
			},
		},
		{
			name:   "JSON format",
			format: domain.OutputFormatJSON,
			check: func(t *testing.T, output string) {
				var result domain.LCOMResponse
				err := json.Unmarshal([]byte(output), &result)
				require.NoError(t, err)
				assert.Equal(t, 2, len(result.Classes))
				assert.Equal(t, 2, result.Summary.TotalClasses)
			},
		},
		{
			name:   "YAML format",
			format: domain.OutputFormatYAML,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "classes:")
				assert.Contains(t, output, "HighLCOM")
			},
		},
		{
			name:   "CSV format",
			format: domain.OutputFormatCSV,
			check: func(t *testing.T, output string) {
				lines := strings.Split(strings.TrimSpace(output), "\n")
				assert.GreaterOrEqual(t, len(lines), 3) // header + 2 data rows
				assert.Contains(t, lines[0], "ClassName")
				assert.Contains(t, lines[0], "LCOM4")
			},
		},
		{
			name:   "HTML format",
			format: domain.OutputFormatHTML,
			check: func(t *testing.T, output string) {
				assert.Contains(t, output, "<!DOCTYPE html>")
				assert.Contains(t, output, "LCOM4")
				assert.Contains(t, output, "HighLCOM")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := formatter.Format(response, tt.format)
			require.NoError(t, err)
			assert.NotEmpty(t, output)
			tt.check(t, output)
		})
	}
}

func TestLCOMFormatter_Write(t *testing.T) {
	formatter := NewLCOMFormatter()
	response := newTestLCOMResponse()

	var buf bytes.Buffer
	err := formatter.Write(response, domain.OutputFormatJSON, &buf)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())

	var result domain.LCOMResponse
	err = json.Unmarshal(buf.Bytes(), &result)
	require.NoError(t, err)
	assert.Equal(t, 2, len(result.Classes))
}

func TestLCOMFormatter_UnsupportedFormat(t *testing.T) {
	formatter := NewLCOMFormatter()
	response := newTestLCOMResponse()

	_, err := formatter.Format(response, domain.OutputFormat("invalid"))
	assert.Error(t, err)
}
