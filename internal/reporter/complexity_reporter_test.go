package reporter

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/internal/config"
	"gopkg.in/yaml.v3"
)

// mockComplexityResult implements ComplexityResult interface for testing
type mockComplexityResult struct {
	complexity        int
	functionName      string
	riskLevel         string
	nodes             int
	edges             int
	ifStatements      int
	loopStatements    int
	exceptionHandlers int
	switchCases       int
}

func (m *mockComplexityResult) GetComplexity() int      { return m.complexity }
func (m *mockComplexityResult) GetFunctionName() string { return m.functionName }
func (m *mockComplexityResult) GetRiskLevel() string    { return m.riskLevel }

func (m *mockComplexityResult) GetDetailedMetrics() map[string]int {
	return map[string]int{
		"nodes":              m.nodes,
		"edges":              m.edges,
		"if_statements":      m.ifStatements,
		"loop_statements":    m.loopStatements,
		"exception_handlers": m.exceptionHandlers,
		"switch_cases":       m.switchCases,
	}
}

func createTestResults() []ComplexityResult {
	return []ComplexityResult{
		&mockComplexityResult{
			complexity:        1,
			functionName:      "simple_function",
			riskLevel:         "low",
			nodes:             1,
			edges:             2,
			ifStatements:      0,
			loopStatements:    0,
			exceptionHandlers: 0,
		},
		&mockComplexityResult{
			complexity:        5,
			functionName:      "medium_function",
			riskLevel:         "low",
			nodes:             5,
			edges:             8,
			ifStatements:      2,
			loopStatements:    1,
			exceptionHandlers: 0,
		},
		&mockComplexityResult{
			complexity:        15,
			functionName:      "complex_function",
			riskLevel:         "medium",
			nodes:             15,
			edges:             28,
			ifStatements:      7,
			loopStatements:    2,
			exceptionHandlers: 1,
		},
		&mockComplexityResult{
			complexity:        25,
			functionName:      "very_complex_function",
			riskLevel:         "high",
			nodes:             25,
			edges:             48,
			ifStatements:      12,
			loopStatements:    3,
			exceptionHandlers: 2,
		},
	}
}

func TestNewComplexityReporter(t *testing.T) {
	t.Run("ValidConfiguration", func(t *testing.T) {
		cfg := config.DefaultConfig()
		var buffer bytes.Buffer

		reporter, err := NewComplexityReporter(cfg, &buffer)
		if err != nil {
			t.Fatalf("Failed to create reporter: %v", err)
		}

		if reporter == nil {
			t.Fatal("Expected reporter instance, got nil")
		}
		if reporter.config != cfg {
			t.Error("Reporter config not set correctly")
		}
		if reporter.writer != &buffer {
			t.Error("Reporter writer not set correctly")
		}
	})

	t.Run("NilConfiguration", func(t *testing.T) {
		var buffer bytes.Buffer

		reporter, err := NewComplexityReporter(nil, &buffer)

		if err == nil {
			t.Fatal("Expected error for nil configuration, but got none")
		}
		if reporter != nil {
			t.Error("Expected nil reporter for nil configuration")
		}
		if !strings.Contains(err.Error(), "configuration cannot be nil") {
			t.Errorf("Expected nil config error, got: %v", err)
		}
	})

	t.Run("NilWriter", func(t *testing.T) {
		cfg := config.DefaultConfig()

		reporter, err := NewComplexityReporter(cfg, nil)

		if err == nil {
			t.Fatal("Expected error for nil writer, but got none")
		}
		if reporter != nil {
			t.Error("Expected nil reporter for nil writer")
		}
		if !strings.Contains(err.Error(), "writer cannot be nil") {
			t.Errorf("Expected nil writer error, got: %v", err)
		}
	})
}

