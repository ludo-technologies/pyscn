package service

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func createTestCBOResponse() *domain.CBOResponse {
	return &domain.CBOResponse{
		Classes: []domain.ClassCoupling{
			{
				Name:      "MyClass",
				FilePath:  "test.py",
				StartLine: 10,
				Metrics: domain.CBOMetrics{
					CouplingCount:    5,
					DependentClasses: []string{"ClassA", "ClassB", "ClassC"},
				},
				RiskLevel: domain.RiskLevelHigh,
			},
			{
				Name:      "SimpleClass",
				FilePath:  "simple.py",
				StartLine: 5,
				Metrics: domain.CBOMetrics{
					CouplingCount:    2,
					DependentClasses: []string{"Helper"},
				},
				RiskLevel: domain.RiskLevelLow,
			},
		},
		Summary: domain.CBOSummary{
			TotalClasses:      2,
			AverageCBO:        3.5,
			MaxCBO:            5,
			LowRiskClasses:    1,
			MediumRiskClasses: 0,
			HighRiskClasses:   1,
			FilesAnalyzed:     2,
		},
	}
}

func createMinimalCBOResponse() *domain.CBOResponse {
	return &domain.CBOResponse{
		Classes: []domain.ClassCoupling{},
		Summary: domain.CBOSummary{
			TotalClasses:  0,
			FilesAnalyzed: 5,
		},
	}
}

func TestNewCBOFormatter(t *testing.T) {
	formatter := NewCBOFormatter()
	assert.NotNil(t, formatter)
}

func TestCBOFormatter_Format_Text(t *testing.T) {
	tests := []struct {
		name          string
		response      *domain.CBOResponse
		expectedParts []string
	}{
		{
			name:     "response with classes",
			response: createTestCBOResponse(),
			expectedParts: []string{
				"CBO (Coupling Between Objects) Analysis Report",
				"MyClass",
				"SimpleClass",
				"Total Classes",
			},
		},
		{
			name:     "no classes",
			response: createMinimalCBOResponse(),
			expectedParts: []string{
				"CBO (Coupling Between Objects) Analysis Report",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewCBOFormatter()

			result, err := formatter.Format(tt.response, domain.OutputFormatText)
			require.NoError(t, err)

			for _, part := range tt.expectedParts {
				assert.Contains(t, result, part, "expected output to contain: %s", part)
			}
		})
	}
}

func TestCBOFormatter_Format_JSON(t *testing.T) {
	formatter := NewCBOFormatter()
	response := createTestCBOResponse()

	result, err := formatter.Format(response, domain.OutputFormatJSON)
	require.NoError(t, err)

	var decoded domain.CBOResponse
	err = json.Unmarshal([]byte(result), &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.Summary.TotalClasses, decoded.Summary.TotalClasses)
	assert.Equal(t, response.Summary.AverageCBO, decoded.Summary.AverageCBO)
	assert.Len(t, decoded.Classes, 2)
}

func TestCBOFormatter_Format_YAML(t *testing.T) {
	formatter := NewCBOFormatter()
	response := createTestCBOResponse()

	result, err := formatter.Format(response, domain.OutputFormatYAML)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &decoded)
	require.NoError(t, err)

	assert.Contains(t, decoded, "classes")
	assert.Contains(t, decoded, "summary")
}

func TestCBOFormatter_Format_CSV(t *testing.T) {
	formatter := NewCBOFormatter()
	response := createTestCBOResponse()

	result, err := formatter.Format(response, domain.OutputFormatCSV)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	assert.GreaterOrEqual(t, len(lines), 2) // header + at least 1 data row

	// Check header
	assert.Contains(t, lines[0], "Class")
	assert.Contains(t, lines[0], "File")
	assert.Contains(t, lines[0], "CBO")
}

func TestCBOFormatter_Format_HTML(t *testing.T) {
	formatter := NewCBOFormatter()
	response := createTestCBOResponse()

	result, err := formatter.Format(response, domain.OutputFormatHTML)
	require.NoError(t, err)

	assert.Contains(t, result, "<!DOCTYPE html>")
	assert.Contains(t, result, "</html>")
	assert.Contains(t, result, "CBO")
}

func TestCBOFormatter_Format_UnsupportedFormat(t *testing.T) {
	formatter := NewCBOFormatter()
	response := createTestCBOResponse()

	_, err := formatter.Format(response, domain.OutputFormat("invalid"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestCBOFormatter_Write(t *testing.T) {
	formatter := NewCBOFormatter()
	response := createTestCBOResponse()
	var buf bytes.Buffer

	err := formatter.Write(response, domain.OutputFormatJSON, &buf)
	require.NoError(t, err)

	var decoded domain.CBOResponse
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.Summary.TotalClasses, decoded.Summary.TotalClasses)
}

func TestCBOFormatter_WriteClassDetails(t *testing.T) {
	formatter := NewCBOFormatter()
	response := createTestCBOResponse()
	var buf bytes.Buffer

	err := formatter.Write(response, domain.OutputFormatText, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "MyClass")
	assert.Contains(t, output, "SimpleClass")
}
