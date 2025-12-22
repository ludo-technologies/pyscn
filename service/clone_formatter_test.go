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

func createTestCloneResponse() *domain.CloneResponse {
	return &domain.CloneResponse{
		Success:  true,
		Duration: 1500,
		Statistics: &domain.CloneStatistics{
			FilesAnalyzed:     10,
			LinesAnalyzed:     1000,
			TotalClones:       5,
			TotalClonePairs:   3,
			TotalCloneGroups:  2,
			AverageSimilarity: 0.85,
			ClonesByType: map[string]int{
				"Type-1": 1,
				"Type-2": 2,
			},
		},
		ClonePairs: []*domain.ClonePair{
			{
				Clone1: &domain.Clone{
					Location: &domain.CloneLocation{
						FilePath:  "file1.py",
						StartLine: 10,
						EndLine:   20,
					},
					LineCount: 11,
				},
				Clone2: &domain.Clone{
					Location: &domain.CloneLocation{
						FilePath:  "file2.py",
						StartLine: 30,
						EndLine:   40,
					},
					LineCount: 11,
				},
				Similarity: 0.95,
				Type:       domain.Type1Clone,
			},
		},
		CloneGroups: []*domain.CloneGroup{
			{
				ID:         1,
				Type:       1,
				Similarity: 0.92,
				Clones: []*domain.Clone{
					{
						Location: &domain.CloneLocation{
							FilePath:  "file1.py",
							StartLine: 10,
							EndLine:   20,
						},
						LineCount: 11,
					},
					{
						Location: &domain.CloneLocation{
							FilePath:  "file2.py",
							StartLine: 30,
							EndLine:   40,
						},
						LineCount: 11,
					},
				},
			},
		},
		Request: &domain.CloneRequest{
			GroupClones: true,
		},
	}
}

func createMinimalCloneResponse() *domain.CloneResponse {
	return &domain.CloneResponse{
		Success:    true,
		Duration:   500,
		Statistics: &domain.CloneStatistics{},
		ClonePairs: []*domain.ClonePair{},
	}
}

func createFailedCloneResponse() *domain.CloneResponse {
	return &domain.CloneResponse{
		Success: false,
		Error:   "analysis failed",
	}
}

func TestNewCloneOutputFormatter(t *testing.T) {
	formatter := NewCloneOutputFormatter()
	assert.NotNil(t, formatter)
}

func TestCloneOutputFormatter_FormatCloneResponse_Text(t *testing.T) {
	tests := []struct {
		name          string
		response      *domain.CloneResponse
		expectedParts []string
	}{
		{
			name:     "full response with clones",
			response: createTestCloneResponse(),
			expectedParts: []string{
				"Clone Detection Analysis Report",
				"Files Analyzed",
				"Clone Pairs",
				"Clone Groups",
				"CLONE TYPES",
				"Type-1",
			},
		},
		{
			name:     "minimal response no clones",
			response: createMinimalCloneResponse(),
			expectedParts: []string{
				"Clone Detection Analysis Report",
				"No clones detected",
			},
		},
		{
			name:     "failed response",
			response: createFailedCloneResponse(),
			expectedParts: []string{
				"Clone detection failed",
				"analysis failed",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewCloneOutputFormatter()
			var buf bytes.Buffer

			err := formatter.FormatCloneResponse(tt.response, domain.OutputFormatText, &buf)
			require.NoError(t, err)

			output := buf.String()
			for _, part := range tt.expectedParts {
				assert.Contains(t, output, part, "expected output to contain: %s", part)
			}
		})
	}
}

func TestCloneOutputFormatter_FormatCloneResponse_JSON(t *testing.T) {
	formatter := NewCloneOutputFormatter()
	response := createTestCloneResponse()
	var buf bytes.Buffer

	err := formatter.FormatCloneResponse(response, domain.OutputFormatJSON, &buf)
	require.NoError(t, err)

	var decoded domain.CloneResponse
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Equal(t, response.Success, decoded.Success)
	assert.Equal(t, response.Statistics.TotalClonePairs, decoded.Statistics.TotalClonePairs)
}

