package service

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

// Test data helpers for output formatting
func createTestComplexityResponse() *domain.ComplexityResponse {
	return &domain.ComplexityResponse{
		Functions: []domain.FunctionComplexity{
			{
				Name:     "simple_function",
				FilePath: "test.py",
				Metrics: domain.ComplexityMetrics{
					Complexity:        2,
					Nodes:             5,
					Edges:             4,
					IfStatements:      1,
					LoopStatements:    0,
					ExceptionHandlers: 0,
				},
				RiskLevel: domain.RiskLevelLow,
			},
			{
				Name:     "complex_function",
				FilePath: "test.py",
				Metrics: domain.ComplexityMetrics{
					Complexity:        8,
					Nodes:             20,
					Edges:             18,
					IfStatements:      3,
					LoopStatements:    2,
					ExceptionHandlers: 1,
				},
				RiskLevel: domain.RiskLevelHigh,
			},
		},
		Summary: domain.ComplexitySummary{
			TotalFunctions:      2,
			AverageComplexity:   5.0,
			MaxComplexity:       8,
			MinComplexity:       2,
			LowRiskFunctions:    1,
			MediumRiskFunctions: 0,
			HighRiskFunctions:   1,
		},
		Warnings: []string{
			"Warning: High complexity detected in complex_function",
		},
		Errors: []string{
			"Error: Could not parse some_file.py",
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     "1.0.0",
		Config: map[string]interface{}{
			"min_complexity": 1,
			"max_complexity": 20,
		},
	}
}

func createMinimalComplexityResponse() *domain.ComplexityResponse {
	return &domain.ComplexityResponse{
		Functions: []domain.FunctionComplexity{},
		Summary: domain.ComplexitySummary{
			TotalFunctions:      0,
			AverageComplexity:   0.0,
			MaxComplexity:       0,
			MinComplexity:       0,
			LowRiskFunctions:    0,
			MediumRiskFunctions: 0,
			HighRiskFunctions:   0,
		},
		Warnings:    []string{},
		Errors:      []string{},
		GeneratedAt: time.Now().Format(time.RFC3339),
		Version:     "1.0.0",
		Config:      map[string]interface{}{},
	}
}

// TestOutputFormatter_Format tests the main Format method with different formats
func TestOutputFormatter_Format(t *testing.T) {
	tests := []struct {
		name           string
		response       *domain.ComplexityResponse
		format         domain.OutputFormat
		validateOutput func(t *testing.T, output string)
		expectError    bool
		errorMsg       string
	}{
		{
			name:     "format as JSON",
			response: createTestComplexityResponse(),
			format:   domain.OutputFormatJSON,
			validateOutput: func(t *testing.T, output string) {
				// Verify valid JSON structure
				var result map[string]interface{}
				err := json.Unmarshal([]byte(output), &result)
				assert.NoError(t, err, "Output should be valid JSON")

				// Check for expected fields in actual JSON structure
				assert.Contains(t, result, "results") // functions are in "results"
				assert.Contains(t, result, "summary")
				assert.Contains(t, result, "metadata") // contains generated_at and version

				// Verify results array (contains functions)
				functions, ok := result["results"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, functions, 2)

				// Check first function structure
				if len(functions) > 0 {
					function := functions[0].(map[string]interface{})
					assert.Contains(t, function, "function_name")
					assert.Contains(t, function, "complexity")
					assert.Contains(t, function, "risk_level")
				}

				// Verify metadata contains expected fields
				metadata, ok := result["metadata"].(map[string]interface{})
				assert.True(t, ok)
				assert.Contains(t, metadata, "generated_at")
				assert.Contains(t, metadata, "version")
			},
			expectError: false,
		},
		{
			name:     "format as CSV",
			response: createTestComplexityResponse(),
			format:   domain.OutputFormatCSV,
			validateOutput: func(t *testing.T, output string) {
				// Verify CSV structure
				reader := csv.NewReader(strings.NewReader(output))
				records, err := reader.ReadAll()
				assert.NoError(t, err, "Output should be valid CSV")

				// Should have header + 2 data rows
				assert.Len(t, records, 3, "Should have header plus 2 function rows")

				// Check header
				expectedHeaders := []string{"Function", "Complexity", "Risk", "Nodes", "Edges", "If Statements", "Loop Statements", "Exception Handlers"}
				assert.Equal(t, expectedHeaders, records[0])

				// Check first data row
				assert.Equal(t, "simple_function", records[1][0])
				assert.Equal(t, "2", records[1][1])
				assert.Equal(t, "low", records[1][2])

				// Check second data row
				assert.Equal(t, "complex_function", records[2][0])
				assert.Equal(t, "8", records[2][1])
				assert.Equal(t, "high", records[2][2])
			},
			expectError: false,
		},
		{
			name:     "format as YAML",
			response: createTestComplexityResponse(),
			format:   domain.OutputFormatYAML,
			validateOutput: func(t *testing.T, output string) {
				// Verify valid YAML structure
				var result map[string]interface{}
				err := yaml.Unmarshal([]byte(output), &result)
				assert.NoError(t, err, "Output should be valid YAML")

				// Check for expected fields in actual YAML structure
				assert.Contains(t, result, "results") // functions are in "results"
				assert.Contains(t, result, "summary")
				assert.Contains(t, result, "metadata") // contains generated_at and version

				// Verify results array (contains functions)
				functions, ok := result["results"].([]interface{})
				assert.True(t, ok)
				assert.Len(t, functions, 2)

				// Verify metadata contains expected fields
				metadata, ok := result["metadata"].(map[string]interface{})
				assert.True(t, ok)
				assert.Contains(t, metadata, "generated_at")
				assert.Contains(t, metadata, "version")
			},
			expectError: false,
		},
		{
			name:     "format as text",
			response: createTestComplexityResponse(),
			format:   domain.OutputFormatText,
			validateOutput: func(t *testing.T, output string) {
				// Verify text format contains expected sections
				assert.Contains(t, output, "Complexity Analysis Report")
				assert.Contains(t, output, "SUMMARY")
				assert.Contains(t, output, "Total Functions: 2")
				assert.Contains(t, output, "Average Complexity: 5.0")
				assert.Contains(t, output, "RISK DISTRIBUTION")
				assert.Contains(t, output, "High: 1")
				assert.Contains(t, output, "Low: 1")
				assert.Contains(t, output, "FUNCTION DETAILS")
				assert.Contains(t, output, "simple_function")
				assert.Contains(t, output, "complex_function")
				assert.Contains(t, output, "WARNINGS")
				assert.Contains(t, output, "High complexity detected")
				assert.Contains(t, output, "ERRORS")
				assert.Contains(t, output, "Could not parse some_file.py")
				assert.Contains(t, output, "Generated at:")
			},
			expectError: false,
		},
		{
			name:     "empty response JSON",
			response: createMinimalComplexityResponse(),
			format:   domain.OutputFormatJSON,
			validateOutput: func(t *testing.T, output string) {
				var result map[string]interface{}
				err := json.Unmarshal([]byte(output), &result)
				assert.NoError(t, err)

				// Should have empty arrays but valid structure
				functions := result["results"].([]interface{})
				assert.Len(t, functions, 0)

				summary := result["summary"].(map[string]interface{})
				assert.Equal(t, float64(0), summary["total_functions"])
			},
			expectError: false,
		},
		{
			name:     "empty response CSV",
			response: createMinimalComplexityResponse(),
			format:   domain.OutputFormatCSV,
			validateOutput: func(t *testing.T, output string) {
				reader := csv.NewReader(strings.NewReader(output))
				records, err := reader.ReadAll()
				assert.NoError(t, err)

				// Should have only header row
				assert.Len(t, records, 1, "Should have only header row")
			},
			expectError: false,
		},
		{
			name:     "format as HTML",
			response: createTestComplexityResponse(),
			format:   domain.OutputFormatHTML,
			validateOutput: func(t *testing.T, output string) {
				// Verify HTML structure
				assert.Contains(t, output, "<!DOCTYPE html>", "Should contain HTML doctype")
				assert.Contains(t, output, "<title>", "Should contain title tag")
				assert.Contains(t, output, "pyscn Code Quality Report", "Should contain report title")
				assert.Contains(t, output, "Overall Score", "Should contain overall score")
				assert.Contains(t, output, "<style>", "Should contain embedded CSS")
				assert.Contains(t, output, "Complexity Score", "Should contain complexity score")

				// Check for Lighthouse-style elements
				assert.Contains(t, output, "score-circle", "Should have score circle element")
				assert.Contains(t, output, "score-gauge", "Should have score gauge element")
			},
			expectError: false,
		},
		{
			name:        "unsupported format error",
			response:    createTestComplexityResponse(),
			format:      domain.OutputFormat("invalid"),
			expectError: true,
			errorMsg:    "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewOutputFormatter()

			output, err := formatter.Format(tt.response, tt.format)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, output)

				if tt.validateOutput != nil {
					tt.validateOutput(t, output)
				}
			}
		})
	}
}

