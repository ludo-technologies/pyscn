package domain

import "testing"

func TestAggregateModuleQuality_GroupsComplexityByFile(t *testing.T) {
	response := &AnalyzeResponse{
		Complexity: &ComplexityResponse{
			Functions: []FunctionComplexity{
				{
					Name:     "hotPath",
					FilePath: "pkg/hot.py",
					Metrics: ComplexityMetrics{
						Complexity:          7,
						CognitiveComplexity: 9,
						ExceptionHandlers:   1,
					},
					RiskLevel: RiskLevelHigh,
				},
				{
					Name:     "warmPath",
					FilePath: "pkg/hot.py",
					Metrics: ComplexityMetrics{
						Complexity:          3,
						CognitiveComplexity: 5,
						ExceptionHandlers:   2,
					},
					RiskLevel: RiskLevelMedium,
				},
				{
					Name:     "simplePath",
					FilePath: "pkg/simple.py",
					Metrics: ComplexityMetrics{
						Complexity:          1,
						CognitiveComplexity: 0,
					},
					RiskLevel: RiskLevelLow,
				},
			},
		},
	}

	modules := AggregateModuleQuality(response)
	if len(modules) != 2 {
		t.Fatalf("expected 2 module-quality entries, got %d", len(modules))
	}

	hot := modules[0]
	if hot.FilePath != "pkg/hot.py" {
		t.Fatalf("expected hottest module first, got %q", hot.FilePath)
	}
	if hot.AnalyzedFunctionCount != 2 {
		t.Errorf("expected 2 analyzed functions, got %d", hot.AnalyzedFunctionCount)
	}
	if hot.AverageComplexity != 5 {
		t.Errorf("expected average complexity 5, got %v", hot.AverageComplexity)
	}
	if hot.AverageCognitiveComplexity != 7 {
		t.Errorf("expected average cognitive complexity 7, got %v", hot.AverageCognitiveComplexity)
	}
	if hot.MaxComplexity != 7 {
		t.Errorf("expected max complexity 7, got %d", hot.MaxComplexity)
	}
	if hot.HighRiskFunctionCount != 1 {
		t.Errorf("expected 1 high-risk function, got %d", hot.HighRiskFunctionCount)
	}
	if hot.ExceptionHandlerCount != 3 {
		t.Errorf("expected 3 exception handlers, got %d", hot.ExceptionHandlerCount)
	}
}