func TestGenerateReport(t *testing.T) {
	cfg := config.DefaultConfig()
	var buffer bytes.Buffer
	reporter, err := NewComplexityReporter(cfg, &buffer)
	if err != nil {
		t.Fatalf("Failed to create reporter: %v", err)
	}

	results := createTestResults()
	report := reporter.GenerateReport(results, 2) // filesAnalyzed count

	// Test basic report structure
	if report == nil {
		t.Fatal("Expected report, got nil")
	}
	if len(report.Results) != 4 {
		t.Errorf("Expected 4 results, got %d", len(report.Results))
	}

	// Test summary
	if report.Summary.TotalFunctions != 4 {
		t.Errorf("Expected 4 total functions, got %d", report.Summary.TotalFunctions)
	}
	if report.Summary.AverageComplexity != 11.5 { // (1+5+15+25)/4
		t.Errorf("Expected average complexity 11.5, got %.1f", report.Summary.AverageComplexity)
	}
	if report.Summary.MaxComplexity != 25 {
		t.Errorf("Expected max complexity 25, got %d", report.Summary.MaxComplexity)
	}
	if report.Summary.MinComplexity != 1 {
		t.Errorf("Expected min complexity 1, got %d", report.Summary.MinComplexity)
	}

	// Test risk distribution
	if report.Summary.RiskDistribution.Low != 2 {
		t.Errorf("Expected 2 low risk functions, got %d", report.Summary.RiskDistribution.Low)
	}
	if report.Summary.RiskDistribution.Medium != 1 {
		t.Errorf("Expected 1 medium risk function, got %d", report.Summary.RiskDistribution.Medium)
	}
	if report.Summary.RiskDistribution.High != 1 {
		t.Errorf("Expected 1 high risk function, got %d", report.Summary.RiskDistribution.High)
	}

	// Test complexity distribution
	if report.Summary.ComplexityDistribution["1"] != 1 {
		t.Errorf("Expected 1 function with complexity 1, got %d", report.Summary.ComplexityDistribution["1"])
	}
	if report.Summary.ComplexityDistribution["2-5"] != 1 {
		t.Errorf("Expected 1 function with complexity 2-5, got %d", report.Summary.ComplexityDistribution["2-5"])
	}
	if report.Summary.ComplexityDistribution["11-20"] != 1 {
		t.Errorf("Expected 1 function with complexity 11-20, got %d", report.Summary.ComplexityDistribution["11-20"])
	}
	if report.Summary.ComplexityDistribution["21+"] != 1 {
		t.Errorf("Expected 1 function with complexity 21+, got %d", report.Summary.ComplexityDistribution["21+"])
	}

	// Test metadata
	if report.Metadata.GeneratedAt.IsZero() {
		t.Error("Expected generated timestamp, got zero time")
	}
	if time.Since(report.Metadata.GeneratedAt) > time.Minute {
		t.Error("Generated timestamp should be recent")
	}
}

func TestFilterAndSortResults(t *testing.T) {
	results := createTestResults()

	t.Run("FilterByMinComplexity", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Output.MinComplexity = 10 // Should filter out first two functions

		var buffer bytes.Buffer
		reporter, err := NewComplexityReporter(cfg, &buffer)
		if err != nil {
			t.Fatalf("Failed to create reporter: %v", err)
		}

		filtered := reporter.filterAndSortResults(results)

		if len(filtered) != 2 {
			t.Errorf("Expected 2 filtered results, got %d", len(filtered))
		}

		for _, result := range filtered {
			if result.GetComplexity() < 10 {
				t.Errorf("Result with complexity %d should have been filtered out", result.GetComplexity())
			}
		}
	})

	t.Run("SortByName", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Output.SortBy = "name"

		var buffer bytes.Buffer
		reporter, err := NewComplexityReporter(cfg, &buffer)
		if err != nil {
			t.Fatalf("Failed to create reporter: %v", err)
		}

		sorted := reporter.filterAndSortResults(results)

		expectedOrder := []string{"complex_function", "medium_function", "simple_function", "very_complex_function"}
		for i, result := range sorted {
			if result.GetFunctionName() != expectedOrder[i] {
				t.Errorf("Expected function %s at position %d, got %s",
					expectedOrder[i], i, result.GetFunctionName())
			}
		}
	})

	t.Run("SortByComplexity", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Output.SortBy = "complexity"

		var buffer bytes.Buffer
		reporter, err := NewComplexityReporter(cfg, &buffer)
		if err != nil {
			t.Fatalf("Failed to create reporter: %v", err)
		}

		sorted := reporter.filterAndSortResults(results)

		// Should be sorted in descending order by complexity
		for i := 1; i < len(sorted); i++ {
			if sorted[i].GetComplexity() > sorted[i-1].GetComplexity() {
				t.Errorf("Results not sorted by complexity: %d > %d at positions %d, %d",
					sorted[i].GetComplexity(), sorted[i-1].GetComplexity(), i, i-1)
			}
		}
	})

	t.Run("SortByRisk", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Output.SortBy = "risk"

		var buffer bytes.Buffer
		reporter, err := NewComplexityReporter(cfg, &buffer)
		if err != nil {
			t.Fatalf("Failed to create reporter: %v", err)
		}

		sorted := reporter.filterAndSortResults(results)

		// Should be sorted high > medium > low
		riskOrder := []string{"high", "medium", "low", "low"}
		for i, result := range sorted {
			if result.GetRiskLevel() != riskOrder[i] {
				t.Errorf("Expected risk level %s at position %d, got %s",
					riskOrder[i], i, result.GetRiskLevel())
			}
		}
	})
}

