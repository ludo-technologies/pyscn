package domain

import (
	"testing"
)

func TestOutputFormat(t *testing.T) {
	tests := []struct {
		name   string
		format OutputFormat
		valid  bool
	}{
		{"Text format", OutputFormatText, true},
		{"JSON format", OutputFormatJSON, true},
		{"YAML format", OutputFormatYAML, true},
		{"CSV format", OutputFormatCSV, true},
		{"Invalid format", OutputFormat("invalid"), false},
	}

	validFormats := map[OutputFormat]bool{
		OutputFormatText: true,
		OutputFormatJSON: true,
		OutputFormatYAML: true,
		OutputFormatCSV:  true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, exists := validFormats[tt.format]
			if exists != tt.valid {
				t.Errorf("Format %s validity = %v, want %v", tt.format, exists, tt.valid)
			}
		})
	}
}

func TestSortCriteria(t *testing.T) {
	tests := []struct {
		name     string
		criteria SortCriteria
		valid    bool
	}{
		{"Sort by complexity", SortByComplexity, true},
		{"Sort by name", SortByName, true},
		{"Sort by risk", SortByRisk, true},
		{"Invalid criteria", SortCriteria("invalid"), false},
	}

	validCriteria := map[SortCriteria]bool{
		SortByComplexity: true,
		SortByName:       true,
		SortByRisk:       true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, exists := validCriteria[tt.criteria]
			if exists != tt.valid {
				t.Errorf("Criteria %s validity = %v, want %v", tt.criteria, exists, tt.valid)
			}
		})
	}
}

func TestRiskLevel(t *testing.T) {
	tests := []struct {
		name  string
		level RiskLevel
		valid bool
	}{
		{"Low risk", RiskLevelLow, true},
		{"Medium risk", RiskLevelMedium, true},
		{"High risk", RiskLevelHigh, true},
		{"Invalid risk", RiskLevel("invalid"), false},
	}

	validLevels := map[RiskLevel]bool{
		RiskLevelLow:    true,
		RiskLevelMedium: true,
		RiskLevelHigh:   true,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, exists := validLevels[tt.level]
			if exists != tt.valid {
				t.Errorf("Risk level %s validity = %v, want %v", tt.level, exists, tt.valid)
			}
		})
	}
}

func TestComplexityMetrics(t *testing.T) {
	metrics := ComplexityMetrics{
		Complexity:        5,
		Nodes:            10,
		Edges:            12,
		IfStatements:     2,
		LoopStatements:   1,
		ExceptionHandlers: 1,
		SwitchCases:      0,
	}

	if metrics.Complexity != 5 {
		t.Errorf("Expected complexity 5, got %d", metrics.Complexity)
	}

	if metrics.Nodes != 10 {
		t.Errorf("Expected nodes 10, got %d", metrics.Nodes)
	}

	if metrics.Edges != 12 {
		t.Errorf("Expected edges 12, got %d", metrics.Edges)
	}
}

func TestFunctionComplexity(t *testing.T) {
	function := FunctionComplexity{
		Name:     "testFunction",
		FilePath: "/path/to/test.py",
		Metrics: ComplexityMetrics{
			Complexity: 3,
			Nodes:     5,
			Edges:     6,
		},
		RiskLevel: RiskLevelLow,
	}

	if function.Name != "testFunction" {
		t.Errorf("Expected function name 'testFunction', got %s", function.Name)
	}

	if function.RiskLevel != RiskLevelLow {
		t.Errorf("Expected risk level low, got %s", function.RiskLevel)
	}
}

func TestComplexitySummary(t *testing.T) {
	summary := ComplexitySummary{
		TotalFunctions:      10,
		AverageComplexity:   2.5,
		MaxComplexity:       8,
		MinComplexity:       1,
		FilesAnalyzed:       3,
		LowRiskFunctions:    8,
		MediumRiskFunctions: 2,
		HighRiskFunctions:   0,
		ComplexityDistribution: map[string]int{
			"1":   5,
			"2-5": 4,
			"6-10": 1,
		},
	}

	if summary.TotalFunctions != 10 {
		t.Errorf("Expected total functions 10, got %d", summary.TotalFunctions)
	}

	if summary.AverageComplexity != 2.5 {
		t.Errorf("Expected average complexity 2.5, got %f", summary.AverageComplexity)
	}

	expectedDistSum := 5 + 4 + 1
	actualDistSum := 0
	for _, count := range summary.ComplexityDistribution {
		actualDistSum += count
	}

	if actualDistSum != expectedDistSum {
		t.Errorf("Expected complexity distribution sum %d, got %d", expectedDistSum, actualDistSum)
	}
}

func TestComplexityResponse(t *testing.T) {
	response := ComplexityResponse{
		Functions: []FunctionComplexity{
			{
				Name: "func1",
				Metrics: ComplexityMetrics{
					Complexity: 1,
				},
				RiskLevel: RiskLevelLow,
			},
			{
				Name: "func2",
				Metrics: ComplexityMetrics{
					Complexity: 5,
				},
				RiskLevel: RiskLevelLow,
			},
		},
		Summary: ComplexitySummary{
			TotalFunctions:    2,
			AverageComplexity: 3.0,
		},
		Warnings:    []string{"warning1"},
		Errors:      []string{},
		GeneratedAt: "2025-01-01T00:00:00Z",
		Version:     "test",
	}

	if len(response.Functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(response.Functions))
	}

	if len(response.Warnings) != 1 {
		t.Errorf("Expected 1 warning, got %d", len(response.Warnings))
	}

	if len(response.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(response.Errors))
	}

	if response.Summary.TotalFunctions != 2 {
		t.Errorf("Expected summary total functions 2, got %d", response.Summary.TotalFunctions)
	}
}