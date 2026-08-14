package domain

import (
	"testing"
)

func TestAggregateComplexityByModule_GroupsUnfilteredFunctions(t *testing.T) {
	modules := AggregateComplexityByModule([]FunctionComplexity{
		{
			Name:      "hotPath",
			ScopeKind: AnalysisScopeFunction,
			FilePath:  "pkg/hot.py",
			Metrics: ComplexityMetrics{
				Complexity:          7,
				CognitiveComplexity: 9,
				ExceptionHandlers:   1,
			},
			RiskLevel: RiskLevelHigh,
		},
		{
			Name:      "warmPath",
			ScopeKind: AnalysisScopeFunction,
			FilePath:  "pkg/hot.py",
			Metrics: ComplexityMetrics{
				Complexity:          3,
				CognitiveComplexity: 5,
				ExceptionHandlers:   2,
			},
			RiskLevel: RiskLevelMedium,
		},
		{
			Name:      "simplePath",
			ScopeKind: AnalysisScopeFunction,
			FilePath:  "pkg/simple.py",
			Metrics:   ComplexityMetrics{Complexity: 1},
			RiskLevel: RiskLevelLow,
		},
	})

	if len(modules) != 2 {
		t.Fatalf("expected 2 module-quality entries, got %d", len(modules))
	}

	hot := modules["pkg/hot.py"]
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

func TestAggregateComplexityByModule_ExcludesModuleScopeFromFunctionAverages(t *testing.T) {
	modules := AggregateComplexityByModule([]FunctionComplexity{
		{
			Name:      ModuleFunctionName,
			ScopeKind: AnalysisScopeModule,
			FilePath:  "pkg/module.py",
			Metrics:   ComplexityMetrics{Complexity: 1},
			RiskLevel: RiskLevelLow,
		},
		{
			Name:      "onlyFunction",
			ScopeKind: AnalysisScopeFunction,
			FilePath:  "pkg/module.py",
			Metrics:   ComplexityMetrics{Complexity: 5, CognitiveComplexity: 7},
			RiskLevel: RiskLevelMedium,
		},
	})

	module := modules["pkg/module.py"]
	if module.AnalyzedFunctionCount != 1 || module.AverageComplexity != 5 || module.AverageCognitiveComplexity != 7 {
		t.Fatalf("expected function-only rollup, got %+v", module)
	}
}

func TestAggregateComplexityByModule_ExcludesClassScopes(t *testing.T) {
	modules := AggregateComplexityByModule([]FunctionComplexity{
		{Name: "work", ScopeKind: AnalysisScopeFunction, FilePath: "pkg/a.py", Metrics: ComplexityMetrics{Complexity: 5}},
		{Name: "Config", ScopeKind: AnalysisScopeClass, FilePath: "pkg/a.py", Metrics: ComplexityMetrics{Complexity: 20}},
	})

	metrics := modules["pkg/a.py"]
	if metrics.AnalyzedFunctionCount != 1 || metrics.AverageComplexity != 5 || metrics.MaxComplexity != 5 {
		t.Fatalf("class scope changed function rollup: %+v", metrics)
	}
}

func TestAggregateDeadCodeByModule_UsesOneUnfilteredPopulation(t *testing.T) {
	modules := AggregateDeadCodeByModule([]FileDeadCode{
		{
			FilePath:      "pkg/hot.py",
			TotalFindings: 4,
			Functions: []FunctionDeadCode{
				{Findings: []DeadCodeFinding{{BlockID: "a"}, {BlockID: "a"}}},
				{Findings: []DeadCodeFinding{{BlockID: "b"}, {BlockID: "c"}}},
			},
		},
	})

	if len(modules) != 1 {
		t.Fatalf("expected 1 module-quality entry, got %d", len(modules))
	}
	hot := modules["pkg/hot.py"]
	if hot.DeadCodeFindingCount != 4 {
		t.Errorf("expected 4 dead-code findings, got %d", hot.DeadCodeFindingCount)
	}
	if hot.DeadCodeBlockCount != 3 {
		t.Errorf("expected 3 dead-code blocks, got %d", hot.DeadCodeBlockCount)
	}
}
