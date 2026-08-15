package domain

import (
	"path/filepath"
)

// ModuleComplexityMetrics is the canonical module-level function-complexity
// contract. The <module> pseudo-record and executable class suites are excluded.
type ModuleComplexityMetrics struct {
	AnalyzedFunctionCount      int     `json:"analyzed_function_count" yaml:"analyzed_function_count"`
	AverageComplexity          float64 `json:"average_complexity" yaml:"average_complexity"`
	AverageCognitiveComplexity float64 `json:"average_cognitive_complexity" yaml:"average_cognitive_complexity"`
	MaxComplexity              int     `json:"max_complexity" yaml:"max_complexity"`
	HighRiskFunctionCount      int     `json:"high_risk_function_count" yaml:"high_risk_function_count"`
	ExceptionHandlerCount      int     `json:"exception_handler_count" yaml:"exception_handler_count"`
}

// ModuleDeadCodeMetrics is the canonical module-level dead-code contract.
// Both counts describe findings enabled by detector options before severity filtering.
type ModuleDeadCodeMetrics struct {
	DeadCodeFindingCount int `json:"dead_code_finding_count" yaml:"dead_code_finding_count"`
	DeadCodeBlockCount   int `json:"dead_code_block_count" yaml:"dead_code_block_count"`
}

// ModuleQualityMetrics is the public per-file view assembled by unified analysis.
type ModuleQualityMetrics struct {
	ModuleName              string `json:"module_name,omitempty" yaml:"module_name,omitempty"`
	FilePath                string `json:"file_path" yaml:"file_path"`
	LinesOfCode             int    `json:"lines_of_code" yaml:"lines_of_code"`
	FunctionCount           int    `json:"function_count" yaml:"function_count"`
	ModuleComplexityMetrics `yaml:",inline"`
	ModuleDeadCodeMetrics   `yaml:",inline"`
}

type moduleComplexityAccumulator struct {
	metrics                  ModuleComplexityMetrics
	totalComplexity          int
	totalCognitiveComplexity int
}

// AggregateComplexityByModule derives module metrics from the complete,
// pre-filter complexity population owned by the complexity service.
func AggregateComplexityByModule(functions []FunctionComplexity) map[string]ModuleComplexityMetrics {
	modules := make(map[string]*moduleComplexityAccumulator)
	for _, function := range functions {
		key := filepath.Clean(function.FilePath)
		module := modules[key]
		if module == nil {
			module = &moduleComplexityAccumulator{}
			modules[key] = module
		}
		if function.ScopeKind != AnalysisScopeFunction {
			continue
		}

		module.metrics.AnalyzedFunctionCount++
		module.totalComplexity += function.Metrics.Complexity
		module.totalCognitiveComplexity += function.Metrics.CognitiveComplexity
		module.metrics.ExceptionHandlerCount += function.Metrics.ExceptionHandlers
		if function.Metrics.Complexity > module.metrics.MaxComplexity {
			module.metrics.MaxComplexity = function.Metrics.Complexity
		}
		if function.RiskLevel == RiskLevelHigh {
			module.metrics.HighRiskFunctionCount++
		}
	}

	result := make(map[string]ModuleComplexityMetrics, len(modules))
	for path, module := range modules {
		if module.metrics.AnalyzedFunctionCount > 0 {
			count := float64(module.metrics.AnalyzedFunctionCount)
			module.metrics.AverageComplexity = float64(module.totalComplexity) / count
			module.metrics.AverageCognitiveComplexity = float64(module.totalCognitiveComplexity) / count
		}
		result[path] = module.metrics
	}
	return result
}

// AggregateDeadCodeByModule derives module metrics from the complete,
// pre-severity-filter dead-code population owned by the dead-code service.
func AggregateDeadCodeByModule(files []FileDeadCode) map[string]ModuleDeadCodeMetrics {
	modules := make(map[string]ModuleDeadCodeMetrics, len(files))
	for _, file := range files {
		key := filepath.Clean(file.FilePath)
		module := modules[key]
		module.DeadCodeFindingCount += file.TotalFindings
		for _, scope := range file.ExecutionScopes() {
			blockIDs := make(map[string]struct{}, len(scope.Findings))
			for _, finding := range scope.Findings {
				if finding.BlockID != "" {
					blockIDs[finding.BlockID] = struct{}{}
				}
			}
			module.DeadCodeBlockCount += len(blockIDs)
		}
		modules[key] = module
	}
	return modules
}
