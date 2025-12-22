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

func createTestDeadCodeResponse() *domain.DeadCodeResponse {
	return &domain.DeadCodeResponse{
		Files: []domain.FileDeadCode{
			{
				FilePath: "test.py",
				Functions: []domain.FunctionDeadCode{
					{
						Name: "test_func",
						Findings: []domain.DeadCodeFinding{
							{
								Location: domain.DeadCodeLocation{
									FilePath:  "test.py",
									StartLine: 10,
									EndLine:   15,
								},
								FunctionName: "test_func",
								Reason:       "code after return",
								Severity:     domain.DeadCodeSeverityCritical,
							},
							{
								Location: domain.DeadCodeLocation{
									FilePath:  "test.py",
									StartLine: 20,
									EndLine:   22,
								},
								FunctionName: "test_func",
								Reason:       "condition always false",
								Severity:     domain.DeadCodeSeverityWarning,
							},
						},
					},
				},
			},
		},
		Summary: domain.DeadCodeSummary{
			TotalFiles:       1,
			TotalFunctions:   1,
			TotalFindings:    2,
			CriticalFindings: 1,
			WarningFindings:  1,
			InfoFindings:     0,
		},
	}
}

func createMinimalDeadCodeResponse() *domain.DeadCodeResponse {
	return &domain.DeadCodeResponse{
		Files: []domain.FileDeadCode{},
		Summary: domain.DeadCodeSummary{
			TotalFiles:    5,
			TotalFindings: 0,
		},
	}
}

func TestNewDeadCodeFormatter(t *testing.T) {
	formatter := NewDeadCodeFormatter()
	assert.NotNil(t, formatter)
}

func TestDeadCodeFormatter_Format_Text(t *testing.T) {
	tests := []struct {
		name          string
		response      *domain.DeadCodeResponse
		expectedParts []string
	}{
		{
			name:     "response with findings",
			response: createTestDeadCodeResponse(),
			expectedParts: []string{
				"Dead Code Analysis Report",
				"test.py",
				"test_func",
				"High",
			},
		},
		{
			name:     "no findings",
			response: createMinimalDeadCodeResponse(),
			expectedParts: []string{
				"Dead Code Analysis Report",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewDeadCodeFormatter()
			var buf bytes.Buffer

			result, err := formatter.Format(tt.response, domain.OutputFormatText)
			require.NoError(t, err)

			for _, part := range tt.expectedParts {
				assert.Contains(t, result, part, "expected output to contain: %s", part)
			}

			// Test Write method too
			err = formatter.Write(tt.response, domain.OutputFormatText, &buf)
			require.NoError(t, err)
			assert.Contains(t, buf.String(), tt.expectedParts[0])
		})
	}
}

func TestDeadCodeFormatter_Format_JSON(t *testing.T) {
	formatter := NewDeadCodeFormatter()
	response := createTestDeadCodeResponse()

	result, err := formatter.Format(response, domain.OutputFormatJSON)
	require.NoError(t, err)

	var decoded domain.DeadCodeResponse
	err = json.Unmarshal([]byte(result), &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.Summary.TotalFindings, decoded.Summary.TotalFindings)
	assert.Equal(t, response.Summary.CriticalFindings, decoded.Summary.CriticalFindings)
}

func TestDeadCodeFormatter_Format_YAML(t *testing.T) {
	formatter := NewDeadCodeFormatter()
	response := createTestDeadCodeResponse()

	result, err := formatter.Format(response, domain.OutputFormatYAML)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = yaml.Unmarshal([]byte(result), &decoded)
	require.NoError(t, err)

	assert.Contains(t, decoded, "files")
	assert.Contains(t, decoded, "summary")
}

func TestDeadCodeFormatter_Format_CSV(t *testing.T) {
	formatter := NewDeadCodeFormatter()
	response := createTestDeadCodeResponse()

	result, err := formatter.Format(response, domain.OutputFormatCSV)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(result), "\n")
	assert.GreaterOrEqual(t, len(lines), 2) // header + at least 1 data row

	// Check header
	assert.Contains(t, lines[0], "File")
	assert.Contains(t, lines[0], "Function")
	assert.Contains(t, lines[0], "Severity")
}

func TestDeadCodeFormatter_Format_HTML(t *testing.T) {
	formatter := NewDeadCodeFormatter()
	response := createTestDeadCodeResponse()

	result, err := formatter.Format(response, domain.OutputFormatHTML)
	require.NoError(t, err)

	assert.Contains(t, result, "<!DOCTYPE html>")
	assert.Contains(t, result, "</html>")
	assert.Contains(t, result, "Dead_code")
}

func TestDeadCodeFormatter_Format_UnsupportedFormat(t *testing.T) {
	formatter := NewDeadCodeFormatter()
	response := createTestDeadCodeResponse()

	_, err := formatter.Format(response, domain.OutputFormat("invalid"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestDeadCodeFormatter_Write(t *testing.T) {
	formatter := NewDeadCodeFormatter()
	response := createTestDeadCodeResponse()
	var buf bytes.Buffer

	err := formatter.Write(response, domain.OutputFormatJSON, &buf)
	require.NoError(t, err)

	var decoded domain.DeadCodeResponse
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.Summary.TotalFindings, decoded.Summary.TotalFindings)
}

func TestDeadCodeFormatter_FormatFinding(t *testing.T) {
	formatter := NewDeadCodeFormatter()
	finding := domain.DeadCodeFinding{
		Location: domain.DeadCodeLocation{
			FilePath:  "test.py",
			StartLine: 10,
			EndLine:   15,
		},
		FunctionName: "my_func",
		Reason:       "code after return",
		Severity:     domain.DeadCodeSeverityCritical,
	}

	result, err := formatter.FormatFinding(finding, domain.OutputFormatText)
	require.NoError(t, err)

	assert.Contains(t, result, "CRITICAL")
	assert.Contains(t, result, "10-15")
	assert.Contains(t, result, "code after return")
}
