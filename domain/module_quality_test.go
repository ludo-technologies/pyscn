package domain

import (
	"path/filepath"
	"testing"
)

func TestAggregateComplexityByModule_GroupsUnfilteredFunctions(t *testing.T) {
	modules := AggregateComplexityByModule([]FunctionComplexity{
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
			Name:      "simplePath",
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

func TestAggregateDeadCodeByModule_UsesOneUnfilteredPopulation(t *testing.T) {
	modules := AggregateDeadCodeByModule([]FileDeadCode{
		{
			FilePath:      "pkg/hot.py",
			TotalFindings: 3,
			Functions: []FunctionDeadCode{
				{DeadBlocks: 2},
				{DeadBlocks: 1},
			},
		},
	})

	if len(modules) != 1 {
		t.Fatalf("expected 1 module-quality entry, got %d", len(modules))
	}
	hot := modules["pkg/hot.py"]
	if hot.DeadCodeFindingCount != 3 {
		t.Errorf("expected 3 dead-code findings, got %d", hot.DeadCodeFindingCount)
	}
	if hot.DeadCodeBlockCount != 3 {
		t.Errorf("expected 3 dead-code blocks, got %d", hot.DeadCodeBlockCount)
	}
}

func TestAggregateModuleQuality_JoinsRootedModuleIdentity(t *testing.T) {
	projectRoot := t.TempDir()
	absoluteHotPath := filepath.Join(projectRoot, "pkg", "hot.py")
	absoluteQuietPath := filepath.Join(projectRoot, "pkg", "quiet.py")

	response := &AnalyzeResponse{
		Complexity: &ComplexityResponse{
			ModuleRollups: map[string]ModuleComplexityMetrics{
				absoluteHotPath: {
					AnalyzedFunctionCount: 1,
					AverageComplexity:     8,
					MaxComplexity:         8,
					HighRiskFunctionCount: 1,
				},
			},
		},
		DeadCode: &DeadCodeResponse{
			ModuleRollups: map[string]ModuleDeadCodeMetrics{
				absoluteHotPath: {
					DeadCodeFindingCount: 3,
					DeadCodeBlockCount:   3,
				},
			},
		},
		System: &SystemAnalysisResponse{
			Summary: SystemAnalysisSummary{ProjectRoot: projectRoot},
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
	if hot.FilePath != absoluteHotPath {
		t.Errorf("expected canonical analysis path, got %q", hot.FilePath)
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

func TestApplyModuleQualityToSystem_PublishesQualityOnDependencyMetrics(t *testing.T) {
	projectRoot := t.TempDir()
	metric := &ModuleDependencyMetrics{FilePath: filepath.Join(projectRoot, "pkg", "hot.py")}
	system := &SystemAnalysisResponse{
		Summary: SystemAnalysisSummary{ProjectRoot: projectRoot},
		DependencyAnalysis: &DependencyAnalysisResult{
			ModuleMetrics: map[string]*ModuleDependencyMetrics{"pkg.hot": metric},
		},
	}
	quality := []ModuleQualityMetrics{
		{
			FilePath: filepath.Join(projectRoot, "pkg", "hot.py"),
			ModuleComplexityMetrics: ModuleComplexityMetrics{
				AnalyzedFunctionCount:      2,
				AverageComplexity:          6.5,
				AverageCognitiveComplexity: 8,
				MaxComplexity:              9,
				HighRiskFunctionCount:      1,
				ExceptionHandlerCount:      3,
			},
			ModuleDeadCodeMetrics: ModuleDeadCodeMetrics{
				DeadCodeFindingCount: 2,
				DeadCodeBlockCount:   4,
			},
		},
	}

	ApplyModuleQualityToSystem(system, quality)

	if metric.AverageComplexity != 6.5 || metric.AverageCognitiveComplexity != 8 {
		t.Fatalf("expected complexity averages to be published, got %+v", metric)
	}
	if metric.MaxComplexity != 9 || metric.HighRiskFunctionCount != 1 {
		t.Fatalf("expected complexity hotspot fields to be published, got %+v", metric)
	}
	if metric.ExceptionHandlerCount != 3 || metric.DeadCodeFindingCount != 2 || metric.DeadCodeBlockCount != 4 {
		t.Fatalf("expected exception and dead-code fields to be published, got %+v", metric)
	}
}