func TestGenerateWarnings(t *testing.T) {
	t.Run("HighComplexityWarnings", func(t *testing.T) {
		cfg := config.DefaultConfig()
		var buffer bytes.Buffer
		reporter, err := NewComplexityReporter(cfg, &buffer)
		if err != nil {
			t.Fatalf("Failed to create reporter: %v", err)
		}

		results := createTestResults()
		warnings := reporter.generateWarnings(results)

		// Should have one high complexity warning for very_complex_function
		highComplexityWarnings := 0
		for _, warning := range warnings {
			if warning.Type == "high_complexity" {
				highComplexityWarnings++
				if warning.FunctionName != "very_complex_function" {
					t.Errorf("Expected high complexity warning for very_complex_function, got %s",
						warning.FunctionName)
				}
			}
		}

		if highComplexityWarnings != 1 {
			t.Errorf("Expected 1 high complexity warning, got %d", highComplexityWarnings)
		}
	})

	t.Run("MaxComplexityWarnings", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Complexity.MaxComplexity = 20 // Should trigger warning for very_complex_function

		var buffer bytes.Buffer
		reporter, err := NewComplexityReporter(cfg, &buffer)
		if err != nil {
			t.Fatalf("Failed to create reporter: %v", err)
		}

		results := createTestResults()
		warnings := reporter.generateWarnings(results)

		// Should have one max complexity warning
		maxComplexityWarnings := 0
		for _, warning := range warnings {
			if warning.Type == "max_complexity_exceeded" {
				maxComplexityWarnings++
				if warning.Complexity != 25 {
					t.Errorf("Expected warning complexity 25, got %d", warning.Complexity)
				}
			}
		}

		if maxComplexityWarnings != 1 {
			t.Errorf("Expected 1 max complexity warning, got %d", maxComplexityWarnings)
		}
	})
}

func TestOutputJSON(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Output.Format = "json"

	var buffer bytes.Buffer
	reporter, err := NewComplexityReporter(cfg, &buffer)
	if err != nil {
		t.Fatalf("Failed to create reporter: %v", err)
	}

	results := createTestResults()
	err = reporter.ReportComplexity(results)

	if err != nil {
		t.Fatalf("Failed to output JSON: %v", err)
	}

	// Verify it's valid JSON
	var report ComplexityReport
	err = json.Unmarshal(buffer.Bytes(), &report)
	if err != nil {
		t.Fatalf("Generated invalid JSON: %v", err)
	}

	// Verify content
	if len(report.Results) != 4 {
		t.Errorf("Expected 4 results in JSON, got %d", len(report.Results))
	}
	if report.Summary.TotalFunctions != 4 {
		t.Errorf("Expected 4 total functions in JSON, got %d", report.Summary.TotalFunctions)
	}
}