func TestCloneOutputFormatter_FormatCloneResponse_YAML(t *testing.T) {
	formatter := NewCloneOutputFormatter()
	response := createTestCloneResponse()
	var buf bytes.Buffer

	err := formatter.FormatCloneResponse(response, domain.OutputFormatYAML, &buf)
	require.NoError(t, err)

	var decoded map[string]interface{}
	err = yaml.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Contains(t, decoded, "success")
	assert.Contains(t, decoded, "statistics")
}

func TestCloneOutputFormatter_FormatCloneResponse_CSV(t *testing.T) {
	formatter := NewCloneOutputFormatter()
	response := createTestCloneResponse()
	var buf bytes.Buffer

	err := formatter.FormatCloneResponse(response, domain.OutputFormatCSV, &buf)
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Check header exists
	assert.GreaterOrEqual(t, len(lines), 1)
	// Check contains expected data
	assert.Contains(t, output, "file1.py")
	assert.Contains(t, output, "file2.py")
}

func TestCloneOutputFormatter_FormatCloneResponse_HTML(t *testing.T) {
	formatter := NewCloneOutputFormatter()
	response := createTestCloneResponse()
	var buf bytes.Buffer

	err := formatter.FormatCloneResponse(response, domain.OutputFormatHTML, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<!DOCTYPE html>")
	assert.Contains(t, output, "</html>")
}

func TestCloneOutputFormatter_FormatCloneResponse_UnsupportedFormat(t *testing.T) {
	formatter := NewCloneOutputFormatter()
	response := createTestCloneResponse()
	var buf bytes.Buffer

	err := formatter.FormatCloneResponse(response, domain.OutputFormat("invalid"), &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestCloneOutputFormatter_FormatCloneStatistics_Text(t *testing.T) {
	formatter := NewCloneOutputFormatter()
	stats := &domain.CloneStatistics{
		FilesAnalyzed:     10,
		LinesAnalyzed:     1000,
		TotalClones:       5,
		TotalClonePairs:   3,
		TotalCloneGroups:  2,
		AverageSimilarity: 0.85,
	}
	var buf bytes.Buffer

	err := formatter.FormatCloneStatistics(stats, domain.OutputFormatText, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Files analyzed")
	assert.Contains(t, output, "10")
}

func TestCloneOutputFormatter_FormatCloneStatistics_JSON(t *testing.T) {
	formatter := NewCloneOutputFormatter()
	stats := &domain.CloneStatistics{
		FilesAnalyzed:     10,
		TotalClonePairs:   3,
		AverageSimilarity: 0.85,
	}
	var buf bytes.Buffer

	err := formatter.FormatCloneStatistics(stats, domain.OutputFormatJSON, &buf)
	require.NoError(t, err)

	var decoded domain.CloneStatistics
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Equal(t, stats.FilesAnalyzed, decoded.FilesAnalyzed)
}

func TestCloneOutputFormatter_FormatCloneStatistics_CSV(t *testing.T) {
	formatter := NewCloneOutputFormatter()
	stats := &domain.CloneStatistics{
		FilesAnalyzed: 10,
	}
	var buf bytes.Buffer

	err := formatter.FormatCloneStatistics(stats, domain.OutputFormatCSV, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "10")
}

func TestCloneOutputFormatter_FormatCloneStatistics_HTML(t *testing.T) {
	formatter := NewCloneOutputFormatter()
	stats := &domain.CloneStatistics{
		FilesAnalyzed: 10,
	}
	var buf bytes.Buffer

	err := formatter.FormatCloneStatistics(stats, domain.OutputFormatHTML, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "<")
}

func TestCloneOutputFormatter_FormatCloneStatistics_UnsupportedFormat(t *testing.T) {
	formatter := NewCloneOutputFormatter()
	stats := &domain.CloneStatistics{}
	var buf bytes.Buffer

	err := formatter.FormatCloneStatistics(stats, domain.OutputFormat("invalid"), &buf)
	assert.Error(t, err)
}
