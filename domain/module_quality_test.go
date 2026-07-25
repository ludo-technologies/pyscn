package domain

import (
	"path/filepath"
	"testing"
)

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

func TestAggregateModuleQuality_JoinsModuleIdentityAndDeadCode(t *testing.T) {
	absoluteHotPath, err := filepath.Abs("pkg/hot.py")
	if err != nil {
		t.Fatalf("resolve test path: %v", err)
	}
	absoluteQuietPath, err := filepath.Abs("pkg/quiet.py")
	if err != nil {
		t.Fatalf("resolve test path: %v", err)
	}

	response := &AnalyzeResponse{
		Complexity: &ComplexityResponse{
			Functions: []FunctionComplexity{
				{
					Name:      "hotPath",
					FilePath:  "pkg/hot.py",
					Metrics:   ComplexityMetrics{Complexity: 8},
					RiskLevel: RiskLevelHigh,
				},
			},
		},
		DeadCode: &DeadCodeResponse{
			Files: []FileDeadCode{
				{
					FilePath:      "pkg/hot.py",
					TotalFindings: 3,
					Functions: []FunctionDeadCode{
						{DeadBlocks: 2},
						{DeadBlocks: 1},
					},
				},
			},
		},
		System: &SystemAnalysisResponse{
			DependencyAnalysis: &DependencyAnalysisResult{
				ModuleMetrics: map[string]*ModuleDependencyMetrics{
					"pkg.hot": {
						ModuleName:    "pkg.hot",
						FilePath:      absoluteHotPath,
						LinesOfCode:   120,
						FunctionCount: 4,
					},
					"pkg.quiet": {
						ModuleName:    "pkg.quiet",
						FilePath:      absoluteQuietPath,
						LinesOfCode:   12,
						FunctionCount: 0,
					},
				},
			},
		},
	}

	modules := AggregateModuleQuality(response)
	if len(modules) != 2 {
		t.Fatalf("expected system analysis to retain 2 modules, got %d", len(modules))
	}

	hot := modules[0]
	if hot.ModuleName != "pkg.hot" {
		t.Errorf("expected module identity pkg.hot, got %q", hot.ModuleName)
	}
	if hot.FilePath != "pkg/hot.py" {
		t.Errorf("expected analysis path to remain user-facing, got %q", hot.FilePath)
	}
	if hot.LinesOfCode != 120 {
		t.Errorf("expected 120 lines of code, got %d", hot.LinesOfCode)
	}
	if hot.FunctionCount != 4 {
		t.Errorf("expected 4 module functions, got %d", hot.FunctionCount)
	}
	if hot.DeadCodeFindingCount != 3 {
		t.Errorf("expected 3 dead-code findings, got %d", hot.DeadCodeFindingCount)
	}
	if hot.DeadCodeBlockCount != 3 {
		t.Errorf("expected 3 dead-code blocks, got %d", hot.DeadCodeBlockCount)
	}

	quiet := modules[1]
	if quiet.ModuleName != "pkg.quiet" {
		t.Errorf("expected zero-finding module to remain visible, got %q", quiet.ModuleName)
	}
}