func TestOutputYAML(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Output.Format = "yaml"

	var buffer bytes.Buffer
	reporter, err := NewComplexityReporter(cfg, &buffer)
	if err != nil {
		t.Fatalf("Failed to create reporter: %v", err)
	}

	results := createTestResults()
	err = reporter.ReportComplexity(results)

	if err != nil {
		t.Fatalf("Failed to output YAML: %v", err)
	}

	// Verify it's valid YAML
	var report ComplexityReport
	err = yaml.Unmarshal(buffer.Bytes(), &report)
	if err != nil {
		t.Fatalf("Generated invalid YAML: %v", err)
	}

	// Verify content
	if len(report.Results) != 4 {
		t.Errorf("Expected 4 results in YAML, got %d", len(report.Results))
	}
	if report.Summary.TotalFunctions != 4 {
		t.Errorf("Expected 4 total functions in YAML, got %d", report.Summary.TotalFunctions)
	}
}

func TestOutputCSV(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Output.Format = "csv"

	var buffer bytes.Buffer
	reporter, err := NewComplexityReporter(cfg, &buffer)
	if err != nil {
		t.Fatalf("Failed to create reporter: %v", err)
	}

	results := createTestResults()
	err = reporter.ReportComplexity(results)

	if err != nil {
		t.Fatalf("Failed to output CSV: %v", err)
	}

	// Parse CSV output
	reader := csv.NewReader(strings.NewReader(buffer.String()))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("Generated invalid CSV: %v", err)
	}

	// Should have header + 4 data rows
	if len(records) != 5 {
		t.Errorf("Expected 5 CSV records (header + 4 data), got %d", len(records))
	}

	// Verify header
	expectedHeader := []string{
		"Function", "Complexity", "Risk", "Nodes", "Edges",
		"If Statements", "Loop Statements", "Exception Handlers",
	}
	for i, field := range records[0] {
		if field != expectedHeader[i] {
			t.Errorf("Expected header field %s, got %s", expectedHeader[i], field)
		}
	}

	// Verify first data row (should be sorted by name)
	firstRow := records[1]
	if firstRow[0] != "complex_function" { // Sorted by name
		t.Errorf("Expected first function to be complex_function, got %s", firstRow[0])
	}
	if firstRow[1] != "15" {
		t.Errorf("Expected complexity 15, got %s", firstRow[1])
	}
	if firstRow[2] != "medium" {
		t.Errorf("Expected risk medium, got %s", firstRow[2])
	}
}

func TestOutputText(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Output.Format = "text"
	cfg.Output.ShowDetails = true

	var buffer bytes.Buffer
	reporter, err := NewComplexityReporter(cfg, &buffer)
	if err != nil {
		t.Fatalf("Failed to create reporter: %v", err)
	}

	results := createTestResults()
	err = reporter.ReportComplexity(results)

	if err != nil {
		t.Fatalf("Failed to output text: %v", err)
	}

	output := buffer.String()

	// Verify key sections are present
	if !strings.Contains(output, "Complexity Analysis Report") {
		t.Error("Missing report title")
	}
	if !strings.Contains(output, "Summary:") {
		t.Error("Missing summary section")
	}
	if !strings.Contains(output, "Total Functions: 4") {
		t.Error("Missing total functions in summary")
	}
	if !strings.Contains(output, "Risk Distribution:") {
		t.Error("Missing risk distribution section")
	}
	if !strings.Contains(output, "Function Details:") {
		t.Error("Missing function details section")
	}
	if !strings.Contains(output, "Generated at:") {
		t.Error("Missing generation timestamp")
	}

	// Verify function details are shown
	for _, result := range results {
		if !strings.Contains(output, result.GetFunctionName()) {
			t.Errorf("Missing function %s in output", result.GetFunctionName())
		}
	}

	// Verify details are shown when enabled
	if !strings.Contains(output, "Nodes") {
		t.Error("Missing detailed columns when ShowDetails is true")
	}
}

