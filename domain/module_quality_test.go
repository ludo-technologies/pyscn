package domain

import (
	"encoding/json"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
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

func TestAggregateComplexityByModule_ExcludesModuleScopeFromFunctionAverages(t *testing.T) {
	modules := AggregateComplexityByModule([]FunctionComplexity{
		{
			Name:      ModuleFunctionName,
			FilePath:  "pkg/module.py",
			Metrics:   ComplexityMetrics{Complexity: 1},
			RiskLevel: RiskLevelLow,
		},
		{
			Name:      "onlyFunction",
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

	payload, err := json.Marshal(metric)
	if err != nil {
		t.Fatalf("marshal dependency metric: %v", err)
	}
	var contract map[string]any
	if err := json.Unmarshal(payload, &contract); err != nil {
		t.Fatalf("decode dependency metric: %v", err)
	}
	if contract["average_complexity"] != 6.5 || contract["dead_code_block_count"] != float64(4) {
		t.Fatalf("expected shared metric field names in system output, got %s", payload)
	}

	yamlPayload, err := yaml.Marshal(metric)
	if err != nil {
		t.Fatalf("marshal dependency metric as YAML: %v", err)
	}
	var yamlContract map[string]any
	if err := yaml.Unmarshal(yamlPayload, &yamlContract); err != nil {
		t.Fatalf("decode dependency metric YAML: %v", err)
	}
	if yamlContract["average_complexity"] != 6.5 || yamlContract["dead_code_block_count"] != 4 {
		t.Fatalf("expected inline shared metrics in system YAML, got %s", yamlPayload)
	}
	if _, nested := yamlContract["modulecomplexitymetrics"]; nested {
		t.Fatalf("unexpected nested complexity contract in system YAML: %s", yamlPayload)
	}
}
