package app

import (
	"fmt"
	"sort"

	"github.com/ludo-technologies/pyscn/domain"
)

// aggregateModuleQuality joins analyzer-owned rollups at the use-case seam.
// Canonical paths are used only as internal identities; the first caller-facing
// path remains reported so unified output preserves the user's path style.
func aggregateModuleQuality(response *domain.AnalyzeResponse, pathIndex analysisPathIndex) ([]domain.ModuleQualityMetrics, error) {
	if response == nil {
		return nil, nil
	}

	modules := make(map[string]*domain.ModuleQualityMetrics)
	if response.Complexity != nil {
		for filePath, complexity := range response.Complexity.ModuleRollups {
			module, err := moduleQualityEntry(modules, pathIndex, filePath)
			if err != nil {
				return nil, err
			}
			module.ModuleComplexityMetrics = complexity
		}
	}

	if response.DeadCode != nil {
		for filePath, deadCode := range response.DeadCode.ModuleRollups {
			module, err := moduleQualityEntry(modules, pathIndex, filePath)
			if err != nil {
				return nil, err
			}
			module.ModuleDeadCodeMetrics = deadCode
		}
	}

	if response.System != nil && response.System.DependencyAnalysis != nil {
		for moduleName, metadata := range response.System.DependencyAnalysis.ModuleMetrics {
			if metadata == nil {
				continue
			}

			module, err := moduleQualityEntry(modules, pathIndex, metadata.FilePath)
			if err != nil {
				return nil, err
			}
			module.ModuleName = metadata.ModuleName
			if module.ModuleName == "" {
				module.ModuleName = moduleName
			}
			module.LinesOfCode = metadata.LinesOfCode
			module.FunctionCount = metadata.FunctionCount
		}
	}

	result := make([]domain.ModuleQualityMetrics, 0, len(modules))
	for _, module := range modules {
		result = append(result, *module)
	}
	sortModuleQuality(result)
	return result, nil
}

func moduleQualityEntry(modules map[string]*domain.ModuleQualityMetrics, pathIndex analysisPathIndex, path string) (*domain.ModuleQualityMetrics, error) {
	identity, err := analysisPathIdentity(path)
	if err != nil {
		return nil, fmt.Errorf("resolve module path %q: %w", path, err)
	}
	module := modules[identity]
	if module == nil {
		reportedPath, err := pathIndex.reportedPath(path)
		if err != nil {
			return nil, fmt.Errorf("resolve reported module path %q: %w", path, err)
		}
		module = &domain.ModuleQualityMetrics{FilePath: reportedPath}
		modules[identity] = module
	}
	return module, nil
}

func sortModuleQuality(modules []domain.ModuleQualityMetrics) {
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