func TestOutputTextWithWarnings(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Output.Format = "text"
	cfg.Complexity.MaxComplexity = 20 // Will trigger warning

	var buffer bytes.Buffer
	reporter, err := NewComplexityReporter(cfg, &buffer)
	if err != nil {
		t.Fatalf("Failed to create reporter: %v", err)
	}

	results := createTestResults()
	err = reporter.ReportComplexity(results)

	if err != nil {
		t.Fatalf("Failed to output text with warnings: %v", err)
	}

	output := buffer.String()

	// Should contain warnings section
	if !strings.Contains(output, "Warnings:") {
		t.Error("Missing warnings section")
	}
	if !strings.Contains(output, "MAX_COMPLEXITY_EXCEEDED") {
		t.Error("Missing max complexity exceeded warning")
	}
	if !strings.Contains(output, "HIGH_COMPLEXITY") {
		t.Error("Missing high complexity warning")
	}
}

func TestFormatComplexityBrief(t *testing.T) {
	t.Run("EmptyResults", func(t *testing.T) {
		brief := FormatComplexityBrief([]ComplexityResult{})
		if brief != "No functions analyzed" {
			t.Errorf("Expected 'No functions analyzed', got %s", brief)
		}
	})

	t.Run("WithResults", func(t *testing.T) {
		results := createTestResults()
		brief := FormatComplexityBrief(results)

		expectedSubstrings := []string{
			"4 functions analyzed",
			"Avg: 11.5",
			"Max: 25",
			"High Risk: 1",
		}

		for _, substring := range expectedSubstrings {
			if !strings.Contains(brief, substring) {
				t.Errorf("Expected brief to contain '%s', got: %s", substring, brief)
			}
		}
	})
}

func TestComplexityReporterEdgeCases(t *testing.T) {
	t.Run("EmptyResults", func(t *testing.T) {
		cfg := config.DefaultConfig()
		var buffer bytes.Buffer
		reporter, err := NewComplexityReporter(cfg, &buffer)
		if err != nil {
			t.Fatalf("Failed to create reporter: %v", err)
		}

		err = reporter.ReportComplexity([]ComplexityResult{})
		if err != nil {
			t.Fatalf("Failed to handle empty results: %v", err)
		}

		// Should generate valid output with zero functions
		report := reporter.GenerateReport([]ComplexityResult{}, 0)
		if report.Summary.TotalFunctions != 0 {
			t.Errorf("Expected 0 total functions, got %d", report.Summary.TotalFunctions)
		}
	})

	t.Run("InvalidOutputFormat", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Output.Format = "invalid"

		var buffer bytes.Buffer
		reporter, err := NewComplexityReporter(cfg, &buffer)

		// Should fail to create reporter with invalid configuration
		if err == nil {
			t.Fatal("Expected error for invalid output format, but got none")
		}

		// Should contain validation error message
		if !strings.Contains(err.Error(), "invalid output.format") {
			t.Errorf("Expected validation error for invalid format, got: %v", err)
		}

		// Reporter should be nil
		if reporter != nil {
			t.Error("Expected nil reporter for invalid configuration")
		}
	})

	t.Run("SingleFunction", func(t *testing.T) {
		cfg := config.DefaultConfig()
		var buffer bytes.Buffer
		reporter, err := NewComplexityReporter(cfg, &buffer)
		if err != nil {
			t.Fatalf("Failed to create reporter: %v", err)
		}

		singleResult := []ComplexityResult{
			&mockComplexityResult{
				complexity:   10,
				functionName: "single_function",
				riskLevel:    "medium",
			},
		}

		report := reporter.GenerateReport(singleResult, 1)

		if report.Summary.AverageComplexity != 10.0 {
			t.Errorf("Expected average complexity 10.0, got %.1f", report.Summary.AverageComplexity)
		}
		if report.Summary.MaxComplexity != 10 {
			t.Errorf("Expected max complexity 10, got %d", report.Summary.MaxComplexity)
		}
		if report.Summary.MinComplexity != 10 {
			t.Errorf("Expected min complexity 10, got %d", report.Summary.MinComplexity)
		}
	})
}