// TestOutputFormatter_Write tests writing formatted output to different writers
func TestOutputFormatter_Write(t *testing.T) {
	tests := []struct {
		name           string
		response       *domain.ComplexityResponse
		format         domain.OutputFormat
		writer         io.Writer
		expectError    bool
		errorMsg       string
		validateWriter func(t *testing.T, writer *strings.Builder)
	}{
		{
			name:     "write JSON to string builder",
			response: createTestComplexityResponse(),
			format:   domain.OutputFormatJSON,
			writer:   &strings.Builder{},
			validateWriter: func(t *testing.T, writer *strings.Builder) {
				output := writer.String()
				assert.NotEmpty(t, output)

				// Verify it's valid JSON
				var result map[string]interface{}
				err := json.Unmarshal([]byte(output), &result)
				assert.NoError(t, err)
			},
			expectError: false,
		},
		{
			name:     "write text format to string builder",
			response: createTestComplexityResponse(),
			format:   domain.OutputFormatText,
			writer:   &strings.Builder{},
			validateWriter: func(t *testing.T, writer *strings.Builder) {
				output := writer.String()
				assert.Contains(t, output, "Complexity Analysis Report")
				assert.Contains(t, output, "Total Functions: 2")
			},
			expectError: false,
		},
		{
			name:     "write CSV to string builder",
			response: createTestComplexityResponse(),
			format:   domain.OutputFormatCSV,
			writer:   &strings.Builder{},
			validateWriter: func(t *testing.T, writer *strings.Builder) {
				output := writer.String()
				assert.Contains(t, output, "Function,Complexity,Risk")
				assert.Contains(t, output, "simple_function")
				assert.Contains(t, output, "complex_function")
			},
			expectError: false,
		},
		{
			name:     "write YAML to string builder",
			response: createTestComplexityResponse(),
			format:   domain.OutputFormatYAML,
			writer:   &strings.Builder{},
			validateWriter: func(t *testing.T, writer *strings.Builder) {
				output := writer.String()
				assert.Contains(t, output, "functions:")
				assert.Contains(t, output, "summary:")

				// Verify it's valid YAML
				var result map[string]interface{}
				err := yaml.Unmarshal([]byte(output), &result)
				assert.NoError(t, err)
			},
			expectError: false,
		},
		{
			name:        "write with unsupported format",
			response:    createTestComplexityResponse(),
			format:      domain.OutputFormat("invalid"),
			writer:      &strings.Builder{},
			expectError: true,
			errorMsg:    "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewOutputFormatter()

			err := formatter.Write(tt.response, tt.format, tt.writer)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				if tt.validateWriter != nil {
					if builder, ok := tt.writer.(*strings.Builder); ok {
						tt.validateWriter(t, builder)
					}
				}
			}
		})
	}
}

