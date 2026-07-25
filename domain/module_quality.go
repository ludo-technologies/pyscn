package domain

import (
	"path/filepath"
	"sort"
)

// ModuleQualityMetrics summarizes function-level quality signals for one Python module.
// AnalyzedFunctionCount reflects functions present in the complexity response after filtering.
type ModuleQualityMetrics struct {
	ModuleName                 string  `json:"module_name,omitempty" yaml:"module_name,omitempty"`
	FilePath                   string  `json:"file_path" yaml:"file_path"`
	LinesOfCode                int     `json:"lines_of_code" yaml:"lines_of_code"`
	FunctionCount              int     `json:"function_count" yaml:"function_count"`
	AnalyzedFunctionCount      int     `json:"analyzed_function_count" yaml:"analyzed_function_count"`
	AverageComplexity          float64 `json:"average_complexity" yaml:"average_complexity"`
	AverageCognitiveComplexity float64 `json:"average_cognitive_complexity" yaml:"average_cognitive_complexity"`
	MaxComplexity              int     `json:"max_complexity" yaml:"max_complexity"`
	HighRiskFunctionCount      int     `json:"high_risk_function_count" yaml:"high_risk_function_count"`
	ExceptionHandlerCount      int     `json:"exception_handler_count" yaml:"exception_handler_count"`
	DeadCodeFindingCount       int     `json:"dead_code_finding_count" yaml:"dead_code_finding_count"`
	DeadCodeBlockCount         int     `json:"dead_code_block_count" yaml:"dead_code_block_count"`
}

type moduleQualityAccumulator struct {
	metrics                  ModuleQualityMetrics
	totalComplexity          int
	totalCognitiveComplexity int
}

// AggregateModuleQuality combines completed analyzer results into per-module metrics.
// Results are ordered by high-risk count, maximum complexity, average complexity,
// and file path so the most actionable modules appear first deterministically.
func AggregateModuleQuality(response *AnalyzeResponse) []ModuleQualityMetrics {
	if response == nil {
		return nil
	}

	modules := make(map[string]*moduleQualityAccumulator)
	if response.System != nil && response.System.DependencyAnalysis != nil {
		for moduleName, metadata := range response.System.DependencyAnalysis.ModuleMetrics {
			if metadata == nil {
				continue
			}

			module := getModuleQualityAccumulator(modules, metadata.FilePath)
			module.metrics.ModuleName = metadata.ModuleName
			if module.metrics.ModuleName == "" {
				module.metrics.ModuleName = moduleName
			}
			module.metrics.LinesOfCode = metadata.LinesOfCode
			module.metrics.FunctionCount = metadata.FunctionCount
		}
	}

	if response.Complexity != nil {
		for _, function := range response.Complexity.Functions {
			module := getModuleQualityAccumulator(modules, function.FilePath)
			module.metrics.FilePath = function.FilePath
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
	}

	if response.DeadCode != nil {
		for _, file := range response.DeadCode.Files {
			module := getModuleQualityAccumulator(modules, file.FilePath)
			module.metrics.FilePath = file.FilePath
			module.metrics.DeadCodeFindingCount += file.TotalFindings
			for _, function := range file.Functions {
				module.metrics.DeadCodeBlockCount += function.DeadBlocks
			}
		}
	}

	result := make([]ModuleQualityMetrics, 0, len(modules))
	for _, module := range modules {
		if module.metrics.AnalyzedFunctionCount > 0 {
			count := float64(module.metrics.AnalyzedFunctionCount)
			module.metrics.AverageComplexity = float64(module.totalComplexity) / count
			module.metrics.AverageCognitiveComplexity = float64(module.totalCognitiveComplexity) / count
		}
		result = append(result, module.metrics)
	}

	sort.Slice(result, func(i, j int) bool {
		left, right := result[i], result[j]
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

	return result
}

func getModuleQualityAccumulator(modules map[string]*moduleQualityAccumulator, path string) *moduleQualityAccumulator {
	key := moduleQualityPathKey(path)
	module := modules[key]
	if module == nil {
		module = &moduleQualityAccumulator{
			metrics: ModuleQualityMetrics{FilePath: path},
		}
		modules[key] = module
	}
	return module
}

func moduleQualityPathKey(path string) string {
	cleaned := filepath.Clean(path)
	absolute, err := filepath.Abs(cleaned)
	if err != nil {
		return cleaned
	}
	return absolute
}
