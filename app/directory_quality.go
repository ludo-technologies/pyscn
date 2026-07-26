package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ludo-technologies/pyscn/domain"
)

// complexityDirectoryRoot returns the common directory explicitly selected by
// the caller. Directory inputs participate directly; file inputs participate
// through their parent directory. It never widens scope via project markers.
func complexityDirectoryRoot(paths []string) (string, error) {
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
		if filepath.VolumeName(root) != filepath.VolumeName(identity) {
			return "", fmt.Errorf("analysis paths do not share a filesystem volume")
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

type directoryComplexityAccumulator struct {
	metrics           domain.DirectoryComplexityMetrics
	totalComplexity   int
	totalNestingDepth int
}

// aggregateComplexityByDirectory groups the reported function population by
// its direct project-root-relative directory. This is the only owner of
// directory grouping and directory-level complexity arithmetic.
func aggregateComplexityByDirectory(functions []domain.FunctionComplexity, projectRoot string) ([]domain.DirectoryComplexityMetrics, error) {
	if len(functions) == 0 {
		return nil, nil
	}

	rootIdentity, err := analysisPathIdentity(projectRoot)
	if err != nil {
		return nil, fmt.Errorf("resolve complexity directory root %q: %w", projectRoot, err)
	}

	directories := make(map[string]*directoryComplexityAccumulator)
	for _, function := range functions {
		fileIdentity, err := analysisPathIdentity(function.FilePath)
		if err != nil {
			return nil, fmt.Errorf("resolve function file path %q: %w", function.FilePath, err)
		}
		relativePath, err := filepath.Rel(rootIdentity, fileIdentity)
		if err != nil {
			return nil, fmt.Errorf("make function file path %q relative to root %q: %w", function.FilePath, projectRoot, err)
		}
		if pathEscapesRoot(relativePath) {
			return nil, fmt.Errorf("function file path %q is outside complexity directory root %q", function.FilePath, projectRoot)
		}

		directoryPath := filepath.Dir(relativePath)
		accumulator := directories[directoryPath]
		if accumulator == nil {
			accumulator = &directoryComplexityAccumulator{
				metrics: domain.DirectoryComplexityMetrics{DirectoryPath: directoryPath},
			}
			directories[directoryPath] = accumulator
		}
		accumulator.addFunction(function)
	}

	result := make([]domain.DirectoryComplexityMetrics, 0, len(directories))
	for _, directory := range directories {
		directory.finishAverages()
		result = append(result, directory.metrics)
	}
	sortDirectoryComplexity(result)
	return result, nil
}

func (a *directoryComplexityAccumulator) addFunction(function domain.FunctionComplexity) {
	a.metrics.FunctionCount++
	a.totalComplexity += function.Metrics.Complexity
	a.totalNestingDepth += function.Metrics.NestingDepth
	if function.Metrics.Complexity > a.metrics.MaxComplexity {
		a.metrics.MaxComplexity = function.Metrics.Complexity
	}
	if function.Metrics.NestingDepth > a.metrics.MaxNestingDepth {
		a.metrics.MaxNestingDepth = function.Metrics.NestingDepth
	}
	if function.RiskLevel == domain.RiskLevelHigh {
		a.metrics.HighRiskFunctionCount++
	}
}

func (a *directoryComplexityAccumulator) finishAverages() {
	count := float64(a.metrics.FunctionCount)
	a.metrics.AverageComplexity = float64(a.totalComplexity) / count
	a.metrics.AverageNestingDepth = float64(a.totalNestingDepth) / count
}

func pathEscapesRoot(relativePath string) bool {
	return relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator))
}

func sortDirectoryComplexity(directories []domain.DirectoryComplexityMetrics) {
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
		return left.DirectoryPath < right.DirectoryPath
	})
}