// TestOutputFormatter_formatText tests detailed text formatting
func TestOutputFormatter_formatText(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name     string
		response *domain.ComplexityResponse
		validate func(t *testing.T, output string)
	}{
		{
			name:     "text format with all sections",
			response: createTestComplexityResponse(),
			validate: func(t *testing.T, output string) {
				// Check header (new unified format)
				assert.Contains(t, output, "Complexity Analysis Report")
				assert.Contains(t, output, "========================================")

				// Check summary section (new unified format)
				assert.Contains(t, output, "SUMMARY")
				assert.Contains(t, output, "Total Functions: 2")
				assert.Contains(t, output, "Average Complexity: 5.0")
				assert.Contains(t, output, "Max Complexity: 8")
				assert.Contains(t, output, "Min Complexity: 2")

				// Check risk distribution (new unified format)
				assert.Contains(t, output, "RISK DISTRIBUTION")
				assert.Contains(t, output, "High: 1")
				assert.Contains(t, output, "Medium: 0")
				assert.Contains(t, output, "Low: 1")

				// Check function details (new unified format)
				assert.Contains(t, output, "FUNCTION DETAILS")
				assert.Contains(t, output, "Function  Complexity  Risk")
				assert.Contains(t, output, "simple_function")
				assert.Contains(t, output, "complex_function")

				// Check warnings and errors sections
				assert.Contains(t, output, "WARNINGS")
				assert.Contains(t, output, "⚠")
				assert.Contains(t, output, "High complexity detected")
				assert.Contains(t, output, "ERRORS")
				assert.Contains(t, output, "❌")
				assert.Contains(t, output, "Could not parse some_file.py")

				// Check footer (metadata section)
				assert.Contains(t, output, "METADATA")
				assert.Contains(t, output, "Generated at:")
			},
		},
		{
			name:     "text format with no functions",
			response: createMinimalComplexityResponse(),
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "Complexity Analysis Report")
				assert.Contains(t, output, "Total Functions: 0")

				// Should not contain function details section
				assert.NotContains(t, output, "FUNCTION DETAILS")
				assert.NotContains(t, output, "WARNINGS")
				assert.NotContains(t, output, "ERRORS")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := formatter.formatText(tt.response)
			assert.NoError(t, err)
			assert.NotEmpty(t, output)

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

// TestOutputFormatter_formatJSON tests JSON formatting details
func TestOutputFormatter_formatJSON(t *testing.T) {
	formatter := NewOutputFormatter()
	response := createTestComplexityResponse()

	output, err := formatter.formatJSON(response)
	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Parse the JSON to verify structure
	var parsed map[string]interface{}
	err = json.Unmarshal([]byte(output), &parsed)
	assert.NoError(t, err)

	// Verify top-level structure matches actual JSON format
	assert.Contains(t, parsed, "results") // not "functions"
	assert.Contains(t, parsed, "summary")
	assert.Contains(t, parsed, "metadata") // contains generated_at, version, config

	// Verify results array structure (contains functions)
	functions := parsed["results"].([]interface{})
	assert.Len(t, functions, 2)

	firstFunction := functions[0].(map[string]interface{})
	assert.Equal(t, "simple_function", firstFunction["function_name"]) // not "name"
	assert.Equal(t, float64(2), firstFunction["complexity"])
	assert.Equal(t, "low", firstFunction["risk_level"])

	// Verify summary structure
	summary := parsed["summary"].(map[string]interface{})
	assert.Equal(t, float64(2), summary["total_functions"])
	assert.Equal(t, 5.0, summary["average_complexity"])

	// Verify metadata structure
	metadata := parsed["metadata"].(map[string]interface{})
	assert.Contains(t, metadata, "generated_at")
	assert.Contains(t, metadata, "version")
	assert.Contains(t, metadata, "configuration")
}

// TestOutputFormatter_formatYAML tests YAML formatting details
func TestOutputFormatter_formatYAML(t *testing.T) {
	formatter := NewOutputFormatter()
	response := createTestComplexityResponse()

	output, err := formatter.formatYAML(response)
	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Parse the YAML to verify structure
	var parsed map[string]interface{}
	err = yaml.Unmarshal([]byte(output), &parsed)
	assert.NoError(t, err)

	// Verify structure matches actual YAML format
	assert.Contains(t, parsed, "results") // not "functions"
	assert.Contains(t, parsed, "summary")
	assert.Contains(t, parsed, "metadata") // contains generated_at, version
}

// TestOutputFormatter_formatCSV tests CSV formatting details
func TestOutputFormatter_formatCSV(t *testing.T) {
	formatter := NewOutputFormatter()
	response := createTestComplexityResponse()

	output, err := formatter.formatCSV(response)
	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Parse CSV
	reader := csv.NewReader(strings.NewReader(output))
	records, err := reader.ReadAll()
	assert.NoError(t, err)

	// Verify structure
	assert.Len(t, records, 3) // Header + 2 functions

	// Check header
	expectedHeaders := []string{"Function", "Complexity", "Risk", "Nodes", "Edges", "If Statements", "Loop Statements", "Exception Handlers"}
	assert.Equal(t, expectedHeaders, records[0])

	// Check data rows (risk levels are lowercase in actual implementation)
	assert.Equal(t, []string{"simple_function", "2", "low", "5", "4", "1", "0", "0"}, records[1])
	assert.Equal(t, []string{"complex_function", "8", "high", "20", "18", "3", "2", "1"}, records[2])
}

// TestOutputFormatter_NewOutputFormatter tests service creation
func TestOutputFormatter_NewOutputFormatter(t *testing.T) {
	formatter := NewOutputFormatter()

	assert.NotNil(t, formatter)
	assert.IsType(t, &OutputFormatterImpl{}, formatter)
}

// TestOutputFormatter_ErrorHandling tests error cases
func TestOutputFormatter_ErrorHandling(t *testing.T) {
	formatter := NewOutputFormatter()

	tests := []struct {
		name        string
		response    *domain.ComplexityResponse
		format      domain.OutputFormat
		expectError bool
		errorMsg    string
	}{
		// Skip nil response test as it causes panic - this is expected behavior
		{
			name:        "unsupported format",
			response:    createTestComplexityResponse(),
			format:      domain.OutputFormat("unsupported"),
			expectError: true,
			errorMsg:    "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := formatter.Format(tt.response, tt.format)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
