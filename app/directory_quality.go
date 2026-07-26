package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

// directoryQualityRoot returns the common directory explicitly selected by the
// caller. Directory inputs participate directly; file inputs participate via
// their parent directory. It does not widen the scope by searching ancestors.
func directoryQualityRoot(paths []string) (string, error) {
	if len(paths) == 0 {
		return "", fmt.Errorf("at least one analysis path is required")
	}

	var root string
	for _, path := range paths {
		identity, err := analysisPathIdentity(path)
		if err != nil {
			return "", fmt.Errorf("resolve analysis path %q: %w", path, err)
		}
		info, err := os.Stat(identity)
		if err != nil {
			return "", fmt.Errorf("inspect analysis path %q: %w", path, err)
		}
		if !info.IsDir() {
			identity = filepath.Dir(identity)
		}

		if root == "" {
			root = identity
			continue
		}
		root = commonDirectory(root, identity)
	}
	return root, nil
}

func commonDirectory(left, right string) string {
	for {
		relative, err := filepath.Rel(left, right)
		if err == nil && !pathEscapesRoot(relative) {
			return left
		}
		parent := filepath.Dir(left)
		if parent == left {
			return left
		}
		left = parent
	}
}

type directoryQualityAccumulator struct {
	metrics                  domain.DirectoryQualityMetrics
	totalComplexity          float64
	totalCognitiveComplexity float64
}

// aggregateDirectoryQuality folds the canonical per-module quality contract
// into recursive, project-root-relative directory rollups. It is the only
// owner of directory grouping and weighted directory metrics.
func aggregateDirectoryQuality(modules []domain.ModuleQualityMetrics, projectRoot string) ([]domain.DirectoryQualityMetrics, error) {
	if len(modules) == 0 {
		return nil, nil
	}

	rootIdentity, err := analysisPathIdentity(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve project root %q: %w", projectRoot, err)
	}

	directories := make(map[string]*directoryQualityAccumulator)
	for _, module := range modules {
		moduleIdentity, err := analysisPathIdentity(module.FilePath)
		if err != nil {
			return nil, fmt.Errorf("resolve module path %q: %w", module.FilePath, err)
		}
		relativePath, err := filepath.Rel(rootIdentity, moduleIdentity)
		if err != nil {
			return nil, fmt.Errorf("make module path %q relative to project root %q: %w", module.FilePath, projectRoot, err)
		}
		if pathEscapesRoot(relativePath) {
			return nil, fmt.Errorf("module path %q is outside project root %q", module.FilePath, projectRoot)
		}

		directDirectory := filepath.Dir(relativePath)
		for _, directoryPath := range directoryAncestors(directDirectory) {
			accumulator := directories[directoryPath]
			if accumulator == nil {
				accumulator = &directoryQualityAccumulator{
					metrics: domain.DirectoryQualityMetrics{DirectoryPath: directoryPath},
				}
				directories[directoryPath] = accumulator
			}
			accumulator.addModule(module, directoryPath == directDirectory)
		}
	}

	result := make([]domain.DirectoryQualityMetrics, 0, len(directories))
	for _, directory := range directories {
		directory.finishAverages()
		result = append(result, directory.metrics)
	}
	sortDirectoryQuality(result)
	return result, nil
}

func (a *directoryQualityAccumulator) addModule(module domain.ModuleQualityMetrics, direct bool) {
	a.metrics.ModuleCount++
	if direct {
		a.metrics.DirectModuleCount++
	}
	a.metrics.LinesOfCode += module.LinesOfCode
	a.metrics.FunctionCount += module.FunctionCount
	a.metrics.AnalyzedFunctionCount += module.AnalyzedFunctionCount
	a.totalComplexity += module.AverageComplexity * float64(module.AnalyzedFunctionCount)
	a.totalCognitiveComplexity += module.AverageCognitiveComplexity * float64(module.AnalyzedFunctionCount)
	if module.MaxComplexity > a.metrics.MaxComplexity {
		a.metrics.MaxComplexity = module.MaxComplexity
	}
	a.metrics.HighRiskFunctionCount += module.HighRiskFunctionCount
	a.metrics.ExceptionHandlerCount += module.ExceptionHandlerCount
	a.metrics.DeadCodeFindingCount += module.DeadCodeFindingCount
	a.metrics.DeadCodeBlockCount += module.DeadCodeBlockCount
}

func (a *directoryQualityAccumulator) finishAverages() {
	if a.metrics.AnalyzedFunctionCount == 0 {
		return
	}
	count := float64(a.metrics.AnalyzedFunctionCount)
	a.metrics.AverageComplexity = a.totalComplexity / count
	a.metrics.AverageCognitiveComplexity = a.totalCognitiveComplexity / count
}

func directoryAncestors(path string) []string {
	cleaned := filepath.Clean(path)
	if cleaned == "." {
		return []string{"."}
	}

	parts := strings.Split(cleaned, string(filepath.Separator))
	ancestors := make([]string, 1, len(parts)+1)
	ancestors[0] = "."
	current := ""
	for _, part := range parts {
		if current == "" {
			current = part
		} else {
			current = filepath.Join(current, part)
		}
		ancestors = append(ancestors, current)
	}
	return ancestors
}

func pathEscapesRoot(relativePath string) bool {
	return relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator))
}

func sortDirectoryQuality(directories []domain.DirectoryQualityMetrics) {
	sort.Slice(directories, func(i, j int) bool {
		left, right := directories[i], directories[j]
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
		return left.DirectoryPath < right.DirectoryPath
	})
}
