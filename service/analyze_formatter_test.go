package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
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
			},
		},
		{
			name:     "minimal response no issues",
			response: createMinimalAnalyzeResponse(),
			expectedParts: []string{
				"Comprehensive Analysis Report",
				"Health Score",
				"100/100",
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

func TestAnalyzeFormatter_Write_TextIncludesModuleQuality(t *testing.T) {
	response := createMinimalAnalyzeResponse()
	response.ModuleQuality = []domain.ModuleQualityMetrics{
		{
			ModuleName:    "pkg.hotspot",
			FilePath:      "pkg/hotspot.py",
			FunctionCount: 4,
			ModuleComplexityMetrics: domain.ModuleComplexityMetrics{
				AnalyzedFunctionCount:      2,
				AverageComplexity:          6.5,
				AverageCognitiveComplexity: 8,
				MaxComplexity:              9,
				HighRiskFunctionCount:      1,
			},
			ModuleDeadCodeMetrics: domain.ModuleDeadCodeMetrics{
				DeadCodeFindingCount: 2,
				DeadCodeBlockCount:   3,
			},
		},
	}

	var output bytes.Buffer
	require.NoError(t, NewAnalyzeFormatter().Write(response, domain.OutputFormatText, &output))

	assert.Contains(t, output.String(), "MODULE QUALITY HOTSPOTS")
	assert.Contains(t, output.String(), "pkg.hotspot (pkg/hotspot.py)")
	assert.Contains(t, output.String(), "Functions: 4 total / 2 analyzed")
	assert.Contains(t, output.String(), "Complexity: avg 6.50, max 9, high-risk 1, handlers 0")
	assert.Contains(t, output.String(), "Cognitive: avg 8.00")
	assert.Contains(t, output.String(), "Dead code: 2 findings, 3 blocks")
}

func TestAnalyzeFormatter_Write_TextIncludesDirectoryQuality(t *testing.T) {
	response := createMinimalAnalyzeResponse()
	response.DirectoryQuality = []domain.DirectoryQualityMetrics{
		{
			DirectoryPath:     "pkg",
			ModuleCount:       3,
			DirectModuleCount: 2,
			LinesOfCode:       180,
			FunctionCount:     7,
			ModuleComplexityMetrics: domain.ModuleComplexityMetrics{
				AnalyzedFunctionCount:      5,
				AverageComplexity:          6.5,
				AverageCognitiveComplexity: 8,
				MaxComplexity:              11,
				HighRiskFunctionCount:      2,
				ExceptionHandlerCount:      3,
			},
			ModuleDeadCodeMetrics: domain.ModuleDeadCodeMetrics{
				DeadCodeFindingCount: 4,
				DeadCodeBlockCount:   6,
			},
		},
	}

	var output bytes.Buffer
	require.NoError(t, NewAnalyzeFormatter().Write(response, domain.OutputFormatText, &output))

	assert.Contains(t, output.String(), "DIRECTORY QUALITY HOTSPOTS")
	assert.Contains(t, output.String(), "pkg")
	assert.Contains(t, output.String(), "Modules: 3 recursive / 2 direct")
	assert.Contains(t, output.String(), "Lines of code: 180")
	assert.Contains(t, output.String(), "Functions: 7 total / 5 analyzed")
	assert.Contains(t, output.String(), "Complexity: avg 6.50, max 11, high-risk 2, handlers 3")
	assert.Contains(t, output.String(), "Cognitive: avg 8.00")
	assert.Contains(t, output.String(), "Dead code: 4 findings, 6 blocks")
}

func TestAnalyzeFormatter_Write_JSON(t *testing.T) {
	formatter := NewAnalyzeFormatter()
	response := createTestAnalyzeResponse()
	response.Complexity.Request = &domain.ComplexityRequest{
		ShowDetails: domain.BoolPtr(false),
		Recursive:   domain.BoolPtr(true),
	}
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
	require.NotNil(t, decoded.Complexity)
	require.NotNil(t, decoded.Complexity.Request)
	if assert.NotNil(t, decoded.Complexity.Request.ShowDetails) {
		assert.False(t, *decoded.Complexity.Request.ShowDetails)
	}
	if assert.NotNil(t, decoded.Complexity.Request.Recursive) {
		assert.True(t, *decoded.Complexity.Request.Recursive)
	}
}

func TestAnalyzeFormatter_Write_JSON_IncludesCommunityAnalysis(t *testing.T) {
	formatter := NewAnalyzeFormatter()
	response := createTestAnalyzeResponse()
	response.Summary.CommunitiesEnabled = true
	response.Communities = &domain.CommunityAnalysisResult{
		Algorithm:        "leiden",
		Scope:            "module",
		TotalCommunities: 2,
		Modularity:       0.42,
		Communities: []domain.CommunityMetrics{
			{ID: "community_1", Modules: []string{"mod.a"}, Size: 1},
			{ID: "community_2", Modules: []string{"mod.b"}, Size: 1},
		},
		BridgeModules: []domain.BridgeModule{
			{
				Module:              "bridge",
				Community:           "community_1",
				CrossCommunityEdges: 1,
				TargetCommunities:   []string{"community_2"},
			},
		},
	}

	var buf bytes.Buffer
	require.NoError(t, formatter.Write(response, domain.OutputFormatJSON, &buf))

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	require.Contains(t, decoded, "community_analysis")

	communityAnalysis := decoded["community_analysis"].(map[string]any)
	assert.Equal(t, "leiden", communityAnalysis["algorithm"])
	assert.Equal(t, float64(2), communityAnalysis["total_communities"])
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

func TestAnalyzeFormatter_Write_SerializesModuleQuality(t *testing.T) {
	response := createMinimalAnalyzeResponse()
	response.ModuleQuality = []domain.ModuleQualityMetrics{
		{
			ModuleName: "pkg.hotspot",
			FilePath:   "pkg/hotspot.py",
			ModuleComplexityMetrics: domain.ModuleComplexityMetrics{
				AnalyzedFunctionCount:      2,
				AverageComplexity:          6.5,
				AverageCognitiveComplexity: 8,
				MaxComplexity:              9,
				HighRiskFunctionCount:      1,
				ExceptionHandlerCount:      3,
			},
			ModuleDeadCodeMetrics: domain.ModuleDeadCodeMetrics{
				DeadCodeFindingCount: 2,
				DeadCodeBlockCount:   4,
			},
		},
	}

	for _, format := range []domain.OutputFormat{domain.OutputFormatJSON, domain.OutputFormatYAML} {
		t.Run(string(format), func(t *testing.T) {
			var output bytes.Buffer
			require.NoError(t, NewAnalyzeFormatter().Write(response, format, &output))

			var decoded domain.AnalyzeResponse
			var contract map[string]any
			switch format {
			case domain.OutputFormatJSON:
				require.NoError(t, json.Unmarshal(output.Bytes(), &decoded))
				require.NoError(t, json.Unmarshal(output.Bytes(), &contract))
			case domain.OutputFormatYAML:
				require.NoError(t, yaml.Unmarshal(output.Bytes(), &decoded))
				require.NoError(t, yaml.Unmarshal(output.Bytes(), &contract))
			default:
				t.Fatalf("unsupported test format %q", format)
			}
			assert.Equal(t, response.ModuleQuality, decoded.ModuleQuality)

			qualityEntries, ok := contract["module_quality"].([]any)
			require.True(t, ok)
			require.Len(t, qualityEntries, 1)
			quality, ok := qualityEntries[0].(map[string]any)
			require.True(t, ok)
			require.Len(t, quality, 12)
			assert.Equal(t, "pkg.hotspot", quality["module_name"])
			assert.Equal(t, "pkg/hotspot.py", quality["file_path"])
			assert.EqualValues(t, 0, quality["lines_of_code"])
			assert.EqualValues(t, 0, quality["function_count"])
			assert.EqualValues(t, 2, quality["analyzed_function_count"])
			assert.EqualValues(t, 6.5, quality["average_complexity"])
			assert.EqualValues(t, 8, quality["average_cognitive_complexity"])
			assert.EqualValues(t, 9, quality["max_complexity"])
			assert.EqualValues(t, 1, quality["high_risk_function_count"])
			assert.EqualValues(t, 3, quality["exception_handler_count"])
			assert.EqualValues(t, 2, quality["dead_code_finding_count"])
			assert.EqualValues(t, 4, quality["dead_code_block_count"])
		})
	}
}

func TestAnalyzeFormatter_Write_SerializesDirectoryQuality(t *testing.T) {
	response := createMinimalAnalyzeResponse()
	response.DirectoryQuality = []domain.DirectoryQualityMetrics{
		{
			DirectoryPath:     "pkg",
			ModuleCount:       3,
			DirectModuleCount: 2,
			LinesOfCode:       180,
			FunctionCount:     7,
			ModuleComplexityMetrics: domain.ModuleComplexityMetrics{
				AnalyzedFunctionCount:      5,
				AverageComplexity:          6.5,
				AverageCognitiveComplexity: 8,
				MaxComplexity:              11,
				HighRiskFunctionCount:      2,
				ExceptionHandlerCount:      3,
			},
			ModuleDeadCodeMetrics: domain.ModuleDeadCodeMetrics{
				DeadCodeFindingCount: 4,
				DeadCodeBlockCount:   6,
			},
		},
	}

	for _, format := range []domain.OutputFormat{domain.OutputFormatJSON, domain.OutputFormatYAML} {
		t.Run(string(format), func(t *testing.T) {
			var output bytes.Buffer
			require.NoError(t, NewAnalyzeFormatter().Write(response, format, &output))

			var decoded domain.AnalyzeResponse
			var contract map[string]any
			switch format {
			case domain.OutputFormatJSON:
				require.NoError(t, json.Unmarshal(output.Bytes(), &decoded))
				require.NoError(t, json.Unmarshal(output.Bytes(), &contract))
			case domain.OutputFormatYAML:
				require.NoError(t, yaml.Unmarshal(output.Bytes(), &decoded))
				require.NoError(t, yaml.Unmarshal(output.Bytes(), &contract))
			default:
				t.Fatalf("unsupported test format %q", format)
			}
			assert.Equal(t, response.DirectoryQuality, decoded.DirectoryQuality)

			qualityEntries, ok := contract["directory_quality"].([]any)
			require.True(t, ok)
			require.Len(t, qualityEntries, 1)
			quality, ok := qualityEntries[0].(map[string]any)
			require.True(t, ok)
			require.Len(t, quality, 13)
			assert.Equal(t, "pkg", quality["directory_path"])
			assert.EqualValues(t, 3, quality["module_count"])
			assert.EqualValues(t, 2, quality["direct_module_count"])
			assert.EqualValues(t, 180, quality["lines_of_code"])
			assert.EqualValues(t, 7, quality["function_count"])
			assert.EqualValues(t, 5, quality["analyzed_function_count"])
			assert.EqualValues(t, 6.5, quality["average_complexity"])
			assert.EqualValues(t, 8, quality["average_cognitive_complexity"])
			assert.EqualValues(t, 11, quality["max_complexity"])
			assert.EqualValues(t, 2, quality["high_risk_function_count"])
			assert.EqualValues(t, 3, quality["exception_handler_count"])
			assert.EqualValues(t, 4, quality["dead_code_finding_count"])
			assert.EqualValues(t, 6, quality["dead_code_block_count"])
		})
	}
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

func TestAnalyzeFormatter_Write_CSVIncludesModuleQuality(t *testing.T) {
	response := createMinimalAnalyzeResponse()
	response.ModuleQuality = []domain.ModuleQualityMetrics{
		{
			ModuleName:    "pkg.hotspot",
			FilePath:      "pkg/hot,spot.py",
			LinesOfCode:   120,
			FunctionCount: 4,
			ModuleComplexityMetrics: domain.ModuleComplexityMetrics{
				AnalyzedFunctionCount:      2,
				AverageComplexity:          6.5,
				AverageCognitiveComplexity: 8,
				MaxComplexity:              9,
				HighRiskFunctionCount:      1,
				ExceptionHandlerCount:      3,
			},
			ModuleDeadCodeMetrics: domain.ModuleDeadCodeMetrics{
				DeadCodeFindingCount: 2,
				DeadCodeBlockCount:   3,
			},
		},
	}

	var output bytes.Buffer
	require.NoError(t, NewAnalyzeFormatter().Write(response, domain.OutputFormatCSV, &output))

	records, err := csv.NewReader(strings.NewReader(output.String())).ReadAll()
	require.NoError(t, err)

	metrics := make(map[string]string, len(records))
	for _, record := range records[1:] {
		require.Len(t, record, 2)
		metrics[record[0]] = record[1]
	}

	assert.Equal(t, "1", metrics["Module Quality Count"])
	assert.Equal(t, "pkg.hotspot", metrics["Module 1 Name"])
	assert.Equal(t, "pkg/hot,spot.py", metrics["Module 1 File Path"])
	assert.Equal(t, "120", metrics["Module 1 Lines of Code"])
	assert.Equal(t, "4", metrics["Module 1 Function Count"])
	assert.Equal(t, "2", metrics["Module 1 Analyzed Function Count"])
	assert.Equal(t, "6.50", metrics["Module 1 Average Complexity"])
	assert.Equal(t, "8.00", metrics["Module 1 Average Cognitive Complexity"])
	assert.Equal(t, "9", metrics["Module 1 Max Complexity"])
	assert.Equal(t, "1", metrics["Module 1 High Risk Function Count"])
	assert.Equal(t, "3", metrics["Module 1 Exception Handler Count"])
	assert.Equal(t, "2", metrics["Module 1 Dead Code Findings"])
	assert.Equal(t, "3", metrics["Module 1 Dead Code Blocks"])
}

type failingAnalyzeWriter struct{}

func (failingAnalyzeWriter) Write([]byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestAnalyzeFormatter_Write_CSVPropagatesWriterErrors(t *testing.T) {
	err := NewAnalyzeFormatter().Write(createMinimalAnalyzeResponse(), domain.OutputFormatCSV, failingAnalyzeWriter{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write CSV output")
	assert.ErrorContains(t, err, "write failed")
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
	assert.Contains(t, output, "Clone")
	assert.Contains(t, output, "Coupling")
}

func TestAnalyzeFormatter_Write_HTMLShowsSortableModuleQuality(t *testing.T) {
	response := createMinimalAnalyzeResponse()
	response.ModuleQuality = []domain.ModuleQualityMetrics{
		{
			ModuleName:    "pkg.hotspot",
			FilePath:      "pkg/hotspot.py",
			LinesOfCode:   120,
			FunctionCount: 4,
			ModuleComplexityMetrics: domain.ModuleComplexityMetrics{
				AnalyzedFunctionCount:      2,
				AverageComplexity:          6.5,
				AverageCognitiveComplexity: 8,
				MaxComplexity:              9,
				HighRiskFunctionCount:      1,
			},
			ModuleDeadCodeMetrics: domain.ModuleDeadCodeMetrics{DeadCodeFindingCount: 2},
		},
	}

	var output bytes.Buffer
	require.NoError(t, NewAnalyzeFormatter().Write(response, domain.OutputFormatHTML, &output))

	html := output.String()
	assert.Contains(t, html, "showTab('module-quality'")
	assert.Contains(t, html, `id="module-quality-table"`)
	assert.Contains(t, html, "pkg.hotspot")
	assert.Contains(t, html, "pkg/hotspot.py")
	assert.Contains(t, html, "sortModuleQuality")
	assert.Contains(t, html, `aria-label="Sort by average complexity"`)
	assert.Contains(t, html, `aria-label="Sort by analyzed function count"`)
	assert.Contains(t, html, `aria-label="Sort by exception handler count"`)
	assert.Contains(t, html, `aria-label="Sort by dead-code blocks"`)
}

func TestAnalyzeFormatter_WriteHTML_ShowsCloneGroupContentWhenEnabled(t *testing.T) {
	formatter := NewAnalyzeFormatter()
	response := createTestAnalyzeResponse()
	response.Clone = &domain.CloneResponse{
		Statistics: &domain.CloneStatistics{
			TotalClones:      2,
			TotalClonePairs:  0,
			TotalCloneGroups: 1,
		},
		Request: &domain.CloneRequest{ShowContent: domain.BoolPtr(true)},
		CloneGroups: []*domain.CloneGroup{
			{
				ID:         1,
				Type:       domain.Type2Clone,
				Similarity: 0.93,
				Size:       2,
				Clones: []*domain.Clone{
					{
						Location:  &domain.CloneLocation{FilePath: "alpha.py", StartLine: 1, EndLine: 2},
						LineCount: 2,
						Content:   "def alpha():\n    return 1",
					},
					{
						Location:  &domain.CloneLocation{FilePath: "beta.py", StartLine: 4, EndLine: 5},
						LineCount: 2,
						Content:   "def beta():\n    return 1",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.Write(response, domain.OutputFormatHTML, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Code Preview")
	assert.Contains(t, output, "def alpha():")
	assert.Contains(t, output, "def beta():")
}

func TestAnalyzeFormatter_WriteHTML_ShowsClonePairContentWhenEnabled(t *testing.T) {
	formatter := NewAnalyzeFormatter()
	response := createTestAnalyzeResponse()
	response.Clone = &domain.CloneResponse{
		Statistics: &domain.CloneStatistics{
			TotalClones:      2,
			TotalClonePairs:  1,
			TotalCloneGroups: 0,
		},
		Request: &domain.CloneRequest{ShowContent: domain.BoolPtr(true)},
		ClonePairs: []*domain.ClonePair{
			{
				Clone1: &domain.Clone{
					Location:  &domain.CloneLocation{FilePath: "alpha.py", StartLine: 1, EndLine: 2},
					LineCount: 2,
					Content:   "def alpha():\n    return 1",
				},
				Clone2: &domain.Clone{
					Location:  &domain.CloneLocation{FilePath: "beta.py", StartLine: 4, EndLine: 5},
					LineCount: 2,
					Content:   "def beta():\n    return 1",
				},
				Similarity: 0.94,
				Type:       domain.Type1Clone,
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.Write(response, domain.OutputFormatHTML, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Clone 1 Preview")
	assert.Contains(t, output, "Clone 2 Preview")
	assert.Contains(t, output, "def alpha():")
	assert.Contains(t, output, "def beta():")
}

func TestAnalyzeFormatter_WriteHTML_HidesCloneContentWhenDisabled(t *testing.T) {
	formatter := NewAnalyzeFormatter()
	response := createTestAnalyzeResponse()
	response.Clone = &domain.CloneResponse{
		Statistics: &domain.CloneStatistics{
			TotalClones:      2,
			TotalClonePairs:  0,
			TotalCloneGroups: 1,
		},
		Request: &domain.CloneRequest{ShowContent: domain.BoolPtr(false)},
		CloneGroups: []*domain.CloneGroup{
			{
				ID:         1,
				Type:       domain.Type2Clone,
				Similarity: 0.93,
				Size:       2,
				Clones: []*domain.Clone{
					{
						Location:  &domain.CloneLocation{FilePath: "alpha.py", StartLine: 1, EndLine: 2},
						LineCount: 2,
						Content:   "def alpha():\n    return 1",
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := formatter.Write(response, domain.OutputFormatHTML, &buf)
	require.NoError(t, err)

	output := buf.String()
	assert.NotContains(t, output, "Code Preview")
	assert.NotContains(t, output, "def alpha():")
}

func TestAnalyzeFormatter_Write_UnsupportedFormat(t *testing.T) {
	formatter := NewAnalyzeFormatter()
	response := createTestAnalyzeResponse()
	var buf bytes.Buffer

	err := formatter.Write(response, domain.OutputFormat("invalid"), &buf)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestAnalyzeFormatter_Write_IncludesCommunitySummary(t *testing.T) {
	response := createMinimalAnalyzeResponse()
	response.Summary.CommunitiesEnabled = true
	response.Communities = &domain.CommunityAnalysisResult{
		Algorithm:        "leiden",
		Scope:            "module",
		TotalCommunities: 2,
		Modularity:       0.42,
		Communities: []domain.CommunityMetrics{
			{ID: "community_1", Size: 3, InternalEdges: 2, ExternalEdges: 1},
			{ID: "community_2", Size: 2, InternalEdges: 1, ExternalEdges: 1},
		},
		BridgeModules: []domain.BridgeModule{
			{
				Module:              "bridge",
				Community:           "community_1",
				CrossCommunityEdges: 1,
				TargetCommunities:   []string{"community_2"},
			},
		},
	}

	formatter := NewAnalyzeFormatter()

	var textBuf bytes.Buffer
	require.NoError(t, formatter.Write(response, domain.OutputFormatText, &textBuf))
	text := textBuf.String()
	assert.Contains(t, text, "COMMUNITY DETECTION")
	assert.Contains(t, text, "BRIDGE MODULES")
	assert.Contains(t, text, "bridge")

	var csvBuf bytes.Buffer
	require.NoError(t, formatter.Write(response, domain.OutputFormatCSV, &csvBuf))
	csv := csvBuf.String()
	assert.Contains(t, csv, "Communities Enabled,true")
	assert.Contains(t, csv, "Total Communities,2")

	var htmlBuf bytes.Buffer
	require.NoError(t, formatter.Write(response, domain.OutputFormatHTML, &htmlBuf))
	html := htmlBuf.String()
	assert.Contains(t, html, "Communities")
	assert.Contains(t, html, "Module Communities")
	assert.Contains(t, html, "bridge")
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
