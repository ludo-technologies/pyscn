package domain

import (
	"path/filepath"
	"sort"
)

// ModuleComplexityMetrics is the canonical module-level complexity contract.
// AnalyzedFunctionCount includes every function-level complexity record before
// presentation filters. The <module> pseudo-function is intentionally excluded.
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
// Metric values are embedded from the same contracts stored on dependency metrics.
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
		if function.Name == ModuleFunctionName {
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
		for _, function := range file.Functions {
			blockIDs := make(map[string]struct{}, len(function.Findings))
			for _, finding := range function.Findings {
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

// AggregateModuleQuality combines analyzer-owned, pre-filter rollups with
// structural module metadata. Unified analysis gives every task the same
// canonical absolute paths, so joins use exact cleaned identities.
func AggregateModuleQuality(response *AnalyzeResponse) []ModuleQualityMetrics {
	if response == nil {
		return nil
	}

	modules := make(map[string]*ModuleQualityMetrics)
	if response.System != nil && response.System.DependencyAnalysis != nil {
		for moduleName, metadata := range response.System.DependencyAnalysis.ModuleMetrics {
			if metadata == nil {
				continue
			}

			module := getModuleQualityMetrics(modules, metadata.FilePath)
			module.ModuleName = metadata.ModuleName
			if module.ModuleName == "" {
				module.ModuleName = moduleName
			}
			module.LinesOfCode = metadata.LinesOfCode
			module.FunctionCount = metadata.FunctionCount
		}
	}

	if response.Complexity != nil {
		for filePath, complexity := range response.Complexity.ModuleRollups {
			module := getModuleQualityMetrics(modules, filePath)
			module.ModuleComplexityMetrics = complexity
		}
	}

	if response.DeadCode != nil {
		for filePath, deadCode := range response.DeadCode.ModuleRollups {
			module := getModuleQualityMetrics(modules, filePath)
			module.ModuleDeadCodeMetrics = deadCode
		}
	}

	result := make([]ModuleQualityMetrics, 0, len(modules))
	for _, module := range modules {
		result = append(result, *module)
	}
	sortModuleQuality(result)
	return result
}

// ApplyModuleQualityToSystem publishes the unified rollup on the existing
// dependency metrics without coupling the parallel analyzer tasks together.
func ApplyModuleQualityToSystem(system *SystemAnalysisResponse, quality []ModuleQualityMetrics) {
	if system == nil || system.DependencyAnalysis == nil {
		return
	}

	byPath := make(map[string]ModuleQualityMetrics, len(quality))
	for _, module := range quality {
		byPath[filepath.Clean(module.FilePath)] = module
	}

	for _, module := range system.DependencyAnalysis.ModuleMetrics {
		if module == nil {
			continue
		}
		qualityMetric, ok := byPath[filepath.Clean(module.FilePath)]
		if !ok {
			continue
		}
		module.ModuleComplexityMetrics = qualityMetric.ModuleComplexityMetrics
		module.ModuleDeadCodeMetrics = qualityMetric.ModuleDeadCodeMetrics
	}
}

func getModuleQualityMetrics(modules map[string]*ModuleQualityMetrics, path string) *ModuleQualityMetrics {
	key := filepath.Clean(path)
	module := modules[key]
	if module == nil {
		module = &ModuleQualityMetrics{FilePath: path}
		modules[key] = module
	}
	return module
}

func sortModuleQuality(modules []ModuleQualityMetrics) {
	sort.Slice(modules, func(i, j int) bool {
		left, right := modules[i], modules[j]
		if left.HighRiskFunctionCount != right.HighRiskFunctionCount {
			return left.HighRiskFunctionCount > right.HighRiskFunctionCount
		}
		if left.MaxComplexity != right.MaxComplexity {
			return left.MaxComplexity > right.MaxComplexity
		}
		if left.AverageComplexity != right.AverageComplexity {
			return left.AverageComplexity > right.AverageComplexity
		}
		if left.DeadCodeFindingCount != right.DeadCodeFindingCount {
			return left.DeadCodeFindingCount > right.DeadCodeFindingCount
		}
		return left.FilePath < right.FilePath
	})
}
