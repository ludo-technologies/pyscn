package reporter

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ludo-technologies/pyscn/domain"
	"github.com/ludo-technologies/pyscn/internal/config"

	"gopkg.in/yaml.v3"
)

// mockComplexityResult implements ComplexityResult interface for testing
type mockComplexityResult struct {
	complexity        int
	functionName      string
	scopeKind         domain.AnalysisScopeKind
	sourceLocation    ComplexitySourceLocation
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
func (m *mockComplexityResult) GetScopeKind() domain.AnalysisScopeKind {
	return m.scopeKind
}
func (m *mockComplexityResult) GetSourceLocation() ComplexitySourceLocation {
	return m.sourceLocation
}
func (m *mockComplexityResult) GetRiskLevel() string { return m.riskLevel }

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
			scopeKind:         domain.AnalysisScopeModule,
			sourceLocation:    ComplexitySourceLocation{FilePath: "a.py", StartLine: 1, StartColumn: 1, EndLine: 4},
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
			scopeKind:         domain.AnalysisScopeFunction,
			sourceLocation:    ComplexitySourceLocation{FilePath: "a.py", StartLine: 6, StartColumn: 1, EndLine: 12},
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
			scopeKind:         domain.AnalysisScopeClass,
			sourceLocation:    ComplexitySourceLocation{FilePath: "b.py", StartLine: 3, StartColumn: 1, EndLine: 24},
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
			scopeKind:         domain.AnalysisScopeFunction,
			sourceLocation:    ComplexitySourceLocation{FilePath: "b.py", StartLine: 30, StartColumn: 1, EndLine: 64},
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
	if len(report.Results) != 3 {
		t.Errorf("Expected 3 legacy results, got %d", len(report.Results))
	}
	for _, result := range report.Results {
		if result.ScopeKind == domain.AnalysisScopeUnknown {
			t.Errorf("Result %q has no scope kind", result.FunctionName)
		}
		if result.FilePath == "" || result.StartLine == 0 {
			t.Errorf("Result %q has no source location: %+v", result.FunctionName, result)
		}
	}
	if len(report.ClassScopes) != 1 {
		t.Fatalf("Expected 1 additive class scope, got %d", len(report.ClassScopes))
	}
	classScope := report.ClassScopes[0]
	if classScope.FunctionName != "complex_function" || classScope.ScopeKind != domain.AnalysisScopeClass {
		t.Fatalf("Unexpected class scope: %+v", classScope)
	}
	if classScope.FilePath != "b.py" || classScope.StartLine != 3 || classScope.StartColumn != 1 || classScope.EndLine != 24 {
		t.Fatalf("Class source location was not preserved: %+v", classScope)
	}

	// Test summary
	if report.Summary.TotalFunctions != 3 {
		t.Errorf("Expected 3 total functions, got %d", report.Summary.TotalFunctions)
	}
	if report.Summary.AverageComplexity != float64(31)/3 {
		t.Errorf("Expected average complexity %.2f, got %.2f", float64(31)/3, report.Summary.AverageComplexity)
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
	if report.Summary.RiskDistribution.Medium != 0 {
		t.Errorf("Expected 0 medium risk functions, got %d", report.Summary.RiskDistribution.Medium)
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
	if report.Summary.ComplexityDistribution["11-20"] != 0 {
		t.Errorf("Expected class scopes to be excluded from function distribution, got %d", report.Summary.ComplexityDistribution["11-20"])
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

	t.Run("TieBreakBySourceLocation", func(t *testing.T) {
		tied := []ComplexityResult{
			&mockComplexityResult{
				complexity:     7,
				functionName:   "duplicate",
				scopeKind:      domain.AnalysisScopeFunction,
				sourceLocation: ComplexitySourceLocation{FilePath: "z.py", StartLine: 1},
				riskLevel:      "medium",
			},
			&mockComplexityResult{
				complexity:     7,
				functionName:   "duplicate",
				scopeKind:      domain.AnalysisScopeFunction,
				sourceLocation: ComplexitySourceLocation{FilePath: "a.py", StartLine: 10},
				riskLevel:      "medium",
			},
			&mockComplexityResult{
				complexity:     7,
				functionName:   "duplicate",
				scopeKind:      domain.AnalysisScopeFunction,
				sourceLocation: ComplexitySourceLocation{FilePath: "a.py", StartLine: 2},
				riskLevel:      "medium",
			},
		}
		want := []ComplexitySourceLocation{
			{FilePath: "a.py", StartLine: 2},
			{FilePath: "a.py", StartLine: 10},
			{FilePath: "z.py", StartLine: 1},
		}

		for _, sortBy := range []string{"complexity", "risk", "name"} {
			cfg := config.DefaultConfig()
			cfg.Output.SortBy = sortBy
			var buffer bytes.Buffer
			reporter, err := NewComplexityReporter(cfg, &buffer)
			if err != nil {
				t.Fatalf("Failed to create reporter: %v", err)
			}

			sorted := reporter.filterAndSortResults(tied)
			for i, result := range sorted {
				if got := result.GetSourceLocation(); got != want[i] {
					t.Errorf("Sort %q location %d = %+v, want %+v", sortBy, i, got, want[i])
				}
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

func TestGenerateReportClassWarningKeepsScopeIdentity(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Complexity.MaxComplexity = 20
	reporter, err := NewComplexityReporter(cfg, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("create reporter: %v", err)
	}
	classScope := &mockComplexityResult{
		complexity:     25,
		functionName:   "Config",
		scopeKind:      domain.AnalysisScopeClass,
		sourceLocation: ComplexitySourceLocation{FilePath: "config.py", StartLine: 3},
		riskLevel:      "high",
	}

	report := reporter.GenerateReport([]ComplexityResult{classScope}, 1)
	if report.Summary.TotalFunctions != 0 || len(report.Results) != 0 || len(report.ClassScopes) != 1 {
		t.Fatalf("class scope changed legacy function populations: %+v", report)
	}
	if len(report.Warnings) != 2 {
		t.Fatalf("warnings = %+v", report.Warnings)
	}
	var warning ReportWarning
	for _, candidate := range report.Warnings {
		if candidate.Type == "max_complexity_exceeded" {
			warning = candidate
			break
		}
	}
	if warning.ScopeKind != domain.AnalysisScopeClass || warning.FilePath != "config.py" || warning.StartLine != 3 {
		t.Fatalf("class warning lost identity: %+v", warning)
	}
	if !strings.Contains(warning.Message, "Class scope complexity") || strings.Contains(warning.Message, "Function complexity") {
		t.Fatalf("class warning label = %q", warning.Message)
	}
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
	if len(report.Results) != 3 {
		t.Errorf("Expected 3 legacy results in JSON, got %d", len(report.Results))
	}
	if len(report.ClassScopes) != 1 {
		t.Errorf("Expected 1 class scope in JSON, got %d", len(report.ClassScopes))
	}
	if report.Summary.TotalFunctions != 3 {
		t.Errorf("Expected 3 total functions in JSON, got %d", report.Summary.TotalFunctions)
	}
	for _, result := range report.Results {
		if result.ScopeKind == domain.AnalysisScopeUnknown {
			t.Errorf("JSON result %q has no scope kind", result.FunctionName)
		}
	}
	if len(report.ClassScopes) == 1 && (report.ClassScopes[0].FilePath != "b.py" || report.ClassScopes[0].StartLine != 3) {
		t.Errorf("Expected JSON class source location, got %+v", report.ClassScopes[0])
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
	if len(report.Results) != 3 {
		t.Errorf("Expected 3 legacy results in YAML, got %d", len(report.Results))
	}
	if len(report.ClassScopes) != 1 {
		t.Errorf("Expected 1 class scope in YAML, got %d", len(report.ClassScopes))
	}
	if report.Summary.TotalFunctions != 3 {
		t.Errorf("Expected 3 total functions in YAML, got %d", report.Summary.TotalFunctions)
	}
	for _, result := range report.Results {
		if result.ScopeKind == domain.AnalysisScopeUnknown {
			t.Errorf("YAML result %q has no scope kind", result.FunctionName)
		}
	}
	if len(report.ClassScopes) == 1 && (report.ClassScopes[0].FilePath != "b.py" || report.ClassScopes[0].StartLine != 3) {
		t.Errorf("Expected YAML class source location, got %+v", report.ClassScopes[0])
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
		"If Statements", "Loop Statements", "Exception Handlers", "Scope Kind",
		"File Path", "Start Line", "Start Column", "End Line",
	}
	for i, field := range records[0] {
		if field != expectedHeader[i] {
			t.Errorf("Expected header field %s, got %s", expectedHeader[i], field)
		}
	}

	// Verify first data row (should be sorted by complexity descending)
	firstRow := records[1]
	if firstRow[0] != "very_complex_function" { // Sorted by complexity (highest first)
		t.Errorf("Expected first function to be very_complex_function, got %s", firstRow[0])
	}
	if firstRow[1] != "25" {
		t.Errorf("Expected complexity 25, got %s", firstRow[1])
	}
	for _, record := range records[1:] {
		if record[8] == "" {
			t.Errorf("CSV result %q has no scope kind", record[0])
		}
	}
	if firstRow[2] != "high" {
		t.Errorf("Expected risk high, got %s", firstRow[2])
	}
	if firstRow[9] != "b.py" || firstRow[10] != "30" || firstRow[11] != "1" || firstRow[12] != "64" {
		t.Errorf("Expected stable source location columns, got %v", firstRow[9:])
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
	if !strings.Contains(output, "Total Functions: 3") {
		t.Error("Missing total functions in summary")
	}
	if !strings.Contains(output, "Risk Distribution:") {
		t.Error("Missing risk distribution section")
	}
	if !strings.Contains(output, "Function Details:") {
		t.Error("Missing function details section")
	}
	if !strings.Contains(output, "Class Scope Details:") {
		t.Error("Missing additive class scope details")
	}
	if !strings.Contains(output, "complex_function (b.py:3:1)") {
		t.Error("Missing class source location")
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
			"3 functions analyzed",
			"Avg: 10.3",
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
				scopeKind:    domain.AnalysisScopeFunction,
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
